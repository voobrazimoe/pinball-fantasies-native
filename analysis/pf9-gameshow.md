# PF9 — native Billion Dollar Gameshow

Billion Dollar Gameshow is playable from frontend F3. F1/Enter starts one
player. The same implementation is available directly:

```sh
./bin/pinballfantasies -data-dir . -pf9
```

Down charges/releases the original launcher; Space pushes the table and supplies
tilt make edges. Held Shift/Ctrl/Alt operate flippers. P, Esc and M use the
established frontend/control paths. F4 remains unavailable. No DOS/x86 execution,
interpreter, translated execution, JIT or hardware emulator is present.

## Original specification and independent extraction

The table identity is established by the banner and MODUL/hi_score_file
records in **reference/original-dos-source/SHOW.ASM**, not inferred from a file
name. SHOW links `INCLUDELIB BILLION` and `INCLUDELIB CLEAR` and declares the
STAGE, HID, MASK, flipper, gate and animation external data symbols. The supplied
linked **TABLE3.PRG** contains those assets. **TABLE3.MOD** supplies the tracker
patterns/samples; **TABLE3.HI** supplies Gameshow's original 64-byte seed.

Shared specification is **FANTASIE.ASM**, **FANTASIE.MAC**, **BALLCODE.ASM**,
**MACROS1.ASM** and **MACROS3.ASM**. Relevant SHOW sections are definitions,
effects/sounds, detection lists, FLASHLIST/LON packets, matrix declarations,
RESET_VARS, LOOSE_BALL, NEW_BALL, gate routines, target/ramp handlers,
UPDATE_COUNTERS, and the final DATA2 animation declarations.

Pinned PRG SHA-256:
`da83ef5a7a471e6a6ad759126907076c81e92ffde6dec8e3de8e6052c6a98858`.
Pinned MOD SHA-256:
`fb7bfd1c96a462cb03999d2e6f843a20d3de69ba05fcbd384a9f1c131b9a563a`.

Extraction precedes native consumption:

```sh
python3 tools/reference_pf8.py
python3 tools/reference_pf9.py
```

The first extractor now includes SHOW alongside the unchanged PLAND/SDEV
records. The second produces `analysis/pf9-assets.json` and the native SHOW
content declarations. Both read original ASM/data only. Static linked references
locate literal artwork/data; machine instructions are never executed. Bitmap
fixtures independently apply original two-plane delta streams and hash every
frame. The fixture producer does not call Go or use native output as an oracle.

## Playfield, geometry and table capacities

All locations below are hexadecimal file offsets in the pinned PRG.

| Content | Recovered value |
|---|---|
| STAGE1_1..4 | 4cb60, 52410, 5a6b0, 634d0; four PBM 320×144 strips |
| Full field / high-mode composition | 320×576; 320×317 field plus original 320×33 matrix |
| Initial viewport | rows 259..575; RASTERPOS=(259+33)*16 |
| SETBALL | (299,530), velocity (10,0), low layer |
| Hidden new-ball position | (284,530); SETBALL waits 80 visits plus its completion visit |
| STARTX / STARTY | (304,535); Billion lock returns to these unadjusted coordinates |
| HID1 / HID2 | 1f870 / 256f0; 23040 bytes each |
| MASK12 / MASK11 / MASK22 | 2b0f0 / 30af0 / 364f0 |
| MASK13 / MASK21 / MASK23 | 6a3f0 / 6fdf0 / 757f0; 23040-byte planes |
| Flipper descriptors | 1f230; masks 3bef0, 3f4a0, 3e510 |
| Flipper delta segment / artwork | b570 / 7b1f0; four planes, 55-byte stride, 220-byte graphics |
| Third flipper | right-key flipper, (240,176), 51 rows, 12 frames; descriptor overrides the unused F3HEIGHT=53 definition |
| SIN / MAT_TABLE | 1ca40 / 1a93d; same values as prior tables |
| Effective ramp gravities | (0,7), (4,9), (0,11), (2,6), (6,10); NO_OF_RAMPS=5 excludes trailing dummy row |
| Spring artwork | 7b2d0; original 10×23 source; destination (304,556) |
| Spring enable / disable regions | (305,512)..(320,576) / (300,400)..(320,450) |
| Bumpers | (44,145)..(68,169), (74,201)..(98,226), (11,231)..(35,247) |
| Target contacts | six original dollar/B/C rectangles from ZonLista_L |
| Layer transitions | five upper entries, ten lower entries, in original list order |
| Tasks / flash slots / WAIT capacity | 20 / 30 / 50 source call-site counters |
| Lamps | 38; original packets; matrix on=153, lattice=114 |
| Gates | eight mask patches; gates 1/7 are upper; others lower |
| DATA2 matrix tables | 41ac0, after collision-frame segments; different segment ordering from F1/F2 |
| Fonts 13 / 11 / 8 / 5 | 1e640 / 1e850 / 1ea10 / 1eb50 |
| Match | 15 banks, 14 syncs between steps |

Gate data is independently extracted from the linked open/closed references.
Gates 7/8 use MOVE_MASK_DATA_B's width-sized copy followed by two widths of
source skipping, with the caller's initial width adjustment. The existing
`physics.PatchMask` handles these after extraction normalizes their rows.

## Portability audit

**A — reused unchanged:** BALLCODE signed integer integration and response,
contact sampling/material lookup, physical flipper dynamics and graphics deltas,
layer/scroll mechanics, indexed foreground/ball composition, VGA DAC conversion,
matrix bitmap/font/scroller drawing, task allocation/run/WAIT primitives,
decimal arithmetic, jingle clock, tracker effects/mixer/interpolation/fixed-point
sample progression, Paula stereo routing, SDL input/queued audio/persistent
window, frontend session ownership and 64-byte high-score codec/storage.

**B — demonstrated generalizations:**

- Presentation extraction accepts SHOW's DATA2 placement, split command words,
  LABEL-before-frame declarations and animation-based CLEARIT. Attract and
  postgame replay consume those original bitmap clears using existing animation
  timing. Existing table data and PF8 golden fixtures are unchanged.
- Display instances own their text maps. SHOW's dynamic player fields cannot
  mutate another session's source content.
- The native MOD loader accepts a verified table-provided order extent.
  TABLE3.MOD reports **61** orders, but its stored order list also contains
  orders 61→pattern62 and 62→pattern63. SHOW KNACKRUT1 explicitly writes
  LASTJINGLE=62, also verified in the linked PRG literal at file 9f1.
  Therefore Gameshow decodes **63 orders / 64 patterns**, with samples starting
  at **1043c**, not fc3c. The shorter header-only interpretation misplaced all
  samples by 2048 bytes. The physical file length equals that corrected pattern
  extent plus the independently summed 153048 sample bytes. Order62's empty
  pattern falls through to order0; no new MOD command is invented. F1/F2 retain
  their existing decoder paths. No tracker/mixer algorithm changed.
- The existing direct-table CLI adapter serves F2 and F3, preserving `-pf7` and
  all earlier flags. Frontend factory/display registration adds table3.

**C — Gameshow-specific:** prize flags/progression, dollar/drop contacts, eight
gates, wheel state/timing, cash pot, jackpots, skills, Money Mania variants,
Billion lock/award, counters, mode cleanup, capture tasks, and matrix semantic
command dispatch stay in `internal/gameshow`. Approximately 1,400 formatted
handwritten lines provide table-local logic/adapters; generated declarations and
roughly 500 lines of focused rule tests are separate. There is no generic rules
DSL and no copied physics, tracker, bitmap/font renderer, SDL backend, score
store or frontend lifecycle.

Two-table assumptions exposed were animation segment/declaration shape, clear
program shape, and the MOD song-header extent. No new physical equation or
scheduler primitive was needed. Common fixed geometry (576-row fields, three
flippers and spring destination) is now confirmed for three tables, but this
report does not assume it applies to Stones 'N Bones.

## Gameshow rules and state

- Dollar targets award 7500 / 510 bonus. Completing the pair enables the wheel
  gate/lamp, with independent 25/60-WAIT lamp tasks. B/C drop targets award
  7500 / 530 or 570 bonus, change original gate collision masks, and reset their
  two-target banks after the source 60-WAIT task.
- Original lower/upper AreaLista records preserve LASTAREA approach history and
  LASTCHECK repeated-entry suppression. CLOSE1 following OPEN1 changes from chute
  to main music/game presentation and closes the launcher gate. Flipper make
  edges do not rotate target groups: SHOW CHECK_SHIFTKEYS only clears its flag.
- TV, TRIP and CAR light from their source timed approach combinations. Lit
  prizes enable collect-prize and the wheel gate. YOU_WIN collects them in source
  order for 500000, 1000000 and 3000000, each with 50000 bonus. Completing this
  group establishes TopThree and the timed jackpot opportunity.
- BOAT/HOUSE/PLANE use the second-group approach windows, original side hints and
  loops. Collection awards 5000000/250000, 7500000/500000 and 10000000/500000.
  AllSix enables the lock/Billion sequence. Partial group flags reset between
  balls; completed TopThree/AllSix groups restore their won lights. Skills,
  decimal score and bonus use the source player-state boundary.
- Jackpot starts at 10000000 and adds 100000 at source contacts. The right ramp
  collects it and lights Super Jackpot; the clockwise ramp collects 50000000 /
  1000000 bonus. Loop Million uses its own 600-sync window and 1000000 / 75000.
- Skill count deliberately changes first 1 to 2, matching ANOTHER_SKILL's jump.
  Six skills starts target Money Mania; twelve lights extra ball; later sixes
  alternate target and loop/trap variants. They accumulate 500000 or 1000000 in
  TM_TOTAL, counted in ball-loss bonus. Intro/25-second countdown/ending programs
  retain SPECIALMODE until their ending command performs cleanup.
- Multiplier advances 2,3,4,6,8,10 from BONUSTABLE, through original animation
  programs and source priority checks. Raising Millions mutates the original
  decimal digit separately from multiplier advancement.
- Cash pot resets to 500000, adds 7130 on source contacts, awards its current
  decimal amount at capture, and applies the timed x5 branch. Capture holds at
  (103,233), uses 160/40-WAIT tasks, then ejects at velocity (100,1700). Special
  mode uses the distinct 30-WAIT release path.
- Wheel capture holds at (4,529), fixes viewport to 187 (SCREENFORCE2=220 includes
  SPLH), seeds its sector from the source random counter/prize branch, and runs
  the exact SPINTIMES_HI slowdown sequence. Ordinary sectors award 25000 through
  5000000 after the 100-WAIT end task. Prize animations/scrollers schedule that
  end task through _END_OF_SPIN rather than awarding/ejecting in the renderer.
- After AllSix, cash capture returns the ball to (304,535) and lights Billion.
  Wheel capture then awards 1000000000 / 5000000 bonus, holds for the original
  250-WAIT cleanup, resets both prize groups and ejects upward at -3500.
- UPDATE_COUNTERS preserves source expiry jumps and early RETN ordering. SHOW
  does not decrement HOUSEcounter/PLANEcounter. These are retained source
  behaviors, not normalized timer policy.

## Ball loss, bonus, tilt and ending

LOOSE_BALL holds at (135,28), forces bottom scroll, disables flippers, exits
SPECIALMODE and starts BALL_LOSTTS. Its matrix task sequence multiplies ordinary
bonus first, adds Skills×100000, then Money Mania total, and counts decimal digits
into score via DO_FLORPA's four-sync cadence and final delay. Original source
branches skip empty components. No Party Land no-score re-shoot branch was added.

Shoot-again uses lamp31 and retains ball number. The ordinary next ball waits
60, resets table-local temporary electronics, holds at (284,530), and schedules
sound/SETBALL tasks. PARTYRUT remains active through shoot-again/new-ball chute
presentation until GAME_ON replaces it. Final balls run the 15×14 match sequence;
a matching tens digit gives the source match ball. _DOBEATEN also lights31 once,
with chute/special guards for the gameplay check and the source ball-loss bypass.

Tilt uses unchanged make-edge/physical-push primitives. SHOW's local cleanup
clears flashing/lamps, disables flippers, plays its Tilt cue and installs TILTTS.
Its tilt detection-area list still permits wheel-hole capture/ejection; the
30-WAIT tilt release prevents a stranded ball. The next ball resets tilt.

The table reports its final decimal result at _2_DEMO_MODE. Shared frontend
ownership then performs qualification, initials, isolated TABLE3.HI persistence,
entry wait, Gameshow postgame/attract and selector return. No table-local high-score
UI or storage implementation was introduced.

## Matrix, lamps and audio

All 18 original animation programs are extracted and independently checked:
prize-side hints, Money Mania, Multiply/x2/x3/x4/x6/x8/x10, CashPot/CashPot5,
Million, Billion, YouWin, ExtraBall, Jackpot and Clear. Text/scrollers, dynamic
numbers, original fonts, score glyphs and bitmap deltas use PF8's renderer.
SHOW's split SPINTS number pointer is resolved as native data identity, not as an
instruction pointer. Commands/timers/tasks own gameplay state; pixels cannot
award scores or advance a second gameplay clock.

Each LON/LOFF call immediately applies its original packet in emitted order.
Frames consume the resulting palette; they do not sort/reconstruct active lamps.
SHOW's 38 lamp packets occupy disjoint DAC ranges, independently checked during
extraction. Source scaling and six-bit VGA conversion remain unchanged.

38 jingle records and 41 effect records come from SHOW declarations. S_SPRING=0,
S_MAIN=1, S_NOHIGH=18 (DEMO_MUSIC return9), Danger=44, Tilt=43, LostBall=30,
Mystery=49, Billion=23, MoneyMania=38, and its end=41. Effects use original
sample/note declarations, including Bygel1's explicit volume32 and spring's
charge×2 volume. The module's tracker commands are already supported; none were
added. Silent/audible clocks and independently extracted sample hashes are
verified. Music remains voices 0/3 left, 1/2 right; existing effect presentation
and mixer behavior are unchanged. Gameplay never reads the host queue clock.

## Deterministic fixtures and visual checkpoints

Focused fixtures cover assets/dimensions/scroll/SETBALL/first sync, spring/flippers,
all target and bumper identities, layers/gravities, gate mask changes, dollar/drop
scores and WAIT boundaries, both prize groups, jackpot/super/Billion, both Money
Mania variants, multiplier, extra ball, captures/wheel/ejects, expiry ordering,
bonus/new ball, top-score guards, match/game over, tilt, all bitmap animation
frames, DAC packets, audio samples/cues, and full frontend initials/persistence.

Examples derived from source:

- First sync: X=299*1024+30, Y=530*1024+24, VY=23, gravityY=7.
- Full spring release, zero jitter: -166*32=-5312.
- Both dollars plus B/C banks: 45000 score, 3220 bonus.
- Bonus fixture: 1000*3 + 2*100000 + 500000 = 703000; ball2.
- Cash capture: 500000/510; x5=2500000; eject (100,1700).
- Match: 15 banks, 14 syncs, repeated-digit inversion 9-digit.
- High-score storage: isolated TABLE3.HI, ABC/150000000; no TABLE1/2 writes.

Logical 320×350 checkpoints in `analysis/pf9-checkpoints/` include attract,
initial-chute, charged-spring, source RMTS text, source JACKPOTTS bitmap, tilt,
bonus and high-score-entry. Text/bitmap checkpoints run actual source command
programs. They are native logical snapshots inspected for layout; they are not
invented DOS temporal golden files. Independent asset/bitmap hashes are the
reference fixtures.

Reproduce snapshots with:

```sh
PF9_CHECKPOINT_DIR="$PWD/analysis/pf9-checkpoints" \
  ./tools/go.sh test -p=1 -count=1 ./internal/gameshow ./internal/frontend
```

## Validation results

- Historical full-suite results predate factory-seed inventory validation;
  see `personal-clean-seeds.md` for current results.
- Requested build succeeds. Direct `-pf9 -ticks 1200 -release-at 0 -png ...`
  completes natively, reaching ball2. Its observed score22080 is a smoke result,
  not an independently derived long-run oracle.
- Party Land: `-pf4 -pf4-script -ticks 1200`: score2300000, ball2, unchanged.
  Speed Devils direct run: score11030, ball2, unchanged; source PF7 fixtures pass.
  All PF8 presentation/input and stereo audio tests pass.
- SDL/X11 three-table smoke passes on an owned private Xvfb display. The physical
  window stays the same window at manually resized 887×1102 through selector,
  unavailable F4, F1/F2/F3 attract/play, launcher/flippers/push, pause/resume,
  abort, chute-abort/attract and selector return. Scores use a temporary store;
  no installation score file is written.
- Dummy audio smoke passes (0 queue resets; 24 scheduling-dependent empty-queue
  observations). The additional **real PulseAudio** smoke passes at 48000 Hz
  stereo with **0 queue resets and 0 empty-queue observations**.
- End-of-game and ABC initials are exercised through the real native frontend
  model/session with isolated storage. The bounded X11 smoke exercises controls
  and handoffs, not an entire naturally played final game.

Smoke command (requires local socket access outside the sandbox):

```sh
PF9_AUDIO_DRIVER=pulseaudio python3 tools/pf8_isolated_x.py python3 tools/smoke_pf9.py
```

## Files added/changed

Added `internal/gameshow/{game,rules,flow,matrix,render,audio,content,game_test}.go`,
`internal/assets/gameshow{,_test}.go`, `internal/physics/gameshow{,_test}.go`,
`internal/frontend/gameshow_test.go`, `tools/reference_pf9.py`,
`tools/smoke_pf9.py`, `analysis/pf9-assets.json`, this report and checkpoint PNGs.

Adapted `tools/reference_pf8.py`, `internal/presentation/{content,matrix,attract,
gameover}.go`, `analysis/pf8-content.json` (additional table3 records),
`internal/audio/module.go`, `internal/frontend/{runtime,render}.go`,
`internal/frontend/speeddevils_test.go` (F4 remains the unavailable-table check),
`cmd/pinballfantasies/{main,speeddevils}.go`, `README.md`, and the built binary.
Original ASM, PRG, MOD and installation HI files are not edited.

## Deliberate limits and fourth-table assessment

Single player only; F4 gameplay and other scope exclusions remain excluded.
The inherited FANTASIE interrupt/audio phase convention and waived PF8 temporal
uncertainties remain. This port has source-derived logical/rendering/state
fixtures, not exhaustive DOS temporal certification or a long-run CPU oracle.
No historical CRT calibration or low-mode/options work was added.

The architecture is mature across three real tables for integer physics,
compositing, tasks, audio, controls, sessions and storage. Stones 'N Bones should
require substantial table-local rule/matrix branch work and careful extraction,
with a smaller integration cost than an engine rewrite. Gameshow demonstrated
that even new gate/wheel/Billion systems can be built on existing mask/task/
integer primitives. However its unusual MOD header and DATA2 placement show why
an independent asset/cue audit remains essential for table4.

Expected difficulty is **moderate for content/integration and high for rule/state
coverage**. This is based on the actual categories of new work above, not ASM
line count. Confidence is high in reuse of the established algorithm families;
it remains conditional for undiscovered Stones modes, captures, format extents
and sequencing variants. No speculative Stones abstraction or gameplay was added.
