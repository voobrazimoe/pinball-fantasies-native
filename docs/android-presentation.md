# A6 Android presentation pacing investigation

Owner physical testing at `1908b44` reported functional gameplay, controls,
orientation and keyboard attach/detach, but visibly jerky/slower presentation
compared with DOS. A6 remains pending physical smoothness retest; A7 is not started.

Before: `ALooper_pollOnce(active ? 16 : -1)` → advance/copy latest engine frame →
RGBA texture upload → GLES2 draw → `eglSwapBuffers`, with swap interval 1. The poll
sleep plus swap/compositor backpressure is a plausible double-pacing cause.
There is no measured physical before/after FPS in this investigation yet.

After: the native glue thread blocks in `ALooper_pollOnce(-1)` for lifecycle,
input and Choreographer dispatch. One Choreographer callback presents at most one
latest frame, then posts the next callback while resumed, focused and attached.
The actual display callback cadence is used; no 60 Hz assumption or interpolation.
Java/native input submission remains immediate and mutex-serialized.

The callback timestamp is used only for diagnostics. `androidEngineFrame` still
passes current `CLOCK_MONOTONIC` real time to the unchanged Go Runner, advancing
all due source tasks at 60 Hz frontend / 71 Hz table cadence. At 60 Hz some source
frames are skipped; at 120 Hz some presentations repeat. Display callback count
never determines game/source ticks or tracker time. This is the same conceptual
separation of source clock and compositor presentation as macOS; no CALayer code
is ported. Oboe/audio focus/route policy, ring and device callback are unchanged.

API 27–28 use `AChoreographer_postFrameCallback` (64-bit `long`, enforced with
`static_assert`). API 29+ resolve `AChoreographer_postFrameCallback64` with `dlsym`
at runtime; a missing symbol falls back to the legacy path. No API 29 symbol is
required by the loader. minSdk 27 and arm64-v8a/x86_64 remain unchanged.
See the [NDK Choreographer contract](https://developer.android.com/ndk/reference/group/choreographer).

Lifecycle commands stop eligibility on pause, focus loss and TERM_WINDOW. A
pending callback retains its original epoch and cannot draw into a replacement
window even if resume occurs before delivery. A new valid callback re-arms once
the stale one is consumed. Callbacks and EGL lifecycle all run on one Looper;
shutdown disables production and drains the sole pending callback before its
Renderer data is destroyed. Engine retention/data/state policy is unchanged.

## Opt-in physical measurement

Enable the existing boolean Activity extra (or `PF_DIAGNOSTICS=1` environment):

```sh
adb shell am start -W -n io.github.voobrazimoe.pinballfantasies/.PinballActivity --ez PF_DIAGNOSTICS true
adb logcat -v threadtime -s PinballFantasies:I
```

Only this opt-in path collects fixed-storage three-second windows:

- `A6_PACING`: API, callback path, swap interval.
- `A6_PRESENT`: callback/draw attempts, successful swaps, wall seconds, callback
  rate/present FPS, source ticks, ticks/presentation and 0/1/2+ distribution.
  Unknown engine samples (e.g. no-data synthetic shell) are excluded from ticks.
- `A6_PRESENT_INTERVAL_MS`: min/average/max draw-start intervals, successful swap
  completion intervals and Choreographer timestamp intervals. Observed vsync
  intervals may be multiples of the true display period if callbacks are missed;
  do not mistake their average for a guaranteed physical refresh rate.
- `A6_PRESENT_COST_MS`: average/max engine advance+copy, upload, draw submission and
  swap CPU time. No `glFinish` or GPU fence is introduced, so deferred GPU work can
  appear as later backpressure rather than as upload time.
- Existing five-second A4 audio counters: compare underrun/overflow counts, stream
  starts and A5 route/focus epochs in the same gameplay/lifecycle run.

Capture steady gameplay for at least 15 seconds at the phone's available display
modes, then pause/resume, rotate, recreate the surface and attach/detach keyboard.
Compare callback rate, swap rate, swap duration and tick distribution alongside
observed smoothness. Successful swaps count submitted buffers, not proof that
SurfaceFlinger displayed every buffer. Do not infer actual scanout from FPS alone.

EGL swap interval stays 1 for the first measurement. If swap still halves cadence,
investigate Android EGL queue/pacing using those measurements. No interval-0
change, frame-rate mode hint, PBO/zero-copy/Vulkan or upload optimization is made.
Texture upload's physical materiality remains unmeasured.

## Validation

Host scheduling tests cover blocking active/inactive polls, at-most-one present,
re-arm eligibility, pause/focus/window retirement, stale callbacks after restart,
shutdown drain and diagnostic interval/tick aggregation. Production wiring checks
verify no timer draw in the Looper, monotonic real-time engine advancement and
immediate input. Native session mocks verify tick delta sampling under the engine
mutex and unknown ticks when inactive. Existing A3/A4/A5 tests and A6 geometry,
semantic/input checks remain in the same suite. Hosted smoke now requires an opt-in
presentation diagnostic window and verifies quiet production launch.

Local A3–A6 host tests, import tests, 23 Android Python tool tests, engine/frontend
Go regressions and actual NDK API 27 compilation for both ABIs passed. The local
ADB inventory is empty and no emulator is installed; there are no local runtime
pacing or physical audio measurements. Hosted emulator checks can establish
callback/lifecycle integration but cannot certify phone smoothness, compiled-ART
performance, listening, hardware routes or input latency.
