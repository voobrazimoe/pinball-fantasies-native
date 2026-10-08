import os
from pathlib import Path
import unittest
from check_demo_2b10_consumers import check, OPS

class Demo2B10Consumers(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        canonical = os.environ.get('PF_RUNTIME_DATA')
        demo = os.environ.get('PF_10MIN_DEMO_DATA')
        if not canonical or not demo:
            raise unittest.SkipTest('private inputs NOT AVAILABLE')
        try:
            import capstone
        except ImportError:
            raise unittest.SkipTest('research Capstone NOT AVAILABLE')
        cls.a = (Path(canonical)/'TABLE1.PRG').read_bytes()
        cls.b = (Path(demo)/'TABLE1.PRG').read_bytes()

    def test_admitted_bounded_consumers(self):
        self.assertTrue(check(self.a, self.b))

    def test_changed_consumers_rejected(self):
        for at in [at for at,_,_ in OPS] + [0x328f, 0x19db0+0x1413, 0x5786, 0x19db0+0x16, 0x19db0+0x16+63]:
            with self.subTest(at=hex(at)):
                b = bytearray(self.b)
                b[at] ^= 1
                with self.assertRaises((ValueError, AssertionError)):
                    check(self.a, bytes(b))

    def test_no_hi_io_or_canonical_fallback(self):
        from unittest.mock import patch
        import builtins, io
        real_open = builtins.open
        def checked_open(file, *args, **kwargs):
            self.assertFalse(str(file).lower().endswith('.hi'), 'unrequested DOS high-score IO')
            return real_open(file, *args, **kwargs)
        with patch('builtins.open', checked_open), patch('io.open', checked_open):
            self.assertTrue(check(self.a, self.b))
        root = Path(__file__).resolve().parents[1]
        consumer = (root/'internal/partyland/demo_high_score.go').read_text()
        body = consumer.split('func (d *demoCore) checkHighScorePreflight()',1)[1].split('func TestDemo',1)[0]
        for forbidden in ('g.checkHighScore(', 'g.beatHighScore(', 'g.matrixTick(', 'ReadFile(', 'WriteFile('):
            self.assertNotIn(forbidden, body)
