#include "abi.h"
#include <android/log.h>
#include <jni.h>
#include <time.h>
#include <array>
#include <mutex>
#include <string>

namespace {
std::mutex lock;
uint64_t persistent = 0;
jlong generation = 0;
bool opened = false;
int64_t now() {
    timespec value{};
    clock_gettime(CLOCK_MONOTONIC, &value);
    return static_cast<int64_t>(value.tv_sec) * 1000000000LL + value.tv_nsec;
}
std::string path(JNIEnv* env, jbyteArray bytes) {
    if (bytes == nullptr) return {};
    std::string result(static_cast<size_t>(env->GetArrayLength(bytes)), '\0');
    env->GetByteArrayRegion(bytes, 0, static_cast<jsize>(result.size()),
                           reinterpret_cast<jbyte*>(result.data()));
    // JNI receives standard UTF-8 bytes, never modified UTF-8 or provider URIs.
    if (result.empty() || result[0] != '/' || result.find('\0') != std::string::npos) return {};
    return result;
}
jstring error(JNIEnv* env, const char* message) {
    __android_log_print(ANDROID_LOG_INFO, "PinballFantasies", "A2_ENGINE_REJECTED %s", message);
    // Shared loader errors are bounded; byte[] -> String below preserves standard UTF-8.
    jclass stringClass = env->FindClass("java/lang/String");
    jmethodID constructor = env->GetMethodID(stringClass, "<init>", "([BLjava/lang/String;)V");
    const auto length = static_cast<jsize>(std::char_traits<char>::length(message));
    jbyteArray bytes = env->NewByteArray(length);
    env->SetByteArrayRegion(bytes, 0, length, reinterpret_cast<const jbyte*>(message));
    jstring encoding = env->NewStringUTF("UTF-8");
    return static_cast<jstring>(env->NewObject(stringClass, constructor, bytes, encoding));
}
void stop() {
    if (persistent != 0) {
        pf_engine_destroy(persistent);
        persistent = 0;
    }
}
}
#define JNI_METHOD(name) Java_io_github_voobrazimoe_pinballfantasies_PinballActivity_##name
extern "C" JNIEXPORT jlong JNICALL JNI_METHOD(nativeOpen)(JNIEnv*, jclass) {
    std::lock_guard<std::mutex> guard(lock);
    stop(); opened = true;
    return ++generation;
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeClose)(JNIEnv*, jclass, jlong token) {
    std::lock_guard<std::mutex> guard(lock);
    if (token == generation) { stop(); opened = false; }
}
extern "C" JNIEXPORT jstring JNICALL JNI_METHOD(nativeEngine)(
        JNIEnv* env, jclass, jlong token, jint operation, jbyteArray dataBytes, jbyteArray stateBytes) {
    std::lock_guard<std::mutex> guard(lock);
    if (!opened || token != generation) return error(env, "Activity closed");
    if (operation == 2) { stop(); return nullptr; }
    if (pf_engine_abi_version() != PF_ABI_VERSION) return error(env, "Packaged engine ABI must be 1");
    std::string data = path(env, dataBytes), state = path(env, stateBytes);
    if (data.empty() || state.empty()) return error(env, "Engine requires absolute filesystem UTF-8 paths");
    if (operation == 1 && persistent != 0) return nullptr;
    std::array<char, 1024> message{};
    const uint64_t handle = pf_engine_create(data.data(), state.data(), now(), message.data(), message.size());
    message.back() = '\0';
    if (handle == 0) return error(env, message[0] ? message.data() : "Engine creation failed");
    if (operation == 0) {
        pf_engine_destroy(handle);
        __android_log_print(ANDROID_LOG_INFO, "PinballFantasies", "A2_CANDIDATE_VALIDATED");
    } else {
        persistent = handle;
        __android_log_print(ANDROID_LOG_INFO, "PinballFantasies", "A2_ENGINE_BOOTSTRAPPED ABI=1");
    }
    return nullptr;
}
