// Typed Go filled-ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFilledEllipseSecond(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !validEllipseFrame(v) || v.Flags()&cpu.DF != 0 {
		return nil
	}
	if v.CS() != 0x48a1 || v.IP() != 0x3cc0 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [1099]byte
	copyView(v, 0x48a10+0x39b2, code[:])
	if sha256.Sum256(code[:]) != filledEllipseCode.sum {
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
	filledEllipseRows(v, p, int32(x), int32(y), s, 367)
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
		imageWord(p, 0x112b, uint16(x), oldN+1, 0x3da6)
		emit(0x114d, gx, 0x3db9, 0x3dbc, oldN+6)
		emit(0x1149, uint32(err), 0x3dc0, 0x3dc4, oldN+8)
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
	imageWord(p, 0x112f, uint16(y), oldN+1, 0x3dc8)
	emit(0x1151, gy, 0x3ddb, 0x3dde, oldN+6)
	emit(0x1149, uint32(err), 0x3dea, 0x3dee, oldN+10)
	borrow := uint16(0)
	if uint16(oldErr) < uint16(term) {
		borrow = 1
	}
	subFlags := ellipseSubFlags(v.Flags(), uint16(oldErr>>16), uint16(term>>16), borrow)
	p.After.Flags = nativeCompareFlags(v.Flags(), oldY, 1)&^cpu.CF | subFlags&cpu.CF | v.Flags()&cpu.DF
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.BX] = uint16(term), uint16(term>>16), uint16(y)
	p.After.IP = 0x3cc0
	p.LastIP = 0x3df5
	if terminal {
		p.After.R[cpu.DI] = v.Read16(stack)
		p.After.R[cpu.SI] = v.Read16(stack + 2)
		p.After.Seg[cpu.DS] = v.Read16(stack + 4)
		p.After.R[cpu.BP] = v.Read16(stack + 6)
		p.After.R[cpu.SP] += 12
		p.After.IP = v.Read16(stack + 8)
		p.After.Seg[cpu.CS] = v.Read16(stack + 10)
		p.LastIP = 0x3dfc
	}
	if p.Instructions > budget {
		return nil
	}
	return p
}
