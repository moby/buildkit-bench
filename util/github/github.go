package github

import (
	"net/http"

	"github.com/pkg/errors"
)

const githubAPIURL = "https://api.github.com"

type Client struct {
	Repo  string
	Token string

	client *http.Client
}

func NewClient(repo, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("missing GitHub token")
	}
	if repo == "" {
		return nil, errors.New("missing GitHub repository")
	}
	return &Client{
		Repo:   repo,
		Token:  token,
		client: &http.Client{},
	}, nil
}
