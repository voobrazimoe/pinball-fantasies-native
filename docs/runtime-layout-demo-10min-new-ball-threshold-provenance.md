# DMO0 NEW_BALL_TASK threshold-age provenance

Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

**NEW_BALL_THRESHOLD_PROVENANCE = NOT_PROVED.**
**EXPIRY_INTERLEAVING = NOT_PROVED.**
**DMO0 NOT CLOSED. DMO1 NOT STARTED.**

This pass binds one short normal-bonus branch route and its relative offset to
private linked operands. It does not establish either a real collision witness
or an exclusion invariant. Canonical-A timing inheritance and native audio
boundary remain premises. No production changes or historical artifact rewrites.

## Predecessor and offset

LOOSE_BALL distinguishes SCORECHANGED=false (unscored PARTY_ON continuation)
from scored drain. The demo scored branch checks expired before electronics and
uses LOSTBALL while it is false: file `0x5d2`, `0x5e3/0x5e6`. Effect record
DS `0x6d5` names matrix `0x1b459`. Effect admission still requires jingle success
and no INH_EFF/SPECIALMODE suppression; a record pointer is not admission proof.

A linked branch candidate with BONUSSIFFRORNA, CYCLONECOUNTERBCD,
HAPPY_HOUR_TOTAL and MEGA_LAUGH_TOTAL all zero follows:

`1b459 CLEAR4 -> 1b45b PRINT13 -> 1b461 WAIT 80 -> 1b465 CLEAR4 ->
1b467 JBCDZ -> 1b49f JBCDZ -> 1b4cb JBCDZ -> 1b4eb JBCDZ ->
1b50b JBCDZ -> 1b531 KOLLA_XXBALL -> 1b533 DEMOVER_CHANGE_PLAYER`.

The branch targets skip CLEAR4 at `1b4c9`, `1b4e9`, `1b509`; counting these
would incorrectly add fifteen visits. With XXBALLE=false, KOLLA_XXBALL tail
calls the producer without an additional visit. This route skips FLORPA,
BEATEN_MATRIX and the high-score/match alternatives. No animation is traversed.
The actual values selecting this route have **not** been proved reachable.

Under the inherited native semantics, CLEAR4 costs five admitted matrix visits,
PRINT13 costs one, WAIT costs eighty, and the zero branches tail dispatch on the
same visit. Total: **91 matrix visits**. DO_MATRIX installs the first routine
during drain before that update's electronics and post-task matrix work. If D
is the counted calculation in that drain update:

`producer_calculation = D + 90 + H`,

where H counts budget-rejected matrix updates through producer dispatch, assuming
no intervening program replacement. For native Sync's budget=true schedule H=0.
The general relative formula is `D + sum(traversed routine visit costs) - 1 + H`;
this pass does not enumerate costs for every nonzero/countdown/XXBALLE route.
It does not assert a universal finite K set or a reachable congruence.

The producer saves/restores BX (`0x73e`, `0x788`), unlike the separately studied
_2_DEMO_MODE handler. Its HU_ tail therefore uses the matrix cursor after adding
the task; it does not call the following task slot through clobbered BX.

## Exact threshold equation and missing predecessor

The zero-age conditional suffix needs producer after 35967. For this branch:

`D + 90 + H = 35967`, hence `D = 35877 - H`.

Consistency correction: the earlier “above zero” wording was a report typo.
The extractor and JSON already required all four aggregates zero; the 91-visit
arithmetic was for that predicate and is unchanged.

For H=0 the selected earlier unresolved fact is exactly:

> Does a real normal-game input prefix reach an admitted scored-drain LOSTBALL
> installation at calculation 35877, with all four aggregates == 0 and
> XXBALLE=false, on the all-budget-true native schedule?

No real prefix or structural delay lemma joined to this drain was established.
Launch-chute continuation is only a possible phase-adjustment candidate; neither
"the player can wait" nor a supplied drain index proves it. Pause contributes
no admitted ElectronicsCalculations. The linked candidate is not an authored
state labeled reachable. No conclusion about other normal-bonus branches follows.

## Conditional suffix boundary

First-free insertion chooses `s=min{i: TASKLIST[i]=DUMRET}`, slots 0..49 at
DS `0x3417+2*s`. Without the real predecessor's list, **actual s is UNKNOWN**.
The 50-slot scan and insertion operands are pinned. Producer matrix work follows
the task scan, so a freshly inserted task first visits on a later scan.

DS:0x36cd belongs to the NEW_BALL_TASK wait callsite. DOADDTASK does not zero it.
A preceding NEW_BALL reset zeros WAITLIST, including this word, but its actual
last reset and absence of a surviving duplicate are not established for a real
prefix. Slot movement/reuse does not transfer or initialize age. Another instance
using the same callsite can increment/reset the word on the same calculation.
The local writes are mismatch increment, equality zero, and RESET_WAITLIST zero.
No concrete reachable suffix exists here on which to restrict every writer; no
global writer closure is claimed.

Given zero initial word, one surviving instance, and one visit per calculation:
visits 35968..35997 leave ages 1..30; visit 35998 fires and zeros the word. Producer
35966 fires on 35997; producer 35968 has age 29 before the equality scan, age 30
after it, and would fire on 35999 if the suffix survives. These are the already
proved arithmetic premises, preserved as regressions rather than reachability.

PARTYFLASH is cleared by CLOSE1 (`0x2591`) after the actual BYGEL12 predecessor;
unscored PARTY_ON_TASK1 sets it (`0x5c9`) before NEW_BALL. VISAKEYS is set by
WHEN_NEW_GAME_RESET (`0x3bae`) and cleared by the non-PARTYFLASH VISAKEYS branch
of WHEN_NEW_BALL_RESET (`0x3b1a`). Neither guard is established false at threshold
from an actual gameplay prefix. They are not inferred from initialized bytes.

Conditional firing would reset tasks/waits before testing these guards and could
replace expiry with SHOWPLAYERSTS, which has no QUIT. This pass supplies no real
counterexample to atomic expiry, no exclusion, and no eventual termination proof.
Slot survival, shared-word uniqueness and guard provenance remain downstream
obligations once the selected predecessor fact is established.

## Verification

New tests cover the relative offset, post-scan production, three threshold phases,
conditional slot survival, reset clearing, shared initial age/interference/duplicate
instances, and guard consequences. Pinned mutation tests reject changed predecessor
and effect edges, branch operands, WAIT, scan, reset, shared word, and both guard
writers. Budget/duration mutations change the candidate equation; tests explicitly
retain UNKNOWN reachability. A proved reachable adjustment/congruence mutation
test is **NOT AVAILABLE** because no such lemma was established; these authored
candidate tests must not be reported as a substitute.

Counts and logs from this pass are recorded in the new JSON artifact; prior-pass
counts remain separate. All outputs contain metadata, not commercial payload.
Machine-readable artifact: `/private/tmp/pf-dmo0-new-ball-threshold-provenance.json`.

Current-pass results: 27 new tests PASS; 295 entire demo research tests PASS;
strict A/native ordering 24 top-level / 347 including subtests PASS;
A/B/C/D 15 top-level / 99 including subtests PASS. All have zero SKIP.
Verified current owner-local harness: 284 tracked Go/module files, zero drift.
Initial older-harness/package attempts are excluded from accepted coverage.

`git diff --check` and whole-file/aligned nontrivial 4 KiB raw/hex/base64 payload
checks PASS. This is a sampled copy check, not an all-substring proof.
No commit, push, tag, release, v0.1.3 or production/runtime changes.
