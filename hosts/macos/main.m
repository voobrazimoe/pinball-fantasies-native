#import "app.h"
#import "frame_view.h"
#import <mach/mach_time.h>
#import "native_input.h"
#include <stdio.h>
static const char *pacingPath;
#ifdef PF_DEMODEV
static const char *demoPath, *canonicalPath, *demoLogPath;
extern uint64_t pf_engine_create_demo(char *,char *,char *,int64_t,char *,uint32_t);
extern int32_t pf_engine_demo_diagnostic(uint64_t,char *,uint32_t);
#endif
static BOOL diagnosticsEnabled(void) {
    const char *value=getenv("PF_DIAGNOSTICS");
    return value && strcmp(value,"1")==0;
}
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
- (void)toggleShiftSides:(NSMenuItem *)sender;
- (void)noteInput:(NSEvent *)event;
- (void)reportPacing;
@end
static void inputEvent(void *context,PFHostEvent event,int32_t a,int32_t b) {
    [(__bridge PFApp *)context emit:event a:a b:b];
}
@implementation PFApp {
    NSWindow *_window;
    PFFrameView *_view;
    dispatch_source_t _timer;
    id _latencyActivity;
    uint64_t _engine, _epoch;
    mach_timebase_info_data_t _timebase;
    PFInput _input;
    PFAudio _audio;
    BOOL _focused, _mouseActive, _cursorHidden, _failed, _sleeping, _stepping;
    id _sleepObserver, _wakeObserver;
    FILE *_pacing;
    int64_t _lastStep, _lastReport;
    uint64_t _lastTicks, _reportTicks;
    uint64_t _frameTicks;
    BOOL _haveFrame;
    double _stepGap, _advanceTime, _frameTime, _drawTime, _eventDelay, _inputDraw;
    NSTimeInterval _inputTimestamp;
    BOOL _inputAdvanced;
    unsigned _draws, _events, _frames;
    uint32_t _mode;
#ifdef PF_DEMODEV
    BOOL _demoDiagnosticSeen;
#endif
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
    if (pacingPath) {
        _pacing=fopen(pacingPath,"w");
        if (!_pacing) NSLog(@"Could not open timing log: %s",pacingPath);
        else {
            fprintf(_pacing,"# Pinball macOS host timing; milliseconds; maxima per interval; input_draw includes event queue and waits for a source tick; draw_ms is layer submission, not scanout\n");
            fprintf(_pacing,"# build=%s arch=%s timer=strict-dispatch input_service=immediate renderer=core-animation shift_mapping=%s\n",
                [[NSBundle.mainBundle objectForInfoDictionaryKey:@"PFHostBuild"] UTF8String] ?: "unknown",
                [[NSBundle.mainBundle objectForInfoDictionaryKey:@"PFHostArchitecture"] UTF8String] ?: "unknown",
                [NSUserDefaults.standardUserDefaults boolForKey:@"SwapShiftKeys"] ? "swapped" : "native");
            fprintf(_pacing,"seconds,mode,ticks,frames,draws,events,step_gap_ms,advance_ms,frame_copy_ms,draw_ms,event_queue_ms,input_draw_ms\n");
            fflush(_pacing);
        }
    }
    pf_input_init(&_input,inputEvent,(__bridge void *)self);
    pf_input_swap_shift(&_input,[NSUserDefaults.standardUserDefaults boolForKey:@"SwapShiftKeys"]);
    pf_audio_init(&_audio);
#ifdef PF_DEMODEV
    char failure[1024]={0};
    _engine=pf_engine_create_demo((char *)demoPath,(char *)canonicalPath,(char *)demoLogPath,
                                 [self now],failure,sizeof(failure));
#else
    NSString *data=nil,*state=nil; NSError *error=nil;
    if (!pf_storage_prepare(&data,&state,&error)) {
        if (error) [self fail:error.localizedDescription]; else [NSApp terminate:nil];
        return;
    }
    char failure[1024]={0};
    _engine=pf_engine_create((char *)data.fileSystemRepresentation,(char *)state.fileSystemRepresentation,
                             [self now],failure,sizeof(failure));
#endif
    if (!_engine) { [self fail:[NSString stringWithUTF8String:failure]]; return; }
    [self check:pf_engine_suspend(_engine)];
    NSRect usable=NSScreen.mainScreen.visibleFrame;
    NSSize size=NSMakeSize(fmin(800,usable.size.width*.85),fmin(766,usable.size.height*.85));
    _window=[[NSWindow alloc] initWithContentRect:NSMakeRect(0,0,size.width,size.height)
        styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable|NSWindowStyleMaskMiniaturizable|NSWindowStyleMaskResizable
        backing:NSBackingStoreBuffered defer:NO];
#ifdef PF_DEMODEV
    _window.title=@"Pinball Fantasies — EXPERIMENTAL 10-minute demo (P pause; Esc exit)";
#else
    _window.title=@"Pinball Fantasies";
#endif
    _window.delegate=self; _window.releasedWhenClosed=NO;
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
    /* The engine remains serialized on main. Strict, zero-leeway wakes avoid
       the additional timer coalescing observed on the Intel host. */
    _timer=dispatch_source_create(DISPATCH_SOURCE_TYPE_TIMER,0,DISPATCH_TIMER_STRICT,dispatch_get_main_queue());
    dispatch_source_set_event_handler(_timer,^{ [weakSelf step]; });
    dispatch_source_set_timer(_timer,_focused?dispatch_time(DISPATCH_TIME_NOW,0):DISPATCH_TIME_FOREVER,
        NSEC_PER_SEC/120,0);
    dispatch_resume(_timer);
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
        _latencyActivity=[NSProcessInfo.processInfo beginActivityWithOptions:
            NSActivityUserInitiatedAllowingIdleSystemSleep|NSActivityLatencyCritical
            reason:@"Responsive pinball input and audio"];
        [self check:pf_engine_resume(_engine,[self now])]; pf_macos_focus(&_input,true,NSEvent.modifierFlags);
        _lastStep=0;
    } else {
        [self check:pf_engine_suspend(_engine)]; pf_macos_focus(&_input,false,NSEvent.modifierFlags);
        pf_audio_pause(&_audio); _mouseActive=NO;
        _inputTimestamp=0; _inputAdvanced=NO;
        if (_latencyActivity) [NSProcessInfo.processInfo endActivity:_latencyActivity];
        _latencyActivity=nil;
    }
    if (_timer) dispatch_source_set_timer(_timer,next?dispatch_time(DISPATCH_TIME_NOW,0):DISPATCH_TIME_FOREVER,
        NSEC_PER_SEC/120,0);
    [self updateCursor];
}
- (void)step {
    if (_stepping) return;
    _stepping=YES;
    @try {
        if (!_engine || _failed) return;
        [self syncFocus];
        if (!_focused) return;
        int64_t start=[self now];
        if (_pacing && _lastStep) _stepGap=fmax(_stepGap,(start-_lastStep)/1e6);
        _lastStep=start;
        uint64_t dropped=atomic_load(&_audio.ring.dropped);
        [self check:pf_engine_advance(_engine,[self now],pf_audio_enqueue,&_audio)];
        if (_pacing) _advanceTime=fmax(_advanceTime,([self now]-start)/1e6);
        if (atomic_load(&_audio.ring.dropped)!=dropped) {
            pf_audio_pause(&_audio);
            if (diagnosticsEnabled()) NSLog(@"Audio queue overrun: discarded stale PCM, source advancement preserved");
        }
        uint64_t ticks; uint32_t mode,table,flags;
        [self check:pf_engine_state(_engine,&ticks,&mode,&table,&flags)];
        _mode=mode;
        if (ticks!=_lastTicks && _inputTimestamp) _inputAdvanced=YES;
        _lastTicks=ticks;
        (void)ticks; (void)table;
        if (flags&2) { [NSApp terminate:nil]; return; }
        _mouseActive=(flags&4)!=0;
        if (mode==PF_MODE_PAUSED || mode==PF_MODE_QUIT_QUESTION) pf_audio_pause(&_audio);
        else pf_audio_play(&_audio);
        [self updateCursor];
        /* Poll at 120 Hz for timely input/source service, but rasterize only new
           source state. Rebuilding identical Go frames on intervening polls wastes
           main-thread time that AppKit needs for keyboard dispatch and drawing.
           Expose/resize draws use the view's retained copy without an engine call. */
        if (!_haveFrame || ticks!=_frameTicks) {
            uint8_t *pixels; int32_t width,height,stride;
            start=[self now];
            [self check:pf_engine_frame(_engine,&pixels,&width,&height,&stride)];
            if (![_view acceptPixels:pixels width:width height:height stride:stride])
                [self fail:@"Invalid framebuffer or insufficient memory"];
            _frameTicks=ticks; _haveFrame=YES; _frames++;
            if (_pacing) _frameTime=fmax(_frameTime,([self now]-start)/1e6);
        }
#ifdef PF_DEMODEV
        if (!_demoDiagnosticSeen) {
            char diagnostic[8192]={0};
            [self check:pf_engine_demo_diagnostic(_engine,diagnostic,sizeof(diagnostic))];
            if (diagnostic[0]) {
                _demoDiagnosticSeen=YES; pf_audio_pause(&_audio);
                _window.title=@"Experimental demo stopped — unsupported transition (Esc closes)";
                NSAlert *alert=[NSAlert new];
                alert.messageText=@"Experimental demo stopped safely";
                alert.informativeText=[NSString stringWithFormat:@"This gameplay transition is not implemented. The session is frozen. Close the window or press Escape.\nLog: %s\n\n%s",demoLogPath,diagnostic];
                [alert addButtonWithTitle:@"Return to stopped session"]; [alert runModal];
            }
        }
#endif
        [self reportPacing];
    } @finally { _stepping=NO; }
}
- (void)noteInput:(NSEvent *)e {
    if (!_pacing || !_focused) return;
    _events++;
    _eventDelay=fmax(_eventDelay,(NSProcessInfo.processInfo.systemUptime-e.timestamp)*1000);
    if (!_inputTimestamp) _inputTimestamp=e.timestamp;
}
- (void)frameDrawnFrom:(uint64_t)start {
    if (!_pacing) return;
    _draws++;
    _drawTime=fmax(_drawTime,pf_clock_ns(mach_absolute_time(),start,_timebase.numer,_timebase.denom)/1e6);
    if (_inputAdvanced && _inputTimestamp) {
        _inputDraw=fmax(_inputDraw,(NSProcessInfo.processInfo.systemUptime-_inputTimestamp)*1000);
        _inputTimestamp=0; _inputAdvanced=NO;
    }
}
- (void)reportPacing {
    if (!_pacing) return;
    int64_t now=[self now];
    if (now-_lastReport<1000000000LL) return;
    fprintf(_pacing,"%.3f,%u,%llu,%u,%u,%u,%.3f,%.3f,%.3f,%.3f,%.3f,%.3f\n",
        now/1e9,_mode,(unsigned long long)(_lastTicks-_reportTicks),_frames,_draws,_events,
        _stepGap,_advanceTime,_frameTime,_drawTime,_eventDelay,_inputDraw);
    fflush(_pacing);
    _lastReport=now; _reportTicks=_lastTicks;
    _draws=_events=_frames=0;
    _stepGap=_advanceTime=_frameTime=_drawTime=_eventDelay=_inputDraw=0;
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
    [self noteInput:e];
    pf_input_key(&_input,e.keyCode,down,e.isARepeat,
                 (e.modifierFlags & NSEventModifierFlagCommand)!=0,
                 (e.modifierFlags & NSEventModifierFlagOption)!=0);
    /* Consume already-due source work using the newly submitted input instead
       of waiting for another timer wake. Never advance a future source tick. */
    [self step];
}
- (void)modifiers:(NSEvent *)e {
    [self noteInput:e];
    pf_macos_modifiers(&_input,e.keyCode,e.modifierFlags);
    if (_pacing) fprintf(_pacing,"# modifier seconds=%.6f key=%u flags=0x%llx sides=%d%d%d%d%d%d flippers=%d%d\n",
        [self now]/1e9,e.keyCode,(unsigned long long)e.modifierFlags,
        _input.physicalModifiers[0],_input.physicalModifiers[1],_input.physicalModifiers[2],
        _input.physicalModifiers[3],_input.physicalModifiers[4],_input.physicalModifiers[5],
        _input.held[PF_LEFT],_input.held[PF_RIGHT]);
    [self step];
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
- (void)toggleShiftSides:(NSMenuItem *)sender {
    BOOL swapped=!_input.swapShift;
    pf_input_swap_shift(&_input,swapped);
    [NSUserDefaults.standardUserDefaults setBool:swapped forKey:@"SwapShiftKeys"];
    sender.state=swapped?NSControlStateValueOn:NSControlStateValueOff;
    if (_pacing) fprintf(_pacing,"# shift_mapping=%s seconds=%.6f\n",swapped?"swapped":"native",[self now]/1e9);
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
    (void)n;
    if (_timer) dispatch_source_cancel(_timer);
    _timer=nil;
    if (_latencyActivity) [NSProcessInfo.processInfo endActivity:_latencyActivity];
    _latencyActivity=nil;
    _focused=NO; _mouseActive=NO; [self updateCursor];
    NSNotificationCenter *workspace=NSWorkspace.sharedWorkspace.notificationCenter;
    if (_sleepObserver) [workspace removeObserver:_sleepObserver];
    if (_wakeObserver) [workspace removeObserver:_wakeObserver];
    pf_audio_unwatch_device(&_audio); pf_audio_close(&_audio);
    if (_engine) { int32_t result=pf_engine_destroy(_engine); _engine=0;
        if (result!=PF_OK) NSLog(@"Settings save failed: %d",result); }
    if (diagnosticsEnabled()) NSLog(@"Audio underruns=%llu dropped frames=%llu",(unsigned long long)atomic_load(&_audio.ring.underruns),
          (unsigned long long)atomic_load(&_audio.ring.dropped));
    if (_pacing) { fclose(_pacing); _pacing=NULL; }
}
@end
static void menu(PFApp *host) {
    NSMenu *bar=[NSMenu new]; NSMenuItem *appItem=[NSMenuItem new]; [bar addItem:appItem];
    NSMenu *app=[NSMenu new]; [app addItemWithTitle:@"Quit Pinball Fantasies" action:@selector(terminate:) keyEquivalent:@"q"];
    appItem.submenu=app;
    NSMenuItem *viewItem=[NSMenuItem new]; [bar addItem:viewItem];
    NSMenu *view=[[NSMenu alloc] initWithTitle:@"View"];
    NSMenuItem *full=[view addItemWithTitle:@"Toggle Full Screen" action:@selector(toggleFullscreen:) keyEquivalent:@"f"];
    full.target=host;
    [view addItem:NSMenuItem.separatorItem];
    NSMenuItem *shift=[view addItemWithTitle:@"Swap Left/Right Shift" action:@selector(toggleShiftSides:) keyEquivalent:@""];
    shift.target=host;
    shift.state=[NSUserDefaults.standardUserDefaults boolForKey:@"SwapShiftKeys"]?NSControlStateValueOn:NSControlStateValueOff;
    viewItem.submenu=view; NSApp.mainMenu=bar;
}
int main(int argc,const char **argv) {
    @autoreleasepool {
        /* A bounded CI launch checks AppKit startup without interactive import,
           originals or a real audio device; production follows the normal path. */
        BOOL smoke=argc==2 && strcmp(argv[1],"--ui-smoke")==0;
        pacingPath=getenv("PF_PACING_LOG");
        if (argc==3 && strcmp(argv[1],"--pacing-log")==0) pacingPath=argv[2];
#ifdef PF_DEMODEV
        if (!smoke) {
            for (int i=1;i<argc;i++) {
                if (strcmp(argv[i],"--experimental-10min-demo")==0 && i+1<argc) demoPath=argv[++i];
                else if (strcmp(argv[i],"--canonical-data-dir")==0 && i+1<argc) canonicalPath=argv[++i];
                else if (strcmp(argv[i],"--demo-log")==0 && i+1<argc) demoLogPath=argv[++i];
                else { fprintf(stderr,"Unknown or incomplete experimental argument: %s\n",argv[i]); return 2; }
            }
            if (!demoPath || !canonicalPath || !demoLogPath) {
                fprintf(stderr,"Usage: pinballfantasies --experimental-10min-demo DIR --canonical-data-dir DIR --demo-log FILE\n"); return 2;
            }
        }
#endif
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
