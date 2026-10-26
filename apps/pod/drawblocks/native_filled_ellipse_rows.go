// Typed Go filled-ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func filledEllipseRows(v CPUView, p *machine.HostWorkPlan, x, y int32, s rectangleSurface, delta uint16) {
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	cx, cy := int32(int16(word(0x112d))), int32(int16(word(0x1131)))
	charge := func(n, c, iterations uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + iterations
		} else {
			p.Cycles += c
		}
	}
	right, left := cx+x, cx-x
	if right < int32(s.left) {
		charge(5, 23, 0)
		return
	}
	charge(4, 21, 0)
	if right < int32(s.right) {
		charge(2, 13, 0)
	} else {
		right = int32(s.right)
		charge(3, 10, 0)
	}
	charge(1, 2, 0)
	imageWord(p, 0x113d, uint16(right), p.Instructions, 0x3b6a+delta)
	if left > int32(s.right) {
		charge(5, 23, 0)
		return
	}
	charge(4, 21, 0)
	if left > int32(s.left) {
		charge(2, 13, 0)
	} else {
		left = int32(s.left)
		charge(3, 10, 0)
	}
	charge(1, 2, 0)
	imageWord(p, 0x113b, uint16(left), p.Instructions, 0x3b86+delta)
	if cy+y < int32(s.top) {
		charge(5, 23, 0)
		return
	}
	charge(4, 21, 0)
	span := func(row int32, wordIP, byteIP uint16) {
		oldN := p.Instructions
		width := uint16(right - left + 1)
		words, tail := width/2, width%2
		at := uint32(row*320 + left)
		color := word(0x1129)
		if words > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color, Count: words, Width: 2, Step: oldN + 19, CS: 0x48a1, IP: wordIP + delta})
		}
		if tail > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(words)*2, Value: color & 255, Count: 1, Width: 1, Step: oldN + 21, CS: 0x48a1, IP: byteIP + delta})
		}
		charge(21, 72+5*uint64(words+tail), uint64(words+tail))
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.DI] = uint16(at) + width
	}
	if cy+y > int32(s.bottom) {
		charge(2, 13, 0)
	} else {
		charge(2, 8, 0)
		span(cy+y, 0x3bd8, 0x3bdc)
	}
	if cy-y < int32(s.top) {
		charge(4, 21, 0)
		return
	}
	charge(4, 16, 0)
	if cy-y > int32(s.bottom) {
		charge(2, 13, 0)
		return
	}
	charge(2, 8, 0)
	span(cy-y, 0x3c2a, 0x3c2e)
}
