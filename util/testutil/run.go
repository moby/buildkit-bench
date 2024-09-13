package testutil

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/gofrs/flock"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/moby/buildkit/util/contentutil"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

var sandboxLimiter *semaphore.Weighted

func init() {
	sandboxLimiter = semaphore.NewWeighted(int64(runtime.GOMAXPROCS(0)))
}

// Backend describes a testing backend.
type Backend interface {
	Address() string
	DebugAddress() string
	ExtraEnv() []string
	BuildxDir() string
	BuilderName() string
}

type Sandbox interface {
	Backend

	Context() context.Context
	Logs() map[string]*bytes.Buffer
	PrintLogs(testing.TB)
	ClearLogs()
	Value(string) interface{} // chosen matrix value
	Name() string
	BinsDir() string
}

// BackendConfig is used to configure backends created by a worker.
type BackendConfig struct {
	Logs         map[string]*bytes.Buffer
	DaemonConfig []ConfigUpdater
}

type Worker interface {
	New(context.Context, *BackendConfig) (Backend, func() error, error)
	Name() string
}

type ConfigUpdater interface {
	UpdateConfigFile(string) string
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

func WithMatrix(key string, m map[string]interface{}) TestOpt {
	return func(tc *testConf) {
		if tc.matrix == nil {
			tc.matrix = map[string]map[string]interface{}{}
		}
		tc.matrix[key] = m
	}
}

func WithMirroredImages(m map[string]string) TestOpt {
	return func(tc *testConf) {
		if tc.mirroredImages == nil {
			tc.mirroredImages = map[string]string{}
		}
		maps.Copy(tc.mirroredImages, m)
	}
}

type testConf struct {
	matrix         map[string]map[string]interface{}
	mirroredImages map[string]string
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

	getMirror := lazyMirrorRunnerFunc(tb, tc.mirroredImages)

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

							runWithSandbox := func(tb testing.TB) {
								sb, closer, err := newSandbox(ctx, br, getMirror(), mv)
								require.NoError(tb, err)
								tb.Cleanup(func() { _ = closer() })
								defer func() {
									if tb.Failed() {
										sb.PrintLogs(tb)
									}
								}()
								runner.Run(tb, sb)
							}
							if b, ok := tb.(*testing.B); ok {
								for i := 0; i < b.N; i++ {
									runWithSandbox(b)
								}
							} else {
								runWithSandbox(tb)
							}
						})
						require.True(tb, ok)
					}(fn, name, br, runner, mv)
				}
			}
		}
	}
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

func writeConfig(updaters []ConfigUpdater) (string, error) {
	tmpdir, err := os.MkdirTemp("", "bkbench_config")
	if err != nil {
		return "", err
	}
	if err := os.Chmod(tmpdir, 0711); err != nil {
		return "", err
	}

	s := ""
	for _, upt := range updaters {
		s = upt.UpdateConfigFile(s)
	}

	if err := os.WriteFile(filepath.Join(tmpdir, buildkitdConfigFile), []byte(s), 0644); err != nil {
		return "", err
	}
	return filepath.Join(tmpdir, buildkitdConfigFile), nil
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

func lazyMirrorRunnerFunc(tb testing.TB, images map[string]string) func() string {
	var once sync.Once
	var mirror string
	return func() string {
		once.Do(func() {
			host, cleanup, err := runMirror(tb, images)
			require.NoError(tb, err)
			tb.Cleanup(func() { _ = cleanup() })
			mirror = host
		})
		return mirror
	}
}

func runMirror(tb testing.TB, mirroredImages map[string]string) (host string, _ func() error, err error) {
	mirrorDir := os.Getenv("REGISTRY_MIRROR_DIR")

	var lock *flock.Flock
	if mirrorDir != "" {
		if err := os.MkdirAll(mirrorDir, 0700); err != nil {
			return "", nil, err
		}
		lock = flock.New(filepath.Join(mirrorDir, "lock"))
		if err := lock.Lock(); err != nil {
			return "", nil, err
		}
		defer func() {
			if err != nil {
				lock.Unlock()
			}
		}()
	}

	mirror, cleanup, err := NewRegistry(mirrorDir)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if err != nil {
			cleanup()
		}
	}()

	if err := copyImagesLocal(tb, mirror, mirroredImages); err != nil {
		return "", nil, err
	}

	if mirrorDir != "" {
		if err := lock.Unlock(); err != nil {
			return "", nil, err
		}
	}

	return mirror, cleanup, err
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

var localImageCache map[string]map[string]struct{}

func copyImagesLocal(tb testing.TB, host string, images map[string]string) error {
	for to, from := range images {
		if localImageCache == nil {
			localImageCache = map[string]map[string]struct{}{}
		}
		if _, ok := localImageCache[host]; !ok {
			localImageCache[host] = map[string]struct{}{}
		}
		if _, ok := localImageCache[host][to]; ok {
			continue
		}
		localImageCache[host][to] = struct{}{}

		// already exists check
		if _, _, err := docker.NewResolver(docker.ResolverOptions{}).Resolve(context.TODO(), host+"/"+to); err == nil {
			continue
		}

		var desc ocispecs.Descriptor
		var provider content.Provider
		var err error
		if strings.HasPrefix(from, "local:") {
			var closer func()
			desc, provider, closer, err = providerFromBinary(strings.TrimPrefix(from, "local:"))
			if err != nil {
				return err
			}
			if closer != nil {
				defer closer()
			}
		} else {
			desc, provider, err = contentutil.ProviderFromRef(from)
			if err != nil {
				return err
			}
		}

		ingester, err := contentutil.IngesterFromRef(host + "/" + to)
		if err != nil {
			return err
		}
		if err := contentutil.CopyChain(context.TODO(), ingester, provider, desc); err != nil {
			return err
		}
		tb.Logf("copied %s to local mirror %s", from, host+"/"+to)
	}
	return nil
}

func OfficialImages(names ...string) map[string]string {
	return officialImages(names...)
}

func withMirrorConfig(mirror string) ConfigUpdater {
	return mirrorConfig(mirror)
}

type mirrorConfig string

func (mc mirrorConfig) UpdateConfigFile(in string) string {
	return fmt.Sprintf(`%s

[registry."docker.io"]
mirrors=["%s"]
`, in, mc)
}

func officialImages(names ...string) map[string]string {
	ns := runtime.GOARCH
	if ns == "arm64" {
		ns = "arm64v8"
	} else if ns != "amd64" {
		ns = "library"
	}
	m := map[string]string{}
	for _, name := range names {
		ref := "docker.io/" + ns + "/" + name
		if pns, ok := pins[name]; ok {
			if dgst, ok := pns[ns]; ok {
				ref += "@" + dgst
			}
		}
		m["library/"+name] = ref
	}
	return m
}

func getFunctionName(i interface{}) string {
	fullname := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	dot := strings.LastIndex(fullname, ".") + 1
	return strings.Title(fullname[dot:]) //nolint:staticcheck // ignoring "SA1019: strings.Title is deprecated", as for our use we don't need full unicode support
}
