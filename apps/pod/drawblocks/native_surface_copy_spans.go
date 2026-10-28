package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// Read-only projection to the typed row geometry already proven for 3F5E.
type surfaceRowProjection struct{ CPUView }

func (v surfaceRowProjection) IP() uint16 { return 0x415e }
func (v surfaceRowProjection) SI() uint16 { return v.AX() }
func (v surfaceRowProjection) DI() uint16 { return v.BX() }

func planSurfaceCopySpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || (v.IP() != 0x4169 && v.IP() != 0x4177) {
		return nil
	}
	modelView, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	model, known := modelView.CPUModel()
	if !known || model != cpu.Model80386 {
		return nil
	}
	// This validates the full source fingerprint, call frame, both descriptor
	// chains, root bounds, remaining rows and the actual AX/BX row origin.
	row, ok := surfaceCopyRowGeometry(surfaceRowProjection{v})
	if !ok {
		return nil
	}
	width := row.width
	back := v.Flags()&cpu.DF != 0
	count, unit := width/2, uint16(2)
	source, dest := v.AX(), v.BX()
	if v.IP() == 0x4169 {
		if v.Flags()&cpu.CF != width%2 {
			return nil
		}
	} else {
		delta := uint16(width / 2 * 2)
		if back {
			source -= delta
			dest -= delta
			source++
			dest++
		} else {
			source += delta
			dest += delta
		}
		mask := uint16(cpu.CF | cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
		if v.Flags()&mask != copyAddFlags(v.Flags(), source-map[bool]uint16{false: 0, true: 1}[back], map[bool]uint16{false: 0, true: 1}[back])&mask {
			return nil
		}
		count, unit = width%2, 1
	}
	if v.CX() != count || v.SI() != source || v.DI() != dest {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, After: boundedMemoryRegisters(v), LastCS: 0x48a1, LastIP: v.IP(), Cycles: 3 + 5*uint64(count)}
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	p.After.IP += 2
	p.After.R[cpu.CX] = 0
	if count > 0 {
		data := make([]byte, int(count)*int(unit))
		shadow := map[uint32]byte{}
		read := func(at uint32) byte {
			if b, ok := shadow[at]; ok {
				return b
			}
			return v.Read8(at)
		}
		si, di := source, dest
		for i := uint16(0); i < count; i++ {
			for c := uint16(0); c < unit; c++ {
				if uint32(si)+uint32(c) >= 64000 || uint32(di)+uint32(c) >= 64000 {
					return nil
				}
				data[int(i)*int(unit)+int(c)] = read(0xa0000 + uint32(si) + uint32(c))
			}
			for c := uint16(0); c < unit; c++ {
				shadow[0xa0000+uint32(di)+uint32(c)] = data[int(i)*int(unit)+int(c)]
			}
			if back {
				si -= unit
				di -= unit
			} else {
				si += unit
				di += unit
			}
		}
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + uint32(dest), Count: count, Width: uint8(unit), Backward: back, Data: data, Step: 1, CS: 0x48a1, IP: v.IP()}}
		p.After.R[cpu.SI], p.After.R[cpu.DI] = si, di
	}
	return p
}

func surfaceCopyFixedRaster(w []machine.HostWrite, ts []machine.HostTransfer) bool {
	for _, x := range w {
		if x.CS == 0x48a1 && (x.IP == 0x4169 || x.IP == 0x4177) {
			return true
		}
	}
	for _, x := range ts {
		if x.CS == 0x48a1 && (x.IP == 0x4169 || x.IP == 0x4177) {
			return true
		}
	}
	return false
}
