// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// Pixel data is reconstructed from the typed plane/row cursor. Guest AX/CX
// validate the handoff; neither is the source of the emitted color/data.
func planBitmap51Span(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 {
		return nil
	}
	p := imageBasePlan(v)
	switch v.IP() {
	case 0x53d8:
		f, _, _, done, ok := bitmap51PlanarState(v)
		if !ok {
			return nil
		}
		color, last, _ := bitmap51Color(v, f, done)
		if v.AX() != uint16(color) || v.BX() != 8*f.stride || byte(v.CX()) != last {
			return nil
		}
		p.Instructions, p.Cycles, p.LastIP = 1, 7, v.IP()
		if v.DOSBoxCost() {
			p.Cycles = 2
		}
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + uint32(f.y)*320 + uint32(f.s.left+done), Value: uint16(color), Count: 1, Width: 1, Step: 1, CS: v.CS(), IP: v.IP()}}
		p.After.R[cpu.DI]++
		p.After.IP++
		return p
	case 0x53f6, 0x53fa:
		f, ok := readBitmap51Frame(v)
		if !ok || f.planar || v.ES() != 0xa000 {
			return nil
		}
		count, size, done := f.width/2, uint8(2), uint16(0)
		if v.IP() == 0x53fa {
			done = count * 2
			count = f.width % 2
			size = 1
		}
		at := uint32(f.y)*320 + uint32(f.s.left+done)
		if v.CX() != count || v.SI() != f.offset+done || uint32(v.DI()) != at || v.AX() != 320 || v.DX() != f.y*64 {
			return nil
		}
		p.Instructions, p.Cycles, p.LastIP = 1, 3+uint64(count)*5, v.IP()
		if v.DOSBoxCost() {
			p.Cycles = 1 + uint64(count)
		}
		if count > 0 {
			data := make([]byte, int(count)*int(size))
			copyView(v, f.source+uint32(done), data)
			p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + at, Data: data, Count: count, Width: size, Step: 1, CS: v.CS(), IP: v.IP()}}
		}
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.SI] += count * uint16(size)
		p.After.R[cpu.DI] += count * uint16(size)
		p.After.IP += 2
		return p
	}
	return nil
}
func planBitmap51Advance(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x53fc {
		return nil
	}
	f, ok := readBitmap51Frame(v)
	if !ok || v.ES() != 0xa000 {
		return nil
	}
	if v.DI() != f.y*320+f.s.left+f.width {
		return nil
	}
	if f.planar {
		done := f.width - 1
		color, last, mask := bitmap51Color(v, f, done)
		if v.SI() != f.offset+done/8 || v.AX() != uint16(color) || v.BX() != 8*f.stride || byte(v.CX()) != last || v.DX() != f.y*64&0xff00|uint16(mask) || v.Read16(0x48a10+0x11fd) != 0 || v.Read16(0x48a10+0x11f9) != f.stride-done/8 || v.Flags()&cpu.CF == 0 {
			return nil
		}
	} else if v.SI() != f.offset+f.width || v.AX() != 320 || v.CX() != 0 || v.DX() != f.y*64 || v.Flags()&cpu.CF != 0 {
		return nil
	}
	p := imageBasePlan(v)
	bitmap51Advance(v, p, f, v.Flags()&cpu.CF)
	if p.Instructions > budget {
		return nil
	}
	return p
}
func bitmap51(v CPUView, budget uint64) *machine.HostWorkPlan {
	if p := planBitmap51Chunky(v, budget); p != nil {
		return p
	}
	if p := planBitmap51PlanarPrefix(v, budget); p != nil {
		return p
	}
	if p := planBitmap51PlanarGroup(v, budget); p != nil {
		return p
	}
	if p := planBitmap51Span(v, budget); p != nil {
		return p
	}
	return planBitmap51Advance(v, budget)
}
func bitmap51FixedRaster(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && (w.IP == 0x53d8 || w.IP == 0x53f6 || w.IP == 0x53fa) {
			return true
		}
	}
	for _, s := range transfers {
		if s.Count > 0 && s.CS == 0x48a1 && (s.IP == 0x53d8 || s.IP == 0x53f6 || s.IP == 0x53fa) {
			return true
		}
	}
	return false
}
