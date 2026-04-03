package store

import (
	"context"
	"database/sql"
	"time"
)

// CaseFolder aggregates how often a scenario was opened and closed cleanly.
type CaseFolder struct {
	ScenarioID    string
	OpenCount     int
	SolvedCount   int
	LastDetective string
	LastOpenedAt  string
	LastOutcome   string
	LastEndedAt   string
}

// TouchCaseOpen records that a detective opened this scenario (increments open_count).
func (s *Store) TouchCaseOpen(ctx context.Context, scenarioID, detective string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO case_folders (scenario_id, open_count, solved_count, last_detective, last_opened_at)
VALUES (?, 1, 0, ?, ?)
ON CONFLICT(scenario_id) DO UPDATE SET
	open_count = case_folders.open_count + 1,
	last_detective = excluded.last_detective,
	last_opened_at = excluded.last_opened_at
`, scenarioID, detective, now)
	return err
}

// RecordCaseFolderOutcome updates last outcome; increments solved_count when outcome is "solved".
func (s *Store) RecordCaseFolderOutcome(ctx context.Context, scenarioID, outcome string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `
UPDATE case_folders SET
	last_outcome = ?,
	last_ended_at = ?,
	solved_count = solved_count + CASE WHEN ? = 'solved' THEN 1 ELSE 0 END
WHERE scenario_id = ?
`, outcome, now, outcome, scenarioID)
	return err
}

// CaseFolderMap returns folder rows keyed by scenario_id (may be incomplete if never opened).
func (s *Store) CaseFolderMap(ctx context.Context) (map[string]CaseFolder, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT scenario_id, open_count, solved_count, last_detective, last_opened_at, last_outcome, last_ended_at
FROM case_folders`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]CaseFolder)
	for rows.Next() {
		var f CaseFolder
		if err := rows.Scan(&f.ScenarioID, &f.OpenCount, &f.SolvedCount,
			&f.LastDetective, &f.LastOpenedAt, &f.LastOutcome, &f.LastEndedAt); err != nil {
			return nil, err
		}
		out[f.ScenarioID] = f
	}
	return out, rows.Err()
}

// FolderOrZero returns one folder row or zero values if never opened.
func (s *Store) FolderOrZero(ctx context.Context, scenarioID string) (CaseFolder, error) {
	var f CaseFolder
	err := s.db.QueryRowContext(ctx, `
SELECT scenario_id, open_count, solved_count, last_detective, last_opened_at, last_outcome, last_ended_at
FROM case_folders WHERE scenario_id = ?`, scenarioID).Scan(
		&f.ScenarioID, &f.OpenCount, &f.SolvedCount,
		&f.LastDetective, &f.LastOpenedAt, &f.LastOutcome, &f.LastEndedAt)
	if err == sql.ErrNoRows {
		return CaseFolder{ScenarioID: scenarioID}, nil
	}
	if err != nil {
		return CaseFolder{}, err
	}
	return f, nil
}
