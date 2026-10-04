# Foundation integration and macOS validation record

## 2026-10-04 input and latency follow-up

The owner's first Intel timing report (from the earlier build without the
`frames` counter) showed playing-mode timer gaps up to 35.321 ms, event delivery
up to 32.070 ms and event-to-draw completion up to 59.162 ms. In those playing
intervals engine work stayed below 0.7 ms and drawing below 1 ms. This points to
host scheduling/delivery delays rather than expensive source updates; screen
scanout latency is not measured by that log.

The next host revision replaces the tolerant AppKit timer with a strict
zero-leeway main-queue dispatch timer, requests foreground latency-critical
scheduling, and services already-due source work and drawing from keyboard event
handlers. It never advances a future source task. Inactive windows stop the timer
and end the activity; a reentry guard protects serialized engine work. Local
live checks cover fullscreen/restoration, minimize/focus pause, fresh input after
explicit resume and clean shutdown. Typical steady timer gaps in the development
run are about 8.4 ms. Native asset-free tests and source/engine/frontend regressions
pass. Intel retesting remains required; local measurements do not prove its lag
resolved. Bundle metadata and timing headers now identify revision/architecture,
and packages also have revision-specific filenames to avoid mixing test builds.

The owner confirmed left/right Shift, Control and Option on an Intel MacBook's
built-in keyboard. Command remains reserved for native shortcuts. Parsec input
on the development Mac was observed to omit modifier side identity in both
AppKit and CoreGraphics events; Z/slash and Left/Right Arrow provide independent
alternate flipper controls without changing normal sided modifier semantics.
The owner still reports severe latency with held keys and the alternate controls
on Intel hardware; that acceptance issue remains unresolved pending its timing log.

The host now retrieves/copies a raster only after source ticks advance instead
of doing that work at every 120 Hz poll. Local AppKit logging confirms about
71 rasters and 60 draws per second during gameplay; a sampled input interval
measured 20.740 ms from event to post-source-tick drawing completion. This is
an Apple Silicon development measurement, not Intel or physical screen latency
acceptance. `RunWithTiming.command` in each zip enables the optional timing log.

Using the owner's imported originals read-only and temporary native state,
all eight macOS loaded-C-library/direct-Go replays passed (four tables in SOFT
and OFF, 2,714 checkpoints each with frame/PCM/state parity). All 32 Apple-linked
native journeys passed (four tables, four scrolling settings, two resolutions).
A direct engine probe observed first flipper movement on the next scheduled
update and full travel after 50 ms at its 60 Hz host polling cadence on all four
tables. Asset-free native tests and source/engine/frontend regressions pass.
The Intel artifact is cross-built and checked for x86_64 architecture, signature
and public bundle hygiene; native Intel test execution remains owner-side.

The historical milestone record below is retained. The macOS PR remains draft
and unmerged; current original-backed automated results do not establish full
physical acceptance or resolve the owner's Intel input-lag report.

This records the two milestones separately. The macOS branch is **not merged**.
Original-backed macOS replay and physical acceptance are **PENDING** on the
owner's authorized Apple Silicon Mac with Codex, as explicitly agreed during
this task. Hosted build/unit checks proceed independently. No Android work,
release/tag, Developer ID credential or notarization was introduced.

## Foundation

| Item | Result |
| --- | --- |
| Original public main | `16d377ba39c9880cd3f5b76be8da27fac345e4d2` |
| Validated foundation branch | `codex/native-host-boundary-macos-android` |
| Validated foundation HEAD | `4a02f91f78d40753565db9f0cbd0c83f556d35ec` |
| Resulting foundation main | `4a02f91f78d40753565db9f0cbd0c83f556d35ec` |
| Integration | Clean fast-forward, pushed to `origin/main`, fetched and SHA verified on 2026-10-03; no force push |
| Final repeated-Alt fix | `internal/platform/win32_windows.go` calls `windowsAltGrControl` from `keys_windows_contract.go` to ignore only the synthetic same-timestamp Left Ctrl/Right Alt pair |
| Alt regressions | `keys_altgr_test.go`: 60,000-cycle repeated/synthetic regression and independent-left preservation; `keys_altgr_windows_test.go`: actual Win32 queue test; PASS under Wine, including the final macOS-branch rerun |

The foundation pre-merge recheck passed full original-backed and asset-free Go
suites; engine/source/input-edge tests; direct-vs-boundary and real C-library
conformance on all four tables in SOFT/OFF; original cheat sequences; Linux and
Windows CGO_ENABLED=0 builds; public/personal packaging and portable-storage
smokes on Linux/Wine; Python regressions; and public source/payload hygiene.
A mutable local PINBALL.CFG was initially rejected by the pinned inventory test;
the suite was rerun green with a checksum-matching pristine test seed before
merging. That historical workaround is not a requirement: PINBALL.CFG is optional
mutable state. The macOS validator now pins only required PRG/MOD payloads, copies
settings into temporary storage, and checks external inputs remain unchanged. Local data/references remained external, ignored and unuploaded.

## macOS branch and implementation

| Item | Result |
| --- | --- |
| Feature branch | `codex/macos-arm64-host`, created from the newly verified foundation `origin/main`, committed in logical pieces and pushed |
| Feature HEAD / exact CI run | Reported by the final task response; GitHub workflow runs identify the tested head SHA |
| Runner | GitHub-hosted standard `macos-15`; actual image `macos-15-arm64` |
| Architecture gate | Actual `uname -m`: `arm64`; any other value fails |
| Observed macOS | 15.7.9, build 24G830 |
| Selected Xcode | 16.4, build 16F6; explicit `DEVELOPER_DIR` |
| Observed clang | Apple clang 17.0.0, clang-1700.0.13.5 |
| Go | 1.27.1 darwin/arm64, matching project policy |
| C ABI build | Apple cgo/compiler/SDK build of archive and shared library; stable ABI 1 unchanged |
| C ABI conformance on Mac | Real archive and loaded shared-library asset-free contract checks PASS; original-backed four-table direct-Runner replay PENDING |
| App build | Native Objective-C/AppKit executable statically linked to existing engine C archive |
| Executable | `file` identifies Mach-O arm64; `lipo -archs` must equal `arm64`; system frameworks inspected using `otool -L` |
| Renderer | CoreGraphics RGBA copy, nearest-neighbour, black bars, shared framebuffer geometry; native offscreen colour/orientation assertions |
| Keyboard/modifiers | Physical keycodes to DOS logical makes; duplicate/repeat filtering; IOKit sided masks through flagsChanged; ordered cheat makes and focus clearing tested |
| Mouse/cursor | Native relative deltaY through canonical ABI plunger functions; one fire edge per press; bounded local hide/unhide while valid/inside/focused; no warp/capture |
| Audio | 48 kHz stereo S16, bounded C SPSC ring, main-thread generation/copy; realtime AudioUnit consumes only host PCM; callback/channel/silence/flush tested without device |
| Asset import | NSOpenPanel, regular-file whitelist, staged copies, shared Go validation through create/destroy, transactional adoption; no commercial payload ships |
| Persistence | `~/Library/Application Support/PinballFantasies/{Data,State}`; shared Go settings/high scores; no writable `.app` or CWD dependency |
| Four-table host journeys on Mac | Compiled for every table, four scrolling settings and NORMAL/HIGH resolution; execution with originals PENDING |
| macOS workflow | Dedicated native build/tests/bundle/hygiene/package and bounded AppKit synthetic-frame launch; green hosted runs are linked by the final task report |
| App artifact | Public `.app` zip uploaded by CI, ad-hoc signed only; no Developer ID/notarization claim |
| Linux regression | Final full Go suite, four-table Go/C replay including cheats, source/desktop plunger, matrix/presentation, and executable build PASS |
| Windows regression | CGO_ENABLED=0 build and repeated/synthetic/actual-queue Alt tests under Wine PASS; existing Linux/Windows hosted CI retained |
| Commercial-data hygiene | Exact app file allowlist and original whole-file hashes checked in hosted CI; original 4 KiB-block scan performed locally with supplied originals against the public artifact |
| Merge status | macOS remains unmerged; foundation remains on `origin/main` at `4a02f91f78d40753565db9f0cbd0c83f556d35ec` |

The eight Linux C replays each compare 2,714 checkpoints and 2,801 scheduled source
tasks, with PCM/frame/state parity. Party Land, Speed Devils, Billion Dollar
Gameshow and Stones 'N Bones pass in SOFT and OFF. These Linux results do not
substitute for the pending Apple-hosted original-backed runs.

No shared Go gameplay/physics/matrix/audio implementation or C ABI change was
needed by the macOS host. The new tests distinguish asset-free ABI checks,
original-backed gameplay replay and physical acceptance. Hosted graphical launch
worked, but its synthetic view and clean exit do not constitute human hardware
acceptance or listening.

## Outstanding physical-Mac acceptance — UNVERIFIED

- Original-backed four-table C ABI/direct-Runner conformance and native host
  journeys, using `tools/validate_macos_originals.py`.
- Actual visible framebuffer/window rendering and every table/presentation mode.
- Keyboard feel and sided modifiers on real hardware/layouts.
- Mouse-plunger feel, fire behaviour and cursor UX.
- Real CoreAudio output, channel order, audio underruns and device reconfiguration.
- Focus/pause/resume and prolonged system sleep/wake.
- Fullscreen/restoration on one and multiple displays.
- NSOpenPanel interactive import, cancellation and invalid input handling.
- Settings/high-score persistence across real app launches, different CWDs and
  moved bundles, with state confined to Application Support.

[Build commands and the physical checklist](macos.md) provide the next-session
handoff. The macOS branch must not merge before this validation is completed and
integration is authorized.

## Intel scroll-lag follow-up (Core Animation presenter)

Owner timing capture from build eaa7be6/x86_64 shows playing-mode drawing maxima
up to 31.403 ms, input queue delay up to 316.520 ms and input-to-draw up to
318.608 ms. Source advance is typically below 0.5 ms; the owner reports severe
lag during table scrolling. The old host forced synchronous view drawing on
keyboard events and scaled the changing bitmap in CGContextDrawImage on main.

The replacement uses immutable CGImages as Core Animation layer contents with
nearest filtering, no implicit animation, and no forced display in key handlers.
Timing draw_ms/input_draw_ms now end at layer submission, not physical scanout;
the renderer policy is included in the CSV header. Hardware scroll/input feel
on the Intel Mac remains pending owner retest.

Local validation: asset-free host/ABI/audio/storage tests and layer presentation
checks PASS, including retained-image immutability, colours/orientation, bars,
640×240 selector geometry and resizing without animations. Original-backed native
journeys PASS for all four tables, four scroll modes and two resolutions (32
journeys, 708 source ticks each). Live arm64 window and fullscreen/restoration
show correct pixels and geometry. x86_64 compilation and bundle checks PASS;
target executable performance is not measured on this arm64 development Mac.

## Built-in Shift follow-up

Owner confirms build 544ce58 is responsive with Control on the Intel MacBook's
built-in keyboard, while Shift is unreliable. Exact IOKit device side flags now
make native modifier updates idempotent instead of toggling on each event;
aggregate-only keycode fallback remains. Native tests cover duplicate makes and
breaks and missed sibling releases for Shift/Control/Option. Timing mode records
modifier-only flags and resulting sides/holds for any remaining hardware issue.
This addresses a found state-tracking weakness; the owner's actual Shift event
sequence has not yet been captured, so hardware acceptance remains pending.

## Shift side mapping follow-up

The owner's requested left-then-right sequence in build 12e4f1e logs key 60 /
flags 0x20104 / flippers 01 first, then key 56 / flags 0x20102 / flippers 10.
Host routing is consistent with the reported native codes and flags. A persistent
View → Swap Left/Right Shift option compensates for the observed reversal on the
owner's keyboard without changing default routing on other Macs. The option
recomputes held actions, preserves all other contributors, and appears in timing
metadata. Logic checks cover both sides, toggling while held, mixed Control,
Option and Z contributions, and focus clearing. Local menu activation and
NSUserDefaults persistence after relaunch PASS. Intel physical retest pending.
