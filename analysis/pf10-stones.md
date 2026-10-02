# PF10 — native Stones ’N Bones

Stones ’N Bones is the fourth native single-player table. F4 selects its original
attract presentation; F1/Enter starts it. Gameplay, bonus/new balls, match,
game-over, initials entry and return to Stones attract use the existing shared
frontend/session lifecycle. Down charges/releases the spring, held Shift/Ctrl/Alt
operate flippers, Space pushes and supplies tilt make edges, P pauses, Esc uses
the shared source abort flow, and M toggles music.

```sh
./bin/pinballfantasies -data-dir . -pf10
```

The public milestone flag delegates to the same `stones.Game` used by F4.
The established `-duration`, `-ticks`, `-release-at` and `-png` development
controls remain available; frontend smoke uses isolated `-high-score-dir` storage. `-pf2`, `-pf4`, `-pf7`
and `-pf9` remain intact. No CPU execution, DOS runtime, interpreter, JIT,
translated instruction execution or hardware emulator was added.

## Original identity and extraction

Identity comes from the banner in `reference/original-dos-source/STONES.ASM`,
its `INCLUDELIB STONES.LIB` and `CLEAR.LIB`, `MODUL 'TABLE4.MOD'`, and
`hi_score_file 'table4.hi'` declarations (lines 4, 13–14, 42 and 52).
The linked PRG, rather than an assumed filename-to-table relationship, supplies
its artwork, masks, descriptors, fonts and animation streams. Shared source is
`FANTASIE.ASM`, `FANTASIE.MAC`, `BALLCODE.ASM`, `MACROS1.ASM` and `MACROS3.ASM`.
The original linked library contents are recovered from the supplied PRG;
the port does not require a DOS linker or execute code to extract them.

| File | Bytes | SHA-256 |
|---|---:|---|
| TABLE4.PRG | 522198 | `88f63edd4c7b50bd057397016d7aa962f0ed1c858f4a746f1ccf976f67494ebf` |
| TABLE4.MOD | 216418 | `31ad7e671ae77c07c3d075e2f1fecd3d918fd921fa23acd9a1b0b6fc07fbbcea` |
| TABLE4.HI | 64 | `10ca2e02da5333360a40c1dfa93547c5c9633ad159d42dee67789b9139850645` |

Independent extraction precedes the native consumers:

```sh
python3 tools/reference_pf8.py
python3 tools/reference_pf10.py
```

These read ASM and original bytes only. They generate `analysis/pf8-content.json`,
`analysis/pf10-assets.json`, and native literal content declarations. Expected
bitmap, sample, mask, flipper and asset hashes are calculated independently in
Python. Rule expectations use source constants and independently stated decimal
arithmetic. No fixture is a score or state recorded from a native smoke run.

## Assets and physical configuration

Offsets in this section are hexadecimal file offsets in the pinned PRG.

| Item | Stones source/data value |
|---|---|
| Logical playfield | 320×576; BANH=576, line 75 |
| STAGE strips | `4bc10`, `54a00`, `5da70`, `66e20` |
| PBM dimensions | first three 320×144; fourth 320×1219 |
| Initial view | rows 259..575; 320×317 field plus 320×33 matrix |
| SETBALL | (297,530), velocity (10,0), lower layer |
| Hidden new ball | (282,530); source SETBALL waits 80 visits plus completion |
| STARTX/STARTY | (302,535); lock teleport uses these unadjusted coordinates |
| Ball artwork | shared literal ball stores, Stones linked offset shift +`7a0` |
| HID1/HID2 | `28eb0` / `2edb0`; each 23040 bytes |
| MASK12/11/22 | `347b0`, `3a1b0`, `3fbb0`; each 23040 bytes |
| MASK13/21/23 | `6e870`, `74270`, `79c70`; each 23040 bytes |
| Flipper descriptors | `1da30`; two entries, followed by kind-zero sentinel |
| Flipper origins | (80,510), (160,510); 53 rows, 20 frames each |
| Collision frames | `455b0`, `47bd0` |
| Flipper delta segment | `ca80`..`166d0` |
| Flipper artwork | `7f670`, 120 bytes, four planes with 30-byte stride |
| SIN / MAT_TABLE | `1b2c0` / `18e77`; independently equal to shared lookup values |
| Spring | `7f6f0`, original 10×23 artwork; destination (304,556) |
| Spring valid region | (305,544)..(320,576) |
| Spring invalid region | (300,520)..(315,540) |
| Tasks / flash slots / WAIT capacity | 20 / 64 / 50, STONES lines 64–66 |
| Lamp packets / attract flash records | 44 / 19 |
| DATA / DATA2 | `166d0` / `1daf0` |
| Font 13 / 11 / 8 / 5 | `1cec0`, `1d0d0`, `1d290`, `1d3d0` |

The fourth PBM's complete ByteRun1 BODY decodes all 1219 rows, and the independent
fixture hashes that complete strip. The table's four 144-row field portions form
BANH=576. Extra linked PBM rows are not substituted for gameplay geometry. Their
asset-production origin is not established; no speculative artwork repair or
replacement is applied.

Eleven effective ramp gravities are `(0,10), (-10,5), (0,-10), (5,0), (5,15),
(-10,12), (2,15), (-8,12), (3,10), (4,13), (7,10)`. The exact lower/upper
transition rectangles, nine STONE/BONE contacts, three bumpers, two slingshots,
and all approach/capture rectangles are decoded or transcribed from the original
lists in source order. `physics/stones.go` contains this table configuration;
`pf10-assets.json` retains independent area and mask records. The first sync is
X=297×1024+30, Y=530×1024+24, VY=26, GY=10. Full spring charge with zero jitter
releases VY=-166×32=-5312. Gameplay uses signed integer BALLCODE throughout.

Six moving collision patches reuse `physics.PatchMask`: GATE2 (bumper entrance,
byte30,row5,2×20), GATE3 (Tower, byte20,row5,2×24), GATE4A/B/C (upper grid ramp,
byte4,rows191/247/303,4×14), and GATE5 (Vault/kickback, byte0,row511,4×13).
Independent extraction follows the linked positive/negative data and
MOVE_MASK_DATA_B's width-sized copy with two widths skipped between rows.

## Four-table portability audit

**A — reused unchanged:** integer integration, contact/material response,
flipper mechanics/delta application, layer and scroll machinery, indexed
foreground/ball composition, VGA DAC conversion, fonts, scrollers, delta bitmap
renderer, decimal arithmetic, task allocation/run/WAIT primitives, jingle clock,
tracker decoder/effects/mixing/interpolation/fixed-point timing, voices 0/3 left
and 1/2 right, shared push/tilt mechanics, launcher equation, SDL input/window/
audio backend, frontend lifecycle and 64-byte high-score storage.

**B — changes demonstrated by table4:**

- PBM decoding accepts an explicit recovered strip height. Stones' last linked
  strip is 1219 rows; the other tables retain their existing 144-row decoding.
- Flipper descriptor decoding recognizes the original kind-zero end marker.
  Stones has two flippers; there is no fabricated third flipper. The fixed
  storage's remaining descriptor is zero, while shared dynamics remain intact.
- Matrix data extraction accepts double-quoted DB strings and repeated RGB
  triples. STONES PL_TEXT is double-quoted, and MATRIXOFF repeats a triple three
  times. DATA2 declaration order and animation-based CLEARIT reuse PF9 support.
- The shared pixel consumer gains `TowerWindow`: STONES showtower (lines
  5790–5856) uses packed artwork at `4a1f0`, rather than a delta frame. Each set
  bit writes MATRIXLO, the lit index. The Stones scheduler owns its row counter;
  drawing never advances a second clock.
- Matrix palette binding accepts source RGB/off packets. MATRIXLO=79 and
  MATRIXHI=231; MATRIXON is 95,83,20 percent. MATRIXOFF writes raw DAC 12,17,21
  into entries 79,80,81, overlapping Mummy colors. Stones applies these at
  command/flash boundaries in chronological order. Rendering does not mutate
  lamp or gameplay state. F1–F3 retain their accepted palette paths.
- Factory/display and the existing direct-table adapter register table4.

**C — table-local:** targets, KEY/RIP rotation guards, ghost progression, Tower
Hunt, Ghost Hunt, Grim Reaper, Multi Demons, captures/locks, award ordering,
mask selection, timers, matrix semantic dispatch, bonus flow and cleanup remain
in `internal/stones`. No full table implementation was copied and mutated.
No new physical equation, tracker command or scheduler primitive was necessary.

Approximately 1700 formatted handwritten runtime lines are table-local Stones
rules/presentation/capture/audio adapters; the roughly 500-line Stones test file
and generated declarations are separate. By type, about 85% of handwritten
runtime additions are table-local content/rules, with the remainder asset/config
loading and registration or narrow presentation/descriptor support. This is an
approximate code-size assessment, not a claim of generic pinball support.

## Major native state machines and scoring

Source PLAYER_STRUC restoration preserves score/bonus, partial RIP and
STONE/BONE banks, completed ghost progression, skill score, screams/next award,
and kickback. Per-ball reset clears transient timers, guards, captures, Tower
awards/values, mode totals, multiplier and mode flags. Source names retained in
content and events allow tracing back to declarations and handler labels.

| System | Trigger, source state and result |
|---|---|
| STONE/BONE | Nine source contacts. Four STONE targets give 27530/510 bonus; five BONE targets give 17520/750 bonus. Completion gives 100000, increments jackpot, lights the next ghost/Vault opportunity, then executes WAIT2 and WAIT70 reset tasks. Repeat guards and GhostInhibit suppress duplicate collection. |
| KEY / skill shot | Three rollovers give 10060/1010. Lit random SkillKey gives accumulating million skill score and increases Vault/Tower/Jackpot/Well values. First skill consumes the selector. Completion opens Tower and advances its award sequence: Million, Five Million, Double Bonus, Hold Bonus, then Five Million. WAIT10/70 and flipper rotation guards match the source. |
| RIP / kickback | Three rollovers give 10070/1080. Complete group lights kickback, patches GATE5, and resets after WAIT70. Individual lights use WAIT20. Lower Vault entry consumes kickback; source high-Vault approach preserves it. |
| Scream | Gives 10060/1050; first count changes 1 to 2. Every ten lights Tower extra ball at ten, then Five Million. Double Screams repeats through WAIT2. Source numeric fields are updated in encoded text. Each scream adds 100000 end-of-ball bonus. |
| Left grid ramp | Gives 10030/1040, adds Grim/Vault contributions, lights Million Plus and Double Screams for 450 syncs and multiplier opportunity for 570 syncs. At 90 remaining, lamps switch to fast flash. Upper GATE4 patches follow each bend. |
| Loop / combo | Loop gives 10030/1020; repeated loop inside 300 syncs gives Million. Grid → Scream → loop inside the 780-sync combo interval awards source LOOPCOMBO. |
| Million Plus | CLOSE1 collects accumulated SULP_SCORE in million steps, plus 10000/1000 ordinary score, closes the bumper gate and resumes Multi Demon countdown. |
| Multipliers | Lit Well opportunity advances 1→2→4→6→8→10 and the five corresponding source bitmap programs/lights. |
| Jackpot | Starts 10000000; qualifying contacts add 100000. Lit Tower collection takes the current value, resets it and enables Super Jackpot for WAIT780. |
| Super Jackpot | Forced-priority Tower award of 50000000, with source matrix/jingle and ordinary award interactions retained. |
| Extra ball / shoot again | Table-local ExtraBalls counter; collected Tower extra balls and top-score awards are consumed before normal ball increment. Match also supplies an additional played ball through source programs. |
| Double / Hold Bonus | Tower collects in source order; Double duplicates the current decimal bonus after effect bonus additions. Hold restores the captured post-calculation bonus after countdown. |

Completing STONE/BONE enables one of eight independent ghost systems, in the
original event-table order. The Vault collects the enabled ghost:

1. BATMAN: Five Million.
2. TOWERHUNT: 2400-sync hunt; Tower stages award Five, Ten, Twenty Million;
   music return advances through source orders up to 50; capture suspends counters.
3. SMILER: lights Tower extra ball.
4. REDDEVIL: Ten Million.
5. GHOSTHUNT: source intro/countdown/ending commands; qualifying target/bumper
   contacts add 1000000 into OR_TOTAL for end-of-ball bonus; enables jackpot.
6. MULTIDEMONS: 2100-sync lock opportunity; locks at Well/Vault raise Scream
   collection from Five to Ten to Twenty Million. Source holds/teleports a single
   ball to STARTX/STARTY, rather than inventing simultaneous multiball physics.
7. MUMMYHEAD: Fifteen Million.
8. GRIMR: source intro/50-second countdown/ending; loops, ramps and traps add
   5000000 into TM_TOTAL; enables jackpot.

After eight ghosts, all ghost lamps flash and WAIT240 clears/inhibits the group.
Starting a timed mode while another is special uses a cooperative deferred task;
its matrix program starts after SPECIALMODE clears. Matrix commands control
special/offroad/turbo enable/disable, countdown and cleanup. Captures retain
mode countdown state and resume the appropriate source BACK_2 program on eject.

The complete 49 effect records, their static decimal score/bonus, matrix program,
priority and jingle are exported in `internal/stones/content.go` and
`analysis/pf10-assets.json`; no manual or remembered rules supplied them.

## Captures, tilt and lifecycle

Tower holds (141,143), upper layer; eject is (0,-4000). Well holds (275,245),
lower layer; eject is (-800,2000). Vault holds (2,532), lower layer; eject is
(0,-2880). Re-entry guards and LASTCHECK suppress repeated captures. Matrix
_TOWEREND/_WELLEND/_VAULTEND schedule source WAIT10 releases; Vault schedules
WAIT30 collision cleanup. Tilt uses the source immediate/scheduled release paths
and prevents resuming a terminated mode.

Tower evaluates Hunt, Super Jackpot, Jackpot, Extra Ball, Double Bonus, Hold
Bonus, Five Million, Million and ordinary Tower value in original order.
INH_EFF suppresses replacement presentation while retaining decimal additions;
EFFECTBRACK preserves the forced jackpot paths. Well/Vault locks use source
STARTX/STARTY, source shoot text and music return, lock lamps, paused Multi Demon
timing, and the Vault teleport presentation guard. No arbitrary capture timers
replace original tasks.

Shared tilt mechanics count Space make edges, retain physical push and warning,
and disable electronics/flippers on tilt. Stones clears table lamps and
uses TILTTS plus capture cleanup. Space remains independent of spring release.

LOOSE_BALL holds (140,210), forces viewport 259, disables flippers, terminates
special state, removes Grim lamp, starts the source loss matrix and schedules
WAIT5 drain sound. Ordinary bonus multiplies first, then adds Screams×100000,
OR_TOTAL and TM_TOTAL; decimal digit countdown adds to score every four visits,
including the final source delay. Tilt takes the no-bonus path. Hold Bonus and
ExtraBalls are evaluated at the source transition. NEW_BALL_TASK is WAIT30;
new ball is initially hidden, sound is WAIT45, SETBALL is WAIT80. The match
program has 36 independently transcribed delay entries and a descending modulo-ten
selection. Game completion hands off to the shared game-over/high-score
presentation, then Stones attract and selector.

Top-score qualification uses the table4 top record, strict score comparison,
and source chute guards. BEATENTS/BEATEN_BH_TS award the source extra ball;
ball-loss qualification bypasses ordinary delay through its source branch.
Frontend initials entry writes only TABLE4.HI, using the unchanged four-record,
64-byte codec. Tests use temporary stores and assert TABLE1/2/3 are untouched.

## Matrix, DAC and audio audit

625 gameplay command records and 58 expanded attract records are extracted.
Fonts, scrollers, numeric fields, clears, title, match, bonus and game-over use
original DATA programs and fonts. DATA2 begins at `1daf0` before HID assets.
Thirteen animations provide 137 independently hashed delta frames: CLEAR,
BONUSX2/4/6/8/10, JACKPOT, EXTRABALL, MILLION, FIVE/TEN/FIFTEEN/TWENTY MILLION.
Tower additionally uses the packed window stream, tested at rows151,73,47,20.
The renderer only consumes existing scheduler/frontend timing.

LON packets are applied in original sequence. OFF halves the converted six-bit
DAC values. Flash slots retain allocation order; shared DAC indices are never
resolved by sorting lamp numbers. MATRIXOFF's overlap with Mummy is explicitly
covered by a later-write-wins test. Original indexed art/palettes are retained;
normal gameplay has no debugging HUD or modern typography.

42 jingle declarations and source sound declarations bind directly to original
MOD orders/samples/notes. Effect/jingle priority and silent cue clocks reuse
shared timing. Source TOUCH1 volume40 and spring charge×2 volume are retained.

TABLE4.MOD declares 66 orders and actually references 66. Highest stored pattern
is63, so the 64-pattern physical extent ends at byte66620 (`1043c`). Its 31
sample lengths sum149798; 66620+149798=216418, the exact physical file size.
Every sample has an independent hash. Source jingle orders reach64; empty order
is52. Unlike TABLE3.MOD, there is no hidden order extension/header trap. Observed
effects are 0,1,2,3,4,6,9,A,B,C,D,E,F; all use existing tracker support. Mixer,
interpolation, fixed-point clocks and Paula stereo routing are unchanged.

## Regression, visuals and runtime validation

Required commands were run:

```sh
./tools/go.sh test -p=1 -count=1 ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -data-dir . -pf10 -ticks 1200 -release-at 0 -png /tmp/pf10-direct.png
python3 tools/pf8_isolated_x.py python3 tools/smoke_pf10.py
PF10_AUDIO_DRIVER=pulseaudio python3 tools/pf8_isolated_x.py python3 tools/smoke_pf10.py
python3 tools/pf8_isolated_x.py python3 tools/pf8_selector_audio.py ./bin/pinballfantasies ./analysis/pf10-runtime-validation
```

Build succeeds, and the direct `-pf10 -duration 2s` SDL/PulseAudio window smoke
also passes with zero queue resets/empty observations. This historical suite predates factory-seed inventory validation; see
`personal-clean-seeds.md` for current full-suite results. Installation/user score files were never rewritten;
their hashes were checked before/after smokes. Party Land's established 1200-tick
script still gives score2300000, ball2; Speed Devils, Gameshow, PF8 and stereo
regressions retain their original expectations.

Focused Stones fixtures cover assets and dimensions, all bitmap frames/samples,
SETBALL/first tick, first flipper frame, two-flipper geometry, ramp/layer/contact
identities, spring saturation/release, STONE/BONE waits, KEY/RIP, all eight ghost
systems, Tower Hunt, multiplier/captures/ejects/locks, jackpots, extra/held/double
bonus, mode totals, screams/combo/timeout, high-score/chute guards, tilt, bonus,
new-ball delay and full frontend match/game-over/initials persistence. Shared
mask patch regressions remain authoritative. The 1200-tick Stones result26070,
ball2 is **smoke only**; no independent full-game scripted score oracle is claimed.

`analysis/pf10-checkpoints` contains 320×350 logical PNGs for original attract
title, initial chute, charged spring, rule text, delta bitmap, Vault capture,
tilt, bonus, game-over and high-score entry. They are visual/layout checkpoints,
not a claim of DOS temporal pixel parity. Attract title selection uses the actual
source PRINT13 program. Existing PF8 timing uncertainties remain accepted.

Four-table X11 smoke passes on an owned Xvfb display, including F1/F2/F3/F4,
launch/flippers/push, pause/resume, abort and selector return. Window identity
survives every transition and manual resize from960×1149 to887×1102 survives.
Temporary aborted sessions write no score records. Dummy-audio smoke has zero
queue resets (empty queue observations are not an audio-quality measure there).
Real PulseAudio reports48000Hz signed16 stereo, zero queue resets and zero empty
queue observations. Forty explicit pause/session lifecycle clears and24 starts
are expected from this control script. The separate selector visual-transition
probe has zero lifecycle clears, one start, zero queue resets/empty observations;
its monitor WAV has distinct, nonzero left/right channels. Logs, timeline and
monitor audio are in `analysis/pf10-runtime-validation`.

## Files and final architecture boundary

Added: `internal/stones` native subsystem/content/tests; Stones asset and physics
loaders/tests; frontend Stones lifecycle test; `tools/reference_pf10.py`;
`tools/smoke_pf10.py`; independent JSON fixtures, logical checkpoints and runtime
validation artifacts; this report.

Changed: shared strip decoder, flipper end-marker decoding, presentation packed
Tower/source palette support and generated content/extractor; TABLE4 MOD identity
adapter; frontend runtime/display registration and F4 expectation; direct-table
CLI adapter/public flag; README. Shared physics equations, scheduler and tracker/
SDL/high-score storage implementations remain unchanged.

All four original single-player tables now use the same proven engine and
lifecycle. Remaining table4-only data differences are the tall linked fourth
PBM, two-flipper sentinel, 30-byte plane stride, packed Tower presentation,
64 flashes and source matrix palette/overlap. None requires a new physical or
scheduler primitive. This demonstrates portability across these four original
tables; it does not establish a generic engine for unrelated games.

Deliberate boundaries: single-player only; accepted PF8 presentation/timing
uncertainties; logical checkpoints do not establish DOS temporal parity; the
extra fourth-PBM rows' production origin is unknown. No options UI, remapping,
mouse/controller UI, video mode selector, modern graphics or audio redesign was
implemented.

## PF11 settings inventory

INTRO.ASM TOGGLAR_STRUCEN (lines938–947) defines the six-byte PINBALL.CFG:
S_BALLS (3/5), S_ANGLE (high/low), S_SCROLLING (hard/medium/soft), S_IM
(in-game music on/off), S_RESOLUTION (normal/high), S_MODE (color/mono).
The labels and update branches are at1052–1083 and4509–4609. SOUND.CFG contains
source device configuration, separate from that six-byte options record;
FANTASIE retains mouse-driver/input paths and runtime M music toggle.

Currently the native shared frontend reads the existing six-byte file's ball
setting and applies 3/5 to all four sessions. Each session exposes music toggle;
original MOD effects and jingle playback remain enabled. Gameplay composition is
fixed to accepted high mode; angle/scrolling/color and mouse selection are not
newly exposed or implemented by PF10. PF11 can start at this precise options
boundary; no F5/settings implementation is included here.
