package testutil

import (
	"context"
	"maps"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

var sandboxLimiter *semaphore.Weighted

func init() {
	sandboxLimiter = semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
}

// Backend describes a testing backend.
type Backend struct{}

type Sandbox interface {
	Context() context.Context
	Value(string) interface{} // chosen matrix value
	Name() string
	BinsDir() string
}

type Worker interface {
	New(context.Context) (Backend, func() error, error)
	Name() string
}

type Runner interface {
	Name() string
	Run(tb testing.TB, sb Sandbox)
}

type Test interface {
	Runner
}

type testFunc struct {
	name string
	run  func(t *testing.T, sb Sandbox)
}

func (f testFunc) Name() string {
	return f.name
}

func (f testFunc) Run(tb testing.TB, sb Sandbox) {
	tb.Helper()
	f.run(tb.(*testing.T), sb)
}

func TestFuncs(funcs ...func(t *testing.T, sb Sandbox)) []Runner {
	var tests []Runner
	names := map[string]struct{}{}
	for _, f := range funcs {
		name := getFunctionName(f)
		if _, ok := names[name]; ok {
			panic("duplicate test: " + name)
		}
		names[name] = struct{}{}
		tests = append(tests, testFunc{name: name, run: f})
	}
	return tests
}

type Bench interface {
	Runner
}

type benchFunc struct {
	name string
	run  func(b *testing.B, sb Sandbox)
}

func (f benchFunc) Name() string {
	return f.name
}

func (f benchFunc) Run(tb testing.TB, sb Sandbox) {
	tb.Helper()
	f.run(tb.(*testing.B), sb)
}

func BenchFuncs(funcs ...func(b *testing.B, sb Sandbox)) []Runner {
	var benchs []Runner
	names := map[string]struct{}{}
	for _, f := range funcs {
		name := getFunctionName(f)
		if _, ok := names[name]; ok {
			panic("duplicate bench: " + name)
		}
		names[name] = struct{}{}
		benchs = append(benchs, benchFunc{name: name, run: f})
	}
	return benchs
}

var defaultWorkers []Worker

func Register(w Worker) {
	defaultWorkers = append(defaultWorkers, w)
}

func List() []Worker {
	return defaultWorkers
}

// TestOpt is an option that can be used to configure a set of tests.
type TestOpt func(*testConf)

type testConf struct {
	matrix map[string]map[string]interface{}
}

func RunTest(tb testing.TB, testFunc func(tb testing.TB, sb Sandbox)) {
	switch t := tb.(type) {
	case *testing.T:
		Run(t, TestFuncs(func(t *testing.T, sb Sandbox) {
			testFunc(t, sb)
		}))
	case *testing.B:
		Run(t, BenchFuncs(func(b *testing.B, sb Sandbox) {
			testFunc(b, sb)
		}))
	}
}

func Run(tb testing.TB, runners []Runner, opt ...TestOpt) {
	var tc testConf
	for _, o := range opt {
		o(&tc)
	}

	matrix := prepareValueMatrix(tc)

	list := List()
	if os.Getenv("BUILDKIT_REF_RANDOM") == "1" && len(list) > 0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // using math/rand is fine in a test utility
		list = []Worker{list[rng.Intn(len(list))]}
	}

	runCount := 1
	if v, ok := os.LookupEnv("TEST_BENCH_RUN"); ok && v != "" {
		if _, ok := tb.(*testing.B); ok {
			if count, err := strconv.Atoi(v); err != nil {
				tb.Fatalf("invalid TEST_BENCH_RUN value: %v", err)
			} else {
				runCount = count
			}
		}
	}

	for _, br := range list {
		for _, runner := range runners {
			for _, mv := range matrix {
				for rc := 1; rc <= runCount; rc++ {
					fn := runner.Name()
					name := fn + "/ref=" + br.Name() + mv.functionSuffix() + "/run=" + strconv.Itoa(rc)
					func(fn, testName string, br Worker, runner Runner, mv matrixValue) {
						ok := runTest(tb, testName, func(tb testing.TB) {
							ctx := appcontext.Context()
							require.NoError(tb, sandboxLimiter.Acquire(context.TODO(), 1))
							defer sandboxLimiter.Release(1)

							ctx, cancel := context.WithCancelCause(ctx)
							defer cancel(errors.WithStack(context.Canceled))

							sb, _, err := newSandbox(ctx, br, mv)
							require.NoError(tb, err)
							runner.Run(tb, sb)
						})
						require.True(tb, ok)
					}(fn, name, br, runner, mv)
				}
			}
		}
	}
}

func ReportMetric(b *testing.B, value float64, unit string) {
	b.ReportMetric(value, "bkbench_"+unit)
}

func runTest(tb testing.TB, name string, f func(tb testing.TB)) bool {
	switch t := tb.(type) {
	case *testing.T:
		return t.Run(name, func(t *testing.T) {
			f(t)
		})
	case *testing.B:
		return t.Run(name, func(b *testing.B) {
			f(b)
		})
	default:
		tb.Fatalf("unsupported testing.TB type: %T", tb)
		return false
	}
}

func prepareValueMatrix(tc testConf) []matrixValue {
	var m []matrixValue
	for featureName, values := range tc.matrix {
		current := m
		m = []matrixValue{}
		for featureValue, v := range values {
			if len(current) == 0 {
				m = append(m, newMatrixValue(featureName, featureValue, v))
			}
			for _, c := range current {
				vv := newMatrixValue(featureName, featureValue, v)
				vv.fn = append(vv.fn, c.fn...)
				maps.Copy(vv.values, c.values)
				m = append(m, vv)
			}
		}
	}
	if len(m) == 0 {
		m = append(m, matrixValue{})
	}
	return m
}

type matrixValue struct {
	fn     []string
	values map[string]matrixValueChoice
}

func (mv matrixValue) functionSuffix() string {
	if len(mv.fn) == 0 {
		return ""
	}
	sort.Strings(mv.fn)
	sb := &strings.Builder{}
	for _, f := range mv.fn {
		sb.Write([]byte("/" + f + "=" + mv.values[f].name))
	}
	return sb.String()
}

type matrixValueChoice struct {
	name  string
	value interface{}
}

func newMatrixValue(key, name string, v interface{}) matrixValue {
	return matrixValue{
		fn: []string{key},
		values: map[string]matrixValueChoice{
			key: {
				name:  name,
				value: v,
			},
		},
	}
}

func getFunctionName(i interface{}) string {
	fullname := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	dot := strings.LastIndex(fullname, ".") + 1
	return strings.Title(fullname[dot:]) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support
}
