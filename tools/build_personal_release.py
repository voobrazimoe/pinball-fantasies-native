#!/usr/bin/env python3
"""Build local-only single-file Windows/Linux releases from validated originals."""
import argparse
import fcntl
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile
from personal_assets import ROOT, inventory

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('data_dir', type=Path)
parser.add_argument('--windows-only', action='store_true', help='build only the portable Windows EXE; no AppImage tools required')
parser.add_argument('--test-cheats', action='store_true', help='build a separately named Windows matrix diagnostic EXE')
parser.add_argument('--tool', type=Path, help='offline pinned appimagetool')
parser.add_argument('--runtime', type=Path, help='offline pinned AppImage runtime')
args = parser.parse_args()
if args.test_cheats and not args.windows_only:
    parser.error('--test-cheats requires --windows-only')
metadata = ROOT/'.build-personal'
metadata.mkdir(exist_ok=True)
with (metadata/'build.lock').open('w') as lock:
    fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
    data, records = inventory(args.data_dir)
    generated = ROOT/'internal/platform/personal_payload_windows.go'
    bundle = ROOT/'internal/platform/.personal-assets/data.zip'
    artifact = ROOT/('release/personal/windows/pinballfantasies-matrix-test.exe' if args.test_cheats else 'release/personal/windows/pinballfantasies.exe')
    # Refuse to copy originals unless every commercial destination is ignored.
    for path in [generated, bundle, artifact, ROOT/'release/personal/linux/PinballFantasies-x86_64.AppImage']:
        subprocess.run(['git', 'check-ignore', '-q', str(path)], cwd=ROOT, check=True)
    bundle.parent.mkdir(parents=True, exist_ok=True)
    artifact.parent.mkdir(parents=True, exist_ok=True)
    try:
        with zipfile.ZipFile(bundle, 'w', compression=zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
            for record in records:
                contents = (data/record['name']).read_bytes()
                if hashlib.sha256(contents).hexdigest() != record['sha256']:
                    raise SystemExit('Original changed during build: '+record['name'])
                info = zipfile.ZipInfo(record['name'], date_time=(2000, 1, 1, 0, 0, 0))
                info.compress_type = zipfile.ZIP_DEFLATED
                archive.writestr(info, contents)
        generated.write_text('//go:build windows && personal\n\npackage platform\n\nimport _ "embed"\n\n//go:embed .personal-assets/data.zip\nvar personalPayload []byte\n')
        tags = 'personal,matrixdebug' if args.test_cheats else 'personal'
        subprocess.run([str(ROOT/'tools/go.sh'), 'build', '-tags='+tags, '-buildvcs=false',
                        '-trimpath', '-p=1', '-ldflags=-H=windowsgui -X pinballfantasies/internal/platform.GUIMode=1',
                        '-o', str(artifact), './cmd/pinballfantasies'], cwd=ROOT, check=True,
                       env=dict(os.environ, GOOS='windows', GOARCH='amd64', CGO_ENABLED='0'))
    finally:
        generated.unlink(missing_ok=True)
        bundle.unlink(missing_ok=True)
        bundle.parent.rmdir()
    artifacts = [artifact]
    if not args.windows_only:
        command = ['python3', str(ROOT/'tools/build_appimage.py'), '--mode', 'personal', '--data-dir', str(data)]
        for flag in ['tool', 'runtime']:
            value = getattr(args, flag)
            if value:
                command += ['--'+flag, str(value.resolve())]
        subprocess.run(command, cwd=ROOT, check=True)
        artifacts.append(ROOT/'release/personal/linux/PinballFantasies-x86_64.AppImage')
    output = {'original_inputs': records, 'artifacts': {str(p.relative_to(ROOT)): hashlib.sha256(p.read_bytes()).hexdigest() for p in artifacts}}
    (metadata/'build-manifest.json').write_text(json.dumps(output, indent=2)+'\n')
    print(json.dumps(output, indent=2))
