# Android host

A6 physical-device preparation and the owner-run checklist are in
[android-a6-acceptance.md](android-a6-acceptance.md). Preparation does not mark
A6 complete; hosted CI cannot certify original-backed physical acceptance.

Android support is being added as a native host around the existing Go engine and stable C ABI 1. The Android host must not duplicate table rules, presentation programs, tracker progression, settings/high-score encodings, source cadence or gameplay semantics.

## Development status (2026-10-04)

Development is on `android-host`, based on desktop `main` at `a848e43`.
This is an asset-free platform prototype, not a playable Android release.

- A0 passed [hosted CI at `8ac4030`](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37189211052):
  ARM64/x86_64 Go shared libraries, 16 KB LOAD alignment, and ABI/missing-data
  smoke on the x86_64 16 KB runtime.
- A1 GameActivity, immersive UI, nearest GLES2 synthetic 320x609 texture,
  letterboxing and orientation handling are implemented. The debug APK builds
  for both ABIs and passes native ELF/ZIP alignment checks.
- The initial A1 runtime run failed in the API 35 ps16k image's `system_server`,
  with a SIGSEGV in precompiled `services.odex` at
  `AppIdSettingMap.getSetting`. The CI guest now restarts ART in `-Xint` mode
  before installation. Native rendering still executes normally; this smoke
  cannot certify Java performance or a normal compiled Android runtime.
- Inspecting the built APK also exposed a missing GameActivity JNI initializer:
  linking only `GameActivity_onCreate` did not pull the JNI implementation out
  of the static archive. CMake now retains
  `Java_com_google_androidgamesdk_GameActivity_initializeNativeCode` explicitly,
  following the [GameActivity integration guidance](https://developer.android.com/games/agdk/game-activity/migrate-native-activity).
  CI checks this symbol, `GameActivity_onCreate` and `android_main` in both
  packaged host libraries before attempting to launch the app.
- The host stops rendering while the Activity is paused. The runtime smoke
  requires new log evidence for portrait → landscape → portrait, the same
  process/native host across rotation and background/resume, a presented frame
  after resume, and a fresh EGL frame after process restart.
- A1 passed [hosted CI at `c72f018`](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37199738092):
  both-ABI APK/symbol/alignment checks, actual synthetic frame presentation,
  both orientation transitions, same-host background/resume, and fresh-process
  presentation on the 16 KB emulator with interpreted ART. A0, asset-free source
  and macOS native-host workflows also passed at that commit. A1 is an empty
  native-shell gate only; normal compiled-runtime and physical-device testing
  remain outstanding.
- A2 passed [hosted Android CI at `e6c5e70`](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37202017313):
  both packaged engine ABIs, engine/host symbols, ELF/ZIP 16 KB alignment,
  commercial filename scan, import transaction tests, packaged-engine JNI rejection,
  and the retained A1 rendering/rotation/lifecycle checks. [A0](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37202017296),
  [asset-free source](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37202017329),
  and [macOS native hosts](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37202017274)
  also passed. A2 supplies SAF import, private Data/State, shared-loader validation
  and persistent engine creation. Successful original-backed Android import/bootstrap
  and physical-device acceptance remain unverified for A6. A3 implementation is described below; A4 output is implemented below; A5 focus and route coordination is implemented below; A6–A7 remain outstanding.

- A3 implementation passed [hosted Android CI at `806d4de`](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37206531433):
  both-ABI APK/exports/16 KB checks, payload scan, A2 import transactions, A3
  controls/frame/native-session tests, real MotionEvent instrumentation and
  no-data rotation/lifecycle/keyboard/Back smoke. Required boundary regressions
  also passed: [A0](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37205273560),
  [shared source matrix](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37205273587),
  and [native macOS](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37205273566).
  Private-original Go render/cadence comparisons passed locally; original-backed
  Android gameplay and physical-device acceptance remain unverified for A6.

Build the engine first with an installed Go 1.27.1, SDK/NDK/JDK and Gradle 9.6.0:

```sh
export ANDROID_NDK_HOME=/path/to/android-ndk
# On macOS also set PF_ANDROID_NDK_HOST_TAG=darwin-x86_64.
sh tools/build_android_engine.sh all
sh tools/test_android_import.sh
gradle -p hosts/android :app:assembleDebug :app:assembleDebugAndroidTest
sh tools/run_android_host_smoke.sh
python3 tools/test_check_android_elf.py
```

The alignment checker needs no SDK. The transaction tests need only JDK 17.
The smoke rotates and restarts the app; use a dedicated emulator/device.

## Private personal APK (local only)

With the Go/NDK/JDK/Gradle prerequisites above installed and `gradle` on PATH:

```sh
python3 tools/build_personal_android.py /path/to/originals
```

Output: `release/personal/android/pinball-fantasies-personal.apk`.
This debug-signed **Pinball Fantasies (Personal)** APK contains commercial data
supplied by the owner. Private/local use only: never upload or redistribute it,
including CI artifacts, issues or GitHub Releases. There are no A7 release
signing keys. Tap the APK to use ordinary Android package installation (allow
installation from that source when Android prompts); adb and SAF are not needed
for ordinary personal use. The application ID and debug signing identity remain
the same; preserve existing app State when installing updates.

The builder reuses `personal_assets.inventory()` before copying inputs, packages
only its 11 required files and optional CFG under `assets/personal-data/`, and
never includes TABLE*.HI. It checks Git ignores for all private destinations,
checks copied bytes against the validated inventory, rebuilds both engine ABIs,
and opts into a temporary asset directory and isolated Gradle output. Temporary
inputs and merged asset/build outputs are removed on success or failure. Original
hashes stay only in ignored `.build-personal/android-manifest.json`. The console
reports the APK path/hash, filenames/sizes and source commit. A separate ordinary
public debug build must pass the asset-free ZIP check before success is reported.
Normal Gradle builds remain asset-free, even while the private payload exists.

On first launch, embedded data uses the existing A2 staging, real-engine
validation and transactional adoption into `noBackupFilesDir/Data`. Valid existing
Data takes precedence; `filesDir/State` is retained. Import is available only in
the central no-data shell; normal engine modes have no import/replacement command. Bundled gameplay/device checks can
inform A6, but do not certify real SAF import; see the separate
[acceptance record](android-a6-acceptance.md). Overall A6 remains NOT TESTED.

## A2 import and engine bootstrap

Tap **Import DOS folder** to open Android's `ACTION_OPEN_DOCUMENT_TREE` picker.
Select the folder containing the original DOS runtime files. Only direct children
are enumerated; subdirectories are not traversed. Provider names are matched
case-insensitively and ambiguous duplicates are rejected. Only these files are
copied, with canonical uppercase private filenames:

- Required: `INTRO.PRG`, `INTRO.MOD`, `MOD2.MOD`, and `TABLE1.PRG`/`TABLE1.MOD`
  through `TABLE4.PRG`/`TABLE4.MOD` (11 files).
- Optional: `PINBALL.CFG`; it is user state and is not pinned to a pristine hash.

Executables, `TABLE*.HI`, other files and directories are ignored. Cursors,
provider input streams and private output streams use try-with-resources.
No persistable URI permission is needed after copying. There is no external
storage output and no commercial fallback in public APKs or tests. A local personal APK can supply
the embedded source described above.

Exact paths (Android resolves the device/user-specific prefixes):

| Purpose | Path |
| --- | --- |
| Adopted commercial inputs | `getNoBackupFilesDir()/Data/` |
| Persistent native player state | `getFilesDir()/State/` |
| Fresh candidate input | `getNoBackupFilesDir()/Data.staging-<random>/` |
| Disposable validation state | `getNoBackupFilesDir()/State.validation-<random>/` |
| Recoverable previous inputs | `getNoBackupFilesDir()/Data.previous/` |

The Java worker copies and syncs selected files into a fresh staging directory.
JNI sends absolute normal filesystem paths as standard UTF-8 byte arrays, with a
native guard against URIs, relative paths and embedded NULs. C++ links the existing
`cmd/pfengine` shared library, checks `pf_engine_abi_version() == PF_ABI_VERSION == 1`,
and calls `pf_engine_create(candidate, disposableState, CLOCK_MONOTONIC ns, error,
1024)`. A zero handle rejects the candidate. The bounded shared-loader error is
logged and shown in a Toast. A successful validation handle is destroyed before
validation State is deleted. There is no Android PRG/MOD parser.

After validation, the worker closes the old persistent engine, renames `Data`
to `Data.previous` if present, then atomically renames the staging directory to
`Data`, all on the same private filesystem. The second rename is the commit.
If it fails, the previous directory is renamed back and its engine is bootstrapped.
Failure of the first rename leaves Data untouched. No deletion of adopted Data
occurs before commit. If rollback itself is prevented by a filesystem error, the
old directory remains at `Data.previous` for relaunch recovery. Backup deletion is
best-effort after commit. On relaunch, an absent Data is restored from the backup;
if both exist the committed Data wins. Stale staging/validation directories are
removed without following symlinks. This supports process interruption recovery;
power-loss durability of directory metadata remains device-dependent.

Picker cancellation performs no transaction. Enumeration, copy or validation
failure preserves adopted Data and persistent State; validation never receives
live State. Successful adoption creates one persistent native engine using
`Data`, `filesDir/State`, and monotonic time. Native session tokens and a mutex
serialize creation/destruction and prevent an old Activity's worker from creating
another persistent instance after close. Activity destruction closes the native
handle. A3 advances that same handle. Relaunch bootstraps adopted Data automatically;
public APKs without Data stay in the shell, and malformed/incomplete Data reports
failure. Personal APKs validate existing Data using disposable State first and
only fall back to their embedded source if existing Data is absent or invalid.
An engine/bootstrap failure after commit leaves validated Data installed and
reports the error; importing again does not require deleting it.

Log events distinguish `A2_SHELL_NO_DATA`, `A2_IMPORT_REQUESTED`,
`A2_IMPORT_CANCELLED`, `A2_IMPORT_REJECTED`, `A2_CANDIDATE_VALIDATED`,
`A2_DATA_ADOPTED`, `A2_ENGINE_BOOTSTRAPPED`, and `A2_BOOTSTRAP_REJECTED`.
The no-data shell retains the A1 synthetic renderer. A3 adds gameplay controls
and real framebuffer upload; A4 adds host audio output.

Asset-free hosted checks build/package both ABIs, check exported engine/host
symbols and every packaged ELF's 16 KB alignment, run ZIP alignment and commercial
filename scans, and retain the A1 rendering/rotation/lifecycle smoke. JDK tests
exercise filtering, optional CFG, closed streams, incomplete/ambiguous sets,
cancellation, copy/validation/rename failures, successful adoption, crash recovery
and safe cleanup using invented bytes and an injected validator. Instrumentation
calls the **packaged real engine** through JNI with missing/malformed invented
inputs, requires bounded rejection, and tests the filesystem-path guard. These
checks do not substitute for a successful original-backed import.

A6 physical-device acceptance must still select real originals through providers,
verify successful validation/adoption/bootstrap and automatic relaunch, optional
CFG and persistent State separation, cancellation/rejection preserving a prior
installation, and normal compiled ART on both supported architectures. Do not
commit those inputs or upload original-backed private directories to CI.

## A3 cadence, framebuffer and input

A3 adds real engine advancement and presentation to the persistent A2 handle.
No SAF, import transaction, Data/State layout or bootstrap policy is redesigned.
The native GameActivity glue thread owns `pf_engine_advance`, `pf_engine_frame`
and EGL/GLES. Each active display-loop wake supplies `CLOCK_MONOTONIC` nanoseconds
rather than a tick count or frame delta. The existing Go `source.Runner.Advance`
executes every due source task at its existing 60/71 Hz cadence, including when
presentation is slower. EGL swap interval and the host's 16 ms poll timeout only
pace presentation/wakes. Neither creates an Android simulation clock.
`pf_engine_advance` now receives A4’s synchronous copy-only PCM sink. The engine
and tracker progression and monotonic Runner deadlines are unchanged. Device
callbacks consume host storage independently.

### Ownership and lifecycle

`engine_host.cpp` has one mutex shared by **every** persistent-engine operation:
A2 worker create/validate/destroy, UI input, UI lifecycle, and native advancement,
frame retrieval and copy. UI callbacks translate events and call the serialized
JNI path; they do not advance source time. The mutex remains held across
advance → frame → validation → copy, so replacement/destruction cannot invalidate
the borrowed pointer mid-copy. No sink reentry and no concurrent ABI calls occur.
GLES receives only the host-owned copy after unlocking; all EGL operations stay
on the glue thread. JNI session tokens reject callbacks from a closed Activity.

Pause or window focus loss synchronously calls `pf_engine_suspend` under that
same mutex **before** clearing Java pointer/key ownership. The existing engine
suspend operation clears held controls, mouse remainder, pending delta/fire and
Runner input, applies the shared focus-pause request, and stops source/tracker
progression. Resuming calls `pf_engine_resume` with fresh monotonic time. The
Runner re-anchors its deadline and preserves the gameplay instance; shared game
pause semantics still apply (a logical key resumes a paused game). The render
loop sleeps while paused or without a surface. Java window focus is the sole
engine focus authority; native glue focus commands are diagnostic rather than
a second presentation gate. While unfocused, a remaining surface may present
its cached frame, but no engine input/advance/frame calls occur. Rotation and EGL surface replacement
only replace graphics resources; they neither destroy the engine nor reset
source/table/player state. A brief surface gap while otherwise active is caught
up by the Runner at the next presentation. Process restart uses A2 bootstrap.
A5 extends this lifecycle with the independent output-focus policy described below.

### Real frame upload

`pf_engine_frame` supplies top-down RGBA8. The host rejects null pixels, nonpositive
or greater-than-4096 dimensions, and strides smaller than `width*4` or larger
than 16384 bytes. Each row is copied using the supplied stride into a reusable
packed vector while the engine mutex is held. The engine pointer is never
written, retained, freed or passed to GLES. Vector capacity is reused; texture
storage changes only when dimensions change, with `glTexSubImage2D` otherwise.
The Go renderer retains its existing allocations; A3 does not redesign it.

GLES2 uses `GL_NEAREST` for both filters and black clearing outside a centred
aspect-preserving viewport. Table pixels are square. The original 640×240
startup/selector/options frontend has its existing desktop logical 640×480
pixel aspect. A no-data installation keeps the legal synthetic 320×609 shell.

### Transient portrait override

The optional additive ABI 1 function
`pf_engine_set_presentation(handle, full_table)` accepts only 0 or 1 and uses
the same invalid-handle/busy/non-reentry rules. Existing exports, signatures and
ABI version remain unchanged; older hosts default to 0 and need no changes.
A new host requires a library containing this optional export (packaged Android
symbol checks enforce it). The previous ABI had no transient presentation hook:
`pf_engine_frame` used saved model/session settings, and logical options keys
would change persistent preferences. This is why the additive hook is needed.

The setter stores only an engine presentation boolean. `Engine.Frame` calls
`Runtime.FramePresentation`; for that retrieval, it scopes a render-only
`physics.Game.PresentationFullTable` flag to the current session and restores
it with `defer`. `PresentationSettings()` returns a **copy** with ScrollOff only
for rendering. Physics raster/camera calculations, model Settings, session
Settings, update/audio tasks, settings storage and future session creation
continue to use the saved preference. Table raster, full-table composition,
attract artwork and frontend matrix text all use the render copy. No F5/options
navigation, preference key, settings write, source tick or audio advancement is
performed by the setter or orientation choice.

The renderer chooses portrait when surface height ≥ width, sets full_table=1
before retrieval, and presents the complete 320×33 matrix above the 320×576
field (320×609) in table modes. Frontend menus keep their own authoritative
raster. Landscape sets 0 and immediately restores saved HARD/MEDIUM/SOFT/OFF
presentation, including preferences changed through the normal options screen.
Orientation never writes or temporarily assigns saved settings.

### A6 touch-first layout and gestures

A6 physical acceptance remains **NOT TESTED** for this revision. The second
owner test found the keyboard-toolbar model intrusive and the relative mouse
swipe undercharged. The default UI now exposes semantic actions driven only by
`pf_engine_state()`. `nativeState` reads mode/table/flags under the existing native
session mutex; the main looper refreshes `SemanticUi` every 100 ms. This single,
framework-free mapper drives the sheets and gameplay overlay idempotently.

| Engine mode | Default touch UI |
| --- | --- |
| STARTUP | Tap to continue (Space underneath) |
| SELECTOR / SELECTOR_TEXT | Named four tables and Options (F1–F5) |
| ATTRACT | Players − / 1..8 / + and Play (selected F1–F8); entry defaults to 1 |
| PLAYING | L/R, tiny mobile menu, Pull ↓ only while flag bit 2 is active |
| OPTIONS | Up / Down / Select / Back |
| PAUSED | Resume / Exit table (Enter / Escape) |
| QUIT_QUESTION | Yes / No (Y / N) |
| INITIALS | Temporary QWERTY letters using existing DOS makes |

The [four-table DOS initials audit](dos-initials-audit.md) confirms that top-row
digits are rejected in all pinned PRGs, including original routine execution.
Initials remain A–Z and Space as `*`; no numeric row is added. Touch controls
(including selector cards and the menu icon) do not take hardware keyboard focus.
| GAME_END / ENTRY_WAIT / QUIT | No contextual controls |

The accepted semantic architecture is preserved. This A6 polish adds only host
geometry; A6 remains **NOT TESTED** pending another physical pass, A7 is not
started, and `main` is not merged.

Portrait selector uses two equal 64 dp table rows, two columns, followed by
Options in the left cell of row three. All five cells have identical dimensions,
two centered text lines, 8 dp gaps and 8 dp inner margins. The panel is
at most 360 dp wide and leaves at least 12 dp outside each safe edge. Landscape
uses three columns then Stones 'N Bones / Options, at most 600 dp wide, with the
same 64 dp cells. Neither selector scrolls. Ordinary contextual panels use 48 dp
rows and a 360 dp maximum width in both orientations.
The playing utility menu (Close / Pause, Music / Advanced keyboard) sits above
flipper guard areas. Only advanced/initials keyboards may use ScrollView. Hidden
panels are GONE and clear their touch exclusion rectangle. The menu has a 48 dp accessible touch target; its viewport-relative placement is described below. System bars, cutouts and gesture insets are respected.

The placement pass anchors SELECTOR / SELECTOR_TEXT, ATTRACT, OPTIONS, PAUSED
and QUIT_QUESTION panels at the bottom of the rendered framebuffer. With enough
bottom letterbox space, the panel starts immediately below the frame; otherwise
it straddles that edge or ends 8 dp above the safe bottom. Top, bottom, cutout and
gesture bounds clamp the result. Selector equal-cell geometry is unchanged;
Options and Players/Play retain compact two-row panels, leaving original content
and settings values visible above. Placement uses the renderer snapshot and
never computes another aspect ratio in Java.

The mobile menu's top is `viewport.top + viewport.height * 33 / source.height`
plus a 4 dp gap, at the framebuffer's right edge with a 4 dp margin, clamped to
safe interactive bounds. Portrait full table therefore uses 33/609; landscape
uses the current rendered source height. The touch target is 48 dp even though
the visible icon is small. The icon stays hidden until a valid viewport exists.

`HardwareKeyboard` enumerates InputManager devices initially and on device
add/change/remove, and rescans after Configuration changes. Only external,
non-virtual, alphabetic devices with SOURCE_KEYBOARD qualify; an IME, touchscreen,
built-in keyboard or gamepad alone does not switch modes. Keyboard mode hides
the whole semantic/menu layer and overlay, including initials, L/R, Pull and
Nudge. Hardware dispatch and existing DOS mappings remain authoritative.
Switching releases touch gestures and held hardware actions without calling
native lifecycle, source-clock, engine creation or audio APIs. Detach restores
the current native mode, including contextual QWERTY during INITIALS, without
restarting the game. All semantic buttons remain non-focusable.

Local deterministic host geometry, input/keyboard parity, native-session and
frontend tests pass; public and instrumentation APKs compile. Instrumentation
covers panel bounds in both orientations, matrix clearance, keyboard policy and
current-mode attach/detach including initials. Physical acceptance of this
placement remains pending another owner test; A7 is not started.

Normal utility menus contain no Data or Import command in any engine mode. The
central first-run/no-data Import DOS folder button remains the only import UX.
Menu → Advanced keyboard preserves the existing DOS keys. Panels remain Views
in the same Activity; focus/audio/cadence ownership is unchanged.

The renderer publishes its exact `a3::letterbox()` viewport, surface size and rendered source dimensions
under a separate host-only mutex. JNI `nativeViewport(int[8])` returns a coherent
snapshot, converting GLES bottom-origin y to Android top-origin y. The UI's
100 ms refresh maps surface pixels to overlay pixels; it does not recalculate
aspect ratio. Stale-orientation snapshots disable playfield gestures until the
new surface snapshot arrives. `InteractionGeometry` clips the frame to safe bounds.

L/R are explicit lower-corner rectangles. Height is min(30% of safe height,
160 dp). Portrait width is min(46% of safe width, 200 dp). Landscape width uses
actual side-letterbox space plus 48 dp, at least 160 dp, bounded by 40% of safe
width and 240 dp. Drawing and hit testing use those same rectangle objects:
faint translucent fill, rounded outline, and L/R at the exact rectangle centers.
No Nudge region is drawn. Transparent guards extend 24 dp above/inward of each
flipper; starts in guards stay dead throughout that gesture. Flipper starts are
immediate and retain pointer ownership, even after crossing or leaving a zone.

Nudge is restricted to the upper/central safe framebuffer, inset 8 dp, ending
above the guards with an 8 dp gap and before the right plunger corridor. UI/menu
rectangles are excluded. One short stationary neutral tap yields exactly one
Space make and 50 ms Tilt pulse; long presses and drags do not. L+R, either
flipper plus plunger, and deliberate Nudge during a pull coexist. Cancellation,
focus loss and stale pulse handling retain the existing safe clearing policy.

The rightmost min(25% of safe framebuffer width, 96 dp) above the guards is the
plunger corridor. Only the authoritative mouse-active flag enables its gesture;
corridor taps never nudge. The accepted downward, initially vertical-dominant
movement exceeding touch slop claims one owner; horizontal jitter then retains
ownership. Downward travel over 25% of safe height (minimum two slops) maps to
absolute rounded 0..32 charge, independent of MOVE count and speed. Release fires
once; cancel resets charge without firing. Accelerometer Nudge is deferred to a
separate sensor/physical-test pass.

Deterministic tests cover 1080×2400, 2400×1080, 1920×1080 and 1080×1920, safe
insets, renderer bounds, equal selector cells, capped/no-scroll panels, region
centers, guards, nudge counts, ownership and absolute plunger behavior. Android
instrumentation checks actual View sizes, no-scroll modes and hidden-panel touch
pass-through. Local A2/A3/A4/A5 and public source checks pass; hosted runtime
results are reported separately for the final commit.

`pf_engine_plunger_target(handle, target)` is an optional additive ABI 1 extension;
old consumers and `pf_engine_plunger_delta` keep their existing behavior. Targets
are clamped at the engine and source spring boundary, accepted only with active,
valid spring input, latched through Runner input, and applied by the existing
shared spring task. Release uses the existing following-source-task mouse fire
and table release physics. Fast swipes and 2 versus 20 MOVE callbacks produce
the same requested charge. No source deadlines or launch velocity formulas change.

Local validation of this redesign passed the A2 import transactions, A3/A4/A5
asset-free tests, the shared package matrix (with the repository's macOS platform
exclusions), and both-ABI engine/APK builds and 16 KB ELF checks. Privately staged
originals passed the unchanged four-table DOS mouse/keyboard regression and new
four-table half/full touch-target release checks. Hosted A0 still runs the real
16 KB ABI loader gate. None of these checks certify this revision on a physical
phone; A6 remains NOT TESTED and A7 is not started.

### Physical keyboard and Back

`Controls` is an Android-framework-independent translation policy shared by
hardware dispatch, menu controls and tests. Each physical key contributes
independently, and repeated/duplicate down events do not emit extra makes.
Tests compare every mapped key against the compiled macOS host make-code
implementation rather than a second expected scan-code table.
Action key-up removes only that contributor; Down release follows the engine's
existing spring release edge. Enter also calls `pf_engine_release` on its make,
matching the desktop native host.

| Android keys | Logical engine semantics |
| --- | --- |
| F1–F8 | DOS makes 59–66 |
| Enter / numpad Enter | release + make 28 |
| Escape / Android Back | make 1 |
| Down | spring held action + make 80 |
| Left Shift/Ctrl/Alt, Left arrow, Z | left flipper contributors |
| Right Shift/Ctrl/Alt, Right arrow, slash | right flipper contributors |
| Space | nudge held action + make 57 |
| P / M | makes 25 / 50 |
| Up, numpad multiply, A–Z | existing desktop logical makes, including initials and Y/N prompts |

Android system Back is registered through the Activity's Back dispatcher,
including predictive-Back integration; hardware Back uses the same translation.
It reaches the shared escape/back/quit state machine. The Activity is never
unconditionally finished by Back; a completed engine quit presents its quit
frame until normal process/lifecycle destruction.

### Validation and acceptance boundary

`tools/test_android_a3.sh` runs deterministic asset-free Java control tests,
C++ frame/stride/viewport tests, and the actual `engine_host.cpp` JNI/session
implementation against mock ABI functions. The session stress test races native
presentation, UI input and pause/focus transitions and asserts no overlapping
ABI calls, advancement while suspended, borrowed-pointer retention, destruction
races or stale-token revival. Java tests cover hit regions, simultaneous flippers,
multiple contributors, pointer-ID changes, cancellation, absolute touch targets, semantic mode/action mapping,
plunger callback-count independence, release, repeat suppression, make codes, Back and focus clearing.

Android instrumentation additionally dispatches real multi-pointer `MotionEvent`
objects through `ControlOverlay`, verifies cancellation/plunger dispatch, and
exercises no-data JNI callbacks. Hosted Android CI retains A2 import/recovery
checks, packaged ABI exports, payload scan, all ELF/ZIP 16 KB checks and A1
rotation/background/resume smoke. Smoke locks rotation through WindowManager rather than relying on a settings
write that emulator startup can race, issues hardware/menu/Back and touch input
in the no-data shell, and requires the same host to survive. Timeout diagnostics
include rotation policy and focused Activity/window state. Smoke waits for
`A3_INPUT_STATE resumed=1 focused=1` before rotation/inputs and after resume
or restart, so disabled early callbacks cannot count as input acceptance.
The disposable CI guest pre-acknowledges Android's first-run immersive education
via `immersive_mode_confirmations=confirmed` before its system-server restart.
Without that acknowledgement, `ImmersiveModeConfirmation` can own window focus
across resume; the game correctly suspends underneath it. This changes only CI
setup, never the APK or its lifecycle/input rules. Standalone dedicated-device
smoke requires dismissing that system tutorial first.

Asset-free Go regressions exercise the actual Engine → Runtime → four-table
render/composition path using invented zero indexed pixels, verify 320×609 and
restored landscape pixels without moving source time, and cover every saved
scrolling mode/resolution without mutating settings. The optional private-original engine regression compares
portrait/landscape rendering for all four tables and saved modes against a second
Runner, including source state, PCM and restored landscape pixels; public CI
explicitly skips commercial fixtures. Shared source/macOS regressions and A0
are required here because the render/ABI boundary changed. No commercial data
is added to any public test or artifact. Original-backed **Android** framebuffer,
touch playability and successful import acceptance still require A6; passing
mock, Go or no-data emulator tests is not proof of those Android outcomes.

Local validation during A3 passed the private-original Go presentation comparison
for all four tables × all four saved modes (80 advancing checkpoints each), plus
asset-free controls/frame/session tests, A2 transactions and the real desktop C
ABI error/extension contract. This is shared-engine evidence only, not
original-backed Android playability acceptance.

Local asset-free checks:

```sh
sh tools/test_android_import.sh
sh tools/test_android_a3.sh
sh tools/go.sh test -count=1 ./internal/physics ./internal/presentation ./internal/engine ./internal/frontend ./internal/source
```

## A4 Oboe output and bounded PCM

A4 integrates the pinned `com.google.oboe:oboe:1.10.0` Prefab package following
[Oboe’s integration guidance](https://github.com/google/oboe/blob/1.10.0/docs/GettingStarted.md).
The output builder requests 48,000 Hz, two channels, interleaved signed I16,
callback output, LowLatency performance and Exclusive sharing. If opening fails,
it retries Shared; a final Shared attempt permits Oboe format/sample-rate
conversion on the host side. The application-facing stream must still report
48 kHz stereo I16. The device buffer target is two native bursts (Oboe/device may
clamp it). No callback size or audio clock is used to advance gameplay.

### Ownership and bounded storage

- The GameActivity native glue/render thread calls `pf_engine_advance` using A3
  monotonic time. A3’s mutex continues to serialize every persistent-engine call.
- The synchronous `pf_pcm_sink` on that same engine/producer call path copies
  borrowed PCM immediately. It only writes host storage; no device write, wait,
  engine reentry, retained pointer or deadline adjustment occurs in the sink.
- Oboe’s realtime data callback is the sole live ring consumer. It performs only
  bounded frame copies, silence filling and lock-free atomic operations. There
  are no allocations, mutexes, log calls or engine ABI calls on this path.
- A dedicated native audio control thread owns open/start/close and deferred
  restart. JNI session/lifecycle calls publish atomic desired state while holding
  the engine mutex; they never wait for this thread or a stream operation. The
  controller polls at 10 ms, independently of rendering or the engine mutex.
  After stream close has joined data callbacks, control may discard stale ring
  frames. Callback objects and their storage are retained with Oboe shared
  ownership, including across delayed error notifications.

The SPSC ring has **4,096 stereo frames**, or **85.33 ms** at 48 kHz. Audio payload
is 16 KiB; per-frame generation tags and padding make the fixed slot storage
64 KiB, plus bounded atomic counters. This accommodates several 60/71 Hz source
batches and ordinary display-wake jitter while bounding queued latency below
macOS’s one-second ring capacity. It is a capacity ceiling, not a latency promise
or a requirement to fill before output. All indexing/capacity accounting uses
stereo frames; each frame copies four bytes, even for an unaligned borrowed input.
Release/acquire cursors protect slot reuse; only the producer writes the write
cursor and only the consumer writes the read cursor. Both public ABIs require
always-lock-free 64-bit atomics at compile time.

On underrun the callback fills missing frames with zeroes. It never repeats old
samples, stretches PCM, calls the engine or advances source time. On overflow it
accepts the available **prefix** of incoming frames and drops the **tail**,
preserving existing queued ordering and the contiguous oldest pending audio,
matching the macOS ring policy. The producer neither waits nor overwrites a slot
being consumed. Prolonged stalls can lose audio; they cannot change Runner cadence
or grow memory. No queue-depth feedback enters gameplay.

### Pause, close and ordinary stream errors

A4 introduced persistent-engine and resumed AND focused eligibility. A5 additionally
requires a live session and granted Android audio focus; background bootstrap never requests focus.
Suspension atomically disables consumption and changes the PCM generation while
preserving the engine and its source/game time. A callback racing that transition
silences its current block if it observes the generation change. The controller
closes output and drops invalidated storage; resume starts a fresh generation and
accepts newly produced PCM. Even a rapid pause/resume before control wakes cannot
replay old queued frames: the consumer skips generation-mismatched slots. Already
submitted device frames may finish for the device’s short buffer duration; there
is no continuing stale ring playback. Engine close, Activity replacement and
import engine replacement invalidate audio in the same way. Live SPSC cursors
are never reset from a producer/UI thread.

An ordinary Oboe error only increments counters and signals deferred restart.
The application takes responsibility for closing the failed stream; the control
thread closes it, then retries after 250 ms if the engine is still active. Failed
opens retry no faster than once per second; pause/close cancels that retry. All
construction/destruction/waits stay outside realtime callbacks and the engine
mutex. A5 adds focus, route events and a finite recovery budget below.

### Diagnostics and asset-free checks

Normal launches have no routine A4 logs. `PF_DIAGNOSTICS=1` or the existing Android
`PF_DIAGNOSTICS` boolean launch extra enables stream-open configuration and an aggregate
`A4_AUDIO` report at most every five seconds. Process-lifetime counters include
PCM frames produced (including overflow drops), valid PCM frames consumed,
callback count, underrun events/missing frames, overflow events/dropped frames,
stale invalidated frames, stream opens/errors/restarts, physical ring depth and
high-water depth. Depth may briefly include invalidated frames pending consumer
cleanup. No per-callback log is emitted; device/API failures do not re-enable the
previously disabled routine host diagnostics.

```sh
sh tools/test_android_a4.sh
sh tools/test_android_a3.sh
sh tools/test_android_import.sh
```

A4’s native tests exercise stereo ordering, unbounded producer stress against a
bounded consumer ring, wraparound, exact full/empty, silence, deterministic partial
overflow, generation invalidation and fresh resume. Allocation guards run on the
actual consumer/error callback paths. The actual A3 JNI/session test emits borrowed
mock PCM, overwrites its input immediately and verifies the copied samples,
suspend/resume/close and retained ABI serialization. The production Oboe
controller links against a device double **without any engine ABI implementation**
and tests Shared fallback, deferred disconnect recovery, bounded open retry and
quiet default diagnostics. Hosted CI additionally runs the ring under TSAN.

Android CI builds both real Oboe/engine ABIs, retains ELF/ZIP 16 KiB alignment,
payload scans and all A1–A3 checks. Packaged JNI instrumentation uses a separate
production audio controller with invented quiet stereo PCM and requires stream
opens, data callbacks, ring consumption, pause/resume and safe close on the 16 KiB
emulator. The dummy host audio backend requires no audible machine output. No
commercial data is used. A0 is unchanged and runs only for its existing boundary
paths; shared engine/audio code is unchanged, so new desktop/macOS regressions
are not required for A4.

This is host/output evidence, **not physical-device audio acceptance**. Physical
latency, original-backed Android audio/playability and normal compiled ART remain
unverified. A5 adds audio focus, interruptions, route changes and
Bluetooth/headset recovery below; A6 owns
original-backed parity and physical-device acceptance; A7 owns signed release
packaging.

## A5 audio focus, interruption and route recovery

The UI-looper `AudioPolicy` reducer requests focus only on an ineligible → eligible
transition. Eligibility means session open, persistent engine available, Activity
resumed and window focused. The platform request result is authoritative: failure
leaves gameplay/render/input valid but output silent, without repeated requests.
A later foreground/playback transition can request again. Import stop and Activity
teardown abandon focus; a background bootstrap cannot acquire it. Native output
independently requires all these prerequisites AND the published focus grant.

One `AudioFocusRequest` instance is reused for request/abandon: `AUDIOFOCUS_GAIN`,
`USAGE_GAME`, `CONTENT_TYPE_MUSIC`, `setWillPauseWhenDucked(true)` and
`setAcceptsDelayedFocusGain(false)`, with a UI Handler listener. Oboe uses matching
`Usage::Game` / `ContentType::Music`. For targetSdk 35+ (this app targets 36),
[Android requires the top app or a foreground service](https://developer.android.com/media/optimize/audio-focus).
This host requires resumed/window-focused playback and adds no foreground service.
Even eligible requests may fail; the host stays silent in that case.

| Event | Output policy | Focus request policy |
| --- | --- | --- |
| Eligible foreground transition | Enable only after platform grant | Request once |
| Transient loss, including call/notification style interruptions | Disable and invalidate PCM; retain engine | Await gain while eligible |
| CAN_DUCK | Same temporary silence; no sample mixer | Await gain while eligible |
| Gain | Enable a fresh PCM generation only while eligible and requested | No duplicate request |
| Permanent loss | Disable and invalidate; retain engine | Abandon; ignore late gain until later eligible transition |
| Pause, window loss, engine/session close | Disable and invalidate | Abandon |
| Route/device/noisy event | Invalidate and defer reopen if still eligible | Retain current grant |

Focus callbacks never suspend/reset/recreate the engine, generate game pause keys,
or change settings. A3 lifecycle/window state alone still governs engine suspension
and source progression. Gain-before-resume, resume-before-gain, loss-while-paused,
duplicate notifications and close-before-callback are harmless. Closed UI reducers
ignore late listeners; native JNI guards every publication with the session token.
Dedicated handler messages queued before abandon are removed. Ordinary rotation
uses the existing manifest configuration handling, retains the Activity, engine,
request and grant, and causes no duplicate focus request or new game.

While granted playback is relevant, `AudioDeviceCallback` tracks output-device
addition/removal, including built-in, wired, USB and Bluetooth A2DP devices, and a
receiver tracks `ACTION_AUDIO_BECOMING_NOISY`. Input-only devices are ignored.
The system selects the route; there is no picker or pinned device ID. Notifications
publish generation changes only. Becoming-noisy is a route transition: it never
pauses/resets gameplay, emits input keys or mutates persistent settings. Removed
headphones may lead to speaker playback if the system selects it and focus remains
held. No Bluetooth latency compensation, resampling of progression or time stretching
is introduced. Oboe disconnect recovery also covers a changed route without a
Java notification.

### Ownership and recovery

* Java/UI looper owns lifecycle, AudioManager request/abandon, focus listener,
  receiver and device notifications. It publishes bounded token-checked native
  state; it never opens, closes or waits for an Oboe stream.
* Native engine/session mutex serializes ABI calls, lifecycle/input, PCM production
  and session-associated output publications. Focus affects audio eligibility only.
* The shared `PcmBuffer` atomic epoch is the sole generation authority. UI/native
  transitions, lightweight Oboe errors and control startup use atomic updates so
  concurrent invalidations cannot lose an epoch. Live SPSC cursors are never reset.
* The Oboe data callback only consumes PCM, fills silence and updates atomics. It
  never calls Go/Java, locks an engine/host mutex or allocates. Error callbacks
  invalidate and signal control, ignore retired stream identities, and never close.
* The existing audio control thread alone opens/closes/restarts streams. Only after
  close joins the callback may control discard ring entries. No close/join happens
  while holding the engine/session mutex, so these domains have no circular wait.

Route bursts coalesce over a 50 ms controller settling interval. Every reopened
stream starts with a fresh epoch. The PCM sink also requires stream readiness:
failed opens cannot continuously refill a silent ring, and unavailable-device PCM
cannot suddenly play later. Each activation/route recovery has at most three open
attempts (each retains Exclusive → Shared → conversion fallback). Failed opens
back off one second; disconnect recovery backs off 250 ms. Repeated immediate
stream disconnects share that budget; five seconds of stable playback renews it.
After exhaustion, only a meaningful lifecycle/focus/route transition renews recovery.
A separate bounded atomic request revision distinguishes those external transitions
from internal PCM invalidations, so reordered error/epoch signals cannot renew retries.
The accepted 4096-frame capacity, tail-drop overflow, 48 kHz stereo I16 callback,
LowLatency, Exclusive preference and two-burst target remain unchanged.

### Diagnostics, tests and scope

Normal public launches remain quiet. Opt-in `PF_DIAGNOSTICS=1` (or the existing
boolean intent extra) adds `A5_FOCUS_REQUEST`, focus class and cumulative loss/gain
counters, route/device notification type/ID, opened stream device ID, noisy event, native route epoch and deferred stream
transition/reopen diagnostics. A4 aggregate callback/ring counters remain throttled
to five seconds. No callback-by-callback logging is added.

`sh tools/test_android_a5.sh` runs the asset-free UI reducer and retains A3/A4
native tests. Coverage includes rejected requests, transient/duck/permanent loss,
fresh gain, lifecycle ordering, rotation idempotence, closed/old listeners and JNI
tokens, route invalidation, route-burst coalescing, deferred disconnect, finite
open-failure budget, no failed-device accumulation, unchanged engine timing and
concurrent render/input/lifecycle/focus/route publication. Hosted CI runs both the
native session and Oboe-controller double under ThreadSanitizer, retaining the A4
SPSC tests and callback allocation guard.

Debug-only `AudioTestActivity` shares the production Android focus/device adapter
but uses an isolated synthetic PCM controller with no engine/assets. Instrumentation
requires a real foreground AudioManager grant, injects transient/permanent/gain
notifications into that adapter, checks silence/fresh callback consumption,
background/resume, deferred route reopen and injected delivery to the registered
becoming-noisy receiver (the system broadcast is protected against application senders). Its Java Activity/manifest and native
JNI hooks are excluded from release builds; no production intent extra enables them.
All A1–A4 packaged ABI, controls, framebuffer, rotation, quiet-launch, background and
16 KiB checks remain. Host-only changes do not trigger A0 or desktop/macOS matrices.

**Hosted A5 tests do not certify physical Bluetooth behavior, actual wired/USB
route behavior, device-specific latency, normal compiled ART performance, or
original-backed gameplay/import parity.** The hosted ps16k emulator still uses
interpreted ART and a dummy host audio backend. These remain A6 physical/original
acceptance, followed by A7 signing/packaging/release work, both requiring owner
approval. A5 does not start either milestone or merge to main.

## Fixed platform decisions

- `minSdk 27`, `compileSdk 36`, `targetSdk 36`.
- Public 64-bit ABIs: `arm64-v8a` and `x86_64`; no 32-bit ABI in the initial port.
- GameActivity/C++ host, EGL/OpenGL ES renderer, Oboe output, Android SAF import and app-private storage.
- Engine PCM remains 48 kHz stereo signed 16-bit. Device conversion and buffering belong to the host.
- Engine/source timing remains owned by the Go Runner. Display callbacks may wake/present but never redefine the 60/71 Hz source cadence.
- Portrait forces an in-memory `SCROLLING OFF` presentation override (full table, 320x609) without overwriting the saved scrolling preference. Landscape uses the saved HARD/MEDIUM/SOFT/OFF preference. Returning to landscape restores that preference.
- Orientation changes must not restart gameplay, reset source time, alter audio progression or change persistent settings merely because the presentation override changed.
- User-supplied commercial DOS files are never packaged in public APK/AAB artifacts.

## Phase A0: native toolchain feasibility

Before building the Android UI, prove that the existing Go `c-shared` engine can be used unchanged on Android:

1. Build `libpfengine.so` with Go 1.27.1 and Android NDK r28+ for `arm64-v8a` and `x86_64`, API 27.
2. Verify every ELF LOAD segment has at least 16 KB alignment.
3. Load the x86_64 library on a real 16 KB Android emulator and call `pf_engine_abi_version`.
4. Call `pf_engine_create` with deliberately missing data and require a clean zero-handle failure plus a bounded error string. A0 contains no commercial input.
5. Do not change ABI 1 to make A0 pass. A Go/NDK/linker compatibility problem is handled in the build boundary first.

Local build:

```sh
export ANDROID_NDK_HOME=/path/to/android-ndk
sh ./tools/build_android_engine.sh all
```

Device/emulator smoke after `adb` sees the target:

```sh
sh ./tools/run_android_abi_smoke.sh x86_64 16384
```

The second argument is the required device page size. Omit it for a normal 4 KB development device.

## Remaining phases

A3 implements cadence, framebuffer upload, touch/keyboard translation and transient
portrait presentation on the A2 engine. A4 implements Oboe plus a bounded host PCM
ring/device output. A5 implements lifecycle, interruption and
audio-focus handling. A6 remains original-backed Android parity, practical touch
charging/layout acceptance, normal compiled ART, hardware keyboards and physical
devices. A7 remains signed APK/AAB packaging and release payload scanning.
A6/A7 and a main merge require owner approval. No physical-device audio acceptance is claimed.
