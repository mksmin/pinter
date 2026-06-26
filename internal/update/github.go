package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/mksmin/pinter/releases/latest"
	githubAPIVersion       = "2026-03-10"
)

type GitHubClient struct {
	httpClient *http.Client
	endpoint   string
}

type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		endpoint: githubLatestReleaseURL,
	}
}

func (c *GitHubClient) LatestRelease(
	ctx context.Context,
) (
	Release,
	error,
) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.endpoint,
		nil,
	)

	if err != nil {
		return Release{}, fmt.Errorf(
			"create latest release request: %w",
			err,
		)
	}
	request.Header.Set(
		"Accept",
		"application/vnd.github+json",
	)
	request.Header.Set(
		"X-GitHub-Api-Version",
		githubAPIVersion,
	)
	request.Header.Set(
		"User-Agent",
		"pinter",
	)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Release{}, fmt.Errorf(
			"request latest release: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf(
			"request latest release: unexpected status code: %s",
			response.Status,
		)
	}

	var payload githubReleaseResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return Release{}, fmt.Errorf(
			"decode latest release: %w",
			err,
		)
	}

	if payload.TagName == "" || payload.HTMLURL == "" {
		return Release{}, fmt.Errorf(
			"decode latest release: missing tag_name or html_url",
		)
	}

	return Release{
		Version: payload.TagName,
		URL:     payload.HTMLURL,
	}, nil
}
