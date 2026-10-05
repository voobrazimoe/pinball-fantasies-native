#!/usr/bin/env python3
"""Public bundle allowlist and commercial-payload hygiene, before artifact upload."""
import argparse, hashlib, json, pathlib, plistlib, re, sys
p=argparse.ArgumentParser(); p.add_argument('app',type=pathlib.Path); p.add_argument('--originals',type=pathlib.Path); a=p.parse_args()
root=pathlib.Path(__file__).resolve().parent.parent
allowed={'Contents/Info.plist','Contents/MacOS/pinballfantasies','Contents/Resources/LICENSE.txt',
         'Contents/Resources/pf-icon.icns','Contents/Resources/Go-version.txt','Contents/Resources/Go-LICENSE.txt','Contents/_CodeSignature/CodeResources'}
files={str(f.relative_to(a.app)):f for f in a.app.rglob('*') if f.is_file()}
assert set(files)==allowed,(set(files)-allowed,allowed-set(files))
assert not any(f.is_symlink() for f in a.app.rglob('*'))
info=plistlib.loads(files['Contents/Info.plist'].read_bytes())
assert info['CFBundleExecutable']=='pinballfantasies' and info['CFBundlePackageType']=='APPL'
assert info['CFBundleIconFile']=='pf-icon.icns'
assert files['Contents/Resources/pf-icon.icns'].read_bytes()==(root/'art/app-icon/pf-icon.icns').read_bytes()
assert files['Contents/MacOS/pinballfantasies'].stat().st_mode & 0o111
assert not re.search(rb'/(?:Users|home)/', files['Contents/MacOS/pinballfantasies'].read_bytes()), 'local build path in public executable'
assert files['Contents/Resources/LICENSE.txt'].read_bytes()==(root/'LICENSE').read_bytes()
originals=json.loads((root/'analysis/game-inventory.json').read_text())
hashes={record['sha256'] for record in originals}
for name,path in files.items():
    assert hashlib.sha256(path.read_bytes()).hexdigest() not in hashes,name
if a.originals:
    blocks=set()
    for source in a.originals.iterdir():
        if source.suffix in ('.PRG','.MOD'):
            data=source.read_bytes()
            blocks.update(data[i:i+4096] for i in range(0,len(data)-4095,4096) if len(set(data[i:i+4096]))>16)
    for name,path in files.items():
        assert not any(block in path.read_bytes() for block in blocks),name
    print(f'PASS {len(blocks)} nontrivial original 4KiB blocks absent from public bundle')
print('PASS macOS public bundle structure, exact file allowlist, no original file hashes/payload paths')
