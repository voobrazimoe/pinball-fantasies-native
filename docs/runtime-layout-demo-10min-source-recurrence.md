# DMO0 source due-entry recurrence — 2026-10-07

**SOURCE_DUE_RECURRENCE = NOT_PROVED**

One narrow uncommitted pass over supplied SDR recurrence producers and the
TABLE1 post-rest tail. Production base is
`306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD is
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
Previous paired and audio-boundary artifacts are read as premises, never rewritten.
New artifact: `/private/tmp/pf-dmo0-source-recurrence.json`.
Extractor: `tools/audit_10min_demo_source_recurrence.py`, expected exit **2**.

The one smallest missing fact is the **reachable joint PIT-time/TABLE1 execution
relation: residual source time after eligible nested L, at outer `0x4758`, and
through `0x479a`**. This is a correlation obligation, not an unknown IRQ producer.
It must account for pending IRQ0, countdown/edge phase and interrupt acceptance
at actual TABLE1 stages. A generic positive reload or a bound measured from L's
publication cannot establish a positive residual when outer P reaches its tail.

## Actual recurrence producer

All eleven pinned SDR modules pass the new operand checks. IRQ0 from PIT channel
0 enters the supplied INT08 handler. Audio device buffer completion does not
produce this scheduler entry; the previously established audio predicate affects
callback AX. The checked vector writes are only the minimal supplied source
installation, not complete IVT writer admission.

| Family | New source event | Publication and recurrence units |
| --- | --- | --- |
| NOSOUND, GUS | PIT0 mode 0 terminal OUT rising edge -> PIC IR0 | Command `0x30`; low/high count at port `0x40`; first record loads `[SI+2]`; others load `u16([SI+2]-[SI-7]+latched_PIT0-10)` |
| PAS16, SB16, SB20, SBLASTER, SBPRO, SM2 | Same PIT0 mode 0 source | Same checked deadline arithmetic and IRQ-atomic PIT write; budget class does not change the producer |
| INTERNAL, ADLIB, THING | PIT0 mode 3 periodic rising edge -> PIC IR0 | Command `0x36`; divisor `floor(0x1234dc/sample_rate)`; **serviced IRQ0 entries**, rather than elapsed edges, decrement the scheduler word |

Mode 0 does not generate another terminal rising edge without rearming. Hardware
count word zero means 65536, not immediate expiry. PIC masking/IF disable can
retain a pending request; periodic edges can coalesce in one pending bit. Therefore
an elapsed hardware edge is not necessarily another serviced countdown decrement.
These hardware interpretations use the primary
[Intel 8254 datasheet](https://www.cs.usfca.edu/~cruse/cs210s07/8254.pdf) and
[Intel 8259A datasheet](https://www.pcjs.org/documents/datasheets/intel/INTEL_8259A_PIC.pdf),
especially PIC fully nested mode, EOI and IMR descriptions on printed pages 14–15.
The normal DOS PIC configuration is the hardware context here; no alternative
PIC programming or arbitrary external executable is admitted.

Indexed handlers issue master EOI and STI **before** SI snapshot and rearm.
Direct handlers issue EOI before countdown decrement, and STI after zero and
active/enabled guards. Under the normal unmasked IRQ0/IF=1 context, the current
ISR0 bit does not exclude a subsequent IRQ0 while callback execution is outstanding.
The PIT continues running independently of TABLE1. This establishes reentry
permission and its producer, not a deadline crossing at a particular instruction.
Each artifact row records the actual installation, EOI, publication and state
field offsets; the priority/callback suffix comes from the previous pass.

GUS has a distinct checked first-record window: `0x6e6–0x6eb` saves IMR and ORs
`0xfd`, masking IRQ0 during raster wait/rearm; `0x717–0x718` restores saved IMR
before cursor advancement and callback. PIT time continues during that window.
The restored mask can expose a pending request. The other indexed prefixes have
no corresponding IMR write. None supplies a temporal tail certificate.

The indexed SI snapshot belongs to the current invocation. Published PIT state
and the advanced global cursor survive nested callbacks; the previous active
priority is restored. PIT rearm itself precedes cursor/priority publication,
while IF is already enabled: expiry in that smaller publication window is also
subject to the same missing clock correlation. The extractor does not claim
that this sequence is globally atomic.

## Direct recurrence, INTERNAL first

INTERNAL's source is installed at decoded `0x1554`, entry `0xac9`; PIT mode 3 is
programmed at `0x14ae` and rearmed at `0x1568`. Divisor DS:`0x66c5` is computed
from source rate DS:`0x78b`. The sample cursor wraps through `0xc2a -> 0xc2e ->
0xaee`; both cursor branches rejoin the **same** countdown decrement. Only this
minimum prefix was inspected; audio rendering is outside the pass.

The precise shared producer prefix is mechanically propagated to ADLIB/THING:

- Each serviced IRQ0 decrements CS countdown modulo 65536. Nonzero returns.
- Zero plus inactive returns retaining zero; the next decrement wraps to 65535.
- Zero plus disabled source reloads **50**, then returns. This differs from inactive.
- Zero plus active/enabled reaches STI, phase selection and callback publication.
- P reloads split when L is registered, otherwise whole period. L reloads
  `u16(whole_period-split)`. No positivity or non-underflow assumption is added.
- A positive reload N needs N serviced IRQ0 entries; zero needs 65536. Serviced
  entries can continue while a callback is outstanding if IF/IMR permit them.
- Phase and countdown are retained across callback return. An IRQ between phase
  publication and reload uses the old countdown. Immediately after expiry that
  word is zero; an intervening raw IRQ wraps it, without a paired callback.
  The outer reload can then overwrite that decrement. Merely publishing phase
  does not create a due callback, and pointers do not prove strict alternation.

INTERNAL's STI `0xb16`, P phase/reload `0xbc9/0xbcf`, and L phase/reload
`0xbf6/0xc03` are recorded as separate stages. ADLIB/THING rows preserve their
own offsets, source divisor programming, decrement, zero path and period fields.
Initial whole period is `floor(sample_rate/100)`; P's raster correction can alter
it. API11 raster-line measurement supplies the split relation. This is source
calibration, not elapsed TABLE1 instruction timing.

## Critical TABLE1 interval and temporal result

The identity-checked `0x4753 -> 0x479a` path has **39 instructions including clear
and RETF**, **38 after completion of clear including RETF**, or **37 before
starting RETF**. There is exactly one own-body path: no conditional branches,
calls, loops, interrupt mask change or source reprogramming. Six word OUTs
restore VGA graphics/sequencer state. Instruction count is not elapsed PIT time:
CPU/memory and VGA I/O transaction duration have no established upper/lower bound
in the supplied calibration.

The smallest semantic window begins at `0x4758`, immediately after completion
of the busy clear, and ends when P returns at `0x479a`. An already-published next
P can enter there only if a source event becomes pending and is accepted before
return. For direct, enough **serviced** periodic IRQs must additionally bring the
word to zero. The residual source state on reaching this window is not derived.
No source clock is sampled in this tail, and there is no hard ordering instruction
that excludes entry throughout it.

`TIMER.BIN` was identity-checked: PIT2 measures a 175-iteration VGA-store loop
and a 2000-iteration multiply/memory loop, then converts elapsed count into five
speed classes using thresholds `0x1a00/0x2500/0x3400/0x4400`. That coarse benchmark
neither measures these six OUTs nor bounds the residual deadline upon reaching
the clear. SDR raster/PIT calibration supplies units, not a callback execution
bound. No historical CPU speed, VGA wait state model or arbitrary IRQ placement
was invented. Hardware variability alone is not proof that both semantic outcomes
are reachable, so no nondeterministic native contract is certified either.

## Critical trace ledger

These are actual source obligations with **UNKNOWN feasibility**, not injected
real traces. R0/R1/R2 below denote the successive current records, not assumed
reachable numeric positions among nine slots.

| Stage | Indexed obligation | Direct obligation |
| --- | --- | --- |
| Outer P | R0 kind P; rearm successor; advance cursor; active priority 100 | expiry 1->0; active/enabled; STI; phase P; publish L then split reload |
| Nested L | real next terminal edge must enter while P outstanding; R1 must be L; 200 admitted; rearm and advance | split-effective count of serviced IRQ0 entries must expire while P outstanding; phase L; publish P and period-minus-split |
| L return | restores active 100; cursor/PIT state remain advanced | countdown/phase remain published |
| No following P | source not yet due, masked, or following record differs: no P entry inferred | not enough serviced IRQ0 entries: nonzero exits; no P entry inferred |
| Nested P while rest busy | another real terminal edge; R2 must be P; equal 100 admitted after L return | another countdown expiry must select P before clear |
| Nested P after clear | same recurrence requirement, but acceptance must lie between `0x4758` and return | same expiry requirement in that smaller window |

`P -> nested L -> outer return`, `P -> nested L -> P while rest busy -> outer
return`, and `P -> nested L -> P after clear -> outer return` all remain UNKNOWN
for all three families. A P record rejected while indexed L is still active
nevertheless advances/rearms; its later return does not undo that state.
The local effects after delivery are earlier premises and were not rediscovered.
The artifact lists source-absent/masked/nonzero impossibilities separately from
unknown real trace feasibility. No pause/resume audit was needed.

## Validation

- Current-pass tests: **29 PASS**: 21 authored source/state cases, eight pinned
  producer/tail mutation tests. `/private/tmp/pf-dmo0-recurrence-new-tests.log`.
- Entire current demo research suite: **190 PASS**, including those 29.
  `/private/tmp/pf-dmo0-recurrence-all-tests.log`. Historical 161 is a prior count.
- Focused A/B/C/D datalayout/frontend/Stones regressions: **PASS, zero SKIP**.
  `/private/tmp/pf-dmo0-recurrence-go-tests-corrected.log`. Initial invocation
  used a GOG container directory without PRGs and failed; retained log is
  `/private/tmp/pf-dmo0-recurrence-go-tests.log`. Only the fixture path was corrected.
- Source gate: **exit 2**, own verdict NOT_PROVED; historical whole-DOS result
  remains a separate TABLE1 30/18/12, CODE2 2/0/2, exit 2 input.
- No-private new pass: **21 PASS**, private class SKIP before Capstone import.
  `/private/tmp/pf-dmo0-recurrence-public-tests.log`.
- `git diff --check`: **PASS**; staged list empty, production path diff empty.
- Raw/encoded payload check: **PASS** for all 23 currently changed/new research
  files plus this pass's owner-local artifact, against 249 private files, eleven
  decoded SDR and 48,898 distinct nontrivial aligned 4 KiB samples. Metadata:
  `/private/tmp/pf-dmo0-recurrence-payload-scan.json`. Whole-file containment plus
  sampled copy check is not an all-substring proof.

**PAIRED_SCHEDULER_CONTRACT = NOT_PROVED**

**NATIVE_AUDIO_BOUNDARY = NOT_PROVED**

No minimal native scheduler/audio contract can be promoted from source permission
without the missing temporal relation. DMO0 NOT CLOSED. DMO1 NOT STARTED.
No production/runtime/internal/hosts/cmd changes, whole-DOS closure expansion,
commit, push, tag, release, v0.1.3 or .DS_Store modification belongs to this pass.
