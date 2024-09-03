package testutil

import (
	"context"
	"log"
	"os"
	"path"
)

var binsDir = "/buildkit-binaries"

type backend struct {
	address  string
	extraEnv []string
}

func (b backend) Address() string {
	return b.address
}

func (b backend) ExtraEnv() []string {
	return b.extraEnv
}

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

func (c *Ref) New(ctx context.Context, cfg *BackendConfig) (b Backend, cl func() error, err error) {
	deferF := &MultiCloser{}
	cl = deferF.F()

	defer func() {
		if err != nil {
			deferF.F()()
			cl = nil
		}
	}()

	// Include use of --oci-worker-labels to trigger https://github.com/moby/buildkit/pull/603
	buildkitdSock, stop, err := runBuildkitd(cfg, []string{
		path.Join(binsDir, c.id, "buildkitd"),
		"--oci-worker=true",
		"--oci-worker-binary=" + path.Join(binsDir, c.id, "buildkit-runc"),
		"--containerd-worker=false",
		"--oci-worker-gc=false",
		"--oci-worker-labels=org.mobyproject.buildkit.worker.sandbox=true",
	}, cfg.Logs, nil)
	if err != nil {
		printLogs(cfg.Logs, log.Println)
		return nil, nil, err
	}
	deferF.Append(stop)

	return backend{
		address: buildkitdSock,
	}, cl, nil
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
