// Horizontal spans and vertical line cursors computed in Go, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var axisLineCode = smallFingerprint(1338, "06d69c6ecaffd16f3d050756a9485ea703e3bb67821f3f9c6f62f45772b46fba")

func planAxisLine(v CPUView, budget uint64) *machine.HostWorkPlan {
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
	if x0 != x1 && y0 != y1 {
		return nil
	}
	horizontal := y0 == y1
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
	loX, hiX := ax, bx
	if int16(loX) > int16(hiX) {
		loX, hiX = hiX, loX
	}
	if horizontal {
		if int16(cy) < int16(s.top) || int16(cy) > int16(s.bottom) || int16(hiX) < int16(s.left) || int16(loX) > int16(s.right) {
			return nil
		}
	} else {
		if int16(x0) < int16(s.left) || int16(x0) > int16(s.right) || int16(dy) < int16(s.top) || int16(cy) > int16(s.bottom) {
			return nil
		}
	}
	n, cost, iterations := uint64(50+4*links), uint64(189), uint64(0)
	if links > 0 {
		cost += 21*links - 10
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	finish := func() *machine.HostWorkPlan {
		phase(6, 42)
		p.After.R[cpu.SP] += 4
		p.After.IP = v.Read16(stack)
		p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		p.Instructions, p.Cycles, p.LastIP = n, cost, 0x3235
		if v.DOSBoxCost() {
			p.Cycles = n + iterations
		}
		if n > budget {
			return nil
		}
		return p
	}
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
	phase(2, 8)
	if !horizontal {
		cost += 5
	}
	p.After.Seg[cpu.ES] = 0xa000
	p.After.R[cpu.CX] = 0
	address := func(x, y uint16) uint32 {
		at := uint32(y-root.top)*uint32(root.right-root.left+1) + uint32(x-root.left)
		phase(10, 58)
		p.After.R[cpu.DX] = 0
		if root.right-root.left+1 == 320 {
			phase(2, 0)
			cost -= 10
			p.After.R[cpu.DX] = (y - root.top) * 64
		}
		return at
	}
	if horizontal {
		left, right := loX, hiX
		phase(2, 8)
		if int16(left) > int16(s.left) {
			cost += 5
		} else {
			left = s.left
			phase(1, 2)
		}
		phase(2, 8)
		if int16(right) < int16(s.right) {
			cost += 5
		} else {
			right = s.right
			phase(1, 6)
		}
		phase(1, 2)
		emit(0x113b, left, 0x2e2b)
		phase(1, 6)
		emit(0x113d, right, 0x2e2e)
		at := address(left, cy)
		width := right - left + 1
		phase(6, 20)
		phase(1, 3+5*uint64(width/2))
		if width/2 > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color&255 | color<<8, Count: width / 2, Width: 2, Step: n, CS: 0x48a1, IP: 0x2e6d})
		}
		phase(1, 2)
		phase(1, 3+5*uint64(width%2))
		if width%2 != 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(width/2)*2, Value: color & 255, Count: 1, Width: 1, Step: n, CS: 0x48a1, IP: 0x2e71})
		}
		iterations = uint64(width/2 + width%2)
		p.After.R[cpu.AX] = color&255 | color<<8
		p.After.R[cpu.BX] = right
		p.After.Flags = nativeCompareFlags(v.Flags(), width/2, 0)
		phase(1, 7)
		return finish()
	}
	// Vertical lines use one Go column cursor; source positions label the
	// original eight stores solely for ordered write notices and timing.
	phase(2, 13) // test of nonzero horizontal distance and its taken zero branch
	top, bottom := cy, dy
	phase(2, 8)
	if int16(top) > int16(s.top) {
		cost += 5
	} else {
		top = s.top
		phase(1, 6)
	}
	phase(2, 8)
	if int16(bottom) < int16(s.bottom) {
		cost += 5
	} else {
		bottom = s.bottom
		phase(1, 6)
	}
	phase(1, 6)
	emit(0x112f, top, 0x2e94)
	phase(1, 6)
	emit(0x1131, bottom, 0x2e98)
	at := address(x0, top)
	phase(5, 14)
	height := bottom - top + 1
	for i := uint16(0); i < height; i++ {
		slot := i % 8
		phase(1, 7)
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0xa0000 + at + uint32(i)*uint32(root.right-root.left+1), Value: byte(color), Step: n, CS: 0x48a1, IP: 0x2ed4 + 11*slot})
		if slot < 7 {
			phase(2, 4)
			if i+1 == height {
				cost += 5
			} else {
				phase(1, 7)
			}
		} else {
			phase(2, 9)
		}
	}
	p.After.R[cpu.AX] = color
	p.After.R[cpu.BX] = x0
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	if height%8 == 0 {
		last := uint16(at) + uint16(uint32(height-1)*uint32(root.right-root.left+1))
		p.After.Flags = imageRowAddFlags(v.Flags()&^cpu.DF, last, root.right-root.left+1)
	}
	phase(1, 7)
	return finish()
}
