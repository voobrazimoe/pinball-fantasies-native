"""Fixture-free independent linked SI+20 cadence, including second-subcall FF."""
import unittest


def scroll(extent, phase):
    text = bytearray(extent + 1)
    text[-1] = 255
    si = visits = 0
    while True:
        visits += 1
        for subcall in range(2):
            if text[si + 20] == 255:
                return visits, phase, subcall
            phase -= 1
            if phase == 0:
                phase = 8
                si += 1


class ExtentTests(unittest.TestCase):
    def test_extra_cell_cost_all_phases(self):
        for phase in range(1, 9):
            for extent in (87, 97):
                old = scroll(extent, phase)
                new = scroll(extent + 1, phase)
                self.assertEqual(new[0] - old[0], 4)
                self.assertEqual(new[1:], old[1:])

    def test_real_extents(self):
        self.assertEqual(scroll(98, 8), (313, 8, 0))
        self.assertEqual(scroll(88, 8), (273, 8, 0))
        self.assertEqual(scroll(98, 4), (311, 8, 0))
        self.assertEqual(scroll(98, 1), (309, 8, 1))


if __name__ == '__main__':
    unittest.main()
