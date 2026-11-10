package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/vinaycharlie01/devsettings-mcp/internal/domain"
)

type Store struct {
	db *sql.DB
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "devsettings.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id TEXT PRIMARY KEY,
			editor TEXT NOT NULL,
			label TEXT NOT NULL DEFAULT '',
			timestamp TEXT NOT NULL,
			settings_data JSON,
			extensions_data JSON,
			checksum TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE TABLE IF NOT EXISTS migration_plans (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			target TEXT NOT NULL,
			items JSON NOT NULL,
			backup_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE TABLE IF NOT EXISTS docs_cache (
			url TEXT PRIMARY KEY,
			entry JSON NOT NULL,
			cached_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
	`)
	return err
}

func (s *Store) SaveBackup(ctx context.Context, backup *domain.Backup) error {
	settingsJSON, err := json.Marshal(backup.SettingsData)
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	extensionsJSON, err := json.Marshal(backup.ExtensionsData)
	if err != nil {
		return fmt.Errorf("marshaling extensions: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO backups (id, editor, label, timestamp, settings_data, extensions_data, checksum)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		backup.ID, string(backup.Editor), backup.Label,
		backup.Timestamp.Format(time.RFC3339),
		string(settingsJSON), string(extensionsJSON), backup.Checksum,
	)
	return err
}

func (s *Store) GetBackup(ctx context.Context, id string) (*domain.Backup, error) {
	var (
		b              domain.Backup
		editor         string
		ts             string
		settingsJSON   sql.NullString
		extensionsJSON sql.NullString
	)

	err := s.db.QueryRowContext(ctx,
		`SELECT id, editor, label, timestamp, settings_data, extensions_data, checksum
		 FROM backups WHERE id = ?`, id,
	).Scan(&b.ID, &editor, &b.Label, &ts, &settingsJSON, &extensionsJSON, &b.Checksum)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("backup not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("querying backup: %w", err)
	}

	b.Editor = domain.EditorID(editor)
	b.Timestamp, _ = time.Parse(time.RFC3339, ts)

	if settingsJSON.Valid && settingsJSON.String != "null" {
		json.Unmarshal([]byte(settingsJSON.String), &b.SettingsData)
	}
	if extensionsJSON.Valid && extensionsJSON.String != "null" {
		json.Unmarshal([]byte(extensionsJSON.String), &b.ExtensionsData)
	}

	return &b, nil
}

func (s *Store) ListBackups(ctx context.Context) ([]domain.Backup, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, editor, label, timestamp, checksum FROM backups ORDER BY timestamp DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing backups: %w", err)
	}
	defer rows.Close()

	var backups []domain.Backup
	for rows.Next() {
		var (
			b      domain.Backup
			editor string
			ts     string
		)
		if err := rows.Scan(&b.ID, &editor, &b.Label, &ts, &b.Checksum); err != nil {
			return nil, fmt.Errorf("scanning backup: %w", err)
		}
		b.Editor = domain.EditorID(editor)
		b.Timestamp, _ = time.Parse(time.RFC3339, ts)
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

func (s *Store) SaveMigrationPlan(ctx context.Context, plan *domain.MigrationPlan) error {
	itemsJSON, err := json.Marshal(plan.Items)
	if err != nil {
		return fmt.Errorf("marshaling items: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO migration_plans (id, source, target, items, backup_id)
		 VALUES (?, ?, ?, ?, ?)`,
		plan.ID, string(plan.Source), string(plan.Target),
		string(itemsJSON), plan.BackupID,
	)
	return err
}

func (s *Store) GetMigrationPlan(ctx context.Context, id string) (*domain.MigrationPlan, error) {
	var (
		p         domain.MigrationPlan
		source    string
		target    string
		itemsJSON string
	)

	err := s.db.QueryRowContext(ctx,
		`SELECT id, source, target, items, backup_id FROM migration_plans WHERE id = ?`, id,
	).Scan(&p.ID, &source, &target, &itemsJSON, &p.BackupID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("migration plan not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("querying migration plan: %w", err)
	}

	p.Source = domain.EditorID(source)
	p.Target = domain.EditorID(target)
	json.Unmarshal([]byte(itemsJSON), &p.Items)

	return &p, nil
}

func (s *Store) CacheDoc(ctx context.Context, entry *domain.DocEntry) error {
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling doc entry: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO docs_cache (url, entry) VALUES (?, ?)`,
		entry.URL, string(entryJSON),
	)
	return err
}

func (s *Store) GetCachedDoc(ctx context.Context, url string) (*domain.DocEntry, error) {
	var entryJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT entry FROM docs_cache WHERE url = ?`, url,
	).Scan(&entryJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("doc not cached: %s", url)
	}
	if err != nil {
		return nil, fmt.Errorf("querying doc cache: %w", err)
	}

	var entry domain.DocEntry
	if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
		return nil, fmt.Errorf("parsing cached doc: %w", err)
	}
	return &entry, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
