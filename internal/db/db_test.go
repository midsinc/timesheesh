package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDBAppliesAllMigrationsToFreshDatabase(t *testing.T) {
	database, err := InitDB(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	assertCurrentSchema(t, database)

	migrations := migrationNames(t, database)
	expected := []string{
		"0001_initial_schema.sql",
		"0002_add_time_entry_billing_code.sql",
		"0003_add_employee_address.sql",
		"0004_add_project_payment_terms.sql",
	}
	if len(migrations) != len(expected) {
		t.Fatalf("expected %d migrations, got %d: %v", len(expected), len(migrations), migrations)
	}
	for i, name := range expected {
		if migrations[i] != name {
			t.Fatalf("expected migration %q at position %d, got %q", name, i, migrations[i])
		}
	}
}

func TestInitDBMigratesLegacyColumns(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy.db")

	legacyDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	legacySchema := []string{
		`CREATE TABLE employees (id INTEGER PRIMARY KEY AUTOINCREMENT, first_name TEXT NOT NULL, last_name TEXT NOT NULL, email TEXT UNIQUE NOT NULL);`,
		`CREATE TABLE projects (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, company_name TEXT NOT NULL, description TEXT);`,
		`CREATE TABLE assignments (id INTEGER PRIMARY KEY AUTOINCREMENT, employee_id INTEGER NOT NULL, project_id INTEGER NOT NULL, billable_rate REAL NOT NULL, pay_rate REAL NOT NULL);`,
		`CREATE TABLE time_entries (id INTEGER PRIMARY KEY AUTOINCREMENT, assignment_id INTEGER NOT NULL, date DATE NOT NULL, hours REAL NOT NULL, task_description TEXT);`,
		`CREATE TABLE billing_codes (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL, code TEXT NOT NULL, description TEXT);`,
	}
	for _, stmt := range legacySchema {
		if _, err := legacyDB.Exec(stmt); err != nil {
			t.Fatalf("create legacy schema: %v", err)
		}
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	migratedDB, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("init migrated db: %v", err)
	}
	t.Cleanup(func() {
		_ = migratedDB.Close()
	})

	assertCurrentSchema(t, migratedDB)

	migrations := migrationNames(t, migratedDB)
	expected := []string{
		"0001_initial_schema.sql",
		"0002_add_time_entry_billing_code.sql",
		"0003_add_employee_address.sql",
		"0004_add_project_payment_terms.sql",
	}
	for i, name := range expected {
		if migrations[i] != name {
			t.Fatalf("expected migration %q at position %d, got %q", name, i, migrations[i])
		}
	}
}

func TestInitDBBackfillsMigrationStateForCurrentDatabase(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "current.db")

	currentDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open current db: %v", err)
	}

	currentSchema := []string{
		`CREATE TABLE employees (id INTEGER PRIMARY KEY AUTOINCREMENT, first_name TEXT NOT NULL, last_name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, address TEXT NOT NULL DEFAULT '');`,
		`CREATE TABLE projects (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, company_name TEXT NOT NULL, description TEXT, default_payment_terms INTEGER NOT NULL DEFAULT 30);`,
		`CREATE TABLE billing_codes (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL, code TEXT NOT NULL, description TEXT, FOREIGN KEY(project_id) REFERENCES projects(id));`,
		`CREATE TABLE assignments (id INTEGER PRIMARY KEY AUTOINCREMENT, employee_id INTEGER NOT NULL, project_id INTEGER NOT NULL, billable_rate REAL NOT NULL, pay_rate REAL NOT NULL, FOREIGN KEY(employee_id) REFERENCES employees(id), FOREIGN KEY(project_id) REFERENCES projects(id));`,
		`CREATE TABLE time_entries (id INTEGER PRIMARY KEY AUTOINCREMENT, assignment_id INTEGER NOT NULL, billing_code_id INTEGER, date DATE NOT NULL, hours REAL NOT NULL, task_description TEXT, FOREIGN KEY(assignment_id) REFERENCES assignments(id), FOREIGN KEY(billing_code_id) REFERENCES billing_codes(id));`,
		`INSERT INTO employees (first_name, last_name, email, address) VALUES ('Willie', 'Pritchett', 'willie@mids-inc.com', '6048 Elderberry Lane');`,
		`INSERT INTO projects (name, company_name, description, default_payment_terms) VALUES ('Apollo', 'Acme', 'Existing data', 30);`,
		`INSERT INTO assignments (employee_id, project_id, billable_rate, pay_rate) VALUES (1, 1, 150, 90);`,
		`INSERT INTO time_entries (assignment_id, billing_code_id, date, hours, task_description) VALUES (1, NULL, '2026-04-17', 8, 'Existing entry');`,
	}
	for _, stmt := range currentSchema {
		if _, err := currentDB.Exec(stmt); err != nil {
			t.Fatalf("seed current schema: %v", err)
		}
	}
	if err := currentDB.Close(); err != nil {
		t.Fatalf("close current db: %v", err)
	}

	migratedDB, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("init current db: %v", err)
	}
	t.Cleanup(func() {
		_ = migratedDB.Close()
	})

	assertCurrentSchema(t, migratedDB)

	var timeEntries int
	if err := migratedDB.QueryRow(`SELECT COUNT(*) FROM time_entries`).Scan(&timeEntries); err != nil {
		t.Fatalf("count time entries: %v", err)
	}
	if timeEntries != 1 {
		t.Fatalf("expected 1 existing time entry, got %d", timeEntries)
	}

	migrations := migrationNames(t, migratedDB)
	if len(migrations) != 4 {
		t.Fatalf("expected 4 backfilled migrations, got %d: %v", len(migrations), migrations)
	}
}

func assertCurrentSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	checks := []struct {
		table  string
		column string
	}{
		{table: "time_entries", column: "billing_code_id"},
		{table: "employees", column: "address"},
		{table: "projects", column: "default_payment_terms"},
	}

	for _, check := range checks {
		hasColumn, err := tableHasColumn(db, check.table, check.column)
		if err != nil {
			t.Fatalf("check %s.%s: %v", check.table, check.column, err)
		}
		if !hasColumn {
			t.Fatalf("expected column %s.%s to exist", check.table, check.column)
		}
	}
}

func migrationNames(t *testing.T, db *sql.DB) []string {
	t.Helper()

	rows, err := db.Query(`SELECT name FROM schema_migrations ORDER BY name`)
	if err != nil {
		t.Fatalf("query schema migrations: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan migration name: %v", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate migration names: %v", err)
	}

	return names
}
