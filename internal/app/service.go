package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pinter/internal/db"
	"pinter/internal/hosts"
	"pinter/internal/model"
	"pinter/internal/sshconfig"
	"pinter/internal/terminal"
)

type Service struct {
	db       *sql.DB
	hosts    *hosts.Repository
	terminal *terminal.Launcher
}

func NewService(
	dbPath string,
) (
	*Service,
	error,
) {
	conn, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		db:       conn,
		hosts:    hosts.NewRepository(conn),
		terminal: terminal.NewLauncher(),
	}, nil
}

func (s *Service) Close() error {
	return s.db.Close()
}

func (s *Service) AddHost(
	ctx context.Context,
	input model.HostInput,
) (
	model.Host,
	error,
) {
	return s.hosts.Create(
		ctx,
		input,
	)
}

func (s *Service) ListHosts(
	ctx context.Context,
	query string,
) (
	[]model.Host,
	error,
) {
	return s.hosts.List(
		ctx,
		query,
	)
}

func (s *Service) Connect(
	ctx context.Context,
	alias string,
) (
	model.ConnectionHistory,
	error,
) {
	host, err := s.hosts.GetByAlias(
		ctx,
		alias,
	)
	if err != nil {
		return model.ConnectionHistory{}, fmt.Errorf(
			"find host %q: %w",
			alias,
			err,
		)
	}

	startedAt := time.Now().UTC()
	command, terminalApp, err := s.terminal.Open(host)
	if err != nil {
		return model.ConnectionHistory{}, err
	}
	if err := s.hosts.MarkConnected(
		ctx,
		host.ID,
		startedAt,
	); err != nil {
		return model.ConnectionHistory{}, err
	}
	return s.hosts.AddHistory(
		ctx,
		host,
		command,
		terminalApp,
		startedAt,
	)
}

func (s *Service) History(
	ctx context.Context,
	limit int,
) (
	[]model.ConnectionHistory,
	error,
) {
	return s.hosts.History(
		ctx,
		limit,
	)
}

func (s *Service) ImportSSHConfig(
	ctx context.Context,
	path string,
) (
	int,
	int,
	error,
) {
	if path == "" {
		path = sshconfig.DefaultConfigPath()
	}
	parsed, err := sshconfig.ParseFile(path)
	if err != nil {
		return 0, 0, err
	}

	created := 0
	updated := 0
	for _, item := range parsed {
		_, wasCreated, err := s.hosts.UpsertByAlias(
			ctx,
			item.HostInput(),
		)
		if err != nil {
			return created, updated, err
		}
		if wasCreated {
			created++
		} else {
			updated++
		}
	}
	return created, updated, nil
}
