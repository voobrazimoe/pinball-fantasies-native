# Architecture

The project has one shared native game implementation with platform-specific host backends.

`cmd/pinballfantasies` starts the frontend and table runtime. `internal/source` owns deterministic source progression. `internal/platform` owns OS event polling, native input translation, windowing, audio-device queues and writable storage. Linux uses SDL2; Windows uses native Win32, GDI and waveOut APIs.

`internal/assets` and `internal/physics` validate and decode the user-supplied DOS data. `internal/partyland`, `internal/speeddevils`, `internal/gameshow` and `internal/stones` implement table-specific rules and state machines in Go. Shared scoring, task scheduling and timing helpers live under `internal/tablelogic`.

`internal/audio` decodes and renders the original tracker/module data and effects. `internal/presentation` renders the matrix display and table presentation from typed native state plus records decoded from the supplied PRG data.

Multiplayer is implemented as a hot-seat session around a single active table engine. Up to eight player records preserve the source-persistent per-player fields; ball-local and transient table state is reset at the same lifecycle boundaries as the original game. Player handoff loads the incoming record before the next launch, including its pre-launch matrix presentation.

`internal/source.Runner` is a concrete, host-independent state object around `frontend.Runtime`. Its injected monotonic `Now` function and next deadline preserve the existing integer-duration cadence: INTRO advances at 60 Hz, table modes at 71 Hz, with the cadence selected before each update. `Submit` snapshots held logical controls and accumulates one-tick edges. `Advance` updates every due source tick and delivers its PCM to a sink before returning. The latest framebuffer is available through `Frame`. Presentation may omit intermediate frames; it must never omit source updates or their PCM. The runner has no SDL, Win32, window, fullscreen, filesystem or audio-device dependency.

The desktop adapter still drains due ticks before fullscreen transactions and before presentation. Win32's source callback also refills PCM between blocking GDI operations, without recursively presenting. The host controls its audio queue, flush/suspend policy and fullscreen/window transactions; these do not define source time.

`internal/gameplay.Controls` is the shared gameplay boundary: left/right flipper holds, nudge hold, plunger pull/release, relative spring adjustment and mouse fire edge. `physics.Inputs` aliases this logical value. Frontend DOS make codes remain for selector/options/initials and pause/quit interactions. The legacy frontend replay fields remain compatible, but desktop hosts submit `Gameplay`. Sided Win32 Shift/Ctrl/Alt reconciliation and SDL scancodes remain in their respective host files.

The mouse plunger restores original MS-DOS behaviour, rather than adding a new gameplay feature. The behavioural oracle is `FANTASIE.ASM`: `INIT_MOUSE`, `SPRINGSTEEN`, `SPRINGIT`, `SPRINGUP`, `SPRINGPOS`, `SPRING_VALID`, `MOUSETOTAL=2` and `MOUSEMIDDLE=1`. The host adapter translates relative counts using the original vertical ratio (64 mickeys per eight positions); `SPRINGSTEEN` adjusts spring position by at most one per source task, down to pull and up to relax, bounded to 0..32. A left-button press schedules `SPRINGUP` on the following source task, using an edge so holding the button cannot repeatedly fire. Both devices use the same `physics.Game.SpringPosition`; the shared spring helper calls each table's existing `Release` callback, preserving canonical `physics.Game.Release(charge, jitter)` arithmetic, spring reset and sound. Mouse controls have no effect outside a valid plunger; keyboard timing remains unchanged.

Linux consumes unscaled SDL relative Y motion and left-button presses. Relative mode is enabled only during active gameplay in the chute, and released on focus loss or leaving that context. Windows registers a native raw mouse device and consumes foreground `WM_INPUT`, excluding absolute pointer devices; it hides the normal pointer only on client hits during active mouse-plunger play, using `WM_SETCURSOR` and `SetCursor`. Mode/focus transitions refresh a stationary client pointer. It does not capture, confine, warp or recenter the pointer and does not use the global `ShowCursor` counter. Neither backend derives sensitivity from client coordinates, DPI, scrolling, framebuffer size or window size. No cursor is added to the framebuffer.

`Suspend` clears held controls, queued gameplay edges and pending mouse fire without updating the game or tracker. `Advance` does nothing while suspended; a close request can still complete shutdown. `Resume` clears input and re-anchors the next deadline to the current monotonic time, so elapsed suspended wall time never causes catch-up. Desktop focus loss additionally submits the existing frontend focus-loss pause request and suspends source advancement. Focus gain re-anchors the clock but leaves the game's pause in place until the existing resume key. Hosts may suspend/flush their audio devices independently. No game/tracker state is rewritten merely to implement suspension.

The tiny `internal/engine` facade and `cmd/pfengine` C ABI expose this same Runner to foreign hosts. See [native engine ownership, timing, threading and conformance](native-engine.md). The Objective-C AppKit macOS arm64 host in `hosts/macos` uses ABI 1 with CoreGraphics presentation and AudioUnit buffering; original-backed and physical-Mac acceptance remain pending. See [macOS build and validation](macos.md). Android remains unimplemented; no second game implementation is introduced.

The public repository contains no DOS/x86 execution path. Original game files are read only as data containers and are never executed.
Matrix pixels belong to source scheduling. The shared presentation layer retains
VGA dot memory, CODE2 digit/comma caches, countdown seconds/phase and scroll
phase. Routine completion, accepted replacement and idle score restoration have
separate source ownership. A lost match reaches the table's source clear and
`_CHECK_HIGH`; the frontend visits high-score qualification one player per sync,
then resumes the installed matrix program and demo task. Game-over presentation
continues DEMOMODE NODOT with retained memory. DOADDTASK clobbers BX in
`_2_DEMO_MODE`, so HU_ reads the following task slot instead of installing
the adjacent long wait. The source AFTERDEMOMODETS program then presents the game-over text and
player scores on their original source visits.
