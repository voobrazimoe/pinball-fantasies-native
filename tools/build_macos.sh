#!/bin/bash
# Must run natively on Apple Silicon with the Apple SDK/compiler.
set -euo pipefail
cd "$(dirname "$0")/.."
test "$(uname -s)" = Darwin
test "$(uname -m)" = arm64
export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1
export CC="$(xcrun --find clang)"
export MACOSX_DEPLOYMENT_TARGET=13.0
sdk=$(xcrun --sdk macosx --show-sdk-path)
mkdir -p bin/macos release/macos
./tools/build_engine.sh c-archive bin/macos/libpfengine.a
./tools/build_engine.sh c-shared bin/macos/libpfengine.dylib
./tools/go.sh build -buildvcs=false -trimpath -o bin/macos/pftrace ./cmd/pftrace
app='release/macos/Pinball Fantasies.app'
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
cp hosts/macos/Info.plist "$app/Contents/Info.plist"
cp LICENSE "$app/Contents/Resources/LICENSE.txt"
./tools/go.sh env GOVERSION > "$app/Contents/Resources/Go-version.txt"
common=(-arch arm64 -isysroot "$sdk" -mmacosx-version-min=13.0 -Wall -Wextra -Werror -O2)
objc=(-fobjc-arc -fblocks -framework AppKit -framework AudioToolbox -framework CoreAudio -framework IOKit -framework Security -framework CoreFoundation -framework CoreGraphics)
"$CC" "${common[@]}" -std=c11 -pthread hosts/macos/host_logic.c hosts/macos/logic_tests.c -o bin/macos/logic-tests
"$CC" "${common[@]}" "${objc[@]}" hosts/macos/main.m hosts/macos/frame_view.m hosts/macos/storage.m hosts/macos/audio_host.m hosts/macos/host_logic.c bin/macos/libpfengine.a -o "$app/Contents/MacOS/pinballfantasies"
"$CC" "${common[@]}" "${objc[@]}" hosts/macos/native_tests.m hosts/macos/frame_view.m hosts/macos/storage.m hosts/macos/audio_host.m hosts/macos/host_logic.c bin/macos/libpfengine.a -o bin/macos/native-tests
codesign --force --sign - "$app"
codesign --verify --strict "$app"
file "$app/Contents/MacOS/pinballfantasies"
test "$(lipo -archs "$app/Contents/MacOS/pinballfantasies")" = arm64
otool -L "$app/Contents/MacOS/pinballfantasies"
plutil -lint "$app/Contents/Info.plist"
bin/macos/logic-tests
bin/macos/native-tests ${PF_ENGINE_DATA_DIR:+"$PF_ENGINE_DATA_DIR"}
if [ -n "${PF_ENGINE_DATA_DIR:-}" ]; then
    python3 tools/test_engine_abi.py --library "$PWD/bin/macos/libpfengine.dylib" --oracle "$PWD/bin/macos/pftrace" --data "$PF_ENGINE_DATA_DIR"
else
    echo 'UNVERIFIED: four-table original-backed macOS C conformance requires external originals'
fi
python3 tools/check_macos_bundle.py "$app"
ditto -c -k --sequesterRsrc --keepParent "$app" 'release/macos/PinballFantasies-arm64.zip'
