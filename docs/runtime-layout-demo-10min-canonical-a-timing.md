# DMO0 canonical-A timing inheritance — 2026-10-07

**CANONICAL_A_TIMING_INHERITANCE = PROVED**, under canonical A's existing
native semantic reference. **NATIVE_AUDIO_BOUNDARY = PROVED** under that same
reference and the supplied SDR service semantics. This is not whole-DOS closure,
a demo implementation, or proof of historical behavior on every DOS machine.

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`. Canonical A remains the sole strict
production oracle. Prior artifacts and their historical verdicts are unchanged.
Owner-local result: `/private/tmp/pf-dmo0-canonical-a-timing.json`.

## Checked linked correspondence

The extractor checks every instruction/operand in these bounded blocks against
reviewed relocation maps. Constants, register identities, operand widths/access,
addressing modes, branch destinations, and call placement must agree. These maps
establish correspondence, not indirect target-domain or writer/alias closure.
Gameplay callees remain paired returning effects outside this narrow pass.

| Slice | Canonical A file interval | Demo file interval | Instructions | Classification |
| --- | --- | --- | --- | --- |
| Primary callback | `44c1..4747` | `4517..479d` | 288 | RELOCATED-IDENTICAL |
| Later callback | `58d6..5a23` | `592b..5a78` | 150 | RELOCATED-IDENTICAL |
| Matrix dispatch | `4747..47d6` | `479d..482c` | 52 | RELOCATED-IDENTICAL |
| Active animation producer | `47f0..4814` | `4846..486a` | 9 | RELOCATED-IDENTICAL |
| Animation routine | `72b6..7390` | `7324..73fe` | 106 | RELOCATED-IDENTICAL |
| Callback registration | `62cb..62f4` | `6345..636e` | 20 | RELOCATED-IDENTICAL |
| Physics boundary (tilt/ramp/level/drain) | `5ca5..5cd4` | `5d1f..5d4e` | 21 | RELOCATED-IDENTICAL |
| Common electronics suffix | `5c86..5ca5` | `5d00..5d1f` | 11 | RELOCATED-IDENTICAL |

Intervals have exclusive ends. No TIMING-SEMANTIC-DIFFERENCE was found in this
slice. Primary AX zero/nonzero is stored before latch/ball guards; later does the
same before its raster/latch/ball guards. Both callbacks share TIME_LEFT
(A DS:36f4, demo DS:37f1). LAST_WAS_VB is A CS:4446/demo CS:449c. Completing ball
work publishes the latch and rest guard; UPDATE_COUNTERS precedes electronics;
electronics/tasks precede the budget read and matrix visit; rest busy clears
before RETF. Later clears the latch only on its admitted completion path.
Normal callback return AX=12345 is common.

Animation DEC at A `72c2`/demo `7330` uses the identical DATA2:042a word. The
zero branch, old-BX frame selection/reload, loop boundary, and routine completion
are common. Relocating DATA2's segment does not change those countdown semantics.
Matrix dispatch invokes DOTRUT/SISA before PRINTTASK in both linked binaries.

## Supplied SDR correspondence

Nine of eleven packed SDR binaries compare equal directly. PAS16.SDR and
SB16.SDR differ; their differences are not represented as complete identity.
Their API11/API12 local registration CFGs and indexed scheduler suffix/full
budget helper are operand-matched. Code shifts by two bytes, the priority slot
moves with it, and PAS16's data segment relocates by a paragraph; fixed data
operands and budget constants remain fixed. Low shared return services do not
move. Registration service calls are paired edges; this does not reopen IRQ or
vector closure.

TABLE1 passes API11 primary DX:ES with BL=100 and API12 later DX:ES with BL=200,
CX=264/174. The common driver semantics retain cursor advance/wrap, priority
save/set/restore, AX=0/65535 budget production, and AX=12345 return handshake.
Identical modules include both the direct and indexed scheduler families. The
families need not become equivalent to each other: each demo driver inherits its
corresponding A service contract.

## Existing chosen native A schedule and evidence

`partyland.Game.Sync` supplies budget=true. `SyncWithMatrixBudget` captures one
explicit deterministic budget for the logical update. Tracker/PCM and logical
jingle callbacks run first. Two early physics passes precede the drain decision;
counters/areas precede targets and shift/keyboard work, matrix blink, tasks,
spring and lamps. The budget-gated matrix visit runs through BeforeLate, before
scroll and the late physics pass (two for fast-ball). Device performance never
rewrites TIME_LEFT during a native update. Runner retains every due source tick
and submits ordered PCM before returning to presentation; table cadence is 71 Hz.

Already-lost balls run audio, blink, tasks, lamps and matrix; an early physical
drain returns from physics before counters/areas and runs the loss task/matrix
branch. A native GameOver branch ends gameplay work. These special branches are
part of the extracted existing A schedule, not proposed changes. Demo timer
accounting must remain at its logical admitted-electronics position after the
ball/drain decision and before common tasks/matrix, including the proved demo
post-drain count; it must not be guessed from Runner.Ticks or physics substeps.
This pass does not implement that accounting or close expiry/task reachability.

Strict original identity is checked by `testinputs.Require` → `oracle.Verify`.
The following evidence constrains the chosen reference schedule:

- `TestPF45IndependentMatrixTraces`: independent PLAND data/TABLE1.MOD generator
  (`PF_REFERENCE_NO_NATIVE=1`) checks exact HIDDEN/MYSTERY/HAPPY/MEGA command ticks.
  The expectation generator observes source data and MOD control, not native
  output or a historical IRQ waveform.
- `TestEveryOriginalAnimationDrawSync` and `TestAllSourceMatrixRoutineBoundaries`:
  original-backed frame/duration checks, zero/wrap/old-BX loops and exact routine
  completion visits. They make a missing matrix visit observable.
- `TestPF45MissedMatrixBudget`: ten missed scans leave matrix state unchanged;
  task completion still occurs at tick 2; later successful scans determine time.
- `TestPF45MatrixBeforeLatePhysics`: matrix expiry precedes late physics.
- `TestPF45TaskSlotAndSharedWaitOrdering`,
  `TestPF45PendingModeStartsAfterNextTaskScan`, and
  `TestPF45IndependentDrainToNextBallTrace`: task scans, next-scan activation and
  original-data drain/bonus/next-ball checkpoints constrain ordering.
- Source Runner tests retain all due updates/PCM, including deadlines becoming due
  during the PCM sink, and verify 60/71 Hz frontend/table cadence.

These are validations of A's chosen semantic schedule. They do not demonstrate
that no other physical DOS schedule exists, and do not themselves validate the
unimplemented demo's gameplay, expiry, persistence or presentation.

## Exact Case A projection

A's linked structure admits the same abstract ambiguity: P AX=0 admits matrix;
the same completing L AX=65535 can overwrite shared TIME_LEFT before the decision.
With an active nonterminal countdown 2, matrix-before-L produces 1; L-before-matrix
retains 2. A next common visit distinguishes frame selection. No demo-specific
callback or animation difference creates this ambiguity.

The existing native A reference chooses the matrix-before-later-work side (A0):
one serialized primary logical visit, stable supplied budget, matrix step before
late work. It does not admit asynchronous L overwrites of the captured budget.
Exact original-backed matrix traces/durations and explicit missed-budget tests
validate this reference visit contract. The choice is admissible in both matched
DOS structures. Inheritance preserves that choice without requiring the alternate
placement to be physically impossible. Arbitrary authored DOS states are not
claimed globally reachable.

## Demo interference and proof boundary

All inspected demo differences are DEMO-SPECIFIC-BUT-NOT-TIMING for this reference
slice, with their distinct state effects retained:

- The word timer increments inside DO_ELECTRONICS before the common suffix. Its
  equality branch at 35998 installs expiry state/program; the ordinary unequal
  branch skips all replacement actions and reaches existing tasks/matrix order.
- Normal-ball continuation replaces canonical player/ball progression and queues
  NEW_BALL_TASK through the existing matrix/task machinery. It changes progression
  and may replace programs, but does not move the task scan or budget decision.
- Expiry executes at the existing electronics position. Interaction with pending
  NEW_BALL/SETBALL/effect replacements remains an independent consumed-state gate.
- Real PLAYERSTEXT print position 336 (A 340) changes presentation content/placement,
  not the callback, budget or active animation countdown implementation.
- S_EMPTY=(62,0,0), S_GAMEOVER2=(13,0,255), versus A priority 1: retain typed cue
  priorities/readiness/program-duration effects. They do not select a host-speed
  budget policy or permit asynchronous native callback preemption.
- Demo INTRO is a separate frontend/table-process mode. TABLE1 installs its own
  checked callback pair; advertising/menu work is not inside its update slice.

Historical PIT due recurrence, CPU-dependent L placement, nested callback timing,
VGA restoration/readback timing and physical audio-budget phase are
**BELOW-NATIVE-REFERENCE-BOUNDARY** for inherited native semantics. Existing
whole-DOS NOT_PROVED results remain valid at the historical boundary. No reference
CPU, hardware timing reconstruction, DMO1 or runtime implementation follows.

Remaining DMO0 blockers excluding audio/scheduler timing: consumed TABLE1/CODE2
indirect and writer/alias closure; expiry versus pending task/effect replacement
reachability; negative persistence and alternate termination/control paths; final
consumed-control differences and presentation coverage. Whole-DOS remains exit 2:
**DMO0 NOT CLOSED. DMO1 NOT STARTED.**

## Current-pass verification

- New pass: **24 PASS**; complete demo research suite: **238 PASS**, zero SKIP.
- Relevant strict A tests: **24 top-level PASS / 347 including sub-tests**, zero SKIP.
- Focused A/B/C/D: **15 top-level PASS / 99 including sub-tests**, zero SKIP.
- Whole-DOS gate: expected **exit 2**, TABLE1 18 bounded/12 UNKNOWN, CODE2 0/2.
- `git diff --check` and raw/encoded payload-copy scan: **PASS**. Payload scan
  checks whole files and aligned nontrivial 4 KiB samples, not every substring.

Logs and initial-run limitations are recorded separately in the JSON. The first
regression invocation swapped C/D fixture directories; the corrected private run
passes. A broad oracle run included one unrelated unavailable live Gear capture;
the required focused private run has zero SKIP. A public-checkout focused run
without root fixtures is NOT AVAILABLE and is not used as oracle evidence.

New-pass counts are separate from the full demo suite. No production/runtime files, historical artifacts,
.DS_Store, v0.1.3, commit, push, tag or release were changed.
