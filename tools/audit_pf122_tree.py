#!/usr/bin/env python3
"""Read every tracked blob: content triage, never legal clearance.

Optional --original-dir compares nontrivial 4 KiB original windows, including
embedded content in differently named files. Report contains hashes only.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

root = Path(__file__).resolve().parent.parent
p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--original-dir', type=Path)
p.add_argument('--font', type=Path, help='optional private removed-font record for exact-byte scanning')
p.add_argument('--output', type=Path, required=True)
a = p.parse_args()
names = subprocess.check_output(['git', 'ls-files'], cwd=root, text=True).splitlines()
old = json.loads((root/'analysis/public-preflight-inventory.json').read_text())
review = {r['path'] for r in old['files'] if r.get('action') == 'MANUAL LEGAL REVIEW'}
review.add('internal/partyland/timing_data.go')
windows = []
if a.original_dir:
    for original in a.original_dir.iterdir():
        if original.suffix.upper() not in {'.PRG', '.MOD', '.HI', '.CFG', '.EXE', '.COM', '.BIN'}:
            continue
        raw = original.read_bytes()
        for i in range(0, len(raw)-4095, 4096):
            chunk = raw[i:i+4096]
            if len(set(chunk)) > 32:
                windows.append((original.name, i, chunk))
font = a.font.read_bytes() if a.font else None
rows, forbidden, embedded, local_paths, font_matches = [], [], [], [], []
for name in names:
    raw = (root/name).read_bytes()
    if name not in {'analysis/pf12.2-public-preflight.md', 'analysis/pf12.2-tree-audit.json'}:
        rows.append({'path': name, 'bytes': len(raw), 'sha256': hashlib.sha256(raw).hexdigest()})
    if (name != 'go.mod' and Path(name).suffix.lower() in {'.prg','.mod','.hi','.cfg','.exe','.com','.bin','.appimage','.wav','.pcm'}) or name.startswith(('reference/','release/','internal/platform/.personal-assets/')) or name == 'internal/platform/personal_payload_windows.go':
        forbidden.append(name)
    if font and font in raw:
        font_matches.append(name)
    for original, offset, chunk in windows:
        if chunk in raw:
            embedded.append({'path':name,'original':original,'offset':offset,'bytes':len(chunk)})
    if re.search(rb'/home/|[CD]:\\|release/personal|\.personal-assets|reference/original-dos-source', raw):
        local_paths.append(name)
result = {'scope':'all tracked working-tree bytes; containing candidate commit binds report',
          'files':rows, 'self_bound_reports':['analysis/pf12.2-public-preflight.md','analysis/pf12.2-tree-audit.json'], 'forbidden_payload_paths':forbidden,
          'original_4k_matches':embedded, 'removed_font_matches':font_matches, 'original_nontrivial_windows_checked':len(windows),
          'manual_copyright_review':sorted(review.intersection(names)),
          'local_path_mentions':local_paths,
          'limitations':'4 KiB comparison is a raw-payload check, not proof of independent authorship. Source-derived programs, quotations and small tables need manual review.'}
a.output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in {'files','local_path_mentions','manual_copyright_review'}},indent=2))
if forbidden or embedded or font_matches:
    raise SystemExit(1)
