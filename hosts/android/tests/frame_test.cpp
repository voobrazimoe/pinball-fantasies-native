#include "a3_host.h"
#include <cassert>
int main() {
    assert(!a3::frameValid(0,609,1280));
    assert(!a3::frameValid(320,609,1279));
    assert(!a3::frameValid(4097,1,16388));
    assert(a3::frameValid(320,609,1296));
    uint8_t source[] = {1,2,3,4,99,99,99,99,5,6,7,8,99,99,99,99};
    std::vector<uint8_t> out;
    assert(a3::copyFrame(source,1,2,8,out));
    assert((out == std::vector<uint8_t>{1,2,3,4,5,6,7,8}));
    auto data = out.data(); assert(a3::copyFrame(source,1,2,8,out)); assert(out.data()==data);
    assert(!a3::copyFrame(nullptr,1,2,8,out));
    auto v=a3::letterbox(1080,1920,320,609);
    assert(v.h==1920 && v.w==1009 && v.x==35 && v.y==0);
    v=a3::letterbox(1920,1080,320,609); assert(v.h==1080 && v.w==567 && v.x==676);
    v=a3::letterbox(1920,1080,640,240); assert(v.h==1080 && v.w==1440 && v.x==240);
    assert(a3::portrait(1080,1920) && !a3::portrait(1920,1080) && a3::portrait(320,320));
}
