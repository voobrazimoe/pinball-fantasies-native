#!/usr/bin/env python3
"""Exercise the production SDL event bridge using real SDL virtual controllers.

Needs SDL2 >= 2.0.14 development files, pkg-config, and a C compiler. No game
assets or physical controller are used. May also run on macOS with native SDL2.
"""
from pathlib import Path
import os
import shlex
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]

HARNESS = r'''
#include <assert.h>
static SDL_GameController *pad;
static int key, alt, repeat, value;
static int event(void) { return pf6_event(&pad,&key,&alt,&repeat,&value); }
static void drain(void) { for(int i=0;i<1000;i++) if(!event()) return; assert(0); }
static void expect(int kind,int code,int data) {
 for(int i=0;i<1000;i++) {
  int e=event();
  if(e==kind) {assert(key==code);assert(value==data);return;}
  // SDL's poll sentinel may end one pump even with new events pending.
 }
 fprintf(stderr,"Missing SDL event kind=%d code=%d data=%d: %s\n",kind,code,data,SDL_GetError());
 assert(0);
}
static int attach(void) {
 int index=SDL_JoystickAttachVirtual(SDL_JOYSTICK_TYPE_GAMECONTROLLER,6,15,0);
 assert(index>=0);
 char guid[33],mapping[1024];
 SDL_JoystickGetGUIDString(SDL_JoystickGetDeviceGUID(index),guid,sizeof(guid));
 SDL_snprintf(mapping,sizeof(mapping),"%s,PF test pad,a:b0,b:b1,x:b2,y:b3,back:b4,guide:b5,start:b6,leftstick:b7,rightstick:b8,leftshoulder:b9,rightshoulder:b10,dpup:b11,dpdown:b12,dpleft:b13,dpright:b14,leftx:a0,lefty:a1,rightx:a2,righty:a3,lefttrigger:a4,righttrigger:a5,",guid);
 assert(SDL_GameControllerAddMapping(mapping)>=0);
 return index;
}
int main(void) {
 assert(SDL_Init(SDL_INIT_VIDEO)==0);
 assert(pf6_controller_init()==0);
 int first=attach();
 SDL_Joystick *joy=SDL_JoystickOpen(first);assert(joy);
 // Open an already-held device: snapshot history suppresses its initial make.
 assert(SDL_JoystickSetVirtualButton(joy,SDL_CONTROLLER_BUTTON_X,SDL_PRESSED)==0);
 SDL_JoystickUpdate();
 assert(event()==12);assert(key & (1<<SDL_CONTROLLER_BUTTON_X));assert(pad);
 drain();
 assert(SDL_JoystickSetVirtualButton(joy,SDL_CONTROLLER_BUTTON_X,SDL_RELEASED)==0);
 SDL_JoystickUpdate();expect(9,SDL_CONTROLLER_BUTTON_X,0);
 assert(SDL_JoystickSetVirtualButton(joy,SDL_CONTROLLER_BUTTON_X,SDL_PRESSED)==0);
 SDL_JoystickUpdate();expect(9,SDL_CONTROLLER_BUTTON_X,1);
 // Raw virtual trigger -32768..32767 maps to standard trigger 0..32767.
 assert(SDL_JoystickSetVirtualAxis(joy,SDL_CONTROLLER_AXIS_TRIGGERLEFT,32767)==0);
 SDL_JoystickUpdate();expect(10,SDL_CONTROLLER_AXIS_TRIGGERLEFT,32767);
 int second=attach();
 SDL_Joystick *other=SDL_JoystickOpen(second);assert(other);drain();
 assert(SDL_JoystickSetVirtualButton(other,SDL_CONTROLLER_BUTTON_Y,SDL_PRESSED)==0);
 SDL_JoystickUpdate();
 for(int i=0;i<1000;i++) {int e=event();assert(e!=9);}
 assert(SDL_JoystickDetachVirtual(first)==0);
 int removed=0;for(int i=0;i<1000;i++){int e=event();if(e==11){removed=1;break;}}
 assert(removed);assert(!pad);
 assert(event()==12);assert(key & (1<<SDL_CONTROLLER_BUTTON_Y));assert(pad);
 SDL_GameControllerClose(pad);pad=NULL;
 SDL_JoystickClose(joy);SDL_JoystickClose(other);
 assert(SDL_JoystickDetachVirtual(0)==0);
 SDL_Quit();
 puts("PASS: SDL controller startup, button edges, trigger, device filtering, disconnect and replacement");
 return 0;
}
'''


def main():
    source = (ROOT / "internal/platform/frontend.go").read_text()
    preamble = source.split("/*", 1)[1].split("*/", 1)[0]
    preamble = "\n".join(line for line in preamble.splitlines()
                         if not line.startswith("#cgo"))
    flags = shlex.split(subprocess.check_output(
        ["pkg-config", "--cflags", "--libs", "sdl2"], text=True))
    with tempfile.TemporaryDirectory(prefix="pf-sdl-gamepad-") as tmp:
        path = Path(tmp)
        (path / "test.c").write_text(preamble + HARNESS)
        subprocess.run([os.environ.get("CC", "cc"), "-Werror=implicit-function-declaration",
                        str(path / "test.c"), "-o", str(path / "test"), *flags],
                       check=True)
        env = dict(os.environ, SDL_VIDEODRIVER="dummy", SDL_AUDIODRIVER="dummy")
        subprocess.run([str(path / "test")], env=env, check=True, timeout=20)


if __name__ == "__main__":
    main()
