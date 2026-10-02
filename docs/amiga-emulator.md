# Amiga-эмулятор для Pinball Fantasies (FS-UAE)

Поставлено локально в `.tools/amiga/`, **без root и без `apt install`** —
пакет `.deb` распакован в каталог проекта (тот же приём, что в
`tools/setup-local.py` для SDL2/Go).

## Быстрый старт

```sh
tools/amiga-emu.sh ocs     # Amiga OCS 1992, 4 дискеты IPF
tools/amiga-emu.sh aga     # Amiga AGA 1993, 4 дискеты IPF
tools/amiga-emu.sh cd32    # Amiga CD32, образ CD
```

Дополнительные опции FS-UAE можно дописать в конец, например:

```sh
tools/amiga-emu.sh aga --fullscreen --window_width=1280 --window_height=960
tools/amiga-emu.sh ocs --warp          # ускорить эмуляцию
tools/amiga-emu.sh cd32 --floppy_drive_volume=0
```

Переустановить/починить эмулятор:

```sh
tools/setup-amiga-emu.sh
```

## Что где лежит

| Путь | Что это |
|---|---|
| `.tools/amiga/pf/bin/fs-uae` | эмулятор FS-UAE 3.2.35 |
| `.tools/amiga/pf/bin/Plugins/CAPSImg/` | плагин CAPS/IPF 5.1.3 (чтение `.ipf`) |
| `.tools/amiga/pf/bin/capsimg.so` | тот же плагин «плоско» — подстраховка |
| `.tools/amiga/pf/bin/Data/ROM/Kickstart/` | сюда кладутся Kickstart ROM'ы (не обязательны) |
| `.tools/amiga/pf/bin/Data/Config/` | конфиги FS-UAE |
| `.tools/amiga/pf/bin/Cache/Logs/fs-uae.log.txt` | лог последнего запуска |

FS-UAE работает в **portable-режиме**: все данные внутри
`.tools/amiga/pf/`, ничего в `$HOME` не пишется.

## Kickstart ROM'ы — не нужны (проверено)

FS-UAE тащит с собой свободную замену Kickstart — **AROS ROM**
(`share/fs-uae/aros-amiga-m68k-rom.bin`). Все три версии запускаются на
ней, реальные ROM'ы не требуются:

```
ROM Type: AROS ROM Operating System and Libraries. ROM Model: 20.05.2015
ROM Type: AROS EXT Extension Libraries.
```

Если хочется максимальной совместимости, положите в
`Data/ROM/Kickstart`:

* OCS 1992 → Kickstart 1.3 (`kick34005.A500`), для A500;
* AGA 1993 → Kickstart 3.1 (`kick40068.A1200`), для A1200;
* CD32 → Kickstart 3.1 CD32 (`kick40060.CD32`) + расширение
  (`kick40060.CD32.ext`).

ROM'ы проприетарные; эмулятор ищет их по именам и подхватывает
автоматически. Без них CD32-режим у меня всё равно загрузился до логотипа
Digital Illusions (см. ниже) — но если что-то пойдёт не так, это первое,
что стоит добавить.

## Проверено (скриншоты в `analysis/amiga/gfx/`)

| Режим | Результат |
|---|---|
| `ocs` | грузится экран `EXTRA MEMORY LOCATED / PINBALL FANTASIES NOW LOADING` (`ocs_boot.png`, `ocs_b.png`) |
| `aga` | грузится и показывает **вопрос защиты по мануалу** (`aga_b.png`) |
| `cd32` | грузится логотип **Digital Illusions** (варяжский корабль), `cd32_b.png` |

Запускалось под Xvfb + программный OpenGL (`llvmpipe`), поэтому кадры
есть; на живом рабочем столе будет обычное окно с аппаратным ускорением.

## Защита по мануалу (AGA/OCS-версии)

AGA-версия при старте спрашивает:

> **Please type in the 1st UPPER-CASE word of Speed Devils' "extra ball"
> feature. (Hyphenated words (i.e. PIT-STOP) count as one word.)**

Список фич Speed Devils есть в `pinfant.txt` (строки 152–194) и в
мануале (`pinbllf_amiga/.../Manual.pdf`, отсканирован без текстового слоя):

```
GEAR / POSITION / EXTRA BALL / JACKPOT / SUPER-JACKPOT / MILES / MILLION /
SPEED / AUTO-FEATURES / GOAL / MULTI-BONUS / OFF-ROAD / TURBO-MODE / JUMP
```

Фича «extra ball» называется **`EXTRA BALL`**, первое слово в верхнем
регистре — **`EXTRA`**. Я это в эмуляторе не набирал (ввод с клавиатуры
из песочницы не подать), так что проверьте сами — но по формулировке
вопроса ответ именно такой.

Ответ вводится прямо в окне эмулятора: игра открывает requester
Intuition, набираете слово и жмёте Return.

## Управление

| Клавиша | Действие |
|---|---|
| `F12` | меню FS-UAE (диски, сохранения, выход) |
| `F1`/`F2` | выбрать стол (Party Land / Speed Devils) |
| `F5`/`F6` | переключить диск в дисководе |
| Стрелки/`Shift` | плунжер и флипперы (как в оригинале) |
| `End`/`Pause` | отпустить мышь из окна (если захвачена) |

Мышь захватывается автоматически (`automatic_input_grab=1`); чтобы
выйти из захвата — `F12` или `End`.

## Если «диск не найден»

1. Проверьте, что файлы на месте:
   `ls "pinbllf_amiga/Pinball Fantasies (1992)(21st Century)[0025]"/`
2. Посмотрите лог: `.tools/amiga/pf/bin/Cache/Logs/fs-uae.log.txt`.
   Признак исправного IPF-плагина:
   ```
   [PLUGINS] DLOPEN: Loaded plugin .../Plugins/CAPSImg/Linux/x86-64/capsimg.so
   CAPS: library version 5.1 (flags=00007FFF)
   caps: type:1 imagetype:2 date:29.1.2003 ... rel:25 rev:1
   ```
   Если вместо этого `IPF support plugin (capsimg) is not installed` —
   запустите `tools/setup-amiga-emu.sh`.
3. Для CD32 нужен именно `.cue` (в нём прописаны ISO + два WAV-трека);
   путь можно переопределить: `CD_CUE=... tools/amiga-emu.sh cd32`.
