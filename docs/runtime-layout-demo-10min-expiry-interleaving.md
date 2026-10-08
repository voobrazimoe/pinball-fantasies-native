# DMO0 expiry interleaving: narrow provenance checkpoint

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

The result is **EXPIRY_INTERLEAVING = NOT_PROVED**. The pass does not establish
a complete set of reachable threshold states or an eventual termination theorem.
Canonical A timing inheritance and native audio boundary remain proved premises.
No timing, production implementation or whole-DOS resolver changes were made.

The single smallest remaining fact selected here is:

> Can NEW_BALL_TASK produced by the real normal demo bonus continuation survive
> to the first ElectronicsCalculation 35998 with shared wait word DS:0x36cd=30
> and PARTYFLASH=VISAKEYS=false at its task visit?

This is a predecessor/age provenance obligation, not a request to close the
entire task list or all writers. Passing conditional tests below does not answer
it. A complete real gameplay prefix to this state has not been established.

All addresses below are demo TABLE1 file offsets unless marked DS. Private
inputs are pinned by existing research identity checks. No payload is exported.

## Producer and age constraints

Normal bonus node `0x1b533` dispatches `_DEMOVER_CHANGE_PLAYER`; `0x782/0x785`
queues NEW_BALL_TASK through first-free task insertion. The linked task at
`0xebb` uses wait limit 30 and shared callsite word DS:0x36cd. This word belongs
to the wait invocation, not to the allocated task slot. DOADDTASK does not zero
it. NEW_BALL reset clears the wait list before queuing new-ball preparation.

From a zero word, 30 visits leave age 30; visit 31 fires and zeroes the word.
Since the normal bonus producer runs in matrix work after the task scan, an
**assumed** enqueue after calculation 35967 first visits at 35968, reaches age 30
after 35997, and fires at 35998. This arithmetic is a conditional suffix, not a
reachable execution witness. Slot survival and the producer's actual calculation
index/guards still require a real prefix. Enqueue after 35968 instead leaves age
29 at equality and can fire on the next scan, provided no intervening reset/exit.

NEW_BALL_PART_TWO inserts SOUNDNEWBALL, SETBALL and SOUNDBRICKUPP in that order.
SETBALL has limit 80 and shared word DS:0x36d3. Reset zeroes waits, but tasks
inserted ahead of the current scan cursor can receive their first visit in that
same scan; slots behind it wait until the next scan. Thus wait age cannot be
inferred simply by subtracting a new-ball creation index. Coexistence of an older
NEW_BALL_TASK and these new tasks is not asserted: NEW_BALL clears the old list.

The direct unscored-drain producer queues PARTY_ON_TASK1 (`0x5b6`). Its wait uses
DS:0x36c9 and limit 30. Its body sets PARTYFLASH before NEW_BALL, so its reset
locally skips SHOWPLAYERSTS. This does not prove its threshold-age provenance.
The scored drain before equality sees the old expired flag because drain precedes
the counted electronics entry. After expiry, the scored-drain branch instead
selects the expiry replay effect. These are distinct predecessor conditions.

## What the local suffix establishes

The new JSON lists candidate classes for ordinary play, normal bonus/new-ball,
SETBALL waiting, scored/unscored drain and other held-ball continuations. Each
has `UNKNOWN_AT_FIRST_EQUALITY`; `classes_complete=false`. Player/bonus phase,
ball/down flags, guards and matrix cursor are marked unknown where no real
predecessor has been proved. No arbitrary task contents are labeled reachable.

For the conditional new-ball firing class, the existing reset edge replaces
expiry with SHOWPLAYERSTS if both guards are false. The narrowly decoded
SHOWPLAYERSTS at `0x1b88e` consists of clear and two print commands followed by
zero. It contains no QUIT and saves no suspended expiry cursor. Reaching its
terminator therefore supplies no automatic restoration or expiry QUIT edge.
A finite trace without QUIT is not an infinite TABLE1 proof.

The existing scored-drain expired branch at `0x622/0x625` selects effect record
`0x1a4a1`; its matrix field points to `0x1ba17`, the expiry entry. This is an
explicit potential reinstallation edge, **not a guaranteed future drain or
admitted effect**. Consumed effect flags and cue admission must permit it.
Reinstallation starts the expiry program again; it does not resume its old cursor.

SETBALL can locally clear HOLDSTILL while expired remains true and expiry remains
current. This does not directly cancel the matrix program or stop counter counting.
Subsequent physics/drain/program effects require their own actual continuation;
no ball trajectory is fabricated by the conditional slice.

Expiry reaches QUIT status zero if its decoded operations complete while it
remains current. The test advances each completed operation as an explicit
premise; it is not a cadence, budget fairness or eventual completion proof.
Replacement, restore, indefinite continuation and alternate exit remain unproved
for actual sessions. The minimal native expiry rule is therefore unresolved.
No atomic expiry rule or guaranteed eventual QUIT may be added from this pass.

Counter increment remains conditional on each admitted ElectronicsCalculation,
even after expired is set. Wrap would require 29,538 additional calculations after
first equality, and re-equality 65,536. Survival that long is unproved, so no
wrap/re-equality execution simulation is used as session evidence.

## Verification

New tests cover conditional no-task expiry, the three NEW_BALL wait cases,
SETBALL same/later release, program replacement, PARTYFLASH/VISAKEYS ordering,
unscored PARTY_ON guard, expiry operation completion, finite replacement without
QUIT, admitted/rejected replay, continued counter increment, shared wait insertion
and scan-slot reuse. Linked operand mutation tests reject changed wait addresses
and limits. Separate assertions prevent conditional results and arithmetic wrap
obligations from upgrading the reachability verdict.

- New pass: **30 PASS**, zero SKIP.
- Entire demo research suite, including this pass: **268 PASS**, zero SKIP.
- Strict A/native ordering: **24 top-level / 347 including subtests PASS**, zero SKIP.
- Focused A/B/C/D: **15 top-level / 99 including subtests PASS**, zero SKIP.
- `git diff --check` and raw/encoded payload scan: **PASS**.

Strict A and final A/B/C/D checks use the existing owner-local original-backed
harness; all 273 Go-source/module files match the current checkout. The initial
Go invocation lacked the local toolchain in PATH; the explicit `.tools/go/bin/go`
rerun succeeds. The public-checkout A/B/C/D invocation skipped one original-backed
matrix integration check; the final private harness rerun has zero SKIP.

Machine-readable evidence: `/private/tmp/pf-dmo0-expiry-interleaving.json`.
Logs: `/private/tmp/pf-dmo0-expiry-{new,all,strict-A,ABCD-private}-tests.log`.
Payload scan: `/private/tmp/pf-dmo0-expiry-payload-scan.json`; whole-file and aligned
nontrivial 4 KiB raw/hex/base64 checks, not an all-substring proof.
Historical artifacts were not rewritten. Only this new report and two new tools
were added. Production/runtime files and `.DS_Store` were not changed by the pass.
No commit, push, tag, release or v0.1.3 change.

**DMO0 NOT CLOSED. DMO1 NOT STARTED.**
