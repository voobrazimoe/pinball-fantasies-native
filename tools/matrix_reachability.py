#!/usr/bin/env python3
"""Source-authoritative matrix-program reachability audit for PF F1/F2/F3/F4.

Authority is the original DOS ASM/data:
  * PLAND.ASM -> Party Land (TABLE1)
  * SDEV.ASM  -> Speed Devils (TABLE2)
  * SHOW.ASM  -> Billion Dollar Gameshow (TABLE3)
  * STONES.ASM -> Stones N Bones (TABLE4)

The tool is independent of the native runtime references. It:
  1. parses every ASM command word and program label in the whole file;
  2. computes source reachability from gameplay entry roots through the
     real branch graph (bounded at the next label / terminator);
  3. resolves the dynamic label families from their source state domains;
  4. emulates the native dispatch over the *registered* command stream to
     find every runtime lookup site the shipping code can actually reach;
  5. diffs source-reachable against registered and fails loudly.

Exit status is non-zero when a source-reachable matrix program is missing, or
when the native runtime can name a target that is not registered.
"""
from __future__ import annotations

import argparse
import ast
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
# Tree that owns the generated Go data files. It never follows --root, so a
# read-only audit of another checkout cannot write into it.
CHECKOUT = ROOT
CHECK = False
STALE = []

def emit(path, body):
    if CHECK:
        if not path.exists() or path.read_text() != body:
            STALE.append(str(path))
    else:
        path.write_text(body)

def current_regions():
    text = (ROOT / "tools/reference_pf8.py").read_text()
    match = re.search(r"commands=\[\];labels=\{\};a,z=(\{[^\n]+\})\[table\]", text)
    if not match:
        raise SystemExit("current PF8 matrix extraction regions not found")
    bounds = ast.literal_eval(match[1])
    return {t: (name, bounds[t]) for t, name in enumerate(("PLAND", "SDEV", "SHOW", "STONES"), 1)}

# Branch args per op: which arg index holds a matrix program label.
BRANCH_ARGS = {
    "_JMP": (0,),
    "_JBCDZ": (1,),
    "_JBONUSX1": (0,),
    "_SHOW_SCORE": (0,),
    "_LOOP_": (1,),
}

# Ops that end the program in the source/registered stream.
ZERO_OP = "0"

# Ops where the native runtime stops dispatching (returns, changes player or
# hands control to the frontend), so source fallthrough past them is not
# required to resolve at runtime.
NATIVE_STOP_OPS = {"0", "_JMP", "_CHECK_XXBALLS", "_2_DEMO_MODE",
                   "_CHANGE_PLAYER", "_TURNONTURBO"}


def load_generated(relpath: str) -> dict:
    text = (ROOT / relpath).read_text()
    m = re.search(r"const \w+ = `(.*)`\n?$", text, re.S)
    if not m:
        raise SystemExit(f"no generated content const in {relpath}")
    data = json.loads(m.group(1))
    if relpath == 'internal/presentation/content.go':
        import prg_content_layout
        data = prg_content_layout.hydrate(data, ROOT)
    elif 'jingles' in data:
        import prg_content_layout
        table = {'partyland':1, 'speeddevils':2, 'gameshow':3, 'stones':4}[relpath.split('/')[1]]
        pres = load_generated('internal/presentation/content.go')[str(table)]
        for label in data['jingles']:
            b = pres['texts'][label]
            assert len(b) == 3
            data['jingles'][label] = dict(position=b[0], repeat=b[1], priority=b[2])
        for label in data.get('animations', {}):
            a = pres['animations'][label]
            data['animations'][label] = dict(header=a['header'], frames=a['durations'])
        for label in data.get('scrolls', {}):
            data['scrolls'][label] = len(pres['texts'][label])
    return data


def normalized_lines(src: str) -> list[str]:
    if not src.endswith(".ASM"):
        src += ".ASM"
    raw = (ROOT / "reference/original-dos-source" / src).read_bytes().decode("latin1")
    out, comment, active, stack = [], False, True, []
    for l in raw.splitlines():
        t = l.strip().upper()
        if t.startswith("COMMENT\\"):
            comment = True
            out.append("")
            continue
        if comment:
            out.append("")
            if t == "\\":
                comment = False
            continue
        if t == "IF DEMOVER":
            stack.append(active); active = False; out.append(""); continue
        if t == "ELSE" and stack:
            active = stack[-1]; out.append(""); continue
        if t == "ENDIF" and stack:
            active = stack.pop(); out.append(""); continue
        out.append(l.upper() if active else "")
    return out


class Source:
    """A whole-file parse of one table's matrix command words."""

    def __init__(self, src, region):
        self.src = src
        self.region = region
        lines = normalized_lines(src)
        self.labels: dict[str, int | None] = {}
        self.commands: list[tuple[str, list[str]]] = []
        self.lines_of: list[int] = []
        cur = None
        for i, raw in enumerate(lines):
            s = raw.split(";")[0].strip()
            if not s:
                continue
            m = re.match(r"^(\w+)?\s*DW\s+(.+)$", s)
            if m:
                args = [x.strip().replace("OFFSET ", "") for x in m[2].split(",")]
                is_cmd = args[0].startswith("_") or args[0] == "0"
                if m[1] and is_cmd:
                    cur = m[1]
                    self.labels[cur] = len(self.commands)
                    self.commands.append((args[0], args[1:]))
                    self.lines_of.append(i + 1)
                elif m[1]:
                    self.labels.setdefault(m[1], None)
                elif is_cmd:
                    self.commands.append((args[0], args[1:]))
                    self.lines_of.append(i + 1)
                continue
            lab = re.match(r"^(\w+)\s+LABEL\s+\w+$", s)
            if lab:
                cur = lab[1]
                self.labels.setdefault(cur, len(self.commands))
                continue
            if re.fullmatch(r"CLEARIT[234]?", s):
                self.commands.append(
                    ("_CLEAR" + (s[-1] if s[-1].isdigit() else "?"), []))
                self.lines_of.append(i + 1)
        # Distinct program starts, used to bound each program body.
        self.starts = sorted({v for v in self.labels.values() if v is not None})
        self.starts_set = set(self.starts)

    def extent(self, name: str):
        """The command stream from a label to its terminator.

        Labels are entry markers into one shared stream, so an extent runs
        through intervening labels until a terminator; it is not clipped at the
        next label (source fallthrough across labels is real).
        """
        start = self.labels.get(name)
        if start is None:
            return None
        end = start
        while end < len(self.commands):
            end += 1
            if self.commands[end - 1][0] == ZERO_OP:
                break
        return start, end

    def line_of(self, idx):
        return self.lines_of[idx] if 0 <= idx < len(self.lines_of) else -1


# Frontend dispatches these through presentation.Content, not gameplay streams.
FRONTEND = {"URBANOVERTS", "SHOWHIGHSTS", "SHOWINFOTS", "SHOWPLAYERSTS",
            "FIRST_NO_OF_PLAYERSTS", "NO_OF_PLAYERSTS", "GAMEOVERTS",
            "ONCE_MORETS", "AFTERDEMOMODETS", "CLEARTS"}

def source_entries(src_name):
    text = "\n".join(normalized_lines(src_name))
    # Literal BX dispatch destinations, including opcode handlers and capture
    # continuations. Exclude frontend ownership; shared high-score roots below.
    roots = set(re.findall(r"MOV\s+BX,\s*OFFSET\s+(\w+TS)\b", text))
    roots |= {"BEATENTS", "BEATEN_BH_TS", "PARTY_ONTS"}
    # SHOW's BONUS_ANIMS is a pointer table, independent of runtime strings.
    for row in re.findall(r"^BONUS_ANIMS\s+DW\s+([^;\n]+)", text, re.M):
        roots.update(re.findall(r"\w+TS", row))
    roots -= FRONTEND
    # F1 ends gameplay at _CHECK_XXBALLS/_KOLLA_XXBALL, then frontend
    # handles game-over. Its renderer owns AFTER_XXBALLTS, not timing.Labels.
    if src_name in ("PLAND", "SDEV"):
        roots.discard("AFTER_XXBALLTS")
    return roots


def implicit_hops(src_name):
    """Source-proven BX hops in the table-local match handlers and shared
    FANTASIE high-score handler. No game-state execution is attempted."""
    text = "\n".join(normalized_lines(src_name))
    shared = "\n".join(normalized_lines("FANTASIE"))
    def between(text, start, end):
        return text.split(start + ":", 1)[1].split(end + ":", 1)[0]
    def targets(body):
        return sorted(set(re.findall(r"MOV\s+BX,\s*OFFSET\s+(\w+TS)\b", body)))
    hops = {"_BEATEN_MATRIX": targets(between(shared, "_BEATEN_MATRIX", "PRINT_SIFFROR5__"))}
    hops["_CHECK_XXBALLS"] = targets(between(text, "_CHECK_XXBALLS", "_KOLLA_XXBALL"))
    hops["_KOLLA_XXBALL"] = targets(between(text, "_KOLLA_XXBALL", "_CHANGE_PLAYER")) + ["SHOOT_AGAIN_ONTS"]
    if src_name == "SDEV":
        hops["_TURNONTURBO"] = ["TURBOTS"]
        hops["_RETURN_OF_THE_EVIL_SUPERMODE"] = ["BACK_2_TURBOTS"]
    return hops


def source_reachable(src: Source, roots):
    hops = implicit_hops(src.src)
    seen_cmds: set[int] = set()
    seen_labels: set[str] = set()
    missing_roots = [r for r in roots if src.labels.get(r) is None]
    queue = [r for r in roots if src.labels.get(r) is not None]
    while queue:
        name = queue.pop()
        if src.src in ("PLAND", "SDEV") and name == "AFTER_XXBALLTS":
            continue  # Explicit frontend interception in these gameplay VMs.
        ext = src.extent(name)
        if ext is None:
            continue
        seen_labels.add(name)
        i, end = ext
        while i < end:
            if i in seen_cmds:
                break
            seen_labels.update(k for k, v in src.labels.items() if v == i and k.endswith("TS"))
            seen_cmds.add(i)
            op, args = src.commands[i]
            i += 1
            for pos in BRANCH_ARGS.get(op, ()):
                if pos < len(args):
                    tgt = args[pos]
                    if src.labels.get(tgt) is None:
                        raise SystemExit(
                            f"{name}: source branch {op} -> {tgt} undefined")
                    queue.append(tgt)
            queue.extend(hops.get(op, []))
            # Shared engine hop: FANTASIE _beaten_matrix jumps to Beaten_bh_TS.
            if op == "_BEATEN_MATRIX" and src.labels.get("BEATEN_BH_TS") is not None:
                queue.append("BEATEN_BH_TS")
            # Stop exactly where the native runtime stops dispatching.
            if op in NATIVE_STOP_OPS:
                break
    return seen_labels, missing_roots


def asm_effects(src):
    """Parse the original EFFECT_STRUC blocks: NAME DW jingle / 2 BCD rows / DW matrix."""
    lines = normalized_lines(src)
    eff = {}
    for i, raw in enumerate(lines):
        s = raw.split(";")[0].strip()
        m = re.match(r"^(\w+)\s+DW\s+(\w+)$", s)
        if not m:
            continue
        rows = []
        j = i + 1
        while j < len(lines) and len(rows) < 4:
            t = lines[j].split(";")[0].strip()
            if t:
                rows.append(t)
            j += 1
        if rows and re.fullmatch(r"DB\s+\d+", rows[0]):
            rows = rows[1:]
        if len(rows) >= 3 and all(re.search(r"\bDB\s+[\d, ]+$", r) and len(re.findall(r"\d+", r.split("DB", 1)[1])) == 12 for r in rows[:2]):
            tm = re.match(r"DW\s+(\w+)", rows[2])
            if tm:
                eff[m[1]] = {"jingle": m[2], "matrix": tm[1]}
    return eff


def emulate_native(commands, labels, roots, hops):
    """Walk the registered command stream exactly as the Go dispatcher does.

    Returns the set of branch-target names the runtime can look up and the set
    of animation/scroll names it can require.
    """
    index = {k: v for k, v in labels.items()}
    seen, targets, assets, ops = set(), set(), set(), set()
    queue = [index[r] for r in roots if r in index]
    unresolved_roots = [r for r in roots if r not in index]
    while queue:
        n = queue.pop()
        while 0 <= n < len(commands):
            if n in seen:
                break
            seen.add(n)
            c = commands[n]
            op = c["op"]
            args = c.get("args", [])
            ops.add(op)
            n += 1
            if op == ZERO_OP:
                break
            if op in BRANCH_ARGS:
                for pos in BRANCH_ARGS[op]:
                    if pos < len(args):
                        targets.add(args[pos])
                        if args[pos] in index:
                            queue.append(index[args[pos]])
                if op == "_JMP":
                    break
            for t in hops.get(op, ()):
                targets.add(t)
                if t in index:
                    queue.append(index[t])
            if op in ("_ANIMATION",) and args:
                assets.add(("animation", args[0]))
            if op in ("_SCROLL",) and args:
                assets.add(("scroll", args[0]))
            if op in ("_JINGLE",) and args:
                assets.add(("jingle", args[0]))
            # Native switch cases that stop the program.
            if op in ("_CHANGE_PLAYER", "_2_DEMO_MODE", "_TURNONTURBO",
                      "_CHECK_XXBALLS", "_COUNTDOWN2"):
                break
    return targets, assets, ops, unresolved_roots


GO_CALL = re.compile(r'(?:beginMatrix|matrixJump|jump)\(\s*"([^"]+)"\s*\)')
GO_ANY_CALL = re.compile(r'beginMatrix\(([^\n]*)\)')


def native_literals(pkg_files):
    lits = set()
    for f in pkg_files:
        text = (ROOT / f).read_text()
        lits |= set(GO_CALL.findall(text))
    return sorted(lits)


def decimal_format_width():
    """Width of the native Decimal.String() formatter, read from source."""
    t = (ROOT / "internal/tablelogic/decimal.go").read_text()
    m = re.search(r'func \(d Decimal\) String\(\) string\s*\{\s*return fmt\.Sprintf\("([^"]+)"', t)
    if m:
        w = re.search(r"%0(\d+)d", m.group(1))
        if w:
            return int(w.group(1))
    return 0


def constructed_families(pkg_files, domains, decimal_width=0):
    """Find beginMatrix() calls whose argument is built at runtime.

    The value domain comes from the ASM table that bounds the selector. When
    the expression formats the selector through the native Decimal.String()
    helper, the produced names use that zero-padded width, which is exactly
    the class of bug this audit exists to catch.
    """
    fams = {}
    for f in pkg_files:
        text = (ROOT / f).read_text()
        for expr in GO_ANY_CALL.findall(text):
            expr = expr.strip()
            if expr.startswith('"') and expr.endswith('"') and "+" not in expr:
                continue  # plain literal, handled elsewhere
            parts = re.findall(r'"([^"]*)"', expr)
            if len(parts) < 2:
                continue
            prefix, suffix = parts[0], parts[-1]
            padded = decimal_width > 0 and "number(" in expr and ".String()" in expr
            for dom_prefix, values in domains.items():
                if prefix == dom_prefix:
                    val = (lambda v: f"{v:0{decimal_width}d}") if padded else str
                    tag = "zero-padded" if padded else "plain"
                    fams[f"{prefix}<{tag} domain>{suffix} ({f})"] = [
                        f"{prefix}{val(v)}{suffix}" for v in values]
    return fams


def check_table(table, native, dyn_effects, dyn_labels, pkg_files):
    src_name, region = current_regions()[table]
    src = Source(src_name, region)
    a, z = region
    registered = set(native["labels"])
    commands = native["commands"]
    labels = native["labels"]

    report = {
        "source_file": f"{src_name}.ASM",
        "region": [a, z],
        "registered": sorted(registered),
        "source_program_count": len(src.starts),
        "source_branch_undefined": [],
        "out_of_region_programs": [],
        "effect_matrix_missing": [],
        "asm_effect_not_native": [],
        "native_literal_missing": [],
        "dynamic_effect_missing": {},
        "dynamic_label_missing": {},
        "runtime_branch_missing": [],
        "runtime_asset_missing": [],
        "reachable_unported_ops": [],
        "unregistered_reachable": [],
        "extra_native": [],
    }

    # 1. Whole-file source: every program label, and every inter-program edge.
    for name, idx in sorted(src.labels.items()):
        if idx is None:
            continue
        if not (a <= src.line_of(idx) - 1 < z):
            report["out_of_region_programs"].append(
                {"name": name, "line": src.line_of(idx), "index": idx})
        ext = src.extent(name)
        for i in range(*ext):
            op, args = src.commands[i]
            for pos in BRANCH_ARGS.get(op, ()):
                if pos < len(args) and src.labels.get(args[pos]) is None:
                    report["source_branch_undefined"].append(
                        {"program": name, "op": op, "target": args[pos]})

    # 2. Effect table matrix targets.
    for eff, spec in sorted(native.get("effects", {}).items()):
        mtx = spec.get("matrix")
        if mtx and mtx != "0" and mtx not in registered:
            report["effect_matrix_missing"].append({"effect": eff, "matrix": mtx})

    # 2b. Original ASM effect blocks are the gameplay entry authority.
    asm_eff = asm_effects(src_name)
    report["asm_effect_count"] = len(asm_eff)
    for eff, spec in sorted(asm_eff.items()):
        if eff not in native.get("effects", {}):
            report["asm_effect_not_native"].append(eff)

    # 3. Native runtime string literals.
    literals = native_literals(pkg_files)
    report["native_runtime_literals"] = literals
    for lit in literals:
        if lit not in registered:
            report["native_literal_missing"].append(lit)

    # 4. Dynamic families.
    for fam, keys in dyn_effects.items():
        missing = [k for k in keys if k not in native.get("effects", {})]
        if missing:
            report["dynamic_effect_missing"][fam] = missing
            continue
        for k in keys:
            mtx = native["effects"][k].get("matrix", "0")
            if mtx != "0" and mtx not in registered:
                report["dynamic_effect_missing"].setdefault(fam, []).append(
                    f"{k}->{mtx}")
    for fam, names in dyn_labels.items():
        miss = [n for n in names if n not in registered]
        if miss:
            report["dynamic_label_missing"][fam] = miss
    for fam, names in native.get("constructed", {}).items():
        miss = [n for n in names if n not in registered]
        if miss:
            report["dynamic_label_missing"][fam] = miss

    # 5. Source-authoritative reachability from gameplay roots.
    asm_roots = {e["matrix"] for e in asm_eff.values() if e["matrix"] != "0"}
    all_roots = sorted(asm_roots | source_entries(src_name))
    report["dynamic_effect_families"] = dyn_effects
    report["dynamic_label_families"] = dyn_labels
    report["native_constructed_families"] = native.get("constructed", {})
    report["runtime_dynamic_calls"] = []
    for f in pkg_files:
        for line, code in enumerate((ROOT / f).read_text().splitlines(), 1):
            m = re.search(r"g\.(?:beginMatrix|matrixJump|jump|effect|effectCore|effectTiming)\((.*)", code)
            if m and not re.match(r'"[^"\n]+"\s*[,)]', m[1]):
                report["runtime_dynamic_calls"].append({"file": f, "line": line, "expression": m[1]})
    report["roots"] = all_roots
    report["implicit_opcode_targets"] = implicit_hops(src_name)
    report["frontend_owned"] = sorted(FRONTEND & set(src.labels))
    report["frontend_boundary"] = ["AFTER_XXBALLTS"] if table in (1, 2) else []
    reach, missing_roots = source_reachable(src, all_roots)
    if missing_roots:
        report["native_literal_missing"].extend(
            f"root-not-in-source:{r}" for r in missing_roots)
    report["source_reachable"] = sorted(reach)
    report["unregistered_reachable"] = sorted(reach - registered)
    report["invalid_reachable_program"] = sorted(label for label in reach & registered
        if not 0 <= labels[label] < len(commands))
    report["extra_native"] = sorted(registered - reach)
    frontend_roots = (FRONTEND & set(src.labels)) | set(report["frontend_boundary"])
    frontend, undefined_frontend = source_reachable(src, frontend_roots)
    # Frontend ownership is checked against the shared renderer registry;
    # F1/F2 intentionally intercept gameplay game-over before these entries.
    report["frontend_reachable"] = sorted(frontend | set(report["frontend_boundary"]))
    report["frontend_program_missing"] = sorted(set(report["frontend_reachable"]) - set(native["frontend_labels"]))
    report["extra_ownership"] = {label: ("frontend" if label in report["frontend_reachable"] else
        "sentinel" if label == "LAST_POS" else "data-only" if label == "SPINSCOREPTR" else
        "declared/unused entry") for label in report["extra_native"]}

    # 6. Native dispatch emulation over the registered stream.
    hops = {}
    for f in pkg_files:
        code = (ROOT / f).read_text()
        for block in re.split(r'case "', code)[1:]:
            op = block.split('"', 1)[0]
            destinations = GO_CALL.findall(block)
            if destinations:
                hops.setdefault(op, []).extend(destinations)
    targets, assets, ops, unresolved = emulate_native(
        commands, labels, all_roots, hops)
    for t in sorted(targets):
        if t not in registered:
            report["runtime_branch_missing"].append(t)
    for kind, name in sorted(assets):
        pool = native.get(kind + "s", {})
        if name not in pool:
            report["runtime_asset_missing"].append(f"{kind}:{name}")
    ported = native.get("ported_ops")
    if ported:
        report["reachable_unported_ops"] = sorted(
            o for o in ops if not o.startswith("_PRINT") and o not in ported)

    missing = set(report["unregistered_reachable"]) | set(report["native_literal_missing"]) | set(report["runtime_branch_missing"]) | set(report["frontend_program_missing"])
    for names in report["dynamic_label_missing"].values():
        missing.update(names)
    for spec in report["effect_matrix_missing"]:
        missing.add(spec["matrix"])
    report["missing_targets"] = sorted(missing)
    return report


def go_string_slice(labels, indent):
    lines = []
    for i in range(0, len(labels), 6):
        lines.append(indent + ", ".join(f'"{l}"' for l in labels[i:i + 6]) + ",")
    return "\n".join(lines)


def write_go_data(out):
    """Emit the source-reachable sets as Go data for the in-tree tests."""
    banner = "// Code generated by tools/matrix_reachability.py; DO NOT EDIT.\n"
    pres = [banner, "package presentation\n",
            "// sourceReachableMatrix is the set of matrix program labels the",
            "// original ASM can reach from gameplay entry points, per table.",
            "var sourceReachableMatrix = map[int][]string{"]
    for t in ("1", "2", "3", "4"):
        labels = out[t]["source_reachable"]
        pres.append(f"\t{t}: {{")
        pres.append(go_string_slice(labels, "\t\t"))
        pres.append("\t},")
    pres.append("}")
    emit(CHECKOUT / "internal/presentation/matrix_reachability_data.go", "\n".join(pres) + "\n")

    def pkg(path, pkgname, varname, labels):
        body = [banner, f"package {pkgname}\n",
                f"var {varname} = []string{{",
                go_string_slice(labels, "\t"), "}"]
        emit(CHECKOUT / path, "\n".join(body) + "\n")

    pkg("internal/partyland/matrix_reachability_data.go", "partyland",
        "sourceReachablePartyLand", out["1"]["source_reachable"])
    pkg("internal/speeddevils/matrix_reachability_data.go", "speeddevils",
        "sourceReachableSpeedDevils", out["2"]["source_reachable"])


def main() -> int:
    global ROOT, CHECK
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", default=".private-cleanup/generated/matrix-reachability.json")
    ap.add_argument("--root", default=str(ROOT),
                    help="tree to audit (defaults to this checkout)")
    ap.add_argument("--check", action="store_true", help="verify generated source sets without writing")
    args = ap.parse_args()
    CHECK = args.check
    cwd = Path.cwd()
    ROOT = Path(args.root).resolve()

    pl = load_generated("internal/partyland/timing_data.go")
    pres = load_generated("internal/presentation/content.go")
    sd = load_generated("internal/speeddevils/content.go")
    gs = load_generated("internal/gameshow/content.go")
    st = load_generated("internal/stones/content.go")

    ported_common = {"0", "_CLEAR1", "_CLEAR2", "_CLEAR3", "_CLEAR4", "_NUMBER",
                     "_FLASHON", "_FLASHOFF", "_MATRIXLGT", "_PARTYON",
                     "_PARTYOFF", "_PARTYONN", "_SETDECCOR"}

    t1 = {"labels": pl["labels"], "commands": pl["commands"],
          "effects": pl["effects"], "animations": pl["animations"],
          "scrolls": pl["scrolls"], "jingles": pl["jingles"],
          "ported_ops": ported_common | {
              "_JMP", "_JBCDZ", "_JBONUSX1", "_WAIT", "_ANIMATION", "_SCROLL",
              "_WAITJINGLE", "_WAITJINGLE2", "_EOSNURR", "_TSEND", "_COUNTDOWN",
              "_COUNTDOWN2", "_JINGLE", "_LASTJINGLE", "_RULLGARDIN_UPP",
              "_RULLGARDIN_NED", "_BONUS_X_CALCS", "_CALC_CYCLO", "_CALC_HAPPY",
              "_CALC_MEGA", "_FLORPA", "_WAITIFMULTI", "_KOLLA_XXBALL",
              "_BEATEN_MATRIX", "_CHANGE_PLAYER", "_NEW_BALL2",
              "_SHOOT_AGAIN_ONN", "_KNACKET", "_CHECK_XXBALLS", "_DOBEATEN",
              "_INIT_SCORE", "_SETLOOP", "_LOOP_", "_SHOW_SCORE"}}
    # PF12.1's Party matrix dispatcher falls back to decoded presentation
    # commands for these shared session panels. Preserve distinct PC spaces by
    # appending their programs, rather than treating their PCs as timing PCs.
    t1["labels"] = dict(t1["labels"])
    t1["commands"] = list(t1["commands"])
    for label in ("FIRST_NO_OF_PLAYERSTS", "NO_OF_PLAYERSTS", "SHOWPLAYERSTS"):
        start = pres["1"]["labels"][label]
        t1["labels"][label] = len(t1["commands"])
        for command in pres["1"]["commands"][start:]:
            t1["commands"].append(command)
            if command["op"] == "0": break
    t1["ported_ops"].add("_WAIT_GAME_ON")
    t2 = {"labels": sd["labels"], "commands": sd["commands"],
          "effects": sd["effects"], "animations": sd["animations"],
          "scrolls": sd["scrolls"], "jingles": sd["jingles"],
          "ported_ops": ported_common | {
              "_JMP", "_JBCDZ", "_JBONUSX1", "_WAIT", "_ANIMATION", "_SCROLL",
              "_WAITJINGLE", "_WAITJINGLE2", "_COUNTDOWN", "_COUNTDOWNCONTINUE",
              "_JINGLE", "_LASTJINGLE", "_RULLGARDIN_UPP", "_RULLGARDIN_NED",
              "_BONUS_X_CALCS", "_CALC_CYCLO", "_CALC_HAPPY", "_CALC_MEGA",
              "_FLORPA", "_WAITIFMULTI", "_KOLLA_XXBALL", "_BEATEN_MATRIX",
              "_WAIT_GAME_ON", "_CHANGE_PLAYER", "_NEW_BALL2", "_SHOOT_AGAIN_ONN", "_KNACKET",
              "_CHECK_XXBALLS", "_DOBEATEN", "_TURNOFFSPECIALMODE",
              "_RETURN_OF_THE_EVIL_SUPERMODE", "_TURNONTURBO", "_ADD50MILLION",
              "_SOUND_EFFECT"}}
    gs_eff = {k: {"matrix": v["matrix"]} for k, v in gs["effects"].items()}
    show_src = (ROOT / "reference/original-dos-source/SHOW.ASM").read_bytes().decode("latin1").upper()
    m = re.search(r"BONUSTABLE\s+DB\s+([\d, ]+)", show_src)
    if not m:
        raise SystemExit("SHOW.ASM BONUSTABLE not found")
    bonus_values = [int(x) for x in re.findall(r"\d+", m.group(1))]
    gs_constructed = constructed_families(
        ["internal/gameshow/rules.go"], {"_BONUSX": bonus_values},
        decimal_width=decimal_format_width())
    t3 = {"labels": pres["3"]["labels"], "commands": pres["3"]["commands"],
          "effects": gs_eff, "animations": pres["3"]["animations"],
          "scrolls": pres["3"]["texts"], "jingles": gs["jingles"],
          "constructed": gs_constructed,
          "ported_ops": ported_common | {
              "_JMP", "_JBCDZ", "_JBONUSX1", "_WAIT", "_ANIMATION", "_SCROLL",
              "_WAITJINGLE2", "_COUNTDOWN", "_JINGLE", "_LASTJINGLE",
              "_RULLGARDIN_UPP", "_RULLGARDIN_NED",
              "_BONUS_X_CALCS", "_CALC_CYCLO", "_CALC_HAPPY", "_FLORPA",
              "_WAITIFMULTI", "_KOLLA_XXBALL", "_BEATEN_MATRIX", "_CHANGE_PLAYER",
              "_NEW_BALL2", "_SHOOT_AGAIN_ONN", "_KNACKET", "_CHECK_XXBALLS",
              "_DOBEATEN", "_SOUND_EFFECT", "_TURNOFFSPECIALMODE",
              "_END_OF_SPIN", "_LIGHTFLASH", "_WAIT_GAME_ON", "_CHECK_HIGH",
              "_2_DEMO_MODE"}}

    # Source lines for the dynamic families (validated against ASM domains):
    #  F1 MULTIBONUS: Multiplier in {1,2,4,6} -> effect M{2,4,6,8}.
    #  F2 offroadLane: first unlit lamp n in 42..49 -> effect M{n-40}.
    #  F2 twoLoops:    Speed clamp 0..11 -> effect SSCORE{s+1}.
    #  F3 clockwise:   BONUSTABLE 2,3,4,6,8,10 -> label _BONUSX{v}TS.
    dyn_effects = {
        1: {"MULTIBONUS->M%d": ["M2", "M4", "M6", "M8"],
            "PARTYSCORE letter": [f"PARTYSCORE{i}" for i in range(1, 6)]},
        2: {"offroadLane M%d": [f"M{i}" for i in range(2, 10)],
            "twoLoops SSCORE%d": [f"SSCORE{i}" for i in range(1, 13)]},
        3: {"prize lit": [p + "LIT" for p in ("TV", "TRIP", "CAR", "BOAT", "HOUSE", "PLANE")],
            "prize won": ["YOUWIN" + p for p in ("TV", "TRIP", "CAR", "BOAT", "HOUSE", "PLANE")]},
    }
    # Declaration order supplies the names; source selector bounds supply the
    # reachable subset (e.g. Party Land has only four multiplier lamps).
    def family(src, pattern):
        return [k for k in asm_effects(src) if re.fullmatch(pattern, k)]
    assert family("PLAND", r"M[2468]") == dyn_effects[1]["MULTIBONUS->M%d"]
    dyn_effects[1]["PARTYSCORE letter"] = family("PLAND", r"PARTYSCORE[1-5]")
    dyn_effects[2]["offroadLane M%d"] = family("SDEV", r"M[2-9]")
    dyn_effects[2]["twoLoops SSCORE%d"] = family("SDEV", r"SSCORE\d+")
    dyn_effects[3]["prize lit"] = family("SHOW", r"(?:TV|TRIP|CAR|BOAT|HOUSE|PLANE)LIT")
    dyn_effects[3]["prize won"] = family("SHOW", r"YOUWIN(?:TV|TRIP|CAR|BOAT|HOUSE|PLANE)")
    dyn_effects[1]["Crazy alternate effect"] = ["ADVANCE3", "ADVANCE3B"]
    dyn_effects[2]["Car part forwarding"] = family("SDEV", r"PART[1-5]")
    dyn_effects[3]["Dollar touch group"] = ["TOUCHB", "TOUCHC"]
    dyn_effects[3]["Money Mania parity"] = ["MONEYMANIA", "MONEYMANIA2"]
    dyn_labels = {
        1: {"Dragon mode jackpot": ["JACKPOTTS", "JACKPOT_SPECIAL_HH_TS", "JACKPOT_SPECIAL_ML_TS"]},
        2: {},
        3: {"_BONUSX%dTS": [f"_BONUSX{v}TS"
                            for v in bonus_values]},
    }

    st_files = [str(p.relative_to(ROOT)) for p in sorted((ROOT / "internal/stones").glob("*.go"))
                if not p.name.endswith("_test.go") and p.name != "content.go"]
    st_effects = {k: {"matrix": v["matrix"]} for k, v in st["effects"].items()}
    st_source = "\n".join(normalized_lines("STONES"))
    # BPOINTER starts at 39 and is bounded by CMP BPOINTER,43. Five
    # contiguous EFFECT_STRUC records beginning at M2 are indexed by it.
    assert re.search(r"MOV\s+BPOINTER,39", st_source)
    assert re.search(r"CMP\s+BPOINTER,43", st_source)
    mkeys = re.findall(r"^(M\d+)\s+DW", st_source, re.M)
    assert mkeys == ["M2", "M4", "M6", "M8", "M10"]
    ghost = re.search(r"GHOSTAREA\s+LABEL BYTE(.*?)EVENT_LIT\s+DW", st_source, re.S)[1]
    ghosts = re.findall(r"DW\s+\w+,(\w+),(EVENT_LIT\d+)", ghost)
    assert len(ghosts) == 8
    tower = sorted(set(re.findall(r"EFFECT\s+(TOWERHUNT\d+)", st_source)))
    dyn_effects[4] = {"Well multiplier BPOINTER 39..43": mkeys,
                      "ghost enable GHOSTAREA": [e for _, e in ghosts],
                      "ghost collect GHOSTAREA": [e for e, _ in ghosts],
                      "Tower Hunt stages": tower,
                      "Multi Demons lock count": ["MILLION5", "MILLION10", "MILLION20"]}
    dyn_effects[4]["Tower flashing award lamps"] = ["EXTRABALL", "SUPERJACK", "JACKPOT", "DOUBLEBONUS", "HOLDBONUS", "TMILLION5", "TMILLION"]
    dyn_effects[4]["Scream threshold text"] = ["JUMP_AT", "JUMP_AT2"]
    dyn_labels[4] = {"Tower special-mode continuation": ["BACK_2_OFFROADTS", "BACK_2_TURBOTS"]}
    t4 = {"labels": pres["4"]["labels"], "commands": pres["4"]["commands"],
          "effects": st_effects, "animations": pres["4"]["animations"],
          "scrolls": pres["4"]["texts"], "jingles": st["jingles"]}

    for table, native in enumerate((t1, t2, t3, t4), 1):
        native["frontend_labels"] = pres[str(table)]["labels"]

    out = {
        "1": check_table(1, t1, dyn_effects[1], dyn_labels[1],
                         [str(p.relative_to(ROOT)) for p in sorted((ROOT / "internal/partyland").glob("*.go"))
                          if not p.name.endswith("_test.go") and p.name not in ("content.go", "timing_data.go", "matrix_reachability_data.go")]),
        "2": check_table(2, t2, dyn_effects[2], dyn_labels[2],
                         [str(p.relative_to(ROOT)) for p in sorted((ROOT / "internal/speeddevils").glob("*.go"))
                          if not p.name.endswith("_test.go") and p.name not in ("content.go", "timing_data.go", "matrix_reachability_data.go")]),
        "3": check_table(3, t3, dyn_effects[3], dyn_labels[3],
                         [str(p.relative_to(ROOT)) for p in sorted((ROOT / "internal/gameshow").glob("*.go"))
                          if not p.name.endswith("_test.go") and p.name not in ("content.go", "timing_data.go", "matrix_reachability_data.go")]),
    }
    out["4"] = check_table(4, t4, dyn_effects[4], dyn_labels[4], st_files)
    if not CHECK:
        (cwd / args.json).parent.mkdir(parents=True, exist_ok=True)
        emit(cwd / args.json, json.dumps(out, indent=2) + "\n")
    write_go_data(out)

    failed = bool(STALE)
    if STALE:
        print("Stale source-reachability data: " + ", ".join(STALE))
    for t, r in out.items():
        title = {"1": "Party Land", "2": "Speed Devils", "3": "Gameshow", "4": "Stones ’N Bones"}[t]
        print(f"{title}: source reachable {len(r['source_reachable'])}, native registered {len(r['registered'])}, missing {len(r['missing_targets'])}")

        for key in ("source_branch_undefined", "effect_matrix_missing",
                    "asm_effect_not_native", "native_literal_missing",
                    "dynamic_effect_missing", "dynamic_label_missing",
                    "runtime_branch_missing", "runtime_asset_missing", "frontend_program_missing",
                    "reachable_unported_ops", "unregistered_reachable", "invalid_reachable_program"):
            if r.get(key):
                failed = True
                print(f"   !! {key}: {r[key]}")
        if r["extra_native"]:
            print(f"   (registered but not source-reachable from roots: "
                  f"{len(r['extra_native'])})")

    print("MISSING" if failed else "OK")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
