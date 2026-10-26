// Typed point geometry and pixel read/write, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var pointWriteCode = smallFingerprint(244, "5ac8a4f4342f02d70c6cf19c77b91d50002271c85a437fc5830e351be6634705")
var pointReadCode = smallFingerprint(248, "740e7fe88b2ec3014aa2d1f05d82b0c8ece8bc5b593d673ed2fe25390566b500")

func planPoint(v CPUView, budget uint64) *machine.HostWorkPlan {
	read := v.IP() == 0x2c04
	if v.CS() != 0x48a1 || (!read && v.IP() != 0x2b10) || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := pointWriteCode
	if read {
		fp = pointReadCode
	}
	var code [248]byte
	copyView(v, 0x48a10+uint32(v.IP()), code[:fp.size])
	if sha256.Sum256(code[:fp.size]) != fp.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 12)
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
	delta := uint16(0)
	if read {
		delta = 0xf4
	}
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
	tailDelta := uint16(0)
	if read {
		tailDelta = 0xf4
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xedb, root.flags, 32, 0x2b70}, {0xecf, root.seg, 34, 0x2b77}, {0xed1, root.extra, 36, 0x2b7e}, {0xed5, root.top, 38, 0x2b85}, {0xed9, root.bottom, 40, 0x2b8c}, {0xedf, root.bottom - root.top + 1, 43, 0x2b94}, {0xed3, root.left, 45, 0x2b9b}, {0xed7, root.right, 47, 0x2ba2}, {0xedd, root.right - root.left + 1, 50, 0x2baa}} {
		imageWord(p, w.off, w.value, w.step+4*links, w.ip+tailDelta)
	}
	x, y := v.Read16(stack+6), v.Read16(stack+8)
	sx, sy := int32(int16(x)), int32(int16(y))
	width := root.right - root.left + 1
	n, cost := uint64(58), uint64(248)
	p.After.R[cpu.AX] = width
	if read {
		p.After.R[cpu.AX] = 0xffff
		n++
		cost += 2
	}
	p.After.R[cpu.BX] = x
	p.After.Flags = nativeCompareFlags(v.Flags(), x, s.left) | v.Flags()&cpu.DF
	valid := false
	if sx >= int32(s.left) {
		n += 2
		cost += 8
		p.After.Flags = nativeCompareFlags(v.Flags(), x, s.right) | v.Flags()&cpu.DF
		if sx <= int32(s.right) {
			n += 3
			cost += 14
			p.After.R[cpu.BX] = y
			p.After.Flags = nativeCompareFlags(v.Flags(), y, s.top) | v.Flags()&cpu.DF
			if sy >= int32(s.top) {
				n += 2
				cost += 8
				p.After.Flags = nativeCompareFlags(v.Flags(), y, s.bottom) | v.Flags()&cpu.DF
				valid = sy <= int32(s.bottom)
			}
		}
	}
	if valid {
		dx, dy := x-root.left, y-root.top
		at := uint32(dy)*uint32(width) + uint32(dx)
		n, cost = 78, 369
		p.After.Seg[cpu.ES] = 0xa000
		p.After.R[cpu.DX] = 0
		p.After.Flags = imageRowAddFlags(v.Flags(), dx, uint16(uint32(dy)*uint32(width))) | v.Flags()&cpu.DF
		if width == 320 {
			n, cost = 80, 359
			p.After.R[cpu.DX] = dy * 64
			p.After.Flags = imageRowAddFlags(v.Flags(), dx+dy*256, dy*64) | v.Flags()&cpu.DF
		}
		p.After.R[cpu.AX] = v.Read16(stack + 10)
		if read {
			n += 2
			p.After.R[cpu.BX] = uint16(at)
			p.After.R[cpu.AX] = uint16(v.Read8(0xa0000 + at))
			p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0) | v.Flags()&cpu.DF
		} else {
			p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + at, Value: p.After.R[cpu.AX] & 255, Count: 1, Width: 1, Step: n - 5 + 4*links, CS: 0x48a1, IP: 0x2bfc}}
		}
	}
	p.After.R[cpu.SP] += 4
	p.After.IP = v.Read16(stack)
	p.After.Seg[cpu.CS] = v.Read16(stack + 2)
	if links > 0 {
		cost += 21*links - 10
	}
	last := uint16(0x2c03)
	if read {
		last = 0x2cfb
	}
	return imageTimingReturn(v, p, budget, n+4*links, cost, last)
}
