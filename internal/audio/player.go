package audio

import "math"

const (
	Channels      = 2
	BytesPerFrame = Channels * 2
)

var periods = [...]int{856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453, 428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226, 214, 202, 190, 180, 170, 160, 151, 143, 135, 127, 120, 113}
var sine = [...]int{0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253, 255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24}

type voice struct {
	centered                                     bool // Runtime sound effects play in both ears on the same sample phase.
	sample                                       int
	phase, step                                  uint64
	period, pitch, volume, target, porta, offset int
	effect, param                                int
	vibSpeed, vibDepth, vibPhase                 int
}
type Player struct {
	Module                  *Module
	Order, Row, Tick, Speed int
	Voices                  [4]voice
	// Jump is deterministic musical control, never a device callback.
	Jump          func(int) int
	jump, br      int
	tickFrames    int
	syncRemainder int
	started       bool
	Frames        uint64
	Jumps         uint64
}

func New(m *Module) *Player { p := &Player{Module: m}; p.Force(0); return p }

// Force resets tracker phase, retaining channel/sample state, matching PF4.5's convention.
func (p *Player) Force(order int) {
	p.Order = order
	p.Row = 0
	p.Tick = 0
	p.Speed = 6
	p.tickFrames = 0
	p.jump = -1
	p.br = -1
	p.started = false
}
func (p *Player) setStep(v *voice) {
	if v.pitch <= 0 || v.sample <= 0 {
		v.step = 0
		return
	}
	fine := p.Module.Samples[v.sample-1].Fine
	hz := 3546895.0 / float64(v.pitch) * math.Pow(2, float64(fine)/96)
	v.step = uint64(hz * 4294967296 / Rate)
}
func (p *Player) row() {
	p.jump = -1
	p.br = -1
	for i, n := range p.Module.Patterns[p.Module.Orders[p.Order]][p.Row] {
		v := &p.Voices[i]
		if n.Sample > 0 || n.Period > 0 {
			v.centered = false
		}
		v.effect = n.Effect
		v.param = n.Param
		if n.Sample > 0 {
			v.sample = n.Sample
			v.volume = p.Module.Samples[n.Sample-1].Volume
		}
		if n.Period > 0 {
			if n.Effect == 3 {
				v.target = n.Period
			} else {
				v.period = n.Period
				v.phase = 0
				v.vibPhase = 0
			}
		}
		switch n.Effect {
		case 3:
			if n.Param != 0 {
				v.porta = n.Param
			}
		case 4:
			if n.Param>>4 != 0 {
				v.vibSpeed = n.Param >> 4
			}
			if n.Param&15 != 0 {
				v.vibDepth = n.Param & 15
			}
		case 9:
			if n.Param != 0 {
				v.offset = n.Param * 256
			}
			if n.Period > 0 {
				v.phase = uint64(v.offset) << 32
			}
		case 11:
			p.jump = n.Param
		case 12:
			v.volume = min(n.Param, 64)
		case 13:
			p.br = (n.Param>>4)*10 + (n.Param & 15)
		case 15:
			p.Speed = n.Param
		}
		v.pitch = v.period
		p.setStep(v)
	}
}
func (p *Player) effects() {
	for i := range p.Voices {
		v := &p.Voices[i]
		v.pitch = v.period
		slideVolume := func() {
			up := v.param >> 4
			if up > 0 {
				v.volume += up
			} else {
				v.volume -= v.param & 15
			}
			v.volume = max(0, min(64, v.volume))
		}
		vibrato := func() {
			s := sine[v.vibPhase&31]
			if v.vibPhase&32 != 0 {
				s = -s
			}
			v.pitch = v.period + s*v.vibDepth/128
			v.vibPhase = (v.vibPhase + v.vibSpeed) & 63
		}
		switch v.effect {
		case 0:
			if v.param != 0 {
				semi := 0
				if p.Tick%3 == 1 {
					semi = v.param >> 4
				}
				if p.Tick%3 == 2 {
					semi = v.param & 15
				}
				v.pitch = int(float64(v.period)/math.Pow(2, float64(semi)/12) + 0.5)
			}
		case 1:
			v.period = max(113, v.period-v.param)
			v.pitch = v.period
		case 2:
			v.period = min(856, v.period+v.param)
			v.pitch = v.period
		case 3:
			if v.target > 0 {
				if v.period < v.target {
					v.period = min(v.target, v.period+v.porta)
				} else {
					v.period = max(v.target, v.period-v.porta)
				}
				v.pitch = v.period
			}
		case 4:
			vibrato()
		case 6:
			vibrato()
			slideVolume()
		case 10:
			slideVolume()
		case 14:
			if v.param>>4 == 9 && v.param&15 != 0 && p.Tick%(v.param&15) == 0 {
				v.phase = 0
			}
		}
		p.setStep(v)
	}
}
func (p *Player) boundary() {
	p.Tick++
	if p.Tick < p.Speed {
		p.effects()
		return
	}
	p.Tick = 0
	if p.jump >= 0 {
		next := p.jump
		if p.Jump != nil {
			next = p.Jump(next)
		}
		p.Order = next
		p.Row = 0
		p.Speed = 6
		p.Jumps++
	} else if p.br >= 0 {
		p.Order = (p.Order + 1) % len(p.Module.Orders)
		p.Row = p.br
	} else {
		p.Row++
		if p.Row == 64 {
			p.Row = 0
			p.Order = (p.Order + 1) % len(p.Module.Orders)
		}
	}
	p.row()
}
func (p *Player) Sound(label string) bool {
	e, ok := Effects[label]
	if !ok {
		return false
	}
	return p.SoundEffect(e)
}

// SoundEffect plays table-provided original SOUND STRUCTURES on channel four.
func (p *Player) SoundEffect(e Effect) bool {
	v := &p.Voices[3]
	*v = voice{centered: true, sample: e.Sample, period: periods[e.Note], pitch: periods[e.Note], volume: p.Module.Samples[e.Sample-1].Volume}
	p.setStep(v)
	return true
}

// Render returns interleaved left/right signed 16-bit LE at 48 kHz.
// Music uses Amiga routing: voices 0/3 left, 1/2 right. Runtime effects are
// centered on their existing voice; each output channel clips independently.
func (p *Player) Render(frames int) []byte {
	out := make([]byte, frames*BytesPerFrame)
	for f := 0; f < frames; f++ {
		if !p.started {
			p.row()
			p.started = true
		}
		if p.tickFrames == 0 {
			p.tickFrames = Rate / 50
		}
		sums := [Channels]int{}
		for i := range p.Voices {
			v := &p.Voices[i]
			if v.sample <= 0 || v.step == 0 {
				continue
			}
			s := &p.Module.Samples[v.sample-1]
			index := int(v.phase >> 32)
			end := len(s.PCM)
			if s.LoopLength > 2 {
				end = s.LoopStart + s.LoopLength
			}
			if index >= end {
				if s.LoopLength > 2 {
					index = s.LoopStart + (index-s.LoopStart)%s.LoopLength
					v.phase = uint64(index)<<32 | v.phase&0xffffffff
				} else {
					v.step = 0
					continue
				}
			}
			if index < len(s.PCM) {
				// Linear interpolation suppresses nearest-neighbor resampling steps.
				next := index + 1
				if next >= end && s.LoopLength > 2 {
					next = s.LoopStart
				}
				adjacent := 0
				if next < len(s.PCM) {
					adjacent = int(s.PCM[next])
				}
				base := int(s.PCM[index])
				channel := 0
				if i == 1 || i == 2 {
					channel = 1
				}
				value := base*v.volume + int((int64(adjacent-base)*int64(v.phase&0xffffffff)*int64(v.volume))>>32)
				sums[channel] += value
				if v.centered {
					sums[1-channel] += value
				}
			}
			v.phase += v.step
		}
		for channel, sum := range sums {
			sum = max(-32768, min(32767, sum))
			offset := f*BytesPerFrame + channel*2
			out[offset] = byte(sum)
			out[offset+1] = byte(sum >> 8)
		}
		p.Frames++
		p.tickFrames--
		if p.tickFrames == 0 {
			p.boundary()
		}
	}
	return out
}

// Sync renders exactly one 1/71 second interval, with retained rational remainder.
func (p *Player) Sync() []byte {
	p.syncRemainder += Rate
	n := p.syncRemainder / 71
	p.syncRemainder %= 71
	return p.Render(n)
}

// SOUND_EFFECT's explicit volume parameter is used by SPRINGUP (charge*64/32).
func (p *Player) SoundEffectVolume(e Effect, volume byte) bool {
	p.SoundEffect(e)
	if volume > 64 {
		volume = 64
	}
	p.Voices[3].volume = int(volume)
	return true
}
func (p *Player) SoundVolume(label string, volume byte) bool {
	e, ok := Effects[label]
	if !ok {
		return false
	}
	return p.SoundEffectVolume(e, volume)
}
