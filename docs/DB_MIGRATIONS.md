# Database Migrations

This repo now has a versioned SQL migration workflow even though the full backend application shell has not landed yet.

The workflow is intentionally simple:

- PostgreSQL runs through `docker compose`.
- The migration CLI runs from the official `migrate/migrate` Docker image on the same Compose network.
- Versioned SQL files live in `backend/migrations/`.

## Current Reality

- The repository is still docs-first.
- The database foundation exists now so schema work can start before the rest of the backend/frontend scaffold is complete.
- The baseline migration pair is intentionally a no-op so phase-0 can verify the tooling without inventing module schema too early.

## Migration File Convention

Create migrations as sequential SQL pairs:

```text
backend/migrations/
├── 000001_phase_0_baseline.up.sql
├── 000001_phase_0_baseline.down.sql
├── 000002_add_users.up.sql
└── 000002_add_users.down.sql
```

Use names that describe the schema change clearly. Later schema beads should extend this directory instead of patching the database manually.

## Commands

```bash
make compose-config
make db-up
make migrate-version
make migrate-create NAME=add_users
make migrate-up
make migrate-down STEPS=1
make migrate-force VERSION=1
```

Command notes:

- `make migrate-create NAME=...` creates the next sequential up/down SQL pair.
- `make migrate-up` applies all pending migrations.
- `make migrate-down STEPS=1` rolls back the most recent migration step.
- `make migrate-version` shows the current schema version and dirty state.
- `make migrate-force VERSION=...` exists only for recovering a dirty migration state after understanding what failed.

## CI Usage

`.github/workflows/ci.yml` now reuses the same non-interactive targets and performs a small migration round-trip:

```bash
make compose-config
make migrate-up
make migrate-version
make migrate-down STEPS=1
make migrate-up
```

The workflow tears the Compose stack down afterward with `make db-down`.

That keeps local development and CI on one migration path instead of separate ad hoc scripts.
