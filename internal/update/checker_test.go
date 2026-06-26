package update

import (
	"context"
	"errors"
	"testing"
)

type fakeReleaseClient struct {
	release Release
	err     error
}

func (f fakeReleaseClient) LatestRelease(
	_ context.Context,
) (
	Release,
	error,
) {
	return f.release, f.err
}

func TestIsNewerVersion(
	t *testing.T,
) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"new version", "v0.1.0", "v0.2.0", true},
		{"same version", "v0.2.0", "v0.2.0", false},
		{"older release", "v0.3.0", "v0.2.0", false},
		{"without prefix", "0.1.0", "0.2.0", true},
		{"development build", "dev", "v0.2.0", false},
		{"invalid current", "unknown", "v0.2.0", false},
		{"invalid latest", "v0.1.0", "latest", false},
		{"stable after prerelease", "v0.2.0-rc.1", "v0.2.0", true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got := IsNewerVersion(
					tt.current,
					tt.latest,
				)

				if got != tt.want {
					t.Fatalf(
						"IsNewerVersion(%q, %q) = %v, want %v",
						tt.current,
						tt.latest,
						got,
						tt.want,
					)
				}
			},
		)
	}
}

func TestCheckerCheck(
	t *testing.T,
) {
	checker := NewChecker(
		fakeReleaseClient{
			release: Release{
				Version: "v0.2.0",
				URL:     "https://example.com/v0.2.0",
			},
		},
	)

	got, err := checker.Check(
		context.Background(),
		"v0.1.0",
	)
	if err != nil {
		t.Fatal(err)
	}

	want := CheckResult{
		CurrentVersion:  "v0.1.0",
		LatestVersion:   "v0.2.0",
		UpdateAvailable: true,
		ReleaseURL:      "https://example.com/v0.2.0",
	}

	if got != want {
		t.Fatalf(
			"Check() = %#v, want %#v",
			got,
			want,
		)
	}
}

func TestCheckerCheckReturnsClientError(
	t *testing.T,
) {
	clientErr := errors.New("network unavailable")

	checker := NewChecker(
		fakeReleaseClient{
			err: clientErr,
		},
	)

	_, err := checker.Check(
		context.Background(),
		"v0.1.0",
	)
	if !errors.Is(err, clientErr) {
		t.Fatalf(
			"Check() error = %v, want %v",
			err,
			clientErr,
		)
	}
}
