package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
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
	client, err := newGitHubClient(cli.Repo, cli.Token)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}
	c, err := getCandidates(client, cli.Refs, cli.LastDays, cli.LastReleases)
	if err != nil {
		return errors.Wrap(err, "failed to get candidates")
	}
	if cli.FileOutput != "" {
		if err := c.WriteFile(cli.FileOutput); err != nil {
			return errors.Wrap(err, "failed to write candidates to output file")
		}
	}
	if cli.GhaOutput != "" {
		if err := c.setGhaOutput(cli.GhaOutput); err != nil {
			return errors.Wrap(err, "failed to set GitHub Actions matrix")
		}
	}
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
