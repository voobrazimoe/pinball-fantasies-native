# Shared runtime compatibility audit

Runtime compatibility and exact reference identity are separate contracts. This audit covers Windows, Linux and macOS; a future Android host should call the same engine/shared loader.

## Before and after, by input

Every PRG previously required the SHA-256 of the entire inventoried executable before decoding fixed offsets. Every MOD also required whole-file SHA-256; frontend hashes additionally selected INTRO's omitted-sample behavior. Those checks protected offset assumptions indirectly, but also rejected changes to executable headers, unused instructions, module titles and PCM bytes that the native decoder does not need to identify its layout.

| File | Previous runtime whole-file SHA | Runtime compatibility now | Actually consumed content |
| --- | --- | --- | --- |
| `INTRO.PRG` | Yes, pinned INTRO build | `dos-retail-linked-v1`: seventeen IFF anchors and exact expected geometry; bounded sidebar/options records | FORM/chunk headers, BMHD, CMAP and BODY at the seventeen addresses below; text `[233806,234046)` |
| `INTRO.MOD` | Yes; its hash selected omitted sample slots | Four-channel 31-slot `M.K.`; at least 44 orders; structurally bounded patterns/samples; verified omitted-silent-slot variant or full sample storage | Sample headers `[20,950)`, length byte 950, active orders from 952, marker `[1080,1084)`, pattern cells from 1084, declared playable sample bytes; title/restart byte and checksum do not select compatibility |
| `MOD2.MOD` | Yes, pinned menu module | Same tracker structure; at least 15 orders, full sample storage | Same tracker records; no INTRO omission rule |
| `TABLE1.PRG` | Yes, pinned Party Land build | Canonical linked DOS profile; structural PBMs plus bounded artwork and regional control/physics fingerprints | Four playfield strips, foreground, ball literals, physics lookups/descriptors/masks/frames, matrix fonts/text/animation records, lamps and duck gates |
| `TABLE1.MOD` | Yes | Same tracker structure; at least 64 orders; required effect slots 7, 22–25, 28–30 | Same module records; native effect sample numbering is preserved; PCM identity is not required |
| `TABLE2.PRG` | Yes, pinned Speed Devils build | Canonical linked DOS profile; structural PBMs plus bounded artwork and regional control/physics fingerprints | Same categories as table 1, with Speed Devils addresses and lamp packet references |
| `TABLE2.MOD` | Yes | Same tracker structure; at least 65 orders; required nonempty effect slots 23–30 | Same module records; source-referenced effect slot 22 is deliberately empty in the oracle and remains safely silent |
| `TABLE3.PRG` | Yes, pinned Gameshow build | Canonical linked DOS profile; structural PBMs plus bounded artwork and regional control/physics fingerprints | Same common categories, Gameshow lamp packets and strided gate masks |
| `TABLE3.MOD` | Yes | Same tracker structure; decode at least 63 source-referenced orders, including orders 61/62 beyond the legacy song-length byte 61; effect slots 7, 15, 22–25, 28–30 | All 64 stored patterns selected by the verified 63-order extent, then samples; no invented edition offsets |
| `TABLE4.PRG` | Yes, pinned Stones build | Canonical linked DOS profile; structural PBMs including the final 320×1219 record; regional control/physics checks | Common categories, strided gates, area rectangles/handler selectors/terminators, packed tower artwork |
| `TABLE4.MOD` | Yes | Same tracker structure; at least 66 orders; effect slots 2, 6, 10, 23–25, 28–30 | Same tracker records, including all 64 selected patterns |
| `PINBALL.CFG` | No in shared settings loader; **yes in the personal builder**, and wrongly documented as required game payload | Optional settings. Valid six-byte DOS seed or native PFNC record; missing/malformed records use shared native defaults. No identity hash | DOS six bytes, or eleven-byte versioned PFNC header/payload. Writable native state takes precedence over installation seed |

Order extents are minimum source-referenced extents, not claims that arbitrary songs provide identical game music. Expected modules may contain different PCM bytes and remain safe to play. Native source-derived cue scheduling is unchanged. Exact PCM equality still requires exact oracle modules.

## PRG read map and safety

The exhaustive, address-only machine-readable read map is [profiles.json](../internal/datalayout/profiles.json). Each entry gives an offset, extent, purpose, optional structural kind and optional consumed-region SHA-256. It contains no original payload bytes. [runtime_layout.py](../tools/runtime_layout.py) documents the source readers behind each category and regenerates metadata **only when explicitly invoked with exact pinned originals**. Tests never regenerate expected hashes.

All offsets use half-open byte extents. The native runtime does not use an MZ header, relocation table or executable entry point. Requiring those would add no protection to its data reads. Byte zero is provably outside every profile region and IFF anchor; the regression mutates that byte in all five PRGs. Whole-file oracle verification rejects the same mutations.

INTRO IFF anchors (width×height and bit planes are pinned separately in the profile):

```text
0x1cb10 0x6b70 0x8b00 0x20430 0x24a50 0x29410 0x2dd40
0x3b810 0x42c50 0x46930 0x4daa0 0x12020 0x17f20 0x10e60
0xa450 0x34d90 0x33410
```

The linked table profile reads these major regions; smaller literal/control records are individually enumerated in the read map rather than hashing surrounding executable bytes:

| Region | TABLE1 | TABLE2 | TABLE3 | TABLE4 |
| --- | --- | --- | --- | --- |
| Four PBM anchors | 0x52430, 0x59660, 0x619a0, 0x6abb0 | 0x50730, 0x583f0, 0x60030, 0x67b00 | 0x4cb60, 0x52410, 0x5a6b0, 0x634d0 | 0x4bc10, 0x54a00, 0x5da70, 0x66e20 |
| Lower/upper foreground (23040 bytes each) | 0x300b0 / 0x35f30 | 0x2d4f0 / 0x33370 | 0x1f870 / 0x256f0 | 0x28f30 / 0x2edb0 |
| Delta lookup start/length | 0xc2e0 / 55888 | 0xbad0 / 54288 | 0xb570 / 54768 | 0xca80 / 40016 |
| Sine lookup (5120 bytes) | 0x1e340 | 0x1d570 | 0x1ca40 | 0x1b2c0 |
| Eight material records (ten consumed bytes per 16-byte stride) | 0x1c05d | 0x1b00b | 0x1a93d | 0x18e77 |
| Flipper descriptors (60-byte stride; only consumed fields) | 0x20690 | 0x1f820 | 0x1f230 | 0x1da30 |
| Flipper collision frames | 0x4c730, 0x4fe10, 0x4ed50 | 0x49b70, 0x4e110, 0x4c190 | 0x3bef0, 0x3f4a0, 0x3e510 | 0x455b0, 0x47bd0; third flipper absent |
| Flipper graphics start/length | 0x82930 / 1748 | 0x7e5c0 / 168 | 0x7b1f0 / 220 | 0x7f670 / 120 |
| Collision masks (six starts) | 0x41330, 0x3b930, 0x46d30, 0x71b30, 0x77530, 0x7cf30 | 0x3e770, 0x38d70, 0x44170, 0x6d7c0, 0x731c0, 0x78bc0 | 0x30af0, 0x2b0f0, 0x364f0, 0x6a3f0, 0x6fdf0, 0x757f0 | 0x3a1b0, 0x347b0, 0x3fbb0, 0x6e870, 0x74270, 0x79c70 |

Masks are 23040 bytes each except TABLE1/2's last two masks, which consume 20400 bytes each. Flipper frame lengths come from the validated descriptor stride and frame count. The native decoder does not infer addresses from untrusted executable instructions.

Sufficient structural checks for graphics are FORM/PBM or FORM/ILBM at each required anchor, bounded chunk extents with padding, a single BMHD/CMAP/BODY, expected dimensions/bit depth, supported masking and ByteRun1 compression, exact row expansion with no trailing compressed bytes, palette extent and pixel indices. All table strips are 320×144×8, except Stones' final strip, which must decode the entire 320×1219 BODY even though the playfield uses its first 144 rows.

Opaque foreground/flipper/spring/tower artwork and ordinary bitmap fonts need bounds, not byte identity. Matrix bitmap counts and animation controls pin the record boundaries and source scheduler assumptions; bitmap pixels remain variable. Glyph pointer records pin the known address layout, while the existing literal drawing-store records are validated for bounded four-byte stores and a terminator. This is a restricted data representation, not an x86 interpreter. Scroll text can vary within the supported glyph character map; unsupported characters are rejected before they can select unknown addresses. Lamp packet headers validate their source-derived extent/index assumptions; RGB payloads remain variable.

Regional fingerprints remain for collision masks and gate masks, sine/delta/material tables, consumed flipper descriptor fields/frames, animation controls and bitmap length words, lamp scheduling, source-coupled jingle/sample/note controls, glyph pointer tables and Stones area records. These values couple the user's data to hard-coded native table geometry, handler identities and source-derived control behavior. Replacing them with arbitrary bounded bytes would weaken gameplay correctness. Unrelated DOS executable bytes are not included.

## MOD safety

The decoder checks the header before accessing any fields, requires a supported four-channel `M.K.` marker and 31 sample slots, bounds order counts/references and pattern extents, rejects sample numbers above 31 or referenced absent samples, and checks sample data lengths, volumes (0–64), finetunes (0–15), and nontrivial loop extents. Bxx jumps must select an existing order and Dxx breaks a valid BCD row 0–63. The current tracker implements Fxx as speed: F00 and BPM-style F20–FF are unsupported layouts. It does not claim complete support for every ProTracker dialect.

INTRO omission is selected by structure and its file role, not SHA: all eleven omitted slots must be unused, zero-volume two-byte silent slots with no nontrivial loops; their declared storage is omitted and followed by the known two-byte checksum extent. A full sample-storage variant is also accepted. Normal modules' sample extent must agree with the patterns selected by their active orders; an optional two-byte trailer is harmless. Extra hidden patterns or shifted sample storage need a separately validated profile.

## Shared boundary and extension

`frontend.LoadConfigured` reads every PRG through `datalayout.PreparePRG`, then the common asset/physics/presentation decoders. MOD roles use explicit shared decoders. Settings always use `settings.Store.Load`. Linux and Windows call this shared frontend loader; the macOS C bridge uses `engine.Load`, which calls the same loader.

macOS validates the source installation before copying, validates its isolated staging directory again to catch changes during copying, and publishes the imported directory only after both shared compatibility checks pass. Personal Windows/Linux packaging also uses the shared runtime validator; it records current input hashes for reproducibility rather than requiring pinned oracle identity. CFG is an optional packaged seed, never required pristine commercial payload.

`runtimeProfile` separates source-layout validation and canonicalization from the canonical engine data contract. Another possessed and validated layout can add an adapter at this shared boundary, with reviewed offsets and control/pointer translations. It can reuse the game engine. No unpossessed edition offsets, heuristic executable scanning or additional edition claims have been added.

Only one DOS linked layout has been validated here. Multiple legitimate DOS revisions can share it **if their consumed artwork/record geometry and native-coupled semantics match**; different unused code, titles and audio bytes alone do not require a new engine. No second DOS revision was available, so cross-revision acceptance is conditional, not experimentally established. No named legitimate DOS edition has been newly declared unsupported without evidence. Moved records, changed native-coupled controls, other module layouts and Amiga/CD32 inputs are outside this profile and return an unsupported-layout error.

## Exact oracle contract

`internal/oracle.Verify` retains the eleven pinned PRG/MOD sizes and whole-file hashes from the existing inventory. `internal/testinputs.Require` uses it for original-backed tests. Inventory/research tooling, exact PCM/image fixtures and deliberate original-payload leakage scans remain strict. The macOS original-backed acceptance script still checks exact immutable PRG/MOD inventory; mutable CFG is excluded. No deterministic oracle hash or existing expected parity fingerprint was updated.

## Validation on 2026-10-04

Executed results: the shared macOS-supported Go package suite passed (16 packages; 176 passing cases/subcases, 789 skips for absent private source/reference captures or root fixture files). Separate original-backed checks passed 56 cases/subcases: strict inventory and Party Land image checks, tracker/PCM checks, and all four tables' two presentation-record fingerprint suites. The original-backed compatibility tests also ran via `PF_RUNTIME_DATA`, and engine replay ran via `PF_ENGINE_DATA_DIR`.

The native Apple Silicon app/archive/shared-library build passed, including the real importer compatibility regression, 32 host journeys (four tables × four scrolling modes × two resolutions), and eight C ABI PCM/frame/state conformance runs. The public bundle scan found none of 836 nontrivial original 4 KiB blocks. Windows amd64 public EXE and host/shared test binaries cross-compiled successfully. Linux amd64 shared test binaries cross-compiled successfully. Native Windows/Linux execution and a native Linux SDL build were not available on this Mac; no native execution result is claimed. An unfiltered macOS `go test ./...` fails in the existing Windows/Linux-only `internal/platform` host package; the supported shared package suite excludes that package and its CLI host, while the Apple host is tested with `tools/build_macos.sh`.

No full historical source/DOSBox acceptance result is claimed: those private inputs/captures are absent. Existing oracle expectations were preserved, and all available selected oracle checks passed. Regression coverage includes optional/modified/malformed/PFNC settings, all four original tables, unused executable-header mutation, opaque artwork and audio variation, strict-oracle rejection of identity changes, broken FORM/geometry, truncated required regions, incompatible descriptors, tracker bounds/loops/references, and the Windows/Linux/macOS call chain. Native platform execution and source/DOSBox fixture availability must be reported separately from cross-compilation and shared tests.
