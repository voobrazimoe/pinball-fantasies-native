# PF7 — native Speed Devils implementation and portability audit

Speed Devils is a native second table, selected with F2 in the normal frontend.
F1/Enter starts one player. The direct development command is:

```sh
./bin/pinballfantasies -data-dir . -pf7
```

The existing `-pf2`, `-pf4` and Party Land regression script retain their meanings.
No original machine instructions are executed, interpreted, translated or JITed.
The original files remain specification inputs. `tools/reference_pf7.py` extracts
content declarations and tracker control metadata into native data; Go handlers
implement the table state machines.

## Specification and recovered content

Primary table specification: `reference/original-dos-source/SDEV.ASM`.
Shared specification: `BALLCODE.ASM`, `FANTASIE.ASM`, `FANTASIE.MAC` and their
original graphics/task macros. `TABLE2.PRG` supplies linked original content;
`TABLE2.MOD` supplies music/sample data; `TABLE2.HI` supplies the 64-byte seed
records. INTRO remains the existing frontend. There is no manual/fan-site rule
substitution.

Pinned TABLE2.PRG SHA-256:
`6689dcef5fd051998bab990b5d243614c7dae2dcdcab9bffbe3c1936a76504b5`.
Pinned TABLE2.MOD SHA-256:
`728629c54311386781271308e181ac0435f0582e90870accff0a42270d467529`.

| Content | Linked location / recovered value |
|---|---|
| STAGE strips | `50730`, `583f0`, `60030`, `67b00` hexadecimal; four 320x144 PBM strips |
| Full playfield | 320x576, original indexed pixels / last strip palette |
| Initial visible field | rows 259..575, 320x317 |
| SETBALL | (300,530), velocity (10,0), low layer |
| New-ball hidden position | (285,530), then source SETBALL task |
| HID1 / upper foreground | `2d4f0` / `33370`, 23040 bytes each |
| MASK12 / MASK11 / MASK22 | `38d70` / `3e770` / `44170`, 23040 bytes each |
| MASK13 / MASK21 / MASK23 | `6d7c0` / `731c0` / `78bc0`; upper material planes end at 20400 bytes |
| Flipper deltas | `bad0`, 54288 bytes |
| Flipper descriptor / frame streams | `1f820`; `49b70`, `4e110`, `4c190` |
| Flipper planar stride | 42, four planes / 168 bytes total |
| Upper flipper | top 168, height 53 (Party Land values do not apply) |
| SIN / MAT_TABLE | `1d570` / `1b00b`; byte-identical lookup values in both tables |
| RAMPTABLE_hi | six effective gravities: (0,7),(0,12),(0,22),(-1,7),(0,17),(12,7) |
| Collision identities | four source bumper regions; six BURNIN contacts; original lower/upper layer areas |
| Lamps | 67 LON records, located in source order, not numerical lamp order |
| Tasks / flashes | 20 cooperative slots / 64 flash slots (Party Land: 50 / 15) |
| Game-over match | 18 banks, 13 syncs between steps (Party Land differs) |
| Attract music | DEMO_MUSIC selects S_NOHIGH, not S_MAIN |

The literal ball pixel palette writes are read as content, at the linked PUTBALL
locations shifted by `-810` hexadecimal from the Party Land locations. Original
HID foreground bits suppress covered sprite pixels. Source flipper streams modify
the shared collision/render masks; the upper layer uses its own foreground and
material masks. No artwork was redrawn.

The independent Python extractor produces `analysis/pf7-assets.json`: indexed,
RGBA, initial-frame, ball and mask hashes. It extracts 52 effects, 37 jingles,
490 matrix commands and 14 animation timing programs. Every animation's header
and frame durations are checked against linked DATA2. Inactive demo and MASM
COMMENT blocks are excluded.

## Portability audit

A: unchanged algorithmic machinery

- BALLCODE signed integer integration, three moves per sync, division/wrap rules,
  contact sampling and collision response.
- Physical flipper dynamics and original delta-stream decoding semantics.
- Shared indexed ball/foreground compositor and gameplay scrolling equations.
- Four-channel module renderer, sample mixing and existing tracker effects.
- SDL event handling and queued-audio backend; input abstraction.
- Native frontend pause/abort/game-end/initials lifecycle.
- Original 12-byte decimal score arithmetic and 64-byte high-score codec/storage.

B: shared concepts with proven Party Land assumptions removed

- `physics.Table` now owns start position, target regions, spring regions,
  ramp gravities and flipper plane stride. Decode offsets, layer rectangles and
  collision masks are table data. The integrator does not select tables.
- `assets.InitialTable` replaces the Party-named shared presentation container;
  a type alias preserves the PF2 API. PBM masking mode 2 is accepted because
  Speed Devils uses transparent-colour metadata, without a separate mask plane.
- Cooperative task allocation/run, call-site WAIT counters, scroll and animation
  clocks moved into `tablelogic`; slot counts remain table-local.
- The rational silent cue clock moved into `tablelogic`. Tables supply jingle
  priority/repeat/return metadata. Decimal arithmetic moved there too, with
  Party Land aliases preserving its public API and existing tests.
- Audio can trigger an explicit table-provided sample effect rather than looking
  up Party Land's sample numbers. No second mixer was introduced.
- Frontend factories and score selection use the selected table. Attract title,
  instruction text and music use Speed Devils content; F3/F4 have no factory.
- PF6.1 independently fixes physical window ownership for every presentation.

C: table-local implementation

Speed Devils contacts, lane history, BURNIN/PIT/GEAR progression, car parts,
miles/speed/overtaking, modes, capture/ejection, bonus, jackpots and extra-ball
branches remain in `speeddevils`. Party Land's arcade/duck/snack/dragon rules stay
in `partyland`. Neither table is a fork of the physics/tracker/platform engine.
Only demonstrated common primitives were extracted; there is no generic table DSL.

For a defined algorithm-family measure, **8 of 10 core engine families (80%)**
retained their existing implementation: integration, contact response, flipper
motion, indexed compositing, tracker decode, tracker mixing, input handling and
score persistence. The other two, cooperative timing and silent cue timing, were
extracted with existing semantics. This is an algorithm-family comparison, not a
claim that 80% of repository files are byte-identical. Table geometry, frontend
selection and window policy needed adapters; new table rules still account for
most of the new handwritten table code.

## Native table behavior

- B/U/R contacts award 7510 score / 550 bonus; N/I/N award 7520 / 570. Source
  cooldowns and group reset tasks light the A/R portions of GEAR and add jackpot.
  Flipper key-down edges rotate BUR, NIN and PIT left as CHECK_SHIFTKEYS does.
- GEAR raises gear and available positions; five gears wrap through the original
  hold-bonus branch. GEAR awards 500000 / 25000.
- PIT completes multiplier stock, with source lane/reset timers. The OffRoad lane
  collects stock; multipliers cap at eight and the full-stock branch awards Million.
- Loops use SSCORE1..12 (25000 increments, capped at 300000), source double-loop
  timing and first-mile behavior. Miles light OffRoad, extra ball and jumps at the
  original thresholds. Speedometer, position advancement and car-part ordering
  remain separate source state machines.
- Car parts award 500000 / 50000; Speedo awards 250000 / 25000; overtaking awards
  500000 / 50000. Goal lights lead into Turbo with original deferred tasks.
- Jump distinguishes upper approach history from the under-jump path. Lit Jump
  awards 10000000; jackpot starts at 5000000 and increments by 100000; SuperJack
  awards 50000000 / 1000000. Timed lamps and Turbo preemption/resumption are native.
- Pitstop captures at (256,41); source 20/150-sync primary waits, secondary wait,
  eject velocity (-2100,800) and reenable task are preserved. The source never sets
  SNACK_DISABLED true on capture; LASTCHECK provides repeated-entry suppression.
- OffRoad and Turbo have separate totals (100000 and 5000000 additions), original
  matrix intros, 25*71+1 countdown and deferred transition ordering.
- Ball loss holds at (280,560), fixes bottom scroll, inhibits gameplay and runs
  BALL_LOSTTS. Bonus multiplication precedes miles and mode totals; DO_FLORPA
  counts decimal digits with original delays. Hold Bonus, Shoot Again, ordinary
  new balls, out-of-balls and the match extra ball have separate paths.
- CHECKHIGHSCORE is chute/mode guarded. Ball-loss _BEATEN_MATRIX deliberately is
  not; the source reaches it only after nonzero bonus counting. _DOBEATEN sets lamp 55 once, preserving Speed Devils' actual shoot-again
  behavior rather than importing Party Land's extra-ball counter.
- Single-player WAITIFMULTI uses the source SISA=2 delay, independently tested.
- Original FLASHLIST drives attract lamps, with original playfield content and
  scrolling. ShowHighsTS title/instructions use Speed Devils identity and scores.

## Audio and determinism

TABLE2.MOD uses existing tracker commands, including E9 retrigger. No new tracker
opcode was needed. Table-local SFX select the original samples/periods. Silent
and audible cue runs agree at every sync for main music, SuperJack, Turbo,
OffRoad and Lost Ball; native PCM order follows the semantic cue position.
Game state never queries the host queued-audio clock.

## Frontend and persistence

F2 loads the native table attract; F1/Enter creates a fresh one-player session.
Pause freezes game syncs. Resume, quit question, chute abort, final-ball drain,
match, qualification, three initials, save, entry wait, table attract and return
to selector use the shared lifecycle. TABLE2.HI remains a 64-byte record file in
the native score directory; TABLE1 records are not used or rewritten by Speed
Devils. F3/F4 remain unavailable and non-destructive.

## Fixtures and validation

The focused Speed Devils fixtures assert source-derived awards and transitions,
not a score invented from a stable native run:

| Fixture | Evidence/checkpoint |
|---|---|
| Assets / initial | independent original index/RGBA/mask/ball hashes, (300,530), viewport259 |
| Motion | first tick X=300*1024+30, Y=530*1024+24, VY=23, GY=7; spring -5312 |
| Collision | byte-identical SIN/materials; original steel response -442; all four bumper identities found in original masks |
| Flippers/layers | left first frame speed -68; upper top/height/stride; original entry/exit areas |
| BURNIN/GEAR | 45090/3360 after both groups, then 545090/28360 on GEAR; reset/hold-bonus tasks |
| PIT | 40080/4040 after PIT; 60120/6130 after stock collection; max-stock Million |
| Loops/modes | source first mile, double-loop Million, positions, thresholds, deferred Turbo |
| Jump/capture | 15600000 from lit jump/jackpot/part; timed lamps; source capture/eject |
| Bonus/new ball | 5302000 = 1000*2 + 2*100000 + 100000 + 5000000; ball2, SETBALL task |
| Top score/match | chute guard, ball-loss bypass, one shoot-again, 18 match banks, game-over |
| Frontend | real session through pause/abort/final drain/ABC initials; TABLE2.HI150000000, no TABLE1 write |
| Party Land | unchanged 1200-tick script: score2300000, ball2; all existing physics/audio/timing fixtures |

Several smaller fixtures are stronger than a single unconstrained Speed Devils
1200-tick score: they isolate source awards, timer boundaries and branch ordering,
and actually exercise modes, bonus and persistence. The direct 1200-tick launch
was additionally exercised (observed score11030, ball2), but that observation is
not used as an independent source oracle.

Commands:

```sh
./tools/go.sh test -p=1 -count=1 ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
DISPLAY=:0 python3 tools/smoke_pf6.py
```

Build succeeds. All package/regression tests pass against original inventoried
inputs in `/tmp/pinballf-pf7-validation`. The requested command in the working
tree reports **one failure**, the unchanged inventory test for TABLE1.HI: the
installation record changed during this session to MES/7310090. Its current
SHA is `0071b43a343ec72ae394d4671e126247818e932d6d6874557b6a5702aa2a40b3`.
That record was preserved, and the inventory test was not weakened. In the
isolated copy only, the source PLAND default fourth record (5000000 / J L)
restores the exact pinned hash
`e7799bb471a403066b549c00510764a0c810ac8d8ba7b298ad523684b36d08a6`.

Real SDL/X11 smoke passes with dummy audio: one window, manual XResizeWindow,
unchanged dimensions/window identity through selector text, unavailable F3/F4,
both tables' attract/play/pause/resume/abort and selector return. Queue resets
and empty-queue observations are zero. High-score/game-over interaction is
covered by real native model/session tests; the bounded X11 smoke does not
simulate an entire game or initials entry. See PF6.1 report for dimensions.

## Files changed / added

PF6.1: `internal/platform/{sizing.go,sizing_test.go,frontend.go,window.go,physics.go}`,
`tools/smoke_pf6.py`, `analysis/pf61-window.md`.

PF7 new: `internal/speeddevils/{game.go,features.go,regions.go,flow.go,matrix.go,
audio.go,render.go,content.go,game_test.go,audio_test.go}`;
`internal/tablelogic/{tasks.go,music.go,decimal.go}`;
`internal/assets/{speeddevils.go,speeddevils_test.go}`;
`internal/physics/{speeddevils.go,speeddevils_test.go}`;
`internal/frontend/speeddevils_test.go`; `cmd/pinballfantasies/speeddevils.go`;
`tools/reference_pf7.py`; `analysis/{pf7-assets.json,pf7-speeddevils.md}`.

PF7 adapted: `internal/assets/{partyland.go,initial.go}`;
`internal/physics/{data.go,ball.go,render.go}`;
`internal/partyland/{game.go,features.go,holes.go,timing.go}`;
`internal/audio/{module.go,player.go}`;
`internal/frontend/{model.go,runtime.go,render.go,scores.go}`;
`cmd/pinballfantasies/main.go`; `tools/reference_extract.py`; `README.md`.
The rebuilt `bin/pinballfantasies` is available. Historical ASM/assets were not edited.

## Deliberate differences / unresolved fidelity

The gameplay score panel is the existing native PF4/PF6 presentation model.
Original matrix commands, priorities, waits and animation frame durations drive
native semantics, but the original animated dot-matrix bitmap pictures are not
rendered. The native panel instead reports score/bonus/gear/miles/position/modes.
Consequently this is playable native session coverage, not a claim of completely
pixel-identical original matrix presentation. Original attract instruction text
also mentions mouse, nudge and music-toggle controls that are outside the existing
native control implementation. The table uses the established deterministic
FANTASIE clock/SDR callback phase convention, whose original device callback phase
remains the inherited PF4.5/PF5 uncertainty. No original CPU oracle was run.

## Remaining tables: concrete engineering assessment

**Moderate overall porting difficulty**, with high table-rule workload. The
completed second table required no new physics equation, mixer or scheduler
ordering algorithm. Formats are shared (indexed PBM strips, bit masks, planar
flipper deltas, module data, decimal/HI records), but offsets, mask lengths,
flipper strides, palette order and per-table capacities demonstrably differ.
A remaining port therefore still needs content archaeology and fixtures; copying
Party Land constants is unsafe.

The cost that remains is understanding and implementing table-local rules and
matrix-driven timing, then testing capture/layer/mode interactions. Speed Devils'
BURNIN/GEAR/miles/position machinery needed substantial new native state even
though its physics and audio were reused. The extracted primitives now eliminate
that repeated infrastructure work for a third table. Rendering surprises were
masking mode2, the short upper material planes and 42-byte flipper stride; audio
only needed different sample mappings.

Planning estimate relative to this completed Speed Devils port: Gameshow about
0.7–1.0 times the effort, Stones 'N Bones about 1.0–1.5 times, excluding the shared
initial extraction/window refactoring already paid here. These are engineering
ranges, not measured results for unported tables. Confidence is strongest in
shared engine reuse and weakest in undiscovered table-specific state machines.
Original dot-matrix bitmap presentation work, if required for full visual fidelity,
is an additional shared rendering milestone rather than already-completed PF7 work.
