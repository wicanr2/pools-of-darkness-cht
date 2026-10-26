package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var fadeColorCode = smallFingerprint(44, "548776df6fec7df3ebb2f57ade4412c8704fee225a928411593571d6c7c3b525")

// planFadeColor prepares the next palette-index/RGB job from the color array.
// Hardware palette output remains at the existing 2AA8 boundary.
func planFadeColor(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || (v.IP() != 0x4b9a && v.IP() != 0x4bc1) || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [44]byte
	copyView(v, 0x48a10+0x4b9a, code[:])
	if sha256.Sum256(code[:]) != fadeColorCode.sum {
		return nil
	}
	used := v.Read16(0x48a10 + 0x116f)
	if used == 0 || used > 256 || v.DI()&1 != 0 || v.DI() > used*2 {
		return nil
	}
	sp := uint32(v.SP())
	prefix := uint64(0)
	if v.IP() == 0x4bc1 {
		sp += 8
		prefix = 2
	}
	if sp > 65535 {
		return nil
	}
	p := fadePlan(v)
	p.After.R[cpu.SP] = uint16(sp)
	p.After.R[cpu.DI] -= 2
	if v.DI() == 0 {
		p.Instructions, p.Cycles = 2+prefix, 9
		if prefix != 0 {
			p.Cycles += 9
		}
		if v.DOSBoxCost() {
			p.Cycles = p.Instructions
		}
		if p.Instructions > budget {
			return nil
		}
		p.After.IP, p.LastIP = 0x4b55, 0x4b9d
		p.After.Flags = nativeCompareFlags(v.Flags(), v.DI(), 2)
		return p
	}
	if v.SI() != 0x1872+(used-v.DI()/2)*3 || sp < 12 {
		return nil
	}
	stack := uint32(v.SS())*16 + sp
	if stack-12 < 0x500 || stack > 0xa0000 || (stack-12 < 0x48a10+0x7580 && stack > 0x48a10) {
		return nil
	}
	color := v.Read16(0x48a10 + 0x1572 + uint32(p.After.R[cpu.DI]))
	rgb := [3]byte{v.Read8(0x48a10 + uint32(v.SI())), v.Read8(0x48a10 + uint32(v.SI()) + 1), v.Read8(0x48a10 + uint32(v.SI()) + 2)}
	if color > 255 || rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return nil
	}
	// Six stack words contribute two normal-model cost units each.
	p.Instructions, p.Cycles = 15+prefix, 80
	if prefix != 0 {
		p.Cycles += 9
	}
	if v.DOSBoxCost() {
		p.Cycles = p.Instructions + 3
	}
	if p.Instructions > budget {
		return nil
	}
	word := func(address uint32, value uint16, step uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: address, Value: byte(value), Step: prefix + step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: address + 1, Value: byte(value >> 8), Step: prefix + step, CS: 0x48a1, IP: ip})
	}
	for i, channel := range rgb {
		word(0x48a10+0x117f+uint32(i)*2, uint16(channel), uint64(5+i*2), uint16(0x4ba2+i*4))
	}
	values := [...]uint16{uint16(rgb[2]), uint16(rgb[1]), uint16(rgb[0]), color, 0x48a1, 0x4bc1}
	ips := [...]uint16{0x4bad, 0x4bb1, 0x4bb5, 0x4bb9, 0x4bbd, 0x4bbe}
	for i, value := range values {
		word(stack-uint32(i+1)*2, value, uint64(10+i), ips[i])
	}
	p.After.R[cpu.AX] = uint16(rgb[2])
	p.After.R[cpu.SI] += 3
	p.After.R[cpu.SP] = uint16(sp - 12)
	p.After.Flags = v.Flags()&^uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF) | cpu.PF | cpu.ZF
	p.After.IP, p.LastIP = 0x2aa8, 0x4bbe
	return p
}
