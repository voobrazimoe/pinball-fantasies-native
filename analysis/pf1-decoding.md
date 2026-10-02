# PF1 Party Land decoding and validation

VERIFIED: this slice reads the original `TABLE1.PRG` in place, SHA-256
`4d7a69e7dc95260ad2541c6981a11ab842e2f1f20e45447e5613688b86e38414`.
It never loads the MZ program, applies relocations or executes its instructions.

VERIFIED: FANTASIE.ASM/INIT_GFX supplies four stage placements at 144-line
intervals. PLAND.ASM/BANH supplies a 576-line field. The original file contains:

| FORM offset | Total FORM bytes | BODY bytes | Decoded pixels |
|---:|---:|---:|---:|
| 336944 | 29220 | 25803 | 320×144 |
| 366176 | 33594 | 29875 | 320×144 |
| 399776 | 37378 | 33645 | 320×144 |
| 437168 | 28544 | 24977 | 320×144 |

VERIFIED: IFF sizes/header words are big-endian, chunk payloads are even-padded,
and all four headers specify chunky PBM, eight bits, no mask, compression=1.
ByteRun1 signed controls copy n+1 literals for n≥0, repeat 1−n times for
−127≤n≤−1, and do nothing for −128. Each row expands to 320 bytes; all BODY
bytes are consumed exactly. Unneeded DPPS/CRNG/TINY metadata is skipped.

INFERRED: increasing block order maps to STAGE1_1..4. The complete artwork is
continuous and visually recognizable as Party Land. This is supported by source
placement and segment order, rather than by a linker map (not present).

VERIFIED: INIT_GFX keeps the last stage's palette in PALLE. The first three CMAPs
have identical hashes; the fourth changes only entries 240–242. The native field
uses that fourth palette for every index. RGB values are the original CMAP bytes;
alpha is 255. No replacement art, scaling of decoded indices or palette painting
is used. Exact later VGA/light initialization behavior remains UNKNOWN.

VERIFIED: `tools/reference_extract.py` is a separate Python implementation with
fixed inspected offsets and independent IFF and signed-run parsing. It generated
`pf1-fixture.json` before the Go implementation existed. Fixtures contain only
metadata and SHA-256 values, with no copied artwork. Go tests compare the file,
each strip's decoded indices/palette, combined dimensions/indices and RGBA output
to these independently derived values; repeated decode/render is also compared.
Hand-authored tests exercise control meanings, row boundaries, truncation,
trailing data and IFF size/offset failures. All 33 original file hashes are checked.

VERIFIED: combined index SHA-256 is
`019eceddf83fc318f193330418064267c5625ad1a630f7b44e29e4482842c8af`.
RGBA framebuffer SHA-256 is
`2d90631a3348512de700102621e1da16c42eb844dd1eb3ea3382c7fa12775753`.
The 184320 indices expand to 737280 RGBA bytes. A PNG export is available for
headless validation; PNG compression bytes are not ground truth.

VERIFIED: the executable opened an Ubuntu SDL2 window with the X11 driver using
`-duration 3s`, then exited successfully. Its independently checked Go PNG was
visually inspected. The window shows the complete static 320×576 playfield,
uses a software texture renderer, redraws at 20 Hz for resize/expose handling,
and exits on Esc/window close. That redraw interval is host UI behavior, not
an emulated original game tick.

UNKNOWN: original UNPKLBM's exact DAC reduction/flags; reason for internal SW=336
versus PBM content width=320; dynamic lamp initialization and overlay pixels.
These do not block the bounded static-artwork milestone. No physics, flippers,
other tables, music or original-game oracle execution was added.
