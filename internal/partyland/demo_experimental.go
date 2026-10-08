//go:build demodev

package partyland

import (
	"encoding/binary"
	"fmt"
	"image"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
)

// ExperimentalDemo adapts the existing staged candidate to a host. Playback is
// a consumer of its typed events, and cannot invoke canonical gameplay rules.
type ExperimentalDemo struct {
	core       *demoConnected
	player     *audio.Player
	pcm        []byte
	fadeVolume uint16
}

func (d *ExperimentalDemo) Sync(in physics.Inputs) error {
	c := d.core
	if c.failure != nil {
		return c.failure
	}
	if c.terminal != nil {
		return nil
	}
	// Audio gets precisely the same one rational interval before electronics.
	wasFade := c.fade.Active
	d.pcm = d.player.Sync()
	err := c.sync(in, true)
	if err != nil {
		d.pcm = nil
		return err
	}
	for _, e := range c.game.Events {
		valid := true
		if e.Kind == "Music" {
			valid = e.Value < uint64(len(d.player.Module.Orders))
		}
		if e.Kind == "Sound" {
			effect, ok := audio.Effects[e.Label]
			valid = ok && effect.Sample > 0 && effect.Sample <= len(d.player.Module.Samples) && effect.Note >= 0 && effect.Note < 36
		}
		if !valid {
			d.pcm = nil
			c.phase = "audio output"
			return c.rejectEdge(e.Label, "host audio", "unadmitted sample/note/order")
		}
	}
	for _, e := range c.game.Events {
		switch e.Kind {
		case "Music":
			d.player.Force(int(e.Value))
		case "Sound":
			if e.Label == "SFJADER" {
				d.player.SoundVolume(e.Label, byte(e.Value))
			} else {
				d.player.Sound(e.Label)
			}
		}
	}
	if !c.fade.Active || !wasFade {
		d.fadeVolume = 256
	}
	if len(c.fade.Volume) > 0 {
		d.fadeVolume = c.fade.Volume[len(c.fade.Volume)-1]
	}
	if c.fade.Active {
		scale := d.fadeVolume
		for i := 0; i+1 < len(d.pcm); i += 2 {
			v := int32(int16(binary.LittleEndian.Uint16(d.pcm[i:]))) * int32(scale) / 256
			binary.LittleEndian.PutUint16(d.pcm[i:], uint16(int16(v)))
		}
	}
	// Test traces are observations only; a live session must not grow them forever.
	c.trace = nil
	c.order = nil
	c.snapshots = nil
	c.childFires = nil
	c.drainFires = nil
	c.toucherFires = nil
	c.bonusNodes = nil
	c.fade.Volume = nil
	return nil
}
func (d *ExperimentalDemo) PCM() []byte          { return d.pcm }
func (d *ExperimentalDemo) Done() bool           { return d.core.terminal != nil }
func (d *ExperimentalDemo) Calculations() uint64 { return d.core.calls }
func (d *ExperimentalDemo) Frame() *image.RGBA {
	g := d.core.game
	if !d.core.fade.Active {
		return g.Frame()
	}
	// Consume the candidate's actual DAC palette, which stays black during
	// the trailing WAIT and QUIT rather than using that command's wait counter.
	var palette [768]byte
	for i, v := range d.core.fade.Palette {
		palette[i] = assets.DACRGB(v)
	}
	display := *g.Display
	return presentation.ComposeNative(g.Physics.FramePalette(palette), &display, palette, 96, 242, g.Physics.PresentationSettings(), g.Physics.ScreenOffset)
}

// Diagnostic contains native state only, never source file/code/image payload.
func (d *ExperimentalDemo) Diagnostic() string {
	c := d.core
	if c.failure == nil {
		return ""
	}
	return fmt.Sprintf("%s\nball=%+v score=%s bonus=%s tasks=%v waits=%v matrix=%s cursor=%d timer=%d expired=%t", c.failure, c.game.Physics.Ball, c.game.Score, c.game.Bonus, c.game.taskIDs, c.game.waitCounters, c.game.matrix.op, c.game.matrix.next, c.counter, c.expired)
}
