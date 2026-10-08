# DMO0 scheduler timing semantic confluence — 2026-10-07

**SCHEDULER_TIMING_CONFLUENCE = NOT_PROVED**

Classification: **REFERENCE_TIMING_REQUIRED**, specifically to preserve exact
animation logical progress for the unresolved L placements. This does not prove
both placements physically reachable, identify a reference CPU, or begin timing
research. It proves that treating those alternatives as equivalent is unsound.

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; committed research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`. Separate uncommitted pass, no production
changes. Prior source-recurrence JSON is unchanged. New artifact:
`/private/tmp/pf-dmo0-scheduler-confluence.json`; extractor:
`tools/audit_10min_demo_scheduler_confluence.py`, expected exit **2**.

## First and minimal native divergence: Case A

The comparison starts after the common outer P electronics/physics/task work,
before its `0x472a` budget decision. It holds the same source event identity L1,
AX, registration, enabled state, source publication order and logical state.
Synchronization is after outer P and L1 have returned, before another independent
event. Equal residual PIT/pending/countdown state at synchronization is an explicit
comparison premise, not a newly proved physical correlation.

A0 delivers L1 after outer P returns; A1 delivers that same L1 before `0x472a`.
AX(P)=0, AX(L1)=65535. Initial latch is set, ball/raster busy clear, rest busy set;
matrix busy and intflag are clear. An active nonterminal animation has frame timer
2 and nonzero SI. PRINTTASK is DUMRET. The witness selects keyboard zero, keyboard
routine already busy, redraw flags 3003/3007 clear and first task slot free.
These are authored semantic states, not an assertion of global state reachability.

The checked direct dependency is:

- `0x472a/0x472f`: TIME_LEFT false skips the matrix call at `0x4746`.
- `0x479d -> 0x47b2`: call animation processing. Its `0x47bc/0x47c0/0x47c4`
  reads DOTRUT/SISA and invokes the active routine. Producer `0x485d/0x4863`
  establishes handler CS:701e, file `0x731e`, and its SI.
- `0x7328/0x732b` establishes DATA2 DS=206c. `0x7330` decrements DS:042a.
  Nonzero branches through `0x7336 -> 0x73e9`, restoring SI and returning.
- A0 therefore has **DATA2:042a=1**; A1 retains **DATA2:042a=2**.

This countdown is the single reported native divergence. On the next common
matrix visit, 1 decrements to zero and enters frame selection at `0x7339`, whereas
2 decrements to 1 and retains the current frame. The frame index/load at
`0x7355/0x735a/0x735d` and eventual program completion depend on that distinction.
The admitted nonterminal path has no VGA input supplying this counter. The
native consumer is `partyland/timing.go` `_ANIMATION` ->
`presentation/matrix.go` StepAnimation -> tablelogic.Animation. Native frameTime
is logical state controlling frame/loop/program advancement, not disposable VGA
register state. Presentation pixels need not be compared to establish the result.

At synchronization, both executions have one outer ElectronicsCalculation and
one ball/physics update; the same completed gameplay, score/progression and
outer task effects; the same active program cursor, handler and PRINTTASK;
TIME_LEFT false; LAST_WAS_VB clear; and all TABLE1 busy flags clear, including
matrix busy. P increments SYNC once; both have the same INT_WAS_HERE fields.
L's scroll helper is mechanically checked to write only main DS:3000/2ffe;
its keyboard busy guard returns immediately, redraw/keyboard guards skip the
matrix-replacement paths, and its queue operation only adds the same KEYTASK to
the first free main-DS slot and increments DS:347b. Neither changes DATA2:042a.
This is the minimum task dependency, not reopened task-list writer closure.

AX(L1)=0 produces equal countdown 1 on both sides. Thus L crossing the decision
is not universally divergent, but AX=65535 supplies a persistent counterexample.
No projection preserving native animation duration can identify timer 1 with 2:
a common next visit distinguishes them. No additional PRINTTASK dependency can
collapse this checked nonterminal countdown difference. Future gameplay equality
is not claimed after the subsequent program/frame transitions diverge.

## Case B and tail commutation limits

B0 and B1 contain the **same next P0**, with equal AX and source event identity.
B0 runs outer tail then P0; B1 runs P0 at a boundary after `0x4758`, before RETF,
then completes outer tail. Synchronization follows both returns. L has already
cleared the latch; all busy flags were cleared before this tail. Under the same
complete native P transformation, each side counts P0 once: outer+next means two
electronics, two ball updates and two SYNC increments, not three. Both end with
LAST_WAS_VB set, busy flags clear, the same final P TIME_LEFT, task/program/matrix
state and callback return AX=12345.

All 38 tail instructions including RETF are decoded and classified in the
artifact. There are no explicit memory operands, native state reads or writes.
Three invocation-local saved words supply GC5/GC8/GC4/GC1/GC0/SEQ2 restoration;
push/pop pairs are local calling machinery. Register moves construct those six
VGA word outputs. The final AX=12345 is callback ABI state; RETF is return
machinery. Raw stack addresses are excluded from the projection.

A decoded operand interpreter checks every one of the 38 insertion boundaries
with an ABI-preserving callback macro that changes logical state once and
saves/restores the six VGA registers. All final projected states, saved-word
consumption and restored VGA values agree. The logical diamond is therefore
proved **for an equal P transformer independent of incoming VGA state**.

A full DOS P diamond is **NOT_PROVED**. P reads incoming VGA registers at
4591/459b/45a4/45ae/45b7/45c1 to save them. Tail restoration changes those readback
values. The save/restore macro checks do not prove that every opaque rendering
callee's native effects are independent of them. Empty native tail footprint
alone is insufficient to discard this dependency. This pass does not assert
that physical VGA interactions are a second proven native divergence or expand
into global callee/alias closure. The A countdown already refutes the requested
general confluence, independently of this remaining B dependency.

The limited diamond can be applied repeatedly to any finite, fixed callback word
under the same noninterference premise. Tests cover up to four P callbacks with
equal intervening L events. Adding an event changes the word; reordering P,L,P
into P,P,L changes latch admission and is not normalization. Additional physical
time permitting more callbacks is not an additional occurrence of the same P.
There is no unbounded scheduler or physical-stack theorem here.

## Architectural result and checks

No universal P/L normalization rule follows: L cannot move across `0x472a`
without preserving its budget and animation-step effect. No unresolved PIT or
countdown placement fact is promoted below the native boundary.

**PAIRED_SCHEDULER_CONTRACT = NOT_PROVED**

**NATIVE_AUDIO_BOUNDARY = NOT_PROVED**

- New pass: **24 PASS** (16 authored, eight private), including AX=0/65535,
  L decision ordering, same-P ordering, latch/counters/busy/matrix comparisons,
  every decoded tail boundary, finite repetition, native-tail-write mutation,
  persistent matrix suppression and operand/branch mutations.
- Entire current demo research suite: **214 PASS**. Historical 190 belongs to
  the previous pass and is not reused as the current count.
- Focused A/B/C/D datalayout/frontend/Stones: **PASS, zero SKIP**.
  `/private/tmp/pf-dmo0-confluence-go-tests-corrected.log`. First invocation failed
  to find Go on PATH; `.tools/go/bin/go` was used. Retained initial log is separate.
- Logs: `/private/tmp/pf-dmo0-confluence-{new-tests,all-tests,gate}.log`.
- `git diff --check` and raw/encoded payload scan: **PASS**. Scan metadata:
  `/private/tmp/pf-dmo0-confluence-payload-scan.json`. Whole-file containment plus
  aligned nontrivial 4 KiB samples is not an all-substring proof.

DMO0 NOT CLOSED. DMO1 NOT STARTED. No production/runtime changes, historical
JSON rewrite, commit, push, tag, release, v0.1.3 or .DS_Store modification.
