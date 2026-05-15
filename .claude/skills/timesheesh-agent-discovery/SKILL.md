---
name: timesheesh-agent-discovery
description: Use this skill when an AI agent needs to discover which Timesheesh projects, assignments, and billing codes are available before attempting to log time. It teaches the agent which CLI commands to run and when to use employee-scoped versus global discovery.
---

# Timesheesh Agent Discovery

Use this skill before any AI-driven time submission.

## Commands

Employee-scoped discovery:

```bash
./timesheesh --json agent projects --employee <employee_ref>
./timesheesh --json agent assignments --employee <employee_ref>
```

Global discovery:

```bash
./timesheesh --json agent projects --all
./timesheesh --json agent assignments --all
```

Optional supporting commands:

```bash
./timesheesh --json billing-code list <project_ref>
./timesheesh --json time list <month_YYYY-MM>
```

## Workflow

1. Resolve the human first.
   Use email when available. It is more stable than display name.

2. Ask for employee-scoped projects.
   This answers "what can this person log against?"

3. Ask for employee-scoped assignments.
   This returns assignment IDs, rates, and project-linked billing codes.

4. Only use `--all` when the agent is doing admin-style discovery or the employee is not known yet.

5. Prefer `--json` on every call.
   The agent should read structured output, not scrape terminal text.

## Guardrails

- Do not invent project names or billing codes.
- Do not submit time until the employee-scoped discovery calls succeed.
- If multiple employees have similar names, rerun using email.
- If no assignment exists for the employee and project, stop and report that clearly.

## Current command contract

See [docs/agent-cli-implementation-plan.md](../../../docs/agent-cli-implementation-plan.md) for the broader roadmap, including planned validate and submit commands.
