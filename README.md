# Watchdog — Uptime / Site Monitoring Service

## Table of Contents
- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Setup](#setup)
- [Available Commands](#available-commands)
- [Folder Structure](#folder-structure)
- [Contributing](#contributing)
- [License](#license)

## Project Overview
Watchdog is an event-driven uptime monitoring service. It periodically checks configured URLs, writes time-series metrics to TimescaleDB/Postgres, and notifies owners when status changes occur.

## Architecture

### Overview
The orchestrator is the main entry point and coordinator for the monitoring system. It registers listeners on the event bus, creates the `Supervisor`, and spins up worker groups for each configured monitoring interval. The runtime flow is:

- `Orchestrator` initializes the system (logger, event bus, supervisor) and registers event listeners.
- For each configured monitoring frequency the orchestrator creates a `ParentWorker`.
- `ParentWorker` in turn spawns multiple `ChildWorker`s which perform the actual periodic HTTP checks.
- `ChildWorker` sends each check result to the `Supervisor` for evaluation.
- The `Supervisor` determines whether a check represents a success or failure and publishes a corresponding event (e.g. `ping.successful`, `ping.unsuccessful`) on the event bus.
- Registered listeners react to those events: they persist time-series measurements, update URL metadata, and trigger notifications (emails) on state transitions.

This is done to ensure separation of concerns: workers perform checks, the supervisor makes state decisions, and listeners handle persistence and notifications.

### Components
- Orchestrator: bootstraps the system, registers listeners, creates the `Supervisor`, and starts `ParentWorker` instances for each time interval.
- ParentWorker: groups `ChildWorker`s for a given monitoring interval and forwards tick signals.
- ChildWorker: performs the periodic HTTP/monitoring checks and forwards the raw check result to the `Supervisor`.
- Supervisor: receives check results, applies decision logic (e.g., thresholds, debounce), and emits domain events (`ping.successful` / `ping.unsuccessful`) to the event bus.
- Event Bus: a lightweight pub/sub mechanism for decoupling event producers (Supervisor) from consumers (listeners).
- Listeners: subscribe to event topics and react (persist metrics into the time-series table, update `Url` status, send notification emails).
- Database Repositories: encapsulate SQL operations for `Url` metadata and `UrlStatus` (time-series) storage.

### Event Flow
1. `Orchestrator` boots: sets up logger, event bus, registers listeners, and instantiates the `Supervisor`.
2. For every monitoring frequency the orchestrator creates and starts a `ParentWorker`.
3. Each `ParentWorker` starts `ChildWorker`s. `ChildWorker`s perform scheduled HTTP checks based on the interval.
4. After a check completes, the `ChildWorker` sends the result to the `Supervisor`.
5. The `Supervisor` evaluates the result and decides whether the check is a success or failure, possibly applying retry/debounce logic.
6. The `Supervisor` publishes a topic event (e.g., `ping.successful` / `ping.unsuccessful`) on the event bus including metadata (url id, url, status, timing).
7. Event listeners receive the event and perform side effects:
   - Persist a time-series data point for historical metrics.
   - Update the canonical `Url` status if there is a state change.
   - Send notification emails when a state change warrants an alert.

### Data Model
- `Url` (metadata): id, url, contact email, current status, monitoring configuration (frequency, thresholds).
- `UrlStatus` (time-series hypertable in Timescale): timestamped health/latency/response metrics.
- `enums`: status values (e.g., `Healthy`, `UnHealthy`).

## Tech Stack
- Go (modules)
- Postgres / TimescaleDB (via `github.com/jackc/pgx/v5/pgxpool`)
- TimescaleDB extension for time-series hypertables
- Goose (https://github.com/pressly/goose) for SQL migrations
- Redis (used for internal pools/caching as configured)
- SMTP for sending notification emails (example uses Mailtrap)
- Structured logging (`log/slog`)

## Setup

### Prerequisites
- Go 1.XX installed
- Postgres / TimescaleDB instance
- (Optional) SMTP credentials for email notifications
- `git`, `make` (if `Makefile` is present)

### Environment Variables
Copy `.env.example` to `.env` and edit values for your environment. The project expects the following variables (from `.env.example`):

- `APP_ENV` — application environment (e.g., `dev`, `prod`).

Redis configuration:
- `REDIS_HOST` — Redis host:port (default `127.0.0.1:6379`).
- `REDIS_PASS` — Redis password (if any).
- `REDIS_DB` — Redis database index used by the app.

Worker / pool configuration:
- `MAXIMUM_CHILD_WORKERS` — maximum number of child workers per parent.
- `MAXIMUM_WORK_POOL_SIZE` — total work pool size (concurrency limit).
- `HTTP_REQUEST_TIMEOUT` — timeout (seconds) for HTTP requests performed by child workers.
- `SUPERVISOR_POOL_FLUSH_TIMEOUT` — flush timeout for supervisor batching (seconds).
- `SUPERVISOR_POOL_FLUSH_BATCHSIZE` — batch size for supervisor flush operations.

Database configuration (used by goose and the app):
- `DB_USER` — Postgres username.
- `DB_PASSWORD` — Postgres password.
- `DB_HOST` — Postgres host.
- `DB_PORT` — Postgres port.
- `DB_DATABASE` — Postgres database name.

Goose migration configuration (see `.env.example`):
- `GOOSE_DRIVER` — goose DB driver (e.g., `postgres`).
- `GOOSE_DBSTRING` — connection string used by goose (often interpolates `DB_*` vars).
- `GOOSE_MIGRATION_DIR` — path to migrations (default `migrations`).

Mail configuration:
- `MAIL_FROM_ADDRESS` — sender address used for notification emails.
- `MAIL_HOST` — SMTP host (example: Mailtrap sandbox).
- `MAIL_PORT` — SMTP port.
- `MAIL_USERNAME` — SMTP username.
- `MAIL_PASSWORD` — SMTP password.

### Database (TimescaleDB/Postgres)
1. Create the database.
2. Install TimescaleDB extension in the database:
    - `CREATE EXTENSION IF NOT EXISTS timescaledb;`
3. Run SQL migrations found in `migrations` to create tables and hypertables.

### Run Locally
1. Clone the repo:
    - `git clone <repo>`
2. Configure environment variables.
3. Run:
    - `go run ./cmd/...` or `go build` and run the binary.
4. Start any worker or scheduler processes required (see `cmd`).

## Available Commands

This project exposes a small CLI with the following commands (located in `cmd/commands`). Each command is registered by the application and can be invoked using the binary or via `go run`.

Note: run the CLI from the repository root. Example: `go run ./cmd/... <command> [args|flags]` (PowerShell examples are shown below).

1) guard (alias: g)
- Purpose: Start the watchdog monitoring process (orchestrator).
- Usage: starts Redis and Postgres connections, instantiates the `Orchestrator`, registers listeners, creates the `Supervisor`, and begins scheduled monitoring intervals.
- Example:

```powershell
# Start the monitoring service (development)
go run ./cmd/... guard
# or using the alias
go run ./cmd/... g
```

2) add (alias: a)
- Purpose: Add a new URL to be monitored.
- Arguments (positional):
  - `url` (string) — The URL to monitor (required).
  - `http_method` (string) — HTTP method to use: `get`, `post`, `patch`, `put`, `delete` (default: `get`).
  - `frequency` (string) — Monitoring frequency. Options: `ten_seconds`, `thirty_seconds`, `one_minute`, `five_minutes`, `thirty_minutes`, `one_hour`, `twelve_hours`, `twenty_four_hours` (default: `five_minutes`).
  - `contact_email` (string) — Email address to notify on state changes (required).
- Behavior: persists the new URL in the database and refreshes the Redis interval list used by the workers.
- Example:

```powershell
# Add a site to be monitored (positional args)
go run ./cmd/... add https://example.com get five_minutes owner@example.com
# Using alias
go run ./cmd/... a https://example.com get five_minutes owner@example.com
```

3) remove (alias: rm)
- Purpose: Remove a monitored URL by ID.
- Arguments (positional):
  - `id` (int) — The ID of the URL to remove (required).
- Behavior: deletes the URL from the database and refreshes the Redis list for that frequency.
- Example:

```powershell
# Remove site with ID 42
go run ./cmd/... remove 42
# Using alias
go run ./cmd/... rm 42
```

4) list (alias: ls)
- Purpose: List the URLs currently being monitored.
- Flags (named):
  - `--page` (int) — Page number (default `1`).
  - `--per_page` (int) — Results per page (default `20`).
  - `--http_method` (string) — Filter by HTTP method (`get`, `post`, ...).
  - `--frequency` (string) — Filter by frequency (see `add` for options).
  - `--status` (string) — Filter by site health status (e.g., `healthy`, `unhealthy`).
- Example:

```powershell
# List first page
go run ./cmd/... list --page=1 --per_page=20
# Filter by frequency and method
go run ./cmd/... list --frequency=five_minutes --http_method=get
# Using alias
go run ./cmd/... ls --status=healthy
```

5) analysis (alias: a)
- Purpose: Run an ad-hoc analysis for a given monitored URL.
- Arguments (positional):
  - `id` (int) — The ID of the URL to analyze (required).
- Behavior: loads the URL from the database and prints a brief status/analysis (currently a simple status output).
- Example:

```powershell
# Run analysis for site id 42
go run ./cmd/... analysis 42
# Using alias
go run ./cmd/... a 42
```

Notes & caveats
- Aliases: be aware that `add` and `analysis` both declare the alias `a` in the code; depending on your CLI invocation this may cause ambiguity — prefer calling the full command name to avoid conflicts.
- Positional vs named arguments: commands in this project use positional arguments (declared in the command definitions) and flags for optional filters or pagination. Make sure to supply arguments in the order shown when using positional syntax.

## Folder Structure

A high-level overview of the top-level folders in this repository and their responsibilities (no file-level details):

- `cmd/` — Application entry points and CLI wiring; contains the executable commands and bootstrapping logic used to run the service and CLI tools.
- `core/` — Shared core utilities and small subsystems used across the app (helpers, common business logic, and integration glue).
- `env/` — Environment loading and configuration helpers (centralizes reading environment variables and simple typed accessors).
- `database/` — Data access layer and repository code that interacts with Postgres/TimescaleDB; abstracts queries and persistence logic.
- `enums/` — Centralized enumerations and parsing utilities used across the codebase to represent domain constants.
- `events/` — Domain event types and the in-process event bus; defines the event contracts used between components.
- `events/listeners/` — Event listener implementations that react to published events (keeps side-effects decoupled from producers).
- `logger/` — Logging configuration and helpers for structured/logging setup used by the rest of the application.
- `orchestrator/` — High-level orchestration logic that wires workers, the supervisor, and the event bus to run monitoring pipelines.
- `supervisor/` — Decision-making component that evaluates raw check results and translates them into domain events.
- `worker/` — Worker implementations: parent/child worker groups responsible for scheduling and performing HTTP checks.
- `migrations/` — SQL migration files for initializing and evolving the database schema and Timescale hypertables.

## Contributing
- Fork, create a branch per feature/fix, open PR with clear description and tests.

## License
This project is licensed under the MIT License — see the `LICENSE` file in the repository root for the full text.
