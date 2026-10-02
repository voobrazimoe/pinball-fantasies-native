#!/bin/bash
# Устанавливает FS-UAE + плагин IPF локально в .tools/amiga (без root, без apt install).
#
#   tools/setup-amiga-emu.sh
#
# Ставит:
#   .tools/amiga/pf/bin/fs-uae                     — эмулятор (из пакета Ubuntu)
#   .tools/amiga/pf/bin/Plugins/CAPSImg/...        — плагин CAPS/IPF
#   .tools/amiga/pf/bin/Data/{ROM/Kickstart,Config,Floppies,CD}
#   .tools/amiga/pf/share/fs-uae/...               — ресурсы FS-UAE
#
# Запуск: tools/amiga-emu.sh ocs|aga|cd32
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DL="$ROOT/.tools/amiga/dl"
PF="$ROOT/.tools/amiga/pf"
mkdir -p "$DL"

FSUAE_DEB_URL="http://archive.ubuntu.com/ubuntu/pool/universe/f/fs-uae/fs-uae_3.2.35-2_amd64.deb"
FSUAE_DEB_SHA="8c77027bc402ea58ef4536c49d08dabe706eed51232005873f897fd1b0ed0a4f"
CAPS_URL="https://fs-uae.net/files/CAPSImg/Stable/5.1.3/CAPSImg_5.1.3_Linux_x86-64.tar.xz"

say() { printf '  %s\n' "$*"; }
fetch() { # url dest sha256(optional)
  local url="$1" dest="$2" want="${3:-}"
  if [ ! -f "$dest" ]; then
    say "скачиваю $(basename "$dest")"
    if ! command -v curl >/dev/null && ! command -v wget >/dev/null; then
      python3 - "$url" "$dest" <<'PY'
import sys, urllib.request
urllib.request.urlretrieve(sys.argv[1], sys.argv[2])
PY
    elif command -v wget >/dev/null; then wget -q -O "$dest" "$url"
    else curl -sSL -o "$dest" "$url"; fi
  fi
  if [ -n "$want" ]; then
    got="$(sha256sum "$dest" | cut -d' ' -f1)"
    [ "$got" = "$want" ] || { echo "ошибка: неверная SHA-256 у $dest" >&2; exit 1; }
  fi
}

if [ ! -x "$PF/bin/fs-uae" ]; then
  say "распаковываю FS-UAE"
  fetch "$FSUAE_DEB_URL" "$DL/fs-uae.deb" "$FSUAE_DEB_SHA"
  rm -rf "$ROOT/.tools/amiga/_deb"
  dpkg-deb -x "$DL/fs-uae.deb" "$ROOT/.tools/amiga/_deb"
  mkdir -p "$PF/bin" "$PF/share"
  cp "$ROOT/.tools/amiga/_deb/usr/bin/fs-uae" "$PF/bin/"
  rm -rf "$PF/share/fs-uae"
  cp -r "$ROOT/.tools/amiga/_deb/usr/share/fs-uae" "$PF/share/fs-uae"
fi

if [ ! -f "$PF/bin/Plugins/CAPSImg/Linux/x86-64/capsimg.so" ]; then
  say "ставлю плагин CAPS/IPF"
  fetch "$CAPS_URL" "$DL/CAPSImg.tar.xz"
  rm -rf "$DL/CAPSImg"
  tar -xJf "$DL/CAPSImg.tar.xz" -C "$DL"
  mkdir -p "$PF/bin/Plugins" "$PF/bin/Data/Plugins"
  cp -r "$DL/CAPSImg" "$PF/bin/Plugins/"
  cp -r "$DL/CAPSImg" "$PF/bin/Data/Plugins/"
  # FS-UAE ищет плагин и «плоско», рядом с бинарником:
  cp "$DL/CAPSImg/Linux/x86-64/capsimg.so" "$PF/bin/capsimg.so"
fi

mkdir -p "$PF/bin/Data/ROM/Kickstart" "$PF/bin/Data/Config" \
         "$PF/bin/Data/Floppies" "$PF/bin/Data/CD"
printf 'portable = 1\n' > "$PF/bin/Portable.ini"

say "готово:"
"$PF/bin/fs-uae" --version | sed 's/^/    FS-UAE /'
say "плагин: $PF/bin/Plugins/CAPSImg/Linux/x86-64/capsimg.so"
say "ROM'ы можно положить в $PF/bin/Data/ROM/Kickstart (не обязательны, см. README)"
