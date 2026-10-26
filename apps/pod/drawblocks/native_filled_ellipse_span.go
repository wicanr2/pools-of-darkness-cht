// Typed Go filled-ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type filledSpanView struct {
	CPUView
	entry uint16
}

func (v filledSpanView) IP() uint16 { return v.entry }
func filledEllipseSpanStage(ip uint16) (entry uint16, negative, tail, ok bool) {
	switch ip {
	case 0x3bd8:
		return 0x3b3d, false, false, true
	case 0x3bdc:
		return 0x3b3d, false, true, true
	case 0x3c2a:
		return 0x3b3d, true, false, true
	case 0x3c2e:
		return 0x3b3d, true, true, true
	case 0x3d47:
		return 0x3cc0, false, false, true
	case 0x3d4b:
		return 0x3cc0, false, true, true
	case 0x3d99:
		return 0x3cc0, true, false, true
	case 0x3d9d:
		return 0x3cc0, true, true, true
	}
	return 0, false, false, false
}
func planFilledEllipseSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, negative, tail, ok := filledEllipseSpanStage(v.IP())
	if !ok || budget == 0 || !validEllipseFrame(v) || v.Flags()&cpu.DF != 0 {
		return nil
	}
	projected := filledSpanView{v, entry}
	var full *machine.HostWorkPlan
	if entry == 0x3b3d {
		full = planFilledEllipseFirst(projected, 1000)
	} else {
		full = planFilledEllipseSecond(projected, 1000)
	}
	if full == nil {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	cx, cy, x, y := int32(int16(word(0x112d))), int32(int16(word(0x1131))), int32(word(0x112b)), int32(word(0x112f))
	left, right := cx-x, cx+x
	if left < int32(word(0xec1)) {
		left = int32(word(0xec1))
	}
	if right > int32(word(0xec5)) {
		right = int32(word(0xec5))
	}
	row := cy + y
	if negative {
		row = cy - y
	}
	if left > right || row < int32(word(0xec3)) || row > int32(word(0xec7)) || word(0x113b) != uint16(left) || word(0x113d) != uint16(right) || v.AX() != word(0x1129) {
		return nil
	}
	width := uint16(right - left + 1)
	count, size := width/2, uint8(2)
	at := uint32(row*320 + left)
	if tail {
		at += uint32(count) * 2
		count, size = width%2, 1
	}
	if uint32(v.DI()) != at || v.CX() != count {
		return nil
	}
	p := imageBasePlan(v)
	p.Instructions, p.Cycles, p.LastIP = 1, 3+uint64(count)*5, v.IP()
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	if count > 0 {
		value := word(0x1129)
		if tail {
			value &= 255
		}
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + at, Value: value, Count: count, Width: size, Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	p.After.R[cpu.DI] += count * uint16(size)
	p.After.R[cpu.CX] = 0
	p.After.IP += 2
	return p
}
func planFilledEllipseBounded(v CPUView, budget uint64) *machine.HostWorkPlan {
	var p *machine.HostWorkPlan
	switch v.IP() {
	case 0x39b2:
		p = planFilledEllipsePrefix(v)
	case 0x3b3d:
		p = planFilledEllipseFirst(v, budget)
		if p == nil {
			p = planFilledEllipseTransition(v, budget)
		}
	case 0x3cc0:
		p = planFilledEllipseSecond(v, budget)
	}
	if p != nil && p.Instructions <= budget {
		return p
	}
	return planFilledEllipseSpan(v, budget)
}

func filledEllipseFixedTransfers(transfers []machine.HostTransfer) bool {
	for _, w := range transfers {
		if _, _, _, ok := filledEllipseSpanStage(w.IP); ok && w.CS == 0x48a1 && w.Count > 0 {
			return true
		}
	}
	return false
}
