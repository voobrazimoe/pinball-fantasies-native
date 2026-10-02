package physics

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"pinballfantasies/internal/assets"
	"pinballfantasies/internal/settings"
	"pinballfantasies/internal/testinputs"
	"testing"
)

// PUTTHEBALL supplies top-left SC_X/SC_Y; BREDDA_MASK reads HID1 at
// HIDDEN1:0580 (UNDANSPR 256 + HIDDA 1152). Stones' linked ES is 287Bh,
// so the file address is 200h+287Bh*10h+580h = 28f30h, not 28eb0h.
func TestSourceInitialChutesBothModes(t *testing.T) {
	decoders := []func([]byte) (*Table, error){DecodePartyLand, DecodeSpeedDevils, DecodeGameshow, DecodeStones}
	for i, decode := range decoders {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			testinputs.Require(t, fmt.Sprintf("../../TABLE%d.PRG", i+1))
			raw, err := os.ReadFile(fmt.Sprintf("../../TABLE%d.PRG", i+1))
			if err != nil {
				t.Fatal(err)
			}
			table, err := decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			call := bytes.Index(raw, []byte{0xbd, 0x96, 0xfb, 0x68})
			if call < 0 {
				t.Fatal("PUTTHEBALL source signature")
			}
			segment := int(raw[call+4]) + int(raw[call+5])*256
			maskStart := 512 + 16*segment + 1408
			for _, res := range []byte{0, 1, 2} {
				g := New(table)
				cfg := settings.Legacy()
				cfg.Resolution = res
				if res == 2 {
					cfg.Resolution = 1
					cfg.ScrollMode = settings.ScrollOff
				}
				g.Configure(cfg, table.gravity)
				palette := assets.VGAPalette(table.Initial.Playfield.Palette)
				for _, pos := range []image.Point{table.start, image.Pt(301, 537)} {
					// The settled Stones position caught the old diagonal mask corruption.
					// Only Stones needs that regression; other tables retain their source start.
					if pos != table.start && i != 3 {
						continue
					}
					g.SetBall(int16(pos.X), int16(pos.Y), 10, 0, false)
					frame := g.FramePalette(palette)
					visible := 0
					masked := 0
					for n, c := range table.Initial.Ball {
						if c == 0 {
							continue
						}
						x, y := pos.X+n%16, pos.Y+n/16
						bit := raw[maskStart+y*40+x/8] & (128 >> uint(x&7))
						xx, yy := x, y-int(g.BottomRaster())
						if res == 2 {
							yy = y
						}
						off := frame.PixOffset(xx, yy)
						if bit == 0 {
							visible++
							want := palette[int(c)*3 : int(c)*3+3]
							if !bytes.Equal(frame.Pix[off:off+3], want) {
								t.Fatalf("res %d pos %v ball pixel %d,%d clipped", res, pos, x, y)
							}
						} else {
							masked++
							index := int(g.indices[y*320+x])
							want := palette[index*3 : index*3+3]
							if !bytes.Equal(frame.Pix[off:off+3], want) {
								t.Fatal("source foreground was overwritten", res, pos, x, y)
							}
						}
					}
					if i == 3 && pos == table.start && (visible != 152 || masked != 25) {
						t.Fatal("source frozen SETBALL visibility", visible, masked)
					}
					if i == 3 && pos == image.Pt(301, 537) && visible != 177 {
						t.Fatalf("settled sprite visible=%d masked=%d", visible, masked)
					}
					if dir := os.Getenv("PF111_VISUAL_DIR"); dir != "" {
						os.MkdirAll(dir, 0755)
						f, e := os.Create(filepath.Join(dir, fmt.Sprintf("f%d-mode%d-chute-%d-%d.png", i+1, res, pos.X, pos.Y)))
						if e != nil {
							t.Fatal(e)
						}
						png.Encode(f, frame)
						f.Close()
					}
				}
			}
		})
	}
}
