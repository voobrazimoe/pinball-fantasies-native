# Native macOS arm64 host

The experimental macOS MVP is an Objective-C AppKit shell around the existing
ABI 1 (`cmd/pfengine/abi.h`). It lives on `codex/macos-arm64-host`. The foundation
was separately fast-forwarded and pushed to main at
`8bc62e361832ac2ae76b2da251d335eef533c9cc` after its Linux/Windows gates passed.
No macOS merge is authorized before original-backed and physical-Mac acceptance.
Android work has not started.

Supported build target: **Apple Silicon arm64, macOS 13 or later**. Hosted
validation uses macOS 15. Deployment to macOS 13/14 has not been exercised.
Intel and universal binaries are outside this milestone. This is not a signed,
notarized production distribution.

## Build and CI

Build on an Apple Silicon Mac with Xcode or its matching command-line tools,
Go **1.27.1** (project policy), Python 3.12 or newer, and Git. The local Linux
checkout does not contain an Apple SDK; a Linux cross-build is not validation.

```sh
uname -m                         # must report arm64
sw_vers
xcodebuild -version
xcrun clang --version
go version
./tools/build_macos.sh
```

The script selects Apple's compiler/SDK through `xcrun`, passes the SDK to cgo,
builds both `bin/macos/libpfengine.a` and `libpfengine.dylib`, and keeps the stable
`abi.h` alongside the Go-generated header. The app statically links the archive
using Apple's compiler and frameworks. It also builds a direct-Runner trace
oracle, native tests and portable C host tests, inspects `file`, `lipo` and
`otool`, checks the plist/bundle, ad-hoc signs, and packages:

```text
release/macos/Pinball Fantasies.app/
  Contents/Info.plist
  Contents/MacOS/pinballfantasies
  Contents/Resources/LICENSE.txt
  Contents/Resources/Go-LICENSE.txt
  Contents/Resources/Go-version.txt
  Contents/_CodeSignature/CodeResources
release/macos/PinballFantasies-arm64.zip
```

`.github/workflows/macos.yml` uses the documented standard **macos-15** Apple
Silicon runner, explicitly selects **Xcode 16.4**, pins Go 1.27.1, and fails if
`uname -m` is not `arm64`. Runner images can receive OS updates; the log reports
actual OS, Xcode, compiler and Go versions. Linux/Windows CI remains independent.
The native host build treats compiler warnings as errors.

Hosted CI requires shared Go tests, real Apple-built C archive/shared library,
asset-free ABI contract tests, native host tests, exact nearest-neighbour bitmap
pixels, arm64 executable/bundle/signature checks, public payload scanning,
a bounded two-second AppKit synthetic-frame launch and clean termination, and
an uploaded public `.app` zip. Hosted WindowServer launch worked; the smoke is
bounded to 15 seconds and fails on crash/timeout. It does not import originals,
exercise a game, certify human-visible output or listen to audio.

The Go desktop CLI and `internal/platform` adapter target Linux/Windows; macOS
uses the C archive plus AppKit rather than introducing fake Go platform stubs.
CI runs every other Go package. Tests requiring absent originals/reference
captures explicitly skip. Those skips **do not satisfy** the original-backed
macOS conformance gate. Asset-free native C archive and loaded C shared-library tests check ABI version, missing-data errors
and every invalid-handle export; they do not claim successful gameplay replay.

## Window, time and input

CoreGraphics presents a copied RGBA8 framebuffer with black bars and
`kCGInterpolationNone`. Storage grows only when geometry requires more capacity;
CGImage/provider wrappers are bounded and released after drawing. The copy is
necessary because ABI frame retrieval can invalidate its borrowed pointer.
The original 640×240 selector is presented as 640×480, matching desktop policy;
other framebuffers preserve their source aspect. No Metal, SDL or gameplay
renderer is introduced.

A non-spinning 120 Hz AppKit timer in common run-loop modes supplies
`mach_absolute_time()` converted by `mach_timebase_info` to monotonic nanoseconds.
Integer conversion uses 128-bit arithmetic to avoid overflow. The Go Runner
owns 60/71 Hz scheduling and drains every due task; presentation can skip frames
without dropping source ticks. Resize and AppKit fullscreen do not change source
time or physical monitor modes. Option-Return and View → Toggle Full Screen
(Command-F) invoke native fullscreen, with a pending-command guard and native
restoration callbacks.

Native physical keycodes translate to the existing DOS makes. Left/right
Shift, Control and Option hold the corresponding flippers; Down pulls the
spring, its break releases, Return supplies the separate release edge, Space
holds tilt, and P/Escape/F1–F8 use shared pause/menu/start/quit semantics.
Command-Q remains the native quit shortcut. Repeats and duplicate keyDowns are
filtered. Cheat letters follow physical QWERTY make mapping, exactly as the
shared DOS key boundary; no cheat menu is added. Unknown ordinary keys map to
the shared generic resume/quit make.

Modifier-only input uses `flagsChanged` and Apple's IOKit device-side masks,
not aggregate flag toggles or Win32 pairing logic. Independent contributors
are combined before issuing held actions. Automated tests exercise both-side
press/release, mixed modifiers, repeat/shortcut suppression and ordered SNAIL
makes. Physical keyboard layouts/devices still need acceptance.

Mouse motion reads the native relative `kCGMouseEventDeltaY`, without screen-Y
strength, warp, disassociation or capture. It sends counts to `plunger_delta`;
only shared Go validity/physics decide canonical SpringPosition. The left button
sends one fire make until release. Mouse policy uses the engine's active-plunger
flag and the pointer being inside the key view. Paired `NSCursor hide/unhide`
calls apply only during that local context and restore on exit, focus loss,
pause/invalidity and shutdown. The pointer remains free to leave the view; no
permanent trap or global hiding is installed. Screen-edge limits and real mouse
sensitivity/cursor feel require physical acceptance.

Loss of app/key-window focus, minimization and system sleep suspend the engine,
clear host controls and stop/flush audio. Suspension applies the shared logical
pause. Regain/wake re-anchors using the monotonic clock, clears stale controls,
and leaves the game paused until ordinary input resumes it. No suspended wall
time is used for background catch-up. Fullscreen uses the same native focus
notifications; its physical multi-display behaviour remains unverified.

## Audio

Each synchronous Go PCM callback copies 48,000 Hz stereo signed 16-bit samples
into a fixed one-second single-producer/single-consumer ring. The main thread is
the sole engine caller. The AudioUnit realtime callback only consumes that ring,
fills silence on underrun, and never allocates, waits, logs or calls Go.
`DefaultOutput` is configured for the engine format and performs any conversion
required by the hardware; neither device rate nor callback cadence controls game
or tracker time. Left/right channel order is retained.

Pause/focus loss stop the AudioUnit before resetting queued PCM; resume starts
with fresh source output. Default-output-device changes are delivered to the
main queue, where the unit is closed/reopened without modifying game state.
Queue backpressure waits in one-millisecond increments while audio drains. A
missing or stalled device cannot hang the UI: a full ring drops remaining PCM
after 100 waits and counts it; source updates still run. Underruns/dropped frames
are logged at shutdown. This bounded failure policy is not a claim of underrun-
free output. Real device output, sustained cadence and route recovery remain
physical-Mac acceptance items.

## Import and persistence

The public bundle contains no commercial files and offers only native import
until valid owner-supplied originals exist. NSOpenPanel accepts the containing
folder or a file within it. The required set is INTRO.PRG, INTRO.MOD, MOD2.MOD,
and TABLE1–4.PRG/MOD. PINBALL.CFG is copied when present; commercial executables,
drivers, existing mutable TABLE*.HI and unrelated files are not copied.

Files must be regular, non-symlink inputs. A sibling staging directory receives
the allowlisted copies. Validation creates/destroys a temporary ABI engine,
using the existing Go hashes/decoders/settings semantics. Only a fully validated
set is moved/replaced as app-owned Data; a failed import preserves the previous
set. Arbitrary external paths are no longer needed after import. No files are
downloaded. Native tests check staged replacement, missing files, symlinks,
failed ABI validation and the whitelist.

Foundation chooses the user-domain application support directory:

```text
~/Library/Application Support/PinballFantasies/Data/
~/Library/Application Support/PinballFantasies/State/PINBALL.CFG
~/Library/Application Support/PinballFantasies/State/TABLE1.HI … TABLE4.HI
```

Explicit Data/State paths go to the existing ABI. Settings/high-score encodings,
factory defaults, save/restart behaviour and hotseat semantics remain shared Go
code. State never targets CWD, originals or `.app`. Interactive first-launch
import and persistence across real app launches remain unverified.

## Original-backed Mac validation and manual acceptance

The owner has an authorized Apple Silicon Mac with Codex for the next session.
Original-backed macOS tests are **PENDING**, distinct from green hosted checks.
With a pristine owner-supplied original directory and installed Go 1.27.1:

```sh
python3 tools/validate_macos_originals.py --data /absolute/path/to/originals
# Optional historical sources can run additional Go reference checks:
python3 tools/validate_macos_originals.py --data /absolute/path/to/originals \
  --reference /absolute/path/to/original-dos-source
```

This command refuses non-Mac/non-arm64 hosts, checks the pinned original hashes,
builds with Apple tools, runs the actual C-library direct-Runner comparison for
all four tables in SOFT/OFF (including cheats/lifecycle/PCM/frame checkpoints),
runs native four-table journeys in HARD/MEDIUM/SOFT/OFF, with NORMAL and HIGH resolution, runs shared Go tests in
a clean temporary checkout with external originals, exercises shared settings
and high-score restart validation, scans the app for original blocks, and checks
that original inputs are unchanged. Missing optional historical captures still
skip with reasons. It never uploads commercial inputs. Preserve its output as
local acceptance evidence; do not put originals in Git or Actions artifacts.

Then launch the real app, import from NSOpenPanel and complete these currently
**UNVERIFIED** physical-Mac checks:

- Actual visible framebuffer/window rendering, resize/aspect/crisp pixels and
  every table in HARD, MEDIUM, SOFT and OFF/full-table at NORMAL and HIGH resolution.
  HARD is the scrolling label exposed by shared settings (the task called it HIGH).
- Keyboard feel, held/released controls, hotseat, pause/quit, original cheat
  sequences, and sided modifiers on real hardware/layouts.
- Relative mouse-plunger feel, one-shot fire, source-invalid input discard,
  cursor hide/restore, leaving the view and focus loss.
- Real CoreAudio output, channel order, underruns, pause/resume, prolonged
  suspension, system sleep/wake and default-output-device changes.
- Native fullscreen entry/restoration on one and multiple displays, with no
  display-mode switch or unwanted held controls.
- NSOpenPanel interactive import, cancellation, corrupt/missing files and
  independence from the selected external folder after successful import.
- Settings/high-score persistence across real app launches from different CWDs,
  with writes confined to Application Support and no changes inside `.app`.

The feature branch must remain unmerged until these original-backed/hardware
gates are completed and the user authorizes integration. No Android implementation
or fake release/tag is part of this milestone.

## Signing

CI uses `codesign --sign -` and verifies the ad-hoc signature. It has no Developer
ID credentials and performs no notarization. Wider distribution would require
an actual Developer ID certificate/private key, appropriate hardened-runtime
signing configuration and Apple notarization credentials/submission/stapling,
followed by Gatekeeper testing. Ad-hoc build artifacts may require owner approval
in macOS security UI; signing/notarization and physical acceptance are not
represented as completed production-distribution work.
