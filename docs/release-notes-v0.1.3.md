# Pinball Fantasies v0.1.3

Prerelease — Android joins Windows, Linux and macOS.

- Android: touch-first contextual controls, external keyboard support, and a clean first-run folder importer.
- Android audio uses Oboe with pause/resume, audio-focus and route handling.
- GLES presentation is paced by Choreographer for smoother display updates.
- The universal Android APK includes ARM64 and x86_64, with 16 KB native-library alignment.
- A new, independently drawn PF application icon appears across all four platforms.

Your own legally obtained supported canonical DOS files are still required:
11 PRG/MOD files, plus optional `PINBALL.CFG`. Android imports the selected
folder through the system picker and copies validated files into private app
storage. Public builds contain no commercial game data. Other DOS layouts,
including `21STCENT/FANTASY`, are not supported.

Desktop packages remain available for Windows/Linux x86_64 and macOS ARM64/Intel.
macOS requires 13 or later; its apps are ad-hoc signed, without Developer ID
signing or notarization. The Android prerelease APK is development-signed.
