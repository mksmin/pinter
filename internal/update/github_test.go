package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubClientLatestRelease(
	t *testing.T,
) {
	handler := http.HandlerFunc(
		func(
			response http.ResponseWriter,
			request *http.Request,
		) {
			if request.Method != http.MethodGet {
				t.Errorf(
					"method = %q, want %q",
					request.Method,
					http.MethodGet,
				)
			}

			if request.URL.Path != "/repos/mksmin/pinter/releases/latest" {
				t.Errorf(
					"path = %q",
					request.URL.Path,
				)
			}

			if got := request.Header.Get("Accept"); got != "application/vnd.github+json" {
				t.Errorf(
					"Accept = %q",
					got,
				)
			}

			if got := request.Header.Get("X-GitHub-Api-Version"); got != githubAPIVersion {
				t.Errorf(
					"X-GitHub-Api-Version = %q",
					got,
				)
			}

			if got := request.Header.Get("User-Agent"); got != "pinter" {
				t.Errorf(
					"User-Agent = %q",
					got,
				)
			}

			response.Header().Set(
				"Content-Type",
				"application/json",
			)

			fmt.Fprint(
				response,
				`{
	"tag_name": "v0.2.0",
	"html_url": "https://github.com/mksmin/pinter/releases/tag/v0.2.0"
}`,
			)
		},
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	client := NewGitHubClient()
	client.httpClient = server.Client()
	client.endpoint = server.URL + "/repos/mksmin/pinter/releases/latest"

	got, err := client.LatestRelease(
		context.Background(),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := Release{
		Version: "v0.2.0",
		URL:     "https://github.com/mksmin/pinter/releases/tag/v0.2.0",
	}

	if got != want {
		t.Fatalf(
			"LatestRelease() = %#v, want %#v",
			got,
			want,
		)
	}
}
