#!/usr/bin/env python3
"""Export original PF SVG to platform icons. Requires Node/sharp and Pillow."""
import copy
import os
from pathlib import Path
import subprocess
import xml.etree.ElementTree as ET
from PIL import Image

ROOT = Path(__file__).resolve().parent.parent
ART = ROOT / 'art/app-icon'
RES = ROOT / 'hosts/android/app/src/main/res'
SVG = '{http://www.w3.org/2000/svg}'
ANDROID = 'http://schemas.android.com/apk/res/android'
ET.register_namespace('', SVG[1:-1])
ET.register_namespace('android', ANDROID)

def write_xml(path, element):
    path.parent.mkdir(parents=True, exist_ok=True)
    ET.indent(element)
    path.write_bytes(ET.tostring(element, encoding='utf-8', xml_declaration=True) + b'\n')

def render(source, dest, size):
    dest.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run([os.environ.get('PF_ICON_NODE', 'node'), str(ROOT/'tools/render_icon.mjs'),
                    str(source), str(dest), str(size)], check=True)

def export():
    master = ET.parse(ART/'pf-icon-master.svg').getroot()
    foreground = master.find(f"{SVG}g[@id='foreground']")
    # 90% scaling brings the entire mark (including orbit/ball) inside the
    # adaptive 66/108 diameter safe circle. Background is full bleed.
    fg = ET.Element(SVG+'svg', width='1024', height='1024', viewBox='0 0 1024 1024')
    group = copy.deepcopy(foreground)
    group.set('transform', 'translate(51.2 51.2) scale(0.9)')
    fg.append(group)
    write_xml(ART/'pf-icon-foreground.svg', fg)
    bg = ET.Element(SVG+'svg', width='1024', height='1024', viewBox='0 0 1024 1024')
    bg.append(ET.Element(SVG+'rect', width='1024', height='1024', fill='#182D42'))
    write_xml(ART/'pf-icon-background.svg', bg)
    for mono in (False, True):
        vector = ET.Element('vector', {f'{{{ANDROID}}}width':'108dp', f'{{{ANDROID}}}height':'108dp',
                                      f'{{{ANDROID}}}viewportWidth':'1024', f'{{{ANDROID}}}viewportHeight':'1024'})
        vg = ET.SubElement(vector, 'group', {f'{{{ANDROID}}}scaleX':'0.9', f'{{{ANDROID}}}scaleY':'0.9',
                                           f'{{{ANDROID}}}translateX':'51.2', f'{{{ANDROID}}}translateY':'51.2'})
        for element in foreground:
            attrs = element.attrib
            if element.tag == SVG+'circle':
                x, y, r = (float(attrs[k]) for k in ('cx','cy','r'))
                data = f'M{x-r},{y}a{r},{r} 0 1,0 {r*2},0a{r},{r} 0 1,0 {-r*2},0'
            else:
                data = attrs['d']
            out = {'pathData':data, 'fillColor': '#FFFFFF' if mono and attrs.get('fill') != 'none' else attrs.get('fill','#00000000')}
            if out['fillColor'] == 'none': out['fillColor'] = '#00000000'
            for key in ('stroke','stroke-width','stroke-linecap','fill-rule'):
                if key in attrs:
                    android_key = {'stroke':'strokeColor','stroke-width':'strokeWidth','stroke-linecap':'strokeLineCap','fill-rule':'fillType'}[key]
                    out[android_key] = '#FFFFFF' if mono and key=='stroke' else ('evenOdd' if key=='fill-rule' else attrs[key])
            ET.SubElement(vg,'path', {f'{{{ANDROID}}}{k}':v for k,v in out.items()})
        write_xml(RES/'drawable'/('ic_launcher_monochrome.xml' if mono else 'ic_launcher_foreground.xml'),vector)
    background=ET.Element('shape')
    ET.SubElement(background,'solid',{f'{{{ANDROID}}}color':'#182D42'})
    write_xml(RES/'drawable/ic_launcher_background.xml', background)
    for version in ('v26','v33'):
        for name in ('ic_launcher','ic_launcher_round'):
            adaptive=ET.Element('adaptive-icon')
            for key in ('background','foreground', *(['monochrome'] if version=='v33' else [])):
                ET.SubElement(adaptive,key,{f'{{{ANDROID}}}drawable':f'@drawable/ic_launcher_{key}'})
            write_xml(RES/f'mipmap-anydpi-{version}'/f'{name}.xml',adaptive)
    for density,size in [('mdpi',48),('hdpi',72),('xhdpi',96),('xxhdpi',144),('xxxhdpi',192)]:
        target=RES/f'mipmap-{density}/ic_launcher.png'
        render(ART/'pf-icon-master.svg',target,size)
        im=Image.open(target).convert('RGBA')
        # Round fallback uses the same identity, clipped to a circle.
        from PIL import ImageDraw
        mask=Image.new('L',im.size); ImageDraw.Draw(mask).ellipse((0,0,size-1,size-1),fill=255)
        im.putalpha(Image.composite(im.getchannel('A'),Image.new('L',im.size),mask))
        im.save(target.with_name('ic_launcher_round.png'))
    render(ART/'pf-icon-foreground.svg',ART/'png/foreground.png',1024)
    for size in (16,24,32,48,64,128,256,512,1024):
        render(ART/'pf-icon-master.svg',ART/f'png/{size}/pinballfantasies.png',size)
    image=Image.open(ART/'png/1024/pinballfantasies.png')
    image.save(ART/'pf-icon.ico', sizes=[(s,s) for s in (16,24,32,48,64,128,256)])
    image.save(ART/'pf-icon.icns', sizes=[(s,s) for s in (16,32,64,128,256,512,1024)])
    # Go's pinned resource compiler produces a deterministic AMD64 COFF object.
    subprocess.run([str(ROOT/'tools/go.sh'),'run','github.com/akavel/rsrc@v0.10.2',
                    '-ico',str(ART/'pf-icon.ico'),'-o',str(ROOT/'internal/platform/pficon_windows_amd64.syso')],check=True,cwd=ROOT)
    print('Exported PF icons; run python3 tools/check_app_icons.py')

if __name__ == '__main__': export()
