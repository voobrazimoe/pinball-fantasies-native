#!/usr/bin/env python3
"""Dependency-free icon/package gate. Optional --exe verifies linked PE resources."""
import argparse
from pathlib import Path
import plistlib
import struct
import xml.etree.ElementTree as ET
import zlib

ROOT = Path(__file__).resolve().parent.parent
ART = ROOT/'art/app-icon'
RES = ROOT/'hosts/android/app/src/main/res'
A = '{http://schemas.android.com/apk/res/android}'

def png_size(path):
    raw=path.read_bytes()
    assert raw[:8]==b'\x89PNG\r\n\x1a\n',path
    return struct.unpack_from('>II',raw,16)

def rgba(path):
    """Decode noninterlaced RGBA PNG for the adaptive safe-circle gate."""
    raw=path.read_bytes(); width,height=png_size(path)
    assert raw[24:29]==bytes([8,6,0,0,0]),'expected RGBA8 PNG'
    offset=8; compressed=b''
    while offset<len(raw):
        size=struct.unpack_from('>I',raw,offset)[0]
        if raw[offset+4:offset+8]==b'IDAT': compressed+=raw[offset+8:offset+8+size]
        offset+=size+12
    data=zlib.decompress(compressed); stride=width*4; previous=bytearray(stride)
    for y in range(height):
        offset=y*(stride+1); kind=data[offset]; row=bytearray(data[offset+1:offset+1+stride])
        for i in range(stride):
            left=row[i-4] if i>=4 else 0; up=previous[i]; corner=previous[i-4] if i>=4 else 0
            if kind==1: predictor=left
            elif kind==2: predictor=up
            elif kind==3: predictor=(left+up)//2
            elif kind==4:
                p=left+up-corner; dl,du,dc=abs(p-left),abs(p-up),abs(p-corner)
                predictor=left if dl<=du and dl<=dc else up if du<=dc else corner
            else: assert kind==0; predictor=0
            row[i]=(row[i]+predictor)&255
        yield y,row
        previous=row

def ico_entries():
    data=(ART/'pf-icon.ico').read_bytes(); reserved,kind,count=struct.unpack_from('<HHH',data)
    assert (reserved,kind)==(0,1)
    entries=[]
    for i in range(count):
        w,h,_,_,planes,bpp,length,offset=struct.unpack_from('<BBBBHHII',data,6+16*i)
        entries.append((w or 256,h or 256,data[offset:offset+length]))
    assert {w for w,h,_ in entries}=={16,24,32,48,64,128,256}
    return entries

def check_exe(path):
    raw=path.read_bytes(); pe=struct.unpack_from('<I',raw,0x3c)[0]
    assert raw[pe:pe+4]==b'PE\0\0'
    machine,sections=struct.unpack_from('<HH',raw,pe+4); assert machine==0x8664
    optional_size=struct.unpack_from('<H',raw,pe+20)[0]
    optional=pe+24; assert struct.unpack_from('<H',raw,optional)[0]==0x20b
    resource_rva,resource_size=struct.unpack_from('<II',raw,optional+112+2*8)
    assert resource_rva and resource_size,'EXE has no resource table'
    section_start=optional+optional_size
    def file_offset(rva):
        for i in range(sections):
            off=section_start+40*i
            vsize,va,size,pointer=struct.unpack_from('<IIII',raw,off+8)
            if va<=rva<va+max(vsize,size): return pointer+rva-va
        raise AssertionError('invalid resource RVA')
    base=file_offset(resource_rva); leaves={}
    def walk(offset,keys=()):
        names,ids=struct.unpack_from('<HH',raw,base+offset+12)
        assert names==0,'unexpected named icon resource'
        for i in range(ids):
            name,child=struct.unpack_from('<II',raw,base+offset+16+i*8)
            if child&0x80000000: walk(child&0x7fffffff,keys+(name,))
            else:
                rva,size=struct.unpack_from('<II',raw,base+child)
                location=file_offset(rva);leaves[keys+(name,)]=raw[location:location+size]
    walk(0)
    entries=ico_entries(); icons=[v for k,v in leaves.items() if k[0]==3]
    assert len(icons)==len(entries)
    assert sorted(icons)==sorted(data for _,_,data in entries),'EXE icon differs from canonical export'
    groups=[v for k,v in leaves.items() if k[0]==14 and k[1]==1]
    assert len(groups)==1 and struct.unpack_from('<HHH',groups[0])==(0,1,len(entries))
    print('PASS Windows AMD64 EXE: seven exact PF RT_ICON images and RT_GROUP_ICON #1')

def check():
    for name in ('master','foreground','background'):
        svg=ET.parse(ART/f'pf-icon-{name}.svg').getroot()
        for element in svg.iter():
            assert element.tag.split('}')[-1] in {'svg','title','g','rect','path','circle'}
            assert not any('href' in k for k in element.attrib)
    for size in (16,24,32,48,64,128,256,512,1024):
        assert png_size(ART/f'png/{size}/pinballfantasies.png')==(size,size)
    ico_entries()
    icns=(ART/'pf-icon.icns').read_bytes();assert icns[:4]==b'icns'
    assert struct.unpack_from('>I',icns,4)[0]==len(icns)
    chunks=set();offset=8
    while offset<len(icns):
        kind,length=struct.unpack_from('>4sI',icns,offset);assert length>=8
        chunks.add(kind);offset+=length
    assert {b'ic07',b'ic08',b'ic09',b'ic10'}<=chunks
    info=plistlib.loads((ROOT/'hosts/macos/Info.plist').read_bytes())
    assert info['CFBundleIconFile']=='pf-icon.icns'
    assert 'cp art/app-icon/pf-icon.icns' in (ROOT/'tools/build_macos.sh').read_text()
    assert (ROOT/'internal/platform/pficon_windows_amd64.syso').read_bytes()[:2]==b'\x64\x86'
    manifest=ET.parse(ROOT/'hosts/android/app/src/main/AndroidManifest.xml').getroot().find('application')
    assert manifest.get(A+'icon')=='@mipmap/ic_launcher'
    assert manifest.get(A+'roundIcon')=='@mipmap/ic_launcher_round'
    for version in ('v26','v33'):
        for name in ('ic_launcher','ic_launcher_round'):
            icon=ET.parse(RES/f'mipmap-anydpi-{version}/{name}.xml').getroot()
            assert icon.tag=='adaptive-icon'
            for element in icon:
                ref=element.get(A+'drawable');assert ref.startswith('@drawable/')
                assert (RES/'drawable'/f'{ref.split("/")[1]}.xml').exists()
            assert {e.tag for e in icon}==({'background','foreground'}|({'monochrome'} if version=='v33' else set()))
    for density,size in [('mdpi',48),('hdpi',72),('xhdpi',96),('xxhdpi',144),('xxxhdpi',192)]:
        for name in ('ic_launcher','ic_launcher_round'):
            assert png_size(RES/f'mipmap-{density}/{name}.png')==(size,size)
    # Entire adaptive mark, including orbit and ball, fits the inscribed safe
    # circle: circle, rounded-square and squircle masks cannot clip it.
    radius=1024*33/108
    for y,row in rgba(ART/'png/foreground.png'):
        for x in range(1024):
            if row[x*4+3]>8:
                assert (x+.5-512)**2+(y+.5-512)**2<=radius**2,(x,y,'outside adaptive safe zone')
    appimage=(ROOT/'tools/build_appimage.py').read_text()
    for text in ('Icon=pinballfantasies','pinballfantasies.png',"appdir/'.DirIcon'",'usr/share/icons/hicolor'):
        assert text in appimage
    print('PASS SVG provenance structure, raster/ICO/ICNS sizes, adaptive safe zone and platform wiring')

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('--exe',type=Path)
    args=parser.parse_args();check()
    if args.exe: check_exe(args.exe)
