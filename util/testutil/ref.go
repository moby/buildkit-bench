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
	if refs, err := getRefs(binsDir); err == nil {
		for _, ref := range refs {
			Register(&Ref{ID: ref})
		}
	} else {
		panic(err)
	}
}

type Ref struct {
	ID string
}

func (c *Ref) Name() string {
	return c.ID
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

func getRefs(dir string) ([]string, error) {
	var refs []string
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				refs = append(refs, entry.Name())
			}
		}
	}
	return refs, nil
}
