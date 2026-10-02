# PF11 source inventory (before implementation)

Authority: reference/original-dos-source/INTRO.ASM, FANTASIE.ASM,
BALLCODE.ASM, PLAND.ASM, SDEV.ASM, SHOW.ASM, STONES.ASM.

INTRO keyread F5 -> f5 -> showmenu -> menu -> showpics. The text-page
path queues menunext then invokes showmenu after showtext. showmenu clears
preview rasters, installs Party Land preview palette via julius slot6, disables
CHANGE16PAL, uses the existing 640x240 planar screen and linked 18px-advance,
14px-high font. Heading OPTIONS MENU at y14; blank line; six rows y50..140,
blank line then SAVE AND EXIT y176. Arrow '>' x175, y50+18*row, with the
last row displaced one extra line. writepage centers within x164..596,
24 characters; fixed-width option rows have their values at character16.
Up/down (72/80) wrap seven rows; Enter (28)/Space (57) invoke the row;
Esc (1) returns without undo. Other keys ignored. No menu mouse reads.
L_SAVE sets LOOSER_BUT_SAVER only, explicitly NOT on disk. Config writes
are at INTRO shutdown/table-program handoff (quit, quit_ing, not_enough),
followed by BEFORE_STARTING resident transfer. No music calls in showmenu.
Fade3B5, fade3 40, fade3b40 use font banks/pages; sidebar PFTASK continues
with options_info instead of general_info. Returning reenables preview rasters.

| Offset | Field | Bytes/display | DOS default | Runtime reader |
|---|---|---|---|---|
|0|S_BALLS|0=3, 1=5 (XOR1)|0|NO_OF_BALLS at table initialization|
|1|S_ANGLE|0=HIGH, 1=LOW (XOR1)|0|TABLE_ANGLE subtracts3 from Y of no_of_ramps entries|
|2|S_SCROLLING|0=HARD,1=MEDIUM,2=SOFT (increment/wrap)|1|F11=20,F10=11,F9=9 SLIME_FACTOR|
|3|S_IM|0=ON,1=OFF (XOR1)|0|MUSIC_TOGGLE at table initialization|
|4|S_RESOLUTION|0=NORMAL,1=HIGH (XOR1)|0|HI_RES before TABLE_ANGLE/INIT_GFX|
|5|S_MODE|0=COLOR,1=MONO (XOR1)|0|COLOR_2_BW and pelle_2_bw at table initialization|

TOGGLAREN starts six zero bytes; failed LOAD_TOGGLAREN sets scrolling1.
DOS reads six bytes without checking AX length or validating contents. Invalid
scrolling indexes can read outside ZARINEN. Native validation must reject these
unsafe records deterministically; use original missing-file defaults. Valid
installation PINBALL.CFG is 01 01 02 01 01 00 and must remain read-only.
No settings are reapplied opportunistically during a running table. M modifies
muzik_off, s_main/s_spring, LASTJINGLE/JINGLEJUMPCNT, not resident/persisted
TOGGLAREN. New table program initializes from config. Native new games in the loaded table retain runtime M; reloading from
the selector restores config. Frontend music ignores S_IM. Other jingles and sound effects
continue; check_m_off redirects returns <=lastmainpos to emptyjingle, preserving
JINGLE_READY_LOGIC/ANIM and deterministic tracker/cue clocks.

SETSCREENSTART: target SC_Y - ((height-33)/2 -8-20), clamp to
0..576-(height-33), add33; SCREENFORCE overrides and immediately sets raster
in sixteenth-lines, SCREENFORCE2 overrides desired target before smoothing.
GLAPP_SIZE=0. delta=(target-(RASTERPOS SAR4))*SLIME_FACTOR (16-bit signed),
RASTERPOS += delta SAR2; gap clamp uses (height-33)/2-20-8 upward and
(height-33)/2+20-8 downward. HIGH thresholds130/170, NORMAL75/115.
SCREENPOSY is added only at hardware output. Table captures force the viewport.

INIT_GFX: SET_MCGAB, SET240, optionally SET_360X350, SETSCREENSTART, SETSPLIT.
SH_LO240, SH_HI350, SPLH33. NORMAL SET240 uses double scan and line compare
414-1 physical scanlines -> 207 logical playfield +33 logical matrix. HIGH
line compare317-1 ->317+33. Matrix stays at VRAM origin, rendered via line
compare, not an overlay. Native logical sizes320x240 and320x350 retain square
logical pixels and independent persistent host-window size. DOS also changes
physics sampling/constants between mode paths (tt/nn=5/6); PF11's requested
same-state physics model retains the established native interrupt schedule.

MONO: pelle_2_bw averages each DAC RGB triplet using integer division by3.
COLOR_2_BW averages every LON RGB triplet after RECALC_LIGHTS conversion,
including MATRIXON. MATRIXOFF is outside LONEND and remains its direct DAC
write (Stones retains chronological overlap with MUMMY). No pixel-index change;
no weighted grayscale. Conversion is gameplay-only, not INTRO selector/options.

SOUND.CFG boundary: INTRO INIT_SOUND (3248), FANTASIE INIT_SOUND (5544) read
only13 bytes into DRV, a zero-terminated DOS sound-driver filename, then execute
that driver via INT21/4B00. INTRO missing-file diagnostic directs the user to
SETSOUND. No F5 path reads/writes SOUND.CFG. Supplied record is25 bytes:
'SBLASTER.SDR\0' then 12 driver-specific bytes 00 01 00 00 03 00 00 04 00 00 FE 01.
SETSOUND.EXE and SDR binaries exist; their source is not in the supplied source
tree. The completed binary-reader audit below resolves the consumed port/IRQ/rate
fields; unused metadata remains explicitly opaque. Native MOD/effect playback replaces the DOS driver; obsolete hardware
configuration is outside the six-option menu.

# Completed native implementation

## Application and persistence

`internal/settings.Config` owns the six typed source bytes in original order.
The menu mutates the frontend copy. Factories snapshot it at table creation;
balls, angle, scrolling, resolution and color mode never change under a running
session. Runtime M is separate live state. `SessionConfig` carries that state to
another game in the same loaded table, while selector reload takes S_IM again.

The default native writable path is `${XDG_CONFIG_HOME}/pinballfantasies/PINBALL.CFG`
(or `os.UserConfigDir`'s fallback). `-config-dir` overrides only settings storage;
`-high-score-dir` remains independent. Native settings take precedence over a
valid read-only installation PINBALL.CFG seed. Missing records use original
00 00 01 00 00 00 defaults. Invalid/truncated/oversized records use those defaults,
not the DOS loader's unchecked partial/out-of-range memory access. Permission and
I/O failures propagate. Writes create a private temporary file, write exactly six
bytes, fsync, close and rename atomically in the destination directory.

SAVE AND EXIT and Esc both retain edits in frontend memory. Neither performs a
DOS disk write. Native writes at table handoff and orderly program quit/SDL close,
matching INTRO's SAVE_TOGGLAREN sites. Reopening Options shows current edits;
closing/restarting shows the saved bytes. Tests use t.TempDir and desktop smoke
uses TemporaryDirectory for both stores. Original `.HI`, PINBALL.CFG, SOUND.CFG,
PRG/MOD and ASM inputs remain untouched. The local score-file digest is recorded
in `pf11-validation/original-user-files.sha256`.

## Balls and angle

S_BALLS0 sets3,1 sets5 through each table's established base-ball setter. Extra
balls, match awards and per-table reset/countdown logic are unchanged. The suite
checks both counts for all four tables and proves changing the frontend record
cannot change an already-created session's limit.

TABLE_ANGLE changes Y only, subtracting3 for LOW. Per-session ramp arrays are
copied, preventing cumulative changes and cross-session asset corruption. The
normalized native source-high constants are:

|Table|HIGH `(GX,GY)` ramp entries|
|---|---|
|Party|0,10; 2,14; -2,14; -4,16|
|Speed|0,10; 0,15; 0,25; -1,10; 0,20; 12,10|
|Gameshow|0,10; 4,12; 0,14; 2,9; 6,13|
|Stones|0,10; -10,5; 0,-10; 5,0; 5,15; -10,12; 2,15; -8,12; 3,10; 4,13; 7,10|

LOW subtracts3 from every listed GY, including negative/zero ramp gravity; GX is
untouched. Initial GRAVY8 stays8 until CHECK_RAMPS, as in source. Tests check every
entry on all tables and controlled signed integration over3 and12 SC_PROGRAM
passes. Existing direct table constructors retain accepted legacy fixtures;
frontend factories explicitly apply configured settings.

## Scrolling and forced viewports

HARD/MEDIUM/SOFT are factors20/11/9 in source signed integer IMUL/SAR2 smoothing,
not floating point speed multipliers. The source clamp, center offset28,
mode-specific up/down gap limits and sixteenth-line remainder are retained.
GLAPP_SIZE is compiled0; there is no invented dead zone. Raster updates do not
change the ball, gameplay Tick, task slots, cue clocks or sync rate.

SCREENFORCE remains an immediate authoritative viewport override. SCREENFORCE2
is a separate smoothed target, before multiplication/gap clamping. Source drain
and Gameshow wheel paths use it in configured sessions: wheel target220 HIGH,
270 NORMAL, then reset-1 on eject. Legacy direct PF9 fixtures retain their
previous187 hard-force path; Options-created sessions follow the recovered
source target. Capture/rule timing and score logic are unchanged. A dedicated
wheel fixture verifies these targets; four-table mode fixtures verify immediate
capture force187. The checker artifact changed only the reported line number of
an existing Gameshow constructed-label expression (515→523), with identical
reachability/registrations.

## Music and M

S_IM1 calls the existing table MUSIC_TOGGLE semantics before tracker attachment:
main/spring cues route to the table's source empty order. Jingles, attract music
and sound effects remain active. Returning background orders are redirected to
empty without dropping ready-animation/ready-logic events. Silent cue timing
continues deterministically, independent of SDL queue state. Frontend INTRO/MOD2
music ignores S_IM. No SOUND.CFG device setting is substituted.

M never writes the config. It survives a new game in the loaded table, including
return to attract. Reloading the table from INTRO restores persistent S_IM. The
native factory initializes the effective music state directly when creating that
new game, avoiding an obsolete queued OFF task when runtime M has turned ON.
Four-table tests cover ON/OFF, M, another game, selector reload and immutable
stored settings. Existing table audio/jingle-return/effect regressions pass.

## Video pipeline and palette

NORMAL is320×207 playfield plus320×33 matrix =320×240 logical frame. HIGH remains
320×317 plus320×33 =320×350. SET240 uses build-time HIRES=false (distinct from
runtime HI_RES), leaves CRTC09h double scan enabled, and SETSPLIT selects line
compare413 physical scanlines. HIGH selects316. `NormalCRTC`/`HighCRTC`, composition
and logical frame tests record these values. Matrix font, text, bitmap, on/off
and chronological command processing are shared unchanged. No overlay or hidden
matrix is invented. Viewport range is0..369 NORMAL and0..259 HIGH. Attract sweeps,
initial chute, spring graphic and frontend pause/entry panels all use the active
height; Party's attract path was specifically fixed and regression-tested.

SDL recreates only the streaming texture for logical-frame changes. It retains
one physical window and renderer. Manual resize stays authoritative. Gameplay
uses square logical pixels in both modes; the established640×240 selector still
has its separate640×480 double-scan mapping. No unconditional aspect stretch or
CRT temporal parity is claimed.

MONO uses integer DAC mean `(R+G+B)/3` for the indexed base palette. Lamp RGB is
first recalculated from percentages via `*162>>8`, then averaged, then dimmed for
OFF lamps. This matches COLOR_2_BW's position after RECALC_LIGHTS. Each later lamp
write applies the converted triplet at its existing chronological write site,
so overlapping lamp indices remain correct. MATRIXON uses the same conversion;
MATRIXOFF retains its direct source DAC data, including Stones' overlapping
three-entry write. Artwork/sprite indices and shared decoded tables are unchanged.
Color/mono six-bit palette and lamp-rounding fixtures accompany screenshots.
Selector and Options stay in their original color palette regardless of S_MODE.

## Options frontend and audio lifecycle

F5 from selector opens Options. F5 from the selector text hold queues Options
through that page's exit fade, matching menunext. The menu uses linked INTRO font
art, Party preview palette, original logo and BIOS sidebar font, the exact heading,
labels, padded values, rows and '>' glyph. Seven rows wrap via Up/Down; Enter and
Space toggle/cycle or leave from SAVE AND EXIT; Esc leaves without undo. There are
no arrow-left/right changes, modern key controls, widgets, key remapping, or mouse
UI. Repeated entry resets the cursor but retains edited settings and table identity.
The forty-step font palette fades use the source denominator and integer rounding.
The first five syncs retain the preceding frontend page, followed by the menu fade;
source VRAM page-switch/raster-interrupt temporal ordering is represented as native
whole frames rather than cycle-level VGA execution.

PFTASK sidebar continues its original clear/character/restore cycle, selecting the
exact options_info strings. INTRO's only menu input calls are the scan-code handler.
Mouse availability is checked in FANTASIE INIT_MOUSE/SPRINGSTEEN and affects DOS
relative mouse plunger input; it neither changes options defaults nor exposes a
mouse-controlled settings menu. Existing native keyboard plunger remains available.

Selector, selector text and Options share the same actual audio producer. The
existing AudioSource identity logic is retained; no menu visual transition replaces
the player or clears SDL output. A sample-exact focused test compares PCM and all
tracker state against an uninterrupted control player over repeated menu entries,
all row changes and return fades, for both INTRO and return MOD2 producers.

## SOUND.CFG boundary, completed evidence

`analysis/pf11-sound-boundary.json` and `pf11-sound-reader-disassembly.txt` add a
read-only static expansion of the supplied SBLASTER.SDR using the proven PF5 SP
packet decoder. No original executable is edited or used by the native runtime.

The supplied25-byte record has filename13, followed by four three-byte cells
(leading byte plus little-endian word). The leading-byte/tag meaning is opaque
and ignored by the inspected driver reader. Reader18eb..1a33 seeks to file14,
17 and20, reads one selector byte at each, masks with7, and indexes its own tables:

|Offset|Supplied selector|Verified SBLASTER reader result|
|---|---|---|
|0..12|SBLASTER.SDR +NUL|DOS driver to execute|
|14|1|base I/O port220h (table210h..260h)|
|17|3|IRQ7 (table2,3,5,7)|
|20|4|21000Hz, extra quality flag255|
|23..24|word510|external setup metadata, not read by this reader|

The complete raw cells and mapping-table words/hashes are in the JSON evidence.
Other SDRs own their own interpretation; this is not asserted to be a universal
IRQ/DMA layout. In particular no DMA field is invented for the inspected reader.
SETSOUND.EXE creates this external record; INTRO's missing-file diagnostic names
that program. INTRO and FANTASIE execute the filename and the sound driver then
reads its hardware parameters. F5 never touches SOUND.CFG. Port and IRQ have no
native equivalent. Driver-specific quality controls its DOS renderer; native
output is the established48kHz signed16 stereo mixer, with no device picker added.
Unused setup metadata is deliberately not reverse engineered into a native UI.

## Fixtures, checkpoints and changed files

`tools/reference_pf11.py` emits `analysis/pf11-settings-fixtures.json`: input-source
hashes, six-field record/defaults/order/choices, original menu geometry/help/keys,
normal/high split geometry and deterministic integer scroll checkpoints. The Go
scroll regression consumes these fixed checkpoints rather than recomputing its
expected values from the implementation. `tools/reference_pf11_sound.py` emits
independent binary boundary evidence.

`pf11-validation/checkpoints/` contains default Options, every one of seven rows
selected, changed options, return, and all four tables in both logical modes:
attract, initial chute, spring, text, source bitmap, normal gameplay and forced
capture. Per-table/per-config palette JSON snapshots cover color and mono. These
are source-derived native logical checkpoints, not measured full CRT parity.
Bitmap names are sorted for deterministic fixture selection.

Primary code changes:

- `internal/settings/{settings.go,settings_test.go}`: typed record and safe storage.
- `internal/frontend/{model.go,runtime.go,options.go,render.go,sidebar.go,options_test.go}`:
  F5 journey, source controls, immutable factory snapshots, M lifecycle and rendering.
- `cmd/pinballfantasies/main.go`, `internal/platform/frontend.go`: isolated config
  root and original F5/Up/Down scan codes; persistent host-window/audio path retained.
- `internal/physics/{settings.go,settings_test.go,ball.go,render.go,rules_boundary.go}`,
  `internal/assets/simulation.go`: session ramps, source scroll/forces and viewport.
- Four tables' `settings.go`, `render.go`, `flow.go`; Party `session.go`; Gameshow
  `rules.go`/`settings_test.go`: settings propagation, palette chronology, attract
  extent and source wheel target; existing scoring/task/matrix rules retained.
- `internal/presentation/{matrix.go,matrix_test.go,video.go}`: explicit viewport
  composition, mode-aware palette conversion and normal CRTC evidence.
- Party `game_test.go`: explicitly preserves the old fixture config and strengthens
  the accepted1200-tick score/ball assertions; no oracle rebaseline.
- `tools/smoke_pf11.py`, `tools/pulse_pf11.py`: own-window X11 and real PulseAudio
  validation using temporary storage. The matrix checker itself is unchanged.

## Validation and deliberate differences

Final full-suite output: `pf11-validation/full-tests.log`. Only failure is
`TestOriginalInventories/game-inventory.json/TABLE1.HI`; that file and test remain
untouched. All other packages pass. Focused Options/config/physics/presentation/
Gameshow tests pass (`focused-options.log`). Build command succeeds (`build.log`).

Party oracle remains1200ticks, score2300000, ball2 with explicit `settings.Legacy()`
(3 balls, LOW angle, SOFT scroll, music ON, HIGH, COLOR). Frame SHA256 remains
`aec01b3a07e1a5ba10b3c635777a6742f4abd41c09899903ea532b98d8522913`.
PF7, PF8, PF9, PF10, stereo, high-score and missing-target panic regressions pass.

Matrix checker (`matrix-check.log`): Party63/63/0; Speed53/66/0;
Gameshow54/67/0; Stones71/88/0. No reachability or panic behavior was weakened.

X11 journey (`x11-smoke.log`) passes Options/change six options/return/F1..F4
play/abort/return/Options/quit; the same window ID remains valid and800×600 manual
resize survives logical changes. Only PINBALL.CFG is written in the temporary
store, with exact expected six bytes; no score records are written.

Real PulseAudio (`pulse-on.log`, `pulse-off.log`, `pulse-options.log` and respective
metrics/monitor/WAV artifacts):48kHz signed16 stereo, different channels, zero
queue resets and zero empty-queue observations. Options-only has lifecycle clears0,
starts1. Full four-table journeys have clears20,starts13 from producer and pause
boundaries; no extra Options clear/restart. Both configured music states and M
transitions, effects and jingle-preserving semantics are exercised. Dummy-audio
X11 output had empty observations under desktop scheduling; real Pulse output has0.
Sample/cue determinism is established by focused PCM and existing audio regressions,
not by treating the monitor capture or host queue as a gameplay clock.

Deliberate native policies: invalid DOS records are rejected safely; installation
assets are seed-only; writes are atomic; physical window size is independent; VGA
raster execution is represented as logical frames. Per the PF11 requirement that
resolution changes presentation without changing gameplay, NORMAL retains the
accepted native71Hz integer physics and source-high normalized constants. Original
DOS normal mode uses60Hz and several5/6-scaled physical constants/material/flipper/
spring paths; those hardware-rate compensation branches are documented, not applied
as an extra native difficulty/rule change. HIGH/LOW angle still applies the exact
source subtraction to the normalized table-specific ramps in either presentation.

All six F5 menu options are implemented end-to-end. No original F5 configuration
feature remains missing. DOS sound-card setup, IRQ/port/driver quality and opaque
external SETSOUND metadata remain outside PF11; no invented native hardware UI was
added. Full CRT temporal parity and universal SDR tail-layout archaeology are not
claimed.
