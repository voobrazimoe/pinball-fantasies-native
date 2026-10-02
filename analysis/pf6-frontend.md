# PF6 — native frontend and single-player lifecycle

Implemented around PF1–PF5, with original content decoded as data. No DOS,
x86 execution, interpreter, JIT, emulator or translated instruction stream is
used. Party Land is the sole playable table. Normal startup now enters PF6;
`-pf4` retains the direct development path. `-pf2` explicitly preserves the
former default initial-frame viewer/export.

## Source references and recovered flow

| Source | Behavioral evidence | Native boundary |
|---|---|---|
| START.ASM `ONCE_AGAIN`, `banor`, `bannumber`, `first_time` | First invocation runs INTRO; selected table program runs; table exit re-enters INTRO without repeating startup | One native process; session factory and selector return |
| INTRO.ASM `enough_mem`, `firsttime_init`, `wait_music`, `checkquit`, `trasselsudd` | INTRO.MOD first time, MOD2.MOD on subsequent frontend invocations; startup logos, Presents, title; Space skips | Startup segments with integer music counter and sync-based palette fades |
| INTRO.ASM `unpkpics2`, `tabletabell`, `vartabell`, `fontlist`, `PUTCHAR` | Original PBM/ILBM pictures, planar fonts, content placements | Native IFF ByteRun1, planar-to-index decode, original glyphs/palettes |
| INTRO.ASM `choosern_init`, `showpics`, `keyread`, `space`, `WAITEND`, `entera`, `savebana` | F1–F4 load tables; Enter alternates previews; Space/540-sync timer alternates preview bank and shows intervening text; 420-sync text hold | Selector and selector-text states, page/text counters |
| INTRO.ASM `RASTATAB`, `rastra`, `rastraclear`, `putraster`, `putraster2`, `fade3`, `fade3b` | 95-line lattice reveal/clear, opposed second-preview direction, DAC 1/12/14 text transitions | Deterministic line groups and integer mono-to-original-color presentation |
| FANTASIE.ASM `checkstartkeys`, `VBLANK_INT_DEMO`, `GO_GAME_MODE` | Table load starts attract, F1 or Enter starts one player | Table-attract state; fresh native Party Land session |
| FANTASIE.ASM `checkpause`, `FEELP`, `judas`, `kaka_calamares`, `lat_den_sta`, `QUIT` | P stops interrupts/audio; next make code resumes; Escape from pause asks REALLY QUIT; Y exits table, another key resumes | Paused and quit-question states; selector return on Y |
| FANTASIE.ASM command-key dispatch, `checkgamequit`, `quit_fragenTS`, `WAITRUT_YN` | Escape in chute ends game and returns to table attract; attract Escape asks to exit; outside chute Escape is cleared | Narrow session `InChute` capability and distinct abort/quit outcomes |
| PLAND.ASM `NO_MORE_BALLS`, `out_of_ballsTS`, `_CHECK_XXBALLS`, `after_xxballTS`, `_check_high`, `SPINTSEL_IN_HIGH`, `_2_DEMO_MODE`, `TO_DEMO_FROM_GAME` | Drain/bonus/ball/match precede score check; after entry/nonqualification return to table attract, not immediately to selector | Existing PF4 reports final result; PF6 owns post-game entry and attract return |
| FANTASIE.ASM `CHECKHIGHSCORE`, `_beaten_matrix`; PLAND.ASM `BeatenTS`, `Beaten_bh_TS`, `_DOBEATEN` | Strictly beating top score once awards an extra ball, including ball-loss branch | Small opt-in Party Land hook plus existing matrix/task/jingle scheduler |
| PLAND.ASM `read_keyboardet`, `wait_a_little`, `hajjskar`; FANTASIE.ASM `ALFA_KEYS` | Three characters, no modern edit/Enter widget; letter make codes and Space mapped to `*`; countdown starts at 60, stars at 30, next-player scan at 2 | Three-character entry, 58-sync completion hold, original additional punctuation glyphs |
| PLAND.ASM `hi_score_list`; INTRO.ASM `HI_SCORE_LIST` through `HI_SCORE_LIST4`, `READ_HIGHS`, `MOVSCORE`; FANTASIE.ASM `INIT_HIGHS`, `SAVE_HIGHS` | Four table-specific 16-byte records; 12 unpacked decimal digits, 3 initials, NUL; strict ordered comparison and downward insertion; DOS `.HI` persistence | Integer score model and original 64-byte format |

The explicit modes are startup, selector, selector text, table attract, playing,
paused, quit question, game over, initials, entry wait and quit. The game-over
state is the five-sync clear before the frontend score decision. Reasons are
active, completed, aborted and program quit. Selection inputs do not leak into
the newly created game. Unavailable choices retain their identity and remain
in the selector; they never create a Party Land session.

INTRO uses a deterministic 60 Hz presentation cadence (inferred from its VGA
presentation and `60*menutime`/`60*texttime` constants); its `wait_music` counter
is represented in 50 Hz integer units. Table, pause and entry use the existing
71 Hz native cadence. Host elapsed time schedules whole ticks only. No PF4.5
scheduler rate, task ordering or audio callback convention was changed.

## Session, pause, scores and persistence

`frontend.Session` exposes sync, release, frame, result, PCM and cue methods.
The concrete implementation remains `partyland.Game`. `SetHighScore` supplies
an immutable top-score value; the gameplay subsystem implements its own beat
and extra-ball logic. The direct PF4 path does not install this optional value,
preserving its established oracle. The PF6-started session also produces that
same 1200-sync oracle with the normal top-score content.

Paused/quit-question ticks call neither gameplay sync nor tracker rendering.
Rendering and the attract camera do not mutate the session. SDL clears queued
samples at suspension and prerolls again on resume. Attract/entry may advance
only presentation audio, never ball/rules/tasks/gameplay Tick. DOS STOP/PLAY
resume phase has not been hardware-validated; native tracker state is retained.

Qualification scans highest to lowest and inserts before the first strictly
smaller score. A tie with the fourth entry fails; a tie with an earlier entry
may still beat a lower entry. Arithmetic uses twelve decimal bytes/uint64,
never floating point. Names consist of exactly three original ALFA_KEYS
characters; Backspace, Enter and digits have no editing meaning. Physical SDL scancode make
edges map to DOS codes independently of the desktop character layout; host repeated keydown events are ignored.

Native storage defaults to `$XDG_CONFIG_HOME/pinballfantasies`, or
`~/.config/pinballfantasies`. `-high-score-dir` overrides it. Filenames remain
`TABLE1.HI` through `TABLE4.HI`, and the records retain the original 64-byte
format. The installation's `.HI` files seed missing native files; source defaults
are used when neither exists. Only Party Land records are saved. A complete
entry writes a temporary file, syncs it, then renames it atomically. Installation
assets are never modified. Malformed score files produce an error instead of
silently discarding scores. Tests inject memory stores or temporary directories.

## Presentation and deliberate limits

All startup, selector, table preview, high-score heading and selector-font images
come from the pinned INTRO.PRG. Both original frontend modules use the existing
PF5 player. No tracker feature or mixer rewrite was required. The supplied
INTRO.MOD omits unused silent samples 21–31 and appends a two-byte checksum;
the decoder normalizes just those known silent slots without reading the
checksum as playable sample content.

Table-attract/entry/pause output uses the existing native PF4 score-panel and
playfield composition, with original FONT13/FONT5 content. The six additional
FONT13 punctuation glyphs are read from original data, including the initials
parentheses and star. The attract sequence follows ShowHighsTS/showithi text,
waits and score ordering with a native triangular camera scroll. Its VGA matrix
wipe/flash presentation is not pixel-exact. The selector retains the original
static sidebar artwork: the original animated instruction sidebar uses a BIOS
ROM font that is not part of the supplied assets, so that ROM-font animation is
omitted. No host/system font or replacement launcher artwork is substituted.
Exact original raster callback phase and DOS audio restart phase remain
unverified. The existing PF0–PF5 accepted differences remain unchanged.

F5's original options label can appear in selector instructions, but configuration
UI is outside scope. Multiplayer F2–F8 within table attract is not implemented;
F1/Enter starts one player. F2–F4 in the global selector preserve their original
artwork/choice identities and have a non-destructive no-op load. There are no
fake playable tables, new menu dialogs, remapping or controller systems.

## Files and validation

Added:

- `internal/frontend/{model,scores,render,runtime}.go`, `model_test.go`.
- `internal/assets/frontend.go`, `frontend_test.go`.
- `internal/audio/frontend_test.go`.
- `internal/partyland/session.go`, `session_test.go`.
- `internal/platform/frontend.go`.
- `tools/reference_pf6.py`, `tools/smoke_pf6.py`.
- `analysis/pf6-art-fixtures.json`, this report.

Changed:

- `cmd/pinballfantasies/main.go`: default frontend, native score path, explicit PF2 bypass.
- `internal/audio/module.go`: pinned frontend modules and known omitted silent slots.
- `internal/platform/audio.go`: queue suspension/preroll boundary.
- `internal/partyland/{game,timing,timing_data}.go`: optional top-score content,
  source extra-ball matrix branch, original last-jingle handling and lifecycle methods.
- `tools/reference_pf45.py`, `analysis/pf45-timing-fixtures.json`: retain the two
  newly reachable original beaten-score matrix programs/jingles; existing timing
  traces and tests are retained.
- `README.md`: normal launch, original lifecycle controls and persistence.

There was no usable Git metadata in this supplied working directory (`git status`
reported “not a git repository”; `.git` is an empty read-only placeholder).
Inspection preceded edits; no broad cleanup or unrelated subsystem rewrite was
performed. Original assets/source were read only.

Validation on Ubuntu, 2026-09-30:

- `python3 tools/reference_pf6.py`: 17 original IFFs decoded independently;
  dimensions, index hashes and VGA-quantized RGBA hashes match Go.
- `./tools/go.sh test -p=1 -count=1 ./...`: PASS, all packages. Focused frontend
  coverage includes automatic startup, selection, handoff, real native three-ball
  completion, both score paths, source key table, insertion/ties, entry/return,
  persistence/seeding, unavailable tables, close in every state, full native
  pause state/PCM freeze, resume, abort and a fresh restart.
- `./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies`:
  PASS.
- `-pf4 -pf4-script -ticks 1200 -png /tmp/pf6-direct-regression.png`:
  score **2,300,000**, ball **2**. A second test checks the PF6-created session
  against the same oracle.
- `-pf2 -png /tmp/pf6-pf2-baseline.png`: original initial-frame export works.
- Actual X11 SDL key smoke (`python3 tools/smoke_pf6.py`): PASS. Exercised
  Space skip, all four selector choices, Party Land start, launch, pause/resume,
  Escape/N cancellation, Escape/Y table abort, frontend return and exit. It
  verified original-mode transitions and no score writes on abort. Audio driver
  was dummy for deterministic UI isolation; zero queue resets/empty observations.
- Separate two-second X11/PulseAudio startup smoke: PASS, real audio opened,
  zero queue resets/empty observations. Offscreen/dummy smoke also passed. A
  sandboxed default PulseAudio probe could not access the host audio service;
  the desktop-access run succeeded.
- Startup, selector and high-score original-content frames were visually inspected.
