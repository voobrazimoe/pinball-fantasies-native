#import "app.h"
#import "frame_view.h"
#import <mach/mach_time.h>
#import "native_input.h"
#include <stdio.h>
/* OS helpers are separate modules; gameplay stays behind abi.h. */
#import "audio_host.h"
#import "storage.h"
@interface PFApp ()
- (void)emit:(PFHostEvent)event a:(int32_t)a b:(int32_t)b;
- (void)syncFocus;
- (void)sleeping:(BOOL)sleeping;
- (void)deviceChanged;
- (void)step;
- (void)toggleFullscreen:(id)sender;
@end
static void inputEvent(void *context,PFHostEvent event,int32_t a,int32_t b) {
    [(__bridge PFApp *)context emit:event a:a b:b];
}
@implementation PFApp {
    NSWindow *_window;
    PFFrameView *_view;
    NSTimer *_timer;
    uint64_t _engine, _epoch;
    mach_timebase_info_data_t _timebase;
    PFInput _input;
    PFAudio _audio;
    BOOL _focused, _mouseActive, _cursorHidden, _failed, _sleeping;
    id _sleepObserver, _wakeObserver;
}
- (int64_t)now { return pf_clock_ns(mach_absolute_time(),_epoch,_timebase.numer,_timebase.denom); }
- (void)fail:(NSString *)message {
    if (_failed) return; _failed=YES;
    NSLog(@"Native host failure: %@",message);
    NSAlert *alert=[NSAlert new]; alert.messageText=@"Pinball Fantasies could not continue";
    alert.informativeText=message; [alert runModal]; [NSApp terminate:nil];
}
- (void)check:(int32_t)result {
    if (result!=PF_OK) [self fail:[NSString stringWithFormat:@"Engine boundary returned %d",result]];
}
- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    (void)notification;
    mach_timebase_info(&_timebase); _epoch=mach_absolute_time();
    pf_input_init(&_input,inputEvent,(__bridge void *)self); pf_audio_init(&_audio);
    NSString *data=nil,*state=nil; NSError *error=nil;
    if (!pf_storage_prepare(&data,&state,&error)) {
        if (error) [self fail:error.localizedDescription]; else [NSApp terminate:nil];
        return;
    }
    char failure[1024]={0};
    _engine=pf_engine_create((char *)data.fileSystemRepresentation,(char *)state.fileSystemRepresentation,
                             [self now],failure,sizeof(failure));
    if (!_engine) { [self fail:[NSString stringWithUTF8String:failure]]; return; }
    [self check:pf_engine_suspend(_engine)];
    NSRect usable=NSScreen.mainScreen.visibleFrame;
    NSSize size=NSMakeSize(fmin(800,usable.size.width*.85),fmin(766,usable.size.height*.85));
    _window=[[NSWindow alloc] initWithContentRect:NSMakeRect(0,0,size.width,size.height)
        styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable|NSWindowStyleMaskMiniaturizable|NSWindowStyleMaskResizable
        backing:NSBackingStoreBuffered defer:NO];
    _window.title=@"Pinball Fantasies"; _window.delegate=self; _window.releasedWhenClosed=NO;
    _window.collectionBehavior=NSWindowCollectionBehaviorFullScreenPrimary;
    _window.contentMinSize=NSMakeSize(320,240); _window.acceptsMouseMovedEvents=YES;
    _view=[[PFFrameView alloc] initWithFrame:NSMakeRect(0,0,size.width,size.height)];
    _view.host=self; _view.autoresizingMask=NSViewWidthSizable|NSViewHeightSizable;
    _window.contentView=_view; [_window makeFirstResponder:_view]; [_window center];
    [_window makeKeyAndOrderFront:nil]; [NSApp activateIgnoringOtherApps:YES];
    [self syncFocus];
    if (!pf_audio_open(&_audio)) NSLog(@"CoreAudio unavailable: %d (game remains playable)",_audio.error);
    __weak PFApp *weakSelf=self;
    pf_audio_watch_device(&_audio, ^{ [weakSelf deviceChanged]; });
    NSNotificationCenter *workspace=NSWorkspace.sharedWorkspace.notificationCenter;
    _sleepObserver=[workspace addObserverForName:NSWorkspaceWillSleepNotification object:nil
        queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n) {
            (void)n; [weakSelf sleeping:YES];
        }];
    _wakeObserver=[workspace addObserverForName:NSWorkspaceDidWakeNotification object:nil
        queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n) {
            (void)n; [weakSelf sleeping:NO];
        }];
    _timer=[NSTimer timerWithTimeInterval:1.0/120 target:self selector:@selector(step) userInfo:nil repeats:YES];
    _timer.tolerance=.001; [NSRunLoop.mainRunLoop addTimer:_timer forMode:NSRunLoopCommonModes];
    [self step];
}
- (void)sleeping:(BOOL)sleeping { _sleeping=sleeping; [self syncFocus]; }
- (void)deviceChanged {
    pf_audio_close(&_audio); pf_ring_reset(&_audio.ring);
    if (!pf_audio_open(&_audio)) NSLog(@"CoreAudio route recovery failed: %d",_audio.error);
    /* Next source wake restarts only if currently audible; no engine mutation. */
}
- (void)syncFocus {
    BOOL next=NSApp.isActive && _window.isKeyWindow && !_window.isMiniaturized && !_sleeping;
    if (!_engine || next==_focused) return;
    _focused=next;
    if (next) {
        [self check:pf_engine_resume(_engine,[self now])]; pf_input_focus(&_input,true);
    } else {
        [self check:pf_engine_suspend(_engine)]; pf_input_focus(&_input,false);
        pf_audio_pause(&_audio); _mouseActive=NO;
    }
    [self updateCursor];
}
- (void)step {
    if (!_engine || _failed) return;
    [self syncFocus];
    if (!_focused) return;
    [self check:pf_engine_advance(_engine,[self now],pf_audio_enqueue,&_audio)];
    uint64_t ticks; uint32_t mode,table,flags;
    [self check:pf_engine_state(_engine,&ticks,&mode,&table,&flags)];
    (void)ticks; (void)table;
    if (flags&2) { [NSApp terminate:nil]; return; }
    _mouseActive=(flags&4)!=0;
    if (mode==PF_MODE_PAUSED || mode==PF_MODE_QUIT_QUESTION) pf_audio_pause(&_audio);
    else pf_audio_play(&_audio);
    [self updateCursor];
    uint8_t *pixels; int32_t width,height,stride;
    [self check:pf_engine_frame(_engine,&pixels,&width,&height,&stride)];
    if (![_view acceptPixels:pixels width:width height:height stride:stride])
        [self fail:@"Invalid framebuffer or insufficient memory"];
}
- (void)emit:(PFHostEvent)e a:(int32_t)a b:(int32_t)b {
    switch(e) {
    case PF_EVENT_ACTION: [self check:pf_engine_set_action(_engine,a,b)]; break;
    case PF_EVENT_KEY: [self check:pf_engine_key(_engine,(uint8_t)a)]; break;
    case PF_EVENT_RELEASE: [self check:pf_engine_release(_engine)]; break;
    case PF_EVENT_FIRE: [self check:pf_engine_plunger_fire(_engine)]; break;
    case PF_EVENT_DELTA: [self check:pf_engine_plunger_delta(_engine,a)]; break;
    case PF_EVENT_FULLSCREEN: [_window toggleFullScreen:nil]; break;
    }
}
- (void)key:(NSEvent *)e down:(BOOL)down {
    pf_input_key(&_input,e.keyCode,down,e.isARepeat,
                 (e.modifierFlags & NSEventModifierFlagCommand)!=0,
                 (e.modifierFlags & NSEventModifierFlagOption)!=0);
}
- (void)modifiers:(NSEvent *)e {
    pf_macos_modifiers(&_input,e.modifierFlags);
}
- (void)motion:(NSEvent *)e {
    [self updateCursor];
    if (e.CGEvent) pf_input_motion(&_input,CGEventGetIntegerValueField(e.CGEvent,kCGMouseEventDeltaY),
                                  _mouseActive && _view.pointerInside);
}
- (void)button:(BOOL)down { pf_input_button(&_input,down,_mouseActive && _view.pointerInside); }
- (void)updateCursor {
    BOOL hide=_focused && _mouseActive && _view.pointerInside;
    if (hide==_cursorHidden) return;
    _cursorHidden=hide;
    if (hide) [NSCursor hide]; else [NSCursor unhide];
}
- (void)toggleFullscreen:(id)sender {
    (void)sender;
    if (_input.fullscreenPending) return;
    _input.fullscreenPending=true; [_window toggleFullScreen:nil];
}
- (void)windowDidEnterFullScreen:(NSNotification *)n { (void)n; pf_input_fullscreen_done(&_input); [self syncFocus]; }
- (void)windowDidExitFullScreen:(NSNotification *)n { (void)n; pf_input_fullscreen_done(&_input); [self syncFocus]; }
- (void)windowDidFailToEnterFullScreen:(NSWindow *)w { (void)w; pf_input_fullscreen_done(&_input); }
- (void)windowDidFailToExitFullScreen:(NSWindow *)w { (void)w; pf_input_fullscreen_done(&_input); }
- (void)windowDidBecomeKey:(NSNotification *)n { (void)n; [self syncFocus]; }
- (void)windowDidResignKey:(NSNotification *)n { (void)n; [self syncFocus]; }
- (void)windowDidMiniaturize:(NSNotification *)n { (void)n; [self syncFocus]; }
- (void)windowDidDeminiaturize:(NSNotification *)n { (void)n; [self syncFocus]; }
- (void)applicationDidBecomeActive:(NSNotification *)n { (void)n; [self syncFocus]; }
- (void)applicationDidResignActive:(NSNotification *)n { (void)n; [self syncFocus]; }
- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)app { (void)app; return YES; }
- (void)applicationWillTerminate:(NSNotification *)n {
    (void)n; [_timer invalidate];
    _focused=NO; _mouseActive=NO; [self updateCursor];
    NSNotificationCenter *workspace=NSWorkspace.sharedWorkspace.notificationCenter;
    if (_sleepObserver) [workspace removeObserver:_sleepObserver];
    if (_wakeObserver) [workspace removeObserver:_wakeObserver];
    pf_audio_unwatch_device(&_audio); pf_audio_close(&_audio);
    if (_engine) { int32_t result=pf_engine_destroy(_engine); _engine=0;
        if (result!=PF_OK) NSLog(@"Settings save failed: %d",result); }
    NSLog(@"Audio underruns=%llu dropped frames=%llu",(unsigned long long)atomic_load(&_audio.ring.underruns),
          (unsigned long long)atomic_load(&_audio.ring.dropped));
}
@end
static void menu(PFApp *host) {
    NSMenu *bar=[NSMenu new]; NSMenuItem *appItem=[NSMenuItem new]; [bar addItem:appItem];
    NSMenu *app=[NSMenu new]; [app addItemWithTitle:@"Quit Pinball Fantasies" action:@selector(terminate:) keyEquivalent:@"q"];
    appItem.submenu=app;
    NSMenuItem *viewItem=[NSMenuItem new]; [bar addItem:viewItem];
    NSMenu *view=[[NSMenu alloc] initWithTitle:@"View"];
    NSMenuItem *full=[view addItemWithTitle:@"Toggle Full Screen" action:@selector(toggleFullscreen:) keyEquivalent:@"f"];
    full.target=host; viewItem.submenu=view; NSApp.mainMenu=bar;
}
int main(int argc,const char **argv) {
    @autoreleasepool {
        /* A bounded CI launch checks AppKit startup without interactive import,
           originals or a real audio device; production follows the normal path. */
        BOOL smoke=argc==2 && strcmp(argv[1],"--ui-smoke")==0;
        NSApplication *app=NSApplication.sharedApplication;
        [app setActivationPolicy:NSApplicationActivationPolicyRegular];
        if (smoke) {
            NSWindow *window=[[NSWindow alloc] initWithContentRect:NSMakeRect(0,0,640,480)
                styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable backing:NSBackingStoreBuffered defer:NO];
            window.releasedWhenClosed=NO;
            PFFrameView *view=[[PFFrameView alloc] initWithFrame:NSMakeRect(0,0,640,480)];
            uint8_t pixels[16]={255,0,0,255,0,255,0,255,0,0,255,255,255,255,255,255};
            if (![view acceptPixels:pixels width:2 height:2 stride:8]) return 1;
            window.contentView=view; [window makeKeyAndOrderFront:nil];
            dispatch_after(dispatch_time(DISPATCH_TIME_NOW,2*NSEC_PER_SEC),dispatch_get_main_queue(),^{
                NSLog(@"PASS AppKit launch, native view and clean termination"); [NSApp terminate:nil];
            });
            [app run]; return 0;
        }
        PFApp *host=[PFApp new]; app.delegate=host; menu(host); [app run];
    }
    return 0;
}
