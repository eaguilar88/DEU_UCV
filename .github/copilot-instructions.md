# Repository instructions for AI coding assistants

This file orients any AI coding assistant (GitHub Copilot, ChatGPT, Claude, etc.) working in this
repository. It covers project layout, backend architecture, conventions to follow, and how to
contribute a change.

## Project overview

DEU UCV is a monorepo for Universidad Central de Venezuela's Dirección de Extensión Universitaria
system. Only one codebase lives directly in this repo:

- `backend/` — the Go API. **This is the only code you should edit here.**

Everything else under the following paths is a **git submodule** — an independent repository owned
by a different team:

- `diplomados/` (`ecp_ucv`) — Next.js portal for course/diploma management
- `grupos/` (`gsu_ucv`) — Next.js portal for extension groups
- `landing/` (`deu-app`) — public landing site (Rails)
- `espacios-universitarios/` — university spaces reservation system

**Do not edit files inside submodule directories from this repo.** They may appear as empty
directories in a fresh checkout (submodules not initialized) — that is expected, not a bug. If you
need their code, ask for `git submodule update --init --recursive` to be run, or work in that
submodule's own repository.

## Commands

Run from `backend/` (the Makefile targets in the root `Makefile` already `cd backend` for you).

```bash
# Dev loop
make start-backend                                          # db (+ backend per committed compose file)
cd backend && go run cmd/deu/main.go                         # run API directly (needs .env, see Configuration)

# Tests
cd backend && go test -count=1 -short -cover ./...           # fast unit tests
cd backend && go test ./... -run TestName -v                 # single test
cd backend && go test ./internal/courses/... -v              # single package

# Full quality gate — run before pushing
cd backend && go vet ./... && go fmt ./... && golangci-lint run && go mod tidy && \
  go test -count=1 -race -short -cover ./...

# Mocks (mockery v3.7.0, config backend/.mockery.yaml)
make generate-mocks                                          # regenerate mocks after changing an interface

# Docker Compose stacks
make start-backend / make start-prod / make start-prod-landing / make start-espacios
make stop-deu / stop-backend / stop-prod / stop-landing / stop-espacios
```

Note: `docker-compose.yml` may have local uncommitted changes stripping the `backend` service down
to just `db` — check `git status` / `git diff docker-compose.yml` before assuming its shape.
The Makefile's `start-db`/`stop-db`/`refresh`/`refresh-prod` targets reference
`docker-compose.dev.yml`, which does not currently exist — those targets will fail until that file
is added.

## Configuration

Copy `.env.example` to `.env` before running anything. `backend/internal/config/config.go`
(parsed via `caarlos0/env` struct tags) is the source of truth for every env var the backend reads.
In `FLAVOR=prod`, Backblaze B2 credentials come from Docker secrets
(`/run/secrets/b2_key_id`, `/run/secrets/b2_application_key`), not env vars.

## Backend architecture

Stack: Echo (HTTP), Squirrel (SQL builder), Zap (structured logging), golang-jwt, AWS SDK v2
against Backblaze B2 (S3-compatible) for file storage.

Entry point: `backend/cmd/deu/main.go`. It wires every module's repository → service → handler and
registers all routes — **read it first** to trace how a request reaches a feature.

### Module layout

Each business capability is a top-level package under `backend/internal/`:

`courses`, `course_periods`, `course_requests`, `providers`, `provider_requests`, `groups`,
`group_requests`, `group_resource_requests`, `activities`, `course_cycle_close_requests`, `users`,
`auth`, `files`, `email`, `storage`.

Shared/infra packages (not business modules): `config`, `entities`, `httperrors`, `jwt`,
`postgres_repository`, `security`, `utils`.

A typical business module contains:

- `endpoints.go` (+ `endpoints_test.go`) — Echo handlers, HTTP binding/validation
- `service.go` (+ `service_test.go`) — business logic; defines its own narrow `Repository`
  (and sometimes `StorageClient`) interface scoped to exactly what it needs
- `request.go` / `response.go` — HTTP DTOs
- `decoders.go` / `encoders.go` — Request→Entity and Entity→Response mapper functions
- `errors.go` — domain-specific sentinel errors
- `mocks/` — mockery-generated testify mocks for that module's interfaces

Every module defines its own narrow `Repository` interface (Interface Segregation) even though a
single `PostgresRepository` (in `backend/internal/postgres_repository/`) implements all of them —
one Postgres repo satisfies every module's interface via method sets, keyed off Squirrel query
builders and hand-written SQL in `postgres_repository/queries/`, with row→entity mapping in
`postgres_repository/mappers/`.

### The three-struct layering

```
Request/Response (internal/*/request.go, response.go)   — HTTP layer, json/validate tags
        ↓ decoders.go / encoders.go
Entities (internal/entities/*.go)                        — domain layer, no struct tags, business rules
        ↓ postgres_repository/mappers
Models (internal/postgres_repository/models/*.go)        — persistence layer, mirrors DB tables
```

**Never let `db` tags leak into `entities`, or `json`/`validate` tags leak past
`request.go`/`response.go`.** That separation is deliberate — it's what lets a module's service
logic be tested without HTTP or SQL involved. See `docs/arquitectura.md` for the full writeup with
a worked example (`CreateUser`).

### Auth & admin routes

JWT auth (`internal/jwt`) issues/validates tokens; `jwt.JWTMiddleware` gates protected routes. In
`main.go`, most modules register a public route group (unauthenticated GETs) and a protected group
(writes) using the same middleware slice. A separate `/admin` group aggregates each module's
`RegisterAdminEndpoints`/`RegisterXAdminEndpoints` callback (see the `RegisterAdminEndpoints` type
in `main.go`). When adding an admin-only endpoint, wire it through that pattern rather than adding
a new top-level route group.

### Known rough edges — don't "fix" these as a drive-by

Per `docs/anti_patterns.md` (documented against `course_periods`, but broadly present):

- Error handling and logging verbosity are **inconsistent across modules** — some handlers return
  raw `echo.Err*`, others wrap with context. Match the surrounding module's existing style in the
  file you're editing; don't unify this repo-wide as a side effect of an unrelated change.
- Some goroutines are fired without a timeout or cancellation path (missing
  `context.WithTimeout` + `select`). Only fix this where it's directly relevant to the task at hand.
- Logging sometimes lacks enough context (request/user IDs) to be useful for debugging.

If your task is specifically about improving one of these, say so explicitly rather than folding it
into an unrelated PR.

## Contribution workflow

- CI (`.github/workflows/pr-analysis.yml`) runs on `feature/*`, `bugfix/*`, and `development`
  branches: `golangci-lint run`, `go test ./... -race -coverprofile=...`, then enforces a **5% total
  coverage floor** (`.github/workflows/gh-scripts/check-coverage.sh`).
- `.github/workflows/deploy.yml` deploys to the VPS on published release or manual dispatch
  (choose `all`/`backend`/`landing`/`diplomados`/`grupos`).
- `.github/workflows/release.yml` cuts a GitHub release from `main` when the `VERSION` file changes.
- PRs use the template at `.github/PULL_REQUEST_TEMPLATE/pull_request_template.md`: **Description**,
  **What Changed**, **Testing Instructions** — fill in all three.
- Run the full quality gate (see Commands above) before opening a PR; it mirrors what CI checks.

## Docs worth reading before large changes

- `docs/arquitectura.md` — the layering/DTO/mapper conventions above, in detail with code samples
- `docs/architecture.md` — C4-style container diagram of how backend, frontends, DB, B2, and email
  fit together
- `docs/design_patterns.md` / `docs/anti_patterns.md` — patterns and known issues, documented
  against `course_periods` but broadly illustrative of the whole backend's style
- `docs/postman/` — a Postman collection + local environment for exercising the API manually

## Example prompts

Use prompts like these with Copilot Chat or ChatGPT once the repo/file context above is loaded
(paste this file's contents, or the relevant module's files, into the chat first):

> **Add a field end-to-end:** "Add an optional `notes` text field to the `groups` module, editable
> only on update. Follow the three-struct layering in `backend/internal/groups/`: add it to the
> entity in `internal/entities`, the DB model and mapper in `internal/postgres_repository`, the
> `UpdateGroupRequest` struct with a `validate` tag, the response DTO, and the encoder/decoder
> functions. Don't let `db` tags reach the entity or `json`/`validate` tags reach past
> `request.go`/`response.go`. Add a migration and a service-layer test using the existing mocks in
> `internal/groups/mocks/`."

> **Add an admin endpoint:** "Add a new admin-only endpoint to the `activities` module that lets an
> admin force-close an activity, following the `RegisterAdminEndpoints` pattern used in
> `backend/cmd/deu/main.go` and the existing admin endpoints in `internal/activities/endpoints.go`.
> Reuse the module's existing `Repository` interface where possible instead of adding new repository
> methods, and match the error-handling style already used in that file rather than introducing a
> new convention."
