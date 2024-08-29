package testutil

import (
	"context"
	"os"
)

var binsDir = "/buildkit-binaries"

func init() {
	if v := os.Getenv("BUILDKIT_BINS_DIR"); v != "" {
		binsDir = v
	}
	for _, ref := range getRefs(binsDir) {
		Register(ref)
	}
}

type Ref struct {
	id string
}

func (c *Ref) Name() string {
	return c.id
}

func (c *Ref) New(ctx context.Context) (b Backend, cl func() error, err error) {
	deferF := &MultiCloser{}
	cl = deferF.F()

	defer func() {
		if err != nil {
			deferF.F()()
			cl = nil
		}
	}()

	return Backend{}, cl, nil
}

func getRefs(dir string) []*Ref {
	var refs []*Ref
	entries, err := os.ReadDir(dir)
	if err != nil {
		return refs
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		refs = append(refs, &Ref{id: entry.Name()})
	}
	return refs
}
