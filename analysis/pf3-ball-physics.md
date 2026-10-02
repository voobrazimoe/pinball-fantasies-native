# PF3 — Party Land ball physics

PF3 ports the coherent ball-update dependency cone, starting at the verified
PF2 SETBALL origin **(297,530)**. The native application can release the ball,
resolve original map collisions, drive physical flippers, follow ramps and
layer changes, scroll, and emit a drain event. PF1/PF2 remain separate baselines.
There is no x86 execution, DOS runtime, instruction translator, generic physics
engine, floating-point physics, or alternate table implementation.

## Original references

- `BALLCODE.ASM`: `KollaKulan`, `sc_program`, `checkpoint` / `checkpoint2`,
  `XY_LIST`, `sc_krock` / `GET_MATERIAL`, `SEARCHBUMPER`, `FINDFLIP`,
  `sc_newdir`, `sc_move`, `draw_flippers`, and the flipper portion of `TILT0`.
- `FANTASIE.MAC`: `SETBALLPOS`, `SETBALLSPEED`, `SETRASTERPOS`.
- `FANTASIE.ASM`: material records and initial state; `TABLE_ANGLE`,
  `INIT_SIN`, `MASKS2GFX`, `INIT_FLIPPERS`, `ORMASKFLIP`, `MOVEFLIPMASK`,
  `VBLANK_INT`, `LATE_RASTER_INTERRUPT`, `DO_PHYSICS`, `CHECK_RAMPS`,
  `GET_LO_RAMP_BYTE`, `get_hi_ramp_byte`, `CHECK_LEVELS`, `SETSCREENSTART`,
  `SPRINGUP`, `FLIPPRA`, `DELTANIM`, `PUTTHEBALL`, and `CHECK_TARGETS`' event boundary.
- `PLAND.ASM`: `SETBALL`, generated flipper definitions, `RAMPTABLE_hi`,
  bumper/slingshot regions, layer-transition regions, target zones, and the
  plunger-enable assignments in `BYGEL12` / `BYGEL28`.

The repository's source archive references missing external libraries for SIN,
collision maps and generated flipper data. The supplied, SHA-pinned TABLE1.PRG
is the content reference for these. Narrow static inspection of linked references
locates their bytes; the native decoder reads those bytes as data only.

## State and numeric conventions

| State | Representation and meaning |
|---|---|
| X/Y_POS_HI:POS | signed 32-bit position, 1024 units per pixel; modulo-2^32 addition |
| SC_X/SC_Y | signed 16-bit drawing origin; signed division by 1024 truncates toward zero |
| X_HAST/Y_HAST | signed 16-bit displacement per elementary `sc_move` pass |
| GRAVX/GRAVY | signed 16-bit velocity increment, applied **after** displacement |
| ROTATION | signed 16-bit spin, participates in surface friction and slows by 2 per move |
| BALLHIGH | selects lower/upper collision, material and foreground planes |
| BALL_DOWN | sticky original drain flag when SC_Y>=576 |
| SC_KV | normal angle, 0..2047 per full revolution |
| ANTALPIX | count of sampled collision pixels, eight bits |
| EXAHITX/Y | normal-derived contact point for target/event dispatch |
| Flippers | source 60-byte descriptors; integer angle, speed, frame, bounds and rotation center |
| RASTERPOS | signed 16-bit screen position in 1/16 scanline units, includes SPLH=33 |

SETBALL sets velocity **(10,0)** and releases HOLDSTILL. Initial gravity is
**(0,8)** from `10*TT/NN`, TT=5, NN=6. The earlier hidden NEW_BALL position
has no empty byte among CHECK_RAMPS' three probes, so it retains this initializer.
On the first visible update pair, CHECK_RAMPS selects the normal high-resolution
slope entry. PINBALL.CFG enables TABLE_ANGLE, subtracting three from the high
ramp gravity entries: `(0,7), (2,11), (-2,11), (-4,13)`.

There are no delta-time multipliers. A PF3 **sync** is one high-resolution
VBLANK/late-raster pair. VBLANK executes two `sc_program` passes, then physical
postprocessing; late raster executes one pass because SHIFTKEYS bit 2 and
INSIDE_RASTINT are both set. This retains the intermediate gravity/layer boundary
rather than replacing the pair with three identical calls. The source names
71 syncs/second in `hires_changes`, with a question mark; the external interrupt
service is absent. Native 14ms presentation pacing is approximate host playback,
not a verified emulation of DOS interrupt timing. State evolution is tick-driven.

## Update order

For each elementary pass:

1. Sample the collision ring and resolve any contact (`sc_krock`, `sc_newdir`).
2. Update physical flipper angles/velocities (`TILT0` flipper branch).
3. Add Y velocity to Y position, derive SC_Y and set the drain flag.
4. Add X velocity to X position and derive SC_X.
5. Add gravity to Y velocity, then X velocity; reduce spin toward zero.
6. Copy overlapping flipper collision frames (`draw_flippers` / `MOVEFLIPMASK`).

After VBLANK's two passes: consume pending active-object hit events, update
ramp gravity, apply one layer transition, and dispatch the drain boundary.
For a live ball, consume target contact events and update the source plunger-enable
regions, update scrolling, then perform the late-raster pass.

The original signed IDIV quotient range is checked: invalid division returns an
error and stops playback instead of silently narrowing an impossible quotient.
Velocity arithmetic wraps at 16 bits. Collision response applies the original
±4100 clamps at its original locations; free gravity increments are not clamped.
The native implementation preserves the product/high-word arithmetic, angle
averaging wrap, friction divisor overflow, impact-angle threshold, and penetration
correction. It does not substitute vector reflection or normalized float normals.

The original flipper `UpFlip` comparison is deliberately unusual: `JLE` leaves
an already more-negative speed alone, and a speed above the negative bound jumps
to that bound. From rest, the left flipper's first two elementary updates are
`speed=-68, angle=68, frame=1`, then `speed=-75, angle=143, frame=2`.
A direct source-derived test pins this branch.

## Collision and original content

A 44-point ring surrounds the nominal 16-pixel ball, rooted at `(SC_X−1,SC_Y−1)`.
Source checkpoint order is retained, including the last-hit material sample.
The average normal uses the source's 16-bit accumulated angles and quadrant-wrap
correction. The normal-derived contact lookup rounds using `1408*SC_KV+0x8000`;
its index-44 edge case preserves the adjacent SC_KV/ANTALPIX values.

Collision/material bitmaps use original 40-byte rows with MSB-first bits. Lower
collision map MASK12 changes under flippers; upper collision map MASK22 is
separate. Three bit planes form the material index. The source's unusual VGA
storage for material maps is decoded directly from its pre-interleaving data.
Ramp identifiers occupy the low nibble of the third material map byte.
Linear byte addressing beyond x=319 can read the next row and is preserved.
Native reads beyond the supplied map return empty rather than reading unrelated
DOS segment memory; these undefined-memory cases are not claimed as emulated.

| Original data | File offset | Bytes |
|---|---:|---:|
| SIN including cosine overlap | 0x1e340 | 5120 |
| 8 material records | 0x1c05d | 128 |
| 3 generated flipper descriptors | 0x20690 | 180 |
| MASK12 / MASK11 / MASK22 | 0x3b930 / 0x41330 / 0x46d30 | 23040 each |
| MASK13 | 0x71b30 | 23040 |
| MASK21 / MASK23 | 0x77530 / 0x7cf30 | 20400 each |
| Left / right / upper collision frames | 0x4c730 / 0x4fe10 / 0x4ed50 | 8904 / 8904 / 4284 |
| Lower / upper foreground | 0x300b0 / 0x35f30 | 23040 each |
| DATAFLIP delta records | 0xc2e0 | 55888 |
| Four-plane flipper graphics | 0x82930 | 1748 |

`analysis/pf3-data.json` records hashes of these original regions. SIN is read
big-endian as INIT_SIN swaps bytes; its signed amplitude is 16384. Generated
flipper descriptors are read from the shipped data rather than regenerated using
the source's disabled CREATE_FLIP_DATA branch. INIT_FLIPPERS divides their
three-plane frame size by three; ORMASKFLIP adds the underlying table collision
bytes to every frame before MOVEFLIPMASK selects a frame.

Bumper/slingshot bounds and upper/lower transition rectangles come directly
from Party Land's source declarations. Material 7 selects bumper search and
−7000 normal impulse; material 3 selects slingshots and the original −2000
impulse/−300 threshold. The original early exit for a candidate above the ball
is retained. Flipper contact uses its actual speed and rotation-center/power-zone
arithmetic. Screenshake is zero because PF3 has no nudge input; the non-flipper
TILT0 prefix therefore has no effect on this supported input state.

Original flipper graphics are rendered using FLIPPRA's direction-dependent delta
records and DElTANIM's four-plane block copies. The renderer owns its background
pixels and redraws the masked ball from that background, avoiding trails.
Upper-plane balls use HID2 foreground. The viewport follows original integer
scrolling with the supplied setting's SLIME_FACTOR=9. Rendering displays the
state after a complete sync; VGA beam scheduling is not reproduced.

## Deferred callbacks and boundaries

Physics exposes `EventBumperHit`, `EventSlingshotHit`, `EventFlipperHit`,
`EventTargetHit`, and `EventDrain`. Bumper indices 0..2 are Bumper1..3;
slingshot indices 0..1 are left/right; flippers 0..2 are lower-left,
lower-right, upper-left. Target indices 0..3 are TOUCHER, DROPA1, DROPA2, DROPA3.

There is no score, sound, light or mission consumer. CHECK_BUMPERS' bookkeeping
and sound callbacks are replaced by pending-hit events. CHECK_TARGETS emits a
contact-zone event instead of invoking scoring/table callbacks. Delayed target
knockdown/restoration, rule-controlled gates, trap/ejection tasks, bonuses and
player rotation remain deferred callbacks; they are not part of the sc_program /
DO_PHYSICS collision solver. The initial raw MASK12 already matches all three
RESTORED target templates byte-for-byte. Dynamic table-rule changes to geometry
are not claimed as implemented. Static original targets/gates remain in the map.
On a drain event native simulation freezes that ball; LOOSE_BALL's score/bonus,
new-ball and player-selection task sequences do not run.

The plunger is a small native release control: release Space to apply the source
SPRINGUP full-charge arithmetic, charge=32 and explicitly fixed jitter=0. The
API accepts charge/jitter as replay inputs rather than sampling host time.
Left/right Shift supply the physical flipper bits. The source plunger-valid area
writes prevent a subsequent release away from the shooter lane. No DOS keyboard
interrupt handler, menus, attract mode, music or other table is ported.

## Independent validation and execution

`tools/reference_pf3.py` is a separate Python numeric model derived from the
assembly and original bytes. It performs the reference ring scan using rotated
16-bit map words/checkpoint records; Go uses direct ring coordinates. Both retain
original numeric narrowing. Python independently composes RGBA pixels from the
PF1 extractor, original foreground and flipper delta content. It never invokes
Go, generates expectations from Go, or executes x86. Its output is stored in
`analysis/pf3-trajectory-fixtures.json`.

Eight cases cover SETBALL and a deterministic release, a wall, bumper, moving
flipper, drain, layer transition, nonzero slope gravity and slingshot. Every
stored sync compares complete ball state, raster position, all flipper
speed/angle/frame values, semantic events and framebuffer SHA-256. Additional
hand-derived tests cover steel/active-bumper response, the unusual flipper
branch, plunger arithmetic, signed division and quotient faults. Repeated runs
compare state bytes, collision masks, raster and framebuffers for 600 syncs.
Existing PF0 inventory/PF1/PF2 tests pass, and their decoder/source fixtures
remain byte-identical.

Reproduce from the project root:

```sh
python3 tools/reference_pf3.py > /tmp/pf3-reference.json
cmp /tmp/pf3-reference.json analysis/pf3-trajectory-fixtures.json
./tools/go.sh test -p=1 ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -pf3
./bin/pinballfantasies -pf3 -ticks 120 -release-at 100 -png /tmp/pf3-native.png
./bin/pinballfantasies -pf3 -ticks 120 -release-at 100 -duration 4s
```

The deterministic 120-sync path releases before sync 100 and ends at
**(301,237)** with velocity **(0,−4892)**. Its scrolled native frame was visually
inspected, and the SDL/X11 live window opened for the release/advance/freeze test.
Use `-ticks N` to freeze live playback after N syncs; zero keeps it live until
close/drain. `-release-at` is a deterministic demo/replay input. `-pf1` and the
unchanged default PF2 view remain available. PF3 adds only this subsystem;
scoring, audio, table-rule consumers and other tables are not begun.
