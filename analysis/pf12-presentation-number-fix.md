# PF12 full-table nudge and numeric formatting

Implementation covers Windows and Linux through shared table/presentation code.
Release blockers remain open for the owner's real-machine retest. Previously
Alt+Enter/audio code is unchanged, but Issue #1 remains open pending owner retest
on real Windows confirming repeated audibly clean transitions.

## Full-table source nudge

`BALLCODE.ASM:TILT0` (825–850) updates SCREENPOS through SCREENHAST, clamps
SCREENPOS to 2048, and derives SCREENPOSY with `>>9`. A newly triggered nudge
therefore has the existing ScreenPosition=1800, ScreenSpeed=600, ScreenOffset=3;
its natural offset range is 0–4. Nothing in this simulation or the table
DANGER/TILT/flipper-inhibit logic was changed.

`FANTASIE.ASM:SETSCREENSTART` (2662 onward) adds SCREENPOSY to the **source**
raster address. Positive offset moves the artwork upward by that exact pixel
count. Full-table composition now draws source row ScreenOffset at destination
row 33, clipping to the playfield rectangle and filling vacated bottom rows
opaque black. The 320×33 matrix is drawn at the original location. There is no
additional animation clock or arbitrary shake amplitude.

`PUTTHEBALL` (2715 onward) adds SCREENPOSY to ball Y unless HOLDSTILL. The
full-table renderer now retains that original sprite offset: after composition,
a free ball stays at its original screen position, while a held ball moves with
the table. This is a rendering-only change in `FramePalette`; collision/ball
integration remains unchanged. Scrolling-mode composition is unchanged.

Regression coverage compares all four tables in OFF/HARD/SOFT using identical
inputs, including the physical state, ball, DANGER counter, TILT and AllowFlip.
It exercises first nudge, DANGER and third-input TILT. Pixel tests verify source
row identity for offsets 1–4, an unchanged matrix, opaque vacated rows, and
byte-identical composition at zero. Existing full-table exact-row tests remain.

## Numeric source semantics and findings

The reproducible [numeric inventory](matrix-numeric-audit.md) lists every
numeric/custom source site, source line, original formatter and mutable text
writer across PLAND/SDEV/SHOW/STONES (36/26/26/31 sites respectively). It also
identifies legacy player/ball/SHOWINFO text paths replaced by the existing native
single-player panel, rather than claiming those buffers are mutated natively.
`python3 tools/matrix_numeric_audit.py --check` verifies the committed reports.

Common `_PRINT5_NUMBER`, `_PRINT8_NUMBER`, `_PRINT8_NUMBER_CENT`,
`_PRINT11_NUMBER`, `_PRINT13_NUMBER` and `_PRINT13_NUMBER_CENT` dispatch through
FANTASIE print tasks (2993–3089) to linked PRINT_NUMBER. The linked routine body
is not present in the supplied ASM; TABLE1.PRG disassembly at 6f09–6f5c confirms
its leading-zero skip and inclusive first-nonzero REPE SCASB centering count.
All-zero BCD prints no glyphs. Native Visit already suppressed these zeros;
there is no special case for CYCLONS. Actual CYCLONECOUNTERBCD commands on each
table are now covered for 0, 2, 10 and 100.

The table-local `Display.Number` path used by bonus/countdown previously padded
with spaces and forced a visible zero. It now shares the exact PRINT_NUMBER
helper. `_NUMBER` and player/high-score panels continue to use CODE2 SCORE,
including its visible zero and comma grouping. SEC_ASC retains its separate
rule: blank only the zero tens digit, with visible ones digit even at zero.
Centered numeric placement retains the source's inclusive SCASB offset.

Stones scream text bypassed Visit: native `%03d` wrote all three MILES/JUMP
buffers after the effect, producing 002/010 and modifying the inactive branch.
`STONES.ASM:Put_In_Text` (5201–5226) writes three encoded digits then replaces
the entire leading zero run with encoded spaces (`*`). Native now does the same:
2 -> **2, 10 -> *10, 100 -> 100, 0 -> ***. Only the source-selected jump buffer
is written, before EFFECT; the matched-award branches retain their original
writer order. Mutable buffers remain private to each Display/session.

Speed Devils Put_In_Text and Put_In_TextTopZero remain separate: miles blank
the full leading run; jump/offroad blank only hundreds. Gameshow SKILLTEXT also
blanks only hundreds and shifts its destination for values >=100. Ordinary
source text is never globally trimmed. Existing table-specific writer tests
and the added Stones tests preserve these differences.

Party and Speed numeric readers also now recognize the source name JACKVALUE
for the live jackpot, retaining JACKPOT as an alias. Reachability's generated
native literal inventory was refreshed for those two aliases; checker rules
and original source data are unchanged.

## Validation and remaining acceptance

- Full unfiltered Go suite: only the previously documented local TABLE1.HI
  inventory mismatch. A full second run excluding precisely that subtest passes;
  the pinned inventory and original files were not modified.
- Numeric regressions: every font/table, 2/10/100, full 12-digit BCD, centered
  output, all-zero semantics, actual bonus counter commands, CODE2 SCORE,
  countdown, ordinary source text and table-specific mutable writers.
- Matrix reachability: zero missing targets. Operand audit: 1367 operands,
  zero unresolved; 1106 source texts, zero byte mismatches or missing writers.
  Operand schema tests and the numeric audit pass.
- Party deterministic oracle: 1200 ticks = 2,300,000 / ball 2, frame SHA256
  `aec01b3a07e1a5ba10b3c635777a6742f4abd41c09899903ea532b98d8522913`.
  Host frame/PCM cadence remains deterministic.
- Windows packages cross-compile; Windows/Wine native focus/input regressions
  and oracle supplement the Linux suite. New numeric regressions run as Windows
  binaries for presentation and all four tables; all-table nudge physics/pixels
  also pass in the Windows frontend binary. Wine cannot close the real-machine
  held-input blocker. The rebuilt Linux AppImage passes the physical XTest
  focus-loss/manual-resume journey under Xvfb.

Fresh public Windows and personal Windows/Linux artifacts are built from this
change. Local logs and commercial-data artifacts remain ignored. Real retest
must confirm flipper release, focus pause, full-table nudge and matrix formatting;
no release blocker is closed by these automated results.

Personal artifact SHA256:

- Windows EXE: `f39ab535e617945b35f168d80d1ce58d8908653862bcb6d1ac877a7b79b394c3`
- Linux AppImage: `0bcd22c99b1250bd1eb8a8126bb64dfaee74cf59196beff095fff7d3d2828141`

All 12 personal build inputs retain their manifest hashes after validation.
