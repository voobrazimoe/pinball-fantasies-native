import os
from pathlib import Path
import unittest
from check_demo_2b11_consumers import check, OPS

class Demo2B11Consumers(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        if not os.getenv('PF_RUNTIME_DATA') or not os.getenv('PF_10MIN_DEMO_DATA'):
            raise unittest.SkipTest('private inputs NOT AVAILABLE')
        try: import capstone
        except ImportError: raise unittest.SkipTest('Capstone NOT AVAILABLE')
        cls.a=(Path(os.environ['PF_RUNTIME_DATA'])/'TABLE1.PRG').read_bytes()
        cls.b=(Path(os.environ['PF_10MIN_DEMO_DATA'])/'TABLE1.PRG').read_bytes()
    def test_linked_consumers(self): self.assertTrue(check(self.a,self.b))
    def test_mutations(self):
        sites=[at for at,_,_ in OPS]+[0x1a485,0x1a485+26,0x1aa50,0x1aa52,0x1a9fd,0x1aa00,0x1b461+2,0x1b467+4,0x1b49f+4,0x1b4cb+4,0x1b4eb+4,0x1b50b+4,0x1b533,0x1b539,0x1b88e+6,0x1b88e+10,0x1bb7e,0x1c135+5]
        for at in sites:
            with self.subTest(site=hex(at)):
                b=bytearray(self.b);b[at]^=1
                with self.assertRaises((ValueError,AssertionError)):check(self.a,bytes(b))
    def test_no_fallback(self):
        root=Path(__file__).resolve().parents[1]
        s=(root/'internal/partyland/demo_scored_drain.go').read_text()
        s=s.split('func (d *demoConnected) scoredDrain()',1)[1].split('func TestDemo',1)[0]
        for forbidden in ('g.drain(', 'g.changeBall(', 'g.newBall(', 'g.advancePlayer(', 'g.startMatrix(', 'g.beginMatrix(', 'g.effect(', 'g.bonus(', 'g.beatHighScore('):self.assertNotIn(forbidden,s)
