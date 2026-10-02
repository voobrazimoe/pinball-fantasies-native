# PF5 — native Party Land audio and callback timing investigation

PF5 decodes TABLE1.MOD at runtime and produces native audible music, jingles and
sample effects. Gameplay remains on PF4.5's deterministic semantic timeline.
No other table, frontend, CPU interpreter or sound hardware emulator is included.

## Original resources and evidence

VERIFIED — PLAND.ASM `SOUND STRUCTURES` supplies the four-byte
sample/note/effect/channel records. `SBASE=0x16`, `EBASE=7`; sound samples are
**inside TABLE1.MOD**, not separate PCM files in TABLE1.PRG. FANTASIE.MAC
`SOUNDEFFECT`/`SOUNDEFFECT2` passes sample in CL, note in BL, optional volume in
BH, and channel+1 in DL to service17. The effect byte in these records is unused
by that macro. All retained Party Land effects select channel3 (fourth voice).
Zero volume argument preserves the instrument's default volume; it does not mute.

VERIFIED — TABLE1.MOD SHA-256:
`a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5`.
Title `pinball2-table1`, M.K. signature, four channels, 64 orders and 64 patterns,
31 sample headers (26 nonempty). The module decoder deliberately rejects any
other input. Big-endian word lengths and loop coordinates are multiplied by two.
Samples are signed eight-bit mono PCM. Volume is 0..64. Only sample10 has nonzero
finetune (+2). Loops longer than two bytes repeat; length2 is the no-loop marker.
All lengths, offsets, effect parameters, loop metadata and raw sample hashes are
in `pf5-audio-fixtures.json`, independently extracted by `tools/reference_pf5.py`.

Representative sound mappings (one-based sample, zero-based note index):

| Source label | Sample | Note index | Period | Length bytes |
|---|---:|---:|---:|---:|
| SBUMPER1 |24 BUMPER|25|202|6116|
| SBUMPER2 |24|23|226|6116|
| SBUMPER3 |24|21|254|6116|
| SBRICKNER |22 BRICKNEDGANG|18|302|3986|
| SBRICKUPP / SGROP |23 BRICKORUPP|23|226|7486|
| SFLIPPUPP |25 FLIPPERUPP|22|240|2238|
| SRINNER / SNEWBALL |28 NEWBALL2|18|302|12810|
| SKICKER |29 SIDOBUMPER|18|302|3900|
| SFJADER |30 UPPSKUTARE|18|302|2910|
| S_SCORELJUD |7|18|302|9352|

VERIFIED — the static selected SBLASTER driver period table begins
856,808,762,720 and contains the same 36 notes used by native sound intents.
Its pitch calculation uses numerator `0x361f0f` = 3546895. Channel selection in
service17 indexes 0x33-byte voices after decrementing DL. The service calls the
same trigger/instrument/effect code used by module rows. Consequently effects
replace the fourth tracker voice and later tracker triggers can replace effects;
there is no invented extra sound lane, duration priority or completion callback.
FANTASIE.ASM `DSSK`/`DSSKR` additionally triggers SFLIPPUPP on keyboard press edges;
the native Playing input boundary now emits those non-blocking sound intents.

## Implemented module subset

VERIFIED — actual data inspection finds only effects:
0 arpeggio, 1 pitch slide up, 2 pitch slide down, 3 tone portamento,
4 vibrato, 6 vibrato plus volume slide, 9 sample offset, A volume slide,
B position jump, C volume, D pattern break, E9 retrigger, F speed.
Every E command is E92/E93/E95; every F selects speed1..15, never BPM.
There is no generic tracker compatibility layer. E6/EE loops/delays, other E
commands, effect5, tremolo and other formats are not implemented.

VERIFIED — instrument selection restores sample volume; triggering restarts
sample phase; tone portamento sets a target without retriggering. Effects on
later ticks update pitch/volume; volume stays 0..64, pitch slides stay113..856.
Portamento, vibrato nibbles and sample offset retain their required memories.
Arpeggio cycles base/high/low nibbles; retrigger restarts on multiples of its tick
parameter. D uses decimal-coded row and advances the order. F changes the current
row speed. B routes its target through the existing game handler's repeat/return
state. The implemented F/B/D timeline is checked for every supplied order against
PF4.5's independent fixtures, including continuation across Dxx.
Supporting tracker effect meanings were checked against the
[ProTracker author's built-in manual, preserved transcription](https://github.com/echolevel/Protracker-2.3D-Helpfile-Manual).
Original game source/data take precedence over general tracker conventions.

## Deterministic clocks and semantic coupling

VERIFIED — one Game.Sync still advances the original PF3 integer physics/rules
and PF4.5 tasks/matrix once. Nothing reads audio-device state to change those
systems. The existing rational semantic clock adds50 per sync and consumes71 per
50-Hz tracker tick. Completion remains before that sync's physics/electronics.
The native renderer emits floor(48000*N/71) total frames after N syncs, retaining
the fractional remainder. Each tracker tick is960 output frames. Rendering frames
or discarding them never advances gameplay tasks independently.

VERIFIED — `Game.AttachAudio` attaches a player without modifying game semantic
state or event fixtures. Accepted `playJingle` forces the renderer to the same
order and clears the established ReadyAnim/ReadyLogic flags. Rejected requests
leave audible position intact while retaining ASM's repeat-swap quirk. The
renderer's B target uses ReturnPosition when JumpCount==1; the established
`audioTick` still runs JINGLE_HANDLER decrement/clamp/ready/priority restoration
and emits AudioCue/AudioComplete. No SDL device callback invokes game code.
Saved return position follows the semantic current order, including D/order
continuations; accepted forces reset phase. Residual tracker time survives B.

INFERRED, deliberately retained — the 50-Hz tick base, full B-row duration before
completion, reset-to-speed6 at cue entry, forced phase reset, and first enclosing
71-Hz game boundary callback convention are the PF4.5 contract. They remain
explicitly qualified, not promoted to verified DOS timing by audible playback.
Initial native spring music starts at order0 without adding new semantic ready
or completion events to the authentic PF4 initial state. CLOSE1 and source new-
ball paths subsequently select music through the established jingle intents.

## SDR callback-phase investigation: PARTIALLY RESOLVED

VERIFIED — SOUND.CFG selects SBLASTER.SDR. The linked SDR files use SP literal/
repeat packet compression. Naively disassembling their packed tail misaligns
branch targets and data offsets. `tools/inspect_pf5_sdr.py` statically expands
these data packets, without executing instructions, applying DOS relocations,
loading drivers, or emulating hardware. It retains the uncompressed prefix and
records original/expanded hashes for SBLASTER, SB16 and NOSOUND. The retained
analysis is narrowly the service registration, force, tracker row/jump and
external game-callback cone. Disposable expanded files go under /tmp.

VERIFIED — offsets below are **expanded image offsets**, after the512-byte MZ
header; original raw-file offsets are not interchangeable. Reproducible selected
instructions are in `pf5-sdr-callback-disassembly.txt`.

- Service19 at image0x19e registers ES:DX in DS:0x66de/0x66e0.
- The effect dispatch table at image `0x1e20+0x4ca` has entry11=0x12e4.
  Row code at0xd73 reads each packed channel, calls0xe05, then0x109f dispatches
  the effect. B's0x12e4 path moves target AH to AL and calls the registered far
  user routine at0x12f8. It stores returned AL in order0x50a, decrements the
  order, and sets break flag0x6fe=1. Row completion subsequently increments the
  order and chooses its first row. Thus the callback is **during B-row entry
  processing**, before that row's remaining ticks and before mixing that row.
  It is not a sample/device buffer-completion event.
- Tick code at0xcf8 decrements countdown0x787; zero reloads speed0x786 and
  calls row processing0xd73. Otherwise it calls per-tick effects0xe19. Mixing
 0xe2b follows row/effect processing at0xd14.
- Service16 at0x1fb exchanges the requested order-minus-one with0x50a, sets
  break flag=1, and sets countdown=1. It does **not** reset mixer fractional
  accumulator0x897 or the row speed byte0x786. Force therefore preserves
  residual mixer time and requests the position change on the next tracker
  processing boundary; it does not simply restart an entire tick at the API call.
- B code similarly does not clear0x897. The mixer updates it after processing,
  at0xd1a..0xd2d. Residual sample time survives the jump callback.
- The raster scheduler has a music-crisis check0x1d08 before the external
  scheduled game routine call0x1cec. This establishes a pre-game audio-budget
  check, but that check itself does not establish when a B callback occurred.

UNKNOWN — exact correspondence of mixer service/fill boundaries, residual time,
forced-position application, and the configured VBLANK/late raster schedule at
runtime. The source installs game callbacks with services11/12 and priorities
100/200, but the inspected cone does not prove a universal before/after game Sync
relationship for every B or every force. Other driver revisions' equivalence is
not claimed. Reconstructing the general IRQ/DMA/raster hardware scheduler is out
of scope and was not undertaken.

PF4.5 adaptation review: **Missing SDR callback phase = PARTIALLY RESOLVED**.
B-row entry and retention of fractional mixer time are now established, while
exact game-boundary and forced-return phase remain UNKNOWN. Per the PF5 request
to retain PF4.5's convention if exact phase cannot be established, **silent
fixture timelines did not change**. Known discrepancies with the selected DOS
binary's entry/force behavior are explicit above. For example, Mystery's native
completion remains sync319; this must not be advertised as an exact DOS timestamp.

## Mixer and host output

VERIFIED — native four-voice fixed-point sample phase (Q32), deterministic linear
interpolation, independent left/right accumulation and clipping, interleaved stereo signed16
little-endian at48 kHz. Fixed Amiga routing sends voices0/3 left and1/2 right;
runtime effects on the fourth voice play in both output channels on the same
sample phase. Tracker notes reclaim that voice and its left routing.
Pitch step is calculated from period with floating-point arithmetic once per
pitch update; offline rendering is byte-identical on this toolchain/platform.
No host clock enters sample stepping. Volume maximum and two signed8-bit voices per side
fit the explicit int16 saturation policy: -32768..32767, never integer wrap.
Interpolation handles looping endpoints and non-looping tails.

VERIFIED — SDL2 queued audio is the minimal host boundary. SDL copies submitted
buffers and consumes them asynchronously. Startup waits for8192 queued bytes
(~42.7 ms) before unpausing; device blocks are1024 frames. Presentation uses71-Hz
absolute deadlines that include frame work; lateness catches up in whole original
Sync calls. The previous additional14-ms sleep after frame work could underfeed
real audio and was corrected. Rendering/game ticks remain fixed-size.
Queues above48000 bytes (~250 ms) may be cleared to bound latency. Output loss,
empty queues and missing devices have no semantic effect. An unavailable device
prints a diagnostic and continues gameplay. Headless PNG/test modes never open
an audio device and can discard all rendered bytes.

Known audible differences from DOS: native stereo48-kHz output and linear
interpolation, no selected sound-card filter/noise/DAC quantization; mathematical
finetune ratio rather than a claim of byte-exact DOS finetune table behavior;
conventional bounded effect behavior rather than undocumented SDR quirk
compatibility. Source optional plunger volume modulation is not reconstructed;
the existing semantic SFJADER intent plays at its instrument default volume.
Keyboard input edge sounds are quantized to native Sync. These audio adaptations
do not alter physics, rules or readiness. No exact acoustic identity is claimed.

## Validation and audible test

VERIFIED — module hash/orders/pattern count, all sample metadata and hashes,
actual effect parameter inventory; every supplied cue's F/B/D duration/target;
synthetic exact tick/row, decimal D12, force, E9, porta, vibrato/volume behavior;
independent900-frame interpolated SBUMPER1 hash; loop/no-loop rendering,
byte-identical chunked vs whole-second output, all effect render repeatability,
negative/positive mixer extremes; fourth-voice replacement and later tracker
reclaim while other music channels continue.

VERIFIED — native-attached vs silent gameplay compares each sync's complete event
list, Audio readiness/priority/position state, score, phase, ball number and matrix
state for Mystery, Crazy letter, collected extra ball, all four audio-dependent
arcade prizes, Jackpot repeat, forced current-order return and drain/new ball.
Discarded output buffers cannot change those comparisons. A1200-sync native
script still reaches score2300000 and ball2 with byte-identical repeated PCM.
PF0–PF4.5 tests, including physics trajectories and matrix/task traces, pass.
Race checks for audio/Party Land/physics, `go vet ./...`, the native build and
headless PNG script all pass.

VERIFIED — real desktop SDL/PulseAudio + X11 scripted20-second smoke tests ran.
The user confirmed music/effects audible, initially reported slight distortion
and stutter, then confirmed **"Smooth; distortion/stutter resolved"** after fixed
pacing, prebuffering and linear interpolation. The second run recorded
queue resets=0, empty-queue observations=0. A final8-second smoke of the build
including flipper press sounds also recorded zero resets/empty observations. No gameplay divergence was observed;
headless comparison provides the deterministic evidence. Audible checks validate
practical playback, not the unresolved historical callback phase.

Files created: `internal/audio/module.go`, `player.go`, `player_test.go`;
`internal/platform/audio.go`; `internal/partyland/audio_test.go`;
`tools/reference_pf5.py`, `inspect_pf5_sdr.py`; this document,
`pf5-audio-fixtures.json`, `pf5-sdr-evidence.json`,
`pf5-sdr-callback-disassembly.txt`.
Files modified: `internal/partyland/game.go`, `timing.go`;
`internal/platform/physics.go`; `cmd/pinballfantasies/rules.go`;
README, build guide, PF4.5 adaptation evidence. Original binaries/data/source and
PF0–PF4.5 fixtures are unchanged. No PF6 work was started.

Exact commands from `/home/mess/work/PINBALLF`:

```sh
python3 tools/reference_pf5.py
python3 tools/inspect_pf5_sdr.py
./tools/go.sh test -p=1 -count=1 ./...
./tools/go.sh test -p=1 -race ./internal/audio ./internal/partyland ./internal/physics
./tools/go.sh vet ./...
./tools/go.sh build -buildvcs=false -p=1 -o bin/pinballfantasies ./cmd/pinballfantasies
./bin/pinballfantasies -data-dir . -pf4
./bin/pinballfantasies -data-dir . -pf4 -pf4-script -duration 20s
./bin/pinballfantasies -data-dir . -pf4 -pf4-script -ticks 1200 -png /tmp/pf5-script.png
```
