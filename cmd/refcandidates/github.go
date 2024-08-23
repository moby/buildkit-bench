package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const githubAPIURL = "https://api.github.com"

type GitHubClient struct {
	apiURL string
	repo   string
	token  string
	client *http.Client
}

type GitHubCommit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date string `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
	} `json:"commit"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

type GitHubTag struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
	Name string `json:"name"`
}

func newGitHubClient(repo, token string) (*GitHubClient, error) {
	if token == "" {
		return nil, fmt.Errorf("missing GitHub token")
	}
	if repo == "" {
		return nil, fmt.Errorf("missing GitHub repository")
	}
	return &GitHubClient{
		apiURL: githubAPIURL,
		repo:   repo,
		token:  token,
		client: &http.Client{},
	}, nil
}

func (c *GitHubClient) GetCommit(ref string) (*GitHubCommit, error) {
	url := fmt.Sprintf("%s/repos/%s/commits/%s", c.apiURL, c.repo, ref)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res *GitHubCommit
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal response: %s", string(body))
	}
	return res, nil
}

func (c *GitHubClient) GetCommits(since time.Time) ([]GitHubCommit, error) {
	var res []GitHubCommit
	url := fmt.Sprintf("%s/repos/%s/commits?since=%s", c.apiURL, c.repo, since.Format(time.RFC3339))

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var commits []GitHubCommit
		if err := json.Unmarshal(body, &commits); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal response: %s", string(body))
		}
		res = append(res, commits...)

		linkHeader := resp.Header.Get("Link")
		if linkHeader == "" {
			break
		}

		nextURL := getNextPageURL(linkHeader)
		if nextURL == "" {
			break
		}
		url = nextURL
	}

	return res, nil
}

func (c *GitHubClient) GetTags() ([]GitHubTag, error) {
	var res []GitHubTag
	url := fmt.Sprintf("%s/repos/%s/tags", c.apiURL, c.repo)

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var tags []GitHubTag
		if err := json.Unmarshal(body, &tags); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal response: %s", string(body))
		}
		res = append(res, tags...)

		linkHeader := resp.Header.Get("Link")
		if linkHeader == "" {
			break
		}

		nextURL := getNextPageURL(linkHeader)
		if nextURL == "" {
			break
		}
		url = nextURL
	}

	return res, nil
}

func getNextPageURL(linkHeader string) string {
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) < 2 {
			continue
		}
		urlPart := strings.Trim(parts[0], "<>")
		relPart := strings.TrimSpace(parts[1])
		if relPart == `rel="next"` {
			return urlPart
		}
	}
	return ""
}
