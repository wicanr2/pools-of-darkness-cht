// Typed Go filled-ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFilledEllipseFirst(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !validEllipseFrame(v) || v.Flags()&cpu.DF != 0 {
		return nil
	}
	if v.CS() != 0x48a1 || v.IP() != 0x3b3d || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
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
	filledEllipseRows(v, p, int32(x), int32(y), s, 0)
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
		imageWord(p, 0x112f, uint16(y), oldN+1, 0x3c37)
		emit(0x1151, gy, 0x3c4b, 0x3c4e, oldN+7)
		emit(0x1149, uint32(err), 0x3c52, 0x3c56, oldN+9)
	}
	oldN := p.Instructions
	x++
	gx += 2 * b
	term := gx + b
	oldErr := uint32(err)
	err += int32(term)
	charge(12, 65)
	imageWord(p, 0x112b, uint16(x), oldN+1, 0x3c5a)
	emit(0x114d, gx, 0x3c6d, 0x3c70, oldN+6)
	emit(0x1149, uint32(err), 0x3c7c, 0x3c80, oldN+10)
	carry := uint16(0)
	if uint32(uint16(oldErr))+uint32(uint16(term)) > 65535 {
		carry = 1
	}
	p.After.Flags = ellipseAddFlags(v.Flags(), uint16(oldErr>>16), uint16(term>>16), carry)
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.BX] = uint16(term), uint16(term>>16), uint16(y)
	p.After.IP = 0x3b3d
	p.LastIP = 0x3c84
	if p.Instructions > budget {
		return nil
	}
	return p
}

var filledEllipseCode = smallFingerprint(1099, "ee2f7dba367fe53d107b8dc0a5fc8ac1e426f7946a63699a9b2f06059d592e8c")
