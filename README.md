# Pinball Fantasies Native

**English** | [Русский](README.ru.md)

First public beta of a native, source-guided reimplementation of the DOS version
of Pinball Fantasies for modern Windows and Linux. All four tables use native
Go game logic, integer ball physics, matrix presentation and tracker/sample audio.
The program reads your original game files as data and never executes their x86 code.

## Quick Start

Download `pinballfantasies.exe` (Windows) or `PinballFantasies-x86_64.AppImage`
(Linux) from [Releases](https://github.com/voobrazimoe/pinball-fantasies-native/releases).
**Public binaries contain no original game data.** Provide your own legally
obtained Pinball Fantasies DOS data, either beside the binary or using `-data-dir`:

```text
pinballfantasies.exe -data-dir "D:\Games\Pinball Fantasies"
```

```sh
chmod +x PinballFantasies-x86_64.AppImage
./PinballFantasies-x86_64.AppImage -data-dir "/path/to/Pinball Fantasies"
```

The supported DOS installation supplies these 12 required files:

```text
INTRO.PRG  INTRO.MOD  MOD2.MOD
TABLE1.PRG TABLE1.MOD
TABLE2.PRG TABLE2.MOD
TABLE3.PRG TABLE3.MOD
TABLE4.PRG TABLE4.MOD
PINBALL.CFG
```

Fingerprints enforce supported data builds. Keep originals read-only. Settings,
high scores and logs go to `userdata/` beside the executable/AppImage, with a
per-user configuration fallback. `-config-dir` and `-high-score-dir` override it.
Optional legacy `TABLE*.HI` seeds are read-only; without them native factory
scores are used. Personal builds never include these high-score files.

## Controls and multiplayer

Before loading a table, **F1–F4 choose a table** (Party Land, Speed Devils,
Billion Dollar Gameshow, Stones ’N Bones); **F5 opens options**.

Inside a loaded table, **F1–F8 start a game with 1–8 players**.
**Enter adds a player before the first ball is launched**, up to eight; in
attract mode Enter starts one player. After an eight-player game, the original
attract-mode behavior ignores Enter: use F1–F8 to start the next game.
Before the first launch, F1–F8 can replace the player count, with a short
source-timed inhibit between changes. The count stays fixed after that launch.
Players rotate each ball round; extra balls stay with the player who earned them.
Each player retains their own score and table state. The matrix presents the
incoming player/ball immediately at handoff, before launch.

| Key | Action |
|---|---|
| Down | Hold to charge the plunger; release to launch |
| Shift, Ctrl, Alt | Flippers |
| Space | Nudge; repeated nudges can tilt |
| P | Pause |
| M | Toggle music |
| Alt+Enter | Toggle borderless fullscreen |
| Esc | Back/quit according to the current screen |

Options include 3/5-ball games and music, scrolling and resolution preferences.
`SCROLLING OFF` displays the entire table using the same physics and rules.
The native window persists across screens and is resizable. Focus loss pauses
an active game; after focus returns, press a game key such as P to resume.

## Build and verification

See [build instructions](docs/build.md). Linux requires a C compiler/libc
headers, SDL2 runtime and Python 3. Setup downloads pinned Go and SDL2 headers;
there are no third-party Go modules:

```sh
python3 tools/setup-local.py
./tools/go.sh test -p=1 -count=1 ./...
python3 tools/check_public_source.py
./tools/go.sh build -buildvcs=false -trimpath -o bin/pinballfantasies ./cmd/pinballfantasies
./tools/build_windows.sh
./tools/build_release.sh
```

[Public CI](.github/workflows/public.yml) runs the asset-free suite and native
builds on Linux and Windows. It downloads no commercial data or historical source.
Tests needing originals, private captures or historical source explicitly skip
when those inputs are absent; source-independent checks still run.
Source-backed reconstruction tools are optional development validation.

See [architecture](docs/architecture.md), [runtime data](docs/runtime-data.md)
and [reconstruction reports](analysis/). Historical reports describe their
recorded development state; their older publication gates are superseded by
this MIT public beta. The preserved Party oracle is 1200 ticks, score
`000002300000`, ball 2, frame SHA256
`f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e`.

## Beta limitations

On some Linux multi-monitor desktops, entering/leaving fullscreen may briefly
flicker or move the window between displays.

A native Windows Win32/GDI/waveOut build exists. Hosted native Windows
build/tests and the documented Wine/live development regressions pass.
This does not claim exhaustive physical Windows acceptance across every
multiplayer/fullscreen/audio combination. Wine is not real Windows hardware.
The independent native 5x7 sidebar font visibly differs from the PC BIOS font.

## Local-only personal builder

For local use, `./tools/build_personal_release.sh "/path/to/original/game"`
creates self-contained builds under `release/personal/` using exactly the 12
inputs above. `TABLE*.HI` are not personal inputs. These builds contain your
commercial data and are **not distributable project releases**. Never commit,
push or upload personal EXE/AppImage files, embedded payloads or extracted data.
Only asset-free binaries are provided in project Releases.

## License and game data

Native code and project-authored material are released under the
[MIT License](LICENSE), copyright (c) 2026 voobrazimoe.
Pinball Fantasies itself is not relicensed. Original commercial game data,
artwork and music are not covered by MIT. Historical/reference source material
is not relicensed by this project. This project claims no ownership of the
original game's trademarks, artwork, music or commercial data.
Users need their own legally obtained DOS game data. Personal self-contained
builds containing local commercial data are not distributable project releases.
See [provenance](analysis/provenance.md).
