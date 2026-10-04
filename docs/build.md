# Build and run

The Linux build uses Go, a C compiler with libc development files, Python 3,
`dpkg-deb` and the SDL2 runtime. `python3 tools/setup-local.py` downloads pinned,
checksum-verified Go and SDL2 development headers into ignored `.tools/`.
The project has no third-party Go modules. Windows uses Go's native Win32,
GDI and waveOut interfaces and needs no SDL DLL. macOS uses an Objective-C
AppKit/Core Animation/AudioUnit host linked to the same Go engine through ABI 1;
build it with Apple's compiler and SDK as described below.

```sh
python3 tools/setup-local.py
mkdir -p bin
./tools/go.sh build -buildvcs=false -trimpath -o bin/pinballfantasies ./cmd/pinballfantasies
./tools/build_windows.sh
./tools/go.sh test ./...
```

Run `bin/pinballfantasies -data-dir /path/to/original/game` on Linux or place the
Windows executable beside the supported original files. Consult the root README
for required files and controls. `-duration 3s` gives a bounded window smoke;
`-png /tmp/frame.png` exports a frame without opening a window.

Tests requiring original files skip explicitly when those files are absent.
Corrupt supplied inputs fail validation. The complete Windows/Linux asset-free checkout runs `./tools/go.sh test ./...`; integration
checks skip with reasons. Local/private integration uses the same command with
legally obtained inputs and historical source supplied separately. CI never
downloads those inputs. Do not use historical reference setup for a public build.

`tools/build_release.sh` builds local Windows and Linux/AppImage artifacts without
commercial assets. AppImage tooling is pinned in `tools/appimage-tools.json`;
`--tool` and `--runtime` accept checksum-matching offline inputs. Packaging requires
`ldd` and the Linux SDL runtime. AppImage includes the bundled runtime libraries’ third-party license notices
and build manifest. These notices are separate from the project MIT license.

GitHub CI uses Go 1.27.1 and Python 3 on Linux, Windows and macOS. On Linux install
`gcc libc6-dev libsdl2-dev pkg-config python3` before building. With system SDL
headers (rather than setup-local.py), set `CGO_CFLAGS="$(pkg-config --cflags sdl2)"`. `tools/go.sh` uses the pinned
local toolchain when installed, otherwise Go from PATH; Windows may run `go test
./...` and `go build -buildvcs=false -trimpath ./cmd/pinballfantasies` directly.
Linux AppImage packaging additionally needs Debian/Ubuntu package metadata,
`dpkg-query`, `ldd` and network access for the pinned AppImage tools.

The host-independent engine can also be built as a C archive/shared library
using `tools/build_engine.sh`. See [native engine build and conformance](native-engine.md)
for the stable header, ownership rules and original-backed replay command. Linux
CI builds both C interfaces without commercial assets. macOS 13+ is supported by separate ARM64 and Intel x86_64 native apps:

```sh
./tools/build_macos.sh           # ARM64
./tools/build_macos.sh x86_64    # Intel
```

See [macOS build and usage](macos.md) for Apple toolchain prerequisites, import,
Application Support storage and ad-hoc signing. The macOS app host is separate
from the Windows/Linux-only `cmd/pinballfantasies` and `internal/platform`.
On macOS, run the shared suite with those two packages excluded, as CI does:

```sh
packages=$(./tools/go.sh list ./... | sed '/\/internal\/platform$/d; /\/cmd\/pinballfantasies$/d')
./tools/go.sh test -p=1 -count=1 $packages
```

Run that snippet in Bash; `tools/build_macos.sh` adds native host, C ABI, signature
and bundle checks. Hosted macOS CI tests/builds ARM64 and cross-builds/packages
Intel. Hosted tests remain asset-free; local original-backed replay/native
journeys and owner physical-Mac acceptance are recorded in
[macOS validation](macos-validation.md). Public bundles contain no commercial
data and are not notarized. The published `v0.1.1` release contains only
Windows/Linux downloads and checksums; Mac assets have not yet been published
as a tagged release. Android remains unimplemented.
