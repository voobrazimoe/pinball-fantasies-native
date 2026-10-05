#include "a3_host.h"
#include "presentation_policy.h"
#include "presentation_diagnostics.h"
#include <android/choreographer.h>
#include <android/api-level.h>
#include <dlfcn.h>
#include <EGL/egl.h>
#include <GLES2/gl2.h>
#include "diagnostics.h"
#include <android/native_window.h>
#include <game-activity/native_app_glue/android_native_app_glue.h>

#include <array>
#include <cmath>
#include <cstdint>
#include <vector>

namespace {

constexpr char kLogTag[] = "PinballFantasies";
constexpr int kFrameWidth = 320;
constexpr int kFrameHeight = 609;

#define LOGI(...) PF_LOGI(__VA_ARGS__)
#define LOGE(...) ((void)__android_log_print(ANDROID_LOG_ERROR, kLogTag, __VA_ARGS__))

GLuint compileShader(GLenum type, const char* source) {
    const GLuint shader = glCreateShader(type);
    if (shader == 0U) {
        LOGE("A1_GL_ERROR glCreateShader failed");
        return 0U;
    }

    glShaderSource(shader, 1, &source, nullptr);
    glCompileShader(shader);

    GLint compiled = GL_FALSE;
    glGetShaderiv(shader, GL_COMPILE_STATUS, &compiled);
    if (compiled == GL_TRUE) {
        return shader;
    }

    std::array<char, 512> message{};
    GLsizei written = 0;
    glGetShaderInfoLog(shader, static_cast<GLsizei>(message.size()), &written, message.data());
    LOGE("A1_GL_ERROR shader compile failed: %s", message.data());
    glDeleteShader(shader);
    return 0U;
}

GLuint createProgram() {
    constexpr char kVertexShader[] =
            "attribute vec2 aPosition;\n"
            "attribute vec2 aTexCoord;\n"
            "varying vec2 vTexCoord;\n"
            "void main() {\n"
            "  gl_Position = vec4(aPosition, 0.0, 1.0);\n"
            "  vTexCoord = aTexCoord;\n"
            "}\n";
    constexpr char kFragmentShader[] =
            "precision mediump float;\n"
            "varying vec2 vTexCoord;\n"
            "uniform sampler2D uTexture;\n"
            "void main() {\n"
            "  gl_FragColor = texture2D(uTexture, vTexCoord);\n"
            "}\n";

    const GLuint vertex = compileShader(GL_VERTEX_SHADER, kVertexShader);
    if (vertex == 0U) {
        return 0U;
    }
    const GLuint fragment = compileShader(GL_FRAGMENT_SHADER, kFragmentShader);
    if (fragment == 0U) {
        glDeleteShader(vertex);
        return 0U;
    }

    const GLuint program = glCreateProgram();
    if (program == 0U) {
        glDeleteShader(fragment);
        glDeleteShader(vertex);
        LOGE("A1_GL_ERROR glCreateProgram failed");
        return 0U;
    }

    glAttachShader(program, vertex);
    glAttachShader(program, fragment);
    glLinkProgram(program);
    glDeleteShader(fragment);
    glDeleteShader(vertex);

    GLint linked = GL_FALSE;
    glGetProgramiv(program, GL_LINK_STATUS, &linked);
    if (linked == GL_TRUE) {
        return program;
    }

    std::array<char, 512> message{};
    GLsizei written = 0;
    glGetProgramInfoLog(program, static_cast<GLsizei>(message.size()), &written, message.data());
    LOGE("A1_GL_ERROR program link failed: %s", message.data());
    glDeleteProgram(program);
    return 0U;
}

std::vector<std::uint8_t> makeSyntheticFrame() {
    std::vector<std::uint8_t> pixels(
            static_cast<std::size_t>(kFrameWidth) * static_cast<std::size_t>(kFrameHeight) * 4U);

    for (int y = 0; y < kFrameHeight; ++y) {
        for (int x = 0; x < kFrameWidth; ++x) {
            const bool checker = (((x / 16) + (y / 16)) & 1) != 0;
            const bool border = x < 3 || x >= kFrameWidth - 3 || y < 3 || y >= kFrameHeight - 3;
            const bool center = x >= (kFrameWidth / 2) - 1 && x <= (kFrameWidth / 2) + 1;
            const bool marker = (y % 64) < 2;

            std::uint8_t red = checker ? 22U : 38U;
            std::uint8_t green = checker ? 54U : 76U;
            std::uint8_t blue = checker ? 78U : 104U;
            if (marker) {
                red = 70U;
                green = 132U;
                blue = 164U;
            }
            if (center) {
                red = 214U;
                green = 180U;
                blue = 72U;
            }
            if (border) {
                red = 238U;
                green = 238U;
                blue = 238U;
            }

            const std::size_t offset =
                    (static_cast<std::size_t>(y) * static_cast<std::size_t>(kFrameWidth) +
                     static_cast<std::size_t>(x)) *
                    4U;
            pixels[offset + 0U] = red;
            pixels[offset + 1U] = green;
            pixels[offset + 2U] = blue;
            pixels[offset + 3U] = 255U;
        }
    }
    return pixels;
}

class Renderer {
public:
    ~Renderer() {
        detach();
    }

    Renderer(const Renderer&) = delete;
    Renderer& operator=(const Renderer&) = delete;
    Renderer() = default;

    bool attach(ANativeWindow* window) {
        detach();
        if (window == nullptr) {
            return false;
        }

        display_ = eglGetDisplay(EGL_DEFAULT_DISPLAY);
        if (display_ == EGL_NO_DISPLAY || eglInitialize(display_, nullptr, nullptr) != EGL_TRUE) {
            LOGE("A1_EGL_ERROR failed to initialize display: 0x%x", eglGetError());
            detach();
            return false;
        }

        constexpr EGLint kConfigAttributes[] = {
                EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
                EGL_RENDERABLE_TYPE, EGL_OPENGL_ES2_BIT,
                EGL_RED_SIZE, 8,
                EGL_GREEN_SIZE, 8,
                EGL_BLUE_SIZE, 8,
                EGL_ALPHA_SIZE, 8,
                EGL_NONE};
        EGLConfig config = nullptr;
        EGLint configCount = 0;
        if (eglChooseConfig(display_, kConfigAttributes, &config, 1, &configCount) != EGL_TRUE ||
            configCount != 1) {
            LOGE("A1_EGL_ERROR no RGBA8 GLES2 window config: 0x%x", eglGetError());
            detach();
            return false;
        }

        EGLint nativeFormat = 0;
        if (eglGetConfigAttrib(display_, config, EGL_NATIVE_VISUAL_ID, &nativeFormat) != EGL_TRUE) {
            LOGE("A1_EGL_ERROR failed to query native visual: 0x%x", eglGetError());
            detach();
            return false;
        }
        if (ANativeWindow_setBuffersGeometry(window, 0, 0, nativeFormat) != 0) {
            LOGE("A1_EGL_ERROR failed to set native window format");
            detach();
            return false;
        }

        constexpr EGLint kContextAttributes[] = {EGL_CONTEXT_CLIENT_VERSION, 2, EGL_NONE};
        context_ = eglCreateContext(display_, config, EGL_NO_CONTEXT, kContextAttributes);
        if (context_ == EGL_NO_CONTEXT) {
            LOGE("A1_EGL_ERROR failed to create GLES2 context: 0x%x", eglGetError());
            detach();
            return false;
        }

        surface_ = eglCreateWindowSurface(display_, config, window, nullptr);
        if (surface_ == EGL_NO_SURFACE) {
            LOGE("A1_EGL_ERROR failed to create window surface: 0x%x", eglGetError());
            detach();
            return false;
        }
        if (eglMakeCurrent(display_, surface_, surface_, context_) != EGL_TRUE) {
            LOGE("A1_EGL_ERROR failed to make context current: 0x%x", eglGetError());
            detach();
            return false;
        }
        (void)eglSwapInterval(display_, 1);

        program_ = createProgram();
        if (program_ == 0U) {
            detach();
            return false;
        }
        positionLocation_ = glGetAttribLocation(program_, "aPosition");
        texCoordLocation_ = glGetAttribLocation(program_, "aTexCoord");
        textureLocation_ = glGetUniformLocation(program_, "uTexture");
        if (positionLocation_ < 0 || texCoordLocation_ < 0 || textureLocation_ < 0) {
            LOGE("A1_GL_ERROR required shader locations missing");
            detach();
            return false;
        }

        glGenTextures(1, &texture_);
        if (texture_ == 0U) {
            LOGE("A1_GL_ERROR failed to create texture");
            detach();
            return false;
        }
        const std::vector<std::uint8_t> pixels = makeSyntheticFrame();
        frameWidth_ = textureWidth_ = kFrameWidth;
        frameHeight_ = textureHeight_ = kFrameHeight;
        glBindTexture(GL_TEXTURE_2D, texture_);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);
        glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);
        glPixelStorei(GL_UNPACK_ALIGNMENT, 1);
        glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, kFrameWidth, kFrameHeight, 0, GL_RGBA,
                     GL_UNSIGNED_BYTE, pixels.data());

        glDisable(GL_BLEND);
        glDisable(GL_DEPTH_TEST);
        glDisable(GL_DITHER);
        firstFrame_ = true;
        lastSurfaceWidth_ = -1;
        lastSurfaceHeight_ = -1;
        LOGI("A1_SURFACE_READY source=%dx%d", kFrameWidth, kFrameHeight);
        return true;
    }

    void detach() {
        if (display_ != EGL_NO_DISPLAY && context_ != EGL_NO_CONTEXT && surface_ != EGL_NO_SURFACE) {
            (void)eglMakeCurrent(display_, surface_, surface_, context_);
            if (texture_ != 0U) {
                glDeleteTextures(1, &texture_);
            }
            if (program_ != 0U) {
                glDeleteProgram(program_);
            }
        }
        texture_ = 0U;
        program_ = 0U;
        positionLocation_ = -1;
        texCoordLocation_ = -1;
        textureLocation_ = -1;

        if (display_ != EGL_NO_DISPLAY) {
            (void)eglMakeCurrent(display_, EGL_NO_SURFACE, EGL_NO_SURFACE, EGL_NO_CONTEXT);
            if (surface_ != EGL_NO_SURFACE) {
                (void)eglDestroySurface(display_, surface_);
            }
            if (context_ != EGL_NO_CONTEXT) {
                (void)eglDestroyContext(display_, context_);
            }
            (void)eglTerminate(display_);
        }
        surface_ = EGL_NO_SURFACE;
        context_ = EGL_NO_CONTEXT;
        display_ = EGL_NO_DISPLAY;
        lastSurfaceWidth_ = -1;
        lastSurfaceHeight_ = -1;
    }

    [[nodiscard]] bool ready() const {
        return display_ != EGL_NO_DISPLAY && surface_ != EGL_NO_SURFACE &&
               context_ != EGL_NO_CONTEXT && program_ != 0U && texture_ != 0U;
    }

    void setResumed(bool resumed) {
        resumed_ = resumed;
        resumeFrame_ = resumed;
        LOGI("A1_ACTIVITY_%s", resumed ? "RESUMED" : "PAUSED");
    }

    [[nodiscard]] bool active() const {
        return resumed_ && focused_ && ready();
    }

    bool draw(int64_t vsync) {
        const bool diagnostic = pfDiagnosticsEnabled();
        const int64_t begin = diagnostic ? presentation::monotonicNs() : 0;
        if (diagnostic) diagnostics_.begin(begin, vsync);
        else if (diagnostics_.start) diagnostics_ = {};
        if (!ready()) {
            if (diagnostic) diagnostics_.finish(presentation::monotonicNs(), false, -1);
            return false;
        }

        EGLint surfaceWidth = 0;
        EGLint surfaceHeight = 0;
        if (eglQuerySurface(display_, surface_, EGL_WIDTH, &surfaceWidth) != EGL_TRUE ||
            eglQuerySurface(display_, surface_, EGL_HEIGHT, &surfaceHeight) != EGL_TRUE ||
            surfaceWidth <= 0 || surfaceHeight <= 0) {
            if (diagnostic) diagnostics_.finish(presentation::monotonicNs(), false, -1);
            return false;
        }

        int64_t sourceTicks = -1;
        const int64_t engineStart = diagnostic ? presentation::monotonicNs() : 0;
        const bool hasFrame = androidEngineFrame(a3::portrait(surfaceWidth, surfaceHeight), pixels_, frameWidth_, frameHeight_, diagnostic ? &sourceTicks : nullptr);
        const int64_t uploadStart = diagnostic ? presentation::monotonicNs() : 0;
        if (diagnostic) diagnostics_.engine.add(uploadStart-engineStart);
        if (hasFrame) {
            glBindTexture(GL_TEXTURE_2D, texture_);
            if (textureWidth_ != frameWidth_ || textureHeight_ != frameHeight_) {
                glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, frameWidth_, frameHeight_, 0,
                             GL_RGBA, GL_UNSIGNED_BYTE, pixels_.data());
                textureWidth_ = frameWidth_; textureHeight_ = frameHeight_;
            } else {
                glTexSubImage2D(GL_TEXTURE_2D, 0, 0, 0, frameWidth_, frameHeight_,
                                GL_RGBA, GL_UNSIGNED_BYTE, pixels_.data());
            }
        }
        const int64_t drawStart = diagnostic ? presentation::monotonicNs() : 0;
        if (diagnostic) diagnostics_.upload.add(drawStart-uploadStart);
        const auto viewport = a3::letterbox(surfaceWidth, surfaceHeight, frameWidth_, frameHeight_);
        const int viewportX = viewport.x, viewportY = viewport.y;
        const int viewportWidth = viewport.w, viewportHeight = viewport.h;

        if (surfaceWidth != lastSurfaceWidth_ || surfaceHeight != lastSurfaceHeight_) {
            LOGI("A1_VIEWPORT surface=%dx%d orientation=%s viewport=%d,%d,%dx%d source=%dx%d",
                 surfaceWidth, surfaceHeight,
                 surfaceHeight >= surfaceWidth ? "portrait" : "landscape",
                 viewportX, viewportY, viewportWidth, viewportHeight,
                 frameWidth_, frameHeight_);
            lastSurfaceWidth_ = surfaceWidth;
            lastSurfaceHeight_ = surfaceHeight;
        }

        glViewport(0, 0, surfaceWidth, surfaceHeight);
        glClearColor(0.0F, 0.0F, 0.0F, 1.0F);
        glClear(GL_COLOR_BUFFER_BIT);
        glViewport(viewportX, viewportY, viewportWidth, viewportHeight);
        androidPublishViewport(viewportX,viewportY,viewportWidth,viewportHeight,surfaceWidth,surfaceHeight,frameWidth_,frameHeight_);

        constexpr GLfloat kVertices[] = {
                -1.0F, -1.0F, 0.0F, 1.0F,
                 1.0F, -1.0F, 1.0F, 1.0F,
                -1.0F,  1.0F, 0.0F, 0.0F,
                 1.0F,  1.0F, 1.0F, 0.0F};

        glUseProgram(program_);
        glActiveTexture(GL_TEXTURE0);
        glBindTexture(GL_TEXTURE_2D, texture_);
        glUniform1i(textureLocation_, 0);
        glBindBuffer(GL_ARRAY_BUFFER, 0U);
        glEnableVertexAttribArray(static_cast<GLuint>(positionLocation_));
        glEnableVertexAttribArray(static_cast<GLuint>(texCoordLocation_));
        glVertexAttribPointer(static_cast<GLuint>(positionLocation_), 2, GL_FLOAT, GL_FALSE,
                              4 * static_cast<GLsizei>(sizeof(GLfloat)), kVertices);
        glVertexAttribPointer(static_cast<GLuint>(texCoordLocation_), 2, GL_FLOAT, GL_FALSE,
                              4 * static_cast<GLsizei>(sizeof(GLfloat)), kVertices + 2);
        glDrawArrays(GL_TRIANGLE_STRIP, 0, 4);
        glDisableVertexAttribArray(static_cast<GLuint>(positionLocation_));
        glDisableVertexAttribArray(static_cast<GLuint>(texCoordLocation_));

        const int64_t swapStart = diagnostic ? presentation::monotonicNs() : 0;
        if (diagnostic) diagnostics_.draw.add(swapStart-drawStart);
        const bool swapped = eglSwapBuffers(display_, surface_) == EGL_TRUE;
        if (diagnostic) {
            const int64_t end = presentation::monotonicNs();
            diagnostics_.swap.add(end-swapStart);
            diagnostics_.finish(end, swapped, sourceTicks);
        }
        if (!swapped) {
            LOGE("A1_EGL_ERROR eglSwapBuffers failed: 0x%x", eglGetError());
            return false;
        }
        if (firstFrame_) {
            firstFrame_ = false;
            LOGI("A1_FRAME_PRESENTED surface=%dx%d viewport=%d,%d,%dx%d",
                 surfaceWidth, surfaceHeight, viewportX, viewportY, viewportWidth, viewportHeight);
        }
        if (resumeFrame_) {
            resumeFrame_ = false;
            LOGI("A1_ACTIVE_FRAME");
        }
        return true;
    }

    void initializePacing() {
        choreographer_ = AChoreographer_getInstance();
        if (!choreographer_) LOGE("A1_PACING_ERROR no Looper Choreographer");
        if (android_get_device_api_level() >= 29)
            post64_ = reinterpret_cast<Post64>(dlsym(RTLD_DEFAULT, "AChoreographer_postFrameCallback64"));
        LOGI("A6_PACING driver=Choreographer api=%d callback=%s swap_interval=1",
             android_get_device_api_level(), post64_ ? "64" : "long64");
    }
    void setFocused(bool value) { focused_ = value; }
    void syncPacing() {
        policy_.setActive(active());
        if (!active()) diagnostics_ = {};
        arm();
    }
    void stopPacing() { policy_.setActive(false); diagnostics_ = {}; }
    bool callbackPending() const { return policy_.pending(); }
private:
    using Post64 = void (*)(AChoreographer*, void (*)(int64_t, void*), void*);
    static void callback64(int64_t vsync, void* data) {
        auto* self = static_cast<Renderer*>(data);
        if (self->policy_.consume() && self->active()) (void)self->draw(vsync);
        self->arm();
    }
    static void callbackLegacy(long vsync, void* data) {
        static_assert(sizeof(long)==sizeof(int64_t), "Android public ABIs must be 64-bit");
        callback64(static_cast<int64_t>(vsync), data);
    }
    void arm() {
        if (!choreographer_ || !policy_.arm()) return;
        if (post64_) post64_(choreographer_, callback64, this);
        else AChoreographer_postFrameCallback(choreographer_, callbackLegacy, this);
    }
    AChoreographer* choreographer_ = nullptr;
    Post64 post64_ = nullptr;
    presentation::Policy policy_;
    presentation::Diagnostics diagnostics_;
    bool focused_ = false;
    std::vector<uint8_t> pixels_;
    int frameWidth_ = kFrameWidth, frameHeight_ = kFrameHeight;
    int textureWidth_ = kFrameWidth, textureHeight_ = kFrameHeight;
    EGLDisplay display_ = EGL_NO_DISPLAY;
    EGLSurface surface_ = EGL_NO_SURFACE;
    EGLContext context_ = EGL_NO_CONTEXT;
    GLuint program_ = 0U;
    GLuint texture_ = 0U;
    GLint positionLocation_ = -1;
    GLint texCoordLocation_ = -1;
    GLint textureLocation_ = -1;
    int lastSurfaceWidth_ = -1;
    int lastSurfaceHeight_ = -1;
    bool firstFrame_ = true;
    bool resumed_ = false;
    bool resumeFrame_ = false;
};

void handleAppCommand(android_app* app, int32_t command) {
    auto* renderer = static_cast<Renderer*>(app->userData);
    if (renderer == nullptr) {
        return;
    }

    switch (command) {
        case APP_CMD_GAINED_FOCUS:
            renderer->setFocused(true);
            LOGI("A3_NATIVE_FOCUS gained");
            break;
        case APP_CMD_LOST_FOCUS:
            renderer->setFocused(false);
            LOGI("A3_NATIVE_FOCUS lost");
            break;
        case APP_CMD_RESUME:
            renderer->setResumed(true);
            break;
        case APP_CMD_PAUSE:
            renderer->setResumed(false);
            break;
        case APP_CMD_INIT_WINDOW:
            renderer->stopPacing();
            if (app->window != nullptr) {
                (void)renderer->attach(app->window);
            }
            break;
        case APP_CMD_TERM_WINDOW:
            renderer->stopPacing();
            renderer->detach();
            break;
        default:
            break;
    }
    renderer->syncPacing();
}

}  // namespace

extern "C" void android_main(struct android_app* app) {
    Renderer renderer;
    renderer.initializePacing();
    app->userData = &renderer;
    app->onAppCmd = handleAppCommand;
    LOGI("A1_HOST_STARTED");

    while (app->destroyRequested == 0) {
        int events = 0;
        android_poll_source* source = nullptr;
        const int timeoutMillis = presentation::Policy::pollTimeoutMillis();
        const int result = ALooper_pollOnce(
                timeoutMillis, nullptr, &events, reinterpret_cast<void**>(&source));
        if (result >= 0 && source != nullptr) {
            source->process(app, source);
        }
        if (app->destroyRequested != 0) {
            break;
        }
    }

    // Native callbacks execute on this Looper. Drain the sole pending callback
    // with production disabled before destroying its data; no dangling pointer,
    // no re-arm, and no rendering after destruction/surface termination.
    renderer.stopPacing();
    while (renderer.callbackPending()) ALooper_pollOnce(-1, nullptr, nullptr, nullptr);
    app->userData = nullptr;
    app->onAppCmd = nullptr;
    renderer.detach();
    LOGI("A1_HOST_STOPPED");
}
