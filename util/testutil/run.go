package testutil

import (
	"context"
	"maps"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"sort"
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

type Test interface {
	Name() string
	Run(t *testing.T, sb Sandbox)
}

type testFunc struct {
	name string
	run  func(t *testing.T, sb Sandbox)
}

func (f testFunc) Name() string {
	return f.name
}

func (f testFunc) Run(t *testing.T, sb Sandbox) {
	t.Helper()
	f.run(t, sb)
}

func TestFuncs(funcs ...func(t *testing.T, sb Sandbox)) []Test {
	var tests []Test
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

func Run(t *testing.T, testCases []Test, opt ...TestOpt) {
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

	for _, br := range list {
		for _, tc := range testCases {
			for _, mv := range matrix {
				fn := tc.Name()
				name := fn + "/ref=" + br.Name() + mv.functionSuffix()
				func(fn, testName string, br Worker, tc Test, mv matrixValue) {
					ok := t.Run(testName, func(t *testing.T) {
						ctx := appcontext.Context()
						require.NoError(t, sandboxLimiter.Acquire(context.TODO(), 1))
						defer sandboxLimiter.Release(1)

						ctx, cancel := context.WithCancelCause(ctx)
						defer cancel(errors.WithStack(context.Canceled))

						sb, _, err := newSandbox(ctx, br, mv)
						require.NoError(t, err)
						tc.Run(t, sb)
					})
					require.True(t, ok)
				}(fn, name, br, tc, mv)
			}
		}
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
