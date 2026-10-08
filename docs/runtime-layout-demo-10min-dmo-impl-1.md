# DMO-IMPL-1: test-only bounded semantic core

2026-10-08. Owner-approved implementation scope:
[runtime-layout-demo-10min-dmo0-scope-decision.md](runtime-layout-demo-10min-dmo0-scope-decision.md).
Production base `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`;
research HEAD `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`.

**Delivered: staged internal harness, not an interactive demo.**
DMO0 NOT CLOSED. DMO1 NOT STARTED. No support registration, packaging,
README download, release/version change, commit, push, tag or research replay.

## Files and production interface

- `internal/partyland/game.go`: extracts the existing UPDATE_COUNTERS prefix
  into unexported `(*Game).updateCounters()`. `beforeTargets` still runs that
  exact prefix followed by `checkAreas`, in the same order. No physics edit.
- `internal/partyland/demo_core_test.go`: isolated table-lifetime state,
  staged admitted-calculation entry, typed expiry stream, guarded task/matrix
  integration and sticky diagnostic failure.
- `internal/partyland/demo_core_checks_test.go`: unit/integration checks,
  including an opt-in pinned-private-input candidate entry.
- This report. Pre-existing research edits/untracked files are outside this change.

Both new Go files require `dmoimpl1` **and** are `_test.go` files. They are
excluded from application builds even if that tag is supplied. No exported
production API, Game demo flag, profile, selector entry or host hook was added.
No oracle or A/B/C/D descriptor changed. Two ordinary build artifacts were
also checked for absence of the prototype diagnostic string.

## Executed behavior and limits

A `newDemoCore` is a fresh TABLE1 lifetime: uint16 counter zero, expired false,
source HOLDSTILL false. Each admitted `calculation(false, budget)` runs the
shared UPDATE_COUNTERS prefix then exactly one wrapping increment, including
an explicitly staged BallLost suffix. Pause admits no work. Ordinary NEW_BALL
ownership is tested against the actual canonical `newBall` reset with private A
inputs; table-lifetime state remains separate. This does not implement demo
NEW_BALL or progression.

Equality is exactly 35998 after increment, with no expired guard. It sets the
sticky expired/source HOLDSTILL flags, requests `(13,0,255)` through shared
`MusicClock.Play`, and installs a typed expiry program before the task scan.
The whole uint16 interval through wrap and second equality is actually executed
in tests with matrix budget disabled. This is a timer test, not a fresh gameplay
reachability witness or eventual-QUIT proof.

The staged order is early/drain observation -> shared UPDATE_COUNTERS -> timer
-> empty electronics -> empty KEYTASK -> shared DO_TASKS -> budgeted matrix
visit -> empty late-physics stage. **Early physics/drain observation and the
empty stages are harness premises, not executed physics/gameplay consumers.**
In particular post-drain accounting is exercised on a supplied BallLost phase;
no real drain-to-demo electronics adapter has been delivered.

Task integration reuses the existing 50-slot first-free allocator, ascending
identity-safe scan and per-callsite shared compare-before-increment WAITLIST.
Tests cover same-scan children above the cursor, next-scan children below it,
wait reset, task-list reset, preservation and replacement of matrix cursors.
Expiry itself never purges tasks. Later explicit task-list resets remain possible.
Synthetic typed task actions exercise preserve, source-hold release, matrix
replacement and task/wait reset. They are primitive tests, **not implementations
of NEW_BALL_TASK, PARTY_ON_TASK1, SETBALL, DROPTASK2 or DURINGFLASH**.

Matrix installation/visits reuse `matrixState`, `matrixDispatch`, `matrixTick`
and `presentation.Display`. CLEAR4 enters with remaining5; a budgeted visit
makes it4. Missed budget preserves the cursor. A same-calculation task replacement
is visited instead of expiry; there is no expiry immunity or saved-cursor resume.
The typed expiry stream records CLEAR4, SCROLL, FLASHON1, score print344,
WAIT100, FLASHOFF1, SCROLL, FADE256, WAIT100 and QUIT0. Only CLEAR4 and explicit
well-formed WAIT primitives are admitted for execution here. Demo text lengths
98/88 are conformance facts, not synthetic assets supplied to the renderer.

The shared jingle priority comparison is tested for lower rejection, equal
admission and subsequent explicit priority lowering. Full demo effect admission
(INH_EFF, SPECIALMODE, cue result, score/bonus side effects) is **not implemented**
and no effect path is exposed by the harness. Source HOLDSTILL is independent
of `Physics.Ball.Hold`; consuming it in real late physics is deferred.

None of the five first-equality gameplay witnesses or their complete QUIT
suffixes is claimed implemented or PASS. They remain future conformance targets.
No wall-clock countdown exists.

## Enforced boundary and integration blockers

The first unsupported requested task action, full task-list allocation,
malformed matrix entry or unimplemented matrix consumer latches
`UNSUPPORTED_DEMO_TRANSITION`, with producer, state/guard and reason.
Subsequent calls return the same error without advancing counter/tasks/matrix.
No canonical progression fallback is available. A replacement with unsupported
entry is rejected instead of running canonical CHANGE_PLAYER.

Before a shared matrix visit could tail-dispatch an unsupported consumer, the
harness checks its next typed command. At expiry CLEAR4 remaining1 / next SCROLL,
it stops **before the final CLEAR4 visit and SCROLL dispatch**, preserving that
cursor. This conservative visit-level boundary is deliberate: it does not claim
completion of the last clear visit. Counter/tasks earlier in the admitted
calculation have already executed; there is no rollback claim. Unknown physics,
areas, bonus, keyboard/restart and reward consumers are outside the staged API;
explicit requests for those paths diagnose unsupported, never run `Game.Sync`.

A reliable interactive boundary is not available in current interfaces:

1. `physics.Sync` returns on drain before BeforeTargets; canonical BallLost
   scheduling is a separate path. A timer-only hook would miss the suffix or
   count at the wrong boundary.
2. BeforeTargets combines electronics counters with area callbacks; physics
   event callbacks and task closures lack an error return for stopping before
   an unsupported consumed edge.
3. Matrix dispatch can immediately tail-call gameplay/progression consumers;
   current production interpreter has no demo FADE/QUIT implementation or
   fallible per-consumer gate. Full effect guards also need typed interception.

Therefore this milestone takes the expressly permitted **test-harness fallback**.
It supplies no interactive candidate, physics adapter or unrestricted equivalence.
The sole production extraction makes the shared counter prefix reviewable
without altering existing callbacks.

## Reproduction and validation

Public fixture-free harness run:

```sh
./tools/go.sh test -tags dmoimpl1 ./internal/partyland -run '^TestDemo' -count=1 -v
```

Private timer candidate: additionally set `PF_10MIN_DEMO_DATA` to the owned
pinbfan directory. It verifies all five existing identity pins through the
existing read-only helper constants, checks fresh timer/threshold/cue inputs,
then executes the staged counter to equality. It neither invokes an audit CLI
nor exports executable/asset bytes or hashes. Missing env skips; supplied wrong
or corrupt data fails. It is not a demo decoder/gameplay conformance test.

Original-backed checks ran in a private temporary source snapshot with links to
owned canonical inputs and historical source; no commercial fixture was placed
in the repository. No missing fixture was manufactured. Logs are owner-local
`/private/tmp/pf-dmo-impl1-*.log`.

| Check | Result |
| --- | --- |
| New harness tests with pinned demo and available A inputs | PASS: all 11 new top-level tests; existing canonical demo-mode matrix test also PASS |
| Fixture-free harness / private timer candidate | PASS; actual canonical NEW_BALL test skips without A |
| Canonical Party Land, source scheduler, presentation suites | PASS with available private inputs; five opt-in capture/export checks skipped |
| Extended physics suite | Available tests PASS; TestStonesSourcePhysicsAndTwoFlippers fails because pf10-assets.json is absent (NOT AVAILABLE fixture) |
| TestOriginalTrajectories | NOT AVAILABLE; explicitly excluded, no fixture generation |
| A/B/C/D detection/factories, descriptor checks, demo/hybrid rejection | PASS: focused datalayout/frontend suite, no skips |
| PF6 gameplay baseline | FAIL: score000002311040, ball2, ticks1200; identical assertion reproduced against unchanged research HEAD in same private snapshot |
| Go vet on Party Land/shared primitives with prototype tag | PASS |
| macOS shared engine and Windows amd64 executable builds | PASS; prototype diagnostic absent in both binaries |
| Broad native-macOS Go compile-only sweep | FAIL in internal/platform: absent macOS AudioDevice/hostWindow/openHost implementation; shared packages and cmd targets otherwise compiled |
| Public source checks | PASS |
| git diff --check | PASS |
| Commercial payload checks | PASS for the four milestone files; aligned 4 KiB block scan against private A/B/C/D/demo runtime files, plus changed-file extension check; sampled block check, not an all-substring proof |

A/B/C/D behavior has no intentional change. Evidence is the unchanged
beforeTargets statement order, unchanged oracle/profiles/physics, canonical
Party Land/reference passes and private A/B/C/D passes. The PF6 failure remains
an explicit baseline failure, not a new regression or a PASS.

## DMO-IMPL-2 work requiring the next scoped milestone

1. Design fallible typed consumption gates for a real admitted calculation,
   including early physics/drain, post-drain electronics, task bodies and late
   HOLDSTILL consumption. Keep A/B/C/D paths unchanged.
2. Map validated private demo presentation operands to the shared display;
   implement the remaining expiry consumers including FADE and native QUIT policy.
3. Implement actual demo reset/new-ball and reviewed zero-bonus continuation,
   explicit demo effect admission and producer/state guards; reject every other
   consumed route before fallback is possible.
4. Implement and replay each supported first-equality witness end-to-end before
   marking it PASS, including matrix replacement, release, repeated equality,
   pause and concrete QUIT continuations. Preserve all unsupported cases.

These are next-milestone tasks, not executed research or implementation here.
