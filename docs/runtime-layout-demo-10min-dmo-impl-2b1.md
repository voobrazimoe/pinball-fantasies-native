# DMO-IMPL-2B1: connected isolated demo calculation

2026-10-08. Owner approval: DMO-IMPL-2B1 ONLY.
Branch `main`, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO-IMPL-1 and DMO-IMPL-2A are retained. DMO0 NOT CLOSED. DMO1 NOT STARTED.
Pre-existing research changes and `.DS_Store` were not edited. No commit, push,
tag, release, profile registration or DMO-IMPL-2B2 work.

CONNECTED_DEMO_CALCULATION = READY

This verdict covers one bounded, candidate-owned calculation through actual
shared physics and the DMO-IMPL-1 semantic core. The candidate remains test-only
under `dmoimpl1`. It is not interactive demo support, a demo gameplay adapter,
trajectory parity, or a reachable first-equality witness. Structural tests use
explicitly arranged ball/timer states and owned canonical TABLE1 physics data;
they do not establish a real demo prefix. Private demo input pins are checked by
the preserved DMO-IMPL-1 test, not by a newly registered runtime loader.

## Changed files and interfaces

- `internal/physics/staged.go`: new internal `CandidateStage(stage, input,
  sourceHold, gate) error`. Executes an individual existing `step`, the existing
  ramps/levels primitives, or callback-free spring/target work. Requires a gate
  before each stage. Pending event and selected target have additional gates
  before publication, pending clear, SpringValid or HitX/HitY writes. Unknown
  event/target consumption remains unimplemented even if a caller admits it.
  No callbacks, drain publication, camera scroll, Sync accounting or global
  policy are installed by this entry.
- `internal/partyland/regions.go`: read-only shared first-match `areaConsumer()`
  extraction. Ordinary `checkAreas` retains first-match, tilt, unsigned coordinate,
  lastCheck, trigger and lastArea behavior and order. Candidate selects without
  running the callback or writing area bookkeeping.
- `internal/partyland/demo_core_test.go`: split the existing semantic calculation
  into sticky `electronicsPrefix` and `taskMatrixSuffix`, with the old harness
  composition/trace retained. Record typed task IDs and owned matrix installation.
  A due task with an unsupported replacement entry rejects before its WAIT reset.
  Add a distinct diagnostic phase field.
- `internal/partyland/demo_connected_test.go`: isolated fixed-policy candidate,
  fallible drain handoff and seven connected execution tests, including negative
  subtests. Calls neither Party Land Game.Sync nor any canonical callback.
- This report. Production `ball.go` and `game.go` changes from 1/2A were retained
  without additional edits in 2B1. No oracle/profile/settings/scheduler changes.

There is no second physics engine or task scheduler. Physics stages call the
existing unexported primitives; tasks still use the shared first-free allocator,
ascending identity-safe scheduler, and WAIT primitive; matrix uses the existing
interpreter after conservative typed admission. Shared Sync remains the ordinary
host entry and does not call CandidateStage. No application calls the new API.

## Connected execution phases

A successful non-drain calculation executes:

1. Two real early `physics.step` invocations, gated independently.
2. Pending event gate, tilt-counter decrement, shared ramps and levels; observe
   actual Ball.Lost at the same boundary as ordinary physics.Sync.
3. One ElectronicsCalculation: shared UPDATE_COUNTERS, one uint16 timer increment
   and equality 35998 handling; read-only area selection and callback-free
   spring/target work. BeforeTargets is never admitted as an aggregate.
4. Candidate typed tasks, then a separately budgeted matrix visit.
5. One real late `physics.step`, gated with the current source HOLDSTILL.

The candidate does not run outer native keyboard/audio/session callbacks, flash,
scroll, canonical AfterTargets/BeforeLate, FastBall scheduling or completed-Sync
accounting. Its monotonic call number is diagnostic context; the demo uint16
counter is the one timer. Stages are explicit and no general scheduler is copied.
An admitted prefix remains on later rejection; there is no rollback claim.

## Drain and HOLDSTILL

The structural drain test starts at y575 with downward velocity; the two actual
early steps set Ball.Lost by the existing PixelY >= 576 calculation. Candidate
observes it after early finish, records the old expired flag before electronics,
and never invokes canonical drain/OnEvent or sets canonical Stopped/Phase.

Both old-expired false and true cases perform admitted post-drain counters,
timer/equality and task/matrix work. Timer becomes 35998 exactly once and Random
increments once. Expiry installs its already supported CLEAR4 and cue state;
that matrix command is the only new Party Land event in this test. Score, bonus,
phase and task slots remain unchanged; there is no NEW_BALL or free plunge.
The handoff then fails explicitly at `LOOSE_BALL`, phase `post-drain`. Sticky
failure prevents any second counting. An already-lost/stopped entry is separately
rejected rather than treated as another new drain calculation. Late physics is
not claimed on this unsupported drain continuation.

Source HOLDSTILL is passed to the staged physics admission and suppresses the
whole step. It never changes native Ball.Hold. At equality, the real late-stage
invocation is held: ball state equals a control that ran only the two early
steps and ramps/levels. A later typed source release opens the real late step.
Clearing native capture hold does not clear source HOLDSTILL. Conversely, a
source release leaves native capture hold set and the ball stationary. No SETBALL,
DROPTASK2 or DURINGFLASH body is implemented.

## Supported envelope and pre-side-effect rejection

| Consumer | Admission / evidence |
| --- | --- |
| Shared early/late physics, ramps/levels | Real production primitives; source hold checked at each step |
| UPDATE_COUNTERS | Shared primitive; TunnelTime 721 or 1 rejected before invocation because nested lamp/flash effects are unsupported |
| Demo timer/equality | Existing uint16 wrap/equality/sticky-expiry behavior; incremented only in electronicsPrefix |
| No-area/no-target electronics | Read-only selectors prove no gameplay callback will run |
| Typed preserve/release/matrix/reset tasks | Existing bounded semantic envelope; only IDs allocated through typed candidate queue admitted |
| CLEAR4/WAIT matrix | Candidate-owned installation and existing conservative lookahead before immediate next dispatch |
| Pending collision callback | Rejected before pending clear, Events append or OnEvent; prior admitted collision response remains |
| Any matching area | Rejected before lastCheck/lastArea/trigger; no complete BeforeTargets aggregate admission |
| Selected target | Rejected before SpringValid, HitX/HitY clear, event append or callback |
| Unowned task/matrix | Rejected before shared dispatch; canonical closures/programs cannot enter the candidate |
| Nested unsupported task replacement | Rejected on due visit before WAIT reset or matrix installation; following guarded tasks make no mutations |
| NEXT_A to SCROLL | Rejected before final CLEAR4 visit/next dispatch; cursor retained and no late step follows |
| Demo LOOSE_BALL | Rejected after the explicitly admitted post-drain electronics suffix |

All unknown gameplay consumers remain unsupported. No score/bonus/progression,
effects, SCROLL/FADE/QUIT or first-equality gameplay bodies were added. Typed
scheduler wrappers check sticky failure before WAIT/action; the shared scheduler
may finish scanning wrappers after a rejection, but cannot execute another body,
reset a wait, remove a task, or proceed to matrix/late physics. The two-task nested
failure test checks retained IDs, occupied slots, both waits and source hold.

Diagnostics include producer, consumer, phase, monotonic calculation, relevant
guard/state and reason, headed by UNSUPPORTED_DEMO_TRANSITION. Native-stage
failures include timer/expired/sourceHold/nativeHold/lost/high/hit/tunnel state;
semantic matrix/task failures retain their specific cursor/due-entry guard plus
connected phase/calculation. First failure is sticky. Subsequent sync and direct
prefix/suffix/stage calls return the same error before mutation. No panic/recover,
silent fallback, canonical callback substitution or fictitious rollback.

## Validation

Public reproduction:

```sh
./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run '^TestDemo' -count=1 -v
```

Original-backed runs used the existing owner-local temporary source snapshot
from 1/2A with the changed Go files copied in. Existing private input/source
links were reused outside Git. No missing capture was created. Owner-local logs:
`/private/tmp/pf-dmo-impl2b1-{connected,reference,layouts,public,vet,mac-engine,windows,compile,baseline,source,payload}.log`.

| Check | Result |
| --- | --- |
| Connected no-callback calculation, explicit phase order, one increment | PASS with owned A physics |
| Drain detection, old expired flag, post-drain accounting, no canonical drain effects, no duplicate counting | PASS structural boundary tests; LOOSE_BALL remains unsupported |
| Expiry before late, held late state, typed release, native/source hold independence | PASS |
| Area/target/counter/nested task/unowned task/unowned matrix rejection and sticky calls | PASS |
| Connected matrix lookahead rejection before late; nested task failure stops later task effects | PASS |
| All existing DMO-IMPL-1/2A tests and private pinned demo check | PASS in private connected run, no skips |
| Public fixture-free TestDemo run | PASS for available tests; original-backed physics/new-ball and private demo checks skip as NOT AVAILABLE |
| Party Land, presentation, source and physics/reference suites | PASS for available tests; five opt-in capture/export checks skipped |
| TestOriginalTrajectories and Stones pf10 capture | NOT AVAILABLE; explicitly excluded, no fixture generation |
| Focused A/B/C/D datalayout/frontend Private/Possessed/Descriptor/RuntimeOriginal/RuntimeReject/RuntimeUnused/RuntimeNative/HostsShare | PASS with private inputs, no skips |
| PF6Session | Known baseline FAIL: score000002311040, ball2, ticks1200; combined layout command exits 1; unchanged |
| Tagged vet: partyland/physics/presentation/tablelogic | PASS |
| macOS c-shared engine / Windows amd64 CGO-disabled executable | PASS |
| Broad macOS compile-only sweep | FAIL: existing missing AudioDevice, hostWindow, openHost backend symbols |
| Same compile-only sweep on unchanged HEAD in separate temporary tree | Same five undefined-symbol diagnostics; baseline not fixed |
| Android / Linux native builds | NOT AVAILABLE: configured NDK / Linux SDL cross toolchain absent |
| Public source checker and git diff --check | PASS |
| Milestone source/report and two build artifacts payload scan | PASS: whole-file and nontrivial aligned 4 KiB samples against owned A/B/C/D/demo inputs; not an all-substring proof |

Reference command: `go test -tags dmoimpl1 ./internal/partyland
./internal/presentation ./internal/source ./internal/physics -skip
'TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers' -count=1 -v`.
Layout command adds PF6Session to the focused selection to expose its known
failure rather than describing the combined run as PASS. A/B/C/D definitions,
canonical A oracle and ordinary physics.Sync execution are unchanged. Available
canonical reference tests validate the area-selector extraction. The build
artifacts contain no test-only unsupported-transition diagnostic.

## Specific blockers beyond 2B1

- Actual demo area/target/event semantics remain unadmitted. Candidate conservatively
  rejects even an area identity that might be a no-op for a particular guard.
- Demo LOOSE_BALL/reset/new-ball continuation, bonus, effects and free plunge need
  separate reviewed consumers. Drain currently terminates there after one suffix.
- Unknown nested matrix/task consumers require typed fallible admission before
  expanding the envelope; CLEAR4 cannot complete its tail dispatch into SCROLL.
- Full demo operands/decoder integration, gameplay trajectory prefixes, five
  witnesses, interactive entry and profile registration remain outside scope.
- FastBall/camera/outer native session scheduling are not connected by this bounded
  calculation API. Their admission must be decided before host integration.

No DMO-IMPL-2B2 is started or implied by this verdict.
