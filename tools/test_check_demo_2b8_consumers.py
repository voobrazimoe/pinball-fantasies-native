import os
from pathlib import Path
import unittest
from check_demo_2b8_consumers import check

class Demo2B8Consumers(unittest.TestCase):
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
        for at in (0x28c8, 0x28db, 0x19be, 0x1bb5, 0x27f8,
                   0x27c1, 0x27de, 0x19db0+0xba8+10, 0x1b25d, 0x2f51, 0x5044, 0x1e2b9, 0x19db0+0xc61):
            with self.subTest(at=hex(at)):
                b = bytearray(self.b)
                b[at] ^= 1
                with self.assertRaises((ValueError, AssertionError)):
                    check(self.a, bytes(b))
