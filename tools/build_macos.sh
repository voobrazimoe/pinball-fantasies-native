#!/bin/bash
# Build on macOS with the Apple SDK/compiler; optional target: arm64 or x86_64.
set -euo pipefail
cd "$(dirname "$0")/.."
test "$(uname -s)" = Darwin
target=${1:-arm64}
case "$target" in
    arm64) goarch=arm64; build=bin/macos; release=release/macos ;;
    x86_64) goarch=amd64; build=bin/macos-x86_64; release=release/macos-x86_64 ;;
    *) echo 'Usage: tools/build_macos.sh [arm64|x86_64]' >&2; exit 2 ;;
esac
export GOOS=darwin GOARCH="$goarch" CGO_ENABLED=1
export CC="$(xcrun --find clang)"
export MACOSX_DEPLOYMENT_TARGET=13.0
sdk=$(xcrun --sdk macosx --show-sdk-path)
export SDKROOT="$sdk"
export CGO_CFLAGS="-isysroot $sdk -arch $target -mmacosx-version-min=13.0"
export CGO_LDFLAGS="-isysroot $sdk -arch $target -mmacosx-version-min=13.0"
mkdir -p "$build" "$release"
./tools/build_engine.sh c-archive "$build/libpfengine.a"
./tools/build_engine.sh c-shared "$build/libpfengine.dylib"
./tools/go.sh build -buildvcs=false -trimpath -o "$build/pftrace" ./cmd/pftrace
app="$release/Pinball Fantasies.app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp hosts/macos/Info.plist "$app/Contents/Info.plist"
build_revision=$(git rev-parse --short HEAD)
if ! git diff --quiet HEAD -- hosts/macos tools/build_macos.sh; then build_revision="$build_revision-dirty"; fi
/usr/libexec/PlistBuddy -c "Add :PFHostBuild string $build_revision" "$app/Contents/Info.plist"
/usr/libexec/PlistBuddy -c "Add :PFHostArchitecture string $target" "$app/Contents/Info.plist"
cp LICENSE "$app/Contents/Resources/LICENSE.txt"
cp "$(./tools/go.sh env GOROOT)/LICENSE" "$app/Contents/Resources/Go-LICENSE.txt"
./tools/go.sh env GOVERSION > "$app/Contents/Resources/Go-version.txt"
common=(-arch "$target" -isysroot "$sdk" -mmacosx-version-min=13.0 -Wall -Wextra -Werror -O2)
objc=(-fobjc-arc -fblocks -framework AppKit -framework AudioToolbox -framework CoreAudio -framework IOKit -framework Security -framework CoreFoundation -framework CoreGraphics -framework QuartzCore)
"$CC" "${common[@]}" -std=c11 -pthread hosts/macos/host_logic.c hosts/macos/logic_tests.c -o "$build/logic-tests"
"$CC" "${common[@]}" "${objc[@]}" hosts/macos/main.m hosts/macos/frame_view.m hosts/macos/storage.m hosts/macos/audio_host.m hosts/macos/native_input.m hosts/macos/host_logic.c "$build/libpfengine.a" -o "$app/Contents/MacOS/pinballfantasies"
"$CC" "${common[@]}" "${objc[@]}" hosts/macos/native_tests.m hosts/macos/frame_view.m hosts/macos/storage.m hosts/macos/audio_host.m hosts/macos/native_input.m hosts/macos/host_logic.c "$build/libpfengine.a" -o "$build/native-tests"
# The Apple linker retains an N_OSO archive path even with Go -trimpath.
# Remove debug symbols before signing so local checkout paths are not shipped.
xcrun strip -S "$app/Contents/MacOS/pinballfantasies"
codesign --force --sign - "$app"
codesign --verify --strict "$app"
file "$app/Contents/MacOS/pinballfantasies"
test "$(lipo -archs "$app/Contents/MacOS/pinballfantasies")" = "$target"
otool -L "$app/Contents/MacOS/pinballfantasies"
plutil -lint "$app/Contents/Info.plist"
if [ "$(uname -m)" = "$target" ]; then
    "$build/logic-tests"
    python3 tools/test_engine_abi_contract.py --library "$PWD/$build/libpfengine.dylib"
    if [ -n "${PF_ENGINE_DATA_DIR:-}" ]; then
        "$build/native-tests" "$PF_ENGINE_DATA_DIR"
    else
        "$build/native-tests"
    fi
    if [ -n "${PF_ENGINE_DATA_DIR:-}" ]; then
        python3 tools/test_engine_abi.py --library "$PWD/$build/libpfengine.dylib" --oracle "$PWD/$build/pftrace" --data "$PF_ENGINE_DATA_DIR"
    else
        echo 'UNVERIFIED: four-table original-backed macOS C conformance requires external originals'
    fi
else
    echo "UNVERIFIED: $target executables cross-built on $(uname -m); run logic/native/ABI tests on the target Mac"
fi
python3 tools/check_macos_bundle.py "$app"
versioned_zip="$release/PinballFantasies-$target-$build_revision.zip"
ditto -c -k --sequesterRsrc --keepParent "$app" "$versioned_zip"
cp "$versioned_zip" "$release/PinballFantasies-$target.zip"
