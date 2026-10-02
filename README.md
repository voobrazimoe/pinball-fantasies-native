# Pinball Fantasies Native

**English** | [Русский](README.ru.md)

First public beta of a native, source-guided reimplementation of the DOS version of **Pinball Fantasies** for modern Windows and Linux.

> **Important:** public builds contain no commercial Pinball Fantasies game data. You need your own legally obtained DOS copy of the game.

## Quick start

Download `pinballfantasies.exe` for Windows or `PinballFantasies-x86_64.AppImage` for Linux from [Releases](https://github.com/voobrazimoe/pinball-fantasies-native/releases).

Keep the original DOS data beside the executable, or point the port to it explicitly:

```text
pinballfantasies.exe -data-dir "D:\Games\Pinball Fantasies"
```

```sh
chmod +x PinballFantasies-x86_64.AppImage
./PinballFantasies-x86_64.AppImage -data-dir "/path/to/Pinball Fantasies"
```

The supported installation provides these 12 required runtime files:

```text
INTRO.PRG  INTRO.MOD  MOD2.MOD
TABLE1.PRG TABLE1.MOD
TABLE2.PRG TABLE2.MOD
TABLE3.PRG TABLE3.MOD
TABLE4.PRG TABLE4.MOD
PINBALL.CFG
```

The originals are treated as read-only data. Native settings, high scores and logs are written to `userdata/` beside the executable/AppImage, with a per-user configuration directory as fallback. Optional legacy `TABLE*.HI` files are read-only score seeds; without them the native factory scores are used.

## What is implemented

All four tables are playable:

- Party Land
- Speed Devils
- Billion Dollar Gameshow
- Stones ’N Bones

The port includes native table rules and state machines, integer ball physics, flippers, plunger, nudging and tilt, matrix display, lamps and animated playfield patches, high scores, options, tracker/module music and effects, persistent resizable windows, borderless fullscreen and native Windows/Linux platform backends.

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
| `Shift`, `Ctrl`, `Alt` | Flippers |
| `Space` | Nudge; repeated nudges can tilt |
| `P` | Pause |
| `M` | Toggle music |
| `Alt+Enter` | Toggle borderless fullscreen |
| `Esc` | Back / quit according to the current screen |

Before the first launch, `F1`–`F8` can replace the player count and `Enter` can add a player up to eight. After the first launch the count is fixed for that game. Players rotate each ball round; extra balls stay with the player who earned them. Each player retains their own score and table state. The matrix presents the incoming player and ball immediately at handoff, before launch.

After an eight-player game, the original attract-mode behavior ignores `Enter`; use `F1`–`F8` to start the next game.

## Options and full-table mode

The native F5 options include 3/5 balls, HIGH/LOW angle, HARD/MEDIUM/SOFT/OFF scrolling, music and NORMAL/HIGH resolution.

`SCROLLING: OFF` is a native extension that shows the complete 320×576 playfield together with the 320×33 matrix, producing a 320×609 logical frame. Gameplay continues through the same physics and rules as the scrolling modes.

## Audio

The native mixer reproduces the original four-channel tracker/module playback and effects as 48 kHz signed 16-bit stereo. Tracker voices 0 and 3 are routed left, and 1 and 2 right, following the Amiga Paula channel layout.

## Build and verification

See [build instructions](docs/build.md). The public CI runs the asset-free test/build path on Linux and Windows and downloads no commercial game data or historical source.

Useful project documentation:

- [Architecture](docs/architecture.md)
- [Runtime data boundary](docs/runtime-data.md)
- [Provenance and licensing boundary](docs/provenance.md)

The preserved Party deterministic oracle is 1200 ticks, score `000002300000`, ball 2, frame SHA256 `f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e`.

## Beta limitations

On some Linux multi-monitor desktops, entering or leaving fullscreen may briefly flicker or move the window between displays.

A native Windows Win32/GDI/waveOut build is provided. Hosted native Windows build/tests and the documented Wine/live development regressions pass, but this beta does not claim exhaustive physical-Windows coverage across every multiplayer/fullscreen/audio combination.

## Local-only personal builder

For local use, `./tools/build_personal_release.sh "/path/to/original/game"` creates self-contained builds under `release/personal/` from the same 12 inputs. These local builds contain your commercial game data and are **not distributable project releases**. Do not commit, push or upload them.

## License and game data

Native code and project-authored material are released under the [MIT License](LICENSE), copyright (c) 2026 voobrazimoe.

Pinball Fantasies itself is not relicensed. Original commercial game data, artwork, music and historical/reference source material are not covered by the MIT license. This project claims no ownership of the original game's trademarks, artwork, music or commercial data.

See [provenance](docs/provenance.md) for the project boundary.