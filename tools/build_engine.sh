#!/bin/sh
# Build on the target SDK/toolchain host. GOOS/GOARCH/CC may be supplied.
set -eu
cd "$(dirname "$0")/.."
mode=${1:-c-archive}
output=${2:-bin/libpfengine.a}
case "$mode" in c-archive|c-shared) ;; *) echo 'mode must be c-archive or c-shared' >&2; exit 2;; esac
mkdir -p "$(dirname "$output")"
CGO_ENABLED=1 ./tools/go.sh build -buildvcs=false -trimpath -buildmode="$mode" -o "$output" ./cmd/pfengine
cp cmd/pfengine/abi.h "$(dirname "$output")/abi.h"
