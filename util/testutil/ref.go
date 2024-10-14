package testutil

import (
	"bufio"
	"context"
	"fmt"
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
	buildkitBinsDir    = "/buildkit-binaries"
	buildkitBinsAltDir = "/opt/buildkit"
	buildxBinsDir      = "/buildx-binaries"
	buildxBinsAltDir   = "/opt/buildx"
	outDir             = "/testout"
)

type backend struct {
	address         string
	debugAddress    string
	extraEnv        []string
	buildxBin       string
	buildxConfigDir string
	builderName     string
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

func (b backend) BuildxBin() string {
	return b.buildxBin
}

func (b backend) BuildxConfigDir() string {
	return b.buildxConfigDir
}

func (b backend) BuilderName() string {
	return b.builderName
}

func init() {
	buildkitBinsDir = getBinsDir(buildkitBinsDir, buildkitBinsAltDir, "BUILDKIT_BINS_DIR")
	buildxBinsDir = getBinsDir(buildxBinsDir, buildxBinsAltDir, "BUILDX_BINS_DIR")
	if v := os.Getenv("TEST_OUT_DIR"); v != "" {
		outDir = v
	}
	for _, ref := range getRefs(buildkitBinsDir) {
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

	buildkitdPath := path.Join(buildkitBinsDir, c.id, "buildkitd")
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
		"--oci-worker-binary=" + path.Join(buildkitBinsDir, c.id, "buildkit-runc"),
		"--containerd-worker=false",
		"--oci-worker-gc=false",
		"--oci-worker-labels=org.mobyproject.buildkit.worker.sandbox=true",
	}, cfg.Logs, []string{
		fmt.Sprintf("PATH=%s:%s", path.Join(buildkitBinsDir, c.id), os.Getenv("PATH")),
	})
	if err != nil {
		printLogs(cfg.Logs, log.Println)
		return nil, nil, err
	}
	deferF.Append(stop)

	if err := lookupBinary(cfg.BuildxBin); err != nil {
		return nil, nil, err
	}

	// Create a remote buildx instance
	builderName := "remote-" + identity.NewID()
	buildxConfigDir := filepath.Join(tmpdir, "buildx")
	cmd := exec.Command(cfg.BuildxBin, "create",
		"--bootstrap",
		"--name", builderName,
		"--driver", "remote",
		buildkitdSock,
	)
	cmd.Env = append(os.Environ(), "BUILDX_CONFIG="+buildxConfigDir)
	if err := cmd.Run(); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create buildx instance %s", builderName)
	}
	deferF.Append(func() error {
		cmd := exec.Command(cfg.BuildxBin, "rm", "-f", builderName)
		cmd.Env = append(os.Environ(), "BUILDX_CONFIG="+buildxConfigDir)
		return cmd.Run()
	})

	// separated out since it's not required in windows
	deferF.Append(func() error {
		return mountInfo(tmpdir)
	})

	return backend{
		address:         buildkitdSock,
		debugAddress:    debugAddress,
		buildxBin:       cfg.BuildxBin,
		buildxConfigDir: buildxConfigDir,
		builderName:     builderName,
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

func getBinsDir(targetdir string, altdir string, env string) string {
	if v := os.Getenv(env); v != "" {
		targetdir = v
	}
	f, err := os.Open(targetdir)
	if err != nil {
		return altdir
	}
	defer f.Close()
	if n, err := f.Readdirnames(1); err != nil || len(n) == 0 {
		return altdir
	}
	return targetdir
}
