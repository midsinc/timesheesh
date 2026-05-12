---
name: timesheesh-agent-discovery
description: Use this skill when Codex needs to discover which Timesheesh projects, assignments, and billing codes are valid for an employee before logging time. It lists the exact CLI commands to call and the expected workflow.
---

# Timesheesh Agent Discovery

This skill is for strict CLI discovery, not natural-language parsing.

## Preferred commands

Employee-scoped:

```bash
./timesheesh --json agent projects --employee <employee_ref>
./timesheesh --json agent assignments --employee <employee_ref>
```

Global:

```bash
./timesheesh --json agent projects --all
./timesheesh --json agent assignments --all
```

Supporting:

```bash
./timesheesh --json billing-code list <project_ref>
./timesheesh --json time list <month_YYYY-MM>
```

## Operating procedure

1. Prefer employee email over name.
2. Run employee-scoped project discovery first.
3. Run employee-scoped assignment discovery second.
4. Read billing codes from the returned project or assignment data before proposing a time entry.
5. Use `--all` only for admin discovery or when the employee is still unknown.

## Failure handling

- If the employee cannot be resolved, stop and ask for a better reference.
- If the project does not appear in employee-scoped discovery, do not try to log time to it.
- If billing codes are empty, treat the project as having no configured billing codes unless a direct billing-code lookup proves otherwise.

## Reference

The implementation roadmap and planned agent validate/submit commands are in [docs/agent-cli-implementation-plan.md](../../../docs/agent-cli-implementation-plan.md).
