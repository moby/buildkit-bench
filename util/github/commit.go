package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
	} `json:"commit"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents"`
}

func (c *Client) GetCommit(ref string) (*Commit, error) {
	url := fmt.Sprintf("%s/repos/%s/commits/%s", githubAPIURL, c.Repo, ref)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
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

	var res *Commit
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal response: %s", string(body))
	}
	return res, nil
}

func (c *Client) GetCommits(since time.Time) ([]Commit, error) {
	var res []Commit
	url := fmt.Sprintf("%s/repos/%s/commits?since=%s", githubAPIURL, c.Repo, since.Format(time.RFC3339))

	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
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

		var commits []Commit
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
