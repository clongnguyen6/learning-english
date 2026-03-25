# Frontend Source Layout

Planned source groups from `PLAN.md` and `docs/IMPLEMENTATION_GUARDRAILS.md`:

- `app`
- `routes`
- `pages`
- `components`
- `features`
- `services`
- `hooks`
- `store`
- `types`
- `utils`

Route and state intent:
- learner and admin routes stay visibly separate early
- server state belongs in TanStack Query
- UI and session state belongs in Zustand or Context
