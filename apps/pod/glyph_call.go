package pod

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

type HostCallCPU interface {
	StepCPU
	AX() uint16
	CX() uint16
	DX() uint16
	BX() uint16
	BP() uint16
	SI() uint16
	DI() uint16
	ES() uint16
	Flags() uint16
	DOSBoxCost() bool
}

const glyphCallCodeSHA = "d1c978b914f28d1a978a2cae1b7a1c52f4b5d233176e8bd6c3029b05c9b4958b"

func GlyphHostCallTargets() []uint32 { return []uint32{0x1c486} }

// PlanGlyphCall is the spec025 whole 8x8 glyph effect. It only reads, and a
// nil proposal leaves all guest state and source ownership untouched.
func PlanGlyphCall(v HostCallCPU, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x1b2e || v.IP() != 0x11a6 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [93]byte
	for i := range code {
		code[i] = v.Read8(0x1c486 + uint32(i))
	}
	sum := sha256.Sum256(code[:])
	if hex.EncodeToString(sum[:]) != glyphCallCodeSHA {
		return nil
	}
	b := machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}
	sp := int(b.R[cpu.SP])
	sb := int(b.Seg[cpu.SS]) * 16
	ds := int(b.Seg[cpu.DS]) * 16
	if sp < 6 || sp+14 > 65536 || sb+sp+14 > 0xa0000 || ds+0x9d8a > 0xa0000 {
		return nil
	}
	word := func(i int) uint16 { return uint16(v.Read8(uint32(i))) | uint16(v.Read8(uint32(i+1)))<<8 }
	stack := sb + sp
	ch, bg, fg := v.Read8(uint32(stack+4)), v.Read8(uint32(stack+6)), v.Read8(uint32(stack+8))
	y, x := word(stack+10), word(stack+12)
	if x >= 40 || y >= 25 {
		return nil
	}
	off := (int(word(ds+0x9d86)) + int(ch)*8) & 65535
	src := int(word(ds+0x9d88))*16 + off
	if off+8 > 65536 || src+8 > 0xa0000 {
		return nil
	}
	overlap := func(a, b, c, d int) bool { return a < d && c < b }
	if overlap(stack-6, stack+14, src, src+8) || overlap(stack-6, stack+14, 0x1c486, 0x1c4e3) || overlap(stack-6, stack+14, ds+0x9d86, ds+0x9d8a) {
		return nil
	}
	var font [8]byte
	ones := 0
	for i := range font {
		font[i] = v.Read8(uint32(src + i))
		ones += bits.OnesCount8(font[i])
	}
	instructions := uint64(656 - ones)
	if instructions > budget {
		return nil
	}
	dest := (int(y)*320 + int(x)) * 8
	p := &machine.HostWorkPlan{After: b, Instructions: instructions, Cycles: uint64(2620 - 2*ones), LastCS: b.Seg[cpu.CS], LastIP: 0x1200, Writes: make([]machine.HostWrite, 0, 84)}
	if v.DOSBoxCost() {
		p.Cycles = instructions + 64
	}
	emit := func(addr int, value byte, n uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: uint32(addr), Value: value, Step: n, CS: b.Seg[cpu.CS], IP: ip})
	}
	push := func(addr int, value uint16, n uint64, ip uint16) {
		emit(addr, byte(value), n, ip)
		emit(addr+1, byte(value>>8), n, ip)
	}
	push(stack-2, b.R[cpu.BP], 1, 0x11a6)
	push(stack-4, b.Seg[cpu.DS], 4, 0x11aa)
	n := uint64(21)
	for row := 0; row < 8; row++ {
		n++
		push(stack-6, uint16(8-row), n, 0x11d7)
		n += 2
		for col := 0; col < 8; col++ {
			on := font[row]&(0x80>>col) != 0
			value, ip := bg, uint16(0x11ea)
			if on {
				value, ip = fg, 0x11f1
			}
			n += 6
			emit(0xa0000+dest+row*320+col, value, n, ip)
			if !on {
				n++
			}
			n += 2
		}
		n += 4
	}
	r := &p.After.R
	r[cpu.AX] = uint16(bg)
	if font[7]&1 != 0 {
		r[cpu.AX] = uint16(fg)
	}
	r[cpu.CX] = 0
	r[cpu.DX] = 0
	r[cpu.BX] = 0
	r[cpu.SI] = uint16(off + 8)
	r[cpu.DI] = uint16(dest + 2560)
	r[cpu.SP] += 14
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = word(stack)
	p.After.Seg[cpu.CS] = word(stack + 2)
	a, add := uint16(dest+7*320+8), uint16(312)
	result := a + add
	f := b.Flags &^ uint16(0xcd5|cpu.DF)
	if uint32(a)+uint32(add) > 65535 {
		f |= cpu.CF
	}
	if (a^add^result)&16 != 0 {
		f |= cpu.AF
	}
	if result == 0 {
		f |= cpu.ZF
	}
	if result&0x8000 != 0 {
		f |= cpu.SF
	}
	if (^(a^add)&(a^result))&0x8000 != 0 {
		f |= cpu.OF
	}
	if bits.OnesCount8(byte(result))%2 == 0 {
		f |= cpu.PF
	}
	p.After.Flags = f
	return p
}

func (l *Live) PrepareHostCall(v HostCallCPU, budget uint64) (*machine.HostWorkPlan, error) {
	if !l.ready || l.Inventory {
		return nil, nil
	}
	if l.Map != nil {
		if l.nextMapSample <= v.Steps() {
			return nil, nil
		}
		if remaining := l.nextMapSample - v.Steps(); remaining < budget {
			budget = remaining
		}
	}
	return PlanGlyphCall(v, budget), nil
}
