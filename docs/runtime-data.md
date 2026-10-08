# Runtime data boundary

The native port requires data from a supported DOS installation of Pinball Fantasies. Public source and public release binaries do not contain that commercial data.

## Required files

```text
INTRO.PRG
INTRO.MOD
MOD2.MOD
TABLE1.PRG
TABLE1.MOD
TABLE2.PRG
TABLE2.MOD
TABLE3.PRG
TABLE3.MOD
TABLE4.PRG
TABLE4.MOD
```

These eleven PRG/MOD files are read-only game data. The shared decoders validate the supported consumed-data layout, not exact whole-file identity. Windows, Linux and the macOS importer/host all use `frontend.LoadConfigured`. See [the compatibility audit](runtime-compatibility-audit.md) for the file-by-file policy and extension rules.

`PINBALL.CFG` is optional mutable settings, separate from commercial payload. A valid six-byte DOS record may seed native settings. Missing or malformed CFG uses native defaults. Native state may contain a valid versioned PFNC record. No CFG hash is required.

## What is decoded

PRG files provide playfield graphics and structured records used by the native implementation, including matrix text and presentation records, lamp/animation data, physics/table records and other table-specific data. MOD files provide the original tracker/module music and samples.

The program never executes original x86 code. Source addresses and record maps identify data to decode; native Go code implements the gameplay behavior.

## Writable state

Native writable data is kept separately from the original installation. Windows/Linux use:

```text
userdata/
```

This includes settings, high scores and logs. If the directory beside the executable/AppImage is not writable, the normal per-user configuration directory is used as fallback. The Windows/Linux CLI provides `-config-dir` and `-high-score-dir` overrides.

macOS imports originals into `~/Library/Application Support/PinballFantasies/Data/` and writes native settings/high scores in the sibling `State/` directory. It does not write `userdata/` beside the app or modify the bundle. The native importer copies only PRG/MOD files and optional CFG, not legacy high-score files. See [macOS storage and import](macos.md).

On Windows/Linux, optional legacy `TABLE*.HI` files may be read as score seeds in a normal asset-free installation. They are never required runtime inputs and are not modified. Without them, the native factory high-score tables are used.

## Personal builds

The local-only personal builder embeds the eleven compatible game-data files above and an optional CFG settings seed into the resulting EXE/AppImage or macOS app. Personal macOS bundles automatically import their Resources/Data contents into Application Support on first launch, preserving an existing valid import and native state. Those artifacts therefore contain the user's commercial game data and must not be distributed as project releases. Public GitHub release binaries are always asset-free.

## Fonts and presentation

Game matrix fonts and presentation records come from the supplied game data where applicable. The small native sidebar font used by the frontend is project-authored code rather than a redistributed PC BIOS or emulator ROM font.

See [provenance](provenance.md) for the licensing boundary.

Supported linked DOS layouts are A (`dos-retail-linked-v1`), Power Pack-marked B (`dos-powerpack-linked-v1`), Deluxe CD-family C (`dos-deluxe-cd-alt-linked-v1`), and the proved Deluxe CD/GOG runtime file-set D (`dos-deluxe-cd-linked-v1`). Each installation must match one coherent profile across all five PRGs. All use the same import UI and engine. B/C/D preserve S_EMPTY priority 0; A retains priority 1. C/D share TABLE relocation metadata, actual 320×123/320×117 startup geometry and typed Stones handler translation. C has separate INTRO source addresses and factory initials. Only A is the pinned canonical oracle. The official 10-minute Party Land demo is a separate five-file profile; see [the demo notes](partyland-10min-demo.md). Other demos and unidentified layouts remain unsupported. No Rev1/Rev2, Gold Pack or Deluxe chronology is asserted. See [C validation](runtime-layout-c-validation.md) and [D validation](runtime-layout-d-validation.md).
