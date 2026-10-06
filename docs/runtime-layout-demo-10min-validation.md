# Party Land 10-minute DOS demo: DMO0 research gate

Date: 2026-10-06. Base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.

**DMO0 NOT CLOSED. DMO1 NOT STARTED. Demo remains unsupported.**
This report records static evidence and the remaining proof obligations; it is
not a production layout descriptor or a demo support announcement. No DOS code
was executed, no DOSBox was installed, and no demo was downloaded.

The updated target is a bundled playable fallback after both gates pass, with
`Play 10-minute demo` / `Import DOS version` on a clean installation. That target
supersedes the original import-only target. Bundling, package allowlisting and
startup UI are deferred with DMO1. Nothing in this research changes release
versions, tags, assets or publication.

## Identity and provenance

Current private inputs were read from `/Users/mess/work/Pinballs/pinbfan` and
canonical A from `/Users/mess/work/PINBALLF`. The earlier private audit is
`/Users/mess/work/Pinballs-audit-2026-10-06/REPORT.ru.md`, with its
`final-comparison.json` and `archive-evidence.json`. Runtime inputs were hashed
again in this task; the source archive itself was not fetched or revalidated.

The five-file runtime fingerprint is
`e7d9aaedf4f06f67d1553d88be0b0f87568bb1da49344b1a7f070f37c160809e`.
It is SHA-256 over UTF-8 `NAME + NUL + lowercase SHA-256 + LF`, in the order
below. ZIP name, launcher, drivers, TIMER.BIN and mutable settings are excluded.

| Required runtime role | Bytes | SHA-256 |
| --- | ---: | --- |
| INTRO.PRG | 347054 | `05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3` |
| INTRO.MOD | 252870 | `f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613` |
| MOD2.MOD | 55394 | `aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21` |
| TABLE1.PRG | 537190 | `44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3` |
| TABLE1.MOD | 210760 | `a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5` |

Prior archive metadata records 725159 bytes, 19 entries, SHA-256
`5a65519053f3e66c323aeea134611de20c30c7872d726b7151532bbade3d3f79`.
This is research provenance, not production acceptance or evidence of an
unaltered publisher master. The prior audit reports all extracted members
matching its archive instance.

The launcher is 2063 bytes, SHA-256
`7acc8be42f23cc56a66ce8839a561c4d7318025dc395935be44b7a760f538dff`.
It contains the user-specified 10-minute closing sentence at file `0x43f`.
INTRO.PRG contains `10 MINUTE DEMO` at `0x582b` and F1 play wording at `0x583f`.
Together with the actual TABLE1 threshold below, these establish the supplied
10-minute demo identity independently of its directory/archive label.
Launcher and driver files are research inputs only, never native runtime roles.
TIMER.BIN (253 bytes, SHA-256
`783f88891a760b3fab648a7ae64ad8998e1349737df07a1e765329f7a6787d0c`)
programs PIT channel 2, measures bounded VGA/multiply work, restores hardware
state and returns a speed class. It is a machine-speed calibrator, not the
ten-minute counter. Its presence does not add a sixth native data role.

Candidate semantic name: `dos-partyland-10min-demo-v1` (Party Land only, bounded
demo edition and reviewed link map). **This ID is not registered.**

## INTRO FORM inventory

All 17 bounded FORM records were reparsed from private bytes. Fourteen complete
FORMs match A; three differ. Table order here follows the existing common
frontend roles, not physical file order. Destinations are semantic roles for a
future descriptor; no prepared demo PRG was created.

| Source offset | Role / prospective common destination | Geometry | Planes | Relation to A |
| --- | --- | --- | ---: | --- |
| `0x1cbd0` | Logo | 640 × 256 | 4 | exact FORM, A `0x1cb10` |
| `0x6c30` | Font | 640 × 28 | 4 | exact, A `0x6b70` |
| `0x8bc0` | MonoFont | 640 × 29 | 4 | exact, A `0x8b00` |
| `0x204f0` | Party Land card | 640 × 95 | 4 | exact, A `0x20430` |
| `0x24b10` | unavailable Speed Devils advertising card | 640 × 200 | 4 | demo-specific |
| `0x29910` | unavailable Gameshow advertising card | 640 × 200 | 4 | demo-specific |
| `0x2e2c0` | unavailable Stones advertising card | 640 × 200 | 4 | demo-specific |
| `0x3bd70` | Startup[0] | 320 × 256 | 8 | exact, A `0x3b810` |
| `0x431b0` | Startup[1] | 320 × 256 | 8 | exact, A `0x42c50` |
| `0x46e90` | Startup[2] | 320 × 130 | 8 | exact, A `0x46930` |
| `0x4e000` | Startup[3] | 320 × 126 | 8 | exact, A `0x4daa0` |
| `0x120e0` | Startup[4] | 320 × 110 | 8 | exact, A `0x12020` |
| `0x17fe0` | Startup[5] | 320 × 130 | 8 | exact, A `0x17f20` |
| `0x10f20` | Startup[6] | 320 × 200 | 8 | exact, A `0x10e60` |
| `0xa510` | Startup[7] | 640 × 178 | 4 | exact, A `0xa450` |
| `0x35360` | HighLogo | 640 × 200 | 4 | exact, A `0x34d90` |
| `0x339e0` | HighMono | 640 × 200 | 4 | exact, A `0x33410` |

The three altered FORM bodies were independently decompressed and visually
inspected in temporary files outside the repository. Each is the corresponding
table advertising card with `NOT AVAILABLE` across it; its lower area is blank.
They are **not demo expiry notices or replacement startup splash screens**.
Canonical replacement would erase the supplied advertising/nonplayable meaning.

MZ relocation entries point at segment references at file `0x6567`, `0x6569`,
`0x656b` for these three assets (segments `0x2471`, `0x2951`, `0x2dec`). INTRO's
selector uses those table references at `0x38907–0x3892e` and
`0x3894f–0x3897a`. Exact placement/cropping through the full unpack/presentation
call graph has **not** been proved; FORM dimensions alone do not establish it.
A future presentation capability must represent advertising cards independently
of table playability, retaining their supplied pixels.

F1 dispatch is live at `0x389cc–0x389ce`. The comparisons for F2/F3/F4 at
`0x389d3`, `0x389d5`, `0x389d7` have no corresponding conditional branches to
their old handlers. Those handlers still exist, but this selector does not
dispatch to them. Full source table-name strings/unused handlers do not make
TABLE2–4 playable. F5 remains dispatched at `0x389d9–0x389db`.

The canonical 240-byte sidebar/options record at `0x3914e` has no demonstrated
demo equivalent. Demo menu text includes the new 10-minute/F1 wording. A fixed
240-byte copy based on an approximate shift is prohibited.

## TABLE1 read-map audit

The read-only tool [audit_10min_demo.py](../tools/audit_10min_demo.py) verifies
the exact research inputs and all 1632 entries of the prior A-based read-map
comparison. It exports offsets, hashes, roles and classifications only. It
does not generate or register a production descriptor. Its per-region output
is owner-local `/private/tmp/pf-dmo0-evidence.json` in this run.

The previous 1625 equivalent entries were byte-rechecked. Four additional
material entries are now proved equivalent at the correct consumer address.
Two jingle differences are typed below. The remaining PLAYERSTEXT entry is not
an equal-length equivalent. The exact-byte candidate translations are:

| Shift | Equivalent candidate entries after material correction |
| ---: | ---: |
| +112 | 358 |
| +130 | 6 |
| +140 | 247 |
| +360 | 43 |
| +362 | 8 |
| +363 | 3 |
| +365 | 18 |
| +368 | 946 |

These counts cover bitmap/glyph records, fonts, animation control and bitmap
records, foreground, palettes, collision masks, flipper records/frames, spring,
lamp schedules/packets, gates, lookup tables, score seeds and matrix text in
the existing read map. They are **not a reviewed per-consumer production map**:
exact short-record matches and prior extended runs do not alone establish
linked pointer/consumer identity. The four playfield PBMs at `0x525a0`,
`0x597d0`, `0x61b10`, `0x6ad20` are complete exact A FORM payloads at +368.

Material source base is `0x1c1c7`, stride 16, five signed 16-bit parameters per
entry. TABLE1 startup loads DS `0x19bb` (file base `0x19db0`) at `0x32ad`, and
its initialization at `0x3411` sets material BX `0x2417`, then walks eight
entries with stride `0x10` at `0x3414–0x342a`. This proves source base and +362.
All eight WallFriction/BallFriction/Bounce/MinSpeed/MaxAngle records match A.
The old audit's four material differences came from +363 nearest-anchor
guesses, one byte after the actual word-aligned records. They are not intentional
demo physics differences. Low-resolution initialization transforms MinSpeed by
5/6; its complete equivalence to all native modes still belongs to the control
graph review, not a new physics algorithm.

## Proved differences and additional control records

| Record | Demo source | Meaning / difference | Current native consumer |
| --- | --- | --- | --- |
| S_EMPTY | `0x1aa38`, 3 bytes | position 62, repeat 0, priority 0; A priority 1, B/C/D priority 0 | decoded JingleSpec → Party Land cue → MusicClock.Play |
| S_GAMEOVER2 | `0x1aa4d`, 3 bytes | position 13, repeat 0, priority 255; A/full common record priority 1 | same cue path; demo timer explicitly invokes it |
| PLAYERSTEXT | `0x1c129` | demo duration label using source glyph encoding; full record is a mutable player-number label | presentation text + Party Land score-display mutations |
| demo bonus continuation | pointer at `0x1b533`, handler file `0x73e` | replaces full player/ball progression with demo-only ball-display progression and NEW_BALL_TASK | canonical compiled matrix commands / `partyland/timing.go` `_CHANGE_PLAYER` |
| expiry program | `0x1ba17`, 42 bytes | new scroll/score/flash/fade/quit program, absent from full common matrix graph | no current decoded demo presentation/control program |
| unavailable table cards | INTRO offsets above | nonplayable advertising instead of full playable labels | frontend cards / table input availability |
| INTRO menu text | marker `0x582b`, F1 wording `0x583f` | new text/organization; old sidebar/options region is not an equivalent record | fixed current frontend text model |

This is the list established in this run, **not a claim of exhaustive semantic
closure**. Factory Party Land seed data at `0x19dc6` (64 bytes) matches A.
All three supplied MODs pass the existing role-specific decoders and their
complete Module models are DeepEqual to the corresponding A models. No demo
audio decoder or scheduler change is warranted by these inputs.

## Timer and expiry: established core, open cadence contract

All offsets below are file offsets, avoiding confusion with linked CS/DS offsets.
TABLE1 code base is `0x300`, data base `0x19db0`. The counter is a zero-initialized
16-bit word at `0x1d27d` (DS:`0x34cd`); the expired flag is `0x1d27f`.

At `0x5cdb` DO_ELECTRONICS increments the counter, then `0x5cdf` compares it
with **35998 = 10 × 60 × 60 − 2**. This is an equality test after increment,
not a wall-clock comparison. At equality it sets the expired flag and HOLDSTILL,
plays the actual S_GAMEOVER2 record and dispatches the program at `0x1ba17`.
The counter has no other direct references in the inspected TABLE1 code range.
Normal new-ball/reset routines inspected here do not reset it.

The direct caller is `0x4723`, in the raster/ball-handler rest-of-update path.
The historical private `FANTASIE.ASM` has the corresponding DO_ELECTRONICS
branch at lines 4601–4618, but uses `DEMOVER=0` and a five-minute threshold.
Its comments distinguish calculations from presentation syncs. It is corroboration
of labels/structure, **not the exact build source for this 10-minute binary**.

Pause code at `0x34ce` disables INTERRUPTS_ON, stops the audio routine and waits
for keyboard input; `0x3530` reenables updates on resume. Thus paused callbacks
cannot be counted as ordinary electronics updates. INTRO has no TABLE1 counter;
its menu/startup time is outside this counter. On each fresh TABLE1 DOS process,
the initial counter and flag are zero. The nominal scope is the table process,
including its running update states, not application launch wall-clock time.

The expiry program has reviewed handler targets for clear, scroll, flash on,
score rendering, wait 100, flash off, second scroll, fade 256, wait 100 and table
QUIT. Its teardown at `0x3cdb–0x3d2f` stops sound, resets mouse/input and exits
TABLE1 with DOS status 0. The launcher executes INTRO again after a successful
table return (`PINBALL.EXE` `0x6f4 → 0x64d`). Therefore the statically observed
path is **table session end → INTRO**, not an immediate launcher exit.
The launcher's closing message belongs to its zero-table-selection path
(`0x70f–0x71e`); the timer itself does not directly print it. F1 in the next
INTRO can load another fresh table process. Relaunch naturally resets the count.

**Remaining timer proof:** the full actual SDR/music/raster callback dispatch
contract, reentry/skipped callback policy, both resolutions and paused/quit
transitions have not been reconstructed. The fixed native Runner tick must be
mapped to the proved counted event before implementing a deterministic limit.
35998 is a proved calculation threshold; claiming that exactly 600 wall-clock
seconds or exactly 36000 native ticks reproduces all DOS states would be a guess.
No native timer was implemented, so there are no native expiry/boundary test results.

## Persistence and the control-graph blocker

TABLE1 retains a `table1.hi` filename at `0x19dbc`, factory seeds, and file read/
write routines around `0x66d0` / `0x6706`. The inspected startup and teardown omit
their calls; scanning the code range found no direct calls to these two routines.
Historical DEMOVER guards omit INIT_HIGHS/SAVE_HIGHS as well. This supports
**factory/transient scores without ordinary full-game persistence**, but complete
indirect reachability and demo score-session paths remain open. A filename or
unused routine is not proof of persistence. No native demo store was enabled.

The decisive consumed-control blocker is the pointer at **TABLE1 `0x1b533`**:
its CS value `0x043e` selects file handler **`0x73e`**. This handler preserves
held bonus, updates two encoded BALLSTEXT display digits, stores player state,
queues NEW_BALL_TASK and continues the matrix program. It does not execute the
ordinary full-game `_CHANGE_PLAYER` ball-limit branch. The historical
`PLAND.ASM` 1640–1648 and 2871–2899 identify this as `_DEMOVER_CHANGE_PLAYER`.

The full native graph instead embeds `_WAITIFMULTI`, `_MATRIXLGT`, clear and
`_CHANGE_PLAYER`; the latter uses `g.changeBall()` and full Session progression.
Copying demo bytes to canonical addresses preserves neither the omitted commands
nor this new transition. A timer-only port could end a demo after three balls
before its timer or report the wrong ball/score state.

The handler's local operations are known, but their complete downstream meaning
through `_KOLLA_XXBALL`, shoot-again, saved per-player state, ordinary new-ball
tasks, display-digit wrap and high-score transitions has **not** been proved for
this linked demo graph. The existing 1632-entry canonical data read map does not
include these linked command words as a decoded runtime program. Required next
proof: recursively decode and review the demo matrix/control graph, identify
every modified edge and its state consumers, then express the proved differences
as common typed capabilities/programs. The companion callback cadence and INTRO
placement/text proofs must close too. These are research obligations, not user
permission or a missing-device requirement.

## Architecture and bundled target after closure

No capability refactor was made while DMO0 is open. Full A/B/C/D still require
all 11 roles. Demo input still fails at the shared production boundary.

The pending model is RuntimeProfile → required roles, available tables,
edition/session/persistence/presentation capabilities → source descriptor →
common decoded model. Full has tables [1,2,3,4]; demo would have [1] plus optional
nonplayable advertising cards. Ball/session continuation also needs a proved
typed capability; availability and time limit alone are insufficient.

The updated RuntimeSource target isolates external installation and bundled
read-only demo sources. Candidate validation must use one source exclusively;
external full-game precedence never fills missing roles from the demo. Import
must retain existing transactional staging/adoption/rollback and explicit
replacement semantics. Android's bundled path may materialize atomically into
private internal storage with post-copy hashes/recovery. Demo launch must not
require SAF or a picker. Desktop resource lookup must survive app relocation.

Future packaging must take the five pinned files from an explicit private input
outside tracked source, verify them before packaging, include each once, and
permit **only these exact identities** in public resources. Full A/B/C/D payload,
unknown PRG/MOD, DOS executables/drivers and archive wrappers remain forbidden.
State stays in native userdata. No packaging destination, approved-demo scanner,
clean-install test or external-source-precedence implementation exists yet.
Existing asset-free package policy was not weakened ahead of that implementation.

## Validation in this task

| Check | Current result |
| --- | --- |
| Five-file research identity; 17 FORM inventory; materials and jingles | PASS through read-only research tool |
| Shared INTRO/MOD2/TABLE1 Module decoder equivalence | PASS, private Go test |
| Production demo support | NOT IMPLEMENTED; fail-closed rejection remains |
| Demo INTRO + full TABLE1; full INTRO + demo TABLE1, A/B/C/D | PASS rejection |
| Demo PRGs plus full TABLE2–4 | PASS rejection for all four sets |
| MOD cross-edition rejection | Not a meaningful mismatch: these three decoded roles are equivalent; no new artificial MOD family discriminator |
| A/B/C/D private mappings and coherent detection | PASS |
| Existing 1024 four-profile PRG matrix | PASS |
| C/D seeds, Stones selectors, priorities and shared construction | PASS existing private regression tests |
| Private fixture absent | Clean SKIP, `PF_10MIN_DEMO_DATA` / `PF_RUNTIME_DATA` |
| Native demo gameplay/intro/expiry and deterministic smoke | NOT RUN: no supported profile |
| Timer before/exact/after, pause/reset/relaunch boundary tests | NOT IMPLEMENTED: counted-event mapping remains open |
| Demo/full imports, replacement, rollback, interrupted extraction | NOT IMPLEMENTED |
| macOS bundled/imported demo journeys | NOT RUN |
| Android demo provider/extraction, two ABIs, APK alignment, physical device | NOT RUN |
| Windows amd64 GUI build | BUILD PASS, PE machine 0x8664; execution NOT RUN on this Mac |
| Linux amd64 trace build | BUILD PASS, ELF machine 62; SDL GUI build/execution NOT RUN |
| New distributables / clean-install bundled demo | NOT PRODUCED / NOT RUN |
| A oracle, captures, reference fixtures | UNCHANGED; no regeneration or allowlist changes |

Regression command used private env variables `PF_RUNTIME_DATA`,
`PF_POWERPACK_DATA`, `PF_DELUXE_CD_ALT_DATA`, `PF_DELUXE_CD_DATA` and
`PF_10MIN_DEMO_DATA`, with `go test ./internal/datalayout ./internal/frontend
-run 'TestPrivate|TestPowerPackDescriptor|TestDeluxe' -count=1 -v`.
There is no variable-role runtime refactor whose acceptance is being claimed.

The four known baseline assertions named in the request were not rerun in this
focused research suite; this task makes no claim they are fixed or newly PASS.
Likewise missing historical/reference captures were not regenerated or silently
substituted. Their prior failure/unavailability remains outside these focused
PASS results.

Changed source is the research tool, private research test, this report and a
compatibility-document link. All payload/images/disassembly output stayed outside
tracked source. The four changed files and newly built Windows/Linux binaries
passed scanning against 1973 distinct nontrivial 4 KiB blocks from private
A/B/C/D/demo PRG/MOD and DOS support inputs, plus whole-file identity checks.
This verifies these sources/binaries, not every pre-existing ignored artifact
in the workspace. No new app bundles/APKs or bundled-demo public packages were
produced. The local research commit SHA is in the final task response.
No push, tag or release was performed.
