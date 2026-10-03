package gameplay

import "strings"

// CheatDecoder is FANTASIE/CHECKCHEAT's twelve-character prefix buffer.
// Space is ALFA_KEYS' '*'; keypad multiply is deliberately not a valid key.
// Invalid make codes leave the partially typed word untouched.
type CheatDecoder struct{ input string }

type Cheat struct{ Word, Program string }

var OriginalCheats = [...]Cheat{
	{"JOHAN", "JOHANTS"}, {"TECH", "TECHTS"}, {"TSP", "TSPTS"},
	{"DANIEL", "DANIELTS"}, {"GABRIEL", "GABRIELTS"}, {"CHEAT", "CHEATTS"},
	{"EARTHQUAKE", "QUAKETS"}, {"EXTRA*BALLS", "BALLSTS"}, {"SNAIL", "SNAILTS"},
	{"FAIR*PLAY", "FAIRPLAYTS"}, {"ROBBAN", "ROBBANTS"}, {"STEIN", "STEINTS"},
	{"GREET", "GREETTS"},
}

func (d *CheatDecoder) Make(scan uint8) string {
	var char byte
	switch {
	case scan >= 16 && scan <= 25:
		char = "QWERTYUIOP"[scan-16]
	case scan >= 30 && scan <= 38:
		char = "ASDFGHJKL"[scan-30]
	case scan >= 44 && scan <= 50:
		char = "ZXCVBNM"[scan-44]
	case scan == 57:
		char = '*'
	default:
		return ""
	}
	d.input += string(char)
	if len(d.input) > 12 {
		d.input = ""
		return ""
	}
	partial := false
	for _, cheat := range OriginalCheats {
		if strings.HasPrefix(d.input, cheat.Word) {
			d.input = ""
			return cheat.Program
		}
		partial = partial || strings.HasPrefix(cheat.Word, d.input)
	}
	if !partial {
		// ENDOFCHEATS carries the last valid character; it does not retry it.
		d.input = string(char)
	}
	return ""
}
