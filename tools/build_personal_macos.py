#!/usr/bin/env python3
"""Build local-only macOS apps with validated originals and automatic first-run import."""
import argparse
import fcntl
import hashlib
import json
from pathlib import Path
import plistlib
import shutil
import subprocess
from personal_assets import ROOT, inventory


def run(*args):
    subprocess.run(args, cwd=ROOT, check=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('data_dir', type=Path)
    parser.add_argument('--arch', choices=['arm64', 'x86_64', 'both'], default='both')
    args = parser.parse_args()
    metadata = ROOT/'.build-personal'
    metadata.mkdir(exist_ok=True)
    with (metadata/'build.lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        data, records = inventory(args.data_dir)
        architectures = ['arm64', 'x86_64'] if args.arch == 'both' else [args.arch]
        artifacts = {}
        for arch in architectures:
            output = ROOT/'release/personal'/('macos-'+arch)
            run('git', 'check-ignore', '-q', str(output/'Pinball Fantasies.app/Contents/Resources/Data/INTRO.PRG'))
            run(str(ROOT/'tools/build_macos.sh'), arch)
            public = ROOT/'release'/('macos' if arch == 'arm64' else 'macos-x86_64')
            app = output/'Pinball Fantasies.app'
            output.mkdir(parents=True, exist_ok=True)
            if app.exists():
                shutil.rmtree(app)
            shutil.copytree(public/app.name, app)
            bundled = app/'Contents/Resources/Data'
            bundled.mkdir()
            for record in records:
                contents = (data/record['name']).read_bytes()
                if hashlib.sha256(contents).hexdigest() != record['sha256']:
                    raise RuntimeError('Original changed during build: '+record['name'])
                (bundled/record['name']).write_bytes(contents)
            info = plistlib.loads((app/'Contents/Info.plist').read_bytes())
            revision = info['PFHostBuild']
            run('codesign', '--force', '--sign', '-', str(app))
            run('codesign', '--verify', '--strict', str(app))
            run('lipo', '-verify_arch', arch, str(app/'Contents/MacOS/pinballfantasies'))
            archive = output/f'PinballFantasies-personal-{arch}-{revision}.zip'
            if archive.exists():
                archive.unlink()
            run('ditto', '-c', '-k', '--sequesterRsrc', '--keepParent', str(app), str(archive))
            artifacts[str(archive.relative_to(ROOT))] = hashlib.sha256(archive.read_bytes()).hexdigest()
        for record in records:
            if hashlib.sha256((data/record['name']).read_bytes()).hexdigest() != record['sha256']:
                raise RuntimeError('Original changed: '+record['name'])
        manifest = {'original_inputs': records, 'artifacts': artifacts}
        (metadata/'macos-build-manifest.json').write_text(json.dumps(manifest, indent=2)+'\n')
        for artifact in artifacts:
            print('Built local personal package: '+artifact)


if __name__ == '__main__':
    main()
