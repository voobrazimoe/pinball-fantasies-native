#import "audio_host.h"
#include <string.h>
#include <time.h>
static const AudioObjectPropertyAddress route={kAudioHardwarePropertyDefaultOutputDevice,
    kAudioObjectPropertyScopeGlobal,kAudioObjectPropertyElementMain};
static OSStatus render(void *ctx,AudioUnitRenderActionFlags *flags,const AudioTimeStamp *time,
                       UInt32 bus,UInt32 frames,AudioBufferList *buffers) {
    (void)flags; (void)time; (void)bus;
    PFAudio *a=ctx;
    /* Real-time path: no engine calls, allocation, logging, locks or waiting. */
    if (buffers->mNumberBuffers==1 && buffers->mBuffers[0].mData &&
        buffers->mBuffers[0].mDataByteSize>=frames*4) {
        pf_ring_read(&a->ring,buffers->mBuffers[0].mData,frames);
        buffers->mBuffers[0].mDataByteSize=frames*4;
    } else {
        for (UInt32 i=0;i<buffers->mNumberBuffers;i++)
            if (buffers->mBuffers[i].mData) memset(buffers->mBuffers[i].mData,0,buffers->mBuffers[i].mDataByteSize);
    }
    return noErr;
}
void pf_audio_init(PFAudio *a) {
    a->unit=NULL; a->error=0; a->running=false; a->listener=nil; pf_ring_init(&a->ring);
}
bool pf_audio_open(PFAudio *a) {
    AudioComponentDescription description={kAudioUnitType_Output,kAudioUnitSubType_DefaultOutput,
        kAudioUnitManufacturer_Apple,0,0};
    AudioComponent component=AudioComponentFindNext(NULL,&description);
    if (!component) { a->error=kAudio_ParamError; return false; }
    a->error=AudioComponentInstanceNew(component,&a->unit);
    if (a->error) return false;
    /* DefaultOutput converts to the hardware format/rate if necessary. Its input
       always remains shared engine 48kHz S16 stereo; device cadence owns no ticks. */
    AudioStreamBasicDescription format={0};
    format.mSampleRate=48000; format.mFormatID=kAudioFormatLinearPCM;
    format.mFormatFlags=kAudioFormatFlagIsSignedInteger|kAudioFormatFlagIsPacked;
    format.mBytesPerPacket=4; format.mFramesPerPacket=1; format.mBytesPerFrame=4;
    format.mChannelsPerFrame=2; format.mBitsPerChannel=16;
    a->error=AudioUnitSetProperty(a->unit,kAudioUnitProperty_StreamFormat,kAudioUnitScope_Input,0,&format,sizeof(format));
    if (!a->error) {
        AURenderCallbackStruct callback={render,a};
        a->error=AudioUnitSetProperty(a->unit,kAudioUnitProperty_SetRenderCallback,kAudioUnitScope_Input,0,&callback,sizeof(callback));
    }
    if (!a->error) a->error=AudioUnitInitialize(a->unit);
    if (a->error) { pf_audio_close(a); return false; }
    return true;
}
void pf_audio_play(PFAudio *a) {
    if (a->unit && !a->running) {
        a->error=AudioOutputUnitStart(a->unit); a->running=a->error==noErr;
    }
}
void pf_audio_pause(PFAudio *a) {
    if (a->unit && a->running) AudioOutputUnitStop(a->unit);
    a->running=false;
    /* Stop joins the render path; both indices may now be reset safely. */
    pf_ring_reset(&a->ring);
}
void pf_audio_close(PFAudio *a) {
    pf_audio_pause(a);
    if (a->unit) { AudioUnitUninitialize(a->unit); AudioComponentInstanceDispose(a->unit); a->unit=NULL; }
}
void pf_audio_enqueue(void *ctx,const uint8_t *pcm,uint32_t bytes) {
    PFAudio *a=ctx; size_t remaining=bytes/4;
    /* Preserve source chunks while the device drains. Never hold a Go pointer
       after this synchronous callback. A stalled/missing device cannot hang UI. */
    unsigned stalled=0;
    while (remaining) {
        size_t free=PF_RING_FRAMES-pf_ring_available(&a->ring);
        if (free) {
            size_t n=remaining<free?remaining:free;
            pf_ring_write(&a->ring,pcm,n); pcm+=n*4; remaining-=n; stalled=0;
        } else if (a->running && stalled++<100) {
            struct timespec wait={0,1000000}; nanosleep(&wait,NULL);
        } else {
            atomic_fetch_add(&a->ring.dropped,remaining); break;
        }
    }
}
void pf_audio_watch_device(PFAudio *a,dispatch_block_t changed) {
    a->listener=^(UInt32 count,const AudioObjectPropertyAddress *addresses) {
        (void)count; (void)addresses; changed();
    };
    OSStatus result=AudioObjectAddPropertyListenerBlock(kAudioObjectSystemObject,&route,dispatch_get_main_queue(),a->listener);
    if (result) a->listener=nil;
}
void pf_audio_unwatch_device(PFAudio *a) {
    if (a->listener) AudioObjectRemovePropertyListenerBlock(kAudioObjectSystemObject,&route,dispatch_get_main_queue(),a->listener);
    a->listener=nil;
}
