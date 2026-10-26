// Typed diagonal cursor; no guest instruction dispatcher.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planLineCursor(v CPUView, budget uint64) *machine.HostWorkPlan {
	steep := false
	decrease := false
	switch v.IP() {
	case 0x309c:
		steep = true
	case 0x30b9:
		steep = true
		decrease = true
	case 0x321c:
	case 0x3206:
		decrease = true
	default:
		return nil
	}
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 || budget == 0 {
		return nil
	}
	expectedDF := uint16(0)
	if !steep && decrease {
		expectedDF = cpu.DF
	}
	if v.Flags()&cpu.DF != expectedDF {
		return nil
	}
	var code [1338]byte
	copyView(v, 0x48a10+0x2cfc, code[:])
	if sha256.Sum256(code[:]) != axisLineCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 24)
	if !ok {
		return nil
	}
	slot := v.Read16(stack + 12)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	s := descriptor(slot * 2)
	root := s
	links := 0
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return nil
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	x0, y0, x1, y1, color := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18), v.Read16(stack+20), v.Read16(stack+22)

	g, ok := clippedGeometry(x0, y0, x1, y1, s)
	if !ok || steep != g.steep || decrease != g.decrease {
		return nil
	}
	x0, x1, y0, y1 = g.x0, g.x1, g.y0, g.y1
	dx, dy, major, minor := g.dx, g.dy, g.major, g.minor
	if v.CX() > major || v.AX() != color || v.DX() != major {
		return nil
	}
	done := int32(major) - int32(v.CX()) - g.endExcess
	if done < g.first || done > int32(major) {
		return nil
	}
	moved := (uint32(major/2) + uint32(done)*uint32(minor)) / uint32(major)
	err := uint16((uint32(major/2) + uint32(done)*uint32(minor)) % uint32(major))
	if v.BX() != err {
		return nil
	}
	currentX, currentY := g.position(done)
	stride := root.right - root.left + 1
	at := uint32(uint16((currentY-int32(root.top))*int32(stride) + currentX - int32(root.left)))
	if v.DI() != uint16(at) {
		return nil
	}
	wantSI, wantBP := uint16(int32(s.bottom)-currentY), dy
	if steep {
		wantSI, wantBP = dx, y1-s.bottom
	}
	if v.SI() != wantSI || v.BP() != wantBP {
		return nil
	}
	selectedBottom := s.bottom
	if !steep {
		_, firstY := g.position(g.first)
		selectedBottom = uint16(firstY)
	}
	for _, w := range [][2]uint16{{0xebd, s.seg}, {0xebf, s.extra}, {0xec1, s.left}, {0xec3, s.top}, {0xec5, s.right}, {0xec7, selectedBottom}, {0xec9, s.flags}, {0xecb, s.right - s.left + 1}, {0xecd, s.bottom - s.top + 1}, {0xecf, root.seg}, {0xed1, root.extra}, {0xed3, root.left}, {0xed5, root.top}, {0xed7, root.right}, {0xed9, root.bottom}, {0xedb, root.flags}, {0xedd, stride}, {0xedf, root.bottom - root.top + 1}, {0x1129, color}, {0x112b, x0}, {0x112d, x1}, {0x112f, y0}, {0x1131, y1}, {0x1137, major / 2}, {0x1149, x0 - x1}, {0x114d, dx}, {0x1151, dy}} {
		if word(w[0]) != w[1] {
			return nil
		}
	}
	remainingMinor := uint16(g.horizontal - int32(moved) + g.firstMoved)
	if steep && (word(0x1139) != uint16(g.topRemaining) || word(0x1133) != remainingMinor) {
		return nil
	}
	p := imageBasePlan(v)
	remaining := v.CX()
	charge := func(n, c, iterations uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + iterations
		} else {
			p.Cycles += c
		}
	}
	pixel := func() {
		if steep {
			p.Writes = append(p.Writes, machine.HostWrite{Address: 0xa0000 + at, Value: byte(color), Step: p.Instructions + 1, CS: 0x48a1, IP: v.IP()})
		} else {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color & 255, Count: 1, Width: 1, Step: p.Instructions + 1, CS: 0x48a1, IP: v.IP(), Backward: decrease})
		}
	}
	for {
		if remaining == 0 {
			n, c, extra := uint64(9), uint64(58), uint64(1)
			if steep {
				n, c, extra = 10, 65, 0
			}
			if n > budget-p.Instructions {
				break
			}
			pixel()
			carry := p.After.Flags & cpu.CF
			charge(n, c, extra)
			p.After.Flags = nativeCompareFlags(p.After.Flags, 0, 1)&^cpu.CF | carry
			p.After.R[cpu.CX] = 0xffff
			p.After.R[cpu.DI] = v.Read16(stack)
			p.After.R[cpu.SI] = v.Read16(stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(stack + 4)
			p.After.R[cpu.BP] = v.Read16(stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(stack + 10)
			p.LastIP = 0x3235
			return p
		}
		nextErr := err + minor
		shift := nextErr >= major
		n, c, extra := uint64(6), uint64(22), uint64(1)
		if shift {
			n, c = 10, 35
		}
		if steep {
			n, c, extra = 7, 29, 0
			if shift {
				n, c = 11, 41
			}
		}

		stop := shift && ((steep && remainingMinor == 0) || (!steep && currentY == int32(s.bottom)))
		if stop {
			if steep {
				n, c = 18, 85
			} else if decrease {
				n, c = 17, 79
			} else {
				n, c = 16, 72
			}
		}
		if n > budget-p.Instructions {
			break
		}
		pixel()
		oldN := p.Instructions
		if shift {
			nextErr -= major
		}
		oldAvailable := uint16(int32(s.bottom) - currentY)
		if steep {
			currentY++
			if shift {
				if decrease {
					currentX--
				} else {
					currentX++
				}
			}
		} else {
			if decrease {
				currentX--
			} else {
				currentX++
			}
			if shift {
				currentY++
			}
		}
		at = uint32(uint16((currentY-int32(root.top))*int32(stride) + currentX - int32(root.left)))
		remaining--
		err = nextErr
		if !shift {
			p.After.Flags = nativeCompareFlags(p.After.Flags, err, major) | expectedDF
			if steep {
				p.LastIP = 0x30ab
				if decrease {
					p.LastIP = 0x30c8
				}
			} else {
				p.LastIP = 0x3224
				if decrease {
					p.LastIP = 0x320e
				}
			}
		} else if steep {
			oldMinor := remainingMinor
			remainingMinor--
			p.After.Flags = nativeCompareFlags(p.After.Flags, oldMinor, 1) &^ cpu.CF
			ip := uint16(0x30b0)
			p.LastIP = 0x30b4
			if decrease {
				ip, p.LastIP = 0x30cd, 0x30d1
			}
			imageWord(p, 0x1133, remainingMinor, oldN+10, ip)
		} else {
			p.After.Flags = nativeCompareFlags(p.After.Flags, oldAvailable, 1)&^cpu.CF | expectedDF
			p.LastIP = 0x322e
			if decrease {
				p.LastIP = 0x3218
			}
		}
		charge(n, c, extra)
		p.After.R[cpu.CX] = remaining
		p.After.R[cpu.BX] = err
		p.After.R[cpu.DI] = uint16(at)
		if !steep {
			p.After.R[cpu.SI] = uint16(int32(s.bottom) - currentY)
		}

		p.After.IP = v.IP()
		if stop {
			p.After.Flags &^= cpu.DF
			p.After.R[cpu.DI] = v.Read16(stack)
			p.After.R[cpu.SI] = v.Read16(stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(stack + 4)
			p.After.R[cpu.BP] = v.Read16(stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(stack + 10)
			p.LastIP = 0x3235
			return p
		}
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}

// Geometry is independent of the ABI and original instruction stream.
type clippedLineGeometry struct {
	x0, y0, x1, y1, dx, dy, major, minor                   uint16
	steep, decrease                                        bool
	first, firstMoved, endExcess, topRemaining, horizontal int32
}

func (g clippedLineGeometry) position(k int32) (int32, int32) {
	moved := (int32(g.major/2) + k*int32(g.minor)) / int32(g.major)
	dir := int32(1)
	if g.decrease {
		dir = -1
	}
	if g.steep {
		return int32(int16(g.x0)) + dir*moved, int32(int16(g.y0)) + k
	}
	return int32(int16(g.x0)) + dir*k, int32(int16(g.y0)) + moved
}
func clippedGeometry(x0, y0, x1, y1 uint16, s rectangleSurface) (clippedLineGeometry, bool) {
	g := clippedLineGeometry{}
	if x0 == x1 || y0 == y1 {
		return g, false
	}
	if int16(y0) > int16(y1) {
		x0, x1, y0, y1 = x1, x0, y1, y0
	}
	dx, dy := int32(int16(x1))-int32(int16(x0)), int32(int16(y1))-int32(int16(y0))
	if dx < 0 {
		dx = -dx
		g.decrease = true
	}
	if dx > 32767 || dy > 32767 {
		return g, false
	}
	g.x0, g.x1, g.y0, g.y1 = x0, x1, y0, y1
	g.dx, g.dy = uint16(dx), uint16(dy)
	g.major, g.minor = uint16(dx), uint16(dy)
	g.steep = dy > dx
	if g.steep {
		g.major, g.minor = uint16(dy), uint16(dx)
	}
	g.first = -1
	for k := int32(0); k <= int32(g.major); k++ {
		x, y := g.position(k)
		if x >= int32(s.left) && x <= int32(s.right) && y >= int32(s.top) && y <= int32(s.bottom) {
			g.first = k
			break
		}
	}
	if g.first < 0 {
		return g, false
	}
	g.firstMoved = (int32(g.major/2) + g.first*int32(g.minor)) / int32(g.major)
	if g.steep {
		g.endExcess = int32(int16(y1)) - int32(s.bottom)
		if g.endExcess < 0 {
			g.endExcess = 0
		}
		skipTop := int32(s.top) - int32(int16(y0))
		if skipTop < 0 {
			skipTop = 0
		}
		g.topRemaining = dx - (dy/2+skipTop*dx)/dy
		margin := int32(0)
		if int32(int16(x1)) < int32(s.left) {
			margin = int32(int16(x1)) - int32(s.left)
		}
		if int32(int16(x1)) > int32(s.right) {
			margin = int32(s.right) - int32(int16(x1))
		}
		g.horizontal = g.topRemaining + margin
	} else {
		if int32(int16(x1)) < int32(s.left) {
			g.endExcess = int32(s.left) - int32(int16(x1))
		}
		if int32(int16(x1)) > int32(s.right) {
			g.endExcess = int32(int16(x1)) - int32(s.right)
		}
	}
	return g, true
}
