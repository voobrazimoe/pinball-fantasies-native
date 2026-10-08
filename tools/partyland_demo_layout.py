#!/usr/bin/env python3
"""Generate the Party Land 10-minute demo layout descriptor.

The descriptor records reviewed offsets, sizes, typed jingle values and
SHA-256 identities only. It never contains original bytes. Inputs are the
owner-local DMO0 evidence JSON (see docs/runtime-layout-demo-10min-validation.md)
and the legally supplied demo directory, which is read to size the playfield
FORM records and to verify every mapped copy before anything is written.
"""
import argparse
import hashlib
import json
import pathlib
import struct
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
PROFILES = ROOT / "internal/datalayout/profiles.json"
OUTPUT = ROOT / "internal/datalayout/partyland_demo.json"
PROFILE = "dos-partyland-10min-demo-v1"
PICTURE_SHIFT = 368  # four complete playfield FORMs, see the validation report
DS = 0x19DB0
# Demo-only records consumed by the native demo rules. Offsets are file offsets.
RECORDS = {
    "PLAYERSTEXT": (DS + 0x2379, 12),        # duration label, zero terminated
    "EXPIRY_TEXT1": (DS + 7313, 99),         # scroll, 255 terminated
    "EXPIRY_TEXT2": (DS + 7412, 89),
}
# Demo-only INTRO records: SHOWTEXT pages (12 zero-terminated rows) and the
# selector/options sidebar record (two 120-byte blocks of 12-column rows).
INTRO_PAGES = {"WELCOME_PAGE": 0x580D, "AVAILABLE_PAGE": 0x5869}
INTRO_SIDEBAR = (0x396D3, 240)
# Launcher zero-table-selection exit: DOS '$'-terminated closing message.
LAUNCHER_MESSAGE = 0x43F


def sha(b):
    return hashlib.sha256(b).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--demo", required=True, help="directory with the supplied demo files")
    ap.add_argument("--evidence", required=True, help="owner-local pf-dmo0-evidence.json")
    ap.add_argument("--output", default=str(OUTPUT))
    args = ap.parse_args()

    evidence = json.loads(pathlib.Path(args.evidence).read_text())
    demo_dir = pathlib.Path(args.demo)
    table = (demo_dir / "TABLE1.PRG").read_bytes()
    intro = (demo_dir / "INTRO.PRG").read_bytes()
    for f in evidence["files"]:
        data = (demo_dir / f["name"]).read_bytes()
        if len(data) != f["size"] or sha(data) != f["sha256"]:
            sys.exit(f"{f['name']}: not the reviewed demo input")
    profiles = json.loads(PROFILES.read_text())
    decoded_size = 536822  # shared canonical decoded address space

    copies = []
    for r in evidence["table1_read_map"]:
        if r.get("source_candidate") is None:
            sys.exit(f"unmapped read-map entry {r['purpose']}")
        copies.append([r["canonical_offset"], r["source_candidate"], r["size"]])
    for p in profiles["TABLE1.PRG"]["pictures"]:
        src = p["offset"] + PICTURE_SHIFT
        if table[src:src + 4] != b"FORM":
            sys.exit(f"missing playfield FORM at {src:#x}")
        n = struct.unpack(">I", table[src + 4:src + 8])[0] + 8
        copies.append([p["offset"], src, n])
    # Merge only contiguous records with the same relocation, in read order.
    merged = []
    for d, s, n in sorted(copies, key=lambda c: (c[0] - c[1], c[0])):
        if merged and merged[-1][0] - merged[-1][1] == d - s and d <= merged[-1][0] + merged[-1][2]:
            last = merged[-1]
            last[2] = max(last[2], d + n - last[0])
        else:
            merged.append([d, s, n])
    merged.sort()
    for d, s, n in merged:
        if s < 0 or s + n > len(table) or d < 0 or d + n > decoded_size:
            sys.exit(f"copy out of bounds {d:#x}")

    jingles = []
    for j in evidence["jingles"]:
        got = tuple(table[j["source"]:j["source"] + 3])
        if got != (j["position"], j["repeat"], j["priority"]):
            sys.exit(f"jingle {j['role']} differs from evidence")
        jingles.append({"role": j["role"], "destination": j["canonical_offset"], "source": j["source"],
                        "position": j["position"], "repeat": j["repeat"], "priority": j["priority"]})

    reviewed = []
    for r in evidence["reviewed_code_ranges"]:
        if sha(table[r["offset"]:r["offset"] + r["size"]]) != r["sha256"]:
            sys.exit(f"reviewed range {r['meaning']} differs")
        reviewed.append({"offset": r["offset"], "size": r["size"], "purpose": r["meaning"], "sha256": r["sha256"]})

    records = {}
    for name, (at, n) in RECORDS.items():
        records[name] = {"source": at, "size": n, "sha256": sha(table[at:at + n])}

    intro_regions, pictures = [], []
    for f in evidence["intro_forms"]:
        if sha(intro[f["offset"]:f["offset"] + f["size"]]) != f["sha256"]:
            sys.exit(f"INTRO {f['role']} differs")
        intro_regions.append({"offset": f["offset"], "size": f["size"], "purpose": "FORM " + f["role"], "sha256": f["sha256"]})
        pictures.append({"offset": f["offset"], "kind": f["kind"].ljust(4), "width": f["width"], "height": f["height"], "planes": f["planes"]})
    intro_records = {}
    for name, at in INTRO_PAGES.items():
        end = at
        for _ in range(12):
            end = intro.index(0, end) + 1
        intro_records[name] = {"source": at, "size": end - at, "sha256": sha(intro[at:end])}
    at, n = INTRO_SIDEBAR
    intro_records["SIDEBAR"] = {"source": at, "size": n, "sha256": sha(intro[at:at + n])}

    launcher = (demo_dir / "PINBALL.EXE").read_bytes()
    for f in evidence["support_research_inputs"]:
        if f["name"] == "PINBALL.EXE" and (len(launcher) != f["size"] or sha(launcher) != f["sha256"]):
            sys.exit("PINBALL.EXE: not the reviewed demo launcher")
    end = launcher.index(b"$", LAUNCHER_MESSAGE) + 1
    launcher_records = {"CLOSING_MESSAGE": {"source": LAUNCHER_MESSAGE, "size": end - LAUNCHER_MESSAGE,
                                            "sha256": sha(launcher[LAUNCHER_MESSAGE:end])}}

    out = {
        "profile": PROFILE,
        "intro": {"regions": intro_regions, "pictures": pictures, "records": intro_records},
        "launcher": {"records": launcher_records},
        "table1": {"decoded_size": decoded_size,
                   "copies": [{"destination": d, "source": s, "size": n} for d, s, n in merged],
                   "jingles": jingles, "reviewed": reviewed, "records": records},
    }
    pathlib.Path(args.output).write_text(json.dumps(out, indent=1) + "\n")
    print(f"{len(merged)} copies, {len(jingles)} jingles, {len(reviewed)} reviewed ranges -> {args.output}")


if __name__ == "__main__":
    main()
