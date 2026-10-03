#!/usr/bin/env python3
"""Compare private DOS matrix captures with source-sync dot-state exports.

No commercial pixels are checked in. Export native states using the opt-in
TestExportScrollOracleStates test, then pass its JSON and the capture folder.
This verifies pixels, not wall-clock durations: capture sampling is not the
DOS source clock. Temporal duration assertions live in the Go source tests.
"""
import argparse
import hashlib
import json
from pathlib import Path
from PIL import Image


def compare(states, captures, groups=('before-interruption', 'after-interruption')):
    known = json.loads(states.read_text())
    results = {}
    for group in groups:
        files = sorted(captures.glob(group + '-*.png'))
        if not files:
            raise ValueError(f'No {group} capture frames')
        counts = dict(blank=0, matched=0, scroll_matched=0, unmatched=0)
        unmatched = []
        candidates = None
        order_failure = None
        for path in files:
            with Image.open(path) as image:
                if image.size != (640, 33):
                    raise ValueError(f'{path}: unexpected matrix crop {image.size}')
                rgb = image.convert('RGB')
                # The native capture tool crops the DOS matrix. One lit dot
                # occupies one sample at x=4*x, y=2+2*y in its doubled raster.
                dots = bytes(int(rgb.getpixel((4*x, 2+2*y))[0] > 81)
                             for y in range(16) for x in range(160))
            if not any(dots):
                counts['blank'] += 1
                continue
            digest = hashlib.sha256(dots).hexdigest()
            traces = known.get(digest)
            if traces is None:
                counts['unmatched'] += 1
                unmatched.append(path.name)
            else:
                counts['matched'] += 1
                if any(trace['op'] == '_SCROLL' for trace in traces):
                    counts['scroll_matched'] += 1
                # Retain a single initial phase through the complete sampled
                # path. Matching isolated frames to different phases or an
                # earlier program cycle is insufficient. Blank illumination
                # frames cannot reveal the retained dot memory, so skip them.
                current = {(t['initial_phase'], t['tick']) for t in traces}
                if candidates is not None:
                    minimum = {}
                    for phase, tick in candidates:
                        minimum[phase] = min(tick, minimum.get(phase, tick))
                    current = {(p, t) for p, t in current
                               if p in minimum and t >= minimum[p]}
                if not current and order_failure is None:
                    order_failure = path.name
                candidates = current
        ordered = dict(first_incompatible_frame=order_failure,
                       compatible_initial_phases=sorted({p for p, _ in candidates or []}),
                       final_sync_range=([min(t for _, t in candidates),
                                          max(t for _, t in candidates)]
                                         if candidates else None))
        results[group] = dict(counts=counts, unmatched=unmatched, ordered=ordered)
    return results


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('states', type=Path)
    parser.add_argument('captures', type=Path)
    parser.add_argument('--output', type=Path)
    parser.add_argument('--groups', nargs='+', default=['before-interruption', 'after-interruption'])
    args = parser.parse_args()
    result = compare(args.states, args.captures, args.groups)
    encoded = json.dumps(result, indent=2) + '\n'
    if args.output:
        args.output.write_text(encoded)
    print(encoded, end='')
    return int(any(v['counts']['unmatched'] or v['ordered']['first_incompatible_frame']
                   for v in result.values()))


if __name__ == '__main__':
    raise SystemExit(main())
