package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
)

type cli struct {
	Repo         string `kong:"name='repo',env='REPO',default='moby/buildkit',help='GitHub repository name.'"`
	Token        string `kong:"name='token',env='GITHUB_TOKEN',required,help='GitHub token.'"`
	Refs         string `kong:"name='refs',env='REFS',default='master',help='Comma-separated list of refs to consider.'"`
	LastDays     int    `kong:"name='last-days',env='LAST_DAYS',default='7',help='Return last merge commit for a number of days.'"`
	LastReleases int    `kong:"name='last-releases',env='LAST_RELEASES',default='3',help='Return last feature releases.'"`
	ResultFile   string `kong:"name='result-file',env='RESULT_FILE',default='./bin/candidates.json',help='File to write the result to.'"`
}

type Result struct {
	Refs     map[string]string `json:"refs"`
	Releases map[string]string `json:"releases"`
	Commits  map[string]string `json:"commits"`
}

var (
	flags cli
)

func run() error {
	kong.Parse(&flags,
		kong.Name("refcandidates"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	log.SetFlags(0)

	client, err := newGitHubClient(flags.Repo, flags.Token)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}

	var res Result

	refs, err := getRefsCandidates(client, strings.Split(flags.Refs, ","))
	if err != nil {
		return errors.Wrap(err, "failed to get refs candidates")
	}
	res.Refs = refs

	releases, err := getReleasesCandidates(client, flags.LastReleases)
	if err != nil {
		return errors.Wrap(err, "failed to get releases candidates")
	}
	res.Releases = releases

	commits, err := getCommitsCandidates(client, flags.LastDays, refs, releases)
	if err != nil {
		return errors.Wrap(err, "failed to get commits candidates")
	}
	res.Commits = commits

	dt, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal result")
	}

	if flags.ResultFile != "" {
		if err := os.MkdirAll(filepath.Dir(flags.ResultFile), 0755); err != nil {
			return errors.Wrap(err, "failed to create result file directory")
		}
		if err := os.WriteFile(flags.ResultFile, dt, 0644); err != nil {
			return errors.Wrap(err, "failed to write result to file")
		}
	}

	log.Printf("%s", string(dt))
	return nil
}

func getRefsCandidates(c *GitHubClient, refs []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, ref := range refs {
		commit, err := c.GetCommit(ref)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch commit for ref %q", ref)
		}
		res[ref] = commit.SHA
	}
	return res, nil
}

func getCommitsCandidates(c *GitHubClient, days int, refs map[string]string, releases map[string]string) (map[string]string, error) {
	commits, err := c.GetCommits(time.Now().AddDate(0, 0, -days))
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch commits")
	}
	res := make(map[string]string)
	for date, commit := range lastCommitByDay(filterMergeCommits(commits)) {
		// skip commits that are already in refs or releases
		if !containsValue(refs, commit.SHA) && !containsValue(releases, commit.SHA) {
			res[date] = commit.SHA
		}
	}
	return res, nil
}

func getReleasesCandidates(c *GitHubClient, last int) (map[string]string, error) {
	tags, err := c.GetTags()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch tags")
	}
	res := make(map[string]string)
	for _, tag := range filterFeatureReleases(tags, last) {
		res[tag.Name] = tag.Commit.SHA
	}
	return res, nil
}

func filterMergeCommits(commits []GitHubCommit) []GitHubCommit {
	var mergeCommits []GitHubCommit
	for _, commit := range commits {
		if len(commit.Parents) > 1 {
			mergeCommits = append(mergeCommits, commit)
		}
	}
	return mergeCommits
}

func lastCommitByDay(commits []GitHubCommit) map[string]GitHubCommit {
	lastCommits := make(map[string]GitHubCommit)
	for _, commit := range commits {
		date := commit.Commit.Committer.Date[:10]
		if existingCommit, exists := lastCommits[date]; !exists || commit.Commit.Committer.Date > existingCommit.Commit.Committer.Date {
			lastCommits[date] = commit
		}
	}
	return lastCommits
}

func filterFeatureReleases(tags []GitHubTag, last int) []GitHubTag {
	latestReleases := make(map[string]GitHubTag)
	zeroReleases := make(map[string]GitHubTag)
	for _, tag := range tags {
		if len(latestReleases) == last && len(zeroReleases) == last {
			break
		}
		if semver.IsValid(tag.Name) {
			mm := semver.MajorMinor(tag.Name)
			if getPatchVersion(tag.Name) == "0" {
				zeroReleases[mm] = tag
			}
			if t, ok := latestReleases[mm]; !ok || semver.Compare(tag.Name, t.Name) > 0 {
				latestReleases[mm] = tag
			}
		}
	}
	var res []GitHubTag
	for mm, lt := range latestReleases {
		res = append(res, lt)
		if zt, ok := zeroReleases[mm]; ok && zt.Name != lt.Name {
			res = append(res, zt)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return semver.Compare(res[i].Name, res[j].Name) > 0
	})
	return res
}

func getPatchVersion(version string) string {
	re := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)
	match := re.FindStringSubmatch(version)
	if len(match) > 1 {
		parts := strings.Split(match[1], ".")
		if len(parts) == 3 {
			return parts[2]
		}
	}
	return ""
}

func containsValue(m map[string]string, value string) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	if err := run(); err != nil {
		log.Printf("error: %+v", err)
		os.Exit(1)
	}
}
