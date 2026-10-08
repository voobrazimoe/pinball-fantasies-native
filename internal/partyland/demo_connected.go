//go:build dmoimpl1 || demodev

package partyland

import (
	"fmt"
	"pinballfantasies/internal/physics"
	"pinballfantasies/internal/presentation"
	"pinballfantasies/internal/tablelogic"
)

// Staged demo execution: no Game.Sync or canonical gameplay callbacks.
type demoConnected struct {
	*demoCore
	calls      uint64
	phase      string
	order      []string
	drains     int
	oldExpired bool
	hiRes      bool // isolated linked HI_RES projection; structural inputs only
	snapshots  []demoDrainSnapshot
}

type demoDrainSnapshot struct {
	Boundary                                                                                  string
	Counter                                                                                   uint16
	Expired, Hold, Lost, Loosing, ScoreChanged, PartyFlash, InhibitEffect, Special, AllowFlip bool
	PukeForbidden                                                                             bool
	Ball                                                                                      physics.Ball
	Matrix                                                                                    matrixState
	Slots                                                                                     [50]uint64
	Wait                                                                                      uint16
	Audio                                                                                     silentJingle
}

func (d *demoConnected) snapshot(boundary string) {
	g := d.game
	s := demoDrainSnapshot{Boundary: boundary, Counter: d.counter, Expired: d.expired, Hold: d.holdStill,
		Lost: g.Physics.Ball.Lost, Loosing: d.loosing, ScoreChanged: g.ScoreChanged, PartyFlash: g.partyFlash,
		InhibitEffect: g.inhibitEffect, Special: d.specialMode, AllowFlip: g.Physics.AllowFlip, PukeForbidden: g.PukeForbidden,
		Ball: g.Physics.Ball, Matrix: g.matrix, Wait: g.waitCounters["PARTY_ON_TASK1"], Audio: g.Audio}
	for i, task := range g.tasks {
		if task != nil {
			s.Slots[i] = g.taskIDs[i]
		}
	}
	d.snapshots = append(d.snapshots, s)
}

// Linked LOOSE_BALL 515..5b9. No canonical drain/party/bonus callback.
func (d *demoConnected) looseBall() error {
	if d.failure != nil {
		return d.failure
	}
	g := d.game
	d.snapshot("LOOSE_BALL entry")

	if g.PukeForbidden {
		return d.rejectEdge("LOOSE_BALL", "PUKEFORBIDDEN return", "alternate drain branch unimplemented")
	}
	d.holdStill = true
	d.screenForce2 = 369
	if d.hiRes {
		d.screenForce2 = 259
	}
	g.Physics.SetBall(15, 47, 0, 0, false)
	g.Physics.AllowFlip = false
	d.specialMode = false
	g.Happy, g.Mega = false, false
	d.loosing = true
	g.Phase = BallLost
	d.snapshot("LOOSE_BALL guards")
	if g.ScoreChanged {
		if d.scoredDrainOperands {
			return d.scoredDrain()
		}
		return d.rejectEdge("LOOSE_BALL", "scored drain", "selected scored branch unimplemented before expired/LOSTBALL consumer")
	}

	g.Audio.Priority = 0
	if err := d.install(demoPartyProgram()); err != nil {
		return err
	}
	a := g.musicClock()
	if !a.Play(tablelogic.JingleSpec{Position: 0, Repeat: 0, Priority: 1}, 62, timing.Cues) {
		return d.rejectEdge("PARTY_ON", "S_SPRING", "priority clear must admit source spring request")
	}
	g.storeMusicClock(a)
	g.musicOK = true
	if err := d.queue(demoTask{Site: "PARTY_ON_TASK1", Delay: 30, Action: demoPartyOnWait}); err != nil {
		return err
	}
	d.snapshot("PARTY_ON installed")
	return nil
}

// Typed reviewed linked operands. Each dispatch passes the fallible gate.
// PARTYRUT retains SI=1; the trailing zero is data, never visited on that path.
func demoPartyProgram() []presentation.Command {
	return []presentation.Command{
		{Op: "_CLEAR4"},
		{Op: "_FLASHON", Args: []string{"3"}, Nums: map[int]int{0: 3}},
		{Op: "_PARTYONN", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "_PRINT13", Args: []string{"PARTY_ON_TEXT", "336"}, Nums: map[int]int{1: 336}},
		{Op: "_PARTYON", Args: []string{"1"}, Nums: map[int]int{0: 1}},
		{Op: "0"},
	}
}

func connected(g *Game) *demoConnected {
	d := newDemoCore()
	d.game = g
	return &demoConnected{demoCore: d}
}

func (d *demoConnected) rejectEdge(producer, consumer, reason string) error {
	if d.failure == nil {
		d.failure = &demoUnsupported{Producer: producer, Consumer: consumer, Phase: d.phase, Calculation: d.calls,
			Guard: fmt.Sprintf("timer=%d expired=%t sourceHold=%t nativeHold=%t lost=%t high=%t hit=(%d,%d) tunnel=%d", d.counter, d.expired, d.holdStill, d.game.Physics.Ball.Hold, d.game.Physics.Ball.Lost, d.game.Physics.Ball.High, d.game.Physics.Ball.HitX, d.game.Physics.Ball.HitY, d.game.TunnelTime), Reason: reason + fmt.Sprintf(" scoreChanged=%t oldExpired=%t loosing=%t partyFlash=%t inhibit=%t special=%t pukeForbidden=%t", d.game.ScoreChanged, d.oldExpired, d.loosing, d.game.partyFlash, d.game.inhibitEffect, d.specialMode, d.game.PukeForbidden)}
	}
	return d.failure
}

func (d *demoConnected) gate(producer, consumer string) error {
	if d.failure != nil {
		return d.failure
	}
	switch consumer {
	case "collision/flipper observation":
		if d.gameplayOperands {
			return nil
		}
	case "scroll":
		if d.gameplayOperands && d.game.ScreenForce == -1 {
			return nil
		}
	case "early.step", "early.finish", "targets", "late.step":
		d.order = append(d.order, consumer)
		return nil
	}
	return d.rejectEdge(producer, consumer, "gameplay callback has no admitted demo implementation")
}

func (d *demoConnected) stage(name string, input physics.Inputs) error {
	if name == "targets" && d.toucherOperands {
		return d.game.Physics.CandidateTargets(d.preflightTarget, d.consumeTarget)
	}
	if err := d.game.Physics.CandidateStage(name, input, d.holdStill, d.gate); err != nil {
		if d.failure == nil {
			return d.rejectEdge(name, "physics arithmetic", err.Error())
		}
		return d.failure
	}
	return nil
}

func (d *demoConnected) sync(input physics.Inputs, budget bool) error {
	if d.terminal != nil {
		return nil
	}
	if d.failure != nil {
		return d.failure
	}
	if d.game.Playback != nil {
		return d.rejectEdge("audioTick", "host playback", "host consumer outside silent candidate")
	}
	if d.gameplayOperands {
		d.phase = "keyboard admission"
		if err := d.keyboardEntry(input); err != nil {
			return err
		}
	}
	d.game.audioTick()
	d.calls++
	d.phase = "early physics"
	p := d.game.Physics

	if p.Stopped || (p.Ball.Lost && !d.loosing && !d.highScoreOperands) {
		return d.rejectEdge("entry", "drain continuation", "previous drain is not a new calculation")
	}
	for i := 0; i < 2; i++ {
		if err := d.stage("early.step", input); err != nil {
			return err
		}
	}
	if err := d.stage("early.finish", input); err != nil {
		return err
	}

	drained := p.Ball.Lost && !d.loosing
	if drained {
		d.phase = "drain handoff"
		d.oldExpired = d.expired
		d.drains++
		d.order = append(d.order, "drain handoff")
		if err := d.looseBall(); err != nil {
			d.failure.Phase, d.failure.Calculation = d.phase, d.calls
			return err
		}
	}
	d.phase = "electronics"

	for i, task := range d.game.tasks {
		if task != nil && !d.ownedTasks[d.game.taskIDs[i]] {
			return d.rejectEdge("DO_TASKS", "unowned task", "canonical/unknown closure must not execute")
		}
	}
	if d.game.matrix.active && !d.ownedMatrix {
		return d.rejectEdge("DO_MATRIX", "unowned matrix", "canonical/unknown matrix must not execute")
	}

	if d.game.TunnelTime == 721 || d.game.TunnelTime == 1 {
		return d.rejectEdge("UPDATE_COUNTERS", "tunnel lamp/flash", "nested gameplay effects unimplemented")
	}
	d.order = append(d.order, "ElectronicsCalculation")
	if err := d.electronicsPrefix(); err != nil {
		return err
	}
	if !p.Ball.Lost {
		d.phase = "electronics areas/targets"
		if d.gameplayOperands {
			if err := d.gameplayArea(); err != nil {
				return err
			}
		} else if area := d.game.areaConsumer(); area != "" {
			return d.rejectEdge("checkAreas", area, "area callback unimplemented; no area bookkeeping started")
		}
		if err := d.stage("targets", input); err != nil {
			return err
		}
	}
	if d.gameplayOperands {
		d.keyboardLamps(input)
	}
	d.trace = append(d.trace, "electronics-no-callback")
	d.phase = "task/matrix"
	d.order = append(d.order, "task/matrix")
	if err := d.taskMatrixSuffix(budget); err != nil {
		d.failure.Phase, d.failure.Calculation = d.phase, d.calls
		return err
	}
	if d.terminal != nil {
		return nil
	}
	if d.loosing {
		d.snapshot("post-drain electronics")
	}

	if d.loosing && !d.holdStill {
		return d.rejectEdge("post-drain", "late lost-ball continuation", "source hold released without admitted reset")
	}
	if d.gameplayOperands {
		if err := d.stage("scroll", input); err != nil {
			return err
		}
	}
	d.phase = "late physics"
	if err := d.stage("late.step", input); err != nil {
		return err
	}
	if d.gameplayOperands {
		d.game.Physics.Syncs++
	}
	d.order = append(d.order, "complete")
	return nil
}
