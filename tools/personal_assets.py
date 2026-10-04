"""Local-only compatible-data inventory shared by personal release builders."""
import hashlib
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent.parent

PERSONAL_INPUTS = (
    'INTRO.PRG', 'INTRO.MOD', 'MOD2.MOD',
    'TABLE1.PRG', 'TABLE1.MOD', 'TABLE2.PRG', 'TABLE2.MOD',
    'TABLE3.PRG', 'TABLE3.MOD', 'TABLE4.PRG', 'TABLE4.MOD',
)

def inventory(directory):
    directory = Path(directory).resolve(strict=True)
    files = PERSONAL_INPUTS + (('PINBALL.CFG',) if (directory/'PINBALL.CFG').is_file() else ())
    missing = [name for name in files if not (directory/name).is_file()]
    if missing:
        raise SystemExit('Missing personal runtime input: '+', '.join(missing))
    records = []
    for name in files:
        data = (directory/name).read_bytes()
        sha = hashlib.sha256(data).hexdigest()
        records.append({'name': name, 'bytes': len(data), 'sha256': sha})
    # Decode required assets and factory scores without consulting local saves.
    subprocess.run([str(ROOT/'tools/go.sh'), 'run', './cmd/personalvalidate',
                    '-data-dir', str(directory)], cwd=ROOT, check=True)
    return directory, records
