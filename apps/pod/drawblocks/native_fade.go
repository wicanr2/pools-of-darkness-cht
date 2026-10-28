package drawblocks

import (
	"crypto/sha256"
	"math/bits"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var fadeDifferenceCode = smallFingerprint(33, "426b9c7bab4ee58a83495f6ee251ade3d04d01ce17e57c163cfdcd18c4b00f86")
var fadeBlendCode = smallFingerprint(33, "391005d80b86b978189cde8e392fbfbca279480fda3f7f3a175cb9c14ed2f2ae")

type fadeRGBLayout struct {
	differenceEntry, differenceNext, blendEntry, blendNext, current, target uint16
	differenceCode, blendCode                                               smallCode
}

var fadeOutRGBLayout = fadeRGBLayout{0x4b19, 0x4b3a, 0x4b6c, 0x4b8d, 0x1872, 0x1b72, fadeDifferenceCode, fadeBlendCode}
var fadeInRGBLayout = fadeRGBLayout{0x4892, 0x48b3, 0x48e5, 0x4906, 0x1b72, 0x1872,
	smallFingerprint(33, "bc07e0842a1185593545ee7c72bc73c057e40a608eb467e3c1ff421c167922c9"),
	smallFingerprint(33, "edd016c91a19d12007eea94e49dc216286ccf3009c9b52ff4022adcf83a0bbbc")}

func fadeInput(v CPUView, ip uint16, fingerprint smallCode) bool {
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.IP() != ip || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 || v.CX() == 0 || uint32(v.SI())+uint32(v.CX()) > 768 {
		return false
	}
	var code [33]byte
	copyView(v, 0x48a10+uint32(ip), code[:])
	return sha256.Sum256(code[:]) == fingerprint.sum
}

func fadePlan(v CPUView) *machine.HostWorkPlan {
	return &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}, LastCS: 0x48a1}
}

// fadeAdvanceFlags is compatibility metadata for the element cursor. Its
// carry comes from the preceding color calculation; other flags follow SI+1.
func fadeAdvanceFlags(before, cursor uint16, carry bool) uint16 {
	value := cursor + 1
	f := before &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if carry {
		f |= cpu.CF
	}
	if cursor&15 == 15 {
		f |= cpu.AF
	}
	if cursor == 0x7fff {
		f |= cpu.OF
	}
	if value == 0 {
		f |= cpu.ZF
	}
	if value&0x8000 != 0 {
		f |= cpu.SF
	}
	if bits.OnesCount8(byte(value))%2 == 0 {
		f |= cpu.PF
	}
	return f
}

// planFadeDifference computes signed direction, distance and maximum distance
// over RGB channels directly. Hardware palette reads remain outside this loop.
func planFadeDifference(v CPUView, budget uint64) *machine.HostWorkPlan {
	return planFadeDifferenceLayout(v, budget, fadeOutRGBLayout)
}
func planFadeDifferenceLayout(v CPUView, budget uint64, layout fadeRGBLayout) *machine.HostWorkPlan {
	if !fadeInput(v, layout.differenceEntry, layout.differenceCode) || byte(v.BX()) > 63 {
		return nil
	}
	p := fadePlan(v)
	maximum := byte(v.BX())
	for x := uint16(0); x < v.CX(); x++ {
		index := uint32(v.SI() + x)
		from, to := v.Read8(0x48a10+uint32(layout.current)+index), v.Read8(0x48a10+uint32(layout.target)+index)
		if from > 63 || to > 63 {
			break
		}
		delta, direction := int(from)-int(to), byte(255)
		steps, cycles, writeStep := uint64(11), uint64(43), uint64(5)
		if delta < 0 {
			delta, direction = -delta, 1
			steps += 2
			cycles--
			writeStep += 2
		}
		belowMax := byte(delta) < maximum
		if belowMax {
			steps--
			cycles += 3
		}
		if steps > budget-p.Instructions {
			break
		}
		p.Writes = append(p.Writes,
			machine.HostWrite{Address: 0x48a10 + 0x2472 + index, Value: direction, Step: p.Instructions + writeStep, CS: 0x48a1, IP: layout.differenceEntry + 0x10},
			machine.HostWrite{Address: 0x48a10 + 0x2172 + index, Value: byte(delta), Step: p.Instructions + writeStep + 1, CS: 0x48a1, IP: layout.differenceEntry + 0x14})
		if !belowMax {
			maximum = byte(delta)
		}
		p.After.R[cpu.AX] = uint16(direction)<<8 | uint16(delta)
		p.After.R[cpu.BX] = v.BX()&0xff00 | uint16(maximum)
		p.After.R[cpu.SI]++
		p.After.R[cpu.CX]--
		p.After.Flags = fadeAdvanceFlags(v.Flags(), uint16(index), belowMax)
		p.Instructions += steps
		if v.DOSBoxCost() {
			cycles = steps
		}
		p.Cycles += cycles
		p.LastIP = layout.differenceEntry + 0x1f
	}
	if p.Instructions == 0 {
		return nil
	}
	if p.After.R[cpu.CX] == 0 {
		p.After.IP = layout.differenceNext
	}
	return p
}

// planFadeBlend advances each RGB channel with its accumulated interpolation
// error. It computes colors directly, without executing instruction effects.
func planFadeBlend(v CPUView, budget uint64) *machine.HostWorkPlan {
	return planFadeBlendLayout(v, budget, fadeOutRGBLayout)
}
func planFadeBlendLayout(v CPUView, budget uint64, layout fadeRGBLayout) *machine.HostWorkPlan {
	if !fadeInput(v, layout.blendEntry, layout.blendCode) {
		return nil
	}
	threshold := v.Read8(0x48a10 + 0x1173)
	if threshold > 63 {
		return nil
	}
	p := fadePlan(v)
	for x := uint16(0); x < v.CX(); x++ {
		index := uint32(v.SI() + x)
		err, distance := v.Read8(0x48a10+0x1e72+index), v.Read8(0x48a10+0x2172+index)
		if err > 63 || distance > 63 {
			break
		}
		value := err + distance
		advance := value > threshold
		steps, cycles, writeStep := uint64(7), uint64(35), uint64(5)
		color, direction := byte(0), byte(0)
		carry := value < threshold
		if advance {
			steps, cycles, writeStep = 10, 48, 8
			color = v.Read8(0x48a10 + uint32(layout.current) + index)
			direction = v.Read8(0x48a10 + 0x2472 + index)
			if color > 63 || (direction != 1 && direction != 255) || (direction == 1 && color == 63) || (direction == 255 && color == 0) {
				break
			}
			value -= threshold
			carry = uint16(color)+uint16(direction) > 255
		}
		if steps > budget-p.Instructions {
			break
		}
		if advance {
			p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + uint32(layout.current) + index, Value: color + direction, Step: p.Instructions + 7, CS: 0x48a1, IP: layout.blendEntry + 0x16})
			p.After.R[cpu.BX] = v.BX()&0xff00 | uint16(direction)
		}
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + 0x1e72 + index, Value: value, Step: p.Instructions + writeStep, CS: 0x48a1, IP: layout.blendEntry + 0x1a})
		p.After.R[cpu.AX] = v.AX()&0xff00 | uint16(value)
		p.After.R[cpu.SI]++
		p.After.R[cpu.CX]--
		p.After.Flags = fadeAdvanceFlags(v.Flags(), uint16(index), carry)
		p.Instructions += steps
		if v.DOSBoxCost() {
			cycles = steps
		}
		p.Cycles += cycles
		p.LastIP = layout.blendEntry + 0x1f
	}
	if p.Instructions == 0 {
		return nil
	}
	if p.After.R[cpu.CX] == 0 {
		p.After.IP = layout.blendNext
	}
	return p
}
