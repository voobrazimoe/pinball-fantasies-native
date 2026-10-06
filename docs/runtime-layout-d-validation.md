# CD-family layout D — semantics and validation, 2026-10-06

Implemented locally from main `cdf608453e2a740e0304cb4b682db330ec401570`. D0 is closed; D1 registers `dos-deluxe-cd-linked-v1`. This name describes the proved Deluxe CD family, including the exact GOG runtime duplicate. It makes no Rev1/Rev2, Gold Pack, first/second revision or chronology claim. At that implementation boundary C and demos remained unsupported; C is now separately supported as described in [C validation](runtime-layout-c-validation.md). A alone remains the pinned oracle.

## Identity and D0 evidence

The eleven required runtime files were compared again, directly in memory, across the unpacked tree, MODE1 raw `pfd.bin` filesystem (+16 sector payload) and MODE2 GOG `game.gog` filesystem (+24). Both ISO9660 trees were traversed after checking CD001. All eleven files are exact-identical; their supplied runtime fingerprint is `633f1acc51ecc2d6cd313ead2a486bfb4c4ab30b284a22ea16590593d6ee94b8`. Installed-file validation does not depend on any container.

Historical `INTRO.ASM`, `STONES.ASM` and `FANTASIE.ASM` were read from the owner-supplied source directory and checked against `analysis/source-inventory.json`. Research used static 16-bit Capstone decoding only. No DOSBox/emulator was installed, downloaded or run.

### INTRO role and actual geometry

Source `INTRO.ASM::unpkpics` loads `t21st1seg` and `t21st2seg`, shares the upper picture palette through pelle1 and presents the first startup scene. These are the upper/lower griffin/21st Century artwork, replacing A's two 320×256 records. Identity is bound by source consumers and relocated segment operands, not by FORM ordinal order.

| Role | A FORM | D FORM | Actual D geometry | Full-build placement |
| --- | --- | --- | --- | --- |
| t21st1seg, upper griffin | 0x3b810 | 0x3b840 | 320×123×8 | x0,y0 |
| t21st2seg, lower griffin | 0x42c50 | 0x3d820 | 320×117×8 | x0,y123 |

D segment operands at file 0x38b8d and 0x38bc7 resolve exactly to these FORM records; calls target UNPKLBM at 0x3b170. The latter placement passes BX=123. The retained UNPKPICS2 routine still contains y139, but its source caller is inside the demo-only branch; the full build's startup uses UNPKPICS. Native stores the actual dimensions and a placement descriptor. It does not pad, stretch or rewrite a 320×256 header. Pixel tests include the y122/y123 boundary and the last y239 row. The same chunky palette half-bright construction is present in A and D; the generic picture decoder and palette path remain shared.

All seventeen actual D FORM descriptions, full copied FORM extents, bounds and decoding were checked. Fifteen retained pictures are byte-equivalent at reviewed addresses; sidebar/options payload remains equivalent. Two Viking/DI pictures relocate to 0x3fa30 and 0x46ba0. No renderer edition branch was added.

### Stones selector identity

`STONES.ASM::AreaLista_L/U/L_T/U_T` supplies labels for all 44 entries, in four lists. `FANTASIE.ASM::CHECK_AREAS` compares the rectangle, suppresses repeated LASTCHECK and consumes the following selector. Native `stones.decodeAreas` resolves that word to a semantic handler string and never executes it. The four terminators and all rectangle coordinates were checked independently.

D's MZ code base is 0x3f4e0, while A's is 0x300. The explicit raw-selector role map below is checked against source labels and the actual routine code: all 34 rooted graphs preserve instruction/register shapes, conditional branch targets and gameplay constants, with reviewed state/effect/task address operands. The check includes 2,085 instruction visits across rooted graphs (shared subgraphs can be visited more than once), plus all eight source GHOSTAREA.G_ROUTINE records and their callback graphs. A second paired traversal checks 2,631 helper instruction positions, following direct calls. It stops at explicitly identified shared-native boundaries: DOS high-score file persistence (native host storage), IRQ wait, matrix dispatch (shared decoded command model), and the eight proved ghost callbacks. This is a static semantic-identity proof for typed handler selection, not a claim of full DOS execution parity.

The operand at GROPC's indirect call is the source-defined seven-byte GHOST_STRUC.G_ROUTINE field, not an unknown arbitrary pointer. Each of its eight values is checked and each callback graph compared. Gate/capture/collision state and presentation were also compared after invoking every native typed role with A/D decoded input. Restoring even one raw D selector to the prepared address view must fail decoded validation.

| D raw selector | Native role | Registered native token | Root graph instruction visits |
| --- | --- | --- | --- |
| 5940 | PARTYOFFAREA | 5944 | 16 |
| 5988 | CLOSE1 | 5992 | 9 |
| 6015 | BYGEL15 | 6019 | 1 |
| 6016 | BYGEL16 | 6020 | 4 |
| 6026 | OPENBUMPERS | 6030 | 3 |
| 6089 | BYGEL5 | 6093 | 196 |
| 6358 | BYGEL6 | 6362 | 196 |
| 6627 | BYGEL7 | 6631 | 198 |
| 7268 | GROPA | 7272 | 318 |
| 8379 | GROPA_T | 8383 | 19 |
| 8638 | GROPB | 8642 | 123 |
| 9145 | GROPB_T | 9149 | 4 |
| 9157 | BYGEL19 | 9161 | 1 |
| 9158 | BYGEL3 | 9162 | 37 |
| 9249 | BYGEL4 | 9253 | 37 |
| 9340 | BYGEL1 | 9344 | 24 |
| 9407 | BYGEL2 | 9411 | 24 |
| 9474 | GROPC | 9478 | 206 |
| 10738 | CLOSE4 | 10742 | 1 |
| 10739 | BYGEL28 | 10743 | 3 |
| 10746 | BYGELSI | 10750 | 3 |
| 10753 | BYGEL8 | 10757 | 61 |
| 11094 | BYGEL12 | 11098 | 91 |
| 11285 | BYGEL13 | 11289 | 91 |
| 11476 | BYGEL14 | 11480 | 91 |
| 11667 | BYGEL8B | 11671 | 2 |
| 11674 | BYGEL11 | 11678 | 168 |
| 12275 | BYGEL9 | 12279 | 97 |
| 12529 | BYGEL10 | 12533 | 46 |
| 12640 | OPEN6 | 12644 | 3 |
| 12647 | CLOSE6 | 12651 | 3 |
| 12654 | OPEN7 | 12658 | 3 |
| 12661 | CLOSE7 | 12665 | 3 |
| 12668 | BYGEL18 | 12672 | 3 |

The numeric native token is an existing internal registration key for the named role. The translation is `raw -> reviewed role -> native token`, not arithmetic `raw+4`. Every selector word is excluded from byte-copy spans and explicitly checked before emitting its native token. An unregistered raw word fails closed. The raw-pointer representation never reaches gameplay.

Tower control initializes row152, decrements before reading, and reads sixteen forty-byte rows. Maximum reachable row151 gives `(151+16)×40 = 6680` bytes. Those bytes are exact at the reviewed D location. D's source descriptor/copy uses this reachable extent. A/B's existing conservative 6720-byte map is unchanged; decoded bounds still satisfy it. The unconsumed tail does not become a fake artwork difference.

## Architecture and typed fields

The boundary remains DetectInstallation -> one coherent registered profile -> PreparePRGForProfile -> shared decoded address contract -> ValidateDecoded -> common runtime. `deluxe_cd.json` contains 1,921 nonoverlapping copy spans (INTRO16, TABLE1 479, TABLE2 424, TABLE3 473, TABLE4 529), source bounds/geometry/fingerprints, a typed cue constraint, and 44 typed selector records. It contains metadata only. `deluxe_cd_layout.py` deterministically checks every reviewed region against private A/D and rejects conflicting destination overlap; production performs no searches, scanning or approximate matching.

Equivalent artwork, fonts/glyph records, control records, physics lookups, masks, matrix/animation records, PBMs and text are address-translated. Actual S_EMPTY `(62,0,0)` bytes are preserved and consumed by the existing JingleSpec/cue/arbitration path; A remains `(62,0,1)` and B remains `(62,0,0)`. The scheduler algorithm is unchanged. Selectors undergo explicit typed identity conversion. Changed INTRO geometry/placement is expressed in the common asset model. D factory seeds are preserved; they are proved A-equivalent. No canonical source bytes are injected into D.

MOD2 and TABLE1–4 MODs are exact/shared; INTRO.MOD's known unconsumed trailer leaves the decoded model unchanged. All six role-specific shared Module decoders were DeepEqual-checked against A. There is no D audio decoder, engine, UI, gameplay-rule or physics-algorithm branch.

## Original D coherence and C rejection evidence

A/B/D detect their respective profile. Tests cover all thirty proper substitutions for each A/D and B/D pair (sixty hybrids), including every PRG role, both INTRO directions, isolated TABLE1/TABLE4 and dangerous combinations. Malformed, unknown and incomplete installs reject. Existing A/B hybrid coverage remains.

C was read from the private CD tree and its supplied fingerprint `6b680be66b3c53286db0e658d48a5c5475e1fb97f67d06374bb1b743a7e5fe33` verified. C as a whole, every individual C PRG under D, and every single-role C substitution into D reject. D source high-score records have explicit fingerprints: C differs in the twelve initials bytes per table, so shared TABLE addresses cannot admit it. Only D is registered. Future C work can reuse the address compiler with separate reviewed seed semantics, but no C allowance exists now.

## Test/build matrix

| Check | Result and boundary |
| --- | --- |
| D0 static source/code check | PASS: 34 roles, 44 records, four terminators, eight ghost callbacks, INTRO role/placement, helper graph |
| Descriptor reproducibility/coverage | PASS, exact regeneration; overlap rejection; all reviewed spans; all actual FORM descriptors |
| Private A/B/D runtime | PASS: frontend, four independent installation-backed factories, graphics/physics/control models, cue path/arbitration, six Module models |
| Native Stones typed roles | PASS: all 34 roles, all 44 entries, gate/capture/collision/presentation equivalence, naive/unknown selector rejection |
| Coherence/C boundary | PASS: sixty A/D and B/D hybrids, existing A/B hybrids, C full/role/mix negatives |
| Shared macOS Go suite | PASS: 18 tested packages, five without tests; 241 passing cases/subcases, 826 external-input skips |
| Canonical A engine regression | PASS: original-backed complete engine suite, 263 seconds |
| Extended A/source suite in disposable external tree | 2,109 PASS, 23 SKIP; 31 failing cases/subcases: 27 from missing reference inputs, four existing baseline frontend failures; see below |
| Python tools | PASS: 40 unittest cases, one unrelated private-input skip; ABI CLI scripts checked separately |
| Desktop/shared validation and launch | PASS through real personalvalidate/Load and C ABI; installed D scores save/restart/reset; native Windows/Linux GUI execution not available on this Mac |
| macOS source/staging/import/adoption/relaunch | PASS for A/B/D; 32 native host journeys per installation (four tables × four scroll modes × two resolutions) |
| Android provider-copy/staging/validation/adoption | PASS: actual DataImport plus shared native validator; D adoption/bootstrap/relaunch, sixty hybrids, C/malformed rejection, rollback and interrupted recovery/cleanup |
| Android SAF/device UI | NOT RUN: provider streams exercise the transaction, not physical-device SAF instrumentation |
| Android Java/C++ host A3/A4/A5 | PASS, including semantic UI, keyboard, frame, presentation, ring/Oboe/focus/recovery tests |
| C ABI versus direct Go Runner | PASS for A/B/D: eight replays per installation, 2,714 checkpoints each; frame/state/PCM parity, not DOS or cross-edition parity |
| macOS public arm64 app/C archive/shared library | PASS; bundle allowlist, ABI contract, codesign, native tests |
| Windows amd64 GUI EXE | BUILD PASS; execution not available here |
| Linux amd64 trace executable | BUILD PASS; native SDL GUI toolchain/execution not available here |
| Android arm64-v8a/x86_64 engines and asset-free debug APK | BUILD PASS; each ELF LOAD alignment is 16 KiB; offline Gradle used existing dependencies |
| Payload/source checks | PASS: metadata schema, changed source extension scan, public bundle/APK allowlists, original-block scans; see cleanliness below |
| Oracle | A inventory/verifier/captures untouched; pinned input checks and canonical engine regression PASS |

The extended suite's missing reference inputs are NOT AVAILABLE, not application failures: pf2-fixture.json, pf2-ball-locations.json, pf6-art-fixtures.json, pf7-assets.json, pf9-assets.json, pf10-assets.json, pf11.2-full-table-fixtures.json and pf8-bios-font.json. Some existing tests fail instead of skipping when these secondary inputs are absent. No fixture was regenerated or oracle weakened to conceal this.

Four real baseline failures are TestNativeGameshowLifecycleAndPersistence, TestNativeSpeedDevilsLifecycleAndPersistence and TestNativeStonesLifecycleAndPersistence (`native content`), plus TestPF6SessionKeepsGameplayOracle (score000002311040, ball2, ticks1200). All four were rerun against a clean `git archive` of cdf6084 with the same A/source inputs and produced identical failures/messages. They remain unresolved pre-existing assertions outside this layout phase. They are not labelled PASS. The new D and A/B regression tests pass.

The first broad Python discovery attempt used system Python without PIL and imported an ABI CLI script without its required argument. The final unittest run used the bundled Python libraries; the ABI script ran separately with the built shared library. Gradle's sandbox file-lock socket was blocked; the authorized offline build succeeded with automatic sandbox approval. No new network dependency was fetched.

## Reproduction

Use owner-provided private fixtures; unset variables skip private tests in public CI:

- PF_RUNTIME_DATA: A
- PF_POWERPACK_DATA: B
- PF_DELUXE_CD_DATA: D unpacked installation (neutral CD family, no chronology)
- PF_DELUXE_CD_ALT_DATA: C, now a supported independent profile (renamed from the former PF_UNSUPPORTED_CD_DATA rejection fixture)
- PF_LAYOUT_AUDIT_EVIDENCE: private final-comparison.json, for descriptor reproducibility

Run `tools/go.sh test ./internal/datalayout ./internal/frontend ./internal/stones`, the supported shared package suite, `sh tools/test_android_import.sh`, and `tools/deluxe_cd_layout.py --evidence ... --canonical ... --data ... --check`. The static proof tool is `tools/deluxe_cd_semantics.py --canonical ... --data ... --source ...`; Capstone is a research-only dependency. It emits only role/address/count metadata and never executes code or exports payload.

## Cleanliness, unchanged scope and files

No commercial PRG/MOD/DOS EXE/CFG/container, reconstructed payload or copied private report was added to tracked or new source files. New descriptor and research tools contain addresses, sizes, hashes, geometry, labels and tests only. Temporary fixture staging/reference trees were outside the checkout. Public macOS/Windows/Linux artifacts and APK were scanned against 1,743 distinct nontrivial 4 KiB blocks from A/B/D, and against required-file payload names. No matches were found. APK DebugProbesKt.bin is the Kotlin dependency metadata, not DOS/CD data. The existing ignored original-backed `release/personal` artifacts predate this task and were left untouched; this is not a claim that every pre-existing ignored workspace artifact is asset-free. The initial untracked .DS_Store is excluded from the commit.

Versions, tags and release assets were untouched. v0.1.3 remains c417afe073505531b6dcafeb45c36ca0d4630668. No push, tag or publication. Remaining limits: missing captures/reference fixtures, the four baseline assertions above, no independent DOS parity or physical Android acceptance, no native Windows/Linux GUI execution on this host. No unresolved D selector or INTRO-role blocker remains.

Changed files:

- internal/datalayout: prepare.go, powerpack.go (shared bounded copy helper), deluxe_cd.go, deluxe_cd.json, deluxe_cd_test.go
- internal/assets: frontend.go, layout.go
- internal/frontend: render.go (placement data only), deluxe_cd_test.go
- internal/stones/deluxe_cd_test.go
- hosts/android/tests/DataImportPrivateTest.java
- tools: deluxe_cd_layout.py, deluxe_cd_semantics.py, test_deluxe_cd_layout.py, check_public_source.py
- docs: runtime-data.md, runtime-compatibility-audit.md, runtime-layout-d-validation.md

The local implementation commit is recorded in the final response; embedding its own SHA here would change that commit.

## C support follow-up

C registration supersedes the original full-C rejection assertion above. The C/D identity boundary remains strict: all 1,024 A/B/C/D PRG combinations are now tested, with C/D mixed seeds rejected. `deluxe_cd.json` and its D metadata are unchanged; C reuses its TABLE spans, cue constraint and 44 typed selectors through a separate INTRO/seed overlay. D geometry, frontend/factories, Modules, typed roles, scores and import/adoption regressions were repeated; see the current [C report](runtime-layout-c-validation.md). A remains the sole oracle.
