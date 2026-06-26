package update

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

type Release struct {
	Version string
	URL     string
}

type CheckResult struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseURL      string
}

type ReleaseClient interface {
	LatestRelease(
		ctx context.Context,
	) (
		Release,
		error,
	)
}

type Checker struct {
	releases ReleaseClient
}

func NewChecker(
	releases ReleaseClient,
) *Checker {
	return &Checker{
		releases: releases,
	}
}

func (c *Checker) Check(
	ctx context.Context,
	currentVersion string,
) (
	CheckResult,
	error,
) {
	release, err := c.releases.LatestRelease(ctx)
	if err != nil {
		return CheckResult{}, fmt.Errorf(
			"get latest release: %w",
			err,
		)
	}
	return CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.Version,
		UpdateAvailable: IsNewerVersion(
			currentVersion,
			release.Version,
		),
		ReleaseURL: release.URL,
	}, nil
}

func IsNewerVersion(
	current string,
	latest string,
) bool {
	if current == "dev" {
		return false
	}

	current = normalizeVersion(current)
	latest = normalizeVersion(latest)

	if !semver.IsValid(current) || !semver.IsValid(latest) {
		return false
	}

	return semver.Compare(
		latest,
		current,
	) > 0
}

func normalizeVersion(
	value string,
) string {
	value = strings.TrimSpace(value)

	if value == "" || strings.HasPrefix(
		value,
		"v",
	) {
		return value
	}
	return "v" + value
}
