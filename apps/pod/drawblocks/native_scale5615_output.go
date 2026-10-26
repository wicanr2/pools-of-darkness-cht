// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planScale5615Output(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x581a {
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
	push := func(value uint16, step uint64) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: f.stack - 2, Value: byte(value), Step: step, CS: 0x48a1, IP: 0x5822}, machine.HostWrite{Address: f.stack - 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: 0x5822})
	}
	charge(2, 31, 0)
	p.After.Seg[cpu.ES] = 0xa000
	width, words, tail := f.width, f.width/2, f.width%2
	data := make([]byte, width)
	copyView(v, 0x48a10+0x1252, data)
	for repeat := uint16(0); repeat < f.factor; repeat++ {
		row := f.destY + repeat
		oldN := p.Instructions
		push(row, oldN+1)
		at := uint32(row)*320 + uint32(f.d.left)
		if words > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Data: append([]byte(nil), data[:words*2]...), Count: words, Width: 2, Step: oldN + 18, CS: 0x48a1, IP: 0x5856})
		}
		if tail > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(words)*2, Data: []byte{data[width-1]}, Count: 1, Width: 1, Step: oldN + 20, CS: 0x48a1, IP: 0x585a})
		}
		charge(20, 68+5*uint64(words+tail), uint64(words+tail))
		p.After.R[cpu.AX] = 320
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.DX] = row + 1
		p.After.R[cpu.SI] = 0x1252 + width
		p.After.R[cpu.DI] = uint16(at) + width
		flags := nativeCompareFlags(v.Flags(), row+1, f.d.bottom)
		if row == f.d.bottom {
			// RETF: base2 + cycFar18 + two stack reads4, after four POPs16.
			charge(9, 59, 0)
			p.After.R[cpu.BX] = f.factor - repeat
			p.After.R[cpu.DI] = v.Read16(f.stack)
			p.After.R[cpu.SI] = v.Read16(f.stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(f.stack + 4)
			p.After.R[cpu.BP] = v.Read16(f.stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(f.stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(f.stack + 10)
			p.After.Flags = flags
			p.LastIP = 0x5885
			if p.Instructions > budget {
				return nil
			}
			return p
		}
		if repeat+1 < f.factor {
			charge(6, 23, 0)
		} else {
			charge(6, 18, 0)
		}
		p.After.R[cpu.BX] = f.factor - repeat - 1
		p.After.Flags = flags
	}
	imageWord(p, 0x1171, uint16(f.y+1), p.Instructions+1, 0x5867)
	charge(2, 13, 0)
	p.After.Flags = copyAddFlags(v.Flags(), uint16(f.y), 1)&^cpu.CF | p.After.Flags&cpu.CF
	p.After.IP = 0x579c
	p.LastIP = 0x586b
	if p.Instructions > budget {
		return nil
	}
	return p
}
