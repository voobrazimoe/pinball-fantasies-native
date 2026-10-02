# PF12 native Windows port

Date: 2026-10-01. Implementation and Linux/Wine validation are complete as described below. **Real Windows 10/11 visual/audio acceptance is pending the user's local test.** Wine was explicitly authorized as interim validation. PF12.1 packaging has not started.

## 1. Platform boundary before and after

Before PF12, platform/frontend.go combined SDL window/input/texture presentation, host timing and shared frontend/audio lifecycle decisions. Other platform files directly provided SDL audio and development presentation loops. Shared command startup chose data/config paths.

After PF12, platform/runtime.go owns one shared frontend runner, deterministic source deadlines, neutral DOS key events and the existing producer/pause lifecycle. A small host boundary provides openHost, Present, Pump/Event/Held, modal ticks, Close and window presentation state. AudioDevice has OS implementations with the same Queue/Suspend/Close contract. Storage and release diagnostics are platform policies. Physics, rules, settings, tracker/mixer and logical presentation remain shared; no second game implementation exists.

## 2. Linux SDL dependencies

Linux keeps SDL2 through its existing CGO bindings in frontend.go, audio.go, physics.go, window.go and fullscreen_sdl.go, now Linux-tagged. Existing diagnostic loops and software renderer behavior were retained. SDL audio production/lifecycle code was not rewritten; a read-only queue-depth accessor was added for the pacing audit. The shared runner's scheduling changed as documented in the separate audit.

## 3. Windows has no SDL dependency

The Windows target is pure Go plus thin syscall bindings to standard system DLLs. CGO_ENABLED=0 builds successfully. The Windows platform file selection excludes SDL files and has no CgoFiles. Static PE imports are Go runtime kernel32 functions; native backend DLLs loaded dynamically are user32, gdi32, winmm and, only for older DPI fallback, shcore. Actual release execution under Wine logged no SDL DLL load. No SDL2.dll is shipped, and no third-party window/audio framework was added.

## 4. Build tags

Common files: runtime.go, fullscreen.go, sizing.go, transfer.go, keys_windows_contract.go and pacing.go. Windows implementations: win32_windows.go, audio_windows.go, diagnostics_windows.go, storage_windows.go and native_windows_test.go. Linux implementations: frontend.go, audio.go, physics.go, window.go, fullscreen_sdl.go and storage_linux.go. Shared geometry/input tests run on Linux as well.

## 5. Win32 window

RegisterClassExW/CreateWindowExW create one persistent HWND. The UI runner locks its OS thread. Selector, Options, attract, play, pause and resolution transitions retain that HWND; Close requests go through the shared frontend lifecycle. Checked API failures identify the failed operation. Native tests verify persistent HWND across logical modes and fullscreen transitions.

## 6. Framebuffer / GDI

Shared RGBA frames are converted at the platform boundary to packed top-down BGRX. A negative DIB height prevents row inversion. COLORONCOLOR StretchDIBits renders to a cached compatible memory bitmap, followed by one visible BitBlt. Black letterboxes are composed offscreen. WM_PAINT redraws retained content; WM_ERASEBKGND is consumed. Origin/stride, channel order, row order and crop geometry are covered by deterministic tests.

## 7. Logical modes

Table frames remain NORMAL 320×240, HIGH 320×350 and OFF 320×609 (33-row matrix over 576-row complete playfield). The selector retains its existing presentation policy. Logical mode changes replace presentation content only; they do not recreate host windows, gameplay or audio producers. Existing full-table camera/layout tests remain shared and pass under Wine as well as Linux.

## 8. Resize / aspect

WM_SIZE retains the frame, invalidates the cached host surface and recomputes a centered destination using actual logical aspect, with letterbox/pillarbox as needed. No arbitrary 4:3 adjustment is applied to table frames. User window dimensions persist through logical changes during runtime; fullscreen restores windowed placement. Initial size uses the established desktop-fitting policy, including tall OFF mode. No forced integer-scale mode jump or new minimum-size rule was introduced.

## 9. DPI

Prefer per-monitor DPI awareness v2, with explicit per-monitor v1 fallback for early Windows 10. WM_DPICHANGED accepts the suggested host rectangle when appropriate; client pixel geometry controls scaling. Logical coordinates never depend on DPI. Geometry tests cover 100%, 125% and 150%. Physical mixed-DPI monitor behavior has not been tested on Windows hardware.

## 10. Borderless fullscreen

Alt+Enter saves window style, WINDOWPLACEMENT and the exact normal screen rectangle, removes overlapped decorations and covers the current MonitorFromWindow monitor's display rectangle. Leaving restores normal or maximized state and geometry. No display-mode, refresh-rate or exclusive fullscreen API is used. Negative-coordinate monitor argument tests and native normal/maximized restore tests pass.

## 11. Alt+Enter consumption

Neutral hostKeys consumes the Return make before normal frontend Enter. Both Alt identities work. Previous-key-state/retained-down state suppress autorepeat and duplicate makes. Shortcut Return never reaches Options, launch, initials or dialogs. Release/focus clears the latch; plain Alt retains flipper semantics.

## 12. Keyboard mapping

Win32 virtual keys and physical scan/extended bits translate to existing DOS-style key codes and held-control bits. Left/right Shift, Control and Alt remain distinct. Extended arrows are distinguished from keypad keys, and physical letter scans preserve game controls across keyboard layouts. Existing frontend F-keys, arrows, Return, initials, pause, quit, M, spring, tilt and flippers use shared logic.

## 13. Focus / stuck keys

WM_KILLFOCUS clears held host keys and pending makes, retains queued close requests and resets the Enter latch. It does not synthesize plunger release or pause/reset/restart gameplay. Native focus and shared make/break tests cover this behavior.

## 14. Timing

Go monotonic time.Now/time.Until and time.Sleep schedule the existing 60 Hz intro / 71 Hz table model. On Windows the Go runtime uses native high-resolution timing facilities, including QueryPerformanceCounter; game time remains source syncs. Every overdue sync and PCM batch executes before the next host frame. Win32 resize modal loops use a checked 8 ms SetTimer callback invoking the same shared deadlines on the UI thread. See [fullscreen pacing audit](pf12-fullscreen-pacing-audit.md) for measured independence from host frame rate and bounded-stall limitations.

## 15. Windows audio

waveOut consumes existing 48000 Hz signed 16-bit stereo interleaved PCM. WAVEFORMATEX and WAVEHDR ABI sizes/offsets are tested. Native LocalAlloc blocks hold headers and copied PCM until completion; waveOut never retains a Go heap pointer. No tracker, sample mixer, interpolation, cue or jingle changes were made.

## 16. Queue / lifecycle

Completed headers are polled and unprepared/freed. Queue depth is bounded to 250 ms and 32 headers, with the existing 2048-frame preroll. Output overflow clears output only, never source state; errors propagate. Pause/producer changes reset and pause the sink, then rebuild preroll on resume. Visual transitions do not touch producers. Shutdown resets, unprepares/frees and closes waveOut, then releases GDI/window resources through deferred cleanup. Counters expose empty queues/output resets/producer starts for validation.

## 17. Native writable storage

Windows and Linux now default native config/high scores/logs to `userdata/` beside the actual executable. AppImage resolves the outer `APPIMAGE` path rather than its temporary mount. Per the user’s subsequent fallback request, an unwritable portable root uses %APPDATA%\PinballFantasies on Windows or the user configuration directory/pinballfantasies on Linux. Explicit -config-dir/-high-score-dir overrides are checked and never silently fall back. A state directory aliasing the original data directory is rejected; original CFG/HI files remain read-only seeds. See [portable report](pf12-portable-release.md).

## 18. CWD / executable / original data

Windows default data discovery explicitly uses os.Executable's directory; it does not assume Explorer CWD. -data-dir is supported, with explicit relative paths resolved against launch CWD. Paths use Go filepath and UTF-16 Win32 bindings, supporting drive letters, spaces, Unicode and backslashes. Linux uses the actual native executable directory or the canonical outer AppImage directory too; source-checkout development runs should pass -data-dir . explicitly. Release launch from /tmp with an explicit Windows data path passed under Wine. Put legally obtained original data beside the executable or supply -data-dir; no folder-picker expansion or copyrighted embedding was added.

## 19. Reproducible build / diagnostics

Run `./tools/build_windows.sh` from Linux to produce `bin/pinballfantasies.exe` with GOOS=windows GOARCH=amd64 CGO_ENABLED=0, -buildvcs=false, -H=windowsgui and the release diagnostics setting. A regular cross-build without the release ldflags retains developer console/panic behavior. Existing -data-dir, -config-dir, -high-score-dir, -pf4, -pf7, -pf9, -pf10 and diagnostic flags remain available.

GUI builds append diagnostics to portable userdata/native.log (or the requested user-directory fallback/explicit -config-dir) and redirect both Go standard files and native standard handles. Fatal startup errors produce a native message box with subsystem/path context and data/log instructions; main-thread panic reports include a saved stack. No console is required for normal launch.

## 20. Executable dependencies

The application artifact is one x86-64 GUI-subsystem executable, `bin/pinballfantasies.exe`. Only Windows system DLLs are required; no SDL, CGO runtime, MinGW runtime or original assets are linked into it. Build metadata, static import inspection and actual Wine loaded-DLL trace are saved in pf12-validation/windows-build-info.log, windows-imports.log and wine-release-dlls.log. Debug/test executables in bin are validation artifacts, not release dependencies.

## 21. NORMAL / HIGH / OFF validation

Shared renderer/full-table tests pass. Native Win32 smoke exercised all three table dimensions, persistent HWND, redraw and resizing. The bounded Wine four-table journey exercised NORMAL/HIGH/OFF; its saved contact sheet shows matrix at top and complete playfields. Linux X11 exercises the same modes. Fullscreen aspect is covered by pure geometry and native host tests; real Windows visual acceptance remains pending.

## 22. Options validation

Settings logic remains shared: BALLS 3/5, ANGLE HIGH/LOW, SCROLLING HARD/MEDIUM/SOFT/OFF, MUSIC ON/OFF, RESOLUTION NORMAL/HIGH, SAVE AND EXIT. No MONO option was added. Shared settings/frontend tests and Linux/Wine bounded Options journeys pass. Legacy DOS import is unchanged.

## 23. Source checks

`python3 tools/matrix_reachability.py --check`: missing targets 0 across four tables. `python3 tools/matrix_operand_audit.py --check`: unresolved operands 0, source text byte mismatches 0, missing runtime text writers 0. Generated source-derived content/checkers were not forked or weakened.

## 24. Party oracle / determinism

Linux and Windows-compiled Party oracle executed under Wine pass at 1200 ticks, score 2300000, ball 2, frame SHA256 aec01b3a07e1a5ba10b3c635777a6742f4abd41c09899903ea532b98d8522913. The new host-frame-cadence regression compares PCM on every tick and final logical frame with every-tick versus every-eighth-tick presentation; both targets pass. Existing physics/rules/matrix/settings/mixer tests use the same expected values.

## 25. Windows visual smoke status

Real Windows: PF12 acceptance remains blocked by the user-reported audible Alt+Enter transition stutter; steady fullscreen pacing acceptance is a separate result. Wine: gated native window tests pass (logical modes, resize, focus, fullscreen, normal/maximized restoration, retained paint and same HWND). The four-table native application journey passes with zero empty observations/output resets. Linux SDL Alt+Enter journey passes. The supplemental Wine fullscreen journey also passes NORMAL/HIGH/OFF and restored-window aspect, Alt+Enter repeats/Enter consumption, exact geometry, stable HWND and unchanged desktop mode. Its harness now crops Win32 client pixels to exclude Wine/GNOME titlebar/shadows; an earlier whole-window aspect assertion was a harness error. This remains Wine evidence, not Windows visual acceptance. No Windows monitor mode change is requested.

## 26. Windows audio smoke status

Real Windows device acceptance: blocked on repeated audibly clean Alt+Enter transitions. The user heard stutter on real Windows during the transaction; it must not be inferred fixed from Wine. Wine waveOut format/queue/pause/resume/shutdown tests pass. Actual output capture confirms 48 kHz S16 stereo and distinct left/right 440/660 Hz tones. Selector/table/Options/M/pause/effect-producing controls are exercised by the native journey; deterministic shared audio tests cover cue/jingle content. Manual real Windows listening for effects/jingles, physical focus/input and fullscreen continuity is still required. Final bounded fullscreen real-output capture has one producer start, zero lifecycle clears and zero observed empty queues/resets. The longer fullscreen UI journey reports 5 empty observations and zero output resets across scene/pause/producer transitions; those aggregate observations are not classified as audible underruns. Queue timing is never used as the game clock.

## 27. Linux regression results

Required full Go suite ran with `./tools/go.sh test -p=1 -count=1 ./...`: this historical run predates factory-seed inventory validation; see `personal-clean-seeds.md` for current full-suite results. Focused platform/frontend tests pass. Linux executable builds with the required command. X11 four-table/Options/resize/modes and borderless fullscreen/geometry/input/audio-lifecycle smoke pass. The post-change X11 geometry journey used SDL dummy audio and recorded 40 empty observations; it is a window/input check, not PulseAudio continuity evidence. The separate real PulseAudio audit/capture below recorded zero. All Windows test packages compile with CGO disabled. Actual Wine Go suite additionally cannot launch Python for two source-checker wrapper tests; those checkers pass directly on Linux, and remaining Windows presentation tests pass with only those environment-dependent wrappers excluded.

## 28. Linux PulseAudio

Rerun was required after the shared scheduling change exposed by the user's pacing regression. SDL audio code only gained a read-only metric accessor; real PulseAudio was exercised in fullscreen present/no-present audits and actual monitor music capture. Both continuous-output audit/capture runs report zero observed queue empties/resets. The separate full Linux UI journey with PulseAudio passes and reports 7 aggregate empty observations, zero resets, 17 lifecycle clears and 10 starts across scene/pause transitions; these counters are preserved as evidence and not presented as zero-underrun interactive acceptance. This was impact-based validation of shared lifecycle/scheduling, not an unrelated audio archaeology experiment.

## 29. Original files

Original PRG/MOD/config/high-score files were not written. Before/after hashes for the 18 recorded original/dependency files are checked in pf12-validation/original-files-check.log. Local mutable scores remain untouched and are now excluded from runtime asset validation. Native smoke writes only isolated writable user/test directories.

## 30. Limitations / remaining acceptance

Real Windows 10/11 hardware smoke remains necessary: Explorer launch, physical make/break/focus, device listening, effects/jingles, resize drag, fullscreen visual continuity, multi-monitor restoration and physical DPI 100/125/150%. Older Windows 10 DPI fallback is implemented but not hardware-tested. GDI cost on the 5120×2880 Wine desktop limits visible frames; fixed source ticks/PCM continue independently of that frame count. A prolonged blocking host call can still exhaust a bounded queue. The Wine X11 bridge can change while HWND remains stable; the harness follows it and validates actual client geometry. No missing data chooser was added; documented -data-dir is the bootstrap.

## 31. Added / changed files

Added: internal/platform/{runtime.go,transfer.go,transfer_test.go,keys_windows_contract.go,win32_windows.go,audio_windows.go,native_windows_test.go,diagnostics_windows.go,storage_linux.go,storage_windows.go,pacing.go}; internal/frontend/host_cadence_test.go; tools/{build_windows.sh,smoke_pf12_wine.py,audit_pf12_pacing.py}; this report and pf12-fullscreen-pacing-audit.md.

Changed: internal/platform/{frontend.go,audio.go,physics.go,window.go,fullscreen_sdl.go,sizing.go}; cmd/pinballfantasies/{main.go,rules.go,speeddevils.go}. Changes in command diagnostic files only apply platform audio-error policy. Validation logs/captures are under analysis/pf12-validation; executable/test artifacts are under bin. Gameplay/core, original data and generated source content were not rewritten. Repository Git metadata is available; the PF12 closure is synchronized with upstream cleanup using a normal push (see the portable report).

## 32. PF12.1 boundary

Portable storage (with the later requested user-directory fallback), Windows executable/README and Linux AppImage/README were explicitly added to PF12 closure by the user and are now implemented. CI/GitHub Releases, additional release polish and Windows version/icon resources have not begun. PF12 is not fully accepted until real Windows Alt+Enter music listening is clean.


## 33. Alt+Enter transition blocker follow-up

The main frontend previously executed Alt+Enter inside input collection before that source update/PCM submission. The runtime now consumes the same make edge but queues the host operation until all deadlines due at that instant have submitted PCM; it rechecks after each transaction and before rendering. Win32 can also service elapsed source deadlines between offscreen clear/stretch/final blit calls and before host message calls during the 500 ms transition settling interval, without recursive Present. No future ticks, duplicate PCM, changed tracker/mixer or device-completion clock is used.

Style/geometry mutation brackets intermediate WM_PAINT/WM_SIZE work. Intermediate paints validate only, background erase remains suppressed, SWP_NOREDRAW/NOACTIVATE retain FRAMECHANGED, and one asynchronous invalidation follows. SetWindowPlacement remains necessary for exact normal/maximized restoration; its redundant ShowWindow was removed. The complete retained surface/pixels remain available. Client-size recreation occurs at normal Present/paint, not inside synchronous WM_SIZE. Letterbox clearing happens only when surface/logical dimensions change. A complete GetDC Present validates the covered client region, preventing a redundant BeginPaint/blit of the same resized frame on the next Pump. Native DPI policy initialization is cached so sequential native host smokes can open windows in one process.

`PF12_TRANSITION_LOG` is an opt-in JSONL trace. It records transaction duration, HWND, pre/post geometry, waveOut incomplete-header queue bytes/ms, independent empty snapshots, cumulative queue-empty/reset/start/clear counters, individual Win32 style/placement/monitor/SetWindowPos/invalidation calls and inclusive synchronous WM_SIZE/WM_PAINT/WM_WINDOWPOS*/erase handler counts/times. Surface recreation and draws for 500 ms after each transition are separate records; first draw is marked. Actual Queue empty observations include time and the last draw call. Snapshot queue milliseconds include the playing header and are an upper bound, not exact playback cushion. Diagnostic file I/O occurs after transaction snapshots and can perturb an instrumented run.

The bounded native `TestNativeFullscreenMusicTransitions` runs continuous selector and Party music for 10 s each with four Alt+Enter transactions at 2/4/6/8 s, including repeat and break handling. It checks HWND, exact restored geometry, unchanged producer/device lifecycle and zero queue empties. It runs the normal deadline/PCM path, not a synthetic tone or ahead-of-time producer. Tests are opt-in; data and output paths are supplied explicitly. Wine evidence, limitations and the outstanding real-Windows acceptance are detailed in [the transition pacing audit](pf12-fullscreen-pacing-audit.md).

Final transition smokes pass under Wine at both 1920×1080 and 5120×2880: selector/table each retain one HWND, restore exact geometry, and report zero queue empties, resets or lifecycle clears with one producer start. The timing/queue table and pre-fix Wine message-settling evidence are in the pacing audit. Real-Windows cause confirmation and listening acceptance remain pending.

## Personal single-file release follow-up

Personal Windows and Linux builds now contain the same 12 validated runtime originals; no original data or personal artifact is committed/pushed. Windows uses an embedded ZIP and private read-only temporary extraction with clean-exit cleanup; Linux reads its mounted bundled data. External userdata/fallback and explicit data overrides remain supported. Gameplay, mixer and deadline scheduling are unchanged. See [portable release report](pf12-portable-release.md) for input inventory, artifact hashes and cold-directory tests.

The personal EXE passed Wine selector/all-four-table controls, captured audio and four Alt+Enter switches. On the 5120x2880 Wine desktop, transitions measured 0.4173, 8.6881, 0.4255 and 8.3713 ms; pre/post queues all 12800 bytes (66.667 ms upper bound), zero empty/reset/clear deltas, starts=1, persistent HWND=65622 and exact windowed restoration (146,90,1288,1566). Wine measurements do not establish the real-Windows cause or listening result. Repeated real-Windows Explorer launch and audible transitions remain pending; PF12 is not fully accepted, and PF12.1 is not started.
