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

## Hosted result for implementation `de3d22e` (2026-10-05)

[Android host CI](https://github.com/voobrazimoe/pinball-fantasies-native/actions/runs/37281345240)
passed, including ThreadSanitizer, packaged instrumentation, quiet default launch,
Choreographer diagnostic window, rotation, background/resume and process restart.
A0, asset-free source and macOS host workflows also passed for that implementation.
The API 35 ps16k emulator selected `callback=64`, retaining `swap_interval=1`.

Representative no-data synthetic shell windows from that run:

| Wall seconds | Callbacks / swaps | Callback Hz / submitted FPS | Draw-start interval min / avg / max ms | Swap CPU avg / max ms |
| --- | --- | --- | --- | --- |
| 3.065 | 30 / 30 | 9.79 / 9.79 | 32.415 / 103.625 / 404.576 | 82.291 / 334.102 |
| 3.020 | 50 / 50 | 16.56 / 16.56 | 34.955 / 61.120 / 112.387 | 41.731 / 97.723 |
| 3.055 | 48 / 48 | 15.71 / 15.71 | 25.408 / 63.702 / 124.904 | 47.723 / 111.161 |

Observed Choreographer timestamp intervals in these windows had a 33.333 ms minimum
and 60.544–103.448 ms averages; missed callbacks make this unsuitable for claiming
the emulator's physical refresh rate. A later fresh-process startup window recorded
6.21 FPS with 128.383 ms average swap CPU time. This slow hosted compositor cannot
establish physical smoothness. Every attempted draw in these windows swapped
successfully; swap/backpressure was the largest measured CPU stage. This does not
isolate an additional factor-of-two cadence reduction caused by swap interval 1,
so the policy remains 1 pending phone measurements.

All these windows had `tick_samples=0` and `upload=0.000/0.000`: the synthetic
texture is initialized at attachment and no real engine framebuffer is uploaded.
Consequently neither original-backed source ticks/presentation nor gameplay upload
cost was measured. Texture upload materiality remains unknown and it is unchanged.
No-data audio counters were zero (no PCM/streams); they are not evidence of physical
gameplay underrun freedom. Host audio/focus/route regressions and packaged synthetic
audio instrumentation passed; real-phone audio and flipper latency remain retests.
