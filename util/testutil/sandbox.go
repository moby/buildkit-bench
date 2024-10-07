package testutil

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

const buildkitdConfigFile = "buildkitd.toml"

var reTestName = regexp.MustCompile(`(.*)/ref=([^/]+)/buildx=([^/]+)/run=([^/]+)`)

type sandbox struct {
	Backend

	logs            map[string]*bytes.Buffer
	cleanup         *MultiCloser
	mv              matrixValue
	ctx             context.Context
	name            string
	buildkitBinsDir string
	outDir          string
}

func (sb *sandbox) Name() string {
	return sb.name
}

func (sb *sandbox) Context() context.Context {
	return sb.ctx
}

func (sb *sandbox) Logs() map[string]*bytes.Buffer {
	return sb.logs
}

func (sb *sandbox) PrintLogs(tb testing.TB) {
	printLogs(sb.logs, tb.Log)
}

func (sb *sandbox) WriteLogs(tb testing.TB) {
	var logs []byte
	for name, l := range sb.logs {
		logs = append(logs, []byte(fmt.Sprintf("%s\n", name))...)
		logs = append(logs, l.Bytes()...)
	}
	sb.WriteLogFile(tb, "buildkitd", logs)
}

func (sb *sandbox) WriteLogFile(tb testing.TB, name string, dt []byte) {
	tname, buildkitRef, buildxRef, run := sb.parseTestName(tb)
	testLogsDir := path.Join(sb.outDir, "logs", path.Join(strings.Split(tname, "/")...), fmt.Sprintf("buildkit-%s", buildkitRef), fmt.Sprintf("buildx-%s", buildxRef))
	if err := os.MkdirAll(testLogsDir, 0755); err != nil {
		tb.Fatalf("failed to create logs directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testLogsDir, fmt.Sprintf("%s-run%s.log", name, run)), dt, 0644); err != nil {
		tb.Fatalf("writing log file: %v", err)
	}
}

func (sb *sandbox) ClearLogs() {
	sb.logs = make(map[string]*bytes.Buffer)
}

func (sb *sandbox) Value(k string) interface{} {
	return sb.mv.values[k].value
}

func (sb *sandbox) BuildKitBinsDir() string {
	return sb.buildkitBinsDir
}

func (sb *sandbox) parseTestName(tb testing.TB) (string, string, string, string) {
	matches := reTestName.FindStringSubmatch(tb.Name())
	if len(matches) < 5 {
		tb.Fatalf("failed to parse test name: %q", tb.Name())
	}
	return matches[1], matches[2], matches[3], matches[4]
}

func newSandbox(ctx context.Context, r Worker, mirror string, mv matrixValue) (s Sandbox, cl func() error, err error) {
	cfg := &BackendConfig{
		Logs: make(map[string]*bytes.Buffer),
	}
	for k, v := range mv.values {
		if u, ok := v.value.(ConfigUpdater); ok {
			cfg.DaemonConfig = append(cfg.DaemonConfig, u)
		} else if p, ok := v.value.(string); ok && k == "buildx" {
			cfg.BuildxBin = p
		}
	}
	if mirror != "" {
		cfg.DaemonConfig = append(cfg.DaemonConfig, withMirrorConfig(mirror))
	}

	deferF := &MultiCloser{}
	cl = deferF.F()

	defer func() {
		if err != nil {
			deferF.F()()
			cl = nil
		}
	}()

	b, closer, err := r.New(ctx, cfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating ref")
	}
	deferF.Append(closer)

	return &sandbox{
		Backend:         b,
		logs:            cfg.Logs,
		cleanup:         deferF,
		mv:              mv,
		ctx:             ctx,
		name:            r.Name(),
		buildkitBinsDir: buildkitBinsDir,
		outDir:          outDir,
	}, cl, nil
}

func printLogs(logs map[string]*bytes.Buffer, f func(args ...interface{})) {
	for name, l := range logs {
		f(name)
		s := bufio.NewScanner(l)
		for s.Scan() {
			f(s.Text())
		}
	}
}
