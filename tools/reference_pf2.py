#!/usr/bin/env python3
"""Independent static PF2 composition from inspected content locations.
No DOS/x86 execution, instruction interpretation, physics or Go invocation.
"""
import hashlib
import json
from pathlib import Path
from reference_extract import extract

def digest(data):
    return hashlib.sha256(data).hexdigest()

def reference():
    data = Path('TABLE1.PRG').read_bytes()
    baseline, rgba = extract('TABLE1.PRG')
    assert baseline['file_sha256'] == '4d7a69e7dc95260ad2541c6981a11ab842e2f1f20e45447e5613688b86e38414'
    offsets = json.loads(Path('analysis/pf2-ball-locations.json').read_text())['offsets']
    ball = bytes(data[o] if o else 0 for o in offsets)
    mask = data[0x300b0:0x300b0+40*576]
    assert len(ball) == 256 and len(mask) == 23040
    full = bytearray(rgba)
    # Read the original palette directly, independently of Go's framebuffer.
    base = 437168
    cmap = data.index(b'CMAP', base)
    palette = data[cmap+8:cmap+8+768]
    visible, masked, hidden = 0, 0, 0
    for i, color in enumerate(ball):
        if not color:
            continue
        dx, dy = i % 16, i // 16
        x, y = 297+dx, 530+dy
        if mask[y*40+x//8] & (128 >> (x%8)):
            masked += 1
        else:
            full[(y*320+x)*4:(y*320+x)*4+3] = palette[color*3:color*3+3]
            visible += 1
        x = 282+dx
        hidden += bool(mask[y*40+x//8] & (128 >> (x%8)))
    frame = full[259*320*4:576*320*4]
    return dict(viewport=[0,259,320,576], ball_origin=[297,530],
                hidden_ball_origin=[282,530], ball_size=16,
                ball_sha256=digest(ball), foreground_sha256=digest(mask),
                sprite_pixels=sum(bool(c) for c in ball), visible_pixels=visible,
                masked_pixels=masked, earlier_hidden_pixels=hidden,
                putball_sha256=digest(data[0x98a0:0xa550]),
                rgba_sha256=digest(frame), pf1_rgba_sha256=baseline['rgba_sha256'])

if __name__ == '__main__':
    print(json.dumps(reference(), indent=2))
