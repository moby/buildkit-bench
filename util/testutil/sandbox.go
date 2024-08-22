package testutil

import (
	"context"

	"github.com/pkg/errors"
)

type sandbox struct {
	Backend

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

func (sb *sandbox) Value(k string) interface{} {
	return sb.mv.values[k].value
}

func (sb *sandbox) BinsDir() string {
	return sb.binsDir
}

func newSandbox(ctx context.Context, r Worker, mv matrixValue) (s Sandbox, cl func() error, err error) {
	deferF := &MultiCloser{}
	cl = deferF.F()

	defer func() {
		if err != nil {
			deferF.F()()
			cl = nil
		}
	}()

	b, closer, err := r.New(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating ref")
	}
	deferF.Append(closer)

	return &sandbox{
		Backend: b,
		cleanup: deferF,
		mv:      mv,
		ctx:     ctx,
		name:    r.Name(),
		binsDir: binsDir,
	}, cl, nil
}
