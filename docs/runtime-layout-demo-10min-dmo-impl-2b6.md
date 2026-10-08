# DMO-IMPL-2B6: complete isolated expiry consumers and terminal QUIT

2026-10-08. Owner approval: DMO-IMPL-2B6 ONLY.
Branch main, unchanged HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_EXPIRY_PROGRAM = PARTIAL

DEMO_TERMINAL_QUIT = READY

All ten commands execute, with actual private scroll data, dot writes, shared
flash/PRINTTASK, internal DAC fade and a typed normal QUIT(0). PARTIAL records
an explicit **FAIL of the requested 1042/37039 timing conformance**, not a missing
opcode: source-driven execution with the actual strings takes **1050 admitted
visits, QUIT37047**. Restart36173 preserves shared scroll phase and takes
**1048 visits, QUIT37220**, rather than the requested37212. No deadline or shorter
text was substituted to turn this mismatch into PASS.

## Concrete timing mismatch and its smallest boundary

Accepted DMO0 post-collision configuration (`tools/audit_10min_demo_post_collision.py`,
configuration) computes scroll lengths as the byte distance to FF, obtaining98/88.
The linked SCROLLE terminator test is byte[SI+20]==FF before each of two subcalls.
`tools/dmo0_post_collision_reference.go.txt`, researchStart, supplies
`make([]byte,n)` placeholder buffers of those lengths; it does not load the
private strings or their terminators. Shared matrixDispatch sets textLeft to
len(text)-21. Consequently that old fixture supplies77/67, whereas actual
FF-terminated private buffers contain99/89 bytes and correctly supply78/68.
Each omitted character costs four visits: the old reference is eight visits
short. This is a bounded inspection of the existing consumer/fixture, not a new
reachability, graph, trajectory or writer-domain investigation. Historical DMO0
reports and their research scripts were left unchanged; their1042 result is not
claimed as conformance evidence for real presentation.

Shared StepScroll's existing independent source/VGA oracle verifies the actual
SI+20 terminator, phase, selector and literal stores for all four canonical tables
and all initial phases1..8. A new assertion checks fallible completion lookahead
against that independent oracle, including termination on the second subcall.

| Consumed routine | Admitted calculations, inclusive | Visits | Next dispatch |
| --- | --- | --- | --- |
| CLEAR4 | 35998..36002 | 5 | SCROLL1 at36002 |
| SCROLL1, private98 bytes before FF | 36003..36315 | 313 | FLASHON1 at36315 |
| FLASHON WAITRUT | 36316 | 1 | PRINT13_NUMBER at36316 |
| PRINT WAITRUT | 36317 | 1 | WAIT100 at36317 |
| WAIT100 | 36318..36417 | 100 | FLASHOFF1 at36417 |
| FLASHOFF WAITRUT | 36418 | 1 | SCROLL2 at36418 |
| SCROLL2, private88 bytes before FF | 36419..36691 | 273 | FADE256 at36691 |
| FADE256 | 36692..36947 | 256 | WAIT100 at36947 |
| WAIT100 | 36948..37047 | 100 | QUIT0 at37047 |

Total1050. QUIT is terminal dispatch on the last WAIT visit, with no synthetic
visit. Tests check every command boundary and total count after execution; these
numbers never drive execution. The structural queued restart dispatches a new
entry before its matrix visit at36173. CLEAR4 is next1/remaining4 after that visit.
The old scroll had advanced half a glyph, so the new scroll inherits phase4 and
saves two visits. The old cursor never returns. This explains1048/37220 without
inventing protection or resetting phase.

The next minimal blocker is reconciling the requested timing oracle with real
FF-terminated text and the reviewed SI+20 consumer. Achieving1042 with these
operands would require truncation or a changed source consumer. Neither is
implemented. No DMO-IMPL-2B7 work is started.

## Presentation and terminal implementation

The pinned private adapter checks the entire existing demo identity before use,
then checks all ten linked handlers and operands at entry0x1ba17 through0x1ba3b.
SCROLL operands are DS7313/7412; PRINT13_NUMBER uses score DS18101 and position344;
FLASHON/OFF operands1; both WAIT operands100; FADE256; QUIT status0. The terminal
is reached before the trailing zero word is consumed.

Private text buffers include their real FF terminators. The demo glyph lookup
bases are canonical+0x100 and glyph code bases canonical+0x70. For every character
actually used by both texts, the adapter validates both complete glyph-store
bodies against canonical A, including every literal MOV store and RET. It then
uses the existing shared source-derived consumer and private text copies.
No fake strings, glyphs, frames or precomputed delay are present. Source font13
is checked by the existing private PARTY_ON presentation correspondence. Missing
or replaced pinned text/font operands reject before their dispatch.

StartMatrix/KillFlash and BeginResolved/Flash remain the shared primitives.
FLASHON speed/count1, enabled/phase=true is set at dispatch; Display.On is retained
until the next blink. Blink precedes tasks and remains independent of matrix
budget. FLASHOFF and replacement disable flashing and set On=true while retaining
the required count/speed/phase words. Tests cover these transitions, budgetless
blink and no extra visit. There is no local flash machine.

Shared PRINTTASK flushes after matrix interpretation. It reads live Score through
matrixNumber at that phase, including a score changed after equality. Tests cover
zero and1234567890, a separate changed live score987654321, and later score changes
that do not re-run the completed print. Expected dots use the shared font13 on
the prior dot-memory background. Source numeric conversion suppresses leading
zero digits; an all-zero value produces no digit stores, retaining the previous
background rather than fabricating an ASCII zero or a diagnostic string.

FADE is implemented only by the test adapter through a narrow optional shared
matrixConsumer extension. Dispatch captures its currently owned768-component DAC
palette. The declared structural initial palette is the current canonical table/
lamp/matrix presentation palette converted from expanded RGB to six-bit DAC; it
is an admitted presentation input, not a claim to have read a historical VGA DAC.
Each admitted FADERUT visit decrements SI and writes high-byte(snapshot*SI),
255 down through0. Every SI multiple of16 records the actual INT66 AL6/CX volume
request,240..0. Tests check all768 components at every visit and all16 requests.
The owned snapshot/output are value arrays, separate from native lamp palette.
Budget=false preserves SI, palette and requests. Replacement stops that routine
and releases candidate fade ownership without restoring an earlier palette/cursor.
The historical FADING ownership latch suppresses DOS DOLIGHTS packets; this
structural candidate owns the DAC output and has no host palette writer. Native
lamp bookkeeping can still advance independently and cannot overwrite that
output. No hardware DAC read/write, volume rendering, new host backend or full
historical palette chronology is claimed. The test candidate's internal palette
and volume requests are the explicit boundary to absent host presentation.

`structuralCalculation` returns `(*demoTerminal,error)`: supported QUIT yields
Status0, terminal calculation and visit count with nil error. Sticky unsupported
execution yields the existing demoUnsupported error and no terminal value.
QUIT disables the current matrix and subsequent candidate calculations return
before counters, tasks, audio, physics, score, display or trace writes. Tests
snapshot the complete candidate/game/physics/display and attempt subsequent core,
electronics, connected and install calls. No os.Exit, process termination, DOS
INTRO, selector or DOS teardown is involved.

## Scope and shared code

Concrete consumers and terminal implementation remain in build-tagged `_test.go`
files. Shared production changes are restricted to the unexported optional
matrixConsumer interface/slot and its dispatch/step hooks in partyland/timing.go,
plus a read-only ScrollCompletes lookahead in presentation/scroll.go. All ordinary
programs leave the extension nil and retain their existing dispatch/step behavior.
The existing independent temporal test gains one lookahead assertion. Production
profiles, oracle data, hosts and canonical programs are unchanged. Untagged Windows
and macOS engine artifacts contain no candidate implementation/error string.

Original-backed structural tests use pinned demo presentation and existing
canonical-A fixtures. They deliberately start at an admitted matrix state and
supply no reachability proof or fresh input-only demo witness. Fixture-free core
checks remain structural; tests needing private inputs skip when unavailable.
The connected candidate retains all area/target gates and the prior BYGEL12
refusal after real SETBALL late physics (TestDemoSetBallFirstEquality and child
sequence checks). No gameplay callback or canonical fallback is introduced. Due
consumer preflight and sticky rejection remain in place; no rollback is invented.

Tests cover actual dot changes in both scrolls; all opcode/WAIT boundaries; flash;
live and zero score; every fade component and volume request; typed terminal freeze;
budget/pause; replacement/restart in partial SCROLL/FADE/WAIT; inherited scroll
phase; direct replacement despite priority255/INH_EFF; missing pinned consumers;
and unknown NEXT_A before its dot/cursor writes. Direct DO_MATRIX receives no
synthetic cursor/priority protection. Gameplay paths that fail before expiry
remain unsupported.

## Validation and baseline comparison

Logs/artifacts: `/private/tmp/pf-dmo-impl2b6-*`. Original-backed checks use existing
private checkout `/private/tmp/pf-dmo-impl1-g75qk5z5` with current candidate/shared
files copied in. No missing fixture was created. A separate archive of unchanged
HEAD, `/private/tmp/pf-dmo-impl2b6-head`, uses the same canonical originals; every Go
source in that archive was compared byte-for-byte with git HEAD before baseline
checks.

| Check | Result |
| --- | --- |
| Tagged TestDemo, milestones1 through2B5 and new2B6 | PASS: 56 top-level tests, no skips |
| Required1042/37039 timing check | FAIL: actual1050/37047, explicit PARTIAL gate |
| Required restart36173 ->37212 | FAIL: actual1048 visits /37220, shared phase retained |
| Public fixture-free tagged TestDemo | PASS available tests; originals NOT AVAILABLE/skipped |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS:153 top-level tests; five opt-in capture/export skips |
| Updated independent four-table SCROLLE oracle, phases1..8 | PASS |
| OriginalTrajectories / Stones pf10 captures | NOT AVAILABLE; excluded, not generated |
| Focused A/B/C/D datalayout/frontend compatibility | PASS:21 top-level tests, no skips |
| Tagged vet partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 CGO-disabled executable and engine package | PASS |
| macOS c-shared engine | PASS |
| Broad macOS compile-only, current and unchanged HEAD | Same baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6Session, current and unchanged HEAD | Same baseline FAIL: score000002311040, ball2, ticks1200 |
| Public source checker / diff whitespace | PASS |
| Milestone source/report and built artifacts payload check | PASS: whole-file/nontrivial aligned4KiB samples; not all-substring proof |
| Fresh input-only candidate witness / host fade output / DOS teardown | NOT AVAILABLE; outside this milestone |

The three broader-frontend lifecycle failures from2B5 were each run against current
candidate sources and unchanged HEAD with the same fixtures:

| Test | Current / HEAD evidence | Classification |
| --- | --- | --- |
| TestNativeGameshowLifecycleAndPersistence | gameshow_test.go:53, native content, both FAIL | baseline |
| TestNativeSpeedDevilsLifecycleAndPersistence | speeddevils_test.go:36, native content, both FAIL | baseline |
| TestNativeStonesLifecycleAndPersistence | stones_test.go:47, native content, both FAIL | baseline |

These are actual baseline failures, not unavailable fixtures or regressions. Their
commands remain FAIL; no unrelated fix is made. Missing pf11.2 full-table and
pf8 BIOS-font fixtures from the older broad run were not generated.

Reproduction: run `go test -tags dmoimpl1 ./internal/partyland ./internal/physics
-run '^TestDemo' -count=1 -v` in the existing private checkout with
PF_10MIN_DEMO_DATA set. Compatibility uses the2B5 focused pattern
`TestPrivate|TestRuntime|TestHostsShareRuntimeBoundary|TestPowerPack|TestDeluxe|TestProfiles|TestPossessed`
and all four existing private installation environment roots. Reference checks
exclude only TestOriginalTrajectories and TestStonesSourcePhysicsAndTwoFlippers.

Earlier worktree edits are preserved. No canonical-A oracle, .DS_Store or v0.1.3
change; no commit/push/tag/release. No scored drain/bonus/progression, new gameplay,
interactive demo, profile registration, five witnesses, DMO0 research or DMO1.
