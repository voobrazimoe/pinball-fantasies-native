import hashlib
import json
import tempfile
import unittest
from pathlib import Path
from PIL import Image
from matrix_oracle_compare import compare

class OrderedOracleTrace(unittest.TestCase):
    def run_trace(self, states, order):
        with tempfile.TemporaryDirectory() as folder:
            root = Path(folder)
            hashes = {}
            for dot, phase, tick in states:
                dots = bytearray(2560)
                dots[dot] = 1
                hashes.setdefault(hashlib.sha256(dots).hexdigest(), []).append(
                    dict(initial_phase=phase, tick=tick, op='_SCROLL'))
            (root/'states.json').write_text(json.dumps(hashes))
            for i, dot in enumerate(order):
                frame = Image.new('RGB', (640, 33))
                frame.putpixel((4*(dot%160), 2+2*(dot//160)), (255, 0, 0))
                frame.save(root/f'trace-{i:04}.png')
            return compare(root/'states.json', root, ['trace'])['trace']

    def test_monotonic_same_phase_path(self):
        result = self.run_trace([(0, 2, 1), (1, 2, 4), (2, 2, 8)], [0, 1, 1, 2])
        self.assertIsNone(result['ordered']['first_incompatible_frame'])
        self.assertEqual(result['ordered']['compatible_initial_phases'], [2])

    def test_isolated_membership_cannot_hide_reversed_trace(self):
        result = self.run_trace([(0, 2, 1), (1, 2, 4)], [1, 0])
        self.assertEqual(result['counts']['unmatched'], 0)
        self.assertEqual(result['ordered']['first_incompatible_frame'], 'trace-0001.png')

    def test_phases_cannot_change_between_sampled_frames(self):
        result = self.run_trace([(0, 2, 1), (1, 4, 2)], [0, 1])
        self.assertEqual(result['ordered']['first_incompatible_frame'], 'trace-0001.png')

if __name__ == '__main__':
    unittest.main()
