package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var floodSpanCode = smallFingerprint(130, "776aafc34711e77e5120c2d65532dd361555f8de46855ba58e1862896bf7a8d1")

// Search the current flood-fill row for the first different color. The original
// caller owns its pending-row queue; this kernel owns the pixel search. Guest
// registers and timing below describe its handoff, not an opcode interpreter.
func planFloodScan(v CPUView, budget uint64) *machine.HostWorkPlan {
	backward := v.IP() == 0x55cd
	if budget == 0 || v.CS() != 0x48a1 || (!backward && v.IP() != 0x5602) || v.Read8(0x449) != 0x13 || v.ES() != 0xa000 || (v.Flags()&cpu.DF != 0) != backward {
		return nil
	}
	model, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	m, known := model.CPUModel()
	if !known || m != cpu.Model80386 {
		return nil
	}
	var code [130]byte
	copyView(v, 0x48a10+0x5593, code[:])
	if sha256.Sum256(code[:]) != floodSpanCode.sum {
		return nil
	}
	count, start := int(v.CX()), int(v.DI())
	width := int(v.Read16(0x48a10 + 0xedd))
	height := int(v.Read16(0x48a10 + 0xedf))
	if width < 1 || width > 320 || height < 1 || height > 200 || start >= width*height || count > width {
		return nil
	}
	if count > 0 && ((!backward && start%width+count > width) || (backward && count > start%width+1)) {
		return nil
	}
	step := 1
	if backward {
		step = -1
	}
	position, visited, last := start, 0, byte(0)
	for visited < count {
		last = v.Read8(0xa0000 + uint32(position))
		position += step
		visited++
		if last != byte(v.AX()) {
			break
		}
	}
	after := boundedMemoryRegisters(v)
	after.R[cpu.CX] -= uint16(visited)
	after.R[cpu.DI] = uint16(position)
	after.IP += 2
	if visited != 0 {
		after.Flags = glyph5aa4Compare8Flags(v.Flags(), byte(v.AX()), last)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: after, Instructions: 1, Cycles: 3 + 5*uint64(visited), LastCS: 0x48a1, LastIP: v.IP()}
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(visited)
	}
	return p
}

// These compatibility blocks contain a row scan. Stop before them even if the
// native kernel declines an unknown context; never classify fixed SCAS as Go.
func fixedFloodScan(start uint16, b block) bool {
	if b.cs != 0x48a1 {
		return false
	}
	for _, ip := range []uint32{0x55cd, 0x5602, 0x2cf4} {
		if uint32(start) <= ip && ip < uint32(start)+uint32(b.size) {
			return true
		}
	}
	return false
}

// The complete point-read call is preferred. A short event budget can leave
// just its pixel read; validate the same surface and caller frame here.
func planPointReadPixel(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || v.IP() != 0x2cf4 || v.Read8(0x449) != 0x13 || v.DS() != 0x48a1 || v.ES() != 0xa000 {
		return nil
	}
	model, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	m, known := model.CPUModel()
	if !known || m != cpu.Model80386 {
		return nil
	}
	var code [248]byte
	copyView(v, 0x48a10+0x2c04, code[:])
	if sha256.Sum256(code[:]) != pointReadCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 18)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	s, r, ok := primitiveSurface(v, v.Read16(stack+12))
	if !ok {
		return nil
	}
	x, y := v.Read16(stack+14), v.Read16(stack+16)
	if int16(x) < int16(s.left) || int16(x) > int16(s.right) || int16(y) < int16(s.top) || int16(y) > int16(s.bottom) {
		return nil
	}
	at := uint32(y-r.top)*uint32(r.right-r.left+1) + uint32(x-r.left)
	if at >= 64000 || v.BX() != uint16(at) || v.DI() != uint16(at) || v.AX()&0xff00 != 0 {
		return nil
	}
	p := primitivePlan(v)
	p.After.R[cpu.AX] = uint16(v.Read8(0xa0000 + at))
	p.After.IP += 3
	p.Cycles = 7
	if v.DOSBoxCost() {
		p.Cycles = 1
	}
	return p
}
