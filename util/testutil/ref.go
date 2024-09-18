package testutil

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/identity"
	"github.com/pkg/errors"
)

var (
	binsDir = "/buildkit-binaries"
	outDir  = "/testout"
)

type backend struct {
	address      string
	debugAddress string
	extraEnv     []string
	buildxDir    string
	builderName  string
}

func (b backend) Address() string {
	return b.address
}

func (b backend) DebugAddress() string {
	return b.debugAddress
}

func (b backend) ExtraEnv() []string {
	return b.extraEnv
}

func (b backend) BuildxDir() string {
	return b.buildxDir
}

func (b backend) BuilderName() string {
	return b.builderName
}

func init() {
	if v := os.Getenv("BUILDKIT_BINS_DIR"); v != "" {
		binsDir = v
	}
	if v := os.Getenv("TEST_OUT_DIR"); v != "" {
		outDir = v
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

	buildkitdPath := path.Join(binsDir, c.id, "buildkitd")
	if err := lookupBinary(buildkitdPath); err != nil {
		return nil, nil, err
	}

	tmpdir, err := os.MkdirTemp("", "bkbench_sandbox")
	if err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(filepath.Join(tmpdir, "tmp"), 0711); err != nil {
		return nil, nil, err
	}
	deferF.Append(func() error { return os.RemoveAll(tmpdir) })

	// Include use of --oci-worker-labels to trigger https://github.com/moby/buildkit/pull/603
	buildkitdSock, debugAddress, stop, err := runBuildkitd(cfg, tmpdir, []string{
		buildkitdPath,
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

	if err := lookupBinary("buildx"); err != nil {
		return nil, nil, err
	}

	// Create a remote buildx instance
	builderName := "remote-" + identity.NewID()
	buildxDir := filepath.Join(tmpdir, "buildx")
	cmd := exec.Command("buildx", "create",
		"--bootstrap",
		"--name", builderName,
		"--driver", "remote",
		buildkitdSock,
	)
	cmd.Env = append(os.Environ(), "BUILDX_CONFIG="+buildxDir)
	if err := cmd.Run(); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create buildx instance %s", builderName)
	}
	deferF.Append(func() error {
		cmd := exec.Command("buildx", "rm", "-f", builderName)
		cmd.Env = append(os.Environ(), "BUILDX_CONFIG="+buildxDir)
		return cmd.Run()
	})

	// separated out since it's not required in windows
	deferF.Append(func() error {
		return mountInfo(tmpdir)
	})

	return backend{
		address:      buildkitdSock,
		debugAddress: debugAddress,
		buildxDir:    buildxDir,
		builderName:  builderName,
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

func mountInfo(tmpdir string) error {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return errors.Wrap(err, "failed to open mountinfo")
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.Contains(s.Text(), tmpdir) {
			return errors.Errorf("leaked mountpoint for %s", tmpdir)
		}
	}
	return s.Err()
}
