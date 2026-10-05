# Original PF application icon

`pf-icon-master.svg` is the canonical, editable, independently drawn source.
A heavy steel P anchors a forward-leaning amber F. One muted rose trajectory
and a small bright ball suggest pinball motion. Flat geometric contours and
negative space remain legible at 32–48 px.

Clean-room provenance: no commercial title image was used as an input, embedded,
rasterized or traced. The general weight contrast, kinetic feeling and early-90s
arcade era inform the concept only. Lettering, trajectory, palette and composition
are independently designed; this is not the commercial logo or wordmark.
Project-authored source and exports use the repository MIT license.

## Regeneration

Requires the repository Go toolchain, Python 3 and Node.js. No proprietary GUI:

```sh
python3 -m venv .cache/icon-export/venv
.cache/icon-export/venv/bin/pip install Pillow==12.3.0
npm install --prefix .cache/icon-export --no-save sharp@0.34.5
.cache/icon-export/venv/bin/python tools/export_app_icons.py
python3 tools/check_app_icons.py
```

The exporter also invokes `tools/go.sh run github.com/akavel/rsrc@v0.10.2`
to create `internal/platform/pficon_windows_amd64.syso`. The exact command is:

```sh
./tools/go.sh run github.com/akavel/rsrc@v0.10.2 \
  -ico art/app-icon/pf-icon.ico -o internal/platform/pficon_windows_amd64.syso
```

For already-installed runtimes, `PF_ICON_NODE` overrides the Node executable and
`PF_ICON_SHARP` the sharp module path. Derived assets are checked in, so normal
platform builds need no rasterizer, Pillow, npm or resource-compiler download.

## Platform exports

- Android foreground SVG/vector scales the mark to 90% about its center. Every
  stroke and ball lies inside the 66/108 adaptive safe circle. The background is
  full bleed. Circle, rounded-square and squircle masks preserve the mark.
  API 33+ receives a white monochrome vector. Five density fallback/round PNG
  pairs and adaptive XMLs are in `hosts/android/app/src/main/res/`.
- Windows ICO includes 16/24/32/48/64/128/256 px. Go links the AMD64 `.syso`
  automatically; Win32 loads group icon #1 for the window/taskbar. Verify the
  linked executable with `python3 tools/check_app_icons.py --exe bin/pinballfantasies.exe`.
- macOS ICNS includes Retina representations through 1024 px; the build copies
  it into `Contents/Resources/pf-icon.icns` before the existing ad-hoc signature.
  `CFBundleIconFile` supplies Finder and AppKit/Dock. Bundle validation requires
  an exact match to this resource.
- Linux uses `png/SIZE/pinballfantasies.png`, the SVG in hicolor/scalable, an
  AppDir root PNG, `.DirIcon`, and matching desktop entries. AppImage packaging
  copies these public exports rather than inventing another logo.

`png/foreground.png` is a transparent mask-check export, not a desktop icon.
The dependency-free gate checks dimensions, icon containers, wiring and every
nontransparent foreground pixel against the Android safe circle. Owner visual
confirmation on launchers, Finder/Dock and Windows remains the final RC gate.
