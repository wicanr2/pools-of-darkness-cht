// Typed Go ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planEllipseSecond(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !validEllipseFrame(v) {
		return nil
	}
	if v.CS() != 0x48a1 || v.IP() != 0x37f6 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
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
	if wide(0x114d) != gx || wide(0x1151) != gy || wide(0x1155) != a || wide(0x1159) != b || wide(0x115d) != 2*a || wide(0x1161) != 2*b {
		return nil
	}
	expected, valid := ellipseSecondError(uint32(rx), uint32(ry), x, y)
	if !valid || int32(wide(0x1149)) != expected {
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
	emit := func(off uint16, value uint32, ipLo, ipHi uint16, step uint64) {
		imageWord(p, off, uint16(value), step, ipLo)
		imageWord(p, off+2, uint16(value>>16), step+1, ipHi)
	}
	charge(1, 4)
	p.Writes = append(p.Writes, machine.HostWrite{Address: stack - 2, Value: byte(y), Step: p.Instructions, CS: 0x48a1, IP: 0x37f6}, machine.HostWrite{Address: stack - 1, Value: byte(y >> 8), Step: p.Instructions, CS: 0x48a1, IP: 0x37f6})
	charge(1, 6)
	ellipseFourPoints(v, p, int32(x), int32(y), s, 0x384e)
	charge(1, 4)
	charge(2, 8)
	err := expected
	if err >= 0 {
		if !v.DOSBoxCost() {
			p.Cycles += 5
		}
	} else {
		oldN := p.Instructions
		x++
		gx += 2 * b
		err += int32(gx)
		charge(9, 46)
		imageWord(p, 0x112b, uint16(x), oldN+1, 0x395b)
		emit(0x114d, gx, 0x396e, 0x3971, oldN+6)
		emit(0x1149, uint32(err), 0x3975, 0x3979, oldN+8)
	}
	oldN := p.Instructions
	oldY := uint16(y)
	oldErr := uint32(err)
	y--
	gy -= 2 * a
	term := gy - a
	err -= int32(term)
	terminal := oldY == 0
	if terminal {
		charge(18, 107)
	} else {
		charge(14, 69)
	}
	imageWord(p, 0x112f, uint16(y), oldN+1, 0x397d)
	emit(0x1151, gy, 0x3990, 0x3993, oldN+6)
	emit(0x1149, uint32(err), 0x399f, 0x39a3, oldN+10)
	borrow := uint16(0)
	if uint16(oldErr) < uint16(term) {
		borrow = 1
	}
	subFlags := ellipseSubFlags(v.Flags(), uint16(oldErr>>16), uint16(term>>16), borrow)
	p.After.Flags = nativeCompareFlags(v.Flags(), oldY, 1)&^cpu.CF | subFlags&cpu.CF | v.Flags()&cpu.DF
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.BX], p.After.R[cpu.CX] = uint16(term), uint16(term>>16), uint16(y), word(0x1129)
	p.After.IP = 0x37f6
	p.LastIP = 0x39aa
	if terminal {
		p.After.R[cpu.DI] = v.Read16(stack)
		p.After.R[cpu.SI] = v.Read16(stack + 2)
		p.After.Seg[cpu.DS] = v.Read16(stack + 4)
		p.After.R[cpu.BP] = v.Read16(stack + 6)
		p.After.R[cpu.SP] += 12
		p.After.IP = v.Read16(stack + 8)
		p.After.Seg[cpu.CS] = v.Read16(stack + 10)
		p.LastIP = 0x39b1
	}
	if p.Instructions > budget {
		return nil
	}
	return p
}
func ellipseSubFlags(before, a, b, borrow uint16) uint16 {
	result := a - b - borrow
	f := nativeCompareFlags(before, result, 0) &^ uint16(cpu.AF|cpu.OF|cpu.CF)
	if uint32(a) < uint32(b)+uint32(borrow) {
		f |= cpu.CF
	}
	if uint32(a&15) < uint32(b&15)+uint32(borrow) {
		f |= cpu.AF
	}
	if ((a^b)&(a^result))&0x8000 != 0 {
		f |= cpu.OF
	}
	return f | before&cpu.DF
}
func ellipseSecondError(rx, ry, tx, ty uint32) (int32, bool) {
	a, b := rx*rx, ry*ry
	x, y := uint32(0), ry
	gx, gy := uint32(0), 2*a*ry
	err := int32(a/4 + b - a*ry)
	for gx < gy {
		if err >= 0 {
			y--
			gy -= 2 * a
			err -= int32(gy)
		}
		x++
		gx += 2 * b
		err += int32(gx + b)
	}
	diff := int32(a - b)
	err += (diff + (diff >> 1) - int32(gx) - int32(gy)) >> 1
	for int32(y) >= 0 {
		if x == tx && y == ty {
			return err, true
		}
		if x > tx || y < ty {
			return 0, false
		}
		if err < 0 {
			x++
			gx += 2 * b
			err += int32(gx)
		}
		y--
		gy -= 2 * a
		err -= int32(gy - a)
	}
	return 0, false
}
