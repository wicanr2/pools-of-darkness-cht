// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planScale5615Prepare(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x579c {
		return nil
	}
	f, ok := readScale5615Frame(v)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	charge := func(n, c, iterations uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + iterations
		} else {
			p.Cycles += c
		}
	}
	push := func(at uint32, value uint16, step uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: at, Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: at + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
	}
	push(f.stack-2, v.CS(), 1, 0x579c)
	p.After.Seg[cpu.ES] = v.CS()
	if f.y < int32(f.s.top) || f.y > int32(f.s.bottom) {
		if f.y < int32(f.s.top) {
			charge(7, 31, 0)
		} else {
			charge(9, 44, 0)
		}
		count := f.samples * f.factor * 2
		step := p.Instructions + 5
		p.Transfers = []machine.HostTransfer{{Address: 0x48a10 + 0x1252, Value: (f.color & 255) * 257, Count: count, Width: 2, Step: step, CS: 0x48a1, IP: 0x5879}}
		charge(8, 31+5*uint64(count), uint64(count))
		p.After.R[cpu.AX] = (f.color & 255) * 257
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.SI] = uint16(f.y)
		p.After.R[cpu.DI] = 0x1252 + count*2
		p.After.Flags = nativeCompareFlags(v.Flags(), count, 0)
		p.After.IP = 0x581a
		p.LastIP = 0x587f
	} else {
		// Row grid and canonical 320-byte source address; data comes from the Go
		// sampling grid. This preserves source samples before overlapping output.
		charge(29, 122, 0)
		push(f.stack-2, f.destY, 9, 0x57b7)
		push(f.stack-4, 0x1252, 10, 0x57b8)
		ax := uint16(320)
		lastCarry := uint16(0)
		for sample := uint16(0); sample <= f.samples; sample++ {
			x := f.x + int32(sample)
			oldN := p.Instructions
			rn, rc, reads := uint64(0), uint64(0), uint64(0)
			if x < int32(f.s.left) {
				rn, rc = 5, 28
				ax = f.color
				lastCarry = nativeCompareFlags(v.Flags(), uint16(x), f.s.left) & cpu.CF
			} else if x > int32(f.s.right) {
				rn, rc = 7, 31
				ax = f.color
				lastCarry = nativeCompareFlags(v.Flags(), uint16(x), f.s.right) & cpu.CF
			} else {
				rn, rc, reads = 8, 61, 1
				ax = ax&0xff00 | uint16(v.Read8(0xa0000+uint32(f.y*320+x)))
				lastCarry = nativeCompareFlags(v.Flags(), uint16(x), f.s.right) & cpu.CF
				push(f.stack-4, v.CS(), oldN+7, 0x580e)
			}
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0x48a10 + 0x1252 + uint32(sample*f.factor), Value: ax & 255, Count: f.factor, Width: 1, Step: oldN + rn + 2, CS: 0x48a1, IP: 0x5813})
			charge(rn+5, rc+20+5*uint64(f.factor), uint64(f.factor)+reads)
		}
		// Final JNS falls through, followed by POP DX.
		p.Instructions++
		if v.DOSBoxCost() {
			p.Cycles++
		} else {
			p.Cycles--
		}
		p.After.R[cpu.AX] = ax
		p.After.R[cpu.BX] = 0xffff
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.SI] = uint16(f.y*320 + f.x + int32(f.samples) + 1)
		p.After.R[cpu.DI] = 0x1252 + (f.samples+1)*f.factor
		p.After.Flags = nativeCompareFlags(v.Flags(), 0, 1)&^cpu.CF | lastCarry
		p.After.IP = 0x581a
		p.LastIP = 0x5819
	}
	if p.Instructions > budget {
		return nil
	}
	return p
}
