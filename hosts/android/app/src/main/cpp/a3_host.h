#pragma once
#include <algorithm>
#include <cmath>
#include <cstdint>
#include <cstring>
#include <vector>
namespace a3 {
inline bool frameValid(int w, int h, int stride) {
    return w > 0 && h > 0 && w <= 4096 && h <= 4096 && stride >= w*4 && stride <= 4096*4;
}
inline bool portrait(int w, int h) { return h >= w; }
struct Viewport { int x, y, w, h; };
inline Viewport letterbox(int sw, int sh, int w, int h) {
    // Original 640x240 frontend pixels have 2:1 vertical pixel aspect.
    const double logicalHeight = (w == 640 && h == 240) ? 480 : h;
    const double scale = std::min(double(sw)/w, double(sh)/logicalHeight);
    int vw = std::max(1, int(std::lround(w*scale)));
    int vh = std::max(1, int(std::lround(logicalHeight*scale)));
    return {(sw-vw)/2, (sh-vh)/2, vw, vh};
}
inline bool copyFrame(const uint8_t* p, int w, int h, int stride, std::vector<uint8_t>& out) {
    if (!p || !frameValid(w,h,stride)) return false;
    out.resize(size_t(w)*h*4);
    for (int y=0; y<h; ++y) std::memcpy(out.data()+size_t(y)*w*4,p+size_t(y)*stride,size_t(w)*4);
    return true;
}
}
// All engine operations share engine_host.cpp's mutex. Returned pixels belong
// to the host vector, never to the engine, after this function returns.
bool androidEngineFrame(bool portrait, std::vector<uint8_t>& pixels, int& width, int& height, int64_t* advancedTicks = nullptr);

// Renderer publishes the exact GL viewport, with surface dimensions.
void androidPublishViewport(int x,int y,int w,int h,int sw,int sh,int fw,int fh);
