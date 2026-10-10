// Exercise the actual JNI/session serialization code with asset-free ABI mocks.
#include "../app/src/main/cpp/engine_host.cpp"
#include <atomic>
#include <cassert>
#include <thread>
a4::PcmBuffer& androidAudioBuffer() { static a4::PcmBuffer buffer; return buffer; }
namespace {
std::atomic<int> entered{0}, calls{0}, suspends{0}, resumes{0}, destroys{0};
bool suspended=false;
uint64_t sourceTicks=12;
struct Call {
    Call() { assert(entered.fetch_add(1)==0); ++calls; std::this_thread::yield(); }
    ~Call() { assert(entered.fetch_sub(1)==1); }
};
uint8_t framePixels[] = {1,2,3,4,99,99,99,99,5,6,7,8,99,99,99,99};
}
extern "C" {
int32_t pf_engine_abi_version() { Call c; return 1; }
uint64_t pf_engine_create(char*,char*,int64_t,char*,uint32_t) { Call c; return 42; }
int32_t pf_engine_destroy(uint64_t) { Call c; ++destroys; return PF_OK; }
int32_t pf_engine_suspend(uint64_t) { Call c; suspended=true; ++suspends; return PF_OK; }
int32_t pf_engine_resume(uint64_t,int64_t) { Call c; suspended=false; ++resumes; return PF_OK; }
int32_t pf_engine_set_action(uint64_t,uint32_t,int32_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_key(uint64_t,uint8_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_release(uint64_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_plunger_delta(uint64_t,int32_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_plunger_target(uint64_t,int32_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_gamepad(uint64_t,int32_t,int32_t,int32_t) { Call c; return PF_OK; }
int32_t pf_engine_state(uint64_t,uint64_t* tick,uint32_t* mode,uint32_t* table,uint32_t* flags) { Call c; *tick=sourceTicks; *mode=PF_MODE_PLAYING; *table=2; *flags=4; return PF_OK; }
int32_t pf_engine_plunger_fire(uint64_t) { Call c; assert(!suspended); return PF_OK; }
int32_t pf_engine_set_presentation(uint64_t,int32_t) { Call c; return PF_OK; }
int32_t pf_engine_advance(uint64_t,int64_t ns,pf_pcm_sink sink,void* context) {
    Call c; assert(ns>0 && !suspended && sink==a4::PcmBuffer::sink);
    sourceTicks+=2;
    uint8_t borrowed[]={12,0,244,255};
    sink(context,borrowed,sizeof(borrowed));
    std::memset(borrowed,99,sizeof(borrowed)); return PF_OK;
}
int32_t pf_engine_frame(uint64_t,uint8_t** p,int32_t* w,int32_t* h,int32_t* stride) {
    Call c; *p=framePixels; *w=1; *h=2; *stride=8; return PF_OK;
}
}
int main() {
    androidPublishViewport(481,1,1439,1077,2400,1080,320,240);
    assert((renderedViewport == std::array<jint,8>{481,2,1439,1077,2400,1080,320,240}));

    const auto token=JNI_METHOD(nativeOpen)(nullptr,nullptr);
    // Simulate the already-validated A2 bootstrap; no commercial loader fixture.
    { std::lock_guard<std::mutex> guard(lock); persistent=42; }
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true);
    assert(!(androidAudioBuffer().current()&1));
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,true,false);
    assert(pfTestInfoLogs == 0);
    JNI_METHOD(nativeDiagnostics)(nullptr,nullptr,true);
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true);
    assert(pfTestInfoLogs == 1);
    JNI_METHOD(nativeDiagnostics)(nullptr,nullptr,false);
    assert(JNI_METHOD(nativeState)(nullptr,nullptr,token)==(PF_MODE_PLAYING | (2<<8) | (4<<16)));
    assert(JNI_METHOD(nativeState)(nullptr,nullptr,token-1)==-1);
    JNI_METHOD(nativeInput)(nullptr,nullptr,token,5,32,0);
    std::vector<uint8_t> pixels; int w=0,h=0;
    assert(androidEngineFrame(true,pixels,w,h) && w==1 && h==2 && pixels[4]==5);
    assert(pixels.data()!=framePixels);
    int64_t advanced=-1;
    assert(androidEngineFrame(true,pixels,w,h,&advanced) && advanced==2);
    int16_t audio[2]; const auto abiCalls=calls.load();
    androidAudioBuffer().render(audio,1);
    assert(audio[0]==12 && audio[1]==-12 && calls==abiCalls);
    assert(androidEngineFrame(true,pixels,w,h));
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,false);
    const auto before=calls.load();
    assert(!androidEngineFrame(false,pixels,w,h,&advanced) && advanced==-1);
    JNI_METHOD(nativeInput)(nullptr,nullptr,token,0,0,1);
    assert(calls==before && suspends==1);
    JNI_METHOD(nativeInput)(nullptr,nullptr,token,6,2,0);
    assert(calls==before+1); // Controller releases maintain physical history while inactive.
    JNI_METHOD(nativeInput)(nullptr,nullptr,token-1,6,2,0);
    assert(calls==before+1); // Stale Activity cannot mutate that history.
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true);
    androidAudioBuffer().render(audio,1); assert(audio[0]==0 && audio[1]==0);
    assert(androidEngineFrame(true,pixels,w,h));
    androidAudioBuffer().render(audio,1); assert(audio[0]==12 && audio[1]==-12);
    const auto sourceSuspends=suspends.load();
    assert(androidEngineFrame(true,pixels,w,h));
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,false,false);
    androidAudioBuffer().render(audio,1); assert(audio[0]==0);
    assert(androidEngineFrame(true,pixels,w,h) && suspends==sourceSuspends);
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,true,false);
    androidAudioBuffer().render(audio,1); assert(audio[0]==0);
    assert(androidEngineFrame(true,pixels,w,h));
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,false,true);
    androidAudioBuffer().render(audio,1); assert(audio[0]==0 && suspends==sourceSuspends);
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,false,true);
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,false,false);
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,true,false); // gain before resume
    assert(!(androidAudioBuffer().current()&1));
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true);
    assert(androidAudioBuffer().current()&1);
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,false,true);
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,false,false);
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true); // resume before gain
    assert(!(androidAudioBuffer().current()&1));
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,true,false);
    assert(androidAudioBuffer().current()&1);
    std::thread focusEvents([&] { for(int i=0;i<1000;i++)
        JNI_METHOD(nativeAudio)(nullptr,nullptr,token,i%2,i%3==0); });
    std::thread render([&] { std::vector<uint8_t> out; int fw=0,fh=0;
        for(int i=0;i<1000;i++) androidEngineFrame(i%2,out,fw,fh); });
    std::thread input([&] { for(int i=0;i<1000;i++)
        JNI_METHOD(nativeInput)(nullptr,nullptr,token,i%6,0,i%2); });
    std::thread lifecycle([&] { for(int i=0;i<1000;i++)
        JNI_METHOD(nativeActive)(nullptr,nullptr,token,i%2,true); });
    render.join(); input.join(); lifecycle.join(); focusEvents.join();
    JNI_METHOD(nativeDetach)(nullptr,nullptr,token);
    assert(persistent==42 && destroys==0 && retained);
    const auto recreated=JNI_METHOD(nativeOpen)(nullptr,nullptr);
    assert(persistent==42 && destroys==0 && recreated!=token);
    JNI_METHOD(nativeActive)(nullptr,nullptr,recreated,true,true);
    JNI_METHOD(nativeClose)(nullptr,nullptr,recreated);
    assert(destroys==1 && persistent==0);
    androidAudioBuffer().render(audio,1); assert(audio[0]==0 && audio[1]==0);
    const auto closedCalls=calls.load();
    JNI_METHOD(nativeActive)(nullptr,nullptr,token,true,true);
    JNI_METHOD(nativeInput)(nullptr,nullptr,token,1,28,0);
    assert(!androidEngineFrame(true,pixels,w,h) && calls==closedCalls);
    const auto next=JNI_METHOD(nativeOpen)(nullptr,nullptr);
    JNI_METHOD(nativeClose)(nullptr,nullptr,token); assert(opened && next!=token);
    JNI_METHOD(nativeAudio)(nullptr,nullptr,token,true,true);
    assert(!(androidAudioBuffer().current()&1));
    JNI_METHOD(nativeClose)(nullptr,nullptr,next);
}
