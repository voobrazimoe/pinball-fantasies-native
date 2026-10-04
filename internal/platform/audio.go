//go:build linux

package platform

/*
#cgo CFLAGS: -I${SRCDIR}/../../.tools/sdl2/usr/include/SDL2 -I${SRCDIR}/../../.tools/sdl2/usr/include/x86_64-linux-gnu
#cgo LDFLAGS: -l:libSDL2-2.0.so.0
#include <SDL.h>
static SDL_AudioDeviceID pf_audio_open(void) {
 SDL_AudioSpec want={0}, have={0};
 want.freq=48000; want.format=AUDIO_S16LSB; want.channels=2; want.samples=1024;
 return SDL_OpenAudioDevice(NULL,0,&want,&have,0);
}
*/
import "C"
import (
	"fmt"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/diagnostics"
	"unsafe"
)

type AudioDevice struct {
	id                      C.SDL_AudioDeviceID
	Dropped                 uint64
	EmptyQueues             uint64
	LifecycleClears, Starts uint64
	started                 bool
}

func OpenAudio() (*AudioDevice, error) {
	if C.SDL_InitSubSystem(C.SDL_INIT_AUDIO) != 0 {
		return nil, fmt.Errorf("SDL audio: %s", C.GoString(C.SDL_GetError()))
	}
	id := C.pf_audio_open()
	if id == 0 {
		C.SDL_QuitSubSystem(C.SDL_INIT_AUDIO)
		return nil, fmt.Errorf("SDL audio: %s", C.GoString(C.SDL_GetError()))
	}
	// Queue a short deterministic pre-roll before starting device consumption.
	diagnostics.Println("SDL audio opened: 48000 Hz stereo signed 16-bit; driver:", C.GoString(C.SDL_GetCurrentAudioDriver()))
	return &AudioDevice{id: id}, nil
}
func (a *AudioDevice) Close() {
	diagnostics.Printf("SDL audio output: queue resets=%d empty-queue observations=%d\n", a.Dropped, a.EmptyQueues)
	diagnostics.Printf("SDL audio continuity: lifecycle clears=%d starts=%d\n", a.LifecycleClears, a.Starts)
	C.SDL_CloseAudioDevice(a.id)
	C.SDL_QuitSubSystem(C.SDL_INIT_AUDIO)
}

// Queue copies the rendered bytes. A slow host may discard output; it never
// changes the renderer or game clock. Cap queued latency at a quarter second.
func (a *AudioDevice) Queue(pcm []byte) error {
	if len(pcm) == 0 {
		return nil
	}
	queued := C.SDL_GetQueuedAudioSize(a.id)
	if a.started && queued == 0 {
		a.EmptyQueues++
	}
	if queued > audio.Rate*audio.BytesPerFrame/4 {
		C.SDL_ClearQueuedAudio(a.id)
		a.Dropped++
	}
	if C.SDL_QueueAudio(a.id, unsafe.Pointer(&pcm[0]), C.Uint32(len(pcm))) != 0 {
		return fmt.Errorf("SDL queue audio: %s", C.GoString(C.SDL_GetError()))
	}
	if !a.started && C.SDL_GetQueuedAudioSize(a.id) >= 2048*audio.BytesPerFrame {
		C.SDL_PauseAudioDevice(a.id, 0)
		a.started = true
		a.Starts++
	}
	return nil
}

// Suspend clears queued output at lifecycle boundaries without advancing tracker
// state. Resume uses the same pre-roll as startup.
func (a *AudioDevice) Suspend(paused bool) {
	a.LifecycleClears++
	C.SDL_PauseAudioDevice(a.id, 1)
	C.SDL_ClearQueuedAudio(a.id)
	a.started = false
}

func (a *AudioDevice) QueuedBytes() int { return int(C.SDL_GetQueuedAudioSize(a.id)) }
