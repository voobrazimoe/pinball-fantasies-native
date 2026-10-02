# Historical source and original data provenance

- VERIFIED: upstream repository: https://github.com/historicalsource/pinballfantasies
- VERIFIED: exact checked-out commit: `aa2dd368d73886bbd666507bd7341001060700d9`.
- VERIFIED: reference downloaded with `git clone` on **2026-09-30**.
- VERIFIED: reference location: `reference/original-dos-source/`; shipping Go never imports or executes this material.
- VERIFIED: `source-inventory.json` lists every tracked source filename, size and SHA-256; 11 ASM/MAC files plus README.md. Git metadata is not content inventory.
- VERIFIED: no historical source contents were edited. Source files are made read-only after inspection. Use labels or byte-preserving Latin-1 decoding for analysis; source has CRLF and non-UTF-8 comments.
- VERIFIED: no license file is present in this pinned tree; README.md is only a project title. Source contains historical copyright notices, including START.ASM's 1993–94 21st Century Entertainment notice.
- UNKNOWN: permissions to modify, distribute, or ship derived implementations and original artwork/music. Public availability is not an explicit software license. No license has been assumed or added for the historical source or supplied installation.
- VERIFIED: the original installation was supplied by the user; its external provenance is not established. No supplied executable, patch, sound driver, or timer is executed.
- VERIFIED: `game-inventory.json` records all 33 supplied files by name, size, SHA-256 and evidence-qualified purpose. `game-inventory.md` is the readable overview. Runtime reads the 12 PRG/MOD/CFG inputs listed in README.md as read-only data containers; optional TABLE*.HI seeds stay external.

Reproduce reference download (into an absent reference directory):

```sh
git clone https://github.com/historicalsource/pinballfantasies.git reference/original-dos-source
git -C reference/original-dos-source checkout --detach aa2dd368d73886bbd666507bd7341001060700d9
python3 tools/inventory.py
```

Do not regenerate inventories to hide changed originals: compare against the saved
manifests first. `go test` validates originals when present and explicitly skips missing external inputs. Public native-content consistency uses `tools/check_public_source.py`.


## Publication boundary

Native Go implementation and tools are authored here but source-guided; this is
not proof of independent copyright ownership or a license grant. Native code and project-authored material are now MIT, copyright 2026 voobrazimoe. Historical ASM/MAC stays in the
ignored external checkout; no upstream redistribution/modification license has
been found or assumed. The owner has approved publication of source-guided native implementation and
audit material. Earlier manual-review gates in historical reports are superseded. Original commercial runtime PRG/MOD/CFG,
legacy HI files, artwork and recorded audio are separate copyrighted inputs,
never public repository/release content. Extracted score/scroller glyphs are
now decoded from the user's PRG at runtime, not embedded in generated code.
The BIOS font blob was removed. Sidebar glyphs are independently authored native
5x7 strokes, with an intentional visual difference; see docs/runtime-data.md.
Matrix texts, animation/flash records and jingle controls now decode from PRGs.
Source-guided native command/effect representations remain in the implementation.

Earlier private history contains deleted originals/reference material. Publish
only a newly audited snapshot with fresh history, never this repository's refs.
