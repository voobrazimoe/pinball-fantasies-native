#include "pcm_ring.h"
#include <cassert>
#include <thread>
#include <vector>
#include <cstdlib>
#include <new>
// Fail any allocation on a marked callback thread. Tests call the exact render
// path, not a duplicate consumer implementation.
thread_local bool realtime=false;
void* operator new(size_t n) { assert(!realtime); if (auto p=std::malloc(n)) return p; throw std::bad_alloc(); }
void operator delete(void* p) noexcept { std::free(p); }
void operator delete(void* p, size_t) noexcept { std::free(p); }
int main() {
    a4::PcmRing<4> ring;
    const int16_t input[]={1,-1,2,-2,3,-3,4,-4,5,-5,6,-6};
    int16_t out[12];
    auto bytes=reinterpret_cast<const uint8_t*>(input);
    assert(ring.push(bytes,4,1)==4 && ring.depth()==4);
    assert(ring.push(bytes+16,2,1)==0 && ring.dropped==2 && ring.overflows==1);
    assert(ring.pull(out,2,1,true)==2 && out[0]==1 && out[3]==-2);
    assert(ring.push(bytes+16,2,1)==2); // Wrap around and preserve stereo order.
    assert(ring.pull(out,6,1,true)==4);
    for(int i=0;i<4;i++) { assert(out[i*2]==i+3); assert(out[i*2+1]==-(i+3)); }
    for(int i=8;i<12;i++) assert(out[i]==0);
    assert(ring.depth()==0 && ring.missing==2 && ring.underruns==1);
    assert(ring.push(bytes,6,1)==4 && ring.dropped==4); // Accept prefix, drop tail.
    ring.discardBefore(2); assert(ring.depth()==0 && ring.invalidated==4);
    assert(ring.push(bytes,1,3)==1); ring.discardBefore(3); assert(ring.depth()==1);
    assert(ring.pull(out,1,3,true)==1 && out[0]==1 && ring.depth()==0);
    assert(ring.pull(out,1,3,true)==0 && out[0]==0 && out[1]==0);
    assert(ring.push(bytes,1,5)==1);
    assert(ring.pull(out,1,3,true)==0 && out[0]==0 && ring.depth()==1);
    ring.discardBefore(4); assert(ring.depth()==1);
    assert(ring.pull(out,1,5,true)==1 && out[0]==1);

    a4::PcmBuffer buffer;
    buffer.setActive(true);
    a4::PcmBuffer::sink(&buffer,bytes,24);
    buffer.setActive(false); buffer.setActive(true);
    const int16_t fresh[]={77,-77};
    a4::PcmBuffer::sink(&buffer,reinterpret_cast<const uint8_t*>(fresh),4);
    realtime=true; buffer.render(out,2); realtime=false;
    assert(out[0]==77 && out[1]==-77 && out[2]==0 && out[3]==0);
    a4::PcmBuffer::sink(&buffer,bytes,24); buffer.reset();
    realtime=true; buffer.render(out,6); realtime=false;
    for(auto x:out) assert(x==0);
    assert(buffer.ring.depth()==0);

    a4::PcmRing<31> concurrent;
    constexpr int count=200000;
    std::thread producer([&] {
        for(int i=0;i<count;) {
            int16_t stereo[]={static_cast<int16_t>(i%30000),static_cast<int16_t>(-(i%30000))};
            if (concurrent.push(reinterpret_cast<uint8_t*>(stereo),1,1)) ++i;
            else std::this_thread::yield();
        }
    });
    std::thread consumer([&] {
        realtime=true;
        for(int i=0;i<count;) {
            int16_t stereo[2];
            if (concurrent.pull(stereo,1,1,true)) {
                assert(stereo[0]==i%30000 && stereo[1]==-(i%30000)); ++i;
            }
        }
        realtime=false;
    });
    producer.join(); consumer.join();
    assert(concurrent.depth()==0 && concurrent.consumed==count && concurrent.highWater<=31);
}
