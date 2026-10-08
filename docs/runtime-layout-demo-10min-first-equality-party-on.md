# DMO0 first-equality PARTY_ON_TASK1

`FIRST_EQUALITY_PARTY_ON_COLLISION_REACHABLE`

`EXPIRY_INTERLEAVING = NOT_PROVED`

DMO0 NOT CLOSED. DMO1 NOT STARTED. This is one concrete fresh-session
input-only witness under the accepted canonical-A native reference. Production
base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD:
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

Artifact: `/private/tmp/pf-dmo0-first-equality-party-on.json`.
Auditor: `tools/audit_10min_demo_party_on.py`; isolated harness template:
`tools/dmo0_party_on_test.go.txt`. No production sources are changed.

## Exact target and fresh input

The linked primary calls DO_PHYSICS at file `0x46b8`. Its BALL_DOWN/LOOSING
checks reach LOOSE_BALL at `0x5d4a`; the unscored branch directly installs
PARTY_ONTS, requests S_SPRING, and inserts PARTY_ON_TASK1 at `0x5b6` before
returning. The admitted common rest then calls UPDATE_COUNTERS at `0x4720`
and DO_ELECTRONICS at `0x4723`. The electronics suffix calls DO_TASKS at
`0x5d16`. First-free insertion walks DS:3417..3479 in ascending order; scan
starts at DS:3417 and visits fifty slots in ascending order. The real drain has
no live tasks: insertion selects slot0, DS:3417. Thus its **first visit is in
the same calculation**, not the next calculation.

Compare-before-increment means D=35998-30=**35968**: visit1 on35968 increments
0 to1; scan35996 increments28 to29; scan35997 increments29 to30;
scan35998 matches30, resets the shared word, and enters the task body.
A next-calculation first visit would instead require D=35967, and is rejected
by both linked ordering and the actual first-visit observer.

The deterministic fresh input is:

- Down=true on calculations **1..35841**.
- Release=true on **35842**, Down=false.
- All other controls false throughout; all controls false after release.

Construction is the accepted `New(DecodePartyLand(A), A)` followed by
`Configure(settings.Legacy())`. There are no timer seeds, task insertions,
teleports or state edits in the witness. Demo-only consumed timer/expiry and
PARTY_ON task semantics run in the isolated copied reference. The source
constructor's initial reset is counted; no later reset occurs before this drain.
The prefix remains SCORECHANGED=false. The real first drain is35968, sees
expired=false and timer35967; its electronics suffix advances timer to35968.
Task/wait/guard observations are passive. The mutation runs are explicitly
separate and do not establish reachability.

Real neighboring drain witnesses also exist: Down1..35837, Release35838 gives
unscored drain35967/firing35997; Down1..35840, Release35841 gives unscored
drain35969/firing35999. Changing the target Release by +1 to35843 instead gives
drain35977; release and drain indices are not assumed to shift linearly.
The finite search is not an exclusion or impossibility proof.

## Concrete suffix and equality

PARTY_ON admission is a **direct DO_MATRIX call**, not a generic inhibited
scoring effect. Its MatrixStarted event, actual command/cursor and allocated
slot establish admission. The stale native `effect_accepted` field is not used
as evidence. DS:36c9 initially equals0 from the fresh WAITLIST clear; ADDTASK
does not clear or otherwise write it. Every calculation35968..35997 visits
exactly one PARTY_ON_TASK1 in slot0. No other task, duplicate, task-list clear,
reset or concrete writer changes that shared word before the equality match.
Linked PARTYRUT is RET. These are concrete suffix checks, not global writer or
indirect-domain closure.

PARTYFLASH is false at the drain. Importantly, the PARTY_ON matrix already
sets it true: _PARTYONN on35973, then _PARTYON on35975. PARTYRUT remains current
through35997 with next cursor `0x1b257`. VISAKEYS=false follows the accepted
linked initial reset/NEW_BALL correspondence; it is a projected linked flag,
not a field exposed by the Go Game. MUSICOK DS:d1 is distinct from VISAKEYS.
There is no fresh-game input/reset or VISAKEYS writer in this concrete suffix.

On35998 the ball is still held/lost; no new drain occurs. UPDATE_COUNTERS
precedes DO_ELECTRONICS. Electronics increments35997 to35998 and directly
installs expired=true, HOLDSTILL=true, S_GAMEOVER2 and expiry `0x1ba17`, with
next cursor `0x1ba19` and initial clear remaining5. Ordinary electronics
continues and DO_TASKS visits slot0 with DS:36c9=30.

The wait resets to0; the task again writes PARTYFLASH=true **before NEW_BALL**.
NEW_BALL clears TASKLIST/WAITLIST, then its PARTYFLASH guard jumps directly to
the reset routine's RET before VISAKEYS/SHOWPLAYERSTS. The matrix therefore
remains expiry: **program, cursor and clear remaining5 are preserved**, with
no restart or replacement. NEW_BALL queues SOUNDNEWBALL/SETBALL/SOUNDBRICKUPP
in slots0/1/2. The scan continues at slot1: SETBALL and SOUNDBRICKUPP age to1;
new SOUNDNEWBALL in slot0 remains0. The admitted matrix visit reduces the clear
remaining from5 to4 on35998.

Clearing PARTYFLASH immediately before reset takes the SHOWPLAYERSTS branch
and replaces expiry. Delaying/removing only the task's flag store preserves
expiry because the real matrix prefix has already set PARTYFLASH. A separate
mutation clears it immediately before the task body; the authentic body
restores true before reset and preserves expiry. These tests distinguish the
actual guard state from an assumption that this task is its sole flag writer.

## This continuation's result

No further input is needed. SOUNDBRICKUPP fires36003, SOUNDNEWBALL36049,
SETBALL36078. SETBALL clears HOLDSTILL and returns the ball to the chute;
expired stays true, and SETBALL does not alter the matrix. With no release,
there is no subsequent drain or program replacement. The decoded expiry
program receives1042 admitted visits including35998 and reaches **QUIT37039**.
The artifact records each suffix calculation, queued-task/body boundaries and
all matrix-operation dispatches. QUIT denotes entry into the linked handler;
DOS teardown is outside this pass.

The reachable first-equality classes currently distinguished are the prior
PARTYFLASH=false/VISAKEYS=false NEW_BALL_TASK replacement by SHOWPLAYERSTS,
and this PARTYFLASH=true PARTY_ON_TASK1 cursor-preserving reset. A SETBALL
release at first equality has no independent reachability evidence asserted
here. `EXPIRY_INTERLEAVING` stays NOT_PROVED; the one selected remaining
dependency is reachability or exclusion and effect classification of SETBALL
clearing HOLDSTILL at first equality with expiry current. This list is not a
completeness proof. The already proved pause counterexample and universal
unpaused termination are not reconsidered.

## Verification

Results are recorded in the artifact and current-pass logs. Required checks:
new linked/replay/mutation tests; entire demo research suite; available
canonical-A/reference tests; focused A/B/C/D regressions; known PF6 baseline;
whitespace, production-scope and payload checks. TestOriginalTrajectories
remains NOT AVAILABLE; no trajectory fixture is generated.

Current results: **26 new tests PASS; 486 complete demo research tests PASS;
35 available canonical-A/reference checks PASS; 14 focused A/B/C/D checks PASS**,
without skips. The deterministic native search/replay/neighbors/guard mutation
test passes. PF6 reproduces the known 2311040 versus2300000 baseline failure;
TestOriginalTrajectories is NOT AVAILABLE. `git diff --check` and production
scope checks PASS. Payload checks PASS for the four new research files and
owner-local JSON: whole-file identity, encoding and48,898 nontrivial aligned
4K samples from249 private files/eleven decoded SDR modules. This sampled
check is not an all-substring proof. An initial broad internal Go invocation
encountered the unrelated platform build boundary; the selected reference
packages above were rerun successfully. No production code or commit changed.
