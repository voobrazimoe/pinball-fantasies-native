#include "abi.h"
#include "a3_host.h"
#include "diagnostics.h"
#include "audio_host.h"
#include <jni.h>
#include <time.h>
#include <array>
#include <mutex>
#include <string>

namespace {
std::mutex lock;
uint64_t persistent = 0;
jlong generation = 0;
bool opened = false, retained = false;
bool resumed = false, focused = false, audioFocus = false;
void syncAudio() { androidAudioBuffer().setActive(opened && persistent && resumed && focused && audioFocus); }
bool active() { return resumed && focused; }
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
    PF_LOGI("A2_ENGINE_REJECTED %s", message);
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
    androidAudioBuffer().reset();
    if (persistent != 0) {
        pf_engine_destroy(persistent);
        persistent = 0;
    }
}
}
#define JNI_METHOD(name) Java_io_github_voobrazimoe_pinballfantasies_PinballActivity_##name
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeDiagnostics)(JNIEnv*, jclass, jboolean enabled) {
    pfDiagnosticsRequested.store(enabled);
}
extern "C" JNIEXPORT jlong JNICALL JNI_METHOD(nativeOpen)(JNIEnv*, jclass) {
    std::lock_guard<std::mutex> guard(lock);
    if (!retained) stop();
    retained = false; opened = true; resumed = focused = audioFocus = false;
    return ++generation;
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeClose)(JNIEnv*, jclass, jlong token) {
    std::lock_guard<std::mutex> guard(lock);
    if (token == generation) { stop(); opened = false; retained = false; }
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
        PF_LOGI("A2_CANDIDATE_VALIDATED");
    } else {
        persistent = handle;
        syncAudio();
        if (!active()) pf_engine_suspend(persistent);
        PF_LOGI("A2_ENGINE_BOOTSTRAPPED ABI=1");
    }
    return nullptr;
}

// UI input, lifecycle, importer and native render calls all take the same lock.
// No ABI call can overlap another, including destruction and borrowed-frame copy.
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeActive)(JNIEnv*, jclass, jlong token, jboolean resume, jboolean focus) {
 std::lock_guard<std::mutex> guard(lock);
 if (!opened || token != generation) return;
 bool before = active(); resumed = resume; focused = focus;
 PF_LOGI("A3_INPUT_STATE resumed=%d focused=%d", resumed, focused);
 if (persistent && before != active()) {
  if (active()) pf_engine_resume(persistent, now());
  else pf_engine_suspend(persistent);
 }
 syncAudio();
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeInput)(JNIEnv*, jclass, jlong token, jint kind, jint a, jint b) {
 std::lock_guard<std::mutex> guard(lock);
 if (!opened || token != generation || !persistent || !active()) return;
 switch (kind) {
 case 0: pf_engine_set_action(persistent, a, b); break;
 case 1: pf_engine_key(persistent, a); break;
 case 2: pf_engine_release(persistent); break;
 case 3: pf_engine_plunger_delta(persistent, a); break;
 case 4: pf_engine_plunger_fire(persistent); break;
 case 5: pf_engine_plunger_target(persistent, a); break;
 }
}
bool androidEngineFrame(bool portrait, std::vector<uint8_t>& pixels, int& width, int& height, int64_t* advancedTicks) {
 std::lock_guard<std::mutex> guard(lock);
 if (advancedTicks) *advancedTicks = -1;
 if (!opened || !persistent || !active()) return false;
 uint64_t before=0, after=0; uint32_t mode=0, table=0, flags=0;
 const bool measured = advancedTicks &&
     pf_engine_state(persistent,&before,&mode,&table,&flags)==PF_OK;
 // Sink only copies borrowed PCM. Runner owns every due 60/71 Hz source tick;
 // queue depth/device callbacks never influence this monotonic deadline.
 if (pf_engine_set_presentation(persistent, portrait ? 1 : 0) != PF_OK ||
     pf_engine_advance(persistent, now(), a4::PcmBuffer::sink, &androidAudioBuffer()) != PF_OK) return false;
 if (measured && pf_engine_state(persistent,&after,&mode,&table,&flags)==PF_OK && after>=before)
     *advancedTicks = static_cast<int64_t>(after-before);
 uint8_t* borrowed = nullptr;
 int32_t w = 0, h = 0, stride = 0;
 if (pf_engine_frame(persistent, &borrowed, &w, &h, &stride) != PF_OK ||
     !a3::copyFrame(borrowed, w, h, stride, pixels)) return false;
 width = w; height = h;
 return true;
}

extern "C" JNIEXPORT jboolean JNICALL JNI_METHOD(nativeHasEngine)(JNIEnv*, jclass, jlong token) {
 std::lock_guard<std::mutex> guard(lock);
 return opened && token == generation && persistent != 0;
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeAudio)(JNIEnv*, jclass, jlong token, jboolean granted, jboolean route) {
 std::lock_guard<std::mutex> guard(lock);
 if (!opened || token != generation) return;
 if (route) {
  androidAudioBuffer().refresh();
  PF_LOGI("A5_ROUTE_GENERATION epoch=%llu", static_cast<unsigned long long>(androidAudioBuffer().current()));
 } else { audioFocus = granted; syncAudio(); }
}

// Packed authoritative snapshot, observed at 10 Hz by the UI looper. No mode
// inference and no unprotected table access. -1 means no current engine.
extern "C" JNIEXPORT jlong JNICALL JNI_METHOD(nativeState)(JNIEnv*, jclass, jlong token) {
 std::lock_guard<std::mutex> guard(lock);
 if (!opened || token != generation || !persistent) return -1;
 uint64_t tick=0; uint32_t mode=0, table=0, flags=0;
 if (pf_engine_state(persistent,&tick,&mode,&table,&flags)!=PF_OK) return -1;
 return static_cast<jlong>(mode | (table<<8) | (flags<<16));
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeDetach)(JNIEnv*, jclass, jlong token) {
 std::lock_guard<std::mutex> guard(lock);
 if (!opened || token != generation) return;
 if (persistent) pf_engine_suspend(persistent);
 opened=false; retained=true; resumed=focused=audioFocus=false; syncAudio();
}

namespace {
std::mutex viewportLock;
std::array<jint,8> renderedViewport{};
}
void androidPublishViewport(int x,int y,int w,int h,int sw,int sh,int fw,int fh) {
 std::lock_guard<std::mutex> guard(viewportLock);
 renderedViewport={x,sh-y-h,w,h,sw,sh,fw,fh}; // GLES bottom origin -> Android top origin
}
extern "C" JNIEXPORT void JNICALL JNI_METHOD(nativeViewport)(JNIEnv* env,jclass,jintArray bounds) {
 if (!bounds || env->GetArrayLength(bounds)!=8) return;
 std::lock_guard<std::mutex> guard(viewportLock);
 env->SetIntArrayRegion(bounds,0,8,renderedViewport.data());
}
