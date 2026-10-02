#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""PF10.2 source-authoritative matrix/presentation operand audit.

Second validation layer on top of ``tools/matrix_reachability.py``.  Where the
reachability checker proves that every source-reachable *program* is
registered, this checker proves that every *operand* of every source-reachable
command is classified, resolved from the original DOS ASM/data, valid for its
opcode, and carried into the generated native command as a typed integer.

Invariant enforced:

    every reachable matrix/presentation operand
        -> classified
        -> resolved
        -> valid for its opcode
        -> no runtime string-expression lookup failure

Exit status is non-zero on any reachable unresolved operand, unknown opcode,
operand-count mismatch, or generated data that disagrees with the source.

Usage:
    python3 tools/matrix_operand_audit.py            # write the inventory
    python3 tools/matrix_operand_audit.py --check    # verify, write nothing
"""
from __future__ import annotations

import argparse
import json
import re
import sys
sys.dont_write_bytecode = True
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CHECK = False
STALE = []

sys.path.insert(0, str(Path(__file__).resolve().parent))
import matrix_operand_schema as schema  # noqa: E402
import matrix_reachability as mr  # noqa: E402

TABLES = {
    1: ("PLAND", "Party Land"),
    2: ("SDEV", "Speed Devils"),
    3: ("SHOW", "Gameshow"),
    4: ("STONES", "Stones ’N Bones"),
}

MODE_LITERAL = "literal-integer"
MODE_ARITH = "arithmetic-expression"
MODE_SYMBOL = "symbolic-constant"
MODE_PLACEHOLDER = "placeholder"
MODE_LABEL = "label"
MODE_TEXT = "text-identity"
MODE_ANIM = "animation-identity"
MODE_VALUE = "value-identity"
MODE_JINGLE = "jingle-identity"
MODE_SOUND = "sound-identity"
MODE_FLAG = "numeric-flag"
MODE_UNKNOWN = "unknown"

# Status of one reachable numeric operand against the *previous* native
# representation (raw source strings plus the ad-hoc holes PF10.1 patched).
OK = "OK"
PANIC = "PANIC"
SILENT_ZERO = "SILENT-ZERO"
MISSING_NATIVE = "MISSING-NATIVE"


def emit(path, body):
    if CHECK:
        if not path.exists() or path.read_text() != body:
            STALE.append(str(path))
    else:
        path.write_text(body)


def load_generated(relpath: str) -> dict:
    return mr.load_generated(relpath)


def table_constants(src_name: str):
    """Recover the table's assembler constants exactly as the extractors do."""
    lines = mr.normalized_lines(src_name)
    constants = {"SW": 336, "TOTCENT": 0, "SPEED": 4}
    pattern = re.compile(r"\s*(\w+)\s*(?:=|EQU\s+)([^;]+)")

    def value(text):
        return schema.resolve_expression(text, constants)

    for line in lines:
        m = pattern.match(line)
        if not m:
            continue
        try:
            constants[m[1].upper()] = value(m[2])
        except (schema.OperandError, KeyError, ValueError, SyntaxError,
                ZeroDivisionError):
            pass
    if not schema.seed_empty_jingle(constants, lines):
        raise SystemExit(f"{src_name}.ASM: S_Empty record not found")
    return constants, lines


def is_expression(text: str) -> bool:
    body = text.strip()
    if re.fullmatch(r"[-+]?\d+", body):
        return False
    return bool(re.search(r"[+\-*/()]", body[1:] if body[:1] in "+-" else body))


def operand_class(op: str, index: int, text: str, constants) -> str:
    kind = schema.classify(op, index)
    if kind == schema.LBL:
        return MODE_LABEL
    if kind == schema.TXT:
        return MODE_TEXT
    if kind == schema.ANIM:
        return MODE_ANIM
    if kind == schema.VAL:
        return MODE_VALUE
    if kind == schema.JIN:
        return MODE_JINGLE
    if kind == schema.SND:
        return MODE_SOUND
    if schema.is_placeholder(text):
        return MODE_PLACEHOLDER
    if kind == schema.FLG:
        return MODE_FLAG
    if kind == schema.NUM:
        if re.fullmatch(r"[-+]?\d+", text.strip()):
            return MODE_LITERAL
        if is_expression(text):
            return MODE_ARITH
        return MODE_SYMBOL
    if kind == schema.IGN:
        return MODE_PLACEHOLDER if schema.is_placeholder(text) else MODE_FLAG
    return MODE_UNKNOWN


# ---------------------------------------------------------------------------
# Previous native representation: what the runtime could resolve before this
# audit's extraction fix.  This is what makes "unresolved before" measurable.
# ---------------------------------------------------------------------------

def legacy_status(table: int, op: str, args, index: int, positions, ticks,
                  origin: str, runtime_reachable: bool) -> tuple:
    """(mode, resolved_value|None, status) for one numeric operand.

    Models the previous native representation exactly: what the shipping code
    could resolve from a raw operand string before this audit's fix.
    """
    raw = args[index] if index < len(args) else ""
    if not runtime_reachable:
        return ("static-only", None, OK)
    if schema.is_placeholder(raw):
        return ("placeholder", None, OK)
    if re.fullmatch(r"[-+]?\d+", raw.strip()):
        return ("literal-atoi", int(raw), OK)
    if origin == "attract":
        # The expander stored resolved values for WAIT/FLASHON/RULLGARDIN.
        if op in ("_WAIT", "_FLASHON", "_RULLGARDIN_UPP", "_RULLGARDIN_NED"):
            return ("attract-inplace", None, OK)
    if origin == "frontend-command":
        # F1/F2 frontend tails are replayed synthetically by GameOver/Attract;
        # their extracted operands are label/op references only.
        return ("synthetic-replay", None, OK)
    if op in ("_WAITIFMULTI", "_SETLOOP", "_LOOP_", "_LIGHTFLASH", "_SETDECCOR"):
        # These operands are never consumed numerically by the native port.
        return ("operand-unused", None, OK)
    if op == "_WAIT" and table in (1, 2):
        # F1/F2 extractors already stored a resolved ticks word.
        if ticks is not None:
            return ("ticks-field", ticks, OK)
        return ("raw-expression", None, PANIC)
    if op == "_TOWER":
        # PF10.1 added the raw expression to the positions map; it is a
        # runtime string lookup, not a typed operand.
        if raw in positions:
            return ("positions-map", positions[raw], OK)
        return ("raw-expression", None, PANIC)
    if op == "_COUNTDOWN" and table == 4:
        # Stones numeric() looked the operand up in the positions map or
        # parsed it; the other tables ignore the countdown digits.
        if raw in positions:
            return ("positions-map", positions[raw], OK)
        return ("raw-expression", None, PANIC)
    if op.startswith("_PRINT") and index == 1:
        # presentation.Display.Position(expr) keyed the generated provenance
        # map by the raw source expression.
        if raw in positions:
            return ("positions-map", positions[raw], OK)
        return ("raw-expression", None, PANIC)
    if op == "_LASTJINGLE":
        if raw == "EMPTYJINGLE":
            return ("runtime-hardcode", None, OK)
        return ("atoi-ignored-error", 0, SILENT_ZERO)
    if op == "_FLASHON":
        # presentation.Visit used strconv.Atoi and dropped the error.
        return ("atoi-ignored-error", 0, SILENT_ZERO)
    if op in ("_RULLGARDIN_UPP", "_RULLGARDIN_NED"):
        # Each table's own dispatcher used Atoi and dropped the error; the
        # shared renderer did the same.
        return ("atoi-ignored-error", 0, SILENT_ZERO)
    if op == "_WAIT":
        # F3 used strconv.Atoi and propagated the error; F4 numeric() panicked.
        return ("atoi-panic", None, PANIC)
    return ("raw-expression", None, PANIC)


CLEAR_OP = {1: ("_CLEAR4", ()), 2: ("_CLEAR1", ()),
            3: ("_ANIMATION", ("_CLEAR",)), 4: ("_ANIMATION", ("_CLEAR",))}


def native_key(table: int, op: str, args):
    """Native extraction key for a source command.

    matrix_reachability.Source renders a bare CLEARIT macro as ``_CLEAR?``;
    the extractors expand it to the table's real clear op.
    """
    if op == "_CLEAR?":
        return CLEAR_OP[table]
    return (op, tuple(args))


def numeric_operands(op: str, args):
    return [i for i, k in enumerate(schema.kinds(op) or ()) if k == schema.NUM]


# ---------------------------------------------------------------------------
# Source text data.
#
# A matrix program names text buffers by identity (TXT/VAL operands), so the
# operand layer cannot see whether the bytes themselves were extracted
# correctly.  Retail text bytes mix character literals with arithmetic
# ('7'+2, 6+'7', '0'+7); this layer re-derives every DB element from source and
# compares it with the generated buffer.  It also inventories the buffers the
# original *writes* at runtime, which the port must reproduce.
# ---------------------------------------------------------------------------

DB_RE = re.compile(r"^(?:(\w+)\s+)?DB\s+(.+)$", re.I)
GO_TEXT_WRITE = re.compile(r'Content\.Texts\[\s*"([A-Z0-9_]+)"\s*\]\s*=')


def source_texts(constants, lines, strict_labels=()):
    out, cur = {}, None
    for line in lines:
        m = DB_RE.match(line.split(";")[0].strip())
        if not m:
            cur = None
            continue
        if m[1]:
            cur = m[1].upper()
            out[cur] = []
        if not cur:
            continue
        try:
            out[cur] += schema.db_bytes(m[2], constants)
        except (schema.OperandError, KeyError, ValueError, SyntaxError):
            if cur in strict_labels: raise
            cur = None
    return {k: v for k, v in out.items() if all(0 <= x <= 255 for x in v)}


def runtime_text_buffers(src_name, lines, eligible, go_files):
    """Text buffers a matrix program prints and the original writes at runtime.

    ``eligible`` restricts the scan to the labels the extracted command stream
    actually prints; BCD rows and other indexed DATA are a different namespace.
    """
    native = "\n".join(p.read_text() for p in go_files)
    direct = {label: re.compile(r"\b" + label + r"\s*\[") for label in eligible}
    offset = {label: re.compile(r"OFFSET\s+" + label + r"\b") for label in eligible}
    written = {}
    for i, line in enumerate(lines, 1):
        code = line.split(";")[0]
        if not code or DB_RE.match(code.strip()):
            continue
        upper = code.upper()
        is_move = "MOV" in upper or "MOVE" in upper
        for label in eligible:
            if label not in upper:
                continue
            if direct[label].search(upper) or (is_move and offset[label].search(upper)):
                written.setdefault(label, []).append(i)
    # Require a write in the function containing the label, rather than a
    # string mention anywhere in the package. This covers direct assignments
    # and finite label loops used by the existing source player/scream writers.
    def native_writer(label):
        for fn in re.split(r"(?m)^func ", native)[1:]:
            if f'"{label}"' not in fn: continue
            if re.search(r'(?:MutableText|sourceTextBuf)\("'+label+r'"', fn): return True
            if re.search(r'Content.Texts\[[^\]]+\]\s*=', fn) and re.search(r'\w+\[[^\]]+\]\s*=', fn): return True
        return label == "BONUS_X_TEXT" and "g.Display.WriteBonusMultiplier(g.Multiplier)" in native
    out = []
    for label, where in sorted(written.items()):
        out.append({
            "label": label,
            "source_file": f"{src_name}.ASM",
            "source_lines": where[:8],
            # The port addresses the buffer by name from Go when it writes it.
            "native_text_write": native_writer(label),
        })
    return out


def check_text_data(table: int, native: dict, constants, lines, eligible) -> dict:
    src_name, title = TABLES[table]
    src = source_texts(constants, lines, eligible)
    generated = native["presentation_texts"]
    problems, changed = [], []
    missing = []
    for label, want in sorted(src.items()):
        if label not in generated:
            problems.append(f"{src_name}.ASM {label}: missing generated buffer")
            continue
        got = list(generated[label])
        if got != want:
            problems.append(
                f"{src_name}.ASM {label}: generated {got[:16]} != source {want[:16]}")
            changed.append(label)
    for label in sorted(set(generated) - set(src)):
        missing.append(label)
    buffers = runtime_text_buffers(src_name, lines, eligible, native["go_files"])
    uncovered = [b["label"] for b in buffers if not b["native_text_write"]]
    return {
        "source_texts": len(src),
        "generated_texts": len(generated),
        "text_byte_mismatches": changed,
        "text_problems": problems,
        "source_only_texts": missing,
        "runtime_text_buffers": buffers,
        "runtime_text_buffers_without_native_writer": uncovered,
    }


def multiline_command_args(src_name, src, index, op, args):
    """Rejoin a source command whose operands continue on following DW lines.

    SHOW.ASM declares ``spinTS DW _PRINT13_NUMBER`` and then puts the data
    label and the position on the next two lines.  The extractor carries a
    hand-written patch for this; the audit re-derives it from the source text
    so the patch itself is checked rather than trusted.
    """
    if args or op != "_PRINT13_NUMBER":
        return args
    lines = mr.normalized_lines(src_name)
    start = src.line_of(index)
    following = []
    for line in lines[start:start + 6]:
        text = line.split(";")[0].strip()
        if text:
            following.append(text)
    if len(following) < 2:
        return args
    m = re.fullmatch(r"(\w+)\s+DW\s+(?:OFFSET\s+)?(\w+)", following[0])
    n = re.fullmatch(r"DW\s+(.+)", following[1])
    if m and n:
        return [m[1], n[1].strip()]
    return args


def check_table(table: int, native: dict, constants) -> dict:
    src_name, title = TABLES[table]
    region = mr.current_regions()[table][1]
    src = mr.Source(src_name, region)
    roots = sorted({e["matrix"] for e in mr.asm_effects(src_name).values()
                    if e["matrix"] != "0"} | mr.source_entries(src_name))
    reach, missing_roots = mr.source_reachable(src, roots)

    positions = native["positions"]
    gameplay = native["gameplay"]
    frontend = native["frontend"]
    native_index = {}
    for origin, group in (("gameplay", gameplay),
                          ("frontend-command", native["presentation_commands"]),
                          ("attract", frontend)):
        for c in group:
            native_index.setdefault(
                (c["op"], tuple(c.get("args") or [])), (c, origin))
    inventory = []
    problems = []
    counts = {"operands": 0, "arithmetic": 0, "symbolic": 0, "labels_assets": 0,
              "numeric": 0, "unresolved_before": 0, "unresolved_after": 0,
              "runtime_reachable": 0, "static_only": 0}
    before_kinds = {}
    opcodes_used = set()
    printed_texts = set()
    # Validate every generated command, including attract and static tails.
    # Flag operands are typed just like numeric counter/position operands.
    for group in (gameplay, frontend, native["presentation_commands"]):
        for c in group:
            try: expected = schema.numeric_values(c["op"],c.get("args") or [],constants)
            except schema.OperandError as exc:
                problems.append(str(exc));continue
            actual = {int(k):v for k,v in (c.get("nums") or {}).items()}
            if actual != expected:
                problems.append(f"{src_name}.ASM generated {c['op']} {c.get('args')}: typed values {actual} != {expected}")


    command_indices = set()
    runtime_indices = set()
    for label in sorted(reach):
        ext = src.extent(label)
        if not ext:
            continue
        for i in range(*ext):
            command_indices.add(i)
        # Walk the same extent but stop exactly where the native VM stops
        # dispatching, so "would crash at runtime" counts stay honest.
        i, end = ext
        while i < end and i not in runtime_indices:
            runtime_indices.add(i)
            op = src.commands[i][0]
            i += 1
            if op in mr.NATIVE_STOP_OPS:
                break

    for i in sorted(command_indices):
        op, args = src.commands[i]
        args = multiline_command_args(src_name, src, i, op, args)
        line = src.line_of(i)
        program = enclosing_program(src, reach, i)
        opcodes_used.add(op)
        kinds = schema.kinds(op)
        if kinds is None:
            problems.append(f"{src_name}.ASM:{line} unknown opcode {op}")
            kinds = ()
        if len(args) != len(kinds):
            problems.append(
                f"{src_name}.ASM:{line} {op} has {len(args)} operands, schema has {len(kinds)}: {args}")
        runtime_reachable = i in runtime_indices
        key = native_key(table, op, args)
        match = native_index.get(key)
        native_cmd, origin = match if match else (None, "none")
        if native_cmd is None and op not in ("0",):
            problems.append(f"{src_name}.ASM:{line} {op} {args} absent from native extraction")
        nums = (native_cmd or {}).get("nums") or {}
        nums = {int(k): v for k, v in nums.items()}
        ticks = (native_cmd or {}).get("ticks")
        counts["runtime_reachable" if runtime_reachable else "static_only"] += 1

        for index, text in enumerate(args):
            klass = operand_class(op, index, text, constants)
            counts["operands"] += 1
            if klass in (MODE_LABEL, MODE_TEXT, MODE_ANIM, MODE_VALUE,
                         MODE_JINGLE, MODE_SOUND):
                counts["labels_assets"] += 1
            if klass == MODE_ARITH:
                counts["arithmetic"] += 1
            if klass == MODE_SYMBOL:
                counts["symbolic"] += 1
            resolved = None
            method = "none"
            status = OK
            if schema.classify(op, index) in (schema.NUM, schema.FLG):
                counts["numeric"] += 1
                try:
                    resolved = schema.resolve_expression(text, constants)
                    method = "assembler-expression"
                except schema.OperandError as exc:
                    problems.append(f"{src_name}.ASM:{line} {op}[{index}] {text!r}: {exc}")
                    status = "UNRESOLVED"
                if schema.classify(op, index) == schema.NUM:
                    mode, before, before_status = legacy_status(
                        table, op, args, index, positions, ticks, origin,
                        runtime_reachable)
                    before_kinds[mode] = before_kinds.get(mode, 0) + 1
                    if before_status != OK:
                        counts["unresolved_before"] += 1
                    if index not in nums:
                        counts["unresolved_after"] += 1
                        status = "UNRESOLVED-AFTER"
                    elif resolved is not None and nums[index] != resolved:
                        problems.append(
                            f"{src_name}.ASM:{line} {op}[{index}] generated nums={nums[index]} != source {resolved}")
                        status = "MISMATCH"
                    if before_status == SILENT_ZERO and index in nums:
                        status = "FIXED-SILENT-ZERO"
                    elif before_status == PANIC and index in nums:
                        status = "FIXED-PANIC"
            elif klass == MODE_PLACEHOLDER and schema.classify(op, index) == schema.IGN:
                status = OK
            if method == "none" and resolved is not None:
                method = "assembler-expression"
            inventory.append({
                "table": table,
                "table_name": title,
                "program": program,
                "source_file": f"{src_name}.ASM",
                "source_line": line,
                "opcode": op,
                "operand_index": index,
                "raw_source_text": text,
                "operand_class": klass,
                "resolved_value": resolved,
                "resolution_method": method,
                "reachable": True,
                "runtime_reachable": runtime_reachable,
                "native_origin": origin,
                "legacy_mode": None,
                "native_representation": (
                    f"nums[{index}]={nums[index]}" if index in nums else
                    "identity" if klass in (MODE_LABEL, MODE_TEXT, MODE_ANIM,
                                            MODE_VALUE, MODE_JINGLE, MODE_SOUND)
                    else "raw-string"),
                "status": status,
            })
            if schema.classify(op, index) == schema.NUM:
                inventory[-1]["legacy_mode"] = mode
            if klass == MODE_TEXT:
                printed_texts.add(text.upper())

    return {
        "source_file": f"{src_name}.ASM",
        "region": list(region),
        "reachable_programs": sorted(reach),
        "missing_roots": missing_roots,
        "opcodes_used": sorted(opcodes_used),
        "counts": counts,
        "legacy_modes": before_kinds,
        "problems": problems,
        "inventory": inventory,
        "text_data": check_text_data(table, native, constants,
                                     mr.normalized_lines(src_name), printed_texts),
    }


def enclosing_program(src, reach, index):
    best = ""
    for label in sorted(reach):
        start = src.labels.get(label)
        if start is None or start > index:
            continue
        ext = src.extent(label)
        if ext and ext[0] <= index < ext[1]:
            if not best or (src.labels[best] or 0) < start:
                best = label
    return best


def load_native(root: Path):
    pres = load_generated("internal/presentation/content.go")
    pl = load_generated("internal/partyland/timing_data.go")
    sd = load_generated("internal/speeddevils/content.go")
    packages = {1: "partyland", 2: "speeddevils", 3: "gameshow", 4: "stones"}
    out = {}
    for table in (1, 2, 3, 4):
        c = pres[str(table)]
        gameplay = {"1": pl["commands"], "2": sd["commands"]}.get(str(table), c["commands"])
        go_files = sorted(p for p in (root / "internal" / packages[table]).glob("*.go")
                          if not p.name.endswith("_test.go"))
        go_files.append(root / "internal" / "presentation" / "attract.go")
        out[table] = {
            "positions": c["positions"],
            "gameplay": gameplay,
            "frontend": c.get("attract", []),
            "presentation_commands": c["commands"],
            "presentation_attract": c.get("attract", []),
            "presentation_texts": c["texts"],
            "go_files": go_files,
        }
    return out


def main() -> int:
    global ROOT, CHECK
    ap = argparse.ArgumentParser()
    ap.add_argument("--json", default=".private-cleanup/generated/matrix-operand-audit.json")
    ap.add_argument("--root", default=str(ROOT))
    ap.add_argument("--check", action="store_true",
                    help="verify generated operand data without writing")
    args = ap.parse_args()
    CHECK = args.check
    cwd = Path.cwd()
    ROOT = Path(args.root).resolve()

    native = load_native(ROOT)
    result = {}
    for table in (1, 2, 3, 4):
        src_name = TABLES[table][0]
        constants, _ = table_constants(src_name)
        result[str(table)] = check_table(table, native[table], constants)

    body = json.dumps(result, indent=2, ensure_ascii=False) + "\n"
    if not CHECK:
        (cwd / args.json).parent.mkdir(parents=True, exist_ok=True)
        emit(cwd / args.json, body)

    failed = bool(STALE)
    if STALE:
        print("Stale source-operand data: " + ", ".join(STALE))
    text_mismatches = 0
    mutation_gaps = []
    for table in (1, 2, 3, 4):
        r = result[str(table)]
        _src, title = TABLES[table]
        c = r["counts"]
        if c["unresolved_after"]: failed = True
        print(f"{title}: unresolved operands {c['unresolved_after']}")
        for problem in r["problems"]:
            failed = True
            print(f"   !! {problem}")
        t = r["text_data"]
        text_mismatches += len(t["text_problems"])
        for problem in t["text_problems"]:
            failed = True
            print(f"   !! {problem}")
        for label in t["runtime_text_buffers_without_native_writer"]:
            mutation_gaps.append(f"{title}:{label}")
    print(f"total operands {sum(result[t]['counts']['operands'] for t in result)}")
    print(f"arithmetic expressions {sum(result[t]['counts']['arithmetic'] for t in result)}")
    print(f"symbolic numeric operands {sum(result[t]['counts']['symbolic'] for t in result)}")
    print(f"labels/assets {sum(result[t]['counts']['labels_assets'] for t in result)}")
    print(f"unresolved before {sum(result[t]['counts']['unresolved_before'] for t in result)}")
    print(f"unresolved after {sum(result[t]['counts']['unresolved_after'] for t in result)}")
    print(f"source texts {sum(result[t]['text_data']['source_texts'] for t in result)}, "
          f"byte mismatches {text_mismatches}")
    print(f"runtime-mutated text buffers without a native writer {len(mutation_gaps)}")
    if mutation_gaps: failed = True
    for gap in mutation_gaps:
        print(f"   -- {gap}")
    print("MISSING" if failed else "OK")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
