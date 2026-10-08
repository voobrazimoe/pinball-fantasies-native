DMO0 NOT CLOSED. DMO1 NOT STARTED.

# DMO0 indirect control domains: unfinished local research

Date: 2026-10-07. Starting local research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`. Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.

## CURRENT RESEARCH SNAPSHOT — 2026-10-07, native audio boundary pass

The separate [native audio/gameplay boundary report](runtime-layout-demo-10min-audio-boundary.md)
records **NATIVE_AUDIO_BOUNDARY = NOT_PROVED**. Its new owner-local artifact is
`/private/tmp/pf-dmo0-audio-boundary.json`; the authoritative historical domain
result remains `/private/tmp/pf-dmo0-admission-domains.json`, unchanged.
TABLE1 **30 total / 18 bounded / 12 UNKNOWN**, CODE2 **2 total / 0 bounded /
2 UNKNOWN**, whole-DOS **exit 2**. DMO0 NOT CLOSED. DMO1 NOT STARTED.

New evidence checks all eleven supplied driver budget/return handshakes and
local callback-body CFGs. Incoming AX and shared TIME_LEFT are observable;
API12 latch release, priority skipping and possible nesting prevent a proved
cadence-only abstraction. Local own-body electronics counts {0,1} are not a
transitive scheduler invariant. Low-res API12 CX is 0x108 (264 decimal).
Physical DOS placement/IVT/stack obligations are classified separately from
the unresolved paired callback semantic transition. No earlier UNKNOWN was
promoted, no previous JSON scope changed, and no production code was added.

This pass's verification is recorded in that report: new tests 14 PASS,
current complete demo suite 145 PASS, focused A/B/C/D regressions PASS.
Historical verification counts below belong to their respective previous passes.

**The callback/vector admission pass retains 0 bounded / 2 UNKNOWN CODE2 sites. TABLE1 remains 18 bounded / 12 UNKNOWN. No site is promoted; all research changes remain uncommitted.** The new evidence exposes missing candidates and classifies far operands; it does not substitute initial literals or installer discovery for complete writer/admission proofs.

Previous-pass machine-readable evidence: `/private/tmp/pf-dmo0-control-domains.json`. It includes definitions, tables, operand classes, module targets, installer stores, writer inventories and explicit unresolved reasons. Candidate sets below must not be represented as mechanically closed domains.

## Callback root and INT33/INT66 admission

Owner-local machine-readable evidence: `/private/tmp/pf-dmo0-admission-domains.json`,
key `callback_and_interrupt_admission`. New extractor:
`tools/audit_10min_demo_admission.py`. All instruction offsets below are semantic
metadata; no private executable slices or instruction listings are exported.

**Result: UNKNOWN. No CODE2 or TABLE1 promotion.** The decisive new binding
obstruction is an unrestricted external executable name, rather than a missing
search for the eleven known supplied installers.

TABLE1 opens `SOUND.CFG` at `0x6622`, requests thirteen bytes at `0x663a` into
DS:`0x23c0`, and supplies that same buffer to DOS normal EXEC at `0x665a`.
The mechanically checked read/close/EXEC interval has no conditional branch,
CALL, filename allowlist, or read-length validation. The EXEC result is not
checked before the subsequent INT66 at `0x6663`. This derives an actual binding
producer path while disproving the inference that only the eleven supplied
SDR identities can execute: neither the configuration filename contents nor
the filesystem executable selected by it are pinned by the research input
certificate. An existing or replacement executable can install another INT66
handler, overwrite it, or leave the initial environmental vector untouched.
This is an admission obligation; it says nothing about cadence.

### Callback registration and invocation candidates

At `0x6345` the registration producer selects INT66 API 11; its DX producer
supplies TABLE1 IP `0x4217`, ES is produced from CS, and the registration
interrupt is at `0x634f`. The linked file root is `0x4517`. The existing
per-driver API-11 pointer stores are reused. API-12 registration remains
separate; its later-callback consumers do not prove admission of this root.

The new discovery follows direct candidate driver graphs and finds a primary
far-call consumer in each supplied SDR. These are **candidate invocation
sources**, not a complete admitted source set. For indexed consumers the
linkage is conditional on SI selecting the first record: the pointer operand
is four bytes before the record, and the primary pointer is at that position.
Record traversal/alias and source admission completeness are not claimed.

| SDR | INT66 handler offset (resident CS) | Candidate callback graph start | Primary far-call consumer | Linkage |
| --- | --- | --- | --- | --- |
| ADLIB | `0x49` | `0xb13` | `0xbdd` | absolute primary pointer |
| GUS | `0xf7` | `0x6b7` | `0x77e` | indexed record pointer |
| INTERNAL | `0x4e` | `0xb15` | `0xbdf` | absolute primary pointer |
| NOSOUND | `0x40` | `0x5bf` | `0x687` | indexed record pointer |
| PAS16 | `0x49` | `0x1b50` | `0x1c0d` | indexed record pointer |
| SB16 | `0x49` | `0x1d0d` | `0x1dca` | indexed record pointer |
| SB20 | `0x49` | `0x1d75` | `0x1e32` | indexed record pointer |
| SBLASTER | `0x49` | `0x1c2f` | `0x1cec` | indexed record pointer |
| SBPRO | `0x49` | `0x1dac` | `0x1e69` | indexed record pointer |
| SM2 | `0x49` | `0x1bb2` | `0x1c6f` | indexed record pointer |
| THING | `0x49` | `0xb0a` | `0xbd4` | absolute primary pointer |

Graph starts are discovered prologue candidates, and are not all interrupt
entry addresses. In particular the direct-pointer family has an earlier
source frame outside the displayed graph start. Candidate graphs retain
indirect frontiers; discovery never certifies an exhaustive interrupt root set.

All eleven discovered consumer instructions are far CALLs, whose own frame
is four bytes: CS then IP are pushed, giving entry `(source SS,
u16(source SP-4))`, with IP at the top. Each has an immediately preceding
two-byte argument push. Thus relative to the state before that argument the
callback SP is `u16(P-6)`. This is an exact local architectural transform,
not a bounded source SP domain. An interrupt source frame, register saves,
and any nested driver calls must be included separately before deriving a
relationship to an interrupted task's stack. No near-call ABI is substituted.

The callback prologue assigns relocated TABLE1 DS. ES at each consumer,
complete callback DS/ES return effects, full object-memory effects, matched
RETF balance, and restored source SS:SP remain unproved. Driver source IRET
candidates and INT66-handler IRET candidates are separately recorded. Neither
is a proof that every callback path returns with a balanced RETF. Callback
admitted invocation set, entry SS/SP and return transform remain null with
`complete=false`; finite candidates are kept in separate fields.

### Vector lifecycle and handler effects

INT66's IVT object is `0000:0198..019b`. The existing eleven entry-path
installer store certificates are retained. Their resident handler candidates
are the offsets in the table above; resident CS identity and physical placement
remain driver-context dependent. The **complete vector-writer set and complete
installed target set are UNKNOWN**. Local EXEC-before-consumer fallthrough
ordering exists, but it does not establish successful installation of a pinned
executable, exclude later overwrites, or close uninstall/restore and alias paths.
Initial IVT state is environmental. A syntactic direct-writer inventory is
included and explicitly cannot exclude indexed or external writers.

For every supplied handler the extractor mechanically checks the common save
prologue: twenty local bytes (PUSHA, DS, ES), in addition to the architectural
six-byte FLAGS/CS/IP frame. Handler entry is `(caller SS,u16(caller SP-6))`.
That prologue establishes a local save count, **not a maximum stack-use bound**.
For every row in the table, maximum additional downward use, final DS/ES/SS/SP,
scoped-object writes, allocation/free/resize effects, and complete IRET return
contracts remain UNKNOWN. IRET candidates alone are insufficient. Arbitrary
other-memory stores are allowed by the certificate validator when their
placement is proved disjoint; tests verify this and both table-overlap failures.

Required INT33 at `0x618f` has AX=3 and IVT object `0000:00cc..00cf`.
No fixed TABLE1 installation is proved: the existing decoded instruction
inventory finds no immediate DOS set-vector producer for it. Absolute
slot-offset stores with unproved/nonzero segments are retained as candidates;
this is not a complete alias exclusion. The target identity is environmental
and unknown. More decisively, no invariant handler stack/register/target-memory
contract has been established. Handler identity variation could be harmless
if a sufficient complete invariant effect certificate existed; the authored
validator test accepts that case. **The real contract itself is UNKNOWN**.
No particular mouse driver or DOS implementation is inferred. Architectural
entry uses the unchanged caller SS and SP minus six, but handler return and
memory effects remain unknown.

### Matched returns and resulting CODE2 entry domains

`0x4726` retains the exact existing two-target candidate set `{0x6182,0x61df}`.
Its target-admission flag remains incomplete. The two return summaries are
UNKNOWN specifically at the already required `0x618f` INT33 and `0x6256` INT66
frontiers. No independent second unknown ABI is invented for these frontiers.

`0x573e` is a direct near CALL to `0x5756`; direct target admission is complete.
Its matched stack summary remains UNKNOWN at nested CALL `0x579c` to
`0x4801`. That transitive summary retains unresolved callee effects at
`0x3cfb`, `0x4aa1`, `0x785`, `0xe4b`, `0xe62`, and a cycle/depth/resource
guard. These are not discharged merely by naming the two required INT66
frontiers; finite rejoin still needs a proof. Neither
matched-return dependency proves complete return balance, finite rejoin depth,
or SS/SP preservation while the external contracts are open. JSON retains
exact targets, return kinds, nested summaries and reasons from the existing
path-sensitive analysis; gameplay semantics are not expanded.

| TABLE1 caller | Required frontiers | Caller SS:SP | CODE2 SS:SP | Status |
| --- | --- | --- | --- | --- |
| `0x66a` | `0x6ec`: INT66 | UNKNOWN | UNKNOWN | UNKNOWN |
| `0xae9` | `0x618f`: INT33; `0x6256`: INT66 | UNKNOWN | UNKNOWN | UNKNOWN |
| `0x5673` | `0x618f`: INT33; `0x6256`: INT66 | UNKNOWN | UNKNOWN | UNKNOWN |
| `0x574d` | `0x55ba`, `0x5fdc`: INT66 | UNKNOWN | UNKNOWN | UNKNOWN |

Startup SS=B/SP=0x100 is not copied into callback entry. Startup INT21 at
`0x32ab` is not added as a callback-stack premise: even a startup preservation
contract would not constrain an unrestricted driver's callback stack.
The existing six-byte startup-frame exclusion is unchanged.

For a finite caller context the preserved calculation is four far-frame bytes
plus ten local CODE2 bytes, `u16([P-14,P))`; physical domains retain all legal
load bases and both modeled A20 states. No real finite source context has been
proved, so neither lookup-table intersection is proved empty. **CODE2 stack
sites: 13 total, 0 closed / 13 unresolved.** The previously proved local
callee/explicit-writer facts are retained without a new global TABLE1 stack
analysis. `0xafa6` and `0xaff7` both remain UNKNOWN; their complete conditional
candidate effects remain bounded finite effects. **CODE2: 2 / 0 bounded /
2 unknown. TABLE1: 30 / 18 bounded / 12 unknown. Default gate: exit 2.**

For TABLE1 external API sites `0x3a01`, `0x3d13`, `0x5640`, binding/admission is
**partially discharged**: supplied producer candidates and the external name
flow are derived, while exhaustive vector admission remains open. Their own
transfer/object-memory contracts are not closed; no automatic promotion occurs.

### Validation and scope preservation

- Complete private Python demo suite: 131 tests PASS, zero SKIP. Includes all
  existing domain/fixed-point, writer/placement, stack/transfer and entry tests,
  27 new authored callback/vector mutations and two new real-evidence tests.
- No-private Python: 90 reported outcomes, 88 PASS and two private classes clean
  SKIP. Capstone is available for authored instruction tests.
- Focused private Go A/B/C/D regression: PASS, zero SKIP. Matrix 1024, C/D seeds,
  Stones mappings, B/C/D priorities, shared construction and production demo
  rejection/hybrids all pass. Focused no-private Go: public checks PASS and
  fixture-dependent cases clean SKIP. A separate broad package run also passes;
  its unrelated optional oracle cases SKIP and are not claimed as acceptance.
- Payload scan: PASS against 249 private files, eleven decoded SDR modules,
  48,898 distinct nontrivial aligned 4 KiB samples and whole-file identities.
  This remains a sampled copy check, not an all-substring proof.
- `git diff --check`: PASS. No production/runtime source changes. `.DS_Store`
  inode/size/mtime/ctime match the entry snapshot; it was not written or staged.

Logs: `/private/tmp/pf-dmo0-admission-{python-tests,public-python-tests,
focused-go-tests,focused-public-go-tests,mutations,gate}.log`.
Payload metadata: `/private/tmp/pf-dmo0-stack-payload-scan.json`.
Changed in this pass: this report, `tools/audit_10min_demo_control.py`,
`tools/audit_10min_demo_entry.py`, `tools/test_audit_10min_demo_domains.py`,
and new `tools/audit_10min_demo_admission.py`,
`tools/test_audit_10min_demo_admission.py`. All earlier research changes retained.

Remaining blockers are exhaustive external executable/vector/callback admission;
bounded source callback SS:SP; complete callback return/object-memory effects;
INT66 handler return/stack/register/placement effects; and INT33's sufficient
invariant contract. Matched-return frontiers reference those contracts; the
`0x573e` nested `0x4801` finite-rejoin/callee-effect guard also remains open. No
callback cadence, IRQ frequency, timer equivalence, task/persistence/INTRO
closure or DMO1 work is added. **DMO0 NOT CLOSED. DMO1 NOT STARTED.**
No commit, push, tag or release; research changes remain uncommitted.

## Stack and transfer-effects closure

Current owner-local JSON: `/private/tmp/pf-dmo0-stack-domains.json`.
The new research-only `audit_10min_demo_stack.py` consumes the existing typed
CFG, writer inventory and placement certificate. Candidate target sets,
relocation identities, DS/ES proofs and placement exclusions are unchanged.
No site is promoted. **DMO0 NOT CLOSED. DMO1 NOT STARTED.**

### CODE2 entry stack and handler contracts

The preceding entry-only pass added `tools/audit_10min_demo_entry.py`. Its proof scope is the typed
candidate TABLE1→CODE2 entry projection, not TABLE1 far-pointer promotion.
JSON `object_writer_evidence[1].placement_effects.entry_stack_provenance`
contains every caller, backward slice, reaching SS/SP definitions, matched
return-effect dependencies, handler frontiers, far frame and physical proofs.
No private executable bytes or instruction listings are included in the report.

| TABLE1 caller | CODE2 entry | Backward instruction slice | Incoming TABLE1 SS / SP | CODE2 entry SS / SP | Required interrupt frontiers |
| --- | --- | --- | --- | --- | --- |
| `0x66a` | `0xaed0` | 328 sites, external root `0x4517` | UNKNOWN / UNKNOWN | UNKNOWN / UNKNOWN | `0x6ec`: INT 66h |
| `0xae9` | `0xaed0` | 275 sites, external root `0x4517` | UNKNOWN / UNKNOWN | UNKNOWN / UNKNOWN | `0x618f`: INT 33h; `0x6256`: INT 66h |
| `0x5673` | `0xaed0` | 263 sites, external root `0x4517` | UNKNOWN / UNKNOWN | UNKNOWN / UNKNOWN | `0x618f`: INT 33h; `0x6256`: INT 66h |
| `0x574d` | `0xaed0` | 312 sites, external root `0x4517` | UNKNOWN / UNKNOWN | UNKNOWN / UNKNOWN | `0x55ba`, `0x5fdc`: INT 66h |

The slice sizes cover structural predecessor edges. Interrupt frontiers in
matched callee-return dependencies are listed separately, including nested
summaries; these are conservative admitted candidate paths, not claims of
state-feasible execution. `0x66a` reaching definitions stop at `0x6ec`.
`0xae9` and `0x5673` require the return effects of `0x4726`, whose existing
two-target domain reaches `0x6182` and `0x61df`; the former encounters
`0x618f`, the latter `0x6256`. Both also reach the uncontracted external root.
`0x574d` requires `0x573e`→`0x5756`→call `0x579c`; nested stack summaries retain
interrupt frontiers and incomplete return/cycle obligations. No unknown
callee is assigned a guessed stack allowance. No reachable explicit SS/SP
assignment is found in these instruction slices and explored matched-return
dependencies; this does not prove the external callback stack contract.

All four structural slices terminate at `0x4517`, the existing primary
callback root. Its registration is already recorded at `0x6345`; registration
of an address supplies no admitted entry SS:SP. No ordinary decoded call edge
connects MZ startup to that external callback root. The initial MZ domain
cannot be copied onto it. INT66 is an actual required dependency of this
projection, so its bindings/scheduling are left UNKNOWN under the requested
scope boundary. No SDR callback/IRQ analysis, task admission, PRINTTASK or
KEYTASK initialization, area or explicit-alias analysis is added.

At the actual MZ entry `0x329f`, the existing header/EXEC certificate derives
**SS=B, SP=0x100**, with legal paragraph bases `16 <= B <= 31968`. The SS token
is `image:0x200`, retaining header-relative/load-base identity; SP is literal.
The mechanically decoded startup prefix reaches `0x32ab` without an SS/SP
assignment. Its INT 21h (AH=4Ah) uses IVT slot `0000:0084`. The immediate
FLAGS/CS/IP stores each occupy two bytes: `[SP-2,SP)`, `[SP-4,SP-2)`, and
`[SP-6,SP-4)`, respectively. Their union is `[0xfa,0x100)`, physically
`16*B+[0xfa,0x100)`, disjoint from both CODE2 lookup tables for all legal B.
The concrete DOS/vector target and relevant handler return/memory/placement
contract are not supplied. This remains an unresolved startup frontier,
separate from the callback-relative slices. Later startup interrupts are not
expanded after this unresolved frontier.

For required INT33/INT66 sites, the fixed IVT slots are `0000:00cc` and
`0000:0198`. Their admitted handler target sets are **UNKNOWN**, not empty
proved sets. JSON uses `admitted_target_set=null`, `target_set_complete=false`.
SS/SP/DS/ES, handler stack writes, scoped object memory, placement and IRET
return effects consequently remain UNKNOWN. Their six-byte architectural
frames are recorded separately; an excluded frame never grants a handler
contract. No generic DOS/BIOS preservation statement or inferred OS version
is used. Existing SDR installer candidates are not promoted into complete
installed-target sets.

For a bounded caller context `(S,P)`, the retained far frame gives CODE2
entry `(S,u16(P-4))`. The unchanged local bound gives the complete possible
stack-write offsets `u16([P-14,P))`: four caller-frame bytes plus ten local
bytes. Physical destinations are `16*S` plus those offsets, with offset wrap,
both A20 states and all correlated legal B preserved. The two half-open
lookup-table physical ranges are:

- `16*B + [0x1f6b0,0x1f8b0)` for file `0x1f8b0..0x1faaf`;
- `16*B + [0x1f8b0,0x1fab0)` for file `0x1fab0..0x1fcaf`.

No admitted real `(S,P)` domain is currently bounded. UNKNOWN SS retains all
real-mode physical destinations and can alias either table. Therefore the
common entry-stack proof remains UNKNOWN: **13 CODE2 stack sites: 0 closed /
13 unresolved**. No per-site allowlist is introduced. Existing 510 callee
DS/ES/SS/SP preservation certificates, 256 raw / 255 distinct targets per
consumer, local ten-byte bound and zero possible explicit CODE2 writers are
retained.

Both transfers `0xafa6` and `0xaff7` remain **UNKNOWN**, with unchanged complete
conditional candidate effects classified **bounded finite effect**. Runtime
lookup-table immutability is **incomplete**: required callback/handler effects
and the separate upstream target-memory/admission certificate remain open.
The existing upstream possible-writer records are retained without analyzing
the 1233 far-pointer aliases or promoting any TABLE1 site.

Global TABLE1 stack completeness is not a premise of the new CODE2-entry
proof. Reaching definitions follow caller predecessors and matched return
dependencies; unrelated stack roots/sites never join its input domains.
A stack-only callee summary ignores unrelated explicit-store uncertainty,
and is explicitly prohibited from authorizing handler-memory or transfer
closure. Authored mutations prove that an unrelated unknown TABLE1 stack
path leaves the entry proof BOUNDED, while a new unknown path reaching the
caller makes it UNKNOWN. Multiple admitted finite contexts are retained.
This path distinction does not discharge the real reachable callback and
interrupt dependencies.

Final global counts remain **TABLE1 30 / 18 BOUNDED / 12 UNKNOWN; CODE2 2 /
0 BOUNDED / 2 UNKNOWN**, default domain command **exit 2**. No real passing
CODE2 closure assertion is added. **DMO0 NOT CLOSED. DMO1 NOT STARTED.**

This pass's validation:

- Complete private Python demo suite: **102 tests PASS**, no SKIP, including
  all domain/fixed-point, writer/placement, stack/transfer tests, 28 new authored
  entry/handler mutations and one new real unresolved-provenance test.
- Focused private A/B/C/D Go: **PASS**, no SKIP; 1024 hybrid matrix, C/D seeds,
  Stones mappings, B/C/D priorities, shared runtime and production demo rejection.
- No-private Python: **63 tests PASS, two private classes clean SKIP**.
  No-private Go: public checks PASS, private input-dependent cases clean SKIP.
- Payload-copy scan of all changed research files: PASS (249 private files,
  eleven decoded SDR modules, 48,898 distinct aligned nontrivial 4 KiB samples;
  whole-file identities also checked). This is a sampled check, not a proof
  over every possible substring.
- `git diff --check`: PASS; no production/runtime source change relative to
  research HEAD. `.DS_Store` inode, size, mtime and ctime match the entry snapshot;
  no write/stage/delete operation was performed on it.

Logs use `/private/tmp/pf-dmo0-entry-{python-tests,public-python-tests,go-tests,
public-go-tests,mutations,gate}.log`. Payload metadata remains at
`/private/tmp/pf-dmo0-stack-payload-scan.json`. Files changed during this pass:
this report, `tools/audit_10min_demo_control.py`,
`tools/audit_10min_demo_stack.py`, `tools/test_audit_10min_demo_domains.py`,
and new `tools/audit_10min_demo_entry.py`, `tools/test_audit_10min_demo_entry.py`.
All earlier research changes are retained. No commit, push, tag or release.

### CODE2: exact local effects, open runtime stack admission

**13 stack sites: 0 closed / 13 unresolved.** They reduce to nine 16-bit PUSH
and four 16-bit near CALL operations. Each writes exactly two bytes, downward
from incoming SP modulo 65536. POP and RET are reads with upward SP changes;
the CODE2 entry returns with RETF, consuming four caller bytes. There is no
admitted CODE2 INT, ENTER/LEAVE, PUSHF/POPF or SS/SP assignment. No IRQ path is
invented. JSON `stack_transitions` retains all decoded stack operations,
including RETs reached from all 510 candidates and segment-register PUSH/POP.

The finite local stack derivation is conditional on the unchanged enumerated
candidate domains. Matched calls/returns and save/restore pairs prove a maximum
**10 bytes downward from CODE2 entry SP**, including the nested CALL DX word.
The far caller contributes another four bytes: **14 bytes from caller SP**.
Balanced loops do not accumulate frames; candidate callees add no stack saves
and perform no further calls. There is no arbitrary stack allowance.

| Site | Operation | Incoming SP relative to CODE2 entry | Written relative interval |
| --- | --- | --- | --- |
| `0xaed0` | PUSH DS | 0 | [-2, 0) |
| `0xaed1` | PUSH relocated image segment | -2 | [-4, -2) |
| `0xaed5` | PUSH relocated image segment | -2 | [-4, -2) |
| `0xaed9` | PUSH BX | -2 | [-4, -2) |
| `0xaeda` | PUSH SI | -4 | [-6, -4) |
| `0xaedb` | near CALL | -6 | [-8, -6) |
| `0xaee0` | near CALL | -2 | [-4, -2) |
| `0xaf5e` | PUSH absolute segment | -8 | [-10, -8) |
| `0xaf62` | PUSH relocated image segment | -8 | [-10, -8) |
| `0xafa6` | CALL DX | -8 | [-10, -8) |
| `0xaff7` | CALL DX | -8 | [-10, -8) |
| `0xb005` | PUSH absolute segment | -4 | [-6, -4) |
| `0xb009` | PUSH relocated image segment | -4 | [-6, -4) |

These relative bounds are not a runtime SS:SP certificate. All four TABLE1
far-call contexts have UNKNOWN incoming SS and SP. The actual MZ initial
SS:SP is B:0x100, but admitted startup INT `0x32ab`, subsequent admitted INT/API
boundaries and callback/root contexts have no closed SS:SP effect/admission
contract. No ABI preservation is assumed. For every CODE2 site, unknown SS
admits the full physical domain (and both A20 states), so overlap with each of
the four objects remains possible. Even a 10-byte local bound cannot exclude
an arbitrary incoming SS:SP. The JSON keeps caller evidence, conditional
relative ranges, architectural width, physical domains and each object's
placement comparison separately.

**Two CODE2 transfer boundaries: both UNKNOWN at runtime.** The reusable
candidate effect summaries are complete for all 255 distinct callees at each
site (256 raw entries retained). Every callee has a matched normal near RET,
preserves DS/ES/SS and caller SP, has zero additional stack extent, and contains
no admitted allocation/free/replacement mechanism or other transfer. All
explicit memory operands are excluded from all four objects by the unchanged
placement certificates; writes occupy the already derived absolute segment
`0xa000`, whose full offset domain is disjoint from the objects for every legal
B. Thus each candidate summary has a **bounded finite effect**: memory writes
exist, so it is not a no-write summary. Its runtime classification is UNKNOWN
because table immutability/admitted target contents are not yet proved.

| Boundary | Candidate domain | DS / ES / SS / SP | Object memory / allocation / placement / return | Runtime classification |
| --- | --- | --- | --- | --- |
| `0xafa6` | unchanged 255 typed CODE2 entries | preserving | explicit writes disjoint; no remapping; matched near return | UNKNOWN: table admission/immutability |
| `0xaff7` | unchanged 255 typed CODE2 entries | preserving | explicit writes disjoint; no remapping; matched near return | UNKNOWN: table admission/immutability |

The shared CODE2 entry summary derives DS/SS/SP preservation, **ES=absolute
0xa000 on return**, matched far return, and the conditional 10-byte extent.
It explicitly references the two still-open nested table boundaries. It does
not mark the four relocated pointer objects immutable.

CODE2 remains **2 indirect sites / 0 BOUNDED / 2 UNKNOWN**. No passing real
closure test is fabricated: the real-obligation test asserts the open caller,
stack and transfer records. Adding an admitted callee with unknown SS or
possible object-memory effects makes candidate effects incomplete as well.

### TABLE1: seven conservative stack equivalence classes

**1420 sites / 7 classes / 0 closed / 1420 unresolved.** The grouping is by
mechanically equal abstract incoming SS/SP domains and instruction effects;
it does not assert that the underlying concrete runtime states are identical.
The JSON includes every group's complete site list and all individual records.
Unknown SP is the entire 16-bit offset domain, not a guessed depth. Physical
unknown SS covers all destinations; no write class is discarded by its size.

| Class | Sites | Representatives | Incoming SS / SP | Maximum architectural write domain | Result |
| --- | --- | --- | --- | --- | --- |
| near CALL, 2 bytes | 821 | `0x304`, `0x307`, `0x310` | UNKNOWN / UNKNOWN | all 65536 offsets; physical domain UNKNOWN | UNKNOWN |
| PUSH, 2 bytes | 482 | `0x31c`, `0x33b`, `0x33c` | UNKNOWN / UNKNOWN | all 65536 offsets; physical domain UNKNOWN | UNKNOWN |
| admitted INT | 69 | `0x61d`, `0x6ec`, `0xfc9` | UNKNOWN / UNKNOWN | six immediate bytes plus unknown handler extent | UNKNOWN |
| far CALL, 4 bytes | 7 | `0x66a`, `0xae9`, `0x3a01` | UNKNOWN / UNKNOWN | all 65536 offsets; physical domain UNKNOWN | UNKNOWN |
| PUSHA, 16 bytes | 36 | `0x185c`, `0x22c5`, `0x22dd` | UNKNOWN / UNKNOWN | all 65536 offsets; physical domain UNKNOWN | UNKNOWN |
| startup INT | 1 | `0x32ab` | image B / 0x100 | immediate [0xfa, 0x100); handler UNKNOWN | UNKNOWN |
| PUSHF, 2 bytes | 4 | `0x3a00`, `0x3d12`, `0x563f` | UNKNOWN / UNKNOWN | all 65536 offsets; physical domain UNKNOWN | UNKNOWN |

The startup INT's **architectural six-byte write** is independently EXCLUDED
from all objects: physical [16*B+0xfa, 16*B+0x100), envelope
[0x1fa, 0x7cf00). Its handler SS/SP/depth/memory effects remain UNKNOWN, so the
whole site's obligation remains UNKNOWN. The 70 actually admitted TABLE1
INT instructions are listed separately; no interrupt scheduler is modeled.
Resource guards, unsupported ENTER/LEAVE/SP adjustments, unmatched stack
operations and recursive SCCs without bounded admission fail closed. No
acyclic/global TABLE1 depth or handler-depth proof is claimed.

### TABLE1: the existing twelve transfer boundaries

All twelve retain **UNKNOWN** runtime classification. The four typed module
boundaries now carry the reusable CODE2 entry summary and explicit CODE2
identity. No unrelated target/admission domain was changed.

| Sites | Typed target/effect evidence | Remaining classification reason |
| --- | --- | --- |
| `0x66a`, `0xae9`, `0x5673`, `0x574d` | exact CODE2 entry `0xaed0`; DS/SS/SP preserved, ES becomes 0xa000; conditional local extent 10 plus far frame 4 | pointer contents/explicit aliases, incoming SS:SP, nested table admission remain UNKNOWN |
| `0x3a01`, `0x3d13`, `0x5640` | existing external/API binding boundary | admitted binding/handler effects on segments, stack, object memory and placement remain UNKNOWN; no new binding analysis |
| `0x47a7` | existing TABLE1 callback candidates retained | initialization/admission and possible writer/callee effects remain UNKNOWN |
| `0x5d0c` | existing sole TABLE1 candidate has a preserving local return summary | initialization/admission still UNKNOWN |
| `0x5eab` | existing TABLE1 task candidates retained | slot/current-slot admission and callee effects remain UNKNOWN |
| `0x6095`, `0x6151` | existing TABLE1 area candidates retained | DS/indexed aliases, table admission and callee effects remain UNKNOWN |

For every boundary JSON records DS/ES/SS/SP, explicit memory comparisons,
allocation/placement, return behavior, target identities, completeness and
unresolved reasons. Candidate-only effects are labeled independently from
runtime completeness. No external ABI or observed register value grants
preservation. No full gameplay, scheduler, persistence or INTRO analysis is
performed.

### Fresh writer evaluation and final scoped statuses

Counts are possible **explicit operands**, unresolved implicit stack sites and
unresolved transfer boundaries for each object. No new explicit exclusion is
granted: the open incoming contexts do not justify refining DS/ES/index aliases.
Existing placement reduction and target candidates are preserved.

| Object (inclusive file bytes) | TABLE1 explicit before → after | TABLE1 stack before → after | TABLE1 transfers before → after | CODE2 explicit before → after | CODE2 stack before → after | CODE2 transfers before → after |
| --- | --- | --- | --- | --- | --- | --- |
| `0x5669..0x566c` | 1233 → 1233 | 1420 → 1420 | 12 → 12 | 0 → 0 | 13 → 13 | 2 → 2 |
| `0x56a8..0x56ab` | 1233 → 1233 | 1420 → 1420 | 12 → 12 | 0 → 0 | 13 → 13 | 2 → 2 |
| `0x1f8b0..0x1faaf` | 1241 → 1241 | 1420 → 1420 | 12 → 12 | 0 → 0 | 13 → 13 | 2 → 2 |
| `0x1fab0..0x1fcaf` | 1241 → 1241 | 1420 → 1420 | 12 → 12 | 0 → 0 | 13 → 13 | 2 → 2 |

All six scoped sites remain UNKNOWN: TABLE1 `0x66a`, `0xae9`, `0x5673`, `0x574d`;
CODE2 `0xafa6`, `0xaff7`. Global TABLE1 is **30 / 18 BOUNDED / 12 UNKNOWN**;
global CODE2 is **2 / 0 BOUNDED / 2 UNKNOWN**. The default domain command writes
JSON and returns **exit 2**. A single known destination is not an immutability
proof, and local CODE2 preservation is not an upstream stack admission proof.

Exact remaining obligations in this family: derive every admitted caller's
SS:SP or a universally disjoint domain; close the effects of actually admitted
INT/API and callback/root contexts without inventing ABI assumptions; prove
TABLE1 explicit writer/index/segment aliases disjoint from each target object;
and prove all transfer target contents, object memory and allocation/placement
effects complete. These gaps cannot be discharged by the conditional local
10-byte CODE2 bound. Unrelated DMO0 blockers remain INT66 binding/admission,
PRINTTASK/KEYTASK initialization, task-list and area aliases, INTRO domains,
SDR cadence/IRQ contracts, ElectronicsCalculation events, persistence, feasible
state/control and expiry interleaving, and final consumed-difference closure.
**DMO0 NOT CLOSED. DMO1 NOT STARTED.**

### Previous stack pass: verification and retained worktree

- Complete private demo Python suite: **73 tests PASS**, no SKIP. Includes all
  existing fixed-point/domain, writer/alias, relocation/placement tests, 20 new
  authored stack/transfer mutations and four new private obligation tests.
  The mutations cover SS/SP overlap, increased acyclic depth, near/far widths,
  PUSHF/POPF, segment PUSH/POP, unmatched adjustments, recursion, preserving /
  bounded-disjoint / unknown SS transfers, memory effects, an added admitted
  CODE2 callee, address independence, legal loader bases and physical wrap.
- Focused A/B/C/D Go regression: **PASS**, no private SKIP; matrix 1024, C/D
  seeds, Stones mappings, B/C/D priorities, shared construction, production
  demo rejection and hybrids. Log: `/private/tmp/pf-dmo0-stack-go-tests.log`.
- No-private Python: **35 public tests PASS; two private classes clean SKIP**.
  No-private Go: public descriptor checks PASS; private cases clean SKIP.
- Payload scan: PASS against 249 private files, eleven decoded SDR modules and
  48,898 distinct aligned nontrivial 4 KiB samples plus whole-file identities.
  Scope is all changed research files; it is a sampled copy check, not an
  all-substring proof. Metadata: `/private/tmp/pf-dmo0-stack-payload-scan.json`.
- `git diff --check`: PASS. No change under `internal`, `hosts` or `cmd` relative
  to starting research HEAD. Production/runtime source remains unchanged.
  Existing `.DS_Store` metadata remains identical; never staged or deleted.
- Files changed by this pass: this report, `tools/audit_10min_demo_control.py`,
  `tools/audit_10min_demo_writers.py`, `tools/test_audit_10min_demo_domains.py`,
  new `tools/audit_10min_demo_stack.py` and
  `tools/test_audit_10min_demo_stack.py`. Earlier uncommitted domain/placement
  passes remain in the worktree. No commit, push, tag or release.

## Previous pass: DOS EXEC placement and stack effects

DMO0 NOT CLOSED. DMO1 NOT STARTED. All changes are uncommitted research.
Current metadata: `/private/tmp/pf-dmo0-placement-domains.json`;
`tools/audit_10min_demo_placement.py` supplies the symbolic certificate consumed
by `tools/audit_10min_demo_writers.py`. Production/runtime sources are unchanged.
No scoped indirect site is promoted. The default global command returns **2**:
TABLE1 **30 / 18 bounded / 12 UNKNOWN**, CODE2 **2 / 0 bounded / 2 UNKNOWN**.
The already established candidate-target, DS/ES preservation and instruction
boundary proofs are retained; this slice adds no new candidate enumeration.

### Loader certificate and its admission boundary

The real launcher has normal DOS EXEC `AX=0x4b00` at file `0x6c3` on its
TABLE-loading path. Its separate MZ header has 32 header paragraphs, 1551 image
bytes, 128 reserved image paragraphs, minimum extra 1, maximum extra `0xffff`,
initial CS:IP `(L+0x3b):0`, and SS:SP `L:0x200`. Startup resize at
`0x5c8` uses the relocated endpoint L+`0x61`, subtracts PSP via NEG BX,
and adds `0x40`: requested PSP block `0xb1` paragraphs, retained image end
L+`0xa1`. Failure does not relocate the parent or child. Its normal-EXEC child receives
its own MZ CS/SS, not the parent's stack. Parent stack restoration is visible
at `0x66c..0x676` and `0x6c5..0x6cf` (saved CS words, POP SS, MOV SP).
The tool records the normal-EXEC site and parent POP SS / MOV SP sites; these parent
restores are not child SS assignments.

Loader semantics are grounded in Microsoft's [EXEC source](https://github.com/microsoft/MS-DOS/blob/main/v4.0/src/DOS/EXEC.ASM)
(`exec_save_start`, `exec_allocate`, `exec_use_ax`, `exec_do_reloc`,
`exec_reloc_one`) and [arena allocator](https://github.com/microsoft/MS-DOS/blob/main/v4.0/src/DOS/ALLOC.ASM)
(`alloc_test`, `alloc_get_size`, `alloc_set_owner`, `$SETBLOCK`). Successful
normal EXEC reserves the header-declared image pages and minimum extra space
inside one allocated block, starts the image 16 paragraphs after the PSP,
relocates CS/SS, and adds the relocation factor to each listed word before entry.
Allocation splits and resizing preserve block position. The admitted arena is
valid and nonwrapping; its endpoint is a segment boundary no higher than 1 MiB.
This certificate does not admit an overlay/custom loader or corrupted arena.
The actual DOS version, free-list endpoints and allocation strategy are not
pinned; the analysis uses their broad valid-arena domain. It assumes neither a
640-KiB ceiling nor an allocation below `0xa000`.

TABLE1's actual header mechanically yields:

| Field | Derived value |
| --- | --- |
| Header | 32 paragraphs / `0x200` bytes |
| Declared file | 1050 pages, last page 102 bytes / 537190 bytes |
| Complete file-backed load module | `0x83066` bytes / 536678 bytes |
| Rounded actual image | `0x8307` paragraphs |
| DOS page-reserved image | `1050*32 - 32 = 0x8320` paragraphs |
| Minimum extra / maximum extra | `0` / `0xffff` paragraphs |
| Minimum PSP block | `0x10 + 0x8320 = 0x8330` paragraphs |
| Initial CS:IP | `(B+0x10):0x2f9f`, file entry `0x329f` |
| Initial SS:SP | `B:0x100` |
| Relocations | 104 unique bounded words; add B modulo 65536 once before entry |

Let P be the PSP, B the TABLE1 load segment and E the valid arena endpoint.
`P=B-0x10`, `P+allocated_paragraphs <= E <= 0x10000`, and
`allocated_paragraphs >= 0x8330` imply **`0x10 <= B <= 0x7ce0`**.
Every integer paragraph in this interval is conservatively admitted; no fixed
load segment is invented. Relocated immediates and CS retain typed image
identity; unrelocated numeric segment literals remain absolute. Header SS/CS
are separately classified as load-base-relative, and entry DS/ES originate at
PSP under normal EXEC. Entry contexts from other roots remain UNKNOWN.

### Image extents and lifetime

All intervals are half-open. Physical addresses use `16*B + relative_offset`.
The following physical envelopes cover every B in the admitted domain. They
are conservative envelopes, not claims that every byte between endpoints can
belong to one particular placement.

| Typed object/image | Relative interval | Physical envelope |
| --- | --- | --- |
| TABLE1 complete initial load image | `[0, 0x83066)` | `[0x100, 0xffe66)` |
| TABLE1 code slice | `[0x100, 0xacd0)` | `[0x200, 0x87ad0)` |
| CODE2 linked slice | `[0xacd0, 0x19bb0)` | `[0xadd0, 0x969b0)` |
| CODE2's demo data segment addressable window | `[0x19bb0, 0x29bb0)` | `[0x19cb0, 0xa69b0)` |
| Post-CODE2 image through startup LASTSEG | `[0x19bb0, 0x523a0)` | `[0x19cb0, 0xcf1a0)` |
| Pointer file `0x5669..0x566c` | `[0x5469, 0x546d)` | `[0x5569, 0x8226d)` |
| Pointer file `0x56a8..0x56ab` | `[0x54a8, 0x54ac)` | `[0x55a8, 0x822ac)` |
| Lookup file `0x1f8b0..0x1faaf` | `[0x1f6b0, 0x1f8b0)` | `[0x1f7b0, 0x9c6b0)` |
| Lookup file `0x1fab0..0x1fcaf` | `[0x1f8b0, 0x1fab0)` | `[0x1f9b0, 0x9c8b0)` |

CODE2 is part of TABLE1's MZ image, not a separately DOS-allocated module:
its far-pointer relocation-derived base remains file `0xaed0` / B+`0xacd`.
The actual relocated startup MOV AX/DS gives data B+`0x19bb` / file `0x19db0`.
The code-slice endpoints retain the preceding typed module identities. The
post-CODE2 range is a containing linked-image region, not a claim that all of
it is the first 64-KiB data segment.

Startup `0x329f..0x32ab` derives BP=PSP, relocated BX=B+`0x523a`, subtracts
BP, adds `0x40`, and calls DOS resize. Requested PSP block is `0x528a`
paragraphs; successful retention ends at B+`0x527a`. The potentially released
tail is relative `[0x527a0,0x83200)`. Failure CF is not checked, so a failed
shrink retains the larger initial block. Both cases preserve all four target
addresses and the initial-base upper bound. A freed tail can later be reused;
no exclusion treats that tail as permanently occupied. A later allocation
does not move the original linked image, but unknown subsequent free/resize,
resident/module and external effects remain lifecycle obligations. No separate
resident allocation is certified or used to exclude a writer. The report does
not claim that all data or the entire initial image is disjoint from `0xa000`;
the four target objects are the proved disjoint regions.

### Absolute-segment exclusions and auditable evidence

Every known absolute literal is classified as `absolute-real-mode-segment`,
without video/system names. For `0xa000` an unknown EA covers all 65536 starts:
physical destination `[0xa0000,0xb0000)`. All four target envelopes end below
`0xa0000`, including the highest lookup endpoint `0x9c8b0`. Hence every possible
byte of every admitted store is disjoint, independently of index uncertainty.
**1741 TABLE1 and 1082 CODE2 operands** are excluded by this universal
certificate. Another **8 TABLE1 operands** with segment zero and bounded low
IVT offsets are disjoint from all four targets' lower endpoints; their actual
offsets, definitions and ranges are in the per-operand metadata. Segment zero
with unknown offsets is not generically excluded.

For example CODE2 store `0xb09f` has DS literal `0xa000`, definitions
`0xaf5e` / `0xaf61`; its BX memory producer at `0xaf76` is unresolved. Its full
segment destination nevertheless cannot meet any target. Every writer record
retains source/operand/width, segment producer sites, literal versus relocated
identity, EA displacement/base/index/scale, index definitions, REP-count and DF
provenance, possible wrapped offset ranges, both A20-state destination domains,
target domains, B domain and exclusion reason. Unknown provenance never earns
an exclusion. Independent envelope overlap remains UNKNOWN; image-relative
comparisons retain shared B correlation only when segment/physical wrap cannot
invalidate it. Uncertified/out-of-image targets fail closed.

Both 16-bit offset wrap and physical 20-bit wrap are modeled. A20-enabled
linear addresses are checked too, so a wrap-dependent low-memory alias is not
removed by choosing an unproved hardware state. A changed initial allocation
requirement or enlarged B domain can change the same `0xa000` store from
EXCLUDED to UNKNOWN. Numeric segment appearance alone has no effect.

### Stack and transfer effects

Initial TABLE1 SS is B, SP `0x100`. Decrement-before-store PUSH/CALL/INT effects
wrap SP modulo 65536. The strongest depth-independent initial-SS interval is
`[0,0x10000)`, not the linker-declared 256-byte stack. That interval can overlap
both pointer objects; it is disjoint from both lookup tables only while SS
remains the initial segment. No maximum dynamic depth, interrupt nesting or
handler stack interval is proved. No direct SS/SP reassignment is found in the
admitted TABLE1/CODE2 CFGs, but this inventory does not cover external handlers
or incoming additional roots. All **1420 TABLE1 / 13 CODE2 implicit stack
sites remain UNKNOWN**. Initial-only table disjointness is emitted separately
and does not discard these sites. TABLE1's **70 explicit INT sites** also have
UNKNOWN handler register, stack and allocation effects; asynchronous interrupt
admission/depth is open. Interrupt writers are not silently omitted.

The **12 TABLE1 / 2 CODE2** unresolved transfer boundaries are inventoried by
source in `placement_effects.boundaries`. DS/ES/SS/SP/allocation/placement are
UNKNOWN at the TABLE1 boundaries. Existing CODE2 candidate DS/ES preservation
is retained as a conditional summary; target contents/admission, SS/SP and
allocation/lifecycle are still UNKNOWN. A target list does not establish a
transfer-effect contract. INT66 binding, PRINTTASK/KEYTASK, task/area semantics,
SDR timing, persistence, gameplay, expiry and DMO1 were not expanded.

### Before/after and exact remaining obligations

| Object | TABLE1 before | TABLE1 after | CODE2 before | CODE2 after |
| --- | ---: | ---: | ---: | ---: |
| pointer `0x5669..0x566c` | 2982 | 1233 | 1082 | 0 |
| pointer `0x56a8..0x56ab` | 2982 | 1233 | 1082 | 0 |
| lookup `0x1f8b0..0x1faaf` | 2990 | 1241 | 1082 | 0 |
| lookup `0x1fab0..0x1fcaf` | 2990 | 1241 | 1082 | 0 |

Total excluded-from-all operands: TABLE1 **401 -> 2150** out of 3392;
CODE2 **5 -> 1087** out of 1087. Implicit stack UNKNOWN counts **1420 -> 1420**
and **13 -> 13**; unresolved transfer counts **12 -> 12** and **2 -> 2**.
These count operands, not unique source addresses or actual writes. There are
no missing operand-access records. Every explicit CODE2 operand is disjoint
from the four static image objects under the placement certificate; this is
not complete CODE2 runtime immutability.

All six scoped indirect sites remain UNKNOWN. Exact completeness obligations:
TABLE1's remaining DS/ES/indexed/REP alias ranges and incoming segment contexts;
SS provenance across all admitted entry/transfer/interrupt paths; either bounded
stack effects or universal stack/object separation on each path; transfer
allocation/free/resize/placement effects and retained-object lifecycle; and
upstream far-pointer/table contents admission. No family promotion is justified
until all those obligations are discharged.

### Current verification

- Complete private demo Python suite: **49 tests PASS**, including existing
  writer/alias/relocation/fixed-point/domain mutations, 15 public placement
  mutations and a real-input placement-integration check.
- Focused A/B/C/D Go suite in datalayout/frontend/Stones: **PASS, no SKIP**,
  including matrix 1024, C/D seeds, Stones mappings, priorities and production
  demo rejection. An initial run pointed D at a C fixture and failed identity
  checks; after live fingerprint verification, corrected fixtures passed.
- No-private Python: **15 public tests PASS, 2 private classes clean SKIP**
  using system Python without Capstone. No-private Go: private cases SKIP,
  public descriptor checks PASS.
- Default global gate: **exit 2**, unchanged 18/12 TABLE1 and 0/2 CODE2.
- Changed-file payload scan: **PASS** against 249 private files, 11 in-memory
  decoded SDR modules and 48,898 nontrivial aligned 4-KiB samples, plus whole-file
  identity checks. This is a sampled copy check, not an all-substring proof.
  Metadata: `/private/tmp/pf-dmo0-placement-payload-scan.json`.
- `git diff --check`: **PASS**. Diff against the starting HEAD has no paths under
  `internal`, `hosts` or `cmd`; all changed/untracked research files are Python
  or Markdown (excluding the pre-existing, untouched `.DS_Store`).
- No production/runtime changes, commit, push, tag or release. `.DS_Store`
  remains untouched and unstaged.

## Previous slice: relocated-object / recursive CODE2 writer pass

Current machine-readable evidence: `/private/tmp/pf-dmo0-writer-domains.json`.
The current global command still returns **2**. TABLE1 remains **30 sites,
18 bounded / 12 UNKNOWN**; CODE2 remains **2 sites, 0 bounded / 2 UNKNOWN**.
DMO0 NOT CLOSED. DMO1 NOT STARTED.

The four scoped TABLE1 consumers retain their operand-derived single candidate:
CODE2 image segment `0xacd`, offset `0`, typed entry file `0xaed0`.
The objects remain file `0x5669` and `0x56a8`; the destination is not substituted
by a consumer-address allowlist. Each has exactly its segment word in the MZ
relocation set (`0x566b` / `0x56aa`). The offset word has no loader relocation.
Additional overlapping or duplicate relocation producers are rejected.

The lifecycle evidence distinguishes three different questions:

- Loader: the MZ loader adds its load segment to the relocated segment word.
  This is established from the relocation records, not inferred runtime mutation.
- Executable initialization: stores are inventoried, but their ordering before
  consumer admission is not proved. No claim of absent initialization is made.
- Runtime: writable aliases and transfer effects are not closed, so immutability
  of either offset word or segment word remains **unproved**.

`tools/audit_10min_demo_writers.py` inventories every memory operand with write
access in every instruction of the admitted TABLE1/CODE2 CFGs. It includes
read/modify/write instructions, XCHG, POP-to-memory, byte/word stores and
MOVS/STOS/INS destinations. Unknown access metadata blocks coverage; OUTS is
explicitly classified as a memory read. Implicit stack writers and unresolved
transfer effects are retained as separate obligations. `complete_inventory`
means known memory-operand stores were enumerated, **not** that writer admission,
alias coverage or whole-image runtime immutability is complete.

The backward alias worklist tracks DS and ES independently, follows admitted
caller entry edges, requires preservation across calls, and distinguishes
relocated image-segment immediates from absolute segment literals. A CS source
retains its owning module identity. Unknown aliases cannot be discarded.
Effective-address overlap includes store width and 16-bit wrap. Unsupported
index arithmetic expands conservatively to the full segment. Bounded REP
counts cover both DF directions; unknown counts/recurrences cover the segment.
The worklist bound fails closed rather than dropping paths.

| Admitted module | Instructions | Explicit writer operands | Excluded from all four objects | Implicit stack writer sites | Unsummarized transfer boundaries |
| --- | ---: | ---: | ---: | ---: | ---: |
| TABLE1 | 14,985 | 3,392 | 401 | 1,420 | 12 |
| CODE2 | 1,723 | 1,087 | 5 | 13 | 2 |

| Object | TABLE1 possible writer operands | CODE2 possible writer operands | Result |
| --- | ---: | ---: | --- |
| pointer `0x5669..0x566c` | 2,982 | 1,082 | UNKNOWN |
| pointer `0x56a8..0x56ab` | 2,982 | 1,082 | UNKNOWN |
| table `0x1f8b0..0x1faaf` | 2,990 | 1,082 | UNKNOWN |
| table `0x1fab0..0x1fcaf` | 2,990 | 1,082 | UNKNOWN |

These are conservative possible-overlap records, **not actual mutation counts**.
Concrete remaining analysis gaps include TABLE1 `0x5c39` (CS:[DI], with DI
arithmetic/recurrence not bounded by this analyzer) and CODE2 glyph stores such
as `0xb09f` (DS is mechanically `0xa000`, BX comes from memory). The analyzer
has no loader-placement contract that compares that absolute segment against
the relocated image, so it cannot exclude it on an assumed conventional DOS
placement. Incoming TABLE1 segment contexts and implicit SS:SP ranges also
remain unresolved. In contrast, TABLE1 REP STOSW `0x5687` is excluded from all
four objects: ES is the relocated demo data segment, DI and CX are immediate,
and either DF direction stays outside the objects. This is a derived range
exclusion, not an exact-address MOV scan.

Both lookup-table candidate sets retain their original mask/shift/load/addend
recipes below: **256 raw entries / 255 distinct targets each**. Recursion still
converges in two rounds, with 511 roots and 1,723 instructions. All 510 admitted
candidate entries have decoded, nonoverlapping instruction boundaries and
closed direct control slices reaching unadjusted near returns. The typed roots
come only from these table recipes and the relocated module entry. This proves
candidate validity under the table-contents premise; it does not prove that
runtime writes cannot change those contents.

For every one of the 255 first-stage candidate callees, separate mechanical
summaries prove **ES preserved and DS preserved**. The same result holds for
all 255 second-stage candidates. PUSHA/POPA handles only general registers;
segment saves/restores are tracked separately. The second consumer's ES
reaching definitions use these first-stage summaries. The new clobber mutation
exposed and fixed a previous research bug: extended-register normalization
incorrectly stripped the initial `e` from `es`. Only actual extended general
register names are now normalized. MOV ES,AX is correctly a clobber; explicit
PUSH ES / POP ES restores it; PUSHA/POPA alone does not.

Mutation evidence includes byte writes into the offset word, word writes into
the segment word, a word starting before the object, direct DS/ES aliases,
indexed writes, table writes, REP writes, unknown aliases and an extra admitted
writer. An authored writer-free typed CFG is BOUNDED by the inventory; injecting
an overlapping writer makes it UNKNOWN. Private-input ES-clobber injection into
an admitted first-stage candidate makes its summary fail and the second-site
ES dependency UNKNOWN. Missing/overlapping relocation producers and ambiguous
table-base grammar fail closed. These tests verify rejection and evidence;
they do **not** establish a passing real-object closure test. A real-table write
and far-pointer write cannot be claimed to revert a previously closed real
family, because that family has not been closed.

The eight unrelated TABLE1 UNKNOWN sites retain their previous classification:
`0x3a01`, `0x3d13`, `0x5640`, `0x47a7`, `0x5d0c`, `0x5eab`, `0x6095`, `0x6151`.
The scoped four far consumers are **also still UNKNOWN**. No INT66, callback,
task, area, SDR timing, persistence, gameplay or expiry proof was broadened.

Current verification:

- Full private demo Python suite: 33 tests PASS, including prior fixed-point and
  mutation tests plus writer/range, relocation lifecycle, target validity,
  address-independent derivation and ES-clobber checks.
- Focused A/B/C/D Go regression including Stones: PASS, no private SKIP.
  Includes matrix 1024, C/D seeds, Stones mappings, B/C/D priorities and
  production demo rejection.
- No-private Python: clean SKIP of both private classes before Capstone import.
  No-private Go: private cases SKIP; public descriptor tests PASS.
- Default global domain command: exit 2; the overall gate is retained.
- Changed-file payload scan and `git diff --check`: PASS. No production runtime
  source paths changed against the starting HEAD. No commit, push, tag or
  release. `.DS_Store` is untouched and unstaged.

## Previous-pass evidence retained

## TABLE1 fixed point

```text
TABLE1 indirect sites: 30
bounded: 18
unknown: 12
```

The candidate fixed point has 7 rounds, 698 TABLE1 roots and 14,985 TABLE1 instructions. The previous 677-root / 14,197-instruction fixed point missed a fourth area-table source. Adding it grows both the area and downstream task candidate sets. No additional TABLE1 indirect site appeared.

## The twelve sites

| Site (file offset) | Operand / category | Demo-derived candidates | Unresolved completeness obligation |
| --- | --- | --- | --- |
| `0x66a` | LCALL ES memory / relocated module pointer | CODE2 image segment `0xacd`, offset `0`, file `0xaed0` | CS pointer immutability requires complete ES/DS/indexed writer aliases |
| `0xae9` | LCALL ES memory / relocated module pointer | CODE2 image segment `0xacd`, offset `0`, file `0xaed0` | CS pointer immutability requires complete ES/DS/indexed writer aliases |
| `0x5640` | LCALL ES memory / external API binding | INT66 vector `0000:0198`; resident SDR CS entries in the table below | vector installer/admitted consumer ordering and writer closure not proved |
| `0x5673` | LCALL ES memory / relocated module pointer | CODE2 image segment `0xacd`, offset `0`, file `0xaed0` | CS pointer immutability requires complete ES/DS/indexed writer aliases |
| `0x574d` | LCALL ES memory / relocated module pointer | CODE2 image segment `0xacd`, offset `0`, file `0xaed0` | CS pointer immutability requires complete ES/DS/indexed writer aliases |
| `0x3a01` | LCALL ES memory / external API binding | INT66 vector `0000:0198`; resident SDR CS entries in the table below | vector installer/admitted consumer ordering and writer closure not proved |
| `0x3d13` | LCALL ES memory / external API binding | INT66 vector `0000:0198`; resident SDR CS entries in the table below | vector installer/admitted consumer ordering and writer closure not proved |
| `0x47a7` | CALL / DS callback field writers plus initial value | `0x48be`, `0x4900`, `0x4921`, `0x6d71`, `0x701a` | initial/derived non-code target requires lifecycle proof |
| `0x5d0c` | CALL / DS callback field writers plus initial value | `0x6d71` | initial/derived non-code target requires lifecycle proof |
| `0x5eab` | CALL / task-list consumer; not a writer | `0x5ba`, `0x5fc`, `0xe77`, `0xebb`, `0xfa8`, `0xfce`, `0xff4`, `0x1408`, `0x1479`, `0x14a3`, `0x151b`, `0x152c`, `0x169b`, `0x16fa`, `0x171d`, `0x177a`, `0x179d`, `0x17fa`, `0x181d`, `0x194f`, `0x1b84`, `0x1c04`, `0x1c44`, `0x1ca5`, `0x1db3`, `0x1dcb`, `0x1e08`, `0x1f13`, `0x1f3a`, `0x20a6`, `0x20fe`, `0x212b`, `0x2166`, `0x2199`, `0x21cc`, `0x220e`, `0x2236`, `0x23e5`, `0x2415`, `0x243b`, `0x2493`, `0x254a`, `0x272e`, `0x2750`, `0x2864`, `0x2993`, `0x2a13`, `0x2a70`, `0x2acd`, `0x2b2a`, `0x2bbf`, `0x3244`, `0x3259`, `0x383c`, `0x655e`, `0x6575`, `0x6d71` | indexed task-list writer domain not fully closed |
| `0x6095` | CALL / CFG rectangle loop: four bounds, handler, zero sentinel | `0x1650`, `0x16b2`, `0x1732`, `0x17b2` | DS entry/callee alias and table writer completeness not proved |
| `0x6151` | CALL / CFG rectangle loop: four bounds, handler, zero sentinel | `0x19ba`, `0x1bb3`, `0x1d12`, `0x1d70`, `0x1e2d`, `0x1fa0`, `0x1fde`, `0x2296`, `0x2561`, `0x2568`, `0x259d`, `0x27c0`, `0x27db`, `0x27f6`, `0x287b`, `0x28c6`, `0x28d9`, `0x28e0`, `0x29ce`, `0x29d2`, `0x2a2f`, `0x2a8c`, `0x2ae9`, `0x2ceb`, `0x2cec`, `0x2ced` | DS entry/callee alias and table writer completeness not proved |

All near targets in that table are TABLE1 file offsets; subtract `0x300` for TABLE1 CS offsets. `PRINTTASK` and `KEYTASK` also have initial zero words, which remain non-code values requiring an initializer-before-consumer admission proof. The report retains zero explicitly in `field_evidence`; it is not silently removed as harmless.

`0x5eab` is **CALL word ptr [BX]**, not a write. The reachable task insertion store is `0x5e9c` (MOV [BX],DX); the reset loop store is `0x3b4b` (MOV [BX],AX). Writer evidence enumerates their candidate base definitions and other [BX] stores without claiming the inventory is complete. All 57 task candidates remain fed back into the TABLE1 worklist; slot/current-slot aliases and initialization/admission remain open.

## Corrected area grammar

The area decoder derives a ten-byte row from five LODSW operations, the zero-sentinel branch, four bound comparisons and four reject branches that restore the saved SI before advancing by ten. It removes only this verified recurrence edge from the base reaching-definition query. It does not use the consumer address as a source allowlist.

`0x6151` has four base definitions: `0x60f7` (SI=`0xe45`), `0x6104` (SI=`0xdb7`), `0x6111` (SI=`0xedf`) and `0x611e` (SI=`0xebf`). The formerly omitted `0xe45` table contributes ten distinct additional targets. `0x6095` retains the SI=`0xd8d` table from `0x6064`. These finite table candidates do not establish DS entry/callee aliases or table-write completeness.

## Far construction and CODE2 recursion

CFG predecessor slices derive ES=TABLE1 CS at `0x66a`, `0xae9`, `0x5673` and `0x574d`. The operands name literal far pointers at file `0x56a8` or `0x5669`; their segment words have actual MZ relocation entries. Both point to CODE2 offset zero. The module is inspected from that typed entry, rather than by interpreting every syntactically valid data word as code.

`0x5640` belongs to the external API domain, despite the original five-site grouping. Its actual ES producer sets ES=0, and its pointer operand is `0x198`.

CODE2 candidate recursion converges in 2 rounds: 511 roots, 1,723 instructions, 2 indirect sites, **0 bounded / 2 UNKNOWN**. Both sites are CALL DX:

| CODE2 site (file offset) | ES word table (file offset) | Exact candidate set definition | Count |
| --- | --- | --- | --- |
| `0xafa6` | `0x1f8b0` | `{0xaed0 + ((u16(table + 2*i) + 0x1a0) & 0xffff) : 0 <= i < 256}` | 255 |
| `0xaff7` | `0x1fab0` | `{0xaed0 + ((u16(table + 2*i) + 0xc40) & 0xffff) : 0 <= i < 256}` | 255 |

The JSON enumerates both complete candidate sets. Mask/shift/load/register-transfer/addend grammar derives them from demo tables. Every candidate is recursively included, and there are no additional indirect sites in this candidate CODE2 CFG. Conditional ES alias evidence resolves to the demo data image segment `0x19bb`; the second site depends on the first candidate glyph calls preserving ES. Segment-save summaries track DS/ES separately from PUSHA/POPA. Table immutability and the upstream far-pointer writer domains remain unproved, so recursive candidate discovery is not reported as closure.

## External installer domain

All eleven supplied SDR identities are checked and their EXEPACK bodies decoded in memory. From each actual entry path the tool verifies ES=0, a literal AX offset store to `0x198`, and a CS segment store to `0x19a`. These are actual binding producers, not an observed-value allowlist:

| Module | Resident CS offset |
| --- | --- |
| ADLIB.SDR | `0x49` |
| GUS.SDR | `0xf7` |
| INTERNAL.SDR | `0x4e` |
| NOSOUND.SDR | `0x40` |
| PAS16.SDR | `0x49` |
| SB16.SDR | `0x49` |
| SB20.SDR | `0x49` |
| SBLASTER.SDR | `0x49` |
| SBPRO.SDR | `0x49` |
| SM2.SDR | `0x49` |
| THING.SDR | `0x49` |

For `0x3a01`, `0x3d13` and `0x5640`, the exact admitted resident-module binding is still UNKNOWN: complete vector writers and installer-before-consumer ordering have not been proved. Discovering all supplied installers does not establish that no unmodeled binding reaches a call. No SDR scheduler/cadence/IRQ semantics were expanded in this pass.

## Previous-pass verification

- Private 10-minute-demo Python suite: 24 tests PASS. This includes prior fixed-point/bounded-target tests and new candidate-set, relocation, installer-source, area-base/recurrence, address-independent glyph and segment-stack mutations. It does **not** include a passing exact twelve-site closure test, because closure is absent.
- Focused existing A/B/C/D Go regression: PASS without private-test SKIP. Includes the 1024 coherence matrix, C/D seeds, Stones mappings, B/C/D priorities, shared construction and production demo rejection/hybrids.
- No-private-fixture Python path: clean SKIP of two private classes before Capstone import. No-private-fixture Go path: private tests SKIP, public descriptor checks PASS.
- Default domain command: exit 2. `--allow-open`: exit 0 for inspection only. The default gate was strengthened to inspect CODE2 statuses and actual TABLE1 statuses as well as the counters; the existing semantic-scope guard is retained.
- Changed-file payload scan: PASS against 249 private PRG/MOD/EXE/SDR/BIN files, eleven in-memory decoded SDR modules and 48,898 distinct nontrivial aligned 4 KiB samples, plus whole-file identity checks. This is a sampled copy check, not an all-substring proof. Metadata: `/private/tmp/pf-dmo0-control-payload-scan.json`.
- `git diff --check`: PASS. Diff against starting research HEAD contains no changes under `internal`, `hosts` or `cmd`. Diff against production under those paths contains only the pre-existing research `internal/frontend/demo_audit_test.go`; production runtime source is unchanged.

## Remaining work

The relocated-object slice still needs complete DS/ES/indexed write aliases and admission/placement/stack contracts for the far-pointer objects and glyph tables. The unrelated prior scopes still need area-table aliases; complete INT66 binding/admission; PRINTTASK/KEYTASK initialization-before-call; and task slot/current-slot/reset writer closure. Finite candidates are retained but cannot discharge these obligations.

Other DMO0 blockers remain separate: INTRO indirect domains, SDR contracts and ElectronicsCalculation event proof, persistence/state lifecycle, gameplay/state-feasible and expiry interleaving proof, and the final consumed-difference gate. DMO1 has not started. Production/runtime source, release assets and v0.1.3 are unchanged. No push, tag or release. The untracked `.DS_Store` was not modified, staged or deleted.

Changed research files: `tools/audit_10min_demo_writers.py`, `tools/audit_10min_demo_control.py`, `tools/audit_10min_demo_domains.py`, `tools/test_audit_10min_demo_domains.py`, and this report. All worktree changes are left uncommitted by instruction.
