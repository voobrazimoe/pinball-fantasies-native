# DMO0 canonical-A launch-jitter reference inheritance

Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
Research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

This pass establishes launch-jitter inheritance under the existing canonical-A
native semantic reference. It does not establish a physical DOS startup phase,
change production, or close DMO0. The new owner-local artifact is
`/private/tmp/pf-dmo0-canonical-a-jitter.json`; previous artifacts stay unchanged.

## A's accepted convention and historical boundary

At production base, `internal/frontend/runtime.go` creates Party Land through
`partyland.New`. `Model.startPlayers` invokes the factory again at game start.
The constructor allocates a fresh Game and fresh physics, initializes ball 1,
one player and reset rule state. Party Land has no `CarryLoadedTableState` hook.
The frontend may carry matrix display memory; it does not carry the selected
trajectory fields or the gameplay clock.

The uint16 `clock` uses Go's zero initialization. PresentationAudioSync only
advances audio; AttractFrame renders without updating gameplay. Sync adds 1030
before audio, phase branches and physics. AfterTargets runs tasks before
springControl; springControl passes `uint8(g.clock)` to Release. Thus a fresh
constructor has clock 0 and its first spring consumer after Sync sees low8=6.
It is not a claim that a DOS counter was reset by NEW_BALL, or that every later
ball resets the native clock.

Historical evidence is retained in Git at `8879eb0^`:

- `analysis/pf8-dos-parity.md:197–204` distinguishes source velocity/spin
  arithmetic and task ordering from the existing deterministic native clock
  convention. Exact DOS SLUMP_COUNTERN callback phase remains unresolved and
  is not represented as authentic. The charged/released checkpoint at line
  281 retains the same hardware-phase caveat.
- `analysis/pf6-frontend.md:18,42–59` describes fresh native Party Land sessions,
  the retained PF4 oracle and attract presentation that never advances gameplay.

Current tests constrain arithmetic and the selected native contract:
TestOriginalRelease supplies jitter 255 explicitly; the direct Down/release test
uses no Sync and therefore clock 0; TestPF6SessionKeepsGameplayOracle checks the
fresh frontend session against the historical deterministic rule oracle. Its
current known baseline assertion fails (2311040 versus 2300000), as already
documented in runtime-layout-c-validation.md:103; it is not labelled PASS.
These tests do not
require a historical attract counter phase. TestOriginalTrajectories remains
NOT AVAILABLE and is neither executed nor regenerated in this pass.

## Minimum linked A/demo comparison

The auditor checks pinned private A/demo identities, pinned historical sources,
and full bounded instruction/operand correspondence with reviewed relocations.
Its output contains metadata, not executable slices or original payload.

| Slice | Classification | Reference implication |
|---|---|---|
| Primary ADD 1030, A 0x4500 / demo 0x4556 | RELOCATED-IDENTICAL | Same word arithmetic |
| MAIN INC, A 0x39e9 / demo 0x39c5 | RELOCATED-IDENTICAL | Same historical foreground producer |
| DEMOMODE test after primary ADD | RELOCATED-IDENTICAL | Demo does not move the increment |
| GO_GAME_MODE | RELOCATED-IDENTICAL | Same gameplay mode boundary |
| Table new-game, RESET_VARS2, RESET_VARS, table new-ball | RELOCATED-IDENTICAL | Same selected reset semantics |
| Generic new-ball reset prefix; task/wait reset | RELOCATED-IDENTICAL | Reset precedes fresh tasks |
| Generic new-game player/current-ball stores | VALUE-EQUIVALENT | Both set player 1 and ball 1 |
| Generic new-game UI/checksum differences | DEMO-SPECIFIC-BUT-REFERENCE-IRRELEVANT | Demo omits UI/checksum failure handling; valid A and demo end with MAKE_BAD false |
| NEW_BALL; NEW_BALL_PART_TWO; SNART wait prefix | RELOCATED-IDENTICAL | Same held-ball construction and delayed SETBALL path |
| SETBALL | RELOCATED-IDENTICAL | Same coordinates, velocity, high flag and HOLDSTILL stores |
| SPRINGTASK including SPRINGUP body | RELOCATED-IDENTICAL | Same charge, release, validity, sound/reset and task-pointer operations |
| High-resolution spring velocity | RELOCATED-IDENTICAL | VX=0, VY=-166*charge-low8(counter) |
| Rotation | RELOCATED-IDENTICAL | counter & 15 |
| SLACK_LIGHTS bounded loop | RELOCATED-IDENTICAL | Same LIGHTSTATUS clearing and bounds |
| Automatic F1 acceptance | DEMO-SPECIFIC-BUT-REFERENCE-IRRELEVANT | Changes pre-session history, not the gameplay jitter algorithm |

No REFERENCE-BLOCKING-DIFFERENCE was found in these launch-jitter slices.
Opaque callees are paired edges, not newly closed runtime domains. Low-resolution
coefficient branches are also paired, but this decision uses the existing A
high-resolution native reference.

## Selected pre-game state audit

This audit uses the already selected fresh one-player/no-input launch construction.
It does not reopen arbitrary input prefixes, task domains or global alias closure.

Score is cleared by RESET_VARS2. SkillTunnel and SkillCyclone are cleared by its
two six-word REP stores; HappyTotal and MegaTotal by RESET_VARS. XXBALLE and
INH_EFF are cleared. SLACK_LIGHTS runs with UPPSTARTAD=true (linked startup store
at demo 0x33dd), clearing LIGHTSTATUS; table new-game initialization supplies the
fresh progression records before new-ball restoration. The selected BYGEL light
is unlit. Task slots and waits reset before fresh SNART/ALLOW and SETBALL work.
Spring charge is loaded zero and attract bypasses SPRINGTASK; the chosen prefix
has no charge input. NEW_BALL holds the ball at 282,530 with zero speed; SETBALL
sets 297,530, VX=10, VY=0, high=false, HOLDSTILL=false. Player and ball are both 1.

The only surviving historical timing residue among the selected fields is
SLUMP_COUNTERN, including held rotation derived from it. Held rotation is not
an independent C0/M/P input: SPRINGUP overwrites it with low4 of the release
counter before the launched flight. Under the chosen fresh native construction
these are initialized by A's clock convention. No additional surviving
pre-game trajectory residue was found in this bounded audit.

The historical artifact's local reset statements remain local; no physical DOS
joined-state or arbitrary opaque-writer theorem is promoted. The new entry
verdict is a native-reference transfer at the chosen fresh construction.

## Reference admissibility and boundary

A already replaces physical foreground/IRQ timing by a deterministic gameplay
clock. Demo shares the relevant increment and spring consumers and supplies no
algorithm requiring a different convention. Therefore it can inherit the same
clock input sequence: u16(1030*n), low8 for velocity and low4 for rotation.
The source arithmetic is preserved at each chosen logical gameplay visit.
Neither this sequence nor native's first low byte 6 is claimed to be the exact
physical DOS sequence. No theorem C0+M+6P ≡ 6 is needed for this reference.

For launch-jitter initialization ONLY, historical C0, pre-game MAIN count M,
pre-game primary count P, automatic-F1 acceptance phase and CPU/IRQ relative
timing before fresh construction are BELOW-NATIVE-REFERENCE-BOUNDARY. They keep
their historical DOS meaning. ATTRACT_PHASE_ALIGNMENT remains historically
NOT_PROVED, but no longer blocks this native equivalence boundary.

FRESH_BYGEL_ENTRY_TRANSFER = PROVED for the bounded selected fresh-session
reference. There is no remaining entry-transfer reason in this scope. The next
pass may run deterministic BYGEL trajectory search. No such search ran here;
flight consumer closure, FRESH_BYGEL_DRAIN_PROVENANCE and drain membership are
not upgraded. DMO0 NOT CLOSED. DMO1 NOT STARTED.

## Verification

Run the new auditor/tests with PF_10MIN_DEMO_DATA, PF_RUNTIME_DATA and
PF_DMO0_HISTORICAL_SOURCE. New-pass tests cover both increments, velocity/masks,
fresh native clock, ordering, reset preservation, changed jitter arithmetic,
a surviving prior score/selected field, and the explicit unproved historical
phase. Verification:

- New pass: 25 tests PASS. No-private run: 12 tests PASS, private class SKIP.
- Full private demo research suite: 390 tests PASS (48.676 seconds).
- Available A spring/physics/reference checks: 28 test/subtest results PASS,
  one reproduced baseline FAIL, zero remaining skips. The baseline failure is
  TestPF6SessionKeepsGameplayOracle (score 2311040, expected 2300000). Initial
  CFG-related skips were rerun against the available real private CFG; the
  test and expected value were not altered. Direct TestDeterministicNativeScript
  PASS. This baseline failure does not supply a historical counter phase.
- A/B/C/D focused regression: 97 test/subtest results PASS, zero skips/failures.
- The external A harness matches all 284 tracked Go/module files in this
  checkout; all of those present at production base also match that base.
- TestOriginalTrajectories: NOT AVAILABLE, not executed or manufactured.
- Raw/hex/base64 sampled payload-copy check: PASS against 249 private files,
  11 decoded SDR modules and 48,898 distinct nontrivial aligned 4 KiB samples.
  This is not an all-substring theorem.
- git diff --check: PASS. Previous tracked diff is preserved byte-for-byte;
  production/runtime diff is empty; .DS_Store metadata is unchanged.

Only the new auditor, new test file and this document are added to the working
tree. No staging, commit, push, tag, release or v0.1.3 change.

CANONICAL_A_JITTER_INHERITANCE = PROVED
HISTORICAL_ATTRACT_PHASE = NOT_PROVED_BUT_BELOW_NATIVE_REFERENCE_BOUNDARY
FRESH_BYGEL_ENTRY_TRANSFER = PROVED
