# Layout B validation — 2026-10-06

Implemented from main `629f88a`. Profile: `dos-powerpack-linked-v1`, identified by the audited Power Pack marker; Gold Pack provenance is unproved. A remains `dos-retail-linked-v1` and the sole pinned oracle. C/D, Deluxe/GOG and demos remain research-only/unsupported.

The shared loader selects a coherent profile across all five PRGs before decoding. B has 918 exact compiled copy spans (16 INTRO, 479 TABLE1, 423 TABLE2), derived from the reviewed per-region mapping. Adjacent equal translations may share a span: this does not reduce the 17 required INTRO FORM records. TABLE3/4 use the existing byte view. The descriptor carries addresses/sizes and one typed S_EMPTY constraint, never commercial payload. Equal records are copied; B priority 0 is preserved, never replaced with A priority 1. Source validation and decoded-address validation are separate. Every validation/start re-detects identity; no state marker or UI is added.

## Results

- Private A and both exact B copies: detection, mapped records and full FORM extents, all INTRO assets/geometry/sidebar, all four tables' decoded graphic/physics/control models, all six role-specific modules, four session factories and real Party Land S_EMPTY cue path PASS. All six INTRO/TABLE1/TABLE2 hybrids reject. Unknown, malformed, incomplete and altered cue semantics reject.
- S_EMPTY typed model is `(62,0,1)` for A, `(62,0,0)` for B. Real `Game.Cue` preserves these requested priorities; shared `MusicClock.Play` at current priority 1 accepts A and rejects B. Existing caller priority reset and scheduler initialization/return state remain unchanged.
- Supported macOS shared Go suite: 18 tested packages PASS, 7 packages without tests; 201 passing cases/subcases, 826 skipped for unavailable private source/captures/root inputs. Separate canonical engine regression PASS (260 seconds). Unfiltered `go test ./...` fails to compile the existing Windows/Linux-only `internal/platform` and its CLI on macOS; shared packages pass.
- Additional pinned-A tests in a disposable external source tree: 509 cases/subcases PASS, 36 SKIP, 7 failures caused solely by absent reference files (`pf2-fixture.json`, `pf2-ball-locations.json`, `pf6-art-fixtures.json`, `pf7-assets.json`, `pf9-assets.json`, `pf10-assets.json`). This extended reference suite is incomplete; its missing evidence was not generated or weakened. Temporary original copies/tree were removed.
- Python unittest scripts: 36 tests, 1 private-input skip, no failures. Parameterized ABI scripts ran separately with their required arguments.
- Android Java/C++ A3/A4/A5 host suite PASS. Private real-file DataImport transaction + shared runtime validator PASS: A/B staging/adoption, validation before stop, all six hybrids, malformed/incomplete rejection, rollback, interrupted adoption/cleanup recovery and relaunch. This uses provider streams and the actual transaction class, not device SAF UI instrumentation.
- macOS native A and B: import/source/staging validation, transactional adoption and relaunch PASS; 32 host journeys each (four tables × four scroll modes × two resolutions).
- A and B C ABI versus direct Go Runner: eight replays each, 2,714 checkpoints per replay, PCM/frame/state parity PASS. These compare implementations for each profile; they do not assert A/B or DOS-oracle equivalence.
- macOS C archive/shared library and asset-free app, Windows amd64 EXE, Linux amd64 trace tool, Android arm64-v8a/x86_64 engines and asset-free debug APK builds PASS. Android ELF LOAD alignment checks PASS at 16 KiB. Gradle required sandbox escalation solely for its local file-lock socket and ran offline with existing dependencies.
- Public source/address-only descriptor check, app icon check, macOS bundle allowlist and Android APK no-assets check PASS. Both public app/APK scans found none of the 836 nontrivial 4 KiB original blocks for each private A/B input set.

## Evidence boundaries and cleanliness

Oracle code/inventory/captures are unchanged. No release/version/tag/push action was performed; `v0.1.3` remains `c417afe073505531b6dcafeb45c36ca0d4630668`. New files contain no PRG/MOD/CFG/EXE/archive/image payload, and all test extraction/copy operations used system temporary directories outside the checkout. No commercial fixtures or research payload were added to tracked or untracked source files.

The checkout already contained ignored original-backed personal bundles and archives under `release/personal` from October 4. They were left untouched under the explicit release-assets restriction; this is not a claim that the entire pre-existing checkout is commercial-data-free. Existing public release/candidate artifacts were also untouched. Initial untracked `.DS_Store` remains outside the commit.

Physical Android/emulator SAF acceptance, native Windows/Linux gameplay execution and independent DOS-source/capture parity have not been established by this task. No new unexplained B consumed semantic difference was found within the audited read map. Changes add only B support; no physics, renderer, UI, game rules or audio scheduling algorithm changes.

## Changed files

- `internal/datalayout/layout.go`, `prepare.go`, `powerpack.go`, `powerpack.json`, `possessed_test.go`, `powerpack_test.go`
- `internal/assets/layout.go`
- `internal/frontend/runtime.go`, `alternate_audit_test.go`, `compatibility_test.go`
- `hosts/android/tests/DataImportPrivateTest.java`, `hosts/macos/native_tests.m`
- `tools/powerpack_layout.py`, `check_public_source.py`, `test_android_import.sh`
- `docs/runtime-compatibility-audit.md`, `runtime-data.md`, `runtime-layout-b-validation.md`

Private reproduction: set `PF_RUNTIME_DATA` to owned A and `PF_POWERPACK_DATA` to owned B; run `tools/go.sh test ./internal/datalayout ./internal/frontend` and `tools/test_android_import.sh` with an installed JDK. Descriptor regeneration requires the private `final-comparison.json` evidence and both owned installations; `tools/powerpack_layout.py --help` documents its arguments. Public CI leaves private variables unset and skips these private tests.
