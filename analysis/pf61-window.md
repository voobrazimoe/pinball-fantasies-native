# PF6.1 — physical window ownership

`ShowFrontend` previously called `SDL_SetWindowSize` whenever an original
framebuffer changed dimensions. It now creates one resizable window and changes
only the renderer logical size and streaming texture at presentation transitions.
The original 640x240 selector retains its doubled-scanline 640x480 aspect.
SDL's logical renderer supplies letterboxing/pillarboxing on resize.

The initial size uses SDL display 0 usable bounds in window coordinates, with an
85% fit budget. Select the largest integer 320x383 gameplay scale from 1 through
4 that fits, and accommodate the largest 640x480 frontend presentation. Clip to
the fit budget on small displays. Typical full usable bounds yield 640x766 at
1080p, 960x1149 at 1440p, and 1280x1532 at 4K. This runs only before window
creation. Renderer output/HiDPI drawable dimensions never feed this policy.
If usable bounds are unavailable, the conservative fallback is 800x600.

Focused tests cover representative and constrained bounds, selector aspect, and
structurally prohibit physical resizing/window recreation in frontend code.
`tools/smoke_pf6.py` now manually resizes through X11 and verifies both the same
window identity and unchanged physical dimensions after each transition.

The 383-row native presentation includes the existing doubled score panel;
this is the largest normal gameplay framebuffer, so it sets the fit budget.
Static and direct gameplay viewers use the same host initial-size policy.

Final real X11 run on the available 4K desktop opened at 1280x1532. A manual
XResizeWindow request produced 1208x1486 window-manager dimensions. The same
window ID and those exact queried dimensions survived selector text/preview,
F3/F4, Party Land attract/play/pause/resume/quit question/abort, Speed Devils
attract/play/pause/resume/quit question/abort, and return to selector. The smoke
passed with dummy audio and zero queue resets/empty-queue observations. It uses
relative resize requests and does not require a particular desktop resolution.
High-score states are covered by native lifecycle tests and the structural ban
on physical resizing; the bounded X11 smoke does not drive initials entry.

Files changed: internal/platform/sizing.go, sizing_test.go, frontend.go,
window.go and physics.go; tools/smoke_pf6.py; this report and README.md.
