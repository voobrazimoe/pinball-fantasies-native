#pragma once
#include "diagnostics.h"
#include <algorithm>
#include <cstdint>
#include <ctime>

namespace presentation {
inline int64_t monotonicNs() {
    timespec t{}; clock_gettime(CLOCK_MONOTONIC, &t);
    return int64_t(t.tv_sec)*1000000000LL+t.tv_nsec;
}
struct Samples {
    uint64_t count=0;
    int64_t sum=0, min=INT64_MAX, max=0;
    void add(int64_t ns) { ++count; sum+=ns; min=std::min(min,ns); max=std::max(max,ns); }
    double averageMs() const { return count ? double(sum)/count/1e6 : 0; }
    double minMs() const { return count ? double(min)/1e6 : 0; }
    double maxMs() const { return double(max)/1e6; }
};
// Fixed storage; one aggregate every three seconds, no per-frame logging or
// GPU synchronization. GL timings measure CPU submission/backpressure only.
struct Diagnostics {
    int64_t start=0, lastDraw=0, lastSwap=0, lastVsync=0;
    uint64_t attempts=0, swaps=0, ticks=0, knownTicks=0, zero=0, one=0, many=0;
    Samples drawIntervals, swapIntervals, vsyncIntervals, engine, upload, draw, swap;
    void begin(int64_t now, int64_t vsync) {
        if (!start) start=now;
        ++attempts;
        if (lastDraw) drawIntervals.add(now-lastDraw);
        if (lastVsync && vsync>lastVsync) vsyncIntervals.add(vsync-lastVsync);
        lastDraw=now; lastVsync=vsync;
    }
    void finish(int64_t now, bool success, int64_t sourceTicks) {
        if (success) {
            ++swaps;
            if (lastSwap) swapIntervals.add(now-lastSwap);
            lastSwap=now;
            if (sourceTicks>=0) {
                ++knownTicks; ticks+=uint64_t(sourceTicks);
                if (!sourceTicks) ++zero; else if (sourceTicks==1) ++one; else ++many;
            }
        }
        if (now-start<3000000000LL) return;
        const double seconds=double(now-start)/1e9;
        PF_LOGI("A6_PRESENT wall_s=%.3f callbacks=%llu attempts=%llu swaps=%llu callback_hz=%.2f present_fps=%.2f ticks=%llu ticks_per_present=%.3f tick_samples=%llu tick_0=%llu tick_1=%llu tick_2plus=%llu",
            seconds, (unsigned long long)attempts, (unsigned long long)attempts,
            (unsigned long long)swaps, attempts/seconds, swaps/seconds,
            (unsigned long long)ticks, knownTicks ? double(ticks)/knownTicks : 0,
            (unsigned long long)knownTicks, (unsigned long long)zero,
            (unsigned long long)one, (unsigned long long)many);
        PF_LOGI("A6_PRESENT_INTERVAL_MS draw=%.3f/%.3f/%.3f swap=%.3f/%.3f/%.3f vsync_observed=%.3f/%.3f/%.3f (min/avg/max; missed callbacks may multiply vsync period)",
            drawIntervals.minMs(),drawIntervals.averageMs(),drawIntervals.maxMs(),
            swapIntervals.minMs(),swapIntervals.averageMs(),swapIntervals.maxMs(),
            vsyncIntervals.minMs(),vsyncIntervals.averageMs(),vsyncIntervals.maxMs());
        PF_LOGI("A6_PRESENT_COST_MS engine=%.3f/%.3f upload=%.3f/%.3f draw=%.3f/%.3f swap=%.3f/%.3f (avg/max CPU; swap_interval=1)",
            engine.averageMs(),engine.maxMs(),upload.averageMs(),upload.maxMs(),
            draw.averageMs(),draw.maxMs(),swap.averageMs(),swap.maxMs());
        *this={};
    }
};
}
