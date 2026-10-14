# Pinball Fantasies v0.1.5

Prerelease — import and polish fixes; gameplay is unchanged.

- Original files are found regardless of letter case, as on DOS, so lower-case
  copies load on Linux and import on macOS. Folders with two case variants of
  one file name are rejected as ambiguous.
- macOS and Android now import optional `TABLE1.HI`–`TABLE4.HI` with the full
  game. As on Windows and Linux, they only seed the starting high-score tables;
  new scores are still saved separately. The demo keeps no high scores.
- Linux shows the game window through a vsynced GPU renderer when available,
  falling back to the previous software renderer.
- Lower CPU use while drawing tables, and on Android when nothing changed
  between display refreshes. Frames are pixel-identical to v0.1.4.
- A failed settings or high-score write is logged instead of ending the game.

The full game still needs your own legally obtained DOS files: 11 PRG/MOD files,
plus optional `PINBALL.CFG` and `TABLE*.HI`. Apart from the 10-minute demo,
builds contain no game data.

Packages: Windows x86_64 EXE, Linux x86_64 AppImage, macOS ARM64 and Intel
(13 or later; ad-hoc signed, not notarized) and a universal Android APK
(ARM64 + x86_64, Android 8.1 or later). The APK is development-signed by CI, so
an existing install may need to be uninstalled before installing this one;
reimport your DOS folder afterwards.
