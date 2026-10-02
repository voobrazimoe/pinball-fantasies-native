#!/usr/bin/env python3
"""Bounded artifact checks: Unicode/spaces, outer roots, overrides and fallback.

Run after build_release.sh. --live additionally opens each backend for 3 s;
run desktop smokes sequentially. Original assets are supplied from the checkout
only in temporary test folders, never in release artifacts.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

root = Path(__file__).resolve().parent.parent
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--backend', choices=['linux', 'wine', 'both'], default='both')
parser.add_argument('--live', action='store_true')
args = parser.parse_args()
out = root/'analysis/pf12-validation/portable'
out.mkdir(parents=True, exist_ok=True)
originals = [p for p in root.iterdir() if p.is_file() and p.suffix.upper() in ['.PRG', '.MOD', '.CFG', '.HI', '.SDR']]
before = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in originals}
results = []
for backend in (['linux', 'wine'] if args.backend == 'both' else [args.backend]):
    with tempfile.TemporaryDirectory(prefix='pf12 portable Ж 日本 ') as base:
        folder = Path(base)/'Game folder Юникод'
        folder.mkdir()
        for p in originals:
            shutil.copy2(p, folder/p.name)
        binary = (root/'release/linux/PinballFantasies-x86_64.AppImage' if backend == 'linux'
                  else root/'release/windows/pinballfantasies.exe')
        game = folder/binary.name
        shutil.copy2(binary, game)
        path = lambda p: str(p) if backend == 'linux' else 'Z:'+str(p).replace('/', '\\')
        fallback = Path(base)/'Fallback user config'
        env = dict(os.environ)
        if backend == 'linux':
            env.update(XDG_CONFIG_HOME=str(fallback), SDL_VIDEODRIVER='x11', SDL_AUDIODRIVER='pulseaudio')
            command = [str(game), '--appimage-extract-and-run']
        else:
            env.update(APPDATA=path(fallback), WINEPREFIX=os.environ.get('PF12_WINEPREFIX', '/tmp/pf12-wine'), WINEDEBUG='-all')
            command = ['wine64', str(game)]

        def run(extra, trace=False):
            runenv = dict(env)
            if trace:
                runenv['LD_DEBUG'] = 'libs'
            launch = command+extra
            if backend == 'wine':
                runenv.update(PF12_PORTABLE_COMMAND=json.dumps([path(game)]+extra), PF12_PORTABLE_APPDATA=path(fallback))
                launch = ['wine64', str(root/'bin/platform-windows.test.exe'), '-test.run', 'TestPF12PortableArtifactLaunch']
            result = subprocess.run(launch, cwd='/tmp', env=runenv,
                                    text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=30)
            if result.returncode:
                raise AssertionError((result.returncode, result.stdout[-1000:], result.stderr[-1000:]))
            return result

        png = Path(base)/'frame.png'
        result = run(['-png', path(png)], trace=backend == 'linux')
        assert png.stat().st_size > 0, 'default external data discovery failed'
        assert (folder/'userdata/native.log').is_file()
        assert not fallback.exists(), 'writable portable folder unexpectedly used user config'
        libraries = []
        if backend == 'linux':
            libraries = [line.strip() for line in result.stderr.splitlines()
                         if 'calling init:' in line and '.so' in line]
            sdl = [line for line in libraries if 'libSDL2' in line]
            assert sdl and all('/usr/lib/' in line and '/tmp/' in line for line in sdl), sdl
            # AppRun's POSIX shell starts with host libc; the game is then
            # exec'd through the bundled loader. Validate that second load.
            start = next(i for i,line in enumerate(libraries) if '/tmp/' in line and 'ld-linux' in line)
            libraries = libraries[start:]
            assert all('/tmp/' in line for line in libraries), libraries
        if args.live:
            run(['-duration', '3s'])
            assert 'window opened' in (folder/'userdata/native.log').read_text(), 'live window did not open'

        override = Path(base)/'Explicit config Ж'
        scores = Path(base)/'Explicit scores Ж'
        run(['-png', path(png), '-config-dir', path(override), '-high-score-dir', path(scores)])
        assert (override/'native.log').exists() and scores.is_dir()

        shutil.rmtree(folder/'userdata')
        (folder/'userdata').write_text('deliberately blocks portable state directory')
        run(['-png', path(png)])
        name = 'pinballfantasies' if backend == 'linux' else 'PinballFantasies'
        assert (fallback/name/'native.log').exists(), 'user fallback missing'
        for p in originals:
            assert hashlib.sha256((folder/p.name).read_bytes()).hexdigest() == before[p.name], p.name
        record = {'backend': backend, 'default_external_data': True, 'unicode_spaces': True,
                  'portable_userdata': True, 'explicit_overrides': True, 'user_fallback': True,
                  'originals_unchanged': True, 'live_seconds': 3 if args.live else 0,
                  'bundled_library_init': libraries}
        results.append(record)
        print(json.dumps(record, ensure_ascii=False), flush=True)
assert {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in originals} == before
(out/'artifact-smoke.json').write_text(json.dumps(results, ensure_ascii=False, indent=2)+'\n')
