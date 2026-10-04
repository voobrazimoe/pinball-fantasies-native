#!/bin/sh
set -eu
cd "$(dirname "$0")/.."

ANDROID_NDK_HOME=${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}
if [ -z "$ANDROID_NDK_HOME" ]; then
    echo 'ANDROID_NDK_HOME (or ANDROID_NDK_ROOT) must point to Android NDK r28 or newer' >&2
    exit 2
fi

API=${PF_ANDROID_API:-27}
HOST_TAG=${PF_ANDROID_NDK_HOST_TAG:-linux-x86_64}
TOOLCHAIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/$HOST_TAG/bin"
OBJDUMP="$TOOLCHAIN/llvm-objdump"
PAGE_LDFLAGS='-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'

if [ ! -x "$OBJDUMP" ]; then
    echo "Android NDK toolchain not found for host tag $HOST_TAG: $TOOLCHAIN" >&2
    exit 2
fi

build_one() {
    abi=$1
    case "$abi" in
        arm64-v8a)
            goarch=arm64
            cc="$TOOLCHAIN/aarch64-linux-android${API}-clang"
            ;;
        x86_64)
            goarch=amd64
            cc="$TOOLCHAIN/x86_64-linux-android${API}-clang"
            ;;
        *)
            echo "unsupported Android ABI: $abi" >&2
            exit 2
            ;;
    esac

    if [ ! -x "$cc" ]; then
        echo "Android compiler not found: $cc" >&2
        exit 2
    fi

    out="bin/android/$abi"
    mkdir -p "$out"
    echo "==> Android engine $abi (API $API)"

    # NDK r28+ defaults to 16 KB-compatible ELF alignment. Keep the flags
    # explicit as well because the final shared object is driven through Go/cgo.
    android_cgo_ldflags="${CGO_LDFLAGS:-} $PAGE_LDFLAGS"
    GOOS=android GOARCH="$goarch" CC="$cc" CGO_LDFLAGS="$android_cgo_ldflags" \
        ./tools/build_engine.sh c-shared "$out/libpfengine.so"

    "$cc" -O2 -fPIE -pie $PAGE_LDFLAGS ./tools/android_abi_smoke.c -ldl \
        -o "$out/pfengine-smoke"

    python3 ./tools/check_android_elf.py "$out/libpfengine.so" --objdump "$OBJDUMP"
    python3 ./tools/check_android_elf.py "$out/pfengine-smoke" --objdump "$OBJDUMP"
}

case "${1:-all}" in
    all)
        build_one arm64-v8a
        build_one x86_64
        ;;
    arm64-v8a|x86_64)
        build_one "$1"
        ;;
    *)
        echo 'usage: tools/build_android_engine.sh [all|arm64-v8a|x86_64]' >&2
        exit 2
        ;;
esac
