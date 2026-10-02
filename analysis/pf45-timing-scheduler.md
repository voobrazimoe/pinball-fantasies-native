# PF4.5 — Party Land task, matrix and silent completion semantics

PF5 review: audible native playback is now implemented. See [PF5 audio evidence](pf5-audio.md). Missing SDR callback phase is PARTIALLY RESOLVED: the selected binary invokes the user callback during B-row processing and preserves fractional mixer time; exact game-boundary/forced-position phase remains UNKNOWN. PF4.5 silent fixtures remain unchanged. The sections below retain the historical PF4.5 scope and inference descriptions.

PF4.5 replaces PF4's independent presentation-delay estimates with Party Land's
original matrix command stream and shared task wait state. Scoring and physics
remain native Go. There is no sound output, sample decoder, mixer, music renderer,
hardware emulation, DOS scheduler, coroutine framework or port of another table.

The state machine is source-guided, not a claim of cycle-exact DOS execution.
The supplied source omits the sound player/SDR implementation. The silent timeline
therefore makes an explicit, tested timing inference; exact DOS driver phase
remains UNKNOWN. This qualification matters when comparing an audible DOS run.

## Evidence and scope

VERIFIED — primary references:

- `reference/original-dos-source/FANTASIE.ASM`: `VBLANK_INT`, `DO_THE_REST`,
  `LATE_RASTER_INTERRUPT`, `DO_ELECTRONICS`, `DOADDTASK`, `DO_TASKS`,
  `DOSUICIDE`, `DOWAITSYNCS`, `RESET_WAITLIST`, `WHEN_NEW_BALL_RESET`,
  `hires_changes`, `DO_THE_ANIMATIONS`, `DO_MATRIX`, `DO_SPEC_MATRIX`,
  `NORMAL_END`, `PRINT_END`, `_WAIT`/`WAITRUT`, `_clear2/3/4`,
  `_COUNTDOWN`, `_countdown2`, `COUNTDOWN`, `_WAITJINGLE`/`_WAITJINGLE2`,
  `DOEFFECT`, `DOPLAYJINGLE`, `JINGLE_HANDLER`, `SETSCREENSTART`.
- `reference/original-dos-source/FANTASIE.MAC`: `WAITSYNCS`, `ADDTASK`,
  `SUICIDE`, `PLAYJINGLE`, `PLAYJINGLE_PENETRATE`.
- `reference/original-dos-source/PLAND.ASM`: effect/jingle structures and
  Party Land matrix tables; `GROPC`, `HIDDENTASK0/1/2`, `GROPD`,
  `WAIT_FOR_TUNNEL_EFFECT`, `LOCK_BALL_IN_TUNNEL`, `GROPB`,
  `WAIT_FOR_SPIN_TASK`, `SPINIT`, `SPIN*_RUT`, `START_DROP_TIMED`,
  `START_DROP_WHEN_READY`, `START_DROP`, `DROPTASK1/2`,
  `RULLGARDIN`/`RULLGARDINSLAV`, `SNACK_HOLE_TASK1/1B`, `BEFOREFLASH`,
  `DURINGFLASH`, `OK_2_BE_HAPPY`, `OK_2_LAUGH`,
  `READ_SPECIAL_MODE_COUNTER`, `HAPPY_START`, `MEGA_START`,
  `LOOSE_BALL`, `ball_lostTS`, `_FLORPA`/`DO_FLORPA`,
  `_BONUS_X_CALCS`, `_CALC_CYCLO/HAPPY/MEGA`, `_CHANGE_PLAYER`,
  `NEW_BALL_TASK`, `NEW_BALL_PART_TWO`, `_knacket`/`knackrut1/2`,
  `_CHECK_XXBALLS`, `_EOSNURR`, `_TSEND`.

VERIFIED — linked `TABLE1.PRG` supplies the external animation/scroll implementation
missing from the source tree. Static disassembly, never execution/emulation,
establishes scrolling at file offsets `0x7197..0x72a9` and animation timing at
`0x72b0` onward. The omitted WAITTS macro is independently resolved in the linked
`HIDDENTASK0` at file `0x1daf`: compare EOTS with 255, return unless equal; add
HIDDENTASK1, call START_DROP, suicide. This resolves its shared-latch semantics.

VERIFIED — `tools/reference_pf45.py` extracts **data directives only**, selects
retail/non-demo branches, and retains 391 matrix commands reachable by Party Land
rules plus the single-player new-ball/match path. It does not translate executable
ASM. All 19 source animation timing arrays are checked against linked DATA2;
only animations used by the retained commands enter native content. The forward
SPEED reference for `_CYCLONE` is resolved to four from the supplied linked data.
No graphic frames or PCM sample contents are copied into generated code.

VERIFIED — original `TABLE1.MOD` has the M.K. four-channel pattern layout, 64 order
entries, Fxx speed commands below 32, Bxx jump cues and Dxx pattern breaks.
No E6 pattern loops, EE row delays or BPM-changing Fxx occur in these patterns.
The oracle follows only those timing/control fields. Its SHA-256 is recorded in
`analysis/pf45-timing-fixtures.json`. Native timing metadata corresponds to this
exact supplied module; no module renderer is loaded at runtime.

## Execution and task ordering

VERIFIED — this is a bounded cooperative slot scan, not task preemption by a
thread scheduler. There are 50 slots. ADDTASK chooses the first free slot.
DO_TASKS scans ascending slots once. A child placed later in that scan executes
in the same tick; a child placed in an already scanned slot waits for the next
scan. A returning task remains installed until it suicides or reset removes it.
The native monotonically assigned task ID prevents a callback that resets the
list from accidentally removing a replacement task in its former slot.

VERIFIED — WAITSYNCS uses a statically assigned WAITLIST word **per macro call
site**. Every invocation compares its uint16 counter with the requested delay.
Equality resets it to zero and continues; otherwise it increments, wraps at
16 bits and returns. Delay N therefore takes N+1 invocations from zero. Multiple
instances share that word. It is not a private elapsed timer per instance.
New-ball reset clears tasks, WAITLIST and flashes. Distinct duck and PUKE call
sites remain distinct. The original byte duck repeat counters have the same
bounded timing on their guarded, supported hit paths.

VERIFIED — source task creation order is retained for new-ball sounds/SETBALL,
tunnel waits/capture, hidden continuation/drop, arcade prize selection and drop
stages. START_DROP_TIMED reads the shared DROP_TIME variable on every invocation.
HANGSAVER/HANGSAVER2/readiness are likewise shared fields. Snack tasks repeat their
position/hold writes before testing the wait, and dynamically bypass the second
wait when SPECIALMODE becomes true.

VERIFIED — camera tasks RULLGARDIN and RULLGARDINSLAV occupy real slots. Their
omission in PF4 could change later ADDTASK ordering even if the camera itself did
not change scoring. They use current START_RASTER (native Raster/16 minus 33),
subtract five per invocation and force zero when the signed result becomes
negative. Drop ejection releases SCREENFORCE. These tasks are no longer collapsed
away.

## Cadence and integer clocks

VERIFIED — one native Sync represents the existing PF3 high-resolution interrupt
pair: two physics passes during VBLANK, rule/electronics work, matrix work, then
one late-raster physics pass. There is one rule/task/matrix scan per pair, not one
per physics pass. Nil PF4.5 hooks leave the PF3 trajectory and framebuffer paths
unchanged.

The established normal order is:

1. Pending plunger sound intent at the native input boundary; deterministic silent completion boundary, then PF3 VBLANK physics.
2. Pending bumper/slingshot event; ramps/levels and drain boundary.
3. UPDATE_COUNTERS; areas; targets; rising-edge Shift behavior.
4. Ascending task scan; lamp flash scan.
5. Matrix routine and at most one ordinary next-command dispatch; print work.
6. Late-raster scrolling/physics. Original RGB palette application is represented
   by logical/visible lamp state; the native renderer consumes that state later.

VERIFIED — original UPDATE_COUNTERS is inhibited while LOOSING. Native ball loss
continues tasks, flashes, matrix and silent audio, but freezes those rule counters.
PF3 already stops the physical drained ball; it does not simulate DOS background
movement during the loss display.

VERIFIED — `hires_changes` selects `sync_per_sec=71`. Ordinary task, matrix wait,
animation and lamp counters count scans. The source adds 1030 to uint16
SLUMP_COUNTERN each VBLANK. The native wrapped random clock retains this addition.
No host time or goroutine scheduling is read by gameplay state.

VERIFIED — source TIME_LEFT=false skips matrix work while task/lamp processing
continues. `SyncWithMatrixBudget(input, timeLeft)` exposes this exact bounded
input for deterministic tests. Normal silent play supplies true. It neither
measures host load nor emulates DOS interrupt/reentrancy overhead. Late-raster
physics remains the PF3 model.

UNKNOWN — actual refresh/SDR interrupt phase and resource-crisis frequency in a
DOS installation. The source requests 71 counter units, not a measured exact
71-Hz hardware oscillator. SDL's existing host presentation delay is pacing only;
changing it does not change a scripted tick sequence's state.

## Matrix completion and interruption

VERIFIED — an accepted DOEFFECT requests its jingle **before** adding score and
raw bonus, then starts the matrix. Rejection/special-mode/inhibition never
suppresses arithmetic. A no-jingle effect still compares its priority. Skill tunnel/cyclone INH_EFF preserves the skill matrix while later letter/table
awards still update arithmetic. Equal
jingle priority replaces; lower priority rejects. Suppression explicitly sets
EOTS=true, even if the interrupted hidden scroller never reached its own TSEND.
EOTS is a shared latch, not a private completion future attached to an effect.

VERIFIED — starting an accepted matrix replaces the prior program; it is not
queued. Its first command is installed immediately. Subsequent scans run the
current routine, then install the next command if that routine completed.
Ordinary command installation does not immediately run that new routine.
Conditional JBCDZ/JBONUSX1 and JMP tail-call the selected handler in the same tick.
PRINT/FLASH/TSEND/EOSNURR installation generally yields a one-scan WAITRUT. CLEAR4
lasts five successful matrix scans, CLEAR2 seventeen, CLEAR3 eighty-one.

VERIFIED — animation timing preserves the linked routine's integer index,
loop counter, initial timer=1 and old-index behavior at loop boundaries. In
particular, `_HAPPY` completes in **92 matrix calls**, rather than summing its
five visible frame durations fifteen times. The linked routine compares index
with the length word before drawing; its saved old index survives the loop-index
write. Flattening that table to a modern animation loop would change rules.

VERIFIED — scrolling performs two substeps per matrix call; its shared phase byte
starts at eight and is **not reset** by a new scroll. A byte advances each eight
substeps. The terminator is tested twenty bytes ahead before each substep. Hidden
completion therefore occurs after 289 successful scans with a fresh phase.
Preemption discards the old stream but preserves scroll phase. Rendering a frame
never advances any timer.

VERIFIED — Happy/Mega scoring flags become active at feature entry. Their
25-second countdown starts only when the intro reaches COUNTDOWN. Jackpot matrix
preemption pauses it; COUNTDOWN2 restores the original remaining displayed
seconds and resets the subsecond phase to one. If it interrupts an intro,
SUPER_INIT's behavior initializes the countdown after the jackpot animation.
Expiry and the two-second quick-flash transition occur from the matrix countdown.
Pending counterpart mode tasks begin their 400 wait on a later task scan.

NON-BEHAVIORAL — actual matrix animation/scroller pixels, flash DAC/fades and
original print positioning remain represented by the PF4 original-font HUD.
Their **control-flow timing** is now processed despite this visual limitation.
High-score-dependent matrix branches are deliberately omitted at BEATEN_MATRIX;
the score/bonus/single-player path continues without high-score UI or its award.

## Silent jingle semantics

VERIFIED — Party Land's gameplay wait dependencies are jingle/module cues, not
completion of ordinary bumper PCM sound effects. The relevant paths are:

- Mystery: WAITJINGLE2 gates EOSNURR; WAIT_FOR_SPIN_TASK polls SNURR_READY.
- Crazy letter and collected extra ball: WAITJINGLE polls ReadyAnim.
- Arcade crazy/5M/1M/500K prizes: START_DROP_WHEN_READY polls ReadyLogic.
- Jackpot's repeat count of three: multiple module cues precede JINGLEREADY.
- Matrix `_JINGLE` commands in bonus calculation replace a cue stream without
  suspending score arithmetic; new-ball reset sets both ready flags true.

VERIFIED — DOPLAYJINGLE clears ReadyAnim/ReadyLogic on acceptance, updates current
priority, repeat count and saved return position. JINGLE_HANDLER decrements the
byte: positive counts continue; reaching zero restores LASTPRIORITY (one for the single-player GAME_MUSIC path),
sets both ready flags and returns LASTJINGLE; a negative signed byte clamps back
to zero without declaring completion. This includes 0->255 for a looping theme.

VERIFIED — the **ASM** swaps JINGLEJUMPCNT before its priority comparison, so a
rejected request can change the pending repeat count. The supplied linked PRG
instead swaps after the comparison. PF4.5 follows the requested primary ASM
semantics and tests this quirk; it does not silently mix the two revisions.

INFERRED — deterministic silent cue timing: use the module's Fxx row speeds,
follow Dxx continuation and Bxx target, and count the cue row's duration before
calling JINGLE_HANDLER. Assume the conventional 50-Hz tracker tick (125-BPM
base). Time is integer rational state: each video sync adds 50 units; each tracker
tick consumes 71 units. The silent cursor advances at Dxx/order boundaries so forced jingles save the
current order, rather than merely the initial cue entry. Residual time survives
a Bxx callback; an accepted forced
position resets the cue phase. No floating point, sample decoding or audio output
is involved. Callback ordering is before that sync's physics/electronics scan.

UNKNOWN — the omitted SDR source does not establish whether its user callback
fires at the beginning or end of the Bxx row, how forced-position phase is
preserved, or exactly where that callback falls relative to VBLANK. Cue-row
choice can shift a DOS comparison by a row duration, not merely one video tick.
The silent reference traces explicitly test the stated convention. Recovering
player control-flow evidence could resolve this without implementing audible
playback; this is not hidden behind a claim that PF5 sound is already needed.

VERIFIED — START_DROP_WHEN_READY decrements the upper bound first, decrements the
minimum while nonzero, and latches completion so a subsequent jingle clearing
ReadyLogic cannot prolong the minimum wait. The four min/max pairs are crazy
140/180, 5M 110/140, 1M 120/150, 500K 45/70. SPECIALMODE bypasses; expiry of the upper
bound releases even without a callback. Release clears ReadyLogic and assigns
HANGSAVER=65535, HANGSAVER2=10. It then adds the source drop/camera tasks.

INTENTIONALLY DEFERRED — generation of audible samples/music and mixing/hardware
services. Sound intents remain `Sound`; accepted jingle intents remain `Music`.
`AudioCue`, `AudioComplete`, `AudioRejected`, `TaskReady`, `MatrixStarted`,
`MatrixCommand`, `MatrixEnded` and existing score/lamp/ball events carry explicit
native Tick values. Ordinary PCM sound intents have no fabricated gameplay wakeup.

## Exact reference traces

VERIFIED — these are direct semantic rule entries at tick zero, with a fresh
scroll phase and no budget loss. They are source/data-derived native timelines;
audio-dependent timestamps use the INFERRED silent convention above.

| Tick | Happy Hour command | Mega Laugh command |
|---:|---|---|
| 0 | EOSNURR | EOSNURR |
| 1 | CLEAR4 | CLEAR4 |
| 6 | HAPPY animation | MEGAL animation |
| 98 | CLEAR4 | — |
| 103 | ALL_TARGA scroll | — |
| 165 | — | CLEAR4 |
| 170 | — | PRINT13 |
| 171 | — | COUNTDOWN installed |
| 172 | — | first decrement: 26→25 |
| 292 | PRINT13 | — |
| 293 | COUNTDOWN installed | — |
| 294 | first decrement: 26→25 | — |

Mystery:

```text
0:   MYSTERY accepted; CLEAR4 installed
5:   MYSTERY animation installed
230: WAITJINGLE2 installed
319: silent B30 cue; JINGLEREADY; matrix installs EOSNURR
320: WAIT_FOR_SPIN_TASK resumes; prize task executes in later slot
```

A matrix-only run without that prize's preemption installs CLEAR4 at 320 and ends
at 325. Hidden: SCROLL at 0, TSEND at 289, HIDDENTASK0 resumes on task scan 290.
A 500K prize alone receives B41 completion at 57; its 45/70 release task runs on
that same tick **after** AudioComplete. Early completion at 10 latches but releases
at 45; missing completion releases at 70.

Drain: direct drain at zero, score1000, bonus1000, multiplier2, cyclones2, no
mode totals/extra ball:

```text
0:   BallLost; ball_lostTS installed
160: BONUS_X_CALCS -> bonus2000
229: CALC_CYCLO -> bonus202000
432: DO_FLORPA awards1000
436: DO_FLORPA awards1000
440: DO_FLORPA awards100000
444: DO_FLORPA awards100000; final count delay starts at -10
520: NEW_BALL; reset tasks/waits; ball number2
525: SOUNDBRICKUPP semantic intent
571: SOUNDNEWBALL semantic intent
600: SETBALL finishes; Playing; score203000, bonus0
```

Source creation order is SOUNDNEWBALL, SETBALL, SOUNDBRICKUPP. Reset during a
running slot means slot zero's new SOUNDNEWBALL first executes on the next scan;
later slots can execute in the reset scan. A synthetic scheduler trace tests
A,B,C on tick1 and D on tick2, where C is added to a later slot and D to a vacated
already scanned slot. Match uses the matrix's initial cue setup followed by 22
banks spaced eleven scans, then the original 120 matrix wait before checking the
single player's tens digit. The deferred frontend is replaced by a frozen native
GameOver; a match resumes the source shoot-again/new-ball path.

## Review of every documented PF4 adaptation

This table supersedes PF4's historical “Known semantic differences” section.

| PF4 adaptation | PF4.5 classification | Current behavior / remaining qualification |
|---|---|---|
| Native random clock instead of MAIN-loop increments and uninitialized DOS state | UNKNOWN | Wrapped VBLANK addition is retained. MAIN-frequency increments and initial undefined bytes cannot be reconstructed from deterministic input ticks. Outcomes remain reproducible native outcomes, not identical DOS randomness. |
| Mystery animation + NOSOUND25 estimate | RESOLVED | Original matrix animation, WAITJINGLE2 audio-ready branch, EOSNURR, next task scan and prize-slot order execute. Silent B30 timing is explicitly INFERRED. |
| Arcade release always using HANGSAVER2 upper bounds | RESOLVED | Original minimum, maximum, readiness latch and callback-driven release execute; fallback remains only when no completion occurs. |
| Missing matrix priority/interruption and dispatch overhead | RESOLVED | ASM priority arbitration, replacement, effect ordering, shared EOTS, ordinary dispatch delays and tail-call branches are implemented. The linked/ASM repeat-swap revision difference is explicit. |
| HIDDEN estimated wait without scroll preemption | RESOLVED | Actual scroll substep/phase semantics and linked WAITTS EOTS polling, including interruption and suppression, execute. |
| Happy/Mega countdown beginning immediately | RESOLVED | Intro commands precede countdown; jackpot interruption/resume and pending-mode task timing are retained. |
| Omitted ball-loss clear/print/jingle dispatch delays | RESOLVED | Original single-player ball_lostTS controls multiplication, category additions, count-up, delays and change-ball scheduling. |
| Explicit hidden/tunnel/arcade hold for stable physical capture | UNKNOWN | PF4's hold protection remains. Source GROPC/GROPD/GROPB do not themselves set HOLDSTILL. Matching their capture geometry/entry hold requires a DOS physical trace; this milestone does not claim those protected trajectories match DOS. Snack/dragon hold writes and wait timing are source-established. |
| Collapsed camera wipe/scroll tasks | RESOLVED | Original slot occupancy, shared ScrollPosition, -5 cadence and screen-force release are represented. |
| Forced new-ball/drain camera display and PF3 viewport integration | NON-BEHAVIORAL | Existing native viewport convention remains; source SCREENFORCE2 and scanout hardware details are not fully recreated. Gameplay slot/timer effects are restored. |
| Full-charge keyboard launch and explicit jitter instead of original charging UI | UNKNOWN | Original charge UI/input timing is not reproduced. Supplied charge/jitter deterministically affect velocity and subsequent game state; equivalence of human launch sequences is not claimed. |
| Non-lamp PF1 CMAP conversion, amber HUD, omitted matrix DAC pulses/fades/layout | NON-BEHAVIORAL | Original-content fonts/lamps and live score state remain; matrix temporal state is independent of those pixels. |
| Missing SDR callback phase | PARTIALLY RESOLVED (PF5) | Static SP expansion establishes B-row entry callback and retained mixer residue. Exact ordering relative to game Sync and forced-return phase remain UNKNOWN. Existing deterministic fixtures are retained; see PF5 evidence. |
| No sound/music renderer | RESOLVED (PF5) | Native module/sample decoding, mixing and SDL output are implemented in PF5. Gameplay completion remains independent of host consumption. |
| Single-player branch, no multiplayer rotation/menu/entry/attract frontend | NON-BEHAVIORAL | Deliberate presentation/player-count scope boundary for the selected single-player loop. No equivalence for multiplayer/frontend branches is claimed. |
| Omitted high-score-dependent award branch | UNKNOWN | Whole-DOS-game progression involving high scores is not validated. This excluded branch can award extra balls; it is not a purely visual difference. High scores remain outside PF4.5 scope. |

## Validation and reproduction

VERIFIED — test suite, race checks, vet, build and SDL offscreen smoke test pass.
Unmodified PF0/PF1/PF2/PF3 and PF4 tests pass, including PF4's physical
awards, modes, bonus/held bonus, extra ball, game over, original-content rendering
and two-launch 1200-tick script. That script still produces score2,300,000 and ball2.

PF4.5 tests additionally check independent intro/mystery/hidden command timelines,
simultaneous runnable tasks, late-slot/early-slot insertion, shared wait counters,
priority/score ordering, jingle preemption/restart/repeat count, current-order
return position, completion before
task scan, arcade continuation, minimum/readiness latch/timeout release, EOTS
preemption, source lamp periods, matrix budget skips, matrix-before-late-physics,
wait-list reset, skill effect inhibition, pending-mode delay, jackpot countdown
resume, independent drain/new-ball checkpoints, every reachable effect
matrix, and an 8000-tick input/state/framebuffer determinism sequence. Expected
command traces and bonus checkpoints come from Python/source reasoning, not Go
snapshot generation. No audible output is used for testing.

Files created: `internal/partyland/timing.go`, `timing_data.go`, `timing_test.go`,
`tools/reference_pf45.py`, `analysis/pf45-timing-fixtures.json`, this document.
Files modified: Party Land `game.go`, `features.go`, `holes.go`, `flow.go`;
Party Land `regions.go` (source CLOSE1 return/matrix semantics);
physics `ball.go` (optional matrix/scroll hooks); PF4 evidence and README.
No original ASM, game content, PF0–PF4 fixtures or existing tests were modified.

Exact commands from the repository root:

```sh
python3 tools/reference_pf45.py
./tools/go.sh test -p=1 -count=1 ./...
./tools/go.sh test -p=1 -race ./internal/partyland ./internal/physics
./tools/go.sh vet ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -data-dir . -pf4
./bin/pinballfantasies -data-dir . -pf4 -pf4-script -ticks 1200 -png /tmp/pf45-script.png
SDL_VIDEODRIVER=offscreen ./bin/pinballfantasies -data-dir . -pf4 -duration 250ms
```

PF5 is implemented; its evidence and validation are documented in [pf5-audio.md](pf5-audio.md).
