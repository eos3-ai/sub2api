package payloadlog

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Wei-Shaw/sub2api/internal/config"
)

//go:embed schema.sql
var schemaSQL string

var databaseNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type ClickHouseStore struct {
	conn     clickhouse.Conn
	database string
}

func NewClickHouseStore(ctx context.Context, cfg config.PayloadLogClickHouseConfig) (*ClickHouseStore, error) {
	if !databaseNamePattern.MatchString(cfg.Database) {
		return nil, fmt.Errorf("invalid ClickHouse database name %q", cfg.Database)
	}
	addresses := strings.Split(cfg.Addr, ",")
	for i := range addresses {
		addresses[i] = strings.TrimSpace(addresses[i])
		if addresses[i] == "" {
			return nil, fmt.Errorf("ClickHouse address is empty")
		}
	}
	compression := clickhouse.CompressionLZ4
	switch cfg.Compression {
	case "none":
		compression = clickhouse.CompressionNone
	case "zstd":
		compression = clickhouse.CompressionZSTD
	}
	opts := &clickhouse.Options{Addr: addresses, Auth: clickhouse.Auth{Database: "default", Username: cfg.Username, Password: cfg.Password}, DialTimeout: cfg.DialTimeout, MaxOpenConns: cfg.MaxOpenConns, Compression: &clickhouse.Compression{Method: compression}}
	if cfg.TLS {
		opts.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open ClickHouse: %w", err)
	}
	s := &ClickHouseStore{conn: conn, database: cfg.Database}
	if err := s.init(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return s, nil
}

func (s *ClickHouseStore) init(ctx context.Context) error {
	if err := s.conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping ClickHouse: %w", err)
	}
	if err := s.conn.Exec(ctx, "CREATE DATABASE IF NOT EXISTS "+s.database); err != nil {
		return fmt.Errorf("create payload log database: %w", err)
	}
	q := strings.ReplaceAll(schemaSQL, "{{database}}", s.database)
	if err := s.conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("create payload log table: %w", err)
	}
	return nil
}
func (s *ClickHouseStore) Ping(ctx context.Context) error { return s.conn.Ping(ctx) }
func (s *ClickHouseStore) Close() error                   { return s.conn.Close() }

const columns = `id, request_id, created_at, username, api_key_id, api_key_name, account_id, model_name, method, path, status_code, is_stream, request_content_type, response_content_type, request_bytes, response_bytes`

func (s *ClickHouseStore) InsertBatch(ctx context.Context, logs []*PayloadLog) error {
	if len(logs) == 0 {
		return nil
	}
	b, err := s.conn.PrepareBatch(ctx, "INSERT INTO "+s.database+".payload_logs ("+columns+", request_body, response_body)")
	if err != nil {
		return err
	}
	for _, l := range logs {
		if l == nil {
			continue
		}
		if err = b.Append(l.ID, l.RequestID, l.CreatedAt.UTC(), l.Username, l.APIKeyID, l.APIKeyName, l.AccountID, l.ModelName, l.Method, l.Path, l.StatusCode, l.IsStream, l.RequestContentType, l.ResponseContentType, l.RequestBytes, l.ResponseBytes, l.RequestBody, l.ResponseBody); err != nil {
			return err
		}
	}
	return b.Send()
}

func (s *ClickHouseStore) List(ctx context.Context, f PayloadLogFilter, p Pagination) ([]PayloadLogMeta, int64, error) {
	where, args := filters(f)
	var total int64
	if err := s.conn.QueryRow(ctx, "SELECT count() FROM "+s.database+".payload_logs"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)
	rows, err := s.conn.Query(ctx, "SELECT "+columns+" FROM "+s.database+".payload_logs"+where+" ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]PayloadLogMeta, 0)
	for rows.Next() {
		var m PayloadLogMeta
		if err := rows.Scan(&m.ID, &m.RequestID, &m.CreatedAt, &m.Username, &m.APIKeyID, &m.APIKeyName, &m.AccountID, &m.ModelName, &m.Method, &m.Path, &m.StatusCode, &m.IsStream, &m.RequestContentType, &m.ResponseContentType, &m.RequestBytes, &m.ResponseBytes); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func filters(f PayloadLogFilter) (string, []any) {
	var clauses []string
	var args []any
	add := func(q string, v any) { clauses = append(clauses, q); args = append(args, v) }
	if f.Username != "" {
		add("username = ?", f.Username)
	}
	if f.APIKeyName != "" {
		add("api_key_name = ?", f.APIKeyName)
	}
	if f.ModelName != "" {
		add("model_name = ?", f.ModelName)
	}
	if f.Start != nil {
		add("created_at >= ?", f.Start.UTC())
	}
	if f.End != nil {
		add("created_at <= ?", f.End.UTC())
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (s *ClickHouseStore) GetByRequestID(ctx context.Context, id string, tr *TimeRange) ([]*PayloadLog, error) {
	where := " WHERE request_id = ?"
	args := []any{id}
	if tr != nil {
		if tr.Start != nil {
			where += " AND created_at >= ?"
			args = append(args, tr.Start.UTC())
		}
		if tr.End != nil {
			where += " AND created_at <= ?"
			args = append(args, tr.End.UTC())
		}
	}
	rows, err := s.conn.Query(ctx, "SELECT "+columns+", request_body, response_body FROM "+s.database+".payload_logs"+where+" ORDER BY created_at ASC, id ASC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*PayloadLog, 0)
	for rows.Next() {
		l := new(PayloadLog)
		if err := rows.Scan(&l.ID, &l.RequestID, &l.CreatedAt, &l.Username, &l.APIKeyID, &l.APIKeyName, &l.AccountID, &l.ModelName, &l.Method, &l.Path, &l.StatusCode, &l.IsStream, &l.RequestContentType, &l.ResponseContentType, &l.RequestBytes, &l.ResponseBytes, &l.RequestBody, &l.ResponseBody); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

var _ Store = (*ClickHouseStore)(nil)
