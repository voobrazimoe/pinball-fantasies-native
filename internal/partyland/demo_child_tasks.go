//go:build dmoimpl1 || demodev

package partyland

import (
	"pinballfantasies/internal/audio"
	"pinballfantasies/internal/physics"
)

type demoChildFire struct {
	Site                string
	Counter             uint16
	Ball                physics.Ball
	SourceHold, InChute bool
	Audio               silentJingle
}

// Complete consumer admission precedes the shared WAITSYNCS reset. The silent
// effect request is a real typed Sound event using the existing shared consumer;
// sample rendering/host playback is outside this candidate.
func (d *demoCore) preflightChild(site string) error {
	if d.failure != nil {
		return d.failure
	}
	if !d.childOperands {
		return d.reject(site, "linked child operands", "verified child consumer unavailable")
	}
	g := d.game
	switch site {
	case "SOUNDBRICKUPP", "SOUNDNEWBALL":
		label, want := "SBRICKUPP", (audio.Effect{Sample: 23, Note: 23})
		if site == "SOUNDNEWBALL" {
			label, want = "SNEWBALL", audio.Effect{Sample: 28, Note: 18}
		}
		if g.Playback != nil || audio.Effects[label] != want {
			return d.reject(site, label, "silent shared SOUND_EFFECT consumer unavailable")
		}
	case "SETBALL":
		if g.Physics.Table == nil || !g.inChute || g.Physics.Ball.Lost || d.loosing || g.Phase != NewBall {
			return d.reject(site, "SETBALL prerequisites", "admitted new-ball table/chute/reset unavailable")
		}
	default:
		return d.reject(site, "child identity", "unknown child")
	}
	return nil
}

func (d *demoCore) executeChild(site string) {
	g := d.game
	switch site {
	case "SOUNDBRICKUPP":
		g.sound("SBRICKUPP")
	case "SOUNDNEWBALL":
		g.sound("SNEWBALL")
	case "SETBALL":
		g.Physics.SetBall(297, 530, 10, 0, false)

		d.holdStill = false
		d.screenForce2 = -1
		g.Physics.TargetRaster = -1
		g.Phase = Playing
	}
	d.childFires = append(d.childFires, demoChildFire{site, d.counter, g.Physics.Ball, d.holdStill, g.inChute, g.Audio})
}
