# Matrix stabilization audit

The source audit is complete, including the final `_2_DEMO_MODE` BX-clobber
finding below. Integration requires the exact-commit validation gate. The full
`go test ./...` pass recorded in `/tmp/stabilization-source-full4.log` includes
the source-driven image-checkpoint reconciliations below. It predates the
latest table-local pseudo-command corrections, which have focused tests.
Earlier build/platform passes do not validate the eventual final commit.
No image fixture has been regenerated.

## Source evidence and implemented corrections

The oracle is the owner-supplied original DOS source and linked retail PRGs.
Neither is included in this document or added to the public repository.

- `FANTASIE.ASM:_FLASHON`, `MATRIX_BLINKOR`, `KILL_FLASHOR`, `MATRIXON_`:
  flash enable, phase, countdown and emitted palette are distinct state.
  The countdown is a wrapping word; zero speed is not an inactive sentinel.
  `KILL_FLASHOR` disables blinking and emits the on palette, preserving the
  countdown, speed and phase words. Matrix entry invokes it only after effect
  acceptance. `_MATRIXLGT` does not rewrite blink phase.
- `DO_THE_ANIMATIONS/ts_slut`: a completed routine with no `NEXT_A` invokes
  `KILL_FLASHOR`. A zero data word itself is not a cleanup command. An animation
  ending with another command queued does not automatically stop flashing.
- `DO_ELECTRONICS`: matrix blinking precedes keyboard/background tasks.
  All four table implementations now follow that order.
- Linked `ANIM` compares the old frame offset to the header's end offset before
  drawing, loads the old frame duration, and preserves the old offset for
  drawing when restarting a loop. Drawing occurs only at timer expiry. All
  four tables use the shared presentation helper for this operation.
- Linked `SCROLLE` calls its routine twice per source sync and tests the byte at
  `SI+20` before each draw. The terminator visit completes without redrawing.
  The phase byte persists across stream replacement. All four tables use the
  shared helper. Literal glyph AL/AH stores now update retained dot memory
  using source base addresses and plane selectors, including interrupted
  streams. Independent tests project raw VGA byte-address memory into dots
  to verify every extracted stream and inherited phase.
- `WAITRUT` decrements a word; `_WAIT n` consumes n visits, with zero wrapping
  to 65536. `rclear1/2/3/4` consume 1/17/81/5 visits respectively. Speed Devils'
  `_CLEAR1` had incorrectly shared `_CLEAR4`'s duration.
- `PLAND:_WAITifmulti` selects two visits for one player and its operand for
  multiple players. Party Land skipped the single-player wait; Stones used the
  single-player duration for all player counts. The PF4.5 reference interpreter
  now includes the source wait; its runtime timing data did not change.
- `PARTYRUT` does not decrement SI. `WHEN_NEW_BALL_RESET` retains its matrix
  state when `PARTYFLASH` is true. Party Land and Speed Devils now preserve this
  ownership until their source `CLOSE1`/party-off path replaces it.
- `WHEN_NEW_BALL_RESET/NONEWPL` dispatches `SHOWPLAYERSTS`, rather than instantly
  drawing its final panel. `NODOT` paints ordinary score on a subsequent source
  visit and can dispatch the player panel using `DO_SPEC_MATRIX`. All four
  tables now execute these transitions through their schedulers. `Frame()`
  composes the retained dot memory without synthesizing idle matrix content.
- `DO_THE_DOTMATRIX` calls its single `PRINTTASK` slot after
  `DO_THE_ANIMATIONS`, then resets it to `DUMRET`. Print handlers now queue
  drawing rather than painting immediately on a gameplay request; a later
  same-sync handler can replace that slot.
- Each table's ordinary `_CHANGE_PLAYER` path ends at `HU_`, tail-calling the
  following handler while `NEW_BALL_TASK` waits independently. Native code
  had ended the outgoing program immediately. Stones also inserted an extra
  visit at `_NEW_BALL2`. Both now follow the source tail-call ordering.
- `_PARTYOFF` only installs `WAITRUT` for one visit. Gameshow and Stones had
  additionally cleared `PARTYFLASH`; the flag belongs to gameplay exit paths.
- `_SETDECCOR` reads its operand into a word counter; it is not filler.
  The schema now types it as numeric. Regenerating presentation and Speed
  Devils content changed only the affected commands' `nums` fields.
  `WAITJINGLE2` now supports the original driver's no-sound capability bit
  and wrapping word countdown. Party Land's GROPB initializes it to 25.
  This capability is distinct from PCM muting or the ingame music option.
- `PLAND:SHOWPLAYERSTS` prints the ball line at source y=10; the former native
  frame renderer synthesized it at y=9. The existing aggregate native frame
  hashes therefore cannot be treated as original DOS matrix oracles.
- Frontend game-over replay follows the actual `_INIT_SCORE`, `_SHOW_SCORE`,
  `_LOOP_` and tail-jump paths. The Party Land single-player trace reaches
  `SHOWHIGHSTS` at sync 298; Speed Devils reaches it at 278. The former native
  handcrafted replay's boundary was 305.

## Party Land reproducer clarification

The clarified gameplay sequence is a side-lane award followed by a drain:
the heart starts, then a new round begins without text. It is distinct from
the arcade `SPINXB -> SIDELANETS` scroll.

`BYGEL1/BYGEL2` award `EXTRABALL2`, set light 51 and increment extra balls.
`LOOSE_BALL` explicitly resets effect priority before requesting `LOSTBALL`.
DOS thus permits `BALL_LOSTTS` to replace an unfinished heart award; it does
not require the award's eventual Extra Ball text to finish. `BALL_LOSTTS`
then has `CLEARIT`, a text print, `_WAIT 80`, and another clear before bonus
and player/extra-ball selection. The new regression checks both side lanes,
starting the heart before drain, and proves that the loss text/wait are not
bypassed. It does **not** establish the cause of the observed sub-second
new-round transition. No artificial hold was inserted.

The owner manually tested the diagnostic Windows build and confirmed the
heart animation, then `GETTING SICK HUH`, then the bonus countdown. That is
consistent with the source drain path. The test-build keys exercise physical
side-lane award/drain callbacks; they do not invoke matrix commands directly.
This confirms the current reproducer but does not identify which historical
condition caused the previously reported missing text.

Independently, the arcade task regression verifies that `SIDELANETS` reaches
its source terminator on matrix visit 197 before physical drop ejection.

## Temporal coverage currently added

Presentation tests cover flash word/phase/palette writes, zero-speed wrap,
routine completion versus a zero word, every decoded animation's source
frame/duration/loop schedule, every extracted scroll's cadence and terminator
with each inherited countdown phase, and a concrete game-over dispatch trace.

Each table has tests for its routine boundaries, jingle-ready polling,
accepted replacement, rejected effects preserving the display, timed new-ball
player-panel execution, and score restoration driven by source syncs. Party
Land additionally covers all five letter animations, persistent party flash,
the arcade scroll/ejection path and the clarified side-lane award/drain path.

Targeted transition suites passed all four tables
(`/tmp/raw-scroll-source.log`, `/tmp/tail-source-regressions.log`,
`/tmp/nosound-source-regressions.log`). Operand, reachability and numeric
checks pass. The complete suite passes after source-derived reconciliation.

## Five image-checkpoint classifications

| Checkpoint | Classification | Source/pixel evidence |
| --- | --- | --- |
| Party Land deterministic 1200-sync image | **PROVEN OLD FIXTURE WRONG** | `PLAND:SHOWPLAYERSTS` prints BALL at y=10 after selection closes. Reconstructing the old invented PLAYERS line at y=9 reproduces the entire old SHA `f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e`; source panel dots and retained matrix illumination are checked independently. |
| Party Land initial full-table chute | **PROVEN OLD FIXTURE WRONG** | `LATE_RASTER_INTERRUPT_DEMO` dispatches `FIRST_NO_OF_PLAYERSTS`, retaining VGA until a routine source visit. This checkpoint occurs before that visit. Reconstructing the old synthetic panel reproduces its complete original image hash. |
| Speed Devils initial full-table chute | **PROVEN OLD FIXTURE WRONG** | Same common interrupt/retained-memory path; complete old hash reproduced by synthetic panel only. |
| Billion Dollar Gameshow initial full-table chute | **PROVEN OLD FIXTURE WRONG** | Same common interrupt/retained-memory path; complete old hash reproduced by synthetic panel only. |
| Stones 'N Bones initial full-table chute | **PROVEN OLD FIXTURE WRONG** | Same common interrupt/retained-memory path; complete old hash reproduced by synthetic BALL panel only. |

The tests retain the historical hashes as checks on unchanged non-matrix
pixels and check current matrix pixels against source operations separately.
Eight additional capture/forced-camera checkpoints had also synthesized an
idle panel without a source sync after explicit bitmap writes; retained VGA
composition now has a direct pixel check, with the historical playfield
hash check preserved. No image fixture was replaced.

## Additional source/pixel evidence

- Linked CODE2 skips leading BCD zero cells and uses a visible-digit cache.
  `UPDAT_SCORE` invalidates this cache and comma count. Tests cover unchanged
  cached digits/commas, invalidation, and the zero-group comma-cache exit.
- Common `HI_1..HI_4` rank prefixes had not been extracted. The extractor now
  resolves their PRG address records. Common native cheat-scroll programs
  likewise use PRG addresses, not embedded original string bytes.
- Linked PRINT_NUMBER's comma stores differ from CODE2's comma stores:
  `SI=DI+height*168+35`, with separate plane4/plane1 byte writes. Independent
  raw-address projection tests now verify their four dots on all four tables.
- `RESET_NUMBERBUF` does not clear pixels. The former native numeric-field
  clear was wrong: DO_FLORPA's final zero print skips BCD cells and retains the
  previous digit during its 14-visit tail. NO_MORE_NUFFROR performs a bounded
  CLEAR_BOX2 afterward. Native number rendering now follows those writes.
- `SDEV:_turnonturbo` always tail-jumps TURBOTS in the same dispatch, even if
  GetTheGoalYouFool defers gameplay via Goliat. Native code had returned.
- `STONES:GROPA/_JACKEND` use saved TOWERHUNTMODEORIG, not the live mode flag.
  Mode expiry during the award must not change the captured gate decision.
- `_SHOOT_AGAIN_ONN` updates a player digit and installs WAITRUT1. It does
  not award a lamp. Extra writes were removed from Party Land and Speed Devils.

## Dynamic interrupted-scroll oracle captures

Owned original DOS payloads and screenshots remain private/ignored. The
opt-in `TestExportScrollOracleStates` exports source-sync states; the new
`tools/matrix_oracle_compare.py` compares each sampled matrix dot to these
states. These are pixel comparisons, not wall-clock timing assertions.

| Table | Nonblank frames before interruption | Nonblank frames after interruption | Unmatched frames |
| --- | ---: | ---: | ---: |
| Party Land | 496 | 704 | 0 |
| Speed Devils | 499 | 711 | 0 |
| Billion Dollar Gameshow | 362 | 588 | 0 |
| Stones 'N Bones | 342 | 557 | 0 |

Both sides include scroll frames and high-score/curtain/text states. Captures
are under `/tmp/pf-matrix-interrupted-table1-c` and corresponding table2/3/4
folders; comparison JSONs are under `/tmp`. The set of inherited phases is
covered by the independent source-time tests. The pixel-state membership
check alone does not establish which phase was interrupted or exact sync
alignment; finish that trace analysis before closing this audit item.

## Audit closure and integration gate

The source discrepancies listed at handoff have source-derived regressions:
interrupted scrolls and one-player startup have private ordered DOS comparisons;
COUNTDOWN, NODOT reentry, all four bonus clears and source selection flags have
boundary tests; NOSOUND has a complete caller audit; original cheats have
positive, negative, activation-condition and physical-effect tests. The later
BX-clobber finding below resolves the remaining non-qualifying game-over path.
The final integration gate is the exact-commit Go/extraction/oracle/build/
platform/storage/payload rerun, followed by safe remote reconciliation.

## Win32 cursor work

Raw Input remains the relative-motion source. Client `WM_SETCURSOR` handling
selects a null cursor only while the focused window's mouse-plunger mode is
active. Mode and focus transitions refresh a stationary client pointer.
There is no production capture, confinement, pointer warp or `ShowCursor`
counter manipulation. Linux mouse code is unchanged. Focus/fullscreen/raw
mouse/cursor Wine checks passed before the most recent matrix-only changes;
the final validation must still rerun the requested complete set.

## Continuation audit: source boundaries and retained CODE2 writes

The initial 85 modified/untracked files were preserved before editing in a
private `/tmp` archive; no checkout, reset or clean was performed. A fresh full
Go suite passed at continuation start. That pass predates the corrections below
and is not final validation.

| Scope | Previous native behavior | DOS evidence and correction | Regression |
| --- | --- | --- | --- |
| Shared countdown/all tables | Convenience ModeTime decrement and simultaneous seconds/value drawing | FANTASIE `_COUNTDOWN`, `_COUNTDOWNCONTINUE`, `_COUNTDOWN2`, `COUNTDOWN`: SEC_ASC starts one higher, SYNC_LEFT starts 1, reloads 71, seconds print replaces value print only on second changes, zero is displayed for 71 visits before uninstall. CONTINUE retains seconds and resets phase. Added shared source state, retained text and exclusive PRINTTASK. | presentation/countdown_test.go; table expiry tests; numeric pixel test |
| All tables idle | A newly installed idle panel consumed its first clear normally; CHECKHIGHSCORE reentry was not handled | `NODOT` returns SI=0 after `DO_SPEC_MATRIX`/`DO_MATRIX`; `DO_THE_ANIMATIONS` dispatches NEXT_A immediately. Both panel and high-score installation reenter on this sync. | TestIdleScoreIsDrivenBySourceSync; TestIdleHighScoreReturnsZeroAndReenters |
| Speed Devils supermode return | Invented a one-visit wait on false inhibit branch | SDEV `INGA_KONSTIGHETER -> NORMAL_END` preserves inherited NODOT/SISA=0; true branch tail-jumps BACK_2_TURBOTS. | TestSupermodeReturnFalsePreservesIdleRoutine |
| Shared CODE2 font | Full 8x16 cell overwrite | Linked CODE2 glyphs write 98 stores (7 columns x14 rows). The eighth column and two lower rows are retained. Decode literal VGA stores rather than paint a bounding box. | TestCODE2RetainedCellsCacheAndWidthTransitions |
| Shared CODE2 commas | Three dots at rows 13,14,15 | Linked comma routine uses SI=BX+0a1bh-200; plane4 writes SI and SI+168, plane1 writes SI+1 and SI+168. Four dots occupy rows14,15. Cache and zero-group early exit remain distinct from PRINT_NUMBER. | Independent raw-address projection in score_retention_test.go; numeric_test.go |
| All-table bonus tail | Last digit retention verified without subsequent erase boundary | `NO_MORE_NUFFROR` clears CLEAR_BOX2 only after the 14-visit final print tail. Added bounded pixel verification on that visit. | bonus_countdown_test.go in all four tables |
| Shared match text | Whole display clear; immediate next-digit draw | KNACKRUT1 prints the full retained LAST_TEXT buffer only in its top five rows. KNACKRUT2 erases the old digit immediately, assigns the new digit to PRINTTASK, and allows later dispatch to replace it. | presentation/match_temporal_test.go |
| Stones match | KNACKRUT1 executed at handler dispatch; updates only on digit changes | STONES `_KNACKET` installs KNACKRUT1 for its next visit; KNACKRUT2 visits every sync. First timeout22, source pointer reload includes the first22 again, then remaining matchTimes. | TestMatchSourceFirstAndLastVisits in all four tables |
| NOSOUND | Caller audit incomplete | Search of supplied ASM/MAC finds capability INT66 function21 bit3 only at WAITJINGLE2. User MUSIC_TOGGLE/S_EMPTY and nil/silent PCM do not set it; virtual tracker readiness still progresses. | TestSilentPCMIsNotDOSNoSoundCapability in all four tables; existing driver-bit wait boundary tests |
| Global original cheats | Positive sequences covered; missing negative/physics cadence coverage | FANTASIE CHEATS/CHECKCHEAT and TILTRUT/SNAILRUT/BALLSRUT/FAIRPLAYRUT are common loaded-table routines. Added every-position wrong-sequence checks and unobstructed VBLANK/late physics-step verification. | TestSourceCheatNearMisses; TestFairPlaySourcePhysicsCadence |

The original five image checkpoints remain **PROVEN OLD FIXTURE WRONG**.
No replacement hashes were accepted. Historical rendering is reconstructed
explicitly for its old hash checks, including its former full score-cell and
three-dot comma writes; independent source-derived pixel tests validate current
retained rendering. The four table aggregate fixtures also preserve their
unchanged non-matrix pixels. CODE2 evidence above explains additional score
pixels in Gameshow/Stones historical camera checkpoints.

## Continuation oracle trace evidence

`tools/matrix_oracle_compare.py` now requires each nonblank capture to fit a
monotonic source-sync trace with one inherited phase across the entire group.
It rejects backward frame order and phase switches; synthetic tests verify both.
All four interrupted-scroll capture groups still have zero unmatched frames
and now also admit an ordered same-phase trace. Captures remain undersampled:
several phases/ticks can produce identical pixels, so these captures do not
prove a unique interruption sync. Independent source tests establish exact
phase retention, terminator ordering, two executions per source sync and
retained dot memory.

Opt-in `TestExportStartupOracleStates` seeds FIRST_NO_OF_PLAYERSTS with each of
several privately captured predecessor VGA grids and holds WAIT_GAME_ON.
All 80 first-player-start captures per table (320 total) have zero unmatched
frames and an ordered retained-memory trace. Native startup already covers
one through eight players; this adds original-DOS one-player pixel evidence.
No original grids/screenshots were added to the repository.

## Continuation lifecycle review (superseded handoff details below)

The frontend's former six-sync GameEnd pause belongs to FANTASIE
`end_gamen -> to_demolation_derby` (keyboard game abort), not match completion.
PLAND/SDEV `_CHECK_XXBALLS` and `_KOLLA_XXBALL` must tail-jump AFTER_XXBALLTS,
not end the session. SHOW/STONES false `_KOLLA_XXBALL` branches use HU_ on the
same sync. All four source AFTER_XXBALL programs clear before installing
SPINTSEL_IN_HIGH. Source CLEARIT is linked `_ANIMATION _CLEAR` on SHOW/STONES,
with four one-visit frames and the final terminator visit.

A source-entry handoff now retains the installed `_CHECK_HIGH` routine while
frontend storage/keyboard logic visits one player per sync. Qualification
prints HAJJSKAR at DI=336 over retained VGA; GET_IT_FROM_KEYBOARD clears queued
SCAN_CODE on the following visit; READ_KEYBOARDET consumes one scan-code slot.
WAIT_A_LITTLE prints STJAERNOR at counter30 and switches back to SPINTSEL at
counter2. Its existing 58-visit boundary is therefore correct and preserved.
Completion resumes `_MATRIXLGT` on the same sync, then `_2_DEMO_MODE` adds the
demo task and tail-dispatches the following command. The next task scan changes
demo state before visiting that installed matrix routine.

The frontend carries a persistent retained-memory continuation. The initial
implementation incorrectly retained the adjacent WAIT20000/WAIT30000 data.
The BX-clobber finding below supersedes that interpretation: DEMOMODE NODOT
starts AFTERDEMOMODETS without installing those waits. Source
`TO_DEMO_FROM_GAME` turns off lamps (Party relights52,53,54), forces the bottom
raster, and ZEROSCORE runs only on Party/SDEV/SHOW. Completed PLAYER_AREA scores
remain available to the frontend timeline. Runtime PCM reads the single session
sync result without advancing presentation audio a second time on a handoff.

These lifecycle changes still need complete regression and source review;
final validation and integration remain outstanding. No commit was pushed or
merged during this continuation.

### Additional lifecycle boundaries resolved

- PLAND/SDEV/SHOW match-round earned Extra Ball branches of `_KOLLA_XXBALL`
  jump directly to SHOOT_AGAIN_ONTS after VARS_2_P_STRUC; they do not reach the
  ordinary bonus/player wait. Party and Stones decrement their original
  counted Extra Ball state. `_KOLLA_XXBALL` does not restore HOLDBONUS; that
  copy belongs to `_CHANGE_PLAYER`. Tests: all-table
  TestMatchEarnedExtraBallOwnsImmediateReplacement.
- `_CHANGE_PLAYER`'s LET_HIM_SHOOT_AGAIN/NO_MORE_BALLS exits are direct source
  tail jumps, not DO_MATRIX requests. They preserve the current flash palette
  until the subsequent source handler changes it. Native StartMatrix(reset=true)
  had incorrectly invoked KILL_FLASHOR here. Tests: all-table
  TestChangePlayerTailJumpRetainsFlashPalette.
- The frontend previously checked ADDPLAYERS after processing make codes.
  Physical CLOSE1 could therefore close selection during one sync and still
  allow F1..F8 on the next. READ_KEYBOARD's source gate must be observed before
  key dispatch. TestSourceSelectionClosesOnActualLaunch uses an actual charge32
  release and physical travel, across all tables, rather than assigning a
  synthetic selection flag.
- TestSourceScoreEntryVisitsAndRetainedPixels covers one-player-per-visit
  qualification, the separate keyboard setup visit, single SCAN_CODE slot,
  HAJJSKAR retained rows, STJAERNOR at counter30 and the counter2 restart on
  all four tables. TestGameOverContinuationKeepsSourceWaitAndMemory checks each
  source post-demo wait, retained phase/memory, and Stones termination/NODOT
  reentry. TestLostMatchRunsSourceScoreEntryClear covers all four clear/entry
  boundaries. These pass alongside the full Go suite.

The earlier sync298/278 game-over numbers refer to the separately tested
AFTERDEMOMODETS helper. They are not the production match-to-demo starting
point on SDEV/SHOW/STONES: the real continuation includes the source long wait
above. No fixture or wait was changed to mimic the former native helper.

### Validation progress

A fresh full `go test ./...` and `-tags matrixdebug` frontend/Party suites are
green after the source changes. Matrix reachability, operand, numeric and
public-source checks pass. PF8/PF7/PF4.5 regeneration reproduces identical
presentation, Speed and Party timing generated files (SHA256 check).
The six extractor/oracle-comparator Python unit tests pass.

Linux and Windows amd64/CGO=0 builds, public Windows/AppImage and personal
Windows/AppImage builds have passed. Personal inventory validated all12 inputs,
all four factory high-score save/restart/reset flows and ignored payload paths.
The separate Windows matrix diagnostic package built successfully.

Linux relative mouse and Win32 host/layout/sided-modifier/focus-pause/raw mouse/
cursor/fullscreen-music tests pass on an isolated1280x1024 Xvfb display with a
fresh private Wine prefix. Attempts on the active desktop exposed focus
interference; no assertions were weakened. SDL's geometry fullscreen journey
requires a window manager, absent from bare Xvfb, and is run separately on the
managed desktop. Visible real-Windows OS cursor acceptance remains recommended.

Package/storage journeys and exact-final-commit reruns are still outstanding.
These progress runs are not a substitute for final-commit validation.

### Investigation: Stones non-qualifying game-over blank matrix (resolved below)

The user reports an empty matrix after a non-qualifying Stones game. This is
initially **UNRESOLVED** despite the full Go suite passing; resolved by the
BX-clobber finding below.
The new retained frontend continuation preserves STONES AFTER_XXBALLSTS
`_MATRIXLGT 0 -> _2_DEMO_MODE -> _WAIT 30000 -> 0`. The linked retail table
confirms that stream: WAIT operand at file0x18756, `_2_DEMO_MODE` CS0x0b5c
adds TO_DEMO_FROM_GAME CS0x0b74, and WAITRUT CS0x5a31 decrements SI.
GO_DEMO_MODE CS0x67a6 only changes mode/keyboard flags. These addresses are
diagnostic metadata; no private bytes were copied into public source.

A first unmodified DOS gameplay capture at /tmp/pf-stones-nohigh-real did not
establish terminal game over: its final visible matrix was TILT. It therefore
cannot validate either the native blank interval or an immediate AFTERDEMOMODETS
handoff. A second capture includes full-frame drain checkpoints to establish
ball progression. The earlier continuation claim that the long waits were
fully validated was premature; only their static operands and native tests
were established. No wait or expected image has been changed to hide this issue.

Party high-score guard review: FANTASIE CHECKHIGHSCORE uses I_UTSKJUT, not
SPRING_VALID, and _BEATEN_MATRIX omits chute/special-mode guards. Party now
uses inChute for the idle guard and a separate source beaten comparison for
_BEATEN_MATRIX. TestHighScoreGuardUsesSourceChuteFlag covers both. The drain
branch of TestHighScoreBranchExtraBall seeds a nonzero bonus because PLAND
JBCDZ BONUSSIFFRORNA bypasses _BEATEN_MATRIX when bonus is zero. Both branches
and the latest full Go suite pass.

Validation correction: the initial Wine journey used a personal GUI diagnostic
build, whose GUIMode suppresses the stdout transition log required by
smoke_pf12_wine. With the current console matrixdebug executable, both the
four-table OFF/NORMAL/HIGH journey and fullscreen/music journey pass
(/tmp/continuation-wine-journey2.log, fullscreen2.log). No input assertion or
implementation was weakened. This still predates an eventual validated commit.

The user confirms original DOS displays GAME OVER promptly after a
non-qualifying game. Treat the native long blank as an implementation bug,
not an accepted DOS defect. The exact source replacement/control-flow owner
remains unresolved; static presence of WAIT30000 does not establish that
it survives the real handoff. Repeated automated DOS launch probes have not
yet left the chute, so they provide no terminal timing proof. A physical
XTest keyboard probe is running to distinguish posted-event input from
actual keyboard makes. All four Linux and all four Wine live hotseat
scenarios now pass with the correct platform test binaries
(/tmp/continuation-hotseat-live2.log).

### Wrapping wait and validation-input repair

Stones used a signed native countdown for `_WAIT 0`, completing on the first
visit. FANTASIE WAITRUT decrements a 16-bit SI before testing completion, so zero
requires 65,536 visits and negative/overflowing operands wrap. Shared
WordWaitTicks now applies that conversion to Stones and frontend projections.
TestWaitUsesWrappingSourceWord and TestContinuationRetainsZeroWaitWord cover
first/last visits. The current full Go suite passes after these corrections.

Live runtime journeys changed the owner's configuration to three balls. The
ignored working-copy configuration formerly shared that owner's symlink; it
now uses a separate copy of the pinned inventory fixture. The owner's current
configuration was preserved untouched. PF10 generation had stopped after
writing an incomplete analysis file because its private TABLE4.HI input was
absent. Restoring that ignored owner input and completing generation reproduced
the existing generated content exactly, apart from its header comment (omitted
from the intended diff). No hash expectation was changed for either issue.

### Resolved: `_2_DEMO_MODE` does not preserve BX

Shared/all four tables. Previous native `_2_DEMO_MODE` installed the next
matrix command through HU_, yielding an empty Stones matrix for 30,000 syncs.
FANTASIE.MAC ADDTASK sets DX and calls DOADDTASK without saving BX.
FANTASIE.ASM DOADDTASK returns BX at its allocated TASKLIST slot. Every table's
`_2_DEMO_MODE` calls ADDTASK without the PUSH BX / POP BX present in `_NEW_BALL2`.
HU_ adds two to this task-slot address and tail-calls that entry, normally
DUMRET. DO_THE_ANIMATIONS had installed NODOT before dispatching this handler;
DUMRET does not install a wait. The next demo task scan changes DEMOMODE, and
NODOT dispatches AFTERDEMOMODETS in that same source sync. The adjacent matrix
WAIT20000/WAIT30000 values are real data but unreachable through this handoff.

Linked Stones confirms `_2_DEMO_MODE` CS0x0b5c calls DOADDTASK CS0x62b2,
which starts BX=DS0x386f and returns it at the allocated slot, then jumps to
HU_ CS0x0b04. No BX restoration exists. A private data-only trigger replacing
FIRST_NO_OF_PLAYERSTS WAIT_GAME_ON with the existing JMP to AFTER_XXBALLSTS
DS0x2076 produced prompt GAME OVER and the subsequent score/high-score/scroll
cycle in DOS (/tmp/pf-stones-nohigh-afterxx). All executable code and target
program operands were unchanged. This controlled probe confirms the handoff,
not physical drains; the owner's unmodified DOS observation supplies that
independent gameplay evidence. Original bytes and screenshots remain ignored.

All four native handlers now follow the returned task slot and preserve VGA
memory. The frontend resumes NODOT on the demo-task sync, then the source
MATRIXLGT/WAIT20/SETLOOP/INIT_SCORE/UrbanOver path. Tests:
TestDemoCommandUsesTaskListTailAndPreservesMatrix (all four tables, empty and
occupied next task slot); TestGameOverContinuationUsesDemoNodotAndRetainsMemory
(all four, retained pixels, wait boundary and prompt text). The older
TestGameOverContinuationKeepsSourceWaitAndMemory test encoded the incorrect
BX-preservation assumption and has been replaced. This resolves the reported
Stones blank-matrix bug; final complete validation remains mandatory.

The private ending oracle exporter compares the corrected persistent timeline
against the controlled DOS capture. All 899 nonblank frames match with no
source-order failure (225 additional blank illumination frames); one inherited
scroll phase fits the entire trace. Capture undersampling admits phases2/4/6/8
and several final syncs, so exact durations come from source boundary tests.
TestNonQualifyingGameShowsSourceGameOverPromptly additionally exercises all
four real native drain/bonus/match/optional match-turn/qualification/demo paths
and checks GAME OVER within its source-defined first 40 demo visits.
