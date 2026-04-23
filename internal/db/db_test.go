package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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

	hasColumn, err := tableHasColumn(migratedDB, "time_entries", "billing_code_id")
	if err != nil {
		t.Fatalf("check migrated column: %v", err)
	}
	if !hasColumn {
		t.Fatal("expected billing_code_id column to be added during migration")
	}

	hasAddress, err := tableHasColumn(migratedDB, "employees", "address")
	if err != nil {
		t.Fatalf("check employee address column: %v", err)
	}
	if !hasAddress {
		t.Fatal("expected address column to be added during migration")
	}

	hasPaymentTerms, err := tableHasColumn(migratedDB, "projects", "default_payment_terms")
	if err != nil {
		t.Fatalf("check project payment terms column: %v", err)
	}
	if !hasPaymentTerms {
		t.Fatal("expected default_payment_terms column to be added during migration")
	}
}
