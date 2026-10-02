#!/usr/bin/env python3
"""Locate literal records in owner-supplied retail PRGs; export addresses only.

Used by private reference generation. Nothing here executes original code.
The public runtime implements the same bounded record reads in Go.
"""
import copy
import struct
from pathlib import Path

DS = {1: 0x19d40, 2: 0x18ee0, 3: 0x18b60, 4: 0x166d0}
ANIM = {1: 0x20750, 2: 0x1f8e0, 3: 0x41ac0, 4: 0x1daf0}


def locate(content, root=Path('.')):
    out = copy.deepcopy(content)
    for key, c in out.items():
        table = int(key)
        b = (root / f'TABLE{table}.PRG').read_bytes()
        refs, exceptions = {}, {}
        for label, values in c.pop('texts').items():
            raw = bytes(values)
            if not raw:
                refs[label] = [0, 0]
                continue
            at = b.find(raw, DS[table], DS[table] + 65536)
            if at < 0:
                at = b.find(raw)
            if at < 0:
                # Unlinked source filename declarations and one unused legacy
                # jingle differ from retail. Keep explicit, reviewable exceptions.
                exceptions[label] = values
            else:
                refs[label] = [at, len(raw)]
        c['texts'], c['text_refs'] = exceptions, refs
        refs = {}
        for label, a in c.pop('animations').items():
            raw = struct.pack('<2H', *a['header'][1:]) + b''.join(
                struct.pack('<2H', p - ANIM[table], d)
                for p, d in zip(a['offsets'], a['durations']))
            at = b.find(raw)
            assert at >= 0, (table, label)
            prefix = a['header'][0] != 0
            if prefix:
                assert b[at-2:at] == struct.pack('<H', a['header'][0])
            refs[label] = [at, len(a['durations']), ANIM[table], int(prefix)]
        c['animation_refs'] = refs
        rows = c.pop('lampflash')
        first = struct.pack('<4H', *rows[0])
        hits = []
        at = b.find(first)
        while at >= 0:
            start = at - 2
            if start >= 0 and all(b[start+i*10+2:start+i*10+10] == struct.pack('<4H', *r)
                                  for i, r in enumerate(rows)):
                hits.append(start)
            at = b.find(first, at+1)
        assert len(hits) == 1, (table, hits)
        c['lamp_flash_ref'] = [hits[0], len(rows)]
    return out


def hydrate(content, root):
    """Private checker view, including all original text and numeric records."""
    out = copy.deepcopy(content)
    for key, c in out.items():
        b = (root / f'TABLE{key}.PRG').read_bytes()
        for label, (at, size) in c.pop('text_refs', {}).items():
            assert 0 <= at <= len(b) and 0 <= size <= len(b)-at
            c['texts'][label] = list(b[at:at+size])
        if 'animation_refs' in c:
            c['animations'] = {}
            for label, (at, count, base, prefix) in c.pop('animation_refs').items():
                assert at >= 2 and at+4+4*count <= len(b)
                h = [struct.unpack_from('<H', b, at-2)[0] if prefix else 0]
                h += list(struct.unpack_from('<2H', b, at))
                frames = [struct.unpack_from('<2H', b, at+4+4*i) for i in range(count)]
                c['animations'][label] = {'header': h, 'durations': [x[1] for x in frames],
                                          'offsets': [base+x[0] for x in frames]}
        if 'lamp_flash_ref' in c:
            at, count = c.pop('lamp_flash_ref')
            assert at >= 0 and at+count*10 <= len(b)
            c['lampflash'] = [list(struct.unpack_from('<4H', b, at+i*10+2)) for i in range(count)]
    return out


def locate_table_records(table, content, root=Path('.')):
    """Remove jingle values, DAC packets and gate masks from table exports."""
    c = copy.deepcopy(content)
    c['jingles'] = {k: {} for k in c['jingles']}
    c.pop('traces', None)
    c.pop('ball_trace', None)
    if table not in (3, 4):
        c['animations'] = {k: {} for k in c['animations']}
        used = {cmd['args'][0] for cmd in c['commands'] if cmd['op'] == '_SCROLL'}
        c['scrolls'] = {k: 0 for k in c['scrolls'] if k in used}
        c.pop('attract', None)
        if table == 2:
            b = (root / 'TABLE2.PRG').read_bytes()
            for label, lamp in c['lamps'].items():
                at = lamp['offset'] - 2
                assert b[at] == lamp['start'] and b[at+1] == lamp['count']
                c['lamps'][label] = {'offset': at}
        return c
    b = (root / f'TABLE{table}.PRG').read_bytes()
    for label, lamp in c['lamps'].items():
        raw = bytes([lamp['start'], len(lamp['rgb'])//3] + lamp['rgb'])
        at = b.find(raw, DS[table]); assert at >= 0
        c['lamps'][label] = {'ref': [at, len(raw)]}
    for gate in c['gates']:
        w, h = gate['width'], gate['height']
        for name in ('opened', 'closed'):
            raw = bytes(gate.pop(name))
            at = b.find(raw, DS[table])
            if at >= 0:
                ref = [at, len(raw), 1, len(raw)]
            elif table == 3:
                offsets = {7: (0x6390, 0x6480), 8: (0x6570, 0x6620)}
                at = DS[table]+offsets[gate['number']][name == 'closed']+w
                ref = [at, w, h, 3*w]
            else:
                sig = struct.pack('<3H', gate['y']*40+gate['x'], w, h)
                at = b.index(sig, DS[table], 0x1da30)-4
                pos, neg = struct.unpack_from('<2H', b, at)
                ref = [DS[table]+(pos if name == 'closed' else neg)+w,w,h,3*w]
            at, width, count, stride = ref
            assert b''.join(b[at+i*stride:at+i*stride+width] for i in range(count)) == raw
            gate[name+'_ref'] = ref
    if table == 4:
        refs, handlers = {}, {}
        for name, rows in c.pop('areas').items():
            hits = [at for at in range(DS[4], 0x1cec0-len(rows)*10)
                    if all(b[at+i*10:at+i*10+8] == struct.pack('<4h', *row['rect'])
                           for i, row in enumerate(rows))
                    and b[at+len(rows)*10:at+len(rows)*10+2] == b'\0\0']
            assert len(hits) == 1, (name, hits)
            at = hits[0]
            refs[name] = [at, len(rows)]
            for i, row in enumerate(rows):
                pointer = struct.unpack_from('<H', b, at+i*10+8)[0]
                assert handlers.get(pointer, row['handler']) == row['handler']
                handlers[pointer] = row['handler']
        c['area_refs'], c['area_handlers'] = refs, handlers
    return c


def score_glyphs(table, b):
    """Private independent glyph hash generation from literal drawing stores."""
    ds = DS[table]
    base = {1:0xae60,2:0xa650,3:0xa0f0,4:0xb600}[table]
    shift = {1:0,2:0x90,3:-0x720,4:0x5f0}[table]
    out = {}
    for char in range(48,58):
        rows = bytearray(16)
        for tab,code,parity in [(0x5a00+shift,base+0x1a0,0),(0x5c00+shift,base+0xc40,1)]:
            q = code+struct.unpack_from('<H',b,ds+tab+2*char)[0]
            while b[q] != 0xc3:
                assert b[q] == 0x88 and b[q+1] in (0x87,0xa7)
                disp = struct.unpack_from('<H',b,q+2)[0]
                x,y = (disp%84)*2+parity,disp//168
                if x<8 and y<16 and b[q+1]==0x87:
                    rows[y] |= 128 >> x
                q += 4
        out[str(char)] = list(rows)
    return out
