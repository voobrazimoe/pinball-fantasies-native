DMO0 NOT CLOSED. DMO1 NOT STARTED.

# Party Land 10-minute DOS demo: DMO0 research gate

Latest 2026-10-07 research snapshot:
[native audio / gameplay callback boundary](runtime-layout-demo-10min-audio-boundary.md)
and [current control-domain snapshot](runtime-layout-demo-10min-control-domains.md).
The new semantic-boundary verdict is NOT_PROVED; the historical whole-DOS gate
below remains separate and unchanged.

Date: 2026-10-06. Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
This continuation starts from research `2a40c3a55d130fa324dda1ceb2fcea40ba0913e1`.

**DMO0 NOT CLOSED. DMO1 NOT STARTED. Demo remains unsupported.**
This report records static evidence and the remaining proof obligations; it is
not a production layout descriptor or a demo support announcement. No DOS code
was executed, no DOSBox was installed, and no demo was downloaded.

This continuation concerns DMO0 semantic obligations only. Identity, FORM
inventory, the 1,632-region read map, material/audio equivalence, jingle typing,
counter threshold and expiry program below retain their prior-pass evidence;
those audits were not repeated. Exact private identities are verified before
the new graph analysis. Bundling, package allowlisting and startup UI remain
deferred with DMO1. Nothing changes release versions, tags, assets or publication.

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

In the prior research pass, the three altered FORM bodies were independently
decompressed and visually inspected outside the repository. Each is the corresponding
table advertising card with `NOT AVAILABLE` across it; its lower area is blank.
They are **not demo expiry notices or replacement startup splash screens**.
Canonical replacement would erase the supplied advertising/nonplayable meaning.

MZ relocation entries for the three advertising FORMs are at `0x6567`,
`0x6569`, `0x656b` (segments `0x2471`, `0x2951`, `0x2dec`). The consumer chain
below now establishes their presentation; this does not use FORM dimensions as
a placement assumption.

### Advertising consumer proof (BLOCKER 3)

INTRO code base is `0x36e90`; its DS file base is `0xc00`. The unpack loop
`0x382c7–0x38306` consumes seven source segments and destination rows from
`0x6551`: 0, 240, 268, 296, 391, 486, 581. The card rows are 296/391/486/581.
The caller passes width 440 for cards, but unpack routine CS:`0x4810` takes the
actual 80-byte planar row stride from the display setup and BMHD geometry.
Thus the 640×200 advertising FORMs are initially unpacked with their own height;
those overlapping temporary storage rows are not the displayed windows.

`CRUNCH_PICS` at `0x39b7d`, called at `0x38330`, preserves the first 55 bytes
of row 296 and compacts the following 379 rows: destination `0x5cb7`, source
`0x5cd0`, copy 55 bytes, skip 25 bytes per row. The four consumed cards therefore
contain exactly the **left 440 pixels of their first 95 rows**, per plane.
Rows 95–199 of each advertising FORM do not contribute to the selector card.

The selector's `0x38907–0x3892e` / `0x3894f–0x3897a` selects palettes for a pair,
then calls the raster presenter at `0x3893a`. Top presenter `0x3944f` uses
DI=`0x334` (x=160, y=10), SI=`0x5c80`; bottom presenter `0x39471` uses
DI=`0x2a44` (x=160, y=135), SI=`0x70e9`. Both copy 55 bytes per plane per row.
The second pair adds BP=`0x28d2` = two compacted cards. Raster packet table
`0x663b` has 18 packets, covering each row 0–94 exactly once, with an interpacket
sync between successive packets; pre-display wait is 20 selector syncs. The
page hold is 540 selector callbacks. This is a packet reveal of fixed windows,
not scrolling a full 640×200 card.

| Demo FORM | Source crop | Selector destination | Palette |
| --- | --- | --- | --- |
| `0x24b10`, Speed Devils | x=0, y=0, 440×95 | first pair, x=160, y=135 | bottom bank 0–15 |
| `0x29910`, Gameshow | x=0, y=0, 440×95 | second pair, x=160, y=10 | top bank 16–31 |
| `0x2e2c0`, Stones | x=0, y=0, 440×95 | second pair, x=160, y=135 | bottom bank 0–15 |

Each card's own 16-color CMAP is consumed through CS:`0x1e86`; the pair copies
48 bytes per card into the current palette buffers. Raster palette switching
uses CHANGE16PAL. The first pair's top card is the shared Party Land FORM.
Pair selection controls advertising presentation, not table availability.

F1 dispatch is live at `0x389cc–0x389ce`. F2/F3/F4 comparisons at `0x389d3`,
`0x389d5`, `0x389d7` have no branches to the retained old table handlers.
F5 at `0x389d9–0x389db` reaches options. Space/automatic page change advances
advertising pairs. Old instructions `0x389ec–0x38a19` are not reached by the
entry/selector direct graph; their stale text-index writes are not consumed
menu semantics.

### DOS consumed menu source → semantic field

A real demo 240-byte record is at **`0x396d3`**, not canonical `0x3914e` plus an
approximate shift. It is two independently consumed 120-byte blocks of ten
12-column rows. Sidebar task `0x398fc` selects CS pointers `0x2843` / `0x28bb`
and printer `0x39954` consumes those rows at a nine-pixel vertical stride.
Options entry `0x3b21e` switches the pointer; exit `0x3b4ed` restores it.

| Consumed source | Consumer | Typed presentation field |
| --- | --- | --- |
| text list `0x5801` → `0x580d` | SHOWTEXT pointer read `0x3a09a`; 12 bounded, zero-delimited rows, ≤24 columns | welcome title: WELCOME TO / PINBALL FANTASIES; notice: 10 MINUTE DEMO; actions: F1 - PLAY PINBALL, F5 - GAME OPTIONS |
| same list → `0x5869` | same page printer | acquisition page: PINBALL FANTASIES / IS AVAILABLE NOW / WRITE TO: / 21ST CENTURY / ENTERTAINMENT INC / P.O. BOX 415 / WEBSTER / NEW YORK 14580 |
| `0x396d3`, first 120 bytes | selector sidebar task/printer | F1 Party Land (two rows), F5 options, Escape quit; other rows blank |
| `0x3974b`, second 120 bytes | options sidebar task/printer | cursor keys select, Enter or Space toggles, Escape quits; bounded ten-row organization |
| `0x5b01` | BX load `0x3b24e` → page printer CS:`0x2ca7` | options page title/labels: balls, angle, scrolling, in-game music, resolution, color mode, save and exit; shared value formatter `0x3af32` writes at row+13 |
| FORM records above | compact/copy/palette consumer | nonplayable table advertising with the supplied NOT AVAILABLE image, independently of menu action labels |

Text list words are `0x4fbd, 0x4c0d, 0x4c69, 0, 0`. SHOWTEXT increments its
index by two and wraps zero entries to index two: the welcome and acquisition
pages are reachable from the entry graph. The high-score-page pointer is the
shared high-score presentation, not another demo notice. Credit text retained
near `0x58e0` is not in this consumed cycle and is not proposed for the model.
The binary options dispatcher `0x3ae21–0x3ae76` handles up/down, Enter/Space,
Escape; ball choice 3/5 remains a displayed option, but the demo continuation
below does not read that limit. No canonical sidebar bytes are substituted.
**BLOCKER 3's placement and consumed presentation obligations are closed.**

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
| INTRO menu text | text list `0x5801`, pages `0x580d`/`0x5869`, sidebar `0x396d3` | separately consumed demo text/pages/sidebar organization | typed presentation fields above |

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

### Counted-event dispatch: local binary proof, unresolved external admission (BLOCKER 2)

TABLE1 installs the primary callback at `0x6345` using INT 66h AX=11,
BL=100, ES:DX=CS:`0x4217` (file `0x4517`). The later callback at `0x6361`
uses AX=12, BL=200, ES:DX=CS:`0x562b` (file `0x592b`), CX=108 in low
resolution or 174 in high resolution. Both resolutions converge on the same
counted call site; a rendering cadence or audio callback is not itself a count.

| Callback edge / gate | Binary condition and counted-event effect |
| --- | --- |
| primary callback `0x4517` | increments SYNC before its update guards |
| `INTERRUPTS_ON` at `0x4536` | disabled update returns without electronics; pause uses this guard |
| `DEMOMODE` attract branch | branches to `0x6386`, outside ordinary electronics |
| slowdown gate `0x456a` | exact demo AND mask is zero, so this gate does not drop alternate updates |
| `LAST_WAS_VB` `0x4571–0x4578` | returns until later callback clears the flag |
| `INSIDE_BALLHANDLER` `0x457a–0x4588` | overlapping ball handler returns |
| low/high/scroll rendering branches | converge through ball calculations and DO_PHYSICS call `0x46b8` |
| `INSIDE_RESTOFVBLANK` `0x46c0` | busy rest-of-update skips electronics; it is set at `0x471b`, cleared at `0x4753` |
| `0x4720 → 0x4723 → 0x5cd9` | UPDATE_COUNTERS then exactly one DO_ELECTRONICS call on this admitted rest-of-update path |
| TIME_LEFT `0x472a` | checked after electronics; budget can skip matrix work, not this counter increment |

At the electronics entry there is no HOLDSTILL, matrix-wait or between-ball
phase test before increment. When those states reach this entry, they count.
The local reentry guards prevent a second admitted rest-of-update while the
first is active, but can cause skipped work. The complete scheduler admission
contract has not yet been proved across every supplied driver and
expiry/task interleaving. The drain return itself is now closed locally below. Do not turn this local result into an unconditional
one-callback/one-native-tick theorem.

The new research tool pins and statically decodes all eleven supplied EXEPACK
SDR containers **in memory only**. This interprets bounded compression blocks;
it neither runs DOS instructions nor writes an unpacked executable. GUS's entry
IP is 22; the other ten entries are 6. Registration producers are now derived for all eleven, rather than assumed
from the two reviewed families:

| SDR | Decoded-module addresses (not packed file offsets) | Established boundary |
| --- | --- | --- |
| NOSOUND | callback install `0x6bc`; IRQ entry `0x5bf`; STI `0x5c7`; priority gate `0x66b`; callback `0x687` | raster-calibrated IRQ0 record scheduler; skips records below the active priority |
| ADLIB | install `0x1a81`; primary callback `0xbdd`; later callback `0xc11`; budget helper `0x1a32` | buffer/interrupt scheduler; callback and budget state differ from NOSOUND |

The INT66 dispatcher is recovered from each module's actual vector installation
and AL compare/branch chain. API 11/12 entries are followed to ES:DX stores;
indexed API 12 insertion exposes the record base, nine-record bound and
nine-byte stride. These are registration/storage proofs, **not** IRQ reachability,
frequency, callback multiplicity or admission invariance proofs.

| Driver | API 11 entry | API 12 entry | API 12 storage family |
| --- | --- | --- | --- |
| ADLIB | `0x1a81` | `0x1ab8` | separate far-pointer field |
| GUS | `0x7b3` | `0x813` | indexed record insertion |
| INTERNAL | `0x1a02` | `0x1a39` | separate far-pointer field |
| NOSOUND | `0x6bc` | `0x71c` | indexed record insertion |
| PAS16 | `0x1c7a` | `0x1cda` | indexed record insertion |
| SB16 | `0x1e37` | `0x1e97` | indexed record insertion |
| SB20 | `0x1e9b` | `0x1efb` | indexed record insertion |
| SBLASTER | `0x1d55` | `0x1db5` | indexed record insertion |
| SBPRO | `0x1ed4` | `0x1f34` | indexed record insertion |
| SM2 | `0x1cd8` | `0x1d38` | indexed record insertion |
| THING | `0x1a37` | `0x1a6e` | separate far-pointer field |

**Still required:** IRQ source/vector and calibration-to-admission closure,
primary/later scheduling, priorities, nesting/skips and driver invariance. No
A/B/C cadence verdict is selected. Historical FANTASIE.ASM supplies labels and
structure only; its five-minute numeric assumptions were not imported.

**Drain order, binary-confirmed:** `0x46b8` calls DO_PHYSICS `0x5d1f`;
`0x5d4a` calls LOOSE_BALL `0x515`, then returns at `0x5d4d`. The primary path
continues through the rest-of-update guard `0x46c0`, UPDATE_COUNTERS `0x4720`
and DO_ELECTRONICS `0x4723`. Drain itself does not bypass this electronics
calculation. In particular, the drain sees the old expired flag before that
update's threshold equality; moving the counter ahead of drain changes semantics.

### Native event mapping and timer contract are not closed

`internal/source/runner.go` advances frontend/application ticks, including states
outside table electronics. `internal/frontend/model.go` uses 71 Hz in table
modes and 60 Hz in INTRO. Neither Runner.Ticks nor presentation frames are a
proved demo counter. `internal/partyland/game.go:244` advances Game.Tick and audio
before a GameOver return, and has separate BallLost display/task/matrix work
without Physics.Sync. `beforeTargets` is therefore not a complete counter hook.
Physics.Sync combines the ball/raster paths with DO_PHYSICS semantics; its
individual physics substeps are not separate electronics calculations.

A proposed explicit **ElectronicsCalculation** logical event must represent the
admitted DOS rest-of-update, before area/target/task work, including held-ball,
wait and between-ball updates that actually reach it. The local DOS order now requires this
hook **after ball/drain processing and UPDATE_COUNTERS, before areas/targets,
shift, KEYTASK and tasks**. The current Physics.Sync drain return precedes
BeforeTargets; the separate Game.BallLost branch also omits that hook. Thus a
future explicit event must cover these paths, not just add a counter to
beforeTargets. The existing PF3 comment about bypassing electronics is not a
proof about this demo binary. No production code was changed. There is **no proved N** for
`one calculation == N Runner ticks`, and no native implementation is added.

Proved constraints for that eventual contract: increment first, compare exactly
35998 by equality, pause-disabled update and INTRO do not count, HOLDSTILL alone
does not suppress the entry, resolution uses the same counted site, a fresh
TABLE1 process starts at zero, inspected new-ball/new-game resets do not reset
it. Subsequent expiry updates still increment when admitted; equality alone
is not a saturation test. The uint16 counter can wrap, and equality can recur
after 65,536 further admitted calculations if TABLE1 has not returned. Neither
an eventual-exit bound nor the first-resume callback/LAST_WAS_VB relation has
been proved; they remain explicit gates. No wall clock, audio IRQ count or presentation count
is proposed as a substitute. Full pause/reset/expiry admission closure remains
BLOCKER 2, not an implemented/tested timer contract.

## Continuation graph (BLOCKER 1)

The bounded normal bonus program was linked as a whole using source-derived
command arities and numeric operands, with one unique candidate per binary.
Canonical A is `0x1b3cd` (119 words); demo is `0x1b459` (114 words). This is a
program/consumer proof, not another nearest-offset read-map audit. The demo tail
is `_KOLLA_XXBALL → _DEMOVER_CHANGE_PLAYER → _CLEAR4 → _WAIT 32000 → 0`.
Canonical instead inserts `_WAITIFMULTI`, `_MATRIXLGT`, `_CLEAR4`, `_CHANGE_PLAYER`
after `_KOLLA_XXBALL`. Demo's handler pointer at `0x1b533` is CS:`0x043e`, file
`0x73e`. It is not `_CHANGE_PLAYER`.

All addresses in this section are TABLE1 file offsets; state names refer to
DS at file base `0x19db0` unless stated otherwise.

| Edge / source label | Condition / state read | State write / queued work | Canonical relation / observable meaning |
| --- | --- | --- | --- |
| `0x73e`, `_DEMOVER_CHANGE_PLAYER` | HOLDBONUSFLAG DS:`0x5b3` true | copy 12 encoded digits TEMPSIFFRORNA `0x34a1` to BONUSSIFFRORNA `0x3495` | keeps the saved bonus snapshot; not a recomputation |
| `0x756–0x77a`, same handler | reads two BALLSTEXT digits `0x2389/0x238a` | increment ones; carry resets ones to glyph zero; tens becomes 1 unless already 1, when it becomes 2 | demo display progression rather than Session ball count |
| `0x77f`, VARS_2_P_STRUC `0x1178` | current PLAYER `0x3819`, stride `0x74` | save current player record; PLAYER unchanged | no ordinary player rotation here |
| `0x782–0x785` | same continuation | queue NEW_BALL_TASK CS:`0x0bbb` / file `0xebb` through DOADDTASK `0x5e80` | proceeds to another ball, then HU_ next matrix node |
| `_KOLLA_XXBALL` `0xd75` | XXBALLE DS:`0xcd` zero | continue to demo handler | ordinary demo bonus tail ignores earned shoot-again branch |
| same, XXBALLE nonzero `0xd84` | saved player state, light 51 (`0x36c3`) | light set: LET_HIM_SHOOT_AGAIN `0xe3a`, decrement XBALLS `0xce`, select SHOOT_AGAIN program `0x17a2`; otherwise test/rotate PLAYER versus PLAYERS and select match continuation | retained match machinery; global reachability is still open |
| retained `_CHANGE_PLAYER`, file `0xdb9` | PLAYER/PLAYERS, ordinary BALLS `0x34db`, NO_OF_BALLS `0x34dc` | ordinary branch `0xdf4` increments BALLS; `0xe51` selects OUT_OF_BALLS `0x17ba` | absent from normal demo bonus program; no pointer to this handler in the reviewed DS extent |
| unscored drain, LOOSE_BALL | SCORECHANGED false | PARTY_ON/S_SPRING; PARTY_ON_TASK1 wait 30 → NEW_BALL | same-ball free plunge; bypasses digit increment |
| scored drain `0x5d2 → 0x622` | expired flag `0x34cf` true | effect/program reference SI=`0x6f1` | additional demo expiry replay guard before ordinary bonus |
| attract start `0x64cf–0x650d` | selected F1–F8 start key | PLAYERS `0x3815`; new-game reset; NEW_BALL | real input path accepts player counts; multiplayer is not disproved by demo title |

**Ball progression proved on the normal demo continuation:** a fresh display
starts with BALL 1, advances through 9, 10–19, 20–29, then returns to 10 and
cycles 10–29. This is a two-glyph counter, not a 3-ball Session. The handler
neither increments ordinary BALLS nor rotates PLAYER. New-game reset changes
the ones glyph to 1; the inspected code does not reset the tens glyph there.
Unscored drain can replay the same displayed ball. The retained start-key path
accepts up to eight player counts, but normal demo continuation stays on the
current player: selection and canonical rotation must not be conflated.

**Bonus:** `_FLORPA` at `0x629` snapshots twelve encoded bonus digits before its
countdown/score transfer. Held bonus restoration uses that snapshot, which may
be an earlier snapshot if the current path skipped `_FLORPA`; it is not safe to
replace it with an unconditional current-bonus copy. Ordinary countdown clears
bonus as it transfers the multiplied/cyclone/happy/mega total to score. The
player save/load pair `0x1178` / `0x104c` carries score, bonus, skill data,
cyclone data and 17 persistent progression lights. Happy/Mega declared record
fields are not saved by these routines; XBALLS is not a saved player field.
RESET_TABLE clears held-bonus flag and restores multiplier 1, then loads the
saved current-player state. Closure still requires the zero-bonus/held snapshot
paths and all match/high-score/cheat roots to be reviewed together.

**Shoot-again:** the normal demo handler does not test light 51 or consume
XBALLS. The retained XXBALLE match branch does. An unscored drain also preserves
the displayed ball without consuming this normal continuation. A claim that
all shoot-again code is unreachable would exceed the present evidence.

### NEW_BALL_TASK: proved local consumer sequence

DOADDTASK inserts the task pointer in a free slot of the 50-entry task list.
DO_TASKS `0x5e9f` visits all slots and CALL `[BX]` at `0x5eab`. Task `0xebb`
uses WAITSYNCS 30: helper `0x5ac7` compares before increment, so from a zero
wait slot its body is reached on visit 31, not visit 30. It calls NEW_BALL
`0xece` and resets tasks/waits through WHEN_NEW_BALL_RESET `0x3abc`.

NEW_BALL clears LOOSING/BALL_DOWN, sets I_UTSKJUT, resets transient table
state and loads the current player record. It sets HOLDSTILL and initial ball
position (282,530), zero velocity. Active-game NEW_BALL_PART_TWO `0xf7d` queues
SOUNDNEWBALL (wait 50), SETBALL (wait 80), SOUNDBRICKUPP (wait 5); it enables
flippers and clears tilt state. SETBALL body `0x1003` places the ball at
(297,530), velocity (10,0), and clears HOLDSTILL at `0x103d`. Shared reset clears
mode/lamp transient state while retaining the saved progression/score fields.
This establishes the local task → preparation → active-ball sequence, not the
unproved global reachability of every matrix entry into it.

The normal branch bypasses the canonical 3/5-ball limit. **Whether any other
reachable demo graph path can reach canonical OUT_OF_BALLS before the timer
is not yet proved.** Remaining domains include XXBALLE/match, high-score
transitions, cheat roots, state-feasible matrix/task/effect roots. Their bounded producer closure is
described below; it does not prove that every retained branch can occur. `_BEATEN_MATRIX`
`0x57a6` locally compares current score with the seed, sets its beaten flag and
selects the shared celebration program; it does not locally write a .HI file.
That is not a proof of the entire high-score lifecycle.

### Additional consumed difference: expiry/task interleaving

DO_ELECTRONICS invokes expiry at `0x5cfd` but then continues area/target/shift,
KEYTASK and DO_TASKS (`0x5d16`) processing. It does not clear pending tasks at
expiry. A pending NEW_BALL_TASK can call reset `0x3abc`: unless PARTYFLASH is
true (`0x3afe`) or VISAKEYS is true (`0x3b10`), reset installs SHOWPLAYERSTS at
`0x3b3a–0x3b3d`, replacing the expiry matrix program. DO_MATRIX `0x4801–0x482c`
kills flashing and dispatches the new program unconditionally; there is no
priority guard that preserves expiry. A pending SETBALL can clear HOLDSTILL
at `0x103d` without checking expired. The scored-drain expired guard and the
unscored-drain branch also differ in their position relative to this flag.

These local replacement/release edges are proved. Exact reachable combinations
at equality 35998, other competing effects and the resulting eventual QUIT
must still be closed. A future model must not silently invent an atomic,
uninterruptible expiry transition. This is a concrete remaining BLOCKER 1/2
obligation and an addition to the consumed-difference audit.

## Persistence reachability (BLOCKER 4 remains open)

Read entry is **`0x66d2`**, after the preceding routine's RET at `0x66d1`;
write entry is `0x6706`, after RET at `0x6705`. Their reviewed bodies occupy
`[0x66d2,0x6706)` and `[0x6706,0x673d)`. Filename `table1.hi` remains at
`0x19dbc`; factory seeds remain equal A. No seed re-audit was needed.

Main CS ends at `0xaed0`, derived from the MZ load base and the relocated
CODE2 segment pointer at `0x56aa`; this corrects the earlier `0xafd0` bound.
The real MZ entry is `0x329f`, not the previous mid-startup root `0x32ad`.
A conservative raw E8/E9 rel16 scan of main CS `[0x300,0xaed0)` still finds no
incoming direct transfer into either body. Entry pointers CS:`0x63d2` / `0x6406`
occur in neither this CS extent nor DS `[0x19db0,0x29db0)`. Neither body appears
in the new direct-or-derived-candidate CFG. Historical DEMOVER guards omit
INIT_HIGHS/SAVE_HIGHS, corroborating that result.

The stronger negative theorem is **not proved**: the twelve TABLE1 UNKNOWN
sites below, INTRO far domains, CODE2 glyph callbacks and driver/launcher
lifecycle remain unclosed. Candidate non-reachability cannot exclude another
segment, computed pointer or aliased control-object write. The tool deliberately
retains `UNKNOWN` for both bodies even though the candidate CFG reaches neither.
**Persistence verdict: neither Variant A nor Variant B proved.** No native
load/save capability can be selected yet.

## Bounded target-domain resolver: current fixed point

[audit_10min_demo_domains.py](../tools/audit_10min_demo_domains.py) starts at the
real MZ entry and the previously proved primary/later callback registrations.
It follows direct conditional edges conservatively, slices register reaching
definitions, uses bounded save/restore summaries for direct callees, derives
callback-field writers and table grammars, adds candidate targets and repeats.
Unsupported definitions and writer/lifecycle assumptions remain explicit UNKNOWN.
This is a bounded overapproximation, not a general x86 executor or a proof of
state-feasible gameplay. The CLI exits **2** with unresolved domains; `--allow-open`
only permits exporting partial metadata and does not change the verdict.

[audit_10min_demo_programs.py](../tools/audit_10min_demo_programs.py) derives
58 command identities from the pinned historical declarations and linked typed
bonus stream. Structural label candidates are used to infer command identity,
never as an independent reachability root: numeric source drift and ambiguous
labels must not erase binary operands. Actual matrix/effect API consumers
supply roots; branch operand words supply successors.

Current fixed point: **seven rounds, 677 code roots, 14,197 instructions,
30 TABLE1 indirect sites**. Eighteen sites have bounded candidate target domains;
**twelve remain UNKNOWN**. This is a TABLE1 transfer-site count only: INTRO,
SDR admission and state/lifecycle obligations remain separately open, so a
global unresolved-domain count has not been established.

| Derived domain | Reproducible result | Remaining limitation |
| --- | --- | --- |
| matrix/control commands | 75 program bodies; zero unknown program heads/bodies; predecessors and branch operands exported | feasible state predicates, cursor/control writer aliases and eventual exits |
| effects | 40 API consumer/source records; every SI producer bounded | cross-effect ordering/state feasibility |
| cheat CALL BX `0x38fb` | 13 word-plus-dollar-string records, sentinel-derived bound | entry/lifetime effects must be included in global state proof |
| spring `0x4726` | targets `0x6182`, `0x61df` | downstream global state proof |
| DOTRUT `0x47c4` | 23 candidate field targets | writer alias/lifecycle proof |
| glyph CALL DX `0x7263`, `0x729a` | each masks byte index to 256 entries; 255 distinct code targets per table | external CODE2 glyph domain remains separate |
| task selector | 128 possible byte indexes; all lookup outputs and task words derived | all task-list writers not closed |
| area consumers | four / sixteen rectangle handlers derived | indexed writes/DS aliases not excluded |

Exact UNKNOWN TABLE1 sites:

| Sites | Unresolved domain |
| --- | --- |
| `0x66a`, `0xae9`, `0x5640`, `0x5673`, `0x574d` | far CODE2 pointer/segment domain, including its internal dispatch |
| `0x3a01`, `0x3d13` | far external callback/API binding |
| `0x47a7`, `0x5d0c` | PRINTTASK/KEYTASK initial-zero exclusion and initialization/admission lifecycle |
| `0x5eab` | indexed task-list writer domain |
| `0x6095`, `0x6151` | area index/DS-alias completeness |

**New consumed presentation difference:** real SHOWPLAYERSTS root `0x1b88e`
prints PLAYERSTEXT at position **336**, versus canonical declaration **340**.
FIRST_NO_OF_PLAYERSTS `0x1b89e` also uses 336. Both have consumer predecessors;
a similarly shaped retained source-label candidate at `0x1e2b3` using 340 is
not substituted for them. Future duration presentation needs an explicit
position field. This does not alter the closed INTRO card/menu results.

**Cheat gate:** `0x3947` tests DEMOMODE; `0x394c` branches past the cheat caller
when not in attract mode. CALL `0x3954` reaches the parser only through that
attract branch. Active-game commands instead cover tilt/music/pause; the demo
omits the retained active chute quit/start-player route. The thirteen retained
cheat targets include EARTHQUAKE (tilt disable), SNAIL (shift bit), EXTRA*BALLS
(binary NO_OF_BALLS value **5**) and FAIR*PLAY (reset to 3 and clear those flags),
plus presentation changes. They do not locally write timer/expired, PLAYER,
ordinary BALLS or XXBALLE. This is not a claim that their complete session
lifetime or every indirect entry has been closed.

XXBALLE direct candidate writers are initialization `0x331` (zero) and match
handler `0xd62` (FF). Establishing the zero invariant for reachable active demo
flow, rather than just listing those writers, remains necessary. Match,
high-score entry, zero-bonus/held snapshot lifetime, all exits and eventual
expiry behavior therefore remain open even though program syntax is decoded.

The parsed expiry body has QUIT at `0x1ba3b`, handler `0x3cdb`; its predecessors
are the threshold call `0x5cfd` and scored-drain replay `0x625`. The quit-question
program `0x1e2c3` has predecessor `0x34b2` and `_WAIT_YN` at `0x1e2cf` with
operand zero. Its answer callback/state domain is not a completed early-exit
proof. The complete set of feasible TABLE1 exits, including teardown/external
callbacks, is **not enumerated**. Consequently neither “all pre-threshold
termination impossible” nor “eventual QUIT guaranteed after equality” is proved.

## Consumed-difference list and research-only typed model

Established differences remain S_EMPTY, S_GAMEOVER2, PLAYERSTEXT duration label,
demo continuation program, expiry program, unavailable cards and demo menu
organization. The expiry replay/interleaving edges and duration-label position above are
additional consumed differences; the attract-only cheat/input gate is also
recorded as a future control capability. Persistence is a pending classification, not a proved
semantic difference. Physics/data materials, factory seeds and decoded audio
models retain their previously established equivalence; they were not re-audited.

The final exhaustive claim is **not made**: unresolved indirect graph domains
can still contain an unknown consumed difference. Direct graph metadata covers
8,375 TABLE1 instructions from the reviewed roots and 5,474 INTRO instructions;
these counts are reachability evidence for those roots, not whole-binary or
whole-native-consumer equivalence. The new resolver expands TABLE1 to 14,197 instructions, while keeping its
twelve unresolved transfer domains and separate scope gates visible.

Minimal proposed common representation, **research design only**:

| Capability / typed node | Known fields | Remaining gate |
| --- | --- | --- |
| required roles | the five pinned runtime roles | identity already proved; no profile registration |
| table availability | Party Land playable; TABLE2–4 advertising only | closed INTRO dispatch evidence |
| INTRO presentation | welcome/notice/actions/acquisition page, selector/options rows, option labels, card crop/windows/palette/reveal schedule | typed fields above; use demo sources |
| DemoBallContinuation | duration-label position 336; encoded display counter state; held snapshot restore; same-player save; QueueNewBall; ContinueProgram | full match/cheat/high-score root closure |
| shared new-ball preparation | existing reset/player restore, visit-based waits and queued tasks | retain DOS order; do not implement raw graph execution |
| CalculationLimit | counted event identity; uint16 initial 0; post-increment equality 35998; persistent table-process scope | external dispatch and exact native event mapping |
| DemoInputPolicy | attract cheat gate; active tilt/music/pause command path | complete entry/lifetime and exit proof |
| ExpiryProgram | clear/scroll/flash/score/wait/fade/quit typed nodes | task/effect preemption and eventual exit closure |
| HighScorePersistence | factory/transient versus load/save lifecycle | Variant A/B reachability proof |

Common matrix commands can retain the shared bonus prefix and have an explicit
demo continuation node; ordinary Session.changeBall must not implement it.
Task queues and program replacement remain common typed control concepts. Do
not encode an unconditional ShootAgain=false, reset-on-new-game counter, or
atomic expiry based on the current partial proof. This design is not yet
mechanical DMO1 input because the stated gates remain open. No production
source, descriptor, timer, importer, package or bundling code was changed.

## Reproducibility and regression boundary

New [audit_10min_demo_graph.py](../tools/audit_10min_demo_graph.py) verifies
exact five-file identity, canonical INTRO/TABLE identities and four historical
source hashes, derives the bounded bonus programs, checks reviewed instruction
consumers, parses menu/card presentation and exports direct graph metadata.
[audit_10min_demo_sdr.py](../tools/audit_10min_demo_sdr.py) additionally pins all
eleven driver identities and interprets EXEPACK blocks in memory. Capstone is a
research-only dependency; no runtime dependency or executable export is added.
The tools fail on private input mutation and print metadata/semantic records
only. They do not prove away an indirect boundary. The new domain resolver
derives tables/definitions and iterates to a fixed point; its default closure
check rejects the current result with exit 2. Program syntax and candidate
target completeness are separate from feasible state/alias/admission proofs.

Owner-local metadata: `/private/tmp/pf-dmo0-graph-evidence.json`. Reproduce with:

```sh
PF_10MIN_DEMO_DATA=/private/demo \
PF_RUNTIME_DATA=/private/canonical \
PF_DMO0_HISTORICAL_SOURCE=/private/historical \
python tools/audit_10min_demo_graph.py --output /private/tmp/pf-dmo0-graph-evidence.json
```

With the same three variables, reproduce the partial domain metadata with
`python tools/audit_10min_demo_domains.py --output /private/tmp/pf-dmo0-domains.json`.
The expected current exit is **2**, after writing explicit UNKNOWN records.
`--allow-open` suppresses that exit only for inspecting partial evidence.

Use a research Python environment with Capstone. The same three variables enable
[private graph tests](../tools/test_audit_10min_demo_graph.py); without them the
suite cleanly skips before reading private data or importing Capstone. Tests
cover linked program/tail, presentation consumer facts, explicit unresolved
indirect verdicts, all eleven registration producers, hash-mutation rejection
for TABLE1 and SDR, fixed-point growth, every glyph/table selector target,
register preservation/clobber handling, the position-336 difference and the
negative-persistence gate. This is static evidence, not a DOS gameplay execution
test; no test claims that open IRQ or state domains have closed.

| Current-pass validation | Result |
| --- | --- |
| New private graph/SDR research tests | PASS, thirteen tests |
| Research tool on pinned private inputs | metadata checks PASS; resolver closure gate correctly rejects with exit 2 |
| Public Python suite without private env | clean SKIP |
| A/B/C/D coherent detection and existing 1024 profile matrix | PASS |
| C/D seed mappings, Stones mappings/selectors, B/C/D priorities and shared construction | PASS existing focused private suite |
| Demo rejection and demo/full PRG hybrids across A/B/C/D | PASS |
| Go private suite with private env removed | clean SKIP of private tests; descriptor tests PASS |
| Production runtime/profile changes | NONE |
| Platform builds, native demo gameplay, timer smoke, bundling | not run; no production implementation |

Focused Go command is `./tools/go.sh test ./internal/datalayout ./internal/frontend
-run 'TestPrivate|TestPowerPackDescriptor|TestDeluxe' -count=1 -v`, with the five
existing private runtime env variables. The actual current-pass log is
`/private/tmp/pf-dmo0-final-go-tests.log`. No private test skipped in that run.
No existing oracle, capture, profile or factory seed was modified. Prior-pass
Windows/Linux build results are not claimed as current-pass validation.

## Remaining DMO0 obligations and checkpoint

1. BLOCKER 1: close the twelve enumerated TABLE1 transfer domains, INTRO/CODE2
   far domains, state-feasible match/high-score/cheat/control roots, held-snapshot
   paths and expiry task/effect interleaving; prove whether
   any alternate graph can reach canonical session end before 35998.
2. BLOCKER 2: finish all eleven SDR admission/IRQ semantics and the exact native
   ElectronicsCalculation event, including pause/reset/expiry boundary contract.
3. BLOCKER 4: close every indirect target domain and launcher/teardown path to
   read/write bodies; classify persistence as A or B.
4. Final consumed-difference exhaustiveness depends on those three closures.

BLOCKER 3 is closed as documented above. No unknown transition was replaced
with a native assumption. **DMO0 NOT CLOSED. DMO1 NOT STARTED.** No closure
checkpoint commit was made, because the user authorizes it only after all
semantic blockers close. Work remains local and reviewable.

Changed files in this pass are this report and six research Python files:
`audit_10min_demo_graph.py`, `audit_10min_demo_sdr.py`,
`audit_10min_demo_programs.py`, `audit_10min_demo_domains.py`,
`test_audit_10min_demo_graph.py`, `test_audit_10min_demo_domains.py`.
The payload scan result below covers all seven changed research files. This checks these sources, not all ignored workspace artifacts.
Scan PASS: 92 private PRG/MOD/EXE/SDR/BIN files, all eleven in-memory decoded SDR
modules, 1,705 distinct nontrivial aligned 4 KiB samples and whole-file identity
comparisons. Samples require more than 32 distinct byte values; this is a
sampled payload-copy check, not an all-substring or all-workspace proof.
Metadata is `/private/tmp/pf-dmo0-payload-scan.json`. Production runtime sources and
descriptors remain identical to research HEAD and the production base.
No executable chunks, images or decoded MOD audio
were exported by this pass; SDR decoded bytes remain in memory. Existing
untracked `.DS_Store` is unrelated and untouched. No push, tag or release.

### Concrete first-equality collision pass

The saved drain-35877 native reference now joins to its actual live task/wait
handoff and the accepted zero-aggregate demo suffix. See
[runtime-layout-demo-10min-first-equality-collision.md](runtime-layout-demo-10min-first-equality-collision.md).
NEW_BALL_TASK is inserted into slot 0 after scan 35967 with shared age zero;
it fires after expiry installation on scan 35998 with PARTYFLASH=VISAKEYS=false,
replacing expiry with SHOWPLAYERSTS. NEW_BALL_THRESHOLD_PROVENANCE = PROVED;
FIRST_EQUALITY_COLLISION_REACHABLE; ATOMIC_EXPIRY_MODEL = DISPROVED.
EXPIRY_INTERLEAVING = NOT_PROVED: eventual termination remains a separate pass.
The CLOSE1 extra DS:0x00d1 store is MUSICOK, not VISAKEYS; the latter's false
value is reconstructed from initial NEW_BALL/reset. DMO0 remains NOT CLOSED;
DMO1 remains NOT STARTED. No production support or publication follows.

### DURINGFLASH first-equality research

The latest narrow report is
[runtime-layout-demo-10min-first-equality-duringflash.md](runtime-layout-demo-10min-first-equality-duringflash.md).
`FIRST_EQUALITY_DURINGFLASH_REACHABLE`: fresh capture35886 → BEFOREFLASH35971
→ DURINGFLASH35998, explicit source HOLDSTILL clear, same-calculation late
movement, scored drain36173 restarting expiry, QUIT37212.
`EXPIRY_INTERLEAVING = NOT_PROVED`; smallest selected independent dependency is
first-equality admission/exclusion of the genuine arcade WAIT_FOR_SPIN_TASK
producer and its priority-clear/reward suffix. The five proved firing classes
do not establish an exhaustive pending-task census.
`DMO0 NOT CLOSED. DMO1 NOT STARTED.`
