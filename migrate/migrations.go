package migrate

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
)

const (
	startUp   = "-- migrate:up"
	startDown = "-- migrate:down"
)

func readFile(dir fs.FS, name string) ([]byte, error) {
	r, err := dir.Open(name)
	defer r.Close()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	inUp := false
	up := make([]byte, 0)
	for scanner.Scan() {
		b := scanner.Bytes()
		trimmed := bytes.TrimSpace(b)
		if bytes.Equal(trimmed, []byte(startDown)) {
			break
		}
		if bytes.Equal(trimmed, []byte(startUp)) {
			inUp = true
		} else if inUp {
			up = append(up, b...)
		}
	}
	return up, nil
}

func initDb(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (\n  version VARCHAR(255) PRIMARY KEY\n)"); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}
	return nil
}

func getMigratedVersions(ctx context.Context, q Queryer) ([]string, error) {
	rows, err := q.QueryContext(ctx, "SELECT * FROM schema_migrations ORDER BY version ASC")
	if err != nil {
		return nil, fmt.Errorf("unable to query schema_migrations for existing versions: %w", err)
	}
	defer rows.Close()
	existing := make([]string, 0)
	for rows.Next() {
		existing = append(existing, "")
		rows.Scan(existing[len(existing)-1])
	}
	return existing, nil
}

func applyMigration(ctx context.Context, exec Execer, name string, b []byte) error {
	if _, err := exec.ExecContext(ctx, string(b)); err != nil {
		fmt.Printf("stmt: \n%s\n", string(b))
		return fmt.Errorf("executing migration: %w", err)
	}
	if _, err := exec.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
		return fmt.Errorf("storing version: %w", err)
	}
	return nil
}

type Execer interface {
	ExecContext(ctx context.Context, stmt string, args ...any) (sql.Result, error)
}

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Tx interface {
	Execer
	Queryer
	Rollback() error
	Commit() error
}

func Migrate(ctx context.Context, dir fs.FS, db *sql.DB) error {
	dirEntries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return fmt.Errorf("reading dir '.': %w", err)
	}
	if err := initDb(ctx, db); err != nil {
		return fmt.Errorf("initializing db: %w", err)
	}
	sort.SliceStable(dirEntries, func(i, j int) bool {
		return dirEntries[i].Name() < dirEntries[j].Name()
	})
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to start tx: %w", err)
	}
	defer tx.Rollback()
	alreadyMigratedVersions, err := getMigratedVersions(ctx, tx)
	for i, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		name := dirEntry.Name()
		if i < len(alreadyMigratedVersions) {
			if name != alreadyMigratedVersions[i] {
				return fmt.Errorf("already ran migration '%s' as version [%d] but got name %s", alreadyMigratedVersions[i], i, name)
			}
		}
		b, err := readFile(dir, dirEntry.Name())
		if err != nil {
			return fmt.Errorf("reading file (%s): %w", name, err)
		}
		if len(b) == 0 {
			continue
		}
		if err := applyMigration(ctx, tx, name, b); err != nil {
			return fmt.Errorf("applying migration (%s): %w", name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("comitting tx: %w", err)
	}
	return nil
}
