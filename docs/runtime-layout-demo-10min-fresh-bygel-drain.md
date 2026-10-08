# DMO0 fresh-session BYGEL scored-drain trajectory provenance

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

`DRAIN_35877_MEMBERSHIP_UNKNOWN`

`FRESH_BYGEL_DRAIN_PROVENANCE = NOT_PROVED`

SCORED_DRAIN_35877_PROVENANCE remains NOT_PROVED. CANONICAL_A_TIMING_INHERITANCE
and NATIVE_AUDIO_BOUNDARY remain PROVED premises. NEW_BALL_THRESHOLD_PROVENANCE
and EXPIRY_INTERLEAVING remain NOT_PROVED. DMO0 NOT CLOSED. DMO1 NOT STARTED.

This pass stops at the fresh entry-state transfer gate. No native trajectory
search was run: zero input scripts and zero logical calculations. This is neither
a failed input search nor an exclusion proof. The cost of a 35,877-calculation
replay is not the blocker.

## Selected correspondence actually verified

The extractor compares complete bounded instruction sequences and explicit
reviewed operand relocations, failing on changed operations, registers,
addressing, constants, destinations, or control edges. Returning callees remain
opaque; block identity does not close their semantics.

| Element | Canonical A file range | Demo file range | Classification |
| --- | --- | --- | --- |
| BYGEL1 unlit consumer | 0x27ed..0x283b | 0x27f6..0x2844 | RELOCATED-IDENTICAL |
| BYGEL2 unlit consumer | 0x2872..0x28bd | 0x287b..0x28c6 | RELOCATED-IDENTICAL |
| SETBALL consumer | 0xff0..0x1048 | 0xff4..0x104c | RELOCATED-IDENTICAL |
| SPRINGUP velocity/rotation consumer | 0x6165..0x61be | 0x61df..0x6238 | RELOCATED-IDENTICAL |
| Drain score-changed/LOOSING selection | 0x572..0x59c | 0x572..0x59c | RELOCATED-IDENTICAL |
| Scored LOSTBALL request after expiry guard | 0x5d2..0x5f1 | 0x5dd..0x5fc | RELOCATED-IDENTICAL |

Ranges are half-open. Demo's preceding expiry guard is a real difference; only
a proved drain before calculation 35998 could classify it as irrelevant. No
such witness was obtained, and the bounded request comparison excludes it.

BYGEL1 is (5,455)..(15,465); BYGEL2 is (284,455)..(294,465). Their actual records
bind demo callbacks at 0x27f6 and 0x287b, respectively. Both test light 39 and
their unlit paths set SCORECHANGED after adding BCD50030 to score. Tests compare
both callbacks, not just a candidate match count.

Eight five-word material parameter records and the full 2,560-word sine record
match at their reviewed relocated locations. The 44 lower collision-ring byte
samples for local SETBALL state (297,530,10,0,low) match individually. These are
selected input-record comparisons, **not consumed records from a witness**.
No flight was run and no claim is made about its dynamically modified masks,
flipper frames, material selection, callbacks, or out-of-bounds reads.

SETBALL's local coordinate/velocity/hold assignments and wait constant 80 match.
That establishes a local initializer transfer, not the complete reachable
fresh-game state. Full new-game/new-ball initialization, physics integration,
collision consumer closure and flight-specific projection were not promoted
after the entry-state gate failed.

## First unresolved entry-state obligation

The native factory creates a new Party Land Game with implicit uint16 `clock=0`.
Native gameplay advances it by 1030 and springControl consumes its low byte.
Native PresentationAudioSync advances audio only; it does not preserve a DOS
pre-game release counter in a newly constructed game.

The linked A and demo words are initially zero in the files, but both increment
SLUMP_COUNTERN by 1030 **before** the DEMOMODE test: A at 0x4500, demo at 0x4556.
The demo word is DS:0x34ec; the release consumer at 0x620d masks it by 255 and
subtracts it from velocity; rotation at 0x622d..0x6238 masks the same word by 15.
Thus pre-game counter history can affect both launch velocity and rotation.
This is a state-construction difference at the native bridge, classified as
TRANSFER-BLOCKING-DIFFERENCE until a reachable entry-state correspondence is
supplied. It is **not** a claim that A and demo binaries have different update
arithmetic, or that every possible demo entry differs from native entry.

For example, phases 0 and 6 produce different full-charge release velocities
(-5312 and -5318) and rotations (0 and 6). This arithmetic counterexample rejects
an inference; it is not a real gameplay prefix. File zero, implicit native zero,
and a modulo-128 counter recurrence do not establish the actual initial phase
or equality of all trajectory-relevant state. No counter was seeded by hand,
and no fast-forward was attempted. This audit does not assert complete exclusion
of alias writes to the counter.

Exactly one smallest unresolved fact is selected:

> Establish a reachable fresh demo NEW_BALL/SETBALL entry whose SLUMP_COUNTERN
> low byte equals the native fresh-session clock phase, through the real
> initialization prefix; neither raw PRG zero nor an authored counter seed
> establishes this correspondence.

This is the first gate, not a claim that all subsequent transfer obligations
are proved. No theorem projecting a native witness to demo semantics exists.
Membership of 35876, 35877 and 35878 remains UNKNOWN. No BYGEL or scored-drain
calculation index, whole-flight zero-aggregate/XXBALLE/INH_EFF state, or LOSTBALL
admission is reported as witnessed. The previously proved local zero predicate
and conditional 91-visit route are unchanged.

## Verification boundary

New tests cover selected physical records, both geometry bindings, both unlit
consumers, local SETBALL/release/drain equivalence, phase dependence and rejection
of a file-zero-to-entry-zero inference. Mutations change material, geometry,
collision sample, sine record, callback score destination, launch coordinate,
release phase consumer, SCORECHANGED guard, LOSTBALL request, counter update and
native clock initialization. They fail closed.

Actual witness replay, one-input replay mutation, whole-witness zero aggregates,
XXBALLE=false, INH_EFF=false, admitted LOSTBALL, exact drain index and reachable
neighbor assertions are NOT AVAILABLE because the transfer gate did not pass.
Passing local tests does not satisfy those acceptance conditions.

Current checks: 22 new tests PASS; whole demo research suite 349 tests PASS;
focused A/B/C/D regressions 15 top-level /
99 including subtests PASS; available strict-A physics/Party Land checks
11 top-level / 15 including subtests PASS. Accepted runs have zero SKIP.
The owner-local harness matches all 284 tracked Go/module files, with zero drift.
The initial broader strict-A invocation failed TestOriginalTrajectories because
its independent generator requires the absent private secondary fixture
`analysis/pf2-ball-locations.json`. That oracle is NOT AVAILABLE. Its failed log
is retained; no fixture was derived from production or manufactured to make it
pass. The available rerun explicitly excludes that unavailable case.

Whole-file and aligned nontrivial 4 KiB raw/hex/base64 payload-copy checks PASS
for this pass's three new research files and owner-local artifact. This sampled
check is not an all-substring proof. `git diff --check` PASS.

Artifact: `/private/tmp/pf-dmo0-fresh-bygel-drain.json`. Only research tools/tests
and this document are added. No production/runtime changes, internal/hosts/cmd
changes, .DS_Store changes, commit, push, tag, release or v0.1.3 changes. The
NEW_BALL_TASK survival, DS:0x36cd and PARTYFLASH/VISAKEYS suffix is untouched.
