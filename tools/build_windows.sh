#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p bin
export GOOS=windows GOARCH=amd64 CGO_ENABLED=0
exec ./tools/go.sh build -buildvcs=false -trimpath -p=1 -ldflags='-H=windowsgui -X pinballfantasies/internal/platform.GUIMode=1' -o bin/pinballfantasies.exe ./cmd/pinballfantasies
