# Portfolio Checker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans for inline execution, or superpowers:subagent-driven-development if the user chooses delegation. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver a working portfolio CSV checker and interview-ready explanation in three days.

**Architecture:** A single Go process validates and calculates portfolios and exposes a generated Protobuf service. Vue calls it using the gRPC-Web protocol through Connect Web. Vite proxies RPC during development; Go serves the built frontend in production. No database or separate proxy is needed.

**Tech Stack:** Go, Connect Go, Protobuf, Buf, Vue 3, TypeScript, Vite, Connect Web, Vitest, Vue Test Utils, Playwright.

**Spec:** `outputs/portfolio-checker-design.md` (approved September 30, 2026).

## Global Constraints

- The exact CSV header is instrument_id,instrument_name,currency,market_value.
- Each value may be at most 1,000,000,000.00.
- Currency must be ZAR for this version; there is no currency conversion.
- Reject missing or extra columns, malformed CSV, header-only files, more than 1,000 holdings and files larger than 1 MiB.
- The limit accepts percentages greater than 0 and at most 100, to two decimal places.
- A breach occurs only when the unrounded allocation is strictly greater than the limit. Equality passes.
- Currency calculations use integer minor units in Go; allocation comparisons use integer arithmetic.
- It keeps no uploaded holdings between requests and does not log file contents.
- Generated Go and TypeScript bindings are checked into the repository with documented regeneration commands.
- UI copy calls this an example concentration rule.
- A local working demo is required; a GCP deployment is a stretch deliverable after local acceptance.
- Use Windows-compatible scripts; local development must not require Docker. Pin resolved dependencies and generators in lock/config files. Install any missing tool into workspace-local tooling rather than changing the system installation.

## Review Focus

1. Values just above a limit must breach even when rounded display percentages equal it (Task 1).
2. Blank lines, BOM and quoted multiline CSV names must retain physical source line numbers (Task 1).
3. Invalid UTF-8 and large protocol requests must fail cleanly without returning partial analysis (Tasks 1–2).
4. A response arriving after inputs change must not be presented as current (Task 3).
5. Uploaded markup must be rendered as text, including in error messages (Task 3).

## Repository layout

Create the deliverable repository at `outputs/portfolio-checker/`. Commands below run there unless stated otherwise. Keep downloaded toolchains and scratch files under workspace `work/`. Initialise a repository only in this new project; do not modify any enclosing repository. Copy the approved design and this plan into `docs/` at execution time.

| Files | Responsibility |
|---|---|
| `go.mod`, `go.sum` | Go module `portfolio-checker`, pinned runtime and tools |
| `internal/portfolio/check.go`, `check_test.go` | CSV domain logic and exact calculations |
| `proto/portfolio/v1/portfolio.proto`, `buf.yaml`, `buf.gen.yaml` | Wire contract and generation |
| `gen/portfolio/v1/` | Generated Go messages and service |
| `internal/rpc/server.go`, `server_test.go` | Adapter and real protocol tests |
| `cmd/server/main.go` | HTTP server, health route and static files |
| `web/src/gen/` | Generated TypeScript messages and service |
| `web/src/api.ts`, `web/src/limit.ts` | RPC transport and percent-input parsing |
| `web/src/App.vue`, `web/src/style.css` | Page orchestration and styles |
| `web/src/components/PortfolioResults.vue`, `ValidationIssues.vue` | Result and error presentation |
| `web/src/App.test.ts`, `limit.test.ts`, `web/e2e/portfolio.spec.ts` | Interaction and browser verification |
| `web/package.json`, `package-lock.json`, `vite.config.ts`, `tsconfig.json`, `index.html`, `src/main.ts` | Frontend build and entry point |
| `web/public/samples/valid.csv`, `invalid.csv` | Downloadable fictional fixtures |
| `scripts/setup.ps1`, `generate.ps1`, `dev.ps1`, `verify.ps1` | Reproducible Windows commands |
| `README.md`, `docs/interview-guide.md` | Startup, architecture and walkthrough |

## Task 1: exact portfolio logic — day 1

**Interfaces:** `Check(csvData []byte, limitBasisPoints uint32) Result` in package `portfolio`. Define `Issue{Row int, Field, Code, Message string}`, `Holding{InstrumentID, InstrumentName, Currency string; MarketValueMinor int64; AllocationBasisPoints uint32; Breached bool}`, `Analysis{TotalMinor int64; Holdings []Holding}`, `Result{Issues []Issue; Analysis *Analysis}`. A failed validation returns nonempty Issues and nil Analysis; a valid result returns Analysis and no Issues.

- [ ] Install the current supported Go release locally from the official download, verify its published SHA-256, record the version and make setup scripts locate it. Check installed Node against Vite's documented requirement. Initialise the Go module and repository.
- [ ] Write `TestCheckAllocationAndBoundary` with these core assertions:

```go
csv := []byte("instrument_id,instrument_name,currency,market_value\nA,Alpha,ZAR,60.00\nB,Beta,ZAR,40.00\n")
r := Check(csv, 5000)
if r.Analysis == nil || len(r.Issues) != 0 { t.Fatalf("invalid result: %+v", r) }
if r.Analysis.TotalMinor != 10000 { t.Fatal(r.Analysis.TotalMinor) }
if !r.Analysis.Holdings[0].Breached || r.Analysis.Holdings[1].Breached { t.Fatal("60/40 at 50%") }
if Check(csv, 6000).Analysis.Holdings[0].Breached { t.Fatal("equality must pass") }
near := []byte("instrument_id,instrument_name,currency,market_value\nA,Alpha,ZAR,5000.01\nB,Beta,ZAR,4999.99\n")
if !Check(near, 5000).Analysis.Holdings[0].Breached { t.Fatal("must compare unrounded weights") }
```

- [ ] Run `go test ./internal/portfolio`; confirm failure because Check is not implemented.
- [ ] Implement Check in `internal/portfolio/check.go` using `encoding/csv`, `utf8.Valid`, field positions and integer minor units. Compare `valueMinor * 10000 > totalMinor * int64(limitBasisPoints)`; the declared caps keep both sides within int64. Round display basis points with `(valueMinor*10000 + totalMinor/2)/totalMinor`. Do not normalise display percentages to sum to 100%.
- [ ] Add `TestCheckRejectsInvalidInputs` table cases: empty/header-only; missing/extra/reordered header; duplicate ` a ` and `A`; empty name; USD; zero/negative money; 1.001; 1e3; 1,000; 1000000000.01; malformed quote; invalid UTF-8; 1 MiB + 1 byte; 1001 holdings; limits 0 and 10001. Each case asserts `r.Analysis == nil` and `len(r.Issues) > 0`. Assert issue codes are stable (`invalid_csv`, `invalid_header`, `invalid_utf8`, `file_too_large`, `too_many_holdings`, `empty_portfolio`, `duplicate_id`, `required`, `unsupported_currency`, `invalid_amount`, `invalid_limit`).
- [ ] Add `TestCheckCSVSourceLines`: optional BOM is accepted, blank lines ignored, quoted names supported, and a malformed value on physical line 5 yields `r.Issues[0].Row == 5`. Add `TestCheckMaximumValues`: 1000 distinct holdings of 1000000000.00 yield `r.Analysis.TotalMinor == 100000000000000` without overflow. Tests must be run before fixing a discovered failure.
- [ ] Run `go test ./internal/portfolio -count=1` and `go vet ./internal/portfolio`; require exit 0. Commit the tested domain slice.

## Task 2: generated contract and real RPC — day 1

**Interfaces:** `portfolio.v1.PortfolioService.CheckPortfolio(CheckPortfolioRequest) returns (CheckPortfolioResponse)`. Request: `bytes csv_data = 1`, `uint32 limit_basis_points = 2`. Response oneof: `ValidationFailure validation_failure = 1` or `PortfolioAnalysis analysis = 2`. ValidationFailure contains repeated ValidationIssue issues; issue fields are uint32 row, string field, string code, string message (tags 1–4). PortfolioAnalysis fields: int64 total_minor (1), repeated Holding holdings (2). Holding fields map domain names in order to tags 1–6; instrument_id, instrument_name, currency are strings, market_value_minor is int64, allocation_basis_points is uint32, breached is bool. Set go_package to `portfolio-checker/gen/portfolio/v1;portfoliov1`.

`rpc.NewHandler() http.Handler` registers the generated service. `Server.CheckPortfolio(context.Context, *connect.Request[portfoliov1.CheckPortfolioRequest]) (*connect.Response[portfoliov1.CheckPortfolioResponse], error)` maps Task 1 results, using the standard wrapped Connect generator mode rather than the optional simple mode.

- [ ] Add schema and Buf configuration. Use local pinned Go generators `protoc-gen-go`, `protoc-gen-connect-go` and npm `@bufbuild/protoc-gen-es` with Buf to avoid a Docker or standalone protoc dependency. Follow current official generator compatibility instructions, record versions, generate both languages and run `buf lint`.
- [ ] Write `TestCheckPortfolioProtocols` against an HTTP/2-enabled TLS httptest server. Exercise generated Connect clients with `connect.WithGRPCWeb()` and `connect.WithGRPC()` separately. For the Task 1 60/40 fixture assert `err == nil`, `res.Msg.GetAnalysis().TotalMinor == 10000`, first breached and second not. For an empty CSV assert `err == nil`, analysis nil, validation issues nonempty. Run `go test ./internal/rpc` and confirm the missing implementation fails.
- [ ] Implement the adapter and `rpc.NewHandler()`. Restrict decoded RPC message size to 2 MiB, accommodating protobuf overhead above the 1 MiB CSV cap. Keep business validation as a typed response; malformed or excessive wire messages use protocol errors. Do not log uploaded content.
- [ ] Add `TestRPCOversizedMessage` sending more than 2 MiB and assert `connect.CodeOf(err) == connect.CodeResourceExhausted`. Verify a 1 MiB + 1-byte CSV that remains below the wire limit returns `file_too_large` with no analysis.
- [ ] Implement `cmd/server/main.go`: default localhost:8080, configurable PORT, HTTP/1 and unencrypted HTTP/2 support, `/healthz`, RPC paths and static `web/dist` files. Use the current Go supported HTTP/2 API and graceful shutdown. Deployment binds all interfaces only in the documented deployment configuration.
- [ ] Run `go test ./...` and `go vet ./...`; require exit 0. Regenerate contracts and confirm no generated diff. Commit service and generated bindings.

## Task 3: usable Vue flow — day 2

**Interfaces:** `checkPortfolio(csv: Uint8Array, limitBasisPoints: number, signal?: AbortSignal): Promise<CheckPortfolioResponse>` in `web/src/api.ts`; instantiate generated service with `createGrpcWebTransport` (not Connect transport). `parseLimit(value: string): number | null` in `limit.ts` accepts 0.01–100.00 and returns integer basis points. PortfolioResults receives `analysis: PortfolioAnalysis`; ValidationIssues receives `issues: ValidationIssue[]`.

- [ ] Scaffold Vue 3 + TypeScript using Vite, pin resolved dependencies in package-lock, add Vitest/jsdom/Vue Test Utils. Proxy `/portfolio.v1.PortfolioService` to localhost:8080. Add typecheck, test and build scripts; include generated TypeScript output from Task 2.
- [ ] Write `limit.test.ts` assertions before implementation:

```ts
expect(parseLimit('20')).toBe(2000)
expect(parseLimit('0.01')).toBe(1)
expect(parseLimit('100.00')).toBe(10000)
for (const invalid of ['0', '100.01', '-1', '1e2', '2.345', '']) {
  expect(parseLimit(invalid)).toBeNull()
}
```

- [ ] Write component tests using injected/mocked `checkPortfolio`: `blocksMissingFile` asserts zero calls and a file-required message; `showsValidationRows` asserts server issue row/code/message and no summary; `showsLoading` asserts disabled submit while a deferred promise is pending; `showsRetryAfterFailure` asserts an error and Retry button; `marksResultsStale` changes limit and asserts an outdated notice; `ignoresObsoleteResponse` changes inputs during a pending request then resolves it and asserts no current result; `rendersNamesAsText` returns `<img src=x onerror=alert(1)>` and asserts `wrapper.find('img').exists() === false` and literal name text is present. Run `npm.cmd test -- --run` and confirm expected failures.
- [ ] Implement `parseLimit`, transport wrapper, page and focused components. Use request sequence IDs plus cancellation to reject obsolete responses. Default limit is 20%; invalid inputs prevent submission. Do not cast int64 totals to floating point for calculations; format bigint minor units with exact cents. Show loading, retry, stale notices and complete validation list. Money and rule logic remain server-side.
- [ ] Style a responsive dashboard with labelled file/limit controls, summary cards, allocation bars, readable holdings, text statuses and fictional-data notice. Use escaped Vue interpolation. Include sample-download links. Use a real button for submission and live status announcements.
- [ ] Add valid.csv with A/Alpha/ZAR/60.00 and B/Beta/ZAR/40.00; invalid.csv contains a duplicate normalised ID and a negative amount. Run `npm.cmd test -- --run`, `npm.cmd run typecheck`, `npm.cmd run build`; all must exit 0. Commit the complete browser workflow.

## Task 4: local delivery and interview preparation — day 3

**Interfaces:** `scripts/setup.ps1` installs/restores pinned tools and dependencies; `generate.ps1` regenerates contracts; `dev.ps1` starts the backend and frontend and cleans up its child processes; `verify.ps1` runs Go tests/vet, frontend tests/typecheck/build. README commands must match these files.

- [ ] Write `web/e2e/portfolio.spec.ts` against the actual server and built frontend: upload valid.csv, set limit to 50%, run check and assert ZAR 100.00 total plus Alpha breach and Beta pass; change to 60%, assert outdated notice, rerun and assert zero breaches; upload invalid.csv and assert row errors and no analysis. Include a narrow mobile viewport and ensure no page-level horizontal overflow. Run once before resolving missing delivery integration.
- [ ] Implement the scripts with paths resolved from `$PSScriptRoot`; surface nonzero child exit codes. Serve the production build from Go and rerun browser tests. Check the actual browser layout at desktop and mobile widths, including keyboard focus and empty/error/loading states.
- [ ] Write README with exact prerequisites, Windows startup, tests, generation, sample schema, supported limits, Mermaid architecture, Connect library versus gRPC-Web wire-protocol distinction, no-persistence behaviour and local-demo limitations. Write `docs/interview-guide.md` with a five-minute demo, C#/Java-to-Go and Angular/React-to-Vue comparisons, exact arithmetic rationale, generated-code boundary and honest limitations.
- [ ] Run `./scripts/verify.ps1` and `npx.cmd playwright test` from the configured frontend directory; require all exit 0. Start from a clean shell using README commands. Record verified commands and any limitations in `docs/verification.md`. Review the complete diff and fix actionable issues before final verification of affected parts.
- [ ] Commit the deliverable and open the running demo for the user. Report its address and the project location. Explain the first learning checkpoint: follow one CSV through request, Go function and response.

## Conditional GCP follow-on

Only after Task 4 passes, inspect availability of Docker and an already-authorised GCP project. If available, add a multi-stage Dockerfile and Cloud Run instructions with HTTP/2 enabled and same-origin UI/RPC hosting. Validate the container locally before deployment. Do not create paid resources or claim deployment without account authorisation and an observed successful request. If prerequisites are absent, document the deployment steps and deliver the verified local demo; this does not block the core project.

## Design sources checked September 30, 2026

- https://connectrpc.com/docs/go/getting-started/ — generated Go service, Buf generation and HTTP/2 support.
- https://connectrpc.com/docs/web/choosing-a-protocol/ — explicit gRPC-Web browser transport.

## Plan self-review

All design requirements map to Tasks 1–4 or the conditional deployment step. The five review-focus cases have assigned tests. Domain money stays int64, browser money stays bigint, and response validation excludes analysis through a Protobuf oneof. No runtime versions are guessed: setup resolves supported versions once and records them before implementation proceeds. No product code or dependencies have been created by this planning step.
