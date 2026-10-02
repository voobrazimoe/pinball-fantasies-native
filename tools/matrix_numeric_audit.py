#!/usr/bin/env python3
"""Inventory all four tables' numeric matrix sites and source text writers.

Source opcodes and line numbers come from original ASM, not Go string guesses.
--check verifies the committed JSON/Markdown inventory without writing.
Pixel regressions exercise formatting; the existing operand/reachability audits
remain independent and unchanged.
"""
import argparse
import json
import re
from pathlib import Path
import matrix_reachability as mr
import matrix_operand_audit as operands

ROOT = Path(__file__).resolve().parent.parent
TABLES = {1: ('PLAND', 'Party Land'), 2: ('SDEV', 'Speed Devils'),
          3: ('SHOW', 'Billion Dollar Gameshow'), 4: ('STONES', "Stones 'N' Bones")}
PRINT = {'_PRINT5_NUMBER', '_PRINT8_NUMBER', '_PRINT11_NUMBER', '_PRINT13_NUMBER',
         '_PRINT8_NUMBER_CENT', '_PRINT13_NUMBER_CENT'}
CUSTOM = {'_NUMBER', '_COUNTDOWN', '_COUNTDOWN2', '_COUNTDOWNCONTINUE',
          '_FLORPA', '_SHOW_SCORE', '_KNACKET', '_INIT_SCORE'}
WRITERS = {
    'BALLSTEXT': ('table player/ball update routines', 'single-player frontend panel prints live ball digit'),
    'PLAYERSTEXT': ('table player selection routines', 'single-player frontend panel prints player 1'),
    'PLAY_TEXT': ('table player/ball update routines', 'legacy SHOWINFO panel; native panel uses player/ball/score'),
    'BONUS_TEXT': ('table bonus multiplier text writes', 'legacy SHOWINFO bonus panel; one/two encoded multiplier digits'),
    'BONUS_X_TEXT': ('_BONUS_X_CALCS', 'single encoded multiplier; SHOW/STONES two cells for X10'),
    'PARTY_ON_TEXT': ('_PARTYON / _PARTYONN', 'one encoded player digit; preserve other source bytes'),
    'SHOOT_AGAIN_TEXT': ('_SHOOT_AGAIN', 'one encoded player digit; preserve other source bytes'),
    'SHOOTTHEBALLTEXT': ('_TURNOFFPARTY', 'one encoded player digit; preserve other source bytes'),
    'SKILLTEXT': ('SHOW Put_In_Text (3739)', 'only hundreds zero blanked; >=100 shifts destination one byte'),
    'MILES_TEXT': ('Put_In_Text / LOOPEN_BERTIL', 'three encoded digits; blank entire leading zero run; zero => ***'),
    'JUMP_AT_TEXT': ('Put_In_Text / LOOPEN_BERTIL', 'STONES: blank full leading run; SDEV: blank hundreds only'),
    'JUMP_AT_TEXT2': ('STONES Put_In_Text (5201)', 'blank full leading run; mutate only selected branch'),
    'OFFROAD_AT_TEXT': ('SDEV lites the off road (4755)', 'blank only hundreds; retain middle zero'),
}

def routine(op):
    if op in PRINT:
        return 'FANTASIE '+op.lower()+' -> print'+re.search(r'\d+', op)[0]+'_task -> PRINT_NUMBER'
    return {
        '_NUMBER': 'FANTASIE _NUMBER -> SIFFRORRUT -> CODE2 SCORE',
        '_COUNTDOWN': 'FANTASIE _countdown -> countdown -> print13_task / SEC_ASC',
        '_COUNTDOWN2': 'FANTASIE _countdown2 -> countdown -> print13_task / SEC_ASC',
        '_COUNTDOWNCONTINUE': 'FANTASIE _COUNTDOWNCONTINUE -> countdown -> print13_task / SEC_ASC',
        '_FLORPA': 'table-local _FLORPA -> print_siffror8 / print_siffror13 / CODE2 SCORE',
        '_SHOW_SCORE': 'table-local _SHOW_SCORE -> CODE2 SCORE',
        '_KNACKET': 'table-local KNACKRUT -> single FONT5 match digits',
        '_INIT_SCORE': 'table-local _INIT_SCORE -> live player BCD selection (no formatting)',
    }[op]


def build():
    content = mr.load_generated('internal/presentation/content.go')
    natives = operands.load_native(ROOT)
    regions = mr.current_regions()
    result = {}
    for table, (source, title) in TABLES.items():
        src = mr.Source(source, regions[table])
        sites = []
        for index, (op, args) in enumerate(src.commands):
            if op not in PRINT | CUSTOM:
                continue
            labels = [(pos, label) for label, pos in src.labels.items() if pos is not None and pos <= index]
            program = max(labels)[1] if labels else '?'
            sites.append({'source_line': src.lines_of[index], 'program': program,
                          'opcode': op, 'operands': args, 'original_routine': routine(op)})
        # Require classification for every native numeric opcode, including
        # macro-expanded attract/high-score commands not in the table region.
        for c in content[str(table)]['commands'] + content[str(table)]['attract']:
            if 'NUMBER' in c['op'] or c['op'].startswith('_COUNTDOWN'):
                assert c['op'] in PRINT | CUSTOM, c
        constants, lines = operands.table_constants(source)
        eligible = {c['args'][0] for c in content[str(table)]['commands'] + content[str(table)]['attract']
                    if c['args'] and (c['op'].startswith('_PRINT') and 'NUMBER' not in c['op'] or
                                      c['op'].startswith('_RULLGARDIN') or c['op'] == '_SCROLL')}
        writers = operands.runtime_text_buffers(source, lines, eligible, natives[table]['go_files'])
        for w in writers:
            if not w['native_text_write']:
                assert w['label'] in {'BALLSTEXT', 'PLAYERSTEXT', 'PLAY_TEXT', 'BONUS_TEXT'}, w
            w['native_path'] = ('source text writer' if w['native_text_write'] else 'native single-player panel replaces legacy frontend text routine')
            assert w['label'] in WRITERS, ('unclassified mutable formatting', source, w['label'])
            w['original_routine'], w['format'] = WRITERS[w['label']]
        result[str(table)] = {'table': title, 'source': source+'.ASM', 'numeric_sites': sites,
                              'mutable_text_writers': writers}
    return result


def markdown(result):
    out = ['# All-table numeric matrix site inventory', '',
           'Generated by `tools/matrix_numeric_audit.py`; verify with `--check`.', '',
           'Every listed source site retains its original formatter. PRINT_NUMBER skips leading BCD zeros; all-zero BCD is blank. Its centered form uses the inclusive first-nonzero SCASB count, rather than modern text centering. CODE2 SCORE uses a visible zero for zero and source comma placement. SEC_ASC blanks only its zero tens digit. Ordinary source text is never globally trimmed.', '',
           '`_PRINT11_NUMBER` is supported by the shared FANTASIE handler; these four source tables have no direct command site for it. The common tests cover it along with fonts 5/8/13.', '']
    for r in result.values():
        out += ['## '+r['table'], '', f"{len(r['numeric_sites'])} numeric/custom sites in `{r['source']}` (including macro bodies and source sites outside the original extraction region).", '',
                '| Line | Program | Opcode / operands | Original formatter |', '|---:|---|---|---|']
        for s in r['numeric_sites']:
            out.append(f"| {s['source_line']} | `{s['program']}` | `{s['opcode']}` {', '.join(s['operands'])} | {s['original_routine']} |")
        out += ['', '| Mutable buffer | Source lines | Original writer | Formatting / native path |', '|---|---|---|---|' ]
        for w in r['mutable_text_writers']:
            out.append(f"| `{w['label']}` | {', '.join(map(str,w['source_lines']))} | {w['original_routine']} | {w['format']}; {w['native_path']} |")
        out += ['']
    out += ['## Other native entry points', '',
            '- Idle player/ball panel uses single decimal player/ball fields and CODE2 SCORE; it does not format BCD via source text.',
            '- Attract/initials high scores and GameOver replay use the same source numeric commands and CODE2 SCORE.',
            '- Custom bonus count-down uses PRINT_NUMBER; mode count-down uses PRINT_NUMBER plus SEC_ASC.',
            '- Match animations print isolated one-digit score/random values, with original table scheduler cadence.',
            '- Table readers map CYCLONECOUNTERBCD to Cyclones / Miles / Skills / Screams and JACKVALUE to the live jackpot. Static 12-byte BCD constants remain BCD data; PRINT_NUMBER owns suppression.', '',
            'Formatting regressions cover every table/font, 2/10/100, full 12-digit BCD, centered and all-zero cases, real bonus counter commands, CODE2 SCORE, countdown, untouched ordinary text, and the table-specific mutable writers. See [the fix report](pf12-presentation-number-fix.md) for findings and validation.', '']
    return '\n'.join(out)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--check', action='store_true')
    args = ap.parse_args()
    result = build()
    if not args.check:
        output = ROOT/'.private-cleanup/generated'
        output.mkdir(parents=True, exist_ok=True)
        (output/'matrix-numeric-audit.json').write_text(json.dumps(result, indent=2, ensure_ascii=False)+'\n')
        (output/'matrix-numeric-audit.md').write_text(markdown(result))
    for r in result.values():
        print(f"{r['table']}: {len(r['numeric_sites'])} numeric/custom sites, {len(r['mutable_text_writers'])} classified mutable writers")
    print('OK')

if __name__ == '__main__':
    main()
