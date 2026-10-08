# Pinball Fantasies Native

**English** | [Русский](README.ru.md)

First public beta of a native, source-guided reimplementation of the DOS version of **Pinball Fantasies** for modern Windows, Linux, macOS and Android.

> **Important:** public builds contain no commercial Pinball Fantasies game data. You need your own legally obtained DOS copy of the game.

## Quick start

Download the published [v0.1.3 prerelease](https://github.com/voobrazimoe/pinball-fantasies-native/releases/tag/v0.1.3) for your platform:

- [Windows x86_64: pinballfantasies.exe](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/pinballfantasies.exe)
- [Linux x86_64: PinballFantasies-x86_64.AppImage](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/PinballFantasies-x86_64.AppImage)
- [macOS Apple Silicon ARM64: PinballFantasies-arm64.zip](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/PinballFantasies-arm64.zip)
- [macOS Intel x86_64: PinballFantasies-x86_64.zip](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/PinballFantasies-x86_64.zip)
- [Android ARM64 + x86_64: PinballFantasies-android.apk](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/PinballFantasies-android.apk)
- [SHA256 checksums](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.3/SHA256SUMS.txt)

For macOS 13 or later, download the ZIP for your Mac, extract it and launch the app. Select your original DOS game folder in the native import flow; files are validated and copied into Application Support. Public Mac bundles are ad-hoc signed and not notarized. See [macOS build and usage](docs/macos.md) for details.

On Windows/Linux, keep the original DOS data beside the executable, or point the port to it explicitly:

```text
pinballfantasies.exe -data-dir "D:\Games\Pinball Fantasies"
```

```sh
chmod +x PinballFantasies-x86_64.AppImage
./PinballFantasies-x86_64.AppImage -data-dir "/path/to/Pinball Fantasies"
```

## Android quick start

Download the public `PinballFantasies-android.apk` from the v0.1.3 prerelease above:

1. Install the APK and open Pinball Fantasies (Android 8.1 or later).
2. Tap **Import DOS folder** and choose the folder containing the supported original DOS files listed below.
3. The app validates and copies the 11 required PRG/MOD files and optional `PINBALL.CFG` into private app storage; the supplied originals remain unchanged.
4. Launch a table using the contextual touch controls, or attach a physical keyboard. Touch panels automatically hide while an external keyboard is active.

Commercial files are not bundled. The universal APK contains
arm64-v8a and x86_64, is development-signed, and supports 16 KB page-size devices.
See [Android usage](docs/android.md) and the [candidate preparation record](docs/release-v0.1.3.md).

Four DOS installations are supported: the original retail release, the Power Pack
`21STCENT/FANTASY` set, and both Deluxe CD `PFD/FANTASY` sets (including the GOG
copy). Point the app at the folder holding the game files; the layout is detected
automatically. See [runtime data](docs/runtime-data.md) for details.

Each supported installation provides these 11 required PRG/MOD runtime files:

```text
INTRO.PRG  INTRO.MOD  MOD2.MOD
TABLE1.PRG TABLE1.MOD
TABLE2.PRG TABLE2.MOD
TABLE3.PRG TABLE3.MOD
TABLE4.PRG TABLE4.MOD
```

The official **10-minute Party Land DOS demo** is also supported: supply a
folder with only its `INTRO.PRG`, `INTRO.MOD`, `MOD2.MOD`, `TABLE1.PRG` and
`TABLE1.MOD`, and the app starts the demo directly in Party Land, ending when
its timer expires. See [the demo notes](docs/partyland-10min-demo.md) for what
is proved and what follows the full-game rules.

`PINBALL.CFG` is an optional legacy settings seed. Missing or malformed settings use native defaults; writable PFNC settings live in native state. Runtime accepts the supported consumed-data layout; exact whole-file hashes are reserved for research and parity fixtures.

The originals are treated as read-only data. Windows/Linux write native settings, high scores and logs to `userdata/` beside the executable/AppImage, with a per-user configuration directory as fallback. macOS imports originals into `~/Library/Application Support/PinballFantasies/Data/` and writes settings and high scores to `~/Library/Application Support/PinballFantasies/State/`, outside the app bundle.

On Windows/Linux, optional legacy `TABLE*.HI` files are read-only score seeds; without them the native factory scores are used. The macOS importer copies only the required PRG/MOD files and optional CFG, using factory scores when there is no saved native state.

## What is implemented

All four tables are playable:

- Party Land
- Speed Devils
- Billion Dollar Gameshow
- Stones ’N Bones

The port includes native table rules and state machines, integer ball physics, flippers, plunger, nudging and tilt, matrix display, lamps and animated playfield patches, high scores, options, tracker/module music and effects, resizable windows, platform fullscreen and native Windows/Linux/macOS platform backends. Windows uses Win32/GDI/waveOut, Linux uses SDL2, and macOS uses AppKit/Core Animation/AudioUnit around the shared Go engine and C ABI.

This is not DOSBox, an x86 interpreter or a wrapper around the original executable. The original files are decoded as data; their x86 code is never executed.

## Controls and multiplayer

Before loading a table:

| Key | Action |
|---|---|
| `F1`–`F4` | Select Party Land / Speed Devils / Billion Dollar Gameshow / Stones ’N Bones |
| `F5` | Options |

Inside a loaded table:

| Key | Action |
|---|---|
| `F1`–`F8` | Start a game with 1–8 players |
| `Enter` | Add a player before the first ball is launched |
| `Down Arrow` | Hold to charge the plunger, release to launch |
| Left/right `Shift`, `Ctrl`, `Alt` (`Option` on macOS) | Corresponding left/right flippers |
| `Space` | Nudge; repeated nudges can tilt |
| `P` | Pause |
| `M` | Toggle music |
| `Alt+Enter` (Windows/Linux) | Toggle borderless fullscreen |
| `Option+Return`, `Command+F` or View → Toggle Full Screen (macOS) | Toggle native fullscreen |
| `Esc` | Back / quit according to the current screen |

On macOS, `Z` / `/` and Left / Right Arrow are alternate left/right flipper controls. If your Mac reports Shift sides reversed, enable **View → Swap Left/Right Shift**; the correction persists across launches and affects only Shift.

Before the first launch, `F1`–`F8` can replace the player count and `Enter` can add a player up to eight. After the first launch the count is fixed for that game. Players rotate each ball round; extra balls stay with the player who earned them. Each player retains their own score and table state. The matrix presents the incoming player and ball immediately at handoff, before launch.

After an eight-player game, the original attract-mode behavior ignores `Enter`; use `F1`–`F8` to start the next game.

## Options and full-table mode

The native F5 options include 3/5 balls, HIGH/LOW angle, HARD/MEDIUM/SOFT/OFF scrolling, music and NORMAL/HIGH resolution.

`SCROLLING: OFF` is a native extension that shows the complete 320×576 playfield together with the 320×33 matrix, producing a 320×609 logical frame. Gameplay continues through the same physics and rules as the scrolling modes.

## Audio

The native mixer reproduces the original four-channel tracker/module playback and effects as 48 kHz signed 16-bit stereo. Tracker voices 0 and 3 are routed left, and 1 and 2 right, following the Amiga Paula channel layout.

## Build and verification

See [build instructions](docs/build.md). Public CI runs asset-free shared tests and native Windows/Linux checks, plus Apple SDK ARM64 host tests/build and an Intel x86_64 cross-build. It downloads no commercial game data or historical source. Hosted asset-free checks are separate from local original-backed validation. macOS original-backed replays and native journeys passed, and physical Mac gameplay/input acceptance is recorded in the [validation report](docs/macos-validation.md).

Useful project documentation:

- [macOS build and usage](docs/macos.md)
- [macOS validation](docs/macos-validation.md)
- [Architecture](docs/architecture.md)
- [Runtime data boundary](docs/runtime-data.md)
- [Provenance and licensing boundary](docs/provenance.md)

The preserved Party deterministic oracle is 1200 ticks, score `000002300000`, ball 2, frame SHA256 `f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e`.

## Beta limitations

On some Linux multi-monitor desktops, entering or leaving fullscreen may briefly flicker or move the window between displays.

A native Windows Win32/GDI/waveOut build is provided. Hosted native Windows build/tests and the documented Wine/live development regressions pass, but this beta does not claim exhaustive physical-Windows coverage across every multiplayer/fullscreen/audio combination.

## Local-only personal builder

For local Windows/Linux use, `./tools/build_personal_release.sh "/path/to/original/game"` creates self-contained builds under `release/personal/` from the same 11 game-data inputs and optional settings seed. These local builds contain your commercial game data and are **not distributable project releases**. Do not commit, push or upload them.

For macOS, `python3 tools/build_personal_macos.py "/path/to/original/game"` creates separate ARM64 and x86_64 apps with automatic first-run import. The same local-only commercial-data boundary applies.

## License and game data

Native code and project-authored material are released under the [MIT License](LICENSE), copyright (c) 2026 voobrazimoe.

Pinball Fantasies itself is not relicensed. Original commercial game data, artwork, music and historical/reference source material are not covered by the MIT license. This project claims no ownership of the original game's trademarks, artwork, music or commercial data.

See [provenance](docs/provenance.md) for the project boundary.
