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
PINBALL.CFG
```

The files are validated against supported fingerprints and treated as read-only inputs.

## What is decoded

PRG files provide playfield graphics and structured records used by the native implementation, including matrix text and presentation records, lamp/animation data, physics/table records and other table-specific data. MOD files provide the original tracker/module music and samples.

The program never executes original x86 code. Source addresses and record maps identify data to decode; native Go code implements the gameplay behavior.

## Writable state

Native writable data is kept separately from the original installation:

```text
userdata/
```

This includes settings, high scores and logs. If the directory beside the executable/AppImage is not writable, the normal per-user configuration directory is used as fallback. `-config-dir` and `-high-score-dir` provide explicit overrides.

Optional legacy `TABLE*.HI` files may be read as score seeds in a normal asset-free installation. They are never required runtime inputs and are not modified. Without them, the native factory high-score tables are used.

## Personal builds

The local-only personal builder embeds exactly the 12 required files above into the resulting EXE/AppImage. Those artifacts therefore contain the user's commercial game data and must not be distributed as project releases. Public GitHub release binaries are always asset-free.

## Fonts and presentation

Game matrix fonts and presentation records come from the supplied game data where applicable. The small native sidebar font used by the frontend is project-authored code rather than a redistributed PC BIOS or emulator ROM font.

See [provenance](provenance.md) for the licensing boundary.