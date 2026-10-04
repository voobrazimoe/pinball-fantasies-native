#import "frame_view.h"
#import "storage.h"
#import "audio_host.h"
#include <assert.h>
#include <stdio.h>
static BOOL acceptTestFiles(NSString *data,NSError **error) {
    (void)error;
    for (NSString *name in pf_required_assets()) {
        NSData *bytes=[NSData dataWithContentsOfFile:[data stringByAppendingPathComponent:name]];
        if (bytes.length!=4) return NO;
    }
    return YES;
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
    [view drawRect:view.bounds]; [NSGraphicsContext restoreGraphicsState];
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
    puts("PASS native bitmap copy, aspect, black bars, exact nearest-neighbour colours");
}
static void abiTests(void) {
    assert(pf_engine_abi_version()==PF_ABI_VERSION);
    char error[256]={0}; assert(!pf_engine_create("/absent-originals","/tmp",0,error,sizeof(error)) && error[0]);
    assert(pf_engine_destroy(0)==PF_INVALID);
    assert(pf_engine_suspend(0)==PF_INVALID); assert(pf_engine_resume(0,0)==PF_INVALID);
    assert(pf_engine_set_action(0,PF_LEFT,1)==PF_INVALID); assert(pf_engine_key(0,57)==PF_INVALID);
    assert(pf_engine_release(0)==PF_INVALID); assert(pf_engine_plunger_delta(0,8)==PF_INVALID);
    assert(pf_engine_plunger_fire(0)==PF_INVALID); assert(pf_engine_advance(0,0,NULL,NULL)==PF_INVALID);
    uint8_t *pixels; int32_t w,h,s; uint64_t ticks; uint32_t mode,table,flags;
    assert(pf_engine_frame(0,&pixels,&w,&h,&s)==PF_INVALID);
    assert(pf_engine_state(0,&ticks,&mode,&table,&flags)==PF_INVALID);
    puts("PASS real Apple-linked C archive: ABI version, load errors, all invalid-handle exports");
}
static void pcm(void *context,const uint8_t *samples,uint32_t bytes) {
    PFRing *ring=context; assert(bytes%4==0);
    pf_ring_write(ring,samples,bytes/4);
    int16_t drain[4096]; while (pf_ring_available(ring)) pf_ring_read(ring,drain,2048);
}
static void journey(NSString *data) {
    for (unsigned table=1;table<=4;table++) for (unsigned scroll=0;scroll<4;scroll++) {
        NSString *state=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
        NSFileManager *fm=NSFileManager.defaultManager;
        assert([fm createDirectoryAtPath:state withIntermediateDirectories:NO attributes:nil error:NULL]);
        uint8_t cfg[]={'P','F','N','C',1,5,0,1,(uint8_t)scroll,0,1};
        assert([[NSData dataWithBytes:cfg length:sizeof(cfg)] writeToFile:[state stringByAppendingPathComponent:@"PINBALL.CFG"] atomically:YES]);
        char error[1024]; uint64_t engine=pf_engine_create((char *)data.fileSystemRepresentation,(char *)state.fileSystemRepresentation,0,error,sizeof(error));
        if (!engine) { fprintf(stderr,"%s\n",error); abort(); }
        PFRing *ring=calloc(1,sizeof(*ring)); pf_ring_init(ring);
        PFFrameView *view=[[PFFrameView alloc] initWithFrame:NSMakeRect(0,0,640,480)];
        int64_t now=0;
        for (unsigned step=0;step<180;step++,now+=16666667) {
            if (step==1) assert(pf_engine_key(engine,pf_make_code(49))==PF_OK);
            const uint16_t nativeF[4]={122,120,99,118};
            if (step==2) assert(pf_engine_key(engine,pf_make_code(nativeF[table-1]))==PF_OK);
            if (step==3) assert(pf_engine_key(engine,pf_make_code(122))==PF_OK);
            if (step==20) assert(pf_engine_set_action(engine,PF_SPRING,1)==PF_OK);
            if (step==40) assert(pf_engine_set_action(engine,PF_SPRING,0)==PF_OK);
            if (step==80) {
                assert(pf_engine_suspend(engine)==PF_OK); now+=3600000000000LL;
                assert(pf_engine_resume(engine,now)==PF_OK);
                assert(pf_engine_key(engine,pf_make_code(35))==PF_OK);
            }
            assert(pf_engine_advance(engine,now,pcm,ring)==PF_OK);
            uint8_t *pixels; int32_t w,h,s;
            assert(pf_engine_frame(engine,&pixels,&w,&h,&s)==PF_OK);
            assert([view acceptPixels:pixels width:w height:h stride:s]);
        }
        uint64_t ticks; uint32_t mode,selected,flags;
        assert(pf_engine_state(engine,&ticks,&mode,&selected,&flags)==PF_OK);
        assert(ticks>150 && selected==table && !(flags&1));
        assert(pf_engine_destroy(engine)==PF_OK); free(ring);
        assert([fm fileExistsAtPath:[state stringByAppendingPathComponent:@"PINBALL.CFG"]]);
        assert([fm removeItemAtPath:state error:NULL]);
        printf("PASS Apple ABI host journey table=%u scroll=%u ticks=%llu mode=%u\n",table,scroll,(unsigned long long)ticks,mode);
    }
}
int main(int argc,const char **argv) {
    @autoreleasepool {
        abiTests(); storageTests(); frameTests();
        if (argc==2) journey([NSString stringWithUTF8String:argv[1]]);
        else puts("UNVERIFIED original-backed four-table macOS journey: external originals not supplied");
    }
    return 0;
}
