# Implementation Guardrails

This document turns the architecture decisions in `PLAN.md` into build-time guardrails.

It is intentionally short. The point is to make the non-negotiable boundaries obvious before the first scaffold or feature code starts to spread.

## Current Reality

- `PLAN.md` is still the canonical product and architecture specification.
- The repository is still in phase 0 and does not have a real frontend or backend implementation yet.
- Any new scaffold or package layout should follow these guardrails instead of inventing a different architecture.

## Non-Negotiable Boundaries

### 1. Keep learner and admin surfaces separate

- Learner APIs live under learner routes.
- Admin content operations live under admin routes from day one.
- Learner pages, state, and API clients must not depend on admin-only shapes.
- Admin workflows can be richer and more operational, but they must not leak draft state into learner reads.

Practical rule:
- Use separate route groups such as `/api/v1/...` for learner/private routes and `/api/v1/admin/...` for content operations.
- Mirror that separation in frontend route trees and feature folders.

### 2. Preserve the revision/publication split

- Draft or source history lives in `content_revisions`.
- Live history lives in `publications`.
- Publish creates publication records; it does not overwrite learner-visible content in place.
- Learner reads resolve from active published projections only.

Practical rule:
- If a design would let a learner read draft rows directly, it is the wrong design.
- If a rollback would mutate history instead of selecting an earlier publication, it is the wrong design.

### 3. Treat write safety as architecture, not polish

Important writes must be designed for retries, races, and partial failure.

- Use `Idempotency-Key` for retry-sensitive writes.
- Use `version` or `ETag` preconditions for mutable resources.
- Record meaningful audit trails for content operations.
- Long-running content mutations go through lease-based jobs, not fire-and-forget goroutines.

Practical rule:
- If a write path can produce duplicate side effects under retry, it is incomplete.
- If a content operation cannot explain who changed what and why, it is incomplete.

### 4. Keep vocabulary sessions deterministic

- Vocabulary sessions must snapshot issued items into `vocab_session_items`.
- Resume or retry must not regenerate a different question sequence from mutable progress state.

Practical rule:
- Session generation is a one-time planning step for a session.
- Question order and prompts are persisted session data, not re-derived every request.

### 5. Keep reading progress publication-aware

- Reading progress writes must carry publication context and concurrency state.
- Delayed client replay must not overwrite progress from a newer publication.

Practical rule:
- `publication_id` and `base_version` are part of the correctness model.
- A simple last-write-wins progress endpoint is not acceptable.

### 6. Keep auth split and session model intact

- Access tokens are short-lived bearer tokens.
- Refresh tokens travel via secure cookie transport, not long-lived browser storage.
- Session rotation, revocation, and reuse detection are backed by database state.

Practical rule:
- Do not store long-lived auth state in `localStorage` or `sessionStorage`.
- Do not treat refresh-token reuse detection as optional future hardening.

## Intended Backend Shape

Use the backend scaffold to make the boundaries visible:

```text
backend/
├── cmd/api/
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   ├── models/
│   ├── repositories/
│   ├── queries/
│   ├── policies/
│   ├── services/
│   ├── handlers/
│   ├── dto/
│   ├── importers/
│   ├── jobs/
│   ├── storage/
│   ├── router/
│   └── swagger/
└── migrations/
```

Boundary rules:

- `handlers`: transport and DTO mapping only
- `services`: business logic and orchestration
- `repositories`: persistence access
- `queries`: read-optimized projections and dashboard/history reads
- `policies`: publish, authorization, and validation gates
- `jobs`: leased background execution

Do not collapse `queries`, `policies`, or `jobs` back into generic service code just because the first implementation is small.

## Intended Frontend Shape

Use the frontend scaffold to keep learner and admin concerns visibly separate:

```text
frontend/src/
├── app/
├── routes/
├── pages/
├── components/
├── features/
│   ├── auth/
│   ├── vocabulary/
│   ├── grammar/
│   ├── reading/
│   └── admin/
├── services/
├── hooks/
├── store/
├── types/
└── utils/
```

Boundary rules:

- TanStack Query owns server state.
- Zustand or Context owns UI/session state.
- Learner routes and admin routes are separated early.
- Pending write recovery is scoped to high-value flows, not bolted onto everything.

## Phase-0 Build Order

When the scaffold starts landing, bias toward this order:

1. Architecture guardrails
2. Repo scaffold
3. Config contract
4. Backend shell
5. Shared API/runtime contract
6. Local PostgreSQL foundation
7. Migration workflow
8. Frontend shell
9. Full local app stack
10. CI and documentation alignment

The point is to unlock backend and schema work early without waiting for the whole app shell to exist.

## Review Checklist

Before landing a scaffold or feature, check:

- Does this preserve learner/admin separation?
- Does this preserve revision/publication separation?
- Does this keep learner reads on published projections only?
- Does this respect idempotency or concurrency requirements for important writes?
- Does this preserve deterministic vocabulary sessions?
- Does this preserve publication-aware reading progress?
- Does this keep transport concerns in handlers and business logic in services?

If the answer to any item is "no" or "not sure," stop and align with `PLAN.md` before continuing.
