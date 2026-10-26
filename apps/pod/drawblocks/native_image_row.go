package drawblocks

import (
	"crypto/sha256"
	"math/bits"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var imageRowCode = smallFingerprint(78, "1c4ab9672accc6ec8dcd86e8703b26cf877efa2ca54e8afa5268d2c5fed7a62b")

// Parse a row index and calculate its clipped surface position.
// Spec025/#49: ABI effects are recorded separately from the geometry.
func planImageRow(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x6850 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 || budget < 5 || v.BP() > 65524 || v.SI() == 65535 {
		return nil
	}
	var code [78]byte
	copyView(v, 0x48a10+0x6850, code[:])
	if sha256.Sum256(code[:]) != imageRowCode.sum {
		return nil
	}
	src := uint32(v.DS())*16 + uint32(v.SI())
	arg := uint32(v.SS())*16 + uint32(v.BP()) + 10
	if src < 0x500 || src >= 0xa0000 || arg < 0x500 || arg+2 > 0xa0000 {
		return nil
	}
	for _, a := range []uint32{0x48a10 + 0x1250, 0x48a10 + 0x120c} {
		if src >= a && src < a+2 || arg < a+2 && arg+2 > a {
			return nil
		}
	}
	raw := v.Read8(src)
	p := &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI() + 1, v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, Flags: v.Flags()}, LastCS: 0x48a1}
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1250, Step: 1, CS: 0x48a1, IP: 0x6850}, {Address: 0x48a10 + 0x1251, Step: 1, CS: 0x48a1, IP: 0x6850}}
	p.After.R[cpu.AX] = uint16(raw)
	if raw == 255 {
		p.Instructions, p.Cycles, p.LastIP, p.After.IP = 5, 25, 0x685c, 0x682d
		p.After.Flags = nativeCompareFlags(v.Flags(), 255, 255)
	} else {
		y := int32(raw) + int32(int16(v.Read16(arg)))
		top, bottom := int32(int16(v.Read16(0x48a10+0xec3))), int32(int16(v.Read16(0x48a10+0xec7)))
		if y < -32768 || y > 32767 || top > bottom {
			return nil
		}
		p.After.R[cpu.AX] = uint16(y)
		switch {
		case y > bottom:
			p.Instructions, p.Cycles, p.LastIP, p.After.IP = 8, 40, 0x6866, 0x682d
			p.After.Flags = nativeCompareFlags(v.Flags(), uint16(y), uint16(bottom))
		case y < top:
			p.Instructions, p.Cycles, p.LastIP, p.After.IP = 10, 49, 0x686d, 0x6832
			p.After.Flags = nativeCompareFlags(v.Flags(), uint16(y), uint16(top))
		default:
			stride := uint32(v.Read16(0x48a10 + 0xedd))
			x := int32(v.Read16(0x48a10+0xec1)) - int32(v.Read16(0x48a10+0xed3))
			dy := y - int32(v.Read16(0x48a10+0xed5))
			if stride == 0 || stride > 320 || x < 0 || uint32(x) >= stride || dy < 0 || dy > 199 {
				return nil
			}
			row := uint32(x) + uint32(dy)*stride
			if row >= 64000 {
				return nil
			}
			p.Instructions, p.Cycles, p.LastIP, p.After.IP = 21, 106, 0x6899, 0x689e
			p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.DI] = uint16(uint32(dy)*stride), 0, uint16(row)
			p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x), uint16(uint32(dy)*stride))
			if stride == 320 {
				p.Instructions, p.Cycles = 23, 96
				p.After.R[cpu.AX], p.After.R[cpu.DX] = 320, uint16(dy*64)
				p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x+dy*256), uint16(dy*64))
			}
			p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + 0x120c, Value: byte(row), Step: p.Instructions, CS: 0x48a1, IP: 0x6899}, machine.HostWrite{Address: 0x48a10 + 0x120d, Value: byte(row >> 8), Step: p.Instructions, CS: 0x48a1, IP: 0x6899})
		}
	}
	if p.Instructions > budget {
		return nil
	}
	if v.DOSBoxCost() {
		p.Cycles = p.Instructions + 1
	}
	return p
}

func imageRowAddFlags(flags, a, b uint16) uint16 {
	n := uint32(a) + uint32(b)
	r := uint16(n)
	f := flags &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if n > 65535 {
		f |= cpu.CF
	}
	if bits.OnesCount8(byte(r))%2 == 0 {
		f |= cpu.PF
	}
	if (a^b^r)&0x10 != 0 {
		f |= cpu.AF
	}
	if r == 0 {
		f |= cpu.ZF
	}
	if r&0x8000 != 0 {
		f |= cpu.SF
	}
	if (^(a^b)&(a^r))&0x8000 != 0 {
		f |= cpu.OF
	}
	return f
}
