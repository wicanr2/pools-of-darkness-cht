// Typed Go ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
)

type ellipsePixelView struct {
	CPUView
	sp, entry, y uint16
}

func (v ellipsePixelView) SP() uint16 { return v.sp }
func (v ellipsePixelView) IP() uint16 { return v.entry }
func (v ellipsePixelView) BX() uint16 { return v.y }
func ellipsePixelStage(ip uint16) (uint16, bool) {
	switch ip {
	case 0x3660, 0x36b6, 0x370c, 0x3762:
		return 0x35f4, true
	case 0x384e, 0x38a4, 0x38fa, 0x3950:
		return 0x37f6, true
	}
	return 0, false
}
func planEllipsePixel(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, ok := ellipsePixelStage(v.IP())
	if !ok || budget < 1 || v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.ES() != 0xa000 {
		return nil
	}
	stack, ok := paletteSourceStack(v, 2, 26)
	if !ok || v.BP() != v.SP()+8 || v.CX() != uint16(byte(v.Read16(stack+24)))*257 {
		return nil
	}
	if v.Read16(0x48a10+0x112d) != v.Read16(stack+16) || v.Read16(0x48a10+0x1131) != v.Read16(stack+18) {
		return nil
	}
	projected := ellipsePixelView{v, v.SP() + 2, entry, v.Read16(stack)}
	var full *machine.HostWorkPlan
	if entry == 0x35f4 {
		full = planEllipseFirst(projected, 1000)
	} else {
		full = planEllipseSecond(projected, 1000)
	}
	if full == nil {
		return nil
	}
	for _, w := range full.Writes {
		if w.CS == 0x48a1 && w.IP == v.IP() && w.Address >= 0xa0000 && w.Address < 0xafa00 {
			if uint32(v.DI()) != w.Address-0xa0000 || byte(v.CX()) != w.Value {
				return nil
			}
			p := imageBasePlan(v)
			w.Step = 1
			p.Writes = []machine.HostWrite{w}
			p.Instructions, p.Cycles, p.LastIP = 1, 7, v.IP()
			if v.DOSBoxCost() {
				p.Cycles = 1
			}
			p.After.IP += 3
			return p
		}
	}
	return nil
}
func planEllipseBounded(v CPUView, budget uint64) *machine.HostWorkPlan {
	var p *machine.HostWorkPlan
	switch v.IP() {
	case 0x346a:
		p = planEllipsePrefix(v)
	case 0x35f4:
		p = planEllipseFirst(v, budget)
		if p == nil {
			p = planEllipseTransition(v, budget)
		}
	case 0x37f6:
		p = planEllipseSecond(v, budget)
	}
	if p != nil && p.Instructions <= budget {
		return p
	}
	return planEllipsePixel(v, budget)
}

func ellipseFixedPixelWrites(writes []machine.HostWrite) bool {
	for _, w := range writes {
		if _, ok := ellipsePixelStage(w.IP); ok && w.CS == 0x48a1 {
			return true
		}
	}
	return false
}
