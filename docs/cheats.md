# Original codes and matrix test build

The original DOS `FANTASIE.ASM/MAIN` calls `CHECKCHEAT` on the loaded table's
attract screen, before starting a game. Type codes there using Latin letters;
spaces separate words. The port retains this restriction and the original
prefix recovery. Pausing during a game does not enable cheat entry.

| Code | Original effect |
| --- | --- |
| `EARTHQUAKE` | Disable tilt penalties; retain the physical table push. |
| `EXTRA BALLS` | Set `NO_OF_BALLS` to 5, as written by `BALLSRUT`. The original scroll says 7. |
| `SNAIL` | Set `SHIFTKEYS` bit 2, skipping the second late-raster ball calculation. The native high-resolution session initially uses this setting. |
| `FAIR PLAY` | Restore tilt penalties, set 3 balls, clear `SHIFTKEYS` bit 2. |
| `CHEAT` | Display the original hint. |
| `JOHAN`, `DANIEL`, `GABRIEL`, `TSP`, `TECH`, `ROBBAN`, `STEIN`, `GREET` | Display the corresponding original developer message. |

Messages are read from the supplied original `.PRG`. Their matrix programs use
each table's source `CLEARIT`, the common scrolling routine and its terminator.
After the scroll, the original attract/high-score program resumes. Codes may
replace a message already in progress. Gameplay flags persist across new games
on that loaded table and reset when a table is loaded again.

## Separate Party Land test build

Build locally with verified originals:

```sh
python3 tools/build_personal_release.py /path/to/originals --windows-only --test-cheats
```

This creates `release/personal/windows/pinballfantasies-matrix-test.exe` without
replacing the normal release. Only this `matrixdebug` build has these keys:

- `X`: arm both original side-lane Extra Ball lamps (39 and 40).
- `L`: arm the lamps and place an already launched ball in the left side lane.
- `R`: arm the lamps and place an already launched ball in the right side lane.

Select Party Land, start a game, launch normally, then press `L` or `R` once.
Placement is ignored in the chute, while the ball is held or after a drain.
Watch the heart, loss text, bonus and extra-ball/new-ball presentation.
No matrix command, award, drain, delay or new-ball handler is called by the
test keys. Initial placement is followed by ordinary physics and table rules.

The diagnostic EXE writes `MATRIX TEST` source ticks/events and matrix operations
to the existing Windows `native.log` next to its portable state or in the normal
state directory. This records which event replaced the display if the visual
problem recurs. Original codes still work on the attract screen.
