#!/usr/bin/env python3
"""Private, read-only comparison of possessed DOS layouts. Never generates profiles.
Output contains hashes/offsets/results, never file contents. Keep output ignored.
Candidate byte matches are investigation leads, not proof of record identity.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import struct
import subprocess

ROOT = Path(__file__).resolve().parent.parent
NAMES = ['INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD'] + [
    f'TABLE{i}.{ext}' for i in range(1, 5) for ext in ('PRG', 'MOD')]


def inventory(root):
    return [{'name': p.name, 'size': p.stat().st_size,
             'sha256': hashlib.sha256(p.read_bytes()).hexdigest()}
            for p in sorted(root.iterdir()) if p.is_file()]


def compare_regions(canonical, alternate, profile, shifts):
    out = []
    for r in profile['regions']:
        at, size = r['offset'], r['size']
        raw = canonical[at:at+size]
        matches = [at+s for s in shifts if at+s >= 0 and
                   alternate[at+s:at+s+size] == raw]
        row = {'canonical_offset': at, 'size': size, 'purpose': r['purpose'],
               'fingerprinted': bool(r.get('sha256')), 'kind': r.get('kind'),
               'reviewed_shift_candidates': matches}
        if not matches:
            # Private search only: unrelated short records can match by accident.
            row['unproven_search_candidate'] = alternate.find(raw)
        out.append(row)
    return out


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--canonical', required=True, type=Path)
    p.add_argument('--alternate', required=True, type=Path)
    p.add_argument('--output', required=True, type=Path)
    p.add_argument('--go', default=str(ROOT/'.tools/go/bin/go'))
    a = p.parse_args()
    output = a.output.resolve()
    try:
        relative = output.relative_to(ROOT)
    except ValueError:
        p.error('output must be in an ignored private directory within the repository')
    if subprocess.run(['git', 'check-ignore', '-q', str(relative)], cwd=ROOT).returncode:
        p.error('output must be Git-ignored')
    profiles = json.loads((ROOT/'internal/datalayout/profiles.json').read_text())
    report = {'canonical_inventory': inventory(a.canonical),
              'alternate_inventory': inventory(a.alternate), 'files': {}}
    env = dict(os.environ, GOCACHE=str(ROOT/'.cache/go-build'), GOMODCACHE=str(ROOT/'.cache/go-mod'))
    for key, source in [('canonical', a.canonical), ('alternate', a.alternate)]:
        report[key+'_validation'] = json.loads(subprocess.check_output(
            [a.go, 'run', './cmd/pflayoutaudit', '--data', str(source.resolve())], cwd=ROOT, env=env))
    for name in NAMES:
        b, d = (a.canonical/name).read_bytes(), (a.alternate/name).read_bytes()
        r = {'identical': b == d, 'canonical_size': len(b), 'alternate_size': len(d)}
        if name in profiles:
            profile = profiles[name]
            shifts = {'INTRO.PRG': [-641], 'TABLE1.PRG': [16,14],
                      'TABLE2.PRG': [16,17]}.get(name, [0])
            r['regions'] = compare_regions(b,d,profile,shifts)
            r['pictures'] = []
            for pic in profile['pictures']:
                at = pic['offset']; size = 8+struct.unpack_from('>I',b,at+4)[0]
                r['pictures'].append(dict(pic, size=size, identical_form_candidate=d.find(b[at:at+size])))
        else:
            differences = [i for i,(x,y) in enumerate(zip(b,d)) if x != y]
            r['differing_byte_count'] = len(differences)
            if differences: r['difference_extent'] = [min(differences), max(differences)+1]
        report['files'][name] = r
    # S_EMPTY identity is anchored by adjacent jingle records, not a global match.
    b, d = (a.canonical/'TABLE1.PRG').read_bytes(), (a.alternate/'TABLE1.PRG').read_bytes()
    at = 0x1a9ac; shift = 16
    report['partyland_s_empty'] = {
        'canonical_offset': at, 'alternate_offset': at+shift,
        'left_neighbor_matches': b[at-3:at] == d[at+shift-3:at+shift],
        's_main_neighbor_matches': b[at+6:at+9] == d[at+shift+6:at+shift+9],
        'changed_fields': [field for i,field in enumerate(['position','repeat','priority'])
                           if b[at+i] != d[at+shift+i]]}
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(report, indent=2)+'\n')
    print('Private metadata audit saved; no payload exported.')


if __name__ == '__main__':
    main()
