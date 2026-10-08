DMO0 NOT CLOSED. DMO1 NOT STARTED.

# DMO0 native audio / gameplay callback boundary — 2026-10-07

**NATIVE_AUDIO_BOUNDARY = NOT_PROVED**

Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
Actual research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
This pass is uncommitted research, with no production changes.

Owner-local result: `/private/tmp/pf-dmo0-audio-boundary.json`.
Extractor: `tools/audit_10min_demo_audio_boundary.py`.
The authoritative whole-DOS input remains
`/private/tmp/pf-dmo0-admission-domains.json`; it was read, never rewritten.
TABLE1 remains **30 total / 18 bounded / 12 UNKNOWN**; CODE2 remains
**2 total / 0 bounded / 2 UNKNOWN**, with whole-DOS gate **exit 2**.
The new command also returns 2 for its own unsuccessful semantic proof. These
are distinct scopes and neither command grants native support.

## What was known before this pass

API11 registration/root/BL=100, API12 registration/root/BL=200, primary update
guards, demo slowdown mask zero, the admitted rest-of-update's single electronics
call, drain ordering, equality threshold 35998 and pause's update-disable effect
were already established. All eleven supplied SDR were decoded and primary
consumer candidates found. Their registration evidence was not scheduler proof.
This pass does not claim those facts as newly discovered, nor revisit arbitrary
SOUND.CFG executable admission, complete IVT writers or physical stack closure.

## A. Semantic entry ABI: proved observations and limits

TABLE1 file offsets use CS base `0x300`; driver offsets below are decoded-module
offsets. The callback assigns its own relocated TABLE1 DS. Incoming DS is not a
semantic input to its own body before that assignment. Incoming **AX zero versus
nonzero is a real input**: `0x4537` tests it, with shared DS:`0x37f1` TIME_LEFT
initially true and changed to false on nonzero AX. API12 repeats this handshake
at `0x5951–0x595d`. Incoming flags are replaced by the explicit AX test; no other
incoming general register selects admission in the local guard prefix.

SYNC_COUNTER DS:`0x364a` increments, and INT_WAS_HERE/INT_WAS_HERE2
DS:`0x364c/0x364d` are set before primary's INTERRUPTS_ON test. Later sets the two
flags before its guards but does not increment SYNC_COUNTER in its own body.
These stores cannot be discarded merely because electronics was suppressed.
SYNC/callback-observed flags also occur outside the ordinary electronics path;
this pass does not replace those consumers with an electronics counter.

Gameplay/update admission reads TABLE1-owned INTERRUPTS_ON, DEMOMODE,
SLOWCNT, CS:LAST_WAS_VB, INSIDE_BALLHANDLER and INSIDE_RESTOFVBLANK.
Later additionally reads INSIDE_RASTINT. Resolution, ball/scroll state, keyboard
state and task state are TABLE1 state, not new parameters delivered by a sound
device. TIME_LEFT is shared mutable state: a nested callback can overwrite it
before an outer callback's budget test.

At `0x472a`, false TIME_LEFT skips the `0x4746 -> 0x479d` dot-matrix/animation
call. It **does not skip electronics, areas/targets/shift, KEYTASK or tasks**:
those have already run inside `0x4723 -> 0x5cd9`. The subsequent lights call at
`0x4750` still occurs. The matrix path includes animation progression and its
indirect PRINTTASK consumer. Prior evidence includes the demo-expiry
continuation ending in table QUIT. Therefore matrix suppression is an observable
continuation/presentation-progress input, not proved to be a harmless device
detail. No PRINTTASK or task-domain closure was undertaken to dismiss it.

All compared driver consumers compute AX from CF, save a copy with PUSH AX,
make the far call, and recover the saved crisis value with POP BX. A nonzero
saved value plus callback return AX different from 12345 triggers the driver's
API8 emergency fallback. Normal primary returns at `0x4536`, `0x4713`,
`0x479a`, `0x6439` use the sentinel 12345. Thus the callback return handshake has
observable audio-service meaning. It is not a gameplay-return value. Later's
extra disabled attract-path return at `0x644b` uses AX=0; this is another reason
not to apply a blanket sentinel theorem to every interleaving.

The saved crisis word is **not proved to be an incoming TABLE1 argument**.
The new local-body walk finds no explicit SS-relative memory reads in either
body (377 primary / 287 later instruction positions). It treats calls as
returning opaque effects and exports every such dependency. This does not
prove that all transitive callees ignore incoming ES/general registers, stack
words or segment-derived addresses. A complete gameplay ABI is consequently
**not certified**. Physical return frames and graphics-register save values
are calling/rendering machinery; they do not by themselves define native
gameplay parameters. Unknown physical stack placement is not evidence that a
native semantic model must implement DOS stacks.

## B. Electronics count per primary

The new local CFG walk follows both conditional successors, keeps matched
fallthrough at every opaque CALL, and rejects cycles, indirect body branches
and adjusted/non-far returns. It covers primary's ordinary and DEMOMODE body
branches. It finds **{0,1} own-body direct calls** to DO_ELECTRONICS. Later has
**{0}**. There is one counted site, `0x4723`; no local path loops back to it.
This upper bound is stronger than observing one admitted execution, but its
callee/return premises are explicit.

| Own-body zero case | Effect |
| --- | --- |
| INTERRUPTS_ON != true | Return before normal update; SYNC/flags already changed |
| DEMOMODE == true | Attract callback body, outside ordinary electronics |
| Nonzero slowdown result | Structural exit; infeasible from the pinned demo AND mask 0 alone |
| LAST_WAS_VB == true | Return until an eligible later callback clears it |
| INSIDE_BALLHANDLER == true | Return without entering another ball handler |
| INSIDE_RESTOFVBLANK != false after ball processing | Ball/physics work can occur, but electronics is skipped |

With those ordinary gates admitted and callees returning, primary reaches
UPDATE_COUNTERS then the single electronics site. Drain itself does not add a
new bypass. Budget, resolution, HOLDSTILL, matrix wait and between-ball state
are not extra pre-electronics guards in this body.

**No >1 direct calculation path exists in the decoded local body.** This is
not a proof that an invocation's entire nested execution interval contains at
most one calculation, nor that every transitive feasible execution returns.
The tool deliberately does not infer effects for the indirect spring/matrix/
task/area dependencies. Those older gameplay domains remain separate. A second
callback entered before the first returns is a second invocation, not a loop
in the first body. An unconditional all-effects per-invocation theorem is
therefore not claimed.

## C. Paired scheduler state is gameplay relevant

Primary sets LAST_WAS_VB at `0x4714` before the admitted rest-of-update, and at
`0x4709` on the rest-busy return. Later clears it at `0x5a69`, after its ordinary
rendering/input path, then clears INSIDE_RASTINT. Its earlier exits do not clear
the latch: slowdown, updates disabled, DEMOMODE diversion, raster handler busy,
latch already false, or ball handler busy all prevent that ordinary clear.
The screen-start call at `0x597b` precedes the ball-busy guard. Thus even a
later invocation is not synonymous with a latch release.

Primary can be skipped until a successful later callback. Order and multiplicity
matter in the simple serial projection: P,L,P admits two calculations whereas
P,P,L admits one. These are **authored abstraction traces**, not assertions of
actual IRQ schedules. Missing/ineligible L can hold off further primary updates.
A late callback may be permitted at the driver but rejected by TABLE1 guards.

Both resolutions use these same latch/busy guards and the same electronics
site. The actual API12 parameters are **CX=0x108 (264 decimal) low-res** and
**CX=0xae (174 decimal) high-res**, rather than decimal 108 for low-res.
Different scheduling positions do not establish equal admissible nested traces.

Pause disables INTERRUPTS_ON at `0x34d8`; resume reenables it at `0x3530`.
Neither store clears LAST_WAS_VB. A paused later callback exits without its
ordinary clear. Consequently, if the latch was set before pause, the first
primary after resume still takes the latch exit until an eligible L clears it.
With a clear latch and other guards admitted, the first resumed P can update.
This is a state-dependent local theorem, not a first-resume driver delivery
ordering theorem. This pass has not proved that audio-stop/resume guarantees a
particular first delivered callback.

## D. Actual supplied-driver differences

The tool revalidates each saved consumer against the current identity-checked
decoded module and derives API11/API12 storage again. It checks the CF-to-AX
producer and sentinel fallback grammar for all eleven primary sites and all
three direct-family later sites. Indexed callbacks use the same dynamic far
consumer for primary and later records. This mechanically propagates the
**local handshake**, not complete IRQ-source or schedule equivalence.

| Semantic family | Drivers | Budget supplied to TABLE1 |
| --- | --- | --- |
| Indexed raster records, constant budget | NOSOUND, GUS | Helper CLC/RET; AX=0 |
| Indexed raster records, audio budget | PAS16, SB16, SB20, SBLASTER, SBPRO, SM2 | AX=0 or 65535 from active/buffer predicate |
| Alternating direct pointers, audio budget | INTERNAL, ADLIB, THING | Same budget predicate grammar; distinct primary/later pointer calls |

NOSOUND uses the indexed nine-byte record traversal, advances the current record
before dispatch, and compares its priority with the currently active priority.
BL=100 primary is below BL=200 later. The local branch suppresses a primary
while priority 200 is active; it permits later at priority 100 and permits equal
priority. Interrupt enablement precedes the callback. TABLE1 busy guards add
their own admission layer. NOSOUND does not produce an audio-budget crisis.

INTERNAL uses a timer-countdown and alternating phase byte (CS:`0xafa`), separate
primary/later fields, and calibrated split position derived from API12 CX.
Its primary consumer is `0xbdf`, later `0xc13`, budget helper `0x19b3`.
The primary path sets the alternating phase before invoking TABLE1; the later
path clears it before invoking TABLE1. Its compared handshake has no indexed
active-priority gate. This does **not** certify that outer source routing has no
other admission conditions. The helper's active flag and remaining-buffer
comparison can return CF=1, which becomes AX=65535. ADLIB/THING share the checked
direct handshake and budget grammar, not a proved identical whole source.

The other eight indexed dispatch suffixes pass the same priority/handshake
checks. GUS shares NOSOUND's constant-clear-CF helper. The six audio-budget
indexed drivers have the same active/threshold predicate grammar, with their
own buffer-position helpers. Audio hardware position, calibration and outer
source state were not collapsed into a single family certificate.

Answers within the evidence scope:

1. **Selected SDR cannot yet be proved to affect cadence only.** Local priority,
   phase and budget differences are real; their full feasible trace projections
   onto TABLE1 updates have not been proved identical. This is not an asserted
   measured divergence between two hardware runs.
2. Different wall-clock/audio cadence is compatible with these mechanisms, but
   preservation of a common logical update sequence is **unproved**.
3. Driver-specific priority skipping exists in the indexed family. If a primary
   is delivered while rest is busy, TABLE1 can perform ball/physics work without
   electronics. Whether all relevant source phases admit such traces and how
   many updates result remain unproved; no constant physics/electronics ratio is
   certified.
4. **Yes, delivery alone is insufficient as an ABI:** the driver budget predicate
   becomes AX and shared TIME_LEFT. The return sentinel affects service fallback.
   No exhaustive absence of further transitive gameplay-visible driver state is
   asserted.

## E. Minimal unresolved fact and native boundary

The minimal concrete counter-obligation is a **later callback completing during
an unfinished primary rest-of-update**, clearing LAST_WAS_VB and overwriting
shared TIME_LEFT, followed by another primary delivery before the first returns.
The indexed priority predicate allows L inside P and, after L returns, equal-
priority P inside the still-active P. TABLE1 rejects/partly admits these entries
using its latch and busy flags. The exact source-phase feasibility and logical
sequence effects are not certified. Such an interval could also change which
budget reaches the outer matrix test. It cannot be dismissed by proving RETF
balance or by counting the single local electronics call.

The **one next dependency** would be a paired scheduler semantic transition
certificate: current-record/active-priority state for indexed dispatch and
countdown/alternating-phase state for direct dispatch, projected only onto
LAST_WAS_VB, TABLE1 busy guards, delivered budget and its shared overwrite.
It must establish the allowed nested/dropped P/L traces (including resume), or
show the precise difference the native model must retain. It does not require
complete IVT writer sets, IRQ physical stacks, resident placement or arbitrary
external executables. This dependency was identified, not pursued as a new
wide closure pass.

**ElectronicsCalculation remains a necessary local event candidate**, after
ball/drain and UPDATE_COUNTERS, before areas/targets/shift/KEYTASK/tasks, including
every running state which reaches that site. Disabled updates, attract mode,
latch and busy guards suppress it as described above. Its count has no proved
mapping to Runner ticks, presentation frames, physics substeps or native audio
callbacks. Neither SDR independence nor a serialized native schedule is proved,
so this report does not publish a complete minimal native contract.

## Old DOS obligations: scope classification

| Obligation | Classification and semantic reason |
| --- | --- |
| Unrestricted SOUND.CFG EXEC | BELOW-NATIVE-BOUNDARY: a native runtime which never executes that file need not reproduce arbitrary replacement DOS programs; no claim of equivalence to them is made |
| Complete INT66 vector admission | BELOW-NATIVE-BOUNDARY for vector identity/writer sets: typed native audio services have no IVT; observable service and callback effects are NATIVE-RELEVANT |
| SDR resident placement | BELOW-NATIVE-BOUNDARY: no native resident real-mode placement; address overlap is not itself a native logical event |
| SDR handler stack depth/frame/IRET/RETF balance | BELOW-NATIVE-BOUNDARY for physical calling machinery; callback order, skipping and nested semantic effects are NATIVE-RELEVANT |
| Callback source stack provenance | UNRESOLVED-BOUNDARY: provenance itself is physical, but no transitive semantic ABI certificate yet excludes stack-derived inputs; local absence of explicit SS reads cannot do that |
| CODE2 stack alias obligations caused solely by external callback stack placement | BELOW-NATIVE-BOUNDARY for that physical alias route; CODE2 lookup contents, roles and consumed effects remain independent native obligations |
| API11/API12 phase, priorities, nesting, budget, latch release and resume delivery | NATIVE-RELEVANT; currently unresolved semantic transition certificate |

These classifications do not turn writer/stack obligations into BOUNDED DOS
domains. Nor do they authorize excluding a gameplay effect merely because its
implementation uses DOS. The arbitrary executable and physical-address portions
need not be closed for the restricted native semantics being researched. The
paired callback semantic dependency remains a blocker independently of the
old whole-DOS UNKNOWNs.

## Verification performed in this pass

- **New-pass private tests: 14 PASS** (9 authored CFG/trace tests, 5 private
  integration/mutation tests). Log: `/private/tmp/pf-dmo0-audio-boundary-new-tests.log`.
- **Entire current demo research suite: 145 PASS**. This includes those 14;
  it is not a prior-pass count. Log: `/private/tmp/pf-dmo0-audio-boundary-all-tests.log`.
- **New-pass no-private tests: 9 PASS**, private class clean SKIP before its
  Capstone import. Log: `/private/tmp/pf-dmo0-audio-boundary-public-new-tests.log`.
- **Focused existing A/B/C/D regressions: PASS, zero SKIP**, for datalayout,
  frontend and Stones. Includes coherence matrix, mappings/seeds, priorities,
  shared runtime, and production demo rejection. Corrected-run log:
  `/private/tmp/pf-dmo0-audio-boundary-go-tests-corrected.log`.
  Initial run failed because D was mistakenly pointed at a GOG container
  directory without PRGs; preserved initial log:
  `/private/tmp/pf-dmo0-audio-boundary-go-tests.log`. Correcting only the fixture
  path to the existing unpacked D installation produced the passing run.
- **Fresh existing whole-DOS resolver: exit 2**, same TABLE1 18/12 and CODE2
  0/2. Its new output is `/private/tmp/pf-dmo0-audio-boundary-dos-gate.json`;
  log `/private/tmp/pf-dmo0-audio-boundary-dos-gate.log`. This is a separate
  output, not an overwrite of the authoritative saved result.
- **git diff --check and new-file encoded-payload checks: PASS**. No changed
  internal/hosts/cmd paths against HEAD; staged file list empty. The saved
  historical JSON and .DS_Store retain the observed inode/size/mtime/ctime.

Reproduce the new command with PF_10MIN_DEMO_DATA and its --output argument;
its default --saved is the authoritative historical admission result. All
inputs remain owner-local. The artifact contains addresses, semantic metadata,
opaque dependencies and authored trace results, not executable/picture/audio
payload. No historical JSON was relabeled or rewritten. No runtime/internal/
hosts/cmd implementation, .DS_Store, release assets, commit, push, tag or DMO1
work is part of this pass.
