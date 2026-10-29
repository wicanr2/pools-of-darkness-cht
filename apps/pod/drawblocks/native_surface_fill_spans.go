package drawblocks

// Spec025/#49: typed surface geometry produces bounded Go fill spans.

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func surfaceFillPixelIP(ip uint16) bool { return ip == 0x42a1 || ip == 0x42a5 }
func planSurfaceFillSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || !surfaceFillPixelIP(v.IP()) || v.Read8(0x449) != 0x13 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 {
		return nil
	}
	mv, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	model, known := mv.CPUModel()
	if !known || model != cpu.Model80386 {
		return nil
	}
	var code [260]byte
	copyView(v, 0x48a10+0x41af, code[:])
	if sha256.Sum256(code[:]) != rectangleCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 16)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	s, r, ok := primitiveSurface(v, v.Read16(stack+12))
	if !ok || v.BX() < s.top || v.BX() > s.bottom {
		return nil
	}
	color := v.Read16(stack + 14)
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	if word(0x1129) != color || word(0x113b) != s.left || word(0x113d) != s.right {
		return nil
	}
	width := s.right - s.left + 1
	count, unit := width/2, uint16(2)
	at := uint32(v.BX()-r.top)*uint32(r.right-r.left+1) + uint32(s.left-r.left)
	if v.IP() == 0x42a5 {
		at += uint32(width/2) * 2
		count, unit = width%2, 1
		if v.Flags()&cpu.CF != 0 {
			return nil
		}
	} else if v.Flags()&cpu.CF != width%2 {
		return nil
	}
	value := (color & 255) * 257
	if v.AX() != value || v.CX() != count || v.DI() != uint16(at) || at+uint32(count)*uint32(unit) > 64000 {
		return nil
	}
	p := primitivePlan(v)
	p.After.IP += 2
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] += count * unit
	p.Cycles = 3 + 5*uint64(count)
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	if count > 0 {
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + at, Value: value, Count: count, Width: uint8(unit), Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	return p
}
func surfaceFillFixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && surfaceFillPixelIP(w.IP) {
			return true
		}
	}
	for _, w := range transfers {
		if w.CS == 0x48a1 && surfaceFillPixelIP(w.IP) {
			return true
		}
	}
	return false
}
