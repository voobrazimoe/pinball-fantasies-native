Fix Windows Alt flipper input: left and right Alt are now driven by their Win32 key messages. Physical polling only clears released keys and no longer activates the opposite flipper without a key-down message.

Includes the public documentation cleanup and restored asset-free test fixtures.

Public binaries require legally obtained original game data, as described in the README. No original game data is bundled.

These Windows and Linux binaries are the unchanged original v0.1.1 artifacts. Before migration they passed Go tests, Windows input tests under Wine, frame-export smoke checks, and commercial-payload checks. Physical Windows gameplay is not certified by these checks. Fresh source CI in this repository is tracked separately.

Downloads: Windows x86_64 EXE, Linux x86_64 AppImage, and SHA256SUMS.txt.
