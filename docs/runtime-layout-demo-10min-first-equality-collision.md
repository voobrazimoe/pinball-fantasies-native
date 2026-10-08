# DMO0 concrete first-equality NEW_BALL_TASK collision

`NEW_BALL_THRESHOLD_PROVENANCE = PROVED`

`FIRST_EQUALITY_COLLISION_REACHABLE`

`FIRST_EQUALITY_EXPIRY_REPLACEMENT = REACHABLE`

`ATOMIC_EXPIRY_MODEL = DISPROVED`

`EXPIRY_INTERLEAVING = NOT_PROVED`

DMO0 NOT CLOSED. DMO1 NOT STARTED. This is one concrete suffix under the
accepted fresh native semantic schedule. It establishes replacement later in
the same logical calculation, not physical DOS timing, eventual QUIT, or an
indefinitely continuing session.

Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
Research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
Owner-local input: `/private/tmp/pf-dmo0-deterministic-bygel-drain.json`.
Output: `/private/tmp/pf-dmo0-first-equality-collision.json`.

## Actual handoff

The saved replay's `task_ids` are insertion identities, not live function
slots. They survive suicide/reset in the native reference. A passive repeat of
the exact saved input script reconstructs live slots and reset counts. Every
original checkpoint and compressed trace row is compared with the saved replay;
none of the original selected state changes.

At LOSTBALL installation on 35877, all 50 live slots are free, represented by
CS DUMRET `0x6a71`. The stored ID in slot 0 belongs to the already completed
ENABLETOUCHER. Its shared word DS:`0x36db` is zero. The drain tail inserts
SOUNDRINNER into slot 0 before that calculation's scan; its word DS:`0x36cb`
becomes 1. The post-drain checkpoint confirms this. It fires at 35882 after
ages 1..5, requests sound and suicides. It cannot reset the task list, replace
the program, or touch DS:`0x36cd`.

The repeated prefix has exactly one native reset, at fresh construction, and
no live NEW_BALL_TASK instance or NEW_BALL_TASK callsite use. Initial reset
zeros all 50 shared wait words. DS:`0x36cd` is therefore zero at drain and at
producer insertion. The artifact records both full 50-word TASKLIST and
WAITLIST projections, including unused zero words.

PARTYFLASH is false at drain. The concrete CLOSE1 boundary on 35481 has
LASTAREA=BYGEL12, selecting its clear store DS:`0x00d0`. No later concrete
prefix/suffix effect sets it.

Correction to the previous CLOSE1 annotation: its additional store at file
`0x2596` is DS:`0x00d1`, the adjacent MUSICOK byte. It does **not** clear
VISAKEYS, which the reset guard reads at DS:`0x34f1`. VISAKEYS=false is instead
reconstructed from the selected fresh entry: WHEN_NEW_GAME_RESET sets it true;
initial PARTYFLASH is false; initial NEW_BALL consumes that true value and
clears it. There is no further new-game/reset entry on the input prefix. The
only live prefix wait tasks are ENABLETOUCHER and SOUNDRINNER, and their
linked effects cannot write VISAKEYS. The concrete continuation has no F1/new
game input. VISAKEYS remains false through the actual guard checks on 35998.

## Concrete suffix and producer

The accepted zero-aggregate, XXBALLE=false program route at `0x1b459` is reused
with its proved 91 admitted visits and H=0. The executable suffix dispatches
the producer when its actual route cursor reaches `0x1b533`; the producer date
is asserted afterwards, rather than scheduling insertion at an authored date.
The four JBCDZ branches and KOLLA_XXBALL take the already proved zero route.
No traversed routine resets tasks/waits or installs a competing matrix program.
The lost, held ball admits no further geometry/target/drain callback on this
suffix. KEYTASK is DUMRET. Sound completion affects cue readiness, and the
traversed matrix operations have no readiness wait or animation.

Immediately before producer insertion on 35967, all 50 slots are free.
First-free slot is **0**, DS:`0x3417`. `_DEMOVER_CHANGE_PLAYER` inserts CS
NEW_BALL_TASK `0x0bbb` there, after DO_TASKS. DS:`0x36cd` is **0**, with
exactly one live callsite instance. PARTYFLASH=false, VISAKEYS=false.
The producer preserves matrix BX; its HU_ tail installs the following CLEAR4,
whose next cursor is `0x1b537`. The next five matrix visits finish that clear;
the following WAIT 32000 has next cursor `0x1b53b`. It stays current until
expiry supersedes it.

For each scan 35968..35997, slot 0 survives and is visited once. Compare-before-
increment changes DS:`0x36cd` from 0 through 30. No reset, duplicate,
callsite interference, control exit, or guard writer occurs. The artifact
records every calculation, live slot list, shared words, visits, matrix cursor,
and guard values. A separate continuation from the actual replay Game executes
native task/reset primitives and checks every task slot and NEW_BALL age
against the linked demo suffix. It supplies the full selected native reset/ball
state; the demo matrix/expiry state is explicitly merged into the final state.
This probe uses the linked demo producer and does not pretend A contains the
demo expiry program.

## Calculation 35998

The inherited phase order reaches UPDATE_COUNTERS and then DO_ELECTRONICS.
The timer increments from 35997 to exactly 35998. The equality branch sets
expired=true, HOLDSTILL=true, requests S_GAMEOVER2 and installs expiry
`0x1ba17`, with first CLEAR4 and next cursor `0x1ba19`. Ordinary work continues
to KEYTASK and DO_TASKS before budgeted matrix work.

Slot 0 is scanned with DS:`0x36cd`=30. WAITSYNCS zeroes the shared word and
fires. NEW_BALL calls reset, clearing all tasks and all wait words before its
guard checks. Both PARTYFLASH and VISAKEYS are false. Reset installs
SHOWPLAYERSTS `0x1b88e`, overwriting expiry before expiry receives a matrix
visit. No saved expiry cursor exists. The reused SHOWPLAYERSTS contract is
clear, two prints, terminator, with no QUIT.

NEW_BALL_PART_TWO inserts SOUNDNEWBALL, SETBALL and SOUNDBRICKUPP into slots
0, 1 and 2. The current scan resumes at slot 1, so SETBALL and SOUNDBRICKUPP
reach age 1 while SOUNDNEWBALL stays at age 0. The artifact binds the actual
demo operands: SOUNDNEWBALL uses DS:`0x36d1`/50 and SOUNDBRICKUPP uses
DS:`0x36cf`/5. Budgeted matrix work then visits the new SHOWPLAYERSTS clear,
leaving remaining=4 and next cursor `0x1b890`.

Post-collision: timer=35998; expired=true; ball=(282,530), held; BALL_DOWN=false;
LOOSING=false; I_UTSKJUT=true; SCORECHANGED=false; NEW_BALL wait=0; reset count=2.
The prior expiry cursor appears only in historical trace metadata, not as a
saved runtime continuation.

## Verification

Results and owner-local log paths are recorded in the artifact. New tests cover
handoff, cursor-derived producer, actual first-free slot, initial age, duplicate
absence, ages 1/29/30, firing, reset, expiry-before-scan order, replacement and
cursor loss. Adversarial mutations cover slot survival, shared wait interference,
PARTYFLASH/VISAKEYS writes, linked operands, SHOWPLAYERSTS content and producer
shifts to 35966/35968; these fire at 35997/35999, away from first equality.

The full current demo research suite, available A ordering/reference tests,
focused A/B/C/D regressions, diff whitespace and sampled payload checks are run.
TestOriginalTrajectories remains NOT AVAILABLE. The known
TestPF6SessionKeepsGameplayOracle baseline is kept separate from new regressions.
Payload checking is whole-file plus nontrivial aligned 4 KiB sample matching,
including decoded SDR modules and encoded candidates, not an all-substring proof.

No production internal/hosts/cmd edits, .DS_Store change, commit, push, tag,
release, v0.1.3 change, predecessor search, whole writer closure, or eventual
termination pass is included.

Final results: 19 new-pass tests PASS; full current demo suite 433 tests PASS;
22 A ordering/reference tests PASS with no private skips; 11 focused A/B/C/D
tests PASS with no private skips. The PF6 baseline was reproduced exactly at
score 2311040 versus 2300000, ball=2, ticks=1200. TestOriginalTrajectories is
NOT AVAILABLE. `git diff --check` PASS; production-path diff from research HEAD
empty. Payload scan PASS against 249 private files, 11 decoded SDR modules and
48,898 nontrivial aligned 4 KiB samples; no encoded payload candidates.
