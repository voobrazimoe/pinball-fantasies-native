#pragma once
#include <algorithm>
#include <array>
#include <atomic>
#include <cstdint>
#include <cstring>

namespace a4 {
static_assert(std::atomic<uint64_t>::is_always_lock_free, "A4 requires lock-free 64-bit atomics");
// SPSC: engine/session mutex serializes the producer. Only the device callback
// consumes; control may consume ONLY after closing/joining that stream.
// 4096 stereo frames = 85.33 ms. Enough for several 60/71 Hz source batches and
// display wake jitter, without the desktop ring's one-second latency ceiling.
template<size_t Capacity = 4096> class PcmRing {
    static_assert(Capacity>0, "PCM ring must have storage");
    struct Frame { uint64_t epoch; int16_t samples[2]; };
    std::array<Frame, Capacity> data{};
    alignas(64) std::atomic<uint64_t> read{0};
    alignas(64) std::atomic<uint64_t> write{0};
public:
    std::atomic<uint64_t> produced{0}, consumed{0}, underruns{0}, missing{0};
    std::atomic<uint64_t> overflows{0}, dropped{0}, highWater{0}, invalidated{0};
    size_t push(const uint8_t* pcm, size_t frames, uint64_t epoch) noexcept {
        const auto w = write.load(std::memory_order_relaxed);
        const auto r = read.load(std::memory_order_acquire);
        const auto n = std::min(frames, Capacity - static_cast<size_t>(w-r));
        for (size_t i=0; i<n; ++i) {
            auto& f = data[(w+i)%Capacity]; f.epoch=epoch;
            std::memcpy(f.samples, pcm+i*4, 4);
        }
        write.store(w+n, std::memory_order_release);
        produced.fetch_add(frames, std::memory_order_relaxed);
        if (n<frames) { overflows.fetch_add(1, std::memory_order_relaxed);
            dropped.fetch_add(frames-n, std::memory_order_relaxed); }
        // Producer alone writes highWater; no CAS/retry loop on realtime paths.
        const auto depth = w+n-r;
        if (depth>highWater.load(std::memory_order_relaxed)) highWater.store(depth, std::memory_order_relaxed);
        return n;
    }
    // Epoch invalidation avoids resetting live SPSC cursors or overwriting a
    // slot being read. Stale frames are discarded, never played after resume.
    size_t pull(void* output, size_t frames, uint64_t epoch, bool active) noexcept {
        auto r=read.load(std::memory_order_relaxed);
        const auto w=write.load(std::memory_order_acquire);
        auto* dst=static_cast<uint8_t*>(output);
        size_t n=0, stale=0;
        while (r<w && n<frames) {
            const auto& f=data[r%Capacity];
            // A rapid resume may publish a newer generation after this callback
            // sampled state. Leave those fresh frames for the next callback.
            if (f.epoch>epoch) break;
            if (active && f.epoch==epoch) { std::memcpy(dst+n*4, f.samples, 4); ++n; }
            else ++stale;
            ++r;
        }
        read.store(r, std::memory_order_release);
        std::memset(dst+n*4, 0, (frames-n)*4);
        consumed.fetch_add(n, std::memory_order_relaxed);
        invalidated.fetch_add(stale, std::memory_order_relaxed);
        if (active && n<frames) { underruns.fetch_add(1, std::memory_order_relaxed);
            missing.fetch_add(frames-n, std::memory_order_relaxed); }
        return n;
    }
    // Consumer-only, e.g. after stream close. Keep any fresh epoch's frames.
    void discardBefore(uint64_t epoch) noexcept {
        auto r=read.load(std::memory_order_relaxed);
        const auto start=r, w=write.load(std::memory_order_acquire);
        while (r<w && data[r%Capacity].epoch<epoch) ++r;
        read.store(r, std::memory_order_release);
        invalidated.fetch_add(r-start, std::memory_order_relaxed);
    }
    size_t depth() const noexcept {
        // Diagnostic snapshot; consumer can move between these loads.
        const auto r=read.load(std::memory_order_acquire), w=write.load(std::memory_order_acquire);
        return std::min(Capacity, static_cast<size_t>(w-r));
    }
};
class PcmBuffer {
    // Low bit = active, upper bits = generation. Requests serialized with PCM
    // production by the A3 mutex. No stream operation or wait on that path.
    std::atomic<uint64_t> state{0};
public:
    PcmRing<> ring;
    std::atomic<bool> ready{true};
    void setActive(bool active) noexcept {
        auto old=state.load(std::memory_order_relaxed);
        do {
            if (bool(old&1)==active) return;
        } while (!state.compare_exchange_weak(old, ((old&~uint64_t{1})+2) | uint64_t(active), std::memory_order_release));
    }
    void refresh() noexcept { state.fetch_add(2, std::memory_order_acq_rel); }
    void reset() noexcept { setActive(false); refresh(); }
    uint64_t current() const noexcept { return state.load(std::memory_order_acquire); }
    static void sink(void* context, const uint8_t* pcm, uint32_t bytes) noexcept {
        auto& self=*static_cast<PcmBuffer*>(context);
        const auto epoch=self.current();
        if ((epoch&1) && self.ready.load(std::memory_order_acquire) && pcm && bytes%4==0) self.ring.push(pcm, bytes/4, epoch);
    }
    void render(void* output, size_t frames) noexcept {
        const auto epoch=current();
        ring.pull(output, frames, epoch, epoch&1);
        // A pause racing this callback silences the whole in-flight block.
        if (current()!=epoch) std::memset(output, 0, frames*4);
    }
};
}
