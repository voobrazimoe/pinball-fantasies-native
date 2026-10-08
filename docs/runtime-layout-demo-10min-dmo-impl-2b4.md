# DMO-IMPL-2B4: isolated unscored new-ball handoff

2026-10-08. Owner approval: DMO-IMPL-2B4 ONLY.
Branch `main`, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

PARTY_ON_TASK1_BODY = READY

DEMO_NEW_BALL_HANDOFF = READY

READY covers the connected, isolated `dmoimpl1` test-only candidate's flagged
unscored branch with verified reset operands, ADDPLAYERS=false and MUSICOK=true.
It ends with a source-held ball and three waiting tasks. No SETBALL release,
child sound body, new gameplay cycle or official fresh demo witness is claimed.
Structural checkpoints deliberately prepare state; they are not input-only proofs.

## Implementation and admission

`internal/partyland/demo_core_test.go` retains the shared scheduler, allocator,
identity-safe task removal and compare-before-increment wait. PARTY_ON_TASK1
uses its existing DS:36c9 site, limit30: installation before electronics receives
age1 in that calculation, reaches age30 after 30 visits and fires on visit31.
Before the due WAITSYNCS reset, `preflightNewBall` admits the complete permitted
body: reset inputs, player index/table, duck restore operands, reset text and
ADDPLAYERS/MUSICOK branch. Missing consumers produce a specific sticky error
with the old task, age30, matrix, ball and PARTYFLASH body store untouched.
Earlier admitted calculation effects, including UPDATE_COUNTERS/timer and
MATRIX_BLINKOR, are not rolled back.

On success, the wait helper resets DS:36c9, the body writes PARTYFLASH=true,
and `newBallHandoff` executes the explicit typed reset in
`internal/partyland/demo_new_ball_test.go`. It calls neither canonical
`Game.newBall` nor canonical `resetBall`. Necessary shared leaf consumers are
LoadPlayerState, palette/lamp writes, duck mask restoration, flash allocation,
SetBall, ResetTilt and the existing first-free task allocator. No second scan,
new dispatcher or ordinary Game.Sync callback is introduced.

Reset operand admission uses the existing pinned private input adapter.
`loadResetOperands` compares all three demo RESTOREDn masks (DS:69b0/69f0/6a40,
30/30/15 bytes) with the shared consumer's owned inputs. It also checks the
actual demo BONUS_TEXT at DS:2325 against the loaded text and owns a mutable
copy before the reset's byte11='8' store. Missing/mismatched data fails closed;
no original bytes or text are stored in this repository. PARTY_ON presentation
admission remains independent: budget=false can hand off without dispatching
PARTYONN or printing its text. The test adapters share identity verification.

The reset projects the reviewed generic/table state explicitly: TASKLIST and
all WAITLIST sites clear; LOOSING/BALL_DOWN/SPECIALMODE clear; SPRING_VALID,
DOT_READY and both jingle readiness flags set; shift/inhibit-countdown and
INH_EFF clear; EOTS sets. Table transient counters, mode/feature/disable flags,
lamps/flash slots, multiplier and BONUS_X reset; three duck targets restore.
P_STRUC_2_VARS loads the saved player's score, bonus, skills, cyclone count and
17 progression lamps rather than retaining current live values. Extra-ball
lamp51 is then restored when XBALLS is nonzero. Happy/Mega totals clear.
The native typed projection uses existing Decimal/player/geometry consumers;
it does not claim a complete raw DOS memory or external-handler proof.

The source SKILLSHOTDOWNCOUNTER is not a RESET_VARS store: SkillTime receives
only the earlier UPDATE_COUNTERS decrement. Flipper angles, spin/gravity and
the table-lifetime demo timer/expired latch are not reset. Source HOLDSTILL is
set true separately from native capture Ball.Hold, which is not written.
SCREENFORCE2 becomes -1. NEW_BALL_PART_TWO allocates the three tasks, enables
ALLOWFLIP, clears TILTFLAG/TILTCOUNTER and clears SCORECHANGED.

ADDPLAYERS=true requires the out-of-scope SNART_NEW_BALL path. MUSICOK=false
requires the not-admitted spring cue path. These are concrete preflight blockers,
not silently skipped consumers. The accepted connected unscored drain itself
sets MUSICOK=true. No new DMO0 reachability/domain research was performed.

## Source evidence reused

The accepted 2B3 report supplies PARTY_ONTS and scheduler ownership.
Existing `runtime-layout-demo-10min-first-equality-party-on.md`,
`runtime-layout-demo-10min-first-equality-setball.md`, expiry-interleaving and
canonical-A-jitter evidence provide the local reset/guard/wait correspondence.
The previously reviewed canonical-A-jitter SLICES cover RESET_VARS,
table_new_ball, generic_new_ball_prefix, reset_tasks, reset_waits, NEW_BALL,
NEW_BALL_PART_TWO and SLACK_LIGHTS. External admission/alias closure remains
outside these local facts and is not promoted to READY.

The existing pinned private input test now checks bounded concrete linked
operands: PARTYFLASH store file5c9; NEW_BALL fileece..f7d;
NEW_BALL_PART_TWO f7d..fa8; reset flags and text/bonus stores; TASKLIST/WAITLIST
extent50; and PARTYFLASH/VISAKEYS guards. This is input drift validation,
not a graph, trajectory or writer-domain search.

## Exact handoff state and matrix ownership

For a normal connected drain with initial DS:36c9=0 and PARTY_ON_TASK1 slot0,
the firing calculation is31. At the end of its task scan:

| State | Value |
| --- | --- |
| Ball position | pixels (282,530), fixed coordinates (282*1024,530*1024) |
| Ball velocity / level | VX=0, VY=0, High=false |
| Hold / capture | source HOLDSTILL=true; Ball.Hold remains false in this drain path |
| Ball/flow flags | Lost=false (BALL_DOWN cleared), LOOSING=false, Phase=NewBall |
| Chute/controls | I_UTSKJUT/inChute=true, SpringValid=true, AllowFlip=true, Tilted=false, TiltCounter=0 |
| Timer / expired | counter31; incoming sticky expired preserved; no reset-generated expiry cue |
| TASKLIST slot0 | SOUNDNEWBALL, wait limit50, DS:36d1, age0 |
| TASKLIST slot1 | SETBALL, wait limit80, DS:36d3, age1 |
| TASKLIST slot2 | SOUNDBRICKUPP, wait limit5, DS:36cf, age1 |
| TASKLIST slots3..49 | empty |
| WAITLIST | old sites all zero/absent; only SETBALL=1 and SOUNDBRICKUPP=1 are nonzero |

The child IDs are new and monotonic. Old IDs in empty slots are metadata, not
live tasks. Slot0's new ID prevents the returning old closure from removing
SOUNDNEWBALL. No old PARTY_ON_TASK1 closure fires twice.

A. With matrix budget=true, PARTYONN already sets PARTYFLASH at calculation6,
and PARTYON installs persistent PARTYRUT at8. Reset preserves its command
program, op_PARTYON, remaining1, next5, dots and matrix flash words. Ordinary
MATRIX_BLINKOR still runs before tasks; no reset KILL_FLASHOR occurs.

B. With matrix budget=false for the first30 calculations, CLEAR4 stays
remaining5/next1 and PARTYFLASH is false before firing. The task body sets it
true before reset, so the same CLEAR4 program/cursor survives. No SHOWPLAYERSTS
is installed. VISAKEYS=true is independently tested: the PARTYFLASH branch
skips its test/clear and retains that flag.

For the structural collision, counter35967 precedes installation/age1;
age30 is at35997. Electronics35998 sets sticky expired, source hold, the expiry
cue and installs expiry CLEAR4 before tasks. The due PARTY_ON_TASK1 reset
preserves this expiry program instead of restoring PARTY_ONTS. Budget=false
leaves expiry remaining5/next1; budget=true gives its ordinary matrix visit,
remaining4/next1. In both cases the held-ball and task states above apply,
with counter35998 and expired=true. Separate non-equality execution proves
an already-set expired latch survives reset without generating another cue.

## Remaining scan, unsupported boundary and tests

Reset stays inside the current ascending DO_TASKS scan. New tasks occupy0/1/2.
When the parent occupied slot1, only new slot2 receives this calculation's
visit (ages0/0/1). With the parent in slot3, all three new slots wait for the
next scan (ages0/0/0). With parent slot0 the ages are0/1/1. The following scan
adds exactly one visit to each; reset never restarts at slot0.

All three children are typed wait-only tasks. Their due checks return sticky
UNSUPPORTED_DEMO_TRANSITION before body execution and before the shared wait
reset; no cue or SETBALL position/velocity/release occurs. In the ordinary
slot0 continuation with matrix budget=false, the first next unsupported
consumer is SOUNDBRICKUPP at calculation36, retaining DS:36cf=5. SOUNDNEWBALL
and SETBALL due rejection are separately tested using declared structural due
checkpoints, without claiming that a run can pass the earlier child failure.
If expiry is active and matrix budget remains true, expiry's unsupported SCROLL
can be encountered first (36002 in the equality fixture); expiry execution
remains outside this milestone.

Connected tests cover installation/age1/age30/firing, body store, complete old
task/wait reset, saved player restoration, transient flags, real duck masks,
held position/velocity/level, matrix budgets and guards, equality-before-task
ordering, allocator/identity survival, both scan age cases, no immediate release,
all child due gates and sticky cessation of subsequent calculations. Direct
reset operand/preflight entry also honors sticky failure. All accepted
DMO-IMPL-1/2A/2B1/2B2/2B3 tests remain; former due-body refusal assertions now
assert the approved handoff while retaining their scheduler/matrix checks.

## Validation

Logs/builds: `/private/tmp/pf-dmo-impl2b4-*`. Private fixture execution uses the
existing canonical-A regression checkout with current candidate test files
copied in and the existing pinned owner demo. Missing fixtures were not created.

| Check | Result |
| --- | --- |
| Tagged TestDemo, preserved milestones plus 2B4, pinned private inputs | PASS, 40 top-level tests |
| Public fixture-free tagged TestDemo | PASS available checks; original-backed checks skip NOT AVAILABLE |
| Canonical-A/reference partyland, presentation, source, physics | PASS available suites; five opt-in capture/export checks skip |
| TestOriginalTrajectories / Stones pf10 capture | NOT AVAILABLE; excluded, no fixture generation |
| A/B/C/D datalayout and frontend private compatibility checks | PASS |
| PF6Session in combined layout command | Known baseline FAIL: score000002311040, ball2, ticks1200; combined exit1 |
| Tagged vet: partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 CGO-disabled executable and engine package | PASS |
| macOS c-shared engine | PASS |
| Broad macOS compile-only and unchanged baseline checkout | Same baseline FAIL: AudioDevice, hostWindow, openHost |
| Public source checker / diff whitespace | PASS |
| Milestone sources/report and two build artifacts payload checks | PASS; whole-file/nontrivial aligned4KiB samples; not an all-substring proof |
| Fresh input-only official demo candidate witnesses | NOT AVAILABLE; not attempted |

Reproduction: `./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run
'^TestDemo' -count=1 -v`. Reference suites skip
`TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers` only because
those inputs/captures are unavailable. Compatibility execution includes the
private installation/layout/profile suites and reports PF6 separately.

Only five existing test-only files and this new reset test-only file/report were
edited for this milestone. Production files match the pre-existing accepted
regression checkout; no oracle/profile/interactive support changed. `.DS_Store`
was not edited. No commit/push/tag/release. DMO-IMPL-2B5 is not started.
