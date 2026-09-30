# Portfolio Checker: project design

## Purpose and success
Build a small, complete interview project in three days that demonstrates the employer's Go, Vue, TypeScript, Protobuf and gRPC-Web stack. The developer brings professional Angular, C#, Java and React experience. Every major decision should be explainable using those familiar concepts.

Success means a reviewer can start the project, load fictional holdings, see validation failures, inspect allocations and concentration breaches, and reproduce the calculations through tests. A local working demo is required; a GCP deployment is a stretch deliverable after local acceptance.

## User journey
One page contains a CSV upload, downloadable fictional sample, concentration-limit input, summary cards, allocation bars and a holdings table. The default limit is 20%. The user selects a CSV and runs a check. Invalid files show row-specific messages and no partial portfolio results. Valid files show total market value, holding count, largest allocation and breach count. Changing the limit requires running the check again; changing an input marks the previous result as outdated.

## Input and financial assumptions
The exact CSV header is instrument_id,instrument_name,currency,market_value. market_value is the current holding value, not a price or quantity. Values must be positive plain decimal amounts with at most two decimal places and no grouping separators, currency symbols or scientific notation. Each value may be at most 1,000,000,000.00. Currency must be ZAR for this version; there is no currency conversion. Instrument IDs are trimmed and uppercased; duplicate normalized IDs are rejected. Names must be nonempty after trimming. UTF-8 CSV with an optional byte-order mark is supported. Empty rows are ignored. Reject missing or extra columns, malformed CSV, header-only files, more than 1,000 holdings and files larger than 1 MiB.

The limit accepts percentages greater than 0 and at most 100, to two decimal places. A breach occurs only when the unrounded allocation is strictly greater than the limit. Equality passes. Currency calculations use integer minor units in Go; allocation comparisons use integer arithmetic. Rounded percentages are for display only. There are no claims of regulatory compliance: UI copy calls this an example concentration rule.

## Architecture and contracts
Use one Go service and one Vue application, without a database. The Go service owns CSV parsing, validation, calculations and concentration checks. It keeps no uploaded holdings between requests and does not log file contents. Vue handles file selection and presents server responses; it does not duplicate portfolio business rules.

Define a portfolio.v1.PortfolioService with a unary CheckPortfolio RPC. Protobuf request fields carry CSV bytes and limit_basis_points (20% = 2000). Response fields carry either row-level validation issues or a complete analysis containing total minor units, holding details, allocation basis points for display and breach flags. Generated Go and TypeScript bindings are checked into the repository with documented regeneration commands. Validation issues carry a row number (header is row 1; zero denotes a file-level issue), field, code and readable message.

Use a Go gRPC-compatible server library with native gRPC and gRPC-Web support, and a generated TypeScript client explicitly configured for gRPC-Web. Prefer this single-server arrangement to a separate proxy for the short deadline. The specific library versions and generator compatibility must be verified against official documentation during setup. Development uses a Vue dev-server proxy to the Go service. Production serves built frontend assets and RPC endpoints from the Go process, simplifying same-origin browser access. Native gRPC support is verified by an integration test.

## Errors and interaction
Use inline validation for missing files and invalid limits, a visible loading state and disabled duplicate submissions. The server remains the source of truth for all validation. Show connection errors with a retry action. Clear outdated validation messages on resubmission, and prevent an older request from replacing newer results. Render uploaded text as plain text. Provide labelled inputs, keyboard-accessible controls and text breach labels alongside colour.

## Scope boundaries
Included: CSV upload, sample data, validation, allocations, configurable concentration rule, generated contracts, meaningful tests, local setup, architecture explanation and interview walkthrough.

Excluded: authentication, persistent storage, live financial feeds, actual trading, rebalancing, AI matching, statutory reporting and multiple currencies. The app is a fictional-data demonstration. These exclusions keep the promised first release achievable.

## Verification and completion criteria
Backend tests cover parsing and row numbers, duplicate IDs, malformed/oversized files, unsupported currencies, invalid money values, portfolio totals, exact-limit equality and immediately-above-limit results without rounding errors. Include a 60/40 fixture whose 50% limit flags only the first holding. RPC integration tests exercise a valid portfolio and a validation failure through gRPC-Web, plus native gRPC compatibility. Frontend checks cover loading, failed requests, stale results and validation rendering. A browser walkthrough uploads both the provided valid and invalid samples and changes the limit. Type checking and a production build must pass.

The README includes reproducible Windows startup instructions, sample schema, architecture diagram, test commands, calculation assumptions and limitations. An interview guide explains Go types and errors, Vue concepts relative to Angular/React, generated contracts and why concentration checks stay deterministic.

## Delivery sequence
Day 1: establish the contract, backend calculations and a working browser-to-service request.
Day 2: complete the CSV workflow, results interface and integration tests.
Day 3: complete verification, documentation and walkthrough; then attempt containerisation and GCP deployment if time and account access permit. Deployment spending or account setup is not assumed.

## Environment findings
The workspace currently contains only work and outputs folders. Node, npm and Git are available on PATH. Go, Docker and protoc were not found on PATH. Setup should install Go locally if needed and use a reproducible generator approach without requiring Docker for local development.
