#!/usr/bin/env python3
"""Create an external DOS oracle helper to read the BIOS font required by INTRO.
No helper bytes enter the native executable. Run only in an isolated oracle dir.
Usage: python3 tools/reference_pf8_bios.py .tools/pf8-font-oracle
Then DOSBox-X: mount c DIR; c:; PF8FONT; exit.
Compare PF8FONT.BIN with analysis/pf8-bios-font.json before replacing any asset.
"""
from pathlib import Path
import sys
root=Path(__file__).resolve().parent.parent
out=Path(sys.argv[1]).resolve()
if out==root or root in out.parents and not out.is_relative_to(root/'.tools'):
 raise SystemExit('The helper must be outside the game directory.')
out.mkdir(parents=True,exist_ok=True)
code=bytearray.fromhex('b8003c31c9ba0000cd2189c31eb800f08ed8ba6efab90004b440cd211fb43ecd21b8004ccd21')
code[6:8]=(0x100+len(code)).to_bytes(2,'little')
(out/'PF8FONT.COM').write_bytes(code+b'PF8FONT.BIN\0')
print(out/'PF8FONT.COM')
