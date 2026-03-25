# Environment Contract

This document explains the versioned configuration contract defined in the root `.env.example` for local development, CI, and future deployment.

Current reality:
- the repo now has a local PostgreSQL foundation, migration workflow, and phase-0 scaffold
- the backend application shell now exists and loads its runtime config from the root env contract
- the frontend application shell still does not exist yet
- this contract exists so later shell work loads config consistently instead of inventing names ad hoc
- `.env.example` is the source of truth for variable names; this document explains how to use them

## Principles

- Fail fast: missing required config must stop startup clearly.
- Separate secret and public config: browser-visible values use `VITE_` prefixes; secrets never do.
- Prefer same-site local defaults first: cross-site auth adds CORS and CSRF complexity and should be opt-in.
- Keep one naming contract across local, CI, staging, and production even if the injection source changes.

## Variable Groups

### Database and local Compose foundation

| Variable | Scope | Required | Notes |
| --- | --- | --- | --- |
| `POSTGRES_DB` | local Compose | yes | Local database name. |
| `POSTGRES_USER` | local Compose | yes | Local database user. |
| `POSTGRES_PASSWORD` | local Compose, secret | yes | Local database password. |
| `POSTGRES_PORT` | local Compose | yes | Host port for PostgreSQL. |
| `POSTGRES_INITDB_ARGS` | local Compose | no | Init-time PostgreSQL flags. |
| `DATABASE_URL` | backend host tools, CI | yes | Host-side DSN for migrations and backend processes running on the machine. |
| `DATABASE_URL_INTERNAL` | Compose services | yes | Service-to-service DSN for containers on the Compose network. |

### Shared runtime

| Variable | Scope | Required | Notes |
| --- | --- | --- | --- |
| `APP_ENV` | backend, frontend build, CI | yes | `development`, `test`, `staging`, or `production`. |
| `APP_PUBLIC_URL` | docs, CI, deployment | yes | Canonical frontend origin, for example `http://localhost:5173`. |
| `API_BASE_URL` | backend, docs, CI | yes | Canonical API origin, for example `http://localhost:8080`. |

### Backend HTTP and auth/session runtime

| Variable | Scope | Required | Notes |
| --- | --- | --- | --- |
| `API_ADDR` | backend | yes | Bind address for the API server, for example `:8080`. |
| `LOG_LEVEL` | backend, CI | yes | Runtime logging verbosity. |
| `REQUEST_ID_HEADER` | backend | yes | Request-ID header name used by middleware and downstream logs. |
| `CORS_ALLOWED_ORIGINS` | backend | yes | Explicit comma-separated allowlist; local default should stay narrow. |
| `JWT_ISSUER` | backend | yes | Token issuer string. |
| `JWT_AUDIENCE` | backend | yes | Intended token audience. |
| `JWT_HS256_SECRET` | backend, secret | yes | Signing secret for access tokens in early environments. |
| `ACCESS_TOKEN_TTL` | backend | yes | Short-lived bearer token TTL. |
| `REFRESH_TOKEN_TTL` | backend | yes | Longer-lived refresh session TTL. |
| `REFRESH_COOKIE_NAME` | backend | yes | Refresh-session cookie name. |
| `REFRESH_COOKIE_DOMAIN` | backend | no | Leave empty for localhost-style local development. |
| `REFRESH_COOKIE_SECURE` | backend | yes | `false` locally, `true` outside local HTTP development. |
| `REFRESH_COOKIE_SAME_SITE` | backend | yes | Expected values: `lax`, `strict`, or `none`. |

### Frontend public runtime

| Variable | Scope | Required | Notes |
| --- | --- | --- | --- |
| `VITE_APP_ENV` | frontend browser bundle | yes | Public frontend environment label. |
| `VITE_APP_NAME` | frontend browser bundle | yes | Public application name for UI metadata. |
| `VITE_APP_ORIGIN` | frontend browser bundle | yes | Canonical frontend origin seen by the browser. |
| `VITE_API_BASE_URL` | frontend browser bundle | yes | Public API origin consumed by the SPA. |

### Storage and media delivery

| Variable | Scope | Required | Notes |
| --- | --- | --- | --- |
| `STORAGE_BACKEND` | backend | yes | `local` now; later environments can switch to a managed backend such as S3. |
| `STORAGE_BUCKET` | backend | conditional | Required when the storage backend needs a bucket/container. |
| `STORAGE_REGION` | backend | conditional | Required when the storage backend is region-aware. |
| `STORAGE_ENDPOINT` | backend | conditional | Optional for AWS, required for S3-compatible providers that need a custom endpoint. |
| `STORAGE_PUBLIC_BASE_URL` | backend | conditional | Public asset base URL when the app serves managed media. |

## Validation Rules

Later backend and frontend shells should implement these checks directly:

- backend startup fails if `DATABASE_URL`, `APP_ENV`, `APP_PUBLIC_URL`, `API_BASE_URL`, `API_ADDR`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_HS256_SECRET`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`, `REFRESH_COOKIE_NAME`, `REFRESH_COOKIE_SECURE`, `REFRESH_COOKIE_SAME_SITE`, or `CORS_ALLOWED_ORIGINS` is missing
- frontend startup fails if `VITE_API_BASE_URL` or `VITE_APP_ORIGIN` is missing
- if refresh/logout ever runs cross-site, `CORS_ALLOWED_ORIGINS` must stay explicit, cookie settings must be reviewed for that topology, and CSRF protection must be added before shipping
- if `STORAGE_BACKEND` moves from `local` to a managed provider, the required bucket/region/public URL and provider credentials must be injected outside git-managed env files

## Environment-by-Environment Rules

### Local development

- Prefer same-site local development first.
- Keep `REFRESH_COOKIE_SECURE=false`, `REFRESH_COOKIE_SAME_SITE=lax`, and leave `REFRESH_COOKIE_DOMAIN` empty.
- Keep `CORS_ALLOWED_ORIGINS` scoped to the local frontend origin only.

### CI

- Inject the same variable names explicitly instead of relying on machine defaults.
- Use disposable secrets for CI-only auth signing and database access.
- Reuse the same migration and startup contract as local development.

### Staging and production

- Secrets move out of repo-managed env files and into a secret manager or deployment secret store.
- `REFRESH_COOKIE_SECURE=true` and HTTPS-only transport become mandatory.
- Cross-site auth flows are allowed only with explicit CORS and CSRF configuration.

## Notes For Later Beads

- `learning-english-2a6.1.6` now loads and validates backend config from this contract.
- `learning-english-2a6.1.11` locks the shared API/runtime contract in `backend/docs/API_RUNTIME_CONTRACT.md` on top of the same request-ID, health/readiness, and auth/runtime env names.
- `learning-english-2a6.1.7` should expose only `VITE_` variables to the browser.
- `learning-english-2a6.1.8` should reuse the same names in CI rather than introducing pipeline-only aliases.
- If media storage grows beyond `STORAGE_BACKEND=local`, add provider-specific secret injection without changing the public/frontend env boundary.
