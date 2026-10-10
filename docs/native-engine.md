# Native engine contract (ABI 1)

There is one Go gameplay implementation. Foreign hosts use `cmd/pfengine/abi.h`;
Go desktop adapters may continue using `internal/source.Runner`. Both paths use
that Runner, `frontend.Runtime`, and the existing table/presentation/audio code.
The native macOS ARM64/x86_64 host is in `hosts/macos`; original-backed checks
and owner physical-Mac acceptance are complete as recorded in
[macOS validation](macos-validation.md). See [macOS build and usage](macos.md).
The Android host uses the same serialized contract; see [Android host](android.md).

```text
Go source engine / frontend.Runtime / source.Runner
    +-- direct Go desktop adapter -> Win32
    +-- direct Go desktop adapter -> SDL/Linux
    +-- internal/engine -> stable C ABI -> AppKit/macOS (ARM64/x86_64 native host; accepted)
    +-- internal/engine -> stable C ABI -> Android native host
```

Build on a machine with Go and the target's C compiler/SDK:

```sh
./tools/build_engine.sh c-archive bin/libpfengine.a
./tools/build_engine.sh c-shared bin/libpfengine.so
```

Distribute `abi.h` with the library. Consumers do not need the Go-generated
header or Go object types. Only fixed-width scalars, opaque integer handles,
borrowed buffers and one synchronous C callback cross the boundary. ABI version
is 1; incompatible changes require a version change. Zero is never a handle;
destroy invalidates a handle permanently. Invalid handles return PF_INVALID,
overlapping/reentrant calls return PF_BUSY, runtime failures return PF_ERROR.
Destroy frees the instance even if settings persistence returns PF_ERROR.
Create copies UTF-8 paths and returns a bounded error message on load failure.

The optional `pf_engine_gamepad(handle, kind, a, b)` addition retains ABI 1.
Kinds describe standard button edges, normalized triggers, a connection hold
snapshot or disconnection; the exact contract is in `abi.h`. The shared Go
mapper selects actions from the authoritative frontend mode. Controller holds
combine with existing keyboard/touch holds, and lifecycle clears suppress
background actions. Older hosts can continue using existing calls.

## Time and input

The host passes nonnegative monotonic nanoseconds from a single clock epoch,
starting at creation. Backward values are rejected. The Runner owns integer
60 Hz intro and 71 Hz table deadlines, selected before each update. Every due
source task runs; a slow rendering wake cannot discard physics or PCM. A vsync,
CADisplayLink or Choreographer callback may wake the host but does not determine
source cadence. State Tick counts scheduled runtime updates, excluding immediate
focus notifications; frontend and table counters retain their original semantics.

`set_action` accepts LEFT, RIGHT, SPRING and TILT holds. Repeated identical holds
cannot invent releases. SPRING's falling edge produces one canonical keyboard
release. `release` supports the original separate release make semantics.
`key` takes original logical DOS make codes for menus, pause and initials/cheats;
the host translates native keys, preserves sided modifiers, combines independent
physical contributors, and suppresses OS repeat. It must not send OS keycodes.
Transient release/fire/key edges survive between wakes and are consumed once.

`plunger_delta` accepts relative vertical counts; eight counts equal one source
adjustment, with at most one adjustment per task. Both mouse and touch must use
this path and the existing 0..32 spring. Excess movement cannot charge faster.
Invalid spring input is discarded and its remainder cleared. One shared frontend
check consults the canonical source SpringValid flag as well as chute context
for both engine and Go desktop adapters. Fire is a make
edge, scheduled for the following task by the original Runner. A held mouse or
touch button must not repeatedly call fire. Hosts use the mouse-active state bit
for local cursor/relative-input policy; no absolute pointer position is involved.

Suspend applies the shared frontend focus-pause request immediately, clears
holds/edges/mouse remainder/scheduled fire, and suspends Runner advancement.
Resume clears stale input and re-anchors to the supplied monotonic timestamp.
It leaves gameplay paused until an ordinary logical resume make. No elapsed
background time catches up. Repeated active Resume calls do not clear holds.
Desktop queued focus notifications and immediate native suspend use the same
Runner pause helper and frontend semantics. Destroy uses shared frontend close,
including settings persistence, even when the engine is suspended.

## Ownership and threads

All operations on an instance run on one host engine execution path. The C bridge
rejects simultaneous/reentrant access, including calls made by the PCM sink.
Different instances have independent locks. The registry protects handle lookup
and destruction; native audio device callbacks never call any engine function.

Go Frame returns a read-only borrowed RGBA raster until the next mutating engine
operation or Frame call. The C frame API stages that raster in engine-owned C
memory, growing only when capacity is insufficient. It performs one raster copy
per retrieval, never exposes a retained Go pointer, and never requires per-frame
host malloc/free. Pixels are top-down RGBA8 with the returned byte stride and
source dimensions, including full-table geometry. The C view survives Advance
and is valid until the next frame retrieval or destroy. Do not mutate or free it.
The underlying existing renderers retain their allocation behavior.

Advance delivers every nonempty source PCM chunk synchronously. Format is 48 kHz,
stereo, signed 16-bit little-endian interleaved samples; byte counts divide by 4.
PCM is borrowed only during the sink callback. Copy it into a bounded host-owned
ring before returning. C never retains that Go pointer. A callback may apply
backpressure while an active audio device drains the ring; it must not reenter
engine calls. Device callbacks read only that ring and may fill silence on an
underrun. Hosts own capacity, underrun measurement, flushing, audio focus, device
route recovery and suspended-device policy. PCM production follows source tasks,
independently of display refresh or audio-device callback frequency.

## Files, display and native responsibilities

Create receives a validated data location and an explicit writable state
location. The shared decoders load all four tables and modules. Shared settings
and high-score stores use state paths, with installation data as the existing
first-load seed; all later writes stay in state. A host must choose Application
Support on macOS or app-private storage on Android, preserving explicit paths
for developer/portable workflows. Hosts own native file selection and staged
copying: NSOpenPanel or Android SAF, required original filenames only, validate
through the shared Go loader before adopting app-owned copies. SAF content URIs
are never passed as filesystem paths. No commercial files ship with the library.
The macOS host stages/imports files through NSOpenPanel and validates through
this ABI, with native transaction tests. Mac host acceptance is recorded in
[macOS validation](macos-validation.md); Android import remains unimplemented.

Hosts own windows/surfaces, focus, native fullscreen, keyboard/touch translation,
pointer IDs and cancellation, nearest-neighbour framebuffer upload, aspect and
letterboxing, audio buffering and devices. Source presentation stays in Go.
AppKit/GLES wrappers must not implement table rules, matrix programs or tracker
progression. macOS build/runtime acceptance must precede Android development.

## Conformance

`go test ./internal/engine ./internal/source` compares direct Runner and Go boundary
on all four tables, SOFT/full-table OFF, complete gameplay/matrix/task checkpoints,
PCM bytes/counts, frame hashes, original cheat strings, keyboard/mouse launch,
flippers/tilt, drains, player/new-ball handoff, suspend/resume and slow wakeups.
Original-backed tests skip explicitly without external user-supplied assets.

To exercise actual C calls against the direct Runner oracle:

```sh
./tools/build_engine.sh c-shared bin/libpfengine.so
./tools/go.sh build -o bin/pftrace ./cmd/pftrace
python3 tools/test_engine_abi.py --library "$PWD/bin/libpfengine.so" \
  --oracle "$PWD/bin/pftrace" --data /path/to/legally-obtained/originals
```

The C replay compares source tasks, table/mode/lifecycle flags, every PCM chunk
hash/count and selected frame hashes/dimensions, and tests invalid handles,
backward clocks, invalid actions, destruction and callback reentry rejection.
No commercial fixture is generated in the repository.
