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
- The host stops rendering while the Activity is paused. The runtime smoke
  requires new log evidence for portrait → landscape → portrait, the same
  process/native host across rotation and background/resume, a presented frame
  after resume, and a fresh EGL frame after process restart.
- A2–A7 are outstanding. There is no SAF import, packaged engine integration,
  real game framebuffer, touch/keyboard controls, Oboe audio, engine lifecycle
  or audio focus handling, original-backed Android parity, physical-device
  acceptance, or signed release APK/AAB yet.

Run the shell build and smoke with an installed SDK/NDK/JDK, Gradle 9.6.0 and
an already booted emulator:

```sh
gradle -p hosts/android :app:assembleDebug
sh tools/run_android_host_smoke.sh
python3 tools/test_check_android_elf.py
```

The last command tests the alignment checker without an Android SDK. The smoke
changes emulator orientation settings and restarts the test app; use a dedicated
test device. CI preserves a successfully built debug APK even if runtime checks
fail. It contains synthetic pixels only.

The next implementation milestone is A2: package the unchanged ABI 1 engine,
copy only the 11 required PRG/MOD files plus optional CFG through SAF to a private
staging directory, validate via the shared loader with disposable state, then
adopt the validated directory without damaging the previous installation.
Keep commercial inputs under `noBackupFilesDir` and player state under
`filesDir`; never pass a `content://` URI to Go. Import rejection/cancellation
must leave the existing Data and State untouched.

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

## Next phases after A0

A1 creates the empty GameActivity + EGL/GLES host with synthetic pixels only. A2 adds SAF staging, app-private Data/State and real engine bootstrap. A3 adds cadence, framebuffer upload and touch/keyboard translation. A4 adds Oboe and the bounded host PCM ring. A5 adds lifecycle/audio-focus handling. A6 runs original-backed parity and physical-device acceptance. A7 adds signed APK/AAB packaging and release payload scanning.
