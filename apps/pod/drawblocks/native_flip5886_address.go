// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFlip5886Address(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x5947 {
		return nil
	}
	f, ok := readFlip5886Frame(v)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	p.Instructions, p.Cycles = 16, 65
	p.LastIP = 0x597a
	row := f.top
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	if f.vertical {
		p.Instructions, p.Cycles, p.LastIP = 29, 110, 0x59a7
		row = f.bottom
		p.After.Flags = copyAddFlags(v.Flags(), f.s.left+row*256, row*64)
	}
	p.After.R[cpu.AX] = 320
	p.After.R[cpu.BX] = row*320 + f.s.left
	p.After.R[cpu.DX] = row * 64
	p.After.R[cpu.SI] = f.top*320 + f.s.left
	p.After.R[cpu.DI] = row*320 + f.s.left
	p.After.IP = 0x59a9
	if v.DOSBoxCost() {
		p.Cycles = p.Instructions
	}
	if p.Instructions > budget {
		return nil
	}
	return p
}
