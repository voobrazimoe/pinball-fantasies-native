# PF4 — Party Land rules and scoring

The native `-pf4` path consumes PF3's physical events, implements Party Land
progression, shows original-content lamps and fonts, and supports a single-player
ball-loss/bonus/new-ball/game-over loop. PF1, PF2 and PF3 remain separate regression
paths. This is a semantic Go port, with no x86 execution, translator, scripting
engine, audio player or other table implementation.

**PF4.5 update:** matrix command timing, task ordering and silent jingle completion
now supersede the historical PF4 timing adaptations described below. See
[PF4.5 evidence and adaptation classifications](pf45-timing-scheduler.md).
Silent driver timing still has explicitly INFERRED/UNKNOWN aspects; this is not
claimed to be cycle-exact DOS execution.

## Source and content references

VERIFIED — `reference/original-dos-source/PLAND.ASM`:

- `BumperLista_L`, `Bumper2Lista_L`, `ZonLista_L`, `AreaLista_L/U`,
  `AREALISTA_L_T/U_T` (tilted alternatives inspected, not enabled).
- Decimal constants and `EFFECT_STRUC` / effect records, lines 232–607.
- `WHEN_NEW_GAME_RESET_TABLE`, `RESET_VARS2`, `WHEN_NEW_BALL_RESET_TABLE`,
  `RESET_VARS`, `PLAYER_STRUC`, `P_STRUC_2_VARS` / `VARS_2_P_STRUC`.
- `TOUCHER`, `DROPA1/2/3`, `DROPSETA`, `CHECKALLDUCKS`, `TURNOFFDUCKS`.
- `CREATE_MASK_DATA`, `KILLGATE1`, `RESTOREGATE1`, `KILLD1/2/3`,
  `RESTORED1/2/3`, `START_DROP`, `DROPTASK1/2`.
- `BYGEL1/2/3/4/5/6/7/8/9/10/11/12/13/14/28`, `CLOSE1/2`,
  `OPEN2`, `BYGEL4B`, `PUKEGROUP`, `PUKETIME`, `PUKEP1/2/4/5`.
- `GROPA/B/C/D/E`, snack release/cooldown tasks, hidden-entry/5X tasks,
  tunnel decay and release, `SPINIT`, `ARCADESLUMP`, `SPIN*_RUT`,
  dragon `CHECK_JACKPOT`, `CHECK_BALL`, `CHECK_5M`, `BEFOREFLASH`, `DURINGFLASH`.
- `CRAZY_LETTER_SPOTTED`, `CHECK_PARTY`, `CHECK_CRAZY`, `OK_2_BE_HAPPY`,
  `OK_2_LAUGH`, `ADDHAPPY`, `ADDMEGALAUGH`, `READ_SPECIAL_MODE_COUNTER`,
  `UPDATE_COUNTERS`, `CHECK_SHIFTKEYS`, `JACKADD`, `INIT_JACK`.
- `LOOSE_BALL`, `ball_lostTS`, `_BONUS_X_CALCS`, `_CALC_CYCLO`, `_CALC_HAPPY`,
  `_CALC_MEGA`, `_FLORPA`, `DO_FLORPA`, `_CHANGE_PLAYER`, `_KOLLA_XXBALL`,
  `_knacket`, `knackrut1/2`, `_CHECK_XXBALLS`, `NEW_BALL`, `SETBALL`.
- `LON1..56`, `LONINDEX`, mode/mystery text and animation timing tables.

VERIFIED — shared `FANTASIE.ASM`: `VBLANK_INT`, `DO_PHYSICS`, `DO_ELECTRONICS`,
`CHECK_BUMPERS`, `CHECK_AREAS`, `CHECK_TARGETS`, `DOEFFECT`, `ADDSCOREBCD`,
`DOADVANCE`, `DOWAITSYNCS`, `DO_TASKS`, `DOLIGHTON/OFF/FLASH`, `DOENDFLASH`,
`FLASHTHELIGHTS`, `RECALC_LIGHTS`, `SLACK_THIS_LIGHT`, `WHEN_NEW_BALL_RESET`,
`WHEN_NEW_GAME_RESET`, `FANTASIES`, `hires_changes`, `_COUNTDOWN`, `_WAITJINGLE2`.
`FANTASIE.MAC`: `ADDSCORE`, `ADDBONUS`, `EFFECT`, lamp/flash macros,
`SETBALLPOS`. Existing `BALLCODE.ASM` physics remains PF3's port.

VERIFIED — the original supplied `TABLE1.PRG` is SHA-validated by the existing
asset decoder. External data missing from the source archive is read from this
file. Static inspection found DATA file base **0x19d40** (linked paragraph
0x19b4 + MZ header 0x200):

| Content | File location / evidence |
|---|---|
| LON palette records | 0x1adb9; source LON1 bytes `128,2,95,0,0,64,0,0` |
| FONT13 | 0x1ff40 = DATA+0x6200, 13 bytes per glyph |
| FONT5 | 0x20450 = DATA+0x6710, 5 bytes per glyph |
| D1 up/down | 0x205f0 / 0x20610, 30 bytes each |
| D2 up/down | 0x20630 / 0x20650, 30 bytes each |
| D3 up/down | 0x20680 / 0x20670, 15 bytes each |
| Skyride gate closed | private copy of MASK22 at byte 14, row 15, 2×18 bytes |

The linked font assignments are at file 0x48c1 / 0x48e2. The bounded linked font
renderer at 0x6fcc selects `font + glyph*height` and uses the bits as original
matrix dots. Linked duck setters at 0x13ae..0x13fd reference DATA offsets
0x68d0,0x68b0,0x6910,0x68f0,0x6930,0x6940. `MOVE_MASK_DATA` at 0x6255 copies
packed source rows to 40-byte map rows. These instructions are inspected offline;
the native program reads data only. The linked scrolling routine at
0x7197..0x72a9 advances one character every four high-resolution syncs and
checks the terminator twenty characters ahead; hidden-entry wait length is derived
from the supplied HIDDENTEXT bytes using that behavior.

`analysis/pf4-rule-fixtures.json` records content hashes and independent numeric
fixtures. `tools/reference_pf4.py` parses the original assembly's decimal
constants/effects and reads original data. It never imports the Go implementation.

## State and arithmetic

VERIFIED — all score-like fields are **12 unpacked decimal bytes**, most
significant first: `SIFFRORNA`, `BONUSSIFFRORNA`, `JACKVALUE`, skill awards,
mode totals. AAA/ADC adds from the least significant digit and discards the final
carry: modulo **1,000,000,000,000**. `Decimal.Add` takes its source by value,
preserving aliased bonus doubling. Display never mutates these numbers.

VERIFIED — key state mapping:

| Go state | Original state |
|---|---|
| Score, Bonus, Jackpot | SIFFRORNA, BONUSSIFFRORNA, JACKVALUE |
| SkillTunnel/Cyclone | SKILL_SCORE / SKILL_SCORE2 |
| HappyTotal/MegaTotal | HAPPY_HOUR_TOTAL / MEGA_LAUGH_TOTAL |
| Multiplier | BONUSMULTIPEL and BONUS_X (reachable values 1,2,4,6,8) |
| BallNumber, ExtraBalls | BALLS[11], XBALLS |
| Lights / Lamps | LIGHTSTATUS logical flags / LON and LOFF palette state |
| SkillTime, LoopTime, ReverseTime | SKILLSHOTDOWNCOUNTER, ANY_LOOP_DOWN_COUNTER, REVERSE_DOWN_COUNTER |
| TunnelTime | TUNNELDOWNCOUNTER (720 / 1440 syncs) |
| MB/HB/DB | MBFLASHING, HBFLASHING, DBFLASHING |
| FiveX/FiveMillion/BallFeature | _5XFLASHING, _5MFLASHING, BALLFLASHING |
| JackpotNormal/Timed | JPFLASHING_NORMAL / JPFLASHING_TIME |
| SnackNext/snacks/Pop | SNACKCOUNTER, SNACKFLASHERS, POPCOUNTER |
| Skyride/Puke/Cyclones | SKYRIDECOUNTER, PUKECOUNTER, CYCLONECOUNTER |
| lastCheck/lastArea | LASTCHECK / LASTAREA |
| Random | SLUMPCOUNTER, byte wrapping; arcade adds 21 before selection |
| clock | SLUMP_COUNTERN, word wrapping with +1030 per native sync |

Counters retain byte/word widths. Game phases replace the DOS frontend's
LOOSING/BALL_DOWN/new-ball/attract control path with observable native states.
The code keeps a 50-slot Party Land callback task list and 15 flash slots, as in
the source. No reusable rule engine is introduced.

## Ordering and awards

VERIFIED — the PF4 consumer is opt-in. PF3 without callbacks remains unchanged:

1. Two original elementary physics/flipper passes.
2. Deferred successful bumper/slingshot callback; original ramp and level checks.
3. Drain handling.
4. UPDATE_COUNTERS, first matching area callback, contact-derived target callback.
5. Shift edge rotates PUKE lamps, tasks in ascending original slot order,
   lamp flashing; the native semantic mode timer advances.
6. Scroll and the late elementary physics pass.

Regions use inclusive unsigned comparisons; first match wins. A repeated callback
identity is debounced by LASTCHECK; leaving all regions clears only LASTCHECK.
LASTAREA remains available to identify loop direction and ramp completion.
Overlapping spring rectangles at y540 resolve to BYGEL12 first. Three CLOSE2
rectangles share the same identity. Low target hits never dispatch on the high
plane. The one deferred physical bumper slot retains PF3's overwrite behavior.

VERIFIED — representative source awards:

| Element | Score | Bonus |
|---|---:|---:|
| Bumper | 1,000 | 0 |
| Slingshot | 500 | 0 |
| Duck | 7,510 | ADDBONUS 750 |
| New PUKE letter | 20,070 | ADDBONUS 1,000 |
| Inlane | 10,040 | raw 1,000 |
| Unlit outlane | 50,030 | 0 |
| Skyride 1/2/3 | 100k / 250k / 500k | 10k / 25k / 50k |
| Reverse 1/2/3 | 250k / 500k / 750k | 10k each |
| Loop 1/2/3 | 100k / 250k / 500k | 10k / 25k / 50k |
| Tunnel 1/2/3 | 1m / 3m / 5m | 25k / 250k / 500k |
| Cyclone normal / 5X | 100k / 2.5m | 25k / 500k |
| PARTY letter | 250k | 25k |
| Snack base | 50k | ADDBONUS 5k |
| ICE / SODA / POP | 250k / 500k / 1m | 25k / 50k / 100k |
| No-snack HSCORE | 50k | 40 |
| Dragon default / 5m | 250k / 5m | 10,030 / 10k |
| Extra-ball collection | 10k | 5k |

DOEFFECT writes SCORECHANGED even for zero-valued effects. Bumpers/slingshots
also add to Happy Hour when active. Target TOUCHER has **no score award**; it
qualifies arcade, with a 20-sync inhibitor.

VERIFIED — quirks intentionally retained: first ordinary cyclone advances the
counter twice but awards CYCLCOUNT once; subsequent 5X increments it by five.
Skill cyclone kills 5X through PARTY_R before its ordinary cyclone branch and
retains SKILLSHOTDOWNCOUNTER. Tunnel skill clears that timer. Both skill values
increase by a million, then award the entire accumulated skill value.

## Features and lamps

VERIFIED — three ducks lower original collision masks after 20 waits; all three
qualify the next ICE/SODA/POP and restore after 71. Snack takes the highest
qualified reward, clears all snack qualifications, and POP qualifies PARTY_A
plus hold bonus on first collection and timed double bonus thereafter.

VERIFIED — three completed skyrides qualify MB unless X8 is already lit.
Reverse collection performs **reverse award → MB → HB → DB**, so DB doubles the
bonus after the reverse and MB contributions. MB advances 1→2→4→6→8.
Three reverse awards cycle balloon indications; three forward loops advance MAD,
then CRAZY in original **Y,Z,A,R,C** order. Five CRAZY advancements start Mega
Laugh. PARTY P/A/R/T/Y comes from reverse→tunnel, POP, reverse→cyclone/skill,
consecutive loops/ramps, and complete PUKE respectively. Completing one mode's
letters during another mode sets its pending flag; the continuation waits 400.

VERIFIED — complete PUKE banks qualify dragon 5m, extra ball, then jackpot.
Dragon collection priority is **jackpot → extra ball → 5m → ordinary award**.
Jackpot starts at 10m, grows in 50k steps on the source's designated callbacks,
and resets to 10m on collection. Extra balls also qualify through arcade then
outlane 39; original right outlane checks logical lamp 39 too.

VERIFIED — lamps use original indexed artwork and all 56 original palette
records. Logical status is separate from flashing: flash never changes
LIGHTSTATUS. LONINDEX's entry 40 aliases entry 39. RGB is converted with the
original `(component*162)>>8`, halved for LOFF, then expanded from DAC values
for native RGBA. Original task ENDFLASH calls that leave the current palette
state intact continue to do so. Full flash lists silently ignore new flashes,
as in the source. No modern lamp graphic is substituted.

## Bonus and ball flow

VERIFIED — **two distinct multipliers** are intentionally represented by the
same reachable multiplier value:

- ADDBONUS adds its constant BONUSMULTIPEL times immediately.
- DOEFFECT adds its TBONUS once, bypassing that macro.
- At ball loss, base BONUSSIFFRORNA is multiplied by BONUS_X.
- Cyclones ×100k, Happy Hour total and Mega Laugh total are added afterward,
  outside that final multiplication.

Happy hits accumulate 1m per eligible hit, Mega qualifying shots accumulate 5m
per shot. These go into bonus at ball loss. Mode expiry itself contributes a raw
1m or 5m bonus via EOHAPPYHOUR/EOMEGALAUGH. Hold bonus retains the final multiplied
and supplemented total (`TEMPSIFFRORNA`), not the pre-multiplied base.

VERIFIED — physical drain emits BallLost, holds the ball in the drop zone and
freezes normal electronics counters. If SCORECHANGED is false, Party On returns
the same numbered ball after 30 waits. Otherwise: LOSTBALL → base bonus display
and multiply → cyclone → happy → mega → total bonus → decimal digit countdown
into score → extra-ball check → numbered-ball progression. Bonus transfer adds
one decimal unit at a time, every four calls, including the original final
`-10` delay. New ball is scheduled after 30; its hidden origin is (282,530),
followed by SETBALL after 80 at (297,530), velocity (10,0). Flipper angles, spin
and ramp gravity are retained across new ball, with Lost/Stopped cleared.

VERIFIED — PLAYER_STRUC's PUKE/MAD/CRAZY/PARTY logical lamps, score, skill values
and cumulative cyclone count survive new balls. Temporary flags, snack/train/
rocket/balloon counters, bonus multiplier, happy/mega totals, duck masks and
flash/task lists reset. Held bonus survives; an ordinary counted bonus becomes
zero. XBALLS decrements on shoot-again, retaining the numbered ball.

VERIFIED — FANTASIES selects three balls for S_BALLS=0 and five otherwise.
The native frontend reads that one byte from the supplied six-byte PINBALL.CFG;
this installation selects **five**. Core deterministic tests explicitly exercise
the source's three-ball option. There is no configuration UI.

VERIFIED — the final match uses the **tens digit** Score[10], 22 high-resolution
banks separated by 11 syncs, and the source's non-repeat modulo-ten choice. A
match supplies the XXBALLE single-player path; earned extras are consumed first.
When the match ball has no extras, or the initial final match fails, GameOver
freezes the displayed result instead of entering attract mode.

## Display and deliberately deferred behavior

VERIFIED — a persistent native **320×66** HUD above the PF3 **320×317** viewport
shows score, ball, bonus, multiplier, PARTY/CRAZY status, mode/arcade/shoot-again
state, cyclone count and jackpot. FONT13 and FONT5 come from original data;
no host font is used. It is a native layout, not a reproduction of the DOS
33-line matrix effect presentation. Rendering can run at different frequencies
without changing rules or eventual pixels.

VERIFIED — sounds and jingles become current-sync `Sound` / `Music` events,
carrying source sound names or the effect identity. There is no playback,
audio device initialization, mixer, music module parser or audio driver port.
No voice-specific rule callback was needed. The original NEW_BALL sound callbacks
remain semantic events. The plunger intent survives the following sync's event
buffer reset.

VERIFIED — multiplayer PLAYER rotation and saved-player branches are reduced to
one player's persistent fields. Title, table selection, high-score entry,
attract mode, options, joystick, other tables, save states and networking remain
absent. There is no tilt/nudge input path beyond PF3's existing supported input.
After GameOver, restarting the executable starts a new game.

## Known semantic differences and uncertainty

INFERRED — native syncs replace interrupt/main-loop time as the deterministic
random clock. The source adds 1030 per VBLANK **and** increments SLUMP_COUNTERN in
MAIN at host-dependent frequency. The latter increments and uninitialized DOS
start state are deliberately not emulated. Arcade, drop jitter and final match
are reproducible native sequences, not identical random outcomes to DOS.

INFERRED — arcade uses the original **NOSOUND** WAITJINGLE2 completion path:
32 mystery frames ×7 syncs plus DECCOR=25. Score-reward hole release uses the
source HANGSAVER2 upper bounds (180/140/150/70) when the missing music-completion
callback would otherwise gate HANGSAVER's lower bound. No invented prize or
award is used, but release timing can differ from audible DOS play.

UNKNOWN — exact full dot-matrix interruption/priority arbitration and its
one-sync dispatch overhead. The original animations/scrollers are not displayed;
HIDDEN wait is derived from the linked text-scroll timing without reproducing
its preemption. Modes become active and begin their 25-second high-resolution
countdown immediately, while DOS starts the countdown after intro animation /
scroll. Native uses 25×71+1 countdown calls, preserving the initial decrement
from 26 to 25. This changes the duration of the initial active scoring window.
Ball-loss arithmetic/order is preserved, but clear/print/jingle dispatch delays
are omitted. These are fidelity gaps, not VERIFIED equivalence.

INFERRED — native saucer capture explicitly holds balls at their prescribed
origins during waits. Original hidden/tunnel/drop tasks can retain HOLDSTILL
from their entry path and rely on collision/capture geometry. Source coordinates,
ejection velocities and timed drop stages are retained, but the explicit hold
is a stability adaptation. Original forced camera wipe/scroll tasks are reduced
to the new-ball bottom viewport and existing PF3 scrolling.

VERIFIED — keyboard launch uses PF3's full-charge release and explicit jitter
input, rather than a reproduced animated plunger-charge UI. Original content
palette conversion is applied to lamp entries; other artwork retains PF1's CMAP
representation. The matrix uses native amber instead of original matrix DAC
pulse/fade sequences. PF0–PF3 pixels are unaffected.

## Validation and commands

`tools/reference_pf4.py` independently extracts 54 effect records, 20 decimal
constants, 56 lamp records and external content hashes. Tests cover real PF3
bumper/slingshot/target trajectories, native score carry/aliasing, lane awards,
duck masks and restoration, PUKE rotation/train features, skyrides, reverse
MB/HB/DB ordering, region overlap/debounce/LASTAREA, area-before-target ordering,
skill/tunnel/cyclone quirks, modes/dragon priority, snacks, extra balls, held
bonus, drain/new-ball/game-over, original ball settings and task reset inside a
running callback. The scripted two-launch/Shift sequence yields **2,300,000**
and reaches ball 2. Two independent native runs compare every step's events,
score/bonus/lamps/ball state and periodic complete framebuffers, with different
intermediate rendering frequency. This is a determinism check, not an independent
DOS trajectory oracle for all rules.

All PF0 inventory, PF1 static frame, PF2 start frame and PF3 independently-derived
physics/framebuffer regressions pass unchanged. Go vet passes. Native SDL opens
both offscreen and on X11; the X11 smoke test required desktop-access escalation
because sandbox-only X11 was unavailable. HUD PNGs were visually inspected.

```sh
python3 tools/reference_pf4.py
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -data-dir . -pf4
./tools/go.sh test -p=1 -count=1 ./...
./tools/go.sh vet ./...
./tools/go.sh test -p=1 -race ./internal/partyland ./internal/physics
./bin/pinballfantasies -data-dir . -pf4 -pf4-script -ticks 1200 -png /tmp/pf4-script.png
```

PF4 work stops here. No PF5, sound/music subsystem or additional table has begun.

## Files changed

Created:

- `internal/partyland/game.go`, `regions.go`, `features.go`, `holes.go`,
  `flow.go`, `render.go`, `game_test.go`.
- `internal/physics/rules_boundary.go`.
- `cmd/pinballfantasies/rules.go`.
- `tools/reference_pf4.py`.
- `analysis/pf4-rule-fixtures.json`, `analysis/pf4-partyland-rules.md`.

Modified:

- `internal/physics/ball.go`, `collision.go`: optional original-order callbacks
  and private mutable upper mask; PF3 callback-free behavior remains tested.
- `cmd/pinballfantasies/main.go`: PF4 and bounded deterministic script options.
- `internal/platform/physics.go`: shared Party Land window title/log wording.
- `README.md`: current play/run/test instructions and fidelity qualification.

Final validation: fresh `go test -p=1 -count=1 ./...`, `go vet ./...`, and
`go test -p=1 -race ./internal/partyland ./internal/physics` all passed.
