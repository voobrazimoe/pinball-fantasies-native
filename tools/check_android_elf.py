#!/usr/bin/env python3
import argparse
import re
import subprocess
import sys
from pathlib import Path

MIN_PAGE_ALIGN = 16 * 1024


def run_tool(tool: str, *args: str) -> str:
    proc = subprocess.run([tool, *args], check=False, text=True, capture_output=True)
    if proc.returncode != 0:
        sys.stderr.write(proc.stderr)
        raise SystemExit(f"{tool} failed with exit code {proc.returncode}")
    return proc.stdout


def parse_load_alignments(text: str) -> list[int]:
    alignments: list[int] = []
    for line in text.splitlines():
        if not re.match(r"^\s*LOAD\b", line):
            continue
        match = re.search(r"\balign\s+2\*\*(\d+)\s*$", line)
        if match:
            alignments.append(1 << int(match.group(1)))
            continue
        match = re.search(r"\s(0x[0-9a-fA-F]+)\s*$", line)
        if match:
            alignments.append(int(match.group(1), 16))
    return alignments


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Check Android ELF load alignment for 16 KB page-size support."
    )
    parser.add_argument("library", type=Path)
    parser.add_argument(
        "--objdump", default="llvm-objdump", help="llvm-objdump path from the Android NDK"
    )
    args = parser.parse_args()

    if not args.library.is_file():
        raise SystemExit(f"missing library: {args.library}")

    output = run_tool(args.objdump, "-p", str(args.library))
    alignments = parse_load_alignments(output)
    if not alignments:
        raise SystemExit("no ELF LOAD segments found")

    if any(value < MIN_PAGE_ALIGN for value in alignments):
        values = ", ".join(str(value) for value in alignments)
        raise SystemExit(f"16 KB alignment check failed: LOAD alignments are {values}")

    values = ", ".join(str(value) for value in alignments)
    print(f"OK: {args.library} LOAD alignments: {values}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
