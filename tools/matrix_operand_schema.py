#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Source-authoritative matrix/presentation opcode operand schema.

Authority is the original DOS retail assembler and linked data:

  * the shared matrix interpreter in ``FANTASIE.ASM`` (``DO_SPEC_MATRIX`` and
    the ``_ANIMATION``/``_SCROLL``/``print``/``_FLASHON``/``_FLASHOFF``/
    ``_MATRIXLGT``/``_SETDECCOR``/``_WAIT``/``_JINGLE``/``_SOUND_EFFECT``/
    ``_COUNTDOWN`` handlers);
  * the table-local handlers in ``PLAND.ASM``/``SDEV.ASM``/``SHOW.ASM``/
    ``STONES.ASM`` (``_JMP``/``_JBCDZ``/``_JBONUSX1``/``_SHOW_SCORE``/
    ``_LOOP_``/``_SETLOOP``/``_tower``/...).

Operand arity is not guessed from the native Go port.  Each handler advances
``BX`` past its own operands:

  * ``NORMAL_END``/``print_end`` progress ``ADD BX,4``/``ADD BX,6`` -> 1/2
    operand words;
  * ``three_end`` progresses ``ADD BX,8`` -> 3 operand words;
  * ``endit``/``_Jmp``/``_JBONUSX1``/``_SHOW_SCORE`` consume 0 or 1 direct
    operand words before jumping through the stream.

The kinds below classify *how the original runtime consumes* each operand:

``NUM``    an integer the runtime copies into a counter, position or row.  It
           must be resolved from the retail source expression at extraction
           time and must never survive into runtime as an expression string.
``LBL``    a matrix program label used as a branch/loop destination.
``TXT``    a text/scroller identity (DATA label), or the ``_CLEAR`` animation
           identity for the ``CLEARIT`` macro.
``ANIM``   an animation identity.
``VAL``    a runtime value/BCD row identity (``BONUSSIFFRORNA``,
           ``(HI_SCORE_LIST+...)``, ``OR_TOTAL``, ...).
``JIN``    a jingle identity.
``SND``    a sound-effect identity.
``FLG``    a numeric operand consumed as a boolean flag (``_MATRIXLGT``).
``IGN``    a word the original handler never reads (``?`` placeholders and
           filler words such as ``_PARTYON,1``).
"""
from __future__ import annotations

import re

NUM = "num"
LBL = "label"
TXT = "text"
ANIM = "animation"
VAL = "value"
JIN = "jingle"
SND = "sound"
FLG = "flag"
IGN = "ignored"

# op -> (operand kinds..., origin)
# Origin names the authoritative handler.  FANTASIE.ASM is the shared engine;
# the other files are table-local handlers reached from the shared dispatcher.
OPCODES: dict[str, tuple[tuple[str, ...], str]] = {
    "0": ((), "matrix terminator (word 0 ends the program)"),
    "_2_DEMO_MODE": ((), "table-local handler, endit"),
    "_ANIMATION": ((ANIM,), "FANTASIE.ASM:_ANIMATION STARTANIM [BX+2]"),
    "_BEATEN_MATRIX": ((), "FANTASIE.ASM:_BEATEN_MATRIX"),
    "_BONUS_X_CALCS": ((IGN,), "table-local handler, NORMAL_END filler"),
    "_CALC_CYCLO": ((), "table-local handler, ADD BX,2"),
    "_CALC_HAPPY": ((), "table-local handler, ADD BX,2"),
    "_CALC_MEGA": ((), "table-local handler, ADD BX,2"),
    "_CHANGE_PLAYER": ((), "FANTASIE.ASM:_CHANGE_PLAYER"),
    "_CHECK_HIGH": ((IGN,), "table-local handler, NORMAL_END filler"),
    "_CHECK_XXBALLS": ((), "table-local handler"),
    "_CLEAR1": ((), "FANTASIE.ASM:_clear1"),
    "_CLEAR2": ((), "FANTASIE.ASM:_clear2"),
    "_CLEAR3": ((), "FANTASIE.ASM:_clear3"),
    "_CLEAR4": ((), "FANTASIE.ASM:_clear4"),
    "_COUNTDOWN": ((NUM, NUM, VAL), "FANTASIE.ASM:_countdown [BX+2],[BX+4],[BX+6]"),
    "_COUNTDOWN2": ((NUM, NUM, VAL), "FANTASIE.ASM:_countdown2 -> MEGA_RESET"),
    "_COUNTDOWNCONTINUE": ((IGN, IGN, VAL), "FANTASIE.ASM:_COUNTDOWNCONTINUE [BX+6], three_end"),
    "_DOBEATEN": ((IGN,), "table-local handler, NORMAL_END filler"),
    "_END_OF_SPIN": ((IGN,), "SHOW.ASM:_END_OF_SPIN filler"),
    "_EOSNURR": ((IGN,), "PLAND.ASM:_EOSNURR filler"),
    "_FLASHOFF": ((IGN,), "FANTASIE.ASM:_FLASHOFF ignores [BX+2]"),
    "_FLASHON": ((NUM,), "FANTASIE.ASM:_FLASHON MOV AX,[BX+2]"),
    "_FLORPA": ((), "FANTASIE.ASM:_FLORPA"),
    "_GRIMOFF": ((IGN,), "STONES.ASM:_GRIMOFF filler"),
    "_INIT_SCORE": ((IGN,), "table-local handler, NORMAL_END filler"),
    "_JACKEND": ((IGN,), "STONES.ASM:_JACKEND filler"),
    "_JBCDZ": ((VAL, LBL), "table-local _JBCDZ IFZEROBCD [BX+2], MOV BX,[BX+4]"),
    "_JBONUSX1": ((LBL,), "table-local _JBONUSX1 MOV BX,[BX+2]"),
    "_JINGLE": ((JIN,), "FANTASIE.ASM:_JINGLE MOV SI,[BX+2]"),
    "_JMP": ((LBL,), "table-local _Jmp MOV BX,[BX+2]"),
    "_KNACKET": ((IGN,), "table-local handler, NORMAL_END filler"),
    "_KOLLA_XXBALL": ((), "table-local handler"),
    "_LASTJINGLE": ((NUM,), "FANTASIE.ASM:_lastjingle MOV AX,[BX+2]"),
    "_LIGHTFLASH": ((NUM,), "SHOW.ASM:_LIGHTFLASH [BX+2]"),
    "_LOOP_": ((NUM, LBL), "table-local _loop_/_setloop [BX+2],[BX+4]"),
    "_MATRIXLGT": ((FLG,), "FANTASIE.ASM:_MATRIXLGT CMP WORD PTR [BX+2],0"),
    "_NEW_BALL2": ((), "FANTASIE.ASM:_NEW_BALL2"),
    "_NUMBER": ((VAL,), "FANTASIE.ASM:_NUMBER MOVE SISA,[BX+2]"),
    "_PARTYOFF": ((IGN,), "table-local _PARTYOFF, NORMAL_END filler"),
    "_PARTYON": ((IGN,), "table-local _PARTYON 'BX NOT USED'"),
    "_PARTYONN": ((IGN,), "table-local _PARTYONN, NORMAL_END filler"),
    "_PRINT11": ((TXT, NUM), "FANTASIE.ASM:print_end [BX+2],[BX+4]"),
    "_PRINT13": ((TXT, NUM), "FANTASIE.ASM:print_end [BX+2],[BX+4]"),
    "_PRINT13_NUMBER": ((VAL, NUM), "FANTASIE.ASM:_print13_number, print_end"),
    "_PRINT13_NUMBER_CENT": ((VAL, NUM), "FANTASIE.ASM:_print13_number_CENT, print_end"),
    "_PRINT5": ((TXT, NUM), "FANTASIE.ASM:print_end [BX+2],[BX+4]"),
    "_PRINT5_NUMBER": ((VAL, NUM), "FANTASIE.ASM:_print5_number, print_end"),
    "_PRINT8": ((TXT, NUM), "FANTASIE.ASM:print_end [BX+2],[BX+4]"),
    "_PRINT8_NUMBER": ((VAL, NUM), "FANTASIE.ASM:_print8_number, print_end"),
    "_PRINT8_NUMBER_CENT": ((VAL, NUM), "FANTASIE.ASM:_print8_number_CENT, print_end"),
    "_RETURN_OF_THE_EVIL_SUPERMODE": ((IGN,), "SDEV.ASM handler filler"),
    "_RULLGARDIN_NED": ((TXT, NUM), "FANTASIE.ASM:_RULLGARDIN_NED (source text, stop row)"),
    "_RULLGARDIN_UPP": ((TXT, NUM), "FANTASIE.ASM:_RULLGARDIN_UPP (source text, stop row)"),
    "_SCROLL": ((TXT,), "FANTASIE.ASM:_SCROLL STARTSCROLL [BX+2]"),
    "_SETDECCOR": ((IGN,), "FANTASIE.ASM:_SETDECCOR MOVE DECCOR,[BX+2]"),
    "_SETLOOP": ((NUM, IGN), "table-local _setloop [BX+2], print_end filler"),
    "_SHOOT_AGAIN_ONN": ((IGN,), "table-local _SHOOT_AGAIN_ONN, NORMAL_END filler"),
    "_SHOW_SCORE": ((LBL,), "table-local _SHOW_SCORE MOV BX,[BX+2]"),
    "_SOUND_EFFECT": ((SND, IGN), "FANTASIE.ASM:_SOUND_EFFECT [BX+2],[BX+4]"),
    "_TILTED": ((IGN,), "STONES.ASM:_TILTED filler"),
    "_TOWER": ((NUM,), "STONES.ASM:_tower MOVE stannalinne,[BX+2]"),
    "_TOWEREND": ((IGN,), "STONES.ASM:_TOWEREND filler"),
    "_TSEND": ((IGN,), "PLAND.ASM:_TSEND filler"),
    "_TURNOFFOFFROADMODE": ((IGN,), "STONES.ASM handler filler"),
    "_TURNOFFSPECIALMODE": ((IGN,), "table-local handler filler"),
    "_TURNOFFTURBOMODE": ((IGN,), "STONES.ASM handler filler"),
    "_TURNONOFFROADMODE": ((IGN,), "STONES.ASM handler filler"),
    "_TURNONSPECIALMODE": ((IGN,), "table-local handler filler"),
    "_TURNONTURBO": ((), "SDEV.ASM:_TURNONTURBO endit"),
    "_TURNONTURBOMODE": ((IGN,), "STONES.ASM handler filler"),
    "_VAULTEND": ((IGN,), "STONES.ASM:_VAULTEND filler"),
    "_WAIT": ((NUM,), "FANTASIE.ASM:_WAIT MOVE SISA,[BX+2]"),
    "_WAITIFMULTI": ((NUM,), "FANTASIE.ASM:_WAITifmulti MOVE SISA,[BX+2]"),
    "_WAITJINGLE": ((IGN,), "FANTASIE.ASM:_WAITJINGLE filler"),
    "_WAITJINGLE2": ((IGN,), "FANTASIE.ASM:_WAITJINGLE2 filler"),
    "_WAIT_GAME_ON": ((IGN,), "table-local handler filler"),
    "_WELLEND": ((IGN,), "STONES.ASM:_WELLEND filler"),
    # matrix_reachability.Source names a bare CLEARIT this way; it has no
    # operand words.  The extractors map it to the table's real clear op.
    "_CLEAR?": ((), "CLEARIT macro, zero operand words"),
}

UNKNOWN_KIND = "unknown"

_EXPR_RE = re.compile(r"[A-Za-z_0-9+*/() -]+")
_IDENT_RE = re.compile(r"\b[A-Za-z_][A-Za-z_0-9]*\b")
_HEX_RE = re.compile(r"\b([0-9][0-9A-Fa-f]*)[Hh]\b")
_IGNORED_TOKENS = {"?"}


class OperandError(ValueError):
    """A reachable source operand could not be resolved to a concrete value."""


def kinds(op: str):
    entry = OPCODES.get(op)
    return entry[0] if entry else None


def origin(op: str) -> str:
    entry = OPCODES.get(op)
    return entry[1] if entry else "unknown opcode"


def numeric_indices(op: str):
    return tuple(i for i, k in enumerate(kinds(op) or ()) if k == NUM)


def classify(op: str, index: int) -> str:
    k = kinds(op)
    if k is None:
        return UNKNOWN_KIND
    if index >= len(k):
        return UNKNOWN_KIND
    return k[index]


def is_placeholder(text: str) -> bool:
    return text.strip().lower() in _IGNORED_TOKENS


def expression_grammar_ok(text: str) -> bool:
    """True when the operand is within the assembler expression grammar."""
    return bool(_EXPR_RE.fullmatch(text.strip()))


def resolve_expression(text: str, constants: dict) -> int:
    """Evaluate an operand exactly as the retail assembler would.

    Only the operators actually present in the retail matrix operands are
    accepted: integer literals (decimal and ``NNh`` hex), the ``SW`` memory
    stride and table equates, ``+ - * /``, parentheses and unary signs.  This
    is deliberately a small closed grammar, not a general evaluator; anything
    outside it raises ``OperandError`` so the checker fails loudly.
    """
    text = text.strip()
    if not expression_grammar_ok(text):
        raise OperandError(f"operand outside retail assembler grammar: {text!r}")
    tokens = re.findall(r"[0-9][0-9A-Fa-f]*[Hh]|[0-9]+|[A-Za-z_][A-Za-z_0-9]*|[()+*/-]", text)
    if "".join(tokens) != re.sub(r"\s+", "", text):
        raise OperandError(f"malformed operand {text!r}")
    at = 0
    def atom():
        nonlocal at
        if at >= len(tokens):
            raise OperandError(f"incomplete operand {text!r}")
        t = tokens[at]; at += 1
        if t in ("+", "-"):
            v = atom()
            return v if t == "+" else -v
        if t == "(":
            v = addition()
            if at >= len(tokens) or tokens[at] != ")":
                raise OperandError(f"unclosed operand {text!r}")
            at += 1
            return v
        if re.fullmatch(r"[0-9][0-9A-Fa-f]*[Hh]", t):
            return int(t[:-1], 16)
        if t.isdecimal():
            return int(t)
        if t.upper() in constants:
            return constants[t.upper()]
        raise OperandError(f"undefined source symbol {t!r} in {text!r}")
    def product():
        nonlocal at
        v = atom()
        while at < len(tokens) and tokens[at] in ("*", "/"):
            op = tokens[at]; at += 1; rhs = atom()
            if op == "*": v *= rhs
            else:
                if rhs == 0: raise OperandError(f"division by zero in {text!r}")
                # Assembler integer division truncates toward zero.
                sign = -1 if (v < 0) != (rhs < 0) else 1
                v = sign * (abs(v) // abs(rhs))
        return v
    def addition():
        nonlocal at
        v = product()
        while at < len(tokens) and tokens[at] in ("+", "-"):
            op = tokens[at]; at += 1; rhs = product()
            v = v + rhs if op == "+" else v - rhs
        return v
    value = addition()
    if at != len(tokens):
        raise OperandError(f"trailing operand tokens in {text!r}")
    return value


def numeric_values(op: str, args, constants: dict) -> dict:
    """Resolve every NUM/FLG operand of one command to an int.

    Returns ``{operand_index: value}``.  Raises ``OperandError`` when a numeric
    operand cannot be resolved, which is exactly the class of defect this
    audit exists to eliminate.
    """
    out = {}
    for index, kind in enumerate(kinds(op) or ()):
        if kind not in (NUM, FLG):
            continue
        if index >= len(args):
            raise OperandError(f"{op} missing numeric operand {index}")
        text = args[index].strip()
        if is_placeholder(text):
            # Only IGN placeholders may be '?'; a NUM operand never is.
            raise OperandError(f"{op} numeric operand {index} is a '?' placeholder")
        try:
            out[index] = resolve_expression(text, constants)
        except OperandError as exc:
            raise OperandError(f"{op} operand {index} {text!r}: {exc}") from exc
    return out


_EMPTY_JINGLE_RE = re.compile(r"^S_Empty\s+DB\s+(\d+)\s*,", re.I)


def empty_jingle_position(lines) -> int | None:
    """Recover the source value of the ``EMPTYJINGLE`` equate.

    ``EMPTYJINGLE`` is referenced by the retail matrix streams but is not
    declared in the surviving ``.ASM`` because it comes from a lost include.
    Every table declares ``S_Empty DB <position>,<repeat>,<priority>`` and
    ``LASTJINGLE``/``EMPTYJINGLE`` is the music position of that empty jingle;
    all four tables agree with the value the shipped native port already used.
    """
    for line in lines:
        m = _EMPTY_JINGLE_RE.match(line.strip())
        if m:
            return int(m.group(1))
    return None


def seed_empty_jingle(constants: dict, lines) -> bool:
    value = empty_jingle_position(lines)
    if value is None:
        return False
    constants["EMPTYJINGLE"] = value
    return True


# ---------------------------------------------------------------------------
# DB byte expressions.
#
# Retail DATA declares text bytes with assembler expressions that mix character
# literals and arithmetic, e.g.
#
#     xb_at_20_text DB '7'+2,'7 LITES EXTRA BALL',0
#     or_at_10_text DB ' ','7'+1,'7 LITES OFF ROAD',0
#     skilltext     DB '  MONEY MANIA AT ',6+'7',' ',0
#     ZEROQ         DB '0'+7,0
#
# Each comma-separated element is one byte (or one quoted string).  A tokenizer
# that splits on quotes turns '7'+2 into two bytes, 0x37 and 0x02, so the digit
# the original encodes becomes garbage.  These helpers evaluate the whole
# element with the assembler's byte semantics.
# ---------------------------------------------------------------------------

_CHAR_BYTE_RE = re.compile(r"'([^'])'")
_DUP_RE = re.compile(r"^(.+?)\s+DUP\s*\((.*)\)$", re.I | re.S)


def _dup_count(text: str, constants: dict) -> int:
    """DUP repeat count; retail writes it as an expression (16*16, 576*2)."""
    count = resolve_expression(text, constants)
    if count < 0:
        raise OperandError(f"negative DUP count {text!r}")
    return count


def split_db_operands(text: str):
    """Split a DB operand list on top-level commas, keeping quotes intact."""
    parts, cur, depth, i = [], [], 0, 0
    while i < len(text):
        c = text[i]
        if c in "'\"":
            j = text.find(c, i + 1)
            if j < 0:
                raise OperandError("unterminated DB quote")
            cur.append(text[i:j + 1])
            i = j + 1
            continue
        if c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
        if c == "," and depth == 0:
            parts.append("".join(cur))
            cur = []
            i += 1
            continue
        cur.append(c)
        i += 1
    if "".join(cur).strip():
        parts.append("".join(cur))
    return [p.strip() for p in parts]


def db_element_bytes(element: str, constants: dict):
    """Bytes produced by one DB element.

    ``'abc'`` yields the string's bytes.  Anything else is one assembler byte
    expression in which a character literal is its code point; the result is
    truncated to a byte exactly as the original linker does.
    """
    element = element.strip()
    if not element:
        return []
    if re.fullmatch(r"'[^']*'|\"[^\"]*\"", element):
        return list(element[1:-1].encode("latin1"))
    expr = _CHAR_BYTE_RE.sub(lambda m: str(ord(m.group(1))), element)
    value = resolve_expression(expr, constants)
    return [value & 0xFF]


def db_bytes(text: str, constants: dict):
    """All bytes a whole DB operand list produces (DUP aware).

    Strict: an element the assembler cannot evaluate raises, which is how the
    extractors have always treated unparseable DATA (``?`` uninitialised rows
    are dropped rather than materialised as zeros).
    """
    out = []
    for element in split_db_operands(text):
        dup = _DUP_RE.match(element)
        if dup:
            count = _dup_count(dup[1], constants)
            values = []
            for item in split_db_operands(dup[2]):
                values.extend(db_element_bytes(item, constants))
            out.extend(values * count)
            continue
        out.extend(db_element_bytes(element, constants))
    return out


def db_byte_count(text: str, constants: dict) -> int:
    """Byte count of a DB operand list, for length-only consumers.

    Each element is one byte regardless of whether it can be evaluated, so a
    text-length table never loses an entry to an uninitialised ``?`` row or a
    symbol from a lost include.  Byte expressions still count as one byte.
    """
    total = 0
    for element in split_db_operands(text):
        dup = _DUP_RE.match(element)
        if dup:
            inner = sum(db_byte_count(item, constants)
                        for item in split_db_operands(dup[2]))
            total += _dup_count(dup[1], constants) * max(inner, 0)
            continue
        element = element.strip()
        if re.fullmatch(r"'[^']*'|\"[^\"]*\"", element):
            total += len(element[1:-1])
            continue
        total += 1
    return total
