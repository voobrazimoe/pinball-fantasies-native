#!/usr/bin/env python3
"""Inspect every candidate file; conservative triage is never legal approval."""
import collections
import hashlib
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parent.parent
BASELINE = 'dc2978b06fbda9be38997b15fecfed153a0f4e74'
DUMPS = {'internal/presentation/content.go', 'internal/partyland/timing_data.go',
         'internal/speeddevils/content.go', 'internal/gameshow/content.go', 'internal/stones/content.go'}
REWRITTEN = {'analysis/provenance.md', 'analysis/matrix-numeric-audit.md',
             'analysis/public-preflight.md', 'docs/runtime-data.md'}


def main():
    baseline = json.loads((ROOT/'analysis/public-preflight-inventory.json').read_text())
    old = {f['path']: f for f in baseline['files']}
    names = set(subprocess.check_output(['git', 'ls-files'], cwd=ROOT, text=True).splitlines())
    names |= set(subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', BASELINE], cwd=ROOT, text=True).splitlines())
    names |= set(subprocess.check_output(['git', 'ls-files', '--others', '--exclude-standard'], cwd=ROOT, text=True).splitlines())
    self_files = {'analysis/public-cleanup-ledger.json', 'analysis/public-cleanup-inventory.json'}
    names |= self_files
    rows, ledger, removed, hazards = [], [], [], []
    for name in sorted(names):
        path = ROOT/name
        previous = old.get(name, {})
        if not path.exists():
            removed.append(name)
            ledger.append(dict(path=name, classification='C' if 'fixtures' in name else 'D', action='REMOVED',
                               reason='Private archaeology/fixture/blob preserved in ignored baseline archive; Party fixtures regenerate at test time.'))
            continue
        if name in self_files:
            continue  # Commit binds these manifests; avoid self-referential hashes.
        raw = path.read_bytes()
        mime = subprocess.check_output(['file', '--brief', '--mime-type', str(path)], text=True).strip()
        if name in DUMPS:
            cls, action, reason = 'D', 'MANUAL LEGAL REVIEW', 'Substantial translated command/effect/cue/region DATA remains; record payload migration does not clear these programs.'
        elif name in REWRITTEN:
            cls, action, reason = 'B', 'KEEP', 'Rewritten factual conclusions and explicit unresolved boundaries.'
        elif name.endswith('_test.go') or ('analysis/' in name and any(k in name for k in ['fixture','assets.json','matches.json','bios-font.json'])):
            cls, action, reason = 'C', 'MANUAL LEGAL REVIEW' if previous.get('action') == 'MANUAL LEGAL REVIEW' else 'KEEP', 'Original-backed hash/structural test or comparison; review residual small expected literals. Large Party expected-data fixtures regenerate privately.'
        elif previous.get('action') == 'MANUAL LEGAL REVIEW' and not name.endswith('.go'):
            cls, action, reason = 'E', 'MANUAL LEGAL REVIEW', 'Retained historical evidence/documentation may contain expressions, quotations or source-derived arrays; targeted text-level review/rewrite remains unresolved.'
        elif name.endswith(('.go','.py','.sh')) or name in {'.gitignore','go.mod'}:
            cls, action, reason = 'A', 'KEEP WITH OWNERSHIP REVIEW', 'Native implementation/tooling; source-guided authorship and ownership remain whole-project review matters.'
        else:
            cls, action, reason = 'B', 'KEEP WITH CONTENT REVIEW', 'Factual metadata/documentation; classification does not grant original-source or third-party redistribution rights.'
        row = dict(path=name, bytes=len(raw), sha256=hashlib.sha256(raw).hexdigest(), mime=mime,
                   classification=cls, action=action, reason=reason)
        rows.append(row)
        if previous.get('action') == 'MANUAL LEGAL REVIEW' or name in DUMPS or name in REWRITTEN:
            ledger.append(row)
        if (name != 'go.mod' and path.suffix.lower() in {'.prg','.mod','.hi','.cfg','.exe','.com','.wav','.pcm','.png','.bin','.appimage','.deb'}) or name.startswith(('reference/','release/','bin/')):
            hazards.append(name)
        if mime == 'application/octet-stream':
            hazards.append(name+': raw binary MIME')
    common = dict(baseline_commit=BASELINE, scope='Candidate tracked files and explicit new source files; all bytes read. No Git history or ignored inputs exported. Content triage, not legal clearance.')
    (ROOT/'analysis/public-cleanup-inventory.json').write_text(json.dumps(dict(common, files=rows, manifest_paths=sorted(self_files),
        removed_tracked_paths=removed, forbidden_payload_matches=hazards), indent=2)+'\n')
    classes = {'A':'native implementation','B':'factual documentation','C':'original-backed fixture/comparison','D':'substantial translated DATA/program/private dump','E':'uncertain; manual review'}
    plan_path = ROOT/'analysis/public-snapshot-plan.json'
    snapshot = {}
    if plan_path.exists():
        plan = json.loads(plan_path.read_text())
        selected = {r['path'] for r in plan['files'] if r['classification'] in {'PUBLIC_REQUIRED','PUBLIC_USEFUL'}}
        snapshot = dict(proposed_public_file_count=len(selected),
                        remaining_D_runtime_paths=sorted(DUMPS & selected),
                        excluded_private_archaeology=plan['counts']['PRIVATE_ARCHAEOLOGY'],
                        excluded_private_validation=plan['counts']['PRIVATE_VALIDATION'],
                        manual_decision_paths=plan['manual_decision_paths'],
                        forbidden_payload_matches=[h for h in hazards if h.split(':')[0] in selected],
                        readiness='NOT READY FOR PUBLIC SNAPSHOT' if DUMPS & selected else 'UNVERIFIED')
    (ROOT/'analysis/public-cleanup-ledger.json').write_text(json.dumps(dict(common, classes=classes, files=ledger,
        removed_tracked_paths=removed, totals=dict(collections.Counter(r['classification'] for r in ledger)),
        intended_snapshot=snapshot), indent=2)+'\n')
    print(f'Inspected {len(rows)} files plus 2 self-bound manifests; removed {len(removed)} paths; forbidden payloads {len(hazards)}; remaining translated program files {len(DUMPS)}')
    assert not hazards, hazards

if __name__ == '__main__':
    main()
