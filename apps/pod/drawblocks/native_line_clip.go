// Interior diagonal rasterization uses Go integer error accumulators.
// Original positions retain only ABI, scratch and model-time notices.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planClippedLinePrefix(v CPUView) *machine.HostWorkPlan {
	var prefix *machine.HostWorkPlan
	if v.CS() != 0x48a1 || v.IP() != 0x2cfc || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := axisLineCode
	var code [1338]byte
	copyView(v, 0x48a10+0x2cfc, code[:])
	if sha256.Sum256(code[:]) != fp.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 16)
	if !ok {
		return nil
	}
	slot := v.Read16(stack + 4)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	s := descriptor(slot * 2)
	root := s
	links := uint64(0)
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
	p := imageBasePlan(v)
	entry := v.IP()
	delta := entry - 0x2b10
	for _, w := range []struct {
		off   uint32
		value uint16
		step  uint64
		ip    uint16
	}{{2, v.BP(), 1, 0x2b10}, {4, v.DS(), 3, 0x2b13}, {6, v.SI(), 4, 0x2b14}, {8, v.DI(), 5, 0x2b15}, {10, v.CS(), 6, 0x2b16}} {
		p.Writes = append(p.Writes, machine.HostWrite{Address: stack - w.off, Value: byte(w.value), Step: w.step, CS: 0x48a1, IP: w.ip + delta}, machine.HostWrite{Address: stack - w.off + 1, Value: byte(w.value >> 8), Step: w.step, CS: 0x48a1, IP: w.ip + delta})
	}
	// Common typed surface fields; source positions only retain ordered notices.
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xebd, s.seg, 11, 0x2b21}, {0xebf, s.extra, 13, 0x2b28}, {0xec1, s.left, 15, 0x2b2f}, {0xec3, s.top, 17, 0x2b36}, {0xec5, s.right, 19, 0x2b3d}, {0xecb, s.right - s.left + 1, 22, 0x2b45}, {0xec7, s.bottom, 24, 0x2b4c}, {0xecd, s.bottom - s.top + 1, 27, 0x2b54}, {0xec9, s.flags, 29, 0x2b5b}} {
		imageWord(p, w.off, w.value, w.step, w.ip+delta)
	}
	// The read variant shares the remaining shape at a different code offset.
	tailDelta := delta
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xedb, root.flags, 32, 0x2b70}, {0xecf, root.seg, 34, 0x2b77}, {0xed1, root.extra, 36, 0x2b7e}, {0xed5, root.top, 38, 0x2b85}, {0xed9, root.bottom, 40, 0x2b8c}, {0xedf, root.bottom - root.top + 1, 43, 0x2b94}, {0xed3, root.left, 45, 0x2b9b}, {0xed7, root.right, 47, 0x2ba2}, {0xedd, root.right - root.left + 1, 50, 0x2baa}} {
		imageWord(p, w.off, w.value, w.step+4*links, w.ip+tailDelta)
	}
	x0, y0, x1, y1, color := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12), v.Read16(stack+14)
	if x0 == x1 || y0 == y1 {
		return nil
	}
	ax, bx, cy, dy := x0, x1, y0, y1
	abs := func(x int32) int32 {
		if x < 0 {
			return -x
		}
		return x
	}
	if abs(int32(int16(x0))-int32(int16(x1))) > 32767 || abs(int32(int16(y0))-int32(int16(y1))) > 32767 {
		return nil
	}
	if int16(cy) >= int16(dy) {
		ax, bx, cy, dy = bx, ax, dy, cy
	}
	fromX, toX := ax, bx
	g, ok := clippedGeometry(x0, y0, x1, y1, s)
	if !ok {
		return nil
	}
	n, cost := uint64(50+4*links), uint64(189)
	if links > 0 {
		cost += 21*links - 10
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	phase(6, 51)
	phase(2, 4)
	if int16(y0) < int16(y1) {
		cost += 5
	} else {
		phase(2, 4)
	}
	phase(1, 2)
	emit(0x112b, fromX, 0x2db2)
	phase(1, 6)
	emit(0x112d, toX, 0x2db5)
	phase(2, 13)
	phase(1, 6)
	emit(0x112f, cy, 0x2dc2)
	phase(2, 13)
	phase(1, 6)
	emit(0x1131, dy, 0x2dcf)
	phase(2, 4)
	if int16(fromX) < int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(4, 26)
	phase(2, 12)
	emit(0x1129, color, 0x2ded)
	difference := fromX - toX
	magnitude := uint16(abs(int32(int16(fromX)) - int32(int16(toX))))
	phase(3, 18)
	emit(0x1149, difference, 0x2df9)
	phase(1, 2)
	if int16(fromX) >= int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(1, 6)
	emit(0x114d, magnitude, 0x2e01)
	phase(3, 18)
	emit(0x1151, dy-cy, 0x2e0d)
	phase(2, 13)
	p.After.Seg[cpu.ES] = 0xa000
	if g.steep {

		decrease := g.decrease
		phase(6, 30)
		phase(2, 4)
		emit(0x1139, magnitude, 0x2f3d)
		phase(4, 16)
		emit(0x1137, (dy-cy)/2, 0x2f48)
		phase(2, 12) // initial SI,DI
		topSkip := int32(s.top) - int32(int16(cy))
		if topSkip < 0 {
			topSkip = 0
		}
		phase(2, 8) // compare top, conditional branch
		if topSkip == 0 {
			cost += 5
		} else {
			phase(4, 13)
			phase(2, 51)
			if decrease {
				phase(3, 15)
			} else {
				phase(5, 14)
			}
			q := uint16(topSkip * int32(magnitude) / int32(dy-cy))
			phase(1, 6)
			emit(0x1139, magnitude-q, 0x2f7b)
			phase(3, 10)
			remainder := uint16(topSkip * int32(magnitude) % int32(dy-cy))
			if (dy-cy)/2+remainder < dy-cy {
				cost += 5
			} else {
				if decrease {
					phase(5, 27)
				} else {
					phase(7, 26)
				}
				emit(0x1139, uint16(g.topRemaining), 0x2f95)
			}
			phase(1, 6)
			if decrease {
				phase(2, 13)
			} else {
				phase(2, 8)
			}
			phase(2, 13)
		}
		topX, _ := g.position(topSkip)
		sideClip := (!decrease && topX < int32(s.left)) || (decrease && topX > int32(s.right))
		if sideClip {
			if decrease {
				phase(8, 33)
			} else {
				phase(7, 31)
			}
			phase(3, 18)
			phase(2, 51)
			phase(4, 12)
			phase(5, 57)
			phase(4, 37)
			phase(6, 26)
			if decrease {
				phase(4, 20)
			} else {
				phase(3, 19)
			}
		} else {
			if decrease {
				phase(7, 31)
			} else {
				phase(7, 36)
			}
		}
		// End-side counter deliberately retains the original side-start behavior.
		if int16(toX) < int16(s.left) {
			phase(5, 31)
		} else if int16(toX) > int16(s.right) {
			phase(8, 45)
		} else {
			phase(9, 42)
		}
		emit(0x1133, uint16(g.horizontal), 0x3053)
		phase(10, 50)
		if root.right-root.left+1 == 320 {
			phase(2, 0)
			cost -= 10
		}
		if int16(dy) > int16(s.bottom) {
			phase(7, 30)
		} else {
			phase(6, 33)
		}
		phase(2, 8)
		if decrease {
			cost += 5
		}
		prefix = imageBasePlan(v)
		prefix.Writes = append(prefix.Writes, p.Writes...)
		prefix.Instructions, prefix.Cycles = n, cost
		if v.DOSBoxCost() {
			prefix.Cycles = n
		}
		prefix.After.R[cpu.AX] = color
		prefix.After.R[cpu.BX] = uint16((int32(g.major/2) + g.first*int32(g.minor)) % int32(g.major))
		prefix.After.R[cpu.CX] = uint16(int32(g.major) - g.first - g.endExcess)
		prefix.After.R[cpu.DX] = dy - cy
		prefix.After.R[cpu.SI] = magnitude
		prefix.After.R[cpu.BP] = dy - s.bottom
		firstX, firstY := g.position(g.first)
		prefix.After.R[cpu.DI] = uint16((firstY-int32(root.top))*int32(root.right-root.left+1) + firstX - int32(root.left))
		prefix.After.R[cpu.SP] -= 8
		prefix.After.Seg[cpu.DS] = v.CS()
		prefix.After.Seg[cpu.ES] = 0xa000
		prefix.After.Flags = nativeCompareFlags(v.Flags(), difference, 0)
		prefix.After.IP = 0x309c
		if decrease {
			prefix.After.IP = 0x30b9
		}
		prefix.LastIP = 0x309a
		return prefix
	}

	decrease := g.decrease
	phase(7, 32)
	phase(4, 16)
	emit(0x1137, magnitude/2, 0x30de)
	phase(4, 20)
	sideSkip := int32(s.left) - int32(int16(fromX))
	if decrease {
		sideSkip = int32(int16(fromX)) - int32(s.right)
	}
	if sideSkip < 0 {
		sideSkip = 0
	}
	if !decrease {
		cost += 5
		if sideSkip > 0 {
			phase(3, 6)
		} else {
			phase(3, 11)
		}
	} else {
		if sideSkip > 0 {
			phase(3, 15)
		} else {
			phase(4, 17)
		}
	}
	if sideSkip > 0 {
		phase(2, 9)
		phase(2, 51)
		phase(4, 12)
		remainder := uint16(sideSkip * int32(dy-cy) % int32(magnitude))
		if magnitude/2+remainder < magnitude {
			cost += 5
		} else {
			phase(2, 8)
		}
		phase(2, 13)
		if decrease {
			phase(3, 19)
		} else {
			phase(4, 20)
		}
	}
	_, sideY := g.position(sideSkip)
	phase(2, 8)
	if sideY >= int32(s.top) {
		cost += 5
	} else {
		phase(4, 24)
		phase(2, 4)
		phase(2, 51)
		phase(4, 12)
		phase(5, 57)
		phase(4, 37)
		phase(3, 11)
		if decrease {
			phase(5, 28)
		} else {
			phase(6, 30)
		}
		phase(1, 6)
	}
	if int16(toX) < int16(s.left) {
		phase(4, 16)
	} else {
		phase(3, 19)
	}
	if int16(toX) > int16(s.right) {
		phase(4, 16)
	} else {
		phase(3, 19)
	}
	phase(10, 50)
	if root.right-root.left+1 == 320 {
		phase(2, 0)
		cost -= 10
	}
	firstX, firstY := g.position(g.first)
	phase(1, 6)
	emit(0xec7, uint16(firstY), 0x31eb)
	phase(6, 28)
	if !decrease {
		cost += 5
	} else {
		phase(1, 2)
	}
	prefix = imageBasePlan(v)
	prefix.Writes = append(prefix.Writes, p.Writes...)
	prefix.Instructions, prefix.Cycles = n, cost
	if v.DOSBoxCost() {
		prefix.Cycles = n
	}
	prefix.After.R[cpu.AX] = color
	prefix.After.R[cpu.BX] = uint16((int32(g.major/2) + g.first*int32(g.minor)) % int32(g.major))
	prefix.After.R[cpu.CX] = uint16(int32(g.major) - g.first - g.endExcess)
	prefix.After.R[cpu.DX] = magnitude
	prefix.After.R[cpu.SI] = uint16(int32(s.bottom) - firstY)
	prefix.After.R[cpu.BP] = dy - cy
	prefix.After.R[cpu.DI] = uint16((firstY-int32(root.top))*int32(root.right-root.left+1) + firstX - int32(root.left))
	prefix.After.R[cpu.SP] -= 8
	prefix.After.Seg[cpu.DS] = v.CS()
	prefix.After.Seg[cpu.ES] = 0xa000
	prefix.After.Flags = nativeCompareFlags(v.Flags(), difference, 0)
	prefix.After.IP = 0x321c
	prefix.LastIP = 0x3203
	if decrease {
		prefix.After.IP = 0x3206
		prefix.After.Flags |= cpu.DF
		prefix.LastIP = 0x3205
	}
	return prefix
}
