package terminal

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"pinter/internal/model"
)

type Launcher struct{}

func NewLauncher() *Launcher {
	return &Launcher{}
}

func (l *Launcher) BuildSSHCommand(
	host model.Host,
) string {
	parts := []string{
		"ssh",
	}
	if host.Port != 0 && host.Port != 22 {
		parts = append(
			parts,
			"-p",
			fmt.Sprint(host.Port),
		)
	}
	if host.IdentityFile != "" {
		parts = append(
			parts,
			"-i",
			host.IdentityFile,
		)
	}
	target := host.Hostname
	if host.Username != "" {
		target = host.Username + "@" + host.Hostname
	}
	parts = append(
		parts,
		target,
	)
	return shellQuote(parts)
}

func (l *Launcher) Open(
	host model.Host,
) (
	string,
	string,
	error,
) {
	command := l.BuildSSHCommand(host)
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(
			`tell application "Terminal"
	activate
	do script %q
end tell`,
			command,
		)
		if err := exec.Command(
			"osascript",
			"-e",
			script,
		).Run(); err != nil {
			return command, "Terminal.app", fmt.Errorf(
				"open Terminal.app: %w",
				err,
			)
		}
		return command, "Terminal.app", nil
	case "windows":
		if err := exec.Command(
			"cmd",
			"/C",
			"start",
			"cmd",
			"/K",
			command,
		).Run(); err != nil {
			return command, "cmd", fmt.Errorf(
				"open Windows terminal: %w",
				err,
			)
		}
		return command, "cmd", nil
	default:
		if err := exec.Command(
			"sh",
			"-lc",
			command,
		).Start(); err != nil {
			return command, "sh", fmt.Errorf(
				"start ssh: %w",
				err,
			)
		}
		return command, "sh", nil
	}
}

func shellQuote(
	parts []string,
) string {
	quoted := make(
		[]string,
		0,
		len(parts),
	)
	for _, part := range parts {
		if part == "" {
			quoted = append(
				quoted,
				"''",
			)
			continue
		}
		if strings.IndexFunc(
			part,
			func(
				r rune,
			) bool {
				return !(r >= 'A' && r <= 'Z') &&
					!(r >= 'a' && r <= 'z') &&
					!(r >= '0' && r <= '9') &&
					!strings.ContainsRune(
						"@%_+=:,./-",
						r,
					)
			},
		) == -1 {
			quoted = append(
				quoted,
				part,
			)
			continue
		}
		quoted = append(
			quoted,
			"'"+strings.ReplaceAll(
				part,
				"'",
				`'\''`,
			)+"'",
		)
	}
	return strings.Join(
		quoted,
		" ",
	)
}
