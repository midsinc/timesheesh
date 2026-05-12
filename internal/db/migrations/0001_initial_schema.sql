CREATE TABLE IF NOT EXISTS employees (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	company_name TEXT NOT NULL,
	description TEXT
);

CREATE TABLE IF NOT EXISTS billing_codes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	code TEXT NOT NULL,
	description TEXT,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS assignments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	employee_id INTEGER NOT NULL,
	project_id INTEGER NOT NULL,
	billable_rate REAL NOT NULL,
	pay_rate REAL NOT NULL,
	FOREIGN KEY(employee_id) REFERENCES employees(id),
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS time_entries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	assignment_id INTEGER NOT NULL,
	date DATE NOT NULL,
	hours REAL NOT NULL,
	task_description TEXT,
	FOREIGN KEY(assignment_id) REFERENCES assignments(id)
);
