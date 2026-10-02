#!/usr/bin/env python3
"""Opt-in live native host test with XTest keys on an isolated Xvfb display.

Build internal/platform tests into bin/pf121-linux.test and
bin/pf121-windows.test.exe first. The test executable contains explicitly
identified score/drain fixtures; production binaries contain none. Wine is a
host compatibility check, never a DOS runtime or gameplay dependency.
"""
import argparse
import ctypes as C
import os
from pathlib import Path
import subprocess
import threading
import time

ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / '.build-personal/pf121'

class XWindowAttributes(C.Structure):
    _fields_ = [(n, C.c_int) for n in ('x', 'y', 'width', 'height', 'border_width', 'depth')] + [
        ('visual', C.c_void_p), ('root', C.c_ulong), ('window_class', C.c_int),
        ('bit_gravity', C.c_int), ('win_gravity', C.c_int), ('backing_store', C.c_int),
        ('backing_planes', C.c_ulong), ('backing_pixel', C.c_ulong), ('save_under', C.c_int),
        ('colormap', C.c_ulong), ('map_installed', C.c_int), ('map_state', C.c_int),
        ('all_event_masks', C.c_long), ('your_event_mask', C.c_long),
        ('do_not_propagate_mask', C.c_long), ('override_redirect', C.c_int), ('screen', C.c_void_p)]


def run(backend, table):
    x = C.CDLL('libX11.so.6')
    xt = C.CDLL('libXtst.so.6')
    x.XOpenDisplay.restype = C.c_void_p
    x.XOpenDisplay.argtypes = [C.c_char_p]
    x.XDefaultRootWindow.restype = C.c_ulong
    x.XDefaultRootWindow.argtypes = [C.c_void_p]
    x.XQueryTree.argtypes = [C.c_void_p, C.c_ulong, C.POINTER(C.c_ulong), C.POINTER(C.c_ulong), C.POINTER(C.POINTER(C.c_ulong)), C.POINTER(C.c_uint)]
    x.XGetWindowAttributes.argtypes = [C.c_void_p, C.c_ulong, C.POINTER(XWindowAttributes)]
    x.XFetchName.argtypes = [C.c_void_p, C.c_ulong, C.POINTER(C.c_char_p)]
    x.XFree.argtypes = [C.c_void_p]
    x.XSetInputFocus.argtypes = [C.c_void_p, C.c_ulong, C.c_int, C.c_ulong]
    x.XFlush.argtypes = [C.c_void_p]
    x.XStringToKeysym.restype = C.c_ulong
    x.XStringToKeysym.argtypes = [C.c_char_p]
    x.XKeysymToKeycode.restype = C.c_uint
    x.XKeysymToKeycode.argtypes = [C.c_void_p, C.c_ulong]
    xt.XTestFakeKeyEvent.argtypes = [C.c_void_p, C.c_uint, C.c_int, C.c_ulong]
    display = ':121'
    server = subprocess.Popen([str(ROOT / '.tools/xvfb/usr/bin/Xvfb'), display,
                               '-screen', '0', '1280x1024x24', '-nolisten', 'tcp', '-ac'],
                              stdout=subprocess.DEVNULL, stderr=open(OUT / 'xvfb.log', 'a'))
    process = None
    try:
        time.sleep(.6)
        d = x.XOpenDisplay(display.encode())
        assert d, 'isolated X11 display unavailable'
        env = dict(os.environ, DISPLAY=display, PF121_LIVE='1', PF121_TABLE=str(table),
                   SDL_VIDEODRIVER='x11', SDL_AUDIODRIVER='pulseaudio',
                   WINEPREFIX=os.environ.get('PF12_WINEPREFIX', str(OUT / 'wine-prefix')), WINEDEBUG='-all')
        Path(env['WINEPREFIX']).mkdir(parents=True, exist_ok=True)
        env['PF121_DATA'] = str(ROOT) if backend == 'linux' else 'Z:' + str(ROOT).replace('/', '\\')
        command = [str(ROOT / 'bin/pf121-linux.test')] if backend == 'linux' else ['wine64', str(ROOT / 'bin/pf121-windows.test.exe')]
        process = subprocess.Popen(command + ['-test.v', '-test.run', '^TestHotseatLiveHost$'],
                                   cwd=ROOT, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        lines = []

        def reader():
            for line in process.stdout:
                lines.append(line.strip())
                print(line.strip(), flush=True)

        thread = threading.Thread(target=reader)
        thread.start()
        root = x.XDefaultRootWindow(d)

        def find(w):
            name = C.c_char_p()
            x.XFetchName(d, w, C.byref(name))
            found = name.value == b'Pinball Fantasies'
            if name:
                x.XFree(name)
            if found:
                attrs = XWindowAttributes()
                if x.XGetWindowAttributes(d, w, C.byref(attrs)) and attrs.map_state == 2:
                    return w
            r, parent = C.c_ulong(), C.c_ulong()
            children, count = C.POINTER(C.c_ulong)(), C.c_uint()
            if x.XQueryTree(d, w, C.byref(r), C.byref(parent), C.byref(children), C.byref(count)):
                ids = [children[i] for i in range(count.value)]
                if children:
                    x.XFree(children)
                for child in ids:
                    got = find(child)
                    if got:
                        return got

        def waitfor(text, timeout=45):
            end = time.monotonic() + timeout
            while time.monotonic() < end:
                if any(text in line for line in lines):
                    return
                assert process.poll() is None, (text, 'native host exited')
                time.sleep(.05)
            raise AssertionError((text, 'timeout'))

        window = None
        for _ in range(200):
            window = find(root)
            if window:
                break
            time.sleep(.05)
        assert window, 'native game window unavailable'
        # find() waits for IsViewable: title assignment precedes mapping in Wine.

        def focus(w):
            x.XSetInputFocus(d, w, 1, 0)
            x.XFlush(d)
            time.sleep(.15)

        def edge(key, down):
            code = x.XKeysymToKeycode(d, x.XStringToKeysym(key.encode()))
            assert code
            xt.XTestFakeKeyEvent(d, code, down, 0)
            x.XFlush(d)

        def press(key, hold=.10):
            edge(key, 1)
            time.sleep(hold)
            edge(key, 0)
            time.sleep(.12)

        focus(window)
        press('space')
        press('F' + str(table))
        press('F2')
        waitfor('PF121 TURN P1 B1')
        press('p')
        time.sleep(.3)
        press('p')
        time.sleep(.3)
        focus(root)
        time.sleep(.3)
        focus(window)
        press('p')
        time.sleep(.3)
        for _ in range(2):
            edge('Alt_L', 1)
            press('Return')
            edge('Alt_L', 0)
            time.sleep(.4)
        for turn in ['P1 B1', 'P2 B1', 'P1 B2', 'P2 B2']:
            waitfor('PF121 TURN ' + turn)
            # NEW_BALL_TASK plus SETBALL hold must finish before the next launch.
            time.sleep(1.8)
            press('Down', .55)
            waitfor('PF121 LAUNCH ' + turn)
            waitfor('PF121 FIXTURE DRAIN ' + turn)
        waitfor('PF121 TURN P1 B3')
        process.wait(timeout=10)
        thread.join()
        assert process.returncode == 0
    finally:
        if process is not None and process.poll() is None:
            process.terminate()
            process.wait(timeout=10)
        server.terminate()
        server.wait(timeout=5)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--backend', choices=['linux', 'wine', 'both'], default='both')
    parser.add_argument('--table', type=int, choices=range(1, 5))
    args = parser.parse_args()
    OUT.mkdir(parents=True, exist_ok=True)
    if args.backend != 'both' and args.table:
        run(args.backend, args.table)
    else:
        for backend in (['linux', 'wine'] if args.backend == 'both' else [args.backend]):
            for table in ([args.table] if args.table else range(1, 5)):
                with (OUT / f'live-{backend}-{table}.log').open('w') as log:
                    result = subprocess.run(['python3', __file__, '--backend', backend, '--table', str(table)],
                                            stdout=log, stderr=subprocess.STDOUT)
                print(backend, table, 'exit', result.returncode, flush=True)
                if result.returncode:
                    raise SystemExit(result.returncode)
