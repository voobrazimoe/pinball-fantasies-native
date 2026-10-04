# Native host phase validation status

## Current status (2026-10-04)

The native macOS ARM64/x86_64 host is implemented and accepted by the owner.
Original-backed replay/native journeys, clean packages and physical Mac gameplay
checks are recorded in [macOS validation](macos-validation.md); see
[macOS build and usage](macos.md). PR #1 was merged into `main` at `97c9777`;
public desktop packages were published as `v0.1.2`.

Android development is separate on `android-host`. A0 native engine builds and
the 16 KB ABI runtime smoke passed hosted CI. A1 has an asset-free GameActivity
and EGL/GLES synthetic renderer; JNI startup and orientation/lifecycle smoke
passed hosted CI at `c72f018` on a 16 KB emulator with interpreted ART.
A playable Android host and physical-device acceptance
remain outstanding. See [Android development status and remaining milestones](android.md).

## Historical foundation report (2026-10-03)

The platform/acceptance restrictions below describe that earlier milestone and
are superseded for macOS by the current status above. The historical test results
and implementation record are retained.

Validated on 2026-10-03. Engine work is ready for native host development;
macOS and Android applications are **not implemented or accepted**. The feature
branch must not be merged into main as a completed platform phase.

- Public starting main: `16d377ba39c9880cd3f5b76be8da27fac345e4d2`.
- Pre-platform parity checkpoint: `81ffea229c4a7b162fda6efa3fc8cde8ef330c23`.
- Engine implementation: `98f5344`.
- Conformance harness: `5477f83`; strengthened coverage: `4cb5b0444452e0363b486a90c5eebccd185a4782`.
- Canonical spring validity gate: `cc92236`.
- Branch: `codex/native-host-boundary-macos-android`.
- No tag or release was created. The checkpoint commit identifies the final
  validated Windows/Linux baseline; existing release/version conventions remain
  unchanged. CI does not automate publishing a new checkpoint release.

The incoming workspace was the separate private development repository, whose
main and preparatory checkout were preserved. An isolated clone of the requested
public repository was used. Its main included the Runner, DOS mouse plunger,
matrix audit (`c5e5472`, closed by `16d377b`) and original attract cheats.

The later repeated/right-Alt fix was committed and pushed in private commit
`5b3763446ef23e2a4c40f776efb13691d25b2779`, but absent from public main. Only the
AltGr pairing fix and its regressions were transferred into the checkpoint.
Windows can synthesize nonextended Left Ctrl followed by extended Right Alt
at the same message timestamp, raising both flippers despite release-only
sided polling. The window procedure now ignores that synthetic Ctrl pair;
independent left modifiers remain functional. `TestWin32AltGrRepeatedGameplay`
covers 60,000 repeat/alias/lost-break/focus cycles, and
`TestWin32AltGrQueuedControl` checks the real Win32 queue path. Both pass under
Wine; the former also runs on Linux. Physical keyboard/layout acceptance of the
reported delayed recurrence is still outstanding.

## Engine and conformance gates

Gate A passes: full original-backed Go regression suite, asset-free suite,
Linux build, Windows CGO_ENABLED=0 build, source payload checks and targeted
Alt regressions. Commercial and historical inputs remain external and untracked.

Gate B passes on Linux: existing full tests, four-table direct-versus-Go-boundary
traces and actual C shared-library traces. Each C trace has 2,714 host checkpoints
and 2,801 scheduled source tasks. Results match for every PCM hash/frame count,
mode/table/lifecycle state and selected framebuffer hash/dimension checkpoint.
The Go harness additionally compares complete table state, score, physics/ball,
spring, matrix, hotseat records and source deadlines at checkpoints.

| Table | SOFT | Full-table OFF |
| --- | --- | --- |
| Party Land | PASS | PASS |
| Speed Devils | PASS | PASS |
| Billion Dollar Gameshow | PASS | PASS |
| Stones 'N Bones | PASS | PASS |

Traces cover intro/selector, table load, all original attract cheat strings,
two-player start, original new-game presentation before launch, keyboard launch,
flippers/tilt, charged mouse fire, drains, new-ball/player handoff, focus loss,
one-hour suspension without catch-up, regain still paused, manual resume and
slow wakes with multiple overdue source tasks. Input tests also prove short
release edges, eight-count accumulated movement, one adjustment per task,
invalid-spring discard even when chute context remains true, scheduled-fire
cancellation and duplicate active Resume. The regression first failed against
the initial facade: a queued invalid fire consumed a later valid adjustment.
A read-only `PlungerValid` accessor now exposes canonical `Physics.SpringValid`;
one shared frontend validity check serves both engine and desktop adapters.
Shared settings persistence runs during destruction while suspended. Existing
four-table score save/restart/reset validation and portable storage checks pass.

The ABI and detailed ownership/threading contract are in
[native-engine.md](native-engine.md) and `cmd/pfengine/abi.h`. The engine is a small
facade over the authoritative Runner, not another game implementation. C exports
version/create/destroy/suspend/resume/set_action/key/release/plunger_delta/
plunger_fire/advance/frame/state. Holds and edges are logical; native codes stay
in hosts. Frame storage is engine-owned reusable C memory, read-only until next
frame retrieval/destroy. One copy stages the existing Go raster; no per-frame
host allocation/free is required. PCM is borrowed only during each synchronous
sink call, 48 kHz stereo signed little-endian int16. Hosts copy it into their own
bounded audio ring; device callbacks never enter the engine. Per-instance calls
are serialized; overlapping calls and callback reentry are rejected with PF_BUSY.

Linux c-archive and c-shared builds pass, as does a plain-C archive link/run
against the stable header. Go race tests pass for the Party Land SOFT trace,
Runner lifecycle/edge tests, native input tests and shutdown storage. Actual C
callback reentry and concurrent native-thread calls are rejected. A race-enabled
shared library builds, but loading it into Python aborts with a ThreadSanitizer
allocation error (`errno 12`); no successful instrumented foreign-runtime run is
claimed. Targeted Go vet passes. Whole-tree vet retains 13 pre-existing unkeyed
`audio.Effect` literals in unchanged `internal/speeddevils/audio.go:65`.

## Platforms and remaining acceptance

Gate C is blocked before AppKit implementation: no Xcode, xcrun, Apple SDK or
Apple-target C compiler is available. The pure-Go engine test binary cross-
compiles for macOS arm64, but C archive cross-building with the available Linux
gcc fails on `-arch`. No native package or runtime test exists. AppKit lifecycle,
nearest bitmap display, fullscreen, native keyboard/mouse, CoreAudio ring and
underrun behavior, NSOpenPanel import, app-support storage, all-table playback
and physical Apple Silicon acceptance all remain to be implemented/tested.

Android host development has not started because macOS has not demonstrated
its required native-host gate. The Android SDK/NDK/JDK, adb and emulator/device
are also absent. Pure-Go engine compilation for Android arm64 succeeds; a shared
library attempted with host gcc fails on arm64 assembly. No APK, JNI bridge,
Surface recreation, pointer-ID touch/cancellation, AAudio/Oboe ring, audio focus,
SAF import, app-private storage or emulator/physical-device acceptance is claimed.

## Desktop/build checks

- Full `go test -p=1 -count=1 ./...` passes with owner-supplied originals and
  private reference inputs; a clean asset-free Git archive also passes with
  original-backed checks explicitly skipped.
- Windows amd64 CGO_ENABLED=0 and native Linux executable builds pass.
- Wine: sided/repeated Alt, real queue pairing, held input, focus/pause,
  fullscreen/music transitions, relative mouse and cursor policy pass on an
  isolated Xvfb display. No real Windows hardware/listening claim is made.
- Linux: SDL relative mouse at different window sizes and the existing four-
  table OFF/NORMAL/HIGH input/presentation/resize/pause/quit journey pass. The
  latter uses dummy audio; its empty-queue observations do not certify audio
  listening/underrun behavior. Desktop window-manager/multi-monitor fullscreen
  and physical audio acceptance remain existing limitations.
- Public and personal Windows EXE/Linux AppImage packaging passes. Personal
  payload validation covers exactly 12 original inputs; TABLE*.HI is excluded.
  Both public and personal portable-storage suites pass on Linux/Wine for
  Unicode/spaces, default paths, explicit overrides and fallback directories;
  originals remain unchanged. These storage runs are headless (`live_seconds=0`).
- All eight discovered Python tool regressions and public source checks pass.
- Tracked tree audit finds zero forbidden commercial payload paths and zero
  matches for 721 nontrivial original 4 KiB PRG/MOD blocks at any byte offset.
  Public builds contain no original inputs; personal artifacts remain ignored
  and local only. No original game data is added or uploaded.

Material additions: `internal/engine/{engine.go,engine_test.go,input_test.go}`,
`cmd/pfengine/{main.go,abi.h}`, `cmd/pftrace/main.go`, `tools/build_engine.sh`,
`tools/test_engine_abi.py`, the two AltGr regression files, this status document
and `docs/native-engine.md`. Material updates: shared Runner focus helper/task
counter, canonical spring validity accessors on the four sessions, shared
frontend/desktop plunger gating, Win32 AltGr message handling, architecture/build
docs and asset-free
Linux CI archive/shared-library builds. Table/physics/matrix/audio gameplay semantics
are unchanged; table code only gains read-only spring-validity accessors. Hosted CI itself has not been rerun by these local checks.

Only the feature branch is to be pushed. No main integration is authorized by
the unfulfilled macOS/Android gates; public main remains the starting SHA.
