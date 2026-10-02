# PF2: Party Land's first visible ball frame

PF2 freezes **PLAND/SETBALL immediately after SETBALLPOS**, before any
KollaKulan call can move the ball. This is the first visually observable ball
state requested by the user, not the earlier hidden internal NEW_BALL state.
The composition is PF1 artwork + original viewport + ball + foreground mask.
No simulation or task scheduler is implemented.

## Source-derived state

VERIFIED — PLAND.ASM lines 64–67 define BALLH=16, BANH=576,
STARTX=310−8=302 and STARTY=543−8=535. NEW_BALL (3604–3669) first
sets HOLDSTILL=true and SETBALLPOS STARTX−5−15, STARTY−5: **(282,530)**.
SETBALL waits 80 syncs and then sets SETBALLPOS STARTX−5, STARTY−5:
**(297,530)**, BALLHIGH=false. The wait and subsequent velocity assignments
are reference evidence only; PF2 starts at this snapshot and never advances.
The pinned executable also contains these coordinate assignments at file
0xef7 and 0xfff respectively, to SC_X/SC_Y offsets 0x2ee2/0x2ee4.

VERIFIED — FANTASIE.MAC/SETBALLPOS (342–356) stores those arguments directly
in SC_X/SC_Y. FANTASIE.ASM/PUTTHEBALL (2673–2740) passes SC_X in DX,
SC_Y+SPLH in SI to PUTBALL. Thus these coordinates are the sprite's upper-left
**draw origin**, not its center. No additional centering/hotspot subtraction
is appropriate. Ball bounds are nominally 16×16; the actual visible art spans
15×15 with transparent edges and a transparent last row.

VERIFIED — PINBALL.CFG is six bytes `01 01 02 01 01 00`.
INTRO.ASM/TOGGLAR_STRUCEN and FANTASIE.ASM/TOGGLAR_STRUCEN place
S_RESOLUTION at byte 4. FANTASIES (630–638) selects HI_RES for nonzero.
PF2 pins that supplied installation's high-resolution geometry; it does not
implement an options reader, menus or alternate display modes.

VERIFIED — FANTASIE.ASM defines SW=336, BPL=84, SPLH=33, SH_HI=350,
SH_LO=240. Startup SCREENFORCE and SETSCREENSTART (2539–2670) establish
a bottom-aligned start: BANH−(SH_HI−SPLH)=**259**. The unforced calculation
at SETBALL also clamps to that same bottom origin. The raster-memory start
is (259+33)×84=24528; FANTASIE.MAC/SETRASTERPOS (428–440) introduces
no horizontal pan. The visible gameplay area is **x=0..319, y=259..575**,
320×317, with the ball at **(297,271)** in viewport coordinates.
Low resolution would give y=369 and 207 gameplay rows; that mode is outside
this slice. The separate 33-row score panel is excluded from PF2.

## Graphics and foreground dependency cone

VERIFIED — PUTBALL and DELBALL are external library declarations at
FANTASIE.ASM 608; their implementation source is absent. The actual original
TABLE1.PRG therefore supplies the remaining graphics/ordering evidence.
Only this bounded routine was statically inspected; it was never executed.
PUTTHEBALL's near call at file 0x44b0 targets **0x98a0**. The MZ header
is 512 bytes and the code segment starts at module paragraph 0x10, giving
file code base 0x300. PUTBALL occupies file **[0x98a0,0xa550)**,
CS offsets [0x95a0,0xa250), SHA-256
`8bf17f7381a9355fa8002e66f2278e80ed407ce5904af586e44ebe31d09de67c`.

VERIFIED — PUTBALL selects VGA planes relative to x modulo four, computes
SI=(SC_Y+33)×84+floor(SC_X/4), and contains unrolled palette-index writes.
It has no separate ball PBM container. Four plane-relative write groups start
at file 0x9901, 0x9c20, 0x9f7f, 0xa29d; their relative planes are 0,1,2,3.
For each literal store's byte displacement d, the graphic coordinate is
x=4×(d mod 84)+relative_plane, y=floor(d/84).
`pf2-ball-locations.json` records the inspected literal-byte file locations
in 16×16 row order, zero for absent pixels. `ball_locations.go` is the same
fixed content-location map used by the native reader. No machine instructions
are interpreted by the Go program; it reads only those 177 palette bytes
from the SHA-pinned original file. It contains no generated ASM-to-Go routines.

VERIFIED — each original palette write is guarded by a test of a foreground
bit in ES:HIDDEN1. A set bit preserves the existing artwork. BALLHIGH=false
selects HID1; the alternate plane is ignored. PUTTHEBALL's linked ES segment
is 0x2f93 and BP is 0xfb96 (−1130). In source, BP=HID1−33×42−576×2,
which gives HID1 offset 1408. The raw original HID1 starts at
512+0x2f93×16+1408=**0x300b0**, 576 rows of 40 bytes, MSB first.
FANTASIE.ASM/BREDDA_MASK (6411–6448) copies these rows to HIDDA at offset
256 with a 42-byte stride and two padding bytes. PF2 reads the 320-bit rows
directly and retains equivalent pixel tests for the visible area.
Mask SHA-256: `472090604b9227c27467757ffc7679faa61d1915305b85d29003edb4b13788c0`.

VERIFIED — NEW_BALL's (282,530) masks all **177** sprite pixels. SETBALL's
(297,530) exposes **152**, with **25** suppressed by foreground. The mask
prevents the ball painting over the lane's foreground edge. Transparent sprite
pixels also preserve the playfield. First-frame background restoration is not
needed; each native render starts from the untouched PF1 framebuffer.

## Validation and limits

`tools/reference_pf2.py` independently composes RGBA bytes in Python using the
PF1 extractor, fixed original literal-byte locations, raw foreground rows and
last-strip palette. It emits `pf2-fixture.json`; fixture files contain content
locations, metadata and hashes, not copied artwork. Run from the project root:

```sh
python3 tools/reference_pf2.py
./tools/go.sh test -p=1 ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -png /tmp/pf2-native.png
./bin/pinballfantasies -pf1 -png /tmp/pf1-regression.png
./bin/pinballfantasies -duration 3s
```

VERIFIED — the 320×317 RGBA hash is
`b921c9933cf6f49b605d78e28257d77c6c5f420ca84afc03a8dac2edda81f7b8`.
The 256-byte ball hash is
`e1286e3cff190e96e8d063dae22155ce387ab85e6d3925b7623c8d84496f3cc4`.
Tests pin viewport/ball origins, sprite shape/hash, foreground hash, individual
pixel ordering, visible/hidden counts, deterministic rerender, file rejection
and unchanged PF1 output. Existing PF0 inventory and all PF1 tests still pass.
PF1 decoder, renderer, fixture and independent extraction tool were not edited.
The native PF2 image was visually inspected and the SDL X11 window opened
successfully in a three-second smoke test. Sandbox-only X11 was unavailable;
the desktop-access retry succeeded. VCS stamping is disabled for builds because
this execution environment's .git placeholder is not a usable Git repository.

The framebuffer deliberately retains PF1 CMAP RGB bytes. Exact VGA DAC/light
palette behavior remains outside this representation, as in PF1. There is no
score panel, lamp animation, additional spring/flipper overlay or table-rule
state; the requested composition uses the decoded playfield as its background.

Parked and ignored: BALLCODE.ASM/sc_program (98–142), sc_krock (207 onward),
sc_newdir, sc_move (761–804), tilt0 and draw_flippers. These combine physics,
collision, motion and flipper work and are not needed to draw this snapshot.
PLAND/NEW_BALL's sounds, lights, reset rules and delayed task sequencing are
not ported. Only its two coordinate assignments and selected static state were
used. No ball velocity, acceleration, collision, gameplay keyboard, scoring,
bumpers, rules, audio, menus, other table, DOSBox, x86 execution, translator
or generic engine has been added. PF2 ends here; PF3 is not started.
