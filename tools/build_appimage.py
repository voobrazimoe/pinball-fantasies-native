#!/usr/bin/env python3
"""Build a public asset-free or personal bundled-data x86_64 AppImage with pinned official packaging tools.

Requires Linux x86_64, the repo Go/C toolchain, SDL2 build headers/library,
Python 3 and ldd. Packaging tools are downloaded once and SHA256 verified.
Use --tool/--runtime for offline copies matching tools/appimage-tools.json.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import urllib.request
import tempfile

root = Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--data-dir', type=Path, help='bundle validated personal originals; local-only output')
parser.add_argument('--mode', choices=['public', 'personal'])
parser.add_argument('--tool', type=Path)
parser.add_argument('--runtime', type=Path)
args = parser.parse_args()
mode = args.mode or ('personal' if args.data_dir else 'public')
if (mode == 'personal') != bool(args.data_dir):
    parser.error('personal mode requires --data-dir; public mode forbids it')
if mode == 'personal':
    from personal_assets import inventory
    originals, records = inventory(args.data_dir)
else:
    originals, records = None, []
print('AppImage mode:', mode, flush=True)
pins = json.loads((root/'tools/appimage-tools.json').read_text())
cache = root/'.cache/appimage-tools'
cache.mkdir(parents=True, exist_ok=True)

def tool(repo, override):
    spec = pins[repo]
    path = (override or cache/(repo+'-'+spec['version'])).resolve()
    if not path.exists():
        urllib.request.urlretrieve(spec['url'], path)
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if digest != spec['sha256']:
        raise SystemExit(f'{path}: checksum mismatch; expected {spec["sha256"]}')
    path.chmod(0o755)
    return path

appimagetool = tool('appimagetool', args.tool)
runtime = tool('type2-runtime', args.runtime)
out = root/('release/personal/linux' if mode == 'personal' else 'release/linux')
out.mkdir(parents=True, exist_ok=True)
stage = tempfile.TemporaryDirectory(prefix='PinballFantasies-'+mode+'-', dir=cache)
appdir = Path(stage.name)
libdir = appdir/'usr/lib'
bindir = appdir/'usr/bin'
libdir.mkdir(parents=True)
bindir.mkdir(parents=True)
subprocess.run([str(root/'tools/go.sh'), 'build', '-buildvcs=false', '-trimpath',
                '-ldflags=-X pinballfantasies/internal/platform.GUIMode=1',
                '-o', str(bindir/'pinballfantasies'), './cmd/pinballfantasies'],
               cwd=root, check=True, env=dict(os.environ, GOOS='linux', GOARCH='amd64', CGO_ENABLED='1'))

# Include the ELF dependency closure and a matching loader/libc. Software GDI/
# SDL rendering needs no proprietary GPU driver in this package. Host display
# and audio servers are still required; optional host ALSA plugins are avoided.
bundled = {}
def dependencies(binary):
    result = subprocess.run(['ldd', str(binary)], text=True, capture_output=True, check=True).stdout
    if 'not found' in result:
        raise SystemExit(result)
    for line in result.splitlines():
        m = re.search(r'(?:=>\s*)?(/\S+)\s+\(', line)
        if m:
            source = Path(m.group(1))
            destination = libdir/source.name
            if source.name in bundled:
                if hashlib.sha256(source.read_bytes()).hexdigest() != bundled[source.name]['sha256']:
                    raise SystemExit(f'conflicting library basename: {source}')
                continue
            shutil.copy2(source, destination, follow_symlinks=True)
            bundled[source.name] = {'source': str(source), 'sha256': hashlib.sha256(source.read_bytes()).hexdigest()}

dependencies(bindir/'pinballfantasies')
if 'libSDL2-2.0.so.0' not in bundled or 'ld-linux-x86-64.so.2' not in bundled:
    raise SystemExit('expected SDL2 and x86_64 loader in dependency closure')
# ALSA may load these independently of DT_NEEDED. Include available modules and
# their closures so ordinary direct ALSA output works without host SDL/plugins.
for base in [Path('/usr/lib/x86_64-linux-gnu/alsa-lib'), Path('/usr/lib/alsa-lib')]:
    if base.exists():
        target = libdir/'alsa-lib'
        target.mkdir(exist_ok=True)
        for plugin in base.glob('*.so'):
            # Optional codec/rate plugins pull unrelated FFmpeg/font material.
            if any(token in plugin.name for token in ('a52', 'lavrate')):
                continue
            shutil.copy2(plugin, target/plugin.name)
            bundled['alsa-lib/'+plugin.name] = {'source': str(plugin), 'sha256': hashlib.sha256(plugin.read_bytes()).hexdigest()}
            dependencies(plugin)
if Path('/usr/share/alsa').exists():
    shutil.copytree('/usr/share/alsa', appdir/'usr/share/alsa')

# Preserve distribution copyright/license notices for every bundled package.
licenses = appdir/'usr/share/licenses'
licenses.mkdir(parents=True)
packages = set()
unowned = []
notice_sources = {record['source'] for record in bundled.values()}
if Path('/usr/share/alsa').exists():
    notice_sources.update(str(p) for p in Path('/usr/share/alsa').rglob('*') if p.is_file())
for source in sorted(notice_sources):
    try:
        owners = subprocess.check_output(['dpkg-query', '-S', source], text=True, stderr=subprocess.DEVNULL).splitlines()
        for owner in owners:
            if ': ' in owner and not owner.startswith('diversion '):
                packages.add(owner.split(': ')[0].split(':')[0])
    except (FileNotFoundError, subprocess.CalledProcessError):
        unowned.append(source)
if unowned:
    raise SystemExit('No package license provenance for: '+', '.join(unowned))
for package in sorted(packages):
    notice = Path('/usr/share/doc')/package/'copyright'
    if not notice.exists():
        raise SystemExit('Missing runtime copyright notice: '+package)
    shutil.copy2(notice, licenses/(package+'.copyright'))
# Debian copyright notices reference these full license texts.
shutil.copytree('/usr/share/common-licenses', licenses/'common-licenses')
shutil.copy2(root/'.tools/go/LICENSE', licenses/'Go-LICENSE.txt')
shutil.copy2(root/'packaging/notices/AppImage-runtime-20251108.txt', licenses/'AppImage-runtime-20251108.txt')
(licenses/'BUILD-SOURCES.txt').write_text(
    'Runtime libraries copied from the build distribution, with notices here.\n'
    'Corresponding source: distribution source packages (apt source PACKAGE),\n'
    'using the versions in BUILD-MANIFEST.json; https://archive.ubuntu.com/ubuntu/pool/ (Ubuntu build); https://deb.debian.org/debian/pool/ (Debian build)\n'
    'SDL upstream source: https://github.com/libsdl-org/SDL\n'
    'GNU libc source: https://www.gnu.org/software/libc/\n'
    'AppImage runtime source: https://github.com/AppImage/type2-runtime/tree/20251108\n'
    'Before binary redistribution, arrange corresponding source/relinking obligations\n'
    'for the exact bundled distribution packages and static AppImage runtime closure.\n')
versions = {}
for package in sorted(packages):
    versions[package] = subprocess.check_output(['dpkg-query', '-W', '-f=${Version}', package], text=True)
manifest = {'mode': mode, 'original_inputs': records, 'tools': pins, 'libraries': bundled, 'packages': versions,
            'go': subprocess.check_output([str(root/'tools/go.sh'), 'version'], text=True).strip()}
(appdir/'BUILD-MANIFEST.json').write_text(json.dumps(manifest, indent=2)+'\n')
manifest_path = root/'.build-personal/linux-build-manifest.json' if mode == 'personal' else out/'BUILD-MANIFEST.json'
manifest_path.parent.mkdir(parents=True, exist_ok=True)
manifest_path.write_text(json.dumps(manifest, indent=2)+'\n')
if mode == 'personal':
    for ignored in [appdir/'usr/share/pinballfantasies/INTRO.PRG', out/'PinballFantasies-x86_64.AppImage']:
        subprocess.run(['git', 'check-ignore', '-q', str(ignored)], cwd=root, check=True)
    data = appdir/'usr/share/pinballfantasies'
    data.mkdir(parents=True)
    for record in records:
        shutil.copyfile(originals/record['name'], data/record['name'])
        (data/record['name']).chmod(0o444)
        if hashlib.sha256((data/record['name']).read_bytes()).hexdigest() != record['sha256']:
            raise SystemExit('Original changed during packaging: '+record['name'])
# Append a default only when no explicit -data-dir override was supplied.
# The outer APPIMAGE still resolves writable userdata through PortableRoot.
bundled_args = ""
if mode == 'personal':
    bundled_args = '''has_data=false
for arg in "$@"; do
 case "$arg" in -data-dir|--data-dir|-data-dir=*|--data-dir=*) has_data=true;; esac
done
if [ "$has_data" = false ]; then set -- "$@" -data-dir "$appdir/usr/share/pinballfantasies"; fi
'''
(appdir/'AppRun').write_text('''#!/bin/sh
set -eu
appdir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
export ALSA_CONFIG_DIR="$appdir/usr/share/alsa"
export ALSA_PLUGIN_DIR="$appdir/usr/lib/alsa-lib"
'''+bundled_args+'''exec "$appdir/usr/lib/ld-linux-x86-64.so.2" --inhibit-cache --library-path "$appdir/usr/lib" "$appdir/usr/bin/pinballfantasies" "$@"
''')
(appdir/'AppRun').chmod(0o755)
(appdir/'pinballfantasies.desktop').write_text('''[Desktop Entry]
Type=Application
Name=Pinball Fantasies
Exec=pinballfantasies
Icon=pinballfantasies
Categories=Game;ArcadeGame;
Terminal=false
''')
# A simple original geometric packaging icon, containing no original game art.
(appdir/'pinballfantasies.svg').write_text('''<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128" viewBox="0 0 128 128"><rect width="128" height="128" rx="24" fill="#152839"/><circle cx="64" cy="43" r="20" fill="#e5edf4"/><path d="M22 102L53 85M106 102L75 85" stroke="#edaf47" stroke-width="14" stroke-linecap="round"/></svg>''')
# Git checkouts and audited exports can have different local umasks. Normalize
# staging permissions so those host permission bits never change release bytes.
for staged in [appdir, *appdir.rglob('*')]:
    if staged.is_dir():
        staged.chmod(0o755)
    elif staged.is_file():
        staged.chmod(0o755 if staged.stat().st_mode & 0o111 else 0o644)

artifact = out/'PinballFantasies-x86_64.AppImage'
subprocess.run([str(appimagetool), '--appimage-extract-and-run', '--no-appstream',
                '--runtime-file', str(runtime), '--mksquashfs-opt', '-processors',
                '--mksquashfs-opt', '1', str(appdir), str(artifact)], check=True,
               env=dict(os.environ, ARCH='x86_64', SOURCE_DATE_EPOCH=os.environ.get('SOURCE_DATE_EPOCH', '1790812800')))
artifact.chmod(0o755)
print(artifact)
print('SHA256:', hashlib.sha256(artifact.read_bytes()).hexdigest())
stage.cleanup()
