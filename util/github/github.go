package github

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/pkg/errors"
)

type Client struct {
	ctx    context.Context
	client *github.Client
	owner  string
	repo   string
}

func NewClient(repo, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("missing GitHub token")
	}
	if repo == "" {
		return nil, errors.New("missing GitHub repository")
	}
	repoParts := strings.SplitN(repo, "/", 2)
	if len(repoParts) != 2 {
		return nil, errors.New("invalid GitHub repository format")
	}
	return &Client{
		ctx:    context.Background(),
		client: github.NewClient(nil).WithAuthToken(token),
		owner:  repoParts[0],
		repo:   repoParts[1],
	}, nil
}

func (c *Client) GetCommit(ref string) (*github.RepositoryCommit, error) {
	commit, _, err := c.client.Repositories.GetCommit(c.ctx, c.owner, c.repo, ref, nil)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (c *Client) GetCommits(since time.Time) ([]*github.RepositoryCommit, error) {
	var res []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		Since: since,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	for {
		commits, resp, err := c.client.Repositories.ListCommits(c.ctx, c.owner, c.repo, opt)
		if err != nil {
			return nil, err
		}
		res = append(res, commits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return res, nil
}

func (c *Client) GetTags() ([]*github.RepositoryTag, error) {
	var res []*github.RepositoryTag
	opt := &github.ListOptions{
		PerPage: 100,
	}
	for {
		tags, resp, err := c.client.Repositories.ListTags(c.ctx, c.owner, c.repo, opt)
		if err != nil {
			return nil, err
		}
		res = append(res, tags...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return res, nil
}
