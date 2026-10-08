DMO0 NOT CLOSED. DMO1 NOT STARTED.

# Paired API11/API12 scheduler transition research — 2026-10-07

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`. Separate uncommitted pass.
Owner-local artifact: `/private/tmp/pf-dmo0-paired-scheduler.json`.
Extractor: `tools/audit_10min_demo_paired_scheduler.py`; expected exit **2**.
Historical audio-boundary/admission JSON is read, not rewritten.

This is a **partial transition certificate**, not a complete allowed trace
contract. The sole remaining scheduler dependency is **due-entry recurrence
while a callback is outstanding**: reachable next-record deadline / direct
countdown (including zero and 16-bit wrap), source enable state, and publication
stage must permit L then P before outer P returns. A branch permitting nesting
plus STI does not prove that the source supplies those due entries. The artifact
marks this edge UNKNOWN rather than inventing a nondeterministic hardware event.
No IVT, physical stack, arbitrary executable or whole-DOS closure is required or
claimed here. **NATIVE_AUDIO_BOUNDARY = NOT_PROVED** consequently remains unchanged.

## Source-derived transitions

All eleven identity-checked supplied modules pass the newly checked transition
prefixes. Prior callback roots, budgets and own-body counts are inputs to this
pass; they are not rediscovered. The models below start at a paired scheduler
source entry and explicitly require a due token. Authored tests supplying that
token establish conditional suffix behavior only.

| State / edge | Indexed constant and audio families | Direct audio family |
| --- | --- | --- |
| Before callback | Snapshot current record; program next deadline; advance cursor by 9, wrap at period record | Decrement 16-bit countdown; only zero reaches active/enabled guards; enable interrupts; select phase |
| Delivery | Record priority >= active priority; save old priority, publish new priority, produce AX, call selected pointer | Registered L: P publishes later phase and split countdown; L publishes primary phase and period-minus-split countdown; produce AX, call selected pointer |
| During callback | Cursor already advanced; L at 200 may enter P at 100; equal priority allowed; P at 100 inside L at 200 is dropped | Published next phase selects next callback; no indexed priority exclusion; another countdown expiry still required |
| After return | Restore saved active priority; do not restore cursor | Restore neither phase nor countdown |
| Drop | Priority-rejected record still advances cursor | Nonzero countdown or inactive/disabled source exits without delivering callback |
| Budget | NOSOUND/GUS AX=0; other six indexed drivers AX=0/65535 | AX=0/65535 |

NOSOUND anchors: advance `0x651`, wrap `0x65e`, STI `0x664`, priority comparison
`0x66b`, restore `0x68b`. Per-driver cursor and priority fields/anchors are in
the artifact. The compared direct prefix has a further non-atomic publication
window: INTERNAL STI `0xb16` precedes phase test `0xb1d`, primary phase store
`0xbc9` and countdown reload `0xbcf`. Later clears phase at `0xbf6` before reload
`0xc03`. ADLIB and THING pass corresponding operand-derived checks.
Therefore strict global P/L alternation cannot be inferred merely from two
pointer calls. The conditional direct model covers published states with the
registered pair; it does not silently collapse the earlier window. Without a
registered L, the primary branch reloads the whole period without publishing L.

The missing edge is a source-state recurrence edge, not a callback ABI return
frame. In the direct prefix, expiry at zero followed by an early active/enabled
exit has no normal phase reload; the next decrement wraps. Reload positivity,
registration and enabled continuity cannot be replaced with an assumed positive
period. Indexed next-record advancement likewise supplies a future deadline,
not a proof that it expires before the outstanding callback finishes. The new
checks certify suffix grammar, not complete writer sets or temporal correlation.

## TABLE1 projection and exact questions

Projection is ordinary gameplay, with the previously proved zero slowdown mask.
Attract mode and opaque task/area/PRINTTASK effects remain outside this model.
Every projection records callback kind, AX, the six requested latch/busy/enable
fields and TIME_LEFT. Matrix budget admission is a necessary condition; another
matrix busy guard can still suppress matrix work.

1. **Real nested L during unfinished admitted P: UNKNOWN.** The source suffix
   permits it after a due entry, in all three families; this is not a real-source
   witness.
2. **Families:** indexed priority permits L at active P, and direct published
   phase selects L. Whether their due recurrence realizes the interval remains
   open for each family.
3. **Latch clear:** conditional yes. An eligible returning L clears LAST_WAS_VB
   even while INSIDE_RESTOFVBLANK is set; rest busy is not its guard. Raster/ball
   busy, absent latch or disabled updates prevent the ordinary clear. A stopped
   L does not count as a release.
4. **Second P before outer return:** real feasibility UNKNOWN. After L returns,
   indexed restores outer priority 100 and permits equal P; direct has published
   primary phase. Neither suffix alone proves a second due event.
5. **If delivered:** while outer rest busy is set, latch-cleared P enters
   ball/physics when ball busy is clear. On ball return it clears ball busy,
   sees rest busy, sets latch, and returns without electronics. If latch is
   already set, it exits before ball; if ball busy is set, the ball guard stops
   it. While nested L is outstanding, indexed rejects P at the scheduler;
   TABLE1 itself has no primary INSIDE_RASTINT guard.
6. **TIME_LEFT overwrite:** conditional yes, before L's raster/latch/ball guards.
   An enabled but rejected L can overwrite it too. Enabled P also writes it
   before its latch/busy exits. Disabled callbacks exit before the budget write.
   There is no callback-local budget restore; outer matrix reads the latest
   shared value at `0x472a`.
7. **Electronics effect:** changing AX alone does not change the direct
   electronics admission/count/order: budget is tested after that work. In the
   rest-busy counterexample the event sequence is `[outer E]`, with nested ball
   work and zero nested E. This is a local conditional statement, not a theorem
   about opaque matrix continuation or future gameplay. A distinct tail case
   matters: `0x4753` clears rest busy **before** return `0x479a`. After eligible L,
   P delivered in this tail can reach a second E: `[outer E, nested E]`. Its
   feasibility remains UNKNOWN. An “unfinished P” is therefore not sufficient
   to prove a single E across its complete nested interval.
8. **Family invariance:** not proved. Constant budget forbids nonzero-AX traces
   admitted by audio predicates; direct phase lacks the indexed priority drop.
   These are source predicate differences, not demonstrated hardware-run
   divergences. The artifact does not label conditional projections as real
   reachable traces.
9. **Native minimum:** no certified minimum yet. Candidate semantic state must
   retain delivery/dropping, outstanding callback stage, latch/busy fields and
   shared budget; indexed cursor/priority or direct countdown/phase/publication
   must first be reduced under the missing recurrence theorem. A serialized
   native schedule is not justified by this pass.
10. **Mechanical invariant:** none established across families. The weaker
    model invariant is that changing only AX cannot change a local P's
    electronics guard outcome; it changes shared matrix budget. Authored tests
    cannot upgrade that invariant to real feasible trace equivalence.

## Minimal pause/resume case

With latch set, disabled updates cause delivered P/L to exit before budget
updates or ordinary latch clear. Resume's enable store retains the latch. If
first feasible delivery is P, it writes budget and takes the latch exit. If L
arrives first with ball/raster guards clear, it releases the latch and the next
admitted P may calculate. Both are conditional continuations; the first real
post-resume ordering depends on the same due-entry recurrence and retained
record/countdown/enable state. No audio restoration or pause UI audit is claimed.

## Validation

- New-pass private tests: **16 PASS** (12 authored semantic tests, four pinned
  integration/mutation tests); `/private/tmp/pf-dmo0-paired-new-tests.log`.
- No-private run: **12 PASS**, private class SKIP before Capstone;
  `/private/tmp/pf-dmo0-paired-public-tests.log`.
- Entire current demo research suite: **161 PASS** including these 16;
  `/private/tmp/pf-dmo0-paired-all-tests.log`.
- Focused A/B/C/D datalayout/frontend/Stones regressions: **PASS, zero SKIP**;
  `/private/tmp/pf-dmo0-paired-go-tests.log`. The first shell invocation lacked
  `go` on PATH; the verified run uses the existing `.tools/go/bin/go`.
- `git diff --check`: **PASS**; staged list empty.
- Three new research files: raw and encoded payload checks **PASS**. Checks
  compare whole private files/decoded SDR and 48,898 distinct nontrivial aligned
  4 KiB samples; long hex/base64 tokens are decoded and checked. This is a sampled
  copy check, not an all-substring proof. Metadata:
  `/private/tmp/pf-dmo0-paired-payload-scan.json`.
- Whole-DOS historical gate remains separate, unchanged: TABLE1 **30 total /
  18 bounded / 12 UNKNOWN**; CODE2 **2 total / 0 bounded / 2 UNKNOWN**, exit **2**.
  No separate whole-DOS closure rerun was performed outside the required suite.
Production/runtime/internal/hosts/cmd are unchanged by this pass. No commit,
push, tag, release, v0.1.3 modification, DMO1 or .DS_Store action.

**PAIRED_SCHEDULER_CONTRACT = NOT_PROVED**
