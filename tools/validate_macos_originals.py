#!/usr/bin/env python3
"""Run local Apple Silicon gates with owner-supplied originals, never upload them.
The clean temporary Go checkout receives only required read-only test inputs.
No state is written to the originals or application bundle.
"""
import argparse, hashlib, json, os, pathlib, platform, subprocess, tarfile, tempfile
p=argparse.ArgumentParser(description=__doc__)
p.add_argument('--data',required=True,type=pathlib.Path)
p.add_argument('--reference',type=pathlib.Path,help='optional owner-supplied original DOS source directory')
a=p.parse_args(); root=pathlib.Path(__file__).resolve().parent.parent
if platform.system()!='Darwin' or platform.machine()!='arm64':
    p.error('this validation requires a real Apple Silicon Mac and Apple toolchain')
data=a.data.resolve(strict=True)
names=['INTRO.PRG','INTRO.MOD','MOD2.MOD','PINBALL.CFG']+[f'TABLE{n}.{ext}' for n in range(1,5) for ext in ('PRG','MOD')]
before={name:hashlib.sha256((data/name).read_bytes()).hexdigest() for name in names}
expected={r['name']:r['sha256'] for r in json.loads((root/'analysis/game-inventory.json').read_text())}
for name,digest in before.items():
    if digest!=expected[name]: p.error(f'{name} differs from the pinned original inventory; use a pristine test copy')
env=dict(os.environ,PF_ENGINE_DATA_DIR=str(data),CGO_ENABLED='1',GOOS='darwin',GOARCH='arm64')
env['CC']=subprocess.check_output(['xcrun','--find','clang'],text=True).strip()
sdk=subprocess.check_output(['xcrun','--sdk','macosx','--show-sdk-path'],text=True).strip()
env.update(SDKROOT=sdk,CGO_CFLAGS=f'-isysroot {sdk} -arch arm64 -mmacosx-version-min=13.0',
           CGO_LDFLAGS=f'-isysroot {sdk} -arch arm64 -mmacosx-version-min=13.0')
subprocess.run([str(root/'tools/build_macos.sh')],cwd=root,env=env,check=True)
with tempfile.TemporaryDirectory(prefix='pf-macos-original-gates-') as tmp:
    tmp=pathlib.Path(tmp)
    archive=tmp/'source.tar'
    with archive.open('wb') as out: subprocess.run(['git','archive','HEAD'],cwd=root,stdout=out,check=True)
    with tarfile.open(archive) as source: source.extractall(tmp,filter='data')
    for name in names: (tmp/name).symlink_to(data/name)
    if a.reference:
        (tmp/'reference').mkdir(exist_ok=True)
        (tmp/'reference/original-dos-source').symlink_to(a.reference.resolve(strict=True),target_is_directory=True)
    # Platform Go adapter remains Linux/Windows; all authoritative engine packages run here.
    packages=subprocess.check_output(['go','list','./...'],cwd=tmp,env=env,text=True).splitlines()
    packages=[pkg for pkg in packages if pkg not in ('pinballfantasies/internal/platform','pinballfantasies/cmd/pinballfantasies')]
    subprocess.run(['go','test','-p=1','-count=1','-v',*packages],cwd=tmp,env=env,check=True)
    subprocess.run(['go','run','./cmd/personalvalidate','-data-dir',str(data)],cwd=tmp,env=env,check=True)
subprocess.run(['python3',str(root/'tools/check_macos_bundle.py'),str(root/'release/macos/Pinball Fantasies.app'),
                '--originals',str(data)],cwd=root,check=True)
assert before=={name:hashlib.sha256((data/name).read_bytes()).hexdigest() for name in names},'original inputs were modified'
print('PASS local macOS original-backed ABI, four-table/all-scroll host journeys, shared Go gates, storage restart, public payload scan')
print('Physical window/input/cursor/audio/fullscreen/import and real restart acceptance still require the docs/macos.md checklist.')
