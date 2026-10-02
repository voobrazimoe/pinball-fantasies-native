# PF8 — native DOS presentation parity

Status2026-10-01: **PF8 implementation accepted by the user within the revised
scope**. Full DOS temporal parity is not claimed. The checkpoint table below
distinguishes measured comparisons from source coverage and unresolved observations.
The user accepts the selector sequence and game-ending flow
as implemented and explicitly waives further validation of those sequences.
The user also listened to the native output and confirms music no longer stutters.
Implementation is accepted within that revised scope; this does not convert
the unobserved DOS temporal details into certified parity. See the continuation.
No x86 execution, DOS service, interpreter, JIT, or hardware emulator is part of
any native package. Original DOS execution occurred only in an external oracle.

## Video modes and final display composition

Authority: `FANTASIE.ASM` SET_360X350, SET240, SETSPLIT, VBLANK_INT,
`INTRO.ASM` video setup and raster callbacks; linked table immediates confirm the
retail build's register choices. `internal/presentation/video.go` records the
final high-mode CRTC registers, including the horizontal override.

| Surface | Dimensions | Meaning |
|---|---:|---|
| Full table artwork | 320 × 576 | Indexed playfield and collision geometry |
| Planar memory pitch | 336 pixels / 84 bytes per plane | Original SW/BPL, including invisible padding |
| Visible high-mode playfield | 320 × 317 | Current scroll window, foreground, flippers, ball, spring |
| Matrix output | 320 × 33 | Bottom raster split, y=317..349 |
| Matrix dots | 160 × 16 | Dots at x=2*n, y=319+2*n; gaps remain black |
| High-mode final logical frame | 320 × 350 | Explicit composition, not the full playfield buffer |
| Startup | 320 × 240 | Original indexed PBM composition |
| Selector | 640 × 240 | Planar 16-color frame, double-scanned to 480 lines |
| Startup high logo | 640 × 480 | Original BIOS 12h presentation |
| SDL physical window | User-controlled | Persistent PF6.1 window; never resized by mode switches |

High CRTC: 00=6b,01=4f,02=5a,03=8e,04=5b,05=87,06=bf,07=1f,
09=00,10=83,11=a5,12=5d,13=2a,14=0f,15=63,16=ba,17=e3,18=3c.
Miscellaneous output is a7. Horizontal display `(4f+1)*4=320` indexed
pixels, vertical display `5d + overflow + 1=350`, pitch `2a*8=336`.
SETSPLIT sets line compare 316; address resets to the matrix page on line 317.
The initial playfield scroll origin is 259, independent of the matrix page.
SET240 changes the vertical register group to 06=0d,07=3e,09=c0,10=ea,
11=0c,12=df,14=00,15=ea,16=06,17=e3. The normal native gameplay path uses
the recovered high mode; a complete user-selectable low-mode option is deferred.

The renderer uses the original indexed field, current scroll, foreground and
flipper delta records, original ball bitmap, and the spring buffer. It then
composes the bottom matrix separately. The previous 66-pixel modern score panel
is removed from the normal Party Land and Speed Devils presentation.
Page flipping and VGA write masks are represented by the resulting native data
layers; the native renderer does not recreate VGA hardware execution.

## Aspect and pixel geometry

The logical frame aspect is 320/350=32/35. The original occupied ball bitmap
has equal 15-pixel X/Y diameters. Native high-mode scaling retains this artwork
geometry (pixel aspect 1:1). Selector pixels receive the source-defined 1:2
vertical double-scan mapping, giving SDL logical size 640×480. Startup uses
320×240; the BIOS logo uses 640×480. SDL fits each logical surface uniformly
inside the persistent window, with black letterbox/pillarbox regions.

**A calibrated historical CRT display aspect has not been established.** A
square source-art circle establishes art geometry, not a monitor's active-area
calibration. The CRTC values establish scan counts and pixel-clock samples;
they do not prove that this custom mode's intended active image filled a 4:3
screen. No unconditional 4:3 transform has been added. The raw DOS oracle's
640×350 table surface doubles X samples per indexed pixel; comparing that
host surface directly against native 320×350 makes circles appear vertically
compressed. Table comparisons below remove this exact X duplication. That
normalization is not a claim about intended analog CRT pixel aspect.

## Palette recovery and fixes

1. Table CMAP components previously bypassed the original UNPKLBM `>>2` DAC
   conversion. Shared `VGAPalette` now quantizes to six bits and `DACRGB`
   expands with high-bit replication `(v<<2)|(v>>4)`. Raw IFF assets and PF1/PF2
   indexed fixtures remain unchanged.
2. Frontend DAC expansion previously used linear `255/63` rounding. PF6 RGB
   fixtures were regenerated from the actual six-bit conversion, with original
   index/asset hashes unchanged; they were not updated from native screenshots.
3. INTRO's linked chunky UNPKLBM branch constructs the half-bright bank 32..63
   from the first 32 colors before programming the DAC. Griffin upper/lower
   planes share pelle1. Viking and FLD halves likewise use the palette selected
   by the original routines. No griffin-specific recoloring lookup was added.
4. Fade macros retain their final integer step: griffin's held palette is 19/20
   in DAC units. Viking/FLD white-to-palette holds retain their source final
   blend rather than jumping to the target palette.
5. Lamp packets can share DAC entries. Party Land lamps 53 and 56 both write
   index 111. Native rendering previously reapplied all lamps in numeric order,
   overwriting the most recent original LON/LOFF write. The palette now changes
   when each lamp packet is emitted. Source-derived test values are lamp56 off
   RGB 73,24,24 and subsequent lamp53 on RGB 97,32,32.
6. SHOWPICS/julius loads upper/lower preview palettes into separate VGA banks;
   CREATRETF/DUMRETF change the active attribute-controller bank. The sidebar
   shares those banks. It previously retained its own IFF palette, making the
   Digital Illusions logo wrong even after preview colors were fixed. The final
   selector now uses both original raster palettes for all indexed content.

`tools/reference_pf8_visual.py` independently decodes original PBM indices,
CMAP, held fade, spring bytes and every matrix animation delta. Griffin selected
palette values and held-frame hashes are pinned in `pf8-visual-fixtures.json`.
The full stable selector capture is pinned independently in `pf8-bios-font.json`.

## Matrix visual implementation

`internal/presentation` consumes existing PF4.5/PF7 command state, frame number,
scroll increment, and integer clocks. It supplies no wall clock and cannot
award scores, advance rule tasks, or change matrix command priority. Begin/Visit
connect to each existing table's matrix dispatch and tick; Flash consumes one
existing sync. Drawing a frame does not advance matrix time.

Original FONT13/FONT11/FONT8/FONT5 are read directly from table data. Party Land
starts at file 1ff40/20150/20310/20450; Speed Devils starts at
1f170/1f380/1f540/1f680. Original numeric text encoding is `'7'+digit`.
Literal `*` selects the original blank glyph; encoded glyph 96 is distinct.
Word positions use the original 336-pixel stride and 168-byte matrix dot-row
pitch. Clipping includes original negative left positions.

Generated SCORE and SCROLL artwork is recovered offline from literal graphics
store records and embedded as bitmap data, never callable CPU code. They are
separate fonts. SCROLLE advances two dot pixels per source sync; character
advance remains four syncs in the existing scheduler. Decimal leading-zero,
right alignment, comma positions and PRINT_NUMBER_CENT's inclusive first-nonzero
count retain linked-source arithmetic, including its non-modern centering.

DATA2 bitmap animations decode two delta planes with starting destination 167,
`dest += byte>>1`, byte254 as a skip, and low bit as the new dot state. All
animation frames are checked against independent source-data hashes. The native
bitmap checkpoint uses real `_GEAR`; an earlier checkpoint accidentally named
nonexistent `_JUMP` and rendered blank. That checkpoint invocation was corrected,
without modifying the animation assets or weakening fixtures.

Lit dots use original percentage 95/70/27, recalculated with `*162>>8` to
DAC60/44/17 (RGB243/178/69). Unlit dots use DAC20 (RGB81). Flash-off dims lit
indices to source literal DAC21 gray, retaining dots. MATRIXOFF lies outside RECALC_LIGHTS' LON..LONEND range:
its 21 is already a DAC component, not a percentage. A second conversion to13
was found and removed during the visual audit. The resulting RGB85 is deliberately
close to the unlit RGB81. Clears, roller transitions,
score/status text, dynamic numbers, animations and match drawing consume source
commands. A winning match enables the original three-sync flash; existing
11/13-sync match cadence and awards are unchanged.

Party Land's and Speed Devils' scheduler priorities/preemption remain tested by
PF4.5/PF7 traces. Source TILT rendering is verified at 648 lit dots; Speed Devils
TILT matches the external original's entire normalized matrix region.

## Frontend and attract sequences

Startup reveal/fade and retained palette steps now use the original images.
Selector preview reveal/clear ordering, original character graphics and text
fades continue through the PF6 state machine. CREATRETF's initial CRTC address
sequence 16..0 is now rendered as the original 128..0-pixel linear scanline pan.
PFTASK's reachable sidebar cycle clears three rows per callback, types one BIOS
character per callback, holds, and restores one saved scanline per callback.
General instructions are the exact original F1/F2/F3/F4/F5/SPACE strings.

WRITEROMCHAR reads BIOS F000:FA6E. The game does not supply that ROM font.
`intro_bios_font.bin` contains the first 128 eight-row characters from the
DOSBox-X default VGA BIOS; its source address, oracle version and SHA256 are
recorded in `pf8-bios-font.json`. The full instruction region was compared with
DOSBox-X. The actual period-PC ROM variant remains unspecified by the game.
`reference_pf8_bios.py` generates an external font-dump helper only in an
isolated oracle directory. Native Go embeds only its recovered bitmap bytes.

Attract uses source triangular one-line scroll and FLASHLIST palette writes;
shared indices preserve chronological packet order. ShowHighsTS is expanded
from each original table's macro. Its rendering consumes the supplied frontend
counter without changing suspended gameplay state. GameOver renders the
single-player AfterDemoModeTS/UrbanOverTS/Once_MoreTS transition to ShowHighsTS.
The score-entry handoff clears the matrix rather than drawing an invented HUD.

Full normal-flow runtime validation of every frontend and bitmap animation
transition is still outstanding; source-derived replay coverage is not itself
proof of full observable DOS cadence or ordering.

## Original input, launcher and tilt

| Input | Source behavior | Native status |
|---|---|---|
| Down (E0 50 / E0 d0) | Hold charge, release SPRINGUP | Both tables, direct and frontend paths |
| Left/right Shift/Ctrl/Alt | Held flippers | Shared original mapping |
| Space | Held physical push plus tilt make-edge latch | Implemented; does not launch |
| P | Pause, any make code resumes; Esc enters quit question | Paused physics/matrix/spring clocks freeze |
| Esc | Source chute/attract/quit transitions | Existing shared frontend path retained |
| M | Toggle main/spring music; jingles remain active | Both tables |
| Mouse | INT33 input and optional button launcher | Deferred with configuration/device options |

There is no separately invented nudge key. Space's TILT0 raises screen position
by600 per physics step to2048, returns by200, and applies the source screen offset
to ball/collision/area behavior. TILTLOGIC adds60 on a make edge, warns above60,
tilts above120, ignores chute/lost/already-tilted states and decays at the source
sync point. Tilt inhibits flippers and source electronic target/bumper effects,
clears lamp/flash state, triggers S_DANGER/S_TILT, and starts TILTTS's long wait.
Party Land also sets its source EOTS state; table-specific remaining area/ejection
logic is retained. Reset occurs at the next-ball source path. It does not invent
a forced drain or bonus forfeiture. Tests cover threshold, latch, inhibition,
message, audio intent, long wait and reset for both tables.

Original spring data is10×23 pixels (Party file83010; Speed7e670), placed at
x304/y556. Charge increases one per sync to32. PUTSPRINGINGFX uses
`real=charge/2-3`, source crop/clear rules and original indexed pixels. Release
uses `vy=-166*charge-low8(jitter)` and source spin masking. SPRINGTASK now follows
DO_ELECTRONICS tasks, as in VBLANK_INT. The native deterministic random input
remains the existing clock convention; exact DOS SLUMP_COUNTERN callback phase
is unresolved and is not described as authentic. Existing explicit-charge
regression entry points remain available for deterministic physics fixtures.

Tilt/plunger audio wiring uses PF5's existing native mixer and fourth voice,
including source plunger volume charge*2. Queued SDL audio never drives gameplay
or matrix clocks. The prior SDR restart/callback-phase uncertainty remains.

## Repeatable comparison workflow and safety

Run from the repository root:

```sh
python3 tools/reference_pf8.py
python3 tools/reference_pf8_visual.py
./tools/go.sh run ./cmd/pf8parity
python3 tools/pf8_oracle.py .tools/pf8-oracle
# External only; manually launch the emitted config with DOSBox-X.
# dosbox-x -defaultconf -conf /absolute/path/to/pf8-oracle.conf
python3 tools/parity_pf8.py list
python3 tools/parity_pf8.py capture WINDOW_ID capture.png
```

The capture helper reports PID and window identity. Select the newly created
oracle window, not any matching title from another running instance. Targeted
XSendEvent is used; there is no global XTest injection. `close` is for the
explicitly created oracle window only. Native screenshot exports use an in-memory
score store; SDL smoke uses a temporary high-score directory. `pf8_oracle.py`
requires an isolated destination and checks original file hashes before copying.

An oracle preparation error was found and corrected during this milestone:
zeroing the disposable TABLE2.HI scores triggered INTRO.ASM MOVSCORE's all-zero
leading-digit counter underflow, followed by a65535-count write across its data
segment. Both DOSBox and DOSBox-X then displayed acid colors and invalid mode
widths. Replacing only that disposable TABLE2.HI restored the palette. The
working game directory's assets/drivers match all original inventory hashes;
only the intentionally user-modified TABLE1.HI differs. Its recorded current
SHA256 is in `pf8-file-audit.json`. No root high-score file was overwritten.
The preparation tool rejects original-INTRO-unsafe scores below10. Faulty oracle
screenshots are diagnostic artifacts, not color or geometry expectations.

Logical-content comparisons remove only demonstrated sample duplication:
640×350 tables ->320×350 by selecting one of each identical X pair; selectors
640×480 ->640×240 by selecting each identical double-scan row. Do not resize
arbitrarily to make artwork fit. Check indexed data/palette/positions separately
from host aspect or physical window proportions. For behavioral checks, record
input edge/hold ordering and source sync phase, not just elapsed milliseconds.

## Scroller boundary records

Linked Party SCROLLE 71d6..727e loads AL=242 (on), AH=96 (off),
uses the DATA tables at 6000/5e00 and advances the original framebuffer.
Its generated literal stores mark run boundaries, rather than every glyph dot.
For A row1, dot1 starts the run and dot6 ends it: dots1..5 are illuminated.
The offline extractor now integrates those boundaries into bitmap artwork.
Speed Devils uses the equivalent linked tables displaced by90h. Score records
remain full glyph records and do not receive this transformation.

The original Party capture `verified-x-party-scroll-long.png` contains
`DO NOT USE DRUGS`. Source SCROLL_TEXT1 at dot offset675 matches all2560
matrix dot states, and all10560 pixels including gaps and palette. This is
an artwork checkpoint at an observed partial SCROLLE visit, not a claim that
675 corresponds to an integral native scheduler tick. Both tables' recovered
fonts pass the original capture test. Drawing still consumes existing state.

## Visual checkpoints and measured scope

Measurements are reproducible with `python3 tools/compare_pf8.py`; exact counts
and normalized oracle hashes are in `pf8-parity-measurements.json`. Raw captures
and native frames are under `pf8-checkpoints/`. Manual inspection accompanied
logical comparison. Counts below describe the stated region only.

| Required checkpoint | Verified result | Difference or remaining work |
|---|---|---|
| Griffin/startup |76800/76800 pixels, full logical frame | Acid-color captures came from the unsafe disposable HI fixture; excluded |
| Selector |153600/153600, whole frame; instructions10080/10080 | Complete temporal raster/wipe sequence has not been sampled at every phase |
| Party attract |10560/10560, table-name matrix | Initial AfterDemoModeTS versus ShowHighsTS prefix ordering still needs a synchronized runtime trace |
| Party initial chute |112000/112000, whole settled frame | Other captures differ in lamp callback phase; do not change source palettes to match a different phase |
| Party scroller/status |10560/10560 matrix scroller pixels | Whole field scroll origin differs in the deliberately isolated glyph checkpoint; not a whole-frame match |
| Party charged/released |170/170 charged spring crop | Release state/velocity and frame sequence have source tests; exact hardware jitter phase remains unresolved |
| Party tilt |10560/10560 matrix, 648 illuminated dots | Late captures can already be preempted by ball loss; whole field timing not certified |
| Speed attract |10560/10560 PF title matrix | Full attract loop ordering/cadence not synchronized with DOS |
| Speed initial chute |112000/112000, whole settled frame | Flash-phase caveat as for Party |
| Speed bitmap animation |Original delta-frame hashes and 69 matching live original GEAR captures | Controlled source-program/data probe; natural rule trigger not claimed. See continuation below |
| Game over/high score |10560/10560 high-score-entry matrix | Entire game-over, initials and return sequence has not been certified phase by phase |

The charged spring crop is source-local, x304..313/y297..313 in the composed
initial frame. Tilt/high-score/scroller comparisons include matrix position,
on/off palette, all gaps, and text shape. Exact settled initial frames establish
viewport/crop and artwork composition for both tables. No claim of whole-game
pixel-perfect parity is made. Historical CRT calibration remains unresolved;
DOSBox raw sample duplication is known presentation geometry, not a native bug.

## Verification and changed files

The required full command `./tools/go.sh test -p=1 -count=1 ./...` completed.
All gameplay, physics, audio, frontend, platform, presentation and PF7 tests
passed. Its only failure is the unchanged inventory assertion for the user's
modified root TABLE1.HI. That test remains strict; the file was preserved.
The later source-scroller capture test also passes. The required build command
`./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies`
succeeded. Party's direct deterministic script reports1200ticks,
score000002300000, ball2, phase0. Original PF7 fixtures were not rebaselined.

Focused coverage includes CRTC dimensions and pitch; visible composition;
occupied ball geometry; startup DAC/bank/fade data; matrix fonts and clipping;
scroller cadence and real captured artwork; all source animation delta hashes;
existing animation/priority scheduling; Down hold/release; Space threshold,
inhibition, cues and per-ball reset; spring charge/crop/release; both table
presentation paths; pause; selector pan/sidebar boundaries; table reload state;
high-score flow; and persistent SDL logical/physical sizing. Source-defined
Party CLEAR4 and Speed CLEAR1 game-over differences are tested explicitly.
Tests do not advance state while rendering, and audio queue size is not a clock.

Principal added/changed areas:

- `internal/presentation/`: shared video composition, matrix artwork renderer,
  extracted content, attract/game-over programs and chronological lamp palettes.
- `internal/assets/`: shared DAC conversion and original BIOS glyph data/provenance.
- `internal/physics/`: original held/edge controls, tilt/push and spring state.
- `internal/partyland/`, `internal/speeddevils/`: original command visual wiring,
  composition, palettes, reset/cue integration and deterministic tests.
- `internal/frontend/`, `internal/platform/`: startup/selector/sidebar effects,
  shared original controls, fresh attract state and persistent-window rendering.
- `cmd/pinballfantasies/`, `cmd/pf8parity/`: shared direct-path controls and named
  logical checkpoint exports; no external oracle invocation from native code.
- `tools/reference_pf8*.py`: static source/data extraction and independent fixtures.
  `tools/pf8_oracle.py`, `parity_pf8.py`, `compare_pf8.py`, `smoke_pf8.py`:
  isolated original preparation, targeted external captures, comparisons and SDL smoke.
- `analysis/pf8-*.json`, this report and checkpoint images: evidence, fixtures,
  hashes, dimensions and measured comparison scope. PF6 RGB fixtures were corrected
  using source DAC conversion; PF0 indexed assets and original ASM were preserved.

## Remaining uncertified behavior

The original validation list remained open before user acceptance. The continuation below validates live bitmap artwork and fixes
selector music continuity. Full startup/selector temporal coverage, synchronized
initial attract ordering, and full game-over/high-score return sequencing remain
partly uncertified, with captured evidence and explicit limitations below.
These are actual validation gaps, not assertions that tests establish runtime parity. Intended analog CRT
pixel aspect remains unproven. Native 1:1 high-mode artwork scaling is an explicit
current policy, not a documented historical monitor calibration. The original
spring jitter/audio callback phase and precise historical BIOS ROM variant remain
uncertain. Mouse/options UI and configurable low-mode selection are deferred.
No artwork alteration, modern HUD, emulator integration, new tables or mixer
redesign was added to compensate for these uncertainties.

Final SDL/X11 smoke passed after these changes, using its own process/window and
an isolated temporary score directory. One physical window survived resizing,
both tables, pause/resume, quit questions and selector returns. SDL audio queue
resets and empty-queue observations were both0. The test used dummy audio;
this confirms lifecycle/control wiring, not a new acoustic parity measurement.

## Runtime-validation continuation — 2026-10-01

This continuation addresses the four remaining runtime items and the separately
reported selector music stutter. Previously verified gameplay, assets, geometry,
physics, PF7 fixtures and mixer semantics were not reimplemented. User input may
have affected the first desktop capture session. Its `original/timeline.json`
is marked provisional; its temporal ordering is excluded from certification.
All decisive new original captures run on an owned private Xvfb display, with
timestamped synthetic inputs and a fresh isolated game directory. No key typed
in the desktop/chat can enter that display. No original user process is killed.
Xvfb was downloaded/extracted locally under `.tools`, without a system install.

Evidence is under `analysis/pf8-runtime-validation/`. `README.md` documents
reproduction and which probes modify copied data. Configurations use surface
output, no aspect correction and no scaler. Sample duplication is checked byte
for byte before normalization; this is not a new CRT/pixel-aspect assumption.
`side-by-side.png` juxtaposes original/native title, Digital Illusions, and text
fade frames. That contact sheet is scaled for inspection; pixel measurements
use the logical frames/crops in `temporal-comparisons.json` and regression tests.

### Selector music stutter: resolved, independently of DOS resume phase

Root cause: `platform.ShowFrontend` called `AudioDevice.Suspend(false)` whenever
a mode change entered Selector. Returning from SelectorText therefore paused
the SDL device, cleared its queued PCM and restarted preroll, although the
same Go tracker was still playing. This was a native queue lifecycle bug, not
a module position reset and not the inherited DOS STOP/PLAY uncertainty.

`Runtime.AudioSource` now identifies the actual player/session independently of
visual mode. The platform resets the queue on a producer change or pause/resume,
not a selector/text transition. Both modes retain the identical tracker object,
order, row, tick, private clocks and sample voice phase. The test exercises
4000 ticks each for initial INTRO and returned-menu MOD2, comparing every PCM
block and complete Player state with an uninterrupted control. No tracker
recreation, rewind, queue clearing or new preroll occurs at these visual changes.

Real SDL **pulseaudio** output was recorded from the Pulse sink monitor using
ffmpeg, with thirteen Space make edges in the normal native frontend and an
isolated score directory. The final run reports lifecycle clears0/starts1,
overrun queue resets0/empty observations0. FFT correlation locates each one-second
PCM block independently; all20 blocks have correlation1, gain1 and a constant
host latency. The entire960000-sample,20-second interval is byte-identical to
continuous native INTRO PCM, with no lost/repeated samples or alignment drift.
This is real output evidence, not a dummy-audio queue test or a claim of subjective
listening. WAV, input timestamps, SDL logs and measurements are retained.

An A/B build using a temporary Go overlay restored only the former platform
mode condition. The same real output probe produced lifecycle clears7/starts8,
alignment drift26757 samples and minimum correlation0.622119, failing the
continuous-PCM comparison. Its old queue-reset/empty counters both stayed0:
those counters alone demonstrably missed the audible gap mechanism. The overlay
did not revert working source. Known original DOS callback/resume uncertainty
remains separate and is not advertised as solved by this fix.

### Startup/title and selector temporal corrections

The reported solid-color Pinball Fantasies title backdrop was a real native
composition bug. INTRO invokes BIOS mode12h (cleared VRAM index0), draws its
640×178 logo at y150, then fades the **whole image DAC palette**. Pixels outside
the logo must therefore use its palette entry0, not permanently black RGB.
Native now fills the whole640×480 surface from index0 before applying the same
source fade. Clean original `startup-0580.png` matches native held title
`startup-13-019.png` on all307200 logical pixels, including all four corners.
No image-specific recoloring table or artwork change was introduced.

INTRO's Digital Illusions/Viking image is loaded at `vikingpos=240`, but its
CRTC address is `vikingadr=80*247`. The upper126-row IFF is followed by the
lower130-row IFF starting at buffer365. Native previously showed buffer240,
seven rows too high. It now shows upper source row7 at screen0 and the lower
image at screen118, including source lower row121 at screen239. The overlapping
upper last row is correctly overwritten. Held frame matches all76800 pixels
after verified original 2× sample duplication is removed. Source crop/fade
regression added; the already-verified griffin path was untouched.

Selector corrections are confined to the previously uncertified text sequence:

- `TEXTLISTA` has ten pages (1,2,5,6,9,10 use HITEXT), then a zero terminator
  wrapping TEXTPEK to slot1. Native's eight-page modulo omitted reachable pages.
- WAITEND toggles BANPEK but selects HITEXT for the preceding preview pair.
  Native now uses PreviousPage for its high scores, including the original
  fixed24-cell headings and rank/score layout. TEXT2/TEXT3/TEXT4 blank rows and
  trailing alignment spaces follow the source strings, rather than approximate
  centered variants.
- During RASTRACLEAR the sidebar keeps the preceding preview banks. SHOWTEXT
  then JULIUS-loads table slot6/Party and disables the raster palette split.
  Text/font/high-logo pixels consume that bank rather than assuming an image's
  own CMAP is still active.
- fade3/fade3b change DAC entries1,12,14. Their integer levels now pass through
  the shared six-bit DAC conversion; the former `value*255/63` caused one-unit
  RGB errors. At fade3b20 step3 (level16/20), the entire text/logo region
  x128..639 matches the clean original capture on all122880 logical pixels.
  Both original scanline duplicates are checked. Captured copied HI bytes are
  pinned separately for this test; it never reads/writes mutable user scores.
- Space/Enter are consumed during wait_sync2 after the reveal, and end its hold
  while preserving fade3b20. Space no longer jumps directly to previews or skips
  the initial fade. Enter toggles the upcoming pair as in wait_sync2's KO branch.
  The old PF6 test expecting immediate text exit encoded a source mismatch;
  it was replaced with assertions for ignored reveal keys and all20 exit syncs,
  not weakened or updated from a native snapshot.

The clean startup capture contains1198 samples at nominal40ms intervals.805
match native full logical frames, spanning85 exported phase states and every
reachable startup picture/fade family. These counts are **not a completeness
percentage**: the capture includes DOS loading text, later selector states,
identical black/hold frames and unsampled fade steps. Original holds rely on
music counters; matching pictures alone does not establish every handoff tick.
Full startup/selector temporal certification remains unresolved. The complete
ten-page selector cycle was source-tested, not all captured live; sidebar
instruction/logo restoration phases remain unaligned, including the original
vertical PF logo reveal. F-key/Escape exit paths through the text fade are also
not certified. These are remaining native temporal coverage/behavior issues,
not palette differences blamed on DOSBox geometry.

### Live bitmap animation: verified controlled renderer/program probe

`bitmap-oracle-fixture.json` records an isolated Speed Devils **data-only** probe:
the existing10-byte GEARTS at file106739 is copied over the first ShowHighsTS
entry at108001. The source macro, original animation deltas, executable code,
VGA renderer and callbacks remain unchanged. Both original/fixture hashes and
before/after bytes are recorded. Superseded INIT_MATRIX/NODOT pointer probes
produced no usable matching frames and are retained as rejected evidence.

`bitmap-data-isolated/` yielded69 matching live matrix captures. Every one of
the five distinct GEAR images matched all2560 dot states and all10560 logical
pixels, including on243/178/69, off81/81/81 and black gaps. GEAR frame0 and5
have identical content; label5 in the match list does not assert a runtime
phase distinction. Transitions follow5,1,2,3,4 repeatedly. Source durations are
three71Hz syncs (~42.25ms); sampled transitions alternate ~25/~50ms and remain
within one25ms sampling interval of that cadence. Exact individual callback
timestamps are not available at this capture rate. A native regression checks
all69 raw captures, including both duplicated horizontal pixels.

This certifies live original bitmap presentation and sequence under the source
program. It does not certify natural gear-award triggering or a whole gameplay
frame. Existing scheduler/priority/timing tests and PF7 fixtures remain intact;
the visual renderer still consumes their state without advancing another clock.

### Initial attract ordering: observed prefix, unresolved absolute handoff

Fresh unmodified Party and Speed captures now start before first table matrix
content, on private displays.111/158 Party crops and242/425 Speed crops match
native source Attract states exactly. Visible order is PF title → THE REAL
SIMULATOR roller/hold → table title → subsequent clears/scroller. Distinct
Party roller states132→157 span ~0.361s (25 source syncs/~71Hz); Speed contains
39 uniquely identified native states through the later clear/scroll phases.
Blank/preload frames and partially visited scroller states are retained, not
filtered into a claim of full-match timing. Full field/lamp phase is excluded.

The unresolved point is the original loader-to-first-matrix dispatch. FANTASIE
JUST_ONE_TIME_RESET initializes NODOTCOUNT to NOT_PLAYING-2; NODOT can select
AfterDemoModeTS versus ShowHighsTS according to DEMOMODE/LOOSING/player state.
The current native fresh Attract starts ShowHighsTS. Captures establish the
first **visible** PF-title prefix, but neither initial callback state nor the
hidden AfterDemo prefix was traced sufficiently to equate source dispatch with
native tick0. `first_time_you_fool` participates in that original flow. No
timing fixture was rebaselined to guess the missing dispatch. This remains
unresolved source/runtime ordering, not an aspect correction problem.

### Game-over → initials → return: explicitly unresolved

The controlled ending data probe copies original after_xxballTS into the first
attract dispatch and sets a saved qualifying score123450/player1, using valid
copied HI100 records. It reaches the correct original high-score-entry screen.
However, only B of separately sent A/B/C make edges reached entry; DEMOMODE's
main checkcheat/SCANCODE reader consumed other edges. It ended with `(B  )`,
not a complete normal return. This probe's entrance through attract differs
from natural ball-loss LOOSING/DEMOMODE state; it cannot certify the ending
reader/return sequence. Code and input handling were not patched to force a pass.

A second private-display run kept TABLE2.PRG entirely original, changed only
copied configuration to three balls and copied safe qualifying HI100, and
attempted10 launch/tilt/drain cycles over108.33 seconds. Its recorded
qualification event is false; gameplay/bonus/extra-ball messages were still
visible. The run did not reach the high-score reader, so no complete game-over,
three-initial acceptance, entry wait and frontend return observation exists.
Exact ball-loss/input phase was not established; a failure to force drains is
not evidence of parity or proof that DOS tilt is broken. Both full timelines
and captures remain available. Existing source-derived ending/score persistence
tests continue to pass, but do not replace this missing runtime oracle.

### Final continuation verification and scope

Required full tests were rerun on final source. All audio, frontend, Party,
physics, platform, presentation and Speed Devils tests passed; only the strict
original inventory assertion for user-modified TABLE1.HI fails. It was neither
overwritten nor exempted. Its preserved SHA256 is
`b78554cfbe71dce39e715634297ec4ac0fe56462ed2c414169e861251b692c2b`.
The required native build succeeded. Party's1200-tick deterministic run remains
score000002300000/ball2/phase0. PF7 expected fixtures were not changed.

Final SDL smoke runs on an owned isolated X display, retaining one window and
the requested physical resize through both tables, pause/resume, quit questions
and returns. The desktop run had an unexpected pause-resume key and is excluded;
the private-display test passed. Dummy audio is used only for this lifecycle
smoke; the separate final selector test uses real PulseAudio and PCM comparison.
Legitimate table/pause lifecycle clears in the smoke are not selector stutter.
The final concurrent smoke logged9 dummy empty-queue observations (earlier
isolated smoke0); the real Pulse selector run still reports0. The dummy smoke
is evidence for window/control wiring only, not an acoustic continuity claim.

Continuation changes: frontend model/render/runtime and their focused tests;
platform frontend audio producer switching and lifecycle counters; live
presentation-capture test; `cmd/pf8temporal`; tools for owned X displays, original
runtime capture/data fixtures, real selector audio and PCM/content comparison;
this report, reproduction notes and evidence artifacts. No original ASM/game
data, user HI, tracker/mixer algorithm, physics or table-rule fixture was altered.

Full DOS temporal parity is **not certified**. Live bitmap renderer validation and native
selector music continuity are resolved. Startup/title/text corrections are
implemented and measured. The remaining temporal/dispatch/ending uncertainties
above have concrete source and runtime evidence; they remain explicitly open.

### User acceptance after validation

The user explicitly said to leave the selector sequence and game-ending flow
as they are and approved them. Further work/certification on those two items is
therefore waived by the user, rather than blocked or claimed complete by tests.
The user additionally confirmed listening to the music and hearing no stutter.
Together with the real Pulse PCM proof, this resolves the requested native
selector audio bug. Previously documented startup/initial-attract callback
alignment uncertainties remain documented differences in evidence, not a
request for more approval or a reason to continue the waived sequencing work.
