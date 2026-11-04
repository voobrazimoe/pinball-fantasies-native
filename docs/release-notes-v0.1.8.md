# Pinball Fantasies v0.1.8

Gamepads on every platform, smoother Windows frames and a cleaner Android
plunger. Gameplay and physics are unchanged.

- Gamepads: SDL2 GameController on Linux and Steam Deck, XInput on Windows,
  Apple GameController on macOS and controller input on Android. Shoulders or
  triggers work the flippers, X charges and releases the plunger, Y nudges and
  Start pauses. Contextual control hints are drawn over the frame.
- Windows reads DualSense, DualSense Edge and DualShock 4 natively over USB or
  Bluetooth, without Steam or DS4Windows.
- Windows now paces frames to the monitor's vertical blank, which removes the
  slight stutter. `PF12_VSYNC=vblank|dwm|off` selects the pacing method.
- Android: the stray amber strip beside the Pull control while pulling is gone.

See [gamepad controls and Steam Deck layouts](https://github.com/voobrazimoe/pinball-fantasies-native/blob/v0.1.8/docs/gamepads-steam-deck.md).

The full game still needs your own legally obtained DOS files: 11 PRG/MOD files,
plus optional `PINBALL.CFG` and `TABLE*.HI`. Apart from the 10-minute demo,
builds contain no game data.

Packages: Windows x86_64 EXE, Linux x86_64 AppImage, macOS ARM64 and Intel
(13 or later; ad-hoc signed, not notarized) and a universal Android APK
(ARM64 + x86_64, Android 8.1 or later). The APK is development-signed by CI, so
an existing install may need to be uninstalled before installing this one;
reimport your DOS folder afterwards.
