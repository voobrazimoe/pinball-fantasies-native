# Build and run

The Linux build uses Go, a C compiler with libc development files, Python 3,
`dpkg-deb` and the SDL2 runtime. `python3 tools/setup-local.py` downloads pinned,
checksum-verified Go and SDL2 development headers into ignored `.tools/`.
The project has no third-party Go modules. Windows uses Go's native Win32,
GDI and waveOut interfaces and needs no SDL DLL.

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
Corrupt supplied inputs fail validation. The complete asset-free checkout runs `./tools/go.sh test ./...`; integration
checks skip with reasons. Local/private integration uses the same command with
legally obtained inputs and historical source supplied separately. CI never
downloads those inputs. Do not use historical reference setup for a public build.

`tools/build_release.sh` builds local Windows and Linux/AppImage artifacts without
commercial assets. AppImage tooling is pinned in `tools/appimage-tools.json`;
`--tool` and `--runtime` accept checksum-matching offline inputs. Packaging requires
`ldd` and the Linux SDL runtime. AppImage includes the bundled runtime libraries’ third-party license notices
and build manifest. These notices are separate from the project MIT license.

GitHub CI uses Go 1.27.1 and Python 3 on Linux and Windows. On Linux install
`gcc libc6-dev libsdl2-dev pkg-config python3` before building. With system SDL
headers (rather than setup-local.py), set `CGO_CFLAGS="$(pkg-config --cflags sdl2)"`. `tools/go.sh` uses the pinned
local toolchain when installed, otherwise Go from PATH; Windows may run `go test
./...` and `go build -buildvcs=false -trimpath ./cmd/pinballfantasies` directly.
Linux AppImage packaging additionally needs Debian/Ubuntu package metadata,
`dpkg-query`, `ldd` and network access for the pinned AppImage tools.

The host-independent engine can also be built as a C archive/shared library
using `tools/build_engine.sh`. See [native engine build and conformance](native-engine.md)
for the stable header, ownership rules and original-backed replay command. Linux
CI builds both C interfaces without commercial assets. The macOS arm64 host is built with Apple tools by `tools/build_macos.sh`; see
[macOS prerequisites, packaging and acceptance](macos.md). Hosted native build/tests
are available; original-backed and physical-Mac acceptance are pending. Android
remains unimplemented.
