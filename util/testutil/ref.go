package testutil

import (
	"context"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/moby/buildkit-bench/util/github"
	"github.com/moby/buildkit/util/bklog"
	"github.com/pkg/errors"
)

var (
	binsDir      = "/buildkit-binaries"
	githubClient *github.Client
)

func init() {
	var err error
	if v := os.Getenv("BUILDKIT_BINS_DIR"); v != "" {
		binsDir = v
	}
	if repo := os.Getenv("BUILDKIT_REPO"); repo != "" {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			githubClient, err = github.NewClient(repo, token)
			if err != nil {
				bklog.L.Errorf("error creating github client: %v", err)
			}
		}
	}
	for _, ref := range getRefs(binsDir) {
		Register(ref)
	}
}

type Ref struct {
	id            string
	committerDate time.Time
}

func (c *Ref) Name() string {
	return c.id
}

func (c *Ref) CommitterDate() time.Time {
	return c.committerDate
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
		if ref, err := getRef(dir, entry); err == nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

func getRef(dir string, entry os.DirEntry) (*Ref, error) {
	ref := &Ref{id: entry.Name()}
	buildkitdPath := path.Join(dir, entry.Name(), "buildkitd")
	output, err := exec.Command(buildkitdPath, "--version").Output()
	if err != nil {
		return ref, errors.Wrap(err, "error running buildkitd --version")
	}
	versionParts := strings.Fields(string(output))
	// buildkitd github.com/moby/buildkit v0.9.3 8d2625494a6a3d413e3d875a2ff7dd9b1ed1b1a9
	if len(versionParts) < 4 {
		return ref, errors.Errorf("unexpected version output: %s", string(output))
	}
	if githubClient == nil {
		return ref, nil
	}
	cm, err := githubClient.GetCommit(versionParts[3])
	if err != nil {
		return ref, errors.Wrap(err, "error getting github commit")
	}
	ref.committerDate = cm.Commit.Committer.Date
	return ref, nil
}
