# learning-english

Planning-first repository for an English-learning web app.

## Current Status

This repo is still phase-0, but it is no longer just a docs-only stub.

Checked-in foundation artifacts now include:
- `docs/IMPLEMENTATION_GUARDRAILS.md` — implementation-facing architecture guardrails
- `docker-compose.yml` — local PostgreSQL foundation
- `Makefile` — DB and migration workflow commands
- `.env.example` — shared config contract for local, CI, and future deployment
- `backend/` — backend scaffold roots and migration home
- `frontend/` — frontend scaffold roots and layout notes
- `backend/migrations/` — versioned SQL migration directory
- `docs/DB_MIGRATIONS.md` — migration workflow runbook

The application shells inside `backend/` and `frontend/` are still intentionally incomplete. The current scaffold locks the repo shape without pretending the runtime is already implemented.

Primary files:
- `PLAN.md` — full product, architecture, schema, API, testing, and roadmap specification
- `AGENTS.md` — working rules and project context for AI coding agents
- `README.md` — short entry point for humans and agents
- `docs/IMPLEMENTATION_GUARDRAILS.md` — short anti-drift guide for the first scaffold

## Planned Product Scope

- Auth: username/password login with safe session management
- Vocabulary: flashcards, MCQ, write mode, progress tracking, deterministic study sessions
- Grammar: Markdown-driven lessons, topic-based exercises, keyword highlighting
- Reading: bilingual section-based reading, progress by section, publication-aware rendering
- Admin content ops: import, preview, validate, commit, publish, rollback

## Planned Architecture

- Frontend: React + TypeScript + Vite
- Backend: Go + layered architecture (`handlers -> services -> repositories`) with `queries`, `policies`, `jobs`
- Database: PostgreSQL
- Content model:
  - `content_revisions` for draft/source history
  - `publications` for live history
  - learner reads from active published projections only

## Config Contract

The root `.env.example` is the versioned config contract for:
- shared local/dev settings
- database and Compose wiring
- planned backend runtime variables
- planned frontend runtime variables

Rules:
- secrets stay out of git; `.env.example` only defines names, safe defaults, and empty required slots
- backend startup should fail fast on missing critical vars instead of silently defaulting
- frontend runtime vars must use the `VITE_` prefix and must never contain secrets
- CI, staging, and production should reuse the same variable names rather than inventing parallel naming

## Start Here

1. Read `PLAN.md` for the canonical project plan.
2. Read `docs/IMPLEMENTATION_GUARDRAILS.md` before scaffolding code or routes.
3. Read `AGENTS.md` for project-specific agent instructions.
4. Review `.env.example` before adding backend/frontend runtime code so env names stay consistent across local, CI, and deployment.
5. Review `docs/DB_MIGRATIONS.md` and use `make compose-config` / `make db-up` / `make migrate-up` when you need the local PostgreSQL foundation.
6. Treat `backend/` and `frontend/` as the canonical scaffold roots for later shell work.
7. Keep these files aligned when architecture, schema, API decisions change.

## Migration Workflow

```bash
make compose-config
make db-up
make migrate-version
make migrate-up
make migrate-down STEPS=1
make migrate-create NAME=add_users
```

See `docs/DB_MIGRATIONS.md` for naming conventions and the future-CI path.

## Full Local Stack

The Compose stack now includes:
- `postgres` for the local database foundation
- `backend` running `go run ./cmd/api`
- `frontend` running the Vite development server on port `5173`

Typical local boot flow:

```bash
make compose-config
make stack-up
make migrate-up
make stack-logs
```

Default local URLs:
- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`
- backend API base: `http://localhost:8080/api/v1`

`make stack-down` stops the full Compose stack when you are done.

## Automation

Baseline CI now lives in `.github/workflows/ci.yml` and runs on pushes to `main`, pull requests, and manual dispatches.

Current jobs:
- repo hygiene: `make compose-config` and `git diff --check`
- frontend build: `cd frontend && npm ci && npm run build`
- backend checks: `cd backend && go test ./... && go vet ./...`
- migration sanity: `make migrate-up`, `make migrate-version`, `make migrate-down STEPS=1`, then `make migrate-up` again

Use the same commands locally before pushing so CI stays a confirmation step instead of a separate undocumented path.
