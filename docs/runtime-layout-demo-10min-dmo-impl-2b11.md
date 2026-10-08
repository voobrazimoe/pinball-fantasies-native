# DMO-IMPL-2B11: scored drain, zero-aggregate bonus and first equality

2026-10-08. Owner approval: DMO-IMPL-2B11 ONLY.
Accepted predecessor: [2B10](runtime-layout-demo-10min-dmo-impl-2b10.md).
Branch main, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e` unchanged.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_SCORED_DRAIN = PARTIAL

DEMO_BONUS_CONTINUATION = PARTIAL

DEMO_FIRST_EQUALITY_NEW_BALL = READY

The real unexpired scored drain35877 and its complete selected zero-aggregate
continuation are implemented and pass. PARTIAL explicitly excludes expired
scored restart and nonzero bonus/match/hold-bonus alternatives. The unchanged
fresh TARGET completes35998, including expiry replacement, and continues through
36724; calculation36725 refuses the next unsupported NODOT/SHOW_HI_ETC consumer.
These are isolated native-reference results, not whole-DOS closure.

## Linked source admission

`check_demo_2b11_consumers.py` extends the2B10 checker; candidate loading still
pins the predecessor identities. No new research search is run. Historical
PLAND.ASM LOOSE_BALL/BALL_LOSTTS/_DEMOVER_CHANGE_PLAYER/NEW_BALL and FANTASIE.ASM
DOEFFECT/WHEN_NEW_BALL_RESET/DO_TASKS are the existing reviewed source basis.
Addresses below are linked demo file offsets, with DS addresses labelled.

- Common LOOSE_BALL stores retain their accepted order before electronics:
  source HOLDSTILL, SCREENFORCE2, parked ball15,47, ALLOWFLIP=false, cleared mode
  flags, LOOSING=true. The new scoped envelope projects SCREENFORCE2 to the
  shared scroll primitive's TargetRaster. No input, ball trajectory, timer or
  RNG value is supplied to force a bonus outcome.
- Scored guard0x5d2 reads expired at DS:34cf. Its true branch0x622 requests
  effect DS:06f1/file0x1a4a1. This entire restart remains unsupported and refuses
  before priority, cue, effect, matrix or task side effects. Earlier common
  ball-loss stores survive; no rollback is invented.
- Unexpired0x5dd clears CURRENT_PRIORITY;0x5e3..5e6 requests LOSTBALL DS:06d5.
  The linked record contains S_LOSTBALL pointer DS:0ca0, twelve zero score
  bytes, twelve zero bonus bytes and matrix DS:16a9/file0x1b459. Its jingle is
  position6/repeat1/priority255. The two zero BCD additions really execute;
  SCORECHANGED stays true, and UPDAT_INFOBAR resets NODOTCOUNT even for this
  zero-valued effect. No duplicate50030 award occurs.
- DOEFFECT0x5f14 onward orders jingle, information-panel handling, score addition,
  SCORECHANGED=true, bonus addition and guarded matrix installation. The linked
  comparison is **AL against byte [DS:0ca0]**: pointer low byte a0 differs from
  cue byte6. Thus the named LOSTBALL exception does not bypass INH_EFF on this
  actual linked call. Structural inhibition suppresses both cue and bonus
  matrix, preserves accounting and still admits the drain-tail sound task.
  SPECIALMODE was already cleared by LOOSE_BALL. Ordinary selected jingle
  admission uses MusicClock; initial priority is cleared, so it accepts.
- After effect,0x5e9 clears priority again,0x5ef sets LASTJINGLE=62, then
  ADDTASK inserts SOUNDRINNER (CS:02fc/file0x5fc). Its shared WAITLIST site is
  DS:36cb, delay5. The due body uses sample28/note18/channel3 (voice4) at
  DS:0c4d, via the shared silent Sound request; it then suicides.
- UPDATE_COUNTERS0x2db5 checks LOOSING and takes its existing return edge.
  The2B11 scored envelope honors this source guard before calling the shared
  counters primitive. Random and ordinary table counters stay fixed during the
  loss; demo timer still increments in ElectronicsCalculation. The predecessor
  structural envelopes retain their accepted behavior. No RNG adjustment is used.

Allocation remains first-free across50 slots; task scanning remains ascending,
identity-safe shared tablelogic.Run; WAITSYNCS remains shared compare-before-
increment. Allocation never zeroes a wait word. If allocation fails after an
admitted effect, that effect remains installed; refusal precedes task insertion.
Mandatory presentation/source operands are checked before their first effects.

## Actual zero-bonus control flow

The typed owned stream executes source handlers in order, retaining the shared
CLEAR/PRINT/WAIT timing and same-visit branch tail dispatch:

| Linked site | Selected operation |
| --- | --- |
| 0x1b459 | CLEAR4, five matrix visits |
| 0x1b45b | PRINT13 GETTING_SICK_TEXT at344, one visit |
| 0x1b461 | WAIT80, eighty visits |
| 0x1b465 | CLEAR4, five visits |
| 0x1b467 | JBCDZ BONUSSIFFRORNA →0x1b49f |
| 0x1b49f | JBCDZ CYCLONECOUNTERBCD →0x1b4cb |
| 0x1b4cb | JBCDZ HAPPY_HOUR_TOTAL →0x1b4eb |
| 0x1b4eb | JBCDZ MEGA_LAUGH_TOTAL →0x1b50b |
| 0x1b50b | JBCDZ BONUSSIFFRORNA →0x1b531 |
| 0x1b531 | KOLLA_XXBALL, actual XXBALLE=false |
| 0x1b533 | DEMOVER_CHANGE_PLAYER, then source HU_ tail |
| 0x1b535 | CLEAR4, followed by WAIT32000 |

All five zero tests execute, reading actual runtime fields. They do not consume
multiplier, FLORPA, bonus countdown, BEATEN_MATRIX, or extra-ball/match consumers:
the linked JBCDZ destinations skip those operations. Source multiplier handling
is therefore not needed on this branch, even when a structural zero case has
multiplier8. Score remains unpacked BCD `000000050030`, bonus remains zero.

A nonzero aggregate refuses its actual unsupported fallthrough before that
consumer's effects. Structural positive effect/accounting cases include a
preserved nonzero bonus under inhibited matrix admission; negative cases include
bonus1, cyclones1, happy1, mega1, XXBALLE=true and hold bonus. No unrestricted
nonzero bonus chain is claimed. Tail dispatch checks source cursor and the next
owned command before recursively entering the shared interpreter; a mutated
_CHANGE_PLAYER cannot reach a canonical consumer.

DEMOVER_CHANGE_PLAYER increments the actual mutable BALLSTEXT digits (DS:2385),
saves VARS_2_P_STRUC through the shared player-state primitive, inserts
NEW_BALL_TASK, and resumes its own matrix cursor. It does not advance the native
full-game player/ball domain. Hold-bonus restoration remains unsupported.

The reset SHOWPLAYERSTS stream at0x1b88e is distinct from the earlier linked
NODOT panel at0x1e2b3. Reset uses PLAYERSTEXT/336 and BALLSTEXT DS:2385/1684;
NODOT's accepted predecessor uses PLAYERSTEXT/340 and its own operand at DS:1e8b.
The candidate owns both typed streams and verifies the exact text/font inputs.
No canonical bonus, LOOSE_BALL, NEW_BALL or progression aggregate is called.

## Fresh input-only execution

TARGET remains `demoFixedInput`: Down35438..35459, Release35460;
d=calculation-35460, Left iff `(d+46)%52<8`, Right iff `(d+22)%30<21`.
Fresh factory score policy is the accepted2B10 verified volatile native policy.
No seeded tasks, changed launch, trajectory search, state restore or direct
ball/timer/RNG manipulation supplies this witness. Canonical callbacks remain
poisoned. Structural cases are separate tests and are not reachability evidence.

| Calculation | Completed candidate observation |
| ---: | --- |
| 35790 | BYGEL1, actual50030 award, bonus0, SCORECHANGED=true; CHECKHIGHSCORE nonqualification and late physics complete |
| 35876 | Actual late physics marks ball lost; score50030/bonus0/changed=true |
| 35877 | Real LOOSE_BALL, expired=false; LOSTBALL installed; SOUNDRINNER slot0 age1; matrix CLEAR4 remaining4 |
| 35882 | SOUNDRINNER age5 fires, resets its shared word and suicides |
| 35967 | All five zero tests execute; actual matrix cursor reaches producer; NEW_BALL_TASK slot0 inserted after scan at age0; saved player score50030 |
| 35968..35997 | Thirty scans increment the same DS:36cd word0→30, with one live parent and no reset/replacement |
| 35998 | Expiry admitted, then due NEW_BALL replaces it through real reset guards; entire calculation completes |
| 36724 | Last completed calculation on unchanged TARGET, score50030/bonus0; no second drain |
| 36725 | Electronics completed; NODOT inactivity count720 refuses SHOW_HI_ETC before presentation effects and late physics |

The independent ordinary canonical-A reference is constructed separately, never
used as a candidate fallback or state source. Original comparisons and saved
DMO0 rows cover every completed1..35876. Through35966 the comparison continues
for ball (normalizing source HOLDSTILL versus native capture Hold), clock,
score/bonus/changed, counter fields, lamps, music, tasks and matrix timing.

Source differences are explicit: native full-game drain freezes its camera and
keeps its ALLOWFLIP projection, while the linked source clears ALLOWFLIP and
consumes SCREENFORCE2 in continued calculations. After the linked demo producer
on35967, full-game A takes different progression. Every completed35877..35998
instead also compares the existing first-equality artifact's122 linked rows
(matrix cursor/op/time, shared waits, live slot occupancy, timer/guards) and its
independent native task-reference ball/score/aggregates/lights/flipper projection.
The artifact producer is not rerun. Native probe-only spring-valid/last-area
fields are not substituted for the actual linked reset stores. After35998,
each completed calculation checks completion count, timer/clock/counter
recurrence, retained score/bonus and single-drain count; it makes no new full
canonical-A or DMO0 research claim for that suffix.

## First equality, full calculation order

Electronics increments35997→35998, sets expired/HOLDSTILL, requests the actual
position13/repeat0/priority255 cue and installs expiry0x1ba17. The due task entry
records that real expiry CLEAR4, cursor0x1ba19, remaining5 and admitted cue
before reset; it is not inferred from a final pointer.

DO_TASKS visits slot0 with DS:36cd=30, resets the word and executes NEW_BALL.
Reset clears the task list and wait words. Actual PARTYFLASH=false and
VISAKEYS=false select SHOWPLAYERSTS0x1b88e, replacing expiry without saving its
cursor. The source reset/table/player/ball operations and all three child
insertions execute. The spring jingle request is rejected by the still-current
priority255; its repeat-store behavior is retained. LASTJINGLE becomes0.

Slot0 now contains SOUNDNEWBALL, not the old parent. The scan resumes at slot1:
SETBALL and SOUNDBRICKUPP become age1; SOUNDNEWBALL stays age0. Matrix processing
visits the **current** reset CLEAR4, leaving remaining4 and linked cursor
0x1b890. Late physics respects source HOLDSTILL; calculation35998 completes.
Score50030/player save survive, SCORECHANGED=false, BALL_DOWN/LOOSING=false,
expired=true. No expiry cursor is restored. Structural PARTYFLASH/VISAKEYS
alternatives prove that replacement follows guards, rather than expectation.
No repeated equality101534 or QUIT102583 continuation is attempted.

## Fallible boundary and scope

Next fresh failure:

```
UNSUPPORTED_DEMO_TRANSITION producer=NODOT
state/guard=SHOW_HI_ETC consumer=NODOT phase=task/matrix
calculation=36725
reason=inactivity high-score presentation not admitted
last completed calculation=36724
```

Failure is sticky across sync and direct candidate entries. Due NEW_BALL refuses
missing reset/child/panel consumers before WAITSYNCS reset; later task, matrix
and physics work stop. Previous legal score/effect/stores remain. Structural
allocation failure likewise retains the already admitted LOSTBALL effect.
No canonical fallback or rollback is introduced.

Only tagged/test-only dmoimpl1 candidate, its tests/checkers and this report are
changed for2B11. No production profile, frontend, persistence, arbitrary modes,
unrestricted bonus, new scheduler, delay, search, original fixtures, DMO0
research or DMO1. No .DS_Store/v0.1.3 change, commit/push/tag/release, or2B12 work.

## Validation

Private logs: `/private/tmp/pf-dmo-impl2b11-*`.
Private source-only harness: `/private/tmp/pf-dmo-impl1-g75qk5z5`.
Final results follow.

| Check | Result |
| --- | --- |
| Fresh fixed TARGET, original prefix1..35876, actual scored drain and saved122-row equality suffix | PASS |
| Two repeated fresh equality runs; continued TARGET through36724 and sticky refusal36725 | PASS |
| Zero bonus, multiplier bypass, nonzero/inhibited accounting, expired guard, priority, allocation, task visits and suicide | PASS, structural cases separate |
| Missing due consumers before WAIT reset, PARTYFLASH/VISAKEYS guards, program replacement, mutated canonical tail and direct entries | PASS |
| DMO-IMPL-1..2B10 retained plus2B11 TestDemo | PASS,77 top-level tests, no skips |
| Linked source admission and adversarial mutation/no-fallback suite | PASS,3 tests,109 mutation subcases |
| Python demo regression suite with existing historical/private inputs | PASS,594 tests, no skips |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS,153 top-level tests;5 optional skips |
| A/B/C/D focused datalayout/frontend | PASS,21 top-level tests, no skips |
| Tagged vet | PASS |
| Windows amd64 executable and engine build | PASS |
| macOS c-shared engine build | PASS |
| Broad macOS compile-only | Known baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL: gameshow53/speeddevils36/stones47, native content |
| Fixture-free TestDemo | PASS available31 tests;46 private-backed SKIP/NOT AVAILABLE |
| Public source checker and diff whitespace | PASS |
| Scoped payload scan | PASS,10 source/report files and2 built artifacts |
| OriginalTrajectories/Stones pf10 captures | NOT AVAILABLE; no fixture generated |

Payload checking covers whole private runtime files plus nontrivial aligned4KiB
raw/hex/base64 samples across92 private files (1677 sample blocks). It is a
sampled check, not an all-substring proof. Originals, saved witnesses and test
logs remain external; no private payload or private-file hashes enter this report.
Build artifacts contain no candidate unsupported-transition diagnostic.

Reproduction in the existing private source-only harness, with legally supplied
originals and the existing research Capstone environment:

```
PF_10MIN_DEMO_DATA=<pinned-demo> \
PF_DEMO_RESEARCH_WITNESS=<saved-drain-json> \
PF_DEMO_COLLISION_WITNESS=<saved-first-equality-json> \
 go test -tags dmoimpl1 ./internal/partyland ./internal/physics \
 -run '^TestDemo' -count=1 -v
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> \
 python3 -m unittest discover -s tools -p test_check_demo_2b11_consumers.py -v
```

The initial implementation checks exposed the difference between the two source
panels and native drain's stopped counter/camera behavior. Exact linked
UPDATE_COUNTERS/LOOSING and reset-panel operands resolve those comparisons;
inputs and trajectories were not adjusted. Early private logs from incomplete
iterations are superseded by the final logs. The final failure is the explicit
36725 unsupported presentation, not an execution or regression failure.
