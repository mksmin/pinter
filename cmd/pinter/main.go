package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"pinter/internal/app"
	"pinter/internal/config"
	"pinter/internal/model"
	"pinter/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(
			os.Stderr,
			"pinter:",
			err,
		)
		os.Exit(1)
	}
}

func run(
	args []string,
) error {
	if len(args) == 0 {
		dbPath, err := config.DBPath()
		if err != nil {
			return err
		}
		svc, err := app.NewService(dbPath)
		if err != nil {
			return err
		}
		defer svc.Close()

		return tui.Run(
			context.Background(),
			svc,
			dbPath,
		)
	}

	dbPath, err := config.DBPath()
	if err != nil {
		return err
	}
	svc, err := app.NewService(dbPath)
	if err != nil {
		return err
	}
	defer svc.Close()

	ctx := context.Background()
	switch args[0] {
	case "add":
		return add(
			ctx,
			svc,
			args[1:],
		)
	case "list", "ls":
		return list(
			ctx,
			svc,
			args[1:],
		)
	case "connect", "ssh":
		return connect(
			ctx,
			svc,
			args[1:],
		)
	case "history":
		return history(
			ctx,
			svc,
			args[1:],
		)
	case "import-ssh-config":
		return importSSHConfig(
			ctx,
			svc,
			args[1:],
		)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		return fmt.Errorf(
			"unknown command %q",
			args[0],
		)
	}
}

func add(
	ctx context.Context,
	svc *app.Service,
	args []string,
) error {
	fs := flag.NewFlagSet(
		"add",
		flag.ContinueOnError,
	)
	alias := fs.String(
		"alias",
		"",
		"host alias",
	)
	hostname := fs.String(
		"host",
		"",
		"hostname",
	)
	port := fs.Int(
		"port",
		22,
		"ssh port",
	)
	user := fs.String(
		"user",
		"",
		"ssh username",
	)
	key := fs.String(
		"key",
		"",
		"identity file",
	)
	notes := fs.String(
		"notes",
		"",
		"notes",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *alias == "" && fs.NArg() > 0 {
		*alias = fs.Arg(0)
	}

	host, err := svc.AddHost(
		ctx,
		model.HostInput{
			Alias:        *alias,
			Hostname:     *hostname,
			Port:         *port,
			Username:     *user,
			IdentityFile: *key,
			Notes:        *notes,
		},
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Added %s (%s@%s:%d)\n",
		host.Alias,
		host.Username,
		host.Hostname,
		host.Port,
	)
	return nil
}

func list(
	ctx context.Context,
	svc *app.Service,
	args []string,
) error {
	fs := flag.NewFlagSet(
		"list",
		flag.ContinueOnError,
	)
	query := fs.String(
		"q",
		"",
		"search query",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	items, err := svc.ListHosts(
		ctx,
		*query,
	)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("No hosts yet. Add one with `pinter add --alias prod --host 10.0.0.1 --user deploy`.")
		return nil
	}

	w := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	fmt.Fprintln(
		w,
		"ALIAS\tTARGET\tKEY\tLAST CONNECTED\tNOTES",
	)
	for _, host := range items {
		last := "-"
		if host.LastConnectedAt != nil {
			last = host.LastConnectedAt.Local().Format(
				"2006-01-02 15:04",
			)
		}
		key := "-"
		if host.IdentityFile != "" {
			key = host.IdentityFile
		}
		fmt.Fprintf(
			w,
			"%s\t%s@%s:%d\t%s\t%s\t%s\n",
			host.Alias,
			host.Username,
			host.Hostname,
			host.Port,
			key,
			last,
			oneLine(host.Notes),
		)
	}
	return w.Flush()
}

func connect(
	ctx context.Context,
	svc *app.Service,
	args []string,
) error {
	if len(args) != 1 {
		return errors.New("usage: pinter connect <alias>")
	}
	entry, err := svc.Connect(
		ctx,
		args[0],
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Opened %s with %s\n",
		entry.AliasSnapshot,
		entry.TerminalApp,
	)
	fmt.Printf(
		"%s\n",
		entry.Command,
	)
	return nil
}

func history(
	ctx context.Context,
	svc *app.Service,
	args []string,
) error {
	fs := flag.NewFlagSet(
		"history",
		flag.ContinueOnError,
	)
	limit := fs.Int(
		"limit",
		20,
		"history limit",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	items, err := svc.History(
		ctx,
		*limit,
	)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("No connection history yet.")
		return nil
	}

	w := tabwriter.NewWriter(
		os.Stdout,
		0,
		0,
		2,
		' ',
		0,
	)
	fmt.Fprintln(
		w,
		"WHEN\tALIAS\tTERMINAL\tCOMMAND",
	)
	for _, item := range items {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\n",
			item.StartedAt.Local().Format(
				time.RFC3339,
			),
			item.AliasSnapshot,
			item.TerminalApp,
			item.Command,
		)
	}
	return w.Flush()
}

func importSSHConfig(
	ctx context.Context,
	svc *app.Service,
	args []string,
) error {
	fs := flag.NewFlagSet(
		"import-ssh-config",
		flag.ContinueOnError,
	)
	path := fs.String(
		"path",
		"",
		"ssh config path",
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	created, updated, err := svc.ImportSSHConfig(
		ctx,
		*path,
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Imported SSH config: %s created, %s updated\n",
		strconv.Itoa(created),
		strconv.Itoa(updated),
	)
	return nil
}

func usage() {
	fmt.Println(
		`pinter - local SSH keeper

Usage:
  pinter add --alias prod --host 10.0.0.1 --user deploy --key ~/.ssh/prod
  pinter list [-q query]
  pinter connect <alias>
  pinter history [--limit 20]
  pinter import-ssh-config [--path ~/.ssh/config]

Environment:
  PINTER_DB_PATH    Override SQLite database path
  PINTER_DATA_DIR   Override data directory
  PINTER_TERMINAL   Windows terminal: auto, wt, pwsh, powershell, cmd`,
	)
}

func oneLine(
	value string,
) string {
	value = strings.ReplaceAll(
		value,
		"\n",
		" ",
	)
	if len(value) > 48 {
		return value[:45] + "..."
	}
	if value == "" {
		return "-"
	}
	return value
}
