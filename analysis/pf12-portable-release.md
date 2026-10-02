# PF12 portable storage and release artifacts

PF12 includes the requested portable storage and release packaging. Later owner acceptance confirmed Windows Alt+Enter audio; older outstanding-acceptance statements below are historical records. Public release remains paused at the user's request. PF12.1 now means native hot-seat multiplayer, not release automation; see [the PF12.1 source and validation report](pf12.1-hotseat.md). No public repository, tag, release or published binary is created by this work. Transient multiplayer records never enter CFG or separate files; every qualifying player feeds the existing table-global userdata high-score store.

## Storage and data roots

The canonical directory of the actual executable is the portable root on Windows and native Linux. AppImage uses the canonical outer `APPIMAGE` filename, not `os.Executable()` inside a temporary mount. Spaces, Unicode and symlink resolution are supported; an invalid AppImage location produces an error instead of guessing a different data root. Explicit `-data-dir` remains supported and is resolved against launch CWD.

Public/default builds discover original files beside the executable/AppImage. Personal builds instead use their bundled original data; they need no original files beside the single distributed artifact. Explicit `-data-dir` overrides either default. Default writable config (`PINBALL.CFG`), native high scores (`TABLE*.HI`) and release `native.log` go to `userdata/` there. Original supplied PRG/MOD/CFG files are read-only runtime data/seeds and remain unchanged. Asset-free installations may supply optional legacy/read-only TABLE*.HI seeds; personal payloads contain none. Explicit state overrides aliasing the original data directory are rejected so native legacy filenames cannot overwrite the seeds. No Registry state is used.

The user's later instruction superseded the earlier no-fallback requirement: if creating/writing portable `userdata/` fails, Windows uses `%APPDATA%\PinballFantasies`; Linux uses `os.UserConfigDir()/pinballfantasies` (XDG_CONFIG_HOME or ~/.config). Writable portable storage never consults that fallback. Explicit `-config-dir`/`-high-score-dir` overrides retain priority and fail with an actionable error instead of falling back. If both default locations fail, the error asks the user to move the game or supply writable overrides.

The directory probe creates, writes, closes and removes a temporary file. State directories are created on first use. Release diagnostics initialize after flag parsing so `-config-dir` also controls `native.log`. Development builds retain terminal diagnostics; they still use portable config/high-score defaults. Source-checkout Linux runs should specify `-data-dir .`.

## Build outputs

- Windows: `release/windows/pinballfantasies.exe`, plus `README.txt`. One x86-64 CGO=0 GUI executable with Win32/GDI/waveOut and no SDL/third-party DLL requirement. Original game data is external.
- Linux: `release/linux/PinballFantasies-x86_64.AppImage`, plus `README.txt` and `BUILD-MANIFEST.json`. One executable distribution with SDL2, its ELF dependency closure, matching loader/libc, available ALSA plugins/configs and runtime copyright notices. Host desktop/audio servers remain external. AppRun uses the bundled loader with cache disabled and a private library path. Its POSIX shell uses the host shell/libc before exec; the game itself loads the bundled dependency closure.

`./tools/build_release.sh` produces both public asset-free releases. `./tools/build_windows.sh` remains the native Windows wrapper. `./tools/go.sh build -o bin/pinballfantasies ./cmd/pinballfantasies` remains the normal Linux development build.

`tools/build_appimage.py` requires Linux x86-64, Python 3, ldd, the repository Go/C toolchain and SDL2 build headers/library. Official appimagetool 1.9.1 and type2 runtime 20251108 are pinned by URL and SHA256 in `tools/appimage-tools.json`; cached downloads are checked before execution. `--tool PATH --runtime PATH` permits offline pinned copies. Build libraries/package versions and hashes are recorded in the manifest; reproducibility requires matching that build distribution/toolchain and source, not only matching appimagetool. SOURCE_DATE_EPOCH is configurable. No distro-specific .deb release is substituted.

Public packaging inputs are an explicit asset-free AppDir built from the native executable, system libraries, manifests, desktop metadata, notices and an original geometric SVG icon. No original game assets or local audit/reference trees are copied. Release outputs and `userdata/` are ignored by Git. Current artifact hashes are recorded in [PF12.2](pf12.2-public-preflight.md);
older portable JSON captures are historical.

## Personal one-file releases

The final personal release model supersedes the personal binary-plus-external-data workflow. Run:

```sh
./tools/build_personal_release.sh /path/to/original/game
# Offline pinned tools, if already downloaded:
./tools/build_personal_release.sh /path/to/original/game \
  --tool /tmp/pf12-appimagetool-x86_64.AppImage \
  --runtime /tmp/pf12-runtime-x86_64
```

This creates exactly two distribution artifacts, with no README, data pack or userdata beside them:

| Personal local artifact | SHA256 |
| --- | --- |
| `release/personal/windows/pinballfantasies.exe` | `7720508f43bd523989dc141d54b8b8d1c4b6c579196347eb12e4b5ba0a496d26` |
| `release/personal/linux/PinballFantasies-x86_64.AppImage` | `f56a89e5aeb719544aef2f1d77999543209396aaa83234d0fab7c7ae19a484c0` |

Copy just the appropriate file into a writable folder and launch it. Writable settings, scores and native.log remain external under userdata next to that file, with the agreed user-config fallback. The original supplied configuration here has music OFF; enable music in Options for table playback. Packaging preserves the seed bytes instead of silently altering them.

`tools/personal_assets.py` allowlists exactly the twelve graphics/module/configuration names below, checks presence and pinned inventory hashes, then invokes the production decoders through `cmd/personalvalidate`. Missing input lists the exact absent names before packaging. High-score files are excluded from both personal payloads. Factory defaults come from `frontend.Defaults(table)` and match pristine INTRO.ASM records. Optional external seeds remain supported in asset-free builds; mutable scores are written only to userdata/fallback state. Current rebuild and validation evidence is in [personal-clean-seeds.md](personal-clean-seeds.md). Rebuilding after replacing all four installation high-score files, then deleting all four, produced byte-identical Windows and Linux SHA256 outputs; the supplied files were restored afterward. Required files total **12**:

```text
INTRO.PRG  INTRO.MOD  MOD2.MOD
TABLE1.PRG TABLE1.MOD
TABLE2.PRG TABLE2.MOD
TABLE3.PRG TABLE3.MOD
TABLE4.PRG TABLE4.MOD
PINBALL.CFG
```

PRG and MOD files contain the graphics, table source data and tracker samples consumed by the native runtime. DOS executables, SDR development/reference files and commercial reference distributions are not runtime inputs. No parser, tracker, mixer or gameplay logic is duplicated or changed.

Windows uses the authorized controlled-extraction implementation to retain the existing filepath loaders. The builder generates an ignored `windows && personal` Go embed file and a deterministic ZIP of these 12 inputs, compiles an amd64 CGO=0 GUI EXE, then deletes both generated source and ZIP in a finally block. Public Windows builds and Linux builds define an empty payload. An explicit data directory bypasses extraction. Default personal launch extracts once into a unique private system temporary directory, marks seed/assets read-only and defers cleanup until clean exit. Failed extraction cleans partial files; archive traversal/nested names, duplicates and malformed bundles are rejected. Cleanup only touches actual children of that private directory. Crashes/forced termination can leave a private temp directory; it is never writable native state or part of the distribution.

Linux uses the same validated 12 files inside `usr/share/pinballfantasies`, with read-only modes in the AppImage. AppRun supplies this embedded `-data-dir` automatically unless an explicit data override was supplied. Outer APPIMAGE continues to select userdata even in extract-and-run mode. The native SDL backend and bundled runtime closure are unchanged. `tools/build_appimage.py --mode personal --data-dir PATH` also builds only the personal Linux artifact; public mode forbids a data input.

Personal artifacts, manifests, captures and staging are local-only. `/release/personal/`, `/.build-personal/`, `/.personal-assets/`, the generated Windows embed source and its asset directory are ignored. AppDirs use unique temporary directories under ignored .cache and are removed after packaging; builders verify their commercial destinations are ignored. Build locking prevents two Windows payload generators from racing. Input names/sizes/hashes and output hashes are retained only in ignored `.build-personal/build-manifest.json`; Linux dependency metadata is recorded there too. Public packaging copies no originals. Commit/push includes only code, scripts, tests and documentation.

### Personal artifact validation

- Both artifacts were copied alone to empty folders with spaces/Cyrillic/Japanese path components, launched from `/tmp`, reached selector and F1–F4 attract/playing modes, accepted spring/flipper controls and four Alt+Enter edges, and created only external userdata. No PRG/MOD/CFG/HI or SDL DLL was placed next to either file.
- Settings changed through Options persist on restart. Production FileStore tests prove high-score seed load, mutable save/reload and unchanged seed bytes; an earned high-score entry on real hardware is not claimed from this automated test. Both artifacts honor explicit data/state overrides and fall back when a file deliberately blocks portable userdata.
- Linux all-table input journey passes on isolated X11/Xvfb with PulseAudio. Synthetic keyboard activation on the GNOME desktop did not produce game input, so that desktop automation is not counted as acceptance. Direct executable launch needs no data flags or sidecar assets. Desktop launcher/double-click behavior with physical input remains a manual check; Windows Explorer and real-Windows audible transition checks remain outstanding.
- Local PulseAudio monitor capture recorded nonzero stereo signed16 audio in all four table intervals on both Linux and Wine after enabling music, including spring/flipper inputs. This proves output for these intervals, not real-Windows audible glitch acceptance. Captures contain commercial music and remain ignored/local-only.
- Personal Wine transition measurements on the 5120x2880 desktop: 0.4173 / 8.6881 / 0.4255 / 8.3713 ms. Queue header depths pre/post were 12800 bytes (66.667 ms upper bound) on all four; empty/reset/waveOut-reset/clear deltas were zero, starts stayed 1 and HWND 65622 persisted. Windowed geometry restored exactly to (146,90,1288,1566). These are Wine timings, not real Windows timings.
- All 12 input hashes and both finished artifact hashes remain unchanged after testing. Windows extraction/read-only/cleanup/override tests pass on Linux and as Windows tests under Wine. No private data directory remains after normal exits. PE/build-info inspection confirms Windows GUI amd64 CGO=0/personal build with no SDL import. LD_DEBUG confirms the personal AppImage uses bundled SDL and all 43 linked libraries rather than host SDL; a separate machine without SDL installed is not available here.
- Source reachability, operands, text and writer checkers remain zero; current full Go suite results are in `personal-clean-seeds.md`. Party 1200 ticks / 2,300,000 / ball 2 and PCM/frame cadence tests pass Linux and Windows/Wine. Normal development Linux and public Windows builds pass.

Reproduction: `tools/smoke_personal_release.py --backend linux|wine --capture-audio` (one desktop journey at a time; Linux test needs libXtst), and `tools/smoke_personal_storage.py --backend both`. Wine requires the native platform test helper built into `bin/platform-windows.test.exe`. Local records live in ignored `.build-personal`, including cold smoke JSON/logs, transition JSONL, storage results, audio captures, oracle/checker/build logs. **PF12 remains blocked on repeated audibly clean real-Windows Alt+Enter transitions. PF12.1 has not begun.**

## Validation

- Portable root/storage tests pass on Linux and in the Windows-compiled native tests under Wine: native roots, outer AppImage selection, invalid APPIMAGE errors, spaces/Unicode, Linux executable symlinks, portable preference, blocked-directory fallback, override errors and CFG/HI seed preservation. Windows symlink/chmod permission semantics are not asserted using Wine's Unix emulation.
- Actual artifacts are copied to temporary folders containing spaces/Cyrillic/Japanese names, supplied data is copied only to those test folders, and launch CWD is `/tmp`. Both produce a PNG with default external data discovery, launch a window/audio backend for 3 s, create portable `userdata/native.log`, honor both state overrides, and use the requested user fallback when a regular file deliberately blocks `userdata/`. Input seeds and checkout originals remain unchanged.
- AppImage runs via `--appimage-extract-and-run` without FUSE/root. LD_DEBUG confirms all 43 linked game libraries, including SDL2, load from the extracted AppImage. This is dependency-path evidence; a second physical distribution with SDL uninstalled and Wayland/direct-ALSA hardware are not yet tested. The temporary AppImage mount never becomes the writable state/data root.
- Wine overwrites APPDATA for a Unix-launched process. The artifact helper launches a native child with an explicit Windows environment to isolate fallback tests, matching native CreateProcess behavior. The game itself retains the standard Windows user-directory policy.
- Current full Go suite results are in `personal-clean-seeds.md`. Source reachability/operand/text/writer checkers are all zero. Party's existing 1200-tick / 2,300,000 / ball 2 PCM/frame cadence oracle passes on Linux and as a Windows-compiled test under Wine. Native Windows build has CGO disabled; its import/dependency records contain no SDL. Linux normal build passes.
- Linux X11 fullscreen/mode/options/geometry/make-edge journey passes with PulseAudio; its 8 aggregate empty observations are across scene/pause/producer transitions and zero output resets. This is separate from continuous transition music acceptance.

Logs: `pf12-validation/portable/{go-suite.log,source-reachability.log,source-operands.log,wine-determinism.log,windows-cgo-deps.log,windows-imports.log,artifact-smoke.json,artifact-smoke.log,linux-x11-fullscreen.log,original-hashes.json}`. Transition evidence and remaining Windows acceptance are in [the pacing audit](pf12-fullscreen-pacing-audit.md).

## Repository synchronization

The requested local-only reference directories and root .deb files are backed up outside the repository before synchronization. PF12 source, reports, tests and packaging scripts are committed, rebased/synchronized onto current origin/main, and pushed normally. Upstream cleanup commits and ignore rules are preserved. Restored reference material remains ignored/untracked; the final tracked tree excludes commercial distributions and generated local audit checkouts. Release binaries are local build outputs, not committed commercial-data bundles.

The final bounded continuous Wine recheck also passed both selector and Party scenes (four switches per scene): selector durations 0.5189/13.8322/0.4399/13.5853 ms, queues 12800→12800, 12800→9600, 12800→12800, 12800→9600 bytes; Party durations 0.4678/9.1245/0.9091/12.7617 ms, queues 13524→13524, 13524→10820, 10820→10820, 10820→10820 bytes. Both scenes retained one producer start, zero empty observations/resets/clears and the same HWND/restored geometry throughout. Local logs are `.build-personal/continuous-wine.log` and `.build-personal/continuous/*-transitions.jsonl`; real Windows remains unmeasured.

## PF12.2 superseding audit

Current candidate hashes, clean-clone evidence, license decisions and publication
gates are in [pf12.2-public-preflight.md](pf12.2-public-preflight.md). Earlier
hashes in this document are historical. Personal builds require exactly the
12 PRG/MOD/CFG inputs, never TABLE*.HI; score validation uses isolated storage.
All public artifacts are built separately with no commercial payload. Optional
ALSA A52/FFmpeg rate plugins are excluded from the AppImage dependency closure.
PF12.1 multiplayer is unchanged. Wine checks do not establish physical Windows
acceptance. Linux multi-monitor fullscreen flicker remains a known beta issue.
