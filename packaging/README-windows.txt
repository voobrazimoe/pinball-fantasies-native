Pinball Fantasies native - Windows x86-64

Put your legally obtained original game files beside pinballfantasies.exe:
INTRO.PRG, TABLE1.PRG through TABLE4.PRG, INTRO.MOD, MOD2.MOD,
TABLE1.MOD through TABLE4.MOD. PINBALL.CFG is an optional legacy settings seed;
missing or malformed settings use native defaults. TABLE1.HI through TABLE4.HI are optional legacy score seeds;
without them, native factory defaults are used. No original game assets are included.

Double-click pinballfantasies.exe. No installation or extra DLL is required.
Alt+Enter switches between a window and borderless desktop fullscreen.
Losing focus pauses gameplay. After returning, press P (or a gameplay key) to resume.

Native options, high scores and native.log go to userdata beside the executable.
If that folder cannot be written, state goes to %APPDATA%\PinballFantasies.
Original supplied files are read only and are never updated in place.
Explicit overrides: -data-dir PATH -config-dir PATH -high-score-dir PATH.
Move the game to a writable folder if neither default state location is writable.

Pre-release/beta: Issue #1 remains open pending owner retest on real Windows.
Repeated Alt+Enter transitions must be confirmed audibly clean.
Also retest held/released flippers, focus-loss pause, full-table nudge and
numeric matrix presentation. Source readiness is separate from binary acceptance.

The flipper held-input blocker also remains pending a real Windows retest.
To collect input diagnostics in PowerShell before launch:
  $env:PF12_INPUT_LOG = '1'
  .\pinballfantasies.exe
input.log is written beside native.log; it includes VK, scan, extended flag,
make/break, focus and physical held bits. Remove PF12_INPUT_LOG after the retest.
Optional PF12_INPUT_LOG_PATH selects a specific writable log file.
