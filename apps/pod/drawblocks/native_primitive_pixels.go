package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func primitivePixelIP(ip uint16) bool {
	switch ip {
	case 0x2bfc, 0x345c, 0x3460, 0x309c, 0x30b9, 0x3206, 0x321c:
		return true
	}
	return false
}

func primitiveSurface(v CPUView, slot uint16) (s, root rectangleSurface, ok bool) {
	if slot > 255 {
		return s, root, false
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	desc := func(i uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + i), word(0x2bd + i), word(0x4bd + i), word(0x6bd + i), word(0x8bd + i), word(0xabd + i), word(0xcbd + i)}
	}
	s = desc(slot * 2)
	root = s
	links := 0
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return s, root, false
		}
		root = desc(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return s, root, false
	}
	for i, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, root.right - root.left + 1, root.bottom - root.top + 1} {
		if word(0xebd+uint16(i)*2) != value {
			return s, root, false
		}
	}
	return s, root, true
}

func primitivePlan(v CPUView) *machine.HostWorkPlan {
	return &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, After: boundedMemoryRegisters(v), LastCS: 0x48a1, LastIP: v.IP()}
}

func planPointPixel(v CPUView) *machine.HostWorkPlan {
	if v.DS() != 0x48a1 || v.ES() != 0xa000 {
		return nil
	}
	var code [244]byte
	copyView(v, 0x48a10+0x2b10, code[:])
	if sha256.Sum256(code[:]) != pointWriteCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 20)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	s, r, ok := primitiveSurface(v, v.Read16(stack+12))
	if !ok {
		return nil
	}
	x, y, color := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18)
	if int16(x) < int16(s.left) || int16(x) > int16(s.right) || int16(y) < int16(s.top) || int16(y) > int16(s.bottom) || v.AX() != color {
		return nil
	}
	at := uint32(y-r.top)*uint32(r.right-r.left+1) + uint32(x-r.left)
	if at >= 64000 || v.DI() != uint16(at) {
		return nil
	}
	p := primitivePlan(v)
	p.After.IP += 3
	p.Cycles = 7
	if v.DOSBoxCost() {
		p.Cycles = 1
	}
	p.Writes = []machine.HostWrite{{Address: 0xa0000 + at, Value: byte(color), Step: 1, CS: 0x48a1, IP: v.IP()}}
	return p
}

func planRegionSpan(v CPUView) *machine.HostWorkPlan {
	if v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 {
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
	s, r, ok := primitiveSurface(v, v.Read16(stack+12))
	if !ok {
		return nil
	}
	x0, y0, x1, y1, color := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18), v.Read16(stack+20), v.Read16(stack+22)
	if int16(x0) > int16(x1) || int16(y0) > int16(y1) {
		return nil
	}
	left, right, top, bottom := int32(int16(x0)), int32(int16(x1)), int32(int16(y0)), int32(int16(y1))
	if left < int32(s.left) {
		left = int32(s.left)
	}
	if right > int32(s.right) {
		right = int32(s.right)
	}
	if top < int32(s.top) {
		top = int32(s.top)
	}
	if bottom > int32(s.bottom) {
		bottom = int32(s.bottom)
	}
	if left > right || top > bottom || int32(v.SI()) < top || int32(v.SI()) > bottom {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	if word(0x113b) != uint16(left) || word(0x113d) != uint16(right) || word(0x1129) != color {
		return nil
	}
	width := uint16(right - left + 1)
	count, unit := width/2, uint16(2)
	at := uint32(v.SI()-r.top)*uint32(r.right-r.left+1) + uint32(left-int32(r.left))
	if v.IP() == 0x3460 {
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

func planDiagonalPixel(v CPUView) *machine.HostWorkPlan {
	c, ok := readLineCursor(v)
	if !ok || c.at >= 64000 {
		return nil
	}
	p := primitivePlan(v)
	p.Cycles = 7
	if c.steep {
		p.After.IP += 3
		if v.DOSBoxCost() {
			p.Cycles = 1
		}
		p.Writes = []machine.HostWrite{{Address: 0xa0000 + c.at, Value: byte(v.AX()), Step: 1, CS: 0x48a1, IP: v.IP()}}
	} else {
		p.After.IP++
		if c.decrease {
			p.After.R[cpu.DI]--
		} else {
			p.After.R[cpu.DI]++
		}
		if v.DOSBoxCost() {
			p.Cycles = 2
		}
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + c.at, Value: v.AX() & 255, Count: 1, Width: 1, Backward: c.decrease, Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	return p
}

func planPrimitivePixel(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || !primitivePixelIP(v.IP()) || v.Read8(0x449) != 0x13 {
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
	switch v.IP() {
	case 0x2bfc:
		return planPointPixel(v)
	case 0x345c, 0x3460:
		return planRegionSpan(v)
	}
	return planDiagonalPixel(v)
}

func primitiveFixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && primitivePixelIP(w.IP) {
			return true
		}
	}
	for _, w := range transfers {
		if w.CS == 0x48a1 && primitivePixelIP(w.IP) {
			return true
		}
	}
	return false
}
