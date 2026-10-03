// Package presentation renders original DATA programs into the VGA split region.
// Timing and priority belong to the table scheduler; this is a pixel consumer.
package presentation

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/tablelogic"
	"strconv"
	"strings"
)

const Width = 320
const FieldHeight = 317
const MatrixHeight = 33
const FrameHeight = 350
const MemoryWidth = 336
const DotWidth = 160
const DotHeight = 16

type Animation struct {
	Header, Durations []uint16
	Offsets           []int
}
type Command struct {
	Op   string
	Args []string
	// Nums holds the integer each source-language numeric operand resolves to,
	// keyed by operand index. It is produced by the ASM extractor: the raw text
	// stays in Args for diagnostics and identity lookups, while gameplay reads
	// only Nums. A missing entry is an invariant violation, not a fallback.
	Nums map[int]int `json:"nums,omitempty"`
}

// Arg returns operand i, or "" when the source command has no such operand.
func (c Command) Arg(i int) string {
	if i < 0 || i >= len(c.Args) {
		return ""
	}
	return c.Args[i]
}

// Num returns the source-resolved integer for numeric operand i. Missing
// values panic: extraction must have resolved every reachable numeric operand
// from the original assembler expression. No default and no re-parsing.
func (c Command) Num(i int) int {
	v, ok := c.Nums[i]
	if !ok {
		panic(fmt.Sprintf("unresolved source operand %s[%d]=%q", c.Op, i, c.Arg(i)))
	}
	return v
}

// WordWaitTicks represents WAITRUT's wrapping SI word as a visit count.
func WordWaitTicks(n int) int {
	word := uint16(n)
	if word == 0 {
		return 65536
	}
	return int(word)
}

type Content struct {
	TextRefs                 map[string][2]int `json:"text_refs"`
	AnimationRefs            map[string][4]int `json:"animation_refs"`
	LampFlashRef             [2]int            `json:"lamp_flash_ref"`
	Texts                    map[string][]byte
	Animations               map[string]Animation
	Commands                 []Command
	Attract                  []Command
	LampFlash                [][4]int
	Labels, Positions, Fonts map[string]int
	ScoreFont, ScrollFont    map[string][]byte
	Spring                   int
	SHA256                   string
}

var contents = func() map[string]Content {
	var v map[string]Content
	if e := json.Unmarshal([]byte(contentJSON), &v); e != nil {
		panic(e)
	}
	return v
}()

type Display struct {
	Content                Content
	table                  int
	data                   []byte
	Dots                   [DotWidth * DotHeight]bool
	On                     bool
	NoSound                bool // Original INT66 function21 bit3, distinct from muted PCM.
	scoreCache             [12]byte
	scoreCommas            int
	jingleCountdown        uint16
	countdown              countdown
	flashSpeed, flashCount uint16
	flashing, flashPhase   bool
	idlePanelPending       bool
	pendingPrint           *Command
	scrollPhase            uint8
	scrollChar             int
	scrollBase             [2]int
	scrollPlane            [2]int
	args                   []string
	nums                   map[int]int
	op                     string
	elapsed                int
	scrollOffset           int
	anim                   string
	sourceMatrixOff        *[3]byte
}

func New(table int, data []byte) *Display {
	c := contents[strconv.Itoa(table)]
	c.ScoreFont, c.ScrollFont = runtimeFonts(table, data)
	texts := make(map[string][]byte, len(c.Texts))
	for k, v := range c.Texts {
		texts[k] = append([]byte{}, v...)
	}
	c.Texts = texts
	c.decodeRecords(data)
	d := &Display{Content: c, table: table, data: data, On: true, scrollPhase: 8, scrollPlane: [2]int{0, 1}}
	d.InvalidateScore()
	return d
}

// CarryMemory retains VGA dots and SCROLLE's self-modifying phase/selectors
// when frontend ownership hands the same loaded table back to gameplay.
func (d *Display) CarryMemory(previous *Display) {
	d.Dots, d.On = previous.Dots, previous.On
	d.scrollPhase, d.scrollPlane = previous.scrollPhase, previous.scrollPlane
}

func (d *Display) Clear() { d.Dots = [DotWidth * DotHeight]bool{} }

// ShowPlayerBall commits the idle panel at a player/new-ball boundary, before
// launch. Replay this table's static SHOWPLAYERSTS prints and use the existing
// CODE2 score renderer; do not run its transition clear over subsequent ticks.
// A copy keeps the live command, animation and elapsed state untouched.
// The idle panel ends the outgoing illumination effect, like source
// DO_THE_ANIMATIONS/ts_slut calling KILL_FLASHOR (MATRIXON, flashing off).
func (d *Display) ShowPlayerBall(score string) {
	start, ok := d.Content.Labels["SHOWPLAYERSTS"]
	if !ok {
		panic("missing source player/ball panel")
	}
	first := d.Content.Commands[start]
	if first.Op != "_CLEAR1" && first.Op != "_CLEAR4" && !(first.Op == "_ANIMATION" && first.Arg(0) == "_CLEAR") {
		panic("unexpected source player/ball transition")
	}
	panel := *d
	panel.InvalidateScore() // DO_MATRIX / DO_SPEC_MATRIX at panel entry.
	panel.Clear()
	for _, c := range d.Content.Commands[start+1:] {
		if c.Op == "0" {
			panel.Score(score)
			d.Dots = panel.Dots
			d.KillFlash()
			return
		}
		if !strings.HasPrefix(c.Op, "_PRINT") {
			panic("non-static source player/ball panel")
		}
		panel.BeginCommand(c)
		panel.Visit(0, 0, func(label string) string {
			if label != "SIFFRORNA" {
				panic("unexpected player/ball number " + label)
			}
			return score
		})
	}
	panic("unterminated source player/ball panel")
}

// MatchStart/MatchStep are KNACKRUT's drawing operations. The table retains
// the 11/13-sync cadence, random counter, awards and completion authority.
func (d *Display) MatchStart(scoreDigit byte) {
	d.MatchPlayers([]byte{scoreDigit}, nil)
}

// MatchWin is gladgnu's MATRIX_SPEED/MATRIX_CNT=3 and flashlast repaint.
func (d *Display) MatchWin(scoreDigit byte) {
	d.MatchPlayers([]byte{scoreDigit}, &scoreDigit)
}
func (d *Display) MatchStep(previous, next uint16) {
	d.Text("*", int(previous)*16, 7, 5)
	d.Content.Texts["SLUTSIFFRA"] = []byte{byte(next) + '7', 0}
	d.pendingPrint = &Command{Op: "_PRINT5", Args: []string{"SLUTSIFFRA", strconv.Itoa(1344 + int(next)*8)}, Nums: map[int]int{1: 1344 + int(next)*8}}
}
func (d *Display) Text(s string, x, y, height int) {
	font := d.Content.Fonts[strconv.Itoa(height)]
	for _, c := range []byte(s) {
		index := -1
		switch {
		case c >= '0' && c <= '9':
			index = int(c - '0')
		case c >= 'A' && c <= 'Z':
			index = int(c-'A') + 10
		}
		switch c {
		case '?':
			index = 36
		case '[', '(':
			index = 37
		case ']', ')':
			index = 38
		case '-':
			index = 39
		case '.':
			index = 40
		case '`':
			index = 41
		}
		// Original PRINT_TEXT's numeric data uses '7'+digit; regular ASCII digits
		// enter only via formatted native BCD values and frontend player substitution.
		for row := 0; row < height; row++ {
			var bits byte
			if index >= 0 && (index < 36 || height == 13) {
				bits = d.data[font+index*height+row]
			}
			for col := 0; col < 8; col++ {
				xx, yy := x+col, y+row
				if xx >= 0 && xx < DotWidth && yy >= 0 && yy < DotHeight {
					d.Dots[yy*DotWidth+xx] = bits&(128>>uint(col)) != 0
				}
			}
		}
		x += 8
	}
}
func (d *Display) SourceText(label string) string {
	b := append([]byte(nil), d.Content.Texts[strings.ToUpper(label)]...)
	for i, c := range b {
		if c >= '7' && c <= '@' {
			b[i] = '0' + c - '7'
		} else if c >= '[' && c <= '`' {
			b[i] = []byte{'?', '[', ']', '-', '.', '`'}[c-'[']
		}

	}
	return string(b)
}

// PositionValue maps an already-resolved source matrix position to a cell.
// Row/column arithmetic is the original two-byte-strip geometry.
func PositionValue(a int) (int, int) {
	row := (a + 84) / 168
	return (a - row*168) * 2, row - 1
}

// Position resolves a raw source position expression through the generated
// provenance map. It exists for synthetic frontend replay programs and
// diagnostics; gameplay uses the extractor-resolved integer instead.
func (d *Display) Position(expr string) (int, int) {
	a, ok := d.Content.Positions[expr]
	if !ok {
		panic("unknown matrix position " + expr)
	}
	return PositionValue(a)
}
func (d *Display) Bitmap(offset int) {
	q := offset
	for plane := 0; plane < 2; plane++ {
		count := int(binary.LittleEndian.Uint16(d.data[q:]))
		q += 2
		di := 167
		for _, v := range d.data[q : q+count] {
			di += int(v >> 1)
			if v>>1 == 127 && v&1 == 0 {
				continue
			}
			x := (di%84)*2 + plane
			y := di/168 - 1
			if x >= 0 && x < 160 && y >= 0 && y < 16 {
				d.Dots[y*160+x] = v&1 != 0
			}
		}
		q += count
	}
}

// Begin records a command, but deferred drawing occurs on its first scheduler
// visit, matching PRINTTASK / DOTRUT. Begin never supplies a second clock.
// Synthetic replay programs without extractor data use this form.
func (d *Display) Begin(op string, args []string) {
	// Hand-built frontend replay commands resolve their operands at construction.
	// Extracted gameplay commands always enter through BeginCommand/Resolved.
	nums := map[int]int{}
	if strings.HasPrefix(op, "_PRINT") {
		nums[1] = d.Content.Positions[args[1]]
		if _, ok := d.Content.Positions[args[1]]; !ok {
			panic("unknown replay position " + args[1])
		}
	}
	if op == "_FLASHON" {
		if len(args) != 1 || args[0] != "1" {
			panic("untyped replay flash")
		}
		nums[0] = 1
	}
	if op == "_MATRIXLGT" {
		if len(args) != 1 {
			panic("untyped replay matrix light")
		}
		switch args[0] {
		case "0":
			nums[0] = 0
		case "1":
			nums[0] = 1
		default:
			panic("untyped replay matrix light")
		}
	}
	d.begin(op, args, nums)
}

// BeginCommand records an extracted source command together with its resolved
// numeric operands. This is the gameplay path.
func (d *Display) BeginCommand(c Command) { d.begin(c.Op, c.Args, c.Nums) }

// BeginResolved records a command whose numeric operands were resolved by a
// table-local extractor with its own command type.
func (d *Display) BeginResolved(op string, args []string, nums map[int]int) {
	d.begin(op, args, nums)
}

func (d *Display) begin(op string, args []string, nums map[int]int) {
	if op != "0" {
		d.InvalidateScore() // DO_SPEC_MATRIX invokes UPDAT_SCORE before each handler.
	}
	d.op = op
	d.args = args
	d.nums = nums
	d.elapsed = 0
	d.scrollOffset = 0
	if op == "_SCROLL" {
		d.scrollChar = 0
		d.scrollBase = [2]int{168, 168} // STARTSCROLL resets both bases, not selectors/phase.
	}
	if strings.HasPrefix(op, "_PRINT") {
		d.pendingPrint = &Command{Op: op, Args: args, Nums: nums}
	}
	d.anim = ""
	// FANTASIE command handlers write illumination at dispatch, before WAITRUT.
	switch op {
	case "_FLASHON":
		d.startFlash(uint16(d.Num(0)))
	case "_FLASHOFF":
		d.KillFlash()
	case "_MATRIXLGT":
		d.On = d.Num(0) != 0
	case "_SETDECCOR":
		d.SetJingleCountdown(uint16(d.Num(0)))
	}
	if op == "_ANIMATION" {
		if _, ok := d.Content.Animations[args[0]]; !ok {
			panic("unknown source matrix animation " + args[0])
		}
		d.anim = args[0]
	}
}

// Num requires a typed operand, including for presentation replay.
func (d *Display) Num(i int) int {
	return (Command{Op: d.op, Args: d.args, Nums: d.nums}).Num(i)
}
func (d *Display) positionAt(i int, expr string) (int, int) {
	return PositionValue(d.Num(i))
}

// MutableText copies a source DATA buffer before each write. Each display owns
// its text map; extracted byte slices remain pristine across sessions.
func (d *Display) MutableText(label string, minimum int) []byte {
	cur, ok := d.Content.Texts[label]
	if !ok || len(cur) < minimum {
		panic("missing source text buffer " + label)
	}
	buf := append([]byte(nil), cur...)
	d.Content.Texts[label] = buf
	return buf
}

// WriteBonusMultiplier is SHOW/STONES _BONUS_X_CALCS, NO_X_BONUS.
// Both bytes are rewritten every time calculation executes, even at X1.
func (d *Display) WriteBonusMultiplier(multiplier uint8) {
	buf := d.MutableText("BONUS_X_TEXT", 10)
	buf[8], buf[9] = '8', '7'
	if multiplier < 10 {
		buf[8], buf[9] = multiplier+'7', ' '
	}
}

// FlushPrint is DO_THE_DOTMATRIX's PRINTTASK call after DO_THE_ANIMATIONS.
// Dispatch replaces the single task slot; it does not draw on an external
// gameplay request. A same-sync tail dispatch may replace an earlier print.
func (d *Display) FlushPrint(number func(string) string) {
	if d.pendingPrint == nil {
		return
	}
	c := *d.pendingPrint
	d.pendingPrint = nil
	print := *d
	print.op, print.args, print.nums, print.elapsed = c.Op, c.Args, c.Nums, 0
	print.Visit(0, 0, number)
	d.Dots = print.Dots
}

func (d *Display) Visit(frame uint16, scrollAdvance int, number func(string) string) {
	d.elapsed++
	a := d.args
	op := d.op
	switch {
	case strings.HasPrefix(op, "_PRINT") && d.elapsed == 1:
		h := 13
		for _, n := range []int{5, 8, 11, 13} {
			if strings.HasPrefix(op, fmt.Sprintf("_PRINT%d", n)) {
				h = n
			}
		}
		x, y := d.positionAt(1, a[1])
		s := strings.TrimRight(d.SourceText(a[0]), "\x00")
		if strings.Contains(op, "NUMBER") {
			d.printNumber(number(a[0]), x, y, h, strings.HasSuffix(op, "_CENT"))
			break
		}
		d.Text(s, x, y, h)
	case op == "_NUMBER":
		d.ScoreAt(number(a[0]), 80)
		d.InvalidateScore() // SIFFRORRUT invokes UPDAT_SCORE after CODE2.
	case op == "_SHOW_SCORE":
		d.Text(strings.TrimRight(d.SourceText("PLAYERSTEXT"), "\x00"), 0, 1, 5)
		d.Score(number("SIFFRORNA"))
	case op == "_SCROLL":
		// SCROLLE executes twice per original visit; two source pixels per sync.
		d.scrollOffset += scrollAdvance
		s := strings.TrimRight(string(d.Content.Texts[a[0]]), "\xff\x00")
		d.Clear()
		d.GlyphText(s, -d.scrollOffset, 0, d.Content.ScrollFont)
	case op == "_RULLGARDIN_UPP" || op == "_RULLGARDIN_NED":
		stop := d.Num(1)
		y := 16 - d.elapsed
		if op == "_RULLGARDIN_NED" {
			y = -13 + d.elapsed
		}
		if op == "_RULLGARDIN_UPP" && y < stop || op == "_RULLGARDIN_NED" && y > stop {
			y = stop
		}
		// RULLGARDIN only erases the one trailing row; unrelated pixels
		// outside the moving glyph rectangle retain their source contents.
		row := y + 13
		if op == "_RULLGARDIN_NED" {
			row = y - 1
		}
		if row >= 0 && row < 16 {
			for x := 0; x < 160; x++ {
				d.Dots[row*160+x] = false
			}
		}
		d.Text(strings.TrimRight(d.SourceText(a[0]), "\x00"), 0, y, 13)
	case op == "_ANIMATION":
		anim := d.Content.Animations[d.anim]
		i := int(frame) / 4
		if i < 0 {
			i = 0
		}
		if i >= len(anim.Offsets) {
			i = len(anim.Offsets) - 1
		}
		if i >= 0 {
			d.Bitmap(anim.Offsets[i])
		}
	case op == "_CLEAR1":
		if d.elapsed == 1 {
			d.Clear()
		}
	case op == "_CLEAR4":
		y := d.elapsed - 1
		if y < 4 {
			for r := y; r < 16; r += 4 {
				for x := 0; x < 160; x++ {
					d.Dots[r*160+x] = false
				}
			}
		}
	case op == "_CLEAR2":
		if d.elapsed <= 16 {
			for x := 0; x < 160; x++ {
				d.Dots[(d.elapsed-1)*160+x] = false
			}
		}
	case op == "_CLEAR3":
		if d.elapsed <= 80 {
			for y := 0; y < 16; y++ {
				d.Dots[y*160+(d.elapsed-1)*2] = false
				d.Dots[y*160+(d.elapsed-1)*2+1] = false
			}
		}

	}
}

// KillFlash is FANTASIE KILL_FLASHOR: stop the blink routine and emit
// MATRIXON. The speed/count/phase words are deliberately left untouched.
func (d *Display) KillFlash() { d.flashing = false; d.On = true }

func (d *Display) SetJingleCountdown(n uint16) { d.jingleCountdown = n }

// JingleDone is WAITJINGLE/WAITJINGLE2. Muting samples does not select NOSOUND;
// the alternate branch is explicitly the original sound-driver capability.
func (d *Display) JingleDone(ready, alternate bool) bool {
	if !alternate || !d.NoSound {
		return ready
	}
	d.jingleCountdown--
	return d.jingleCountdown == 0
}

// StartMatrix is DO_MATRIX before its DO_SPEC_MATRIX tail call: KILL_FLASHOR
// and BEHOVS_PROVAD=true. DO_SPEC_MATRIX itself performs neither write.
func (d *Display) StartMatrix() {
	d.KillFlash()
	d.idlePanelPending = true
}

// TakeIdlePanel is NODOT's first BEHOVS_PROVAD visit outside I_UTSKJUT.
// It requests SHOWPLAYERSTS via DO_SPEC_MATRIX and also renders score on
// that visit. The installed clear routine first runs on the next sync.
func (d *Display) TakeIdlePanel(inChute bool) bool {
	if inChute || !d.idlePanelPending {
		return false
	}
	d.idlePanelPending = false
	return true
}

// FinishRoutine is DO_THE_ANIMATIONS after DOTRUT returns SI=0. A zero
// NEXT_A invokes ts_slut/KILL_FLASHOR; a following command inherits flash.
// The zero data word itself is not a command handler and performs no writes.
func (d *Display) FinishRoutine(hasNext bool) {
	if !hasNext {
		d.KillFlash()
	}
}

func (d *Display) startFlash(speed uint16) {
	d.flashSpeed, d.flashCount = speed, speed
	d.flashing, d.flashPhase = true, true
	// _FLASHON writes MATRIX_ONOFF, but emits no palette packet.
}
func (d *Display) Flash() {
	if d.flashing {
		d.flashCount--
		if d.flashCount == 0 {
			d.flashCount = d.flashSpeed
			d.flashPhase = !d.flashPhase
			d.On = d.flashPhase
		}
	}
}

// StepAnimation follows the linked ANIM routine (TABLE1.PRG 72be..738f).
// It draws old BX only when the timer expires, before storing the next index.
// The length boundary ends without drawing; loop restart preserves old BX.
func (d *Display) StepAnimation(a Animation, frame, loops, timer *uint16) bool {
	if *timer == 1 && !(*frame == a.Header[2] && *loops == 1) {
		d.Bitmap(a.Offsets[int(*frame)/4])
	}
	return tablelogic.Animation(a.Header, a.Durations, frame, loops, timer)
}

// StepScroll is linked SCROLLE (TABLE1.PRG 7197..729e): two calls per
// sync, test SI+20 before drawing, then shift/advance. The phase byte is shared
// across replacement streams. A terminator visit preserves the final pixels.
func (d *Display) StepScroll(left *uint16) bool {
	for i := 0; i < 2; i++ {
		if *left == 0 {
			return true
		}
		d.scrollStores()
		d.scrollOffset++
		d.scrollPlane[0] ^= 1
		d.scrollPlane[1] ^= 1
		if d.scrollPlane[1] == 0 {
			d.scrollBase[0]--
		} else {
			d.scrollBase[1]--
		}
		d.scrollPhase--
		if d.scrollPhase == 0 {
			d.scrollPhase = 8
			*left--
			d.scrollChar++
			d.scrollBase = [2]int{168, 168}
		}
	}
	return false
}

func (d *Display) Draw(out *image.RGBA, p [768]byte, off, on byte) {
	d.DrawAt(out, p, off, on, out.Rect.Dy()-33)
}

// DrawAt changes placement only; matrix dots and palette chronology are unchanged.
func (d *Display) DrawAt(out *image.RGBA, p [768]byte, off, on byte, height int) {
	draw.Draw(out, image.Rect(0, height, 320, height+33), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.Point{}, draw.Src)
	for y := 0; y < 16; y++ {
		for x := 0; x < 160; x++ {
			i := off
			if d.Dots[y*160+x] {
				i = on
			}
			if !d.On && i == on {
				q := int(i) * 3
				// MATRIXOFF is outside LON..LONEND and stores DAC units already.
				gray := assets.DACRGB(21)
				p[q], p[q+1], p[q+2] = gray, gray, gray
				if d.sourceMatrixOff != nil {
					copy(p[q:q+3], d.sourceMatrixOff[:])
				}
			}
			q := int(i) * 3
			out.SetRGBA(2*x, height+2+2*y, color.RGBA{p[q], p[q+1], p[q+2], 255})
		}
	}
}
func Compose(field *image.RGBA, d *Display, p [768]byte, off, on byte) *image.RGBA {
	return ComposeHeight(field, d, p, off, on, 317)
}
func ComposeHeight(field *image.RGBA, d *Display, p [768]byte, off, on byte, height int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, 320, height+33))
	draw.Draw(out, image.Rect(0, 0, 320, height), field, image.Point{}, draw.Src)
	d.Draw(out, p, off, on)
	return out
}

// ComposeFullTable places the unchanged table-space renderer below the matrix.
func ComposeFullTable(field *image.RGBA, d *Display, p [768]byte, off, on byte) *image.RGBA {
	return ComposeFullTableOffset(field, d, p, off, on, 0)
}

// SETSCREENSTART adds SCREENPOSY to the source raster: positive nudge exposes
// later table rows, moving artwork upward by exactly that many pixels. Keep
// the matrix fixed and clip the field to its own rectangle. No shake clock.
func ComposeFullTableOffset(field *image.RGBA, d *Display, p [768]byte, off, on byte, screenOffset int16) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, 320, 609))
	area := image.Rect(0, 33, 320, 609)
	if screenOffset != 0 {
		draw.Draw(out, area, image.NewUniform(color.Black), image.Point{}, draw.Src)
	}
	draw.Draw(out, area, field, image.Pt(0, int(screenOffset)), draw.Src)
	d.DrawAt(out, p, off, on, 0)
	return out
}
func ComposeNative(field *image.RGBA, d *Display, p [768]byte, off, on byte, c settings.Config, screenOffset int16) *image.RGBA {
	if c.ScrollMode == settings.ScrollOff {
		return ComposeFullTableOffset(field, d, p, off, on, screenOffset)
	}
	return ComposeHeight(field, d, p, off, on, c.FieldHeight())
}

// Original MATRIXON is percentage data recalculated by RECALC_LIGHTS.
func MatrixPalette(p [768]byte, on byte) [768]byte {
	for i, v := range []int{95, 70, 27} {
		p[int(on)*3+i] = assets.DACRGB(byte(v * 162 >> 8))
	}
	return p
}
func (d *Display) SpringGraphics() []byte {
	return append([]byte(nil), d.data[d.Content.Spring:d.Content.Spring+230]...)
}

// GlyphText consumes bitmap records exported from original generated artwork.
func (d *Display) GlyphText(s string, x, y int, font map[string][]byte) {
	for _, c := range []byte(s) {
		rows := font[strconv.Itoa(int(c))]
		for r, b := range rows {
			for col := 0; col < 8; col++ {
				xx, yy := x+col, y+r
				if xx >= 0 && xx < 160 && yy >= 0 && yy < 16 {
					d.Dots[yy*160+xx] = b&(128>>uint(col)) != 0
				}
			}
		}
		x += 8
	}
}
func (d *Display) Score(s string) { d.ScoreAt(s, 160) }

// InvalidateScore is FANTASIE.MAC UPDAT_SCORE: oldbuf=12h, lastcommas=0.
func (d *Display) InvalidateScore() {
	for i := range d.scoreCache {
		d.scoreCache[i] = 0x12
	}
	d.scoreCommas = 0
}
func (d *Display) ScoreAt(s string, end int) {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		s = "0"
	}
	// Linked CODE2 starts BP at oldbuf after skipping leading zeros. Cache
	// indices therefore follow visible digits, rather than their BCD cells.
	for i, c := range []byte(s) {
		digit := c - '0'
		if d.scoreCache[i] != digit {
			d.scoreGlyph(c, end-8*len(s)+8*i)
			d.scoreCache[i] = digit
		}
	}
	commas := (len(s) - 1) / 3
	if commas == 0 || commas == d.scoreCommas {
		return // CODE2's all-zero-group exit does not write lastcommas.
	}
	d.scoreCommas = commas
	// CODE2 restores the caller's BX before comma scanning. SI=BX+0a1bh
	// -200; plane4 writes SI/SI+168, plane1 writes SI+1/SI+168.
	// BX=168+(end-96)/2, so each comma occupies four retained dots.
	for i := 1; i <= commas; i++ {
		x := end - 24*i - 1
		for _, point := range [][2]int{{x, 14}, {x, 15}, {x + 1, 14}, {x - 1, 15}} {
			if point[0] >= 0 && point[0] < DotWidth {
				d.Dots[point[1]*DotWidth+point[0]] = true
			}
		}
	}
}

func (d *Display) Number(s string, x, y, height int) {
	// PRINT*_TASK resets NUMBERBUF, but PRINT_NUMBER skips leading zero
	// cells without writing VGA memory. Source callers own any later clear.
	d.printNumber(s, x, y, height, false)
}

// FANTASIE print*_task/print_siffror* share the linked PRINT_NUMBER routine.
// Leading BCD zeros are skipped; an all-zero BCD prints no glyphs. This is
// deliberately separate from ordinary source text and generated CODE2 SCORE.
func (d *Display) printNumber(s string, x, y, height int, centered bool) {
	s = strings.TrimLeft(s, "0")
	if centered {
		// REPE SCASB includes the first nonzero byte in its count before
		// subtracting twice that count from DI (TABLE1.PRG 6f21..6f36).
		x += 4 * (11 - len(s))
	} else {
		x += 8 * (12 - len(s))
	}
	d.Text(s, x, y, height)
	// Linked PRINT_NUMBER -> 706a (TABLE1.PRG): the three leading BCD
	// groups select comma count. SI=DI+height*168+35; plane4 writes SI
	// and SI+168, plane1 writes SI+1 and SI+168. These stores differ from
	// CODE2 SCORE's three-dot comma and do not erase skipped leading cells.
	end := x + 8*len(s)
	for i := 1; i <= (len(s)-1)/3; i++ {
		left, top := end-24*i-1, y+height
		for _, point := range [][2]int{{left, top}, {left + 1, top}, {left, top + 1}, {left - 1, top + 1}} {
			if point[0] >= 0 && point[0] < DotWidth && point[1] >= 0 && point[1] < DotHeight {
				d.Dots[point[1]*DotWidth+point[0]] = true
			}
		}
	}
}

// FinishBonusField is each table's NO_MORE_NUFFROR -> CLEAR_BOX2.
// DI=7*2*SW/4, width=16*(12-1)/4-16*4/4 bytes, height=10.
func (d *Display) FinishBonusField() {
	for y := 6; y < 16; y++ {
		for x := 0; x < 56; x++ {
			d.Dots[y*DotWidth+x] = false
		}
	}
}

func (d *Display) Argument(n int) string {
	if n >= len(d.args) {
		return ""
	}
	return d.args[n]
}

// TowerWindow draws STONES showtower's packed four-dot artwork window.
// This consumes the table scheduler's row; it owns no animation clock.
func (d *Display) TowerWindow(offset, row int) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 160; x++ {
			q := offset + (row+y)*40 + x/4
			bit := uint(7 - 2*(x%4))
			d.Dots[y*160+x] = d.data[q]&(1<<bit) != 0
		}
	}
}

// SourceMatrixPalette supports a table's original MATRIXON percentage packet.
// Stones demonstrates a different RGB triplet from the accepted F1-F3 path.
func (d *Display) SourceMatrixPalette(p [768]byte) [768]byte {
	b := d.Content.Texts["MATRIXON"]
	for i := 0; i < 3; i++ {
		p[int(b[0])*3+i] = assets.DACRGB(byte(uint16(b[2+i]) * 162 >> 8))
	}
	return p
}
func (d *Display) UseSourceMatrixOff() {
	b := d.Content.Texts["MATRIXOFF"]
	v := [3]byte{assets.DACRGB(b[2]), assets.DACRGB(b[3]), assets.DACRGB(b[4])}
	d.sourceMatrixOff = &v
}

// MatrixPaletteMode uses COLOR_2_BW after RECALC_LIGHTS; MATRIXOFF stays direct.
func MatrixPaletteMode(p [768]byte, mode byte, on byte) [768]byte {
	p = MatrixPalette(p, on)
	if mode == 1 {
		q := int(on) * 3
		v := (int(p[q]>>2) + int(p[q+1]>>2) + int(p[q+2]>>2)) / 3
		rgb := assets.DACRGB(byte(v))
		p[q], p[q+1], p[q+2] = rgb, rgb, rgb
	}
	return p
}

// SetPlayers writes the original encoded DATA text buffers. DOS PRINT_TEXT
// decodes numeric glyphs stored as digit+7; the native source writer does too.
func (d *Display) SetPlayers(player, count, ball int) {
	for _, v := range []struct {
		label        string
		index, value int
	}{
		{"PLAYERSTEXT", 7, player}, {"NO_OF_PLAYERS_TEXT", 8, count},
		{"BALLSTEXT", 5, ball}, {"PLAY_TEXT", 8, player}, {"PLAY_TEXT", 18, ball},
	} {
		if b := d.Content.Texts[v.label]; len(b) > v.index {
			d.MutableText(v.label, v.index+1)[v.index] = byte(v.value) + '7'
		}
	}
}
func (d *Display) MatchPlayers(scores []byte, winner *byte) bool {
	label := "LAST_TEXT"
	if winner != nil {
		label = "FLASHLAST"
	}
	buf := d.MutableText(label, 18)
	any := false
	for i, digit := range scores {
		if winner == nil {
			buf[2*i], buf[2*i+1] = digit+'7', ' '
			flash := d.MutableText("FLASHLAST", 18)
			flash[2*i], flash[2*i+1] = buf[2*i], buf[2*i+1]
		} else if digit == *winner {
			any = true
		} else {
			buf[2*i], buf[2*i+1] = '*', '*'
		}
	}
	d.Text(strings.TrimRight(d.SourceText(label), "\x00"), 0, 0, 5)
	if winner != nil && any {
		d.startFlash(3)
	}
	return any
}

// CurrentOperation exposes the source handler for temporal diagnostics.
func (d *Display) CurrentOperation() string { return d.op }
