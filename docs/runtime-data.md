# Runtime DATA boundary

The public architecture is native Go plus owner-supplied PRG/MOD data. Native
code reads structured records; it never executes the original x86 routines.
The project license covers native code and project-authored material, not the
original commercial data or historical reference source.

## Records now read from PRGs

`presentation.New` reads matrix text buffers, mutable text seeds, score seeds,
animation headers/frame pointers/durations and lamp flash records. Address and
length maps identify records in the supported retail builds; they contain no
matrix prose, animation timings, or flash schedules. Each display owns its text
buffers, so a table's native writers cannot mutate the input or another session.
Table loaders validate the supplied builds against their existing fingerprints.
Record reads reject truncation rather than substituting empty game data.

All four schedulers read tracker jingle position/repeat/priority bytes from
those decoded buffers. Party and Speed use the shared decoded animation timing
and scroll lengths. Gameshow and Stones read lamp DAC packets and gate masks
from PRG records, including interleaved mask rows. The intro reads both 120-byte
instruction buffers from INTRO.PRG. Existing matrix fonts, graphics, samples,
physics records and playfields continue to come from the owner's originals.

A nil PRG remains supported for metadata-only structural tests. It provides
registered identities and typed legacy commands, but no fabricated text or
animation timing. Tests exercising original records require their external PRG.
Synthetic tests exercise bounds, ownership and record shapes without originals.

The location generator searches for exact accepted bytes in the retail inputs
and verifies every record before exporting only addresses. Some labels share
identical bytes and may use the same record address; each mutable buffer is
copied separately. Hash-only parity manifests pin all pre-migration records;
private independent ASM checks hydrate these addresses from PRGs before comparing
all 1,106 declarations and native text writers. This retains source assertions
without a committed full-content dump. Numeric runtime expressions are not parsed.

Six unlinked legacy filename declarations per table and Speed's unused legacy
three-byte game-over jingle remain explicit small exceptions in the shared
metadata. Their retail bytes differ or are absent. These are source-guided metadata exceptions; matrix text is read from PRGs. The jingle exception
preserves the previously accepted source-guided audio behavior.

## Sidebar font provenance

INTRO's writer reads the PC ROM at F000:FA6E. Its font is not supplied by the
game, and the removed 1,024-byte oracle font is absent as a complete record from
all five supplied PRGs. Retaining that font was not a defensible game-data path.

`internal/assets/frontend_bios.go` now draws a minimal uppercase 5x7 alphabet,
digits and required punctuation inside the existing 8x8 cells. These simple
strokes were independently authored during this cleanup. No BIOS, emulator ROM,
font library, external bitmap, or external font design was used as a reference.
This project-authored code is covered by the project MIT license. The sidebar letters visibly differ from the DOSBox-X
ROM. Character cadence, placement, palette, clearing, restoration, instruction
text and all game matrix fonts retain the accepted behavior. The dedicated
raster test explicitly pins this intentional replacement, while retaining the
unaffected original selector hash and sidebar restoration assertions.

## Private reference generation

Large content/operand reports, capture timelines and state dumps were removed
from the tracked candidate. Their complete baseline copies remain locally in
ignored `.private-cleanup/baseline/`; external/private archives were not changed.

Party rule, timing and trajectory tests generate expected results into their
temporary directory using the owner's PRG/MOD/reference inputs. Missing inputs
skip clearly; present/corrupt inputs and generator failures fail. Generators do
not rewrite native code when invoked by a test. Manual generators and detailed
operand/numeric reports default to `.private-cleanup/generated/`. `--check`
recomputes and asserts source invariants without requiring committed report dumps;
reachability still checks freshness of the native registration declarations.

## Remaining program migration

The shared `content.go`, Party `timing_data.go` and three other table `content.go`
files still contain substantial translated command programs, source operand
expressions, effect data and tracker cue summaries. The owner approved publication of these source-guided native representations;
historical reference source and original commercial payload remain outside MIT.

A private structured-word investigation finds candidate locations for most
matrix groups, but it does not establish full topology/operand parity. Repeated
programs, macro expansions and source/retail differences require further layout
work. Unverified candidates have not replaced runtime commands.

Speed lamp start/count headers now come from PRG bytes instead of generated
metadata. The 67 address registrations contain only header locations; each game
owns its packet copy. Bounds checks reject truncated packets and DAC overflow.

Stones reads four terminated area lists (44 rectangles) from PRG records at
startup. Linked handler words select registered native Go behaviors; the words
are never executed. The layouts contain only offsets/counts and address-to-native
handler registration. Each game owns its decoded lists. Hash-only assertions
compare rectangle values, handler identities and ordering to the private baseline.
No new compatibility constant or replacement program dump was introduced.
