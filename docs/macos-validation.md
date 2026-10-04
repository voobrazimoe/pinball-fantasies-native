# macOS release validation

## Accepted on 2026-10-04

The owner tested the Intel build on a 2017 MacBook Pro using its built-in
keyboard. After the Core Animation presenter and Shift-side correction, the
owner confirmed that scrolling and flippers are responsive and the game runs
smoothly. The owner accepts the macOS version as ready and requested clean
ARM64/Intel builds and a Git push. PR #1 is ready for review; merging is a
separate repository operation.

The accepted gameplay implementation is commit `865d300`. The final packaging
commit adds this record and Intel CI packaging without changing gameplay.

## Delivered behavior

- Native AppKit `.app`, minimum macOS 13, separate ARM64 and x86_64 packages.
- Existing Go engine and ABI 1; shared gameplay, physics and source timing intact.
- Core Animation presents immutable RGBA CGImages, nearest-neighbour scaling,
  preserved aspect ratio and black bars, without implicit animations or forced
  synchronous drawing inside input handlers.
- Main-thread engine serialization, strict 120 Hz wake-ups, source-paced frame
  retrieval and immediate service of already-due source work on keyboard events.
- Independent sided Shift/Control/Option flippers; Z/slash and arrows are
  alternatives. Command is reserved for native shortcuts.
- View → Swap Left/Right Shift persists per Mac and corrects the observed
  reversed Shift events. Other contributors retain their normal sides.
- Application Support storage, validated import of user-owned originals,
  bounded AudioUnit PCM ring and focus/minimize/sleep handling.
- Optional timing logs record revision, architecture, renderer, Shift mapping,
  modifier events and host timing. Normal launches do not create a log.

## Validation evidence

| Check | Result |
| --- | --- |
| Apple SDK ARM64 executable/archive/shared-library builds | PASS |
| Intel executable architecture, link, ad-hoc signature and package | PASS; cross-built on ARM64 |
| Native input/logic, modifier duplicate and focus regressions | PASS |
| Shift swap, mixed contributors and per-Mac preference persistence | PASS |
| Audio render callback, stereo ring/concurrency, bounded storage/import | PASS |
| Layer colours/orientation, immutable old frames, nearest scaling, resize/selector geometry | PASS |
| Real loaded-C-library/direct-Go conformance | Eight replays PASS, four tables in SOFT/OFF, 2,714 checkpoints each, frame/PCM/state parity |
| Apple-linked original-backed native journeys | 32 PASS, four tables × four scroll modes × two resolutions, 708 ticks each |
| Local live ARM64 window/fullscreen/restoration/minimize/input | PASS |
| Owner Intel gameplay/scroll/input acceptance | PASS with Shift swap enabled |
| Exact public bundle allowlist and commercial-payload hygiene | PASS |

The clean release build reruns the local original-backed validation via
`tools/validate_macos_originals.py`; external inputs are staged into temporary
storage and their hashes are checked unchanged. Its clean archived checkout runs
the same asset-free Go suite as public CI; original-backed coverage is provided
by the separate real ABI replays, native journeys and personalvalidate run.
Historical independent fixture tests requiring absent private reference files
are not included in this release gate. CI uses no commercial originals
and runs asset-free tests, bounded AppKit launch, ARM64 build and Intel cross-build.

## Diagnosis and resolution

The first Intel timing logs showed timer gaps around 35 ms. A later build with
synchronous input-triggered drawing showed drawing up to 31.403 ms and input
queue delay up to 316.520 ms. Shared source updates remained fast. Replacing the
per-frame CGContext drawing path with Core Animation eliminated the long host
stalls in the owner's next capture (`544ce58`): 59–61 layer updates per interval,
input queue delay at most 7.106 ms, typical maximum timer gaps around 8.7 ms.
Layer submission timing does not measure physical display scanout.

Native modifier snapshots make repeated events idempotent. In the requested
left-then-right Shift capture (`12e4f1e`), macOS reported key 60 / flags 0x20104 /
right hold first, then key 56 / flags 0x20102 / left hold. The persistent Shift
swap option (`865d300`) compensates for that observed order. The owner then
confirmed both responsiveness and correct control.

Observed Parsec modifier events omitted side identity; Z/slash and arrows remain
available for such input sources. Native keyboard acceptance does not establish
sided modifier support for remote events that omit those details.

## Distribution details

Packages contain the public app and optional timing launcher, without commercial
originals. They are ad-hoc signed; Developer ID signing and notarization have not
been performed. Settings and scores stay outside the bundle. The owner accepts
the release based on the reported gameplay checks; this record does not claim
exhaustive testing of every audio route, display topology or sleep duration.

[Build and usage documentation](macos.md).

## Local personal macOS packages

The local personal builder validates the supplied installation, copies only the
11 required PRG/MOD files and optional PINBALL.CFG into Resources/Data, re-signs
both architectures and writes ignored local ZIP packages. Production first-run
storage automatically imports bundled originals using the same validated,
staged import path as manual import. Existing valid Data and State are preserved.
Original-backed native tests PASS for automatic first import, relaunch with an
existing import, unchanged bundled originals and all 32 gameplay journeys.
Public bundle allowlist/source checks remain unchanged; CI never packages data.
