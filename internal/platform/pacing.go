package platform

import (
	"encoding/json"
	"os"
	"time"
)

// Optional host diagnostics. No audio/device measurement enters the game clock.
type pacingLog struct {
	file  *os.File
	start time.Time
	seq   int
}

func openPacingLog() (*pacingLog, error) {
	path := os.Getenv("PF12_PACING_LOG")
	if path == "" {
		return nil, nil
	}
	f, e := os.Create(path)
	if e != nil {
		return nil, e
	}
	return &pacingLog{file: f, start: time.Now()}, nil
}
func (p *pacingLog) Close() {
	if p != nil {
		p.file.Close()
	}
}
func (p *pacingLog) mark(stage string, start time.Time, fullscreen bool, device *AudioDevice) {
	if p == nil {
		return
	}
	queued := 0
	var empty, resets uint64
	if device != nil {
		queued = device.QueuedBytes()
		empty, resets = device.EmptyQueues, device.Dropped
	}
	p.seq++
	json.NewEncoder(p.file).Encode(struct {
		Sequence          int    `json:"sequence"`
		Stage             string `json:"stage"`
		Elapsed, Duration int64
		Fullscreen        bool
		QueuedBytes       int
		Empty, Resets     uint64
	}{p.seq, stage, time.Since(p.start).Nanoseconds(), time.Since(start).Nanoseconds(), fullscreen, queued, empty, resets})
}
