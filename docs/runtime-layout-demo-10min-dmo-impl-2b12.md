# DMO-IMPL-2B12: complete SHOW_HI_ETC presentation transition

2026-10-08. Owner approval: DMO-IMPL-2B12 ONLY.
Accepted predecessor: [2B11](runtime-layout-demo-10min-dmo-impl-2b11.md).
Branch main, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e` unchanged.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_SHOW_HI_ETC = READY

DEMO_FRESH_POST_EQUALITY_CONTINUATION = PARTIAL

READY covers the complete bounded source branch, including all ShowInfoTS
consumers and its return, and the actual fresh spring-guard return. It does
**not** claim fresh reachability of the installed ShowInfoTS program.
The unchanged fresh TARGET completes36725 and continues through50000 without
an unsupported consumer. PARTIAL records that the requested next actual
unsupported boundary has not been reached within this bounded suffix. There is
no missing ShowInfoTS consumer and no invented refusal at50001. No claim of
unrestricted continuation, impossibility or eventual termination follows.

## Actual fresh branch: SpringValid is true

The2B11 refusal occurred before inspecting SHOW_HI_ETC's own guard. Admission of
that guard changes the verdict at36725, without changing its incoming state:

- NODOT file0x56ac requires enabled keyboard;0x56b7 compares the actual word
  NODOTCOUNT at DS:37f5 against720, and0x56bd selects CS:5406/file0x5706.
  This equality precedes AFTER_CHEAT, DEMOMODE and ordinary panel guards.
- SHOW_HI_ETC0x5706 tests SPRING_VALID at DS:3486 against true=ff;
 0x570b jumps to CS:543e/file0x573e when true.
- In the fresh replay, SpringValid really is true. No FJANTTEXT, PLAY_TEXT,
  DO_MATRIX, flash reset, matrix replacement or high-score presentation occurs.
- ONLY_SCORE0x573e executes the admitted CHECKHIGHSCORE guard/return; the actual
  I_UTSKJUT remains true.0x5741 increments NODOTCOUNT720→721. PEKOR renders the
  retained score50030 at200,0x5752 returns SI=0, and late physics completes.
- The current inactive matrix/cursor and dot memory survive that visit.
  Normal audioTick still runs before electronics; the test compares its separate
  MusicClock.Sync projection instead of falsely freezing audio for a calculation.
  Expired remains true. No displaced expiry cursor is restored.

This is a source-defined continuation, not a reason to force SpringValid=false,
relaunch the ball, change TARGET, reset inactivity, or seed presentation state.

## Full source-consistent false-guard branch

Historical FANTASIE.ASM NODOT/SHOW_HI_ETC/DO_MATRIX and PLAND.ASM
ShowInfoTS/showithi provide the bounded source basis. The new checker extends
2B11 and verifies the actual linked stream0x1b6ac..0x1b795, its handlers,
operands, text correspondence and new roll-handler control operands. It does
not start trajectory, graph, writer-domain or IRQ research.

On SPRING_VALID!=ff, linked0x5710 writes FJANTTEXT word DS:37f7=00ff;
0x5716..0x5718 writes PLAY_TEXT[8]='7'+1;0x571b..0x571e copies the actual
BALLSTEXT[5] at DS:238a into PLAY_TEXT[18].0x5721 supplies BX=DS:18fc and
0x5724 calls DO_MATRIX CS:4501/file0x4801, selecting ShowInfoTS file0x1b6ac.
0x5727 returns to ONLY_SCORE, rather than to ShowHighsTS.

DO_MATRIX kills flashes, clears DOT_READY, resets NODOTCOUNT and sets
BEHOVS_PROVAD before its DO_SPEC_MATRIX tail dispatch. It is a direct request:
INH_EFF, SPECIALMODE, audio priority and sticky expired are not extra matrix
eligibility guards. No jingle is requested. CHECKHIGHSCORE remains admitted
before its possible side effects; reward remains unsupported.

DO_MATRIX first dispatches CLEAR4 at0x1b6ac. NODOT then returns SI=0, so NEXT_A
dispatches the second CLEAR4 at0x1b6ae in that same visit; the first clear does
not run a WAITRUT visit. Structural tests check current op, next cursor and
remaining5 at this boundary. Subsequent dispatch uses the existing interpreter,
Display/PRINTTASK, fonts, matrix budget and shared scheduler/WAITLIST.

| Linked stream | Actual consumers |
| --- | --- |
| 0x1b6ac,0x1b6ae | Two CLEAR4 dispatches; only the second routine executes |
| 0x1b6b0 | RULLGARDIN_NED PLAY_TEXT, stop1;14 visits |
| 0x1b6b6..0x1b6cc | WAIT120, MATRIXLGT0, CLEAR4, PRINT13 JACK_TEXT/336, PRINT13_NUMBER JACKVALUE/368, FLASHON1 |
| 0x1b6d0..0x1b6e4 | WAIT120, FLASHOFF1, CLEAR4, RULLGARDIN_UPP BONUS_TEXT/stop1 (15 visits), WAIT120, CLEAR4 |
| 0x1b6e6..0x1b6f0 | PRINT13 ALLTIME_TEXT/338, WAIT40, CLEAR4 |
| 0x1b6f2..0x1b781 | Four expanded rank groups, in source order |
| Each rank group | FLASHOFF1, MATRIXLGT0, PRINT13 rank/336, PRINT13_NUMBER record digits/368, PRINT13 record name/344, FLASHON1, WAIT140, CLEAR4 |
| 0x1b782..0x1b794 | FLASHOFF1, CLEAR4, PRINT5 PLAYERSTEXT/336, PRINT5 BALLSTEXT/1684, zero terminator |

There are55 linked command records, including the overwritten first CLEAR4 and
zero terminator. The remaining routine execution takes1072 matrix visits.
PRINTTASK paints on dispatch in its actual scheduler visit; each new print is
compared against independently constructed glyph/number output and raw position.
CLEAR4 uses five visits. FLASH and MATRIXLGT have one WAITRUT visit; WAIT120/140
use their actual counts. The roll handlers start Y=-13/16, increment/decrement
before drawing, erase the trailing row and stop at the linked stop value1.
Their completion returns SI=0 and resets clipping; normal NEXT_A continues.
There is no animation, SCROLL, FADE, task insertion, task wait or nested jump in
this stream. Reusing their unrelated consumers is unnecessary.

The terminator ends the current program and kills flash through the shared
FinishRoutine path. BEHOVS_PROVAD remains pending. On the next idle visit outside
I_UTSKJUT, the accepted NODOT panel uses DO_SPEC_MATRIX, consumes that pending
flag and takes its same-visit panel continuation. In the chute, that panel guard
continues to skip the request. No alternate high-score lifecycle is introduced.

Replacement uses ordinary owned matrix installation and source guards. A new
request interrupts the current program without preserving its cursor; flashing
is killed by DO_MATRIX, dot memory is retained until consumers overwrite it.
Sticky expired neither vetoes ShowInfoTS nor gets cleared by it. A structurally
independent expiry replacement still runs its ordinary stream through QUIT.

## Data provenance and admission

The accepted `verified-native-factory-volatile` policy now supplies all four
records. Each is16 bytes: twelve unpacked BCD digits, three source-encoded name
bytes, and a zero terminator. Linked DS:0016/0026/0036/0046 supply the records;
name fields are respectively+12. Scores are50,000,000 /25,000,000 /10,000,000 /
5,000,000. The2B10 predecessor verifies all64 compiled seed bytes against the
native factory seed and pristine inventory identity. No DOS .HI file is opened.
The top-score comparator retains its own admitted Decimal; canonical
Game.highScore=nil is never a data source.

All records, names and presentation text are copied from pinned owner-local
inputs into owned memory before replay. An owned in-memory seal prevents
changed record names from acquiring admission together with changed display
bytes. Each digit/name consumer rechecks policy, record identity and its actual
input before dispatch. Missing, persistent/unverified or changed records refuse.
The four name bytes and glyph payload are not copied into Git or this report.

PLAY_TEXT DS:2339, JACK_TEXT DS:231a, BONUS_TEXT DS:2325, ALLTIME_TEXT DS:21e6,
rank texts DS:3733/373b/3743/374b and player/ball inputs are verified linked
operands. Static presentation texts match canonical private text references;
player/ball use the predecessor's actual demo inputs. SourceText performs the
existing source encoding conversion before glyph painting. PLAY_TEXT's two
mutable bytes use the linked stores; BONUS_TEXT retains the accepted NEW_BALL
byte11 reset. Invalid/missing inputs or changed font5/font13 admission refuse.
Content maps and byte slices are owned to prevent cross-candidate mutation.

JACKVALUE is the live shared Jackpot Decimal, not the zero bytes in the raw
uninitialized DS field. Historical WHEN_NEW_GAME_RESET_TABLE's MOVEBCD and
linked0x316..0x321 copy JACKINIT DS:0124 to JACKVALUE DS:010c, length12;
the checker verifies its10,000,000 initializer against the native initial value.
The selected TARGET has no jackpot writer; it retains that value. Printing reads
the current Decimal and rejects invalid BCD. No score/bonus/ball state is
supplied by expected frames or reference execution.

## Fresh execution and comparison boundary

TARGET remains Down35438..35459; Release35460; d=calculation-35460;
Left iff `(d+46)%52<8`, Right iff `(d+22)%30<21`.
Fresh construction has zero score/timer/clock, no tasks or active program and the
accepted volatile factory inputs. All earlier calculations execute. No count,
ball, matrix, score, timer or RNG injection feeds the replay. Canonical physics
callbacks remain poisoned. Structural setups are separate and are not witnesses.

| Calculation | Fresh observation |
| ---: | --- |
| 35877 | Accepted actual scored drain,50030/bonus0, LOSTBALL and SOUNDRINNER |
| 35967 | Accepted five zero-aggregate branches and NEW_BALL_TASK producer |
| 35998 | Complete expiry installation, due NEW_BALL replacement and late physics |
| 36724 | Complete; actual NODOTCOUNT720, SpringValid=true, expired=true |
| 36725 | Complete; SHOW_HI_ETC spring guard returns, NODOTCOUNT721; no ShowInfoTS installation |
| 37047 | Ordinary supported continuation; no artificial expiry deadline |
| 50000 | Complete; score50030, bonus0, NODOTCOUNT13996, expired=true; no second drain or unsupported consumer |

Two final fresh runs repeat this result. Accepted comparisons against independent
canonical-A and the fixed DMO0 witness remain through the same source-comparable
prefix; the122 saved linked collision rows cover35877..35998. No witness producer
is rerun. After35998 the full canonical-A progression is different: the suffix
checks completed calculations, actual timer and clock/counter recurrences,
score/bonus, single-drain count and source guard/cursor observations. The shared
physics regression suite remains the arithmetic reference; no false whole-state
equality with full-game A is claimed for this suffix.

The real installed presentation branch is covered structurally for complete
consumer order, all print data/positions, real dot memory, flash, unchanged
score/bonus/audio/tasks/waits and retained expired. Separate cases exercise719,
720,721, spring eligibility, changed/missing records and text, invalid BCD,
font admission, all-consumer budget starvation and pause, nested consumer
refusal, replacement, sticky failure and independent terminal QUIT.

Next unsupported consumer: **NOT REACHED through50000**. The bound is a test
limit, not a source boundary. Finding a later actual refusal remains unresolved;
this pass does not run toward repeated equality101534 or reopen termination
research. Structural completion of ShowInfoTS is not fresh reachability evidence.

## Fallible execution and validation

All mandatory ShowInfoTS consumers are checked before its first source writes.
Each current/next consumer is checked again before its own effects. Sticky
UNSUPPORTED_DEMO_TRANSITION, no canonical fallback, retained earlier legal
writes, no rollback and no late physics after refusal remain intact. The older
admission envelopes retain their accepted36725 refusal and predecessor tests.
Only test/tagged dmoimpl1 candidate files, source checks and this report change.
No production/frontend profile, .HI persistence, initials, bonus/match family,
scored expired restart, new gameplay family, scheduler or timing change is added.

Private logs: `/private/tmp/pf-dmo-impl2b12-*`.
Existing source-only private harness: `/private/tmp/pf-dmo-impl1-g75qk5z5`.
Final validation:

| Check | Result |
| --- | --- |
| Two repeated unchanged fresh TARGET runs, completed36725..50000 | PASS |
| Complete installed ShowInfoTS,55 records/1072 visits, all PRINTTASK data/positions and real dot memory | PASS, structural |
| Guards/provenance/budget/pause/replacement/nested refusal/sticky failure/independent QUIT | PASS, structural |
| Retained DMO-IMPL-1..2B11 plus2B12 TestDemo | PASS,84 top-level tests, no skips |
| New linked source/admission/no-HI-IO/no-fallback suite | PASS,3 tests,167 mutation subcases |
| Python demo source/admission/regressions | PASS,597 tests, no skips |
| Canonical-A partyland/physics/presentation/tablelogic/source | PASS,153 top-level tests;5 optional skips |
| A/B/C/D focused datalayout/frontend | PASS,21 top-level tests, no skips |
| Tagged vet | PASS |
| Windows amd64 executable and engine build | PASS |
| macOS c-shared engine build | PASS |
| Broad macOS compile-only | Known baseline FAIL: AudioDevice/hostWindow/openHost |
| PF6SessionKeepsGameplayOracle | Known baseline FAIL: score000002311040, ball2, ticks1200 |
| Three native lifecycle tests | Known baseline FAIL: gameshow53/speeddevils36/stones47, native content |
| Fixture-free TestDemo | PASS available32 tests;52 private-backed SKIP/NOT AVAILABLE |
| Public source checker and diff/scoped whitespace | PASS |
| Scoped payload scan | PASS,8 source/report files and2 built artifacts |
| OriginalTrajectories/Stones pf10 captures | NOT AVAILABLE; no fixture generated |

The payload scan covers whole private runtime files and nontrivial aligned4KiB
raw/hex/base64 samples across92 private files (1677 sample blocks). It is a
sampled check, not an all-substring proof. Original payload, strings/names/fonts,
witnesses and logs stay outside Git. Built production artifacts contain no
candidate unsupported-transition diagnostic.

Initial test iterations compared an entire calculation against frozen audio and
omitted SourceText's existing character decoding in the independent print
expectation. The final checks use MusicClock.Sync and SourceText respectively.
Neither correction changes TARGET, physics, presentation implementation or
source inputs. Final `*-final.log` results supersede those incomplete assertions.

Reproduction with the existing private source-only harness and legally supplied
originals/research Capstone:

```
PF_10MIN_DEMO_DATA=<pinned-demo> \
PF_DEMO_RESEARCH_WITNESS=<saved-drain-json> \
PF_DEMO_COLLISION_WITNESS=<saved-first-equality-json> \
 go test -tags dmoimpl1 ./internal/partyland ./internal/physics \
 -run '^TestDemo' -count=1 -v
PF_10MIN_DEMO_DATA=<pinned-demo> PF_RUNTIME_DATA=<canonical-A> \
 python3 -m unittest discover -s tools -p test_check_demo_2b12_consumers.py -v
```


No canonical-A oracle, .DS_Store, v0.1.3, commit/push/tag/release, new fixtures,
DMO0 research or DMO1 change. DMO-IMPL-2B13 is not started.
