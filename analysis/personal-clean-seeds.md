# Personal release: clean factory high scores

Rebuilt and validated 2026-10-02 from `57d6e3e4acc044ea8d3853f1a68fe2283208716d`,
including the single/multiplayer pre-launch matrix fix. Personal builds require exactly 12 original inputs. No
binary high-score files are bundled, tracked or copied into either artifact.
PINBALL.CFG input bytes and behavior are unchanged.

```text
INTRO.PRG
INTRO.MOD
MOD2.MOD
TABLE1.PRG
TABLE1.MOD
TABLE2.PRG
TABLE2.MOD
TABLE3.PRG
TABLE3.MOD
TABLE4.PRG
TABLE4.MOD
PINBALL.CFG
```

The Windows embedded ZIP and Linux AppImage data directory were inspected:
exactly these names, matching each input SHA256, with no `.HI` payload entries.
Both ignored build manifests contain the same 12 input records.

| Artifact | SHA256 |
| --- | --- |
| `release/personal/windows/pinballfantasies.exe` | `7720508f43bd523989dc141d54b8b8d1c4b6c579196347eb12e4b5ba0a496d26` |
| `release/personal/linux/PinballFantasies-x86_64.AppImage` | `f56a89e5aeb719544aef2f1d77999543209396aaa83234d0fab7c7ae19a484c0` |

`FileStore.Load` reads mutable userdata first, then an explicitly configured
optional data seed. When neither exists it returns `frontend.Defaults(table)`.
An empty SeedDirectory no longer implicitly reads scores from the working
directory. Save creates the mutable file only in the selected state directory.

`TestDefaultScoresSourceFidelity` checks all four serialized factory records
against `analysis/game-inventory.json`. TABLE3/4 inventory hashes formerly
represented saved EMH scores; they now describe pristine
INTRO.ASM defaults. The inventory generator reconstructs all four seed hashes
from the original source declarations, without reading or distributing local
binary `.HI` files. Defaults themselves were already source-faithful and did
not require changes. Original local score files remain unchanged after validation.

Validation passed:

- Full `./tools/go.sh test -p=1 -count=1 ./...`, without score hash exceptions.
- Fresh state, save/reload, deletion/reset and optional seed tests for all tables.
- Working-directory isolation when no seed directory is configured. The personal
  validator uses isolated writable and empty seed directories, validates factory
  records plus save/restart/reset for all four tables, and passes with malformed
  scores in the input/CWD as well as with no score files in either fixture.
- `python3 tools/test_personal_assets.py`: missing and malformed/changed local
  score files do not change required input records or validation.
- Reachability, operand and numeric matrix audits with `--check`: all pass.
- Party oracle: 1200 ticks, score 2,300,000, ball 2; established frame hash retained.
- Linux and Windows/Wine artifact storage smoke: artifact alone in empty
  Unicode/space paths; bundled data, userdata, explicit state overrides and
  fallback configuration directory all pass. Storage checks include a three-second live window on each backend.
- Three complete builds from the installation directory: baseline, all four
  TABLE*.HI replaced with malformed bytes, and all four removed. Both artifacts
  have byte-identical SHA256 in all three cases. Installation score bytes/modes
  are restored after the test. Reproduce with `tools/test_personal_reproducibility.py`.
- Linux and Wine live personal smoke: all four tables, settings restart, Unicode
  paths, external userdata, explicit data overrides and unchanged artifact hashes.
- Windows/Wine factory-score, persistence/reset and optional-seed tests pass.
- Case-insensitive tracked-tree search for the removed unused legacy file:
  zero matches. Historical external source repositories are unchanged.

Local logs and manifests remain ignored under `.build-personal/`:
`cleanup-go-suite.log`, `cleanup-builder-tests.log`, `cleanup-schema-tests.log`,
`cleanup-{reachability,operands,numeric,party-oracle,wine-oracle,wine-scores}.log`,
`cleanup-storage.log`, `cleanup-linux-smoke.log`, `cleanup-wine-smoke.log`,
`cleanup-build-{baseline,replaced-hi,removed-hi}.log`,
`cleanup-reproducibility.json`, `build-manifest.json` and
`linux-build-manifest.json`. Historical reports link here for current results;
archived raw logs remain historical evidence.
