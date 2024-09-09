package testutil

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
)

const buildkitdConfigFile = "buildkitd.toml"

type sandbox struct {
	Backend

	logs    map[string]*bytes.Buffer
	cleanup *MultiCloser
	mv      matrixValue
	ctx     context.Context
	name    string
	binsDir string
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

func (sb *sandbox) ClearLogs() {
	sb.logs = make(map[string]*bytes.Buffer)
}

func (sb *sandbox) Value(k string) interface{} {
	return sb.mv.values[k].value
}

func (sb *sandbox) BinsDir() string {
	return sb.binsDir
}

func newSandbox(ctx context.Context, r Worker, mirror string, mv matrixValue) (s Sandbox, cl func() error, err error) {
	cfg := &BackendConfig{
		Logs: make(map[string]*bytes.Buffer),
	}
	for _, v := range mv.values {
		if u, ok := v.value.(ConfigUpdater); ok {
			cfg.DaemonConfig = append(cfg.DaemonConfig, u)
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
		Backend: b,
		logs:    cfg.Logs,
		cleanup: deferF,
		mv:      mv,
		ctx:     ctx,
		name:    r.Name(),
		binsDir: binsDir,
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
