# PF0 source map and PF1 dependency cone

Every architecture assertion below has a status. VERIFIED means directly visible
in the pinned source or supplied bytes, INFERRED means a supported interpretation,
and UNKNOWN identifies missing evidence. Labels are stable navigation anchors;
line numbers refer to untouched commit aa2dd368d73886bbd666507bd7341001060700d9.
This is a subsystem map, not a full source reconstruction.

## Entry and high-level flow

- VERIFIED — `START.ASM`, `START`, names `Intro.prg`, `Table1.Prg` through `Table4.Prg`, and invokes DOS load/execute (`21h/4B00h`, lines 214–217, 251–252).
- INFERRED — supplied `PINBALL.EXE` is this bootstrap's built counterpart; its MZ signature and filename support this, but we have not compared every instruction.
- VERIFIED — `INTRO.ASM` includes MACROS1.ASM, MACROS3.ASM and FANTASIE.MAC; its startup calls `LOAD_TOGGLAREN` and `UNPKPICS`. It defines option/intro graphics references, an INIT_SOUND driver loader and sound-service calls.
- INFERRED — INTRO is the menu/options and resident-services preparation stage before a table program. Exact process/residency lifetimes are outside PF1.
- VERIFIED — `FANTASIE.ASM`, `FANTASIES` (line 613), initializes settings, SIN, scrolling, ball speeds, matrix, graphics, high scores, masks/flippers, spring, input and sound, then installs interrupts (`INIT_INTS`). `MAIN` (1221) processes command keys and calls the 66h service before returning to MAIN.
- VERIFIED — table choice is compile-time `bana`: 1 includes PLAND.ASM, 2 SDEV.ASM, 3 SHOW.ASM, 4 STONES.ASM (lines 92–103). File headers identify Party Land, Speed Devils, Billion Dollar Game Show, and Stones 'n' Bones respectively.
- VERIFIED — FANTASIE's `VBLANK`/raster-related code calls `KOLLAKULAN` (e.g. 2824–2849); MAIN alone is not all simulation work.
- UNKNOWN — exact timing schedule and hardware-dependent interrupt frequencies have not been reconstructed. A comment on FlipVinkelHast says speed per 1/200 second; that is not sufficient to assert a universal simulation tick rate.

## Rendering and PF1

- VERIFIED — `FANTASIE.ASM` defines `SW=336`, `BPL=SW/4`, `SPLH=33`, low/high display heights 240/350 (lines 146–165). Its VGA setup, split, scrolling, palette, matrix and delta-animation routines implement presentation.
- VERIFIED — `PLAND.ASM` defines `BANH=576`, `BALLH=16` (lines 73–74), spring/flipper positions and table-specific state. It declares STAGE1_1..4 segments at lines 6062–6076, with external st1..st4 data. It also declares mask and flipper graphics segments.
- VERIFIED — `FANTASIE.ASM`, `INIT_GFX` (5787–5876), calls external `UNPKLBM` four times for STAGE1_1..4, x=0 and y=SPLH+[0,144,288,432]. The final stage supplies a 768-byte palette copied to `PALLE` (5849–5854).
- VERIFIED — supplied `TABLE1.PRG` has four well-formed FORM/PBM containers at file offsets 336944, 366176, 399776, 437168. Their BMHD headers say 320×144, eight bits, no mask and ByteRun1 compression; each has a 768-byte CMAP. The complete BODY streams decode to four 46080-byte strips. See `pf1-fixture.json`.
- INFERRED — these four blocks correspond in order to the four STAGE1 references. Segment order, strip dimensions, the assembly's placement and the independently inspected, continuous recognizable Party Land image support the mapping. No linker map is available to prove each symbol's file offset.
- VERIFIED — PF1 renders exactly 320×576 PBM content pixels. It uses the fourth CMAP globally, consistent with INIT_GFX retaining the fourth palette. Only entries 240–242 differ between the first three palettes and the fourth.
- UNKNOWN — external `UNPKLBM` implementation and its exact BP flags, palette/DAC conversion, clipping and handling of SW=336 are absent. PF1 preserves PBM RGB bytes and does not claim exact VGA DAC output or a 336-wide memory-layout reproduction.
- VERIFIED — external `UNPKLBM` is not needed for PF1: the supplied containers decode using their explicit format headers and bounded native ByteRun1 data handling.
- VERIFIED — PF1 does not include the separate 33-line score panel, ball, dynamic lamps, spring/flipper overlays, scrolling, audio or simulation. It displays all four artwork strips at once with original pixels.

## Physics, state and numbers (map only)

- VERIFIED — `BALLCODE.ASM` exports `KollaKulan`, imports FANTASIE state and provides `sc_program`, `sc_krock`, `sc_newdir`, `sc_move` and `draw_flippers`. `sc_program` calls collision, direction, tilt, movement and flipper routines; some branches perform two passes.
- VERIFIED — FANTASIE exports paired `X_POS/X_POS_HI`, `Y_POS/Y_POS_HI`, `SC_X/SC_Y`, velocities `X_HAST/Y_HAST`, gravity, rotation, material/hit flags, scroll state, and mask offsets (lines 199–277). BALLCODE imports the same names.
- VERIFIED — BALLCODE `sc_move` (761–804) adds sign-extended 16-bit velocity to paired 32-bit position words using ADD/ADC, then divides signed positions by 1024 to obtain SC_X/SC_Y. This establishes 1/1024 position units for this routine; signed IDIV truncation and word overflow will matter to a later port.
- UNKNOWN — other numeric conventions (angle units, sine table scale, collision intermediates and complete timing) remain unverified. No physics has been implemented.
- VERIFIED — FANTASIE `flipstruc` (105–137) includes position/bounds, rotation center, speed/angle/frame fields, and pointers to delta graphics and collision tables. It holds task lists, wait lists, player area, TOGGLAREN options, RGB/light buffers and flags.
- VERIFIED — FANTASIE.MAC provides score/light/task macros; `ADDBCD` calls `ADDSCOREBCD`, and score addition uses carry handling. Table modules contain per-table light palette/state/task data.
- INFERRED — score representations are decimal/BCD-oriented based on macro names and arithmetic. Exact packed/unpacked interpretation must be checked before porting scoring.

## Input, sound and supporting dependencies

- VERIFIED — FANTASIE `INIT_KEY` (2239) installs interrupt 9; `KEYINT` reads port 60h. `MAIN` uses IFSCAN macros, scan-code state, SHIFTKEYS, pause and tilt flags. `INIT_MOUSE` is also called during startup. These DOS handlers are references only.
- VERIFIED — INTRO/FANTASIE use interrupt 66h services for audio-related setup/playback; FANTASIE names SOUND.CFG and a historical RESIPLAY path. INTRO `INIT_SOUND` executes a configured sound driver. Table modules name TABLEn.MOD; PLAND names TABLE1.MOD.
- INFERRED — supplied `.SDR` files are hardware-specific driver implementations; their exact protocol and music engine are not fully represented in this source tree. They are never loaded by native PF1.
- VERIFIED — MACROS1.ASM and MACROS3.ASM provide assembler conveniences, data alignment, arithmetic/register helpers and graphics/debug macros; FANTASIE.MAC contains game-specific light, scoring, control/task helpers.
- VERIFIED — source uses absolute historical include paths such as `C:\source\MACROS1.ASM` and INCLUDELIB references (fantasie, pland, billion). Referenced libraries and external fonts/SIN/material/mask/stage/UNPKLBM implementations are not supplied as complete source in this repository.
- UNKNOWN — original build recipe, linker map, generated data library construction and precise match between this source revision and the supplied 1994 executables. The Go port explicitly pins the supplied TABLE1.PRG hash rather than accepting arbitrary installations.

## Bounded dependency cone

- VERIFIED — implemented path: PLAND identity/BANH and stage declarations → FANTASIE INIT_GFX placements/last palette → TABLE1.PRG PBM/BMHD/CMAP/BODY bytes → Go decoder → RGBA framebuffer → SDL2 software window.
- VERIFIED — the native executable uses ordinary Go content-decoding code and a small cgo SDL boundary. There is no DOS/x86 execution, interpreter, instruction decoder, JIT or general ASM translator. Python tools are offline analysis only.
- VERIFIED — parked discoveries: BALLCODE/sc_move — position scale evidence; FANTASIE/DELTANIM — flipper delta graphics; PLAND/READ_SPECIAL_MODE_COUNTER — table mode logic; INTRO/INIT_SOUND — SDR loader. None is in the PF1 implementation scope.
