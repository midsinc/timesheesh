package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS employees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			address TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			company_name TEXT NOT NULL,
			description TEXT,
			default_payment_terms INTEGER NOT NULL DEFAULT 30
		);`,
		`CREATE TABLE IF NOT EXISTS billing_codes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			code TEXT NOT NULL,
			description TEXT,
			FOREIGN KEY(project_id) REFERENCES projects(id)
		);`,
		`CREATE TABLE IF NOT EXISTS assignments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,
			billable_rate REAL NOT NULL,
			pay_rate REAL NOT NULL,
			FOREIGN KEY(employee_id) REFERENCES employees(id),
			FOREIGN KEY(project_id) REFERENCES projects(id)
		);`,
		`CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			assignment_id INTEGER NOT NULL,
			billing_code_id INTEGER,
			date DATE NOT NULL,
			hours REAL NOT NULL,
			task_description TEXT,
			FOREIGN KEY(assignment_id) REFERENCES assignments(id),
			FOREIGN KEY(billing_code_id) REFERENCES billing_codes(id)
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	if err := migrateTables(db); err != nil {
		return err
	}

	return nil
}

func migrateTables(db *sql.DB) error {
	hasBillingCodeID, err := tableHasColumn(db, "time_entries", "billing_code_id")
	if err != nil {
		return err
	}
	if !hasBillingCodeID {
		if _, err := db.Exec(`ALTER TABLE time_entries ADD COLUMN billing_code_id INTEGER REFERENCES billing_codes(id)`); err != nil {
			return err
		}
	}

	hasEmployeeAddress, err := tableHasColumn(db, "employees", "address")
	if err != nil {
		return err
	}
	if !hasEmployeeAddress {
		if _, err := db.Exec(`ALTER TABLE employees ADD COLUMN address TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}

	hasPaymentTerms, err := tableHasColumn(db, "projects", "default_payment_terms")
	if err != nil {
		return err
	}
	if !hasPaymentTerms {
		if _, err := db.Exec(`ALTER TABLE projects ADD COLUMN default_payment_terms INTEGER NOT NULL DEFAULT 30`); err != nil {
			return err
		}
	}

	return nil
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
