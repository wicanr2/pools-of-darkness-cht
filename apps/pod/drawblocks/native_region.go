// Clipped rectangle spans computed in Go, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var regionCode = smallFingerprint(306, "48c54ccea05f6f319457cf4bcc454a759cb81650a14c89b26b722b3a3fce7a3e")

func planRegion(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x3338 || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := regionCode
	var code [306]byte
	copyView(v, 0x48a10+0x3338, code[:])
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
	x0, y0, x1, y1 := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12)
	if int16(x0) > int16(x1) || int16(y0) > int16(y1) {
		return nil
	}
	color := v.Read16(stack + 14)
	rootSlot := slot * 2
	for word(0xcbd+rootSlot)&64 != 0 {
		rootSlot = word(0xbd + rootSlot)
	}
	p.After.R[cpu.BX] = rootSlot
	p.After.Seg[cpu.ES] = 0xa000
	n, cost, iterations := uint64(54+4*links), uint64(224), uint64(0)
	if links > 0 {
		cost += 21*links - 10
	}
	imageWord(p, 0x1129, color, n, 0x33de)
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	flags := func(a, b uint16) { p.After.Flags = nativeCompareFlags(v.Flags(), a, b) }
	finish := func() *machine.HostWorkPlan {
		phase(5, 40)
		p.After.R[cpu.SP] += 4
		p.After.IP = v.Read16(stack)
		p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		p.Instructions, p.Cycles, p.LastIP = n, cost, 0x3469
		if v.DOSBoxCost() {
			p.Cycles = n + iterations
		}
		if n > budget {
			return nil
		}
		return p
	}
	p.After.R[cpu.AX] = x1
	flags(x1, s.left)
	phase(3, 14)
	if int16(x1) < int16(s.left) {
		cost += 5
		return finish()
	}
	phase(2, 8)
	right := x1
	if int16(x1) < int16(s.right) {
		cost += 5
	} else {
		right = s.right
		phase(1, 2)
	}
	phase(1, 2)
	imageWord(p, 0x113d, right, n, 0x33f3)
	p.After.R[cpu.AX] = x0
	flags(x0, s.right)
	phase(3, 14)
	if int16(x0) > int16(s.right) {
		cost += 5
		return finish()
	}
	phase(2, 8)
	left := x0
	if int16(x0) > int16(s.left) {
		cost += 5
	} else {
		left = s.left
		phase(1, 2)
	}
	p.After.R[cpu.AX] = left
	phase(1, 2)
	imageWord(p, 0x113b, left, n, 0x3408)
	phase(3, 14)
	top := y0
	if int16(y0) > int16(s.top) {
		cost += 5
	} else {
		top = s.top
		phase(1, 6)
	}
	width := right - left + 1
	stride := root.right - root.left + 1
	for y := top; ; y++ {
		phase(2, 8)
		flags(y, s.bottom)
		if y > s.bottom {
			cost += 5
			return finish()
		}
		phase(2, 8)
		flags(y, y1)
		if int16(y) > int16(y1) {
			cost += 5
			return finish()
		}
		rowSteps, rowCycles, wordStep, byteStep := uint64(21), uint64(91), uint64(17), uint64(19)
		if stride == 320 {
			rowSteps, rowCycles, wordStep, byteStep = 23, 81, 19, 21
		}
		at := uint32(y-root.top)*uint32(stride) + uint32(left-root.left)
		if width/2 > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color&255 | color<<8, Count: width / 2, Width: 2, Step: n + wordStep, CS: 0x48a1, IP: 0x345c})
		}
		if width%2 != 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(width/2)*2, Value: color & 255, Count: 1, Width: 1, Step: n + byteStep, CS: 0x48a1, IP: 0x3460})
		}
		count := uint64(width/2 + width%2)
		phase(rowSteps, rowCycles+5*count)
		iterations += count
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.AX] = color&255 | color<<8
		p.After.R[cpu.DX] = 0
		if stride == 320 {
			p.After.R[cpu.DX] = (y - root.top) * 64
		}
	}
}
