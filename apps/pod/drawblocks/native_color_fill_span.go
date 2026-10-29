package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// Resume the shared color fill at its typed VGA span boundary. The complete
// seven-instruction call remains the preferred path when its entry budget fits.
func planColorFillSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x1ed4 || v.IP() != 0x1a8e || v.Read8(0x449) != 0x13 || v.Flags()&cpu.DF != 0 {
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
	var raw [20]byte
	copyView(v, 0x1ed40+0x1a7f, raw[:])
	if sha256.Sum256(raw[:]) != smallCalls[0x1ed41a7f].sum {
		return nil
	}
	sp := uint32(v.SP())
	stack := uint32(v.SS())*16 + sp
	if sp+12 > 65536 || stack < 0x500 || stack+12 > 0xa0000 || v.BX() != v.SP() {
		return nil
	}
	color, count := v.Read16(stack+4), v.Read16(stack+6)
	offset, segment := v.Read16(stack+8), v.Read16(stack+10)
	address := uint32(segment)*16 + uint32(offset)
	if uint32(offset)+uint32(count) > 65536 || address < 0xa0000 || address+uint32(count) > 0xafa00 || byte(v.AX()) != byte(color) || v.CX() != count || v.DI() != offset || v.ES() != segment {
		return nil
	}
	after := boundedMemoryRegisters(v)
	after.IP = 0x1a90
	after.R[cpu.CX] = 0
	after.R[cpu.DI] += count
	p := &machine.HostWorkPlan{Algorithmic: true, After: after, Instructions: 1, Cycles: 3 + 5*uint64(count), LastCS: 0x1ed4, LastIP: 0x1a8e}
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	if count != 0 {
		p.Transfers = []machine.HostTransfer{{Address: address, Value: uint16(byte(color)), Count: count, Width: 1, Step: 1, CS: 0x1ed4, IP: 0x1a8e}}
	}
	return p
}

func colorFillFixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x1ed4 && w.IP == 0x1a8e && w.Address >= 0xa0000 {
			return true
		}
	}
	for _, tr := range transfers {
		if tr.CS == 0x1ed4 && tr.IP == 0x1a8e && tr.Address >= 0xa0000 {
			return true
		}
	}
	return false
}
