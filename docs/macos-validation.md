# Foundation integration and macOS validation record

This records the two milestones separately. The macOS branch is **not merged**.
Original-backed macOS replay and physical acceptance are **PENDING** on the
owner's authorized Apple Silicon Mac with Codex, as explicitly agreed during
this task. Hosted build/unit checks proceed independently. No Android work,
release/tag, Developer ID credential or notarization was introduced.

## Foundation

| Item | Result |
| --- | --- |
| Original public main | `b5363935e5c4085ef197cb33f4fd9f778ddbe9fe` |
| Validated foundation branch | `codex/native-host-boundary-macos-android` |
| Validated foundation HEAD | `8bc62e361832ac2ae76b2da251d335eef533c9cc` |
| Resulting foundation main | `8bc62e361832ac2ae76b2da251d335eef533c9cc` |
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
merging. Local data/references remained external, ignored and unuploaded.

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
| Merge status | macOS remains unmerged; foundation remains on `origin/main` at `8bc62e361832ac2ae76b2da251d335eef533c9cc` |

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
