package tui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"pinter/internal/app"
	"pinter/internal/model"
)

type action int

const (
	actionNone action = iota
	actionQuit
)

type menuItem struct {
	title       string
	description string
	run         func() (
		action,
		string,
	)
}

type menu struct {
	title string
	items []menuItem
}

type key int

const (
	keyUnknown key = iota
	keyUp
	keyDown
	keyEnter
	keyQuit
	keyBack
	keyConnect
)

const logo = ` ____  _       _            
|  _ \(_)_ __ | |_ ___ _ __ 
| |_) | | '_ \| __/ _ \ '__|
|  __/| | | | | ||  __/ |   
|_|   |_|_| |_|\__\___|_|   
Local SSH keeper
Made by mksmin: https://github.com/mksmin/pinter
`

const (
	ansiReset          = "\033[0m"
	ansiCursorBlinkOff = "\033[?12l"
	ansiCursorBlinkOn  = "\033[?12h"
	ansiCursorHide     = "\033[?25l"
	ansiCursorShow     = "\033[?25h"

	colorBlue    = "\033[38;5;75m"
	colorCyan    = "\033[38;5;51m"
	colorGreen   = "\033[38;5;83m"
	colorMuted   = "\033[38;5;245m"
	colorRed     = "\033[38;5;203m"
	colorYellow  = "\033[38;5;221m"
	colorMagenta = "\033[38;5;177m"
)

type runner struct {
	ctx      context.Context
	svc      *app.Service
	dbPath   string
	oldState *term.State
	reader   *bufio.Reader
}

func Run(
	ctx context.Context,
	svc *app.Service,
	dbPath string,
) error {
	oldState, err := term.MakeRaw(
		int(
			os.Stdin.Fd(),
		),
	)
	if err != nil {
		return fmt.Errorf(
			"enable raw terminal: %w",
			err,
		)
	}
	defer term.Restore(
		int(
			os.Stdin.Fd(),
		),
		oldState,
	)
	write(ansiCursorBlinkOff + ansiCursorHide)
	defer write(ansiCursorBlinkOn + ansiCursorShow + ansiReset)

	r := &runner{
		ctx:      ctx,
		svc:      svc,
		dbPath:   dbPath,
		oldState: oldState,
		reader:   bufio.NewReader(os.Stdin),
	}
	return r.main()
}

func (r *runner) main() error {
	status := ""
	for {
		result, nextStatus, err := r.show(
			menu{
				title: "Main",
				items: []menuItem{
					{
						title:       "Hosts",
						description: "Browse and connect",
						run:         r.hosts,
					},
					{
						title:       "Add",
						description: "Create new SSH host",
						run:         r.addHost,
					},
					{
						title:       "Import",
						description: "Import ~/.ssh/config",
						run:         r.importSSHConfig,
					},
					{
						title:       "History",
						description: "Show connection launches",
						run:         r.history,
					},
					{
						title:       "Status",
						description: "Show local DB status",
						run:         r.status,
					},
					{
						title:       "Quit",
						description: "Exit Pinter",
						run: func() (
							action,
							string,
						) {
							return actionQuit, ""
						},
					},
				},
			},
			status,
		)
		if err != nil {
			return err
		}
		if result == actionQuit {
			clear()
			return nil
		}
		status = nextStatus
	}
}

func (r *runner) show(
	m menu,
	status string,
) (
	action,
	string,
	error,
) {
	index := 0
	for {
		r.render(
			m,
			index,
			status,
		)

		k, err := r.readKey()
		if err != nil {
			return actionNone, status, err
		}
		switch k {
		case keyUp:
			if index > 0 {
				index--
			}
		case keyDown:
			if index < len(m.items)-1 {
				index++
			}
		case keyEnter:
			result, nextStatus := m.items[index].run()
			return result, nextStatus, nil
		case keyQuit:
			return actionQuit, "", nil
		case keyBack:
			return actionNone, "", nil
		}
	}
}

func (r *runner) render(
	m menu,
	selected int,
	status string,
) {
	clear()
	renderLogo()
	line()
	line(colorMuted + "Update check disabled. Local-only MVP." + ansiReset)
	line()

	for i, item := range m.items {
		cursor := "  "
		titleColor := colorMuted
		descColor := colorMuted
		if i == selected {
			cursor = colorCyan + "> " + ansiReset
			titleColor = colorCyan
			descColor = colorCyan
		}
		format(
			"%s%s%d. %-10s%s %s%s%s\r\n",
			cursor,
			titleColor,
			i+1,
			item.title,
			ansiReset,
			descColor,
			item.description,
			ansiReset,
		)
	}

	line()
	line(helpText("Up/Down | Enter | B Back | Q Quit"))
	if status != "" {
		line()
		line(statusText(status))
	}
}

func (r *runner) hosts() (
	action,
	string,
) {
	items, err := r.svc.ListHosts(
		r.ctx,
		"",
	)
	if err != nil {
		return actionNone, err.Error()
	}
	if len(items) == 0 {
		return actionNone, "No hosts yet. Use Add or Import."
	}

	index := 0
	status := "Enter shows details. C connects selected host. B returns."
	for {
		clear()
		renderLogo()
		line()
		line(colorYellow + "Hosts" + ansiReset)
		line()

		for i, host := range items {
			cursor := "  "
			aliasColor := colorMuted
			targetColor := colorMuted
			if i == index {
				cursor = colorCyan + "> " + ansiReset
				aliasColor = colorCyan
				targetColor = colorCyan
			}
			last := "-"
			if host.LastConnectedAt != nil {
				last = host.LastConnectedAt.Local().Format(
					"2006-01-02 15:04",
				)
			}
			format(
				"%s%s%-18s%s %s%s@%s:%d%s  %slast=%s%s\r\n",
				cursor,
				aliasColor,
				host.Alias,
				ansiReset,
				targetColor,
				host.Username,
				host.Hostname,
				host.Port,
				ansiReset,
				colorMuted,
				last,
				ansiReset,
			)
		}

		line()
		line(helpText("Up/Down | Enter Details | C Connect | B Back | Q Quit"))
		line()
		line(statusText(status))

		k, err := r.readKey()
		if err != nil {
			return actionNone, err.Error()
		}
		switch k {
		case keyUp:
			if index > 0 {
				index--
			}
		case keyDown:
			if index < len(items)-1 {
				index++
			}
		case keyEnter:
			result, nextStatus := r.hostDetails(items[index])
			if result == actionQuit {
				return actionQuit, ""
			}
			status = nextStatus
		case keyConnect:
			entry, err := r.svc.Connect(
				r.ctx,
				items[index].Alias,
			)
			if err != nil {
				status = err.Error()
			} else {
				status = "Opened " + entry.AliasSnapshot + " in " + entry.TerminalApp
			}
		case keyBack:
			return actionNone, ""
		case keyQuit:
			return actionQuit, ""
		}
	}
}

func (
	r *runner,
) hostDetails(
	host model.Host,
) (
	action,
	string,
) {
	status := ""
	for {
		clear()
		renderLogo()
		line()
		line(colorYellow + "Host details" + ansiReset)
		line()

		line(label("Alias:          ") + colorCyan + host.Alias + ansiReset)
		line(label("Target:         ") + colorGreen + host.Username + "@" + host.Hostname + ":" + strconv.Itoa(host.Port) + ansiReset)

		keyPath := "–"
		if host.IdentityFile != "" {
			keyPath = host.IdentityFile
		}
		line(label("IdentityFile:   ") + colorMuted + keyPath + ansiReset)

		favorite := "no"
		if host.Favorite {
			favorite = "yes"
		}
		line(label("Favorite:       ") + colorMuted + favorite + ansiReset)

		last := "-"
		if host.LastConnectedAt != nil {
			last = host.LastConnectedAt.Local().Format(
				"2006-01-02 15:04",
			)
		}
		line(label("Last connected: ") + colorMuted + last + ansiReset)

		notes := "-"
		if host.Notes != "" {
			notes = host.Notes
		}
		line(label("Notes:          ") + colorMuted + notes + ansiReset)

		line()
		line(helpText("C Connect | B Back | Q Quit"))
		if status != "" {
			line()
			line(statusText(status))
		}

		k, err := r.readKey()
		if err != nil {
			return actionNone, err.Error()
		}
		switch k {
		case keyConnect:
			entry, err := r.svc.Connect(
				r.ctx,
				host.Alias,
			)
			if err != nil {
				status = err.Error()
			} else {
				status = "Opened " + entry.AliasSnapshot + " in " + entry.TerminalApp
			}
		case keyBack:
			return actionNone, ""
		case keyQuit:
			return actionQuit, ""
		}
	}
}

func (r *runner) addHost() (
	action,
	string,
) {
	input, err := r.promptHost()
	if err != nil {
		return actionNone, err.Error()
	}
	if input.Alias == "" {
		return actionNone, "Add cancelled."
	}

	host, err := r.svc.AddHost(
		r.ctx,
		input,
	)
	if err != nil {
		return actionNone, err.Error()
	}
	return actionNone, "Added " + host.Alias
}

func (r *runner) importSSHConfig() (
	action,
	string,
) {
	clear()
	line(colorYellow + "Import SSH config" + ansiReset)
	line()
	line(colorMuted + "Path empty = ~/.ssh/config" + ansiReset)
	path, err := r.readLine("Path: ")
	if err != nil {
		return actionNone, err.Error()
	}

	created, updated, err := r.svc.ImportSSHConfig(
		r.ctx,
		strings.TrimSpace(path),
	)
	if err != nil {
		return actionNone, err.Error()
	}
	return actionNone, fmt.Sprintf(
		"Import done: %d created, %d updated.",
		created,
		updated,
	)
}

func (r *runner) history() (
	action,
	string,
) {
	items, err := r.svc.History(
		r.ctx,
		20,
	)
	if err != nil {
		return actionNone, err.Error()
	}

	clear()
	renderLogo()
	line()
	line(colorYellow + "History" + ansiReset)
	line()
	if len(items) == 0 {
		line(colorMuted + "No connection history yet." + ansiReset)
	} else {
		for _, item := range items {
			format(
				"%s%s%s  %s%-18s%s  %s%s%s  %s%s%s\r\n",
				colorMuted,
				item.StartedAt.Local().Format(
					time.RFC3339,
				),
				ansiReset,
				colorCyan,
				item.AliasSnapshot,
				ansiReset,
				colorGreen,
				item.TerminalApp,
				ansiReset,
				colorMuted,
				item.Command,
				ansiReset,
			)
		}
	}
	line()
	line(helpText("Press any key."))
	_, _ = r.readKey()
	return actionNone, ""
}

func (r *runner) status() (
	action,
	string,
) {
	hosts, err := r.svc.ListHosts(
		r.ctx,
		"",
	)
	if err != nil {
		return actionNone, err.Error()
	}
	history, err := r.svc.History(
		r.ctx,
		200,
	)
	if err != nil {
		return actionNone, err.Error()
	}

	clear()
	renderLogo()
	line()
	line(colorYellow + "Status" + ansiReset)
	line()
	line(label("DB:      ") + colorMuted + r.dbPath + ansiReset)
	line(label("Hosts:   ") + colorCyan + strconv.Itoa(len(hosts)) + ansiReset)
	line(label("History: ") + colorCyan + strconv.Itoa(len(history)) + ansiReset)
	line()
	line(helpText("Press any key."))
	_, _ = r.readKey()
	return actionNone, ""
}

func (r *runner) promptHost() (
	model.HostInput,
	error,
) {
	clear()
	line(colorYellow + "Add host" + ansiReset)
	line()
	line(colorMuted + "Alias empty cancels." + ansiReset)

	alias, err := r.readLine("Alias: ")
	if err != nil {
		return model.HostInput{}, err
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return model.HostInput{}, nil
	}

	hostname, err := r.readLine("Hostname [" + alias + "]: ")
	if err != nil {
		return model.HostInput{}, err
	}
	username, err := r.readLine("Username [root]: ")
	if err != nil {
		return model.HostInput{}, err
	}
	portText, err := r.readLine("Port [22]: ")
	if err != nil {
		return model.HostInput{}, err
	}
	key, err := r.readLine("Identity file: ")
	if err != nil {
		return model.HostInput{}, err
	}
	notes, err := r.readLine("Notes: ")
	if err != nil {
		return model.HostInput{}, err
	}

	port := 22
	portText = strings.TrimSpace(portText)
	if portText != "" {
		parsed, err := strconv.Atoi(portText)
		if err != nil {
			return model.HostInput{}, fmt.Errorf(
				"invalid port: %w",
				err,
			)
		}
		port = parsed
	}

	return model.HostInput{
		Alias:        alias,
		Hostname:     strings.TrimSpace(hostname),
		Port:         port,
		Username:     strings.TrimSpace(username),
		IdentityFile: strings.TrimSpace(key),
		Notes:        strings.TrimSpace(notes),
	}, nil
}

func (r *runner) readKey() (
	key,
	error,
) {
	runeValue, _, err := r.reader.ReadRune()
	if err != nil {
		return keyUnknown, err
	}

	switch runeValue {
	case 'q', 'Q', 'й', 'Й':
		return keyQuit, nil
	case 'b', 'B', 'и', 'И':
		return keyBack, nil
	case 'c', 'C', 'с', 'С':
		return keyConnect, nil
	case '\r', '\n':
		return keyEnter, nil
	case 'k', 'K', 'л', 'Л':
		return keyDown, nil
	case 'j', 'J', 'о', 'О':
		return keyUp, nil
	case 27:
		next, err := r.reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		if next != '[' {
			return keyUnknown, nil
		}
		final, err := r.reader.ReadByte()
		if err != nil {
			return keyUnknown, nil
		}
		switch final {
		case 'A':
			return keyUp, nil
		case 'B':
			return keyDown, nil
		}
	}

	if runeValue >= '1' && runeValue <= '9' {
		return keyEnter, nil
	}
	return keyUnknown, nil
}

func (r *runner) readLine(
	label string,
) (
	string,
	error,
) {
	if err := r.suspendRaw(); err != nil {
		return "", err
	}
	defer func() {
		_ = r.resumeRaw()
	}()

	fmt.Print("\r\n")
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(
		value,
		"\r\n",
	), nil
}

func (r *runner) suspendRaw() error {
	if !term.IsTerminal(
		int(
			os.Stdin.Fd(),
		),
	) {
		return nil
	}
	write(ansiCursorBlinkOn + ansiCursorShow + ansiReset)
	return term.Restore(
		int(
			os.Stdin.Fd(),
		),
		r.oldState,
	)
}

func (r *runner) resumeRaw() error {
	if !term.IsTerminal(
		int(
			os.Stdin.Fd(),
		),
	) {
		return nil
	}
	state, err := term.MakeRaw(
		int(
			os.Stdin.Fd(),
		),
	)
	if err != nil {
		return err
	}
	r.oldState = state
	r.reader = bufio.NewReader(os.Stdin)
	write(ansiCursorBlinkOff + ansiCursorHide)
	return nil
}

func renderLogo() {
	lines := strings.Split(
		strings.TrimRight(
			logo,
			"\n",
		),
		"\n",
	)
	for i, text := range lines {
		if i < 6 {
			line(colorGreen + text + ansiReset)
			continue
		}
		prefix, link, ok := strings.Cut(
			text,
			": ",
		)
		if !ok {
			line(colorMuted + text + ansiReset)
			continue
		}
		line(colorMuted + prefix + ": " + colorBlue + link + ansiReset)
	}
}

func label(
	value string,
) string {
	return colorMuted + value + ansiReset
}

func helpText(
	value string,
) string {
	return colorMuted + value + ansiReset
}

func statusText(
	value string,
) string {
	if strings.Contains(
		strings.ToLower(value),
		"error",
	) || strings.Contains(
		strings.ToLower(value),
		"invalid",
	) {
		return colorRed + value + ansiReset
	}
	return colorGreen + value + ansiReset
}

func clear() {
	write("\033[H\033[2J\033[3J")
}

func line(
	values ...any,
) {
	write(fmt.Sprintln(values...))
}

func format(
	template string,
	values ...any,
) {
	write(fmt.Sprintf(
		template,
		values...,
	))
}

func write(
	value string,
) {
	value = strings.ReplaceAll(
		value,
		"\r\n",
		"\n",
	)
	fmt.Print(
		strings.ReplaceAll(
			value,
			"\n",
			"\r\n",
		),
	)
}
