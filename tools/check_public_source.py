#!/usr/bin/env python3
"""Check address-only record layouts and typed native operands without assets."""
import json
from pathlib import Path
import re

root = Path(__file__).resolve().parent.parent
source = (root/'internal/presentation/content.go').read_text()
content = json.loads(re.search(r'const contentJSON = `(.*)`', source, re.S).group(1))
assert set(content) == {'1', '2', '3', '4'}
for table, data in content.items():
    for field in ('scorefont', 'scrollfont', 'animations', 'lampflash'):
        assert field not in data, f'original records embedded: {field}'
    # Explicit small unresolved retail/source exceptions; not matrix prose.
    exceptions = {'FLIPDATAFIL_L','FLIPDATAFIL_R','FLIPDATAFIL_L2','FLIPGFXFIL','FLIPSTRUCFIL','HI_SCORE_FILE'}
    if table == '2': exceptions.add('S_GAMEOVER2')
    assert set(data['texts']) == exceptions
    assert all(len(v) <= 13 for v in data['texts'].values())
    for ref in data['text_refs'].values():
        assert len(ref) == 2 and all(isinstance(v,int) and v >= 0 for v in ref)
    for ref in data['animation_refs'].values():
        assert len(ref) == 4 and ref[0] >= 2 and ref[1] > 0 and ref[2] >= 0 and ref[3] in (0,1)
    assert len(data['lamp_flash_ref']) == 2 and all(v > 0 for v in data['lamp_flash_ref'])
    for command in data['commands'] + data['attract']:
        assert isinstance(command['op'], str)
        for index, value in command.get('nums', {}).items():
            assert 0 <= int(index) < len(command['args'])
            assert isinstance(value, int)
assert not (root/'internal/assets/intro_bios_font.bin').exists()
assert not (root/'analysis/pf8-content.json').exists()
print('PASS: PRG record layouts; typed operands; no embedded matrix prose/animation/flash/font records')

for package,filename in [('partyland','timing_data.go'),('speeddevils','content.go'),('gameshow','content.go'),('stones','content.go')]:
    source=(root/'internal'/package/filename).read_text()
    data=json.loads(re.search(r'const \w+ = `(.*)`',source,re.S).group(1))
    assert all(v == {} for v in data['jingles'].values()), 'embedded jingle records'
    if package in ('partyland','speeddevils'):
        assert all(v == {} for v in data['animations'].values()), 'embedded animation timings'
        assert all(v == 0 for v in data['scrolls'].values()), 'embedded source text lengths'
        assert 'attract' not in data, 'duplicate lamp flash records'
        if package == 'speeddevils':
            assert len(data['lamps']) == 67
            assert all(set(v) == {'offset'} for v in data['lamps'].values()), 'embedded lamp headers'
    else:
        assert all(set(v)=={'ref'} for v in data['lamps'].values()), 'embedded DAC packets'
        assert all('opened' not in g and 'closed' not in g for g in data['gates']), 'embedded gate masks'
        if package == 'stones':
            assert 'areas' not in data, 'embedded area records'
            assert len(data['area_refs']) == 4
            assert sum(v[1] for v in data['area_refs'].values()) == 44
            assert all(len(v) == 2 and v[0] >= 0 and v[1] >= 0 for v in data['area_refs'].values())
print('PASS: native jingle/animation registration and PRG DAC/gate layouts')
