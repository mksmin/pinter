package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func DataDir() (
	string,
	error,
) {
	if override := os.Getenv("PINTER_DATA_DIR"); override != "" {
		return override, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf(
			"resolve user config dir: %w",
			err,
		)
	}
	return filepath.Join(
		base,
		"pinter",
	), nil
}

func DBPath() (
	string,
	error,
) {
	if override := os.Getenv("PINTER_DB_PATH"); override != "" {
		return override, nil
	}

	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(
		dir,
		"pinter.sqlite",
	), nil
}
