#!/usr/bin/env python3
"""Asset-free helper safety/collection tests. These NEVER certify physical A6."""
import contextlib
import io
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile

import android_a6_device as a6


class FakeDevice(a6.Device):
    def __init__(self, serial="physical", props=None):
        super().__init__(serial)
        self.props = props or {}
        self.calls = []
        self.current_pid = "123"

    def run(self, *args, optional=False):
        self.calls.append(args)
        if args[:2] == ("shell", "getprop"):
            return self.props.get(args[2], "")
        if args[:2] == ("shell", "pidof"):
            return self.current_pid
        if args[:2] == ("shell", "date"):
            assert " " not in args[2], "adb shell date format must survive remote argument joining"
            return "10-04T12:34:56.000"
        if args[:3] == ("shell", "am", "force-stop"):
            self.current_pid = ""
        if args[:3] == ("shell", "am", "start"):
            self.current_pid = self.current_pid or "456"
        if args[:3] == ("shell", "pm", "path"):
            return "package:/data/app/base.apk"
        if "package" in args:
            return "versionCode=1 versionName=0.0.0-a4 targetSdk=36 minSdk=27\nsecret provider URI"
        if "meminfo" in args:
            return "TOTAL PSS: 12345\nNative Heap: 678\nsecret contents of INTRO.PRG"
        if "cpuinfo" in args:
            return "5.4% 123/" + a6.PACKAGE + ": 4% user\nother app secret"
        if "displays" in args:
            return "rotation=1 source-provider-secret"
        if "audio" in args:
            return "TYPE_BUILTIN_SPEAKER secret bluetooth address and name"
        if "run-as" in args:
            return "/data/app/lib/libpfengine.so\nmemfd:jit-code-cache\n/data/user/0/private-source-secret"
        return ""


class FakeLogcat:
    def __init__(self, *args, **kwargs):
        self.stdout = io.StringIO("historical line\nPinballFantasies: A4_AUDIO underruns=2 high=480 dropped=0 SECRET-PAYLOAD\n")
        self.returncode = None

    def terminate(self):
        self.returncode = -15

    def wait(self, timeout):
        return self.returncode


class SafetyTests(unittest.TestCase):
    def test_refuses_emulators_and_interpreted_runtime_without_mutation(self):
        for serial, props in (
            ("emulator-5554", {}), ("physical", {"ro.kernel.qemu": "1"}),
            ("physical", {"ro.boot.qemu": "1"}), ("physical", {"ro.hardware": "ranchu"}),
            ("physical", {"ro.product.model": "sdk_gphone64_arm64"}),
            ("physical", {"dalvik.vm.extra-opts": "-Xint -verbose:gc"}),
            ("physical", {"dalvik.vm.usejit": "false"}),
            ("physical", {"dalvik.vm.dex2oat-flags": "--compiler-filter=interpret-only"}),
        ):
            device = FakeDevice(serial, props)
            with self.assertRaises(RuntimeError):
                device.preflight()
            self.assertTrue(all(call[:2] == ("shell", "getprop") for call in device.calls))

    def test_output_repo_symlink_and_overwrite_guards(self):
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(ValueError):
                a6.output_path(directory)
            alias = Path(directory) / "repo-alias"
            alias.symlink_to(a6.ROOT, target_is_directory=True)
            for path in (a6.ROOT / "evidence", alias / "evidence"):
                with self.assertRaises(ValueError):
                    a6.output_path(path)

    def test_apk_rejects_payload_and_missing_abis(self):
        with tempfile.TemporaryDirectory() as directory:
            apk = Path(directory) / "debug.apk"
            def write(extra=None, abi_missing=False):
                with zipfile.ZipFile(apk, "w") as archive:
                    for abi in ("arm64-v8a", "x86_64"):
                        if abi_missing and abi == "arm64-v8a":
                            continue
                        for lib in ("libpfengine.so", "libpinball_android.so"):
                            archive.writestr(f"lib/{abi}/{lib}", b"invented test bytes")
                    if extra:
                        archive.writestr(extra, b"invented rejection bytes")
            write()
            self.assertEqual(len(a6.inspect_apk(apk)["sha256"]), 64)
            for name in ("assets/intro.prg", "TABLE4.MOD", "assets/table1.HI", "PINBALL.CFG"):
                write(name)
                with self.assertRaises(ValueError):
                    a6.inspect_apk(apk)
            write(abi_missing=True)
            with self.assertRaises(ValueError):
                a6.inspect_apk(apk)

    def test_logs_strip_contents_and_scope_crashes(self):
        log = a6.Logs()
        log.accept("PinballFantasies: A2_IMPORT_REJECTED source-uri SECRET-PAYLOAD")
        for _ in range(200):
            log.accept("PinballFantasies: A4_AUDIO high=4096 dropped=2 SECRET-PAYLOAD")
        log.accept("AndroidRuntime: FATAL EXCEPTION: main")
        log.accept("AndroidRuntime: Process: unrelated.app, PID: 50")
        self.assertEqual(log.crashes, 0)
        log.accept("AndroidRuntime: FATAL EXCEPTION: main")
        log.accept("AndroidRuntime: Process: " + a6.PACKAGE + ", PID: 51")
        log.accept("DEBUG: Cmdline: " + a6.PACKAGE)
        self.assertEqual(log.crashes, 2)
        self.assertEqual(len(log.counters), 120)
        self.assertNotIn("SECRET", json.dumps([log.events, list(log.counters)]))

    def test_snapshot_filters_private_and_unrelated_metadata(self):
        device = FakeDevice()
        result = a6.snapshot(device)
        self.assertEqual(result["cpu_percent"], ["5.4"])
        self.assertEqual(result["native_libraries"], ["libpfengine.so"])
        self.assertTrue(result["jit_mapping_seen"])
        self.assertNotIn("secret", json.dumps(result).lower())
        self.assertTrue(all("pull" not in call and "root" not in call for call in device.calls))
        self.assertEqual([call for call in device.calls if "run-as" in call],
                         [("shell", "run-as", a6.PACKAGE, "cat", "/proc/123/maps")])

    def test_commands_diagnostics_default_and_manual_boundary(self):
        for mode, diagnostics in (("snapshot", False), ("home-return", False),
                                  ("session", False), ("restart", True)):
            device = FakeDevice()
            with tempfile.TemporaryDirectory() as directory:
                out = Path(directory) / "new-evidence"
                argv = ["helper", "--serial", "physical", "--output", str(out),
                        "--mode", mode, "--seconds", "5"]
                clock = iter(range(0, 1000, 10))
                with patch("sys.argv", argv), patch.dict(os.environ, {"PF_DIAGNOSTICS": "1" if diagnostics else "0"}), \
                     patch.object(a6, "Device", return_value=device), \
                     patch.object(a6.subprocess, "Popen", FakeLogcat), \
                     patch.object(a6.time, "monotonic", side_effect=lambda: next(clock)), \
                     patch.object(a6.time, "sleep"), contextlib.redirect_stdout(io.StringIO()):
                    self.assertEqual(a6.main(), 0)
                report = json.loads((out / "metadata.json").read_text())
                self.assertEqual(report["A6"], "NOT TESTED")
                self.assertIn("NOT TESTED", report["manual"])
                self.assertNotIn("SECRET", json.dumps(report))
                starts = [call for call in device.calls if call[:3] == ("shell", "am", "start")]
                if mode in ("session", "restart"):
                    self.assertEqual(starts[0][-3:], ("--ez", "PF_DIAGNOSTICS", "true" if diagnostics else "false"))
                if mode == "snapshot":
                    self.assertFalse(starts)
                self.assertFalse(any("settings" in call or "pull" in call or "clear" in call for call in device.calls))


if __name__ == "__main__":
    unittest.main()
