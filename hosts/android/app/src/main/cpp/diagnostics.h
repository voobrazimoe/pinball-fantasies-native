#pragma once
#include <android/log.h>
#include <atomic>
#include <cstdlib>
#include <cstring>

// Shared by JNI and the render thread; configured before GameActivity startup.
inline std::atomic<bool> pfDiagnosticsRequested{false};
inline bool pfDiagnosticsEnabled() {
    const char* value = std::getenv("PF_DIAGNOSTICS");
    return pfDiagnosticsRequested.load() || (value && std::strcmp(value, "1") == 0);
}
#define PF_LOGI(...) do { if (pfDiagnosticsEnabled()) \
    (void)__android_log_print(ANDROID_LOG_INFO, "PinballFantasies", __VA_ARGS__); } while (0)
