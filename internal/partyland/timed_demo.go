package partyland

import (
	"encoding/binary"

	"pinballfantasies/internal/assets"

	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
)

// Official 10-minute DOS demo, layered on the canonical Party Land rules.
// Proved demo differences replace their canonical counterparts; every other
// path, including states the DMO0 research left open, continues canonically.
// See docs/partyland-10min-demo.md for the proved/fallback boundary.

// DemoThreshold is the linked DO_ELECTRONICS equality, 10*60*60-2.
const DemoThreshold = 35998

// DemoInputs are the demo-only records outside the shared decoded space.
type DemoInputs struct {
	PlayersText []byte    // fixed duration label replacing the player label
	ExpiryTexts [2][]byte // the two expiry scrolls, 255 terminated
}

type timedDemo struct {
	g           *Game
	counter     uint16 // DS:34cd, table-lifetime, never reset by new ball/game
	expired     bool   // DS:34cf, sticky once set
	counted     bool   // one admitted ElectronicsCalculation per Sync
	equalities  int
	quit        bool
	fade        bool
	fadeLevel   uint16    // remaining FADE steps out of 256, also the DAC scale
	volume      uint16    // AL=6 volume request, stepped every sixteen visits
	dac         [768]byte // VGA DAC snapshot (6-bit units) when FADE starts
	playersText []byte
	playersNo   []byte // DS:1e8b as loaded; the attract start writes the count
}

// Linked expiry stream at 0x1ba17. CLEAR/SCROLL/FLASH/PRINT/WAIT use the
// shared interpreter; only FADE and QUIT are demo consumers.
func timedDemoExpiryProgram() []presentation.Command {
	return []presentation.Command{
		{Op: "_CLEAR4"}, {Op: "_SCROLL", Args: []string{"DEMO_EXPIRY_TEXT1"}},
		{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_PRINT13_NUMBER", Args: []string{"SIFFRORNA", "344"}, Nums: map[int]int{1: 344}},
		{Op: "_WAIT", Args: []string{"100"}, Nums: map[int]int{0: 100}},
		{Op: "_FLASHOFF", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_SCROLL", Args: []string{"DEMO_EXPIRY_TEXT2"}},
		{Op: "_DEMO_FADE", Args: []string{"256"}, Nums: map[int]int{0: 256}},
		{Op: "_WAIT", Args: []string{"100"}, Nums: map[int]int{0: 100}},
		{Op: "_DEMO_QUIT", Args: []string{"0"}, Nums: map[int]int{0: 0}},
	}
}

const timedDemoExpiryLabel = "DEMO_EXPIRYTS"

// The demo's SHOWPLAYERSTS (DS:1ade, file 0x1b88e) prints the duration label
// at 336 and the demo ball counter. NODOT (file 0x56ef) and the NEW_BALL reset
// both install it. FIRST_NO_OF_PLAYERSTS (0x1b89e) shows the players record.
func timedDemoPanels() map[string][]presentation.Command {
	panel := func(at int, second string) []presentation.Command {
		return []presentation.Command{{Op: "_CLEAR4"},
			{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", ""}, Nums: map[int]int{1: at}},
			{Op: "_PRINT5", Args: []string{second, "1684"}, Nums: map[int]int{1: 1684}}, {Op: "0"}}
	}
	return map[string][]presentation.Command{
		"DEMO_SHOWPLAYERSTS": panel(336, "BALLSTEXT"),
		// FIRST_NO_OF_PLAYERSTS 0x1b89e: the full-game stream with the
		// duration label at 336.
		"DEMO_FIRST_NO_OF_PLAYERSTS": {{Op: "_CLEAR4"},
			{Op: "_PRINT5", Args: []string{"PLAYERSTEXT", ""}, Nums: map[int]int{1: 336}},
			{Op: "_PRINT13", Args: []string{"NOLLA", ""}, Nums: map[int]int{1: 412}},
			{Op: "_PRINT5", Args: []string{"NO_OF_PLAYERS_TEXT", "1684"}, Nums: map[int]int{1: 1684}},
			{Op: "_WAIT_GAME_ON", Args: []string{"?"}}, {Op: "0"}},
	}
}

func (g *Game) firstPlayersPanel() string {
	if g.timed != nil && g.Session.PlayerCount <= 1 {
		return "DEMO_FIRST_NO_OF_PLAYERSTS"
	}
	return "FIRST_NO_OF_PLAYERSTS"
}

// Multiplayer is a native extension: the demo binary never rotates PLAYER.
// A multiplayer demo uses the full game's PLAYER n / BALL m panel so the
// players can see whose turn it is; single player keeps the demo panel.
func (g *Game) playerPanel() string {
	if g.timed != nil && g.Session.PlayerCount <= 1 {
		return "DEMO_SHOWPLAYERSTS"
	}
	return "SHOWPLAYERSTS"
}

// NewDemo creates a fresh table lifetime: timer zero, expired false.
func NewDemo(table *physics.Table, data []byte, in DemoInputs) *Game {
	g := New(table, data)
	d := &timedDemo{g: g, fadeLevel: 256, volume: 256, playersText: append([]byte(nil), in.PlayersText...),
		playersNo: append([]byte(nil), g.Display.Content.Texts["NO_OF_PLAYERS_TEXT"]...)}
	g.timed = d
	c := &g.Display.Content
	// Content commands/labels are shared by every Display; own them first.
	labels := make(map[string]int, len(c.Labels)+3)
	for k, v := range c.Labels {
		labels[k] = v
	}
	commands := append([]presentation.Command(nil), c.Commands...)
	programs := timedDemoPanels()
	programs[timedDemoExpiryLabel] = timedDemoExpiryProgram()
	for _, label := range []string{timedDemoExpiryLabel, "DEMO_SHOWPLAYERSTS", "DEMO_FIRST_NO_OF_PLAYERSTS"} {
		labels[label] = len(commands)
		commands = append(commands, programs[label]...)
	}
	c.Commands, c.Labels = commands, labels
	c.Texts["DEMO_EXPIRY_TEXT1"] = append([]byte(nil), in.ExpiryTexts[0]...)
	c.Texts["DEMO_EXPIRY_TEXT2"] = append([]byte(nil), in.ExpiryTexts[1]...)
	g.playerText()
	// A fresh table process has not run the attract start yet: the players
	// record keeps its loaded placeholder until StartPlayers writes PLAYERS.
	c.Texts["NO_OF_PLAYERS_TEXT"] = append([]byte(nil), d.playersNo...)
	return g
}

// Demo reports the timer state; ok is false for an ordinary game.
func (g *Game) Demo() (counter uint16, expired, ok bool) {
	if g.timed == nil {
		return 0, false, false
	}
	return g.timed.counter, g.timed.expired, true
}

// DemoFinished is the linked QUIT(status 0) handler, the demo's normal end.
func (g *Game) DemoFinished() bool { return g.timed != nil && g.timed.quit }

// demoElectronics is the DO_ELECTRONICS timer: increment, then compare for
// exact equality. It runs after UPDATE_COUNTERS and before areas/targets.
// Pause and attract never reach it; uint16 wraparound repeats the equality.
func (g *Game) demoElectronics() {
	d := g.timed
	if d == nil || d.counted || d.quit {
		return
	}
	d.counted = true
	d.counter++
	if d.counter != DemoThreshold {
		return
	}
	d.expired = true
	d.equalities++
	g.Physics.Ball.Hold = true // source HOLDSTILL; a later SETBALL/release clears it
	g.emit("DemoExpired", "DEMO_TIMER", uint64(d.equalities))
	g.music("S_GAMEOVER2")
	d.install()
}

// install is DO_MATRIX of the expiry entry: a new installation every time,
// never resumption of an older cursor.
func (d *timedDemo) install() {
	d.g.startMatrix(timedDemoExpiryLabel, true)
	d.g.matrix.consumer = d
}

// Expired scored drain requests effect 0x1a4a1 instead of LOSTBALL: cue
// S_GAMEOVER2, zero arithmetic, matrix = expiry entry. Only an admitted
// effect installs it. Returns false when the canonical drain applies.
func (g *Game) demoScoredDrain() bool {
	d := g.timed
	if d == nil || !d.expired {
		return false
	}
	accepted := true
	if !g.inhibitEffect && !g.special() {
		accepted = g.playJingle("S_GAMEOVER2")
	}
	g.score("LOSTBALL", 0)
	g.effectAccepted = accepted && !g.inhibitEffect && !g.special()
	if g.effectAccepted {
		d.install()
	} else {
		g.effectEnded = true
	}
	return true
}

// DEMOVER_CHANGE_PLAYER: keep the player, advance the displayed ball number
// (1..9, 10..29, then 10 again), save the player and queue NEW_BALL_TASK.
// There is no ball limit. Earned shoot-again and match states have no proved
// demo route and fall back to the canonical change.
func (g *Game) demoChangePlayer() (handled bool) {
	if g.timed == nil || g.Lights[51] || g.matchBall {
		return false
	}
	g.savePlayer()
	// Native multiplayer: rotate like the full game, without a ball limit.
	// The displayed ball advances once per round, i.e. when player 1 is next.
	s := &g.Session
	if s.PlayerCount > 1 {
		s.SelectionOpen = false
		s.CurrentPlayer = s.CurrentPlayer%s.PlayerCount + 1
		g.LoadPlayerState(s.Load())
	}
	if s.CurrentPlayer == 1 {
		text := g.Display.MutableText("BALLSTEXT", 6)
		text[5]++
		if text[5] >= '7'+10 {
			text[5] = '7'
			if text[4] == '8' {
				text[4]++
			} else {
				text[4] = '8'
			}
		}
	}
	g.playerText()
	g.waitAt("NEW_BALL_TASK", 30, g.newBall)
	return true
}

// The reviewed demo NEW_BALL reset sets SPRING_VALID; canonical play waits
// for the placed ball to reach the spring area.
func (g *Game) demoNewBall() {
	if g.timed != nil {
		g.Physics.SpringValid = true
	}
}

// demoPlayerText keeps the fixed duration label and the demo ball counter;
// the canonical player/ball digits are not demo consumers.
func (g *Game) demoPlayerText(ballsText []byte) {
	if g.timed == nil {
		return
	}
	g.Display.Content.Texts["PLAYERSTEXT"] = append([]byte(nil), g.timed.playersText...)
	if g.Session.PlayerCount > 1 {
		// The full game's player label, glyph digit '7'+n.
		g.Display.Content.Texts["PLAYERSTEXT"] = []byte{'P', 'L', 'A', 'Y', 'E', 'R', ' ', byte(g.Session.CurrentPlayer) + '7', 0}
	}
	if ballsText != nil {
		g.Display.Content.Texts["BALLSTEXT"] = ballsText
	}
}

func (d *timedDemo) dispatch(c matrixCommand) bool {
	switch c.Op {
	case "_DEMO_FADE":
		d.g.matrix.remaining = uint16(c.Num(0))
		d.fade = true
		d.fadeLevel = d.g.matrix.remaining
		for i, v := range d.g.Display.SourceMatrixPalette(d.g.lampPalette) {
			d.dac[i] = v >> 2
		}
		return true
	case "_DEMO_QUIT":
		d.quit = true
		d.g.matrix.active = false
		d.g.emit("DemoQuit", "QUIT", uint64(c.Num(0)))
		return true
	}
	return false
}
func (d *timedDemo) step(op string, left *uint16) bool {
	if op != "_DEMO_FADE" {
		return false
	}
	*left--
	d.fadeLevel = *left
	if *left&15 == 0 {
		d.volume = *left
	}
	return true
}

// demoFadePalette is the faded VGA DAC. It stays black through the trailing
// WAIT and QUIT; the frame is recomposed with it, not dimmed afterwards.
func (g *Game) demoFadePalette() ([768]byte, bool) {
	var p [768]byte
	d := g.timed
	if d == nil || !d.fade {
		return p, false
	}
	for i, v := range d.dac {
		p[i] = assets.DACRGB(byte(uint16(v) * d.fadeLevel >> 8))
	}
	return p, true
}

// demoPCM applies the expiry fade's audio volume steps.
func (g *Game) demoPCM(pcm []byte) []byte {
	d := g.timed
	if d == nil || !d.fade || pcm == nil {
		return pcm
	}
	out := append([]byte(nil), pcm...)
	for i := 0; i+1 < len(out); i += 2 {
		v := int32(int16(binary.LittleEndian.Uint16(out[i:]))) * int32(d.volume) / 256
		binary.LittleEndian.PutUint16(out[i:], uint16(int16(v)))
	}
	return out
}
