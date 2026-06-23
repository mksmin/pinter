package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"pinter/internal/model"
)

const (
	TerminalAuto            = "auto"
	TerminalCMD             = "cmd"
	TerminalPowerShell      = "powershell"
	TerminalPwsh            = "pwsh"
	TerminalWindowsTerminal = "wt"
)

type Launcher struct {
	terminal string
}

func NewLauncher() *Launcher {
	return &Launcher{
		terminal: resolveTerminal(),
	}
}

func (l *Launcher) BuildSSHCommand(
	host model.Host,
) string {
	return shellQuote(
		buildSSHArgs(host),
	)
}

func buildSSHArgs(
	host model.Host,
) []string {
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
	return parts
}

func (l *Launcher) Open(
	host model.Host,
) (
	string,
	string,
	error,
) {
	args := buildSSHArgs(host)
	command := shellQuote(args)

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
		terminal := l.terminal
		if terminal == TerminalAuto {
			terminal = resolveWindowsTerminal()
		}
		return openWindowsTerminal(
			terminal,
			args,
		)
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

func cmdQuote(
	part string,
) string {
	if part == "" {
		return `""`
	}

	needsQuote := strings.IndexFunc(
		part,
		func(
			r rune,
		) bool {
			return r == ' ' ||
				r == '\t' ||
				r == '&' ||
				r == '|' ||
				r == '<' ||
				r == '>' ||
				r == '^' ||
				r == '"'
		},
	) != -1

	escaped := strings.ReplaceAll(
		part,
		`"`,
		`\"`,
	)

	if !needsQuote {
		return escaped
	}
	return `"` + escaped + `"`
}

func cmdCommand(
	parts []string,
) string {
	quoted := make(
		[]string,
		0,
		len(parts),
	)
	for _, part := range parts {
		quoted = append(
			quoted,
			cmdQuote(part),
		)
	}
	return strings.Join(
		quoted,
		" ",
	)
}

func resolveTerminal() string {
	value := strings.ToLower(
		strings.TrimSpace(
			os.Getenv("PINTER_TERMINAL"),
		),
	)

	switch value {
	case "", TerminalAuto:
		return TerminalAuto
	case TerminalCMD, TerminalPowerShell, TerminalPwsh, TerminalWindowsTerminal:
		return value
	default:
		return TerminalAuto
	}
}
func resolveWindowsTerminal() string {
	if _, err := exec.LookPath(
		"wt",
	); err == nil {
		return TerminalWindowsTerminal
	}
	if _, err := exec.LookPath(
		"pwsh",
	); err == nil {
		return TerminalPwsh
	}
	if _, err := exec.LookPath("powershell"); err == nil {
		return TerminalPowerShell
	}
	return TerminalCMD
}

func resolveWindowsShell() string {
	if _, err := exec.LookPath("pwsh"); err == nil {
		return TerminalPwsh
	}
	if _, err := exec.LookPath("powershell"); err == nil {
		return TerminalPowerShell
	}
	return TerminalCMD
}

func powershellQuote(
	part string,
) string {
	if part == "" {
		return "''"
	}

	needsQuote := strings.IndexFunc(
		part,
		func(
			r rune,
		) bool {
			return r == ' ' ||
				r == '\t' ||
				r == '\'' ||
				r == '&' ||
				r == '|' ||
				r == '<' ||
				r == '>' ||
				r == '(' ||
				r == ')'
		},
	) != -1

	if !needsQuote {
		return part
	}
	return "'" + strings.ReplaceAll(
		part,
		"'",
		"''",
	) + "'"
}

func powershellCommand(
	parts []string,
) string {
	quoted := make(
		[]string,
		0,
		len(parts),
	)

	for _, part := range parts {
		quoted = append(
			quoted,
			powershellQuote(part),
		)
	}
	return strings.Join(
		quoted,
		" ",
	)

}

func windowsTerminalCommand(
	shell string,
	command string,
) *exec.Cmd {
	switch shell {
	case TerminalPwsh:
		return exec.Command(
			"wt",
			"new-tab",
			"--title",
			"pinter",
			"pwsh",
			"-NoExit",
			"-Command",
			command,
		)
	case TerminalPowerShell:
		return exec.Command(
			"wt",
			"new-tab",
			"--title",
			"pinter",
			"powershell",
			"-NoExit",
			"-Command",
			command,
		)
	default:
		return exec.Command(
			"wt",
			"new-tab",
			"--title",
			"pinter",
			"cmd",
			"/K",
			command,
		)
	}
}

func openWindowsTerminal(
	terminal string,
	args []string,
) (
	string,
	string,
	error,
) {
	switch terminal {
	case TerminalWindowsTerminal:
		if _, err := exec.LookPath("wt"); err != nil {
			return openWindowsTerminal(
				resolveWindowsShell(),
				args,
			)
		}

		shell := resolveWindowsShell()
		command := renderWindowsCommand(
			shell,
			args,
		)
		if err := windowsTerminalCommand(
			shell,
			command,
		).Run(); err != nil {
			return openWindowsTerminal(
				shell,
				args,
			)
		}
		return command, "wt/" + shell, nil
	case TerminalPwsh:
		command := powershellCommand(args)
		if err := exec.Command(
			"cmd",
			"/C",
			"start",
			"pwsh",
			"-NoExit",
			"-Command",
			command,
		).Run(); err != nil {
			return command, "pwsh", fmt.Errorf(
				"open PowerShell: %w",
				err,
			)
		}
		return command, "pwsh", nil
	case TerminalPowerShell:
		command := powershellCommand(args)
		if err := exec.Command(
			"cmd",
			"/C",
			"start",
			"powershell",
			"-NoExit",
			"-Command",
			command,
		).Run(); err != nil {
			return command, "powershell", fmt.Errorf(
				"open Windows PowerShell: %w",
				err,
			)
		}
		return command, "powershell", nil
	default:
		command := cmdCommand(args)
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
	}
}

func renderWindowsCommand(
	terminal string,
	args []string,
) string {
	if terminal == TerminalCMD {
		return cmdCommand(args)
	}
	return powershellCommand(args)
}
