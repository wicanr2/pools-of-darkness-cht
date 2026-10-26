// Typed bounded rectangle row cursor and ABI prefix, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planRegionRows(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x3418 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [306]byte
	copyView(v, 0x48a10+0x3338, code[:])
	if sha256.Sum256(code[:]) != regionCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 24)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	s := rectangleSurface{word(0xebd), word(0xebf), word(0xec1), word(0xec3), word(0xec5), word(0xec7), word(0xec9)}
	root := rectangleSurface{word(0xecf), word(0xed1), word(0xed3), word(0xed5), word(0xed7), word(0xed9), word(0xedb)}
	if root.seg != 0xa000 || root.flags&64 != 0 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	x0, y0, x1, y1 := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18), v.Read16(stack+20)
	if int16(x0) > int16(x1) || int16(y0) > int16(y1) {
		return nil
	}
	left, right, top, end := int32(int16(x0)), int32(int16(x1)), int32(int16(y0)), int32(int16(y1))
	if left < int32(s.left) {
		left = int32(s.left)
	}
	if right > int32(s.right) {
		right = int32(s.right)
	}
	if top < int32(s.top) {
		top = int32(s.top)
	}
	if end > int32(s.bottom) {
		end = int32(s.bottom)
	}
	if left > right || top > end || word(0x113b) != uint16(left) || word(0x113d) != uint16(right) || word(0x1129) != v.Read16(stack+22) || int32(v.SI()) < top || int32(v.SI()) > end+1 {
		return nil
	}
	if word(0xedd) != root.right-root.left+1 || word(0xedf) != root.bottom-root.top+1 {
		return nil
	}
	width, stride, color := uint16(right-left+1), root.right-root.left+1, word(0x1129)
	p := imageBasePlan(v)
	charge := func(n, c, iterations uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + iterations
		} else {
			p.Cycles += c
		}
	}
	for y := v.SI(); ; y++ {
		terminal := y > s.bottom || int16(y) > int16(y1)
		if terminal {
			n, c, bound := uint64(7), uint64(53), s.bottom
			if y <= s.bottom {
				n, c, bound = 9, 61, y1
			}
			if n > budget-p.Instructions {
				break
			}
			charge(n, c, 0)
			p.After.Flags = nativeCompareFlags(v.Flags(), y, bound)
			p.After.R[cpu.DI] = v.Read16(stack)
			p.After.R[cpu.SI] = v.Read16(stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(stack + 4)
			p.After.R[cpu.BP] = v.Read16(stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(stack + 10)
			p.LastIP = 0x3469
			return p
		}
		n, c, wordStep, byteStep := uint64(25), uint64(107), uint64(21), uint64(23)
		if stride == 320 {
			n, c, wordStep, byteStep = 27, 97, 23, 25
		}
		if n > budget-p.Instructions {
			break
		}
		at := uint32(y-root.top)*uint32(stride) + uint32(left-int32(root.left))
		iterations := uint64(width/2 + width%2)
		if width/2 > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color&255 | color<<8, Count: width / 2, Width: 2, Step: p.Instructions + wordStep, CS: 0x48a1, IP: 0x345c})
		}
		if width%2 != 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(width/2)*2, Value: color & 255, Count: 1, Width: 1, Step: p.Instructions + byteStep, CS: 0x48a1, IP: 0x3460})
		}
		charge(n, c+5*iterations, iterations)
		p.After.R[cpu.AX] = color&255 | color<<8
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.DX] = 0
		if stride == 320 {
			p.After.R[cpu.DX] = (y - root.top) * 64
		}
		p.After.R[cpu.DI] = uint16(at) + width
		p.After.R[cpu.SI] = y + 1
		p.After.Flags = fadeAdvanceFlags(v.Flags(), y, false)
		p.After.IP = 0x3418
		p.LastIP = 0x3463
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}

func planRegionBounded(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() == 0x3418 {
		return planRegionRows(v, budget)
	}
	full := planRegion(v, 10000)
	if full == nil {
		return nil
	}
	if full.Instructions <= budget {
		return full
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	slot := v.Read16(stack+4) * 2
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	left, right, top, bottom := word(0x4bd+slot), word(0x8bd+slot), word(0x6bd+slot), word(0xabd+slot)
	x0, y0, x1, y1 := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12)
	xlo, xhi, ylo, yhi := int32(int16(x0)), int32(int16(x1)), int32(int16(y0)), int32(int16(y1))
	if xlo < int32(left) {
		xlo = int32(left)
	}
	if xhi > int32(right) {
		xhi = int32(right)
	}
	if ylo < int32(top) {
		ylo = int32(top)
	}
	if yhi > int32(bottom) {
		yhi = int32(bottom)
	}
	if xlo > xhi || ylo > yhi {
		return nil
	}
	for word(0xcbd+slot)&64 != 0 {
		slot = word(0xbd + slot)
	}
	stride := word(0x8bd+slot) - word(0x4bd+slot) + 1
	rows := uint64(yhi - ylo + 1)
	width := uint64(xhi - xlo + 1)
	rowN, rowC := uint64(25), uint64(107)
	if stride == 320 {
		rowN, rowC = 27, 97
	}
	rowC += 5 * (width/2 + width%2)
	endN, endC := uint64(7), uint64(53)
	if yhi < int32(bottom) {
		endN, endC = 9, 61
	}
	n := full.Instructions - rows*rowN - endN
	if n > budget {
		return nil
	}
	c := n
	if !v.DOSBoxCost() {
		c = full.Cycles - rows*rowC - endC
	}
	p := imageBasePlan(v)
	p.Writes = append(p.Writes, full.Writes...)
	p.Instructions, p.Cycles = n, c
	p.After.R[cpu.AX] = uint16(xlo)
	p.After.R[cpu.BX] = slot
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.BP] = v.SP() - 2
	p.After.R[cpu.SI] = uint16(ylo)
	p.After.Seg[cpu.ES] = 0xa000
	p.After.Seg[cpu.DS] = 0x48a1
	p.After.Flags = nativeCompareFlags(v.Flags(), y0, top)
	p.After.IP = 0x3418
	p.LastIP = 0x3412
	if int16(y0) <= int16(top) {
		p.LastIP = 0x3414
	}
	return p
}
