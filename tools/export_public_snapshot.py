#!/usr/bin/env python3
"""Export an explicitly reviewed commit allowlist, without any Git history.

Requires a locally prepared approval JSON with commit, legal_decisions (text),
files ({path: SHA256}), and excluded ({path: reason}). Every commit path must
be accounted for. This tool does not determine rights or create/push a repo.
"""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import re
import subprocess

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--approval', required=True, type=Path)
p.add_argument('--output', required=True, type=Path)
a = p.parse_args()
root = Path(__file__).resolve().parent.parent
review = json.loads(a.approval.read_text())
commit = review['commit']
assert re.fullmatch(r'[0-9a-f]{40}', commit), 'pin a full commit SHA'
assert isinstance(review.get('legal_decisions'), str) and review['legal_decisions'].strip(), 'record actual owner/legal decisions'
files, excluded = review['files'], review['excluded']
assert files and not set(files).intersection(excluded)
records = subprocess.check_output(['git', 'ls-tree', '-rz', '--full-tree', commit], cwd=root).split(b'\0')
tree = {}
for record in filter(None, records):
    meta, name = record.split(b'\t', 1)
    mode, kind, oid = meta.decode().split()
    tree[name.decode()] = (mode, kind, oid)
assert set(tree) == set(files) | set(excluded), 'review must account for EVERY commit path'
assert all(isinstance(v, str) and v.strip() for v in excluded.values())
assert not a.output.exists(), 'output must be a new absent directory'
# Validate every blob before creating output. Never copy filesystem originals,
# local caches, .git, refs, submodules, symlinks or ignored build artifacts.
blobs = {}
for name, digest in files.items():
    path = PurePosixPath(name)
    assert not path.is_absolute() and '..' not in path.parts
    assert path.parts[0] not in {'reference','release','bin','userdata'}, name
    assert not any(part.startswith('.') and part not in ('.gitignore', '.github') for part in path.parts), name
    assert name == 'go.mod' or path.suffix.lower() not in {'.prg','.mod','.hi','.cfg','.exe','.com','.wav','.pcm','.appimage','.pyc','.deb','.patch','.diff'}, name
    mode, kind, oid = tree[name]
    assert kind == 'blob' and mode in ('100644', '100755'), name
    data = subprocess.check_output(['git', 'cat-file', 'blob', oid], cwd=root)
    assert hashlib.sha256(data).hexdigest() == digest, 'unreviewed bytes: '+name
    blobs[name] = (data, mode)
a.output.mkdir(parents=True)
for name, (data, mode) in blobs.items():
    dest = a.output/name
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_bytes(data)
    dest.chmod(0o755 if mode == '100755' else 0o644)
print(f'Exported {len(blobs)} reviewed files from {commit}; no .git or external inputs copied')
