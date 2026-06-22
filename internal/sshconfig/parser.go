package sshconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pinter/internal/model"
)

type ParsedHost struct {
	Alias        string
	Hostname     string
	Port         int
	Username     string
	IdentityFile string
}

func Parse(
	r io.Reader,
) ([]ParsedHost, error) {
	scanner := bufio.NewScanner(r)
	var hosts []ParsedHost
	var current *ParsedHost

	commit := func() {
		if current == nil {
			return
		}
		if current.Hostname == "" {
			current.Hostname = current.Alias
		}
		if current.Port == 0 {
			current.Port = 22
		}
		if current.Username == "" {
			current.Username = "root"
		}
		hosts = append(hosts, *current)
		current = nil
	}

	for scanner.Scan() {
		line := cleanLine(scanner.Text())
		if line == "" {
			continue
		}

		key, value, ok := splitDirective(line)
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "host":
			commit()
			current = nil
			fields := strings.Fields(value)
			if len(fields) != 1 || isWildcardHost(fields[0]) {
				continue
			}
			current = &ParsedHost{Alias: fields[0]}
		case "hostname":
			if current != nil {
				current.Hostname = value
			}
		case "port":
			if current != nil {
				port, err := strconv.Atoi(value)
				if err == nil {
					current.Port = port
				}
			}
		case "user":
			if current != nil {
				current.Username = value
			}
		case "identityfile":
			if current != nil {
				current.IdentityFile = expandIdentityFile(value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse ssh config: %w", err)
	}
	commit()
	return hosts, nil
}

func ParseFile(
	path string,
) ([]ParsedHost, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open ssh config: %w", err)
	}
	defer file.Close()
	return Parse(file)
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}

func (h ParsedHost) HostInput() model.HostInput {
	return model.HostInput{
		Alias:        h.Alias,
		Hostname:     h.Hostname,
		Port:         h.Port,
		Username:     h.Username,
		IdentityFile: h.IdentityFile,
	}
}

func cleanLine(
	line string,
) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}

	var out strings.Builder
	inSingle := false
	inDouble := false
	for _, r := range line {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimSpace(out.String())
			}
		}
		out.WriteRune(r)
	}
	return strings.TrimSpace(out.String())
}

func splitDirective(
	line string,
) (string, string, bool) {
	if idx := strings.Index(line, "="); idx > 0 {
		key := strings.TrimSpace(line[:idx])
		value := strings.Trim(strings.TrimSpace(line[idx+1:]), `"'`)
		return key, value, key != "" && value != ""
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", "", false
	}
	return fields[0], strings.Trim(strings.Join(fields[1:], " "), `"'`), true
}

func isWildcardHost(
	value string,
) bool {
	return strings.ContainsAny(value, "*?!")
}

func expandIdentityFile(
	value string,
) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, value[2:])
		}
	}
	return value
}
