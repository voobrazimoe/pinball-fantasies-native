# PF8 continuation evidence

Run commands from `/home/mess/work/PINBALLF`. External DOSBox-X processes are
oracle tools only; none of these tools is imported into native packages.
Root high scores are never written. New oracle directories must be empty.

## Native states and real audio

```sh
./tools/go.sh run -buildvcs=false ./cmd/pf8temporal
python3 tools/pf8_isolated_x.py python3 tools/pf8_selector_audio.py
/home/mess/.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python3 tools/pf8_pcm_continuity.py
python3 tools/pf8_isolated_x.py python3 tools/smoke_pf8.py
python3 tools/compare_pf8_temporal.py
```

PulseAudio, ffmpeg, Pillow and numpy are needed for real output capture/comparison.
The PCM comparison reports byte equality of a continuous20-second interval,
gain, correlation and independent per-second latency drift. `audio-before/`
is the A/B probe of only the former native SDL lifecycle reset condition.
Dummy-audio smoke checks persistent window/control flow, not audible parity.

`pf8_isolated_x.py` and `pf8_isolated_oracle.py` use the locally extracted
`.tools/xvfb/usr/bin/Xvfb`. On another host, provide Xvfb at that path or adapt
the helper to the installed executable. No system package change is required.

## Fresh original captures

```sh
python3 tools/pf8_oracle.py .tools/pf8-new-reference
python3 tools/pf8_isolated_oracle.py .tools/pf8-new-reference.conf analysis/pf8-new-capture normal
python3 tools/pf8_isolated_oracle.py .tools/pf8-new-reference.conf analysis/pf8-new-speed attract-speed
```

Recorded continuation configurations used cycles200000 instead of the default
30000, with aspect=false/scaler=none/output=surface. CPU cycle settings affect
the external oracle only; timestamps are not native game-clock authorities.
`normal-isolated/` covers startup, first selector/text and Party entry.
Its final scripted F2 transition did not enter Speed; those trailing frames
are not used as Speed evidence. `speed-attract-isolated/` is a separate fresh
unmodified Speed entry. `original/` is provisional desktop capture because the
user may have pressed keys; do not use its temporal ordering as certification.

## Controlled source-data probes

```sh
python3 tools/pf8_runtime_fixture.py bitmap .tools/pf8-new-bitmap
python3 tools/pf8_isolated_oracle.py .tools/pf8-new-bitmap.conf analysis/pf8-new-bitmap bitmap
python3 tools/pf8_runtime_fixture.py ending-probe .tools/pf8-new-ending
python3 tools/pf8_isolated_oracle.py .tools/pf8-new-ending.conf analysis/pf8-new-ending ending-speed
python3 tools/pf8_runtime_fixture.py ending-real .tools/pf8-new-real-ending
python3 tools/pf8_isolated_oracle.py .tools/pf8-new-real-ending.conf analysis/pf8-new-real-ending ending-real
```

Bitmap copies the existing GEARTS10-byte source program to ShowHighsTS in an
isolated TABLE2.PRG; executable code and animation assets remain original.
`bitmap-data-isolated/matches.json` contains69 live matches across five distinct
images; native tests compare the entire raw matrix including duplicated X
samples, palette and gaps. Metadata records hashes and rejected earlier probes.
This validates presentation/program playback, not the natural gear-award trigger.

Ending-probe changes copied command/state data only and copied HI to100. It is
**not valid evidence of a complete normal ending**: attract-mode keyboard reading
consumed two initials. Ending-real leaves PRG unchanged, sets copied three-ball
configuration and safe HI100; the recorded108.33-second attempt never qualified.
These are concrete unresolved results, not successful flow certifications.
Further selector/ending validation was subsequently waived by the user.

`selector-text-scores.json` pins the copied original score bytes for the text
capture regression, independent of future changes to user scores. Values0..9
are forbidden in oracle HI because original INTRO/MOVSCORE underflows its loop
and corrupts data; no root HI is changed to prepare these probes.

`side-by-side.png` contains original/native title, Digital Illusions and text
fade comparisons. It is a contact sheet, not a geometry measurement. Actual
comparison removes sample duplication only after verifying all duplicates.
The report in `analysis/pf8-dos-parity.md` distinguishes full-frame matches,
region matches, animation cadence sampling limits and uncertified timing.
