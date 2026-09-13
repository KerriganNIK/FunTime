package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("state not found")
var validKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,100}$`)

type Store interface {
	Load(context.Context, string, any) error
	Save(context.Context, string, any) error
	Keys(context.Context) ([]string, error)
	Close()
}

// Each service opens its own store and database. File mode is for a single local process.
func Open(ctx context.Context, databaseURL, dataDir string) (Store, error) {
	if databaseURL != "" {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			return nil, err
		}
		if _, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS state_documents (id text PRIMARY KEY, payload jsonb NOT NULL, updated_at timestamptz NOT NULL DEFAULT now())`); err != nil {
			pool.Close()
			return nil, err
		}
		return &postgres{pool: pool}, nil
	}
	if dataDir == "" {
		return nil, errors.New("DATABASE_URL or DATA_DIR is required")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	return &files{dir: dataDir}, nil
}

type files struct{ dir string }

func (s *files) Load(_ context.Context, key string, dest any) error {
	if !validKey.MatchString(key) {
		return ErrNotFound
	}
	b, err := os.ReadFile(filepath.Join(s.dir, key+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
func (s *files) Save(_ context.Context, key string, value any) error {
	if !validKey.MatchString(key) {
		return errors.New("invalid state key")
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(s.dir, ".state-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, filepath.Join(s.dir, key+".json"))
}
func (s *files) Keys(context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			keys = append(keys, entry.Name()[:len(entry.Name())-5])
		}
	}
	return keys, nil
}
func (*files) Close() {}

type postgres struct{ pool *pgxpool.Pool }

func (s *postgres) Load(ctx context.Context, key string, dest any) error {
	var b []byte
	err := s.pool.QueryRow(ctx, `SELECT payload FROM state_documents WHERE id=$1`, key).Scan(&b)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
func (s *postgres) Save(ctx context.Context, key string, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO state_documents(id,payload) VALUES ($1,$2) ON CONFLICT(id) DO UPDATE SET payload=EXCLUDED.payload,updated_at=now()`, key, b)
	return err
}
func (s *postgres) Keys(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT id FROM state_documents ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}
func (s *postgres) Close() { s.pool.Close() }

func Clone[T any](value *T) (*T, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("clone state: %w", err)
	}
	var copy T
	err = json.Unmarshal(b, &copy)
	return &copy, err
}
