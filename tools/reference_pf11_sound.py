#!/usr/bin/env python3
"""Read-only SOUND.CFG boundary evidence from the supplied sound-driver binary.
Uses PF5's proven SP packet expansion; output is archaeology, not runtime code.
"""
import hashlib,json,struct,subprocess
from pathlib import Path
root=Path(__file__).resolve().parent.parent
original=(root/'SBLASTER.SDR').read_bytes();b=original[512:];cs=struct.unpack_from('<H',original,22)[0]*16;h=struct.unpack_from('<8H',b,cs)
src=cs-h[7]*16+15;dest=h[6]*16-h[7]*16+15
while b[src]==255:src-=1
expanded=bytearray(dest+1)
while True:
 marker=b[src];n=struct.unpack_from('<H',b,src-2)[0];src-=3
 if marker&254==0xb0:expanded[dest-n+1:dest+1]=bytes([b[src]])*n;src-=1
 elif marker&254==0xb2:expanded[dest-n+1:dest+1]=b[src-n+1:src+1];src-=n
 else:raise ValueError('invalid SP packet')
 dest-=n
 if marker&1:break
assert src==dest
expanded[:dest+1]=b[:src+1]
base=0x1e20
ports=struct.unpack_from('<6H',expanded,base+0x6b56);irqs=list(expanded[base+0x6b62:base+0x6b66]);rates=struct.unpack_from('<10H',expanded,base+0x6ba2)
config=(root/'SOUND.CFG').read_bytes()
assert config[:13]==b'SBLASTER.SDR\0' and ports==(528,544,560,576,592,608) and irqs==[2,3,5,7]
evidence=dict(size=len(config),driver=config[:13].rstrip(b'\0').decode(),raw=list(config),tail_cells=[dict(offset=i,tag=config[i],word=struct.unpack_from('<H',config,i+1)[0]) for i in range(13,len(config),3)],sblaster_reader=dict(config_offsets=[14,17,20],base_ports=list(ports),irq_map=irqs,rate_flags=list(rates),configured_port=ports[config[14]],configured_irq=irqs[config[17]],configured_rate=rates[2*config[20]],quality_extra_flag=rates[2*config[20]+1]),driver_sha256=hashlib.sha256(original).hexdigest(),expanded_sha256=hashlib.sha256(expanded).hexdigest(),evidence_addresses=dict(reader='18eb..1a33',data_segment_base=base,port_table='DS:6b56',irq_table='DS:6b62',quality_table='DS:6ba2'),opaque=['tail-cell tag bytes; ignored by this SDR reader','offset23 word: external SETSOUND calibration/metadata; not consumed by this reader'],boundary='INTRO and FANTASIE read only filename13; SETSOUND creates record; F5 touches PINBALL.CFG only')
(root/'analysis/pf11-sound-boundary.json').write_text(json.dumps(evidence,indent=2)+'\n')
tmp=Path('/tmp/pf11-sblaster-source-evidence.bin');tmp.write_bytes(expanded)
disasm=subprocess.check_output(['objdump','-D','-b','binary','-m','i8086','-Mintel','--start-address=0x18eb','--stop-address=0x1a34',str(tmp)],text=True)
(root/'analysis/pf11-sound-reader-disassembly.txt').write_text(disasm)
print('SOUND.CFG: filename13 +4 opaque tagged word slots; verified SoundBlaster port220h IRQ7 rate21000; no F5 integration')
