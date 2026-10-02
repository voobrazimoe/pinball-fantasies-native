# Supplied game data inventory

Inventory date: 2026-09-30. Original files remain in the installation root; no game assets are embedded or copied into public builds. Full SHA-256 values and sizes are in `game-inventory.json`. The four `.HI` entries describe factory records from INTRO.ASM, not mutable local DOS score files.

| File | Bytes | Probable purpose and evidence |
|---|---:|---|
| ADLIB.SDR | 10966 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| GUS.SDR | 10502 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| INTERNAL.SDR | 10711 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| INTRO.MOD | 252870 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| INTRO.PRG | 345678 | VERIFIED: MZ executable container; INFERRED: intro/table program by START.ASM filenames |
| MOD2.MOD | 55394 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| NOSOUND.SDR | 2883 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| PAS16.SDR | 10712 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| PINBALL.CFG | 6 | VERIFIED: options configuration name referenced by INTRO.ASM |
| PINBALL.EXE | 1742 | INFERRED: DOS launcher corresponding to START.ASM (loads Intro.prg and TableN.Prg) |
| SB16.SDR | 11349 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| SB20.SDR | 11814 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| SBLASTER.SDR | 11421 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| SBPRO.SDR | 11886 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| SETSOUND.EXE | 5652 | INFERRED: sound setup utility from filename; implementation absent |
| SM2.SDR | 11348 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| SOUND.CFG | 25 | VERIFIED: sound configuration name referenced by INTRO.ASM/FANTASIE.ASM |
| TABLE1.HI | 64 | VERIFIED: pristine factory high scores from INTRO.ASM HI_SCORE_LIST; optional seed only, local mutable files are not runtime/build inputs |
| TABLE1.MOD | 210760 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| TABLE1.PRG | 536822 | VERIFIED: MZ executable container with four embedded PBM playfield strips; PLAND.ASM names Party Land |
| TABLE2.HI | 64 | VERIFIED: pristine factory high scores from INTRO.ASM HI_SCORE_LIST; optional seed only, local mutable files are not runtime/build inputs |
| TABLE2.MOD | 211912 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| TABLE2.PRG | 517974 | VERIFIED: MZ executable container; INFERRED: intro/table program by START.ASM filenames |
| TABLE3.HI | 64 | VERIFIED: pristine factory high scores from INTRO.ASM HI_SCORE_LIST; optional seed only, local mutable files are not runtime/build inputs |
| TABLE3.MOD | 219668 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| TABLE3.PRG | 504758 | VERIFIED: MZ executable container; INFERRED: intro/table program by START.ASM filenames |
| TABLE4.HI | 64 | VERIFIED: pristine factory high scores from INTRO.ASM HI_SCORE_LIST; optional seed only, local mutable files are not runtime/build inputs |
| TABLE4.MOD | 216418 | VERIFIED: ProTracker-compatible M.K. marker at byte 1080; INFERRED: music, corresponding MODUL names in ASM (MOD2 role UNKNOWN) |
| TABLE4.PRG | 522198 | VERIFIED: MZ executable container; INFERRED: intro/table program by START.ASM filenames |
| THING.SDR | 10369 | INFERRED: sound driver; INTRO.ASM INIT_SOUND loads configured SDR; hardware name from filename |
| TIMER.BIN | 253 | UNKNOWN: small binary; no verified consumer in PF1 dependency cone |
| pinfant.txt | 16081 | INFERRED: user-facing game documentation from text contents |
