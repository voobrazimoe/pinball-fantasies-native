# Provenance and licensing boundary

Pinball Fantasies Native is a source-guided native reimplementation. The public repository contains the project-authored Go implementation, build tooling, tests and documentation. It does not contain the original commercial game payload.

## Historical reference

Development consulted the historical source repository:

- upstream: `https://github.com/historicalsource/pinballfantasies`
- pinned reference commit: `aa2dd368d73886bbd666507bd7341001060700d9`

That historical source is not bundled in this repository or in release binaries. No open-source license was identified in the pinned historical tree, so this project does not purport to relicense it.

## Original game data

Users provide their own legally obtained DOS copy of Pinball Fantasies. PRG, MOD, legacy HI, configuration, artwork, music and other original commercial data remain outside the project license.

The public release artifacts are asset-free. Original game files are read as data containers and their x86 code is never executed.

## Project license

The native implementation and project-authored material are released under the MIT License, copyright (c) 2026 voobrazimoe.

The MIT license applies only to material for which this project can grant those terms. It does not grant rights to Pinball Fantasies itself, its original commercial data, artwork, music, trademarks or historical/reference source material.

## Public history

The public repository was created from an audited source snapshot with fresh Git history rather than by exposing the private development history. Local personal builds that embed commercial inputs are kept outside the public repository and are not project release artifacts.