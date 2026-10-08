# Pinball Fantasies v0.1.4

Prerelease — more original DOS installations work, on every platform.

- Four DOS installations are now supported: the original retail release, the
  Power Pack `21STCENT/FANTASY` set, and both Deluxe CD `PFD/FANTASY` sets
  (including the GOG copy). The layout is detected automatically when you pick
  the game folder; nothing else changes.
- All four tables use the same rules and physics on every installation. Each
  keeps its own edition details: the Deluxe CD sets have their own startup
  artwork and factory high-score initials, and Power Pack and Deluxe CD sets
  keep their original music-cue priority, so a few cues can play differently
  from the retail release.
- The official 10-minute Party Land DOS demo is playable from a folder holding
  only its five files, with its own intro, table select, options and closing
  text.

Your own legally obtained DOS files are still required: 11 PRG/MOD files, plus
optional `PINBALL.CFG`. Public builds contain no commercial game data.

Packages: Windows x86_64 EXE, Linux x86_64 AppImage, macOS ARM64 and Intel
(13 or later; ad-hoc signed, not notarized) and a universal Android APK
(ARM64 + x86_64, Android 8.1 or later). The APK is development-signed by CI, so
an existing v0.1.3 install may need to be uninstalled before installing this one;
reimport your DOS folder afterwards.
