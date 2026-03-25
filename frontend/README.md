# Frontend Scaffold

This directory is reserved for the React and TypeScript application described in `PLAN.md`.

Current scope:
- `src/` for the application code
- `public/` for static assets served by the frontend

Config contract:
- frontend runtime config should come from the root `.env.example` contract using `VITE_` variables only
- secrets do not belong in frontend env files
- the first required frontend values are the app origin and API base URL so routing, auth refresh, and API clients do not drift

This bead only establishes the repository shape. Real frontend shell work belongs to `learning-english-2a6.1.7`.
