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

	"github.com/moby/buildkit-bench/util/github"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/mod/semver"
)

var reSemverRelease = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)

type candidates struct {
	res struct {
		Refs     map[string]string `json:"refs"`
		Releases map[string]string `json:"releases"`
		Commits  map[string]string `json:"commits"`
	}
	ghc *github.Client
}

func getCandidates(ghc *github.Client, refs string, lastDays int, lastReleases int) (*candidates, error) {
	c := &candidates{
		ghc: ghc,
	}
	if err := c.setRefs(strings.Split(refs, ",")); err != nil {
		return nil, errors.Wrap(err, "failed to set refs candidates")
	}
	if err := c.setReleases(lastReleases); err != nil {
		return nil, errors.Wrap(err, "failed to set releases candidates")
	}
	if err := c.setCommits(lastDays); err != nil {
		return nil, errors.Wrap(err, "failed to set commits candidates")
	}
	log.Printf("%d ref(s), %d release(s) and %d commit(s) marked as candidates", len(c.res.Refs), len(c.res.Releases), len(c.res.Commits))
	return c, nil
}

func (c *candidates) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.res)
}

func (c *candidates) WriteFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrap(err, "failed to create output file directory")
	}
	dt, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal candidates")
	}
	if err := os.WriteFile(path, dt, 0644); err != nil {
		return errors.Wrap(err, "failed to write candidates to output file")
	}
	return nil
}

func (c *candidates) setGhaOutput(name string) error {
	type include struct {
		Name   string `json:"name"`
		Ref    string `json:"ref"`
		Commit string `json:"commit"`
	}
	var includes []include
	for ref, sha := range c.res.Refs {
		includes = append(includes, include{
			Name:   ref,
			Ref:    ref,
			Commit: sha,
		})
	}
	for release, sha := range c.res.Releases {
		includes = append(includes, include{
			Name:   release,
			Ref:    release,
			Commit: sha,
		})
	}
	for day, sha := range c.res.Commits {
		includes = append(includes, include{
			Name:   day,
			Ref:    sha,
			Commit: sha,
		})
	}
	dt, err := json.Marshal(includes)
	if err != nil {
		return errors.Wrap(err, "failed to marshal includes")
	}
	githubactions.SetOutput(name, string(dt))
	return nil
}

func (c *candidates) setRefs(refs []string) error {
	res := make(map[string]string)
	for _, ref := range refs {
		commit, err := c.ghc.GetCommit(ref)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch commit for ref %q", ref)
		}
		res[ref] = commit.SHA
	}
	c.res.Refs = res
	return nil
}

func (c *candidates) setReleases(last int) error {
	tags, err := c.ghc.GetTags()
	if err != nil {
		return errors.Wrap(err, "failed to fetch tags")
	}
	res := make(map[string]string)
	for _, tag := range filterFeatureReleases(tags, last) {
		if containsValue(c.res.Refs, tag.Commit.SHA) {
			log.Printf("skipping tag %s (%s), already in refs", tag.Name, tag.Commit.SHA)
		} else {
			res[tag.Name] = tag.Commit.SHA
		}
	}
	c.res.Releases = res
	return nil
}

func (c *candidates) setCommits(lastDays int) error {
	commits, err := c.ghc.GetCommits(time.Now().AddDate(0, 0, -lastDays))
	if err != nil {
		return errors.Wrap(err, "failed to fetch commits")
	}
	res := make(map[string]string)
	for date, commit := range lastCommitByDay(filterMergeCommits(commits)) {
		if containsValue(c.res.Refs, commit.SHA) {
			log.Printf("skipping commit %s, already in refs", commit.SHA)
		} else if containsValue(c.res.Refs, commit.SHA) {
			log.Printf("skipping commit %s, already in releases", commit.SHA)
		} else {
			res[date] = commit.SHA
		}
	}
	c.res.Commits = res
	return nil
}

func filterMergeCommits(commits []github.Commit) []github.Commit {
	var mergeCommits []github.Commit
	for _, commit := range commits {
		if len(commit.Parents) > 1 {
			mergeCommits = append(mergeCommits, commit)
		}
	}
	return mergeCommits
}

func lastCommitByDay(commits []github.Commit) map[string]github.Commit {
	lastCommits := make(map[string]github.Commit)
	for _, cm := range commits {
		day := cm.Commit.Committer.Date.Format("2006-01-02")
		if existingCommit, exists := lastCommits[day]; !exists || cm.Commit.Committer.Date.After(existingCommit.Commit.Committer.Date) {
			lastCommits[day] = cm
		}
	}
	return lastCommits
}

func filterFeatureReleases(tags []github.Tag, last int) []github.Tag {
	latestReleases := make(map[string]github.Tag)
	zeroReleases := make(map[string]github.Tag)
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
	var res []github.Tag
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
	match := reSemverRelease.FindStringSubmatch(version)
	if len(match) > 1 {
		parts := strings.Split(match[1], ".")
		if len(parts) == 3 {
			return parts[2]
		}
	}
	return ""
}

func containsValue(m map[string]string, value string) bool {
	if m == nil {
		return false
	}
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}
