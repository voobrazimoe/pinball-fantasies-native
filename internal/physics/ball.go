package physics

import (
	"fmt"
	"pinballfantasies/internal/gameplay"
	"pinballfantasies/internal/settings"
)

type EventKind uint8

const (
	EventBumperHit EventKind = iota + 1
	EventSlingshotHit
	EventFlipperHit
	EventTargetHit
	EventDrain
)

type Event struct {
	Kind   EventKind
	Object int
	X, Y   int16
}
type Ball struct {
	X, Y                                     int32
	VX, VY, GX, GY, PixelX, PixelY, Rotation int16
	High, Hold, Lost                         bool
	HitX, HitY                               int16
	CollisionAngle                           uint16
	ContactCount, Material                   uint8
}
type Inputs = gameplay.Controls
type Game struct {
	Configured   bool
	TiltDisabled bool  // FANTASIE/TILTDISABLED; physical TILT0 push still runs.
	FastBall     bool  // FAIRPLAY clears SHIFTKEYS bit2; HI_RES initially sets it.
	TargetRaster int16 // SCREENFORCE2: desired viewport, still smoothed

	Settings      settings.Config
	ReferenceMode byte // Historical DOS palette reference; never a native user setting.
	gravity       [][2]int16
	Table         *Table
	Ball          Ball
	Flippers      [3]Flipper
	mask12        []byte
	mask22        []byte
	// Optional Party Land consumer, called at the original callback boundaries.
	OnEvent                                   func(Event)
	BeforeLate                                func()       // PF4.5 matrix work finishes VBLANK before late-raster physics.
	ScrollForce                               func() int16 // nil preserves the PF3 camera exactly.
	BeforeTargets                             func()
	AfterTargets                              func(Inputs)
	Raster                                    int16 // original RASTERPOS, 1/16 scanline including SPLH
	Events                                    []Event
	Syncs                                     uint64
	Stopped                                   bool
	pendingBumper                             bool
	pendingEvent                              Event
	indices                                   []byte
	graphicsFrame                             [3]int16
	SpringValid                               bool
	AllowFlip, Tilted, tiltLatched            bool
	TiltCounter                               uint16
	ScreenPosition, ScreenSpeed, ScreenOffset int16
	SpringPosition                            uint8
	SpringGraphics                            []byte
}

func New(t *Table) *Game {
	g := &Game{Table: t, TargetRaster: -1, Settings: settings.Legacy(), gravity: append([][2]int16(nil), t.gravity...), Flippers: t.Flippers, mask12: append([]byte(nil), t.mask12...), mask22: append([]byte(nil), t.mask22...), Raster: (259 + 33) * 16}
	g.indices = append([]byte(nil), t.Initial.Playfield.Indices...)
	g.SpringValid = true
	g.AllowFlip = true
	x, y := int16(t.start.X), int16(t.start.Y)
	g.Ball = Ball{X: int32(x) * 1024, Y: int32(y) * 1024, VX: 10, GY: 8, PixelX: x, PixelY: y}
	for i := range g.Flippers {
		g.copyFlipper(i)
	}
	return g
}

// Release is the high-resolution SPRINGUP arithmetic. Charge is the original
// springpos byte (0..32); jitter is SLUMP_COUNTERN's low eight bits, supplied
// explicitly so replay never depends on host timing or a random generator.
func (g *Game) Release(charge, jitter uint8) {
	if charge == 0 || charge > 32 || g.Ball.Lost || !g.SpringValid {
		return
	}
	g.Ball.VX = 0
	g.Ball.VY = -166*int16(charge) - int16(jitter)
	g.Ball.Rotation = int16(jitter & 15)
	// SPRING_VALID is initially true; later table-rule callbacks are events.
}

// Sync reproduces a high-resolution gameplay interrupt pair: VBLANK executes
// two sc_program passes, then DO_PHYSICS; late raster executes one pass.
// One step always checks collision, resolves it, updates flipper state, moves
// the ball, then copies any overlapping flipper mask.
func (g *Game) Sync(input Inputs) error {
	g.Events = g.Events[:0]
	if g.Stopped {
		return nil
	}
	for i := 0; i < 2; i++ {
		if err := g.step(input); err != nil {
			return err
		}
	}
	if g.pendingBumper {
		g.Events = append(g.Events, g.pendingEvent)
		g.pendingBumper = false
		if g.OnEvent != nil {
			g.OnEvent(g.pendingEvent)
		}
	}
	if g.TiltCounter > 0 {
		g.TiltCounter--
	}
	g.checkRamps()
	g.checkLevels()
	if g.Ball.Lost {
		g.Stopped = true
		e := Event{Kind: EventDrain, X: g.Ball.PixelX, Y: g.Ball.PixelY}
		g.Events = append(g.Events, e)
		if g.OnEvent != nil {
			g.OnEvent(e)
		}
		return nil
	}
	if g.BeforeTargets != nil {
		g.BeforeTargets()
	}
	g.checkSpringAndTargets()
	if g.AfterTargets != nil {
		g.AfterTargets(input)
	}
	if g.BeforeLate != nil {
		g.BeforeLate()
	}
	g.scroll()
	lateSteps := 1
	if g.FastBall {
		lateSteps = 2
	}
	for i := 0; i < lateSteps; i++ {
		if err := g.step(input); err != nil {
			return err
		}
	}
	g.Syncs++
	return nil
}

func (g *Game) step(input Inputs) error {
	if !g.Ball.Hold {
		if c, ok := g.collision(); ok {
			if err := g.respond(c); err != nil {
				return err
			}
		}
	}
	g.push(input.Tilt)
	g.moveFlippers(input)
	if !g.Ball.Hold {
		b := &g.Ball
		b.Y += int32(b.VY)
		y, err := divide(b.Y, 1024)
		if err != nil {
			return err
		}
		b.PixelY = y
		if b.PixelY >= 576 {
			b.Lost = true
		}
		b.X += int32(b.VX)
		x, err := divide(b.X, 1024)
		if err != nil {
			return err
		}
		b.PixelX = x
		b.VY += b.GY
		b.VX += b.GX
		if b.Rotation > 0 {
			b.Rotation -= 2
			if b.Rotation <= 0 {
				b.Rotation = 0
			}
		} else if b.Rotation < 0 {
			b.Rotation += 2
			if b.Rotation >= 0 {
				b.Rotation = 0
			}
		}
	}
	for i, f := range g.Flippers {
		if f.Bounds.contains(g.Ball.PixelX, g.Ball.PixelY) {
			g.copyFlipper(i)
		}
	}
	return nil
}

func divide(n int32, d int16) (int16, error) {
	if d == 0 {
		return 0, fmt.Errorf("original physics IDIV divisor zero")
	}
	q := int64(n) / int64(d)
	if q < -32768 || q > 32767 {
		return 0, fmt.Errorf("original physics IDIV quotient overflow: %d / %d", n, d)
	}
	return int16(q), nil
}
func clamp(v int16) int16 {
	if v < -4100 {
		return -4100
	}
	if v > 4100 {
		return 4100
	}
	return v
}
func highProduct(a, b int16) int16 { return int16((int32(a) * int32(b)) >> 16) }

func (g *Game) moveFlippers(input Inputs) {
	for i := range g.Flippers {
		f := &g.Flippers[i]
		up := g.AllowFlip && ((f.Kind == 2 && input.Left) || (f.Kind == 1 && input.Right))
		if up {
			f.Speed += f.AccelerationUp
			// Original JLE retains speeds already at/below the negative bound;
			// speeds above it jump straight to MaxSpeedUp. Do not modernize it.
			if f.Speed > f.MaxSpeedUp {
				f.Speed = f.MaxSpeedUp
			}
		} else {
			f.Speed += f.AccelerationDown
		}
		f.Angle -= f.Speed
		// ADJUSTTABLE: floor(angle/55), populated for 24*55 entries.
		// Downward overshoot is represented by a negative angle; the original
		// adjacent initialized zero words make that lookup zero at rest.
		frame := int16(0)
		if f.Angle > 0 {
			frame = f.Angle / 55
		}
		if frame == 0 {
			f.Speed = 0
			f.Angle = 0
		}
		if frame >= f.Frames {
			f.Speed = 0
			f.Angle = f.MaxAngle
			frame = f.Frames
		}
		f.Frame = frame
	}
}
func (g *Game) copyFlipper(i int) {
	f := g.Flippers[i]
	start := int(f.Frame) * f.stride
	for y := 0; y < int(f.Height); y++ {
		dst := (int(f.Top)+y)*40 + int(f.Left)/8
		src := start + y*int(f.Words)*2
		copy(g.mask12[dst:dst+int(f.Words)*2], f.masks[src:src+int(f.Words)*2])
	}
}

func (g *Game) checkRamps() {
	b := &g.Ball
	si := int(uint16(b.PixelY+8+g.ScreenOffset))*40 + int(uint16(b.PixelX+8)>>3) - 1
	for n := 0; n < 3; n++ {
		i := si + n
		if i < 0 || i >= 23040 {
			continue
		}
		var occupied byte
		var ramp uint8
		if b.High {
			if i >= len(g.Table.mask21) {
				continue
			}
			occupied = g.Table.mask21[i] | g.mask22[i]
			ramp = g.Table.mask23[i] & 15
		} else {
			occupied = g.Table.mask11[i] | g.mask12[i]
			ramp = g.Table.mask13[i] & 15
		}
		if occupied == 0 {
			if int(ramp) < len(g.gravity) {
				b.GX = g.gravity[ramp][0]
				b.GY = g.gravity[ramp][1]
			}
			return
		}
	}
}
func (g *Game) checkLevels() {
	index := 0
	if g.Ball.High {
		index = 1
	}
	for _, r := range g.Table.levels[index] {
		if r.contains(g.Ball.PixelX+8, g.Ball.PixelY+8+g.ScreenOffset) {
			g.Ball.High = !g.Ball.High
			return
		}
	}
}
func (g *Game) scroll() {
	if g.ScrollForce != nil {
		force := g.ScrollForce()
		if force >= 0 {
			g.Raster = (force + 33) * 16
			return
		}
	}
	middle := int16(g.Settings.FieldHeight() / 2)
	target := g.Ball.PixelY - (middle - 28)
	if target < 0 {
		target = 0
	}
	if target > g.BottomRaster() {
		target = g.BottomRaster()
	}
	if g.TargetRaster >= 0 {
		target = g.TargetRaster
	}
	target += 33
	delta := (target - (g.Raster >> 4)) * g.Settings.ScrollFactor()
	g.Raster += delta >> 2
	gap := target - (g.Raster >> 4)
	if gap < 0 {
		gap += middle - 28
		if gap <= 0 {
			g.Raster += gap << 4
		}
	} else {
		gap -= middle + 12
		if gap >= 0 {
			g.Raster += gap << 4
		}
	}
}

// Only the physical plunger-enable writes of BYGEL12/BYGEL28 are retained.
// CHECK_TARGETS emits the source zone identity instead of scoring or target
// rule callbacks. Those callbacks do not run inside collision response.
func (g *Game) checkSpringAndTargets() {
	b := &g.Ball
	if g.Tilted {
		return
	}
	if !b.High {
		x, y := b.PixelX+8, b.PixelY+8
		if g.Table.springInvalid.contains(x, y) {
			g.SpringValid = false
		} else if g.Table.springValid.contains(x, y) {
			g.SpringValid = true
		}
	}
	hx, hy := b.HitX, b.HitY+g.ScreenOffset
	b.HitX = 0
	b.HitY = 0
	if b.High || (hx == 0 && hy == 0) {
		return
	}
	for i, r := range g.Table.targets {
		if r.contains(hx, hy) {
			e := Event{Kind: EventTargetHit, Object: i, X: hx, Y: hy}
			g.Events = append(g.Events, e)
			if g.OnEvent != nil {
				g.OnEvent(e)
			}
			return
		}
	}
}
