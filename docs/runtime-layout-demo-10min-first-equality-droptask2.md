# DMO0: DROPTASK2 at first equality

`FIRST_EQUALITY_DROPTASK2_REACHABLE` under the accepted fresh canonical-A native
reference. `EXPIRY_INTERLEAVING = NOT_PROVED`. `DMO0 NOT CLOSED. DMO1 NOT STARTED.`
Production base is `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; committed research
HEAD is `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

DROPTASK2 fires after expiry installation at calculation 35998, preserves the
expiry program/cursor **and HOLDSTILL**, and sets the ball high at (15,47) with
VX=0, VY=52. It does **not** release the ball after expiry. The same calculation
visits expiry clear 5→4, then executes a held late physics pass. No further
input leads to QUIT at 37039, without movement, drain, capture, or replacement.

## Linked chain and arithmetic

The selected producer is the ordinary, non-special tunnel GROPD, bound by the
linked lower area record at file 0x1ab85: rectangle (47,122)..(67,146), callback
0x1e2d. Canonical A's corresponding record/callback is 0x1aaf9/0x1e29. The actual
SKILLTUNNEL effect installs program 0x1b904 and inhibits the later TSCORE1 matrix.
All linked GROPD branches set TUNNELWAITER DS:0xac to 130. No START_DROP_TIMED or
START_DROP_WHEN_READY path is consumed by this witness.

| Task/procedure | Demo file | WAITLIST DS | Limit | Actual slot and first visit |
| --- | --- | --- | --- | --- |
| GROPD | 0x1e2d | none | none | capture E=35810, before task scan |
| RULLGARDIN | 0x151b | none | none | slot0, first E; inserts slave into later slot3 |
| RULLGARDINSLAV | 0x152c | none | none | slot3, first E; dies in E because scroll position becomes negative |
| WAIT_FOR_TUNNEL_EFFECT | 0x1f13 | 0x36f1 | DS:0xac=130 | slot1, first E; fires E+130=35940 |
| LOCK_BALL_IN_TUNNEL | 0x1f3a | 0x36f3 | 2 | slot2, first E; fires E+2=35812 |
| START_DROP | 0x142a | none | none | called by the surviving slot1 task at C=35940 |
| second RULLGARDIN | 0x151b | none | none | inserted into earlier slot0 at C; first C+1 |
| second RULLGARDINSLAV | 0x152c | none | none | inserted into later slot1 at C+1; first/dies C+1 |
| DROPTASK1 | 0x1479 | 0x36d7 | 30 | inserted slot2 at C; first C, fires C+30=35970 |
| DROPTASK2 | 0x14a3 | 0x36d9 | 27 | inserted slot0 at 35970; first 35971; fires 35998 |

Parent slots remain occupied until SUICIDE. DOADDTASK takes the first free slot
and does not zero any wait word. DO_TASKS scans slots in ascending order. Each
WAITSYNCS compares before incrementing and resets its shared word on equality.
Thus WAIT130 fires on visit131, DROP1 wait30 on visit31, DROP2 wait27 on visit28:

`C=E+130; F1=C+30; first_DROP2=F1+1; F=F1+1+27=E+188`.

Solving F=35998 gives E=35810, C=35940, F1=35970. Treating DROP2's first visit as
same-scan incorrectly gives 35997; delaying DROP1's first visit incorrectly gives
35999. The JSON includes actual per-visit ages and persistent task identities.

The correspondence file verifies entire bounded producer/task/helper blocks,
with reviewed operand relocations. Linked operands independently bind wait
words, limits, insertion targets, coordinates, velocity and high flag.
Historical PLAND.ASM and FANTASIE.MAC corroborate the linked result.

## Fresh input witnesses

The first capture family used release124, charge27, Left iff
`(n-124+63)%119 < 59`, Right iff `(n-124+117)%208 < 47`, from release through331;
Down on97..123. All remaining controls are false. Capture331 → START_DROP461 →
DROP1 fire491 → DROP2 fire519. It releases normally before expiry. This fixed
script is also reproduced twice in the final harness.

The target script is:

- Down on calculations **35516..35533**; release at **35534**.
- On **35534..35810**, Left iff `(n-35534+13)%95 < 57` and Right iff
  `(n-35534+22)%84 < 34`.
- All other controls are false; **all controls false from35811 through QUIT**.

Both are fresh `New(DecodePartyLand(A), A)` + `Configure(Legacy)` sessions. The
script has an explicit cutoff, without capture feedback, timer seeding, state
edits, teleportation, or task insertion. The target and short family each
replay identically twice. The target has one initial reset, no prior capture or
drain, and no tasks live at capture. Its preceding switches are BYGEL12,
BYGEL28, BYGEL12, CLOSE1, BYGEL9, BYGEL11, then GROPD.

A finite 513-script long-release search failed. A separate deterministic short
search found a candidate after2574 trials whose drop fire had the desired low
seven-bit clock phase; one delayed candidate then replayed fresh to35998. The
residue only proposed the candidate. No linear-shift or impossibility premise
is used in the acceptance test.

## HOLDSTILL binding and exact equality

The native capture abstraction uses `Ball.Hold=true` after LOCK_BALL_IN_TUNNEL.
This is **not** a linked HOLDSTILL store. GROPD's position macro, LOCK's macro,
START_DROP and DROPTASK2 do not write DS:0x3026. The source HOLDSTILL remains
false after the initial SETBALL until expiry sets it true. Native capture freeze
and source HOLDSTILL are recorded separately. This remains a native-reference
proof, not a DOS gameplay execution claim.

DROPTASK2 0x14b2..0x151a writes SCREENFORCE DS:0x347f=-1, derives BP from
SLUMP_COUNTERN DS:0x34ec &127, writes pixel position (15,47), fixed position
(15360,48128), VX=0, VY=BP, BALLHIGH DS:0x3416=true, plays SNEWBALL, ends flashes
3/56, and suicides. The linked body and bounded flash/suicide helpers have no
HOLDSTILL, expired, BALL_DOWN, LOOSING, chute, effect or matrix write.

The isolated demo overlay preserves an installed expiry hold while removing the
native capture freeze for an ordinary drop. Transplanting native `Hold=false`
unconditionally would fabricate the claimed post-expiry release. Its explicit
mutation test demonstrates the error; production files are unchanged.

At35997 the slot0 task has shared DS:0x36d9=27, no duplicate, no reset, no exit,
no competing wait instance, and expired=false. The captured reference ball is
low at (15,47), velocity0/0. SKILLTUNNEL ends at35920; the ordinary idle panel
runs35921..35923. Before equality DOTRUT is NODOT/cursor0, the last panel is
SHOWPLAYERSTS at0x1b88e, and no effect remains active.

Calculation35998 executes two early captured-ball physics passes, UPDATE_COUNTERS,
then DO_ELECTRONICS changes35997→35998. It sets expired=true and HOLDSTILL=true,
plays S_GAMEOVER2 with priority255, installs program0x1ba17/cursor0x1ba19 with
clear remaining5. Ordinary electronics continues and DO_TASKS visits slot0.
The wait compares27, resets to0, and fires. BP is52 under the accepted fresh
clock construction. Program/cursor/clear remain unchanged; HOLDSTILL stays true.
BALL_DOWN=false, LOOSING=false, chute=false. Suicide leaves no live tasks.

The first matrix visit consumes clear5→4 without changing its cursor. The late
physics pass is admitted but skips collision response, movement and gravity
because HOLDSTILL is true. There are no late collision-ring, material, sine,
ramp or region reads; only flipper2/frame0 copying occurs. Material6/contact1
remaining in the snapshot is old metadata, not a consumed late collision.
Early ramp indices2201/2202 and five negative low-level tests precede the task.
Area/target/drain dispatch has already returned; no later same-calculation
callback replaces or restarts expiry.

Without input, the high ball stays at(15,47), VX=0/VY=52. No capture, drain,
effect, task, or reset intervenes. Expiry remains current for1042 visits and
executes QUIT at37039. This is the concrete drop continuation.

## Bounded inventory review

The old nine entries are reclassified in the JSON. NEW_BALL_TASK, PARTY_ON_TASK1
and SETBALL retain their previously proved classes; their nonfiring ages add no
independent equality body. The early scored-drain matrix is superseded by
expiry. DROPTASK2 is now a proved firing/preservation class.

The old “other held-ball continuation” bucket still has one selected independent
candidate: **DURINGFLASH**, file0x2750, wait DS:0x3701, limit27. Its linked store at
0x2778 explicitly clears HOLDSTILL, then emits (257,310), VX=-575, VY=1575,
high=false. Its play-field physics differs from SETBALL's chute release and this
hold-preserving drop. No fresh first-equality dragon witness or exclusion is
asserted. This pass only binds those operands to justify the remaining class;
it does not expand its producer chain or start a second trajectory search.

Therefore `EXPIRY_INTERLEAVING = NOT_PROVED`, with exactly one smallest selected
remaining dependency: fresh input-only DURINGFLASH surviving at age27 on first
equality, followed by its actual release suffix. This conclusion follows current
linked evidence, rather than copying the former `classes_complete=false` field.

## Artifacts and validation

Owner-local metadata: `/private/tmp/pf-dmo0-first-equality-droptask2.json`.
Research tool: `tools/audit_10min_demo_droptask2.py`; correspondence:
`tools/dmo0_droptask2_correspondence.json`; isolated Go template:
`tools/dmo0_droptask2_test.go.txt`; tests:
`tools/test_audit_10min_demo_droptask2.py`.

With the same private fixture environment and prior accepted replay artifacts,
run the tool in a Python environment with Capstone. Its default regenerates the
isolated reference and replays the two fixed scripts and four mutations.
`--replay /private/tmp/pf-drop-witness.json` rebuilds the report from saved metadata.
No original payload is copied to the repository or output artifact.

Validation covers linked chain correspondence, shared words/limits, same/next
scan ordering, exact arithmetic, deterministic capture and equality witnesses,
slot/identity survival, wait-age boundaries, expiry ordering, drop state,
absence of HOLDSTILL clear, expired/program/cursor preservation, first matrix
visit, held late physics, and concrete no-input QUIT. Actual allocator mutation
moves child toslot3 and fire35997; wait mutation fires35999; state mutation
changes coordinates/high; incorrect Hold-clear mutation moves Y48128→48180 in
same-calculation late physics. All four are counterfactual test runs,
not reachability witnesses for their mutated states.

New-pass tests: PASS (30). Available canonical-A/reference tests and focused
A/B/C/D regressions: PASS. Known PF6 baseline is reproduced, not promoted to
PASS. TestOriginalTrajectories: NOT AVAILABLE (missing private
`analysis/pf2-ball-locations.json`). Full demo research suite: PASS (543). git diff --check, untracked-file
whitespace checks and payload checks: PASS. Detailed logs and their precise
validation scope are recorded in the final JSON.

No production internal/hosts/cmd edits, profile, DMO1 work, .DS_Store change,
commit, push, tag, release or v0.1.3 change.
