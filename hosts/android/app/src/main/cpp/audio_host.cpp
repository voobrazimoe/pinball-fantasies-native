#include "audio_host.h"
#include "diagnostics.h"
#include <oboe/AudioStream.h>
#include <oboe/AudioStreamBuilder.h>
#include <chrono>
#include <memory>
#include <thread>
#include <jni.h>

namespace a4 {
// Oboe retains these shared callbacks (including any detached error handler).
// Their buffer/atomics outlive the controller and stream close, preventing UAF.
class AudioCallbacks final : public oboe::AudioStreamDataCallback, public oboe::AudioStreamErrorCallback {
public:
    PcmBuffer buffer;
    std::atomic<bool> restart{false};
    std::atomic<uint64_t> callbacks{0}, opens{0}, errors{0}, restarts{0};
    oboe::DataCallbackResult onAudioReady(oboe::AudioStream*, void* output, int32_t frames) override {
        buffer.render(output, static_cast<size_t>(frames));
        callbacks.fetch_add(1, std::memory_order_relaxed);
        return oboe::DataCallbackResult::Continue;
    }
    bool onError(oboe::AudioStream*, oboe::Result) override {
        // Application takes responsibility for close. Signal control; never
        // construct/destroy streams or race Oboe's automatic close thread.
        errors.fetch_add(1, std::memory_order_relaxed);
        restart.store(true, std::memory_order_release);
        return true;
    }
};
class AudioOutput final {
    std::shared_ptr<AudioCallbacks> state=std::make_shared<AudioCallbacks>();
    std::atomic<bool> exiting{false};
    std::shared_ptr<oboe::AudioStream> stream;
    std::thread control;
    void close() {
        if (stream) { stream->close(); stream.reset(); }
        // close joins the data callback. Control temporarily owns consumption.
        buffer.ring.discardBefore(buffer.current());
    }
    bool open() {
        oboe::AudioStreamBuilder builder;
        builder.setDirection(oboe::Direction::Output)
            ->setSampleRate(48000)->setChannelCount(2)->setFormat(oboe::AudioFormat::I16)
            ->setPerformanceMode(oboe::PerformanceMode::LowLatency)
            ->setSharingMode(oboe::SharingMode::Exclusive)
            ->setDataCallback(state)->setErrorCallback(state);
        auto result=builder.openStream(stream);
        if (result!=oboe::Result::OK) {
            builder.setSharingMode(oboe::SharingMode::Shared);
            result=builder.openStream(stream);
        }
        if (result!=oboe::Result::OK) {
            // Allow Oboe's host-side format conversion if native I16 is unavailable.
            builder.setFormatConversionAllowed(true)
                ->setSampleRateConversionQuality(oboe::SampleRateConversionQuality::Medium);
            result=builder.openStream(stream);
        }
        if (result!=oboe::Result::OK || !stream) return false;
        if (stream->getSampleRate()!=48000 || stream->getChannelCount()!=2 ||
            stream->getFormat()!=oboe::AudioFormat::I16) { close(); return false; }
        // Two device bursts; the ring absorbs source batches independently.
        stream->setBufferSizeInFrames(stream->getFramesPerBurst()*2);
        PF_LOGI("A4_STREAM_OPEN rate=48000 channels=2 format=I16 sharing=%d performance=%d burst=%d",
            int(stream->getSharingMode()), int(stream->getPerformanceMode()), stream->getFramesPerBurst());
        if (stream->requestStart()!=oboe::Result::OK) { close(); return false; }
        opens.fetch_add(1);
        return true;
    }
    void run() {
        using namespace std::chrono;
        uint64_t epoch=0;
        auto retry=steady_clock::now(), report=retry+seconds(5);
        while (!exiting.load()) {
            const auto desired=buffer.current();
            const bool failed=state->restart.exchange(false);
            if (desired!=epoch || failed) {
                close(); epoch=desired;
                if (failed) { restarts.fetch_add(1); retry=steady_clock::now()+milliseconds(250); }
                else retry=steady_clock::now();
            }
            if ((epoch&1) && !stream && steady_clock::now()>=retry) {
                if (!open()) { errors.fetch_add(1); restarts.fetch_add(1); }
                retry=steady_clock::now()+seconds(1);
            }
            if (steady_clock::now()>=report) {
                PF_LOGI("A4_AUDIO produced=%llu consumed=%llu callbacks=%llu underruns=%llu missing=%llu overflows=%llu dropped=%llu invalidated=%llu depth=%zu high=%llu opens=%llu errors=%llu restarts=%llu",
                    value(buffer.ring.produced),value(buffer.ring.consumed),value(callbacks),
                    value(buffer.ring.underruns),value(buffer.ring.missing),value(buffer.ring.overflows),
                    value(buffer.ring.dropped),value(buffer.ring.invalidated),buffer.ring.depth(),
                    value(buffer.ring.highWater),value(opens),value(errors),value(restarts));
                report=steady_clock::now()+seconds(5);
            }
            std::this_thread::sleep_for(milliseconds(10));
        }
        close();
    }
    static unsigned long long value(const std::atomic<uint64_t>& n) { return n.load(std::memory_order_relaxed); }
public:
    PcmBuffer& buffer=state->buffer;
    std::atomic<uint64_t>& callbacks=state->callbacks;
    std::atomic<uint64_t>& opens=state->opens;
    std::atomic<uint64_t>& errors=state->errors;
    std::atomic<uint64_t>& restarts=state->restarts;
    AudioOutput() { control=std::thread([this] { run(); }); }
    ~AudioOutput() { exiting.store(true); control.join(); }

};
}
a4::PcmBuffer& androidAudioBuffer() {
    static a4::AudioOutput output;
    return output.buffer;
}

// Asset-free packaged-stream test, isolated from the persistent engine/session.
// Only instrumentation calls this; synthetic PCM never enters gameplay.
extern "C" JNIEXPORT jboolean JNICALL
Java_io_github_voobrazimoe_pinballfantasies_ImportInstrumentation_nativeAudioSmoke(JNIEnv*, jclass) {
    using namespace std::chrono;
    a4::AudioOutput output;
    std::array<int16_t, 960> pcm{};
    for (size_t i=0; i<pcm.size(); i+=2) { pcm[i]=32; pcm[i+1]=-32; }
    auto exercise=[&] {
        const auto before=output.callbacks.load();
        const auto consumed=output.buffer.ring.consumed.load();
        output.buffer.setActive(true);
        const auto deadline=steady_clock::now()+seconds(5);
        while (steady_clock::now()<deadline) {
            a4::PcmBuffer::sink(&output.buffer, reinterpret_cast<const uint8_t*>(pcm.data()), sizeof(pcm));
            if (output.callbacks.load()>before+2 && output.buffer.ring.consumed.load()>consumed) return true;
            std::this_thread::sleep_for(milliseconds(10));
        }
        return false;
    };
    if (!exercise()) return false;
    output.buffer.setActive(false);
    std::this_thread::sleep_for(milliseconds(100));
    const auto paused=output.callbacks.load();
    std::this_thread::sleep_for(milliseconds(100));
    if (output.callbacks.load()!=paused) return false;
    if (!exercise()) return false;
    output.buffer.reset();
    return output.opens.load()>=2 && output.errors.load()==0;
}
