# Android host

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
- A2 implementation adds SAF import, private Data/State, packaged ABI 1 validation
  and persistent engine creation. Hosted acceptance is pending; no original-backed
  Android acceptance is claimed. A3–A7 remain outstanding.

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
storage output and no commercial fallback in the APK or tests.

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
handle. A2 does not advance it. Relaunch bootstraps adopted Data automatically;
missing Data stays in the shell, and malformed/incomplete Data reports failure.
An engine/bootstrap failure after commit leaves validated Data installed and
reports the error; importing again does not require deleting it.

Log events distinguish `A2_SHELL_NO_DATA`, `A2_IMPORT_REQUESTED`,
`A2_IMPORT_CANCELLED`, `A2_IMPORT_REJECTED`, `A2_CANDIDATE_VALIDATED`,
`A2_DATA_ADOPTED`, `A2_ENGINE_BOOTSTRAPPED`, and `A2_BOOTSTRAP_REJECTED`.
The A1 synthetic renderer and its timing are unchanged and never clock the engine.
There are no gameplay controls, framebuffer upload, or audio device operations.

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

A2 adds SAF staging, app-private Data/State and real engine bootstrap. A3 adds cadence, framebuffer upload and touch/keyboard translation, including the temporary portrait scrolling override. A4 adds Oboe and the bounded host PCM ring. A5 adds engine lifecycle/audio-focus handling. A6 runs original-backed parity and physical-device acceptance. A7 adds signed APK/AAB packaging and release payload scanning.
