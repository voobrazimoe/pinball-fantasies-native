#!/usr/bin/env python3
"""Run local Apple Silicon gates with owner-supplied originals, never upload them.
Downstream tests receive isolated copies of required payload and optional settings.
No state is written to the originals or application bundle.
"""
import argparse, hashlib, json, os, pathlib, platform, subprocess, tarfile, tempfile
import shutil

IMMUTABLE_NAMES = ['INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD'] + [
    f'TABLE{n}.{ext}' for n in range(1, 5) for ext in ('PRG', 'MOD')]


def input_hashes(data):
    names = IMMUTABLE_NAMES + (['PINBALL.CFG'] if (data/'PINBALL.CFG').exists() else [])
    return {name: hashlib.sha256((data/name).read_bytes()).hexdigest() for name in names}


def check_inventory(before, expected):
    for name in IMMUTABLE_NAMES:
        if before[name] != expected[name]:
            raise ValueError(f'{name} differs from the pinned original inventory; use a pristine test copy')


def stage_inputs(data, destination):
    destination.mkdir(parents=True, exist_ok=True)
    for name in IMMUTABLE_NAMES:
        shutil.copyfile(data/name, destination/name)
    cfg = data/'PINBALL.CFG'
    if cfg.exists():
        shutil.copyfile(cfg, destination/'PINBALL.CFG')
    else:
        # DOS encoding of shared settings.Defaults(): medium scrolling, other fields zero.
        (destination/'PINBALL.CFG').write_bytes(bytes([0, 0, 1, 0, 0, 0]))


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--data', required=True, type=pathlib.Path)
    p.add_argument('--reference', type=pathlib.Path, help='optional owner-supplied original DOS source directory')
    a = p.parse_args()
    root = pathlib.Path(__file__).resolve().parent.parent
    if platform.system() != 'Darwin' or platform.machine() != 'arm64':
        p.error('this validation requires a real Apple Silicon Mac and Apple toolchain')
    data = a.data.resolve(strict=True)
    before = input_hashes(data)
    expected = {r['name']: r['sha256'] for r in json.loads((root/'analysis/game-inventory.json').read_text())}
    try:
        check_inventory(before, expected)
    except ValueError as error:
        p.error(str(error))
    try:
        with tempfile.TemporaryDirectory(prefix='pf-macos-original-gates-') as directory:
            tmp = pathlib.Path(directory)
            staged = tmp/'inputs'
            stage_inputs(data, staged)
            env = dict(os.environ, PF_ENGINE_DATA_DIR=str(staged), CGO_ENABLED='1', GOOS='darwin', GOARCH='arm64')
            local_go = root/'.tools/go/bin'
            if (local_go/'go').is_file():
                env['PATH'] = str(local_go) + os.pathsep + env.get('PATH', '')
            env.setdefault('GOCACHE', str(root/'.cache/go-build'))
            env.setdefault('GOMODCACHE', str(root/'.cache/go-mod'))
            env['CC'] = subprocess.check_output(['xcrun', '--find', 'clang'], text=True).strip()
            sdk = subprocess.check_output(['xcrun', '--sdk', 'macosx', '--show-sdk-path'], text=True).strip()
            env.update(SDKROOT=sdk, CGO_CFLAGS=f'-isysroot {sdk} -arch arm64 -mmacosx-version-min=13.0',
                       CGO_LDFLAGS=f'-isysroot {sdk} -arch arm64 -mmacosx-version-min=13.0')
            subprocess.run([str(root/'tools/build_macos.sh')], cwd=root, env=env, check=True)
            checkout = tmp/'checkout'
            checkout.mkdir()
            archive = tmp/'source.tar'
            with archive.open('wb') as out:
                subprocess.run(['git', 'archive', 'HEAD'], cwd=root, stdout=out, check=True)
            with tarfile.open(archive) as source:
                if hasattr(tarfile, 'data_filter'):
                    source.extractall(checkout, filter='data')
                else:
                    # Python 3.9 on macOS lacks filters; this archive is generated
                    # above from the trusted local Git checkout, never downloaded.
                    source.extractall(checkout)
            # Exercise the current inventory regression fix even before it is committed.
            shutil.copyfile(root/'internal/assets/inventory_test.go', checkout/'internal/assets/inventory_test.go')
            stage_inputs(data, checkout)
            if a.reference:
                (checkout/'reference').mkdir(exist_ok=True)
                (checkout/'reference/original-dos-source').symlink_to(a.reference.resolve(strict=True), target_is_directory=True)
            packages = subprocess.check_output(['go', 'list', './...'], cwd=checkout, env=env, text=True).splitlines()
            packages = [pkg for pkg in packages if pkg not in ('pinballfantasies/internal/platform', 'pinballfantasies/cmd/pinballfantasies')]
            subprocess.run(['go', 'test', '-p=1', '-count=1', '-v', *packages], cwd=checkout, env=env, check=True)
            subprocess.run(['go', 'run', './cmd/personalvalidate', '-data-dir', str(staged)], cwd=checkout, env=env, check=True)
            subprocess.run(['python3', str(root/'tools/check_macos_bundle.py'), str(root/'release/macos/Pinball Fantasies.app'),
                            '--originals', str(staged)], cwd=root, check=True)
    finally:
        if before != input_hashes(data):
            raise RuntimeError('owner-supplied inputs were modified')
        print('PASS owner-supplied PRG/MOD and optional CFG unchanged')
    print('PASS local macOS original-backed ABI, four-table/all-scroll host journeys, shared Go gates, storage restart, public payload scan')
    print('Physical window/input/cursor/audio/fullscreen/import and real restart acceptance still require the docs/macos.md checklist.')


if __name__ == '__main__':
    main()
