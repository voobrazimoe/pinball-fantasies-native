import os
from pathlib import Path
import unittest
from check_demo_2b9_consumers import check

class Demo2B9Consumers(unittest.TestCase):
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
        for at in (0x1ab45,0x1651,0x1656,0x165d,0x166b,0x166f,
                   0x1674,0x167a,0x1684,0x1689,0x1692,0x1694,0x169c,
                   0x169f,0x16a2,0x16ab,0x16b0,0x5ac8,0x5aca,0x5acf,
                   0x5ad3,0x5ad6,0x19db0+0xc79,0x19db0+0xc7a,0x19db0+0xc7c):
            with self.subTest(at=hex(at)):
                b = bytearray(self.b)
                b[at] ^= 1
                with self.assertRaises((ValueError, AssertionError)):
                    check(self.a, bytes(b))
