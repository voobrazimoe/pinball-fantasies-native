package physics

var ring = [44][2]int16{
	{16, 8}, {16, 9}, {16, 10}, {15, 11}, {15, 12}, {14, 13}, {13, 14}, {12, 15}, {11, 15}, {10, 16}, {9, 16}, {8, 16}, {7, 16}, {6, 16}, {5, 15}, {4, 15}, {3, 14}, {2, 13}, {1, 12}, {1, 11}, {0, 10}, {0, 9}, {0, 8}, {0, 7}, {0, 6}, {1, 5}, {1, 4}, {2, 3}, {3, 2}, {4, 1}, {5, 1}, {6, 0}, {7, 0}, {8, 0}, {9, 0}, {10, 0}, {11, 1}, {12, 1}, {13, 2}, {14, 3}, {15, 4}, {15, 5}, {16, 6}, {16, 7},
}
var normalAngles = [44]uint16{0, 41, 80, 132, 169, 226, 286, 343, 380, 432, 471, 512, 553, 592, 644, 681, 738, 798, 855, 892, 944, 983, 1024, 1065, 1104, 1156, 1193, 1250, 1310, 1367, 1404, 1456, 1495, 1536, 1577, 1616, 1668, 1705, 1762, 1822, 1879, 1916, 1968, 2007}
var sampleOrder = [44]int{35, 34, 33, 32, 31, 37, 36, 30, 29, 38, 28, 39, 27, 40, 26, 41, 25, 42, 24, 43, 23, 0, 22, 1, 21, 2, 20, 3, 19, 4, 18, 5, 17, 6, 16, 7, 8, 14, 15, 9, 10, 11, 12, 13}

type contact struct {
	angle      uint16
	count      uint8
	material   uint8
	xadd, yadd int16
	object     int
	kind       EventKind
}

func (g *Game) collision() (contact, bool) {
	b := &g.Ball
	m := g.mask12
	if b.High {
		m = g.mask22
	}
	var sum uint16
	var quadrants uint8
	count, down, last := uint8(0), uint8(0), 0
	for _, i := range sampleOrder {
		x, y := b.PixelX-1+ring[i][0], b.PixelY-1+g.ScreenOffset+ring[i][1]
		if !bit(m, x, y) {
			continue
		}
		count++
		last = i
		sum += normalAngles[i]
		if i <= 10 {
			quadrants |= 1
			down++
		} else if i <= 22 {
			quadrants |= 2
			down++
		} else if i <= 32 {
			quadrants |= 4
		} else {
			quadrants |= 8
		}
	}
	if count == 0 {
		return contact{}, false
	}
	if quadrants == 11 || quadrants == 9 || quadrants == 13 {
		sum += uint16(down) << 11
	}
	angle := (sum / uint16(count)) & 2047
	b.CollisionAngle = angle
	b.ContactCount = count
	x, y := b.PixelX-1+ring[last][0], b.PixelY-1+g.ScreenOffset+ring[last][1]
	if y >= 576 {
		return contact{}, false
	}
	var mat uint8
	if b.High {
		if bit(g.Table.mask21, x, y) {
			mat |= 1
		}
		if bit(g.mask22, x, y) {
			mat |= 2
		}
		if bit(g.Table.mask23, x, y) {
			mat |= 4
		}
	} else {
		if bit(g.Table.mask11, x, y) {
			mat |= 1
		}
		if bit(g.mask12, x, y) {
			mat |= 2
		}
		if bit(g.Table.mask13, x, y) {
			mat |= 4
		}
	}
	b.Material = mat
	// Original rounded 1408*angle/65536 table index. Index 44 falls into
	// the adjacent SC_KV/ANTALPIX words; preserve that original edge case.
	idx := (uint32(angle)*1408 + 32768) >> 16
	hx, hy := int16(angle), int16(count)
	if idx < 44 {
		hx = ring[idx][0]
		hy = ring[idx][1]
	}
	b.HitX = b.PixelX + hx
	b.HitY = b.PixelY + hy
	c := contact{angle: angle, count: count, material: mat, object: -1}
	if !g.Tilted && (mat == 7 || mat == 3) {
		list := g.Table.bumper
		c.kind = EventBumperHit
		if mat == 3 {
			list = g.Table.kicker
			c.kind = EventSlingshotHit
		}
		for i, r := range list {
			// SEARCHBUMPER short-circuits when y is above a candidate region.
			if uint16(b.HitX) < uint16(r.X1) {
				continue
			}
			if uint16(b.HitY) < uint16(r.Y1) {
				break
			}
			if r.contains(b.HitX, b.HitY) {
				c.object = i
				g.pendingBumper = true
				g.pendingEvent = Event{Kind: c.kind, Object: i, X: b.HitX, Y: b.HitY}
				break
			}
		}
	}
	if mat == 2 {
		for i, f := range g.Flippers {
			if !f.Bounds.contains(b.HitX, b.HitY) {
				continue
			}
			c.xadd, c.yadd = flipperVelocity(f, b.HitX, b.HitY)
			g.Events = append(g.Events, Event{Kind: EventFlipperHit, Object: i, X: b.HitX, Y: b.HitY})
			break
		}
	}
	return c, true
}

func flipperVelocity(f Flipper, x, y int16) (int16, int16) {
	x -= f.CenterX
	if f.Kind == 1 {
		if x >= 0 {
			return 0, 0
		}
	} else if x < 0 {
		return 0, 0
	}
	y -= f.CenterY
	var a int16
	if f.PowerZone != 0 {
		x, y = y, x
		a = y >> 1
		if a < 0 {
			a = -a
		}
	} else {
		if f.Kind == 1 {
			x = -x
			y = -y
		}
		a = y
		if a < 0 {
			a = -a
		}
		a >>= 2
	}
	x = -(x + a)
	return f.Speed * y, f.Speed * x
}

func (g *Game) respond(c contact) error {
	b := &g.Ball
	m := g.Table.Materials[c.material]
	x, y := clamp(b.VX+c.xadd), clamp(b.VY+c.yadd)
	inv := (2048 - c.angle) & 2047
	s, co := g.Table.Sin[inv], g.Table.Sin[int(inv)+512]
	// The original two signed products are combined as wrapped 32-bit values,
	// shifted left three, then their signed high words are retained.
	n := int16((uint32(int32(x)*int32(co)-int32(y)*int32(s)) << 3) >> 16)
	t := int16((uint32(int32(x)*int32(s)+int32(y)*int32(co)) << 3) >> 16)
	if n <= 0 {
		g.pendingBumper = false
		return nil
	}
	n = -n
	boost := false
	if n < m.MinSpeed {
		a, err := divide(16*int32(t), n)
		if err != nil {
			return err
		}
		if a < 0 {
			a = -a
		}
		if uint16(a) < uint16(m.MaxAngle) {
			if g.pendingBumper {
				if c.material != 3 {
					n += -7000
					boost = true
				} else if n <= -300 {
					n += -2000
					boost = true
				}
			}
		} else {
			n = 0
		}
	} else {
		n = 0
	}
	if !boost {
		g.pendingBumper = false
	}
	q, err := divide(int32(n)*256, m.Bounce)
	if err != nil {
		return err
	}
	n -= q
	wf, bf := m.WallFriction, m.BallFriction
	if n >= -1023 {
		factor := (-n >> 6) + 1
		wf *= factor
		bf *= factor
	}
	spinSlip := b.Rotation + g.ScreenSpeed - t
	deltaWall, deltaBall := spinSlip, spinSlip
	if wf != 0 {
		deltaWall, err = divide(int32(spinSlip)*256, wf)
		if err != nil {
			return err
		}
	} else {
		deltaWall = int16(int32(spinSlip) * 256)
	}
	if bf != 0 {
		deltaBall, err = divide(int32(spinSlip)*256, bf)
		if err != nil {
			return err
		}
	} else {
		deltaBall = int16(int32(spinSlip) * 256)
	}
	t += deltaWall
	b.Rotation -= deltaBall
	t, err = divide(int32(t)*2048, 2049)
	if err != nil {
		return err
	}
	s, co = g.Table.Sin[c.angle], g.Table.Sin[int(c.angle)+512]
	b.VX = clamp(int16((uint32(int32(n)*int32(co)-int32(t)*int32(s))<<1)>>16) - c.xadd)
	b.VY = clamp(int16((uint32(int32(n)*int32(s)+int32(t)*int32(co))<<1)>>16) - c.yadd)
	if c.count >= 6 {
		b.X += int32(highProduct(-1024, co))
		b.Y += int32(highProduct(-1024, s))
	}
	return nil
}
