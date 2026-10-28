package drawblocks

import (
	"crypto/sha256"
	"math/bits"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteListCode = smallFingerprint(14, "1814a41894eea9e581384ac7c8a316353039634bf747a66ceb5ec19aff3361af")
var paletteTargetCode = smallFingerprint(14, "343202170a7507bd27ba58e074710568546eac5f1fc99691bcac8d8e11a98349")
var paletteErrorCode = smallFingerprint(27, "592b32248b4af06920a2979be064570ce726975b43eae84d9591126394638a7d")

func paletteSetupInput(v CPUView, ip uint16, code smallCode) bool {
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.ES() != 0x48a1 || v.IP() != ip || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return false
	}
	var raw [27]byte
	copyView(v, 0x48a10+uint32(ip), raw[:code.size])
	return sha256.Sum256(raw[:code.size]) == code.sum
}

// planPaletteList filters the used-color bitmap into a descending index list.
func planPaletteList(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, next, store, last := uint16(0x4a83), uint16(0x4a91), uint16(0x4a8d), uint16(0x4a8f)
	if v.IP() == 0x47d6 {
		entry, next, store, last = 0x47d6, 0x47e4, 0x47e0, 0x47e2
	}

	if !paletteSetupInput(v, entry, paletteListCode) || v.BX() > 255 || v.DI() < 0x1572 || v.DI()&1 != 0 || uint32(v.DI())+2*(uint32(v.BX())+1) > 0x1772 {
		return nil
	}
	p := fadePlan(v)
	for index := int(v.BX()); index >= 0; index-- {
		used := v.Read8(0x48a10 + 0x1772 + uint32(index))
		steps, cycles := uint64(5), uint64(19)
		if used != 0 {
			steps, cycles = 7, 23
		}
		if index > 0 {
			cycles += 5
		}
		if steps > budget-p.Instructions {
			break
		}
		p.After.R[cpu.AX] = p.After.R[cpu.AX]&0xff00 | uint16(used)
		if used != 0 {
			p.After.R[cpu.AX] = uint16(index)
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0x48a10 + uint32(p.After.R[cpu.DI]), Value: uint16(index), Count: 1, Width: 2, Step: p.Instructions + 5, CS: 0x48a1, IP: store})
			p.After.R[cpu.DI] += 2
		}
		p.After.R[cpu.BX]--
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(index), 1) &^ cpu.CF
		p.Instructions += steps
		if v.DOSBoxCost() {
			cycles = steps
			if used != 0 {
				cycles++
			}
		}
		p.Cycles += cycles
		p.LastIP = last
	}
	if p.Instructions == 0 {
		return nil
	}
	if p.After.R[cpu.BX] == 0xffff {
		p.After.IP = next
	}
	return p
}

// planPaletteTarget repeats the three target channels over the color array.
func planPaletteTarget(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, next, store, last := uint16(0x4afd), uint16(0x4b0b), uint16(0x4b00), uint16(0x4b09)
	if v.IP() == 0x4867 {
		entry, next, store, last = 0x4867, 0x4875, 0x486a, 0x4873
	}

	if !paletteSetupInput(v, entry, paletteTargetCode) || v.CX() == 0 || v.CX() > 256 || v.DI() < 0x1b72 || (v.DI()-0x1b72)%3 != 0 || uint32(v.DI())+3*uint32(v.CX()) > 0x1e72 || budget < 7 {
		return nil
	}
	count := uint64(v.CX())
	if count > budget/7 {
		count = budget / 7
	}
	rgb := [3]byte{v.Read8(0x48a10 + 0x1179), v.Read8(0x48a10 + 0x117b), v.Read8(0x48a10 + 0x117d)}
	p := fadePlan(v)
	for x := uint64(0); x < count; x++ {
		for channel, value := range rgb {
			p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + uint32(v.DI()) + uint32(x)*3 + uint32(channel), Value: value, Step: x*7 + uint64(channel*2+2), CS: 0x48a1, IP: store + uint16(channel*4)})
		}
	}
	p.Instructions = count * 7
	p.Cycles = count * 29
	if v.DOSBoxCost() {
		p.Cycles = count * 10
	}
	p.After.R[cpu.AX] = v.AX()&0xff00 | uint16(rgb[2])
	p.After.R[cpu.CX] -= uint16(count)
	p.After.R[cpu.DI] += uint16(count) * 3
	p.LastIP = last
	if p.After.R[cpu.CX] == 0 {
		p.After.IP = next
	}
	return p
}

// planPaletteError initializes the interpolation accumulator to half its span.
func planPaletteError(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, next, store, last := uint16(0x4b3a), uint16(0x4b55), uint16(0x4b3e), uint16(0x4b53)
	if v.IP() == 0x48b3 {
		entry, next, store, last = 0x48b3, 0x48ce, 0x48b7, 0x48cc
	}

	if !paletteSetupInput(v, entry, paletteErrorCode) || budget < 9 || byte(v.BX()) > 63 {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count == 0 || count > 256 {
		return nil
	}
	n := count * 3
	maximum := byte(v.BX())
	p := fadePlan(v)
	p.Instructions, p.Cycles = 9, 35+5*uint64(n)
	if v.DOSBoxCost() {
		p.Cycles = 9 + uint64(n)
	}
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1173, Value: maximum, Step: 1, CS: 0x48a1, IP: entry}, {Address: 0x48a10 + 0x1174, Value: maximum, Step: 2, CS: 0x48a1, IP: store}}
	p.Transfers = []machine.HostTransfer{{Address: 0x48a10 + 0x1e72, Value: uint16(maximum / 2), Count: n, Width: 1, Step: 9, CS: 0x48a1, IP: last}}
	p.After.R[cpu.AX] = v.AX()&0xff00 | uint16(maximum/2)
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] = 0x1e72 + n
	p.After.Flags = v.Flags() &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if ((count*2)&15)+(count&15) > 15 {
		p.After.Flags |= cpu.AF
	}
	if bits.OnesCount8(byte(n))%2 == 0 {
		p.After.Flags |= cpu.PF
	}
	p.After.IP, p.LastIP = next, last
	return p
}
