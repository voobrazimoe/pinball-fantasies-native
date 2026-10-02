#!/bin/bash
# Запуск Amiga-версий Pinball Fantasies в FS-UAE.
#
#   tools/amiga-emu.sh ocs    — Amiga OCS 1992 (4 дискеты IPF)
#   tools/amiga-emu.sh aga    — Amiga AGA 1993 (4 дискеты IPF)
#   tools/amiga-emu.sh cd32   — Amiga CD32 (образ CD)
#   tools/amiga-emu.sh ocs --warp --fullscreen   — доп. опции FS-UAE
#
# Эмулятор и плагин IPF лежат в .tools/amiga/pf (установлены без root).
# Переопределить пути: PF_DIR, OCS_DIR, AGA_DIR, CD_DIR, CD_CUE.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PF_DIR="${PF_DIR:-$ROOT/.tools/amiga/pf}"
EMU="$PF_DIR/bin/fs-uae"

OCS_DIR="${OCS_DIR:-$ROOT/pinbllf_amiga/Pinball Fantasies (1992)(21st Century)[0025]}"
AGA_DIR="${AGA_DIR:-$ROOT/pinbllf_amiga/Pinball Fantasies (1993)(21st Century)(AGA)[0047]}"
CD_DIR="${CD_DIR:-$HOME/Загрузки/Pinball Fantasies (1993)(21st Century Entertainment)[!]}"
CD_CUE="${CD_CUE:-$CD_DIR/Pinball Fantasies (1993)(21st Century Entertainment)[!].cue}"

die() { echo "ошибка: $*" >&2; exit 1; }

[ -x "$EMU" ] || die "нет $EMU — эмулятор не установлен"

mode="${1:-}"; shift || true
[ -n "$mode" ] || die "укажи режим: ocs | aga | cd32"

common=(
  --window_width=960 --window_height=720
  --floppy_drive_volume=60
  --joystick_port_0_mode=joystick
  --automatic_input_grab=1
)

case "$mode" in
  ocs)
    [ -d "$OCS_DIR" ] || die "нет каталога $OCS_DIR"
    set -- \
      --amiga_model=A500 \
      --chip_memory=1024 \
      --floppy_drive_0="$OCS_DIR/PinballFantasies_Disk1.ipf" \
      --floppy_drive_1="$OCS_DIR/PinballFantasies_CourseDisk1.ipf" \
      --floppy_drive_2="$OCS_DIR/PinballFantasies_CourseDisk2.ipf" \
      --floppy_drive_3="$OCS_DIR/PinballFantasies_CourseDisk3.ipf" \
      "$@"
    ;;
  aga)
    [ -d "$AGA_DIR" ] || die "нет каталога $AGA_DIR"
    set -- \
      --amiga_model=A1200 \
      --chip_memory=2048 \
      --floppy_drive_0="$AGA_DIR/PinballFantasiesAGA_PF1.ipf" \
      --floppy_drive_1="$AGA_DIR/PinballFantasiesAGA_PF2.ipf" \
      --floppy_drive_2="$AGA_DIR/PinballFantasiesAGA_PF3.ipf" \
      --floppy_drive_3="$AGA_DIR/PinballFantasiesAGA_PF4.ipf" \
      "$@"
    ;;
  cd32)
    [ -f "$CD_CUE" ] || die "нет образа CD: $CD_CUE (задай CD_CUE=...)"
    set -- \
      --amiga_model=CD32 \
      --chip_memory=2048 \
      --cdrom_drive_0="$CD_CUE" \
      "$@"
    ;;
  *) die "неизвестный режим: $mode (ожидается ocs | aga | cd32)" ;;
esac

echo "FS-UAE: режим $mode"
echo "  эмулятор : $EMU"
echo "  данные   : $PF_DIR/bin/Data"
exec "$EMU" "${common[@]}" "$@"
