#!/usr/bin/env python3
"""Bounded, allowlisted local A6 metadata collection; never reads game data."""
import argparse
import collections
import hashlib
import json
from pathlib import Path
import re
import subprocess
import threading
import time
import zipfile

PACKAGE = "io.github.voobrazimoe.pinballfantasies"
COMPONENT = PACKAGE + "/.PinballActivity"
ROOT = Path(__file__).resolve().parent.parent
COMMERCIAL = re.compile(r"(?:^|/)(?:INTRO\.(?:PRG|MOD)|MOD2\.MOD|TABLE.*\.(?:PRG|MOD|HI)|PINBALL\.CFG)$", re.I)
EVENT = re.compile(r"\b(A[1-5]_[A-Z_]+)\b")
NUMERIC = re.compile(r"\b([a-zA-Z_]+)=(-?\d+(?:\.\d+)?)\b")


class Device:
    def __init__(self, serial):
        self.prefix = ["adb", "-s", serial]

    def run(self, *args, optional=False):
        try:
            result = subprocess.run(self.prefix + list(args), capture_output=True,
                                    text=True, timeout=20)
            if result.returncode:
                raise RuntimeError("adb command failed: " + args[0])
            return result.stdout.strip()
        except (subprocess.TimeoutExpired, RuntimeError):
            if optional:
                return ""
            raise

    def prop(self, name):
        return self.run("shell", "getprop", name)

    def pid(self):
        value = self.run("shell", "pidof", PACKAGE, optional=True)
        return value if re.fullmatch(r"\d+", value) else ""

    def preflight(self):
        metadata = {key: self.prop(key) for key in (
            "ro.product.manufacturer", "ro.product.model", "ro.product.cpu.abilist",
            "ro.build.version.release", "ro.build.version.sdk", "ro.kernel.qemu",
            "ro.boot.qemu", "ro.hardware", "dalvik.vm.usejit", "dalvik.vm.extra-opts",
            "dalvik.vm.dex2oat-flags")}
        emulator = (self.prefix[2].startswith("emulator-") or
                    metadata["ro.kernel.qemu"] == "1" or metadata["ro.boot.qemu"] == "1" or
                    metadata["ro.hardware"] in ("goldfish", "ranchu") or
                    re.search(r"sdk|emulator|generic", metadata["ro.product.model"], re.I))
        if emulator:
            raise RuntimeError("Emulator rejected: A6 requires owner-confirmed physical hardware")
        if (metadata["dalvik.vm.usejit"].lower() == "false" or
                "-Xint" in metadata["dalvik.vm.extra-opts"] or
                "--compiler-filter=interpret-only" in metadata["dalvik.vm.dex2oat-flags"]):
            raise RuntimeError("Interpreted/disabled-JIT runtime rejected; do not change device settings with this helper")
        metadata["page_size"] = self.run("shell", "getconf", "PAGE_SIZE")
        return metadata


class Logs:
    """Never persists raw logcat, provider errors, artwork, payloads or stacks.

    Pinball tags retain event names and numeric counters only. Crash blocks for
    this package retain a marker only, not arbitrary app/system message bodies.
    """
    def __init__(self):
        self.events = collections.Counter()
        self.counters = collections.deque(maxlen=120)
        self.crashes = 0
        self.pending_crash = False
        self.lines = 0

    def accept(self, line):
        self.lines += 1
        if "AndroidRuntime" in line:
            if "FATAL EXCEPTION" in line:
                self.pending_crash = True
            if "Process:" in line:
                if self.pending_crash and re.search(r"Process: " + re.escape(PACKAGE) + r"(?:,|\s|$)", line):
                    self.crashes += 1
                self.pending_crash = False
        if "DEBUG" in line and re.search(r"(?:Cmdline: |>>> )" + re.escape(PACKAGE) + r"(?:\s|$)", line):
            self.crashes += 1
        if "PinballFantasies" not in line:
            return
        match = EVENT.search(line)
        if not match:
            return
        event = match[1]
        self.events[event] += 1
        if event in ("A4_AUDIO", "A4_STREAM_OPEN", "A5_STREAM_TRANSITION", "A5_FOCUS", "A5_DEVICE"):
            self.counters.append({"event": event, **dict(NUMERIC.findall(line))})


def output_path(value):
    path = Path(value).expanduser().resolve()
    if path == ROOT or ROOT in path.parents:
        raise ValueError("Evidence must be outside the repository")
    if path.exists():
        raise ValueError("Use a fresh evidence directory; refusing to overwrite")
    return path


def inspect_apk(value):
    path = Path(value).expanduser().resolve(strict=True)
    with zipfile.ZipFile(path) as archive:
        names = archive.namelist()
        if any(COMMERCIAL.search(name) for name in names):
            raise ValueError("APK contains a commercial-input filename; refusing installation")
        for abi in ("arm64-v8a", "x86_64"):
            for library in ("libpfengine.so", "libpinball_android.so"):
                if f"lib/{abi}/{library}" not in names:
                    raise ValueError("Expected both packaged native ABIs")
    return {"path": str(path), "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}


def snapshot(device):
    pid = device.pid()
    package = device.run("shell", "dumpsys", "package", PACKAGE, optional=True)
    # Only version fields; never persist the full dumpsys (URIs/provider names).
    version = re.findall(r"\b(?:versionCode|versionName|targetSdk|minSdk)=[\w.+-]+", package)
    mem = device.run("shell", "dumpsys", "meminfo", "-s", PACKAGE, optional=True)
    memory = [line.strip() for line in mem.splitlines()
              if re.fullmatch(r"\s*(?:TOTAL(?: PSS| RSS| SWAP PSS)?|Native Heap|Dalvik Heap|Java Heap)\s*:?[\d\s]+", line)]
    cpu = device.run("shell", "dumpsys", "cpuinfo", optional=True)
    cpu_samples = re.findall(r"([\d.]+)%\s+\d+/" + re.escape(PACKAGE) + r"\s*:", cpu)
    window = device.run("shell", "dumpsys", "window", "displays", optional=True)
    orientation = re.findall(r"\b(?:rotation|mRotation|orientation|mLastOrientation)\s*[=:]\s*(?:ROTATION_)?[0-9]+", window)
    audio = device.run("shell", "dumpsys", "audio", optional=True)
    # Numeric/enum metadata only. Device addresses/names/media titles omitted.
    audio_types = sorted(set(re.findall(r"\b(?:AUDIO_DEVICE_OUT_|TYPE_)[A-Z0-9_]+\b", audio)))
    maps = ""
    if pid:
        maps = device.run("shell", "run-as", PACKAGE, "cat", f"/proc/{pid}/maps", optional=True)
    libraries = sorted(set(re.findall(r"/(lib(?:pfengine|pinball_android|art|oboe|c\+\+_shared)\.so)\b", maps)))
    gfx = device.run("shell", "dumpsys", "gfxinfo", PACKAGE, optional=True)
    gfx_summary = [line.strip() for line in gfx.splitlines() if re.fullmatch(
        r"\s*(?:Total frames rendered: \d+|Janky frames: \d+ \([\d.]+%\)|\d+th percentile: \d+ms)\s*", line)]
    return {"monotonic_seconds": round(time.monotonic(), 3), "pid": pid,
            "package_version": version, "memory_summary": memory, "cpu_percent": cpu_samples,
            "orientation": orientation, "audio_type_enums": audio_types,
            "native_libraries": libraries, "jit_mapping_seen": "jit-code-cache" in maps,
            "gfxinfo_summary": gfx_summary}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--serial", required=True, help="adb serial of the owner's physical device")
    parser.add_argument("--output", required=True, help="fresh local directory OUTSIDE this repository")
    parser.add_argument("--mode", choices=("snapshot", "session", "home-return", "restart"), default="snapshot",
                        help="session/restart force-stop and relaunch; home-return sends Home then resumes")
    parser.add_argument("--seconds", type=int, default=60, help="bounded observation window, 5..600 seconds")
    parser.add_argument("--apk", help="optional adb install -r before session; never uninstalls/clears data")
    args = parser.parse_args()
    diagnostics = __import__("os").environ.get("PF_DIAGNOSTICS", "0") == "1"
    if not re.fullmatch(r"[a-zA-Z0-9_.:-]+", args.serial):
        parser.error("invalid serial")
    if not 5 <= args.seconds <= 600:
        parser.error("--seconds must be 5..600")
    if (args.apk or diagnostics) and args.mode not in ("session", "restart"):
        parser.error("install/diagnostics require an explicit fresh session or restart")
    out = output_path(args.output)
    apk = inspect_apk(args.apk) if args.apk else None
    device = Device(args.serial)
    metadata = device.preflight()
    if apk:
        device.run("install", "-r", apk["path"])
    if not device.run("shell", "pm", "path", PACKAGE, optional=True).startswith("package:"):
        raise RuntimeError("Package not installed")
    out.mkdir(parents=True, mode=0o700)
    report = {"A6": "NOT TESTED", "mode": args.mode, "diagnostics_requested": diagnostics,
              "metadata": metadata, "apk": apk,
              "automated": {"collection_complete": "NOT TESTED"}, "samples": [],
              "manual": "All owner-observed checks remain NOT TESTED. Use docs/android-a6-acceptance.md."}
    logs = Logs()
    process = None
    reader = None
    try:
        # adb shell joins remote arguments. Keep the date format space-free;
        # logcat's own adb command accepts the resulting timestamp argument.
        since = device.run("shell", "date", "+%m-%dT%H:%M:%S.000").replace("T", " ")
        if not re.fullmatch(r"\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.000", since):
            raise RuntimeError("Cannot establish a device logcat time boundary")
        report["logcat_since_device_time"] = since
        process = subprocess.Popen(device.prefix + ["logcat", "-T", since, "-v", "threadtime", "-s",
                                   "PinballFantasies:I", "AndroidRuntime:E", "DEBUG:F", "*:S"],
                                   stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True)
        def read_logs():
            for line in process.stdout:
                logs.accept(line)
        reader = threading.Thread(target=read_logs, daemon=True)
        reader.start()
        before = device.pid()
        if args.mode in ("session", "restart"):
            device.run("shell", "am", "force-stop", PACKAGE)
            device.run("shell", "am", "start", "-W", "-n", COMPONENT,
                       "--ez", "PF_DIAGNOSTICS", "true" if diagnostics else "false")
        elif args.mode == "home-return":
            if not before:
                raise RuntimeError("home-return requires an already-running game")
            device.run("shell", "input", "keyevent", "KEYCODE_HOME")
            time.sleep(2)
            device.run("shell", "am", "start", "-W", "-n", COMPONENT)
        print(f"Observe/play for {args.seconds}s. Automated collection cannot certify gameplay or A6.", flush=True)
        deadline = time.monotonic() + args.seconds
        while True:
            report["samples"].append(snapshot(device))
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                break
            time.sleep(min(10, remaining))
        after = device.pid()
        report["automated"]["collection_complete"] = "PASS"
        report["automated"]["process_alive_at_end"] = "PASS" if after else "FAIL"
        if args.mode == "home-return":
            report["automated"]["same_pid_home_return"] = "PASS" if before == after else "FAIL"
        if args.mode == "restart":
            report["automated"]["fresh_pid_after_force_stop"] = "PASS" if after and after != before else "FAIL"
    finally:
        if process:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
        if reader:
            reader.join(timeout=5)
        available = bool(process and logs.lines and process.returncode in (0, -15))
        report["automated"]["bounded_crash_scan"] = ("FAIL" if logs.crashes else "PASS" if available else "NOT TESTED")
        report["log_events"] = dict(logs.events)
        report["audio_counter_samples"] = list(logs.counters)
        report["crash_markers"] = logs.crashes
        report["limits"] = "Crash scan covers this window only. Missing/denied metadata is NOT TESTED. gfxinfo may omit native GLES frames. Properties show no forced interpreter; owner must confirm physical hardware and normal ART. No original-data or State content was read."
        (out / "metadata.json").write_text(json.dumps(report, indent=2) + "\n")
        (out / "README.txt").write_text("A6 NOT TESTED: owner must complete the manual report.\nLocal metadata only; do not upload original-backed storage, screenshots or raw logs.\n")
    print(f"Local evidence: {out}/metadata.json; A6 remains NOT TESTED")
    return 1 if "FAIL" in report["automated"].values() else 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, RuntimeError, subprocess.TimeoutExpired, zipfile.BadZipFile) as failure:
        # Avoid leaking arbitrary provider/device stderr into evidence.
        raise SystemExit(f"A6 helper stopped ({type(failure).__name__}): {failure}")
