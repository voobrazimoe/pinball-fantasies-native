# Pinball Fantasies v0.1.7

Prerelease: the Android Options menu fits landscape screens. Gameplay is unchanged.

- In landscape, when the screen leaves room to the right of the table, the
  Options menu puts Up and Down in a column there and Select and Back in a strip
  under the OPTIONS MENU heading, so the list itself stays visible. The attract
  screen, pause and question prompts use the same side column.
- Includes everything from v0.1.6: the redrawn Android overlays, the menu tab,
  the compact keyboard, the steady flipper zones, the plunger meter and the new
  8x8 sidebar font.

The full game still needs your own legally obtained DOS files: 11 PRG/MOD files,
plus optional `PINBALL.CFG` and `TABLE*.HI`. Apart from the 10-minute demo,
builds contain no game data.

Packages: Windows x86_64 EXE, Linux x86_64 AppImage, macOS ARM64 and Intel
(13 or later; ad-hoc signed, not notarized) and a universal Android APK
(ARM64 + x86_64, Android 8.1 or later). The APK is development-signed by CI, so
an existing install may need to be uninstalled before installing this one;
reimport your DOS folder afterwards.
