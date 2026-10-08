# DMO0 timing oracle correction: FF-inclusive SCROLLE

2026-10-08. Owner approval: DMO-IMPL-2B7 ONLY.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

This correction supersedes the exact expiry/QUIT timing derived from the shortened
research buffers. It does not retract the proved first-equality reachability or
make a whole-DOS timing claim. Historical reports and private artifacts are kept;
their old dates describe the defective reference model, not actual demo text.

## Defect and reproduced cause

`audit_10min_demo_post_collision.configuration` measured the byte distance from
DS7313/7412 to the first FF: 98/88. Those distances exclude FF. Five research
constructors then used `make([]byte,n)`, omitting the terminator cell. Shared
matrix dispatch initializes `textLeft = len(text)-21`, so those buffers supplied
77/67 instead of 78/68. The error was in the fixture, not the working interpreter.

The pinned demo's buffers have 99/89 bytes including FF, with no earlier FF.
The linked demo consumer is checked mechanically at file offsets7206 (CX=2),
721a (ES:[SI+20]==FF),721f (continue drawing if unequal),72cc (shared phase
 decrement),72d6 (reset phase to8),72dd (increment SI),72f7/72fb (second subcall).
Each subcall tests FF before any glyph stores. Odd glyph stores precede even
glyph stores; selector changes precede phase decrement. A successful FF test
ends the routine immediately, including when it occurs on the second subcall.
Dispatch resets the text pointer/window, retaining the shared phase.

For every character used by both pinned texts, and the owned representation's
zero cell, configuration now verifies both glyph bodies as bounded literal MOV
stores followed by RET. No glyph-dependent branch or phase writer is admitted.
The private candidate additionally checks glyph correspondence to canonical A.
Thus glyph pixels vary with content, while logical visit cadence depends on
extent and shared phase. This is a logical calculation result, not a claim that
all glyphs take equal CPU time on DOS.

The corrected research representation owns `extent+1` cells and places FF at
index `extent`. No private text bytes are added to Git. A single helper supplies
all five research constructors. Actual private rendering remains in the existing
candidate adapter. `test_dmo0_scroll_extent.py` independently interprets SI+20,
two subcalls and phase1..8: adding one consumed character costs four visits in
every case. Both omitted characters therefore explain this fixture's eight-visit
error; this is not a general rule for arbitrary interrupted continuations.

## Authoritative reference basis

- `tools/audit_10min_demo_post_collision.py`: pinned program, source operands,
  FF extents, linked cadence anchors and bounded glyph-body validation.
- `tools/dmo0_post_collision_reference.go.txt`: FF-inclusive owned timing buffers;
  passive phase/admitted-visit observers in isolated research code.
- `internal/partyland/demo_timing_oracle_test.go`: independent SI+20 execution
  over actual private buffers, compared with candidate at every calculation,
  including every command boundary and restart36173. It calls neither StepScroll
  nor ScrollCompletes to decide completion.
- Existing `TestEverySourceScrollBoundaryAndInheritedPhase`: independent
  literal-store/VGA oracle versus StepScroll for four tables and phases1..8.
- `tools/audit_10min_demo_timing_correction.py`: only saved fixed input scripts,
  separate private output directory, accepted source base, exact saved prefix and
  first-equality comparison, concrete execution to QUIT. No search entry point.
- `tools/test_audit_10min_demo_timing_correction.py`: corrected artifacts, all
  command transitions, observed admission counts and retained pause theorem.

New report metadata uses `FF-inclusive-v1`. Historical artifact validators still
accept unmarked archived results under their original model; marking an old
artifact with the new basis does not validate its old QUIT date. New report
builders use corrected expectations. Archived tests are history regression,
not evidence of real-text timing. The new fixed replay/test pair is authoritative
for the corrected concrete continuations.

## Command boundaries

These are dispatch calculations; QUIT dispatch occurs on the last WAIT visit,
with no additional synthetic visit. A replacement at36173 runs before that
calculation's admitted matrix visit. The old scroll's phase is4, retained through
CLEAR4 and into the restarted scroll. Its old SI/cursor is not restored.

| Command dispatched | Uninterrupted | Restart36173 | Restart cumulative visits |
| --- | ---: | ---: | ---: |
| CLEAR4 | 35998 | 36173 | 1 |
| SCROLL1 (78 windows) | 36002 | 36177 | 5 |
| FLASHON1 | 36315 | 36488 | 316 |
| PRINT13_NUMBER | 36316 | 36489 | 317 |
| WAIT100 | 36317 | 36490 | 318 |
| FLASHOFF1 | 36417 | 36590 | 418 |
| SCROLL2 (68 windows) | 36418 | 36591 | 419 |
| FADE256 | 36691 | 36864 | 692 |
| WAIT100 | 36947 | 37120 | 948 |
| QUIT0 | 37047 | 37220 | 1048 |

Uninterrupted costs:5+313+1+1+100+1+273+256+100=1050.
Restart costs:5+311+1+1+100+1+273+256+100=1048.
The interpreter produces these totals; they do not drive execution.

## Affected exact calculations and retained verdicts

| Concrete case | Historical QUIT / visits | Corrected QUIT / observed visits | Result |
| --- | --- | --- | --- |
| Uninterrupted structural expiry35998 | 37039 /1042 | 37047 /1050 | Independent private oracle agrees with candidate |
| PARTY_ON_TASK1 equality, Release35842 | 37039 /1042 | 37047 /1050 | Fixed replay reconfirmed; reachability retained |
| SETBALL equality, Release35764 | 37039 /1042 | 37047 /1050 | Fixed replay reconfirmed; release/late physics retained |
| DROPTASK2 equality, saved TARGET | 37039 /1042 | 37047 /1050 | Fixed replay reconfirmed; HOLDSTILL preservation retained |
| DURINGFLASH equality, saved TARGET; scored drain36173 | 37212 /1040 after restart | 37220 /1048 after restart | Fixed replay reconfirmed; phase4 and real effect admission retained |
| Structural restart36173 | 37212 /1040 expected by old basis | 37220 /1048 | Queued restart, independent oracle agrees |
| NEW_BALL collision; no input; equality101534 | 102575 /1042 | 102583 /1050 | Fixed post-collision suffix reconfirmed; initial replacement retained |
| Post-collision single launch; unscored drain36288; equality101534 | 102575 /1042 | 102583 /1050 | Concrete suffix reconfirmed; PARTY_ON and second SETBALL retained |
| Post-collision periodic launch; scored drain37084 | 38125 /1042 | 38133 /1050 | Real scored admission/restart reconfirmed |

The first equality remains35998, wrap remains65536, repeated equality remains
101534. These dates are not shifted. Scored drains36173/37084 and task firing
calculations also stay fixed. The scored-drain35877 predecessor report itself
contains no exact terminal QUIT assertion: its drain, bonus producer35967 and
NEW_BALL replacement35998 remain established. Its later termination claims come
from the separate post-collision report and are corrected in the table above.

Affected historical reports: `runtime-layout-demo-10min-first-equality-party-on`,
`first-equality-setball`, `first-equality-droptask2`, `first-equality-duringflash`,
`post-collision-termination`, `dmo0-scope-decision`, the accumulated `validation`
report, and the requested timing expectations recorded by `dmo-impl-2b6`.
The corresponding `/private/tmp/pf-dmo0-first-equality-*.json` and
`pf-dmo0-post-collision-termination.json` retain their original values. Earlier
harness outputs/configurations containing truncated buffers are historical only.
No old artifact was overwritten by this correction.

FIRST_EQUALITY_COLLISION_REACHABLE and the four preservation/release class
reachability verdicts remain proved within their existing native-reference scope.
The exact old terminal dates are superseded, not silently reclassified as PASS.
POST_COLLISION_TERMINATION remains proved for the re-executed concrete scripts.
EVENTUAL_QUIT_NOT_GUARANTEED retains its closed frontend pause counterexample.
Universal active-only termination, EXPIRY_INTERLEAVING, untested input suffixes,
historical DOS/PIT phase and other DMO0 domains remain unproved. No timing is
assigned to an unreplayed continuation by adding eight to an old date.
