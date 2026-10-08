import os
from pathlib import Path
import unittest
from unittest.mock import patch
from check_demo_2b12_consumers import check, OPS, PROGRAM, REFS

class Demo2B12Consumers(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not os.getenv('PF_RUNTIME_DATA') or not os.getenv('PF_10MIN_DEMO_DATA'):
            raise unittest.SkipTest('private inputs NOT AVAILABLE')
        try: import capstone
        except ImportError: raise unittest.SkipTest('Capstone NOT AVAILABLE')
        cls.a=(Path(os.environ['PF_RUNTIME_DATA'])/'TABLE1.PRG').read_bytes()
        cls.b=(Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes()
    def test_linked_consumers_no_hi_io(self):
        read=Path.read_bytes
        def guarded(p,*args,**kwargs):
            self.assertNotEqual(p.suffix.upper(),'.HI')
            return read(p,*args,**kwargs)
        with patch.object(Path,'read_bytes',guarded): self.assertTrue(check(self.a,self.b))
    def test_mutations(self):
        sites=[at for at,_,_ in OPS]
        at=0x1b6ac
        for op,args in PROGRAM:
            sites.extend(range(at,at+2*(1+len(args)),2));at+=2*(1+len(args))
        sites.extend(0x19db0+v for k,v in REFS.items() if k not in ('JACKVALUE','PLAYERSTEXT','DEMO_BALLSTEXT'))
        for at in sites:
            with self.subTest(site=hex(at)):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises((ValueError,AssertionError)):check(self.a,bytes(b))
    def test_no_fallback(self):
        root=Path(__file__).resolve().parents[1]
        s=(root/'internal/partyland/demo_info.go').read_text()
        for forbidden in ('g.Sync(', 'g.startMatrix(', 'g.beginMatrix(', 'g.checkHighScore(', 'StartCheat(', 'ReadFile(', '.HI', 'highScore =', 'SetBall(', 'SpringValid =', 'expired = false'):
            self.assertNotIn(forbidden,s)
        self.assertIn('g.matrixDispatch()',s)
        self.assertIn('verified-native-factory-volatile',s)
