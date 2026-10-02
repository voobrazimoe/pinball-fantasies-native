# Architecture

`cmd/pinballfantasies` selects the frontend or one of four tables.
`internal/platform` owns windows, input, audio queues, timing and writable storage.
Linux uses SDL2; Windows uses native operating-system APIs.

`internal/assets` and `internal/physics` validate and decode user-supplied data.
`internal/partyland`, `speeddevils`, `gameshow` and `stones` implement table rules
and cooperative tasks as native Go. `internal/tablelogic` shares decimal scoring,
tracker cue clocks and task scheduling. `internal/audio` decodes and renders tracker
modules. `internal/presentation` draws matrix pixels from typed commands and
external text, fonts and animation records. Mutable buffers belong to each game.

Deterministic simulation advances by table syncs, independently of the host audio
queue. Private validation compares original-derived records, reachable commands,
operands, writers, state and frame hashes. Public validation exercises bounded
synthetic records and native algorithms without committing original payloads.

Five generated program/effect declarations remain outside the completed runtime
boundary. Their presence currently prevents approval of the proposed public source
snapshot. Excluding development reports alone cannot resolve a runtime dependency.
See `runtime-data.md` and the private preflight report for the remaining work.
