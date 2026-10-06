# Deluxe CD alternate runtime layout C — validation, 2026-10-06

Implemented from `main` `d67859c45225e4ce4778174bc2ddd746f8e16ad5`. C is supported as `dos-deluxe-cd-alt-linked-v1`, alongside A/B/D. “alt” distinguishes an independently coherent Deluxe CD-family runtime file-set; it does not claim Rev1/Rev2, first/second edition, Gold Pack or release chronology. A alone remains the pinned canonical oracle. Demos and unidentified layouts remain unsupported.

## Identity and private evidence

The eleven-file fingerprint is `6b680be66b3c53286db0e658d48a5c5475e1fb97f67d06374bb1b743a7e5fe33`. It was recalculated from the existing private fixture in the named `Pinball Fantasies Deluxe CD (1995)(21st Century Entertainment)` directory and independently from the directory named `(Rev1)`; both match. The prior private audit also records this fingerprint for the reconstructed multi-volume ARJ installation. That temporary reconstruction is no longer available and was not reconstructed during this task. Directory/archive names are representation labels, not evidence of official chronology.

Research read the private `REPORT.ru.md`, `final-comparison.json`, `installations.json`, historical `INTRO.ASM`, and actual C/D PRGs/MODs. No payload or private report was copied into the repository. No DOS program, DOSBox or installer was run or downloaded.

## Exact INTRO mapping

The descriptor stores exact addresses, not approximate deltas. These are the common native presentation roles in decoder order:

| Common role | Decoder address | C source FORM | Actual geometry |
| --- | --- | --- | --- |
| Logo | 0x1cb10 | 0x1c920 | 640×256×4 |
| Font | 0x6b70 | 0x6980 | 640×28×4 |
| MonoFont | 0x8b00 | 0x8910 | 640×29×4 |
| Tables[0] | 0x20430 | 0x20240 | 640×95×4 |
| Tables[1] | 0x24a50 | 0x24860 | 640×95×4 |
| Tables[2] | 0x29410 | 0x29220 | 640×200×4 |
| Tables[3] | 0x2dd40 | 0x2db50 | 640×200×4 |
| Startup[0], t21st1seg | 0x3b810 | 0x3b960 | 320×123×8 |
| Startup[1], t21st2seg | 0x42c50 | 0x3d940 | 320×117×8 |
| Startup[2] | 0x46930 | 0x3fb50 | 320×130×8 |
| Startup[3] | 0x4daa0 | 0x46cc0 | 320×126×8 |
| Startup[4] | 0x12020 | 0x11e30 | 320×110×8 |
| Startup[5] | 0x17f20 | 0x17d30 | 320×130×8 |
| Startup[6] | 0x10e60 | 0x10c70 | 320×200×8 |
| Startup[7] | 0xa450 | 0xa260 | 640×178×4 |
| HighLogo | 0x34d90 | 0x34ba0 | 640×200×4 |
| HighMono | 0x33410 | 0x33220 | 640×200×4 |

Sidebar/options source is `[233978,234218)`, mapped to `[233806,234046)`: 120 bytes per role. Its payload is exact D-equivalent. Early retained forms cross-check at A−496; late retained forms cross-check at A−28128; sidebar at A+172. These deltas are research checks only.

`deluxe_cd_alt_semantics.py` reuses D's established source-role binding. Actual C segment operands at `0x38caa` and `0x38ce4` resolve to `0x3b960` and `0x3d940`, respectively. The corresponding full-build callsite instruction/register shapes agree with the reviewed D callsites. C calls UNPKLBM at `0x3b290`, with BX=0 and BX=123. C/D code locations differ, but the source roles and placement remain the same. The historical INTRO source is inventory-checked. D0's Stones handler proof is not repeated from scratch.

All seventeen entire FORM extents are checked against their corresponding D payloads, including both changed CD-family images, bounds and geometry. Prepared C INTRO is byte-identical to prepared D INTRO. Native tests preserve upper row122, lower start row123 and final row239, the common palette/half-bright path, and sidebar/options/table-card models. No padding, stretching, image payload injection or renderer profile branch is used.

## TABLE reuse and factory seed plumbing

`deluxe_cd.json` is unchanged. Its 1,905 TABLE copy spans, reviewed source regions, pictures, S_EMPTY record and 44 typed Stones selectors are reused directly. `deluxe_cd_alt.json` contains only the C INTRO descriptor and four small typed factory seed records. Go extracts the existing D translation into one helper; C/D use the same bounded address compiler and selector translation. There is no second engine, table map, Stones selector table, or gameplay/render/audio profile branch.

The compiler reads actual private C/D files and proves that each entire TABLE differs at exactly twelve positions, exclusively all three initials bytes of all four factory entries. All remaining bytes are exact, not merely similar. The following typed seed names were read from those bytes:

| TABLE | Source record | Decoder record | Four C factory initials |
| --- | --- | --- | --- |
| TABLE1 | 0xdd93 | 0x19d56 | CLS / RLZ / CLS / RLZ |
| TABLE2 | 0xd743 | 0x18ef6 | CLS / RLZ / CLS / RLZ |
| TABLE3 | 0xd923 | 0x18b76 | CLS / RLZ / CLS / RLZ |
| TABLE4 | 0xa013 | 0x16776 | CLS / RLZ / CLS / RLZ |

Each record is four sixteen-byte entries. C source validation checks the exact sixty-four-byte semantic fingerprint and each typed three-character name. D retains its original four seed fingerprints. Preparation copies each actual input seed unchanged; no A/D seed is inserted into C. Tests compare all copied spans and prove prepared TABLE C/D differ only at these twelve initials positions. They reject every individual swapped character in both directions, every entry substitution, and all fourteen proper subsets of four seed entries in both directions.

The existing native `Defaults(table)` names were hardcoded from canonical INTRO factory records. `LoadConfigured` now reads validated TABLE initials into the typed `Scores` model, before constructing the frontend. The table presentation also retains its actual decoded HI_SCORE_LIST. FileStore uses installation-bound defaults when writable and optional seed files are missing. Saved `.HI` state still wins, then explicitly supplied optional `.HI` seeds, then installation defaults. Save uses the existing atomic persistence logic. Relaunch binds the defaults from the current installation again; clearing the writable record restores those defaults. Bare `New`/FileStore APIs retain their previous defaults for callers without installation data. Custom Store implementations retain control of their own Load policy.

Existing numeric native factory values are preserved for A/B/C/D. In particular, canonical TABLE3's attract HI_SCORE_LIST is 50M/25M/10M/5M, whereas the existing native INTRO-derived factory defaults are 100M/50M/25M/10M. This pre-existing distinction is also present in D and C; changing numeric defaults was outside the requested initials change and would alter A/B/D behavior. Actual decoded TABLE score digits remain intact. All A/B/D native default scores are verified unchanged; C receives its actual factory initials on the same numeric defaults.

State storage intentionally remains one common namespace. Existing scores survive switching editions; missing records use that installation's defaults. Tests check C save/restart/clear, optional seed precedence, C-to-D saved-state precedence and absence of global/default leakage. No profile-specific state directory or format was introduced.

## Cue and Stones invariants

C TABLE1's actual source record at `0xe9e9` is S_EMPTY `(62,0,0)`. The compiler reads this from the private input. Copying preserves it at `0x1a9ac`, and the decoded JingleSpec, Party Land Cue and MusicClock priority/arbitration tests pass. A remains `(62,0,1)`; B/C/D remain `(62,0,0)`. Scheduler code and algorithm are unchanged.

C TABLE4 source selector words are exact D representation. The single reviewed `raw selector -> semantic role -> native token` map covers 44 area records and 34 roles, with four terminators. Unknown selectors fail closed, and reverting any prepared selector to raw DOS representation fails validation. No selector+4 arithmetic is introduced. All typed native handlers are exercised against A/D and A/C; rectangles, collision/capture/gate state, control and presentation agree. D's proved ghost callback/source semantics remain the shared invariant. The copied tower extent is the same reviewed 6,680 bytes. Full C/D file equality outside factory initials makes another handler-source proof unnecessary.

## Detection and imports

DetectInstallation validates all five PRGs under one registered profile. Full A/B/C/D installations detect the correct distinct IDs. The generated matrix evaluates all 4^5=1,024 combinations. Ten are byte-coherent installations (A/B TABLE3/4 are identical); all 1,014 incoherent combinations reject. This covers each individual role, INTRO crosses, pairs of mixed TABLE families, all proper C/D subsets and INTRO-versus-all-TABLE dangerous cases. Malformed, incomplete and unknown fixtures also reject. The whole-file fingerprint is a research identity check, not the production acceptance algorithm.

C uses the existing eleven-file import flow with optional CFG. Real desktop/shared validation and Load pass. The macOS host validates source and staging, adopts transactionally, preserves prior data on rejection, relaunches and runs all native journeys. Actual Android DataImport provider streams copy/stage, call the shared validator, adopt/bootstrap/relaunch, roll back injected failures and recover interrupted transactions. Android tests reject sixty C/D mixed installs, sixty A/B/C mixed installs and the existing A/B/D matrix; C-specific UI is not added. Physical Android SAF/device instrumentation was NOT RUN.

## Regression/build matrix

| Check | Result and evidence boundary |
| --- | --- |
| C identity | PASS: two available named private copies recalculated; ARJ copy is prior audit evidence, current temporary fixture NOT AVAILABLE |
| C/D descriptor reproducibility | PASS: deterministic compiler/checks, bounded FORM payload comparison, exact TABLE twelve-byte difference proof, shared mapping coverage |
| C INTRO source roles | PASS: static source/callsite operands, actual geometry, placement and common decoded presentation |
| C/D seeds | PASS: every entry/character, partial seed hybrids, fresh/nil-store models, save/restart/clear, optional seeds and common-state precedence |
| A/B/C/D detection | PASS: 1,024 combinations; 1,014 hybrid rejections; malformed/incomplete/unknown rejects |
| A/B/C/D frontend/factories | PASS: installation-only staging, all four factories per installation, no canonical payload dependency |
| Graphics/physics/control | PASS: all four prepared C/D table models against A; twelve seed bytes remain installation-specific |
| MOD | PASS: all six role-specific decoded C/D Module models DeepEqual against A; no new audio decoder |
| Stones/cue/audio/persistence | PASS: shared CD selector invariants and every native role; S_EMPTY decode/Cue/arbitration; actual C initials persist |
| Supported shared Go suite | PASS: eighteen tested packages, five without tests; 289 passing cases/subcases, 826 external-input skips |
| Extended A/source suite outside checkout | 2,158 PASS, 22 SKIP; 31 failing cases/subcases = 27 absent-reference cases/subcases and four reproduced baseline assertions |
| Canonical A engine suite | PASS: original-backed complete engine regression, 267.191 seconds |
| Public fixture-free private tests | Clean SKIP for missing A/B/C/D inputs; no private fixture dependency in public CI |
| Python tools | PASS: 41 unittest cases, one unrelated private-input skip; five layout compiler cases include A/D and C/D reproducibility |
| Android import | PASS: real provider/DataImport/shared validator, transactions/recovery and hybrid rejection; physical device UI NOT RUN |
| Android Java/C++ host A3/A4/A5 | PASS: semantic UI/keyboard, native lifecycle/frame, presentation, PCM ring/Oboe, focus/route policy and recovery |
| macOS host | PASS: public app/archive/shared-library, source/staging/adoption/relaunch, 32 native journeys per A/B/C/D installation |
| C ABI versus direct Go Runner | PASS for A/B/C/D: eight replays per installation, 2,714 checkpoints each, PCM/frame/state parity; not DOS or cross-edition parity |
| Windows amd64 GUI | BUILD PASS; execution NOT RUN on this Mac |
| Linux amd64 trace | BUILD PASS; SDL GUI build/native GUI execution NOT AVAILABLE on this Mac |
| Android arm64-v8a/x86_64 | BUILD PASS; both shared engines and smoke binaries have 16 KiB ELF LOAD alignment |
| Asset-free debug APK | BUILD PASS with cached offline Gradle dependencies; exact APK contents/payload scan checked |
| Public source/payload checks | PASS: descriptor metadata schemas, changed-file extensions, nontrivial original-block scan of changed sources and public outputs |
| Oracle | A inventory, allowlist, verifier, reference captures and fixtures unchanged; canonical engine regression passes |

The four baseline assertions are `TestNativeGameshowLifecycleAndPersistence`, `TestNativeSpeedDevilsLifecycleAndPersistence`, `TestNativeStonesLifecycleAndPersistence` (`native content`) and `TestPF6SessionKeepsGameplayOracle` (`score=000002311040 ball=2 ticks=1200`). They were repeated on a clean external `git archive` of the task's actual base `d67859c`, with identical messages to the patched extended suite. They are not labelled PASS or hidden.

Missing reference inputs are `pf2-fixture.json`, `pf2-ball-locations.json`, `pf6-art-fixtures.json`, `pf7-assets.json`, `pf9-assets.json`, `pf10-assets.json`, `pf11.2-full-table-fixtures.json`, and `pf8-bios-font.json`. Existing tests sometimes fail instead of skipping when these inputs are missing. These are NOT AVAILABLE, not application acceptance. No canonical fixture was regenerated to make C pass.

The first Android import attempt lacked JAVA_HOME; it was rerun with the existing private JDK and passed. Gradle's local file-lock socket required sandbox escalation; the authorized offline build passed. No dependency download was needed.

## Reproduction and cleanliness

Private variables:

- `PF_RUNTIME_DATA`: canonical A
- `PF_POWERPACK_DATA`: supported B
- `PF_DELUXE_CD_ALT_DATA`: existing private C fixture (replaces the former test-only `PF_UNSUPPORTED_CD_DATA` name)
- `PF_DELUXE_CD_DATA`: supported D
- `PF_LAYOUT_AUDIT_EVIDENCE`: private final-comparison metadata
- `PF_ENGINE_DATA_DIR`: installation for original-backed engine tests

Run `tools/go.sh test ./internal/datalayout ./internal/frontend ./internal/stones`, the supported shared suite, `tools/test_android_import.sh`, and the existing host/ABI tools. Check C metadata with `python3 tools/deluxe_cd_alt_layout.py --evidence "$PF_LAYOUT_AUDIT_EVIDENCE" --data "$PF_DELUXE_CD_ALT_DATA" --deluxe "$PF_DELUXE_CD_DATA" --check`. The static INTRO tool additionally accepts `--source` pointing to private historical source and requires the already available research-only Capstone dependency. Production uses neither tool.

Changed files contain no PRG/MOD, DOS EXE/COM/CFG, ISO/BIN/GOG/ARJ, extracted images, executable chunks, PCM, or copied private evidence report. Metadata contains only offsets, lengths, fingerprints, geometry, roles and typed seed labels. Public macOS bundle, Windows/Linux binaries, Android engines/APK and changed source were checked against nontrivial 4 KiB blocks from A/B/C/D. Existing ignored personal/release originals were left untouched; `.DS_Store` is excluded.

No release/version, tag, push or publication changes. No demo support, DOSBox, gamepad/sensors, gameplay/physics/scheduler changes, edition renderer branch, Gold Pack identity or historical revision claim. Remaining limits are the baseline assertions, missing reference inputs, unavailable temporary ARJ reconstruction, physical Android UI and native Windows/Linux GUI execution. No unresolved C semantic or identity blocker remains. The local commit SHA is provided in the final task response.
