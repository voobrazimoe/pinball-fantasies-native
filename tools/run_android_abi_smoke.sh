#!/bin/sh
set -eu
cd "$(dirname "$0")/.."

abi=${1:-x86_64}
required_page_size=${2:-}
local_dir="bin/android/$abi"
remote_dir="/data/local/tmp/pfengine-a0-$$"

if [ ! -f "$local_dir/libpfengine.so" ] || [ ! -f "$local_dir/pfengine-smoke" ]; then
    echo "missing Android A0 artifacts for $abi; run tools/build_android_engine.sh first" >&2
    exit 2
fi

cleanup() {
    adb shell "rm -rf '$remote_dir'" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

page_size=$(adb shell getconf PAGE_SIZE | tr -d '\r')
echo "Android page size: $page_size"
if [ -n "$required_page_size" ] && [ "$page_size" != "$required_page_size" ]; then
    echo "expected Android page size $required_page_size, got $page_size" >&2
    exit 1
fi

adb shell "mkdir -p '$remote_dir'"
adb push "$local_dir/libpfengine.so" "$remote_dir/libpfengine.so" >/dev/null
adb push "$local_dir/pfengine-smoke" "$remote_dir/pfengine-smoke" >/dev/null
adb shell "chmod 755 '$remote_dir/pfengine-smoke' && cd '$remote_dir' && LD_LIBRARY_PATH=. ./pfengine-smoke ./libpfengine.so"
