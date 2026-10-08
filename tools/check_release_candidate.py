#!/usr/bin/env python3
"""Verify a complete public release candidate before owner package smoke.

Requires dissect.squashfs (pip install dissect.squashfs==1.12).
No original data is needed; --originals adds an owner-local byte-block scan.
"""
import argparse
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import plistlib
import re
import struct
import subprocess
import sys
import zipfile

ROOT=Path(__file__).resolve().parent.parent
VERSION,VERSION_CODE='0.1.4','4'
sys.path.insert(0,str(ROOT/'tools'))
from check_app_icons import check_exe
from bundled_demo import DEMO,demo_bytes
from dissect.squashfs import SquashFS

NAMES=['pinballfantasies.exe','PinballFantasies-x86_64.AppImage',
       'PinballFantasies-arm64.zip','PinballFantasies-x86_64.zip','PinballFantasies-android.apk']
GATES={'Android host','Android A0','Asset-free source','macOS native hosts','Desktop release candidate artifacts'}
FORBIDDEN=re.compile(r'\.(?:prg|mod|cfg|hi)$',re.I)


def bundled_demo(name):
    """The shipped 10-minute demo is the only original data a package may carry."""
    path=PurePosixPath(name)
    return path.name in DEMO and path.parent.name in ('demo','Demo')


def verify(directory,source_sha,evidence,sdk,originals=None):
    assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip()==source_sha
    assert not subprocess.check_output(['git','diff','HEAD','--name-only'],cwd=ROOT,text=True).strip(),'tracked source is dirty'
    runs=json.loads(evidence.read_text())
    assert {r['name'] for r in runs}==GATES
    assert all(r['sha']==source_sha and r['status']=='completed' and r['conclusion']=='success' for r in runs),'CI gate not green at candidate SHA'
    inventory=json.loads((ROOT/'analysis/game-inventory.json').read_text())
    commercial_hashes={r['sha256'] for r in inventory}
    demo=demo_bytes()
    blocks=set()
    if originals:
        for path in originals.iterdir():
            if path.suffix.upper() in ('.PRG','.MOD'):
                data=path.read_bytes()
                blocks.update(data[i:i+4096] for i in range(0,len(data)-4095,4096) if len(set(data[i:i+4096]))>16)
        # Retail blocks that the bundled demo shares are expected in every package.
        blocks={b for b in blocks if not any(b in d for d in demo.values())}
    inspected=0
    def scan(name,data):
        nonlocal inspected
        if bundled_demo(name):
            assert demo[PurePosixPath(name).name]==data,('altered bundled demo',name)
            inspected+=1
            return
        assert hashlib.sha256(data).hexdigest() not in commercial_hashes,('commercial file',name)
        assert not any(block in data for block in blocks),('commercial payload',name)
        assert not re.search(rb'/Users/mess(?:/|\x00)|/home/mess(?:/|\x00)|m[e]ss@mini',data),('private build path',name)
        inspected+=1
    def safe_path(name):
        path=PurePosixPath(name)
        assert not path.is_absolute() and '..' not in path.parts,name
        assert bundled_demo(name) or not FORBIDDEN.search(name),('commercial filename',name)
        assert not any(p in {'personal-data','.personal-assets'} for p in path.parts),name
    for name in subprocess.check_output(['git','ls-files'],cwd=ROOT,text=True).splitlines():
        if name!='go.mod': safe_path(name)
        scan('source/'+name,(ROOT/name).read_bytes())
    for name in NAMES:
        assert (directory/name).is_file(),('missing candidate',name)
        scan(name,(directory/name).read_bytes())
    check_exe(directory/NAMES[0])
    for name,architecture in ((NAMES[2],0x100000c),(NAMES[3],0x1000007)):
        with zipfile.ZipFile(directory/name) as archive:
            members={i.filename:archive.read(i) for i in archive.infolist() if not i.is_dir()}
            for member,data in members.items(): safe_path(member);scan(name+'/'+member,data)
            app='Pinball Fantasies.app/Contents/'
            info=plistlib.loads(members[app+'Info.plist'])
            assert info['CFBundleShortVersionString']==VERSION
            assert info['CFBundleIconFile']=='pf-icon.icns'
            assert source_sha.startswith(info['PFHostBuild'])
            assert members[app+'Resources/pf-icon.icns']==(ROOT/'art/app-icon/pf-icon.icns').read_bytes()
            executable=members[app+'MacOS/pinballfantasies']
            assert executable[:4]==b'\xcf\xfa\xed\xfe'
            assert struct.unpack_from('<I',executable,4)[0]==architecture
            assert app+'_CodeSignature/CodeResources' in members
    raw=(directory/NAMES[1]).read_bytes()
    for match in re.finditer(b'hsqs',raw):
        try: fs=SquashFS(io.BytesIO(raw[match.start():]));break
        except (ValueError,NotImplementedError): continue
    else: raise AssertionError('AppImage has no readable SquashFS')
    def walk(node,prefix=''):
        for child in node.iterdir():
            name=(prefix+'/'+child.name).lstrip('/');safe_path(name)
            if child.is_dir():walk(child,name)
            elif child.is_file():scan('AppImage/'+name,child.open().read())
            elif child.is_symlink():safe_path(child.link)
            else:raise AssertionError(('unexpected AppImage inode',name))
    walk(fs.root)
    assert fs.get('pinballfantasies.png').open().read()==(ROOT/'art/app-icon/png/256/pinballfantasies.png').read_bytes()
    assert fs.get('.DirIcon').link=='pinballfantasies.png'
    assert b'Icon=pinballfantasies\n' in fs.get('pinballfantasies.desktop').open().read()
    binary=fs.get('usr/bin/pinballfantasies').open().read()
    assert binary[:4]==b'\x7fELF' and struct.unpack_from('<H',binary,18)[0]==62
    manifest=json.loads(fs.get('BUILD-MANIFEST.json').open().read())
    assert manifest['mode']=='public' and manifest['original_inputs']==[]
    for size in (16,24,32,48,64,128,256,512):
        assert fs.get(f'usr/share/icons/hicolor/{size}x{size}/apps/pinballfantasies.png').open().read()==(ROOT/f'art/app-icon/png/{size}/pinballfantasies.png').read_bytes()
    libraries=[]
    apk_raw=(directory/NAMES[4]).read_bytes()
    with zipfile.ZipFile(directory/NAMES[4]) as archive:
        for member in archive.infolist():
            safe_path(member.filename)
            if member.is_dir(): continue
            data=archive.read(member);scan('APK/'+member.filename,data)
            if member.filename.startswith('lib/') and member.filename.endswith('.so'):
                _,abi,library=member.filename.split('/')
                assert abi in {'arm64-v8a','x86_64'}
                assert data[:6]==b'\x7fELF\x02\x01'
                assert struct.unpack_from('<H',data,18)[0]==(183 if abi=='arm64-v8a' else 62)
                phoff=struct.unpack_from('<Q',data,32)[0];size,count=struct.unpack_from('<HH',data,54)
                loads=0
                for i in range(count):
                    offset=phoff+i*size
                    if struct.unpack_from('<I',data,offset)[0]==1:
                        align=struct.unpack_from('<Q',data,offset+48)[0]
                        assert align>=16384 and align&(align-1)==0,(member.filename,'ELF alignment')
                        loads+=1
                assert loads
                assert member.compress_type==zipfile.ZIP_STORED
                name_length,extra_length=struct.unpack_from('<HH',apk_raw,member.header_offset+26)
                assert (member.header_offset+30+name_length+extra_length)%16384==0,(member.filename,'ZIP alignment')
                libraries.append(member.filename)
        for abi in ('arm64-v8a','x86_64'):
            for lib in ('libpfengine.so','libpinball_android.so'):
                assert f'lib/{abi}/{lib}' in libraries
        assert 'resources.arsc' in archive.namelist()
    buildtools=sdk/'build-tools/36.0.0'
    apk=directory/NAMES[4]
    badging=subprocess.check_output([str(buildtools/'aapt2'),'dump','badging',str(apk)],text=True)
    assert f"versionName='{VERSION}'" in badging and f"versionCode='{VERSION_CODE}'" in badging
    assert "name='io.github.voobrazimoe.pinballfantasies'" in badging
    assert "application-label:'Pinball Fantasies'" in badging
    assert 'ic_launcher' in badging and 'Personal' not in badging
    subprocess.run([str(buildtools/'zipalign'),'-c','-P','16','4',str(apk)],check=True)
    signature=subprocess.check_output([str(buildtools/'apksigner'),'verify','--verbose','--print-certs',str(apk)],text=True)
    assert 'Verified using v2 scheme (APK Signature Scheme v2): true' in signature
    result={'source_sha':source_sha,'result':'PASS','ci':runs,'commercial_blocks_checked':len(blocks),
            'source_and_package_members_scanned':inspected,'android_libraries':libraries,'android_signature':signature.splitlines(),
            'artifacts':{name:hashlib.sha256((directory/name).read_bytes()).hexdigest() for name in NAMES},
            'physical_rc_smoke':'PENDING OWNER: launcher masks/icon, first-run/SAF/play/touch/smoothness; desktop icon/launch/gameplay'}
    return result

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory',type=Path);parser.add_argument('--source-sha',required=True)
    parser.add_argument('--ci-evidence',required=True,type=Path);parser.add_argument('--originals',type=Path)
    parser.add_argument('--android-sdk',required=True,type=Path)
    parser.add_argument('--report',type=Path)
    args=parser.parse_args()
    result=verify(args.directory,args.source_sha,args.ci_evidence,args.android_sdk,args.originals)
    report=args.report or args.directory/'validation.json'
    report.parent.mkdir(parents=True,exist_ok=True)
    report.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))
