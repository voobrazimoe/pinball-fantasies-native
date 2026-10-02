# Matrix numeric presentation audit

All four tables use typed numeric operands. Runtime rendering preserves BCD
formatting, centering and mutable text writers. The current private source audit
classifies 36/26/26/31 numeric or custom sites and 6/10/8/10 mutable writers for
Party, Speed, Gameshow and Stones respectively.

Run `python3 tools/matrix_numeric_audit.py --check` with the owner's originals
and external historical reference to recompute these assertions. Detailed site
listings are private output in `.private-cleanup/generated/`, rather than
committed source-derived operand inventories. Public structural, synthetic and
formatting tests run without the historical checkout. Original-backed text and
record tests skip when their genuine external inputs are absent.

The remaining translated programs still require migration or rights review;
see [public preflight](public-preflight.md) and [runtime DATA](../docs/runtime-data.md).
