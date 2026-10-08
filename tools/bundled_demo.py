"""The official 10-minute Party Land demo that public builds ship (internal/demodata)."""
import hashlib
from pathlib import Path

ROOT=Path(__file__).resolve().parent.parent
DIRECTORY=ROOT/'internal/demodata/assets/demo'
DEMO={
    'INTRO.PRG':'05bdba35e0a9a31a87b944428ddad27a00ba8983e97963290ab2ee1f57910fa3',
    'INTRO.MOD':'f36beae00efec1dd9e1c4e977bea264b7ec41ab18f258528ab66577a9ec66613',
    'MOD2.MOD':'aa5003c275b494062f37f44e8c77105b8a420555f4bd6ff53d7698f89c540f21',
    'TABLE1.PRG':'44b8f4b76ee16c47cda26904681e83e8f4420974bcea249d099790bfbd7369e3',
    'TABLE1.MOD':'a0877e4372abe64b70d9e361bf257ea5a84c948771f0eace3433d5f6399060b5',
    'PINBALL.EXE':'7acc8be42f23cc56a66ce8839a561c4d7318025dc395935be44b7a760f538dff',
}

def demo_bytes():
    """Exact bundled demo contents; packages may carry these and nothing else original."""
    found={p.name:p.read_bytes() for p in DIRECTORY.iterdir()}
    assert set(found)==set(DEMO),set(found)^set(DEMO)
    for name,data in found.items():
        assert hashlib.sha256(data).hexdigest()==DEMO[name],name
    return found
