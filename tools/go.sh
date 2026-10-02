#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
export GOMAXPROCS=1
export GOCACHE="$PWD/.cache/go-build"
export GOMODCACHE="$PWD/.cache/go-mod"
if [ -x "$PWD/.tools/go/bin/go" ]; then
    exec "$PWD/.tools/go/bin/go" "$@"
fi
exec go "$@"
