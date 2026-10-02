#!/usr/bin/env python3
"""Private, read-only source archaeology index; no execution or source dumps.

The output defaults under the ignored personal audit directory. It records
source locations, player-record declarations and actual save/load references,
including fields declared but never copied. The behavioral report lives in
analysis/pf12.1-hotseat.md.
"""
import hashlib
import json
import re
from pathlib import Path

root = Path(__file__).resolve().parent.parent
source = root / 'reference/original-dos-source'
out = root / '.build-personal/pf121/source-index.json'
pattern = re.compile(r'PLAYER|BALL|GAME.?OVER|NEW.?GAME|START.?GAME|SIFFROR|HIGH.?SCORE|XBALL|XXBALL|ADDPL|F1F8|SAVE|LOAD', re.I)
index = {}
for path in sorted(source.glob('*')):
    if path.suffix.upper() not in ('.ASM', '.MAC'): continue
    raw = path.read_bytes()
    text = raw.decode('latin1')
    lines = text.splitlines()
    entries = []
    routine = ''
    for n, line in enumerate(lines, 1):
        code = line.split(';')[0].strip()
        match = re.match(r'(\w+)(?:\s+PROC\b|:)', code, re.I)
        if match: routine = match[1]
        if pattern.search(code): entries.append({'line': n, 'routine': routine})
    item = {'sha256': hashlib.sha256(raw).hexdigest(), 'paths': entries}
    record = re.search(r'PLAYER_STRUC STRUC(.*?)\n(?:ENDS|PLAYER_STRUC ENDS)', text, re.I | re.S)
    if record:
        fields = re.findall(r'^\s*(P_\w+)\s+D[BW]', record[1], re.I | re.M)
        item['declared_fields'] = fields
        for label, key in [('VARS_2_P_STRUC', 'saved_fields'), ('P_STRUC_2_VARS', 'loaded_fields')]:
            body = re.search(r'^'+label+r' PROC.*?^'+label+r' ENDP', text, re.I | re.S | re.M)
            assert body, (path.name, label)
            references = {f.upper() for f in re.findall(r'\.?(P_\w+)\b', body[0], re.I)}
            item[key] = [f for f in fields if f.upper() in references]
        saved = {f.upper() for f in item['saved_fields']}
        item['declared_not_saved'] = [f for f in fields if f.upper() not in saved]
    index[path.name] = item
assert re.search(r'NO_OF_PLAYERS\s*=\s*8', (source/'FANTASIE.ASM').read_text(encoding='latin1'))
for name in ['PLAND.ASM','SDEV.ASM','SHOW.ASM','STONES.ASM']:
    assert 'P_SIFFRORNA' in {f.upper() for f in index[name]['saved_fields']}
out.parent.mkdir(parents=True, exist_ok=True)
out.write_text(json.dumps(index, indent=2)+'\n')
for name, item in index.items():
    if 'saved_fields' in item:
        print(name, 'saved',len(item['saved_fields']), 'declared not saved', item['declared_not_saved'])
print('Private location:', out.relative_to(root))
