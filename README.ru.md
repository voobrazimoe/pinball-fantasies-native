# Pinball Fantasies Native

[English](README.md) | **Русский**

Первый публичный beta-релиз нативной реконструкции DOS-версии **Pinball Fantasies** для современных Windows, Linux и macOS.

> **Важно:** публичные сборки не содержат коммерческих файлов Pinball Fantasies. Для запуска нужна собственная легально полученная DOS-версия игры.

## Быстрый старт

Скачайте опубликованный [prerelease v0.1.2](https://github.com/voobrazimoe/pinball-fantasies-native/releases/tag/v0.1.2) для своей платформы:

- [Windows x86_64: pinballfantasies.exe](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.2/pinballfantasies.exe)
- [Linux x86_64: PinballFantasies-x86_64.AppImage](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.2/PinballFantasies-x86_64.AppImage)
- [macOS Apple Silicon ARM64: PinballFantasies-arm64.zip](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.2/PinballFantasies-arm64.zip)
- [macOS Intel x86_64: PinballFantasies-x86_64.zip](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.2/PinballFantasies-x86_64.zip)
- [Контрольные суммы SHA256](https://github.com/voobrazimoe/pinball-fantasies-native/releases/download/v0.1.2/SHA256SUMS.txt)

Для macOS 13 и новее скачайте ZIP для своего Mac, распакуйте его и запустите приложение. Выберите каталог оригинальной DOS-игры через нативный диалог импорта; файлы проверяются и копируются в Application Support. Публичные Mac-сборки имеют ad-hoc подпись и не прошли нотариализацию. Подробнее — в [инструкции по сборке и использованию macOS-версии](docs/macos.md).

На Windows/Linux оригинальные DOS-данные можно положить рядом с исполняемым файлом либо указать каталог явно:

```text
pinballfantasies.exe -data-dir "D:\Games\Pinball Fantasies"
```

```sh
chmod +x PinballFantasies-x86_64.AppImage
./PinballFantasies-x86_64.AppImage -data-dir "/path/to/Pinball Fantasies"
```

Для запуска используются 11 файлов PRG/MOD из поддерживаемой DOS-установки:

```text
INTRO.PRG  INTRO.MOD  MOD2.MOD
TABLE1.PRG TABLE1.MOD
TABLE2.PRG TABLE2.MOD
TABLE3.PRG TABLE3.MOD
TABLE4.PRG TABLE4.MOD
```

`PINBALL.CFG` — необязательный импорт старых настроек. Отсутствующий или некорректный CFG заменяется встроенными настройками; PFNC хранится отдельно в пользовательском состоянии. Совместимость проверяется по реально используемым данным; точные хеши целых файлов нужны только для эталонных проверок.

Оригиналы используются только для чтения. На Windows/Linux нативные настройки, рекорды и логи записываются в `userdata/` рядом с EXE/AppImage; если туда нельзя писать, используется пользовательский каталог конфигурации. macOS импортирует оригиналы в `~/Library/Application Support/PinballFantasies/Data/`, а настройки и рекорды записывает в соседний каталог `State/`, вне приложения.

На Windows/Linux необязательные `TABLE*.HI` могут использоваться как исходные рекорды только для чтения; без них берутся встроенные заводские значения. Импортёр macOS копирует только обязательные PRG/MOD и необязательный CFG; при отсутствии сохранённого нативного состояния используются заводские рекорды.

## Что реализовано

Работают все четыре стола:

- Party Land
- Speed Devils
- Billion Dollar Gameshow
- Stones ’N Bones

Порт включает нативные правила и машины состояний столов, целочисленную физику шара, флипперы, пружину, толчки и tilt, матричное табло, лампы и анимированные элементы поля, рекорды, настройки, tracker/module-музыку и эффекты, изменяемые по размеру окна, платформенный fullscreen и отдельные нативные backend'ы для Windows/Linux/macOS. Windows использует Win32/GDI/waveOut, Linux — SDL2, macOS — AppKit/Core Animation/AudioUnit вокруг общего Go-движка и C ABI.

Это не DOSBox, не интерпретатор x86 и не оболочка вокруг оригинального EXE. Оригинальные файлы читаются как данные; их x86-код не выполняется.

## Управление и мультиплеер

До загрузки стола:

| Клавиша | Действие |
|---|---|
| `F1`–`F4` | Party Land / Speed Devils / Billion Dollar Gameshow / Stones ’N Bones |
| `F5` | Настройки |

После загрузки стола:

| Клавиша | Действие |
|---|---|
| `F1`–`F8` | Начать игру на 1–8 игроков |
| `Enter` | Добавить игрока до запуска первого шара |
| `Down Arrow` | Удерживать для натяжения пружины, отпустить для запуска |
| Левые/правые `Shift`, `Ctrl`, `Alt` (`Option` на macOS) | Соответствующие левый/правый флипперы |
| `Space` | Толчок стола; повторные толчки могут вызвать tilt |
| `P` | Пауза |
| `M` | Включить/выключить музыку |
| `Alt+Enter` (Windows/Linux) | Borderless fullscreen |
| `Option+Return`, `Command+F` или View → Toggle Full Screen (macOS) | Нативный fullscreen |
| `Esc` | Назад / выход в зависимости от экрана |

На macOS `Z` / `/` и стрелки влево / вправо — дополнительные клавиши левого/правого флиппера. Если Mac сообщает стороны Shift наоборот, включите **View → Swap Left/Right Shift**. Настройка сохраняется после перезапуска и влияет только на Shift.

До первого запуска `F1`–`F8` меняют количество игроков, а `Enter` добавляет игрока до восьми. После первого запуска количество фиксируется на всю игру. Игроки чередуются по раундам шаров; extra ball остаётся у того, кто его заработал. У каждого игрока собственные очки и сохраняемое состояние стола. При переходе к следующему игроку матрица сразу показывает его счёт и номер шара, ещё до запуска.

После игры на восемь игроков оригинальное поведение attract mode игнорирует `Enter`; следующую игру нужно начинать через `F1`–`F8`.

## Настройки и режим полного стола

В F5 доступны 3/5 шаров, HIGH/LOW angle, HARD/MEDIUM/SOFT/OFF scrolling, музыка и NORMAL/HIGH resolution.

`SCROLLING: OFF` — нативное расширение. Оно показывает целиком игровое поле 320×576 вместе с матрицей 320×33, итоговый логический кадр — 320×609. Физика и правила остаются теми же, что и в режимах с прокруткой.

## Звук

Нативный mixer воспроизводит исходную четырёхканальную tracker/module-музыку и эффекты как 48 кГц signed 16-bit stereo. Каналы 0 и 3 направлены влево, 1 и 2 — вправо, по аппаратной схеме Amiga Paula.

## Сборка и проверки

См. [инструкцию по сборке](docs/build.md). Публичный CI выполняет общие asset-free тесты и нативные проверки Windows/Linux, а также тесты и сборку Mac-хоста ARM64 с Apple SDK и кросс-сборку Intel x86_64. Коммерческие игровые данные и исторические исходники не скачиваются. Hosted asset-free проверки отделены от локальной проверки с оригиналами. macOS replay-проверки с оригиналами и нативные сценарии прошли; приёмка игры и управления на физическом Mac записана в [отчёте](docs/macos-validation.md).

Полезная документация:

- [Сборка и использование macOS-версии](docs/macos.md)
- [Проверки macOS-версии](docs/macos-validation.md)
- [Архитектура](docs/architecture.md)
- [Граница runtime-данных](docs/runtime-data.md)
- [Происхождение и границы лицензии](docs/provenance.md)

Сохранённый детерминированный oracle Party Land: 1200 ticks, счёт `000002300000`, ball 2, frame SHA256 `f9b5160b3173c35f2798dd5e270e7ded40c33642605e21c0935cd083ff231a5e`.

## Ограничения beta

На некоторых Linux-системах с несколькими мониторами вход/выход из fullscreen может на мгновение вызвать flicker или перемещение окна между дисплеями.

Для Windows используется нативный backend Win32/GDI/waveOut. Hosted Windows build/tests и документированные Wine/live-регрессии проходят, но этот beta-релиз не заявляет исчерпывающую проверку всех сочетаний multiplayer/fullscreen/audio на физическом Windows-компьютере.

## Локальный self-contained builder

Для личного использования на Windows/Linux `./tools/build_personal_release.sh "/path/to/original/game"` создаёт self-contained сборки в `release/personal/` из тех же 11 игровых файлов и необязательного файла настроек. Эти сборки содержат ваши коммерческие игровые данные и **не являются распространяемыми релизами проекта**. Их нельзя коммитить, пушить или загружать в Releases.

Для macOS команда `python3 tools/build_personal_macos.py "/path/to/original/game"` создаёт отдельные приложения ARM64 и x86_64 с автоматическим импортом при первом запуске. На них распространяется та же граница локального использования коммерческих данных.

## Лицензия и игровые данные

Нативный код и материалы, созданные для этого проекта, распространяются по [MIT License](LICENSE), copyright (c) 2026 voobrazimoe.

Сам Pinball Fantasies не перелицензируется. Оригинальные коммерческие игровые данные, графика, музыка и исторические/reference-исходники не покрываются MIT. Проект не заявляет права собственности на торговые марки, графику, музыку или коммерческие данные оригинальной игры.

Подробнее: [происхождение и границы проекта](docs/provenance.md).
