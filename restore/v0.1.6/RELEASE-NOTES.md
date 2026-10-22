# Pinball Fantasies v0.1.6

Prerelease: a new Android interface and a cleaner sidebar font. Gameplay is unchanged.

- Android overlays are redrawn in the selector's own colours, with raised VGA
  panels and the sidebar's 8x8 lettering. Each button shows the DOS key it sends.
- The Android menu is a small tab under the dot-matrix display. Tap it or pull it
  down for a short list: Pause, Music and the keyboard. The tab, Back or a tap
  outside closes the list.
- The Android keyboard (Enter, Esc, arrows, F1–F8 and letters) is a compact board
  with equal-size keys.
- Flipper touch zones are a steady, faint outline that no longer flashes when
  pressed. Pulling the plunger shows a rail, a knob and a charge meter, and the
  knob springs back on release.
- The table selector's sidebar uses a redrawn 8x8 font in the style of the VGA
  text mode on every platform, replacing the rough stand-in for the BIOS font.

The full game still needs your own legally obtained DOS files: 11 PRG/MOD files,
plus optional `PINBALL.CFG` and `TABLE*.HI`. Apart from the 10-minute demo,
builds contain no game data.

Packages: Windows x86_64 EXE, Linux x86_64 AppImage, macOS ARM64 and Intel
(13 or later; ad-hoc signed, not notarized) and a universal Android APK
(ARM64 + x86_64, Android 8.1 or later). The APK is development-signed by CI, so
an existing install may need to be uninstalled before installing this one;
reimport your DOS folder afterwards.
