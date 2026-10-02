# PF12 fullscreen pacing audit

Date: 2026-10-01. Linux X11 desktop: 5120×2880. SDL output used real PulseAudio; Windows x86-64 output used waveOut under Wine 10. This is Wine evidence, not Windows 10/11 hardware acceptance.

## Method

`tools/audit_pf12_pacing.py --label after` runs four sequential, bounded 12-second fullscreen selector runs: Linux and Wine, each with and without host presentation. `PF12_NO_PRESENT=1` removes the host blit only; shared source updates, logical frame generation, mixer and PCM submission remain active. JSONL records update/frame/present durations, queue depth after each phase, empty-queue observations and output resets. Queue metrics are diagnostic only; they never drive the simulation clock.

The selector uses its existing 60 Hz clock, so the expected count is 720 source syncs. Tables retain 71 Hz. Results are in [pacing](pf12-validation/pacing/after-summary.json); before measurements are in [before-summary.json](pf12-validation/pacing/before-summary.json). These are bounded observations, not proof against arbitrarily long OS/device stalls.

## Before

| Backend | Host present | Updates | Presents | Present median / p95 / max ms | Empty queue observations | Resets |
|---|---|---:|---:|---|---:|---:|
| Linux SDL | yes | 649 | 650 | 17.30 / 18.40 / 99.16 | 266 | 0 |
| Linux SDL | no | 720 | 0 | — | 0 | 0 |
| Win32/Wine | yes | 331 | 332 | 29.93 / 35.85 / 58.60 | 323 | 0 |
| Win32/Wine | no | 720 | 0 | — | 0 | 0 |

One source update per host frame underproduced PCM whenever presentation took longer than the source interval. Removing present restored both source cadence and audio supply. This supports a common scheduling cause for crackle; it does not attribute the problem to mixer content or a Windows-specific simulation rate.

## Change

The shared loop advances all overdue source deadlines and submits each tick's PCM before producing the next visible host frame. Deadlines advance by the existing source interval, not by the time a blit finishes. No game ticks are discarded, and queue depth/device completion does not determine game time. Intermediate host frames may be omitted. A Win32 modal resize timer invokes the same deadline-driven work on the UI thread; no concurrent simulation or mixer was introduced.

The loop remains single-threaded. A sufficiently long blocking host call can still exhaust the bounded output queue. The audit establishes that the measured fullscreen costs no longer reduce PCM production to the host frame rate; it does not promise independent refill during an indefinitely blocked Win32/SDL call.

## After

| Backend | Host present | Updates | Presents | Present median / p95 / max ms | Steady queue min / median / max bytes | Empty / resets |
|---|---|---:|---:|---|---|---|
| Linux SDL | yes | 720 | 654 | 17.34 / 17.96 / 56.03 | 8960 / 13312 / 17024 | 0 / 0 |
| Linux SDL | no | 720 | 0 | — | 9472 / 11520 / 13696 | 0 / 0 |
| Win32/Wine | yes | 720 | 314 | 32.44 / 34.67 / 37.56 | 22400 / 25600 / 32000 | 0 / 0 |
| Win32/Wine | no | 720 | 0 | — | 12800 / 12800 / 12800 | 0 / 0 |

At 192000 PCM bytes/s those median queues are approximately 69 ms (Linux) and 133 ms (Wine); these are output-sink depths, not gameplay timing. Wine's cheap phase measurements are sometimes quantized to zero. Present cost is still appreciable at this large desktop resolution, but 720 source updates execute in both runs.

Separate real monitor captures exercised fullscreen music through actual SDL/PulseAudio and waveOut/Wine output. Both logs report zero empty observations/resets, one producer start and no lifecycle clears. Captures are 48000 Hz, signed 16-bit, stereo. Linux had 575915 distinct-channel frames; Wine had 567468. See [capture log](pf12-validation/pacing/fullscreen-output-capture.log) and adjacent WAV/metrics files. Earlier isolated 440 Hz left / 660 Hz right tone capture also verified distinct routing. Queue observations and recordings are complementary evidence; no comparison of device latency to gameplay clocks was made.

## GDI repaint / flicker

Previously, clearing the visible DC and then stretching the frame exposed a black intermediate surface. The backend now clears letterboxes and performs top-down BGRX StretchDIBits into a cached compatible memory bitmap, then transfers the complete client image with one BitBlt. The surface is resized only when actual client dimensions change.

WM_PAINT uses BeginPaint/EndPaint and the retained complete surface; it neither updates gameplay nor restarts audio. WM_ERASEBKGND returns 1 because the complete surface owns the background. WM_SIZE invalidates without background erase and retains the logical framebuffer. The native smoke forces an erase request and UpdateWindow, checks repaint completion and verifies the erase handler. The fullscreen capture recorded 2 paints, 3 erase requests and 316 complete blits. The no-present control suppresses WM_PAINT blits too.

This removes the identified partial-frame/background path. Physical fullscreen flicker acceptance on Windows remains pending. The additional Wine fullscreen journey now passes all mode aspect checks, restore geometry, repeat suppression, Enter consumption and persistent HWND. The harness crops the actual Win32 client obtained with GetClientRect/ClientToScreen, excluding Wine/GNOME decorations; its earlier restored-window aspect failure measured a titlebar/shadow. Wine may recreate its X11 bridge while the application's HWND stays unchanged. The longer Linux UI journey with real PulseAudio passes with 7 empty queue observations and zero resets. The longer Wine journey reports 5 empty queue observations and zero resets across scene/pause/producer transitions; it is not the continuous-output timing audit, and those aggregate observations are not classified as audible underruns.

## Deterministic regression

`TestHostFrameCadenceLeavesTicksAndPCMIdentical` compares identical shared Party sessions for 1200 source ticks with logical frames consumed every tick versus every eighth tick. Every tick's PCM is byte-identical; final logical hashes match, score is 2300000 and ball is 2. It passed on Linux and in the Windows-compiled test under Wine. The source checkers and existing oracle remain unchanged.


## Alt+Enter transition blocker (PF12 closure)

Real Windows acceptance reported audible stutter specifically during windowed ↔ borderless switching, while steady continuous fullscreen output was clean. This is a separate blocker. No real-Windows transition trace or post-fix listening result is available from this Linux host; the hardware cause and acceptance must not be inferred from Wine timings.

### Identified source ordering and host work

The old shared `step` executed Alt+Enter in input collection, before its due source update and Queue. It now consumes the same make edge/repeats/Enter but defers each host mutation until `advance` has submitted every source deadline already due. It rechecks after transitions and before Present. Win32 can service deadlines that elapsed between GDI clear/stretch/blit calls without recursive presentation or future ticks.

The old host permitted synchronous paint during style/geometry changes. WM_PAINT could resize and render the client-sized surface while SetWindowPos was still on the shared thread. The style bracket now validates intermediate paints without rendering, WM_SIZE only marks dirty, background erase stays suppressed, and SWP_NOREDRAW/NOACTIVATE keep FRAMECHANGED semantics. One asynchronous invalidation follows. SetWindowPlacement remains for correct normal/maximized restoration; redundant restore ShowWindow is removed. The complete retained surface and logical pixels stay intact until the normal next Present/WM_PAINT. Letterboxes are cleared only when surface/logical dimensions change. A full GetDC Present validates its covered client region, consuming a redundant subsequent paint.

Wine exposed additional settling work outside the direct transaction: after entering fullscreen, PeekMessageW blocked 48.75–51.61 ms while its pre-call queue was 9600 bytes (at most 50 ms, including the playing header). Two selector empty observations resulted. Those measurements are preserved in `pf12-validation/transitions/wine-5120-before-pump-refill/`; preliminary first-render evidence is in the `*-before-refill` traces. The final path services already-due source deadlines before each message call during the 500 ms settling interval following a fullscreen mutation. This services elapsed real time only; it changes neither steady input pumping outside that interval nor the audio queue policy. The same calls still block; a single call longer than the actual available PCM cushion can still drain the device.

This is a measured Wine starvation mechanism and a fix for confirmed scheduling/render-work deficiencies. Whether real Windows was blocked in SetWindowPos, placement, synchronous paint, message settling or another stage remains pending its trace. No tracker/mixer/sample/PCM format change, duplicated PCM, skipped source ticks, source ticks ahead of time, waveOutReset/restart, pause/resume or extra queue latency was introduced. The queue remains bounded at 250 ms/32 headers with its existing 2048-frame preroll.

### Focused diagnostic records

Set `PF12_TRANSITION_LOG` to a writable JSONL path before launching the native Windows application. Transaction records include pre/post queue bytes/ms, independent empty snapshots, cumulative empty/overflow/all-waveOut-reset/start/clear counters, duration, HWND and geometry. Per-call records cover style, monitor/placement, SetWindowPos, ShowWindow where necessary and invalidation. Synchronous WM_SIZE, WM_PAINT, WM_WINDOWPOSCHANGING/CHANGED and erase handler counts/times are inclusive (nested handler times must not be added to caller times). Surface recreation and draws during the next 500 ms are separate records; the first draw is marked. GetDC/ReleaseDC, BeginPaint/EndPaint, validation and message-pump calls are also traced during that interval. Queue empty events include time and the last draw call.

waveOut queue depth counts incomplete headers, including all bytes of the playing header. The milliseconds are an upper bound, not an exact remaining cushion. Measurements do not influence deadlines; transaction snapshots precede diagnostic file I/O. File I/O can perturb an instrumented run and should be compared with an uninstrumented listening run. No compulsory diagnostics or log-volume changes affect normal release playback.

### Bounded music transition smoke

`TestNativeFullscreenMusicTransitions` runs the normal runtime for 10 s with selector music, then 10 s with Party music. Alt+Enter make/repeat/break goes through the Win32 key translator and shared key/deadline path at 2/4/6/8 s:

`windowed → fullscreen → windowed → fullscreen → windowed`.

It checks persistent HWND, exact restored geometry, Enter consumption, zero empty observations and unchanged device lifecycle. The test positions its window at (100,100) to avoid Wine/GNOME edge snapping. It does not synthesize replacement music, queue duplicate output or fill with future ticks. Opt in on Windows with `PF12_TRANSITION_SMOKE=1`, `PF12_DATA_DIR=<original data directory>` and optionally `PF12_TRANSITION_OUTPUT_DIR=<existing writable output directory>`, then run `go test -count=1 -v ./internal/platform -run TestNativeFullscreenMusicTransitions` with CGO_ENABLED=0. Alternatively run the compiled `bin/platform-windows.test.exe` with the corresponding `-test.run` selector. Manual real-Windows release listening should cover the same sequence with and without instrumentation.

Final automated results (Wine only; `summary.json` and raw JSONL/native logs are under `pf12-validation/transitions/`):

| Desktop | Music | Transition | Duration ms | Queue bytes pre / post | Empty / resets / clears | Starts | HWND |
|---|---|---|---:|---|---|---:|---|
| 1920 Wine | selector | 1 enter | 0.517 | 12800 / 12800 | 0 / 0 / 0 | 1 | `0x30052` |
| 1920 Wine | selector | 2 leave | 3.162 | 12800 / 12800 | 0 / 0 / 0 | 1 | `0x30052` |
| 1920 Wine | selector | 3 enter | 0.437 | 12800 / 12800 | 0 / 0 / 0 | 1 | `0x30052` |
| 1920 Wine | selector | 4 leave | 3.084 | 12800 / 9600 | 0 / 0 / 0 | 1 | `0x30052` |
| 1920 Wine | table | 1 enter | 0.507 | 13524 / 13524 | 0 / 0 / 0 | 1 | `0x40052` |
| 1920 Wine | table | 2 leave | 2.740 | 13524 / 10820 | 0 / 0 / 0 | 1 | `0x40052` |
| 1920 Wine | table | 3 enter | 0.487 | 13524 / 10820 | 0 / 0 / 0 | 1 | `0x40052` |
| 1920 Wine | table | 4 leave | 3.030 | 13524 / 10820 | 0 / 0 / 0 | 1 | `0x40052` |
| 5120 Wine | selector | 1 enter | 0.422 | 12800 / 12800 | 0 / 0 / 0 | 1 | `0x30052` |
| 5120 Wine | selector | 2 leave | 14.806 | 12800 / 9600 | 0 / 0 / 0 | 1 | `0x30052` |
| 5120 Wine | selector | 3 enter | 0.443 | 12800 / 12800 | 0 / 0 / 0 | 1 | `0x30052` |
| 5120 Wine | selector | 4 leave | 14.374 | 12800 / 9600 | 0 / 0 / 0 | 1 | `0x30052` |
| 5120 Wine | table | 1 enter | 1.149 | 13524 / 13524 | 0 / 0 / 0 | 1 | `0x40052` |
| 5120 Wine | table | 2 leave | 17.528 | 10820 / 8116 | 0 / 0 / 0 | 1 | `0x40052` |
| 5120 Wine | table | 3 enter | 0.462 | 13524 / 13524 | 0 / 0 / 0 | 1 | `0x40052` |
| 5120 Wine | table | 4 leave | 13.367 | 13524 / 10820 | 0 / 0 / 0 | 1 | `0x40052` |

Both final runs pass the complete native window/music/storage smoke. Fullscreen geometry is (0,0,1920,1080) or (0,0,5120,2880); windowed geometry returns to the saved (100,100) rectangle with original dimensions. Each transaction has zero surface recreation within the style operation. The same HWND is retained throughout each scene. Styles, placement and normal/maximized restoration remain tested; NORMAL/HIGH/OFF, make-edge/repeat suppression and Enter consumption remain shared.

Windows-compiled PCM/frame cadence/oracle passes under Wine (1200 ticks / 2,300,000 / ball 2); current Linux full-suite results are in `personal-clean-seeds.md`; source checkers are all zero. Linux X11/PulseAudio fullscreen journey passes. Portable executable/AppImage storage/build validation is in [the portable report](pf12-portable-release.md). The new portable work was explicitly added to PF12 closure by the user; PF12.1 automation/polish has not begun.

**Remaining acceptance: PF12 is still blocked on repeated audibly clean real-Windows transitions.** These automated Wine results are regression evidence, not proof of the real-Windows cause or audible fix. Collect the real transition call/event/queue trace and listen for stutter/click with continuous selector and table music before declaring PF12 fully accepted.

## Personal single-file release follow-up

Personal Windows and Linux builds now contain the same 12 validated runtime originals; no original data or personal artifact is committed/pushed. Windows uses an embedded ZIP and private read-only temporary extraction with clean-exit cleanup; Linux reads its mounted bundled data. External userdata/fallback and explicit data overrides remain supported. Gameplay, mixer and deadline scheduling are unchanged. See [portable release report](pf12-portable-release.md) for input inventory, artifact hashes and cold-directory tests.

The personal EXE passed Wine selector/all-four-table controls, captured audio and four Alt+Enter switches. On the 5120x2880 Wine desktop, transitions measured 0.4173, 8.6881, 0.4255 and 8.3713 ms; pre/post queues all 12800 bytes (66.667 ms upper bound), zero empty/reset/clear deltas, starts=1, persistent HWND=65622 and exact windowed restoration (146,90,1288,1566). Wine measurements do not establish the real-Windows cause or listening result. Repeated real-Windows Explorer launch and audible transitions remain pending; PF12 is not fully accepted, and PF12.1 is not started.

The final bounded continuous Wine recheck also passed both selector and Party scenes (four switches per scene): selector durations 0.5189/13.8322/0.4399/13.5853 ms, queues 12800→12800, 12800→9600, 12800→12800, 12800→9600 bytes; Party durations 0.4678/9.1245/0.9091/12.7617 ms, queues 13524→13524, 13524→10820, 10820→10820, 10820→10820 bytes. Both scenes retained one producer start, zero empty observations/resets/clears and the same HWND/restored geometry throughout. Local logs are `.build-personal/continuous-wine.log` and `.build-personal/continuous/*-transitions.jsonl`; real Windows remains unmeasured.
