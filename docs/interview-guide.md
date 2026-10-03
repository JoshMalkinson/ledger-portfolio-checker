# Explain Ledger in an interview

## The honest introduction

“My professional background is Angular, React, C# and Java. I built this portfolio checker with AI assistance to learn Go, Vue and Protobuf/gRPC in a workflow relevant to your team. It validates fictional holdings, calculates allocations and checks a configurable concentration rule. I can walk through the design, tests and limitations.”

Only say you can explain or change a part once you have actually worked through it. Keep this project separate from professional experience on your CV.

## Five-minute walkthrough

**0:00–0:45 — The problem.** A data problem should not masquerade as a valid portfolio result. We validate the entire import before computing any totals.

**0:45–1:45 — The demo.** Upload valid.csv, choose 50%, show Alpha above the limit and Beta within it. Change to 60%: the old result becomes outdated, and the next check passes both holdings. Upload invalid.csv and show the duplicate ID and bad value, both on row 3.

**1:45–2:45 — Follow one request.** Start at `web/src/App.vue`: read the selected file, parse the limit, pass bytes to `checkPortfolio`. In `web/src/api.ts`, the generated service uses gRPC-Web. The RPC handler calls `portfolio.Check`; the result is mapped to the Protobuf oneof and rendered by Vue.

**2:45–3:45 — A deliberate decision.** Show `internal/portfolio/check.go`. Money is cents, the limit is basis points, and the comparison uses integer multiplication. A boundary test proves equality passes; another proves a value just over 50% fails even when its display rounds down to 50.00%.

**3:45–4:30 — Evidence.** Run the tests. Point out that protocol integration tests call the actual generated client over HTTP, and the browser test checks that the wire content type is gRPC-Web. Mocks are limited to timing/failure cases in component tests.

**4:30–5:00 — Limits and next steps.** No real regulation is encoded. No live positions or trades are involved. Before a real application, you would need identity and authorization, versioned data snapshots and rules, persistence, operational observability, deployment controls and domain review.

## Transfer your existing knowledge

| Familiar concept | This project's equivalent | Difference to explain |
|---|---|---|
| C#/Java DTO | Go struct; generated Protobuf message | Domain structs remain separate from transport types |
| Interface implementation | Go method set | Satisfaction is implicit rather than an `implements` declaration |
| Exceptions | Go error return values | Expected invalid CSV data is a typed result, while transport failures use errors |
| ASP.NET/Spring handler | Connect Go handler | Protobuf generates the typed service boundary |
| Angular service / React API module | `web/src/api.ts` | Client methods and messages come from the schema |
| Angular signals / React state | Vue `ref`, `shallowRef`, `computed` | Vue tracks reactive reads; templates unwrap refs |
| Input binding | Vue `v-model` | It synchronizes the limit text with local state |
| Component props | Vue `defineProps` | Result and error components receive generated typed data |
| Component teardown | `onBeforeUnmount` | Cancels work and invalidates late responses |

## Questions to practise

**Why Protobuf instead of hand-written JSON interfaces?** One schema generates both sides and provides explicit message evolution rules. It does not replace runtime input validation. Adding fields is generally easier than changing the meaning or type of an existing field; never reuse removed field numbers.

**Why gRPC-Web?** Browser APIs do not directly expose all native gRPC transport capabilities. The browser uses gRPC-Web and the Go library accepts it. Native gRPC is also tested. The choice here matches the target stack; a conventional REST API would also be viable for this small workflow.

**Why no microservices?** One stateless service covers this bounded workflow. Splitting validation and calculation across deployments would add network failures and operational work without a demonstrated need.

**Why reject duplicate instrument IDs?** The file contract expects one holding per instrument. Silently merging rows can hide a data-quality issue. A real aggregation feature would need explicit rules and auditability.

**Why a Protobuf oneof?** The API makes it impossible to populate both the error and analysis variants simultaneously. The client still handles a missing variant as a failure.

**Why no AI matching yet?** The core checks have exact answers and should be deterministic. A future LLM could propose matches for ambiguous instrument names, with human approval and a retained audit record; it should not decide the arithmetic.

**How do you prevent a stale result?** Abort the request when inputs change, and use a monotonically increasing request ID before accepting any response. Aborting alone is insufficient when a response is already resolving.

**How would you scale it?** Measure first. Larger files could become asynchronous jobs with object storage, bounded workers, import IDs and status endpoints. Preserve reproducibility by versioning source data and rules. This demo intentionally caps synchronous requests.

**What did you learn?** Answer from experience after reading and changing the code. Good concrete examples are integer money handling, Go's explicit errors, Vue reactivity and protocol-generated contracts. Do not claim years of experience with the new tools.

## Your first hands-on exercise

1. Run the app and complete the two-minute demo without assistance.
2. Open the Protobuf schema. Locate the request fields and response variants.
3. Follow the request through `api.ts`, `server.go`, and `check.go`.
4. Change the sample values to 50 and 50, predict the result at a 50% limit, then verify it.
5. Add a third fictional holding and explain why every percentage changes.
6. Pick one existing boundary test and explain exactly which incorrect implementation would make it fail.

Once comfortable, make a small feature yourself—for example, a client-side filter showing only breaches—and explain how you verified it. That gives you a concrete contribution to discuss beyond running generated code.
