#!/usr/bin/env python3
"""Rebuild both personal artifacts with changed and absent installation scores.

Run locally with original runtime assets present. Restores installation score
bytes/modes even on failure; all payloads and evidence remain ignored.
"""
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tempfile
import zipfile
from personal_assets import ROOT, PERSONAL_INPUTS

ARTIFACTS = [ROOT/'release/personal/windows/pinballfantasies.exe',
             ROOT/'release/personal/linux/PinballFantasies-x86_64.AppImage']

def build(label):
    log = ROOT/'.build-personal'/f'cleanup-build-{label}.log'
    with log.open('w') as output:
        subprocess.run(['python3', str(ROOT/'tools/build_personal_release.py'), str(ROOT)],
                       cwd=ROOT, stdout=output, stderr=subprocess.STDOUT, check=True)
    return {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in ARTIFACTS}

def verify_payloads():
    exe = ARTIFACTS[0].read_bytes()
    start = exe.index(b'PK\x03\x04')
    end = exe.index(b'PK\x05\x06', start) + 22
    with zipfile.ZipFile(io.BytesIO(exe[start:end])) as archive:
        assert archive.namelist() == list(PERSONAL_INPUTS), archive.namelist()
        for name in archive.namelist():
            assert archive.read(name) == (ROOT/name).read_bytes(), name
    with tempfile.TemporaryDirectory(prefix='personal-payload-') as folder:
        subprocess.run([str(ARTIFACTS[1]), '--appimage-extract'], cwd=folder,
                       stdout=subprocess.DEVNULL, check=True)
        data = Path(folder)/'squashfs-root/usr/share/pinballfantasies'
        assert sorted(p.name for p in data.iterdir()) == sorted(PERSONAL_INPUTS)
        for name in PERSONAL_INPUTS:
            assert (data/name).read_bytes() == (ROOT/name).read_bytes(), name
        manifest = json.loads((Path(folder)/'squashfs-root/BUILD-MANIFEST.json').read_text())
        assert [r['name'] for r in manifest['original_inputs']] == list(PERSONAL_INPUTS)
    for manifest in ['build-manifest.json', 'linux-build-manifest.json']:
        records = json.loads((ROOT/'.build-personal'/manifest).read_text())['original_inputs']
        assert [r['name'] for r in records] == list(PERSONAL_INPUTS)

if __name__ == '__main__':
    scores = [ROOT/f'TABLE{table}.HI' for table in range(1, 5)]
    backup = {p: (p.read_bytes(), p.stat().st_mode) if p.exists() else None for p in scores}
    try:
        baseline = build('baseline')
        verify_payloads()
        for p in scores:
            p.write_bytes(b'replaced malformed installation scores: '+p.name.encode())
        replaced = build('replaced-hi')
        assert replaced == baseline, (baseline, replaced)
        for p in scores:
            p.unlink()
        removed = build('removed-hi')
        assert removed == baseline, (baseline, removed)
        verify_payloads()
        result = {'baseline': baseline, 'all_four_replaced': replaced, 'all_four_removed': removed,
                  'byte_identical': True, 'payload_input_count': 12, 'payload_hi_count': 0}
        (ROOT/'.build-personal/cleanup-reproducibility.json').write_text(json.dumps(result, indent=2)+'\n')
        print(json.dumps(result, indent=2))
    finally:
        for p, original in backup.items():
            if original is None:
                p.unlink(missing_ok=True)
            else:
                p.write_bytes(original[0])
                p.chmod(original[1])
