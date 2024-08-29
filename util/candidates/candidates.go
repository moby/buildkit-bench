package candidates

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

type Candidates struct {
	Refs     map[string]Commit `json:"refs"`
	Releases map[string]Commit `json:"releases"`
	Commits  map[string]Commit `json:"commits"`

	ghc *github.Client
}

type Commit struct {
	SHA  string    `json:"sha"`
	Date time.Time `json:"date"`
}

type Ref struct {
	Name   string
	Commit Commit
}

func Get(ghc *github.Client, refs string, lastDays int, lastReleases int) (*Candidates, error) {
	c := &Candidates{
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
	log.Printf("%d ref(s), %d release(s) and %d commit(s) marked as candidates", len(c.Refs), len(c.Releases), len(c.Commits))
	return c, nil
}

func Load(f string) (*Candidates, error) {
	dt, err := os.ReadFile(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read candidates")
	}
	var c Candidates
	if err := json.Unmarshal(dt, &c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal candidates")
	}
	return &c, nil
}

func (c *Candidates) List() map[string]Commit {
	res := make(map[string]Commit)
	for k, v := range c.Refs {
		res[k] = v
	}
	for k, v := range c.Releases {
		res[k] = v
	}
	for k, v := range c.Commits {
		res[k] = v
	}
	return res
}

func (c *Candidates) Sorted() []Ref {
	var sortedCandidates []Ref
	for ref, cm := range c.List() {
		sortedCandidates = append(sortedCandidates, Ref{
			Name:   ref,
			Commit: cm,
		})
	}
	sort.Slice(sortedCandidates, func(i, j int) bool {
		return sortedCandidates[i].Commit.Date.Before(sortedCandidates[j].Commit.Date)
	})
	return sortedCandidates
}

func (c *Candidates) WriteFile(path string) error {
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

func (c *Candidates) SetGhaOutput(name string) error {
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
	dt, err := json.Marshal(includes)
	if err != nil {
		return errors.Wrap(err, "failed to marshal includes")
	}
	githubactions.SetOutput(name, string(dt))
	return nil
}

func (c *Candidates) setRefs(refs []string) error {
	res := make(map[string]Commit)
	for _, ref := range refs {
		cm, err := c.ghc.GetCommit(ref)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch commit for ref %q", ref)
		}
		res[ref] = Commit{
			SHA:  cm.SHA,
			Date: cm.Commit.Committer.Date,
		}
	}
	c.Refs = res
	return nil
}

func (c *Candidates) setReleases(last int) error {
	tags, err := c.ghc.GetTags()
	if err != nil {
		return errors.Wrap(err, "failed to fetch tags")
	}
	res := make(map[string]Commit)
	for _, tag := range filterFeatureReleases(tags, last) {
		if containsCommitSha(c.Refs, tag.Commit.SHA) {
			log.Printf("skipping tag %s (%s), already in refs", tag.Name, tag.Commit.SHA)
		} else {
			cm, err := c.ghc.GetCommit(tag.Commit.SHA)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch commit for tag commit %q", tag.Commit.SHA)
			}
			res[tag.Name] = Commit{
				SHA:  cm.SHA,
				Date: cm.Commit.Committer.Date,
			}
		}
	}
	c.Releases = res
	return nil
}

func (c *Candidates) setCommits(lastDays int) error {
	commits, err := c.ghc.GetCommits(time.Now().AddDate(0, 0, -lastDays))
	if err != nil {
		return errors.Wrap(err, "failed to fetch commits")
	}
	res := make(map[string]Commit)
	for date, cm := range lastCommitByDay(filterMergeCommits(commits)) {
		if containsCommitSha(c.Refs, cm.SHA) {
			log.Printf("skipping commit %s, already in refs", cm.SHA)
		} else if containsCommitSha(c.Releases, cm.SHA) {
			log.Printf("skipping commit %s, already in releases", cm.SHA)
		} else {
			res[date] = Commit{
				SHA:  cm.SHA,
				Date: cm.Commit.Committer.Date,
			}
		}
	}
	c.Commits = res
	return nil
}

func filterMergeCommits(commits []github.Commit) []github.Commit {
	var mergeCommits []github.Commit
	for _, cm := range commits {
		if len(cm.Parents) > 1 {
			mergeCommits = append(mergeCommits, cm)
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

func containsCommitSha(m map[string]Commit, sha string) bool {
	if m == nil {
		return false
	}
	for _, cm := range m {
		if cm.SHA == sha {
			return true
		}
	}
	return false
}
