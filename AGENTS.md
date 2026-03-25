# AGENTS.md — learning-english (le)

> Guidelines for AI coding agents working in this codebase.

---

## RULE 0 - THE FUNDAMENTAL OVERRIDE PREROGATIVE

If I tell you to do something, even if it goes against what follows below, YOU MUST LISTEN TO ME. I AM IN CHARGE, NOT YOU.

---

## RULE NUMBER 1: NO FILE DELETION

**YOU ARE NEVER ALLOWED TO DELETE A FILE WITHOUT EXPRESS PERMISSION.** Even a new file that you yourself created, such as a test code file. You have a horrible track record of deleting critically important files or otherwise throwing away tons of expensive work. As a result, you have permanently lost any and all rights to determine that a file or folder should be deleted.

**YOU MUST ALWAYS ASK AND RECEIVE CLEAR, WRITTEN PERMISSION BEFORE EVER DELETING A FILE OR FOLDER OF ANY KIND.**

---

## Irreversible Git & Filesystem Actions — DO NOT EVER BREAK GLASS

1. **Absolutely forbidden commands:** `git reset --hard`, `git clean -fd`, `rm -rf`, or any command that can delete or overwrite code/data must never be run unless the user explicitly provides the exact command and states, in the same message, that they understand and want the irreversible consequences.
2. **No guessing:** If there is any uncertainty about what a command might delete or overwrite, stop immediately and ask the user for specific approval. "I think it's safe" is never acceptable.
3. **Safer alternatives first:** When cleanup or rollbacks are needed, request permission to use non-destructive options (`git status`, `git diff`, `git stash`, copying to backups) before ever considering a destructive command.
4. **Mandatory explicit plan:** Even after explicit user authorization, restate the command verbatim, list exactly what will be affected, and wait for a confirmation that your understanding is correct. Only then may you execute it—if anything remains ambiguous, refuse and escalate.
5. **Document the confirmation:** When running any approved destructive command, record (in the session notes / final response) the exact user text that authorized it, the command actually run, and the execution time. If that record is absent, the operation did not happen.

---

## Git Branch: ONLY Use `main`, NEVER `master`

**The default branch is `main`. The `master` branch exists only for legacy URL compatibility.**

- **All work happens on `main`** — commits, PRs, feature branches all merge to `main`
- **Never reference `master` in code or docs** — if you see `master` anywhere, it's a bug that needs fixing
- **The `master` branch must stay synchronized with `main`** — after pushing to `main`, also push to `master`:
  ```bash
  git push origin main:master
  ```

**If you see `master` referenced anywhere:**
1. Update it to `main`
2. Ensure `master` is synchronized: `git push origin main:master`

---

## Toolchain

### Current Repo State

- This repository is currently **docs-first**, but phase-0 execution and structural scaffold work have started.
- The main planning and operating docs right now are:
  - `PLAN.md`
  - `AGENTS.md`
  - `README.md`
  - `docs/IMPLEMENTATION_GUARDRAILS.md`
- Phase-0 foundation and runbook artifacts currently present in the repo workspace:
  - `docs/IMPLEMENTATION_GUARDRAILS.md`
  - `docs/DB_MIGRATIONS.md`
  - `docker-compose.yml`
  - `Makefile`
  - `.env.example`
- Structural `frontend/` and `backend/` directories now exist, including `backend/migrations/`.
- There is still **no** `package.json` or `go.mod` in this repo, and the app shells are still layout/readme placeholders rather than runnable implementations.
- Do **not** invent code paths, commands, or test results that depend on files that do not exist yet.

### Planned Toolchain

- Frontend: React + TypeScript + Vite
- Frontend routing/state/forms: React Router, TanStack Query, React Hook Form, Zustand or Context
- Content rendering: `react-markdown` + `remark-gfm`
- Backend: Go + Gorilla/mux or equivalent lightweight router + GORM + Swagger/OpenAPI
- Database / delivery: PostgreSQL + SQL migrations + Docker Compose + Makefile + CI
- Runtime reliability: object storage, background jobs, publications, audit logs, idempotency keys

### Key Dependencies

- Frontend runtime:
  - `react`
  - `react-dom`
  - `react-router-dom`
  - `@tanstack/react-query`
  - `react-hook-form`
  - `zustand`
  - `react-markdown`
  - `remark-gfm`
- Frontend testing:
  - `vitest`
  - `@testing-library/react`
  - `playwright`
- Backend runtime:
  - `gorilla/mux` or equivalent router
  - `gorm`
  - PostgreSQL driver
  - Swagger / OpenAPI tooling
- Backend testing:
  - Go `testing`
  - `httptest`
  - `testify`

---

## Code Editing Discipline

### No Script-Based Changes

**NEVER** run a script that processes/changes code files in this repo. Brittle regex-based transformations create far more problems than they solve.

- **Always make code changes manually**, even when there are many instances
- For many simple changes: use parallel subagents
- For subtle/complex changes: do them methodically yourself

### No File Proliferation

If you want to change something or add a feature, **revise existing code files in place**.

**NEVER** create variations like:
- `foo_new.ts`
- `foo_v2.ts`
- `foo_fixed.ts`
- `foo_copy.tsx`
- `PLAN-final.md`
- `PLAN_v2.md`
- `README-new.md`
- `AGENTS-copy.md`

New files are reserved for **genuinely new functionality** that makes zero sense to include in any existing file. The bar for creating new files is **incredibly high**.

---

## Backwards Compatibility

We do not care about backwards compatibility—we're in early development with no users. We want to do things the **RIGHT** way with **NO TECH DEBT**.

- Never create "compatibility shims"
- Never create wrapper functions for deprecated APIs
- Just fix the code directly

---

## Compiler Checks (CRITICAL)

**After any substantive code changes, you MUST verify no errors were introduced:**

```bash
# Current repo reality: docs + local DB foundation checks
rg -n "TODO|TBD|FIXME|XXX" AGENTS.md PLAN.md README.md docs/IMPLEMENTATION_GUARDRAILS.md docs/DB_MIGRATIONS.md .env.example
make compose-config
git diff --check

# Expected checks once runnable frontend/backend app shells exist
# frontend
cd frontend && npm run lint && npm run test && npm run build

# backend
cd backend && go test ./... && go vet ./...
```

If you see errors, **carefully understand and resolve each issue**. Read sufficient context to fix them the RIGHT way.

If package manifests or runnable app shells do not exist yet, do not pretend frontend/backend checks ran. State clearly which phase-0 or docs-level checks were actually available.

---

## Testing

### Testing Policy
This repo still does not have automated app tests wired up yet. The current phase-0 foundation can be validated with docs checks plus `make compose-config`. Once the scaffold exists, tests must cover:
- Happy path
- Edge cases (empty input, max values, boundary conditions)
- Error conditions

Planned testing split:
- Backend: unit tests, integration tests, API tests
- Frontend: unit/component tests
- E2E: Playwright
- Contract drift: OpenAPI-generated client sync + contract tests

### Unit Tests
```bash
# Current repo reality
rg -n "TODO|TBD|FIXME|XXX" AGENTS.md PLAN.md README.md docs/IMPLEMENTATION_GUARDRAILS.md docs/DB_MIGRATIONS.md .env.example
make compose-config
git diff --check

# Expected once runnable implementation exists
cd backend && go test ./...
cd frontend && npm run test
cd frontend && npm run test:e2e
```

---

## Third-Party Library Usage

If you aren't 100% sure how to use a third-party library, **SEARCH ONLINE** to find the latest documentation and current best practices.

---
## Learning-english - This project
This repository is the planning and operating handbook for a future English-learning web app. `PLAN.md` is the source of truth for product scope, architecture, schema, API contracts, reliability model, testing strategy, and roadmap. `AGENTS.md` exists to help agents work consistently with that plan.

### What It Does

- Builds an English-learning product with 4 main modules:
  - Auth
  - Vocabulary
  - Grammar
  - Reading
- Supports an admin content operations loop:
  - import
  - preview
  - validate
  - commit
  - publish
  - rollback
- Uses a reliability-first model:
  - hybrid auth
  - background jobs
  - idempotent writes
  - optimistic concurrency
  - audit logs
  - publication-aware learner reads

### Components

- Learner surface:
  - dashboard
  - vocabulary study sessions
  - grammar lessons and exercises
  - reading viewer
- Admin surface:
  - content import
  - validation report
  - diff preview
  - publish / unpublish / rollback
  - import and publication history
- Content lifecycle:
  - `content_revisions` for draft/source history
  - `publications` for live history
  - published projections for learner-facing reads
- Reliability/runtime components:
  - `queries`
  - `policies`
  - `jobs`
  - `idempotency_keys`
  - `content_audit_logs`
  - publication-scoped reading progress/highlights

### Architecture

- Frontend:
  - React SPA
  - learner routes separated from admin routes
  - server state separated from UI state
  - local persistence for pending learner/admin writes where appropriate
- Backend:
  - layered architecture: `handlers -> services -> repositories`
  - helper boundaries: `queries`, `policies`, `jobs`
- Database:
  - PostgreSQL
  - strong constraints for status/version/uniques
  - publication boundary for learner reads
- Content model:
  - source history lives in `content_revisions`
  - live history lives in `publications`
  - learner reads from active publication only
- Write safety:
  - `Idempotency-Key`
  - version / ETag preconditions
  - lease-based jobs
  - content audit logs
- Module-specific decisions:
  - Vocabulary uses deterministic `vocab_session_items`
  - Reading uses section-based published projections
  - Reading progress/highlights are publication-aware

### Project Structure

Current repo/worktree state:

```text
.
├── .env.example
├── AGENTS.md
├── Makefile
├── PLAN.md
├── README.md
├── backend/
│   ├── README.md
│   ├── cmd/
│   │   └── api/
│   │       └── README.md
│   ├── docs/
│   │   └── README.md
│   ├── internal/
│   │   └── LAYOUT.md
│   └── migrations/
│       ├── 000001_phase_0_baseline.down.sql
│       ├── 000001_phase_0_baseline.up.sql
│       └── README.md
├── docker-compose.yml
├── docs/
│   ├── DB_MIGRATIONS.md
│   └── IMPLEMENTATION_GUARDRAILS.md
└── frontend/
    ├── README.md
    ├── public/
    │   └── README.md
    └── src/
        ├── LAYOUT.md
        ├── features/
        │   └── README.md
        └── ...
```

Target structure once runnable implementation starts:

```text
learning-english/
├── frontend/
├── backend/
├── docker-compose.yml
├── Makefile
├── .env.example
├── README.md
├── PLAN.md
└── AGENTS.md
```

Important:
- The repo now has a structural scaffold, but most work is still documentation/spec alignment and phase-0 foundation refinement until real app entrypoints and manifests land.
- If the user asks to scaffold implementation, follow the target structure and contracts in `PLAN.md` instead of inventing an alternate layout.

### Key Design Decisions

- Optimize for a complete `content -> validate -> publish -> learner consume -> progress` loop before adding feature sprawl
- Keep learner APIs separate from admin content APIs from day one
- Learner must never read draft/source state; only active published projections
- Publish creates publication records instead of overwriting live content state
- Long-running content writes go through background jobs
- Vocabulary questions are snapshotted per session for deterministic resume/retry
- Reading progress carries publication context to prevent stale overwrite after republish/rollback
- Prefer reliability, operability, and clarity over extra product surfaces

### Agent Priorities In This Repo

- Treat `PLAN.md` as canonical. If `AGENTS.md` and `PLAN.md` disagree, align them instead of improvising.
- When updating architecture/schema/API docs, update all related sections so the document stays internally consistent.
- Be explicit about **current repo reality** versus **planned future structure**.
- Do not add speculative product surfaces unless they strengthen the core learning loop.
- If you later scaffold code, preserve these contracts:
  - learner/admin API split
  - revision + publication model
  - idempotent write paths
  - deterministic vocabulary sessions
  - publication-aware reading progress

---

## MCP Agent Mail — Multi-Agent Coordination

A mail-like layer that lets coding agents coordinate asynchronously via MCP tools and resources. Provides identities, inbox/outbox, searchable threads, and advisory file reservations with human-auditable artifacts in Git.

### Why It's Useful

- **Prevents conflicts:** Explicit file reservations (leases) for files/globs
- **Token-efficient:** Messages stored in per-project archive, not in context
- **Quick reads:** `resource://inbox/...`, `resource://thread/...`

### Same Repository Workflow

1. **Register identity:**
   ```
   ensure_project(project_key=<abs-path>)
   register_agent(project_key, program, model)
   ```

2. **Reserve files before editing:**
   ```
   file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true)
   ```

3. **Communicate with threads:**
   ```
   send_message(..., thread_id="FEAT-123")
   fetch_inbox(project_key, agent_name)
   acknowledge_message(project_key, agent_name, message_id)
   ```

4. **Quick reads:**
   ```
   resource://inbox/{Agent}?project=<abs-path>&limit=20
   resource://thread/{id}?project=<abs-path>&include_bodies=true
   ```

### Macros vs Granular Tools

- **Prefer macros for speed:** `macro_start_session`, `macro_prepare_thread`, `macro_file_reservation_cycle`, `macro_contact_handshake`
- **Use granular tools for control:** `register_agent`, `file_reservation_paths`, `send_message`, `fetch_inbox`, `acknowledge_message`

### Common Pitfalls

- `"from_agent not registered"`: Always `register_agent` in the correct `project_key` first
- `"FILE_RESERVATION_CONFLICT"`: Adjust patterns, wait for expiry, or use non-exclusive reservation
- **Auth errors:** If JWT+JWKS enabled, include bearer token with matching `kid`

---
## Beads (br) — Dependency-Aware Issue Tracking

Beads provides a lightweight, dependency-aware issue database and CLI (`br` - beads_rust) for selecting "ready work," setting priorities, and tracking status. It complements MCP Agent Mail's messaging and file reservations.

**Important:** `br` is non-invasive—it NEVER runs git commands automatically. You must manually commit changes after `br sync --flush-only`.

### Conventions

- **Single source of truth:** Beads for task status/priority/dependencies; Agent Mail for conversation and audit
- **Shared identifiers:** Use Beads issue ID (e.g., `br-123`) as Mail `thread_id` and prefix subjects with `[br-123]`
- **Reservations:** When starting a task, call `file_reservation_paths()` with the issue ID in `reason`

### Typical Agent Flow

1. **Pick ready work (Beads):**
   ```bash
   br ready --json  # Choose highest priority, no blockers
   ```

2. **Reserve edit surface (Mail):**
   ```
   file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true, reason="br-123")
   ```

3. **Announce start (Mail):**
   ```
   send_message(..., thread_id="br-123", subject="[br-123] Start: <title>", ack_required=true)
   ```

4. **Work and update:** Reply in-thread with progress

5. **Complete and release:**
   ```bash
   br close 123 --reason "Completed"
   br sync --flush-only  # Export to JSONL (no git operations)
   ```
   ```
   release_file_reservations(project_key, agent_name, paths=["src/**"])
   ```
   Final Mail reply: `[br-123] Completed` with summary

### Mapping Cheat Sheet

| Concept                   | Value                             |
| ------------------------- | --------------------------------- |
| Mail `thread_id`          | `br-###`                          |
| Mail subject              | `[br-###] ...`                    |
| File reservation `reason` | `br-###`                          |
| Commit messages           | Include `br-###` for traceability |

---
## bv — Graph-Aware Triage Engine

bv is a graph-aware triage engine for Beads projects (`.beads/beads.jsonl`). It computes PageRank, betweenness, critical path, cycles, HITS, eigenvector, and k-core metrics deterministically.

**Scope boundary:** bv handles *what to work on* (triage, priority, planning). For agent-to-agent coordination (messaging, work claiming, file reservations), use MCP Agent Mail.

**CRITICAL: Use ONLY `--robot-*` flags. Bare `bv` launches an interactive TUI that blocks your session.**

### The Workflow: Start With Triage

**`bv --robot-triage` is your single entry point.** It returns:
- `quick_ref`: at-a-glance counts + top 3 picks
- `recommendations`: ranked actionable items with scores, reasons, unblock info
- `quick_wins`: low-effort high-impact items
- `blockers_to_clear`: items that unblock the most downstream work
- `project_health`: status/type/priority distributions, graph metrics
- `commands`: copy-paste shell commands for next steps

```bash
bv --robot-triage        # THE MEGA-COMMAND: start here
bv --robot-next          # Minimal: just the single top pick + claim command
```

### Command Reference

**Planning:**
| Command            | Returns                                         |
| ------------------ | ----------------------------------------------- |
| `--robot-plan`     | Parallel execution tracks with `unblocks` lists |
| `--robot-priority` | Priority misalignment detection with confidence |

**Graph Analysis:**
| Command                                         | Returns                                                                                                           |
| ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `--robot-insights`                              | Full metrics: PageRank, betweenness, HITS, eigenvector, critical path, cycles, k-core, articulation points, slack |
| `--robot-label-health`                          | Per-label health: `health_level`, `velocity_score`, `staleness`, `blocked_count`                                  |
| `--robot-label-flow`                            | Cross-label dependency: `flow_matrix`, `dependencies`, `bottleneck_labels`                                        |
| `--robot-label-attention [--attention-limit=N]` | Attention-ranked labels                                                                                           |

**History & Change Tracking:**
| Command                           | Returns                                               |
| --------------------------------- | ----------------------------------------------------- |
| `--robot-history`                 | Bead-to-commit correlations                           |
| `--robot-diff --diff-since <ref>` | Changes since ref: new/closed/modified issues, cycles |

**Other:**
| Command                                             | Returns                                              |
| --------------------------------------------------- | ---------------------------------------------------- |
| `--robot-burndown <sprint>`                         | Sprint burndown, scope changes, at-risk items        |
| `--robot-forecast <id\|all>`                        | ETA predictions with dependency-aware scheduling     |
| `--robot-alerts`                                    | Stale issues, blocking cascades, priority mismatches |
| `--robot-suggest`                                   | Hygiene: duplicates, missing deps, label suggestions |
| `--robot-graph [--graph-format=json\|dot\|mermaid]` | Dependency graph export                              |
| `--export-graph <file.html>`                        | Interactive HTML visualization                       |

### Scoping & Filtering

```bash
bv --robot-plan --label backend              # Scope to label's subgraph
bv --robot-insights --as-of HEAD~30          # Historical point-in-time
bv --recipe actionable --robot-plan          # Pre-filter: ready to work
bv --recipe high-impact --robot-triage       # Pre-filter: top PageRank
bv --robot-triage --robot-triage-by-track    # Group by parallel work streams
bv --robot-triage --robot-triage-by-label    # Group by domain
```

### Understanding Robot Output

**All robot JSON includes:**
- `data_hash` — Fingerprint of source beads.jsonl
- `status` — Per-metric state: `computed|approx|timeout|skipped` + elapsed ms
- `as_of` / `as_of_commit` — Present when using `--as-of`

**Two-phase analysis:**
- **Phase 1 (instant):** degree, topo sort, density
- **Phase 2 (async, 500ms timeout):** PageRank, betweenness, HITS, eigenvector, cycles

<!-- bv-agent-instructions-v1 -->

---

## Beads (br) Workflow Integration

**Note:** `br` is non-invasive and never executes git commands. After `br sync --flush-only`, you must manually run `git add .beads/` and `git commit`.

This project uses [beads_rust](https://github.com/Dicklesworthstone/beads_rust) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View issues (launches TUI - avoid in automated sessions)
bv

# CLI commands for agents (use these instead)
br ready --json         # Show issues ready to work (no blockers)
br list --status open   # All open issues
br show <id>            # Full issue details with dependencies
br create --title="..." --type=task --priority=2
br update <id> --status in_progress
br close <id> --reason "Completed"
br close <id1> <id2>    # Close multiple issues at once
br sync --flush-only    # Export JSONL only; commit .beads/ separately
git add .beads/
git commit -m "sync beads"
```

### Workflow Pattern

1. **Start**: Run `br ready --json` to find actionable work
2. **Claim**: Use `br update <id> --status in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `br close <id>`
5. **Sync**: Always run `br sync --flush-only`, then `git add .beads/` and `git commit`

### Key Concepts

- **Dependencies**: Issues can block other issues. `br ready --json` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers, not words)
- **Types**: task, bug, feature, epic, question, docs
- **Blocking**: `br dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
git commit -m "..."     # Commit code
br sync --flush-only    # Export beads changes to JSONL
git add .beads/
git commit -m "sync beads"
git push                # Push to remote
```

### Best Practices

- Check `br ready --json` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `br create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always `br sync --flush-only`, then commit `.beads/`, before ending session

<!-- end-bv-agent-instructions -->

---

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Sync beads** - `br sync --flush-only` to export to JSONL
5. **Hand off** - Provide context for next session

---

Note for Codex/GPT-5.2:

You constantly bother me and stop working with concerned questions that look similar to this:

```
Unexpected changes (need guidance)

- Working tree still shows edits I did not make in Cargo.toml, Cargo.lock, src/main.rs, src/patterns.rs. Please advise whether to keep/commit/revert these before any further work. I did not touch them.

Next steps (pick one)

1. Decide how to handle the unrelated modified files above so we can resume cleanly.
```

NEVER EVER DO THAT AGAIN. The answer is literally ALWAYS the same: those are changes created by the potentially dozen of other agents working on the project at the same time. This is not only a common occurrence, it happens multiple times PER MINUTE. The way to deal with it is simple: you NEVER, under ANY CIRCUMSTANCE, stash, revert, overwrite, or otherwise disturb in ANY way the work of other agents. Just treat those changes identically to changes that you yourself made. Just fool yourself into thinking YOU made the changes and simply don't recall it for some reason.

---

## Note on Built-in TODO Functionality

Also, if I ask you to explicitly use your built-in TODO functionality, don't complain about this and say you need to use beads. You can use built-in TODOs if I tell you specifically to do so. Always comply with such orders.
