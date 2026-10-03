package presentation

import "strconv"

// countdown retains the original SEC_ASC bytes and SYNC_LEFT word across
// matrix replacement. The units byte starts one higher than the operand.
// COUNTDOWN displays zero for a full second before uninstalling the routine.
type countdown struct {
	text [2]byte
	left uint16
}

func (d *Display) StartCountdown(tens, units int) {
	d.countdown.text = [2]byte{byte('7' + tens), byte('8' + units)}
	d.ContinueCountdown()
}
func (d *Display) ContinueCountdown() { d.countdown.left = 1 }

// CountdownRemaining is the number of source visits until the zero callback,
// retained as ModeTime for gameplay observers. It is not the uninstall timer.
func (d *Display) CountdownRemaining() uint16 {
	tens := int(d.countdown.text[0] - '7')
	if d.countdown.text[0] == '*' {
		tens = 0
	}
	seconds := tens*10 + int(d.countdown.text[1]-'7')
	if seconds == 0 {
		return 0
	}
	return uint16((seconds-1)*71) + d.countdown.left
}

// StepCountdown models FANTASIE COUNTDOWN/INH_CD/uninstall_count. Only one
// PRINTTASK is installed per visit; READ_SPECIAL_MODE_COUNTER runs before the
// seconds print is installed, and may replace the running matrix program.
func (d *Display) StepCountdown(value string, inhibit bool, onSecond func(int)) bool {
	c := &d.countdown
	if !inhibit {
		c.left--
		if c.left == 0 {
			c.left = 71
			if c.text == [2]byte{'*', '7'} {
				return true
			}
			c.text[1]--
			if c.text[1] == '6' {
				c.text[1] += 10
				c.text[0]--
			}
			if c.text[0] == '7' {
				c.text[0] = '*'
			}
			tens := int(c.text[0] - '7')
			if c.text[0] == '*' {
				tens = 0
			}
			if onSecond != nil {
				onSecond(tens*10 + int(c.text[1]-'7'))
			}
			// SEC_ASC is a private DATA buffer, not extracted artwork.
			d.Content.Texts["SEC_ASC"] = append([]byte{c.text[0], c.text[1]}, 0)
			d.pendingPrint = &Command{Op: "_PRINT11", Args: []string{"SEC_ASC", "576"}, Nums: map[int]int{1: 576}}
			return false
		}
	}
	d.pendingPrint = &Command{Op: "_PRINT13_NUMBER", Args: []string{value, "344"}, Nums: map[int]int{1: 344}}
	return false
}

// CountdownText exposes the source digits for diagnostics without a new clock.
func (d *Display) CountdownText() string {
	tens := d.countdown.text[0]
	if tens == '*' {
		return " " + strconv.Itoa(int(d.countdown.text[1]-'7'))
	}
	return strconv.Itoa(int(tens-'7')) + strconv.Itoa(int(d.countdown.text[1]-'7'))
}
