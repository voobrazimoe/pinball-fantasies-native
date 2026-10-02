package audio

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"pinballfantasies/internal/testinputs"
	"reflect"
	"strconv"
	"testing"
)

func module(t *testing.T) *Module {
	t.Helper()
	testinputs.Require(t, "../../TABLE1.MOD")
	b, e := os.ReadFile("../../TABLE1.MOD")
	if e != nil {
		t.Fatal(e)
	}
	m, e := Decode(b)
	if e != nil {
		t.Fatal(e)
	}
	return m
}
func TestModuleContent(t *testing.T) {
	var f struct {
		ModuleSHA256 string `json:"module_sha256"`
		Orders       []int
		Patterns     int
		Effects      map[string][]int
		Samples      []struct {
			Length, Fine, Volume int
			LoopStart            int `json:"loop_start"`
			LoopLength           int `json:"loop_length"`
			SHA256               string
		}
	}
	b, _ := os.ReadFile("../../analysis/pf5-audio-fixtures.json")
	if e := json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	m := module(t)
	if f.ModuleSHA256 != ModuleSHA256 || len(m.Patterns) != f.Patterns || !reflect.DeepEqual(m.Orders, f.Orders) {
		t.Fatal("module metadata")
	}
	found := map[string]map[int]bool{}
	for _, p := range m.Patterns {
		for _, r := range p {
			for _, n := range r {
				k := strconv.Itoa(n.Effect)
				if found[k] == nil {
					found[k] = map[int]bool{}
				}
				found[k][n.Param] = true
			}
		}
	}
	for k, v := range f.Effects {
		if len(found[k]) != len(v) {
			t.Fatalf("effect %s", k)
		}
		for _, p := range v {
			if !found[k][p] {
				t.Fatal(k, p)
			}
		}
	}
	for i, s := range m.Samples {
		want := f.Samples[i]
		raw := make([]byte, len(s.PCM))
		for j, v := range s.PCM {
			raw[j] = byte(v)
		}
		if len(s.PCM) != want.Length || s.Volume != want.Volume || s.Fine != want.Fine || s.LoopStart != want.LoopStart || s.LoopLength != want.LoopLength || fmt.Sprintf("%x", sha256.Sum256(raw)) != want.SHA256 {
			t.Fatalf("sample%d metadata/decode", i+1)
		}
	}
	if _, e := Decode(nil); e == nil {
		t.Fatal("unpinned input accepted")
	}
}
func TestAllCueTimingMatchesPF45(t *testing.T) {
	var f struct {
		Cues map[string]struct {
			TrackerTicks int `json:"tracker_ticks"`
			Next         int
		}
	}
	b, _ := os.ReadFile(testinputs.Generated(t, "reference_pf45.py", "TABLE1.MOD", "reference/original-dos-source/PLAND.ASM"))
	if e := json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	m := module(t)
	for order := 0; order < len(m.Orders); order++ {
		p := New(m)
		p.Force(order)
		cue := f.Cues[strconv.Itoa(order)]
		called := 0
		at := uint64(0)
		p.Jump = func(next int) int {
			called++
			at = p.Frames
			if next != cue.Next {
				t.Fatalf("order%d B target%d want%d", order, next, cue.Next)
			}
			return next
		}
		for tick := 0; tick < cue.TrackerTicks; tick++ {
			p.Render(Rate / 50)
			if tick < cue.TrackerTicks-1 && called != 0 {
				t.Fatalf("order%d early cue at tick%d", order, tick)
			}
		}
		if called != 1 || at != uint64(cue.TrackerTicks*Rate/50) {
			t.Fatalf("order%d cue=%d frame=%d", order, called, at)
		}
	}
}
func TestRowTicksEffectsAndBreak(t *testing.T) {
	m := &Module{Orders: []int{0, 1}, Patterns: make([][64][4]Note, 2)}
	m.Patterns[0][0][0] = Note{Effect: 15, Param: 3}
	m.Patterns[0][1][0] = Note{Effect: 13, Param: 0x12}
	m.Patterns[1][12][0] = Note{Effect: 15, Param: 2}
	p := New(m)
	p.Render(960)
	if p.Row != 0 || p.Tick != 1 || p.Speed != 3 {
		t.Fatal("F03", p)
	}
	p.Render(1920)
	if p.Row != 1 || p.Tick != 0 {
		t.Fatal("row", p)
	}
	p.Render(2880)
	if p.Order != 1 || p.Row != 12 || p.Tick != 0 || p.Speed != 2 {
		t.Fatal("D12", p)
	}
	p.Render(123)
	p.Force(0)
	if p.tickFrames != 0 || p.Tick != 0 || p.Row != 0 {
		t.Fatal("force convention")
	}
	// Reachable E92 / tone slide / vibrato+volume: ticks 1 onward only.
	v := &p.Voices[0]
	v.sample = 1
	v.period = 428
	v.pitch = 428
	v.volume = 32
	v.effect = 14
	v.param = 0x92
	v.phase = 100 << 32
	p.Tick = 2
	p.effects()
	if v.phase != 0 {
		t.Fatal("retrigger")
	}
	v.effect = 3
	v.target = 400
	v.porta = 4
	p.effects()
	if v.period != 424 {
		t.Fatal("portamento")
	}
	v.effect = 6
	v.vibSpeed = 1
	v.vibDepth = 4
	v.param = 1
	p.effects()
	if v.volume != 31 || v.vibPhase != 1 {
		t.Fatal("vibrato volume")
	}
}
func TestPCMAndMixer(t *testing.T) {
	m := module(t)
	p := New(m)
	p.Force(62)
	p.Render(1)
	if !p.Sound("SBUMPER1") {
		t.Fatal("effect")
	}
	pcm := p.Render(900)
	var f struct {
		Hash string `json:"bumper900_sha256"`
	}
	b, _ := os.ReadFile("../../analysis/pf5-audio-fixtures.json")
	json.Unmarshal(b, &f)
	if fmt.Sprintf("%x", sha256.Sum256(pcm)) != f.Hash {
		t.Fatal("independent bumper PCM sequence", fmt.Sprintf("%x", sha256.Sum256(pcm)), f.Hash)
	}
	if p.Sound("missing") {
		t.Fatal("unknown sound accepted")
	}
	// Two worst-case voices per output channel fit without wrapping.
	for _, value := range []int8{-128, 127} {
		synthetic := &Module{Orders: []int{0}, Patterns: make([][64][4]Note, 1)}
		synthetic.Samples[0] = Sample{PCM: []int8{value, value, value, value}, Volume: 64, LoopLength: 4}
		for c := 0; c < 4; c++ {
			synthetic.Patterns[0][0][c] = Note{Sample: 1, Period: 428}
		}
		q := New(synthetic)
		pcm := q.Render(1)
		for channel := 0; channel < Channels; channel++ {
			got := int16(binary.LittleEndian.Uint16(pcm[channel*2:]))
			if int(got) != int(value)*64*2 {
				t.Fatalf("channel%d mix wraps %d", channel, got)
			}
		}
	}
	// Chunking must not affect music, sample phase, row/tick or output bytes.
	a, c := New(m), New(m)
	a.Force(1)
	c.Force(1)
	want := a.Render(48000)
	var got []byte
	for i := 0; i < 71; i++ {
		got = append(got, c.Sync()...)
	}
	if !bytes.Equal(want, got) || !reflect.DeepEqual(a, c) {
		t.Fatal("render chunk nondeterminism")
	}
	for label := range Effects {
		q := New(m)
		q.Force(62)
		q.Render(1)
		q.Sound(label)
		r := New(m)
		r.Force(62)
		r.Render(1)
		r.Sound(label)
		if !bytes.Equal(q.Render(4800), r.Render(4800)) {
			t.Fatal("effect nondeterminism", label)
		}
	}
}

func TestAmigaStereoRoutingAndClipping(t *testing.T) {
	m := &Module{Orders: []int{0}, Patterns: make([][64][4]Note, 1)}
	m.Samples[0] = Sample{PCM: []int8{100, -100, 50, -50}, LoopLength: 4}
	for voiceIndex := 0; voiceIndex < 4; voiceIndex++ {
		p := New(m)
		p.Render(1)
		p.Voices[voiceIndex] = voice{sample: 1, volume: 64, step: 1 << 31}
		pcm := p.Render(4)
		if len(pcm) != 4*BytesPerFrame || p.Voices[voiceIndex].phase != 2<<32 {
			t.Fatal("stereo changed frame count/sample phase")
		}
		for frame, value := range []int{6400, 0, -6400, -1600} {
			for channel := 0; channel < Channels; channel++ {
				want := 0
				if (channel == 0) == (voiceIndex == 0 || voiceIndex == 3) {
					want = value
				}
				got := int16(binary.LittleEndian.Uint16(pcm[frame*BytesPerFrame+channel*2:]))
				if int(got) != want {
					t.Fatalf("voice%d frame%d channel%d got%d want%d", voiceIndex, frame, channel, got, want)
				}
			}
		}
	}
	// Deliberately overdrive one side to verify saturation cannot wrap or affect
	// the other side. Normal module volume is bounded to 64.
	for clippedChannel := 0; clippedChannel < Channels; clippedChannel++ {
		for _, sign := range []int{-1, 1} {
			p := New(m)
			p.Render(1)
			p.Voices[clippedChannel] = voice{sample: 1, volume: sign * 400, step: 1 << 32}
			p.Voices[1-clippedChannel] = voice{sample: 1, volume: -sign * 10, step: 1 << 32}
			pcm := p.Render(1)
			clipped := int16(binary.LittleEndian.Uint16(pcm[clippedChannel*2:]))
			other := int16(binary.LittleEndian.Uint16(pcm[(1-clippedChannel)*2:]))
			want := 32767
			if sign < 0 {
				want = -32768
			}
			if int(clipped) != want || int(other) != -sign*1000 {
				t.Fatalf("channel%d independent clipping: got%d other%d", clippedChannel, clipped, other)
			}
		}
	}
}

func TestSampleLoopAndEnd(t *testing.T) {
	for _, loop := range []bool{false, true} {
		m := &Module{Orders: []int{0}, Patterns: make([][64][4]Note, 1)}
		m.Samples[0] = Sample{PCM: []int8{1, 2, 3, 4}, Volume: 64}
		if loop {
			m.Samples[0].LoopStart = 1
			m.Samples[0].LoopLength = 3
		}
		p := New(m)
		p.Render(1)
		p.Voices[0] = voice{sample: 1, volume: 64, step: 1 << 32}
		pcm := p.Render(10)
		for i := 0; i < 10; i++ {
			want := 0
			if i < 4 {
				want = i + 1
			} else if loop {
				want = 2 + (i-4)%3
			}
			got := int16(binary.LittleEndian.Uint16(pcm[BytesPerFrame*i:]))
			if int(got) != want*64 {
				t.Fatalf("loop=%v frame%d got%d want%d", loop, i, got, want*64)
			}
		}
	}
}

func TestEffectAndTrackerShareFourthVoice(t *testing.T) {
	m := module(t)
	p := New(m)
	p.Force(1)
	p.Render(1)
	firstThree := [3]voice{p.Voices[0], p.Voices[1], p.Voices[2]}
	p.Sound("SBUMPER1")
	if firstThree != [3]voice{p.Voices[0], p.Voices[1], p.Voices[2]} {
		t.Fatal("effect interrupted music on another channel")
	}
	if p.Voices[3].sample != 24 || p.Voices[3].period != 202 {
		t.Fatal("source effect channel/pitch")
	}
	p.Sound("SKICKER")
	if p.Voices[3].sample != 29 || p.Voices[3].phase != 0 {
		t.Fatal("ordinary effects must replace, not queue")
	}
	// A later music note reclaims the same voice, rather than leaving an extra lane.
	m2 := &Module{Orders: []int{0}, Patterns: make([][64][4]Note, 1), Samples: m.Samples}
	m2.Patterns[0][0][0] = Note{Effect: 15, Param: 1}
	m2.Patterns[0][1][3] = Note{Sample: 1, Period: 428}
	q := New(m2)
	q.Render(1)
	q.Sound("SBUMPER1")
	q.Render(959)
	if q.Voices[3].sample != 1 || q.Voices[3].phase != 0 {
		t.Fatal("tracker did not reclaim effect channel")
	}
}

func TestRuntimeEffectsAreCentered(t *testing.T) {
	m := &Module{Orders: []int{0}, Patterns: make([][64][4]Note, 1)}
	m.Samples[0] = Sample{PCM: []int8{100, -100, 50, -50}, Volume: 64, LoopLength: 4}
	p := New(m)
	p.Render(1)
	p.SoundEffect(Effect{Sample: 1, Note: 12})
	step := p.Voices[3].step
	pcm := p.Render(100)
	audible := false
	for frame := 0; frame < 100; frame++ {
		left := int16(binary.LittleEndian.Uint16(pcm[frame*BytesPerFrame:]))
		right := int16(binary.LittleEndian.Uint16(pcm[frame*BytesPerFrame+2:]))
		if left != right {
			t.Fatal("effect is not centered", frame, left, right)
		}
		audible = audible || left != 0
	}
	if !audible || p.Voices[3].phase&0xffffffff != (100*step)&0xffffffff {
		t.Fatal("effect phase or sound changed")
	}
	// A tracker note reclaims the existing effect voice and its Amiga routing.
	m.Patterns[0][1][3] = Note{Sample: 1, Period: 428}
	p.Row = 1
	p.row()
	pcm = p.Render(1)
	if p.Voices[3].centered || binary.LittleEndian.Uint16(pcm) == 0 || binary.LittleEndian.Uint16(pcm[2:]) != 0 {
		t.Fatal("tracker did not restore voice3 left routing")
	}
}
