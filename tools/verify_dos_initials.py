#!/usr/bin/env python3
"""Private original-backed initials oracle; emits addresses/results, never PRG bytes.
Requires capstone and unicorn. Only reads --data. Runs linked IRQ/setup/input
instructions in isolated 16-bit memory, stopping before matrix rendering/I/O.
This is a routine oracle, not a full DOS gameplay/device acceptance run.
"""
import argparse
import hashlib
import json
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
from unicorn import Uc, UC_ARCH_X86, UC_MODE_16, UC_HOOK_CODE, UC_HOOK_INSN
from unicorn.x86_const import (UC_X86_REG_CS, UC_X86_REG_DS, UC_X86_REG_ES,
    UC_X86_REG_SS, UC_X86_REG_SP, UC_X86_REG_EFLAGS, UC_X86_INS_IN, UC_X86_INS_OUT)

# File offsets recovered from each linked retail PRG, not source assumptions.
PROFILES = [
    (0x19d40, 0x1d396, 0x8ad, 0x909, 0x36eb, 0x605, 0x607, 0x947, 0x4134, 0x4158),
    (0x18ee0, 0x1c5cb, 0x772, 0x7ce, 0x3780, 0x4ca, 0x4cc, 0x80c, 0x391a, 0x393e),
    (0x18b60, 0x1ba70, 0x740, 0x79c, 0x2fa5, 0x498, 0x49a, 0x7da, 0x33c7, 0x33eb),
    (0x166d0, 0x1a2f4, 0x818, 0x874, 0x3cb9, 0x570, 0x572, 0x8b2, 0x48e5, 0x4909),
]
ROOT = Path(__file__).resolve().parent.parent

def verify(table, data, inventory):
    ds, alpha, setup, reader, scan, pointer, count, accepted, irq, irq_end = PROFILES[table-1]
    name = f'TABLE{table}.PRG'
    b = (data/name).read_bytes()
    assert hashlib.sha256(b).hexdigest() == inventory[name], name
    dis = Cs(CS_ARCH_X86, CS_MODE_16)
    ins = list(dis.disasm(b[reader:reader+64], reader))
    lookup = next(i for i in ins if i.mnemonic == 'xlatb')
    assert any(i.mnemonic == 'mov' and i.op_str == f'bx, {hex(alpha-ds)}'
               for i in ins if i.address < lookup.address)
    assert any(i.mnemonic == 'cmp' and i.op_str == 'al, 0' for i in ins)
    irq_ins = list(dis.disasm(b[irq:irq_end], irq))
    assert any(i.mnemonic == 'in' and i.op_str == 'al, 0x60' for i in irq_ins)
    assert irq_ins[-1].mnemonic == 'mov' and irq_ins[-1].op_str == f'byte ptr [{hex(scan)}], al'
    # ALFA_KEYS is 128 scan-code entries. Keep original bytes in memory only.
    mapping = b[alpha:alpha+128]
    assert not any(mapping[2:12]), name
    uc = Uc(UC_ARCH_X86, UC_MODE_16)
    uc.mem_map(0, 0x100000)
    # Linked CS offset zero corresponds to file offset 0x300. DS paragraph
    # immediates correspond to file offsets minus 0x200 in this isolated image.
    uc.mem_write(0x100, b[0x300:])
    ds_linear = ds-0x200
    for reg, value in [(UC_X86_REG_CS,0x10),(UC_X86_REG_DS,ds_linear//16),
                       (UC_X86_REG_ES,ds_linear//16),(UC_X86_REG_SS,0x9000),
                       (UC_X86_REG_SP,0xff00),(UC_X86_REG_EFLAGS,2)]:
        uc.reg_write(reg,value)
    stops = set()
    halted = [False]
    def hook(cpu, address, size, _):
        # Stop at routine RET or just after the entry write/count decrement.
        if address in stops or cpu.mem_read(address,1) == b'\xc3':
            halted[0] = True
            cpu.emu_stop()
    uc.hook_add(UC_HOOK_CODE, hook)
    port_scan = [0]
    uc.hook_add(UC_HOOK_INSN, lambda cpu,port,size,_: port_scan[0] if port==0x60 else 0,
                None, 1, 0, UC_X86_INS_IN)
    uc.hook_add(UC_HOOK_INSN, lambda *args: None, None, 1, 0, UC_X86_INS_OUT)
    def run(at, stop=None):
        stops.clear()
        halted[0] = False
        if stop is not None: stops.add(stop-0x200)
        uc.reg_write(UC_X86_REG_SP,0xff00)
        uc.emu_start(at-0x200,0xfffff,count=1000)
        assert halted[0], "routine exceeded instruction bound"
    def word(at): return int.from_bytes(uc.mem_read(at,2),'little')
    # Execute GET_IT_FROM_KEYBOARD: rank/player one, queued make cleared.
    rank_load = next(i for i in dis.disasm(b[setup:reader],setup)
                     if i.mnemonic=='mov' and i.op_str.startswith('cx, word ptr cs:'))
    rank = int(rank_load.op_str.split('[')[1].split(']')[0],16)
    uc.mem_write(0x100+rank,b'\x01\x00')  # rank=1; player byte immediately before rank
    uc.mem_write(0x100+rank-1,b'\x01')
    uc.mem_write(ds_linear+scan,b'\x1e')
    run(setup)
    assert word(0x100+count)==3 and uc.mem_read(ds_linear+scan,1)==b'\xff'
    # Use a disposable name destination, preserving original counters/reader.
    def reset():
        uc.mem_write(0x100+pointer,(0x100).to_bytes(2,'little'))
        uc.mem_write(0x100+count,(3).to_bytes(2,'little'))
        uc.mem_write(ds_linear+0x100,b'???')
    def key(code):
        port_scan[0]=code
        run(irq,irq_end)
        assert uc.mem_read(ds_linear+scan,1)==bytes([code]) # IRQ stores raw make
        run(reader,accepted)
        assert uc.mem_read(ds_linear+scan,1)==b'\xff'
    results={}
    for code in range(2,12):
        reset(); key(code)
        assert word(0x100+count)==3 and word(0x100+pointer)==0x100
        assert uc.mem_read(ds_linear+0x100,3)==b'???'
    for label,codes,want in [('A1B',[30,2,48],b'AB?'),('7UP',[8,22,25],b'UP?'),
                             ('R2D',[19,3,32],b'RD?'),('ABC',[30,48,46],b'ABC'),
                             ('***',[57,57,57],b'***')]:
        reset()
        for code in codes: key(code)
        got=bytes(uc.mem_read(ds_linear+0x100,3))
        assert got==want and word(0x100+count)==got.count(b'?'),(name,label,got)
        results[label]=got.decode()
    return dict(table=table,inventory_match=True,alfa_keys=hex(alpha),
                get_it_from_keyboard=hex(setup),read_keyboard=hex(reader),irq=hex(irq),
                top_row_scans='0x02..0x0b: all rejected',routine_oracle=results)

if __name__=='__main__':
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--data',type=Path,required=True)
    args=ap.parse_args()
    inventory={r['name']:r['sha256'] for r in json.loads((ROOT/'analysis/game-inventory.json').read_text())}
    print(json.dumps([verify(t,args.data,inventory) for t in range(1,5)],indent=2))
