#import "frame_view.h"
#import <QuartzCore/QuartzCore.h>
#import <IOKit/hidsystem/IOLLEvent.h>
#import "storage.h"
#import "audio_host.h"
#import "native_input.h"
#import "gamepad.h"
#import <GameController/GameController.h>
#import <mach/mach_time.h>
#include <assert.h>
#include <stdio.h>
static void audioTests(void) {
    PFAudio audio={0}; pf_audio_init(&audio);
    int16_t input[4]={123,-456,789,-1024}, output[8];
    memset(output,0x7f,sizeof(output));
    pf_audio_enqueue(&audio,(const uint8_t *)input,sizeof(input));
    AudioBufferList buffers={0}; buffers.mNumberBuffers=1;
    buffers.mBuffers[0]=(AudioBuffer){2,sizeof(output),output};
    assert(pf_audio_render(&audio,NULL,NULL,0,4,&buffers)==noErr);
    assert(!memcmp(output,input,sizeof(input)));
    assert(output[4]==0 && output[7]==0 && pf_ring_available(&audio.ring)==0);
    assert(atomic_load(&audio.ring.underruns)==1);
    pf_audio_enqueue(&audio,(const uint8_t *)input,sizeof(input));
    pf_audio_pause(&audio); assert(pf_ring_available(&audio.ring)==0);
    int16_t *full=calloc(PF_RING_FRAMES*2,sizeof(int16_t)); assert(full);
    pf_audio_enqueue(&audio,(const uint8_t *)full,PF_RING_FRAMES*4);
    audio.running=true; /* Simulated active-but-stalled device: enqueue must not wait. */
    pf_audio_enqueue(&audio,(const uint8_t *)input,sizeof(input));
    assert(atomic_load(&audio.ring.dropped)==2 && pf_ring_available(&audio.ring)==PF_RING_FRAMES);
    pf_audio_pause(&audio); assert(!audio.running && pf_ring_available(&audio.ring)==0); free(full);
    puts("PASS actual AudioUnit render callback: S16 stereo, silence, bounded overrun and stopped flush (no device required)");
}
static unsigned modifierMakes, modifierActions, modifierFullscreen;
static void modifierEmit(void *ctx,PFHostEvent event,int32_t a,int32_t b) {
    (void)ctx; (void)b;
    if (event==PF_EVENT_KEY) { assert(a==127); modifierMakes++; }
    if (event==PF_EVENT_ACTION) modifierActions++;
    if (event==PF_EVENT_FULLSCREEN) modifierFullscreen++;
}
static void modifierEvent(PFInput *input,uint16_t key,NSEventModifierFlags flags) {
    NSEvent *event=[NSEvent keyEventWithType:NSEventTypeFlagsChanged
        location:NSZeroPoint modifierFlags:flags timestamp:0 windowNumber:0
        context:nil characters:@"" charactersIgnoringModifiers:@"" isARepeat:NO keyCode:key];
    assert(event && event.keyCode==key && event.modifierFlags==flags);
    pf_macos_modifiers(input,event.keyCode,event.modifierFlags);
}
static void modifierTests(void) {
    const uint16_t keys[3][2]={{56,60},{59,62},{58,61}};
    const NSEventModifierFlags flags[3]={NSEventModifierFlagShift,
        NSEventModifierFlagControl,NSEventModifierFlagOption};
    PFInput input;
    const uint64_t sideFlags[3][2]={{NX_DEVICELSHIFTKEYMASK,NX_DEVICERSHIFTKEYMASK},
        {NX_DEVICELCTLKEYMASK,NX_DEVICERCTLKEYMASK},{NX_DEVICELALTKEYMASK,NX_DEVICERALTKEYMASK}};
    for (unsigned c=0;c<3;c++) {
        pf_input_init(&input,modifierEmit,NULL); pf_macos_focus(&input,true,0);
        modifierMakes=modifierActions=0;
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][0]);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][0]); /* duplicate make */
        assert(input.held[PF_LEFT] && !input.held[PF_RIGHT] && modifierMakes==1);
        modifierEvent(&input,keys[c][1],flags[c]|sideFlags[c][0]|sideFlags[c][1]);
        assert(input.held[PF_LEFT] && input.held[PF_RIGHT]);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][1]);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][1]); /* duplicate break */
        assert(!input.held[PF_LEFT] && input.held[PF_RIGHT] && modifierMakes==2);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][0]); /* missed right break */
        assert(input.held[PF_LEFT] && !input.held[PF_RIGHT] && modifierMakes==3);
        pf_macos_focus(&input,false,flags[c]|sideFlags[c][0]);
        pf_macos_focus(&input,true,flags[c]|sideFlags[c][0]);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][0]);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]); /* held-before-focus stays suppressed */
        modifierEvent(&input,keys[c][0],0);
        modifierEvent(&input,keys[c][0],flags[c]|sideFlags[c][0]);
        assert(input.held[PF_LEFT] && !input.held[PF_RIGHT]);
        modifierEvent(&input,keys[c][0],0);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
    }
    for (unsigned c=0;c<3;c++) for (unsigned first=0;first<2;first++) {
        pf_input_init(&input,modifierEmit,NULL); pf_macos_focus(&input,true,0);
        modifierMakes=modifierActions=0;
        modifierEvent(&input,keys[c][first],flags[c]);
        assert(input.held[first] && !input.held[1-first]);
        modifierEvent(&input,keys[c][1-first],flags[c]);
        assert(input.held[PF_LEFT] && input.held[PF_RIGHT]);
        modifierEvent(&input,keys[c][first],flags[c]);
        assert(!input.held[first] && input.held[1-first]);
        modifierEvent(&input,keys[c][1-first],0);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
        assert(modifierMakes==2 && modifierActions==4);

        /* The old side remains physically held across loss. A fresh opposite
           side works, its release keeps the old side suppressed, and the old
           side's break with the sibling still held cannot become a make. */
        modifierEvent(&input,keys[c][first],flags[c]);
        pf_macos_focus(&input,false,flags[c]);
        pf_macos_focus(&input,true,flags[c]);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
        modifierMakes=0;
        modifierEvent(&input,keys[c][1-first],flags[c]);
        assert(!input.held[first] && input.held[1-first] && modifierMakes==1);
        modifierEvent(&input,keys[c][1-first],flags[c]);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT] && modifierMakes==1);
        modifierEvent(&input,keys[c][1-first],flags[c]);
        assert(!input.held[first] && input.held[1-first] && modifierMakes==2);
        modifierEvent(&input,keys[c][first],flags[c]);
        assert(!input.held[first] && input.held[1-first] && modifierMakes==2);
        modifierEvent(&input,keys[c][first],flags[c]);
        assert(input.held[first] && input.held[1-first] && modifierMakes==3);
        modifierEvent(&input,keys[c][first],flags[c]);
        modifierEvent(&input,keys[c][1-first],0);

        /* A break missed while inactive is reconciled by public aggregate
           flags on regain, so the next same-side press works immediately. */
        modifierEvent(&input,keys[c][first],flags[c]);
        pf_macos_focus(&input,false,flags[c]); pf_macos_focus(&input,true,0);
        modifierEvent(&input,keys[c][first],flags[c]); assert(input.held[first]);
        pf_macos_focus(&input,false,flags[c]);
        modifierEvent(&input,keys[c][first],0); /* break delivered inactive */
        modifierEvent(&input,keys[c][1-first],flags[c]); /* inactive make */
        pf_macos_focus(&input,true,flags[c]);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
        modifierEvent(&input,keys[c][1-first],0);
        assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
    }
    pf_input_init(&input,modifierEmit,NULL); pf_macos_focus(&input,true,0);
    modifierMakes=modifierActions=0;
    modifierEvent(&input,56,NSEventModifierFlagShift);
    modifierEvent(&input,59,NSEventModifierFlagShift|NSEventModifierFlagControl);
    modifierEvent(&input,58,NSEventModifierFlagShift|NSEventModifierFlagControl|NSEventModifierFlagOption);
    assert(modifierActions==1 && modifierMakes==3);
    modifierEvent(&input,56,NSEventModifierFlagControl|NSEventModifierFlagOption);
    modifierEvent(&input,59,NSEventModifierFlagOption);
    assert(input.held[PF_LEFT] && modifierActions==1);
    modifierEvent(&input,55,NSEventModifierFlagCommand);
    modifierEvent(&input,54,0); modifierEvent(&input,57,NSEventModifierFlagCapsLock);
    modifierEvent(&input,63,0);
    assert(input.held[PF_LEFT] && modifierActions==1 && modifierMakes==3);
    modifierEvent(&input,58,0); assert(!input.held[PF_LEFT] && modifierActions==2);

    /* A different class held over focus loss cannot suppress a fresh press. */
    modifierEvent(&input,60,NSEventModifierFlagShift);
    pf_macos_focus(&input,false,NSEventModifierFlagShift);
    pf_macos_focus(&input,true,NSEventModifierFlagShift);
    modifierEvent(&input,58,NSEventModifierFlagShift|NSEventModifierFlagOption);
    assert(input.held[PF_LEFT] && !input.held[PF_RIGHT]);
    modifierFullscreen=0;
    pf_input_key(&input,36,true,false,false,true);
    assert(modifierFullscreen==1);
    pf_input_key(&input,36,false,false,false,false); pf_input_fullscreen_done(&input);
    pf_input_key(&input,3,true,false,true,false); assert(modifierFullscreen==2);
    pf_input_key(&input,12,true,false,true,false); assert(!input.down[12]); /* Command-Q */
    pf_macos_focus(&input,false,NSEventModifierFlagShift|NSEventModifierFlagOption);
    assert(!input.held[PF_LEFT] && !input.held[PF_RIGHT]);
    mach_timebase_info_data_t scale; assert(mach_timebase_info(&scale)==KERN_SUCCESS);
    uint64_t epoch=mach_absolute_time(); assert(pf_clock_ns(epoch,epoch,scale.numer,scale.denom)==0);
    int64_t previous=0;
    for (unsigned i=0;i<1000;i++) {
        int64_t now=pf_clock_ns(mach_absolute_time(),epoch,scale.numer,scale.denom);
        assert(now>=previous); previous=now;
    }
    puts("PASS public AppKit modifier events, sided holds, focus reconciliation, shortcuts and monotonic clock");
}
static BOOL acceptTestFiles(NSString *data,NSError **error) {
    (void)error;
    for (NSString *name in pf_required_assets()) {
        NSData *bytes=[NSData dataWithContentsOfFile:[data stringByAppendingPathComponent:name]];
        if (bytes.length!=4) return NO;
    }
    return YES;
}
static BOOL acceptDemoTestFiles(NSString *data,NSError **error) {
    (void)error;
    for (NSString *name in pf_demo_assets()) {
        NSData *bytes=[NSData dataWithContentsOfFile:[data stringByAppendingPathComponent:name]];
        if (bytes.length!=4) return NO;
    }
    return YES;
}
static void demoStorageTests(void) {
    NSFileManager *fm=NSFileManager.defaultManager; NSError *error=nil;
    NSString *root=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
    NSString *source=[root stringByAppendingPathComponent:@"pinbfan"];
    NSString *data=[root stringByAppendingPathComponent:@"Data"];
    assert([fm createDirectoryAtPath:source withIntermediateDirectories:YES attributes:nil error:&error]);
    for (NSString *name in [pf_demo_assets() arrayByAddingObject:@"PINBALL.EXE"]) assert([[@"test" dataUsingEncoding:NSUTF8StringEncoding]
        writeToFile:[source stringByAppendingPathComponent:name] atomically:YES]);
    assert([pf_installation_assets(source) isEqualToArray:pf_demo_assets()]);
    assert(pf_import_assets(source,data,acceptDemoTestFiles,&error));
    assert([fm contentsOfDirectoryAtPath:data error:&error].count==6); /* five roles + launcher text */
    assert([fm fileExistsAtPath:[data stringByAppendingPathComponent:@"PINBALL.EXE"]]);
    assert([fm removeItemAtPath:[source stringByAppendingPathComponent:@"MOD2.MOD"] error:&error]);
    assert(!pf_import_assets(source,data,acceptDemoTestFiles,&error));
    assert([fm contentsOfDirectoryAtPath:data error:NULL].count==6);
    assert([[@"test" dataUsingEncoding:NSUTF8StringEncoding] writeToFile:[source stringByAppendingPathComponent:@"TABLE3.PRG"] atomically:YES]);
    assert([pf_installation_assets(source) isEqualToArray:pf_required_assets()]);
    [fm removeItemAtPath:root error:NULL];
    /* The demo shipped in Resources/Demo must load in the shared engine. */
    NSString *bundled=[fm.currentDirectoryPath stringByAppendingPathComponent:@"internal/demodata/assets/demo"];
    assert(pf_validate_assets(bundled,&error));
    puts("PASS demo-shaped import selects the five demo runtime roles; bundled demo validates");
}
static void storageTests(void) {
    NSFileManager *fm=NSFileManager.defaultManager; NSError *error=nil;
    NSString *root=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
    NSString *source=[root stringByAppendingPathComponent:@"Original files 日本"];
    NSString *data=[root stringByAppendingPathComponent:@"Data"];
    assert([fm createDirectoryAtPath:source withIntermediateDirectories:YES attributes:nil error:&error]);
    for (NSString *name in pf_required_assets()) assert([[@"test" dataUsingEncoding:NSUTF8StringEncoding]
        writeToFile:[source stringByAppendingPathComponent:name] atomically:YES]);
    assert([[@"not imported" dataUsingEncoding:NSUTF8StringEncoding] writeToFile:[source stringByAppendingPathComponent:@"TABLE1.HI"] atomically:YES]);
    assert(pf_import_assets(source,data,acceptTestFiles,&error));
    assert([fm contentsOfDirectoryAtPath:data error:&error].count==11);
    assert(![fm fileExistsAtPath:[data stringByAppendingPathComponent:@"TABLE1.HI"]]);
    assert(pf_import_assets(source,data,acceptTestFiles,&error)); /* replacement transaction */
    NSString *input=[source stringByAppendingPathComponent:@"TABLE4.PRG"];
    assert([fm removeItemAtPath:input error:&error]);
    assert(!pf_import_assets(source,data,acceptTestFiles,&error));
    assert([fm contentsOfDirectoryAtPath:data error:NULL].count==11); /* original adopted set preserved */
    assert([fm createSymbolicLinkAtPath:input withDestinationPath:[source stringByAppendingPathComponent:@"TABLE3.PRG"] error:&error]);
    assert(!pf_import_assets(source,data,acceptTestFiles,&error));
    assert(!pf_validate_assets(data,&error)); /* real ABI rejects synthetic commercial files */
    NSString *path=pf_support_path(@"/Users/Test/Library/Application Support");
    NSString *cwd=fm.currentDirectoryPath; assert([fm changeCurrentDirectoryPath:@"/"]);
    assert([path isEqualToString:pf_support_path(@"/Users/Test/Library/Application Support")]);
    assert([fm changeCurrentDirectoryPath:cwd]); assert([fm removeItemAtPath:root error:&error]);
    puts("PASS import whitelist, staged replacement, missing/symlink rejection, ABI validation, CWD-independent storage");
}
static void frameTests(void) {
    PFFrameView *view=[[PFFrameView alloc] initWithFrame:NSMakeRect(0,0,8,4)];
    uint8_t pixels[16]={255,0,0,255,0,255,0,255,0,0,255,255,255,255,255,255};
    assert([view acceptPixels:pixels width:2 height:2 stride:8]);
    memset(pixels,0,sizeof(pixels)); /* proves the presenter owns its copy */
    assert(![view acceptPixels:pixels width:2 height:2 stride:1]);
    NSBitmapImageRep *bitmap=[[NSBitmapImageRep alloc] initWithBitmapDataPlanes:NULL pixelsWide:8 pixelsHigh:4
        bitsPerSample:8 samplesPerPixel:4 hasAlpha:YES isPlanar:NO colorSpaceName:NSDeviceRGBColorSpace
        bytesPerRow:32 bitsPerPixel:32];
    NSGraphicsContext *ctx=[NSGraphicsContext graphicsContextWithBitmapImageRep:bitmap];
    [NSGraphicsContext saveGraphicsState]; NSGraphicsContext.currentContext=ctx;
    CGContextTranslateCTM(ctx.CGContext,0,4); CGContextScaleCTM(ctx.CGContext,1,-1);
    CGContextSetInterpolationQuality(ctx.CGContext,kCGInterpolationNone);
    [view updateLayer];
    [view.layer renderInContext:ctx.CGContext]; [NSGraphicsContext restoreGraphicsState];
    /* Black bars plus exact source colours: no blended/scaled filtering. */
    unsigned red=0,green=0,blue=0,white=0,black=0;
    for (int y=0;y<4;y++) for (int x=0;x<8;x++) {
        unsigned char *p=bitmap.bitmapData+y*32+x*4;
        if (p[0]==255 && p[1]==0 && p[2]==0) red++;
        else if (p[0]==0 && p[1]==255 && p[2]==0) green++;
        else if (p[0]==0 && p[1]==0 && p[2]==255) blue++;
        else if (p[0]==255 && p[1]==255 && p[2]==255) white++;
        else if (p[0]==0 && p[1]==0 && p[2]==0) black++;
    }
    assert(red==4 && green==4 && blue==4 && white==4 && black==16);
    /* Top source row remains at the top; green stays to the right of red. */
    unsigned char *top=bitmap.bitmapData+2*4;
    unsigned char *bottom=bitmap.bitmapData+3*32+2*4;
    assert(top[0]==255 && top[1]==0 && top[2]==0);
    assert(bottom[0]==0 && bottom[1]==0 && bottom[2]==255);
    CALayer *frame=view.layer.sublayers.firstObject;
    assert([frame.magnificationFilter isEqualToString:kCAFilterNearest]);
    assert([frame.minificationFilter isEqualToString:kCAFilterNearest]);
    CGImageRef previous=CGImageRetain((__bridge CGImageRef)frame.contents);
    assert([view acceptPixels:pixels width:2 height:2 stride:8]);
    [view updateLayer];
    CFDataRef oldData=CGDataProviderCopyData(CGImageGetDataProvider(previous));
    assert(CFDataGetBytePtr(oldData)[0]==255); /* compositor's old image is immutable */
    CFRelease(oldData); CGImageRelease(previous);
    uint8_t *selector=calloc(640*240,4); assert(selector);
    assert([view acceptPixels:selector width:640 height:240 stride:640*4]); free(selector);
    [view updateLayer];
    assert(fabs(frame.frame.size.width-16.0/3)<1e-6 && frame.frame.size.height==4);
    view.frame=NSMakeRect(0,0,4,8); [view layout];
    assert(frame.frame.size.width==4 && frame.frame.size.height==3 && frame.frame.origin.y==2.5);
    assert(![frame animationForKey:@"contents"] && ![frame animationForKey:@"bounds"]);
    puts("PASS layer presenter: immutable frames, orientation, nearest colours, bars, selector aspect, resize without animation");
}
static void abiTests(void) {
    assert(pf_engine_abi_version()==PF_ABI_VERSION);
    char error[256]={0}; assert(!pf_engine_create("/absent-originals","/tmp",0,error,sizeof(error)) && error[0]);
    assert(pf_engine_destroy(0)==PF_INVALID);
    assert(pf_engine_suspend(0)==PF_INVALID); assert(pf_engine_resume(0,0)==PF_INVALID);
    assert(pf_engine_set_action(0,PF_LEFT,1)==PF_INVALID); assert(pf_engine_key(0,57)==PF_INVALID);
    assert(pf_engine_gamepad(0,0,0,1)==PF_INVALID);
    assert(pf_engine_release(0)==PF_INVALID); assert(pf_engine_plunger_delta(0,8)==PF_INVALID);
    assert(pf_engine_plunger_fire(0)==PF_INVALID); assert(pf_engine_advance(0,0,NULL,NULL)==PF_INVALID);
    uint8_t *pixels; int32_t w,h,s; uint64_t ticks; uint32_t mode,table,flags;
    assert(pf_engine_frame(0,&pixels,&w,&h,&s)==PF_INVALID);
    assert(pf_engine_state(0,&ticks,&mode,&table,&flags)==PF_INVALID);
    puts("PASS real Apple-linked C archive: ABI version, load errors, all invalid-handle exports");
}
typedef struct { uint64_t engine, pcmFrames; PFRing *ring; } PFJourney;
static void journeyInput(void *context,PFHostEvent event,int32_t a,int32_t b) {
    PFJourney *host=context; int32_t result=PF_OK;
    switch (event) {
    case PF_EVENT_ACTION: result=pf_engine_set_action(host->engine,a,b); break;
    case PF_EVENT_KEY: result=pf_engine_key(host->engine,(uint8_t)a); break;
    case PF_EVENT_RELEASE: result=pf_engine_release(host->engine); break;
    case PF_EVENT_DELTA: result=pf_engine_plunger_delta(host->engine,a); break;
    case PF_EVENT_FIRE: result=pf_engine_plunger_fire(host->engine); break;
    case PF_EVENT_FULLSCREEN: break; /* window command tested separately */
    }
    assert(result==PF_OK);
}
static void pcm(void *context,const uint8_t *samples,uint32_t bytes) {
    PFJourney *host=context; assert(bytes%4==0); host->pcmFrames+=bytes/4;
    assert(pf_ring_write(host->ring,samples,bytes/4)==bytes/4);
    int16_t drain[4096];
    while (pf_ring_available(host->ring)) pf_ring_read(host->ring,drain,2048);
}
static void tap(PFInput *input,uint16_t native) {
    pf_input_key(input,native,true,false,false,false);
    pf_input_key(input,native,false,false,false,false);
}
static void journey(NSString *data) {
    for (unsigned table=1;table<=4;table++) for (unsigned scroll=0;scroll<4;scroll++)
    for (unsigned resolution=0;resolution<2;resolution++) {
        @autoreleasepool {
        NSString *state=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
        NSFileManager *fm=NSFileManager.defaultManager;
        assert([fm createDirectoryAtPath:state withIntermediateDirectories:NO attributes:nil error:NULL]);
        uint8_t cfg[]={'P','F','N','C',1,5,0,1,(uint8_t)scroll,0,(uint8_t)resolution};
        assert([[NSData dataWithBytes:cfg length:sizeof(cfg)] writeToFile:[state stringByAppendingPathComponent:@"PINBALL.CFG"] atomically:YES]);
        char error[1024]; uint64_t engine=pf_engine_create((char *)data.fileSystemRepresentation,(char *)state.fileSystemRepresentation,0,error,sizeof(error));
        if (!engine) { fprintf(stderr,"%s\n",error); abort(); }
        PFJourney host={engine,0,calloc(1,sizeof(PFRing))}; assert(host.ring); pf_ring_init(host.ring);
        PFInput input; pf_input_init(&input,journeyInput,&host); pf_input_focus(&input,true);
        PFFrameView *view=[[PFFrameView alloc] initWithFrame:NSMakeRect(0,0,640,480)];
        int64_t now=0; uint64_t ticks=0; uint32_t mode=0,selected=0,flags=0;
        for (unsigned step=0;step<600;step++,now+=16666667) {
            if (step==1) tap(&input,49);
            const uint16_t nativeF[4]={122,120,99,118};
            if (step==2) tap(&input,nativeF[table-1]);
            if (step==3) tap(&input,122);
            if (step==220) pf_input_key(&input,125,true,false,false,false);
            if (step==252) pf_input_key(&input,125,false,false,false,false);
            bool modifiers[6]={step>=350 && step<370,step>=360 && step<380,false,false,false,false};
            pf_input_modifiers(&input,modifiers);
            pf_input_motion(&input,8,(flags&4)!=0);
            pf_input_button(&input,step==280,(flags&4)!=0);
            if (step==300) {
                uint64_t before=ticks, count=host.pcmFrames;
                assert(pf_engine_suspend(engine)==PF_OK); pf_input_focus(&input,false);
                now+=3600000000000LL;
                assert(pf_engine_advance(engine,now,pcm,&host)==PF_OK);
                assert(pf_engine_state(engine,&ticks,&mode,&selected,&flags)==PF_OK);
                assert(ticks==before && host.pcmFrames==count && (flags&1));
                assert(mode==PF_MODE_PAUSED);
                assert(pf_engine_resume(engine,now)==PF_OK); pf_input_focus(&input,true);
                assert(pf_engine_state(engine,&ticks,&mode,&selected,&flags)==PF_OK);
                assert(ticks==before && mode==PF_MODE_PAUSED && !(flags&1));
                tap(&input,35); /* ordinary P make; no implicit unpause */
            }
            uint64_t before=ticks;
            assert(pf_engine_advance(engine,now,pcm,&host)==PF_OK);
            assert(pf_engine_state(engine,&ticks,&mode,&selected,&flags)==PF_OK);
            if (step==3) assert(mode==PF_MODE_PLAYING && selected==table);
            if (step==300) assert(ticks<=before+1); /* reanchored, never one-hour catch-up */
            uint8_t *pixels; int32_t w,h,stride;
            assert(pf_engine_frame(engine,&pixels,&w,&h,&stride)==PF_OK);
            assert([view acceptPixels:pixels width:w height:h stride:stride]);
        }
        assert(ticks>500 && selected==table && !(flags&1) && host.pcmFrames>0);
        assert(pf_engine_destroy(engine)==PF_OK); free(host.ring);
        assert([fm fileExistsAtPath:[state stringByAppendingPathComponent:@"PINBALL.CFG"]]);
        assert([fm removeItemAtPath:state error:NULL]);
        printf("PASS Apple ABI host journey table=%u scroll=%u resolution=%u ticks=%llu mode=%u PCM=%llu\n",
            table,scroll,resolution,(unsigned long long)ticks,mode,(unsigned long long)host.pcmFrames);
        }
    }
}

static void bundledStorageTests(NSString *originals) {
    NSFileManager *fm=NSFileManager.defaultManager;
    NSString *root=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
    NSString *resources=[root stringByAppendingPathComponent:@"Resources"];
    NSString *bundled=[resources stringByAppendingPathComponent:@"Data"];
    NSString *base=[root stringByAppendingPathComponent:@"Support"];
    assert([fm createDirectoryAtPath:bundled withIntermediateDirectories:YES attributes:nil error:NULL]);
    for (NSString *name in pf_required_assets())
        assert([fm copyItemAtPath:[originals stringByAppendingPathComponent:name]
            toPath:[bundled stringByAppendingPathComponent:name] error:NULL]);
    NSString *data=nil,*state=nil; NSError *error=nil;
    assert(pf_storage_prepare_at(base,resources,&data,&state,&error));
    assert([data isEqualToString:[[base stringByAppendingPathComponent:@"PinballFantasies"] stringByAppendingPathComponent:@"Data"]]);
    assert([fm fileExistsAtPath:state] && pf_validate_assets(data,&error));
    NSString *marker=[data stringByAppendingPathComponent:@"preserved"];
    assert([@"existing import" writeToFile:marker atomically:YES encoding:NSUTF8StringEncoding error:NULL]);
    assert(pf_storage_prepare_at(base,resources,&data,&state,&error));
    assert([fm fileExistsAtPath:marker]); /* relaunch preserves existing valid import */
    for (NSString *name in pf_required_assets()) {
        assert([[NSData dataWithContentsOfFile:[bundled stringByAppendingPathComponent:name]]
            isEqualToData:[NSData dataWithContentsOfFile:[originals stringByAppendingPathComponent:name]]]);
    }
    assert([fm removeItemAtPath:root error:NULL]);
    puts("PASS personal bundle first launch, existing import preservation and read-only bundled originals");
}

static void compatibilityImportTests(NSString *originals) {
    NSFileManager *fm=NSFileManager.defaultManager;
    NSString *root=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
    NSString *source=[root stringByAppendingPathComponent:@"source"];
    NSString *destination=[root stringByAppendingPathComponent:@"Data"];
    assert([fm createDirectoryAtPath:source withIntermediateDirectories:YES attributes:nil error:NULL]);
    for (NSString *name in pf_required_assets()) {
        assert([fm copyItemAtPath:[originals stringByAppendingPathComponent:name]
                          toPath:[source stringByAppendingPathComponent:name] error:NULL]);
    }
    NSError *error=nil;
    assert(pf_import_assets(source,destination,pf_validate_assets,&error)); /* CFG absent */
    const uint8_t legacy[]={1,1,2,1,1,1};
    assert([[NSData dataWithBytes:legacy length:sizeof(legacy)] writeToFile:[source stringByAppendingPathComponent:@"PINBALL.CFG"] atomically:YES]);
    for (NSString *name in pf_required_assets()) {
        if (![name hasSuffix:@".PRG"]) continue;
        NSString *path=[source stringByAppendingPathComponent:name];
        NSMutableData *bytes=[NSMutableData dataWithContentsOfFile:path];
        assert(bytes.length); ((uint8_t *)bytes.mutableBytes)[0]^=1; /* unconsumed MZ byte */
        assert([bytes writeToFile:path atomically:YES]);
    }
    assert(pf_import_assets(source,destination,pf_validate_assets,&error));
    NSData *accepted=[NSData dataWithContentsOfFile:[destination stringByAppendingPathComponent:@"TABLE1.PRG"]];
    NSString *path=[source stringByAppendingPathComponent:@"TABLE1.PRG"];
    NSMutableData *bytes=[NSMutableData dataWithContentsOfFile:path];
    bytes.length=100; /* truncated consumed data rejects every registered layout */
    assert([bytes writeToFile:path atomically:YES]);
    assert(!pf_import_assets(source,destination,pf_validate_assets,&error));
    assert([accepted isEqualToData:[NSData dataWithContentsOfFile:[destination stringByAppendingPathComponent:@"TABLE1.PRG"]]]);
    assert([fm removeItemAtPath:root error:NULL]);
    puts("PASS shared macOS import compatibility: absent/modified CFG, unused PRG bytes, truncation rejection, prior import preserved");
}

static void gamepadTests(void) {
    GCController *first=[GCController controllerWithExtendedGamepad];
    GCController *second=[GCController controllerWithExtendedGamepad];
    [first.extendedGamepad.buttonX setValue:1];
    __block NSArray<GCController *> *inventory=@[first,second];
    NSMutableArray *events=[NSMutableArray new];
    PFGamepad *input=[[PFGamepad alloc] initWithSink:^(int32_t kind,int32_t a,int32_t b) {
        [events addObject:@[@(kind),@(a),@(b)]];
    } controllers:^{return inventory;}];
    assert(([events containsObject:@[@2,@4,@0]]));
    [first.extendedGamepad.buttonX setValue:0];
    [first.extendedGamepad.leftTrigger setValue:1];
    // Snapshot writes do not deliver physical-device events. Invoke Apple's
    // installed callback with those real framework elements to test decoding.
    first.extendedGamepad.valueChangedHandler(first.extendedGamepad,first.extendedGamepad.buttonX);
    [NSRunLoop.currentRunLoop runUntilDate:[NSDate dateWithTimeIntervalSinceNow:.05]];
    assert(([events containsObject:@[@0,@2,@0]]));
    assert(([events containsObject:@[@1,@4,@32767]]));
    NSUInteger before=events.count;
    [second.extendedGamepad.buttonY setValue:1];
    [NSRunLoop.currentRunLoop runUntilDate:[NSDate dateWithTimeIntervalSinceNow:.01]];
    assert(events.count==before);
    inventory=@[second];
    [NSNotificationCenter.defaultCenter postNotificationName:GCControllerDidDisconnectNotification object:first];
    assert(([events containsObject:@[@3,@0,@0]]));
    assert(([events containsObject:@[@2,@8,@0]]));
    [input close];before=events.count;
    [second.extendedGamepad.buttonY setValue:0];
    [NSRunLoop.currentRunLoop runUntilDate:[NSDate dateWithTimeIntervalSinceNow:.01]];
    assert(events.count==before);
    puts("PASS GameController snapshot: connection history, buttons, trigger, device filtering, replacement and cleanup");
}

int main(int argc,const char **argv) {
    @autoreleasepool {
        abiTests(); gamepadTests(); modifierTests(); audioTests(); storageTests(); demoStorageTests(); frameTests();
        if (argc==2) { NSString *originals=[NSString stringWithUTF8String:argv[1]]; compatibilityImportTests(originals); bundledStorageTests(originals); journey(originals); }
        else puts("UNVERIFIED original-backed four-table macOS journey: external originals not supplied");
    }
    return 0;
}
