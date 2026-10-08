# DMO-IMPL-2A: conservative fallible consumption boundary

2026-10-08. Owner approval: DMO-IMPL-2A ONLY. DMO-IMPL-1 remains accepted.
HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`, branch `main`.
Pre-existing dirty research files and `.DS_Store` were left alone.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

Delivered a conservative real-physics boundary and a test-only native candidate.
This is **partial infrastructure**, not a demo gameplay adapter or trajectory
parity proof. No DMO-IMPL-2B, profile registration, oracle change, demo drain
rules, effects, SCROLL/FADE/QUIT, witnesses or interactive entry was implemented.
No commit/push/tag/release was performed.

## Production interfaces and isolation

Only `internal/physics/ball.go` changes production code in this milestone:

- Existing `(*physics.Game).Sync(Inputs) error` delegates to an unexported shared
  implementation with a nil gate. Existing callback signatures and return values
  are unchanged. A/B/C/D hosts continue calling this entry.
- New `SyncWithGate(Inputs, func(producer, consumer string) error) error` is the
  explicit candidate entry. A nil gate returns an error before any mutation.
  The candidate-entry guard precedes even the Events slice reset. Each consumer
  guard returns through the existing physics error path immediately.
- No global gate, panic/recover, host flag, profile or shared policy was added.
  The API is in an internal package. No application calls it.

`demoNativeCandidate`, its fixed admission policy and diagnostic are `_test.go`
code under the existing `dmoimpl1` tag. It accepts only entry, shared physics
steps and ramps/levels; **all canonical gameplay aggregates are rejected**.
There is no configurable acceptance escape hatch and no call to `Game.Sync`.
This entry intentionally does not execute the outer Party Land Sync prefix
(keyboard/audio/session), staged timer or post-drain electronics.
The two candidate APIs are separate experiments; no combined calculation is
claimed. Do not call `Game.Sync` on the candidate's underlying Game.

The lower-level physics API is not intrinsically sticky: its caller owns failure
lifetime. The test-only candidate checks its stored error before any calculation,
trace or physics mutation and retains the first rejection. Native arithmetic
errors also become sticky candidate diagnostics. Diagnostic fields are producer,
consumer, phase/lost/hold/high state, calculation number and reason, headed by
`UNSUPPORTED_DEMO_TRANSITION`. No rollback occurs: admitted early physics changes
remain when the subsequent gameplay callback is rejected.

The DMO-IMPL-1 diagnostic now includes consumer and calculation context. Its
existing field order users remain valid through named initialization. Its direct
`matrixVisit` entry now checks sticky failure before accessing/mutating the
matrix. Timer, tasks, waits and matrix execution logic otherwise remain intact.
The staged harness remains test-only, with its original conservative matrix
lookahead. Its diagnostic calculation context is the staged uint16 counter;
the separate native candidate uses a monotonic uint64 call number.

## Boundary inventory and pre-effect guarantees

Admission names below denote whole operations, not proof of their demo semantics.
A future policy must not admit an aggregate merely because one child is supported.

| Production path | Guard now | Guarantee and remaining limitation |
| --- | --- | --- |
| physics Sync entry / each early or late step | Before invocation | Rejecting a step prevents its collision/push/flipper/movement effects; admitted shared steps can enqueue a deferred event |
| pending bumper/event callback | Before event publication, pending flag clear and OnEvent | No callback score/sound/reward effect starts on rejection; earlier collision response is retained |
| ramps / levels | Before aggregate | Entire shared operation admitted by current candidate |
| drain return | Before Stopped, drain event append and OnEvent | Canonical drain, phase change, effects and task allocation never start |
| BeforeTargets | Before callback | UPDATE_COUNTERS and checkAreas both blocked; not a timer-only hook |
| areas / area callbacks | Enclosed by denied BeforeTargets | No individual area gate yet; parent calls can already have effects before a nested consumer |
| spring / target dispatch | Before entire checkSpringAndTargets | Rejection also precedes SpringValid/HitX/HitY changes; individual target callback selection is not exposed |
| AfterTargets | Before callback | Tilt, previousInput, flash, tasks, spring and lamp work all blocked |
| task bodies / nested insertion | Enclosed by denied gameplay callbacks | No production per-body or insertion gate; existing first-free allocator/identity scan unchanged |
| BeforeLate / matrix visit | Before callback | Entire presentationTick blocked; no production per-op gate |
| matrix dispatch / immediate tail dispatch | Enclosed by denied callbacks | No individual production dispatch gate; inserting one inside dispatch alone cannot stop a void caller's suffix |
| effects / admission | Enclosed by denied gameplay callbacks | No individual effect gate; effectTiming may request a jingle before score/bonus/matrix work |
| scroll / ScrollForce | Before aggregate | Blocks callback and camera changes |
| Party Land Game.Sync outer prefix and BallLost suffix | Not intercepted | Candidate never enters them; attaching this boundary to Game.Sync is not safe yet |

Consequently unknown gameplay cannot run within the current fixed candidate
policy, but precise selective demo gameplay consumption remains unimplemented.
The physics guards preserve ordinary nil-gate callback order. They do not make
arbitrary existing void callbacks fallible.

## Real execution tests

New tests are in `internal/partyland/demo_gate_test.go`:

- `TestDemoNativeGateBeforeTargets`: constructs actual canonical-backed native
  Games, starts at their real initial ball state without teleportation or HOLD,
  executes two physics steps and ramps/levels, and rejects BeforeTargets. It
  checks the admitted sequence and actual ball movement, unchanged gameplay
  counter/score/events, no completed sync, structured diagnostic, identical
  sticky error on repeated calls and no further state change. An ordinary
  `Physics.Sync` on the other Game from the same initial state executes the
  callback, increments the counter and completes the sync.
- `TestDemoNativeGateDrainBeforeCallback`: actual physics Sync reaches its drain
  branch from an explicitly arranged out-of-table state. It rejects before
  Stopped/event publication, phase change, scoring or task insertion. This is
  a structural boundary test, **not a reachable demo drain witness**.
- `TestDemoNativeGateRejectsAllGameplayAggregates`: fixed-policy negative checks
  for all canonical gameplay aggregates and an unknown consumer. These are
  policy tests, not execution coverage of every callback site.
- `TestDemoGateStickyMatrixEntry`: a supported expiry matrix is installed, an
  error latched, and a direct later matrixVisit cannot advance the cursor.

The supported consumers demonstrated here are shared native physics operations,
not demo-specific gameplay consumers. No demo trajectory parity is established.
There is no synthetic replacement for the initial native execution test.

## Regressions and reproduction

Public run:

```sh
./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run '^TestDemo' -count=1 -v
```

Original-backed runs reuse the owner-local DMO-IMPL-1 temporary source snapshot,
with only this milestone's Go files copied in. Existing private originals and
historical source links remain outside Git. No fixture was created to fill a
missing capture. Logs: `/private/tmp/pf-dmo-impl2a-*.log`.

| Check | Result |
| --- | --- |
| New positive/negative native gate tests, existing DMO-IMPL-1 suite and canonical demo matrix test with available originals | PASS; no skips in gate run |
| Fixture-free TestDemo suite | PASS; original-backed new-ball/native tests and opt-in private timer skip when inputs are absent |
| Party Land, presentation, source and physics suites | PASS for available tests; five opt-in capture/export checks skipped; OriginalTrajectories and missing Stones capture explicitly excluded |
| A/B/C/D focused datalayout/frontend: Private, Possessed, Descriptor, RuntimeOriginal, RuntimeReject, RuntimeUnused, RuntimeNative, HostsShare | PASS with private inputs, no skips |
| PF6Session in same focused run | Known FAIL: score000002311040, ball2, ticks1200; same DMO-IMPL-1 baseline result; combined command exits 1 |
| go vet with dmoimpl1 on partyland/physics/presentation/tablelogic | PASS |
| macOS shared engine, c-shared | PASS |
| Windows amd64 executable, CGO disabled | PASS |
| Broad macOS compile-only sweep | FAIL: internal/platform lacks AudioDevice, hostWindow, openHost; cmd/pinballfantasies also fails |
| Same sweep on unchanged HEAD exported by git archive into a separate temporary tree | Same FAIL with the same five undefined-symbol diagnostics; unrelated backend was not changed |
| Android / Linux native builds | NOT AVAILABLE here: no configured Android NDK or Linux SDL cross toolchain |
| Public source checker / git diff --check | PASS |
| Milestone payload scan | PASS for four source/report files and two build artifacts against owned A/B/C/D/demo runtime inputs: whole-file and nontrivial aligned 4 KiB samples; not an all-substring proof |

Reference command used `go test -tags dmoimpl1` on partyland, presentation,
source and physics, with `-skip
'TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers'` and `-count=1`.
Neither TestOriginalTrajectories nor missing private captures were generated.
All DMO-IMPL-1 wrapping/equality35998/sticky-expiry/shared-wait/task-order/matrix
cursor tests passed. Canonical A oracle and all runtime profile definitions are
unchanged. Build artifacts contain no test-only unsupported diagnostic.

## Can real demo drain / HOLDSTILL be connected safely?

**No, not yet.** The current candidate safely refuses those canonical consumers,
but cannot resume into a fallible demo electronics/task/matrix/late-physics
suffix. Removing the aggregate rejection would remove the safety guarantee.

The concrete blocker is continuation after a nested failure in existing void
callbacks and task closures. Minimal next design options, not implemented here:

1. Candidate-owned fallible callbacks for BeforeTargets/AfterTargets/BeforeLate
   and drain, with explicit returns after each admitted child; retain existing
   void callbacks exclusively for ordinary hosts. Add fallible task iteration
   and dispatch propagation where a callback may continue after a rejection.
2. Candidate-owned staged execution of those aggregates, reusing only validated
   shared primitives and admitting typed task/matrix/effect operations before
   invocation. This requires an explicit post-drain and late-HOLDSTILL adapter.

Either needs explicit boundary tests for each newly admitted consumer before
expanding the fixed envelope. This milestone stops here; it does not silently
inherit canonical gameplay semantics to make progress.

FALLIBLE_CONSUMPTION_GATE = PARTIAL
