package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/moby/buildkit-bench/util/candidates"
	"github.com/moby/buildkit-bench/util/github"
	"github.com/moby/buildkit-bench/util/github/gha"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
)

var cli struct {
	Repo         string `kong:"name='repo',default='moby/buildkit',help='GitHub repository name.'"`
	Token        string `kong:"name='token',env='GITHUB_TOKEN',required,help='GitHub API token.'"`
	Refs         string `kong:"name='refs',default='master',help='Comma-separated list of refs to consider.'"`
	LastDays     int    `kong:"name='last-days',default='7',help='Return last merge commit for a number of days.'"`
	LastReleases int    `kong:"name='last-releases',default='3',help='Return last feature releases.'"`
	FileOutput   string `kong:"name='file-output',help='File to write the JSON output to.'"`
	GhaOutput    string `kong:"name='gha-output',help='Set GitHub Actions output parameter to be used as matrix includes.'"`
}

func run() error {
	client, err := github.NewClient(cli.Repo, cli.Token)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}

	c, err := candidates.New(client, cli.Refs, cli.LastDays, cli.LastReleases)
	if err != nil {
		return errors.Wrap(err, "failed to create candidates")
	}
	log.Printf("%d ref(s), %d release(s) and %d commit(s) marked as candidates", len(c.Refs), len(c.Releases), len(c.Commits))

	if cli.FileOutput != "" {
		if err := writeFile(cli.FileOutput, c); err != nil {
			return errors.Wrap(err, "failed to write candidates to output file")
		}
	}
	if cli.GhaOutput != "" {
		if err := setGhaOutput(cli.GhaOutput, c); err != nil {
			return errors.Wrap(err, "failed to set GitHub Actions output")
		}
	}
	return nil
}

func writeFile(f string, c *candidates.Candidates) error {
	if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
		return errors.Wrap(err, "failed to create output file directory")
	}
	dt, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal candidates")
	}
	if err := os.WriteFile(f, dt, 0644); err != nil {
		return errors.Wrap(err, "failed to write candidates to output file")
	}
	log.Printf("Candidates written to %q", f)
	return nil
}

func setGhaOutput(name string, c *candidates.Candidates) error {
	if !gha.IsRunning() {
		return errors.New("not running in GitHub Actions")
	}

	type include struct {
		Name   string `json:"name"`
		Ref    string `json:"ref"`
		Commit string `json:"commit"`
	}

	var includes []include
	for ref, cm := range c.Refs {
		includes = append(includes, include{
			Name:   ref,
			Ref:    ref,
			Commit: cm.SHA,
		})
	}
	for release, cm := range c.Releases {
		includes = append(includes, include{
			Name:   release,
			Ref:    release,
			Commit: cm.SHA,
		})
	}
	for day, cm := range c.Commits {
		includes = append(includes, include{
			Name:   day,
			Ref:    cm.SHA,
			Commit: cm.SHA,
		})
	}

	if gha.IsPullRequestEvent() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if len(includes) > 2 {
			s := make([]include, 0, 2)
			si := make(map[int]struct{})
			for len(s) < 2 {
				idx := r.Intn(len(includes))
				if _, exists := si[idx]; !exists {
					si[idx] = struct{}{}
					s = append(s, includes[idx])
				}
			}
			includes = s
		}
		log.Printf("Reducing candidates to %d for pull request event in GHA output", len(includes))
	}

	dt, err := json.Marshal(includes)
	if err != nil {
		return errors.Wrap(err, "failed to marshal includes")
	}
	githubactions.SetOutput(name, string(dt))
	log.Printf("GitHub Actions matrix includes set to %q", name)

	return nil
}

func main() {
	log.SetFlags(0)
	ctx := kong.Parse(&cli,
		kong.Name("refcandidates"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	ctx.FatalIfErrorf(run())
}
