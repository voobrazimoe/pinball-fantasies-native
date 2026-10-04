#!/usr/bin/env python3
"""Regressions for the Android native page-size gate (no SDK required)."""
import contextlib
import io
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import check_android_elf as gate


class AlignmentGateTests(unittest.TestCase):
    def check(self, output):
        with tempfile.TemporaryDirectory() as directory:
            library = Path(directory) / "libtest.so"
            library.touch()
            with patch("sys.argv", ["check_android_elf.py", str(library)]), \
                    patch.object(gate, "run_tool", return_value=output), \
                    contextlib.redirect_stdout(io.StringIO()):
                return gate.main()

    def test_ndk_objdump_multiline_headers(self):
        self.assertEqual(self.check("""Program Header:
    LOAD off 0x00000000 vaddr 0x00000000 paddr 0x00000000 align 2**14
         filesz 0x00002000 memsz 0x00002000 flags r-x
    LOAD off 0x00004000 vaddr 0x00004000 paddr 0x00004000 align 2**16
         filesz 0x00001000 memsz 0x00002000 flags rw-
"""), 0)

    def test_hex_alignment(self):
        self.assertEqual(self.check("  LOAD 0x0 0x0 0x0 0x20 0x20 R E 0x4000"), 0)

    def test_one_4k_segment_fails(self):
        with self.assertRaisesRegex(SystemExit, "alignment check failed"):
            self.check("LOAD align 2**14\nLOAD align 2**12")

    def test_unrecognized_segment_cannot_be_skipped(self):
        with self.assertRaisesRegex(SystemExit, "cannot parse"):
            self.check("LOAD align 2**14\nLOAD align unknown")

    def test_empty_headers_fail(self):
        with self.assertRaisesRegex(SystemExit, "no ELF LOAD"):
            self.check("Program Header:\n DYNAMIC align 2**3")

    def test_non_power_of_two_alignment_fails(self):
        with self.assertRaisesRegex(SystemExit, "alignment check failed"):
            self.check("LOAD 0x6000")


if __name__ == "__main__":
    unittest.main()
