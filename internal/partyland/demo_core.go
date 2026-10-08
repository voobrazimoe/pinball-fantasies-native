//go:build dmoimpl1 || demodev

package partyland

// Staged demo semantics shared by development tests and the experimental host.
// No unguarded canonical rule callback is admitted.
import (
	"bytes"
	"fmt"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
)

type demoUnsupported struct {
	Producer, Guard, Reason string
	Consumer, Phase         string
	Calculation             uint64
}

func (e *demoUnsupported) Error() string {
	return fmt.Sprintf("UNSUPPORTED_DEMO_TRANSITION producer=%s state/guard=%s consumer=%s phase=%s calculation=%d reason=%s", e.Producer, e.Guard, e.Consumer, e.Phase, e.Calculation, e.Reason)
}

type demoAction uint8

const (
	demoPreserve demoAction = iota // no matrix write
	demoReleaseSourceHold
	demoReplaceMatrix
	demoResetTasks
	demoPartyOnWait // DS:36c9, PARTYFLASH then typed demo NEW_BALL
	demoEnableToucher
	demoScoredWait
	demoChildWait // reviewed child body after source-specific wait
)

type demoTask struct {
	Site    string
	Delay   uint16
	Action  demoAction
	Program []presentation.Command
}

// Linked expiry stream; presentation admission requires pinned private operands.
func demoExpiryProgram() []presentation.Command {
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

type demoCore struct {
	infoFactorySeal [32]byte

	infoPresentation bool
	infoTexts        map[string][]byte
	infoFactory      [64]byte
	infoStarts       int

	game                           *Game
	counter                        uint16
	expired, holdStill             bool // source HOLDSTILL, never an alias of Ball.Hold
	failure                        *demoUnsupported
	trace                          []string
	cueRequests                    int
	ownedTasks                     map[uint64]bool
	ownedMatrix                    bool
	loosing, specialMode           bool
	visaKeys, addPlayers           bool // linked guards; PARTYFLASH skips VISAKEYS entirely
	shiftPressed, inhibitCountdown bool
	keyTaskEmpty, dotReady         bool
	handoffs                       int
	bonusX                         uint8 // linked BONUS_X reset store, separate from BONUSMULTIPEL
	screenForce2                   int16
	childOperands                  bool // pinned sample/note/channel and SETBALL operands
	childFires                     []demoChildFire
	resetOperands                  bool // reset masks and mutable BONUS_TEXT verified
	expiryTexts                    [2][]byte
	expiryFont                     int
	expiryPresentation             bool
	terminal                       *demoTerminal
	fade                           demoFade
	visits                         int
	partyPresentation              bool // pinned demo text and consumed font13 verified before admission
	gameplayFont                   int
	panelTexts                     [2][]byte
	infoCount                      uint16
	fjantText                      bool
	scoredDueMatrix                matrixState
	scoredDueExpiry                bool
	scoredDueAudio                 silentJingle
	scoredNewBall                  bool
	scoredDrainOperands            bool
	drainText                      []byte
	drainFont                      int
	drainBallText                  []byte
	drainFires                     []demoChildFire
	bonusNodes                     []uint16
	highScoreOperands              bool
	factoryTop                     *Decimal // verified compiled/native factory, volatile candidate lifetime
	factoryTopSource               string
	toucherOperands                bool
	toucherFires                   []uint16
	gameplayOperands               bool         // bounded 2B8 consumers, never canonical callbacks
	controls                       func() error // fresh keyboard consumer, after task scan
}

func newDemoCore() *demoCore {
	return &demoCore{game: &Game{Physics: &physics.Game{}, Display: &presentation.Display{}, waitCounters: make(map[string]uint16), matrixTimeLeft: true, inChute: true}}
}
func (d *demoCore) reject(producer, guard, reason string) error {
	if d.failure == nil {
		d.failure = &demoUnsupported{Producer: producer, Guard: guard, Reason: reason, Consumer: producer, Calculation: uint64(d.counter)}
	}
	return d.failure
}

// Called only after the private adapter pins the demo installation. The
// canonical display supplies the existing glyph consumer, never substitute text.
func (d *demoCore) loadPartyPresentation(demo, canonical []byte) error {
	if d.failure != nil {
		return d.failure
	}
	const textAt, fontAt = 0x19db0 + 7647, 0x19db0 + 0x6300
	font, ok := d.game.Display.Content.Fonts["13"]
	if !ok || font < 0 || font+42*13 > len(canonical) || textAt+20 > len(demo) || fontAt+42*13 > len(demo) {
		return d.reject("PARTY_ON presentation", "text/font extent", "pinned presentation inputs unavailable")
	}
	text := demo[textAt : textAt+20]
	if text[19] != 0 || !bytes.Equal(text, d.game.Display.Content.Texts["PARTY_ON_TEXT"]) || !bytes.Equal(canonical[font:font+42*13], demo[fontAt:fontAt+42*13]) {
		return d.reject("PARTY_ON presentation", "text/font correspondence", "shared glyph consumer has no verified demo operands")
	}
	d.game.Display.Content.Texts["PARTY_ON_TEXT"] = append([]byte(nil), text...)
	d.partyPresentation = true
	return nil
}

func (d *demoCore) command(c presentation.Command) bool {
	if d.infoCommand(c) {
		return true
	}
	if d.scoredCommand(c) {
		return true
	}
	switch c.Op {
	case "_PRINT5":
		return d.gameplayOperands && d.game.Display.Content.Fonts["5"] == d.gameplayFont && len(c.Args) == 2 && len(c.Nums) == 1 && ((c.Args[0] == "PLAYERSTEXT" && c.Args[1] == "340" && c.Nums[1] == 340 && bytes.Equal(d.panelTexts[0], d.game.Display.Content.Texts[c.Args[0]])) || (c.Args[0] == "BALLSTEXT" && c.Args[1] == "1684" && c.Nums[1] == 1684 && bytes.Equal(d.panelTexts[1], d.game.Display.Content.Texts[c.Args[0]])))
	case "_PARTYOFF":
		return d.gameplayOperands && len(c.Args) == 1 && c.Args[0] == "1" && len(c.Nums) == 1 && c.Nums[0] == 1
	case "_DEMO_PARTY_FLASHOFF":
		return d.gameplayOperands && len(c.Args) == 1 && c.Args[0] == "1" && len(c.Nums) == 1 && c.Nums[0] == 1
	case "_CLEAR4":
		return len(c.Args) == 0 && len(c.Nums) == 0
	case "_SCROLL", "_PRINT13_NUMBER", "_FLASHOFF", "_DEMO_FADE", "_DEMO_QUIT":
		return d.expiryCommand(c)
	case "_FLASHON":
		return len(c.Args) == 1 && len(c.Nums) == 1 && ((c.Args[0] == "3" && c.Nums[0] == 3) || (d.expiryPresentation && c.Args[0] == "1" && c.Nums[0] == 1))
	case "_PARTYONN":
		return d.partyPresentation && len(d.game.Display.Content.Texts["PARTY_ON_TEXT"]) == 20 && len(c.Args) == 1 && c.Args[0] == "1" && len(c.Nums) == 1 && c.Nums[0] == 1
	case "_PARTYON":
		return len(c.Args) == 1 && c.Args[0] == "1" && len(c.Nums) == 1 && c.Nums[0] == 1
	case "_PRINT13":
		return d.partyPresentation && len(c.Args) == 2 && c.Args[0] == "PARTY_ON_TEXT" && c.Args[1] == "336" && len(c.Nums) == 1 && c.Nums[1] == 336
	case "0":
		return len(c.Args) == 0 && len(c.Nums) == 0
	case "_WAIT":
		n, ok := c.Nums[0]
		return ok && n >= 0 && n <= 65535 && len(c.Args) == 1 && len(c.Nums) == 1
	}
	return false
}
func (d *demoCore) install(program []presentation.Command) error {
	return d.installOwned(program, true)
}
func (d *demoCore) installOwned(program []presentation.Command, reset bool) error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if len(program) == 0 || !d.command(program[0]) {
		return d.reject("DO_MATRIX", "entry", "unimplemented or malformed matrix entry")
	}
	d.fade.Active = false

	commands := make([]presentation.Command, len(program))
	for i, c := range program {
		commands[i] = c
		commands[i].Args = append([]string(nil), c.Args...)
		commands[i].Nums = make(map[int]int)
		for k, v := range c.Nums {
			commands[i].Nums[k] = v
		}
	}
	d.game.Display.Content.Commands = commands
	d.ownedMatrix = true
	if reset {
		d.game.Display.StartMatrix()
	}
	d.game.matrix = matrixState{active: true, sourceProgram: true, consumer: d}
	d.game.matrixDispatch()
	if d.failure != nil {
		return d.failure
	}
	return nil
}
func (d *demoCore) queue(t demoTask) error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if t.Action == demoEnableToucher && (t.Site != "ENABLETOUCHER" || t.Delay != 20 || !d.toucherOperands) {
		return d.reject("ADDTASK", "ENABLETOUCHER", "unadmitted task body/wait")
	}
	if t.Action == demoScoredWait && (!d.scoredDrainOperands || !((t.Site == "SOUNDRINNER" && t.Delay == 5) || (t.Site == "NEW_BALL_TASK" && t.Delay == 30))) {
		return d.reject("ADDTASK", "scored task", "unadmitted scored producer")
	}
	if t.Action == demoChildWait && !demoChild(t.Site, t.Delay) {
		return d.reject("ADDTASK", "child wait site", "unvalidated child producer")
	}
	if t.Action > demoChildWait || t.Site == "" || (t.Action == demoPartyOnWait && (t.Site != "PARTY_ON_TASK1" || t.Delay != 30)) {
		return d.reject("ADDTASK", "action/site", "unvalidated task producer")
	}
	free := false
	for _, f := range d.game.tasks {
		if f == nil {
			free = true
			break
		}
	}
	if !free {
		return d.reject("ADDTASK", "50 occupied slots", "no first-free task slot")
	}

	d.game.task(func() bool {
		if d.failure != nil {
			return false
		}

		if d.game.waitCounters[t.Site] == t.Delay {
			if t.Action == demoScoredWait && d.preflightScoredTask(t.Site) != nil {
				return false
			}
			if t.Action == demoEnableToucher && !d.toucherOperands {
				d.reject("ENABLETOUCHER", "body operands", "unadmitted due consumer before WAIT reset")
				return false
			}
			if t.Action == demoPartyOnWait && d.preflightNewBall() != nil {
				return false
			}
			if t.Action == demoChildWait {
				if d.preflightChild(t.Site) != nil {
					return false
				}
			}
		}

		if t.Action == demoReplaceMatrix && d.game.waitCounters[t.Site] == t.Delay && (len(t.Program) == 0 || !d.command(t.Program[0])) {
			d.reject("DO_TASKS", "due replacement entry", "unimplemented nested matrix consumer")
			return false
		}
		if !d.game.waitReady(t.Site, t.Delay) {
			return false
		}
		switch t.Action {
		case demoScoredWait:
			d.executeScoredTask(t.Site)
		case demoEnableToucher:
			d.game.touchDisabled = false
			d.toucherFires = append(d.toucherFires, d.counter)
		case demoChildWait:
			d.executeChild(t.Site)
		case demoPartyOnWait:
			d.game.partyFlash = true
			d.newBallHandoff()
		case demoReleaseSourceHold:
			d.holdStill = false
		case demoReplaceMatrix:
			if d.install(t.Program) != nil {
				return false
			}
		case demoResetTasks:
			d.game.tasks = [50]func() bool{}
			d.game.waitCounters = make(map[string]uint16)
		}
		return true
	})
	if d.ownedTasks == nil {
		d.ownedTasks = make(map[uint64]bool)
	}
	d.ownedTasks[d.game.nextTaskID] = true
	return nil
}
func (d *demoCore) matrixVisit() error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	g := d.game
	if d.gameplayOperands && g.matrixTimeLeft && !g.matrix.active {
		return d.gameplayIdle()
	}
	if !g.matrixTimeLeft || !g.matrix.active {
		return nil
	}
	if !g.matrix.sourceProgram || g.matrix.next < 1 || g.matrix.next > len(g.Display.Content.Commands) || !d.command(g.Display.Content.Commands[g.matrix.next-1]) {
		return d.reject("DOTRUT", "current routine", "unvalidated matrix consumer")
	}
	if d.matrixCompletes() {
		if g.matrix.next >= len(g.Display.Content.Commands) || !d.command(g.Display.Content.Commands[g.matrix.next]) {
			return d.reject("NEXT_A", fmt.Sprintf("op=%s remaining=1 next=%d", g.matrix.op, g.matrix.next), "next matrix consumer not implemented; visit stopped before dispatch")
		}
	}
	d.visits++
	g.matrixTick()
	if d.failure != nil {
		return d.failure
	}
	return nil
}

// calculation is an admitted suffix after the harness's early-physics/drain
// observation. It intentionally cannot run real physics/areas/keyboard rules.
// A caller requesting those consumers must use unsupported, never Game.Sync.
func (d *demoCore) calculation(paused, matrixBudget bool) error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if paused {
		return nil
	}
	d.trace = append(d.trace, "early/drain-observed")
	if err := d.electronics(matrixBudget); err != nil {
		return err
	}
	d.trace = append(d.trace, "late-physics-empty")
	return nil
}

// electronics is the same semantic core used by staged native execution.
func (d *demoCore) electronics(matrixBudget bool) error {
	if d.terminal != nil {
		return nil
	}
	if err := d.electronicsPrefix(); err != nil {
		return err
	}
	d.trace = append(d.trace, "electronics-empty")
	return d.taskMatrixSuffix(matrixBudget)
}

func (d *demoCore) electronicsPrefix() error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if !d.scoredDrainOperands || !d.loosing {
		d.game.updateCounters()
	}
	d.trace = append(d.trace, "UPDATE_COUNTERS")
	d.counter++
	if d.counter == 35998 {
		d.expired, d.holdStill = true, true
		d.cueRequests++
		a := d.game.musicClock()
		a.Play(tablelogic.JingleSpec{Position: 13, Repeat: 0, Priority: 255}, 62, timing.Cues)
		d.game.storeMusicClock(a)
		if err := d.install(demoExpiryProgram()); err != nil {
			return err
		}
	}
	d.trace = append(d.trace, "timer")
	return nil
}

func (d *demoCore) taskMatrixSuffix(matrixBudget bool) error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}

	d.game.Display.Flash()
	d.trace = append(d.trace, "KEYTASK-empty", "DO_TASKS")
	d.game.runTasks()
	if d.failure != nil {
		return d.failure
	}
	if d.controls != nil {
		if err := d.controls(); err != nil {
			return err
		}
	}
	d.game.flashTick()
	d.game.matrixTimeLeft = matrixBudget
	d.trace = append(d.trace, "matrix-budget")
	if err := d.matrixVisit(); err != nil {
		return err
	}
	return nil
}
func (d *demoCore) unsupported(producer, guard string) error {
	return d.reject(producer, guard, "physics/gameplay/control path outside DMO-IMPL-1 staged harness")
}
