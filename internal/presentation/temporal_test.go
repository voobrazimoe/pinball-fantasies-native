package presentation

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

func TestDOSPrintTaskSlotAndVisibleSync(t *testing.T) {
	// FANTASIE _PRINT5/13 only install PRINTTASK; DO_THE_DOTMATRIX
	// calls that single slot after DO_THE_ANIMATIONS, then clears it.
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		var prints []Command
		for _, c := range d.Content.Commands {
			if c.Op == "_PRINT5" && len(d.Content.Texts[c.Arg(0)]) > 1 {
				prints = append(prints, c)
				if len(prints) == 2 {
					break
				}
			}
		}
		if len(prints) != 2 {
			t.Fatal("missing source prints", table)
		}
		for _, c := range prints {
			d.BeginCommand(c)
		}
		if d.Dots != [DotWidth * DotHeight]bool{} {
			t.Fatal("external handler drew before matrix sync", table)
		}
		d.BeginCommand(Command{Op: "_WAIT", Args: []string{"1"}, Nums: map[int]int{0: 1}})
		d.FlushPrint(nil)
		want := original(t, table)
		c := prints[1]
		x, y := PositionValue(c.Num(1))
		want.Text(strings.TrimRight(want.SourceText(c.Arg(0)), "\x00"), x, y, 5)
		if d.Dots != want.Dots {
			t.Fatal("PRINTTASK did not retain the last handler's slot", table)
		}
		d.Clear()
		d.FlushPrint(nil)
		if d.Dots != [DotWidth * DotHeight]bool{} {
			t.Fatal("PRINTTASK was not reset to DUMRET", table)
		}
	}
}

// FANTASIE:_FLASHON writes speed/count/phase immediately; MATRIX_BLINKOR
// decrements a word once per electronics sync. KILL_FLASHOR changes neither
// counter nor phase. MATRIXON_/OFF_ only emit palette, not the phase word.
func TestDOSFlashWordAndPaletteTimeline(t *testing.T) {
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			d := original(t, table)
			d.On = false
			d.BeginCommand(Command{Op: "_FLASHON", Args: []string{"3"}, Nums: map[int]int{0: 3}})
			if d.On || !d.flashing || !d.flashPhase || d.flashCount != 3 {
				t.Fatal("dispatch changed palette or failed to install flash")
			}
			for tick := 1; tick <= 12; tick++ {
				d.Flash()
				want := tick >= 6 && tick < 9 || tick >= 12
				if d.On != want {
					t.Fatalf("sync %d on=%v want=%v", tick, d.On, want)
				}
			}
			d.BeginCommand(Command{Op: "_MATRIXLGT", Args: []string{"0"}, Nums: map[int]int{0: 0}})
			if !d.flashPhase || !d.flashing || d.On {
				t.Fatal("MATRIXOFF changed blink phase")
			}
			d.Flash()
			d.Flash()
			d.Flash()
			if d.On {
				t.Fatal("next blink must use phase, not current palette")
			}
			count, phase := d.flashCount, d.flashPhase
			d.KillFlash()
			if d.flashCount != count || d.flashPhase != phase || !d.On || d.flashing {
				t.Fatal("KILL_FLASHOR writes")
			}
			for tick := 0; tick < 1000; tick++ {
				d.Flash()
				if !d.On {
					t.Fatal("flash leaked into idle")
				}
			}
			// Speed zero is an enabled 16-bit counter, not an inactive sentinel.
			d.BeginCommand(Command{Op: "_FLASHON", Args: []string{"0"}, Nums: map[int]int{0: 0}})
			for i := 0; i < 65535; i++ {
				d.Flash()
				if !d.On {
					t.Fatal("zero speed toggled before word wrap")
				}
			}
			d.Flash()
			if d.On {
				t.Fatal("zero speed did not toggle on wrap")
			}
		})
	}
}

// Linked ANIM 72c2..72f3 compares old BX with length before drawing, reloads
// the old frame duration, and leaves old BX intact when restarting a loop.
// Expand source data to absolute draw syncs independently of StepAnimation.
func TestEveryOriginalAnimationDrawSync(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		for name, a := range d.Content.Animations {
			t.Run(fmt.Sprintf("%d/%s", table, name), func(t *testing.T) {
				type draw struct{ tick, index int }
				var trace []draw
				index, loops, tick := 0, int(a.Header[1]), 1
				for {
					if index == int(a.Header[2]) {
						loops--
						if loops == 0 {
							break
						}
						indexNext := int(a.Header[0]) + 4
						trace = append(trace, draw{tick, index / 4})
						tick += int(a.Durations[index/4])
						index = indexNext
					} else {
						trace = append(trace, draw{tick, index / 4})
						tick += int(a.Durations[index/4])
						index += 4
					}
					if tick > 100000 {
						t.Fatal("invalid source animation")
					}
				}
				live, want := original(t, table), original(t, table)
				live.BeginCommand(Command{Op: "_ANIMATION", Args: []string{name}})
				frame, remaining, timer := uint16(0), a.Header[1], uint16(1)
				next := 0
				for sync := 1; sync <= tick; sync++ {
					if next < len(trace) && trace[next].tick == sync {
						want.Bitmap(a.Offsets[trace[next].index])
						next++
					}
					done := live.StepAnimation(a, &frame, &remaining, &timer)
					if live.Dots != want.Dots {
						t.Fatalf("sync %d old-BX draw mismatch", sync)
					}
					if done != (sync == tick) {
						t.Fatalf("sync %d completion=%v source end=%d", sync, done, tick)
					}
				}
			})
		}
	}
}

// Linked SCROLLE 71ac tests byte [SI+20] before each of the two subcalls;
// 725e decrements the shared phase after drawing and advances SI on zero.
func TestEverySourceScrollBoundaryAndInheritedPhase(t *testing.T) {
	for table := 1; table <= 4; table++ {
		d := original(t, table)
		seen := map[string]bool{}
		for _, c := range append(append([]Command{}, d.Content.Commands...), d.Content.Attract...) {
			if c.Op != "_SCROLL" || seen[c.Arg(0)] {
				continue
			}
			seen[c.Arg(0)] = true
			t.Run(fmt.Sprintf("%d/%s", table, c.Arg(0)), func(t *testing.T) {
				text := d.Content.Texts[c.Arg(0)]
				end := -1
				for i, b := range text {
					if b == 255 {
						end = i
						break
					}
				}
				if end < 20 {
					t.Fatal("missing SI+20 terminator")
				}
				for initial := uint8(1); initial <= 8; initial++ {
					live, want := original(t, table), original(t, table)
					live.scrollPhase = initial
					live.scrollPlane = [2]int{int(initial % 2), 1 - int(initial%2)}
					live.BeginCommand(c)
					left := uint16(end - 20)
					phase, si, offset := int(initial), 0, 0
					// Independent VGA plane memory model of the linked literal
					// MOV [BX+disp],AL/AH glyph routines and selector XOR 5.
					var memory [2][4096]bool
					bases := [2]int{168, 168}
					masks := [2]int{1, 4}
					if initial%2 != 0 {
						masks = [2]int{4, 1}
					}
					for tick := 1; tick <= 4*len(text)+1; tick++ {
						done := false
						for sub := 0; sub < 2; sub++ {
							if text[si+20] == 255 {
								done = true
								break
							}
							oracleScrollStores(table, want.data, text[si:si+21], bases, masks, &memory)
							masks[0] ^= 5
							masks[1] ^= 5
							if masks[1] == 1 {
								bases[0]--
							} else {
								bases[1]--
							}
							offset++
							phase--
							if phase == 0 {
								phase = 8
								si++
								bases = [2]int{168, 168}
							}
						}
						for y := 0; y < 16; y++ {
							for x := 0; x < 160; x++ {
								want.Dots[y*160+x] = memory[x%2][168+y*168+x/2]
							}
						}
						if live.ScrollCompletes(left) != done {
							t.Fatal("scroll preflight differs from source terminator", tick, initial, left)
						}
						got := live.StepScroll(&left)
						if got != done || live.Dots != want.Dots || live.scrollOffset != offset || int(live.scrollPhase) != phase {
							t.Fatalf("phase %d sync %d scroll state differs", initial, tick)
						}
						if done {
							break
						}
					}
				}
			})
		}
	}
}

func TestGameOverDOSCommandSyncs(t *testing.T) {
	// PLAND AfterDemoModeTS/Once_MoreTS/UrbanOverTS, FANTASIE WAITRUT
	// and PLAND _LOOP_: taken loop tail-calls INIT_SCORE on the same sync.
	expected := map[int]string{0: "_MATRIXLGT", 1: "_WAIT", 21: "_SETLOOP", 22: "_INIT_SCORE", 23: "_CLEAR4", 28: "_FLASHON", 29: "_PRINT13", 30: "_WAIT", 60: "_FLASHOFF", 61: "_CLEAR4", 66: "_SHOW_SCORE", 67: "_INIT_SCORE", 68: "_WAIT", 148: "_CLEAR4", 153: "_FLASHON", 154: "_PRINT13", 155: "_WAIT", 185: "_FLASHOFF", 186: "_CLEAR4", 191: "_SHOW_SCORE", 192: "_LOOP_", 193: "_WAIT", 273: "_CLEAR4", 278: "_WAIT"}
	d := original(t, 1)
	r := replay{d: d, commands: d.Content.Commands, pc: d.Content.Labels["AFTERDEMOMODETS"], scores: []string{"123456"}, active: true}
	r.dispatch()
	for sync := 0; sync <= 298; sync++ {
		if sync > 0 {
			r.step()
		}
		if op, ok := expected[sync]; ok && d.op != op {
			t.Fatalf("sync %d got %s source %s", sync, d.op, op)
		}
		if r.handoff != (sync == 298) {
			t.Fatalf("sync %d handoff=%v", sync, r.handoff)
		}
	}
}

func TestZeroWordAndRoutineCompletionAreDistinct(t *testing.T) {
	d := original(t, 1)
	d.BeginCommand(Command{Op: "_FLASHON", Args: []string{"1"}, Nums: map[int]int{0: 1}})
	d.BeginCommand(Command{Op: "0"})
	d.Flash()
	if d.On || !d.flashing {
		t.Fatal("zero data word invented cleanup")
	}
	d.FinishRoutine(true)
	if d.On || !d.flashing {
		t.Fatal("completion with NEXT_A killed deliberate flash")
	}
	d.FinishRoutine(false)
	if !d.On || d.flashing {
		t.Fatal("ts_slut failed source cleanup")
	}
}

// Source SCROLLE calls the literal glyph routines selected from DATA tables.
// Keep raw VGA byte-address memory here, rather than native dot coordinates.
func oracleScrollStores(table int, data, text []byte, bases, masks [2]int, memory *[2][4096]bool) {
	i := table - 1
	ds := [4]int{0x19d40, 0x18ee0, 0x18b60, 0x166d0}[i]
	shift := [4]int{0, 0x90, -0x720, 0x5f0}[i]
	tables := [2]int{0x5e00 + shift, 0x6000 + shift}
	code := [2]int{[4]int{0x7c00, 0x73f0, 0x6e90, 0x83a0}[i] + 0x300, [4]int{0x7090, 0x6880, 0x6320, 0x7830}[i] + 0x300}
	for glyph := 1; glyph >= 0; glyph-- {
		bank := 0
		if masks[glyph] == 4 {
			bank = 1
		}
		for char, v := range text {
			q := code[glyph] + int(binary.LittleEndian.Uint16(data[ds+tables[glyph]+2*int(v):]))
			for data[q] != 0xc3 {
				address := bases[glyph] + char*4 + int(binary.LittleEndian.Uint16(data[q+2:]))
				memory[bank][address] = data[q+1] == 0x87
				q += 4
			}
		}
	}
}

func TestDOSScoreCacheAndCommaStores(t *testing.T) {
	// Linked CODE2 SCORE+65h skips leading zero BCD cells, then starts BP
	// at oldbuf. SCORE+135h returns when the comma count equals lastcommas.
	// FANTASIE.MAC UPDAT_SCORE writes 12h to oldbuf and zero to lastcommas.
	for table := 1; table <= 4; table++ {
		t.Run(fmt.Sprint(table), func(t *testing.T) {
			d := original(t, table)
			d.Score("1234567")
			first := d.Dots
			d.Clear() // VGA erasure alone is not UPDAT_SCORE.
			d.Score("1234567")
			if d.Dots != [DotWidth * DotHeight]bool{} {
				t.Fatal("cached digits/commas redrawn without source invalidation")
			}
			d.InvalidateScore()
			d.Score("1234567")
			if d.Dots != first {
				t.Fatal("UPDAT_SCORE did not restore all digit and comma stores")
			}
			// Zero groups terminate the comma scanner without touching its cache.
			d.Score("0")
			if d.scoreCommas != 2 {
				t.Fatal("zero-group exit changed lastcommas")
			}
			d.BeginCommand(Command{Op: "_WAIT", Nums: map[int]int{0: 1}})
			if d.scoreCache[0] != 0x12 || d.scoreCommas != 0 {
				t.Fatal("DO_SPEC_MATRIX omitted UPDAT_SCORE")
			}
		})
	}
}
