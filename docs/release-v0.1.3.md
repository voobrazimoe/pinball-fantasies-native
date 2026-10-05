# v0.1.3 release candidate preparation

v0.1.3 remains a **prerelease candidate** until the owner accepts the staged
packages. The owner has accepted A6 physical Android testing. This task does
not authorize a main merge, a v0.1.3 tag or a GitHub Release; v0.1.2 stays intact.

## Scope

First public Android host: canonical DOS SAF import with transactional adoption,
a clean no-data import shell and concise unsupported-data errors, contextual
touch controls and external-keyboard auto-hide, Oboe audio/focus/lifecycle,
GLES/Choreographer presentation, ARM64 + x86_64 and 16 KB compatibility.
Desktop Windows/Linux x86_64 and macOS ARM64/Intel packages remain included.
An independently drawn PF monogram is integrated across all platforms.
Engine, ABI, gameplay and import transaction behavior are unchanged.
`21STCENT/FANTASY` data is not supported.

## Source and version boundary

The repository's existing Android `versionName` and macOS
`CFBundleShortVersionString` are 0.1.3; their existing numeric build fields are 3.
Windows/Linux had no maintained application-version field to update.
No new runtime versioning scheme is introduced. Candidate CI evidence and the
local validation report identify one exact source SHA for all five packages.
The owner must approve these exact staged bytes; rebuilding after approval
requires a new package smoke. Historical docs/releases are retained.

The icon source is `art/app-icon/pf-icon-master.svg`; regeneration commands and
clean-room provenance are in [the icon README](../art/app-icon/README.md).
The first icon gate uncovered an unrelated Google Play Services crash in the
host smoke's global log filter. Crash attribution now uses the PF process ID
and native package tombstone; dedicated tests retain PF Java/JNI/native failures.
No gameplay or emulator acceptance assertion was relaxed.

## Candidate files

Stage a public set from one exact commit under `release/v0.1.3-candidate/`:

- `pinballfantasies.exe`
- `PinballFantasies-x86_64.AppImage`
- `PinballFantasies-arm64.zip`
- `PinballFantasies-x86_64.zip`
- `PinballFantasies-android.apk` (universal, development-signed)
- `SHA256SUMS.txt`

Obtain desktop packages from the **Desktop release candidate artifacts** and
**macOS native hosts** workflows, and the APK from **Android host**. Verify all
workflow/artifact head SHAs and archive digests before extracting. The owner-local
staging record retains workflow/artifact provenance separately from release files.
If preparing an installable update to an existing local development install,
re-sign the public CI APK using the existing local debug signing identity with
Android SDK `apksigner`; leave every non-signature ZIP member unchanged and
record that transformation. Never stage or upload the personal APK or signing key.
This candidate does not introduce production signing keys or an AAB pipeline.

## Required validation

Require successful **Android host**, **Android A0**, **Asset-free source**,
**macOS native hosts**, and **Desktop release candidate artifacts** at the exact
candidate commit. These retain shared Go regressions, Windows/Linux builds,
native Apple host/ABI/signature checks, packaged Android JNI/input/audio tests,
public APK/ABI/export checks, and the interpreted 16 KB emulator smoke.
Hosted checks do not replace the accepted physical A6 evidence or final RC smoke.

The dependency-free `tools/check_app_icons.py` checks source structure, raster
sizes, ICO/ICNS, Android references and the entire adaptive safe circle. Its
`--exe` mode compares all seven linked Windows icon resources to the exports.
AppImage CI checks the actual root PNG, `.DirIcon` and desktop icon reference.
macOS bundle validation requires the exact ICNS before ad-hoc signing.

For final local package inspection (Python 3.10+, Android SDK Build Tools 36.0.0
and JDK 17; `JAVA_HOME` configured):

```sh
python3 -m venv .cache/rc-inspection
.cache/rc-inspection/bin/pip install dissect.squashfs==1.12
.cache/rc-inspection/bin/python tools/check_release_candidate.py \
  release/v0.1.3-candidate --source-sha "$(git rev-parse HEAD)" \
  --ci-evidence release/v0.1.3-evidence/ci-evidence.json \
  --android-sdk "$ANDROID_HOME" \
  --report release/v0.1.3-evidence/validation.json
# Optional owner-local stronger scan: add --originals /outside/repo/originals.
(cd release/v0.1.3-candidate && shasum -a 256 \
  pinballfantasies.exe PinballFantasies-x86_64.AppImage \
  PinballFantasies-arm64.zip PinballFantasies-x86_64.zip \
  PinballFantasies-android.apk > SHA256SUMS.txt)
(cd release/v0.1.3-candidate && shasum -a 256 -c SHA256SUMS.txt)
```

`ci-evidence.json` is an array with one record per named gate: `name`, `id`,
`sha`, `status`, `conclusion`, `url`; require `completed`/`success` throughout.
The package checker inspects Windows icon resources, both Mach-O architectures
and version/icon metadata, the full AppImage filesystem/runtime manifest,
all APK entries/native libraries, ELF and ZIP 16 KB alignment, manifest version,
launcher icon and APK signing. It scans source and unpacked payloads for known
commercial file hashes/forbidden filenames/private build paths. Optional private
originals are used only in memory for nontrivial 4 KiB byte-block searches; no
original bytes, hashes or paths are written into the report.

## Owner RC smoke — stop here

Android: confirm the launcher icon and mask/cropping, first-run shell, SAF import,
gameplay launch, touch controls and smoothness. Where available on macOS/Windows:
confirm app/executable icon, launch and basic gameplay. This is a final visual
and package check, not a new broad parity campaign.

After preparation, stop and await explicit owner RC acceptance. The next,
separately authorized operation may verify main, integrate the approved commit,
confirm main CI, tag v0.1.3, publish a GitHub prerelease, upload only approved
packages/checksums, download and re-hash them, and update live README links.
The intended English/Russian README publication changes are staged in
[release-v0.1.3-readme.patch](release-v0.1.3-readme.patch); apply them only after
all v0.1.3 download URLs exist. Current live links stay at v0.1.2.

The concise draft is [release-notes-v0.1.3.md](release-notes-v0.1.3.md).
