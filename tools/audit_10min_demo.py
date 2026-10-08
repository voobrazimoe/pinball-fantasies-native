#!/usr/bin/env python3
"""Read-only DMO0 evidence. Exports metadata, never PRG/MOD/code/image bytes.

This is NOT a production descriptor generator or a compatibility validator.
Exact-file identity pins this research; production support is still gated by
the unresolved callback/control-flow proof in the validation report.
"""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import struct

ROOT = Path(__file__).resolve().parent.parent
FILES = {
    'INTRO.PRG': (347054, '05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3'),
    'INTRO.MOD': (252870, 'f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613'),
    'MOD2.MOD': (55394, 'aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21'),
    'TABLE1.PRG': (537190, '44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3'),
    'TABLE1.MOD': (210760, 'a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5'),
}
FINGERPRINT = 'e7d9aaedf4f06f67d1553d88be0b0f87568bb1da49344b1a7f070f37c160809e'
TABLE_DS = 0x19db0
INTRO_OFFSETS = [0x1cbd0, 0x6c30, 0x8bc0, 0x204f0, 0x24b10, 0x29910,
                 0x2e2c0, 0x3bd70, 0x431b0, 0x46e90, 0x4e000, 0x120e0,
                 0x17fe0, 0x10f20, 0xa510, 0x35360, 0x339e0]
INTRO_ROLES = ['Logo', 'Font', 'MonoFont', 'PartyLandCard',
               'UnavailableSpeedDevilsCard', 'UnavailableGameshowCard',
               'UnavailableStonesCard'] + [f'Startup{i}' for i in range(8)] + ['HighLogo', 'HighMono']


def digest(b):
    return hashlib.sha256(b).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def bounded(b, at, size):
    require(at >= 0 and size >= 0 and at + size <= len(b), 'truncated research record')
    return b[at:at+size]


def form(b, at):
    require(bounded(b, at, 4) == b'FORM', 'FORM anchor')
    size = 8 + struct.unpack_from('>I', bounded(b, at+4, 4))[0]
    raw = bounded(b, at, size)
    pos, header = 12, None
    while pos < size:
        tag = bounded(raw, pos, 4)
        n = struct.unpack('>I', bounded(raw, pos+4, 4))[0]
        data = bounded(raw, pos+8, n)
        if tag == b'BMHD':
            require(header is None and len(data) == 20, 'BMHD')
            header = data
        pos += 8+n+(n & 1)
    require(pos == size and header is not None, 'FORM extent')
    w, h = struct.unpack_from('>2H', header)
    return dict(offset=at, size=size, kind=raw[8:12].decode('ascii'),
                width=w, height=h, planes=header[8], sha256=digest(raw))


def audit(data, canonical, evidence):
    src = {}
    for name, (size, sha) in FILES.items():
        b = (data/name).read_bytes()
        require(len(b) == size and digest(b) == sha, f'not the pinned research input: {name}')
        src[name] = b
    identity = ''.join(n+'\0'+digest(src[n])+'\n' for n in FILES)
    require(digest(identity.encode()) == FINGERPRINT, 'runtime fingerprint')
    out = dict(status='DMO0 NOT CLOSED; not production metadata',
               fingerprint=FINGERPRINT,
               files=[dict(name=n, size=len(src[n]), sha256=digest(src[n])) for n in FILES])
    inventory = json.loads((ROOT/'analysis/game-inventory.json').read_text())
    ref = {}
    for name in ('INTRO.PRG', 'TABLE1.PRG'):
        ref[name] = (canonical/name).read_bytes()
        expected = next(r for r in inventory if r['name'] == name)
        require(digest(ref[name]) == expected['sha256'], f'noncanonical A: {name}')
    intro, table = src['INTRO.PRG'], src['TABLE1.PRG']
    out['identity_markers'] = dict(intro_10min_offset=intro.find(b'10 MINUTE DEMO'))
    require(out['identity_markers']['intro_10min_offset'] == 0x582b, '10-minute data marker')
    # Launcher/calibrator are static research inputs, not required runtime roles.
    out['support_research_inputs'] = []
    for name, size, sha in [
        ('PINBALL.EXE', 2063, '7acc8be42f23cc56a66ce8839a561c4d7318025dc395935be44b7a760f538dff'),
        ('TIMER.BIN', 253, '783f88891a760b3fab648a7ae64ad8998e1349737df07a1e765329f7a6787d0c'),
    ]:
        b = (data/name).read_bytes()
        require(len(b) == size and digest(b) == sha, 'support research identity: '+name)
        out['support_research_inputs'].append(dict(name=name, size=size, sha256=sha))
        if name == 'PINBALL.EXE':
            at = b.find(b'You have been playing a 10 minute demo of Pinball Fantasies.')
            require(at == 0x43f, 'launcher 10-minute marker')
            out['identity_markers']['launcher_10min_offset'] = at
    profiles = json.loads((ROOT/'internal/datalayout/profiles.json').read_text())
    out['intro_forms'] = []
    for at, role, old in zip(INTRO_OFFSETS, INTRO_ROLES, profiles['INTRO.PRG']['pictures']):
        pic = form(intro, at)
        old_pic = form(ref['INTRO.PRG'], old['offset'])
        pic.update(role=role, canonical_offset=old['offset'],
                   canonical_form_equal=pic['sha256'] == old_pic['sha256'])
        out['intro_forms'].append(pic)
    require(sum(p['canonical_form_equal'] for p in out['intro_forms']) == 14, 'INTRO relations')
    out['jingles'] = []
    for role, dest, at, want in [('S_EMPTY', 0x1a9ac, 0x1aa38, (62, 0, 0)),
                               ('S_GAMEOVER2', 0x1a9c1, 0x1aa4d, (13, 0, 255))]:
        fields = tuple(bounded(table, at, 3))
        require(fields == want, role)
        out['jingles'].append(dict(role=role, source=at, canonical_offset=dest,
                                   position=fields[0], repeat=fields[1], priority=fields[2]))
    out['materials'] = []
    for i in range(8):
        dest, at = 0x1c05d+16*i, 0x1c1c7+16*i
        require(bounded(table, at, 10) == bounded(ref['TABLE1.PRG'], dest, 10), 'material difference')
        out['materials'].append(dict(index=i, source=at, canonical_offset=dest,
                                     shift=362, equivalent=True))
    require(struct.unpack_from('<HB', table, TABLE_DS+0x34cd) == (0, 0), 'timer initial state')
    out['timer'] = dict(counter_offset=TABLE_DS+0x34cd, expired_flag_offset=TABLE_DS+0x34cf,
                        increment_site=0x5cdb, comparison_site=0x5cdf,
                        threshold=struct.unpack_from('<H', table, 0x5ce3)[0],
                        caller=0x4723, electronics_entry=0x5cd9,
                        script_offset=0x1ba17, callback_cadence_proof='OPEN')
    require(out['timer']['threshold'] == 35998, 'timer boundary')
    # Hashes bind manually reviewed code slices without exporting instructions.
    out['reviewed_code_ranges'] = []
    for start, size, sha, meaning in [
        (0x5cd9, 70, '016ac291fea17c79de24a8d53cb16731872e3812696c13012d2399bc745dfe3a', 'electronics timer'),
        (0x73e, 78, 'a00a14fa364a1328f45097bdf7b6b3d62973aaf1074efd7253b4f9f07c2f5557', 'demo ball continuation'),
        (0x1ba17, 42, '14f0a23672396dad451a474a61b9035afd3d5c04742f1ee9b042a4d819c976c4', 'expiry script'),
        (0x34ce, 109, 'bf34a70116f5d9aa85a79c231fc3b31ce62c62e8c7ace95cda7d78c6f1954c13', 'pause'),
        (0x3cdb, 84, '5bc3ab52eea26fd740a154ef078f7bb9113ce587b70898d3401c20d9eb86ec3d', 'table teardown'),
    ]:
        require(digest(bounded(table, start, size)) == sha, meaning+' evidence drift')
        out['reviewed_code_ranges'].append(dict(offset=start, size=size, sha256=sha, meaning=meaning))
    prior = next(a for a in json.loads(evidence.read_text()) if a['source'].endswith('/pinbfan'))
    prior_map = {(r['canonical_offset'], r['size'], r['purpose'])
                 for r in prior['files']['TABLE1.PRG']['regions']}
    current_map = {(r['offset'], r['size'], r['purpose'])
                   for r in profiles['TABLE1.PRG']['regions']}
    require(prior_map == current_map, 'prior audit does not cover the current TABLE1 read map')
    regions = []
    for r in prior['files']['TABLE1.PRG']['regions']:
        row = dict(canonical_offset=r['canonical_offset'], size=r['size'], purpose=r['purpose'])
        at = r.get('chosen_offset')
        if r['purpose'] == 'material parameters':
            at = r['canonical_offset']+362
        if at is not None:
            require(bounded(table, at, r['size']) == bounded(ref['TABLE1.PRG'], r['canonical_offset'], r['size']),
                    'prior candidate no longer byte-equivalent')
            row.update(source_candidate=at, shift=at-r['canonical_offset'],
                       status='byte-equivalent; not an instruction-level consumer proof')
        elif r['purpose'] in ('matrix record S_EMPTY', 'matrix record S_GAMEOVER2'):
            row.update(source_candidate=r['canonical_offset']+140, shift=140,
                       status='reviewed typed jingle difference')
        else:
            row.update(status='OPEN consumer proof', source_candidate=r.get('nearest_anchor_candidate'))
        regions.append(row)
    out['table1_read_map'] = regions
    out['table1_candidate_shift_counts'] = dict(Counter(r['shift'] for r in regions if 'shift' in r))
    out['open_gate'] = ['sound/raster callback cadence and skipped/reentrant callback policy',
                        'complete linked demo matrix/control graph and canonical-read-map coverage',
                        'INTRO text records and exact advertising-card placement']
    return out


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    for name in ('data', 'canonical', 'evidence', 'output'):
        p.add_argument('--'+name, required=True, type=Path)
    a = p.parse_args()
    result = audit(a.data, a.canonical, a.evidence)
    a.output.write_text(json.dumps(result, indent=2)+'\n')
    print(f"DMO0 metadata verified: {FINGERPRINT}; production gate remains OPEN")
