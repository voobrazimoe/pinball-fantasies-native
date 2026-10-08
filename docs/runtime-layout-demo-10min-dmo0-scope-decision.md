# DMO0: scope and implementation decision gate

Дата: 2026-10-08. Official 10-minute demo `pinbfan`.
Production base: `306d11a0c479c7ac5ee6e245f3f72eacbc665abd`.
Committed research HEAD: `37ff8bc38d7af4a09a70319675d6e505b0bd6a5e`;
текущий HEAD совпадает. Последующие research reports/tools находятся в рабочем
дереве. **DMO0 NOT CLOSED. DMO1 NOT STARTED.** Production demo support отсутствует.

**Рекомендация: STOP автоматическому расширению DMO0 research; GO только на
решение владельца о явно ограниченном implementation scope.** Доказательств
достаточно для реализации semantic core в принятом canonical-A native reference
с проверяемой границей поддержки. Их недостаточно для unrestricted demo support
или полного equivalence claim. Это предложение, не разрешение начать реализацию,
зарегистрировать partial profile или перейти к DMO1.

## 1. Основание и смысл verdicts

Основные источники — существующие отчёты и соответствующие owner-local JSON
`/private/tmp/pf-dmo0-<suffix>.json`:

| Report | JSON suffix / используемое evidence |
| --- | --- |
| [Validation](runtime-layout-demo-10min-validation.md) | `evidence`, `graph-evidence`, `linked`: timer, typed programs, continuation, presentation, persistence |
| [Canonical-A timing](runtime-layout-demo-10min-canonical-a-timing.md) | `canonical-a-timing`: callback/order/budget/SDR native-reference inheritance |
| [Canonical-A jitter](runtime-layout-demo-10min-canonical-a-jitter.md) | `canonical-a-jitter`: выбранная fresh construction и launch arithmetic |
| [Deterministic BYGEL drain](runtime-layout-demo-10min-deterministic-bygel-drain.md) | `deterministic-bygel-drain`: реальный fresh prefix до scored drain35877 |
| [NEW_BALL collision](runtime-layout-demo-10min-first-equality-collision.md) | `first-equality-collision`: producer35967, firing35998, replacement |
| [Post-collision termination](runtime-layout-demo-10min-post-collision-termination.md) | `post-collision-termination`: pause, wrap, repeated equality, реальные continuations |
| [PARTY_ON](runtime-layout-demo-10min-first-equality-party-on.md) | `first-equality-party-on` |
| [SETBALL](runtime-layout-demo-10min-first-equality-setball.md) | `first-equality-setball` |
| [DROPTASK2](runtime-layout-demo-10min-first-equality-droptask2.md) | `first-equality-droptask2` |
| [DURINGFLASH — последний отчёт](runtime-layout-demo-10min-first-equality-duringflash.md) | `first-equality-duringflash`: последний JSON, включая independent inventory producer |
| [Control domains](runtime-layout-demo-10min-control-domains.md) | `control-domains`, `admission-domains`, `canonical-a-whole-dos`: сохранившиеся whole-DOS UNKNOWN |

JSON просмотрены как сохранённое evidence, без обновления и без replay.
Старые NOT_PROVED в ранних отчётах не отменяются задним числом: более поздние
fresh witnesses устанавливают конкретные facts в своём scope. Например,
ранний scored-drain provenance был открытым; deterministic drain и collision
позже доказали выбранный prefix. Аналогично ранний audio-boundary verdict
сменяется **PROVED под canonical-A reference**, не whole-DOS closure.

`CANONICAL_A_TIMING_INHERITANCE`, `NATIVE_AUDIO_BOUNDARY` и
`CANONICAL_A_JITTER_INHERITANCE` — PROVED в своих принятых native scopes.
Canonical A остаётся единственным strict production oracle. Физическая DOS
IRQ/PIT phase и pre-game attract phase не становятся доказанными от этого.
Relocated operand correspondence не означает arbitrary caller/alias closure.

## 2. Доказанный semantic contract

### Fresh construction и counted event

Выбранный reference создаёт `partyland.New(DecodePartyLand(A), A)` и применяет
Legacy settings: fresh one-player Game, player1/ball1, сброшенные rule/task/wait
states и fresh physics. Native clock начинается с0; Sync добавляет1030 до audio
и physics, первый spring consumer получает low8=6. Release использует low8
для скорости и low4 для rotation. Это принятая A convention, не восстановленная
физическая DOS startup phase; clock не сбрасывается на каждом новом шаре.

Demo timer — отдельный uint16 DS:34cd, expired — DS:34cf. Fresh TABLE1 process
начинается с timer0/expired=false. В исследованных new-game/new-ball resets
timer не сбрасывается; его lifetime — table process. Если native frontend
создаёт Game заново внутри того же table lifetime, нельзя автоматически
обнулить этот timer вслед за gameplay clock. Bounded предложение ниже исключает
внутрипроцессный restart/multiplayer; fresh table entry задаёт новый lifetime.

`ElectronicsCalculation` означает admitted rest-of-update: early physics/drain
decision, затем UPDATE_COUNTERS, затем timer increment, затем ordinary
electronics/areas/targets/shift/keyboard/tasks, затем budget-gated matrix и late
physics. Drain читает старый expired **до** increment; его локальный возврат
не исключает electronics suffix. Held ball, между шарами и matrix wait сами
по себе не запрещают count. Native BallLost требует явного accounting hook:
одного BeforeTargets недостаточно. Ни Runner.Ticks, ни rendering frames, ни
каждый physics substep не являются этим событием.

Принятый A schedule сериализован, сохраняет due source updates/PCM и stable
matrix budget. Cadence native table71Hz не превращает threshold в wall-clock
600 секунд. При пропущенном matrix budget electronics может посчитать событие,
а matrix cursor не продвинется. INTRO/attract и pause не считаются ordinary
active electronics. Whole-DOS admission/reentrancy по всем средам остаётся open.

### Equality, expired, pause и expiry program

DO_ELECTRONICS сначала увеличивает uint16, затем сравнивает **ровно35998**.
При equality устанавливает expired=true и source HOLDSTILL=true, играет
`S_GAMEOVER2=(13,0,255)` и устанавливает expiry entry0x1ba17/cursor0x1ba19.
Нет saturation, проверки `>=` или запрета повторного equality при expired=true.
Неравный counter не очищает expired. Inspected continuation/reset consumers
также не очищают его: sticky expired доказан для consumed native scope, не
для всех возможных DOS writers. `S_EMPTY=(62,0,0)` отличается от A priority1.

Expiry stream: CLEAR4 → SCROLL(length98) → FLASHON1 → PRINT13_NUMBER(score,
position344) → WAIT100 → FLASHOFF1 → SCROLL(length88) → FADE256 → WAIT100 →
QUIT(status0). Первый admitted matrix visit после install уменьшает clear5→4.
Matrix progress зависит от visits и сохранённой scroll phase, поэтому QUIT
нельзя назначить фиксированной датой от threshold для всех состояний.
QUIT в witnesses означает вход в linked handler; полный DOS teardown не replayed.

Pause отключает DOS update interrupts/audio. В принятом native frontend P
при35998 даёт замкнутый suspended transition: timer, expired, matrix и session
state не меняются при последующих empty inputs. Поэтому
`EVENTUAL_QUIT_NOT_GUARANTEED` уже доказан без forced-resume premise.
Это не доказательство бесконечной **unpaused** игры; её universal termination
остаётся UNKNOWN. Esc/Y из pause даёт Aborted/Selector/Session=nil в проверенном
native пути; обычный Esc имеет SessionReady guard.

После NEW_BALL collision no-input continuation действительно достигает wrap
при65536, repeated equality при101534 и QUIT102575. Это реальный continuation,
не только uint16 арифметика. Повторное equality ставит expiry заново. Нельзя
обещать quit при первом equality или считать этот один suffix universal bound.

### Drain, bonus и new-ball

Unscored drain (`SCORECHANGED=false`) выбирает PARTY_ON до expired scored-drain
guard, даже если expired=true. Он ставит PARTY_ONTS, очередь PARTY_ON_TASK1,
которая после compare-before-increment wait30 ставит PARTYFLASH и вызывает
NEW_BALL. Free plunge не увеличивает displayed ball counter.

Scored drain при expired=false проходит normal bonus. При expired=true
запрашивает effect0x1a4a1: только **admitted** effect устанавливает expiry
с entry, не восстанавливает старый cursor. Нужны фактические INH_EFF,
SPECIALMODE, cue priority и effect result. Реальные admitted examples:
drain37084 → QUIT38125 и DURINGFLASH drain36173 → QUIT37212; priority255
принимается при старом priority255. Нельзя заменять guard безусловным replay.

Normal linked bonus tail — KOLLA_XXBALL → DEMOVER_CHANGE_PLAYER → CLEAR4 →
WAIT32000 → terminator. При XXBALLE=false demo handler сохраняет текущего
player, обновляет encoded BALLSTEXT, сохраняет player record и ставит
NEW_BALL_TASK. Он не выполняет canonical player rotation/ordinary BALLS
increment и не применяет normal3/5-ball limit. Displayed progression:
1..9,10..19,20..29, затем10..29; это не Session ball-count theorem.
NEW_BALL_TASK wait30 вызывает NEW_BALL/reset; queued SETBALL wait80 ставит
(297,530), VX10/VY0 и явно очищает source HOLDSTILL.

BONUS snapshot/transfer и player save/load установлены локально. HOLDBONUSFLAG
восстанавливает сохранённые12 digits; snapshot может быть прежним при пропуске
FLORPA. Не заменять его unconditional current-bonus copy. Реальный drain35877
имеет четыре zero aggregates, XXBALLE=false, H=0: программа потребляет91 visits,
ставит NEW_BALL_TASK после scan35967, который fires35998. Это доказанный route;
произвольные nonzero/held-bonus/match paths не становятся proved от него.

### Пять first-equality классов

Все пять имеют реальные fresh input-only prefixes в принятом reference.
Исходный source HOLDSTILL отличается от native capture Hold; его отдельные
stores нельзя потерять при снятии capture abstraction.

| Класс | Доказанное поведение | Конкретный outcome / граница |
| --- | --- | --- |
| NEW_BALL_TASK | Заменяет expiry на SHOWPLAYERSTS | PARTYFLASH=VISAKEYS=false; expired остаётся true; старый expiry cursor не сохраняется; no-input QUIT102575 через repeated equality |
| PARTY_ON_TASK1 | Сохраняет expiry | PARTYFLASH=true перед NEW_BALL reset; no-input QUIT37039 |
| SETBALL | Сохраняет expiry, снимает HOLDSTILL | Реальная late movement в35998; no-input QUIT37039 |
| DROPTASK2 | Сохраняет expiry и HOLDSTILL | Drop state/position меняются, expiry hold остаётся; no-input QUIT37039 |
| DURINGFLASH | Сохраняет expiry, снимает HOLDSTILL; scored drain перезапускает expiry | Play-field release35998; scored drain36173; новый entry; QUIT37212 |

**Пять классов не исчерпывают все pending tasks/effects или feasible states.**
Это пять установленных outcomes, а не enum всех возможных демо сессий.
Уже доказанное `ATOMIC_EXPIRY_MODEL = DISPROVED` сохраняется.

## 3. WAIT_FOR_SPIN_TASK: ограниченное чтение consumer

Нового gameplay/replay не было. Прочитаны существующий DURINGFLASH JSON,
`arcade_pending_spin_and_reward_chain` correspondence и bounded linked
TABLE1 consumer0x20a6..0x2296, его existing effect/matrix/drop helpers.
Private inputs сверены существующим `pinned()`; executable bytes, original
text/assets и private identity hashes в этот отчёт не экспортируются.
Ниже адреса code — file offsets, operands state — DS offsets.

### PROVED_LOCAL_EFFECT

WAIT_FOR_SPIN_TASK0x20a6 ждёт SPECIALMODE(34e1)=FF **или** SNURR_READY(d3)=FF.
Если оба false, возвращает без selected body. Затем0x20bc проверяет DS:240c:
FF ведёт0x20f8 → START_DROP0x142a → suicide, минуя priority clear/SPIN.
Paired GROPB и исторический label связывают240c с TILTFLAG. В прежнем
`independent_pending_capture_candidate` этот guard назван jingle-ready;
это неточная аннотация: реальный JINGLE_READY_LOGIC(348b) проверяется позже,
в START_DROP_WHEN_READY0x2262. Прежние JSON/report сохранены неизменными.
Уточнение не меняет five-class witnesses и не закрывает admission UNKNOWN.

Selected non-tilt branch восстанавливает LASTJINGLE(2405) из saved2406,
ставит JINGLEJUMPCNT(2408)=1 и в0x20d3 обнуляет **CURRENT_PRIORITY,
однобайтный DS:3485**, после чего вызывает SPIN0x20df и suicides.
Это table-level jingle/effect admission priority, не SDR callback record
priority и не matrix priority. Если этот body выполняется после expiry cue,
он может понизить255→0. Сам store не пишет expired34cf, HOLDSTILL3026,
matrix program/cursor; текущая expiry routine/cursor от этого не меняется.

SPIN выбирает reward по SLUMPCOUNTER93, SHR1, lookup ARCADESLUMP5b4 и
ARCADEEFFECTS634, затем ADDTASK selected pointer. Это gameplay random counter,
не demo timer34cd и не launch-jitter SLUMP_COUNTERN. Allocation/scan position
определяет, получит ли child visit в этом scan или следующем; выбранный
threshold slot/child не установлен.

### POTENTIAL_EXPIRY_INTERFERENCE

| Selected task / effect | Linked effect/program edge | Последующая task edge |
| --- | --- | --- |
| SPINXB_RUT0x20fe | effect DS:640, inline priority100 → matrix0x1b930 | LIGHT39; admitted branch delay160 → START_DROP_TIMED |
| SPINCL_RUT0x212b | crazy letter consumer выбирает arcade effect DS:a56, priority90 → matrix0x1b936; CHECK_CRAZY может запросить mode effect DS:bfd, priority151 → matrix0x1b30b | CRAL/result и crazy-completion state имеют значение; accepted letter path min140/max180 → START_DROP_WHEN_READY |
| SPIN5M_RUT0x2199 | effect DS:65d, cue priority90 → matrix0x1b265 | admitted min110/max140 → START_DROP_WHEN_READY |
| SPIN1M_RUT0x2166 | effect DS:679, cue priority90 → matrix0x1b203 | admitted min120/max150 → START_DROP_WHEN_READY |
| SPIN500K_RUT0x21cc | effect DS:695, cue priority30 → matrix0x1b96a | admitted min45/max70 → START_DROP_WHEN_READY |
| SPINNS_RUT0x220e | effect DS:6b1, inline priority100 → matrix0x1b952 | delay45 → START_DROP_TIMED |

Rejected ordinary rewards route through delay10 → START_DROP_TIMED; SPINNS
сохраняет45. SPECIALMODE может bypass timed wait. Readiness drop path содержит
minimum/maximum, readiness latch и SPECIALMODE guards. START_DROP сам не ставит
новый matrix program: пишет drop position/state и ставит DROPTASK1/DROPTASK2
и camera work. Последующий физический drain требует отдельного реального prefix.

Matrix-changing edge находится **в effect consumer**, не в priority store:
0x5f14..0x5fb3 сохраняет admission flags, применяет score/bonus arithmetic,
и лишь при разрешённом result, nonzero program, INH_EFF3494=false и
SPECIALMODE34e1=false достигает DO_MATRIX call0x5f9f. Cue helper сравнивает
requested priority с CURRENT_PRIORITY unsigned; ниже — reject, equal — permit.
Для no-cue effect используется inline priority после zero cue pointer;
его program word смещён на один byte относительно cue-backed record.
DO_MATRIX заменяет программу без отдельного expired-protection guard.

Таким образом понижение priority делает lower-priority rewards **потенциально
admissible** после expiry, но не гарантирует replacement: нужны selected task,
guards, actual effect/cue result и slot order. SPECIALMODE, например, допускает
WAIT body, но блокирует обычную effect matrix installation. Crazy mode имеет
дополнительный state/carry predicate. Нельзя объявить все rewards reached либо
считать direct priority clear cancellation expiry.

### FIRST_EQUALITY_REACHABILITY_UNKNOWN

Сохранённый independent inventory prefix реально достигает GROPB1001 и pending
`arcade.func1`, два fresh replays согласуются. Это доказывает gameplay producer,
не его selected body/child при35998. Не доказаны pending-spin guards, slot,
selected reward и admitted effect suffix на первом equality либо их исключение.
Для такого witness может понадобиться отдельный trajectory pass; он здесь
не запускался. Практическая значимость высокая для **arbitrary playable demo
equivalence**, поскольку matrix replacement способен отложить QUIT. Для
restricted core с явно unsupported unvalidated expiry/reward interaction эта
reachability theorem не является prerequisite самого implementation start.

## 4. Gap triage без открытия новых доменов

В столбце «bounded blocker» рассматривается только предложение из раздела5.
«Можно исключить» означает явное ограничение и обнаружение выхода за него;
отсутствие свидетельства не является доказательством невозможности.

| Вопрос | Уже proved | UNKNOWN / возможное observable изменение | Bounded blocker | Strict equivalence blocker | Честная граница |
| --- | --- | --- | --- | --- | --- |
| Expiry/interleaving | Пять классов, replacement/release, repeated equality, два scored expiry replays | Полнота pending-task/effect census; WAIT selected threshold suffix; дальнейшие конкурирующие replacements могут менять hold, score, display и дату QUIT | Да для unrestricted controls; нет для enforceable consumed-path envelope | Да | Не поддерживать непроверенный competing suffix как verified; не сделать atomic expiry или priority immunity |
| Active-game termination | Конкретные QUIT; legitimate pause counterexample | Universal unpaused liveness, alternate pre-threshold session ends; отсутствие replacement cycles не доказано | Нет при отказе от universal deadline/liveness promise | Да для полноты exit behavior; паузную universal guarantee уже опровергли | Pause разрешена без time cap; гарантировать только проверенные continuations; остальные outcomes explicit unsupported/unverified |
| Score/bonus/new-ball | Fresh scored zero-aggregate route, локальный normal tail, save/load, waits, same-player continuation, free unscored plunge | Nonzero/held snapshot lifetime, XXBALLE/match/extra-ball/high-score state feasibility; может менять score, progression, program и exit | Нет для zero-route/core; да для обещания всех bonus/progression states | Да | Fresh single-player Legacy; только проверенные routes и state dependencies; не перенести canonical3/5-ball progression |
| Persistence/high scores | Factory seeds=A; read/write bodies0x66d2/0x6706 локализованы; нет direct/candidate predecessor в reviewed graph | Ни VariantA, ни VariantB не proved; indirect/lifecycle путь может прочитать/записать HI | Нет для явного volatile native policy; да для faithful persistent demo claim | Да | Factory/transient scores, без load/save; это выбранное ограничение, не theorem «DOS never persists» |
| Controls/alternative exits | Flippers/plunger реальные prefixes; active tilt/music/pause и attract-only cheat gate локально; native P/Esc/Y witness | Полная lifetime/guard closure, multiplayer/F1–F8, cheat effects, quit-question и остальные exits могут менять state/end | Нет для названного control subset; да для всех DOS controls | Да | Назвать поддержанные controls; unsupported starts/cheats/restarts explicit; host abort как native policy, не полная DOS exit parity |
| INTRO/launcher handoff | BLOCKER3 consumer presentation closed: crops/menu/availability; локальный TABLE1 status0→launcher INTRO edge | Все indirect/teardown/environment outcomes, failures, exact DOS restart lifecycle | Нет при исключённом DOS INTRO/launcher; да для faithful whole-app flow | Да | Direct native Party Land entry и возврат в native selector; INTRO parity deferred; не объявлять launcher exits исчерпанными |
| Indirect/callback domains | TABLE1 18 bounded/12 UNKNOWN; CODE2 0/2; native timing/audio inheritance proved | DS/ES/indexed aliases, SS:SP/callers, INT66 admission, PRINTTASK/KEYTASK init, task/reset writer completeness, area writers, INTRO far domains; могут добавлять consumed effects | DOS machine obligations нет для chosen native reference; unknown semantic producers да, если входят в support envelope | Да | Shared typed native consumers в reviewed scope, исключён arbitrary DOS environment; никакой address allowlist/observed-target «closure» |

SOUND.CFG допускает unpinned external EXEC без filename/read-length/result
closure. Это реальный whole-DOS admission gap, а не требование запустить ещё
один нативный spin search. Physical placement/stack/PIT и attract history
остаются below accepted native reference; изучать их повторно для bounded
варианта не требуется. Semantic UNKNOWN нельзя скрыть этим classification.

## 5. Два стандарта готовности и acceptance criteria

### Strict DMO0 equivalence — NO GO сегодня

Нужна полнота consumed demo behavior относительно явно названного reference,
а не только пять удачных traces. Для original DOS equivalence дополнительно
нужен environment/scheduler scope: нельзя выдать A logical convention за
все физические DOS machines. Existing whole-DOS closure gate сохраняется.

Acceptance criteria:

1. Замкнуть relevant indirect/callback/writer/lifecycle domains либо доказать
   непричастность каждой оставшейся неопределённости к заявленному contract.
   Finite candidate sets и невстреченные targets не закрывают domain.
2. Доказать полный feasible expiry/task/effect interaction contract: сохранять
   установленные пять outcomes и классифицировать остальные consumed cases,
   включая WAIT guards/rewards. First equality census не заменяет последующие
   interleavings, repeated equality и terminal-state coverage.
3. Покрыть весь progression/bonus/held-snapshot/match/high-score/control/start
   lifecycle; классифицировать persistence A/B без negative inference от CFG.
4. Установить полный termination/alternate-exit/INTRO handoff contract.
   Не требовать уже опровергнутой unconditional eventual QUIT: явно определить
   pause/fairness domain и доказать поведение внутри него.
5. Сохранить accepted A timing/jitter/audio premises, demo counted-event и
   matrix-budget order; для exact DOS claim закрыть физическую admission часть.
   Exhaustive consumed differences и production data consumer map должны быть
   reviewable, не основаны на ближайшем byte match.
6. После будущей реализации — original-backed conformance/negative checks,
   A/B/C/D regression и честная классификация недоступных проверок. Known PF6
   baseline не превращать в PASS; missing OriginalTrajectories не подменять.

Успешный WAIT witness не выполняет пункты1,3,4,5 и сам по себе не выполняет2.

### Bounded native demo implementation — YES, предложение для owner decision

YES означает **достаточность evidence для начала реализации ограниченного
semantic core**, а не готовый продукт, approval или equivalence certificate.
Предлагаемая первая граница: canonical-A fresh one-player Legacy reference,
direct Party Land entry, table-lifetime timer, reviewed shared primitives,
demo normal zero-aggregate continuation, five classified expiry interactions,
их проверенные suffixes, pause и native abort. Persistence, DOS INTRO/launcher,
multiplayer/restarts/cheats и unvalidated bonus/mode/expiry interactions
исключены из verified support. Unrestricted interactive demo readiness остаётся
UNKNOWN. Нельзя назвать эту границу полной official demo implementation.

Acceptance criteria для такого отдельного решения:

1. Владелец явно принимает support envelope, volatile scores, native entry/exit
   policy и отсутствие общего10-minute wall-clock/unpaused QUIT promise.
   Пользовательская документация различает official input data и bounded native
   behavior; исключённые пути не объявляются невозможными в оригинале.
2. Реализация вводит явный admitted ElectronicsCalculation и uint16 timer:
   old-flag drain order, counted post-drain/BallLost path, pause/INTRO exclusion,
   equality35998, sticky expired, wrap/re-equality, lifetime сохранение.
   Нельзя использовать Runner tick, wall clock или physics-step counter.
3. Matrix/effect/task primitives сохраняют accepted A order/budget, first-free
   50-slot scan, shared waits compare-before-increment и source HOLDSTILL.
   Реализуются typed demo continuation/expiry/cue differences; никакого
   uninterruptible expiry, forced task purge, priority clamp или cursor resume.
4. Support envelope проверяется по **semantic state, producers, guards и
   consumed paths**, не по task names или пяти checkpoint snapshots. До первого
   неподтверждённого consumed edge должен быть явный unsupported outcome,
   без guessed transition, default canonical progression или silent fallback.
   WAIT/reward near expiry и непроверенный held/nonzero bonus сюда относятся.
   Если эту границу нельзя надёжно обнаруживать, YES не даёт interactive GO:
   остаётся только ограниченный test/reference prototype до нового решения.
5. Conformance harness проверяет существующие fixed fresh scripts/checkpoints,
   пять разных full-calculation outcomes и указанные concrete QUIT continuations,
   repeated equality и pause; predicates получают meaningful negative tests
   для first unsupported edge. Это критерий **будущей реализации**: никаких
   replay в текущем проходе. Времена не вшиваются вместо программы/physics.
6. Before support registration нужны reviewable decoder/profile boundaries,
   coherent detection/hybrid rejection и A/B/C/D regressions; canonical A oracle
   не меняется. Здесь нет разрешения добавить partial profile. Bundling, rights,
   packaging и DMO1 требуют отдельного owner scope/решения, не следуют из YES.

Такой bounded подход допускает implementation unknowns как документированные
границы; он не допускает runtime guesses, представленные как proved behavior.
Если владельцу нужна свободно играемая демоверсия без unsupported guard, этот
предложенный envelope недостаточен: нужен другой scope и явное принятие
unverified paths либо targeted semantic closure до их verified поддержки.

## 6. STOP/GO и следующий этап

Ещё один узкий WAIT pass имеет **ограниченную, условную пользу**. Он может
доказать selected body на35998 и конкретный reward/slot/effect suffix либо
механическое exclusion invariant. Unsuccessful finite search ничего не
исключает. Даже successful witness расширит число concrete outcomes; он не
докажет exhaustive pending-task census, persistence, alternate exits,
whole-DOS callbacks или universal unpaused termination.

Для решения «можно ли начать bounded core» local interference edge уже
достаточно понятен. Новый trajectory witness не снимает необходимости выбрать
support standard, поэтому сейчас его marginal decision value недостаточен.
**STOP следующему автоматическому WAIT trajectory pass.** Возобновлять research
разумно лишь с owner-selected вопросом, который реально блокирует выбранный
product scope, заранее определёнными acceptance/stop conditions и budget.

Для перехода к реализации требуется owner choice: strict closure остаётся
целью или принимается отдельный bounded contract из раздела5; затем review
конкретного implementation scope/guard policy и отдельная авторизация кода.
Данный проход ничего из этого не начинает. Можно разумно заморозить DMO0
research сейчас: сохранить отчёты/JSON и UNKNOWN inventory, статус NOT CLOSED,
отдельно зафиксировать доказанный core. Freeze не означает proof closure.

## 7. Проверка текущего прохода

Research/report only. Выполнены чтение существующих metadata, pinned-input
проверка и bounded linked-consumer review, без trajectory searches, full-horizon
replays или запуска audit CLI, который мог бы их вызвать. Старые validation
counts принадлежат своим отчётам; не представлены как проверки этого прохода.
Для нового Markdown выполнены whitespace и source-link checks, а сохранность
существующих docs/tools/production files и DMO0 JSON сверена по pre-write manifest.

Production `internal`, `hosts`, `cmd`, существующие артефакты и `.DS_Store`
не изменены. DMO1, commit/push/tag/release, v0.1.3, PF6 fix и создание missing
TestOriginalTrajectories fixture не выполнялись.

DMO0_STRICT_CLOSURE = NOT_PROVED

BOUNDED_IMPLEMENTATION_READINESS = YES

EXPIRY_INTERLEAVING = NOT_PROVED

Рекомендуемый следующий этап: owner scope decision по bounded semantic core
с enforceable unsupported boundary. Он использует уже доказанные правила и
останавливает расширение research без притворства, что остальные DOS paths
закрыты. До такого решения DMO0 research можно заморозить; реализация не начата.
