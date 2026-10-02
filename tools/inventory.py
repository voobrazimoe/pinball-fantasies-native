#!/usr/bin/env python3
"""Read-only inventory of supplied originals and pinned reference sources."""
import hashlib, json, re
from pathlib import Path
GAME_NAMES='ADLIB.SDR GUS.SDR INTERNAL.SDR INTRO.MOD INTRO.PRG MOD2.MOD NOSOUND.SDR PAS16.SDR PINBALL.CFG PINBALL.EXE SB16.SDR SB20.SDR SBLASTER.SDR SBPRO.SDR SETSOUND.EXE SM2.SDR SOUND.CFG TABLE1.HI TABLE1.MOD TABLE1.PRG TABLE2.HI TABLE2.MOD TABLE2.PRG TABLE3.HI TABLE3.MOD TABLE3.PRG TABLE4.HI TABLE4.MOD TABLE4.PRG THING.SDR TIMER.BIN pinfant.txt'.split()
def purpose(p,b):
    if p.suffix=='.SDR':return 'INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename'
    if p.name=='PINBALL.EXE':return 'INFERRED: DOS launcher corresponding to START.ASM (loads Intro.prg and TableN.Prg)'
    if p.name=='SETSOUND.EXE':return 'INFERRED: sound setup utility from filename; implementation absent'
    if p.suffix=='.PRG':return 'VERIFIED: MZ executable container' + (' with four embedded PBM playfield strips; PLAND.ASM names Party Land' if p.name=='TABLE1.PRG' else '; INFERRED: intro/table program by START.ASM filenames')
    if p.suffix=='.MOD':return 'VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN)'
    if p.suffix=='.HI':return 'VERIFIED: pristine factory high scores from INTRO.ASM HI_SCORE_LIST; optional seed only, local mutable files are not runtime/build inputs'
    if p.name=='SOUND.CFG':return 'VERIFIED: sound configuration name referenced by INTRO.ASM/FANTASIE.ASM'
    if p.name=='PINBALL.CFG':return 'VERIFIED: options configuration name referenced by INTRO.ASM'
    if p.name=='TIMER.BIN':return 'UNKNOWN: small binary; no verified consumer in PF1 dependency cone'
    return 'INFERRED: user-facing game documentation from text contents'
def factory_scores(name):
    table = int(name[5])
    label = 'HI_SCORE_LIST' + (str(table) if table != 1 else '')
    source = Path('reference/original-dos-source/INTRO.ASM').read_text(encoding='latin1')
    section = source.split(label+' LABEL BYTE', 1)[1]
    lines = section.splitlines()[1:5]
    data = bytearray()
    for line in lines:
        declaration = line.strip().split(';', 1)[0]
        assert declaration.lower().startswith('db ')
        for literal, number in re.findall(r"'([^']*)'|(\d+)", declaration[3:]):
            data.extend(literal.encode('ascii') if literal else bytes([int(number)]))
    assert len(data) == 64
    return bytes(data)

def record(p, data=None):
    b=p.read_bytes() if data is None else data;return dict(name=p.name,size=len(b),sha256=hashlib.sha256(b).hexdigest())
root=Path('.')
game=[]
for name in GAME_NAMES:
    p=root/name;b=factory_scores(name) if p.suffix=='.HI' else p.read_bytes();r=record(p,b);r['purpose']=purpose(p,b)
    if p.suffix=='.MOD':assert b[1080:1084]==b'M.K.'
    if p.suffix=='.PRG':assert b[:2]==b'MZ'
    game.append(r)
Path('analysis/game-inventory.json').write_text(json.dumps(game,indent=2)+'\n')
source=[record(p) for p in sorted(Path('reference/original-dos-source').iterdir()) if p.is_file()]
Path('analysis/source-inventory.json').write_text(json.dumps(source,indent=2)+'\n')
lines=['# Supplied game data inventory','','Inventory date: 2026-09-30. Original files remain in the installation root; no game assets are embedded or copied into shipping code. Full SHA-256 values and sizes are in `game-inventory.json`.','','| File | Bytes | Probable purpose and evidence |','|---|---:|---|']
for r in game:lines.append(f"| {r['name']} | {r['size']} | {r['purpose']} |")
Path('analysis/game-inventory.md').write_text('\n'.join(lines)+'\n')
