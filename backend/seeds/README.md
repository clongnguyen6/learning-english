# Seed Data and Sample Import Artifacts

This directory holds the reproducible local fixture corpus for `learning-english`.

Its purpose is to give the team one small, realistic set of inputs for QA, smoke demos,
and early content-ops work instead of ad hoc local data setup.

## Current Reality

- No automated seed-loader exists yet.
- The backend shell and auth schema are still landing.
- These files are still useful now because they define the canonical local fixture set that
  future seed/load paths and admin import flows should consume.

## Layout

```text
backend/seeds/
├── README.md
├── manifest.json
├── auth/
│   └── users.json
└── imports/
    ├── grammar/
    │   └── present-perfect.md
    ├── reading/
    │   ├── a-visit-to-da-lat.md
    │   └── street-food-weekend.txt
    └── vocabulary/
        └── travel-basics.csv
```

## How To Use This Seed Corpus

Use these files as the default local fixture inputs when the corresponding runtime paths land:

- `auth/users.json`
  Local-only account manifest for future auth seeding.
- `imports/vocabulary/travel-basics.csv`
  Happy-path vocabulary import sample.
- `imports/grammar/present-perfect.md`
  Happy-path grammar Markdown lesson.
- `imports/reading/a-visit-to-da-lat.md`
  Primary bilingual Markdown reading sample.
- `imports/reading/street-food-weekend.txt`
  Alternate plain-text reading sample that stresses the second parser path.

`manifest.json` is the machine-readable catalog. Update it whenever a sample is added,
renamed, or materially changed.

## Scope Choices

- Keep the dataset small and deterministic.
- Prefer one high-signal fixture per flow over a large noisy corpus.
- Keep at least one artifact that exercises a less common path or edge case.
- Do not invent new import contracts here. The samples should follow formats already
  described in `PLAN.md`.

## Notes

- Plaintext passwords in `auth/users.json` are for local development only.
- Vocabulary JSON is intentionally omitted for now because `PLAN.md` names CSV/JSON
  support but does not yet lock a canonical JSON payload shape.
