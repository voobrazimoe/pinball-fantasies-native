# DMO0 deterministic fresh BYGEL → scored drain

Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
Research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

This pass obtains a real input-only fresh native-reference replay and projects
its consumed path to the official `pinbfan` demo. It does not establish a
physical DOS trajectory oracle or close DMO0. The previously accepted timing,
audio, jitter and fresh-entry premises are retained. Existing earlier research
artifacts and their historical conclusions are preserved.

## Concrete witness

Construct `partyland.New(physics.DecodePartyLand(A), A)` and apply
`settings.Legacy()` as in the accepted A reference. The isolated harness copies
the tracked source and verifies every file present at the production base
against that base before adding passive observers. No ball placement, direct
Release call, timer seed, state edit or tilt is authored by the replay.

The script uses only native `Down`, release edge, and left/right held controls:

- Calculations 1–35437: no inputs.
- Calculations 35438–35459: Down held (22 charge visits).
- Calculation 35460: release edge.
- From calculation 35460 through drain, put `d = calculation - 35460`.
  Left is held iff `(d+46)%52 < 8`; Right iff `(d+22)%30 < 21`.
- All other inputs are false. Stop the replay at the drain.

| Boundary | Calculation | Selected state |
| --- | ---: | --- |
| Actual SPRINGUP release | 35460 | charge 22, clock low8 24; source velocity `-166*22-24`, rotation 8 |
| CLOSE1 | 35481 | actual fresh chute exit |
| BYGEL9 | 35524 | previous area CLOSE1, so no loop award |
| BYGEL11 | 35545 | previous BYGEL9, reverse inhibit counter nonzero; score/aggregate path bypassed |
| Unlit BYGEL1 entry | 35790 | light 39 false, score before callback 0; score-only BCD50030 |
| Actual scored drain | 35877 | score 50030, SCORECHANGED true, four totals zero, XXBALLE false |
| LOSTBALL request/admission | 35877 | INH_EFF false, SPECIALMODE false; accepted matrix BALL_LOSTTS |

The exact release boundary is observed inside the real spring callback, after
the two early physics passes and before the late pass. Every logical calculation
is recorded, including the actual drain ball before LOOSE_BALL performs its own
source-authentic held-ball assignment. The clock is encoded as `u16(1030*n)`;
low8 is `(6*n)%256`. Inclusive run ranges and dictionaries preserve all ordered
physics observations, inputs, ball/velocity/rotation, material/region identities,
events, score, aggregate predicates, flags, matrix state and relevant task state.

Two full fresh replays produce identical traces. Removing precisely the first
Down visit changes charge 22 to 21 and changes/invalidates the witness; no game
state is edited. The accepted historical PF6 failing oracle is not used as
evidence. The new trajectory depends on a concrete replay, paired consumers and
source-derived arithmetic.

## Consumed-path projection

`tools/dmo0_replay_correspondence.json` freezes per-site operand correspondences
for named bounded consumers. The auditor verifies operations, instruction sizes,
registers, addressing, constants and destinations. A changed operand cannot be
absorbed by recomputing the map. These are source-consumer relocations, not an
indirect-target admission list or a global writer-closure claim.

The actual route uses BYGEL12/BYGEL28, CLOSE1, BYGEL9's rejected-loop guard,
BYGEL11's inhibited-reverse prefix and counter-clear exit, TOUCHER/ENABLETOUCHER,
then unlit BYGEL1 and the scored LOOSE_BALL/LOSTBALL branch. No duck award, hole,
loop award, aggregate effect or match transition occurs. The four totals and
XXBALLE predicates are asserted at every calculation through drain.

The projection compares the integer collision/response/move/flipper consumers,
ramp/target/level/area consumers, high-resolution low-angle initialization,
adjusttable and flipper initialization, actual callback binding records and
score arithmetic. Materials 2/3/6/7 are consumed. Full bounded sine, collision,
material and ramp maps, flipper numeric descriptors and collision frames,
initial duck mask patches, level/area/target rectangles and ring-coordinate data
match at the explicitly linked relocated locations. Complete map equality also
covers negative collision tests, not just observed positive contacts.

The accepted native A `bit()` policy treats samples outside the map as empty.
Such samples occur on this route and are explicitly recorded. Demo projection
uses that same bounded-map reference policy and equal dimensions/data. This is
not a claim that this Go flight reproduces physical DOS adjacent-segment reads;
neither historical DOS timing nor whole physics closure is promoted.

CLOSE1 contains one extra demo `VISAKEYS=false` store at file 0x2596, classified
DEMO-DIFFERENT-BUT-IRRELEVANT for this prefix. Its source consumer belongs to the
deferred new-ball task. The extra store does not change ball, input spring,
flipper, region, score, aggregate or LOSTBALL admission arithmetic on this path.
The demo expiry guard is also irrelevant on this concrete prefix: drain 35877
precedes the inherited first equality 35998. Its pre-electronics check sees
expired=false. No native timer is seeded.

At admission LOOSE_BALL has cleared SPECIALMODE and reset priority. INH_EFF is
false. The linked zero-arithmetic LOSTBALL effect admits its jingle and selects
normal demo bonus program **0x1b459**. No transfer-blocking difference was found
on the selected route. This is path-specific correspondence, not closure of
unvisited physics paths or arbitrary callback contexts.

## Delay data and finite search

First simple witness: Down visits 1–60, release 61, right held from release:
unlit BYGEL2 255 → scored drain 349; score 50030, totals zero, LOSTBALL admitted.
It preceded exact-index targeting.

The target family's shorter script uses charge 22 and the same periodic inputs:

| Release | BYGEL | Drain | Valid zero-aggregate scored BYGEL drain |
| ---: | ---: | ---: | --- |
| 132 | 462 | 549 | yes |
| 260 | 590 | 677 | yes |
| 388 | 718 | 805 | yes |
| 1156 | 1486 | 1573 | yes |
| 16516 | 16846 | 16933 | yes |
| 35458 | — | 35645 | no |
| 35459 | — | 36458 | no |
| 35460 | 35790 | **35877** | **yes** |
| 35461 | — | 35740 | no |
| 35462 | — | 36249 | no |

Every row is a full replay from fresh construction. The 128-period clock
arithmetic suggested a candidate; it did not replace the long replay or prove
periodicity of every state. Random counters, lamp flash phases, audio/matrix
state and task state are not assumed frozen or globally periodic.

Separately, holding Down for all 35516 pre-release calculations was tested.
The ball remains in the chute, charge remains 32 from visit 32, electronics run
35516 times, score/SCORECHANGED remain zero/false, four totals remain zero,
XXBALLE and INH_EFF remain false. Tasks/lights/modes are observed in checkpoints.
Fresh deterministic releases after long holds also give real valid routes:
61 → 255 → 349, 189 → 391 → 479, and 35517 → 35719 → 35807. These results
are finite dynamic evidence; no arbitrary-N chute theorem is claimed.

Search coverage is recorded exactly in the artifact: 511 no-flipper scripts;
1716 fixed/periodic-flipper scripts; 601 directly replayed late constant-right
scripts; 28416 right-interval scripts; 63360 charge/right-interval scripts;
and 8610 deterministic generated periodic-control trials (157 valid selected
scored BYGEL drains). The latter stops at the successful candidate; aggregate
producer prefixes are pruned as invalid. The generated controller uses seed 1
only to design player inputs, never to seed game state.

35876 and 35878 remain UNKNOWN; nearby input variations did not land at them.
No exclusion theorem follows. Exact 35877 membership requires no interval or
general controllability theorem because the concrete projected replay suffices.

## Existing suffix and verification

Reuse the already proved zero-aggregate/XXBALLE=false branch: 91 matrix visits,
producer after 35967, conditional NEW_BALL_TASK firing at 35998. Its arithmetic
was not rediscovered. Actual slot, shared DS:0x36cd initial/unique age, task
survival, PARTYFLASH and VISAKEYS remain the next pass. NEW_BALL_TASK collision
reachability is not promoted.

Owner-local artifact:
`/private/tmp/pf-dmo0-deterministic-bygel-drain.json`.
Reproduce with the existing three private input environment variables and:

```sh
python tools/audit_10min_demo_deterministic_replay.py
python -m unittest discover -s tools -p test_audit_10min_demo_deterministic_replay.py
```

The Python environment must provide research-only Capstone, and `--go` may
select the available Go toolchain. The separate Go template creates no package
or runtime extension inside production `internal`, `hosts` or `cmd`.

New-pass Python tests: 24 PASS; deterministic replay/one-input mutation and
long-hold/release-variation Go tests PASS. Full private demo research suite:
414 PASS. Available canonical-A reference checks: 27 test/subtest PASS and the
one reproduced known PF6 baseline FAIL (2311040 vs 2300000). OriginalTrajectories
is NOT AVAILABLE and was not executed or manufactured.

Focused A/B/C/D regressions: 97 test/subtest PASS, zero SKIP/FAIL.
`git diff --check` PASS. The pre-existing tracked diff remains byte-for-byte
unchanged. Production `internal`, `hosts`, `cmd`, module files and .DS_Store
are not edited; no staging, commit, push, tag, release or v0.1.3 changes.
Whole-file and raw/hex/base64 sampled payload checks PASS against 249 private
files, 11 decoded SDR modules and 48898 distinct nontrivial aligned 4 KiB
samples. This is a sampled check, not an all-substring theorem.

FRESH_BYGEL_DRAIN_PROVENANCE = PROVED

DRAIN_35877_REACHABLE

SCORED_DRAIN_35877_PROVENANCE = PROVED
