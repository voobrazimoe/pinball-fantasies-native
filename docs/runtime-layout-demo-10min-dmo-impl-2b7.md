# DMO-IMPL-2B7: reconcile the FF-terminated timing oracle

2026-10-08. Owner approval: DMO-IMPL-2B7 ONLY.
Branch main, unchanged HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_TIMING_ORACLE = RECONCILED

DEMO_EXPIRY_PROGRAM = READY

DEMO_TERMINAL_QUIT = READY

These verdicts cover the accepted isolated expiry candidate and the corrected
native-reference timing oracle. READY does not register demo support or prove
whole-DOS closure, gameplay completeness, host fade output or DOS teardown.

The independent reference agrees with the real-text candidate:1050 admitted
visits / QUIT37047 from35998, and1048 / QUIT37220 after restart36173. The shared
matrix and physics implementation needed no correction. The 2B6 PARTIAL timing
gate is resolved by repairing the research fixture that omitted FF, with evidence
from pinned linked consumers, private texts and source-driven execution.

See [the correction record](runtime-layout-demo-10min-timing-oracle-correction.md)
for root cause, all command boundaries, affected reports/artifacts, historical
versus corrected dates, retained verdicts and authoritative reference basis.

## Changes

The research configuration validates linked SCROLLE's two subcalls, SI+20 test,
phase decrement/reset and SI advance. It validates the actual text extents and
bounded literal MOV/RET glyph bodies, including the owned representation cell.
The common research helper now supplies extent+1 cells ending in FF. The old
77/67 window counts become78/68; no private text is copied into tracked files.
New report builders mark the corrected `FF-inclusive-v1` basis. Unmarked archived
artifact validation remains explicitly historical.

A new independent test runs literal SI+20 logic on actual private texts, alongside
the candidate, at every calculation and command boundary. The existing independent
four-table VGA/StepScroll oracle still passes. A fixture-free oracle tests every
phase1..8 and second-subcall completion; one extra character costs four visits.

A dedicated fixed replay driver reconstructs the accepted research source base
`306d11a0c479c7ac5ee6e245f3f72eacbc665abd` in a private snapshot, preserving the
existing source identity checks. Later candidate worktree changes are not used
as a replacement research baseline. It runs only saved scripts, compares every
saved prefix and first-equality row/boundary, and passively records scroll phase
and actual admitted visits after budget/active checks. Each real expiry install
resets that observation counter; it does not control termination. The DROPTASK2 comparison reconstructs its existing derived `linked_HOLDSTILL`
report field from expired/ball.Hold; it does not seed runtime state. Existing fixed
handoff checks also revalidate scored drain35877 and NEW_BALL equality replacement.

## Previously asserted exact QUIT calculations

| Case | Historical | Corrected | Current timing verdict |
| --- | --- | --- | --- |
| Uninterrupted expiry35998 | 37039 /1042 visits | 37047 /1050 | RECONCILED, private independent oracle |
| PARTY_ON equality | 37039 /1042 | 37047 /1050 | Fixed input-only replay reconfirmed |
| SETBALL equality | 37039 /1042 | 37047 /1050 | Fixed input-only replay reconfirmed |
| DROPTASK2 equality | 37039 /1042 | 37047 /1050 | Fixed input-only replay reconfirmed |
| DURINGFLASH; scored-drain restart36173 | 37212 /1040 after restart | 37220 /1048 after restart | Fixed input-only replay; phase4 retained |
| Structural queued restart36173 | 37212 /1040 old expected basis | 37220 /1048 | Independent oracle agrees |
| NEW_BALL collision; no input; repeated equality101534 | 102575 /1042 | 102583 /1050 | Fixed concrete continuation reconfirmed |
| Post-collision single launch; unscored drain36288 | 102575 /1042 | 102583 /1050 | Fixed concrete continuation reconfirmed |
| Post-collision periodic launch; scored drain37084 | 38125 /1042 | 38133 /1050 | Fixed real admission/restart reconfirmed |
| Scored-drain35877 predecessor report itself | No exact QUIT asserted | No date invented | Drain/bonus/task provenance retained |

All five existing first-equality reachability verdicts survive the prefix and
full-calculation comparison. POST_COLLISION_TERMINATION remains proved for the
replayed scripts; EVENTUAL_QUIT_NOT_GUARANTEED retains the actual closed frontend
pause-cycle proof. EXPIRY_INTERLEAVING remains NOT_PROVED. Old exact dates are
superseded, while their original reports/artifacts remain preserved. No date was
corrected solely by adding eight, and no untouched suffix receives a new proof.

## Validation

Private logs/artifacts: `/private/tmp/pf-dmo-impl2b7-*`; final fixed replay outputs
under `/private/tmp/pf-dmo-impl2b7-final`. Candidate originals use the pre-existing
`/private/tmp/pf-dmo-impl1-g75qk5z5` checkout with current Go sources copied in.
Existing private installation roots were verified by the pinned consumers; no
missing fixture was created.

| Check | Result |
| --- | --- |
| Independent FF/SI+20 oracle, private uninterrupted and restart | PASS, every calculation and command boundary |
| Corrected fixed PARTY_ON/SETBALL/DROPTASK2/DURINGFLASH/post continuations | PASS, observed visits and saved prefix/equality comparisons |
| Interrupted scroll and real restart phase | PASS, phase4 survives restart; second scroll phase8 |
| DMO-IMPL-1..2B6 tagged TestDemo plus new oracle | PASS:57 top-level tests, no skips |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS:153 top-level tests; five opt-in export/capture skips |
| Four-table independent SCROLLE/VGA oracle, phases1..8 | PASS |
| A/B/C/D datalayout/frontend focused compatibility | PASS:21 top-level tests, no skips |
| Tagged vet partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 CGO-disabled executable / internal engine package | PASS |
| macOS c-shared engine | PASS |
| Broad macOS compile-only | Known baseline FAIL:AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL:score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL:gameshow:53, speeddevils:36, stones:47, native content |
| Public fixture-free tagged TestDemo | PASS available tests; original-backed tests SKIP/NOT AVAILABLE |
| Python DMO0 regression suite including corrected tests | PASS:396 tests,20 skips; archived timing stays historical |
| Corrected fixed-witness tests / fixture-free extent oracle | PASS:5 /2 tests, no skips |
| Public source checker / diff whitespace / scoped payload scan | PASS; payload scan covers whole files and nontrivial aligned4KiB samples, not all substrings |
| OriginalTrajectories / Stones pf10 capture tests | NOT AVAILABLE; excluded, no fixtures generated |

The known baseline failures match the accepted 2B6 current/unchanged-HEAD evidence;
none was repaired or relabeled PASS. Initial compatibility attempts with C/D
roots interchanged failed identity checks; correcting the environment mapping
made the focused suite pass. An initial Windows command targeted the host-only
`cmd/pfengine` with CGO disabled and was unavailable by build constraints; the
required internal engine package and Windows executable build both passed.

Reproduce the correction with the three existing private paths:

```
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> \
PF_DMO0_HISTORICAL_SOURCE=<pinned-source> python3 \
 tools/audit_10min_demo_timing_correction.py --output <new-private-directory>
PF_DMO_TIMING_CORRECTION=<new-private-directory>/summary.json \
 python3 -m unittest discover -s tools -p 'test_audit_10min_demo_timing_correction.py'
python3 tools/test_dmo0_scroll_extent.py
```

Use the existing research Python/Capstone environment and bundled Go toolchain.
For the candidate run `go test -tags dmoimpl1 ./internal/partyland ./internal/physics
-run '^TestDemo' -count=1 -v` in the existing private checkout with PF_10MIN_DEMO_DATA.
The complete command boundary table is in the correction record.

No shared matrix/physics, actual private rendering, source phase, FADE256, typed
QUIT, interruption/restart, sticky UNSUPPORTED or BYGEL12 rejection was changed.
No hardcoded deadline, new trajectory search, producer, scored-drain implementation,
BYGEL12 implementation, interactive profile, DMO1 or unrelated closure work.
No canonical-A oracle, .DS_Store or v0.1.3 change; no commit/push/tag/release.
DMO-IMPL-2B8 is not started.
