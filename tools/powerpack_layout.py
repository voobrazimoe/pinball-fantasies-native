#!/usr/bin/env python3
"""Export reviewed B address descriptors from private audit evidence, never payload.
Run explicitly with --evidence final-comparison.json --data <possessed B>.
All mapped bytes must match possessed A except the proved S_EMPTY priority.
No discovery or heuristic search is performed here or by the production loader.
"""
import argparse
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

def generate(evidence, canonical, data):
    audit = json.loads(evidence.read_text())
    b = next(s for s in audit if s['fingerprint'] == 'a6074cece2b37bd139381c51493ceea2670f60b776f0145b0bd55b53ba5a78e1')
    names = ['INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD'] + [f'TABLE{i}.{ext}' for i in range(1,5) for ext in ('PRG','MOD')]
    identity = ''.join(name+'\0'+hashlib.sha256((data/name).read_bytes()).hexdigest()+'\n' for name in names)
    assert hashlib.sha256(identity.encode()).hexdigest() == b['fingerprint'], 'not the audited B installation'
    profiles = json.loads((ROOT / 'internal/datalayout/profiles.json').read_text())
    inventory = {r['name']: r for r in json.loads((ROOT/'analysis/game-inventory.json').read_text())}
    for name in profiles:
        assert hashlib.sha256((canonical/name).read_bytes()).hexdigest() == inventory[name]['sha256'], name
    out = {}
    for name in ('INTRO.PRG', 'TABLE1.PRG', 'TABLE2.PRG'):
        f = b['files'][name]
        src, ref = (data/name).read_bytes(), (canonical/name).read_bytes()
        p = profiles[name]
        regions = {(r['canonical_offset'], r['size'], r['purpose']): r for r in f['regions']}
        translated = []
        spans = []
        for r in p['regions']:
            e = regions[(r['offset'], r['size'], r['purpose'])]
            assert e['classification'] in ('payload-equivalent at relocated offset', 'semantic field difference')
            at, dest, size = e['chosen_offset'], r['offset'], r['size']
            observed, expected = src[at:at+size], ref[dest:dest+size]
            assert len(observed) == size
            if e['classification'] == 'semantic field difference':
                assert name == 'TABLE1.PRG' and r['purpose'] == 'matrix record S_EMPTY'
                assert expected == bytes((62, 0, 1)) and observed == bytes((62, 0, 0))
            else:
                assert observed == expected, (name, r)
            q = dict(r, offset=at)
            if 'sha256' in q:
                q['sha256'] = hashlib.sha256(observed).hexdigest()
            translated.append(q)
            spans.append((dest, size, at-dest))
        pictures = []
        for pic, e in zip(p['pictures'], f['pictures']):
            assert pic['offset'] == e['offset']
            dest, at, size = e['offset'], e['match_offset'], e['size']
            assert src[at:at+size] == ref[dest:dest+size]
            pictures.append(dict(pic, offset=at))
            spans.append((dest, size, at-dest))
        # Only merge overlapping/adjacent intervals with identical translation.
        merged = []
        for delta in sorted(set(s[2] for s in spans)):
            group = []
            for at, size, shift in sorted(spans):
                if shift != delta:
                    continue
                if group and at <= group[-1][1]:
                    group[-1][1] = max(group[-1][1], at+size)
                else:
                    group.append([at, at+size])
            merged += [dict(destination=a, source=a+delta, size=z-a) for a,z in group]
        merged.sort(key=lambda r: r['destination'])
        for left, right in zip(merged, merged[1:]):
            assert left['destination']+left['size'] <= right['destination'], (name, left, right)
        out[name] = dict(decoded_size=f['canonical_size'], copies=merged)
        if name == 'TABLE1.PRG':
            out[name]['jingles'] = [dict(role='S_EMPTY', destination=0x1a9ac, source=0x1a9bc,
                                        position=62, repeat=0, priority=0)]
    return out

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--evidence', type=Path, required=True)
    parser.add_argument('--canonical', type=Path, required=True)
    parser.add_argument('--data', type=Path, required=True)
    args = parser.parse_args()
    result = generate(args.evidence, args.canonical, args.data)
    (ROOT/'internal/datalayout/powerpack.json').write_text(json.dumps(result, indent=2)+'\n')
