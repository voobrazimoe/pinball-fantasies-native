//go:build dmoimpl1 || demodev

package partyland

import (
	"fmt"
)

func (d *demoConnected) preflightTarget(index int) error {
	if d.failure != nil {
		return d.failure
	}
	if index != 0 || !d.toucherOperands || d.game.Playback != nil {
		return d.rejectEdge(fmt.Sprintf("checkSpringAndTargets target=%d", index), "target/OnEvent", "unadmitted linked target or sound")
	}

	if d.game.touchDisabled {
		return nil
	}
	for _, task := range d.game.tasks {
		if task == nil {
			return nil
		}
	}
	return d.rejectEdge("TOUCHER", "ENABLETOUCHER allocation", "50 occupied slots; reject before sound/guard writes")
}
func (d *demoConnected) consumeTarget(index int) error {
	if err := d.preflightTarget(index); err != nil {
		return err
	}
	g := d.game
	if g.touchDisabled {
		return nil
	}

	g.sound("S_TOUCH2")
	g.touchDisabled = true
	if err := d.queue(demoTask{Site: "ENABLETOUCHER", Delay: 20, Action: demoEnableToucher}); err != nil {
		return err
	}
	if !g.Arcade {
		g.Arcade = true
		g.flash(7, 12, 0, false)
		g.flash(55, 12, 0, false)
	}
	d.trace = append(d.trace, "target:TOUCHER")
	return nil
}
