package candidates

import (
	"fmt"
	"testing"
	"time"

	ggithub "github.com/google/go-github/v90/github"
	"github.com/stretchr/testify/require"
)

func TestSetRefsUsesPullRequestMergeCommitSHA(t *testing.T) {
	updatedAt := time.Date(2026, 8, 12, 10, 11, 12, 0, time.UTC)
	client := &fakeGithubClient{
		pulls: map[int]*ggithub.PullRequest{
			42: {
				MergeCommitSHA: new("merge-sha"),
				UpdatedAt:      &ggithub.Timestamp{Time: updatedAt},
				Merged:         new(true),
			},
		},
	}
	c := &Candidates{ghc: client}

	require.NoError(t, c.setRefs([]string{"pr-42"}))
	require.Equal(t, map[string]Commit{
		"pr-42": {
			SHA:    "merge-sha",
			Date:   updatedAt,
			Merged: true,
		},
	}, c.Refs)
	require.Empty(t, client.commitRefs)
}

func TestSetRefsFallsBackToPullRequestMergeRef(t *testing.T) {
	commitDate := time.Date(2026, 8, 12, 10, 11, 12, 0, time.UTC)
	client := &fakeGithubClient{
		commits: map[string]*ggithub.RepositoryCommit{
			"refs/pull/42/merge": repositoryCommitFixture("merge-ref-sha", commitDate),
		},
		pulls: map[int]*ggithub.PullRequest{
			42: {
				Merged: new(false),
			},
		},
	}
	c := &Candidates{ghc: client}

	require.NoError(t, c.setRefs([]string{"pr-42"}))
	require.Equal(t, map[string]Commit{
		"pr-42": {
			SHA:  "merge-ref-sha",
			Date: commitDate,
		},
	}, c.Refs)
	require.Equal(t, []string{"refs/pull/42/merge"}, client.commitRefs)
}

func TestSetRefsReturnsErrorWhenPullRequestMergeCommitIsUnavailable(t *testing.T) {
	client := &fakeGithubClient{
		pulls: map[int]*ggithub.PullRequest{
			42: {},
		},
	}
	c := &Candidates{ghc: client}

	err := c.setRefs([]string{"pr-42"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to resolve commit for pull request 42")
}

func TestSetRefsResolvesBranchCommit(t *testing.T) {
	commitDate := time.Date(2026, 8, 12, 10, 11, 12, 0, time.UTC)
	client := &fakeGithubClient{
		commits: map[string]*ggithub.RepositoryCommit{
			"master": repositoryCommitFixture("master-sha", commitDate),
		},
	}
	c := &Candidates{ghc: client}

	require.NoError(t, c.setRefs([]string{"master"}))
	require.Equal(t, map[string]Commit{
		"master": {
			SHA:  "master-sha",
			Date: commitDate,
		},
	}, c.Refs)
	require.Equal(t, []string{"master"}, client.commitRefs)
}

type fakeGithubClient struct {
	commits    map[string]*ggithub.RepositoryCommit
	pulls      map[int]*ggithub.PullRequest
	commitRefs []string
}

func (c *fakeGithubClient) GetCommit(ref string) (*ggithub.RepositoryCommit, error) {
	c.commitRefs = append(c.commitRefs, ref)
	cm := c.commits[ref]
	if cm == nil {
		return nil, fmt.Errorf("missing commit for ref %q", ref)
	}
	return cm, nil
}

func (c *fakeGithubClient) GetCommits(time.Time) ([]*ggithub.RepositoryCommit, error) {
	return nil, nil
}

func (c *fakeGithubClient) GetPullRequest(number int) (*ggithub.PullRequest, error) {
	pr := c.pulls[number]
	if pr == nil {
		return nil, fmt.Errorf("missing pull request %d", number)
	}
	return pr, nil
}

func (c *fakeGithubClient) GetTags() ([]*ggithub.RepositoryTag, error) {
	return nil, nil
}

func repositoryCommitFixture(sha string, date time.Time) *ggithub.RepositoryCommit {
	return &ggithub.RepositoryCommit{
		SHA: new(sha),
		Commit: &ggithub.Commit{
			Committer: &ggithub.CommitAuthor{
				Date: &ggithub.Timestamp{Time: date},
			},
		},
	}
}
