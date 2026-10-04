# Android host

Android support is being added as a native host around the existing Go engine and stable C ABI 1. The Android host must not duplicate table rules, presentation programs, tracker progression, settings/high-score encodings, source cadence or gameplay semantics.

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
./tools/build_android_engine.sh all
```

Device/emulator smoke after `adb` sees the target:

```sh
./tools/run_android_abi_smoke.sh x86_64 16384
```

The second argument is the required device page size. Omit it for a normal 4 KB development device.

## Next phases after A0

A1 creates the empty GameActivity + EGL/GLES host with synthetic pixels only. A2 adds SAF staging, app-private Data/State and real engine bootstrap. A3 adds cadence, framebuffer upload and touch/keyboard translation. A4 adds Oboe and the bounded host PCM ring. A5 adds lifecycle/audio-focus handling. A6 runs original-backed parity and physical-device acceptance. A7 adds signed APK/AAB packaging and release payload scanning.
