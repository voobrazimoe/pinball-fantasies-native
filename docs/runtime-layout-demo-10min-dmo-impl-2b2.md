# DMO-IMPL-2B2: unscored LOOSE_BALL → PARTY_ON

2026-10-08. Owner approval: DMO-IMPL-2B2 ONLY.
Branch `main`, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_UNSCORED_DRAIN = READY

POST_DRAIN_ORDER = SOURCE_CONSISTENT

These verdicts cover the bounded, isolated candidate accepted in 1/2A/2B1,
which remains behind `dmoimpl1` in test files. A real shared native physics drain
now invokes an actual demo LOOSE_BALL consumer before counters/electronics.
They do not assert complete PARTY_ON presentation, a fresh input-only prefix,
interactive support, a registered profile, or whole-DOS equivalence.

## Changed files and interfaces

- `internal/partyland/demo_connected_test.go`: fallible demo drain consumer,
  typed PARTY_ONTS metadata, isolated LOOSING/SPECIALMODE/HI_RES/SCREENFORCE2
  projection, and boundary snapshots. Connected execution resumes pending
  lost-ball calculations without counting another drain. Sticky rejection also
  covers direct drain calls. Existing sticky tests now exercise direct prefix,
  suffix, matrix, physics, allocation and installation calls after failure.
- `internal/partyland/demo_core_test.go`: typed PARTY_ON_TASK1 wait action with
  exact site/limit validation and pre-body rejection. Uses the existing shared
  allocator, scheduler and wait; does not introduce another scheduler.
- `internal/partyland/demo_core_checks_test.go`: private pinned input check now
  checks reviewed linked drain store operands, producer operands, source spring
  tuple, SCORECHANGED guard and PARTY_ONTS entry/operands. No source search,
  reachability search or new DMO0 audit is performed.
- `internal/partyland/demo_drain_test.go`: new structural linked-calculation,
  ordering, cursor, task, guard, rejection and continuation tests; preserved
  structural post-drain timer assertion separated from gameplay execution.
- This report.

No production interface needed modification in 2B2. The previously accepted
production changes in game.go, regions.go, ball.go and staged.go are retained.
No canonical drain, PARTY_ON, bonus or NEW_BALL callback is used as a substitute.
No oracle, runtime descriptor, profile, host/backend or scheduler policy changes.
Pre-existing work and `.DS_Store` were not edited. No commit/push/tag/release.

## Actual execution order and source state

1. Two gated shared early physics steps, then the existing pending-event,
   tilt-counter, ramps/levels boundary. Native movement sets Ball.Lost at y≥576.
2. Observe BALL_DOWN and isolated LOOSING; only a new loss enters LOOSE_BALL.
   Record old expired, actual Game.ScoreChanged, ball, live slots and guards.
3. Reject the unimplemented PUKEFORBIDDEN alternate return before entering it.
   For the admitted common prefix, store source HOLDSTILL=true, SCREENFORCE2
   369/259 for projected low/high resolution, place the ball at (15,47), velocity
   (0,0), high=false, and clear AllowFlip, SPECIALMODE, HAPPY_HOUR and MEGA_LAUGH.
   Set LOOSING=true and candidate-visible Phase=BallLost. BALL_DOWN remains true.
   Source HOLDSTILL does not set native capture Ball.Hold or canonical Stopped.
4. Select the branch from actual Game.ScoreChanged. For false, clear audio
   priority, directly install PARTY_ONTS, request S_SPRING `(0,0,1)` through
   shared MusicClock.Play, store MUSICOK=true, and allocate PARTY_ON_TASK1.
   No effect admission or bonus callback is invoked.
5. Run shared UPDATE_COUNTERS and the existing electronics timer prefix once.
   Equality35998 may replace PARTY_ONTS with the already admitted expiry entry.
6. Scan typed tasks in the shared ascending 50-slot scheduler.
7. Visit the matrix only when budget admits it and typed current/next gates pass.
8. Invoke the gated shared late physics step with source HOLDSTILL. The drain
   branch's hold suppresses movement and canonical continuation. A typed source
   release while LOOSING is rejected before late physics.

Subsequent pending calculations retain BALL_DOWN/LOOSING and the task, perform
one electronics increment each, and do not re-enter LOOSE_BALL. Areas/targets
remain skipped for the lost ball. Shared early ramps/levels may update gravity;
source HOLDSTILL holds step movement, not every ramps/levels state variable.

Game.ScoreChanged is never rewritten to make the branch pass. Expired and
HOLDSTILL belong to the existing isolated semantic core. LOOSING is separate
from native Stopped. SPECIALMODE has an independent projected flag; the drain
clears it without clearing ModeTime or calling a canonical reset. HI_RES and
its SCREENFORCE2 result are isolated structural projection inputs, not a claim
that a demo decoder or fresh initializer is connected. PARTYFLASH, INH_EFF,
PUKEFORBIDDEN, task slots and audio use actual candidate/Game state.

Old expired is recorded before timer increment; unscored PARTY_ON does not use
it as an admission condition. PARTYFLASH is preserved, not fabricated or set by
the task. INH_EFF and SPECIALMODE are not effect-admission guards for this direct
DO_MATRIX branch. Tests exercise INH_EFF=true and incoming SPECIALMODE=true;
source mode clears happen before installation. Audio starts at priority255 in
the structural test: the source priority-clear admits priority1. Shared Play
preserves its actual return-position/repeat admission behavior. No host playback
callback or outer audio tick is introduced.

Reviewed linked sites: LOOSE_BALL HOLDSTILL515; placement532..566;
AllowFlip56c; SPECIALMODE572; mode flags578/57d; PUKEFORBIDDEN582;
LOOSING58c; SCORECHANGED592; priority59c; direct matrix5a5; spring5ab;
MUSICOK5ae; ADDTASK5b6; PARTY_ON_TASK1 wait5ba..5c8. These are semantic metadata,
not copied executable bodies. PARTY_ONTS starts at linked file1b243.

## Task and matrix consumption boundaries

PARTY_ON_TASK1 is inserted through the shared first-free allocator. Insertion
never zeros DS:36c9: this is the shared `PARTY_ON_TASK1` wait key. Limit30 uses
compare-before-increment. An empty structural wait is0 at installation and1
following the same calculation's scan. A hole at slot1 between live slots0/2
is selected and visited in that scan. A pre-existing shared age7 becomes8,
proving insertion did not reset it. Existing 2B1/core tests separately retain
ascending scan, identity-safe removal and insertion-behind-cursor behavior.

At age30 the wrapper rejects with UNSUPPORTED_DEMO_TRANSITION before shared
WAIT resets the word, before PARTYFLASH=true, and before NEW_BALL. Age30, task
identity, slot, phase and matrix remain intact. Later wrappers see the sticky
failure. There is no body that silently removes or completes this task.

Direct PARTY_ONTS installation admits only its CLEAR4 entry. It owns the
reviewed command/operand records: CLEAR4, FLASHON(3), PARTYONN(1),
PRINT13(PARTY_ON_TEXT,336), PARTYON(1). Only CLEAR4 is executable in this
milestone. The first budgeted visit changes remaining5→4 and retains next1;
a budgetless calculation preserves both. Replacing a previous owned WAIT
program with next2 resets the cursor to the new entry's next1.
At remaining1, existing conservative lookahead rejects before dispatching
FLASHON(3), including before the final CLEAR4 visit. Later PARTY_ON operations
are metadata only and are not assumed admitted.

## Unsupported branches and rejection guarantees

- SCORECHANGED=true selects a named scored-drain rejection after the admitted
  common ball-loss stores, before its expired/LOSTBALL branch consumers. Timer,
  Random, audio, matrix, task slots, score, bonus and events do not acquire
  scored-branch effects. Old expired=false and true are both tested, including
  counter35997. No fictitious expiry program or canonical bonus is installed.
  The later expired scored-drain effect0x1a4a1 is not implemented or inherited.
- PUKEFORBIDDEN return, unowned tasks/matrices, pending event/area/target
  callbacks and nested tunnel effects remain unsupported.
- Full TASKLIST rejects before allocator invocation and before electronics;
  the already admitted placement/matrix/audio prefix remains. No rollback is
  claimed. Nested unsupported matrix replacement rejects before its WAIT reset.
- The PARTY task body, FLASHON and later presentation consumers reject before
  their bodies. Matrix failure stops late physics; task failure stops subsequent
  task effects, matrix and late physics.
- First failure is sticky: repeated sync and direct consumer calls return the
  same error before any additional counter, wait, cursor, task or ball mutation.
  Diagnostics retain producer/consumer, phase/calculation and relevant guards.

No demo bonus, scored continuation, NEW_BALL/reset, SCROLL/FADE/QUIT, witness
search, interactive demo, profile registration, DMO0 research or DMO1 is added.

## Evidence and validation

Structural evidence uses an explicitly prepared ball at y575 with downward
velocity and owned canonical-A physics data. Two real native steps set Lost;
the new demo consumer executes in the same connected path. Boundary snapshots
include physical ball state, source hold/loss/guards, timer/expired, actual audio,
matrix routine/cursor/remaining, all occupied task identities and shared age.
They prove ordering through state, independently of the retained call trace.
Sentinel callbacks fail tests if any canonical physics gameplay aggregate runs.

Snapshots show: native loss before placement; placement and source hold before
matrix installation; matrix remaining5/task age0/timer old before electronics;
remaining4/age1/timer old+1 after it. Equality tests show spring priority1 before
expiry priority255, with the task surviving and aging once. Next-calculation
snapshots prove no double drain and persistent task/cursor state.

Reachability: NOT AVAILABLE for this candidate. No fresh input-only prefix was
constructed or replayed here. Prior research witnesses are not relabeled as
2B2 candidate witnesses. Structural tests are not official demo witnesses.
The preserved `TestDemoStructuralPostDrainTimerSuffix` and 1's timer tests
explicitly exercise only arranged semantic suffix state, not complete drains.

Public reproduction:

```sh
./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run '^TestDemo' -count=1 -v
```

Private tests reuse the existing owner-local snapshot and originals outside
Git, with this milestone's test files copied in. Production files were compared
against the working tree before regressions. No missing fixture was created.
Logs: `/private/tmp/pf-dmo-impl2b2-*.log`.

| Check | Result |
| --- | --- |
| DMO-IMPL-1, 2A, preserved 2B1 no-drain/hold/rejection tests, new 2B2 tests and private linked operands | PASS with available originals |
| Old expired false/true; real native drain; ordering snapshots; one timer increment; first-free slot/shared age; pending survival; equality replacement | PASS structural |
| Scored, PUKEFORBIDDEN, full allocator, nested task, matrix-next and lost-ball release rejection; sticky calls | PASS |
| Fresh input-only reachability / official candidate witness | NOT AVAILABLE; not attempted |
| Public fixture-free TestDemo | PASS for available checks; original-backed checks skip as NOT AVAILABLE |
| Available Party Land/presentation/source/physics canonical-A/reference suites | PASS; five opt-in capture/export checks skip |
| TestOriginalTrajectories / missing Stones pf10 capture | NOT AVAILABLE; excluded without fixture generation |
| Focused A/B/C/D datalayout/frontend Private/Possessed/Descriptor/RuntimeOriginal/RuntimeReject/RuntimeUnused/RuntimeNative/HostsShare | PASS with private inputs |
| PF6Session in combined layout run | Baseline FAIL: score000002311040, ball2, ticks1200; combined command exits1 |
| Tagged vet: partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 executable, CGO disabled / macOS c-shared engine | PASS |
| Broad macOS compile-only / unchanged HEAD baseline tree | Same baseline FAIL: AudioDevice, hostWindow, openHost; not fixed |
| Public source checker / git diff --check | PASS |
| Five milestone source/report files and two build artifacts payload checks | PASS; whole-file and nontrivial aligned4KiB samples against owned A/B/C/D/demo inputs, not an all-substring proof |

Reference command retains `-tags dmoimpl1` on partyland, presentation, source
and physics with `-skip 'TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers'`.
The layout command includes PF6Session to expose the baseline. An initial layout
invocation reversed the two private CD environment roots; profile-identity tests
caught it. The corrected run passes the selected A/B/C/D checks and retains only
the known PF6 failure. This was a test configuration error, not a product fix.

## Next blocker

With regular matrix budget, the first unsupported continuation is PARTY_ONTS
FLASHON(3) after CLEAR4. Without matrix budget, the independent blocker is
PARTY_ON_TASK1's due body at DS:36c9=30 (PARTYFLASH store and NEW_BALL/reset).
Scored-drain continuation remains separately blocked at its selected branch.
No DMO-IMPL-2B3 is started.
