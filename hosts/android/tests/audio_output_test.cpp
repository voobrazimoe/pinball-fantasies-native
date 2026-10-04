#include "../app/src/main/cpp/audio_host.cpp"
#include <cassert>
#include <cstdlib>
#include <new>
thread_local bool oboe::realtime=false;
void* operator new(size_t n) { assert(!oboe::realtime); if(auto p=std::malloc(n)) return p; throw std::bad_alloc(); }
void operator delete(void* p) noexcept { std::free(p); }
void operator delete(void* p,size_t) noexcept { std::free(p); }
template<class F> void until(F predicate) {
    const auto end=std::chrono::steady_clock::now()+std::chrono::seconds(3);
    while(!predicate()) { assert(std::chrono::steady_clock::now()<end);
        std::this_thread::sleep_for(std::chrono::milliseconds(2)); }
}
int main() {
    // The production consumer/controller compiles and links with no engine ABI
    // implementation at all. Allocation guards cover data and error callbacks.
    oboe::exclusiveFailures=1;
    a4::AudioOutput output;
    output.buffer.setActive(true);
    until([&] { return output.opens==1; });
    int16_t pcm[]={31,-31};
    a4::PcmBuffer::sink(&output.buffer,reinterpret_cast<uint8_t*>(pcm),4);
    until([&] { return output.callbacks>2 && output.buffer.ring.consumed==1; });
    assert(output.opens==1 && pfTestInfoLogs==0);
    oboe::injectError=true;
    until([&] { return output.opens>=2; });
    assert(output.errors==1 && output.restarts==1);
    const auto initialEpoch=output.buffer.current(), routeOpens=output.opens.load();
    for(int i=0;i<20;i++) output.buffer.refresh();
    assert(output.buffer.current()>initialEpoch);
    until([&] { return output.opens>routeOpens; });
    assert(output.opens<=routeOpens+2); // notifications coalesce during settle
    output.buffer.setActive(false);
    until([&] { return !output.buffer.ready.load(); });
    std::this_thread::sleep_for(std::chrono::milliseconds(30));
    const auto paused=output.callbacks.load();
    std::this_thread::sleep_for(std::chrono::milliseconds(30)); assert(output.callbacks==paused);
    const auto resumeOpens=output.opens.load();
    output.buffer.setActive(true);
    until([&] { return output.opens>resumeOpens; });
    a4::PcmBuffer::sink(&output.buffer,reinterpret_cast<uint8_t*>(pcm),4);
    until([&] { return output.opens>=3 && output.buffer.ring.consumed>=2; });
    output.buffer.reset();
    until([&] { return !output.buffer.ready.load(); });
    // Ordinary open failure retries with bounded backoff and can be cancelled.
    oboe::openFailures=100;
    output.buffer.setActive(true);
    until([&] { return output.errors>=2; });
    std::this_thread::sleep_for(std::chrono::milliseconds(2400));
    const auto failures=output.errors.load();
    for(int i=0;i<10000;i++) a4::PcmBuffer::sink(&output.buffer,reinterpret_cast<uint8_t*>(pcm),4);
    assert(output.buffer.ring.depth()==0);
    std::this_thread::sleep_for(std::chrono::milliseconds(1100));
    assert(output.errors==failures); // finite three-attempt budget
    output.buffer.reset(); oboe::openFailures=0;
    assert(pfTestInfoLogs==0);
    assert(Java_io_github_voobrazimoe_pinballfantasies_ImportInstrumentation_nativeAudioSmoke(nullptr,nullptr));
}
