# DMO0 fresh-session attract phase alignment

`ATTRACT_PHASE_ALIGNMENT = NOT_PROVED`.
`FRESH_BYGEL_ENTRY_TRANSFER = NOT_PROVED`.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

This narrow pass preserves `CANONICAL_A_TIMING_INHERITANCE = PROVED` and
`NATIVE_AUDIO_BOUNDARY = PROVED`. It does not run a BYGEL trajectory search.
The previous fresh/scored drain verdicts and membership remain unchanged.
Evidence: `/private/tmp/pf-dmo0-attract-phase-alignment.json`;
reproducer: `tools/audit_10min_demo_attract_phase.py` with the existing private
`PF_10MIN_DEMO_DATA`, `PF_RUNTIME_DATA`, and `PF_DMO0_HISTORICAL_SOURCE` inputs.
Exit 2 is the expected open phase gate, not a test failure.

## Actual startup and attract prefix

All addresses below are linked demo TABLE1 file offsets unless marked DS.
The actual process entry is `0x329f`. Startup establishes DS at `0x32ad..0x32b0`,
passes the configuration destination DS:`0x372d` to INT65, disables callbacks
at `0x336e`, and calls JUST_ONE_TIME_RESET at `0x3373`. That reset writes
**F1F8_CODE = 0x3b at `0x3a74`**: the demo has an automatic one-player start
request. This is linked binary evidence, also explained by the historical
conditional demo startup declaration. It is not an injected input.

Startup calls the table reset at `0x3376` and GO_DEMO_MODE at `0x337a`;
`0x6380` sets DEMOMODE true. `0x33a2` calls the callback-registration routine
at `0x6345`; INT66 registrations retain primary and later identities.
`0x33a5` separately calls INIT_INTS at `0x5c45` (the light initialization
loop). `0x33e2` enables callbacks; `0x33e7` enters MAIN at `0x38fe`.
The artifact preserves these identities and the intervening opaque initialization
calls; this is not a new closure of API/rendering/indirect domains.

The enabled primary callback ADD at `0x4556` increments DS:`0x34ec` by
1030 **before** the DEMOMODE comparison `0x455c` and attract jump `0x4563`.
The disabled callback returns before this ADD. The attract primary own body
returns at `0x6439`; its local control path has no 128-count exit test.
However, a no-key session already has a pending start request. The enabled
later attract callback tests it at `0x64c5`, consumes it at `0x64cf..0x64d2`,
and enters the start path. Consequently, this pass establishes no controlled
128-primary-callback attract dwell. Waiting without a key does not establish
that theorem.

CHECKSTARTKEYS at `0x36fa` samples SCAN_CODE in the foreground. F1/F8 or Enter
can latch a request into DS:`0x3813`; the no-key path returns without clearing
the automatic request. Acceptance occurs in the later callback, not in the
primary ADD's instruction stream. On the selected inherited serialized P/L
prefix, the primary ADD precedes later acceptance. This does not establish a
number of foreground iterations between callbacks.

## The additional counter producer

The linked word starts at zero in the file image, at DS:`0x34ec`. That fact
is not used as a certified runtime entry phase. MAIN contains **INC word
[0x34ec] at `0x39c5`**, with its backedge at `0x3a06`. It is a second concrete
producer in addition to the primary ADD. While HOLDSTILL, MAIN also reads
the word at `0x39d3` and stores held rotation at `0x39d6`.

Thus, with C0 the actual initial runtime word, P enabled primary ADD visits,
and M foreground INC visits, the counter expression is:

```
word = u16(C0 + 1030*P + M)
phase = (C0 + 6*P + M) mod 256
```

The existing 128-step arithmetic recurrence holds only for the isolated ADD
contribution: 127/128/129 increments have low bytes 250/0/6. One MAIN visit
changes that residue. A formula depending only on attract callback count N
would omit an actual writer.

## Start, NEW_BALL and conditional SETBALL contribution

The accepted start calls GO_GAME_MODE (`0x6501` → `0x636e`), generic new-game
reset (`0x6504` → `0x3b53`), table new-game reset (`0x6507` → `0x30b`), music,
and NEW_BALL (`0x650d` → `0xece`). NEW_BALL resets the generic and table
new-ball state (`0xee0`, `0xee3`). The start path clears ADDPLAYERS only at
`0x6510`, after NEW_BALL; its true branch at `0xf67..0xf76` therefore installs
SNART_NEW_BALL. The subsequent start path installs ALLOW_ADDPL.

A **conditional finite task-prefix calculation**, using the local first-free
50-slot scan and compare-before-increment waits, gives:

- SNART_NEW_BALL waits 30 and fires on gameplay scan 31.
- It installs SETBALL; the newly occupied later slot is visited in that same scan.
- SETBALL waits 80 and runs its body on gameplay scan 111.

This calculation is not a new admission/closure of every task or a physical
foreground/callback timing witness. Its task events and scope are explicit in
JSON. The fixed primary contribution is `1030*111 = 114330`, low byte 154;
MAIN contributes separately and is not a fixed K.

| Point | Conditional low-byte expression |
| --- | --- |
| Start acceptance | `(C0 + 6*N + M_accept) mod 256` |
| GO_GAME_MODE / NEW_BALL | Same within the serialized later callback |
| SETBALL installation | `(C0 + 6*(N+31) + M_install) mod 256` |
| SETBALL body | `(C0 + 6*(N+111) + M_body) mod 256` |
| First gameplay spring calculation | `(C0 + 6*(N+1) + M_first) mod 256` |
| First post-SETBALL SPRINGTASK | Same as SETBALL body within the serialized primary callback |

## Native comparison and equation

The native factory constructs an already prepared ball and leaves clock zero.
Its first actual springControl follows one Sync increment, so its low byte
is **6**, not constructor phase zero. The selected semantic comparison is
native first springControl versus demo first post-SETBALL SPRINGTASK.
Native source ordering is checked directly; no trajectory is executed.

Under the explicitly conditional C0=0, M_body=0 arithmetic model, the equation
is `6*(N+111) = 6 mod 256`, giving `N = 18 mod 128`. N=128 does not solve it.
For conditional N=1, the required MAIN residue is 102. Both produce low byte 6
arithmetically; neither is reported as a reachable prefix. JSON contains all
128 conditional N/M residue pairs and a null reachable witness.

## Selected reset-state inspection

Own-body direct stores and recognized ES=DS string destinations were enumerated
for JUST_ONE_TIME_RESET, GO_GAME_MODE, new-game/new-ball routines, RESET_VARS,
RESET_VARS2, task/wait reset, NEW_BALL, its part two and SETBALL. None of these
reviewed destinations overlaps the counter word. Redirecting a direct reset
store or the score REP destination onto the word fails the audit. Opaque callee
and runtime alias effects are not silently classified as preserving, and no
global writer/alias closure is reopened.

Local reset checks cover score's six words, aggregate totals, XXBALLE,
INH_EFF, 50 task words, 50 wait words, held-ball coordinates/velocity, and the
SETBALL coordinates/velocity/HOLDSTILL stores. Light-reset calls, linked spring
charge zero and foreground held rotation are recorded with their limited scope.
These observations do not certify a joined fresh runtime state or whole physics
equivalence. Fresh construction is not promoted while the first counter-phase
transfer fact is missing.

## One smallest unresolved fact

Establish one reachable joint foreground-MAIN/callback residue at the automatic
fresh start's first post-SETBALL SPRINGTASK: actual MAIN INC visits M and enabled
primary ADD visits P, with actual initial C0, must satisfy
`C0 + M + 6*P = 6 (mod 256)` through the real startup/start prefix.

The existence question is unresolved, not disproved. There is no chosen dwell
N or reachable alignment witness. `FRESH_BYGEL_ENTRY_TRANSFER = NOT_PROVED`;
the deterministic BYGEL trajectory search remains for a later pass.

## Verification

- New pass: 16 tests PASS, including increment/branch ordering, automatic-start,
  foreground writer, reset-store/REP mutations, waits, reset-field bindings,
  127/128/129 arithmetic and the explicitly open real-prefix gate.
- Complete private demo research suite: 365 tests PASS (49.311 seconds).
- Available strict-A/native ordering/physics: 18 test/subtest results PASS,
  zero skips; canonical-A harness matches all 284 tracked Go/module files.
- Focused A/B/C/D regression: 97 test/subtest results PASS. The two task-tail
  tests skipped for absent checkout fixtures both PASS separately in the same
  checked canonical-A harness. The first run used reversed C/D fixture paths;
  correcting environment assignments resolved those identity failures.
- No-private new pass: four arithmetic tests PASS, private class clean SKIP.
- `TestOriginalTrajectories` is not run or manufactured; its fixture is not
  required for this phase pass.
- Payload scan: PASS against 249 private files and 11 decoded SDR modules,
  using whole-file comparison and 48,898 distinct nontrivial aligned 4 KiB
  samples, including hex/base64 candidates. This is a sampled payload check,
  not an all-substring theorem.
- `git diff --check`: PASS. Pre-existing tracked diff preserved byte-for-byte;
  production/runtime diff empty; `.DS_Store` size/mtime/ctime/inode unchanged.
No production/runtime implementation is changed. All research changes remain
uncommitted; `.DS_Store` is untouched. No commit, push, tag or release.
