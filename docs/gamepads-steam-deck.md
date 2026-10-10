# Gamepads and Steam Deck

The Linux SDL2 host accepts one standard SDL GameController at a time. It opens
the first recognized controller, handles hotplug, ignores other controllers,
and selects a remaining device after the active controller disconnects.
Unmapped raw joysticks need an SDL mapping; button numbers are not guessed.
SDL_GAMECONTROLLERCONFIG can supply custom mappings. Labels below use Xbox /
Steam Deck positions; A/B/X/Y on PlayStation are Cross/Circle/Square/Triangle.

Actions enter the existing logical input and source runner. Charging remains
one adjustment per source task. Triggers act as digital flipper buttons with
hysteresis. Disconnect or focus loss clears actions without launching the ball.
Controls held across connection, focus or screen changes need release and a
fresh press. Keyboard and controller holds combine. Physics, clock and ABI
remain unchanged.

## Other platforms

The same standard buttons and menu policy below are shared by all four hosts.

| Platform | Native device API | Scope |
| --- | --- | --- |
| Linux / Steam Deck | SDL2 GameController | Recognized standard SDL controllers; first active device |
| Windows | XInput 1.4, fallback 9.1.0 | Xbox/XInput-compatible devices, including compatible Steam Input emulation |
| macOS ARM64 / Intel | Apple GameController | Controllers with an extended gamepad profile |
| Android | InputManager, KeyEvent, MotionEvent | Devices advertising GAMEPAD or JOYSTICK; trigger axes/digital buttons and D-pad keys/hats |

Windows reads snapshots at host input collection, with a 4 ms minimum polling
interval and a one-second scan when no device is connected. A press/release
entirely between two XInput snapshots cannot be observed. Raw DirectInput
controllers need external XInput emulation. macOS callbacks run on the main
queue; background controller events maintain physical history while the engine
remains suspended. Android retains historical motion samples, chooses supported
LTRIGGER/RTRIGGER axes or BRAKE/GAS fallback, and filters other devices. Android
touch UI stays available; a gamepad does not qualify as an external alphabetic
keyboard. Only one active controller per host is supported.

On desktop, initials still use a keyboard. On Android the existing contextual
QWERTY panel remains available. Gyro/stick nudging, vibration and interactive
rebinding are not implemented. PlayStation labels use physical button positions.

## Layout 1: native gamepad

Add the Linux AppImage as a non-Steam game in Steam Desktop Mode, then launch
from Gaming Mode. Select Steam Input's **Gamepad** template. Keep the standard
gamepad outputs; also assigning keyboard keys to these same controls can submit
two actions. Use the Linux AppImage directly.

| Screen | Control | Action |
| --- | --- | --- |
| Intro | A or Start | Skip intro |
| Selector | A / B / X / Y | Party Land / Speed Devils / Billion Dollar Gameshow / Stones ’N Bones |
| Selector | Start | Options |
| Selector | View / Back | Back / quit |
| Options | D-pad Up / Down | Select option |
| Options | A | Change value / accept |
| Options | B or View / Back | Leave options |
| Loaded table attract | A | Start one player, also after an eight-player game |
| Loaded table attract | X | Start / add player through Enter |
| Playing | L1 / L2 | Left flipper |
| Playing | R1 / R2 | Right flipper |
| Playing | X held, then released | Charge plunger, then launch |
| Playing | Y | Nudge (repeated nudges can tilt) |
| Before first launch | A | Add player |
| Playing | Start | Pause |
| Playing / attract | View / Back | Existing Escape action |
| Paused | A or Start | Resume |
| Paused | B or View / Back | Quit question |
| Quit question | A / B | Yes / No |
| Demo closing text | B or View / Back | Close |

The demo accepts only Party Land and retains its original one-player behavior.
Back retains mode-dependent Escape behavior, including the demo's omitted
in-chute quit route. To leave during play, pause and use the quit prompt.

Optional Steam Input rear bindings: L4 = Alt+Enter (fullscreen), R4 = M (music),
L5 = Esc, R5 = F1 (one-player start). Rear buttons and trackpads are not distinct
SDL standard controller buttons. Gyro/stick nudging and vibration are not
implemented. For initials, open Steam's keyboard with **Steam + X** and enter
three letters. Native controller initials selection is not implemented.

## Layout 2: keyboard emulation

This uses existing desktop keyboard input and also works with earlier Linux
builds. In the shortcut's Steam Input editor assign these keyboard outputs
instead of standard gamepad outputs for the same controls.

| Deck control | Keyboard output | Action |
| --- | --- | --- |
| L1 / L2 | Left Shift | Left flipper |
| R1 / R2 | Right Shift | Right flipper |
| X | Down Arrow | Hold to charge, release to launch |
| Y | Space | Nudge / skip intro |
| A | Enter | Start / add player, confirm options, resume |
| B | Esc | Back / quit |
| Start | P | Pause / resume |
| View / Back | F5 | Options in selector |
| D-pad Up / Right / Down / Left | F1 / F2 / F3 / F4 | Select table / 1–4-player start |
| L4 / R4 | Up Arrow / Down Arrow | Navigate options |
| L5 | M | Toggle music |
| R5 | Alt+Enter | Fullscreen |
| Right trackpad | Mouse | Pointer in desktop dialogs |
| Right trackpad click | Left mouse button | Click |

For quit prompts enter Y or N with Steam's keyboard; Enter does not confirm Yes.
For 5–8 players use F5–F8 through an external keyboard or a custom Steam Input
menu. After an eight-player game use F1 rather than Enter to restart. These are
binding instructions, not a published Steam community configuration or a tested
importable VDF preset.

## Validation and next steps

### Contextual controls on screen

A translucent controller panel in a separate strip below the source image shows
table/button identities in the selector,
options navigation, start/player selection, pause and quit actions. Before a
ball is launched it also shows the flipper and charged-launch controls. It
disappears during normal ball play and leaves the entire source image unobstructed.
The 640x240 frontend keeps its doubled-scanline aspect.
These are host presentation pixels; source frames and source game logic are
unchanged. Disconnecting the controller removes the panel.

Linux uses SDL's controller type; macOS uses the GameController profile and
reported face-button symbols; Android uses vendor/device identity. Xbox and
PlayStation controls have their corresponding letters or shapes, with a generic
ABXY fallback. Windows XInput and Steam Input can expose a virtual Xbox device,
so its Xbox glyphs are shown even if the physical pad is a PlayStation pad.
Nintendo profiles use the corresponding physical face-button letters.

Android also highlights its actual launch touch region with a translucent blue
rectangle and “DRAG DOWN / RELEASE TO LAUNCH” instruction while the source
plunger is available. The same rectangle defines touch hits and drawing; it
starts below the upper matrix and avoids the flipper regions. With a controller
connected, the automatic Android menu sheet gives way to controller hints;
the optional menu button and initials input remain available.

Contextual visibility, table identities, glyph overrides, source-frame
immutability, framebuffer reuse, complete source-image preservation and Android draw/hit bounds
are covered by host tests.

Android emulator visual check (2026-10-10): bundled demo, portrait launch zone,
downward drag launching the ball, and automatic removal of the blue rectangle
after launch: PASS.
Windows startup enumeration and focus-loss connection retention have regression
tests; physical startup acceptance remains pending. This does not replace physical-device/controller acceptance.

Pure adapter tests cover repeats, short release edges, trigger hysteresis,
simultaneous aliases, focus/screen boundaries, connection history, disconnect
and menu DOS key mappings. Run the SDL event bridge check with
`python3 tools/test_sdl_gamepad.py` (SDL2 >= 2.0.14, C compiler and pkg-config).
It uses real SDL virtual controllers without game assets or physical hardware.

Local validation on 2026-10-10: shared Go tests on macOS PASS; gamepad/source
race tests PASS; Windows cross-build PASS; public source check PASS; production
SDL C event bridge with SDL2 2.32.10 virtual controllers on macOS PASS. Linux
host build/runtime and GitHub CI are NOT RUN locally. The virtual event check
is included in Linux CI for subsequent runs.

Physical Linux gamepad and Steam Deck acceptance remains: intro/table/options
navigation; all four tables; trigger and shoulder flippers; charged launch;
pause/resume/quit; detach/reconnect while charging; focus loss with buttons held;
initials with Steam's keyboard; both layouts in Gaming Mode; fullscreen and
audio continuity. The adapters are implemented; physical-device acceptance remains pending.

References: [SDL controller API](https://wiki.libsdl.org/SDL2/CategoryGameController),
[SDL device events](https://wiki.libsdl.org/SDL2/SDL_ControllerDeviceEvent),
[Valve non-Steam shortcuts](https://help.steampowered.com/en/faqs/view/671A-4453-E8D2-323C/),
[Valve Deck FAQ](https://www.steamdeck.com/en/faq).

### Additional platform validation (2026-10-10)

- Shared mapper, XInput snapshot identities/edges, and engine ownership/source
  charging/lifecycle tests: PASS.
- macOS ARM64 host build, existing native/ABI tests and GameController snapshot
  decoding with explicit callback delivery: PASS. Snapshot tests do not certify
  hardware event delivery, pairing, controller firmware or physical latency.
- macOS Intel host cross-build/package: PASS; Intel execution NOT RUN.
- Android debug APK (both ABIs): PASS.
- Windows executable cross-build: PASS; real XInput device execution NOT RUN.
- Android ARM64/x86_64 engine libraries and 16 KB ELF checks: PASS. Native JNI
  serialization/lifecycle and standard button mapping tests: PASS.
- Physical controllers on Windows/macOS/Android and Steam Deck: NOT RUN.
