package engine

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"os"
	"strings"
	"testing"

	"pinballfantasies/internal/demodata"
	"pinballfantasies/internal/frontend"
	"pinballfantasies/internal/partyland"
	"pinballfantasies/internal/settings"
)

// The bundled Party Land demo is public, so this golden run is the one
// end-to-end regression check every CI job can execute: INTRO, selector,
// a played table with launches and drains, the demo timer expiring, and
// INTRO starting again. Frames and PCM are hashed at fixed checkpoints so a
// failure names the first stretch that changed.
//
// After an intended change in behaviour, regenerate with
// PF_UPDATE_DEMO_GOLDEN=1 go test -run TestDemoGolden -v ./internal/engine
// and paste the printed table.
const (
	demoGoldenSteps      = 40000
	demoGoldenCheckpoint = 5000
)

var demoGolden = map[settings.ScrollMode][]string{
	settings.ScrollHard: {
		"frames dc7f52ae6085 pcm f8122cbfc436",
		"frames fe5ac4a0d15f pcm 150cdb7de271",
		"frames b579bf5686c6 pcm 9430fdfddce7",
		"frames f96ad7330f16 pcm 012d58e64e36",
		"frames 55b8ac7b49cb pcm 555fc53f99c1",
		"frames 74df3bfb1615 pcm 99a15c49fc80",
		"frames 2319074168a4 pcm d1003835ae45",
		"frames a074a9d28c82 pcm 2f7a60bc6ff9",
	},
	settings.ScrollMedium: {
		"frames de26d7f4d620 pcm f8122cbfc436",
		"frames 97fe876ce97e pcm 150cdb7de271",
		"frames b90fc322e471 pcm 9430fdfddce7",
		"frames a27ae2f0c3ba pcm 012d58e64e36",
		"frames 9eb8b9b9aa4c pcm 555fc53f99c1",
		"frames 25fdf6eddf75 pcm 99a15c49fc80",
		"frames 5b670056f22d pcm d1003835ae45",
		"frames 9d45bd6a0a06 pcm 2f7a60bc6ff9",
	},
	settings.ScrollSoft: {
		"frames d84bb17505a9 pcm f8122cbfc436",
		"frames 8bdb11abadb8 pcm 150cdb7de271",
		"frames a18b32dc4922 pcm 9430fdfddce7",
		"frames 33431346bddb pcm 012d58e64e36",
		"frames d48bca26e51e pcm 555fc53f99c1",
		"frames 83d554f72d1b pcm 99a15c49fc80",
		"frames 1664d5ac0707 pcm d1003835ae45",
		"frames 603667cd698d pcm 2f7a60bc6ff9",
	},
	settings.ScrollOff: {
		"frames d3da13b527ce pcm f8122cbfc436",
		"frames 9abd021b88f4 pcm 150cdb7de271",
		"frames 5fcaff47af5e pcm 9430fdfddce7",
		"frames 2bff84cc2b63 pcm 012d58e64e36",
		"frames 8239f711d0b1 pcm 555fc53f99c1",
		"frames 140fa4383e99 pcm 99a15c49fc80",
		"frames c5c25fb67631 pcm d1003835ae45",
		"frames 5fd1764c0da0 pcm 2f7a60bc6ff9",
	},
}

func TestDemoGolden(t *testing.T) {
	update := os.Getenv("PF_UPDATE_DEMO_GOLDEN") != ""
	var table strings.Builder
	for _, scroll := range []settings.ScrollMode{settings.ScrollHard, settings.ScrollMedium, settings.ScrollSoft, settings.ScrollOff} {
		t.Run(fmt.Sprintf("scroll%d", scroll), func(t *testing.T) {
			got := runDemoGolden(t, scroll)
			fmt.Fprintf(&table, "\tsettings.Scroll%s: {\n", scrollName(scroll))
			for _, h := range got {
				fmt.Fprintf(&table, "\t\t%q,\n", h)
			}
			table.WriteString("\t},\n")
			if update {
				return
			}
			want := demoGolden[scroll]
			if len(want) != len(got) {
				t.Fatalf("golden has %d checkpoints, run produced %d; regenerate with PF_UPDATE_DEMO_GOLDEN=1", len(want), len(got))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("first divergence before step %d: got %s, want %s", (i+1)*demoGoldenCheckpoint, got[i], want[i])
				}
			}
		})
	}
	if update {
		t.Logf("var demoGolden = map[settings.ScrollMode][]string{\n%s}", table.String())
	}
}

func scrollName(s settings.ScrollMode) string {
	return [...]string{"Hard", "Medium", "Soft", "Off"}[s]
}

// runDemoGolden plays a fixed input script against the bundled demo and
// returns the running frame and PCM digests at each checkpoint.
func runDemoGolden(t *testing.T, scroll settings.ScrollMode) []string {
	t.Helper()
	dir, cleanup, err := demodata.Extract()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	e, err := Load(dir, t.TempDir(), 0)
	if err != nil {
		t.Fatal(err)
	}
	e.runner.Runtime.Model.Settings.ScrollMode = scroll

	frames, pcm := sha256.New(), sha256.New()
	var digests []string
	played, restarted := false, false
	drains, launches := 0, 0
	lost, valid := false, false
	for step := 0; step < demoGoldenSteps; step++ {
		mode := frontend.Mode(e.State().Mode)
		switch mode {
		case frontend.Startup, frontend.Selector, frontend.SelectorText:
			// Space leaves the startup pictures and pages; F1 picks Party Land,
			// which the demo starts as a one-player game at once.
			if step%200 == 199 {
				if step%400 == 399 {
					e.Key(uint8(frontend.F1))
				} else {
					e.Key(uint8(frontend.Space))
				}
			}
			if played {
				restarted = true
			}
		case frontend.Playing:
			played = true
			mustAction(t, e, Left, step%93 < 18)
			mustAction(t, e, Right, step%71 < 14)
			mustAction(t, e, Spring, step%300 < 40)
		default:
			if step%50 == 49 {
				e.Key(uint8(frontend.Enter))
			}
		}
		if g, ok := e.runner.Runtime.Model.Session.(*partyland.Game); ok {
			if g.Physics.Ball.Lost && !lost {
				drains++
			}
			if !g.Physics.SpringValid && valid {
				launches++
			}
			lost, valid = g.Physics.Ball.Lost, g.Physics.SpringValid
		}
		if err := e.Advance(e.runner.Next.UnixNano(), func(p []byte) error { pcm.Write(p); return nil }); err != nil {
			t.Fatal(err)
		}
		if e.State().Done {
			t.Fatalf("demo program ended at step %d", step)
		}
		if step%3 == 0 {
			// Every 50th sampled frame also covers the full-table presentation.
			e.SetPresentation(step%150 == 0)
			hashFrame(frames, e)
		}
		if (step+1)%demoGoldenCheckpoint == 0 {
			digests = append(digests, "frames "+hex.EncodeToString(frames.Sum(nil)[:6])+" pcm "+hex.EncodeToString(pcm.Sum(nil)[:6]))
		}
	}
	t.Logf("drains=%d launches=%d restarted=%t", drains, launches, restarted)
	if !played || !restarted || drains == 0 || launches == 0 {
		t.Fatalf("script no longer covers the demo: played=%t restarted=%t drains=%d launches=%d", played, restarted, drains, launches)
	}
	return digests
}

func mustAction(t *testing.T, e *Engine, a Action, down bool) {
	t.Helper()
	if err := e.SetAction(a, down); err != nil {
		t.Fatal(err)
	}
}

var castagnoli = crc32.MakeTable(crc32.Castagnoli)

// hashFrame feeds the frame size and a hardware CRC of its pixels into h;
// hashing every sampled frame with SHA-256 would dominate the run.
func hashFrame(h hash.Hash, e *Engine) {
	f := e.Frame()
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:4], uint32(f.Rect.Dx()))
	binary.LittleEndian.PutUint32(b[4:8], uint32(f.Rect.Dy()))
	binary.LittleEndian.PutUint32(b[8:], crc32.Checksum(f.Pix, castagnoli))
	h.Write(b[:])
}
