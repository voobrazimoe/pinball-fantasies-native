# DMO-IMPL-2B10: scored NODOT / CHECKHIGHSCORE continuation

2026-10-08. Owner approval: DMO-IMPL-2B10 ONLY.
Accepted predecessor: [2B9](runtime-layout-demo-10min-dmo-impl-2b9.md).
Branch main, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e` unchanged.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_CHECKHIGHSCORE_CONSUMER = READY

DEMO_FIRST_SCORED_CALCULATION = READY

READY covers the actually selected nonqualification branch under the authorized
volatile native factory policy. Reward dispatch remains unsupported. The fresh
candidate completes calculation35790, with the actual BYGEL1 award50030 and
late physics, then completes through35876. Calculation35877 reaches the existing
LOOSE_BALL scored-drain refusal. It does not complete35877 or execute LOSTBALL /
bonus. This is a native-reference candidate result, not whole-DOS closure.

## Bounded linked/source review

`check_demo_2b10_consumers.py` extends predecessor operand admission. All addresses
below are linked demo file offsets unless explicitly labelled CS/DS.

- NODOT entry file0x56ac (CS:53ac); ONLY_SCORE callsite0x573e calls
  CS:5456, CHECKHIGHSCORE entry file0x5756. Entry checks keyboard enable,
  NODOTCOUNT==720, AFTER_CHEAT and DEMOMODE. The ordinary playing path checks
  I_UTSKJUT and BEHOVS_PROVAD; if needed, DO_SPEC_MATRIX(SHOWPLAYERSTS) runs before
  ONLY_SCORE. These predecessor branches are retained; their unadmitted lifecycle
  paths are not implemented by this slice.
- CHECKHIGHSCORE tests I_UTSKJUT at DS:34df, SPECIALMODE at DS:34e1,
  ALREADY_BEATEN at DS:34f0, each against true=ff. Any true guard returns before
  reading high-score digits. PUSHA/POPA preserve general registers.
- DI=DS:0016 identifies HI_SCORE_LIST, SI=DS:46b5 identifies SIFFRORNA.
  CX=12, BX=0; each unsigned byte comparison reads current score against the
  corresponding top-score digit, most significant first. JB returns without
  qualification; JA selects reward; all twelve equal also return without reward.
  Names and lower three score records are not comparator inputs. No bonus or
  SCORECHANGED read participates in this comparison. CPU CF/ZF drive JB/JA;
  there is no returned CPU flag contract used by the following NODOT code.
- The nonqualifying return0x5792..5793 performs no state writes. Source NODOT
  then increments NODOTCOUNT at0x5741, draws SIFFRORNA at BX=200 through PEKOR,
  and returns SI=0 at0x5752. Existing same-visit matrix dispatch continues.
- The qualifying branch0x5794 sets ALREADY_BEATEN, calls
  DO_MATRIX(BEATENTS=DS:1413), then sets BEHOVS_PROVAD=true. It does **not** call
  `_BEATEN_MATRIX`. BEATENTS's first linked handler is `_DOBEATEN`, file0x3283;
  its reward producer includes the extra-ball byte increment at0x328f.
  The separate `_BEATEN_MATRIX` body starts0x57a6 and omits chute/special guards;
  it is not part of the selected call chain. No reward is represented as a no-op.

The review uses existing historical FANTASIE.ASM NODOT/CHECKHIGHSCORE/
_BEATEN_MATRIX and PLAND.ASM default declarations, plus the concrete linked
operands. Checker mutations cover comparator operands, guards, unsigned branches,
loop/return, reward producer, NODOT continuation, padding and factory seed.
It does not reopen DOS handler, stack, alias, placement or IRQ research.

## Comparator source, policy and lifetime

The admitted source is `verified-native-factory-volatile`, owned by this fresh
candidate. The loader first pins the existing demo/canonical identities and
predecessor correspondences, verifies the new operands, then copies the first
12 compiled default digits from linked DS:0016 into a candidate-owned Decimal.
It never takes top score from canonical Game.highScore or from a persistent file.

The four compiled default scores are50,000,000 /25,000,000 /10,000,000 /5,000,000.
The checker reconstructs the existing native Defaults(1) records and verifies
their identity against the pristine TABLE1.HI entry in the existing inventory,
then checks all64 linked compiled seed bytes. The independent frontend factory
seed tests also run. No TABLE1.HI content is opened by admission or execution.
The inventory is metadata; no fixture is created.

The top remains unchanged for the candidate's lifetime. Every comparison that
passes the three source guards requires the admitted source marker, a present
valid Decimal, and the verified factory top. Missing, persistent/unverified,
changed or invalid state refuses before display/matrix/score side effects.
The ordinary native development highScore=nil convention is not evidence here.
Structural tests poison canonical highScore with zero, which would qualify if
its consumer were accidentally used; candidate behavior still follows its own
verified top.

This is the explicitly authorized bounded native policy. It is not a proof that
DOS runtime HI_SCORE_LIST always retains compiled defaults. Historical
INIT_HIGHS/SAVE_HIGHS consumers can load/store four16-byte records, and mutable
DOS high-score state may differ. This pass neither admits that lifecycle nor
claims that the original linked demo never reads `.HI`. An unverified persistent
source cannot enter this candidate's comparator.

At35790 the compared fields are:

| Field | Value |
| --- | --- |
| SIFFRORNA, candidate actual BYGEL1 award | BCD `000000050030` =50030 |
| HI_SCORE_LIST first record, verified factory policy | BCD `000050000000` =50000000 |
| I_UTSKJUT /SPECIALMODE /ALREADY_BEATEN | false /false /false |
| First differing digit, zero-based index4 | current0 < top5 |
| Selected branch | JB nonqualification; no BEATENTS or reward |

CHECKHIGHSCORE leaves score, bonus, SCORECHANGED, ALREADY_BEATEN, panel request,
lamps, tasks, audio and matrix program/cursor unchanged. NODOT paints the score
and advances its scored idle count0→1. Matrix remains inactive, op0, owned local
cursor4 at the completed35790 boundary. This cursor is normalized against the
reference SHOWPLAYERSTS stream; raw cross-layout cursor addresses are not equal.
The old zero-score native-entry projection remains the accepted predecessor;
this slice adds scored idle accounting, not a new attract/high-score lifecycle.
The unadmitted scored inactivity branch at count720 is explicitly refused.

## Fallible execution and tests

Implementation is confined to tagged/test-only `dmoimpl1`. The ordinary canonical
CHECKHIGHSCORE /NODOT consumers are not called. Existing2B9 tests retain their
original operand envelope and expected35790 refusal; new tests explicitly admit
2B10 through its loader. No production profile or frontend entry is added.

The new preflight runs before idle-panel consumption, matrix installation,
score painting/flush or scored idle count changes. Qualification refuses before
ALREADY_BEATEN, BEATENTS installation, BEHOVS_PROVAD or `_DOBEATEN` effects.
Unsupported state stays sticky across sync and direct entries. Earlier awards
are not rolled back. The existing task scan and shared matrix dispatcher retain
their order, including the same-visit NEXT_A after the supported panel consumer.

Structural tests are separate from the fresh replay:

- `TestDemoHighScoreComparatorStructural`: independent integer oracle, strict
  tie behavior, every differing decimal position, invalid digits in either field.
- `TestDemoHighScoreGuardsStructural`: below/equal/reward, all three source
  guards before top reads, missing/persistent/changed/invalid top, invalid current
  score, inactivity refusal, unchanged preflight state and sticky failure.
- Python source tests: linked admission, adversarial mutations, IO interception
  rejecting `.HI` access and checks against canonical consumer/IO fallback.

`TestDemoFreshCompletedScoredCalculation` constructs new factory native-reference
and isolated demo candidates, with verified volatile factory state installed
before replay. Timer/clock/score/bonus start at zero; no initial tasks or active
matrix. The reference independently uses its ordinary comparator with the same
verified native policy. It supplies no candidate award, ball or state.

The fixed `demoFixedInput` helper is unchanged:
Down35438..35459, Release35460; d=calculation-35460;
Left iff `(d+46)%52 < 8`, Right iff `(d+22)%30 < 21`.
There is no trajectory search, input change, restore, teleport, task insertion,
timer seed or direct release in the fresh test. Structural setups are not
reachability witnesses. All canonical physics callback hooks remain poisoned.

Every completed calculation1..35876 compares ball/physics, flippers, spring,
camera, score/bonus/SCORECHANGED, clock/random, timers, guards, lights/lamps,
flashes, music clock, wait words, live task slots/IDs, matrix activity/op/timing
and program-local cursor against independent A. The reviewed demo FLASHOFF alias
is normalized to its shared source role. Known demo CLOSE1 VISAKEYS and panel
text differences remain as accepted in2B8/2B9. Every completed row also matches
the pre-existing private DMO0 witness; its producer is not run.

| Calculation | Actual candidate boundary |
| ---: | --- |
| 35789 | Complete; score0, exactly the accepted preaward prefix |
| 35790 | BYGEL1, SBYGEL1, ScoreAwarded50030; CHECKHIGHSCORE nonqualification; complete, including late physics |
| 35876 | Complete; actual late physics marks ball lost, score50030/bonus0/SCORECHANGED=true |
| 35877 | Same early pair, real lost-ball detection, existing LOOSE_BALL guards; unsupported scored branch before expired/LOSTBALL; incomplete |

The first late physics after the award changes Y458267→459054,
VY787→794, rotation1582→1580; X4325/VX0 remain. Pixel position is4,448.
No score is added twice. The fresh final scored idle count is87.

## Exact next boundary

```
UNSUPPORTED_DEMO_TRANSITION
producer=LOOSE_BALL consumer=scored drain
phase=drain handoff calculation=35877
SCORECHANGED=true expired=false
reason=selected scored branch unimplemented before expired/LOSTBALL consumer
```

The selected source edge is the unexpired scored LOSTBALL effect/bonus matrix
continuation, not PARTY_ON or demo expiry. The candidate refuses at the existing
scored-branch preflight, before invoking that effect. Supported LOOSE_BALL guard
writes already set source HOLDSTILL/LOOSING, disable flippers, clear mode flags
and park the ball at15,47; they are retained, as in the predecessor structural
implementation. Score50030, BCD, bonus0 and SCORECHANGED=true survive failure.
No LOSTBALL sound, bonus consumer, extra ball or following calculation executes.

The earlier generic already-lost entry rejection remains for old admission
envelopes. With2B10 admission, a ball lost by the preceding late step is allowed
through the same normal early pair to the existing LOOSE_BALL guard/refusal.
This exposes the real consumer without implementing scored-drain bonus.

## Validation

Private logs: `/private/tmp/pf-dmo-impl2b10-*`.
Existing private harness: `/private/tmp/pf-dmo-impl1-g75qk5z5`.

| Check | Result |
| --- | --- |
| Fresh fixed TARGET, independent A and saved witness through35876 | PASS |
| Two identical fresh runs, completed35790/late physics and sticky35877 refusal | PASS |
| Comparator/source/qualification/guards/invalid state/reward preflight | PASS, structural tests separate from fresh |
| DMO-IMPL-1..2B9 retained plus2B10 TestDemo | PASS,72 top-level tests, no skips |
| New source/IO/fallback suite | PASS,3 tests,56 mutation subcases |
| Python demo regression suite with private historical fixtures | PASS,591 tests, no skips |
| Independent native factory seed identity/values | PASS,2 top-level tests, all four tables |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS,153 top-level tests;5 optional skips |
| A/B/C/D focused datalayout/frontend | PASS,21 top-level tests, no skips |
| Tagged vet | PASS |
| Windows amd64 executable and engine build | PASS |
| macOS c-shared engine build | PASS |
| Broad macOS compile-only | Known baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL: gameshow53/speeddevils36/stones47, native content |
| Fixture-free TestDemo | PASS available29 tests;43 private-backed SKIP/NOT AVAILABLE |
| Public source checker and diff whitespace | PASS |
| Scoped payload scan | PASS,8 source/report files and2 built artifacts |
| OriginalTrajectories /Stones pf10 captures | NOT AVAILABLE; excluded, no fixture generated |

Payload validation checks whole private runtime files and nontrivial aligned4KiB
raw/hex/base64 samples across92 private files (1677 sample blocks). It is a
sampled check, not an all-substring proof. Built artifacts contain no candidate
transition diagnostic. Private originals and logs stay external.

The initial expanded matrix projection checks exposed a demo FLASHOFF alias and
reference program-label lookup mismatch in the test. They were corrected to
use the actual timing/source program label and local cursor; the final full
suite and repeated fresh runs pass. No input or physics change resolved those
test assertions. The final A/B/C/D run uses the already accepted corrected C/D
fixture mapping.

Reproduction (existing private fixtures and research Capstone required):

```
PF_10MIN_DEMO_DATA=<pinned-demo> PF_DEMO_RESEARCH_WITNESS=<saved-json> \
 go test -tags dmoimpl1 ./internal/partyland ./internal/physics \
 -run '^TestDemo' -count=1 -v
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> \
 python3 -m unittest discover -s tools -p test_check_demo_2b10_consumers.py -v
```

The Go fixture-backed command runs from the existing private source-only harness,
which has the same current sources and authorized original data. The public
checkout fixture-free command remains valid with the explicit skips above.

No canonical-A oracle, .DS_Store or v0.1.3 change. No commit/push/tag/release,
payload publication, high-score persistence/UI/initials, full arcade/mystery,
scored-drain bonus, new search, DMO0 research, DMO1, interactive profile or
frontend integration. DMO-IMPL-2B11 is not started.
