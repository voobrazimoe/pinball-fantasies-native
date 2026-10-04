#!/usr/bin/env python3
"""Build a private, debug-signed Android APK from owner-supplied originals."""
import argparse
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import sys
import zipfile
from personal_assets import ROOT, PERSONAL_INPUTS, inventory


def ignored(root, paths):
    for path in paths:
        # A symlink must not redirect an ignored destination into tracked/outside storage.
        if not path.resolve().is_relative_to(root.resolve()):
            raise RuntimeError('Destination escapes repository: ' + str(path))
        for parent in [path, *path.parents]:
            if parent == root: break
            if parent.is_symlink():
                raise RuntimeError('Symlink in private destination: ' + str(parent))
        subprocess.run(['git', 'check-ignore', '-q', str(path)], cwd=root, check=True)


def verify_apk(path, records=None):
    with zipfile.ZipFile(path) as apk:
        names = apk.namelist()
        if len(names) != len(set(names)):
            raise RuntimeError('Duplicate APK entries')
        for abi in ('arm64-v8a', 'x86_64'):
            for lib in ('libpfengine.so', 'libpinball_android.so', 'libc++_shared.so', 'liboboe.so'):
                if f'lib/{abi}/{lib}' not in names:
                    raise RuntimeError(f'Missing native library: {abi}/{lib}')
        commercial = {name for name in names if
                      Path(name).name.upper() in (*PERSONAL_INPUTS, 'PINBALL.CFG') or
                      re.fullmatch(r'TABLE.*\.HI', Path(name).name, re.I) or
                      name.startswith('assets/personal-data')}
        expected = {f"assets/personal-data/{r['name']}" for r in records or []}
        if commercial != expected:
            raise RuntimeError('Unexpected commercial APK entries: ' + str(sorted(commercial ^ expected)))
        for record in records or []:
            data = apk.read('assets/personal-data/' + record['name'])
            if len(data) != record['bytes'] or hashlib.sha256(data).hexdigest() != record['sha256']:
                raise RuntimeError('APK input differs from validated original: ' + record['name'])


def build(data_dir, root=ROOT, gradle='gradle'):
    root = Path(root).absolute()
    metadata = root / '.build-personal'
    assets = root / 'hosts/android/app/.personal-assets'
    work = metadata / 'android-work'
    output = root / 'release/personal/android/pinball-fantasies-personal.apk'
    manifest = metadata / 'android-manifest.json'
    # Check before creating metadata/lock too. No cached prior APK can imply success.
    ignored(root, [metadata/'android.lock', manifest, assets/'personal-data/.ignore-probe',
                   work/'app/.ignore-probe', work/'gradle/.ignore-probe', output])
    metadata.mkdir(exist_ok=True)
    with (metadata/'android.lock').open('w') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        output.unlink(missing_ok=True)
        manifest.unlink(missing_ok=True)
        try:
            data, records = inventory(data_dir)
            destinations = [assets/'personal-data'/r['name'] for r in records]
            ignored(root, destinations)
            # Always isolate Gradle merged assets, intermediates and APK output.
            if assets.exists(): shutil.rmtree(assets)
            if work.exists(): shutil.rmtree(work)
            (assets/'personal-data').mkdir(parents=True)
            for record, destination in zip(records, destinations):
                contents = (data/record['name']).read_bytes()
                if hashlib.sha256(contents).hexdigest() != record['sha256']:
                    raise RuntimeError('Original changed during build: ' + record['name'])
                destination.write_bytes(contents)
            env = dict(os.environ)
            if not env.get('ANDROID_NDK_HOME') and not env.get('ANDROID_NDK_ROOT'):
                sdk = env.get('ANDROID_HOME') or env.get('ANDROID_SDK_ROOT')
                if sdk: env['ANDROID_NDK_HOME'] = str(Path(sdk)/'ndk/28.2.13676358')
            if sys.platform == 'darwin': env.setdefault('PF_ANDROID_NDK_HOST_TAG', 'darwin-x86_64')
            subprocess.run(['sh', str(root/'tools/build_android_engine.sh'), 'all'], cwd=root, env=env, check=True)
            command = [gradle, '-p', str(root/'hosts/android'), ':app:assembleDebug',
                       '--no-build-cache', '--no-configuration-cache', '--no-daemon']
            subprocess.run(command + ['--project-cache-dir', str(work/'gradle'),
                           '-PpersonalAssetsDir='+str(assets), '-PpersonalBuildDir='+str(work/'app')],
                           cwd=root, env=env, check=True)
            private_apk = work/'app/outputs/apk/debug/app-debug.apk'
            verify_apk(private_apk, records)
            # Mandatory separate ordinary build, without either opt-in property.
            subprocess.run(command, cwd=root, env=env, check=True)
            public_apk = root/'hosts/android/app/build/outputs/apk/debug/app-debug.apk'
            verify_apk(public_apk)
            output.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(private_apk, output)
            verify_apk(output, records)
            commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip()
            result = {'apk': str(output), 'sha256': hashlib.sha256(output.read_bytes()).hexdigest(),
                      'source_commit': commit, 'included': [{k: r[k] for k in ('name', 'bytes')} for r in records],
                      'public_apk_asset_free': True}
            manifest.write_text(json.dumps(dict(result, original_inputs=records), indent=2)+'\n')
        except BaseException:
            output.unlink(missing_ok=True)
            manifest.unlink(missing_ok=True)
            raise
        finally:
            try:
                if assets.exists(): shutil.rmtree(assets)
                if work.exists(): shutil.rmtree(work)
            except BaseException:
                output.unlink(missing_ok=True)
                manifest.unlink(missing_ok=True)
                raise
        print(json.dumps(dict(result, temporary_commercial_assets_removed=True), indent=2))
        return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('data_dir', type=Path)
    args = parser.parse_args()
    # SIGINT already raises KeyboardInterrupt; give termination the same cleanup path.
    def terminate(signum, frame):
        raise KeyboardInterrupt('Personal build terminated')
    signal.signal(signal.SIGTERM, terminate)
    build(args.data_dir)


if __name__ == '__main__':
    main()
