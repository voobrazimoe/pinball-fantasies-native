#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
exec python3 ./tools/build_personal_release.py "$@"
