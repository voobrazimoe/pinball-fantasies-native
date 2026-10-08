package datalayout

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Fixture-free: the descriptor stays inside the decoded space and carries
// exactly the two reviewed typed jingle differences.
func TestPartyLandDemoDescriptor(t *testing.T) {
	l := partyLandDemoLinked()
	if l.DecodedSize != 536822 || len(l.Copies) == 0 {
		t.Fatal(l.DecodedSize, len(l.Copies))
	}
	for _, c := range l.Copies {
		if c.Destination < 0 || c.Size <= 0 || c.Destination+c.Size > l.DecodedSize || c.Source < 0 {
			t.Fatal(c)
		}
	}
	want := map[string][3]uint8{"S_EMPTY": {62, 0, 0}, "S_GAMEOVER2": {13, 0, 255}}
	if len(l.Jingles) != len(want) {
		t.Fatal(l.Jingles)
	}
	for _, j := range l.Jingles {
		if want[j.Role] != [3]uint8{j.Position, j.Repeat, j.Priority} {
			t.Fatal(j)
		}
	}
	// Every hashed canonical consumer must map into the demo file.
	p := partyLandDemoProfile(true)
	if len(p.Regions) != len(profiles["TABLE1.PRG"].Regions)+len(partyLandDemo.Table1.Reviewed)+len(partyLandDemo.Table1.Records) {
		t.Fatal(len(p.Regions))
	}
	if _, err := DetectInstallation(map[string][]byte{"INTRO.PRG": nil, "TABLE1.PRG": nil}); err == nil {
		t.Fatal("demo pair must not satisfy a full installation")
	}
	if _, err := DetectDemoInstallation(nil, nil); err == nil {
		t.Fatal("empty demo accepted")
	}
}

func privateDemo(t *testing.T) (intro, table []byte) {
	t.Helper()
	dir := os.Getenv("PF_10MIN_DEMO_DATA")
	if dir == "" {
		t.Skip("set PF_10MIN_DEMO_DATA to the private 10-minute demo")
	}
	var err error
	if intro, err = os.ReadFile(filepath.Join(dir, "INTRO.PRG")); err != nil {
		t.Fatal(err)
	}
	if table, err = os.ReadFile(filepath.Join(dir, "TABLE1.PRG")); err != nil {
		t.Fatal(err)
	}
	return intro, table
}

func TestPrivatePartyLandDemoProfile(t *testing.T) {
	intro, table := privateDemo(t)
	id, err := DetectDemoInstallation(intro, table)
	if err != nil || id != PartyLandDemoProfile {
		t.Fatal(id, err)
	}
	out, err := PreparePRGForProfile(id, "TABLE1.PRG", table)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDecoded("TABLE1.PRG", out); err != nil {
		t.Fatal(err)
	}
	if err := Validate("TABLE1.PRG", out); err == nil {
		t.Fatal("demo jingles must not pass as canonical A")
	}
	rec, err := DemoRecords(table)
	if err != nil || len(rec.PlayersText) != 12 || len(rec.ExpiryTexts[0]) != 99 || len(rec.ExpiryTexts[1]) != 89 {
		t.Fatal(err, rec)
	}
	for _, r := range append(append([]Region(nil), partyLandDemo.Table1.Reviewed...), partyLandDemo.Intro.Regions[4]) {
		src := table
		if r.Purpose[:4] == "FORM" {
			src = intro
		}
		bad := bytes.Clone(src)
		bad[r.Offset+r.Size/2] ^= 1
		if r.Purpose[:4] == "FORM" {
			_, err = DetectDemoInstallation(bad, table)
		} else {
			_, err = DetectDemoInstallation(intro, bad)
		}
		if err == nil {
			t.Fatal("mutation accepted", r.Purpose)
		}
	}
	if _, err := DetectDemoInstallation(table, intro); err == nil {
		t.Fatal("swapped roles accepted")
	}
	// With canonical A supplied, the decoded demo equals A except the typed
	// jingles and the unconsumed PLAYERSTEXT slot.
	dir := os.Getenv("PF_RUNTIME_DATA")
	if dir == "" {
		return
	}
	a, err := os.ReadFile(filepath.Join(dir, "TABLE1.PRG"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range profiles["TABLE1.PRG"].Regions {
		if r.Purpose == "matrix record S_EMPTY" || r.Purpose == "matrix record S_GAMEOVER2" || r.Purpose == "matrix record PLAYERSTEXT" {
			continue
		}
		if !bytes.Equal(out[r.Offset:r.Offset+r.Size], a[r.Offset:r.Offset+r.Size]) {
			t.Fatal("region differs", r.Purpose)
		}
	}
}
