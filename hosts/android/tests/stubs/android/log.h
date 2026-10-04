#pragma once
#define ANDROID_LOG_INFO 4
#define ANDROID_LOG_ERROR 6
inline int pfTestInfoLogs = 0;
inline int __android_log_print(int priority, const char*, const char*, ...) {
    if (priority == ANDROID_LOG_INFO) ++pfTestInfoLogs;
    return 0;
}
