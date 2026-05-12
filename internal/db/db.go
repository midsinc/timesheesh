package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	name string
	sql  string
}

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if err := ensureMigrationTable(db); err != nil {
		return err
	}
	if err := bootstrapLegacyMigrationState(db); err != nil {
		return err
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(db, migration.name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.name, err)
		}
	}

	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(migrationFiles, "migrations/"+entry.Name())
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, migration{
			name: entry.Name(),
			sql:  string(content),
		})
	}

	sort.Slice(migrations, func(i int, j int) bool {
		return migrations[i].name < migrations[j].name
	})

	return migrations, nil
}

func applyMigration(db *sql.DB, migration migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, statement := range splitSQLStatements(migration.sql) {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (name, applied_at) VALUES (?, CURRENT_TIMESTAMP)`,
		migration.name,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func splitSQLStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		statement := strings.TrimSpace(part)
		if statement == "" {
			continue
		}
		statements = append(statements, statement)
	}
	return statements
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL
	)`)
	return err
}

func migrationApplied(db *sql.DB, name string) (bool, error) {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func bootstrapLegacyMigrationState(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	existingTables, err := existingAppTables(db)
	if err != nil {
		return err
	}
	if len(existingTables) == 0 {
		return nil
	}

	applied := make([]string, 0, 4)
	hasBaseTables := existingTables["employees"] &&
		existingTables["projects"] &&
		existingTables["assignments"] &&
		existingTables["time_entries"] &&
		existingTables["billing_codes"]
	if hasBaseTables {
		applied = append(applied, "0001_initial_schema.sql")
	}

	if existingTables["time_entries"] {
		hasBillingCodeID, err := tableHasColumn(db, "time_entries", "billing_code_id")
		if err != nil {
			return err
		}
		if hasBillingCodeID {
			applied = append(applied, "0002_add_time_entry_billing_code.sql")
		}
	}

	if existingTables["employees"] {
		hasAddress, err := tableHasColumn(db, "employees", "address")
		if err != nil {
			return err
		}
		if hasAddress {
			applied = append(applied, "0003_add_employee_address.sql")
		}
	}

	if existingTables["projects"] {
		hasPaymentTerms, err := tableHasColumn(db, "projects", "default_payment_terms")
		if err != nil {
			return err
		}
		if hasPaymentTerms {
			applied = append(applied, "0004_add_project_payment_terms.sql")
		}
	}

	if len(applied) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, name := range applied {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO schema_migrations (name, applied_at) VALUES (?, CURRENT_TIMESTAMP)`,
			name,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func existingAppTables(db *sql.DB) (map[string]bool, error) {
	tableNames := []string{"employees", "projects", "billing_codes", "assignments", "time_entries"}
	existing := make(map[string]bool, len(tableNames))
	for _, name := range tableNames {
		present, err := tableExists(db, name)
		if err != nil {
			return nil, err
		}
		existing[name] = present
	}
	return existing, nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		tableName,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func tableHasColumn(db *sql.DB, tableName string, columnName string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}
