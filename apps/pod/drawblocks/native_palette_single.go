package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

func palettePhaseAdmit(v CPUView, entry uint16, code smallCode) bool {
	if v.CS() != 0x48a1 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return false
	}
	var raw [33]byte
	copyView(v, 0x48a10+uint32(entry), raw[:code.size])
	return sha256.Sum256(raw[:code.size]) == code.sum
}
func palettePhasePlan(v CPUView) *machine.HostWorkPlan {
	return &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, After: boundedMemoryRegisters(v), LastCS: 0x48a1, LastIP: v.IP()}
}
func paletteAddColorFlags(before uint16, color, direction byte) uint16 {
	sum := uint16(color) + uint16(direction)
	value := byte(sum)
	f := before &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if sum > 255 {
		f |= cpu.CF
	}
	if (color^direction^value)&16 != 0 {
		f |= cpu.AF
	}
	if value == 0 {
		f |= cpu.ZF
	}
	if value&128 != 0 {
		f |= cpu.SF
	}
	if (^(color^direction)&(color^value))&128 != 0 {
		f |= cpu.OF
	}
	if bits.OnesCount8(value)%2 == 0 {
		f |= cpu.PF
	}
	return f
}

// One RGB channel is reconstructed from the driver's typed interpolation
// buffers. ABI flags/locations only retain the observable scheduler boundary.
func planPaletteRGBPhase(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 {
		return nil
	}
	layout := fadeOutRGBLayout
	if v.IP() == 0x48fb || v.IP() == 0x48ff {
		layout = fadeInRGBLayout
	} else if v.IP() != 0x4b82 && v.IP() != 0x4b86 {
		return nil
	}
	if !palettePhaseAdmit(v, layout.blendEntry, layout.blendCode) || v.DS() != 0x48a1 || v.CX() == 0 {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count == 0 || count > 256 || uint32(v.SI())+uint32(v.CX()) != uint32(count)*3 {
		return nil
	}
	index := uint32(v.SI())
	maximum := v.Read8(0x48a10 + 0x1173)
	err := v.Read8(0x48a10 + 0x1e72 + index)
	distance := v.Read8(0x48a10 + 0x2172 + index)
	color := v.Read8(0x48a10 + uint32(layout.current) + index)
	if maximum > 63 || err > 63 || distance > 63 || color > 63 {
		return nil
	}
	sum := err + distance
	advance := sum > maximum
	value := sum
	if advance {
		value -= maximum
	}
	if byte(v.AX()) != value {
		return nil
	}
	p := palettePhasePlan(v)
	p.Cycles = 6
	if v.DOSBoxCost() {
		p.Cycles = 1
	}
	p.After.IP = v.IP() + 4
	mask := uint16(cpu.CF | cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
	if v.IP() == layout.blendEntry+0x16 {
		direction := v.Read8(0x48a10 + 0x2472 + index)
		if !advance || byte(v.BX()) != direction || (direction != 1 && direction != 255) || (direction == 1 && color == 63) || (direction == 255 && color == 0) {
			return nil
		}
		if v.Flags()&mask != glyph5aa4Compare8Flags(v.Flags(), sum, maximum)&mask {
			return nil
		}
		p.Writes = []machine.HostWrite{{Address: 0x48a10 + uint32(layout.current) + index, Value: color + direction, Step: 1, CS: 0x48a1, IP: v.IP()}}
		p.After.Flags = paletteAddColorFlags(v.Flags(), color, direction)
	} else {
		expected := glyph5aa4Compare8Flags(v.Flags(), sum, maximum)
		if advance {
			direction := v.Read8(0x48a10 + 0x2472 + index)
			if byte(v.BX()) != direction || (direction != 1 && direction != 255) {
				return nil
			}
			previous := int(color) - 1
			if direction == 255 {
				previous = int(color) + 1
			}
			if previous < 0 || previous > 63 {
				return nil
			}
			expected = paletteAddColorFlags(v.Flags(), byte(previous), direction)
		}
		if v.Flags()&mask != expected&mask {
			return nil
		}
		p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1e72 + index, Value: value, Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	return p
}

func planPaletteSourceFlag(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.IP() != 0x4a6a && v.IP() != 0x47bd {
		return nil
	}
	entry := uint16(0x4a67)
	if v.IP() == 0x47bd {
		entry = 0x47ba
	}
	if !palettePhaseAdmit(v, entry, paletteScanCode) || v.DS() != 0xa000 || v.SI() == 0 || v.SI() > 64000 || v.CX() == 0 || v.CX() > 320 || uint32(v.SI()-1)%320+uint32(v.CX()) > 320 {
		return nil
	}
	color := v.Read8(0xa0000 + uint32(v.SI()) - 1)
	if byte(v.AX()) != color || v.BX() != uint16(color) {
		return nil
	}
	p := palettePhasePlan(v)
	p.Cycles = 7
	if v.DOSBoxCost() {
		p.Cycles = 1
	}
	p.After.IP = v.IP() + 6
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1772 + uint32(color), Value: 1, Step: 1, CS: 0x48a1, IP: v.IP()}}
	return p
}
func planPaletteListItem(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.IP() != 0x4a8d && v.IP() != 0x47e0 {
		return nil
	}
	entry := uint16(0x4a83)
	if v.IP() == 0x47e0 {
		entry = 0x47d6
	}
	if !palettePhaseAdmit(v, entry, paletteListCode) || v.DS() != 0x48a1 || v.ES() != 0x48a1 || v.BX() > 255 || v.AX() != v.BX() || v.DI() < 0x1572 || v.DI()&1 != 0 || v.DI() > 0x1770 || v.Read8(0x48a10+0x1772+uint32(v.BX())) == 0 {
		return nil
	}
	p := palettePhasePlan(v)
	p.Cycles = 7
	if v.DOSBoxCost() {
		p.Cycles = 2
	}
	p.After.IP = v.IP() + 1
	p.After.R[cpu.DI] += 2
	p.Transfers = []machine.HostTransfer{{Address: 0x48a10 + uint32(v.DI()), Value: v.BX(), Count: 1, Width: 2, Step: 1, CS: 0x48a1, IP: v.IP()}}
	return p
}
func planPaletteTargetChannel(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 {
		return nil
	}
	entry, first := uint16(0x4afd), uint16(0x4b00)
	switch v.IP() {
	case 0x486a, 0x486e, 0x4872:
		entry, first = 0x4867, 0x486a
	case 0x4b00, 0x4b04, 0x4b08:
	default:
		return nil
	}
	if !palettePhaseAdmit(v, entry, paletteTargetCode) || v.DS() != 0x48a1 || v.ES() != 0x48a1 || v.DI() < 0x1b72 {
		return nil
	}
	channel := (v.IP() - first) / 4
	count := v.Read16(0x48a10 + 0x116f)
	offset := uint32(v.DI() - 0x1b72)
	if count == 0 || count > 256 || offset%3 != uint32(channel) || offset >= uint32(count)*3 || uint32(v.CX()) != uint32(count)-offset/3 {
		return nil
	}
	value := v.Read8(0x48a10 + 0x1179 + uint32(channel)*2)
	if value > 63 || byte(v.AX()) != value {
		return nil
	}
	p := palettePhasePlan(v)
	p.Cycles = 7
	if v.DOSBoxCost() {
		p.Cycles = 2
	}
	p.After.IP = v.IP() + 1
	p.After.R[cpu.DI]++
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + uint32(v.DI()), Value: value, Step: 1, CS: 0x48a1, IP: v.IP()}}
	return p
}
func planPaletteErrorSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.IP() != 0x4b53 && v.IP() != 0x48cc {
		return nil
	}
	entry := uint16(0x4b3a)
	if v.IP() == 0x48cc {
		entry = 0x48b3
	}
	if !palettePhaseAdmit(v, entry, paletteErrorCode) || v.DS() != 0x48a1 || v.ES() != 0x48a1 {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	maximum := v.Read8(0x48a10 + 0x1173)
	if count == 0 || count > 256 || maximum > 63 || byte(v.BX()) != maximum || v.Read8(0x48a10+0x1174) != maximum || byte(v.AX()) != maximum/2 || v.CX() != count*3 || v.DI() != 0x1e72 {
		return nil
	}
	p := palettePhasePlan(v)
	p.Cycles = 3 + 5*uint64(count)*3
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)*3
	}
	p.After.IP = v.IP() + 2
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] += count * 3
	p.Transfers = []machine.HostTransfer{{Address: 0x48a10 + 0x1e72, Value: uint16(maximum / 2), Count: count * 3, Width: 1, Step: 1, CS: 0x48a1, IP: v.IP()}}
	return p
}
func planPaletteSingle(v CPUView, budget uint64) *machine.HostWorkPlan {
	if p := planPaletteRGBPhase(v, budget); p != nil {
		return p
	}
	if p := planPaletteSourceFlag(v, budget); p != nil {
		return p
	}
	if p := planPaletteListItem(v, budget); p != nil {
		return p
	}
	if p := planPaletteTargetChannel(v, budget); p != nil {
		return p
	}
	return planPaletteErrorSpan(v, budget)
}
func paletteSingleWriteIP(ip uint16) bool {
	switch ip {
	case 0x4b82, 0x4b86, 0x48fb, 0x48ff, 0x4a6a, 0x47bd, 0x4a8d, 0x47e0, 0x4b00, 0x4b04, 0x4b08, 0x486a, 0x486e, 0x4872, 0x4b53, 0x48cc:
		return true
	}
	return false
}
func paletteFixedSingleWrites(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && paletteSingleWriteIP(w.IP) {
			return true
		}
	}
	for _, s := range transfers {
		if s.CS == 0x48a1 && paletteSingleWriteIP(s.IP) {
			return true
		}
	}
	return false
}
