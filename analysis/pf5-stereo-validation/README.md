# PF5 stereo validation — 2026-10-01

PF5 now renders 48-kHz signed16 LE interleaved left/right frames, with fixed
voices 0/3 → left and 1/2 → right. Each side accumulates and saturates separately.
SDL requests exactly two channels; its pre-roll remains 2048 frames (~42.7 ms)
and queue cap remains 12000 frames (250 ms), now 8192 and 48000 bytes.
The initial stereo change retained tracker code, interpolation, sample advancement,
cue code, and game clocks. Follow-up fixes are recorded below. No module bytes or panning commands changed. The only changed pinned
fixture field is `bumper900_sha256`, initially generated independently with a silent right
channel (now duplicated for centered runtime effects) by `tools/reference_pf5.py`.

Validation:

- Full `./tools/go.sh test -p=1 -count=1 ./...` and full `-race` runs:
  these historical runs predate factory-seed inventory validation.
  Current full-suite results are in `../personal-clean-seeds.md`.
- Full `./tools/go.sh vet ./...` reports existing unkeyed `audio.Effect` literals
  in `internal/speeddevils/audio.go:58`. Vet for audio/platform passes.
- Native build passed. Headless 1200-sync Party Land script retained score
  2300000 and ball 2; Speed Devils retained score 11030 and ball 2.
- Routing tests isolate all four voices, including voice 3's tracker left routing;
  check interpolated fractional phases, output frame size, both polarities of
  independent saturation on both sides, full-volume mixes, loops and sample ends.
  Existing chunking, cue timing, priority, replacement, frontend continuity,
  and silent/native gameplay comparisons pass.
- An additional temporary comparison against the pre-change mono Render method
  ran 4000 blocks for each of TABLE1.MOD, TABLE2.MOD, INTRO.MOD and MOD2.MOD,
  with periodic Force calls. Full Player state remained identical after every
  block; summed stereo contributions matched the mono PCM for every frame.
  The temporary baseline code was removed after passing.

Real desktop smoke used `tools/pf8_selector_audio.py` with X11 and PulseAudio,
22 seconds of playback and 24 seconds of stereo monitor recording. Thirteen
Space edges exercised selector/text transitions. `selector-pulse.log` reports
zero resets, zero empty queues, zero lifecycle clears and one device start.
Both monitor channels contain distinct non-silent PCM (`stereo-smoke.json`).
`tools/pf8_pcm_continuity.py` compared offline stereo INTRO PCM with the real
monitor: **960000 frames (20 seconds) were byte-exact on both channels**, with
constant 6656-frame host lag, gain 1, correlation 1 and zero drift.
See `selector-pcm-continuity.json` and `selector-audio-inputs.json`.

Large WAV/PCM recordings and exported gameplay PNGs remain in
`/tmp/pf5-stereo-smoke`; small evidence records are kept here.

## Follow-up: centered effects and persistent music mute

Runtime sound effects now duplicate their interpolated contribution to both
output channels using the existing fourth voice and sample phase. Music retains
Amiga stereo routing; a tracker note or instrument reclaims that voice's left
routing. The independent bumper fixture now contains the effect in both ears.

M now clears the saved background return as well as stopping current background
music. Both tables redirect background returns to their empty order while muted,
including returns overwritten by rule scripts. Active jingles continue, and M
can restore spring/main music. Tests exercise successive finite jingles, native
and silent cue state, rule return overrides, and M during active jingles.

Follow-up validation: this historical regression run predates factory-seed
inventory validation; see `../personal-clean-seeds.md` for current results. Audio, Party Land, Speed Devils and frontend race
checks pass; audio/platform/Party Land/frontend vet and native build pass.
Tests additionally verify that PCM is silent after finite muted jingles.
The real PulseAudio/X11 gameplay smoke enters both tables, toggles M, exercises
both flippers, pauses and returns to the selector. It passes with zero queue
resets and empty queues (`fixes-pulse.log`; reproduction `fixes-pulse-smoke.py`,
run from the repository root). An initial older UI smoke helper missed a
selector transition; the final run used the existing desktop key helper and
longer transition delays.
