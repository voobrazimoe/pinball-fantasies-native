#pragma once
// Small device double for the production AudioOutput controller/callback. Real
// Oboe is separately built and exercised in packaged Android instrumentation.
#include <atomic>
#include <chrono>
#include <memory>
#include <thread>
#include <cstdint>
namespace oboe {
enum class Usage { Game };
enum class ContentType { Music };
enum class Direction { Output };
enum class AudioFormat { I16 };
enum class PerformanceMode { LowLatency };
enum class SharingMode { Exclusive, Shared };
enum class SampleRateConversionQuality { Medium };
enum class Result { OK, ErrorDisconnected };
enum class DataCallbackResult { Continue };
class AudioStream;
class AudioStreamDataCallback { public: virtual ~AudioStreamDataCallback()=default;
    virtual DataCallbackResult onAudioReady(AudioStream*,void*,int32_t)=0; };
class AudioStreamErrorCallback { public: virtual ~AudioStreamErrorCallback()=default;
    virtual bool onError(AudioStream*,Result)=0; };
inline std::atomic<int> exclusiveFailures{0}, openFailures{0}, closeCount{0};
inline std::atomic<bool> injectError{false};
extern thread_local bool realtime;
class AudioStream {
    std::atomic<bool> stopped{false};
    std::thread worker;
public:
    std::shared_ptr<AudioStreamDataCallback> callback;
    std::shared_ptr<AudioStreamErrorCallback> error;
    SharingMode sharing=SharingMode::Exclusive;
    int getSampleRate() { return 48000; }
    int getChannelCount() { return 2; }
    AudioFormat getFormat() { return AudioFormat::I16; }
    SharingMode getSharingMode() { return sharing; }
    PerformanceMode getPerformanceMode() { return PerformanceMode::LowLatency; }
    int getDeviceId() { return 0; }
    int getFramesPerBurst() { return 192; }
    Result setBufferSizeInFrames(int) { return Result::OK; }
    Result requestStart() {
        worker=std::thread([this] {
            int16_t output[384];
            while (!stopped.load()) {
                realtime=true;
                callback->onAudioReady(this,output,192);
                if (injectError.exchange(false)) { error->onError(this,Result::ErrorDisconnected); realtime=false; break; }
                realtime=false;
                std::this_thread::sleep_for(std::chrono::milliseconds(2));
            }
        }); return Result::OK;
    }
    Result close() { stopped=true; if(worker.joinable()) worker.join(); ++closeCount; return Result::OK; }
    ~AudioStream() { if(worker.joinable()) close(); }
};
class AudioStreamBuilder {
    std::shared_ptr<AudioStreamDataCallback> data;
    std::shared_ptr<AudioStreamErrorCallback> error;
    SharingMode sharing=SharingMode::Exclusive;
public:
    AudioStreamBuilder* setUsage(Usage) { return this; }
    AudioStreamBuilder* setContentType(ContentType) { return this; }
    AudioStreamBuilder* setDirection(Direction) { return this; }
    AudioStreamBuilder* setSampleRate(int n) { if(n!=48000) std::abort(); return this; }
    AudioStreamBuilder* setChannelCount(int n) { if(n!=2) std::abort(); return this; }
    AudioStreamBuilder* setFormat(AudioFormat) { return this; }
    AudioStreamBuilder* setPerformanceMode(PerformanceMode) { return this; }
    AudioStreamBuilder* setSharingMode(SharingMode s) { sharing=s; return this; }
    AudioStreamBuilder* setFormatConversionAllowed(bool) { return this; }
    AudioStreamBuilder* setSampleRateConversionQuality(SampleRateConversionQuality) { return this; }
    AudioStreamBuilder* setDataCallback(std::shared_ptr<AudioStreamDataCallback> c) { data=c; return this; }
    AudioStreamBuilder* setErrorCallback(std::shared_ptr<AudioStreamErrorCallback> c) { error=c; return this; }
    Result openStream(std::shared_ptr<AudioStream>& stream) {
        if (openFailures.load()>0) { --openFailures; return Result::ErrorDisconnected; }
        if(sharing==SharingMode::Exclusive && exclusiveFailures.load()>0) {
            --exclusiveFailures; return Result::ErrorDisconnected;
        }
        stream=std::make_shared<AudioStream>(); stream->callback=data; stream->error=error; stream->sharing=sharing;
        return Result::OK;
    }
};
}
