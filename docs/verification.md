# Verification and delivery record

Verified locally on October 3, 2026, on Windows x64 with Go 1.27.1 and Node 24.15.0.

## Results

| Check | Result |
|---|---|
| `scripts/setup.ps1` | Passed; reused checksum-verified Go toolchain, restored locked npm dependencies |
| `scripts/verify.ps1` | Passed: Go tests, Go vet, 10 Vue/unit tests, Vue typecheck, production build |
| `scripts/generate.ps1` | Passed: Buf lint and generation; no generated-file differences |
| `npm run e2e` from web with live server | 2 passed: full real gRPC-Web flow and mobile layout |
| `scripts/run.ps1` from fresh shell | Built frontend/backend and served the app on localhost:8080 |
| Visual review | Empty and successful desktop screens, successful and invalid mobile screens inspected |
| Independent whole-project review | No Critical or Important findings; one Minor finding below |

Backend tests cover 60/40 allocations, exact equality, near-boundary breaches, maximum portfolio values, duplicate IDs, required fields, currencies, decimal formats, limits, UTF-8, BOM, quoted fields, physical holding-row numbers, malformed CSV, size caps, native gRPC, gRPC-Web, health and static serving.

The development launcher was inspected, but its process cleanup was not exercised interactively. The tested launch path is `scripts/run.ps1`. The first-time download path was exercised during initial toolchain setup and the exact published checksum was verified; the subsequent scripted setup reused that installation.

## Deferred minor findings

- A malformed or incorrect header after leading blank lines reports row 1 instead of the actual physical line. Ordinary holding-row errors use the correct physical line. Validation still rejects the file and no partial analysis is returned. This is a location-reporting defect, not a calculation defect.
- The reviewer recommended extending the existing escaped-markup component test to validation messages. Validation messages already use escaped interpolation; the current automated markup test covers holding names.

## Decisions and deviations

1. Created a standalone repository on `build/portfolio-checker`, rather than a linked worktree, because this workspace had no existing repository. If the project later belongs elsewhere, it can be moved or pushed; no existing branch was changed.
2. Used equivalent PowerShell ledger operations instead of the skill's Bash bookkeeping wrappers on Windows. This changes internal tracking, not application behaviour.
3. Pinned TypeScript to 5.9.3 because the Vue type checker could not load the compiler entry point from TypeScript 7.0.2. Limited Vitest to one worker after worker-start timeouts. These choices may need revisiting when upgrading dependencies.

## Delivery boundaries

The app is local, with no remote repository, public URL, authentication or GCP deployment. No cloud resources or paid services were created. The Git branch is preserved as the project deliverable; there is no pre-existing base branch to merge into. The independent reviewer intentionally did not assess cloud runtime behaviour or excluded features such as authentication, persistence, regulatory rules and FX conversion. Those boundaries are consistent with the approved design; applying this demo to production would require a separate design and review.

The independent reviewer did not rerun the suites; execution evidence above comes from the implementation session. The review separately inspected the code, contracts, scripts and tests.
