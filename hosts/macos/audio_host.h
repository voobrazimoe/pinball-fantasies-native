#ifndef PF_MAC_AUDIO_HOST_H
#define PF_MAC_AUDIO_HOST_H
#import <AudioToolbox/AudioToolbox.h>
#import <CoreAudio/CoreAudio.h>
#import "host_logic.h"
typedef struct {
    PFRing ring;
    AudioUnit unit;
    OSStatus error;
    bool running;
    AudioObjectPropertyListenerBlock listener;
} PFAudio;
void pf_audio_init(PFAudio *);
bool pf_audio_open(PFAudio *);
void pf_audio_play(PFAudio *);
void pf_audio_pause(PFAudio *);
void pf_audio_close(PFAudio *);
void pf_audio_enqueue(void *,const uint8_t *,uint32_t);
void pf_audio_watch_device(PFAudio *,dispatch_block_t);
void pf_audio_unwatch_device(PFAudio *);
#endif
