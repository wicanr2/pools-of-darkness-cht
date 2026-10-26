// Typed Go ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func ellipseAddFlags(before, a, b uint16, carry uint16) uint16 {
	result := a + b + carry
	f := nativeCompareFlags(before, result, 0) &^ uint16(cpu.AF|cpu.OF|cpu.CF)
	if uint32(a)+uint32(b)+uint32(carry) > 65535 {
		f |= cpu.CF
	}
	if uint32(a&15)+uint32(b&15)+uint32(carry) > 15 {
		f |= cpu.AF
	}
	if (^(a^b)&(a^result))&0x8000 != 0 {
		f |= cpu.OF
	}
	return f | before&cpu.DF
}
func ellipseFourPoints(v CPUView, p *machine.HostWorkPlan, x, y int32, s rectangleSurface, pixelBase uint16) {
	cx, cy := int32(int16(v.Read16(0x48a10+0x112d))), int32(int16(v.Read16(0x48a10+0x1131)))
	charge := func(n, c uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n
		} else {
			p.Cycles += c
		}
	}
	for index, xy := range [][2]int32{{cx + x, cy + y}, {cx + x, cy - y}, {cx - x, cy + y}, {cx - x, cy - y}} {
		px, py := xy[0], xy[1]
		charge(5, 26)
		p.After.R[cpu.DI], p.After.R[cpu.SI], p.After.R[cpu.BX] = uint16(px), uint16(py), uint16(px)
		if px < int32(s.left) {
			charge(2, 13)
			continue
		}
		charge(2, 8)
		if px > int32(s.right) {
			charge(2, 13)
			continue
		}
		charge(2, 8)
		charge(1, 2)
		p.After.R[cpu.BX] = uint16(py)
		if py < int32(s.top) {
			charge(2, 13)
			continue
		}
		charge(2, 8)
		if py > int32(s.bottom) {
			charge(2, 13)
			continue
		}
		charge(2, 8)
		charge(13, 47)
		at := uint32(py*320 + px)
		p.After.R[cpu.DI] = uint16(at)
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0xa0000 + at, Value: byte(v.Read16(0x48a10 + 0x1129)), Step: p.Instructions, CS: 0x48a1, IP: pixelBase + uint16(index)*0x56})
	}
}
func planEllipseFirst(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !validEllipseFrame(v) {
		return nil
	}
	if v.CS() != 0x48a1 || v.IP() != 0x35f4 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [1352]byte
	copyView(v, 0x48a10+0x346a, code[:])
	if sha256.Sum256(code[:]) != ellipseOutlineCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 2, 24)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	wide := func(off uint16) uint32 { return uint32(word(off)) | uint32(word(off+2))<<16 }
	rx, ry := v.Read16(stack+18), v.Read16(stack+20)
	if rx == 0 || rx > 160 || ry == 0 || ry > 100 {
		return nil
	}
	x, y := uint32(word(0x112b)), uint32(word(0x112f))
	if x > uint32(rx)+1 || y > uint32(ry) || v.BX() != uint16(y) {
		return nil
	}
	a, b := uint32(rx)*uint32(rx), uint32(ry)*uint32(ry)
	gx, gy := 2*b*x, 2*a*y
	if gx >= gy || wide(0x114d) != gx || wide(0x1151) != gy || wide(0x1155) != a || wide(0x1159) != b || wide(0x115d) != 2*a || wide(0x1161) != 2*b {
		return nil
	}
	expected := int32(b*(x+1)*(x+1) + a*y*(y-1) + a/4 - a*uint32(ry)*uint32(ry))
	if int32(wide(0x1149)) != expected {
		return nil
	}
	s := rectangleSurface{word(0xebd), word(0xebf), word(0xec1), word(0xec3), word(0xec5), word(0xec7), word(0xec9)}
	if s.left > s.right || s.right > 319 || s.top > s.bottom || s.bottom > 199 || word(0xecf) != 0xa000 || word(0xed3) != 0 || word(0xed5) != 0 || word(0xed7) != 319 || word(0xed9) != 199 || word(0xedd) != 320 {
		return nil
	}
	p := imageBasePlan(v)
	charge := func(n, c uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n
		} else {
			p.Cycles += c
		}
	}
	emit := func(off uint16, value uint32, ipLo, ipHi uint16, lowStep uint64) {
		imageWord(p, off, uint16(value), lowStep, ipLo)
		imageWord(p, off+2, uint16(value>>16), lowStep+1, ipHi)
	}
	charge(5, 27)
	charge(1, 4)
	p.Writes = append(p.Writes, machine.HostWrite{Address: stack - 2, Value: byte(y), Step: p.Instructions, CS: 0x48a1, IP: 0x3608}, machine.HostWrite{Address: stack - 1, Value: byte(y >> 8), Step: p.Instructions, CS: 0x48a1, IP: 0x3608})
	charge(1, 6)
	ellipseFourPoints(v, p, int32(x), int32(y), s, 0x3660)
	charge(1, 4)
	charge(2, 8)
	err := expected
	if err < 0 {
		if !v.DOSBoxCost() {
			p.Cycles += 5
		}
	} else {
		oldN := p.Instructions
		y--
		gy -= 2 * a
		err -= int32(gy)
		charge(10, 48)
		imageWord(p, 0x112f, uint16(y), oldN+1, 0x376d)
		emit(0x1151, gy, 0x3781, 0x3784, oldN+7)
		emit(0x1149, uint32(err), 0x3788, 0x378c, oldN+9)
	}
	oldN := p.Instructions
	x++
	gx += 2 * b
	term := gx + b
	oldErr := uint32(err)
	err += int32(term)
	charge(12, 65)
	imageWord(p, 0x112b, uint16(x), oldN+1, 0x3790)
	emit(0x114d, gx, 0x37a3, 0x37a6, oldN+6)
	emit(0x1149, uint32(err), 0x37b2, 0x37b6, oldN+10)
	carry := uint16(0)
	if uint32(uint16(oldErr))+uint32(uint16(term)) > 65535 {
		carry = 1
	}
	p.After.Flags = ellipseAddFlags(v.Flags(), uint16(oldErr>>16), uint16(term>>16), carry)
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.BX], p.After.R[cpu.CX] = uint16(term), uint16(term>>16), uint16(y), word(0x1129)
	p.After.IP = 0x35f4
	p.LastIP = 0x37ba
	if p.Instructions > budget {
		return nil
	}
	return p
}

var ellipseOutlineCode = smallFingerprint(1352, "4b118914e4b4a0f6286f6b8be0eb508710457be02e61eb9887f6dc2c5a9548d5")
