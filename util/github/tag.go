package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Tag struct {
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
	Name string `json:"name"`
}

func (c *Client) GetTags() ([]Tag, error) {
	var res []Tag
	url := fmt.Sprintf("%s/repos/%s/tags", githubAPIURL, c.Repo)

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

		var tags []Tag
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
