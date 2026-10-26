package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteReadCode = smallFingerprint(37, "1d90a923f1fa7125eb8bdf8b166538ed7a7c2feaf1f5c4e039e2c48403de50de")

// Spec025/#49: read one typed RGB color through the existing VGA service;
// register, pointer and stack effects only preserve the original ABI.
func planPaletteRead(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x2aeb || v.Read8(0x449) != 0x13 || budget < 20 {
		return nil
	}
	var code [37]byte
	copyView(v, 0x48a10+0x2aeb, code[:])
	if sha256.Sum256(code[:]) != paletteReadCode.sum {
		return nil
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if v.SP() < 2 || v.SP() > 65524 || stack < 0x502 || stack+12 > 0xa0000 || overlapCode(stack-2, stack+12) {
		return nil
	}
	index := v.Read16(stack + 4)
	if index > 255 {
		return nil
	}
	reader, ok := v.(interface{ PaletteRGB(byte) ([3]byte, bool) })
	if !ok {
		return nil
	}
	rgb, known := reader.PaletteRGB(byte(index))
	if !known || rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return nil
	}
	addresses := [3]uint32{}
	pointers := [3]uint16{}
	for i := range addresses {
		pointers[i] = v.Read16(stack + 6 + uint32(i)*2)
		at := uint32(v.DS())*16 + uint32(pointers[i])
		addresses[i] = at
		if pointers[i] > 65534 || at < 0x500 || at+2 > 0xa0000 || overlapCode(at, at+2) || at < stack+12 && at+2 > stack-2 {
			return nil
		}
	}
	p := imageBasePlan(v)
	p.Instructions, p.Cycles, p.LastIP = 20, 150, 0x2b0f
	p.After.R[cpu.AX] = uint16(rgb[2])
	p.After.R[cpu.BX] = pointers[2]
	p.After.R[cpu.DX] = 0x3c9
	p.After.R[cpu.SP] += 4
	p.After.IP = v.Read16(stack)
	p.After.Seg[cpu.CS] = v.Read16(stack + 2)
	p.After.Flags = v.Flags()&^uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF) | cpu.PF | cpu.ZF | cpu.IF
	p.Writes = []machine.HostWrite{{Address: stack - 2, Value: byte(v.BP()), Step: 1, CS: 0x48a1, IP: 0x2aeb}, {Address: stack - 1, Value: byte(v.BP() >> 8), Step: 1, CS: 0x48a1, IP: 0x2aeb}}
	for i, at := range addresses {
		step, ip := uint64(11+i*3), uint16(0x2aff+i*6)
		p.Writes = append(p.Writes, machine.HostWrite{Address: at, Value: rgb[i], Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: at + 1, Value: 0, Step: step, CS: 0x48a1, IP: ip})
	}
	job := machine.HostPaletteRead{Index: byte(index), RGB: rgb, Steps: [4]uint64{7, 9, 12, 15}, Cycles: [4]uint64{34, 52, 80, 108}}
	if v.DOSBoxCost() {
		p.Cycles = 20
		job.Cycles = job.Steps
	}
	p.PaletteReads = []machine.HostPaletteRead{job}
	return p
}
