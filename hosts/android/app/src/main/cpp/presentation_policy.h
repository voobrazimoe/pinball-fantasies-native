#pragma once
#include <cstdint>

namespace presentation {
// Looper-thread-only policy. A pending callback survives stop/start, but its
// epoch cannot present into a replacement surface. No display-derived clock.
class Policy {
public:
    static constexpr int pollTimeoutMillis() { return -1; }
    void setActive(bool value) {
        if (active_ != value) { active_ = value; ++epoch_; }
    }
    bool arm() {
        if (!active_ || pending_) return false;
        pending_ = true; postedEpoch_ = epoch_; return true;
    }
    bool consume() {
        if (!pending_) return false;
        pending_ = false;
        return active_ && postedEpoch_ == epoch_;
    }
    bool pending() const { return pending_; }
private:
    bool active_ = false, pending_ = false;
    uint64_t epoch_ = 0, postedEpoch_ = 0;
};
}
