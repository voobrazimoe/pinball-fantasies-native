#!/usr/bin/env python3
"""68k scan: MOVE.W/L <ea>,(d16,An) and LEA (d16,An) with custom-chip offsets.
Copper lists are exactly MOVE.W #imm,(d16,An) instructions, so this finds both
code and copper data."""
import sys, os, collections

def regname(off):
    if 0x0E0 <= off <= 0x0FC:
        n = (off - 0x0E0)//4 + 1
        return f'BPL{n}PT' + ('(AGA)' if n >= 7 else '')
    if off == 0x100: return 'BPLCON0'
    if off == 0x102: return 'BPLCON1'
    if off == 0x104: return 'BPLCON2'
    if off == 0x106: return 'BPLCON3(AGA)'
    if off == 0x108: return 'BPLCON4(AGA)'
    if off == 0x10C: return 'BPLCON4b(AGA)'
    if off == 0x110: return 'BPL1MOD'
    if off == 0x112: return 'BPL2MOD'
    if 0x180 <= off <= 0x1BE and (off & 1) == 0: return f'COLOR{(off-0x180)//2:02d}'
    if off == 0x08E: return 'DIWSTRT'
    if off == 0x090: return 'DIWSTOP'
    if off == 0x092: return 'COPJMP1'
    if off == 0x094: return 'DDFSTRT'
    if off == 0x096: return 'DDFSTOP'
    if off == 0x09A: return 'INTREQ'
    if off == 0x09C: return 'INTENA'
    if off == 0x09E: return 'ADKCON'
    return None

def scan(name, b, quiet=False):
    named = collections.Counter()
    colimm = collections.Counter()
    allofs = collections.Counter()
    for i in range(len(b) - 4):
        op = (b[i] << 8) | b[i+1]
        off = (b[i+2] << 8) | b[i+3]
        if not (0x080 <= off <= 0x1FF):
            continue
        is_move = ((op & 0xF000) in (0x1000, 0x2000, 0x3000)) and ((op >> 6) & 7) == 5
        is_lea = (op & 0xF1C0) == 0x41C0
        if not (is_move or is_lea):
            continue
        allofs[off] += 1
        nm = regname(off)
        if nm:
            named[nm] += 1
            if nm.startswith('COLOR') and (op & 0x3F) == 0x3C:
                colimm[b[i+2] << 8 | b[i+3]] += 1
    aga = {k: v for k, v in named.items() if 'AGA' in k}
    bpl = sorted({k for k in named if k.startswith('BPL') and k.endswith('PT')},
                 key=lambda s: int(s[3:-2]))
    ncolors = len({k for k in named if k.startswith('COLOR')})
    print(f'== {name}')
    print(f'   bitplane pointers referenced: {bpl if bpl else "none"}')
    print(f'   AGA registers: {aga if aga else "NONE"}')
    print(f'   distinct COLORxx registers written: {ncolors}')
    print(f'   max custom offset seen: 0x{max(allofs) if allofs else 0:03x}')
    return ncolors, bool(aga), bpl

if __name__ == '__main__':
    for f in sys.argv[1:]:
        scan(os.path.basename(f), open(f,'rb').read())
