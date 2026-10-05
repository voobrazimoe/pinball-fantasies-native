# DOS initials character audit (2026-10-05)

The owner's four local retail TABLE PRGs match the pinned SHA-256 inventory in
`analysis/game-inventory.json`. All four accept A–Z and Space (stored as `*`).
The normal top-row numeric makes are rejected. No initials semantics or numeric
touch buttons were added. This finding applies to these four pinned DOS builds;
it does not establish what another DOS revision or the Jaguar release accepts.

## Linked path inspected in every PRG

Addresses below are **file offsets**, not CS-relative instruction addresses.
Linked code starts at file offset `0x300`.

| Table | DS file base | ALFA_KEYS | GET_IT_FROM_KEYBOARD | READ_KEYBOARDET | IRQ handler |
| --- | --- | --- | --- | --- | --- |
| Party Land | 0x19d40 | 0x1d396 | 0x8ad | 0x909 | 0x4134 |
| Speed Devils | 0x18ee0 | 0x1c5cb | 0x772 | 0x7ce | 0x391a |
| Billion Dollar Gameshow | 0x18b60 | 0x1ba70 | 0x740 | 0x79c | 0x33c7 |
| Stones 'N Bones | 0x166d0 | 0x1a2f4 | 0x818 | 0x874 | 0x48e5 |

The IRQ reads port `0x60`, preserves AL while acknowledging port `0x61`, tests
bit 7 for a break, then writes the raw make to SCAN_CODE. There is no numeric
translation before that store. GET_IT_FROM_KEYBOARD inserts the score row,
sets the remaining count to three, clears SCAN_CODE to `0xff`, and installs
READ_KEYBOARDET for its following task visit.

READ_KEYBOARDET reads SCAN_CODE, skips `0xff`, clears the slot, loads BX with
ALFA_KEYS' DS-relative address and executes XLAT. A zero result takes the exit
branch before either name buffer write or count decrement. A nonzero result
is stored directly, with no subsequent character lookup before storage.

| Top-row key | DOS make | Lookup result in each of TABLE1–4 |
| --- | --- | --- |
| 1 | 0x02 | zero / rejected |
| 2 | 0x03 | zero / rejected |
| 3 | 0x04 | zero / rejected |
| 4 | 0x05 | zero / rejected |
| 5 | 0x06 | zero / rejected |
| 6 | 0x07 | zero / rejected |
| 7 | 0x08 | zero / rejected |
| 8 | 0x09 | zero / rejected |
| 9 | 0x0a | zero / rejected |
| 0 | 0x0b | zero / rejected |

## Original-backed routine execution

`tools/verify_dos_initials.py --data /absolute/path/to/owned/DOS/data` reads the
PRGs without changing them, verifies their pinned hashes, and uses Capstone and
Unicorn to run their linked 16-bit IRQ/setup/input instructions in disposable
memory. The original make is supplied through port `0x60`. Execution stops
after the raw scan store and after name/count mutation, before rendering.
Only addresses and behavioral results are printed; no PRG payload is exported.

The local run passed for all four tables with Capstone 5.0.9 / Unicorn 2.1.4.
Unicorn's JIT needed execution outside the desktop sandbox. Each of the ten
numeric makes cleared the scan slot but left the destination, pointer and
remaining character count unchanged. Setup also cleared a queued letter make.

| Sequence | Stored characters | Remaining characters |
| --- | --- | --- |
| A1B | AB | 1 |
| 7UP | UP | 1 |
| R2D | RD | 1 |
| ABC | ABC | 0 |
| three Space makes | *** | 0 |

These mixed sequences do **not** form complete three-character DOS names.
`?` in the oracle output denotes an untouched sentinel slot, not a DOS glyph.
This is original instruction execution for the input routines, not a full DOS
boot/gameplay run or physical Android acceptance.

The shared Go test compares all 128 lookup entries against **each** original
PRG. Deterministic tests also reject every numeric make and verify the three
mixed sequences remain in Initials without saving. Both checks passed locally
using temporary copies of the four originals; the copies were removed and
owner input hashes were unchanged.

Android keeps its contextual QWERTY keyboard and automatic disappearance on
ENTRY_WAIT. Hardware digits do not acquire new game semantics. Touch buttons,
including selector cards and the menu icon, no longer take hardware keyboard
focus. A6 remains NOT TESTED pending the physical owner pass.

Validation also passed the asset-free frontend package, A3 geometry/semantic UI/
keyboard parity/native session tests, public source checks, and Android main plus
instrumentation Java compilation. The new View focus assertions compiled but
were not executed on a device in this turn.
