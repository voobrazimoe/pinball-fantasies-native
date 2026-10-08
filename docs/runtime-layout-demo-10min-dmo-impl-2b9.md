# DMO-IMPL-2B9: TOUCHER chain and first actual score

2026-10-08. Owner approval: DMO-IMPL-2B9 ONLY.
Accepted predecessor: [2B8](runtime-layout-demo-10min-dmo-impl-2b8.md).
Branch main, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e` unchanged.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_TOUCHER_CHAIN = READY

DEMO_FIRST_SCORED_PREFIX = PARTIAL

The unchanged fresh input-only TARGET reaches and executes unlit BYGEL1 at
calculation35790: score becomes50030, BCD `000000050030`, bonus0,
SCORECHANGED=true. That calculation subsequently refuses NODOT/CHECKHIGHSCORE
in task/matrix, before late physics and Sync accounting. Last completed
calculation is35789. PARTIAL explicitly does not claim a completed35790 boundary.
The score award is real candidate execution, not a structural placement or a
value inherited from canonical A. No scored-drain continuation is implemented.

## Fixed fresh entry

`TestDemoFreshFirstScoredPrefix` constructs a new canonical-data native-reference
factory game and applies Legacy, exactly as accepted in2B8. It creates a new
isolated demo candidate, timer0/clock0/score0/bonus0, player1/ball1, empty tasks,
no captured/source-held/lost ball and inactive matrix. The existing pinned demo
and canonical identities and linked correspondences are checked before admission.
There is no restore, timer seed, task insertion, ball placement or direct release
in this fresh test. Structural injections reside in separately named tests.

`demoFixedInput` remains unchanged:

- 1..35437: no controls.
- 35438..35459: Down.
- Release35460.
- From35460, d=calculation-35460; Left iff `(d+46)%52 < 8`, Right iff
  `(d+22)%30 < 21`; all other controls false.

No search, alternate trajectory or input mutation was run. Two fresh runs pass.
The old `TestDemoFreshGameplayBoundary` remains intact and still verifies the
2B8 admission envelope1..35711 and its explicit unadmitted target boundary35712.

## Reviewed binding and implementation

`check_demo_2b9_consumers.py` reuses the complete frozen
`TOUCHER_and_ENABLETOUCHER` instruction correspondence in
`tools/dmo0_replay_correspondence.json`, through predecessor admission. It checks
additional concrete consumed operands: target0 rectangle130,196..146,204 and
callback file0x1650; repeat byte DS:94; task entry file0x169b; wait20;
WAITLIST word DS:36db; ENABLETOUCHER clear and SUICIDE jump. The shared WAITSYNCS
body at file0x5ac7 compares before increment, resets the word only on equality,
and returns readiness through carry. Source mutation tests cover binding,
guards, sound request/record, task pointer, flashes, wait operand/site/body,
clear and suicide. These bounded correspondences do not close DOS admission,
handler, alias, stack or IRQ domains.

S_TOUCH2's actual source record is sample7/note10/fourth voice: stored channel3
is incremented by the INT66 request. It is a channel sound, not a jingle;
MusicClock priority cannot suppress it. The candidate uses the shared sound
request primitive and keeps host playback closed. Native sound request testing
does not claim physical audio output.

Linked order: repeat guard; S_TOUCH2; touchDisabled=true; ADDTASK ENABLETOUCHER;
if Arcade was false, set true and allocate flashes7/55, speed12, phase0; RET.
An inhibited contact consumes the physics target/contact observation but returns
without a second callback body, sound, flash or task, even if all slots are full.
If Arcade is already true, sound/repeat/task still run, without new flashes.
ENABLETOUCHER clears only touchDisabled, leaves Arcade true, and suicides.

The implementation remains tagged/test-only under `dmoimpl1`. It uses shared
sound, flash, score/BCD, first-free task allocation, ascending task scan, wait
and identity-safe removal. `physics.CandidateTargets` uses the existing shared
spring/target selector on detached state, admits the complete consumer before
publishing spring/contact/event changes, then runs the isolated callback.
It neither invokes nor rebinds the candidate's canonical OnEvent hook. All four
canonical physics hooks remain poisoned in the fresh test.

Unsupported target/nested allocation is rejected before target writes or sound/
guard effects. Due ENABLETOUCHER operand admission is checked before WAIT reset,
guard clear and suicide. All errors are sticky across subsequent sync and direct
consumer entries. No parallel scheduler is added.

## Actual chronology

| Calculation | Observed result |
| ---: | --- |
| 1 /19 | BYGEL12 /BYGEL28, matching accepted fresh prefix |
| 35460 | Real release, charge22, clock20248/low8=24 |
| 35461 | BYGEL12 |
| 35481 | CLOSE1, source chute exit and demo VISAKEYS clear |
| 35524 | BYGEL9, previous CLOSE1, no loop award |
| 35545 | BYGEL11, inhibited reverse, counter cleared |
| 35712 | TOUCHER target0; S_TOUCH2; guards set; task inserted live slot0, persistent ID1; flashes inserted slots2/3; same-calculation DO_TASKS first visit increments wait0→1 |
| 35713 | Same live slot0/ID1, wait2, inhibited=true |
| 35731 | Wait19→20; task still live, repeat guard true |
| 35732 | Compare20==20; wait reset0; guard false; task suicide; Arcade remains true |
| 35789 | Last full electronics/task/matrix/late-physics boundary, score0 |
| 35790 | BYGEL1 unlit Light39, SBYGEL1, BCD50030, SCORECHANGED=true, score-info count0; later NODOT refuses CHECKHIGHSCORE |

At end35712, ball pixel119,204, VX=-1129/VY=1624, rotation-957;
new flash phases are1 after the shared same-calculation flash tick.
End35732: pixel53,311, VX=-1129/VY=2044, rotation-837. The firing calculation
was measured from the actual slot scan and first visit, then asserted; it was
not inferred merely by adding20 to insertion time. Persistent ID1 remains in
slot metadata after suicide; the live function slot is empty.

At the35790 refusal, ball is `X=4325 Y=458267 VX=0 VY=787 GX=0 GY=7`,
pixel4,447, rotation1582, low/not held/not lost, HitX/Y0. Source score effects
already completed in order: Switch BYGEL1, Sound SBYGEL1, ScoreAwarded BCD50030
(value50030). Light39 is false, so EXTRABALL2 was not invoked. Score is exactly
BCD `000000050030`, bonus `000000000000`, SCORECHANGED=true, infoCount0.
Arcade=true, touchDisabled=false, Happy/Mega=false; no bonus or extra-ball award.
No award or target state is taken from the reference.

## Independent comparisons and precise next boundary

Every completed calculation1..35789 compares the full accepted2B8 applicable
physics projection (ball, flippers, spring, camera offset/raster/speed/position)
and score/bonus/SCORECHANGED, clock/random, skill/reverse counters, area state
and lights. This stage additionally compares Arcade/repeat guard, lamps,
flash slots/phases, MusicClock and shared wait words. They match independent A.
At35790 the candidate ball matches an independently observed A pre-target/
pre-late ball; score/bonus/SCORECHANGED match A's actual award. A alone finishes
that calculation; the candidate does not claim its final physics state.

Known semantic differences remain explicit: CLOSE1 has the demo VISAKEYS=false
store; player/ball panel texts are the pinned demo texts loaded by2B8. They are
not replaced by A text/state. Matrix cursor addresses and owned streams are not
claimed numerically equal across layouts. No newly differing gameplay rule
appeared on the admitted TOUCHER/BYGEL path.

With `PF_DEMO_RESEARCH_WITNESS` pointing at the existing private
`pf-dmo0-deterministic-bygel-drain.json`, the test reads the saved artifact
without running its producer. Every completed calculation1..35789 is compared
against its saved row for ball, spring/validity, score/SCORECHANGED, four totals,
Light39/all lights, area state, skill/reverse counters, Arcade/repeat guard,
flippers and waits. Saved35790 score/flag/Light39 also match. This is validation
of TARGET, not new DMO0 research. Without that external artifact the test logs
NOT AVAILABLE for the saved comparison while retaining the independent A test.

The exact first remaining boundary is:

```
UNSUPPORTED_DEMO_TRANSITION
producer=NODOT consumer=NODOT state/guard=CHECKHIGHSCORE
phase=task/matrix calculation=35790
reason=nonzero score idle chain not admitted
```

The existing2B8 idle consumer intentionally refused any nonzero-score
CHECKHIGHSCORE continuation. Its actual linked high-score guard/operands/body
have not been admitted by this slice. A's development `highScore=nil` is not
substituted for the demo's source state to bypass it. The failure is stable;
no matrix fallback, late movement,35791, drain35877, expiry or QUIT is executed.
This is a concrete presentation continuation blocker after a successful award,
not an assertion that BYGEL1 is unreachable.

## Validation

Private logs: `/private/tmp/pf-dmo-impl2b9-*`. Existing private harness:
`/private/tmp/pf-dmo-impl1-g75qk5z5`. No missing fixture is generated.

| Check | Result |
| --- | --- |
| Fixed fresh replay, saved witness and independent A | PASS, completed1..35789, actual award/refusal35790 |
| Two identical fresh runs | PASS |
| TOUCHER guards/repeat, voice request/flash, allocation/age/firing/suicide | PASS |
| Nested allocation/due-body refusal, target preflight, poisoned canonical callback, sticky failure | PASS |
| DMO-IMPL-1..2B8 plus2B9 TestDemo | PASS,69 top-level tests, no skips |
| New source suite | PASS,2 tests,25 mutation subcases |
| Python demo regressions with private source fixtures | PASS,588 tests, no skips |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS,153 top-level tests;5 optional skips |
| A/B/C/D focused datalayout/frontend | PASS,21 top-level tests, no skips |
| Tagged vet | PASS |
| Windows amd64 executable and engine build | PASS |
| macOS c-shared engine build | PASS |
| Broad macOS compile-only | Known baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL: gameshow53/speeddevils36/stones47, native content |
| Fixture-free TestDemo | PASS available tests; private-backed tests SKIP/NOT AVAILABLE |
| Public source checker and diff whitespace | PASS |
| Scoped payload scan | PASS,9 source/report files and2 built artifacts |
| OriginalTrajectories / Stones pf10 captures | NOT AVAILABLE; excluded, no fixture generated |

The initial A/B/C/D command reused the pre-correction C/D environment mapping
and failed descriptor identities. Rerunning with the accepted corrected mapping
passed all21 tests; the first attempt is not counted as a successful gate.
The first Python run lacked the historical-source environment and skipped16
private suites; it is superseded by the explicit private-source run.

Reproduction (existing fixtures and research Capstone environment required):

```
PF_10MIN_DEMO_DATA=<pinned-demo> PF_DEMO_RESEARCH_WITNESS=<saved-json> \
 go test -tags dmoimpl1 ./internal/partyland ./internal/physics \
 -run '^TestDemo' -count=1 -v
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> \
 python3 -m unittest discover -s tools -p test_check_demo_2b9_consumers.py -v
```

Payload validation covers whole private runtime files and nontrivial aligned4KiB
raw/hex/base64 samples across92 private files (1677 sample blocks). This is a
sampled check, not an all-substring proof. Built artifacts contain no candidate
transition diagnostic; private originals remain external.

No canonical-A oracle, production profile, .DS_Store or v0.1.3 changes; no
commit/push/tag/release. No full arcade/mystery/reward system, multiplayer,
cheats, interactive demo, DMO1 or scored-drain bonus continuation.
DMO-IMPL-2B10 is not started.
