# Time Sheesh Development Guidelines

## Build & Deploy Workflow
Every time an action, feature, or fix is implemented, the following sequence MUST be performed:
1. **Build**: Run `go build -o timesheesh ./cmd/timesheesh/main.go` to ensure the project compiles.
2. **Test**: Verify the change via the CLI or Web UI (and run any applicable unit tests).
3. **Push**: Commit the changes and push to the GitHub remote (`origin main`).

## Project Structure
- `cmd/timesheesh/`: CLI entry point and command definitions.
- `internal/db/`: SQLite database initialization and migrations.
- `internal/models/`: Core data entities.
- `internal/services/`: Business logic (Time tracking and PDF Invoicing).
- `internal/web/`: Gin-based web server and API handlers.
- `static/`: Frontend HTML/JS/CSS.

## Tech Stack
- Language: Go
- Database: SQLite
- CLI: spf13/cobra
- Web: gin-gonic/gin
- PDF: go-pdf/fpdf
