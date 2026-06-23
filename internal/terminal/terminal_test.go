package terminal

import (
	"testing"

	"pinter/internal/model"
)

func TestBuildSSHArgsWindowsPath(
	t *testing.T,
) {
	host := model.Host{
		Hostname:     "example.com",
		Username:     "deploy",
		Port:         22,
		IdentityFile: `C:\Users\username\Рабочий стол\.ssh\ключ`,
	}

	args := buildSSHArgs(host)

	want := []string{
		"ssh",
		"-i",
		`C:\Users\username\Рабочий стол\.ssh\ключ`,
		"deploy@example.com",
	}

	if len(args) != len(want) {
		t.Fatalf(
			"len mismatch: got %#v, want %#v",
			args,
			want,
		)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf(
				"arg %d: got %q, want %q",
				i,
				args[i],
				want[i],
			)
		}
	}

}

func TestCmdCommandQuotesRussianDesktopPath(
	t *testing.T,
) {
	args := []string{
		"ssh",
		"-i",
		`C:\Users\moype\Рабочий стол\.ssh\ключ`,
		"deploy@example.com",
	}
	got := cmdCommand(args)
	want := `ssh -i "C:\Users\moype\Рабочий стол\.ssh\ключ" deploy@example.com`

	if got != want {
		t.Fatalf(
			"got %q, want %q",
			got,
			want,
		)
	}
}

func TestResolveTerminalDefaultsToAuto(
	t *testing.T,
) {
	t.Setenv(
		"PINTER_TERMINAL",
		"",
	)

	got := resolveTerminal()
	if got != TerminalAuto {
		t.Fatalf(
			"got %q, want %q",
			got,
			TerminalAuto,
		)
	}
}

func TestResolveTerminalSupportsPwsh(
	t *testing.T,
) {
	t.Setenv(
		"PINTER_TERMINAL",
		"pwsh",
	)

	got := resolveTerminal()
	if got != TerminalPwsh {
		t.Fatalf(
			"got %q, want %q",
			got,
			TerminalPwsh,
		)
	}
}

func TestResolveTerminalSupportsWindowsTerminal(
	t *testing.T,
) {
	t.Setenv(
		"PINTER_TERMINAL",
		"wt",
	)

	got := resolveTerminal()
	if got != TerminalWindowsTerminal {
		t.Fatalf(
			"got %q, want %q",
			got,
			TerminalWindowsTerminal,
		)
	}
}

func TestPowerShellCommandQuotesRussianDesktopPath(
	t *testing.T,
) {
	args := []string{
		"ssh",
		"-i",
		`C:\Users\moype\Рабочий стол\.ssh\ключ`,
		"deploy@example.com",
	}

	got := powershellCommand(args)
	want := `ssh -i 'C:\Users\moype\Рабочий стол\.ssh\ключ' deploy@example.com`

	if got != want {
		t.Fatalf(
			"got %q, want %q",
			got,
			want,
		)
	}
}

func TestPowerShellQuoteEscapesSingleQuote(
	t *testing.T,
) {
	got := powershellQuote(
		`C:\Users\moype\Рабочий стол\key's\.ssh\ключ`,
	)
	want := `'C:\Users\moype\Рабочий стол\key''s\.ssh\ключ'`

	if got != want {
		t.Fatalf(
			"got %q, want %q",
			got,
			want,
		)
	}
}

func TestResolveTerminalInvalidDefaultsToAuto(
	t *testing.T,
) {
	t.Setenv(
		"PINTER_TERMINAL",
		"bad",
	)

	got := resolveTerminal()
	if got != TerminalAuto {
		t.Fatalf(
			"got %q, want %q",
			got,
			TerminalAuto,
		)
	}
}
