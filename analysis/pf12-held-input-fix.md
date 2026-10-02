# PF12 Windows held input and focus pause

Status: implementation and automated validation complete; **the real-Windows
flipper blocker stays open until physical-key retesting**. Issue #1 remains open: repeated real-Windows
windowed/borderless audio transitions require an owner retest of the current build.
No issue is closed by this fix; prior acceptance wording was unsupported.

## Held input

The Win32 host previously derived held controls solely from its accumulated
`windowsKeys.down[256]` make/break history. A missing break could keep a flipper
raised until focus loss. Physics already releases on false Left/Right input and
is unchanged. The companion full-table presentation fix changes only rendering
in `internal/physics/rules_boundary.go`; simulation and flipper physics are unchanged.

`Pump()` and `Held()` now sample `GetAsyncKeyState`'s high bit for VK_LSHIFT,
VK_RSHIFT, VK_LCONTROL, VK_RCONTROL, VK_LMENU, VK_RMENU, VK_DOWN and VK_SPACE.
The low historical bit is ignored. See the [Microsoft API contract](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getasynckeystate).
Each sided modifier is independent; overlapping modifiers OR into Left/Right.
A physical press can supply a held control even without a received make.

Held bits require both WM_SETFOCUS state and current GetFocus/GetForegroundWindow
ownership. WM_KILLFOCUS clears the complete message history, discards queued
input while retaining close, and generates the existing focus-loss event. Global
keyboard state never supplies gameplay held bits while the window lacks focus.

Reconciliation clears physically released control entries from `down[]` before
key decoding as well as at pump/tick boundaries. Thus stale Alt cannot turn a
later plain Enter into Alt+Enter. Polling does not insert makes into message
history, fabricate release events or mark the first queued Space/Down make as a
repeat. The existing scan/extended decoding, F1–F5, letter makes, Enter/Down
release events and repeat suppression remain intact. A real queued Alt+Enter
retains its message-time Alt context bit even if Alt is physically released
before the message is pumped.

## Pause on Windows and Linux

Win32 WM_KILLFOCUS and SDL_WINDOWEVENT_FOCUS_LOST feed the shared frontend
FocusLost input. Active Playing becomes Paused before session sync, regardless
of the manual P debounce. Focus loss takes precedence over queued gameplay keys
and release edges, is idempotent when already paused, and preserves close.
Returning focus leaves gameplay paused until an ordinary gameplay make resumes
it. Existing pause behavior suspends PCM production and the audio device.
Selectors/options retain their existing behavior.

The standalone physics diagnostic hosts pause on focus loss too; this is host
scheduling only. Linux frontend held controls additionally require the SDL
window's current keyboard-focus flag.

## Temporary opt-in trace

PowerShell, from the directory containing the new EXE:

```powershell
$env:PF12_INPUT_LOG = '1'
.\pinballfantasies.exe
# After testing:
Remove-Item Env:PF12_INPUT_LOG
```

`input.log` is appended beside `native.log` in the resolved state directory
(normally `userdata`; explicit `-config-dir` and the existing fallback apply).
Optional `PF12_INPUT_LOG_PATH` specifies a writable file path. Disabled tracing
opens no log. Each key entry records raw VK, scan, extended, make/break,
message-time Alt/previous state, focus, physical bits and gameplay held bits.
Pump and Held entries record reconciliation, including stale message-held bits.
The log header defines the per-key bit mask; trace write failures do not stop
normal gameplay. Keep tracing enabled only during retesting.

## Validation

- Lost-break regression for all eight controls: make without break, then
  physical false returns zero and clears the history.
- All 256 physical combinations: sided modifiers, simultaneous modifiers,
  Down/Space and controls pressed without a make; zero held without focus.
- Independent releases, complete focus-history clearing, stale Alt followed by
  plain Enter, legitimate message-context Alt+Enter, makes/repeats, F1–F5 and
  Enter/Down release events pass on Linux and as Windows binaries under Wine.
- The opt-in native Win32 window smoke exercises actual Held/GetAsyncKeyState
  after injected lost breaks with no physical key down; resize/fullscreen/focus
  and existing Win32 contract tests pass under isolated Xvfb/Wine.
- The shared frontend regression verifies no session/ball/score/mixer advance
  during focus pause, repeated losses, queued makes, regain, manual resume and
  close priority. The native Win32 frontend pause/resume journey passes under
  Wine. Physical XTest focus changes in the freshly built Linux AppImage also
  verify loss -> pause, regain -> still paused, manual resume, and loss/regain
  while already paused. Wine's XTest key delivery did not support the analogous
  physical UI journey; that is not claimed as real-Windows input evidence.
- Full Go suite has only the pre-existing documented local TABLE1.HI inventory
  mismatch. A second full run excluding exactly that inventory subtest passes;
  no inventory expectations or originals were changed.
- Reachability: zero missing targets on all four tables. Operand/text/writer
  checkers: zero unresolved operands, mismatched source text or missing writers.
  Matrix operand schema tests pass. Every Windows test package cross-compiles.
- Linux and Windows/Wine Party oracle: 1200 ticks, score 2,300,000, ball 2,
  frame SHA256 `aec01b3a07e1a5ba10b3c635777a6742f4abd41c09899903ea532b98d8522913`.
  PCM/frame host-cadence regression passes on both targets.
- Normal Linux build, public Windows GUI EXE and personal Windows GUI EXE/Linux
  AppImage build. The 12 personal build inputs retain their pre-build hashes.

Local evidence is in ignored `.build-personal/held-input-*.log` and
`.build-personal/release-*.log`; generated
binaries and bundled commercial data remain ignored and are not pushed.

Build the Windows personal artifact without Linux packaging dependencies:

```sh
./tools/build_personal_release.sh . --windows-only
```

Build both personal artifacts with the existing default command:

```sh
./tools/build_personal_release.sh .
```

Artifacts:

- `release/personal/windows/pinballfantasies.exe` (self-contained personal data)
- `release/personal/linux/PinballFantasies-x86_64.AppImage` (self-contained personal data)
- `release/windows/pinballfantasies.exe` (public build requiring external data)

## Required real-Windows retest

Test each left/right Shift, Ctrl and Alt, simultaneous modifiers and rapid
alternating presses/releases. After every release, verify the associated
flipper falls; include Down and Space. Alt+Tab away while controls are pressed,
release outside the game, return, confirm the game stays paused, then resume.
A plain Enter after released Alt must not toggle fullscreen; deliberate
Alt+Enter must still toggle once per make, with repeat suppressed. Retain
`input.log` from any stuck event. Wine evidence is additional automated evidence
and does not close the real-Windows blocker.
