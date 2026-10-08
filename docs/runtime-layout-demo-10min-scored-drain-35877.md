# DMO0 scored-drain phase provenance at calculation 35877

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

**SCORED_DRAIN_35877_PROVENANCE = NOT_PROVED.**

CANONICAL_A_TIMING_INHERITANCE and NATIVE_AUDIO_BOUNDARY remain PROVED premises.
NEW_BALL_THRESHOLD_PROVENANCE and EXPIRY_INTERLEAVING remain NOT_PROVED.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

## Consistency correction

The previous report's “above zero” was a report typo. The inspected owner-local
JSON's route condition and smallest remaining dependency both require zero.
The extractor's UNKNOWN string and route condition also require zero. Its linked
JBCDZ targets describe that branch. Previous tests pinned 91 visits but lacked an
explicit regression rejecting the positive-aggregate predicate. This pass adds
that regression and corrects the report; the previous JSON is unchanged.

The retained predicate is BONUSSIFFRORNA == 0, CYCLONECOUNTERBCD == 0,
HAPPY_HOUR_TOTAL == 0, MEGA_LAUGH_TOTAL == 0, XXBALLE == false.
The existing conditional result remains 91 admitted matrix visits, producer at
D+90 with H=0, and D=35877 for producer 35967. The suffix was not rediscovered.

## Phase adjustment first

SPRINGIT at file `0x6166` tests held Down, caps SPRINGPOS at 32 (`0x6170/0x6175`),
and otherwise increments it (`0x617a`). SPRINGSTEEN calls it at `0x61db`.
Thus additional held-Down visits preserve an already saturated charge. This is
a control-state recurrence, not an active-ball/gameplay recurrence.

SPRINGUP reads charge at `0x61df`. On the high-resolution release path,
`0x6206/0x6209` computes -166*charge; `0x620d..0x6213` subtracts
SLUMP_COUNTERN & 255. Rotation also uses SLUMP_COUNTERN & 15 (`0x6233`).
Consecutive counter values therefore give different release velocities even
when charge is stable. Counter values separated by 256 preserve those release
operands, but do not prove ball position, counters, lamps, task state, effect
inhibition or callback ordering recur. No calculation-index congruence follows
from this counter period. The local tests deliberately reject that upgrade.

Ordinary admitted electronics continue under the inherited Sync schedule while
inputs charge the spring; pause is excluded. Neither charge saturation nor a
chute rectangle proves the ball can stay in that rectangle for arbitrary N.
No preserved active-ball recurrence joined to a real selected drain was proved.
Consequently this pass establishes no interval, consecutive reachable N, drain
congruence, fixed phase or exclusion bound. It does not simulate 35,877 updates.

## Actual predecessor consumers

The necessary graph is fresh TABLE1 -> new-game/new-ball resets -> real ball
motion and release -> score-only lane callback -> actual drain -> LOOSE_BALL ->
admitted LOSTBALL. The linked consumers are established; the geometry/input
edges from fresh initialization through the lane and drain remain unjoined.
No coordinates, score flag, drain index or task state are authored as reachable.

BALLCODE's sc_move sets BALL_DOWN when SC_Y >= banh. The linked DO_PHYSICS
checks BALL_DOWN at `0x5d36`, then LOOSING at `0x5d40`, and calls LOOSE_BALL at
`0x5d4a`. LOOSE_BALL also has the PUKEFORBIDDEN early return (`0x582..0x589`);
an actual predecessor must avoid it. It clears SPECIALMODE at `0x572` and
eventually sets LOOSING at `0x58c`. SCORECHANGED==false selects PARTY_ON
(`0x592/0x597`); the scored branch checks the expired flag (`0x5d2`) and selects
LOSTBALL (`0x5e3/0x5e6`) only before expiry. These branches are regression tested.

A useful narrower candidate is an unlit side lane. The typed BYGEL1 region
record at `0x1abcb` binds rectangle (5,455)..(15,465) to callback `0x27f6`.
The callback tests light 39, then, on its unlit path, adds BCD50030 to SIFFRORNA
and writes SCORECHANGED=true (`0x2815..0x281e`). BYGEL2 has the corresponding
score-only path. Its sound and info-bar work do not add a bonus award. This
demonstrates that score-changed is not a logical implication of nonzero bonus;
it does not demonstrate that a fresh ball physically reaches this path.

## Aggregate and XXBALLE provenance

New-game RESET_VARS2 zeroes BONUSSIFFRORNA and CYCLONECOUNTERBCD
(`0x354..0x374`). RESET_VARS zeroes HAPPY_HOUR_TOTAL and MEGA_LAUGH_TOTAL
(`0x3ac..0x3c1`). The unlit lane's local transfer preserves all four values.
LOSTBALL's own score and bonus fields are zero. The latter is checked directly,
so effect admission cannot silently introduce a bonus into the selected branch.

Relevant producers to avoid on the candidate flight are effect bonus fields and
bonus arithmetic/restores; BYGEL13 normal/5x cyclone additions and player restore;
ADDHAPPY under HAPPY_HOUR; ADDMEGALAUGH under MEGA_LAUGH. PLAND's bonus tail can
also fold mode totals/cyclone awards into bonus, but that later suffix is outside
this pass. No universal nonzero-aggregate invariant or complete producer-domain
closure is asserted. Zero preservation is proved for the local lane/drain
transfer, not for a complete physically reachable play prefix.

The new-game table reset writes XXBALLE=false at `0x331`. The relevant true
writer is LET_HIM_MATCH. A first-ball lane/drain candidate avoids that continuation
and neither local lane nor drain sets XXBALLE. Absence of a match transition along
an actual input prefix is not established by the initializer alone. Global match
and shoot-again machinery was not reopened.

## Effect admission

LOSTBALL's record DS:0x6d5 refers to jingle DS:0xca0, decoded as (6,1,255),
and normal bonus matrix file `0x1b459`. The scored drain writes current priority
zero immediately before the request (`0x5dd`). The jingle comparison accepts
priority equality and rejects only lower priority (`0x5fc5/0x5fc9`); priority 255
cannot lose to another byte priority, even if the zeroing is hypothetically
removed. Under the inherited synchronous call ordering, no other effect request
intervenes between that write and admission.

DOEFFECT preserves the admission carry across score arithmetic (`0x5f7b/0x5f7c`)
and separately rejects INH_EFF or SPECIALMODE (`0x5f89/0x5f93`) before DO_MATRIX.
SPECIALMODE is cleared by this drain. New-ball reset clears INH_EFF (`0x3adc`),
but no whole-flight exclusion of its setters was established. Thus matrix
admission is conditional on INH_EFF=false in the same real predecessor; the
record pointer and maximum priority do not discharge that condition.

## Exact index and remaining fact

For calculations 35876, 35877 and 35878, reachable scored-drain membership is
UNKNOWN. Their conditional producers remain 35966, 35967 and 35968 respectively.
Neither `35877 ∈ reachable_scored_drain_indices` nor its negation is established.
No wall-clock inference or new threshold class is justified.

Exactly one remaining predecessor fact is selected:

> Establish a geometry-derived fresh-session launch/outlane/drain transfer
> relation, including its admitted calculation indices, that reaches an unlit
> BYGEL1 or BYGEL2 without an aggregate producer or INH_EFF setter before scored drain.

This restricts the earlier general scored-drain question to a source-identified
score-only predecessor family. Its actual input/geometry transfer is still
unknown; the tests do not make that family reachable. Task-slot survival,
DS:0x36cd uniqueness and PARTYFLASH/VISAKEYS remain downstream and untouched.

## Verification

Current pass: 32 new tests PASS. Entire current demo research suite: 327 tests
PASS. Strict A/native ordering: 24 top-level / 347 with subtests PASS. Focused
A/B/C/D: 16 top-level / 100 with subtests PASS. Accepted runs have zero SKIP.
The owner-local Go harness matches all 284 tracked Go/module files, zero drift.
An initial ordering invocation named nonexistent internal/runner; it is excluded
from accepted coverage. The corrected run uses internal/source and presentation.

Coverage includes the consistency predicate, preserved 91-visit equation,
scored/unscored/inhibited drain selection, local zero preservation, XXBALLE
initializer, admission/suppression mutations, charge saturation and a mutation
breaking that local recurrence, jitter dependence, and conditional index neighbors.
Actual gameplay phase-lemma mutation and reachable-index neighbor checks are
NOT AVAILABLE: no such lemma or real trajectory was established. Local authored
transfers and metadata tests are explicitly insufficient to prove reachability.

`git diff --check` and raw/hex/base64 payload checks PASS. The payload check covers
whole-file copies and aligned nontrivial 4 KiB samples; it is not an all-substring
proof. Artifact: `/private/tmp/pf-dmo0-scored-drain-35877.json`.
Only this pass's research files and the preceding report typo were changed.
No production/runtime changes, historical unrelated artifact rewrites, .DS_Store
changes, commit, push, tag, release or v0.1.3 changes.
