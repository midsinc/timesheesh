# Agent CLI Implementation Plan

## Goal

Make `timesheesh` predictable for AI agents that need to submit time on behalf of a human.

The agent should:
- discover which projects and assignments are available
- see valid billing codes for those projects
- validate a proposed entry before writing
- submit time with a machine-readable response

The CLI should not parse natural language. The AI agent should convert notes into structured fields, then call strict `timesheesh` commands.

## Command Roadmap

### Discovery

- `timesheesh agent projects --employee <employee_ref> --json`
- `timesheesh agent assignments --employee <employee_ref> --json`
- `timesheesh agent projects --all --json`
- `timesheesh agent assignments --all --json`

These commands let an agent answer "what can Willie log against right now?" before attempting validation or submit.

### Validation

- `timesheesh agent validate --employee <employee_ref> --project <project_ref> --hours <n> --task <text> [--date YYYY-MM-DD] [--billing-code <ref>] --json`

This should resolve references, verify the assignment exists, confirm the billing code belongs to the project, and return structured warnings or errors without writing.

### Submit

- `timesheesh agent submit --employee <employee_ref> --project <project_ref> --hours <n> --task <text> [--date YYYY-MM-DD] [--billing-code <ref>] [--idempotency-key <key>] --json`
- `timesheesh agent submit-batch --input <file.json> --json`

## JSON Contracts

### `agent projects --employee`

```json
{
  "employee": {
    "id": 1,
    "first_name": "Willie",
    "last_name": "Pritchett",
    "email": "willie@mids-inc.com"
  },
  "projects": [
    {
      "id": 1,
      "name": "MMIS TAS Lead Sr Analyst",
      "company_name": "MIDS",
      "description": "",
      "assignment_id": 1,
      "billing_codes": []
    }
  ]
}
```

### `agent assignments --employee`

```json
{
  "employee": {
    "id": 1,
    "email": "willie@mids-inc.com"
  },
  "assignments": [
    {
      "id": 1,
      "employee_id": 1,
      "project_id": 1,
      "project_name": "MMIS TAS Lead Sr Analyst",
      "billable_rate": 150,
      "pay_rate": 90,
      "billing_codes": []
    }
  ]
}
```

### Error shape

```json
{
  "status": "rejected",
  "error_code": "employee_not_found",
  "message": "Employee reference \"foo\" not found"
}
```

## Implementation Order

1. Add agent discovery commands:
   - `agent projects`
   - `agent assignments`
2. Add service-layer query helpers to return employee-scoped project and assignment views.
3. Add tests for employee-scoped and global discovery.
4. Add `agent validate`.
5. Add `agent submit`.
6. Add idempotent submit support with a migration-backed receipt table.
7. Add `submit-batch`.

## Schema Work

The discovery commands do not need schema changes.

Idempotent submit will need a migration like:

`internal/db/migrations/0005_add_submission_receipts.sql`

Proposed table:

```sql
CREATE TABLE submission_receipts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  idempotency_key TEXT NOT NULL UNIQUE,
  time_entry_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(time_entry_id) REFERENCES time_entries(id)
);
```

## Current Scope In This Branch

This branch implements the discovery layer:
- `timesheesh agent projects`
- `timesheesh agent assignments`

That gives an AI agent enough information to discover valid employee assignments before validation and submit are added.
