package gameplay

import "testing"

func cheatScan(char byte) uint8 {
	if char == '*' {
		return 57
	}
	for _, first := range []uint8{16, 30, 44} {
		for scan := first; scan < first+10; scan++ {
			d := CheatDecoder{}
			d.Make(scan)
			if d.input == string(char) {
				return scan
			}
		}
	}
	panic("missing cheat scan")
}

func TestSourceCheatTableAndPrefixRecovery(t *testing.T) {
	// FANTASIE/CHEATS order, CHECKCHEAT/FOUND and ENDOFCHEATS.
	for _, cheat := range OriginalCheats {
		d := CheatDecoder{}
		for i := range cheat.Word {
			got := d.Make(cheatScan(cheat.Word[i]))
			if i+1 < len(cheat.Word) && got != "" || i+1 == len(cheat.Word) && got != cheat.Program {
				t.Fatalf("%s character %d: %q", cheat.Word, i, got)
			}
		}
	}
	d := CheatDecoder{}
	d.Make(16) // Q is not a prefix; retain this last character.
	d.Make(18) // E replaces Q, then the remaining EARTHQUAKE matches.
	for _, c := range []byte("ARTHQUAK") {
		d.Make(cheatScan(c))
	}
	for _, invalid := range []uint8{1, 28, 55, 59, 127, 128} {
		if d.Make(invalid) != "" || d.input != "EARTHQUAK" {
			t.Fatal("ALFA_KEYS invalid key changed prefix", invalid)
		}
	}
	if d.Make(18) != "QUAKETS" || d.input != "" {
		t.Fatal("FOUND must empty the prefix")
	}
}

func TestSourceCheatNearMisses(t *testing.T) {
	for _, cheat := range OriginalCheats {
		for index := range cheat.Word {
			wrong := []byte(cheat.Word)
			wrong[index] = 'Z'
			if cheat.Word[index] == 'Z' {
				wrong[index] = 'X'
			}
			d := CheatDecoder{}
			for _, c := range wrong {
				if got := d.Make(cheatScan(c)); got != "" {
					t.Fatalf("near miss %q activated %s", wrong, got)
				}
			}
		}
	}
}
