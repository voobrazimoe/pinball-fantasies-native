//go:build windows

package platform

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"time"
)

// Measurements never enter scheduling. Queue bytes count incomplete headers
// (including the playing header), so queue milliseconds are an upper bound.
type transitionAudio struct {
	QueuedBytes                   int
	QueueMS                       float64
	Empty, Resets, Starts, Clears uint64
	WaveOutResets                 uint64
	EmptyObserved                 bool
}

func audioSnapshot(a *AudioDevice) transitionAudio {
	if a == nil {
		return transitionAudio{}
	}
	n := a.QueuedBytes()
	return transitionAudio{n, float64(n) / 192, a.EmptyQueues, a.Dropped, a.Starts, a.LifecycleClears, a.ResetCalls, a.started && n == 0}
}

type transitionTime struct {
	Name       string
	DurationNS int64
}
type transitionEvents struct {
	Count      uint64
	DurationNS int64
}
type fullscreenTransition struct {
	Kind                      string
	Sequence                  uint64
	Fullscreen                bool
	HWND                      uintptr
	DurationNS                int64
	Pre, Post                 transitionAudio
	GeometryPre, GeometryPost WindowGeometry
	Calls                     []transitionTime
	Events                    map[string]*transitionEvents
	SurfaceRecreations        uint64
	Error                     string `json:",omitempty"`
}

func transitionMessage(msg uint32) string {
	switch msg {
	case 0x5:
		return "WM_SIZE"
	case 0xf:
		return "WM_PAINT"
	case 0x14:
		return "WM_ERASEBKGND"
	case 0x46:
		return "WM_WINDOWPOSCHANGING"
	case 0x47:
		return "WM_WINDOWPOSCHANGED"
	}
	return ""
}
func (t *fullscreenTransition) event(name string, d time.Duration) {
	e := t.Events[name]
	if e == nil {
		e = &transitionEvents{}
		t.Events[name] = e
	}
	e.Count++
	e.DurationNS += d.Nanoseconds()
}

//go:uintptrescapes
func (h *hostWindow) transitionCall(name string, args ...uintptr) (uintptr, uintptr, error) {
	if h.transition == nil {
		return up(name).Call(args...)
	}
	start := time.Now()
	v, r, e := up(name).Call(args...)
	h.transition.Calls = append(h.transition.Calls, transitionTime{name, time.Since(start).Nanoseconds()})
	return v, r, e
}

func writeTransition(v any) error {
	path := os.Getenv("PF12_TRANSITION_LOG")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(v)
	ce := f.Close()
	if err != nil {
		return err
	}
	return ce
}

func (h *hostWindow) ToggleFullscreen(a *AudioDevice) error {
	if os.Getenv("PF12_TRANSITION_LOG") == "" {
		var clears, starts uint64
		if a != nil {
			clears, starts = a.LifecycleClears, a.Starts
		}
		if err := h.presentation.ToggleFullscreen(); err != nil {
			return err
		}
		h.transitionSettlingUntil = time.Now().Add(500 * time.Millisecond)
		fmt.Printf("PF12 fullscreen=%t window ID=%d\n", h.presentation.IsFullscreen(), h.hwnd)
		h.FullscreenTrace(clears, starts, a)
		return nil
	}
	h.transitionSequence++
	t := &fullscreenTransition{Kind: "transition", Sequence: h.transitionSequence,
		Fullscreen: !h.presentation.IsFullscreen(), HWND: h.hwnd, Events: map[string]*transitionEvents{}}
	t.GeometryPre = h.Geometry()
	h.transition = t
	t.Pre = audioSnapshot(a)
	start := time.Now()
	err := h.presentation.ToggleFullscreen()
	t.DurationNS = time.Since(start).Nanoseconds()
	t.Post = audioSnapshot(a) // Immediately after transaction, before diagnostic I/O.
	// Observe a drained playing queue here even if the next Queue has not run.
	// Queue's production counter is deliberately unchanged by instrumentation.
	t.GeometryPost = h.Geometry()
	h.transition = nil
	h.transitionDrawPending, h.transitionDevice = true, a
	h.transitionDrawUntil = time.Now().Add(500 * time.Millisecond)
	h.transitionSettlingUntil = h.transitionDrawUntil
	if err != nil {
		t.Error = err.Error()
	}
	if e := writeTransition(t); e != nil && err == nil {
		err = e
	}
	fmt.Printf("PF12 transition %d fullscreen=%t HWND=%#x duration=%.3fms queue=%d/%d empty=%d/%d resets=%d/%d starts=%d/%d clears=%d/%d geometry=%+v\n",
		t.Sequence, t.Fullscreen, t.HWND, float64(t.DurationNS)/1e6, t.Pre.QueuedBytes, t.Post.QueuedBytes,
		t.Pre.Empty, t.Post.Empty, t.Pre.Resets, t.Post.Resets, t.Pre.Starts, t.Post.Starts, t.Pre.Clears, t.Post.Clears, t.GeometryPost)
	return err
}

func (h *hostWindow) surfaceTrace(start time.Time, size image.Point) {
	if h.transition != nil {
		h.transition.SurfaceRecreations++
	}
	if os.Getenv("PF12_TRANSITION_LOG") == "" {
		return
	}
	if e := writeTransition(struct {
		Kind             string
		Sequence         uint64
		DurationNS       int64
		Size             image.Point
		DuringTransition bool
	}{"surface", h.transitionSequence, time.Since(start).Nanoseconds(), size, h.changingStyle}); e != nil {
		h.err = fmt.Errorf("transition surface log: %w", e)
	}
}

//go:uintptrescapes
func (h *hostWindow) drawCall(times *[]transitionTime, name string, args ...uintptr) (uintptr, uintptr, error) {
	h.lastDrawCall = name
	if times == nil {
		return gdi32.NewProc(name).Call(args...)
	}
	start := time.Now()
	v, r, e := gdi32.NewProc(name).Call(args...)
	*times = append(*times, transitionTime{name, time.Since(start).Nanoseconds()})
	return v, r, e
}

//go:uintptrescapes
func (h *hostWindow) drawHostCall(name string, args ...uintptr) (uintptr, uintptr, error) {
	if os.Getenv("PF12_TRANSITION_LOG") == "" || !time.Now().Before(h.transitionDrawUntil) {
		return up(name).Call(args...)
	}
	start, pre := time.Now(), audioSnapshot(h.transitionDevice)
	v, r, e := up(name).Call(args...)
	post, duration := audioSnapshot(h.transitionDevice), time.Since(start).Nanoseconds()
	if err := writeTransition(struct {
		Kind, Name string
		Sequence   uint64
		DurationNS int64
		Pre, Post  transitionAudio
	}{"host-call", name, h.transitionSequence, duration, pre, post}); err != nil {
		h.err = err
	}
	return v, r, e
}
