package drawblocks

// Spec025/#49: a typed font bit and cell cursor produce each resident pixel.

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planResidentGlyphPixel(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x1b2e || (v.IP() != 0x11ea && v.IP() != 0x11f1) || v.Read8(0x449) != 0x13 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 {
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
	var code [93]byte
	for i := range code {
		code[i] = v.Read8(0x1c486 + uint32(i))
	}
	digest := sha256.Sum256(code[:])
	if hex.EncodeToString(digest[:]) != "d1c978b914f28d1a978a2cae1b7a1c52f4b5d233176e8bd6c3029b05c9b4958b" {
		return nil
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	overlap := func(a, b, c, d uint32) bool { return a < d && c < b }
	if uint32(v.SP())+20 > 65536 || stack < 0x500 || stack+20 > 0xa0000 || v.BP() != v.SP()+4 || overlap(stack, stack+20, 0x1c486, 0x1c4e3) {
		return nil
	}
	originalDS := v.Read16(stack + 2)
	table := uint32(originalDS)*16 + 0x9d86
	if table+4 > 0xa0000 || overlap(stack, stack+20, table, table+4) {
		return nil
	}
	ch, bg, fg := v.Read8(stack+10), v.Read8(stack+12), v.Read8(stack+14)
	y, x := v.Read16(stack+16), v.Read16(stack+18)
	if x >= 40 || y >= 25 {
		return nil
	}
	base, seg := v.Read16(table), v.Read16(table+2)
	off := uint16(uint32(base) + uint32(ch)*8)
	src := uint32(seg)*16 + uint32(off)
	if uint32(off)+8 > 65536 || src+8 > 0xa0000 || overlap(stack, stack+20, src, src+8) || overlap(src, src+8, 0x1c486, 0x1c4e3) {
		return nil
	}
	rows := v.Read16(stack)
	if rows < 1 || rows > 8 || v.CX() < 1 || v.CX() > 8 || v.DS() != seg {
		return nil
	}
	row, col := 8-rows, 8-v.CX()
	mask := uint16(0x80 >> col)
	on := v.Read8(src+uint32(row))&byte(mask) != 0
	ip, value := uint16(0x11ea), bg
	if on {
		ip, value = 0x11f1, fg
	}
	expected := uint16(0)
	if on {
		expected = mask
	}
	flags := uint16(0)
	if expected == 0 {
		flags = cpu.PF | cpu.ZF
	}
	arithmetic := uint16(cpu.CF | cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
	at := (uint32(y)*320+uint32(x))*8 + uint32(row)*320 + uint32(col)
	if v.IP() != ip || v.AX() != uint16(value) || v.BX() != mask || v.SI() != off+row || v.DI() != uint16(at) || v.Flags()&arithmetic != flags&arithmetic {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, Cycles: 7, LastCS: v.CS(), LastIP: v.IP(),
		After:  machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI() + 1}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP() + 1, Flags: v.Flags()},
		Writes: []machine.HostWrite{{Address: 0xa0000 + at, Value: value, Step: 1, CS: v.CS(), IP: v.IP()}}}
	if v.DOSBoxCost() {
		p.Cycles = 2
	}
	return p
}
