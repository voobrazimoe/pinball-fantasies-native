# DMO-IMPL-2B3: isolated PARTY_ONTS matrix program

2026-10-08. Owner approval: DMO-IMPL-2B3 ONLY.
Branch `main`, HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

DEMO_PARTY_ONTS = READY

PARTY_ON_TASK1_BODY = NOT_IMPLEMENTED

READY covers all five consumers in the isolated `dmoimpl1` test-only candidate,
with pinned private demo presentation inputs. The program's final routine is
intentionally persistent: PARTYRUT returns without changing SI=1. Completing
this implementation does not mean reaching its trailing zero word, firing the
task body, registering a profile, or producing an official fresh demo witness.
Without verified presentation input, the candidate fails closed at PARTYONN.

## Interfaces and changed files

- `internal/partyland/demo_core_test.go`: `loadPartyPresentation(demo,
  canonical []byte) error`, per-command admission, persistent PARTYRUT lookahead,
  bounds checking and shared flash phases. The private adapter verifies the
  existing pinned installation identity before calling this loader. It checks
  text extent/terminator, byte correspondence with the loaded canonical display,
  and all 42 font13 glyphs, then owns a copy of the actual demo text. The existing
  canonical-backed Display supplies the byte-identical glyph consumer. No
  guessed text, replacement glyphs or demo decoder/profile is introduced.
- `internal/partyland/demo_connected_test.go`: trailing zero metadata, updated
  program contract and sticky-loader checks.
- `internal/partyland/demo_core_checks_test.go`: bounded linked handler, operand,
  store, player-encoding and PARTYRUT checks in the existing private input test.
  No graph, reachability or trajectory search is called.
- `internal/partyland/demo_drain_test.go`: accepted 2B2 rejection expectations
  move from FLASHON to unverified PARTYONN; nested rejection uses a malformed
  FLASHON operand. Shared lamp ticking permits only LampChanged suffix events
  after the initial CLEAR4 event in the preserved drain test.
- `internal/partyland/demo_party_matrix_test.go`: new connected structural
  tests, private adapter and read-only flash-word observations via reflection.
- This report.

No production interface or production file was changed by 2B3. The existing
accepted game.go, regions.go, ball.go and staged.go work is retained. Shared
production files were byte-compared with the private regression checkout.
The ordinary A/B/C/D execution order and canonical oracle remain unchanged.
Pre-existing work and `.DS_Store` were not edited. No commit/push/tag/release.

## Source consumers and calculation boundaries

Linked program entry is file `0x1b243`; reviewed words resolve to CLEAR4,
FLASHON(3), PARTYONN(1), PRINT13(DS:7647,336), PARTYON(1), zero. Bounded linked
handlers are file51be,489b,2ef8,4ea4,2f33. Source correspondence is
FANTASIE's CLEAR4/FLASHON/PRINT13/NORMAL_END/print_END and PLAND's
PARTYONN/PARTYON/PARTYRUT. The pinned input check verifies the concrete handler
identities and stores; it does not claim a whole-DOS writer/admission proof.

| Calculation from structural drain | Active routine after matrix phase | NEXT command index | Effects |
| --- | --- | --- | --- |
| Installation, before electronics | CLEAR4, remaining5 | 1 | StartMatrix/KILL_FLASHOR; no visit yet |
| 1..4 | CLEAR4, remaining4..1 | 1 | Existing four clear visits |
| 5 | FLASHON, remaining1 | 2 | Final CLEAR4 visit; dispatch FLASHON |
| 6 | PARTYONN, remaining1 | 3 | FLASHON WAITRUT finishes; PARTYFLASH=true; player text store |
| 7 | PRINT13, remaining1 | 4 | PARTYONN WAITRUT finishes; PRINTTASK executes after animation phase |
| 8 | PARTYON, remaining1 | 5 | PRINT13 WAITRUT finishes; PARTYFLASH=true; PARTYRUT installed |
| 9 onward, if admitted | PARTYON, remaining1 | 5 | PARTYRUT preserves SI; no next dispatch |

These counts use matrix budget=true throughout. The actual shared dispatch
rules are used; no extra calculation is invented for the zero data word and no
local dispatcher bypasses the production primitives. FLASHON, PARTYONN and
PRINT13 each install WAITRUT/SISA=1; their next handler dispatch occurs on its
following admitted visit. PRINTTASK draws in calculation7, the same matrix
phase that dispatches PRINT13, rather than waiting for its WAITRUT visit.

PARTYONN at linked2efc writes DS:d0=true, reads PLAYER DS:3819, adds0x37,
and stores DS:1df1, i.e. PARTY_ON_TEXT[18]. The shared consumer copies the
mutable text buffer before this write. PARTYON at2f33 writes DS:d0=true and
installs PARTYRUT. Its operand1 is the reviewed stream word, not a duration.
Neither consumer starts NEW_BALL or resets the task wait. In scenario A,
PARTYFLASH is true at calculation6, before PARTY_ON_TASK1 can fire.

PRINT13 pins text pointer7647, position336 and font pointerDS:6300. Its source
writes include SISA=1, WAITRUT, height13, row adjustment0xf77c, FONT_ADR,
zero clipping offsets, AX_PRINT/DI_PRINT and PRINTTASK. The existing Display
BeginResolved/FlushPrint/SourceText/Text path consumes the demo bytes and the
verified identical glyph memory. Tests compare actual nonempty dot memory to
an independent display invocation with the same owned inputs and position336;
existing canonical presentation/source suites supply glyph/PRINTTASK coverage.
No payload bytes, rendered captures or commercial text are stored in sources,
reports or test artifacts.

## Flash and presentation state

FLASHON's operand3 is the whole dot-matrix blink period, not lamp channel3 or
a mask bit. The linked handler writes MATRIX_SPEED DS:348c=3, MATRIX_CNT
DS:348e=3, MATRIX_ONOFF DS:3490=true and MATRIX_IS_FLASHING DS:3492=true.
The existing Display.startFlash does these semantic writes. It emits no palette
change at dispatch: Display.On changes on a later blink or KILL_FLASHOR.

The candidate now calls shared Display.Flash at MATRIX_BLINKOR before KEYTASK
and DO_TASKS. Shared lamp flashTick runs after successful tasks, matching
FLASHTHELIGHTS. The two mechanisms remain distinct. No local blink simulator
or ordinary Game.Sync reorder is introduced. A failure during tasks prevents
lamp ticking, matrix and late physics; the earlier MATRIX_BLINKOR mutation is
an already admitted side effect and is retained.

MATRIX_BLINKOR decrements a word once per admitted calculation, independently
of matrix budget. At zero it reloads3, toggles phase and changes Display.On.
FLASHON dispatch at calculation5 starts count3; calculations6/7 leave2/1;
calculation8 reloads3 and turns the matrix off. Three budgetless calculations
then turn it on while matrix cursor and dot memory stay unchanged. Tests also
pause independently at CLEAR4, FLASHON, PARTYONN, PRINT13 and PARTYON.

Normal next-command dispatch inherits flash words. Direct program installation,
including equality35998 replacement, calls StartMatrix/KILL_FLASHOR: flashing
stops and On becomes true, while speed/count/phase remain intact. PARTYFLASH
and the modified text are not reset. Expiry replacement tests cover all four
new active consumers, task identity/age survival, preserved dots with no budget,
new CLEAR4 remaining5/next1 and later rejection before expiry SCROLL. No old
PARTY_ON cursor resumes.

PARTYRUT stays active at next5 and does not invoke the terminator or idle score
consumer. The trailing zero remains metadata. A separate generic terminator
test proves shared FinishRoutine(false) disables blink, emits matrix-on state,
ends the program, and subsequent isolated idle visits do not call canonical
score/panel fallback. This is not described as PARTY_ONTS termination.

## Fallible boundary and retained task gate

Before every new handler the current/next typed gate checks availability and
operands. The former premature final-CLEAR4 rejection shifts correctly: its
fifth visit executes and FLASHON is dispatched. If presentation is unavailable,
the next check rejects before PARTYONN and its text/PARTYFLASH writes. A
malformed next PRINT13 operand rejects after preserving PARTYONN's completed
stores, before the print and before late physics. The final persistent PARTYON
routine does not look ahead into the unreachable zero as if its SI decremented.

UNSUPPORTED_DEMO_TRANSITION remains sticky. Direct loader, installation,
allocation, drain, matrix, electronics and physics calls after failure retain
the first diagnostic and produce no further state mutation. There is no
canonical fallback or rollback of admitted effects.

PARTY_ON_TASK1 remains distinct from PARTY_ONTS: shared first-free insertion,
DS:36c9 wait, limit30, compare-before-increment, same-calculation first visit
and identity/task survival are unchanged. At age30 its next due visit rejects
before WAIT reset, PARTYFLASH body store or NEW_BALL. Scenario A reaches this
gate with PARTYFLASH already true and PARTYRUT active. Scenario B uses no matrix
budget for 30 calculations: CLEAR4 stays remaining5/next1, PARTYFLASH stays
false, age reaches30, and calculation31 rejects before the body. Matrix budget
is therefore not used to conceal an unsupported program consumer.

Scored drain, bonus/progression, PARTY_ON_TASK1 body, NEW_BALL/reset, expiry
SCROLL/FADE/QUIT, interactive demo/profile and five full witnesses remain out
of scope. DMO-IMPL-2B4 is not started.

## Validation

All connected drains are structural: an explicitly prepared ball at y575 moves
through real shared physics into Lost, then the accepted unscored consumer
installs PARTY_ONTS before electronics. They are not fresh official demo
witnesses. Original inputs remain owner-local; missing fixtures were not created.
Logs and build outputs are `/private/tmp/pf-dmo-impl2b3-*`.

| Check | Result |
| --- | --- |
| Preserved DMO-IMPL-1/2A/2B1/2B2 and new 2B3 TestDemo checks, pinned inputs | PASS |
| Connected budget=true, actual dots/stores/blink, persistent PARTYRUT; independent budget=false task age30 | PASS structural |
| Pause at each consumer; expiry at four new consumers; fallible print; terminator/idle; sticky failure | PASS |
| Public fixture-free TestDemo | PASS available checks; original-backed tests skip NOT AVAILABLE |
| Canonical-A/reference Party Land, presentation, source and physics | PASS available suites; five opt-in capture/export checks skip |
| TestOriginalTrajectories and missing Stones pf10 capture | NOT AVAILABLE; excluded, no fixture generation |
| A/B/C/D selected datalayout/frontend compatibility suites | PASS with private inputs |
| PF6Session in combined layout command | Baseline FAIL: score000002311040, ball2, ticks1200; combined exit1 |
| Tagged vet, partyland/physics/presentation/tablelogic | PASS |
| Windows amd64 executable, CGO disabled; macOS c-shared engine | PASS |
| Broad macOS compile-only and unchanged HEAD baseline checkout | Same baseline FAIL: AudioDevice, hostWindow, openHost |
| Public source checker and diff whitespace checks | PASS |
| Six milestone source/report files and two build artifacts payload checks | PASS; whole-file and nontrivial aligned4KiB samples against owned A/B/C/D/demo inputs; not an all-substring proof |
| Fresh official demo candidate witnesses | NOT AVAILABLE; not attempted |

Reproduction uses `./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run
'^TestDemo' -count=1 -v`. Private execution uses the existing canonical-A
snapshot with these test files copied in and PF_10MIN_DEMO_DATA pointing to the
existing pinned owner installation. Reference command includes partyland,
presentation, source and physics, with `-skip
'TestOriginalTrajectories|TestStonesSourcePhysicsAndTwoFlippers'`.

The only remaining unscored post-drain task blocker is PARTY_ON_TASK1's body.
PARTY_ONTS itself has all its consumers implemented and admitted with verified
presentation inputs, ending in its source-defined persistent PARTYRUT state.
