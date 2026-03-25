# Internal Package Layout

Planned package groups from `PLAN.md` and `docs/IMPLEMENTATION_GUARDRAILS.md`:

- `config`
- `database`
- `middleware`
- `models`
- `repositories`
- `queries`
- `policies`
- `services`
- `handlers`
- `dto`
- `importers`
- `jobs`
- `storage`
- `router`
- `swagger`
- `utils`

Boundary intent:
- `handlers` keep transport concerns at the edge
- `services` own orchestration and business logic
- `repositories` own persistence access
- `queries`, `policies`, and `jobs` stay explicit instead of collapsing into generic service code

Current phase-0 shell status:
- `config/` is materialized and owns runtime env loading/validation
- `router/` is materialized and wires the root plus `healthz` / `readyz` surface
- `handlers/` is materialized with the root shell handler, health handlers, and response helpers
- the remaining package groups are stubbed intentionally so later beads land in stable boundaries instead of inventing ad hoc directories
