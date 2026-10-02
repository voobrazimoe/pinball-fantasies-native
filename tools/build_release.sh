#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
./tools/build_windows.sh
mkdir -p release/windows release/linux
cp bin/pinballfantasies.exe release/windows/pinballfantasies.exe
cp packaging/README-windows.txt release/windows/README.txt
cp .tools/go/LICENSE release/windows/Go-LICENSE.txt
cp packaging/README-linux.txt release/linux/README.txt
python3 tools/build_appimage.py --mode public "$@"
