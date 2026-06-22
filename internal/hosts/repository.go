package hosts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"pinter/internal/ids"
	"pinter/internal/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(
	db *sql.DB,
) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	input model.HostInput,
) (
	model.Host,
	error,
) {
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return model.Host{}, err
	}

	id, err := ids.New("hst")
	if err != nil {
		return model.Host{}, err
	}

	now := time.Now().UTC()
	host := model.Host{
		ID:           id,
		Alias:        input.Alias,
		Hostname:     input.Hostname,
		Port:         input.Port,
		Username:     input.Username,
		IdentityFile: input.IdentityFile,
		Notes:        input.Notes,
		Favorite:     input.Favorite,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO hosts
		(id, alias, hostname, port, username, identity_file, notes, favorite, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		host.ID,
		host.Alias,
		host.Hostname,
		host.Port,
		host.Username,
		host.IdentityFile,
		host.Notes,
		boolToInt(host.Favorite),
		formatTime(host.CreatedAt),
		formatTime(host.UpdatedAt),
	)
	if err != nil {
		return model.Host{}, fmt.Errorf(
			"create host: %w",
			err,
		)
	}
	return host, nil
}

func (r *Repository) UpsertByAlias(
	ctx context.Context,
	input model.HostInput,
) (
	model.Host,
	bool,
	error,
) {
	existing, err := r.GetByAlias(
		ctx,
		input.Alias,
	)
	if err == nil {
		input = normalizeInput(input)
		if input.Notes == "" {
			input.Notes = existing.Notes
		}
		if err := r.Update(
			ctx,
			existing.ID,
			input,
		); err != nil {
			return model.Host{}, false, err
		}
		updated, err := r.GetByID(
			ctx,
			existing.ID,
		)
		return updated, false, err
	}
	if !errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return model.Host{}, false, err
	}

	created, err := r.Create(
		ctx,
		input,
	)
	return created, true, err
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	input model.HostInput,
) error {
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return err
	}

	res, err := r.db.ExecContext(
		ctx,
		`UPDATE hosts
		SET alias = ?, hostname = ?, port = ?, username = ?, identity_file = ?, notes = ?, favorite = ?, updated_at = ?
		WHERE id = ?`,
		input.Alias,
		input.Hostname,
		input.Port,
		input.Username,
		input.IdentityFile,
		input.Notes,
		boolToInt(input.Favorite),
		formatTime(time.Now().UTC()),
		id,
	)
	if err != nil {
		return fmt.Errorf(
			"update host: %w",
			err,
		)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"update host rows affected: %w",
			err,
		)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) List(
	ctx context.Context,
	query string,
) (
	[]model.Host,
	error,
) {
	query = strings.TrimSpace(query)
	args := []any{}
	sqlQuery := `SELECT id, alias, hostname, port, username, identity_file, notes, favorite,
		last_connected_at, created_at, updated_at FROM hosts`
	if query != "" {
		like := "%" + query + "%"
		sqlQuery += ` WHERE alias LIKE ? OR hostname LIKE ? OR username LIKE ? OR notes LIKE ?`
		args = append(
			args,
			like,
			like,
			like,
			like,
		)
	}
	sqlQuery += ` ORDER BY favorite DESC, COALESCE(last_connected_at, created_at) DESC, alias ASC`

	rows, err := r.db.QueryContext(
		ctx,
		sqlQuery,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list hosts: %w",
			err,
		)
	}
	defer rows.Close()

	var out []model.Host
	for rows.Next() {
		host, err := scanHost(rows)
		if err != nil {
			return nil, err
		}
		out = append(
			out,
			host,
		)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"list hosts rows: %w",
			err,
		)
	}
	return out, nil
}

func (r *Repository) GetByAlias(
	ctx context.Context,
	alias string,
) (
	model.Host,
	error,
) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, alias, hostname, port, username, identity_file, notes, favorite,
		last_connected_at, created_at, updated_at FROM hosts WHERE alias = ?`,
		strings.TrimSpace(alias),
	)
	return scanHost(row)
}

func (r *Repository) GetByID(
	ctx context.Context,
	id string,
) (
	model.Host,
	error,
) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, alias, hostname, port, username, identity_file, notes, favorite,
		last_connected_at, created_at, updated_at FROM hosts WHERE id = ?`,
		id,
	)
	return scanHost(row)
}

func (r *Repository) MarkConnected(
	ctx context.Context,
	hostID string,
	at time.Time,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE hosts SET last_connected_at = ?, updated_at = ? WHERE id = ?`,
		formatTime(at),
		formatTime(time.Now().UTC()),
		hostID,
	)
	if err != nil {
		return fmt.Errorf(
			"mark connected: %w",
			err,
		)
	}
	return nil
}

func (r *Repository) AddHistory(
	ctx context.Context,
	host model.Host,
	command string,
	terminalApp string,
	startedAt time.Time,
) (
	model.ConnectionHistory,
	error,
) {
	id, err := ids.New("hstlog")
	if err != nil {
		return model.ConnectionHistory{}, err
	}

	entry := model.ConnectionHistory{
		ID:            id,
		HostID:        host.ID,
		AliasSnapshot: host.Alias,
		Command:       command,
		StartedAt:     startedAt,
		TerminalApp:   terminalApp,
	}
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO connection_history
		(id, host_id, alias_snapshot, command, started_at, terminal_app)
		VALUES (?, ?, ?, ?, ?, ?)`,
		entry.ID,
		entry.HostID,
		entry.AliasSnapshot,
		entry.Command,
		formatTime(entry.StartedAt),
		entry.TerminalApp,
	)
	if err != nil {
		return model.ConnectionHistory{}, fmt.Errorf(
			"add history: %w",
			err,
		)
	}
	return entry, nil
}

func (r *Repository) History(
	ctx context.Context,
	limit int,
) (
	[]model.ConnectionHistory,
	error,
) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, host_id, alias_snapshot, command, started_at, exit_status, terminal_app
		FROM connection_history ORDER BY started_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list history: %w",
			err,
		)
	}
	defer rows.Close()

	var out []model.ConnectionHistory
	for rows.Next() {
		entry, err := scanHistory(rows)
		if err != nil {
			return nil, err
		}
		out = append(
			out,
			entry,
		)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"list history rows: %w",
			err,
		)
	}
	return out, nil
}

func normalizeInput(
	input model.HostInput,
) model.HostInput {
	input.Alias = strings.TrimSpace(input.Alias)
	input.Hostname = strings.TrimSpace(input.Hostname)
	input.Username = strings.TrimSpace(input.Username)
	input.IdentityFile = strings.TrimSpace(input.IdentityFile)
	input.Notes = strings.TrimSpace(input.Notes)
	if input.Port == 0 {
		input.Port = 22
	}
	if input.Username == "" {
		input.Username = "root"
	}
	if input.Hostname == "" {
		input.Hostname = input.Alias
	}
	return input
}

func validateInput(
	input model.HostInput,
) error {
	if input.Alias == "" {
		return errors.New("alias is required")
	}
	if input.Hostname == "" {
		return errors.New("hostname is required")
	}
	if input.Port <= 0 || input.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if input.Username == "" {
		return errors.New("username is required")
	}
	return nil
}

type hostScanner interface {
	Scan(dest ...any) error
}

func scanHost(
	scanner hostScanner,
) (
	model.Host,
	error,
) {
	var host model.Host
	var favorite int
	var lastConnected sql.NullString
	var createdAt, updatedAt string
	if err := scanner.Scan(
		&host.ID,
		&host.Alias,
		&host.Hostname,
		&host.Port,
		&host.Username,
		&host.IdentityFile,
		&host.Notes,
		&favorite,
		&lastConnected,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.Host{}, err
	}

	host.Favorite = favorite == 1
	if lastConnected.Valid && lastConnected.String != "" {
		parsed, err := parseTime(lastConnected.String)
		if err != nil {
			return model.Host{}, err
		}
		host.LastConnectedAt = &parsed
	}
	var err error
	host.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return model.Host{}, err
	}
	host.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return model.Host{}, err
	}
	return host, nil
}

func scanHistory(
	scanner hostScanner,
) (
	model.ConnectionHistory,
	error,
) {
	var entry model.ConnectionHistory
	var startedAt string
	var exitStatus sql.NullInt64
	if err := scanner.Scan(
		&entry.ID,
		&entry.HostID,
		&entry.AliasSnapshot,
		&entry.Command,
		&startedAt,
		&exitStatus,
		&entry.TerminalApp,
	); err != nil {
		return model.ConnectionHistory{}, err
	}
	parsed, err := parseTime(startedAt)
	if err != nil {
		return model.ConnectionHistory{}, err
	}
	entry.StartedAt = parsed
	if exitStatus.Valid {
		status := int(exitStatus.Int64)
		entry.ExitStatus = &status
	}
	return entry, nil
}

func formatTime(
	t time.Time,
) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(
	value string,
) (
	time.Time,
	error,
) {
	return time.Parse(
		time.RFC3339Nano,
		value,
	)
}

func boolToInt(
	v bool,
) int {
	if v {
		return 1
	}
	return 0
}
