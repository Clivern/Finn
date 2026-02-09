# Finn agent knowledge

This document is meant to help AI agents (and new contributors) understand how the Finn repository is organized and how to make safe, correct changes.

## What Finn is

Finn is a platform to run, schedule, and automate `Ansible` and `Terraform` workflows from a web UI or an HTTP API (OpenAPI/Swagger). It includes:

- A Go management API server (HTTP, sessions/auth, admin APIs)
- A Go runner server (gRPC) that actually executes work
- Background task execution (management worker)
- A database for management state (driver-configurable: SQLite/Postgres)
- A separate local bolt DB for runner identity/state

## High-level architecture

1. `core/server.go` builds the HTTP router (using `chi`) and registers all `api/**/*.go` action handlers.
2. `cli/server` runs the management server:
   - Loads config via `viper`
   - Sets up logging
   - Starts the HTTP server and the management background worker (`mgmt.NewWorker`)
3. The runner is started separately via `cli/runner`:
   - Opens the runner bolt DB
   - Starts a gRPC server (`proto/runner/v1/runner.proto`)
   - Runs the runner worker (`deps.Worker.Run`)
4. `finn.go` embeds the built web assets (`web/dist/*`) and embeds locales (`locale/locales`) into the binary.

## Where to find things

- HTTP routing / entrypoints
  - `core/server.go` (the route table and middleware wiring)
  - `cli/server.go` (cobra command: `finn server`)
  - `finn.go` (main entry; embeds static assets)
- Management logic (HTTP handlers)
  - `api/` (notably `api/workspace/*` for workspace CRUD and workflow-related APIs)
- Runner logic (gRPC service + execution)
  - `pkg/runner/` (runner core; bolt storage; worker)
  - `pkg/runner/tasks/*` (task implementations: ansible/terraform/clone/etc.)
  - `proto/runner/v1/` (gRPC contracts and generated code)
- Background task system (management worker)
  - `pkg/mgmt/` and `core/server.go` (worker setup)
- Database/migrations
  - `db/` (DB access + repositories)
  - `migration/` (migration definitions + manager)
  - `cli/migrate.go` (cobra command: `finn migrate {up,down,status}`)
- Web frontend
  - `web/` (Vue app built with Vite)
  - `web/src/locales/` (UI strings)
  - `locale/` (embedded server-side locales)
- API spec / clients
  - `openapi/definition.yaml` (OpenAPI 3.0 spec)
  - `sdk/v1/` (Go SDK for Finn API)

## API overview

- Base path is `/api/v1` as documented in `openapi/definition.yaml`.
- Authentication/authorization is applied via middleware in `core/server.go`.
- Endpoint registration pattern:
  - `Public` routes (health/readiness, setup/login/logout, register/invite token, etc.)
  - Authenticated routes with role gating (admin / regular)

If you add or modify endpoints:

1. Update `openapi/definition.yaml` (`paths` + relevant schemas in `components/schemas`).
2. Add/modify the Go handler under `api/`.
3. Register the handler in `core/server.go` under the correct router group and role middleware.

## Runner overview (gRPC)

- Runner server is gRPC-based (separate process from the HTTP server).
- `cli/runner` starts:
  - `core.InitRunner()` (opens runner bolt DB + ensures identity)
  - `core.RunRunner()` (starts the gRPC server with auth/log interceptors)

To add a new runner task:

- Implement it under `pkg/runner/tasks/*`.
- Ensure it is registered/available via the runner task wiring (see `pkg/runner/tasks` imports in `core/runner.go` and related runner dependency setup).

## Config

The server uses YAML configs loaded into `viper`, with environment-variable interpolation (see `mgmt_config.*.yml` / `runner_config.*.yml`).

Management server config (`mgmt_config.*.yml`):

- `app.port`, `app.hostname`
- TLS: `app.tls.status`, `app.tls.crt_path`, `app.tls.key_path`
- Logging: `app.log.level`, `app.log.output`, `app.log.format`
- Database: `app.database.driver`, `host`, `port`, `username`, `password`, `name`, `max_open_conns`, `max_idle_conns`, `conn_max_lifetime`, `datasource`
- AI config: `app.ai.enabled`, `provider`, `model`, `api_key`, `base_url`, and provider-specific fields (ex: `project_id`, `location`)

Runner server config (`runner_config.*.yml`):

- `app.port`, TLS, logging
- `app.database.driver` (bolt), `app.database.path`

Migrations:

- Run `finn migrate up` with the management config file.

## Development workflow

### Common commands

- Run the full management server (includes migrations step in the Makefile target):
  - `make run`
- Run the management HTTP server only (if you want to control migrations separately):
  - `go run finn.go server -c mgmt_config.prod.yml`
- Run the runner only:
  - `go run finn.go runner -c runner_config.prod.yml`
- Build the web frontend assets:
  - `make web`
- Build protobuf/gRPC code:
  - `make proto`
- Run tests:
  - `make test_short`, `make test`
- Lint:
  - `make lint`

### Code style/conventions to follow

- Go doc comments: use simple, human doc-comment sentences for exported symbols (see `.cursor/rules/go-doc-comments.mdc`).
- Zerolog formatting: when chaining fields, split into multiple lines (see `.cursor/rules/zerolog-log-format.mdc`).

## Logging and audit context

When working in `api/**/*.go`:

- Controller/audit logging should include contextual IDs to keep logs traceable (workspace/user/entity IDs as available).
- Make sure success and error logs carry the same contextual identifiers for easier debugging.

## Safety checklist for changes

- Do you need to update the OpenAPI spec in `openapi/definition.yaml`?
- Did you register the handler in `core/server.go` under the correct role guard?
- Did you add/adjust tests (unit/integration) for the new behavior?
- Are you storing secrets/tokens only where intended (DB/encrypted fields), and not logging raw secrets?

