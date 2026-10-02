#!/usr/bin/env python3
"""Prepare disposable external runtime probes, never change root game data.
Usage: pf8_runtime_fixture.py bitmap|ending-probe|ending-real NEW_DIRECTORY
"""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parent.parent
mode, directory = sys.argv[1:]
assert mode in ('bitmap', 'ending-probe', 'ending-real')
out = Path(directory).resolve()
subprocess.run([sys.executable, str(root/'tools/pf8_oracle.py'), str(out)], check=True)
conf = out.parent/(out.name+'.conf')
conf.write_text(conf.read_text().replace('cycles=30000', 'cycles=200000'))
if mode == 'bitmap':
    f = json.loads((root/'analysis/pf8-runtime-validation/bitmap-oracle-fixture.json').read_text())
    data = bytearray((out/'TABLE2.PRG').read_bytes())
    assert hashlib.sha256(data).hexdigest() == f['original_sha256']
    patch = f['patch']
    offset = patch['file_offset']
    before, after = bytes.fromhex(patch['before']), bytes.fromhex(patch['after'])
    assert data[offset:offset+len(before)] == before
    # Source program copied from its original location; no invented commands.
    source = f['gear_program_file_offset']
    assert data[source:source+len(after)] == after
    data[offset:offset+len(after)] = after
    assert hashlib.sha256(data).hexdigest() == f['fixture_sha256']
    (out/'TABLE2.PRG').write_bytes(data)
else:
    # Valid BCD100 avoids INTRO MOVSCORE's original underflow for scores0..9.
    (out/'TABLE2.HI').write_bytes((bytes([0]*9+[1,0,0])+b'TST'+bytes(1))*4)
    if mode == 'ending-real':
        cfg = bytearray((out/'PINBALL.CFG').read_bytes())
        cfg[0] = 0  # source three-ball setting; original PRG unchanged
        (out/'PINBALL.CFG').write_bytes(cfg)
    else:
        f = json.loads((root/'analysis/pf8-runtime-validation/ending-oracle-fixture.json').read_text())
        data = bytearray((out/'TABLE2.PRG').read_bytes())
        for patch in f['patches']:
            offset = patch['file_offset']
            before, after = bytes.fromhex(patch['before']), bytes.fromhex(patch['after'])
            assert data[offset:offset+len(before)] == before
            data[offset:offset+len(after)] = after
        assert hashlib.sha256(data).hexdigest() == f['sha256']
        (out/'TABLE2.PRG').write_bytes(data)
print(conf)
