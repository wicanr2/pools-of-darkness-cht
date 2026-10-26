// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func bitmap51Advance(v CPUView, p *machine.HostWorkPlan, f bitmap51Frame, carry uint16) {
	oldN := p.Instructions
	terminal := f.rows == 1
	n, c := uint64(4), uint64(23)
	if terminal {
		n, c = 9, 64
	}
	imageWord(p, 0xec3, f.y+1, oldN+1, 0x53fc)
	imageWord(p, 0x11ff, f.rows-1, oldN+2, 0x5401)
	p.Instructions += n
	if v.DOSBoxCost() {
		p.Cycles += n
	} else {
		p.Cycles += c
	}
	p.After.Flags = nativeCompareFlags(v.Flags(), f.rows, 1)&^cpu.CF | carry
	p.After.IP = 0x5313
	p.LastIP = 0x5408
	if terminal {
		p.After.R[cpu.AX] = v.Read16(0x48a10 + 0x1214)
		p.After.R[cpu.DI] = v.Read16(f.stack)
		p.After.R[cpu.SI] = v.Read16(f.stack + 2)
		p.After.Seg[cpu.DS] = v.Read16(f.stack + 4)
		p.After.R[cpu.BP] = v.Read16(f.stack + 6)
		p.After.R[cpu.SP] += 12
		p.After.IP = v.Read16(f.stack + 8)
		p.After.Seg[cpu.CS] = v.Read16(f.stack + 10)
		p.LastIP = 0x5413
	}
}
func planBitmap51Chunky(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x5373 {
		return nil
	}
	f, ok := readBitmap51Frame(v)
	if !ok || f.planar || v.SI() != f.offset {
		return nil
	}
	words, tail := f.width/2, f.width%2
	at := uint32(f.y)*320 + uint32(f.s.left)
	p := imageBasePlan(v)
	p.Instructions, p.Cycles = 20, 106+5*uint64(words+tail)
	if v.DOSBoxCost() {
		p.Cycles = 20 + uint64(words+tail)
	}
	if words > 0 {
		data := make([]byte, int(words)*2)
		copyView(v, f.source, data)
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Data: data, Count: words, Width: 2, Step: 18, CS: 0x48a1, IP: 0x53f6})
	}
	if tail > 0 {
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at + uint32(words)*2, Data: []byte{v.Read8(f.source + uint32(words)*2)}, Count: 1, Width: 1, Step: 20, CS: 0x48a1, IP: 0x53fa})
	}
	p.After.R[cpu.AX], p.After.R[cpu.CX], p.After.R[cpu.DX] = 320, 0, f.y*64
	p.After.R[cpu.SI] += f.width
	p.After.R[cpu.DI] = uint16(at) + f.width
	p.After.Seg[cpu.ES] = 0xa000
	bitmap51Advance(v, p, f, 0)
	if p.Instructions > budget {
		return nil
	}
	return p
}
