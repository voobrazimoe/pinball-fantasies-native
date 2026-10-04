#!/usr/bin/env python3
"""Classify current tracked paths by publication purpose, without exporting them.

This is a proposal, never rights approval. Runtime D blockers stay visible as
PUBLIC_REQUIRED with unresolved provenance; excluding them would break the build.
"""
import collections
import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent.parent
DUMPS = {
    'internal/presentation/content.go', 'internal/partyland/timing_data.go',
    'internal/speeddevils/content.go', 'internal/gameshow/content.go',
    'internal/stones/content.go',
}
DOCS = {'docs/runtime-compatibility-audit.md', 'README.md', 'docs/build.md', 'docs/runtime-data.md', 'docs/architecture.md'}
BUILD = {
    '.gitignore', 'go.mod', 'tools/go.sh', 'tools/setup-local.py',
    'tools/build_windows.sh', 'tools/build_release.sh', 'tools/build_appimage.py',
    'tools/appimage-tools.json', 'packaging/README-linux.txt',
    'packaging/README-windows.txt', 'packaging/notices/AppImage-runtime-20251108.txt',
}
PUBLIC_TESTS = {
    'internal/audio/compatibility_test.go', 'internal/datalayout/layout_test.go',
    'internal/settings/settings_test.go',
    'internal/speeddevils/lamp_decode_test.go', 'internal/stones/area_decode_test.go',
    *('internal/platform/'+name+'_test.go' for name in (
        'fullscreen','keys_reconcile','personal_data','sizing','storage','transfer')),
}
NEW = {
       'docs/runtime-compatibility-audit.md', 'internal/assets/layout.go',
       'internal/datalayout/layout.go', 'internal/datalayout/prepare.go', 'internal/datalayout/picture.go',
       'internal/datalayout/profiles.json', 'internal/datalayout/layout_test.go',
       'internal/audio/compatibility_test.go','analysis/public-snapshot-plan.json', 'analysis/program-data-manifest.json',
       'analysis/final-cleanup-validation.json', 'docs/architecture.md',
       'tools/plan_public_snapshot.py',
       'internal/speeddevils/lamp_decode_test.go', 'internal/speeddevils/lamp_parity_test.go',
       'internal/stones/area_decode_test.go', 'internal/stones/area_parity_test.go'}

def classify(name):
    if name in DUMPS:
        return 'PUBLIC_REQUIRED', 'Runtime dependency; substantial historical DATA remains. Technical migration is mandatory before approval.', 'D / MANUAL LEGAL REVIEW'
    if name in DOCS:
        return 'PUBLIC_USEFUL', 'Concise build, usage, architecture or external-data/provenance documentation.', None
    if name in BUILD:
        return 'PUBLIC_REQUIRED', 'Build/package dependency or required third-party packaging notice.', None
    if name in PUBLIC_TESTS:
        return 'PUBLIC_USEFUL', 'Native platform/settings or synthetic record validation without original game/source fixtures.', None
    if name.startswith('internal/oracle/'):
        return 'PRIVATE_VALIDATION', 'Exact original identity gate for research and parity fixtures; not a runtime dependency.', None
    if name.endswith('_test.go') or name.endswith('matrix_reachability_data.go') or name.startswith('internal/testinputs/'):
        return 'PRIVATE_VALIDATION', 'Original-backed/reference-dependent suite or development verification; retain privately.', None
    if name == 'internal/datalayout/profiles.json':
        return 'PUBLIC_REQUIRED', 'Address-only runtime compatibility profiles; no commercial payload.', None
    if name.startswith('internal/') and name.endswith('.go') or name.startswith('cmd/pinballfantasies/'):
        return 'PUBLIC_REQUIRED', 'Native runtime implementation; subject to whole-project ownership review.', None
    if name.startswith(('cmd/pf8','cmd/personalvalidate')):
        return 'PRIVATE_VALIDATION', 'Private comparison/personal-data validation command; not needed by the minimal runtime build.', None
    if name.startswith('analysis/'):
        cls = 'PRIVATE_VALIDATION' if any(k in name for k in ('validation','fixture','parity','checkpoints','manifest','matches','preflight','cleanup','snapshot')) or name.endswith('.log') else 'PRIVATE_ARCHAEOLOGY'
        return cls, 'Detailed reconstruction evidence is retained privately; not required to use or build the project.', None
    if name.startswith('tools/'):
        return 'PRIVATE_VALIDATION' if any(k in name for k in ('test_','audit_','check_','matrix_','smoke_','reference_','parity','pf8','compare_','export_public','plan_public','prg_')) else 'PRIVATE_ARCHAEOLOGY', 'Development/reference/personal or unrelated platform tooling; exclude from minimal snapshot.', None
    return 'PRIVATE_ARCHAEOLOGY', 'Historical document or unrelated development artifact; preserve privately.', None

def main():
    names=set(subprocess.check_output(['git','ls-files'],cwd=ROOT,text=True).splitlines())|NEW
    rows=[]
    for name in sorted(names):
        cls, reason, blocker=classify(name)
        path=ROOT/name
        row={'path':name,'classification':cls,'reason':reason}
        if blocker: row['publication_blocker']=blocker
        if path.exists() and name!='analysis/public-snapshot-plan.json':
            row['sha256']=hashlib.sha256(path.read_bytes()).hexdigest()
        rows.append(row)
    counts=dict(collections.Counter(r['classification'] for r in rows))
    plan={'baseline_commit':'136e8cc40c87758cfda3c9f5fc771316b1c09a47',
          'status':'NOT READY FOR PUBLIC SNAPSHOT',
          'scope':'Every tracked path plus explicit new files. Proposal only; no export, rights clearance, license choice or publication.',
          'classifications':['PUBLIC_REQUIRED','PUBLIC_USEFUL','PRIVATE_ARCHAEOLOGY','PRIVATE_VALIDATION','REMOVE_OBSOLETE','MANUAL_DECISION'],
          'counts':{k:counts.get(k,0) for k in ['PUBLIC_REQUIRED','PUBLIC_USEFUL','PRIVATE_ARCHAEOLOGY','PRIVATE_VALIDATION','REMOVE_OBSOLETE','MANUAL_DECISION']},
          'proposed_public_file_count':counts.get('PUBLIC_REQUIRED',0)+counts.get('PUBLIC_USEFUL',0),
          'unresolved_public_runtime_blockers':sorted(DUMPS),'manual_decision_paths':[],
          'files':rows}
    (ROOT/'analysis/public-snapshot-plan.json').write_text(json.dumps(plan,indent=2)+'\n')
    print(json.dumps(plan['counts']), 'proposed public:',plan['proposed_public_file_count'])

if __name__=='__main__':main()
