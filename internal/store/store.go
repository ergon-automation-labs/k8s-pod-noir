package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "history.db")
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY,
	scenario_id TEXT NOT NULL,
	detective_name TEXT NOT NULL,
	started_at TEXT NOT NULL,
	ended_at TEXT,
	outcome TEXT
);
CREATE TABLE IF NOT EXISTS case_notes (
	id INTEGER PRIMARY KEY,
	session_id INTEGER NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
	kind TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_case_notes_session ON case_notes(session_id);
`)
	return err
}

func (s *Store) StartSession(ctx context.Context, scenarioID, detective string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (scenario_id, detective_name, started_at) VALUES (?, ?, ?)`,
		scenarioID, detective, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) EndSession(ctx context.Context, sessionID int64, outcome string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET ended_at = ?, outcome = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), outcome, sessionID,
	)
	return err
}

func (s *Store) AddNote(ctx context.Context, sessionID int64, kind, body string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO case_notes (session_id, kind, body, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, kind, body, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) Notes(ctx context.Context, sessionID int64) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT kind, body, created_at FROM case_notes WHERE session_id = ? ORDER BY id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.Kind, &n.Body, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

type Note struct {
	Kind      string
	Body      string
	CreatedAt string
}
