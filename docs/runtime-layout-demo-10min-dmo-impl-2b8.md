# DMO-IMPL-2B8: first fresh gameplay prefix

2026-10-08. Owner approval: DMO-IMPL-2B8 ONLY.
Accepted predecessor: [2B7](runtime-layout-demo-10min-dmo-impl-2b7.md).
Branch main, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e` unchanged.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_FRESH_GAMEPLAY_PREFIX = READY

DEMO_BYGEL_CONSUMER_FAMILY = READY

READY covers the bounded branches below and the fixed fresh prefix ending before
TOUCHER at calculation 35712. It does not mean unrestricted BYGEL modes, complete
gameplay, a historical DOS trajectory, demo profile registration or host output.
There is no claim of a fresh score award: the actual admitted path has zero score
and bonus. Award arithmetic is tested separately with explicit structural states.

## Fresh entry and fixed inputs

`TestDemoFreshGameplayBoundary` creates a new
`New(DecodePartyLand(canonicalA), canonicalA)` and applies `settings.Legacy()`.
This is the accepted fresh native-reference factory convention, with a new demo
core/TABLE1 lifetime, timer=0, clock=0, player 1, ball 1, fresh score and gameplay
state. The factory is the entry; it is not a replay of DOS initialization.
No ball/timer placement, task insertion, direct Release call, timer seed or
state restoration occurs in this test. Pinned demo identity and existing linked
consumer correspondences are checked before gameplay admission. Private demo
panel operands are loaded in memory; no payload is added to the repository.

The exact saved script is `TARGET` in
`tools/audit_10min_demo_deterministic_replay.py`, also recorded in
[the deterministic witness](runtime-layout-demo-10min-deterministic-bygel-drain.md):

- 1..35437: no controls.
- 35438..35459: Down held, 22 charge visits.
- 35460: release edge.
- From 35460, let `d = calculation - 35460`; Left iff `(d+46)%52 < 8`,
  Right iff `(d+22)%30 < 21`. All other inputs false.

`demoFixedInput` transcribes these inputs. No new trajectory search or neighboring
script is run. TIME_LEFT is true on this fresh replay. Two fresh runs pass.
The convention remains uint16 clock +=1030 before early physics, jitter low8
`(6*n)%256`. At release35460, clock=20248/low8=24, charge22; shared release
arithmetic sets VY=-3676 and rotation8 before the late physics pass. End of that
calculation: `(301,533)`, VY=-3669, rotation6, spring position0.

An independent fresh canonical-A game consumes the same inputs as a comparison
control. It never provides callbacks or state to the candidate. Every completed
calculation compares the complete ball, flippers, spring position, camera offset,
raster, screen position/speed, score/bonus/SCORECHANGED, random/clock, skill/reverse
counters, LASTAREA/LASTCHECK and progression lights. Candidate canonical physics
hooks are replaced by test failures; none executes. Full completion includes the
late physics pass and Sync accounting, rather than ending at a callback.

## First real boundary, before implementation

`TestDemoFreshInitialRejection` reconstructs the unadmitted connected candidate
from the same fresh entry and script. First rejection is **BYGEL12 at calculation
1**, before launch; this was established by a fresh run, not assumed from its name.
Producer `checkAreas`, phase electronics areas/targets, timer1; low ball,
not tilted/lost/captured/source-held, center `(305,538)` in the inclusive low
rectangle `(305,455)..(320,540)`. Early physics ball:

`X=304148 Y=542728 VX=10 VY=16 GX=0 GY=7 Pixel=(297,530)`;
contact `(297,538)`, angle1024, count13, material7, rotation0.

Score/bonus0, SCORECHANGED=false, tasks empty, matrix inactive, LASTCHECK/AREA
empty. Rejection precedes area bookkeeping and all callback effects.
Linked demo BYGEL12 file `0x28c6..0x28d9` sets skill counter300,
SPRING_VALID=false and inhibit-reverse counter120. It requires no sound, score,
bonus, matrix or task. BYGEL28 (`0x28d9..0x28e0`) restores SPRING_VALID.
CLOSE1, BYGEL9 and BYGEL11 are the connected family on the saved script.

## Bounded implementation and admission

All new gameplay is test-only under `dmoimpl1`. Production canonical callbacks,
oracle and default builds are unchanged. Shared score/BCD, lamp/flash, MusicClock,
spring arithmetic, integer physics, camera, task scan and matrix primitives are
reused; no Game.Sync/trigger/reverse/loop/effect/newBall fallback is called by the
candidate.

The source adapter reuses existing fixed operand correspondences, validates
canonical-A and demo identities, checks the bounded BYGEL3/4 bodies and their
actual binding/effect/sound records, and validates PARTY_OFFTS/SHOWPLAYERSTS
linked words. Shared font5 bytes are checked, while the actual demo player/ball
texts are owned in memory. Those texts differ from A and are not replaced with
canonical text. New mutation tests reject changed instructions, arithmetic,
guards, commands and voice operands. This is bounded implementation admission;
it does not reopen DMO0 external handler/alias/IRQ closure.

| Consumer | Supported branch/effect | Refusal before nested effects |
| --- | --- | --- |
| BYGEL12/28 | Skill/reverse counters and spring validity | Missing pinned operands |
| BYGEL9 | Previous area is not BYGEL11; return | Loop award chain |
| BYGEL11 | Previous area differs, or inhibited reverse clears counter; INH_REV return | Uninhibited reverse award chain |
| CLOSE1 | Previous BYGEL12: ADDPLAYERS=false, S_MAIN request/return1, PARTY_OFFTS, chute/party/VISAKEYS clear; otherwise return | Unadmitted program/operands |
| BYGEL1/2 | Light39 unlit: SBYGEL1, BCD50030, SCORECHANGED, score-info bookkeeping | Lit EXTRABALL2 chain; unadmitted nested score panel |
| BYGEL3/4 | Adjacent matrix-free BYGELSETB: score10040, bonus1000, EOTS, SBYGEL2 | Unadmitted nested score panel |

BYGEL1/2 and the nearest score/bonus siblings BYGEL3/4 are structural coverage of
the bounded family; the candidate fresh prefix does not reach them. Matrix=0 and
jingle=0 in BYGELSETB mean no matrix or music request at any priority. Its raw
bonus1000 is not multiplied by BONUSMULTIPEL. Priority/mode suppression cannot
suppress its arithmetic or channel-four sound. Lit extra-ball, loop/reverse awards,
other areas/targets, hole systems and scored-drain continuation remain closed.

The fresh envelope adds keyboard edges and spring control, after-task spring
ordering, lamp rotation and shared camera/Sync accounting. Non-keyboard/tilt and
FastBall entries are rejected. Task ownership, ascending scan, matrix budget and
sticky errors remain enforced. Area admission precedes LASTCHECK, Switch and
callback effects. Leaving all regions clears LASTCHECK; a continuing area does
not call the consumer twice.

Required presentation continuation is implemented rather than skipped:
PARTY_OFFTS's clear/one-visit PARTYOFF/FLASHOFF/zero and NODOT's pending player
panel through DO_SPEC_MATRIX, real demo PRINT5 operands and NEXT_A on that visit.
DO_SPEC_MATRIX does not reset BEHOVS_PROVAD. A nonzero-score idle high-score chain
is refused; that branch is outside the fresh zero-score prefix. No mandatory
startup consumer is missing on the claimed path.

Collision flipper events are numeric physics observations, not calls to
Party Land OnEvent. Explicit observation admission preserves real flipper
response. An initial broad collision-event rejection at35593 was investigated
and corrected on this basis; it is not reported as the final gameplay boundary.
Bumper/kicker and target gameplay callbacks remain gated before live effects.

## Actual fresh suffix and next boundary

These rows are end-of-calculation states, including late physics:

| Calculation | Actual callback | Pixel ball | VX/VY | Skill / inhibit-reverse | Score / bonus |
| ---: | --- | --- | --- | --- | --- |
| 1 | BYGEL12 | 297,530 | 10 /23 | 300 /120 | 0 /0 |
| 19 | BYGEL28 | 301,537 | 231 /7 | 282 /102 | 0 /0 |
| 35461 | BYGEL12 | 301,522 | 0 /-3648 | 300 /120 | 0 /0 |
| 35481 | CLOSE1 | 301,321 | 0 /-3228 | 280 /100 | 0 /0 |
| 35524 | BYGEL9, no loop | 228,20 | -1661 /-899 | 237 /57 | 0 /0 |
| 35545 | BYGEL11, inhibited reverse | 120,10 | -1703 /447 | 216 /0 | 0 /0 |

The test asserts these six callbacks' counts (BYGEL12 twice, all others once)
and matches the canonical reference's electronics/physics through35711. The
first unsupported **gameplay** boundary is calculation35712:

- Producer `checkSpringAndTargets target=0`; selected callback TOUCHER
  (linked demo callback file0x1650).
- Phase electronics areas/targets; two early steps and UPDATE_COUNTERS/timer
  completed. Timer35712, completed Physics.Syncs35711, Random128, clock17664.
- Ball `X=123674 Y=207591 VX=-1129 VY=1617 GX=0 GY=7`, pixel `(120,202)`,
  rotation-959, low/not lost/not held. Hit `(132,199)`, angle1623, count5,
  material3. SPRING_VALID=false, skill49, inhibit-reverse0.
- TOUCHER repeat guard is open: touchDisabled=false; Arcade=false.
  Required effects are repeat inhibition, S_TOUCH2, ENABLETOUCHER wait20,
  Arcade=true and flashes7/55 speed12. They have not executed.
- Score/bonus0, SCORECHANGED=false, no active tasks or waits. Matrix inactive,
  owned program's next=4/op=0/remaining=1; LASTCHECK empty, LASTAREA=BYGEL11.
  Lamp bits52/53/54 on. Existing flash slots `(14,15,8)` and `(26,17,9)`;
  the other thirteen slots empty. Audio position/return/priority1,
  JumpCount0, active, elapsed11550/cue136320.

Rejection retains HitX/Y before target consumption, does not publish the target
callback or queue its task, and prevents keyboard/task/matrix/late continuation.
Repeated sync/direct candidate entries return the same sticky failure without
state mutation. Implementing TOUCHER would start target/arcade admission, outside
this owner-approved BYGEL slice. No attempt is made to reach BYGEL1 at35790,
drain35877, equality35998 or QUIT with this candidate.

## Validation and compatibility

Private logs: `/private/tmp/pf-dmo-impl2b8-*`; candidate inputs use the existing
`/private/tmp/pf-dmo-impl1-g75qk5z5` harness. No missing fixture was created.

| Check | Result |
| --- | --- |
| Fresh fixed input-only candidate vs independent A reference | PASS, every completed calculation1..35711; refusal at35712 |
| Two fresh runs of the exact fixed script | PASS |
| Fresh unadmitted boundary | PASS, BYGEL12/calculation1 |
| Structural family, score/bonus/SCORECHANGED, priority, guards, nested admission, task/matrix ordering | PASS |
| Canonical callback poison hooks and sticky failure | PASS |
| DMO-IMPL-1..2B7 plus 2B8 tagged TestDemo | PASS,64 top-level tests, no skips |
| New bounded source mutation suite | PASS,2 tests with13 mutation subcases |
| Existing Python demo regression tests | PASS,584 tests, no skips; no new search |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS,153 top-level tests;5 optional export/capture skips |
| A/B/C/D focused datalayout/frontend | PASS,21 top-level tests, no skips |
| Tagged vet | PASS |
| Windows amd64 executable and internal engine build | PASS |
| macOS c-shared engine build | PASS |
| Broad macOS compile-only | Known baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL: gameshow53, speeddevils36, stones47; native content |
| Public fixture-free tagged TestDemo | PASS available tests; original-backed tests SKIP/NOT AVAILABLE |
| Public source checker, diff whitespace, scoped payload scan | PASS |
| OriginalTrajectories / Stones pf10 private captures | NOT AVAILABLE; excluded, no fixture generated |

Initial broad compatibility used the earlier pre-correction C/D environment
mapping and failed profile identities; the corrected mapping passes. The first
PF6 command matched no tests and is NOT RUN; the corrected frontend test above
ran and reproduced the baseline FAIL. Neither is counted as a successful gate.
A/B/C/D descriptors, recognition and runtime remain unchanged. Builds contain no
demo candidate; the implementation lives in tagged test files.

Reproduction in the existing private harness, with bundled Go and the existing
research Python/Capstone environment on PATH:

```
PF_10MIN_DEMO_DATA=<pinned-demo> go test -tags dmoimpl1 \
 ./internal/partyland ./internal/physics -run '^TestDemo' -count=1 -v
PF_10MIN_DEMO_DATA=<pinned-demo> go test -tags dmoimpl1 \
 ./internal/partyland -run '^TestDemoFreshGameplayBoundary$' -count=2 -v
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> python3 \
 -m unittest discover -s tools -p test_check_demo_2b8_consumers.py -v
```

Payload checks cover this milestone's source/report files and built executable/
engine: whole originals and nontrivial aligned4KiB raw/hex/base64 samples. This
sampled check is not an all-substring proof. Private text/font bytes remain
external and in memory. No canonical-A oracle, .DS_Store or v0.1.3 changes;
no commit/push/tag/release. DMO-IMPL-2B9 is not started.
