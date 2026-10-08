# DMO-IMPL-2B5: isolated child tasks and first new-ball late physics

2026-10-08. Owner approval: DMO-IMPL-2B5 ONLY.
Branch `main`, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_CHILD_TASKS = READY

DEMO_NEW_BALL_RELEASE = READY

READY is the supported structural unscored-drain continuation in the isolated
`dmoimpl1` candidate. All three children execute and suicide; SETBALL is followed
by the first real shared late physics step in the same calculation. The next
ordinary gameplay consumer is unsupported. No fresh input-only candidate witness,
full gameplay, expiry termination or QUIT PASS is claimed.

## Source bodies and admission

The accepted 2B4 report and existing reviewed linked first-equality SETBALL and
post-collision correspondence supply the bounded consumers. Historical
`PLAND.ASM` SOUNDBRICKUPP/SOUNDNEWBALL/SETBALL and `FANTASIE.MAC`
SOUNDEFFECT/SETBALLPOS were checked against the actual linked bodies. No new
DMO0 graph, writer-domain, reachability or trajectory research was run.

The two sound bodies are **SOUNDEFFECT**, not DOPLAYJINGLE. After their existing
WAITSYNCS they request INT66 AL=17 and jump to SUICIDE. Linked body extents are
file fa8..fce and fce..ff4. Their effect record pointers are DS:c39 and DS:c51.

| Task | Shared WAITLIST | Limit | Actual consumer operands |
| --- | --- | --- | --- |
| SOUNDBRICKUPP | DS:36cf | 5 | SBRICKUPP: sample23, note23, volume0, channel4 |
| SOUNDNEWBALL | DS:36d1 | 50 | SNEWBALL: sample28, note18, volume0, channel4 |
| SETBALL | DS:36d3 | 80 | SETBALLPOS(297,530,10,0,false), HOLDSTILL=false, SCREENFORCE2=-1 |

SOUNDEFFECT reads record bytes0/1/3; byte3=3 is incremented to channel4. Byte2
is not a repeat or priority operand. Neither sound task reads or writes jingle
position, repeat, priority, readiness, return position or cue elapsed state.
Current jingle priority255 therefore does not reject these sample requests.
Tests cover priorities0/255 and verify the entire MusicClock is unchanged by
the body, followed by its normal shared Sync advancement. There is no invented
jingle request or priority-dependent omission of SUICIDE. The existing expiry
jingle request remains separate.

The silent candidate uses the existing `Game.sound` request consumer, with
verified `audio.Effects` sample/note correspondence and Playback=nil. It records
real typed Sound events; no host backend or sample renderer was added. A missing
operand/map consumer or attached playback consumer fails UNSUPPORTED before the
due wait reset. The connected calculation now advances the existing MusicClock
via audioTick before its early physics pair; host playback is refused before
that call. Rendering and an external DOS INT66 ABI remain outside this boundary.

`loadResetOperands` also admits the linked effect records. The private identity
check validates the concrete child macro reads, volume/channel/INT66 operands,
SUICIDE targets and SETBALL stores. These checks reuse known reviewed offsets;
they do not traverse new domains or embed private payloads.

All due bodies are preflighted in the existing task closure before waitReady can
reset their shared site. A successful body returns true to the shared scheduler's
identity-safe removal. No task runner, replacement scheduler or second scan is
introduced. Failure leaves the due word, live task and other identities intact;
later closures observe sticky failure before consuming their waits.

## SETBALL projection and first physics

SETBALL uses shared Physics.SetBall and writes fixed X=297*1024, Y=530*1024,
pixel position(297,530), VX=10, VY=0 and High=false. It clears source HOLDSTILL,
writes source/native SCREENFORCE2=-1 and projects native Phase=Playing. I_UTSKJUT
/inChute is already true from NEW_BALL and is retained; it is an admitted
prerequisite, not a fabricated extra linked store. SETBALLPOS retains gravity,
rotation, hit/contact state and native capture Ball.Hold. SETBALL never clears
native Hold. A captured native ball stays captured through late physics even
when source HOLDSTILL becomes false.

Missing linked operands, table, chute/reset state, a lost ball, LOOSING=true or
wrong native phase refuse SETBALL before WAIT reset or release. The successful
body does not reset expired, timer, matrix program/cursor, tasks, player state,
score or demo expiry; it does not execute NEW_BALL again.

The supported structural sequence is:

`real physics drain -> LOOSE_BALL -> PARTY_ONTS -> PARTY_ON_TASK1 -> NEW_BALL
-> NEW_BALL_PART_TWO -> SOUNDBRICKUPP -> SOUNDNEWBALL -> SETBALL -> late physics`.

For parent slot0, H=31. Visit-by-visit assertions check each live child's age,
slot identity, due reset and suicide through calculation111:

| Parent slot | Ages after H: SOUNDNEWBALL / SETBALL / SOUNDBRICKUPP | Brick / NewBall / SetBall firing |
| --- | --- | --- |
| 0 | 0 / 1 / 1 | 36 / 82 / 111 |
| 1 | 0 / 0 / 1 | 36 / 82 / 112 |
| 3 | 0 / 0 / 0 | 37 / 82 / 112 |

Only the first row establishes H+5/H+51/H+80. A separate competing-instance test
proves that two SOUNDBRICKUPP instances consume one shared wait word: their ages
and firing dates cannot be inferred from that row. Each suicide preserves the
other instance's identity.

The calculation111 early pair sees the source-held position(282,530). DO_TASKS
then executes SETBALL, the current admitted persistent PARTYON matrix routine
runs, and late.step sees sourceHold=false. CandidateStage calls the actual
shared physics step. Observed post-step state is X304138, Y542720, VX10, VY7,
GX0, GY7, pixels(297,530), High/Hold/Lost=false, Hit(297,538), angle1024,
contactCount13, material7. Expected state is independently computed with a fresh
canonical-A physics instance using its ordinary Sync: native capture suppresses
the early pair and the reference releases capture at BeforeTargets for exactly
one late pass. The candidate receives no expected coordinates or reference state.
The complete Ball values match.

Assertions preserve saved player/score, timer progression, expired, matrix
cursor, dot memory and the expected shared Flash advancement. WAITLIST, live
identities, BALL_DOWN/LOOSING, handoff/drain count and source/native hold are
checked. Canonical callbacks are poisoned in the connected fixture. The next
calculation112 refuses `checkAreas -> BYGEL12` before area bookkeeping or its
callback. Other parent slots stop at the same consumer on calculation113.

CandidateStage is now guarded by the `dmoimpl1` build tag. Its collision
lookahead uses a detached Game value and detached Events before live collision
writes, refusing unadmitted collision/event consumers. A structural bumper
fixture verifies that even Ball/contact/pending/event state remains unchanged on
rejection. The successful step still runs the existing integrator directly.
Canonical physics and production profiles/oracles are unchanged.

## Structural first-equality SETBALL collision

The declared structural initial timer35887 gives real unscored drain35888,
PARTY_ON_TASK1 handoff35918, SOUNDBRICKUPP35923, SOUNDNEWBALL35969 and
SETBALL35998. Every calculation from the prepared drain through release runs
the same connected candidate with matrix budget=true.

At35998, held early physics precedes electronics equality. Equality sets
expired=true, source hold and the expiry jingle/program before DO_TASKS.
SETBALL then clears only source hold and releases the ball. Expired remains true;
expiry CLEAR4 remains current with next1/remaining4 after its ordinary matrix
visit. The same calculation executes real late physics. Neither PARTY_ONTS nor
a prior cursor is restored; handoff count stays1.

On35999 the first next unsupported consumer is BYGEL12, before the next matrix
visit. Thus this full candidate continuation does not reach expiry SCROLL.
The preserved matrix-only expiry tests still refuse NEXT_A before SCROLL; with
uninterrupted matrix visits from35998 its due boundary would be36002. Matrix
budget is never disabled to manufacture a full expiry PASS. No full fresh
candidate witness for the research prefix is claimed.

## Validation and limits

Logs/artifacts: `/private/tmp/pf-dmo-impl2b5-*`. Original-backed candidate and
reference checks use the existing private canonical-A regression checkout
`/private/tmp/pf-dmo-impl1-g75qk5z5`, with the current candidate files copied in.
No missing fixture was generated.

| Check | Result |
| --- | --- |
| All tagged TestDemo, 1/2A/2B1/2B2/2B3/2B4 plus 2B5, private inputs | PASS: 47 top-level tests, no skips |
| Public fixture-free tagged TestDemo | PASS available checks; original-backed tests skip NOT AVAILABLE |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS: 153 top-level tests; five opt-in capture/export checks skip |
| OriginalTrajectories / Stones pf10 captures | NOT AVAILABLE; excluded without generation |
| A/B/C/D focused datalayout/frontend compatibility | PASS: 21 top-level tests, no skips |
| Public-root broader datalayout/frontend | PASS available tests; original-dependent checks skip |
| Private broader frontend attempt | FAIL: PF6 baseline plus missing pf11.2-full-table-fixtures/pf8-bios-font and three native-content lifecycle checks; not reported as a green regression gate |
| PF6Session | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Tagged vet: partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 CGO-disabled executable and internal/engine package | PASS |
| macOS c-shared engine | PASS |
| Broad macOS compile-only and unchanged baseline checkout | Same baseline FAIL: AudioDevice/hostWindow/openHost |
| Public source checker / diff whitespace | PASS |
| Milestone sources/report and two build artifacts payload checks | PASS: whole-file/nontrivial aligned4KiB samples; not an all-substring proof |
| Fresh input-only demo candidate witness | NOT AVAILABLE; not attempted |

The broader frontend native-content failures are outside this test-only change;
that command's exit remains FAIL. The focused compatibility command is
`go test ./internal/datalayout ./internal/frontend -run
'TestPrivate|TestRuntime|TestHostsShareRuntimeBoundary|TestPowerPack|TestDeluxe|TestProfiles|TestPossessed'`.
Reference tests exclude only `TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers`.
Candidate reproduction: `./tools/go.sh test -tags dmoimpl1 ./internal/partyland
./internal/physics -run '^TestDemo' -count=1 -v` with the existing private inputs.

This milestone changes only test candidate files and the build-tagged staging
helper/report. Production tracked edits predate this milestone. No scored drain,
bonus/progression, new gameplay effects, expiry SCROLL/FADE/QUIT, interactive
demo, profile registration, DMO1 or new DMO0 research was implemented.
`.DS_Store`, canonical-A oracle and v0.1.3 were not changed. No commit/push/tag/
release. DMO-IMPL-2B6 is not started.
