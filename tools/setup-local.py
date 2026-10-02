#!/usr/bin/env python3
"""Pinned, checksum-verified toolchain/headers; no sudo or system mutation."""
import hashlib, os, subprocess, tarfile, urllib.request
from pathlib import Path
os.chdir(Path(__file__).resolve().parent.parent)
root=Path('.tools');root.mkdir(exist_ok=True)
def download(url,name,expected):
    p=root/name
    if not p.exists():urllib.request.urlretrieve(url,p)
    if hashlib.sha256(p.read_bytes()).hexdigest()!=expected:raise SystemExit('SHA-256 mismatch: '+str(p))
    return p
if not (root/'go/bin/go').exists():
    archive=download('https://go.dev/dl/go1.27.1.linux-amd64.tar.gz','go1.27.1.linux-amd64.tar.gz','63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445')
    with tarfile.open(archive) as t:t.extractall(root,filter='data')
if not (root/'sdl2/usr/include/SDL2/SDL.h').exists():
    package=download('https://archive.ubuntu.com/ubuntu/pool/universe/libs/libsdl2/libsdl2-dev_2.32.10+dfsg-6_amd64.deb','libsdl2-dev_2.32.10+dfsg-6_amd64.deb','b8e94803098e9299afc031fb898d0908387d88308eb62ea239918f455c69e985')
    subprocess.run(['dpkg-deb','-x',str(package),str(root/'sdl2')],check=True)
subprocess.run([str(root/'go/bin/go'),'version'],check=True)
print('Requires existing gcc/libc development files and libSDL2-2.0.so.0 runtime; see docs/build.md.')
