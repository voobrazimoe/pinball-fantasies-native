# Architecture

The project has one shared native game implementation with platform-specific host backends.

`cmd/pinballfantasies` starts the frontend and table runtime. `internal/platform` owns windowing, input, timing, audio-device queues and writable storage. Linux uses SDL2; Windows uses native Win32, GDI and waveOut APIs.

`internal/assets` and `internal/physics` validate and decode the user-supplied DOS data. `internal/partyland`, `internal/speeddevils`, `internal/gameshow` and `internal/stones` implement table-specific rules and state machines in Go. Shared scoring, task scheduling and timing helpers live under `internal/tablelogic`.

`internal/audio` decodes and renders the original tracker/module data and effects. `internal/presentation` renders the matrix display and table presentation from typed native state plus records decoded from the supplied PRG data.

Multiplayer is implemented as a hot-seat session around a single active table engine. Up to eight player records preserve the source-persistent per-player fields; ball-local and transient table state is reset at the same lifecycle boundaries as the original game. Player handoff loads the incoming record before the next launch, including its pre-launch matrix presentation.

Deterministic simulation advances from source-time ticks independently of the host audio queue and window refresh. Platform presentation and audio buffering therefore do not define gameplay time.

The public repository contains no DOS/x86 execution path. Original game files are read only as data containers and are never executed.