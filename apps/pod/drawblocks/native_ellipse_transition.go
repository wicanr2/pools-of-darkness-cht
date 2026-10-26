// Typed Go ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planEllipseTransition(v CPUView, budget uint64) *machine.HostWorkPlan {
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
	if gx < gy || wide(0x114d) != gx || wide(0x1151) != gy || wide(0x1155) != a || wide(0x1159) != b || wide(0x115d) != 2*a || wide(0x1161) != 2*b {
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
	p.Instructions, p.Cycles, p.LastIP = 26, 109, 0x37f5
	if v.DOSBoxCost() {
		p.Cycles = 26
	}
	if budget < 26 {
		return nil
	}
	p.Writes = append(p.Writes, machine.HostWrite{Address: stack - 2, Value: byte(y), Step: 7, CS: 0x48a1, IP: 0x37bd}, machine.HostWrite{Address: stack - 1, Value: byte(y >> 8), Step: 7, CS: 0x48a1, IP: 0x37bd})
	diff := int32(a - b)
	delta := (diff + (diff >> 1) - int32(gx) - int32(gy)) >> 1
	value := uint32(expected + delta)
	imageWord(p, 0x1149, uint16(value), 24, 0x37ed)
	imageWord(p, 0x114b, uint16(value>>16), 25, 0x37f1)
	carry := uint16(0)
	if uint32(uint16(expected))+uint32(uint16(delta)) > 65535 {
		carry = 1
	}
	p.After.Flags = ellipseAddFlags(v.Flags(), uint16(uint32(expected)>>16), uint16(uint32(delta)>>16), carry)
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.CX] = uint16(delta), uint16(uint32(delta)>>16), uint16(uint32(diff)>>16)
	p.After.IP = 0x37f6
	return p
}
