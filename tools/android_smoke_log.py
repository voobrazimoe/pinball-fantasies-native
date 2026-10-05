#!/usr/bin/env python3
"""Attribute brief-logcat crash records to the package under smoke test."""
import argparse
from pathlib import Path
import re

PACKAGE='io.github.voobrazimoe.pinballfantasies'
LINE=re.compile(r'^[A-Z]/([^ (]+)\s*\(\s*(\d+)\s*\):\s*(.*)$')
ERROR=re.compile(r'A1_.*ERROR|A2_BOOTSTRAP_REJECTED|FATAL EXCEPTION|UnsatisfiedLinkError|Fatal signal')

def host_errors(text, package=PACKAGE):
    records=[]; pids=set(); errors=[]
    for line in text.splitlines():
        match=LINE.match(line)
        if not match: continue
        tag,pid,message=match.groups();records.append((tag,pid,message,line))
        if tag in {'PinballFantasies','GameActivity'}:
            pids.add(pid)
        if tag=='AndroidRuntime' and re.search(r'\bProcess: '+re.escape(package)+r'(?:[:,\s]|$)',message):
            pids.add(pid)
        # Tombstones are emitted by crash_dump/debuggerd, not by the target PID.
        if tag=='DEBUG' and f'>>> {package} <<<' in message:
            errors.append(line)
    for tag,pid,message,line in records:
        if pid in pids and ERROR.search(message): errors.append(line)
    return errors

if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('log',type=Path)
    parser.add_argument('--package',default=PACKAGE);args=parser.parse_args()
    errors=host_errors(args.log.read_text(errors='replace'),args.package)
    if errors:
        print('\n'.join(errors));raise SystemExit(1)
