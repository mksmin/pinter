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
	run         func() (action, string)
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
)

const logo = ` ____  _       _            
|  _ \(_)_ __ | |_ ___ _ __ 
| |_) | | '_ \| __/ _ \ '__|
|  __/| | | | | ||  __/ |   
|_|   |_|_| |_|\__\___|_|   
Local SSH keeper`

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
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enable raw terminal: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

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
						run: func() (action, string) {
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
) (action, string, error) {
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
	fmt.Println(logo)
	fmt.Println()
	fmt.Println("Update check disabled. Local-only MVP.")
	fmt.Println()

	for i, item := range m.items {
		cursor := "  "
		if i == selected {
			cursor = "> "
		}
		fmt.Printf(
			"%s%d. %-10s %s\r\n",
			cursor,
			i+1,
			item.title,
			item.description,
		)
	}

	fmt.Println()
	fmt.Println("Up/Down | Enter | B Back | Q Quit")
	if status != "" {
		fmt.Println()
		fmt.Println(status)
	}
}

func (r *runner) hosts() (action, string) {
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
	status := "Enter connects selected host. B returns."
	for {
		clear()
		fmt.Println(logo)
		fmt.Println()
		fmt.Println("Hosts")
		fmt.Println()

		for i, host := range items {
			cursor := "  "
			if i == index {
				cursor = "> "
			}
			last := "-"
			if host.LastConnectedAt != nil {
				last = host.LastConnectedAt.Local().Format("2006-01-02 15:04")
			}
			fmt.Printf(
				"%s%-18s %s@%s:%d  last=%s\r\n",
				cursor,
				host.Alias,
				host.Username,
				host.Hostname,
				host.Port,
				last,
			)
		}

		fmt.Println()
		fmt.Println("Up/Down | Enter Connect | B Back | Q Quit")
		fmt.Println()
		fmt.Println(status)

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

func (r *runner) addHost() (action, string) {
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

func (r *runner) importSSHConfig() (action, string) {
	clear()
	fmt.Println("Import SSH config")
	fmt.Println()
	fmt.Println("Path empty = ~/.ssh/config")
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

func (r *runner) history() (action, string) {
	items, err := r.svc.History(
		r.ctx,
		20,
	)
	if err != nil {
		return actionNone, err.Error()
	}

	clear()
	fmt.Println(logo)
	fmt.Println()
	fmt.Println("History")
	fmt.Println()
	if len(items) == 0 {
		fmt.Println("No connection history yet.")
	} else {
		for _, item := range items {
			fmt.Printf(
				"%s  %-18s  %s  %s\r\n",
				item.StartedAt.Local().Format(time.RFC3339),
				item.AliasSnapshot,
				item.TerminalApp,
				item.Command,
			)
		}
	}
	fmt.Println()
	fmt.Println("Press any key.")
	_, _ = r.readKey()
	return actionNone, ""
}

func (r *runner) status() (action, string) {
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
	fmt.Println(logo)
	fmt.Println()
	fmt.Println("Status")
	fmt.Println()
	fmt.Println("DB:      " + r.dbPath)
	fmt.Println("Hosts:   " + strconv.Itoa(len(hosts)))
	fmt.Println("History: " + strconv.Itoa(len(history)))
	fmt.Println()
	fmt.Println("Press any key.")
	_, _ = r.readKey()
	return actionNone, ""
}

func (r *runner) promptHost() (model.HostInput, error) {
	clear()
	fmt.Println("Add host")
	fmt.Println()
	fmt.Println("Alias empty cancels.")

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
			return model.HostInput{}, fmt.Errorf("invalid port: %w", err)
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

func (r *runner) readKey() (key, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return keyUnknown, err
	}

	switch b {
	case 'q', 'Q':
		return keyQuit, nil
	case 'b', 'B':
		return keyBack, nil
	case '\r', '\n':
		return keyEnter, nil
	case 'k', 'K':
		return keyUp, nil
	case 'j', 'J':
		return keyDown, nil
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

	if b >= '1' && b <= '9' {
		return keyEnter, nil
	}
	return keyUnknown, nil
}

func (r *runner) readLine(
	label string,
) (string, error) {
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
	return strings.TrimRight(value, "\r\n"), nil
}

func (r *runner) suspendRaw() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	return term.Restore(
		int(os.Stdin.Fd()),
		r.oldState,
	)
}

func (r *runner) resumeRaw() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	r.oldState = state
	r.reader = bufio.NewReader(os.Stdin)
	return nil
}

func clear() {
	fmt.Print("\033[H\033[2J")
}
