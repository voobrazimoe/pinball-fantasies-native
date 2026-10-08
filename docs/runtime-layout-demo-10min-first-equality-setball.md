# DMO0 SETBALL release at first equality

`FIRST_EQUALITY_SETBALL_REACHABLE`

`EXPIRY_INTERLEAVING = NOT_PROVED`

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`, research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`. This is a fresh input-only witness
under the accepted canonical-A native reference and consumed demo overlays.
DMO0 NOT CLOSED. DMO1 NOT STARTED. No production implementation is changed.

Artifact: `/private/tmp/pf-dmo0-first-equality-setball.json`.
Auditor: `tools/audit_10min_demo_setball.py`; template:
`tools/dmo0_setball_test.go.txt`; tests: `tools/test_audit_10min_demo_setball.py`.
Addresses are linked demo TABLE1 file offsets unless marked DS.

## Predecessor arithmetic and real input

Unscored drain D inserts PARTY_ON_TASK1 in first-free slot0 before electronics;
its same-calculation visit increments shared DS:36c9 from0 to1. The limit30
compares before increment: D+29 leaves30 and D+30 fires, writes PARTYFLASH=true,
and calls NEW_BALL. NEW_BALL clears TASKLIST and WAITLIST. NEW_BALL_PART_TWO
queues SOUNDNEWBALL, SETBALL, SOUNDBRICKUPP in slots0/1/2. The already executing
old slot0 returns and the scan advances to slot1. Thus SETBALL's first visit is
D+30 and increments shared DS:36d3 from0 to1. It reaches80 after D+109;
visit81 on D+110 matches80 and fires. **D=35998-110=35888**.
Assuming first visit on the next scan would incorrectly select35887.

The independent visit-by-visit arithmetic tests give drain35887 -> PARTY35917,
SETBALL35997; drain35889 -> PARTY35919, SETBALL35999. These arithmetic mutations
are not trajectories or reachability evidence.

Exact deterministic target script:

- Down=true on calculations **1..35763**.
- Release=true on **35764**, Down=false.
- Every other control false throughout; all controls false after Release.

Construction is `partyland.New(DecodePartyLand(A), A)`, then
`Configure(settings.Legacy())`. No timer seed, teleport, insertion or state edit
creates the witness. The isolated reference executes linked demo timer/expiry,
PARTY_ON body and expiry operations; observers read its concrete state.
Two full fresh executions have identical snapshots, events and consumed reads.
The prefix has SCORECHANGED=false and only the constructor's initial reset.

A second real script, Down1..35757/Release35758, gives drain35887 and SETBALL35997.
A finite search of long-hold Release35300..35780 and additional short-charge
Release35720..35780 with charges1..16 did not find an unscored drain35889.
That neighbor remains UNKNOWN. The artifact distinguishes it from the proved
35888 target and arithmetic35889 expectation; no exclusion theorem is claimed.
Earlier search logs retain finite-search failures, not semantic test failures.

## Drain and NEW_BALL handoff

At drain35888, early physics has BALL_DOWN=true; drain sees timer35887,
expired=false, SCORECHANGED=false, PARTYFLASH=false, VISAKEYS=false, no live task,
and reset_count1. The observer records the exact lost ball, existing native
matrix cursor and full task/wait state before insertion. Direct PARTY_ONTS
admission installs program `0x1b243`, cursor `0x1b245`, clear remaining5.
PARTY_ON_TASK1 allocates slot0 DS:3417, wait DS:36c9 initially0. Its first visit
occurs on35888. Every scan through35917 visits it once; after35917 age=30.

On35918 its real body writes PARTYFLASH=true before NEW_BALL. PARTY_ON matrix
has already set that flag on35893/35895. The reset clears every task and wait
and increments reset_count to2, but preserves PARTY_ON program/cursor because
PARTYFLASH is true. No expiry has yet been installed. VISAKEYS remains false
under the accepted initial-reset projection and this concrete writer suffix.

After NEW_BALL_PART_TWO, immediately before resumed scan:

| Task | Slot / DS task word | DS wait word | Initial | After scan35918 | Fires |
|---|---|---|---:|---:|---:|
| SOUNDNEWBALL | 0 / 3417 | 36d1 | 0 | 0 | 35969 |
| SETBALL | 1 / 3419 | 36d3 | 0 | 1 | 35998 |
| SOUNDBRICKUPP | 2 / 341b | 36cf | 0 | 1 | 35923 |

The ball is held at (282,530), velocity(0,0), BALL_DOWN=false, expired=false.
The full handoff states and DS-word projections are in the artifact. The suffix
has exactly one SETBALL instance, the same slot1, one visit per scan and no
reset, duplicate, TABLE1 exit or concrete non-wait writer changing DS:36d3.
The sound tasks retire on the dates above. Before scan35996 SETBALL is78,
after it79; before scan35997 it79, after it80; scan35998 compares80 and zeroes it.
Linked local task store operands, common task scan, SETBALL correspondence,
complete collision masks and consumed material/sine/geometry data are checked.
This is concrete suffix evidence, not global DOS writer/interrupt closure.

## Complete calculation35998

Early physics examines a held, live ball at (282,530); no drain occurs.
DO_ELECTRONICS advances35997 ->35998 and directly installs expired=true,
HOLDSTILL=true, S_GAMEOVER2 (position13/repeat0/priority255), expiry program
`0x1ba17`, cursor `0x1ba19`, clear remaining5. Ordinary electronics continues.
DO_TASKS reaches the real slot1 with DS:36d3=80 and fires SETBALL.

SETBALL stores position(297,530), velocity(10,0), high=false, HOLDSTILL=false,
and I_UTSKJUT=ffff (native in_chute=true). Its body and the actual consumed
state both preserve expired=true, expiry program/cursor and clear remaining5.
It retires from the task list; the remaining scan finds no live task.

The subsequent admitted matrix visit keeps cursor `0x1ba19` and reduces clear
remaining5 ->4. Then one late sc_program pass executes on the newly active
ball **in this same calculation**. It reads collision material7, angle1024,
contact count13. The response takes its n<=0 exit and leaves velocity unchanged
before gravity. It updates contact metadata and advances the ball to:

- fixed-point X=304138, Y=542720;
- pixel position(297,530), velocity(10,7), gravity(0,7);
- high=false, held=false, lost=false.

The late pass can perform collision response, but area/target/drain dispatch
phases have already returned before DO_TASKS. This material7 contact produces
no pending scoring callback. Loss, if a late step set it, would be consumed by
the next early DO_PHYSICS; here no loss occurs. There is no same-calculation
area/drain/effect edge replacing or restarting expiry. The end-of-calculation
snapshot and every intervening boundary are recorded, including actual ring,
response and sine reads. SETBALL-body observation alone is not used as the result.

## Actual no-input continuation

Every subsequent control is false. The ball stays in the chute, settles at
pixel(302,537), fixed-point(309580,549955), velocity(0,7), and remains lost=false.
The first settled calculation is **36095**; the artifact records every suffix calculation.
No further drain, scored or unscored, occurs; no matrix program restarts or is
replaced. Expiry receives1042 admitted visits including35998 and enters
**QUIT on37039**. This is the concrete SETBALL continuation; no PARTY_ON-class
continuation or universal future-input termination theorem is substituted.

## Bounded inventory reconsideration

The native-relevant proved first-equality classes are:

| Class | Real witness | Immediate outcome |
|---|---|---|
| NEW_BALL_TASK | Previous collision pass | SHOWPLAYERSTS replaces expiry before first visit |
| PARTY_ON_TASK1 | Previous PARTY_ON pass | Flagged reset preserves expiry program/cursor |
| SETBALL | This pass | Preserves expiry, releases ball for same-calculation late physics |

The earlier `audit_10min_demo_expiry_interleaving.audit` inventory explicitly
has classes_complete=false. Its nonfiring new-ball/SETBALL ages are neighboring
phases of these producers. An early scored drain sees expired=false before
electronics, so its early installation is superseded by equality. The ordinary
active-ball entry is conditional on no suffix effect. The inventory also leaves
**other held-ball continuation**, with capture/mode producer not traversed.
Those premises cannot be converted into an exclusion by these three witnesses.

Exactly one smallest next dependency is selected: **DROPTASK2 capture release
at first equality** — prove or exclude an input-only surviving wait at firing
age, then classify its actual post-expiry release and same-calculation suffix.
Its native source clears HOLDSTILL at a different drop location and high=true;
it is independently produced by START_DROP, rather than NEW_BALL_PART_TWO.
First-equality reachability is UNKNOWN. This names one unresolved instance of
the existing bounded held-ball bucket; it does not reopen all effects/tasks.
Thus **EXPIRY_INTERLEAVING = NOT_PROVED**. Global task/alias closure and universal
unpaused termination remain outside this pass.

## Verification

Current-pass results and logs are included in the artifact. New tests cover
arithmetic and both drain shifts, fresh replay, drain predicate, PARTY firing,
initial resumed-scan ages, slot survival, ages79/80/fire, expiry-before-SETBALL,
HOLDSTILL/expired/cursor effects, first matrix visit, actual late physics,
no-input QUIT and both actual SETBALL limit/replacement mutations.
A +1 drain test is explicitly arithmetic, since no real35889 witness was found.

Final results: **27 new-pass tests PASS; 513 entire-demo tests PASS;
35 available canonical-A/reference tests PASS; 14 focused A/B/C/D tests PASS**,
zero skips. The isolated full replay/determinism/neighbor/mutation test also PASS.
PF6 reproduces2311040 versus2300000; TestOriginalTrajectories cannot run because
`analysis/pf2-ball-locations.json` is unavailable. The reference source/module
identity check verifies284 files equal to the checkout. An initial check named
a nonexistent runner package; the corrected source/frontend packages PASS.
`git diff --check`, production-scope check and payload checks PASS. Payload
scope is four new research files plus private JSON; whole-file, raw/hex/base64
and nontrivial aligned4KiB samples, not an all-substring theorem.
Logs use `/private/tmp/pf-setball-{new-tests,all-tests,canonical,reference,abcd,
replay,pf6,trajectories}.log`; payload metadata is
`/private/tmp/pf-setball-payload-scan.json`. Finite-search logs separately record
the unsuccessful35889 neighbor searches and are not reported as passed tests.

No production `internal`, `hosts` or `cmd` changes; no `.DS_Store` changes,
commit, push, tag, release or v0.1.3 changes. Known PF6 failure remains baseline;
TestOriginalTrajectories remains NOT AVAILABLE. DMO0 NOT CLOSED. DMO1 NOT STARTED.
