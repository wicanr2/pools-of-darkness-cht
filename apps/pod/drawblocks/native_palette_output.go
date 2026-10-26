package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteOutputCode = smallFingerprint(27, "7134f5f3fbaf2adf007afa01632ff5ec980bd4c127a095defaa954fd831b4801")

// planPaletteOutput decodes one index/RGB job and submits it to the VGA DAC
// service. Guest stack, flags and return effects are separate compatibility data.
func planPaletteOutput(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x2aa8 || v.Read8(0x449) != 0x13 || budget < 16 || v.SP() < 2 || v.SP() > 65524 {
		return nil
	}
	var code [27]byte
	copyView(v, 0x48a10+0x2aa8, code[:])
	if sha256.Sum256(code[:]) != paletteOutputCode.sum {
		return nil
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if stack-2 < 0x500 || stack+12 > 0xa0000 || stack-2 < 0x48a10+0x7580 && stack+12 > 0x48a10 {
		return nil
	}
	index := v.Read16(stack + 4)
	rgb := [3]uint16{v.Read16(stack + 6), v.Read16(stack + 8), v.Read16(stack + 10)}
	if index > 255 || rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 16, Cycles: 130, LastCS: 0x48a1, LastIP: 0x2ac2, After: machine.HostRegisters{R: [8]uint16{rgb[2], v.CX(), 0x3c9, v.BX(), v.SP() + 4, v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.Read16(stack + 2), v.SS(), v.DS()}, IP: v.Read16(stack), Flags: v.Flags()&^uint16(cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF) | cpu.PF | cpu.IF}}
	p.Writes = []machine.HostWrite{{Address: stack - 2, Value: byte(v.BP()), Step: 1, CS: 0x48a1, IP: 0x2aa8}, {Address: stack - 1, Value: byte(v.BP() >> 8), Step: 1, CS: 0x48a1, IP: 0x2aa8}}
	job := machine.HostPaletteWrite{Index: byte(index), RGB: [3]byte{byte(rgb[0]), byte(rgb[1]), byte(rgb[2])}, Steps: [4]uint64{6, 9, 11, 13}, Cycles: [4]uint64{32, 56, 78, 100}}
	if v.DOSBoxCost() {
		p.Cycles = 16
		job.Cycles = job.Steps
	}
	p.PaletteWrites = []machine.HostPaletteWrite{job}
	return p
}
