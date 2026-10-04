//go:build windows

package platform

import (
	"fmt"
	"os"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/diagnostics"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

var winmm = syscall.NewLazyDLL("winmm.dll")

//go:uintptrescapes
func mm(name string, args ...uintptr) error {
	code, _, _ := winmm.NewProc(name).Call(args...)
	if code != 0 {
		return fmt.Errorf("waveOut %s: multimedia error %d", name, code)
	}
	return nil
}

type waveFormat struct {
	Tag, Channels      uint16
	Rate, Average      uint32
	Align, Bits, Extra uint16
}
type waveHeader struct {
	Data             uintptr
	Length, Recorded uint32
	User             uintptr
	Flags, Loops     uint32
	Next, Reserved   uintptr
}
type AudioDevice struct {
	id                                            uintptr
	blocks                                        []uintptr // headers and PCM in native LocalAlloc memory; never retained Go pointers
	Dropped, EmptyQueues, LifecycleClears, Starts uint64
	ResetCalls                                    uint64
	started                                       bool
	err                                           error
}

func header(p uintptr) *waveHeader { return (*waveHeader)(unsafe.Pointer(p)) }
func OpenAudio() (*AudioDevice, error) {
	a := &AudioDevice{}
	format := waveFormat{Tag: 1, Channels: 2, Rate: audio.Rate, Average: audio.Rate * audio.BytesPerFrame, Align: audio.BytesPerFrame, Bits: 16}
	if e := mm("waveOutOpen", uintptr(unsafe.Pointer(&a.id)), 0xffffffff, uintptr(unsafe.Pointer(&format)), 0, 0, 0); e != nil {
		return nil, e
	}
	if e := mm("waveOutPause", a.id); e != nil {
		mm("waveOutClose", a.id)
		return nil, e
	}
	diagnostics.Println("waveOut opened: 48000 Hz stereo signed 16-bit")
	return a, nil
}
func (a *AudioDevice) reap(force bool) (int, error) {
	queued := 0
	keep := a.blocks[:0]
	var failure error
	for _, p := range a.blocks {
		h := header(p)
		if force || atomic.LoadUint32(&h.Flags)&1 != 0 {
			if e := mm("waveOutUnprepareHeader", a.id, p, unsafe.Sizeof(waveHeader{})); e != nil {
				failure = e
				keep = append(keep, p)
			} else {
				kernel32.NewProc("LocalFree").Call(p)
			}
		} else {
			queued += int(h.Length)
			keep = append(keep, p)
		}
	}
	a.blocks = keep
	return queued, failure
}
func (a *AudioDevice) reset() error {
	if e := mm("waveOutPause", a.id); e != nil {
		return e
	}
	if e := mm("waveOutReset", a.id); e != nil {
		return e
	}
	a.ResetCalls++
	// Reset may resume device consumption; pause again before accumulating pre-roll.
	if e := mm("waveOutPause", a.id); e != nil {
		return e
	}
	if _, e := a.reap(true); e != nil {
		return e
	}
	if len(a.blocks) != 0 {
		return fmt.Errorf("waveOut reset retained %d buffers", len(a.blocks))
	}
	a.started = false
	return nil
}
func (a *AudioDevice) Queue(pcm []byte) error {
	if a.err != nil {
		return a.err
	}
	if len(pcm) == 0 {
		return nil
	}
	if len(pcm)%audio.BytesPerFrame != 0 {
		return fmt.Errorf("waveOut PCM is not stereo frame aligned")
	}
	queued, e := a.reap(false)
	if e != nil {
		return e
	}
	if a.started && queued == 0 {
		a.EmptyQueues++
		if h := activeWindow; h != nil && os.Getenv("PF12_TRANSITION_LOG") != "" {
			if err := writeTransition(struct {
				Kind         string
				Sequence     uint64
				AtUTC        time.Time
				Drawing      bool
				LastDrawCall string
				Empty        uint64
			}{"queue-empty", h.transitionSequence, time.Now().UTC(), h.drawing, h.lastDrawCall, a.EmptyQueues}); err != nil {
				return err
			}
		}
	}
	// Bound bytes and header count. Slow hosts lose output only, never simulation ticks.
	if queued+len(pcm) > audio.Rate*audio.BytesPerFrame/4 || len(a.blocks) >= 32 {
		if e := a.reset(); e != nil {
			return e
		}
		queued = 0
		a.Dropped++
	}
	if len(pcm) > audio.Rate*audio.BytesPerFrame/4 {
		return fmt.Errorf("waveOut PCM batch exceeds queue bound")
	}
	size := unsafe.Sizeof(waveHeader{})
	p, _, e := kernel32.NewProc("LocalAlloc").Call(0x40, size+uintptr(len(pcm)))
	if p == 0 {
		return apiError("audio LocalAlloc", e)
	}
	h := header(p)
	h.Data = p + size
	h.Length = uint32(len(pcm))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(h.Data)), len(pcm)), pcm)
	if e := mm("waveOutPrepareHeader", a.id, p, size); e != nil {
		kernel32.NewProc("LocalFree").Call(p)
		return e
	}
	if e := mm("waveOutWrite", a.id, p, size); e != nil {
		if ue := mm("waveOutUnprepareHeader", a.id, p, size); ue != nil {
			a.blocks = append(a.blocks, p)
			return fmt.Errorf("%v; %w", e, ue)
		}
		kernel32.NewProc("LocalFree").Call(p)
		return e
	}
	a.blocks = append(a.blocks, p)
	if !a.started && queued+len(pcm) >= 2048*audio.BytesPerFrame {
		if e := mm("waveOutRestart", a.id); e != nil {
			return e
		}
		a.started = true
		a.Starts++
	}
	return nil
}
func (a *AudioDevice) Suspend(paused bool) { a.LifecycleClears++; a.err = a.reset() }
func (a *AudioDevice) Close() {
	if e := a.reset(); e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	if e := mm("waveOutClose", a.id); e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	diagnostics.Printf("waveOut output: queue resets=%d empty-queue observations=%d lifecycle clears=%d starts=%d\n", a.Dropped, a.EmptyQueues, a.LifecycleClears, a.Starts)
}

func (a *AudioDevice) QueuedBytes() int {
	n := 0
	for _, p := range a.blocks {
		h := header(p)
		if atomic.LoadUint32(&h.Flags)&1 == 0 {
			n += int(h.Length)
		}
	}
	return n
}
