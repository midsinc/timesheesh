# CLI Reference

`timesheesh` exposes the same core records as the web app and supports machine-readable output with `--json`.

The CLI also supports human-friendly references so you do not have to memorize internal IDs for normal workflows.

## Global Flags

- `--db <path>`: use a specific SQLite database file.
- `--json`: emit structured JSON instead of human-formatted text.

By default, the CLI now looks for the existing `timesheesh.db` in the current working tree and next to the executable before creating a new file, which avoids accidentally booting against an empty database from a different directory.

## Commands

## Reference Rules

- `employee_ref`: employee ID, full name, or email.
- `project_ref`: project ID or project name.
- `assignment_ref`: assignment ID or `employee_ref::project_ref`.
- Use `timesheesh employee list`, `timesheesh project list`, and `timesheesh assignment list` to browse valid values before creating related records.

### Employees

- `timesheesh employee add <first> <last> <email> <address>`
- `timesheesh employee update <employee_ref> <first> <last> <email> <address>`
- `timesheesh employee list`

Example:

```bash
./timesheesh employee add Ada Lovelace ada@example.com "123 Example St, Indianapolis, IN"
./timesheesh employee list --json
```

### Projects

- `timesheesh project add <name> <company> <description> <invoice_due_days>`
- `timesheesh project update <project_ref> <name> <company> <description> <invoice_due_days>`
- `timesheesh project list`

Example:

```bash
./timesheesh project add Apollo Acme "Migration work" 30
./timesheesh project list --json
```

### Assignments

- `timesheesh assignment add <employee_ref> <project_ref> <bill_rate> <pay_rate>`
- `timesheesh assignment update <assignment_ref> <employee_ref> <project_ref> <bill_rate> <pay_rate>`
- `timesheesh assignment list`

Example:

```bash
./timesheesh assignment add "Ada Lovelace" Apollo 150 90
./timesheesh assignment update "Ada Lovelace::Apollo" "Ada Lovelace" Apollo 160 95
./timesheesh assignment list --json
```

### Billing Codes

- `timesheesh billing-code add <project_ref> <code> <description>`
- `timesheesh billing-code list <project_ref>`

Example:

```bash
./timesheesh billing-code add Apollo DEV-01 Development
./timesheesh billing-code list Apollo --json
```

### Time Entries

- `timesheesh time add <assignment_ref> <date:YYYY-MM-DD> <hours> <task_description> [billing_code_ref]`
- `timesheesh time log <project_ref> <hours> <task_description>`
- `timesheesh time update <id> <assignment_ref> <date:YYYY-MM-DD> <hours> <task_description> [billing_code_ref]`
- `timesheesh time list [month_YYYY-MM]`

Example:

```bash
./timesheesh time add "Ada Lovelace::Apollo" 2026-04-13 8 "Feature build" 1
export TIMESHEESH_EMPLOYEE="ada@example.com"
./timesheesh time log Apollo 8 "Feature build" --billing-code DEV-01
./timesheesh time list 2026-04 --json
```

Quick logging notes:

- `time log` defaults the date to today.
- `time log` resolves the employee from `--employee`, `TIMESHEESH_EMPLOYEE`, or the only employee in the database.
- `billing_code_ref` can be either a numeric ID or the billing code itself, such as `DEV-01`.

### Invoices

- `timesheesh invoice <project_ref> <employee_ref> <year> <month> <output>`
- `timesheesh invoice --description-mode task ...`
- `timesheesh invoice --description-mode project ...`

Example:

```bash
./timesheesh invoice --description-mode project Apollo "Ada Lovelace" 2026 4 invoice_april.pdf
```

### Web Server

- `timesheesh server`
- `timesheesh migrate`

Example:

```bash
./timesheesh server
./timesheesh --db /path/to/timesheesh.db migrate
```

## Agent Notes

- Prefer `--json` whenever an AI agent needs to read command output reliably.
- Human-readable references are accepted anywhere the command uses `employee_ref`, `project_ref`, or `assignment_ref`.
- IDs returned by create/list commands still map directly to the same records used by the web app.
- Invoice description mode accepts exactly `task` or `project`.
- Schema migrations live in `internal/db/migrations/*.sql` and can be applied explicitly with `timesheesh migrate`.
