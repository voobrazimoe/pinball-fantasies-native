Pinball Fantasies native - Linux x86-64 AppImage

Put your legally obtained original game files beside the AppImage:
INTRO.PRG, TABLE1.PRG through TABLE4.PRG, INTRO.MOD, MOD2.MOD,
TABLE1.MOD through TABLE4.MOD. PINBALL.CFG is an optional legacy settings seed;
missing or malformed settings use native defaults. TABLE1.HI through TABLE4.HI are optional legacy score seeds;
without them, native factory defaults are used. No original game assets are included.

Make the AppImage executable and run it. No installation or root is needed:
chmod +x PinballFantasies-x86_64.AppImage
./PinballFantasies-x86_64.AppImage
If FUSE is unavailable, add --appimage-extract-and-run before game arguments.

The package contains SDL2, its ELF dependencies and a matching libc/loader.
A working X11/Wayland desktop and audio server/device are still required.
Losing focus pauses gameplay. After returning, press P (or a gameplay key) to resume.
Native options, high scores and native.log go to userdata beside the OUTER
AppImage, not its temporary mount. If that directory cannot be written, state
goes to the user config directory (XDG_CONFIG_HOME or ~/.config)/pinballfantasies.
Original supplied files are read only and are never updated in place.
Explicit overrides: -data-dir PATH -config-dir PATH -high-score-dir PATH.

The normal native development build remains available:
./tools/go.sh build -o bin/pinballfantasies ./cmd/pinballfantasies
Pass -data-dir . when running that build from the source checkout.
