# DMO0 post-collision eventual termination

`POST_COLLISION_TERMINATION = PROVED`

`EVENTUAL_QUIT_NOT_GUARANTEED`

`EXPIRY_INTERLEAVING = NOT_PROVED`

The negative universal result includes legitimate pause: P from the supplied
post-35998 state reaches a closed suspended native transition. It does not
claim an infinite unpaused gameplay trajectory. With an added mandatory-resume
or unpaused-only fairness condition, universal active-game termination remains
UNKNOWN. No such condition is part of this pass's supplied contract.

DMO0 NOT CLOSED. DMO1 NOT STARTED. Production base is
`306d11a0c479c7ac5ee6e245f3f72eacbc665abd`; research HEAD is
`37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
Input: `/private/tmp/pf-dmo0-first-equality-collision.json`.
Output: `/private/tmp/pf-dmo0-post-collision-termination.json`.
Research tool: `tools/audit_10min_demo_post_collision.py`.

## Exact starting boundary and immediate suffix

The actual replay Game is reconstructed through the previously accepted
handoff, rather than constructing a hypothetical threshold state. All supplied
ball, flipper, live-slot, shared-word, guard, score and matrix fields are checked.
The one admitted matrix visit omitted by the earlier primitive-only native
probe is supplied at 35998: SHOWPLAYERSTS clear has four visits remaining.

A consumed audio omission is made explicit without rewriting the input artifact:
its native A `audio_priority=1` did not include the linked demo expiry cue.
DS:c9d supplies position=13, repeat=0, priority=255, played before the established
NEW_BALL reset. That reset sets readiness but does not clear current priority;
the lower-priority spring request is rejected. The completed demo state has
priority=255. This completion changes none of the task/ball/timer/collision
claims, and is required for honest subsequent effect admission.

| Task | Slot / initial age | Shared word / limit | Firing calculation |
| --- | --- | --- | --- |
| SOUNDBRICKUPP | 2 / 1 | DS:36cf / 5 | 36003 |
| SOUNDNEWBALL | 0 / 0 | DS:36d1 / 50 | 36049 |
| SETBALL | 1 / 1 | DS:36d3 / 80 | 36078 |

The waits compare before increment. SETBALL's body sets (297,530), VX=10,
VY=0, HOLDSTILL=false; I_UTSKJUT=true and expired=true. Timer=36078.
Matrix is idle; there is no saved expiry cursor. Its slot is still executing
at the immediate body observer, then suicides on return; all slots are free
at the calculation's end. The late physics pass then changes VY to 7 and X
from 304128 to 304138. Both boundaries are recorded separately.

The linked queued-task store inventory, exact shared wait words and current-slot
suicide stores do not write DS:34cf. The admitted native CLEAR4/PRINT5 consumers
write bounded Display state, not expired. This is a consumed native trace
statement, not a new whole-DOS writer/IRQ closure claim.

SHOWPLAYERSTS takes **seven admitted visits**, including 35998, hence six more
from the starting boundary. Clear completes and dispatches first print on
36002; second print dispatches on 36003; terminator is consumed on 36004.
NEXT_A becomes zero and NODOT/DUMRET is idle. In the chute, idle work shows
score without automatically reinstalling expiry or another player panel.
SETBALL does not replace it. The artifact distinguishes native array indices
from the linked runtime cursor; a stale diagnostic array index is not a saved
expiry continuation.

## Actual continuations

No additional input: after SETBALL the ball settles in the chute at (301,537),
X=308400, Y=549914, VX=0, VY=7. SCORECHANGED=false, score=50030. No subsequent
drain, cue/effect producer, live task or program replacement is invented.
The compressed trajectory retains every examined calculation range; event
rows and complete expiry-state rows are also retained.

This real continuation survives until timer wraps to zero at calculation
65536. At 101534 the counter reaches 35998 again; the equality code has no
expired guard. It directly reinstalls expiry at 0x1ba17, next cursor 0x1ba19.
No tasks remain to replace it. QUIT is reached on **102575**, after 1042 admitted
matrix visits including the install calculation. Wrap is relevant on this
witness; it was not investigated as an unrelated hypothetical case.

Single-launch input: Down on 36079..36100, release at 36101, Right held through
37100. Its first actual drain at **36288 is unscored**, despite expired=true.
The unscored guard precedes the expired scored-drain test. It installs PARTY_ONTS
and queues PARTY_ON_TASK1, which fires at 36318, sets PARTYFLASH=true, and calls
NEW_BALL. That reset preserves the party program. SETBALL fires at 36398;
expired remains true. With no further release, the next equality again installs
expiry at 101534 and QUIT occurs at 102575. Logical electronics accounting
includes the accepted post-drain common task/matrix suffix; it is not simply
the number of physics BeforeTargets callbacks.

Periodic launch: Down on 36206..36227, release at 36228, then with
`d=calculation-36228`, Left iff `(d+46)%52<8`, Right iff `(d+22)%30<21`.
The actual path visits GROPE/dragon and BYGEL3. Its existing capture/ejection
rules are executed normally; no research seed or teleport is added.
Additional bounded A/demo instruction correspondence checks cover those
callbacks, their guards and BEFOREFLASH/DURINGFLASH tasks, plus arithmetic and
region bindings. Optional branches are paired but not claimed reached.

The real **scored drain is at 37084**. It sees old expired=true, SCORECHANGED=true,
INH_EFF=false, SPECIALMODE=false and priority=255. The actual request consumes
effect 0x1a4a1: cue priority=255 is admitted at equality with the old priority.
Arithmetic is zero and its matrix pointer is 0x1ba17. The effect consumer,
not the presence of that pointer, proves installation. Its initial clear has
remaining=5 and next cursor 0x1ba19, before the first matrix visit. No old
cursor is resumed. Both dragon tasks have finished; no NEW_BALL/SETBALL or
other live task remains. Held/stopped BallLost skips controls and further
geometry/drain requests. The traversed matrix operations have no replacement
edge or jingle-readiness wait.

| Expiry operation dispatch | No input | Scored-drain replay |
| --- | --- | --- |
| clear / first admitted visit | 101534 | 37084 |
| first scroll | 101538 | 37088 |
| flash on | 101847 | 37397 |
| score print | 101848 | 37398 |
| wait100 | 101849 | 37399 |
| flash off | 101949 | 37499 |
| second scroll | 101950 | 37500 |
| fade256 | 102219 | 37769 |
| wait100 | 102475 | 38025 |
| QUIT | **102575** | **38125** |

Scroll consumes the actual linked lengths 98 and 88 under the retained native
scroll phase. Only placeholder cells are supplied in the isolated harness;
their values do not feed cadence or gameplay. The accepted native StepScroll
executes the real visits; durations are not supplied as an external completion
premise. Fade binds the linked decrement from 256. Its palette/volume writes
have no NEXT_A admission predicate. QUIT means entry into its linked handler,
which disables gameplay interrupts; final DOS teardown is outside this pass.
Wrap is unreachable on the scored-drain witness.

## Input and theorem boundary

Down/plunger can change launch/drain timing; withholding release postpones drain
but does not defeat the later timer equality on the no-input witness. Flippers
are meaningful, but arbitrary unpaused control cycles remain UNKNOWN. Space
can push the ball and, outside the chute, tilt can install TILTTS; no indefinite
active replacement cycle was established. During the actual BallLost expiry
run, the native reference skips these controls.

The external native frontend test applies **P, empty, empty** to the reconstructed
post-collision Game. The session state, timer, expired and matrix are identical
across both real empty transitions. Structurally, Model.Update has an empty key
loop and no Paused arm in its outer mode switch; Runtime.Update returns before
session audio when suspended. All future empty transitions preserve the same
session. Only the unconsumed uint64 frontend Tick increments: this gives a
full-state recurrence after 2^64 host updates, proved by the transition rule,
not by simulating a long finite run. There is no forced resume or exit. This
is a legitimate replay/cycle witness for EVENTUAL_QUIT_NOT_GUARANTEED.

The same test then executes Esc and Y from pause and verifies Selector,
Aborted and Session=nil: an alternate accepted table exit exists. Ordinary
Esc requires the native SessionReady gate; at the immediate supplied boundary
SpringValid=false, so direct Esc is not yet admitted. Pause and later exit
must not be confused with an expiry operation.

The required witness behavior includes sticky expired, active SETBALL,
unscored PARTY_ON despite expiry, admitted scored replay from entry, later
uint16 equality, and suspended pause semantics. Atomic timer shutdown remains
disproved. These concrete results do **not** close EXPIRY_INTERLEAVING for all
unpaused input/replacement continuations. The smallest active-only dependency
is whether every infinite unpaused continuation admits an expiry installation
that cannot subsequently be replaced before QUIT. Separately, the first-equality
class with PARTY_ON_TASK1 age30 still has UNKNOWN reachability; the later
unscored drain here does not prove that independent threshold class.

## Validation and scope

New tests cover state identity, slot-age asymmetry, exact task firing, SETBALL
body/end boundaries, expired writers, panel termination/absence of restoration,
real unscored and scored drains, actual effect admission, fresh expiry entry,
complete second expiry, QUIT, pause-cycle closure and actual alternate exit.
Runtime mutations clear expired at SETBALL, block effect admission at the real
drain, or replace the freshly installed expiry. Each changes the consumed
result and fails positive-witness validation. These finite mutation prefixes
are never counted as nontermination witnesses. Linked operand mutations also
reopen the certificate. Validation results and log paths are in the artifact.

No production internal/hosts/cmd source changes, fixture generation for
TestOriginalTrajectories, .DS_Store edits, commit, push, tag, release or v0.1.3
change. Earlier threshold/predecessor, historical phase, PIT/IRQ, persistence,
INTRO, DMO1 and global control/writer obligations are not reopened.

Current-pass results: 27 new Python tests PASS; both native continuation/pause
harness tests PASS; all 460 demo research tests PASS; 22 available canonical-A
ordering/reference tests PASS without skips; 11 focused A/B/C/D tests PASS
without skips. The PF6 baseline was reproduced exactly at 2311040 versus
2300000, ball2/ticks1200. TestOriginalTrajectories remains NOT AVAILABLE and was
not generated. Diff whitespace and production-path checks PASS. Payload scan
PASS: 249 private files, eleven decoded SDR modules, 48,898 nontrivial aligned
4 KiB samples and zero encoded candidates. This checks whole-file identity and
sampled copies, not every possible substring.
