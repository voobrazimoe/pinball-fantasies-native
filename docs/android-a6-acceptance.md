# A6 physical Android acceptance record

**Core physical A6: owner-reported PASS at `f35f27d` (2026-10-05).** A0–A5 are accepted.
The checklist below remains a per-device template; unreported individual checks
are NOT TESTED. The final import UI still needs owner confirmation on-device.
Hosted, asset-free CI and emulator results cannot certify this phase. A7 signing,
release publication and merging to `main` are outside this work.

Copy this template outside the repository for each device. Every status below
must be **PASS**, **FAIL**, or **NOT TESTED**, with short factual notes. Leave
untested checks as NOT TESTED. Keep commercial screenshots, originals, provider
archives and original-backed app storage out of Git, uploads and CI artifacts.
Local human-comparison screenshots, if needed, belong outside the repository.

## Build and run

Use the ordinary asset-free `:app:assembleDebug` APK or the local personal APK
from `android-host`. Debug signing is
sufficient. Do not change SDK levels, engine behavior, ABIs, or runtime options.
The existing build packages arm64-v8a and x86_64; diagnostics default to OFF.
The CI guest's `-Xint` workaround is a device setting, not part of the APK.
Do not use `run_android_host_smoke.sh` on the owner's acceptance device: that CI
smoke changes global rotation settings and exercises a no-data shell.

With the existing Go/NDK/JDK/Gradle setup described in [android.md](android.md):

```sh
sh tools/build_android_engine.sh all
gradle -p hosts/android :app:assembleDebug
shasum -a 256 hosts/android/app/build/outputs/apk/debug/app-debug.apk
```

For convenient private gameplay/device testing, build locally:

```sh
python3 tools/build_personal_android.py /path/to/originals
# Output: release/personal/android/pinball-fantasies-personal.apk
```

This debug-signed APK contains the owner's commercial originals. Private/local
use only; never upload or redistribute it. Public builds remain asset-free; A7
release signing is not started. It can be installed by tapping the APK with the
ordinary Android installer and launches from bundled data without adb or SAF.
Existing valid Data wins and State is preserved. The personal builder separately
builds and checks an asset-free public APK.

Record **personal bundled** versus **public SAF** as the input source for each
run. A personal APK may support compiled ART, four-table gameplay, touch,
orientation, Oboe/focus/routes, lifecycle, persistence and practical performance
observations. These do **not** certify any real SAF picker/provider/transaction
check below. Leave those NOT TESTED until explicitly exercised through the central **Import DOS folder** control in a fresh no-data
public installation. Normal gameplay has no replacement/reimport menu. Overall A6 stays NOT TESTED until every
required acceptance area has real evidence.

The evidence helper's `--apk` option deliberately rejects commercial payloads.
For the personal APK, install by tapping it (or ordinary `adb install -r`), then
run the helper **without `--apk`** and record its APK hash independently. Do not
weaken the helper's public-APK gate to collect personal gameplay evidence.

Connect/unlock a physical Android device and authorize adb. Set the actual
serial from `adb devices`; the helper never chooses an arbitrary device.
Python 3 and adb on PATH are required. No root is used.

```sh
SERIAL='your-physical-device-serial'
APK='/absolute/path/to/app-debug.apk'
EVIDENCE="${TMPDIR:-/tmp}/pf-a6-$(date +%Y%m%d-%H%M%S)"
adb -s "$SERIAL" install -r "$APK"
sh tools/run_android_a6_device.sh --serial "$SERIAL" --mode session \
  --seconds 120 --output "$EVIDENCE/normal"
```

`session` deliberately force-stops then launches a fresh process, with diagnostics
OFF. Install uses `-r` only; it never uninstalls, clears storage or repairs a
signature conflict. An existing app signed with a different debug key requires
owner planning to preserve its state; do not uninstall it to make this pass.
Alternatively `--apk "$APK" --mode session` installs and records its SHA256
after checking expected native libraries and rejecting commercial filenames.
Record build provenance and APK hash independently even when using plain adb.

During the session manually select the DOS folder through the central no-data **Import DOS
folder** control, then perform the observations below. For a representative diagnostic
session, explicitly opt in using the existing Activity diagnostic extra:

```sh
PF_DIAGNOSTICS=1 sh tools/run_android_a6_device.sh --serial "$SERIAL" \
  --mode session --seconds 120 --output "$EVIDENCE/diagnostics"
```

This fresh launch resets the in-memory game, so start a representative game
after launch. The helper translates `PF_DIAGNOSTICS=1` to the supported boolean
Intent extra. No persistent device property is changed. To turn it OFF, run a
new ordinary session, or force-stop and launch from the launcher.

Explicit lifecycle commands (run after saving/recording the intended state):

```sh
# Read-only snapshot of an already-running game; no launch or input.
sh tools/run_android_a6_device.sh --serial "$SERIAL" --mode snapshot \
  --seconds 30 --output "$EVIDENCE/snapshot"
# Home, two-second wait, return; same PID checked, gameplay continuity is manual.
sh tools/run_android_a6_device.sh --serial "$SERIAL" --mode home-return \
  --seconds 30 --output "$EVIDENCE/home-return"
# Force-stop/relaunch; A2 bootstrap and State reload must be observed manually.
sh tools/run_android_a6_device.sh --serial "$SERIAL" --mode restart \
  --seconds 30 --output "$EVIDENCE/restart"
# Optional practical process-death scenario: background first. Android may decline
# to kill a foreground/persistent process. Observe PID death; otherwise NOT TESTED.
adb -s "$SERIAL" shell input keyevent KEYCODE_HOME
adb -s "$SERIAL" shell am kill io.github.voobrazimoe.pinballfantasies
adb -s "$SERIAL" shell pidof io.github.voobrazimoe.pinballfantasies
adb -s "$SERIAL" shell am start -W -n io.github.voobrazimoe.pinballfantasies/.PinballActivity
```

Rotate, lock/unlock, switch recent apps and generate focus/route interruptions
manually. The helper never locks rotation, changes global settings, enables
root, copies source files, or reads private Data/State. Do not use global
"Don't keep activities" settings just to force Activity recreation.

The helper refuses known emulators and forced interpreted/disabled-JIT runtime
properties. These checks cannot prove physical hardware or actual ART execution:
the owner must attest to both. Native/ART library and JIT mappings corroborate
the runtime where `run-as` is allowed; absence of a JIT mapping alone is not a
failure (AOT execution is legitimate). Metadata APIs may be denied on some
devices; mark the affected observations NOT TESTED. Empty metadata is not zero.

Collection is bounded to 5–600 seconds of observation plus command overhead,
with 20-second command timeouts, up to 120 retained audio counter records, and
no raw log file. Only app event names/numeric A4/A5 fields and package-associated
crash markers are retained. Provider error strings and stack bodies are omitted.
The crash scan starts at the recorded device-time second (up to one second
before collection) and cannot establish crash absence outside that window.
`gfxinfo` may omit native GLES frames. CPU is a system
sampling summary, not a gameplay benchmark. Audio enums are metadata, not proof
of an audible route. This helper never assigns a manual PASS or overall A6 PASS.

## Run identity

| Field | Value |
| --- | --- |
| Date / owner tester | To fill |
| Tooling commit / engine source commit | To fill |
| APK type (personal bundled / public SAF), build run / local command | To fill |
| APK absolute path / SHA256 | To fill |
| Debug signing identity / installed package version | To fill |
| Physical model / manufacturer / Android version | To fill |
| ABI / page size | To fill |
| Normal ART evidence / no `-Xint` | To fill |
| Legitimate-owner originals attestation (no hashes/contents) | To fill |
| Local evidence directory (never upload app storage) | To fill |
| Headset / Bluetooth / USB availability | To fill |

## Automated evidence

PASS means only the narrowly described check, never overall acceptance.

| Check | Status | Short notes / local evidence |
| --- | --- | --- |
| Debug APK built, arm64-v8a + x86_64, unchanged SDKs | NOT TESTED | |
| Native ELF + ZIP 16 KB alignment, debug signature | NOT TESTED | |
| Separately built public APK has no DOS/personal payload | NOT TESTED | Personal APK itself contains owner inputs |
| Ordinary Android package install / package and version metadata | NOT TESTED | |
| Physical-device preflight / no forced interpreted ART | NOT TESTED | Owner attestation also required |
| ABI / page size / native mappings | NOT TESTED | |
| Process alive at end of collection | NOT TESTED | |
| Bounded package-associated crash scan | NOT TESTED | State exact time window |
| Home/return command, same PID | NOT TESTED | Does not prove same game |
| Force-stop/relaunch command, fresh PID | NOT TESTED | Does not prove State reload |
| A2 validated/adopted/bootstrap event sequence | NOT TESTED | Diagnostics run only |
| A4/A5 counter observations | NOT TESTED | First/last samples, deltas and duration |
| Memory/CPU samples, route metadata, orientation, frame summary | NOT TESTED | Unavailable fields NOT TESTED |
| Asset-free A1–A5 hosted regressions | NOT TESTED | Commit and CI link; not A6 evidence |

## Owner-observed SAF and transaction checks

Required direct children: `INTRO.PRG`, `INTRO.MOD`, `MOD2.MOD`, `TABLE1.PRG`,
`TABLE1.MOD`, `TABLE2.PRG`, `TABLE2.MOD`, `TABLE3.PRG`, `TABLE3.MOD`,
`TABLE4.PRG`, `TABLE4.MOD`. `PINBALL.CFG` is optional. `TABLE*.HI` is never
required/imported. Use only the owner's legitimate originals, with a safe backup
under their control. Never alter their only copy for rejection tests.

| Check | Status | Short notes |
| --- | --- | --- |
| Real SAF folder selection, all 11 inputs accepted | NOT TESTED | Provider type only; omit source URI |
| Optional CFG present accepted / absent accepted | NOT TESTED | Test variants in owner-controlled copies |
| Irrelevant files and TABLE*.HI ignored | NOT TESTED | |
| Real shared engine validates, staging adopted, persistent engine starts | NOT TESTED | |
| Inputs in app-private noBackupFilesDir/Data; state in filesDir/State | NOT TESTED | Architecture + behavior; do not export contents |
| Successful import no longer needs source provider | NOT TESTED | Safely disconnect/revoke source access, relaunch |
| No source-folder writeback / no broad storage permission | NOT TESTED | Owner checks source locally, no published fingerprints |
| Cancel another SAF selection, existing Data/State still playable | NOT TESTED | |
| Reject separate missing-input folder, previous Data/State intact | NOT TESTED | Empty folder is safe |
| Reject invented invalid substitutes, previous Data/State intact | NOT TESTED | Never corrupt originals |
| Re-import valid originals via staging/adoption/rollback design | NOT TESTED | State preserved; no induced destructive failures |

## Each table and controls

Record each table separately. For **every** table observe correct visuals;
tracker/music progression; left and right flippers; simultaneous flippers;
plunger pull/release; nudge/tilt; pause/menu; exit/Back; score and ball
progression; and no obvious simulation-speed drift. A cell may contain PASS,
FAIL or NOT TESTED and short notes. Any unperformed component remains NOT TESTED.

| Table | Start / visuals | Music | L/R + simultaneous | Plunger | Nudge/tilt | Pause/menu + exit/Back | Score/ball + speed |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 Party Land | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED |
| 2 Speed Devils | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED |
| 3 Billion Dollar Gameshow | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED |
| 4 Stones 'N Bones | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED | NOT TESTED |

Touch polish retest: the owner reports the first personal APK launches and works,
but camera-cutout overlap, the permanent F-row and bottom-right Pull geometry
require a new physical pass. A6 remains incomplete; this report does not fill
individual acceptance cells.

Touch-first revision after the second physical report: **NOT TESTED**.
The permanent keyboard toolbar and F row are removed. Verify Tap to continue,
named tables/Options, attract Players 1..8 + Play, contextual options,
Resume/Exit, Yes/No and temporary initials letters. During Playing verify an
unobstructed 320×33 matrix, tiny safe menu, centered L at 25% and R at 75%, and
Pull ↓ only when the authoritative engine flag is active. Menu → Advanced
keyboard is a dismissible in-Activity overlay; opening/closing it must not
interrupt audio or suspend gameplay.

Bottom 35% of the safe area is split into left/right flipper halves. Neutral
short taps nudge once on release. A right-side downward drag can claim the
plunger only while available; 25% of safe height maps to full **32** charge.
Test half/full fast and slow swipes, immediate release at full charge, no
Nudge contamination, horizontal jitter, cancel without launch, L+R,
L/R+Nudge and L/R+Pull. Verify safe menu access and matrix visibility in
portrait and landscape, no held controls after focus loss, and reconstruction
from the current engine state after rotation. Check physical keyboard
compatibility when available; otherwise leave that check NOT TESTED.

## Orientation, lifecycle and persistent state

| Check | Status | Short notes |
| --- | --- | --- |
| Portrait full 320×609 table, usable touch, transient ScrollOff | NOT TESTED | |
| Portrait does not change saved scrolling setting | NOT TESTED | |
| Landscape HARD/MEDIUM/SOFT/OFF restored, aspect/letterboxing | NOT TESTED | Test all four saved preferences |
| Repeated active rotation: no texture corruption/engine restart | NOT TESTED | |
| Rotation preserves game/score/ball/player and coherent tracker | NOT TESTED | |
| Rotation: no stuck input or stale audio burst | NOT TESTED | |
| Home → return and recent-app switch → return | NOT TESTED | Same session where architecture allows |
| Screen off → unlock → return | NOT TESTED | |
| Temporary window focus loss | NOT TESTED | |
| Activity recreation when Android chooses to recreate | NOT TESTED | Record whether actually observed |
| Force-stop then relaunch: A2 bootstrap, saved native state reloads | NOT TESTED | No promise of mid-ball process-death save |
| Practical process kill/relaunch | NOT TESTED | Confirm kill occurred |
| No duplicate engine, stuck input, stale PCM, crash or deadlock | NOT TESTED | |
| Scrolling preference survives restart | NOT TESTED | |
| Existing volume/music settings behave as designed | NOT TESTED | |
| High scores/state persist separately from original inputs | NOT TESTED | Complete real game / initials where needed |
| Optional legacy CFG seeds only as currently designed | NOT TESTED | Do not redesign persistence |

## Audio and practical performance

Use built-in output first. Diagnostics are local evidence only. For a
representative session compare first/last A4_AUDIO fields (`underruns`,
`missing`, `overflows`, `dropped`, `high`, `opens`, `errors`, `restarts`), A5
focus/route transitions, and memory/CPU samples. Counters are cumulative within
their instance: compare deltas within a single process, record lifecycle/route
events and duration, and do not compare unrelated instances. Explain audible
failures even if counters look acceptable. Judge practical presentation/input
latency and memory trend; no new profiler or timing compensation is involved.

| Check | Status | Short notes |
| --- | --- | --- |
| Built-in speaker: continuous real Oboe tracker/music, no crackle | NOT TESTED | |
| Pause/resume: fresh audio, no stale/repeat burst or long delay | NOT TESTED | |
| No runaway underrun/overflow/restart behavior | NOT TESTED | Counter deltas + listening |
| Frame stability / practical input and output latency | NOT TESTED | Human observation, not synthetic benchmark |
| Memory growth / CPU load representative session | NOT TESTED | Duration and limitations |
| Gameplay speed independent of device audio timing | NOT TESTED | |
| Wired headset connect/disconnect during play | NOT TESTED | Unavailable hardware remains NOT TESTED |
| Bluetooth A2DP start/disconnect/reconnect, background/foreground | NOT TESTED | System-dependent latency accepted |
| Bluetooth transient focus event where practical | NOT TESTED | |
| USB audio route-change safety | NOT TESTED | Unavailable hardware remains NOT TESTED |
| Route changes: game continues, no crash/deadlock/stale burst | NOT TESTED | Record each tested route |
| Another media app / transient notification interruption | NOT TESTED | Record reproducible event and A5 response |
| Permanent focus loss, later legitimate gain | NOT TESTED | Loss may require background/foreground to request again |
| Focus obeys A5 policy, game state preserved, fresh resumed audio | NOT TESTED | |
| Phone call if safe/practical | NOT TESTED | May remain NOT TESTED |

## Completion decision

**Extended checklist: NOT TESTED where no specific result was supplied.** Do not
interpret the owner-reported core PASS as certification of every route below.
Full checklist PASS requires actual physical
testing with legitimate originals demonstrates normal compiled ART, successful
real SAF/bootstrap, four playable tables, working controls, both orientations
without reset, built-in Oboe audio, safe pause/resume/rotation/process relaunch,
native state persistence, cancel/invalid re-import preservation, and no
commercial payload in the separately built public APK, repository or report.
A local personal APK contains commercial originals and must remain private. Headset/Bluetooth/USB must each
be explicitly PASS or NOT TESTED according to available hardware; report any
FAIL honestly. Record unresolved failures and missing evidence here.

For a discovered bug, reproduce narrowly, add an asset-free regression where
feasible, fix only that bug without changing source/gameplay semantics, rerun
applicable A1–A5 CI, and put meaningful fixes in clear separate commits.
This preparation does not authorize A7 or a merge to `main`.


## Final A6 import acceptance report (2026-10-05)

| Evidence | Result and scope |
| --- | --- |
| Physical gameplay/touch | Owner-reported PASS at f35f27d; no new per-table hardware trace supplied. |
| External keyboard attach/detach | Owner-reported PASS. |
| Audio/lifecycle | Owner-reported PASS for exercised behavior; wired/Bluetooth/USB routes remain NOT TESTED unless separately recorded. |
| Physical smoothness | Owner-reported PASS after Choreographer pacing fix. |
| Clean public PINBALLF SAF import | Owner-reported PASS: imported and ran. Establishes real A2/SAF success. |
| Alternate installation | Unsupported; INTRO/TABLE1/TABLE2 offsets differ, and Party Land S_EMPTY priority changes. Modules and TABLE3/4 validate independently. No second profile; alternate four-table gameplay NOT TESTED. |
| First-run cleanup | Implemented black native background and Android shell, hidden no-data game controls, busy/restore/success states. New instrumentation covers these transitions; new on-device visual confirmation pending. |
| Import diagnostics | Concise allowlisted filename messages; parser details only in explicitly enabled logs/internal errors. |
| Transaction safety | Existing staging/disposable validation/rename/rollback/State behavior retained; rejection-preservation regression includes unsupported-layout failure. |

The source audit is [runtime-compatibility-audit.md](runtime-compatibility-audit.md).
No commercial inputs or whole-file hashes are published. PINBALLF remains the
canonical runtime and strict oracle input. No A7 or main merge is performed.

Local final validation: canonical PINBALLF independent validation passes all 11
inputs; targeted datalayout/frontend regressions and the engine's four-table
SOFT/OFF direct-Runner conformance pass. Alternate independent validation and
module-equivalence/S_EMPTY regressions pass with expected rejection. Java import
transactions/error sanitizer, A3 geometry/controls/native-session/pacing, A4 audio
and A5 policy tests pass. Both engine ABIs build with 16 KB LOAD alignment, and
public debug APK plus instrumentation APK build. New first-run instrumentation
is pending hosted execution; no connected device was available locally.

A broad original-backed `go test ./...` in an isolated private checkout was also
attempted. It is not a passing gate: historical private art/trajectory/presentation
reference captures are absent, and the macOS desktop platform has existing
undefined host symbols in this invocation. These do not invalidate the passing
targeted runtime/engine tests, but the full historical oracle suite is unverified.
Hosted CI results should be checked against the pushed code commit. Public
packaging/source scans remain asset-free; private reports stay under ignored
`.cache/edition-audit`. The production canonical profiles were not regenerated.
