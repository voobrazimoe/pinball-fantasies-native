#!/bin/bash
# Owner-local experimental executable; no original data is copied into the app.
set -euo pipefail
cd "$(dirname "$0")/.."
test "$(uname -s)" = Darwin
arch=$(uname -m)
case "$arch" in arm64) goarch=arm64;; x86_64) goarch=amd64;; *) exit 2;; esac
build=bin/demo-experimental
app="$build/Pinball Fantasies Demo.app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
export GOOS=darwin GOARCH="$goarch" CGO_ENABLED=1 MACOSX_DEPLOYMENT_TARGET=13.0
export CC="$(xcrun --find clang)"
sdk=$(xcrun --sdk macosx --show-sdk-path)
export SDKROOT="$sdk"
export CGO_CFLAGS="-isysroot $sdk -arch $arch -mmacosx-version-min=13.0"
export CGO_LDFLAGS="$CGO_CFLAGS"
./tools/go.sh build -tags demodev -buildvcs=false -trimpath -buildmode=c-archive -o "$build/libpfengine.a" ./cmd/pfengine
cp hosts/macos/Info.plist "$app/Contents/Info.plist"
cp art/app-icon/pf-icon.icns "$app/Contents/Resources/pf-icon.icns"
cp LICENSE "$app/Contents/Resources/LICENSE.txt"
"$CC" -DPF_DEMODEV -arch "$arch" -isysroot "$sdk" -mmacosx-version-min=13.0 -Wall -Wextra -Werror -O2 \
 -fobjc-arc -fblocks -framework AppKit -framework AudioToolbox -framework CoreAudio -framework IOKit -framework Security -framework CoreFoundation -framework CoreGraphics -framework QuartzCore -framework GameController \
 hosts/macos/main.m hosts/macos/gamepad.m hosts/macos/frame_view.m hosts/macos/storage.m hosts/macos/audio_host.m hosts/macos/native_input.m hosts/macos/host_logic.c \
 "$build/libpfengine.a" -o "$app/Contents/MacOS/pinballfantasies"
xcrun strip -S "$app/Contents/MacOS/pinballfantasies"
codesign --force --sign - "$app"
codesign --verify --strict "$app"
printf '%s\n' "$PWD/$app/Contents/MacOS/pinballfantasies"
