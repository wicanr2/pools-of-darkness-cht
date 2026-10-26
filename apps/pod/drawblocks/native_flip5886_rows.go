// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFlip5886Rows(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x59a9 {
		return nil
	}
	f, ok := readFlip5886Frame(v)
	if !ok {
		return nil
	}
	bottom := f.top
	if f.vertical {
		bottom = f.bottom
	}
	topAt, bottomAt := f.top*320+f.s.left, bottom*320+f.s.left
	if v.AX() != 320 || v.SI() != topAt || v.BX() != bottomAt || v.DI() != bottomAt || v.DX() != bottom*64 {
		return nil
	}
	p := imageBasePlan(v)
	width, words, tail := f.width, f.width/2, f.width%2
	iterations := uint64(words + tail)
	charge := func(n, c, it uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + it
		} else {
			p.Cycles += c
		}
	}
	push := func(at uint32, value uint16, step uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: at, Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: at + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
	}
	emitCopy := func(at uint32, data []byte, wordStep, byteStep uint64, wordIP, byteIP uint16) {
		if words > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: at, Data: append([]byte(nil), data[:words*2]...), Count: words, Width: 2, Step: wordStep, CS: 0x48a1, IP: wordIP})
		}
		if tail > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: at + uint32(words)*2, Data: []byte{data[width-1]}, Count: 1, Width: 1, Step: byteStep, CS: 0x48a1, IP: byteIP})
		}
	}
	push(f.stack-2, topAt, 1, 0x59a9)
	push(f.stack-4, v.CS(), 3, 0x59ad)
	saved := make([]byte, width)
	for x := uint16(0); x < width; x++ {
		sx := x
		if f.horizontal {
			sx = width - 1 - x
		}
		saved[x] = v.Read8(0xa0000 + uint32(topAt+sx))
	}
	ax := uint16(320)
	if f.horizontal {
		charge(10, 63, 0)
		for x, value := range saved {
			p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + 0x1252 + uint32(x), Value: value, Step: p.Instructions + 3, CS: 0x48a1, IP: 0x59c8})
			charge(4, 17, 1)
			ax = ax&0xff00 | uint16(value)
		}
		charge(1, 7, 0)
	} else {
		emitCopy(0x48a10+0x1252, saved, 10, 12, 0x59cf, 0x59d3)
		charge(12, 69+5*iterations, iterations)
	}
	if f.horizontal {
		charge(8, 55, 0)
		shadow := make(map[uint32]byte, int(width))
		for x := uint16(0); x < width; x++ {
			source := uint32(0xa0000) + uint32(bottomAt+width-1-x)
			value, found := shadow[source]
			if !found {
				value = v.Read8(source)
			}
			at := uint32(0xa0000) + uint32(topAt+x)
			p.Writes = append(p.Writes, machine.HostWrite{Address: at, Value: value, Step: p.Instructions + 3, CS: 0x48a1, IP: 0x59f1})
			shadow[at] = value
			charge(4, 17, 1)
			ax = ax&0xff00 | uint16(value)
		}
		charge(1, 7, 0)
	} else {
		data := make([]byte, width)
		copyView(v, 0xa0000+uint32(bottomAt), data)
		emitCopy(0xa0000+uint32(topAt), data, p.Instructions+8, p.Instructions+10, 0x59f8, 0x59fc)
		charge(10, 61+5*iterations, iterations)
	}
	push(f.stack-2, v.CS(), p.Instructions+1, 0x59fe)
	emitCopy(0xa0000+uint32(bottomAt), saved, p.Instructions+7, p.Instructions+9, 0x5a0c, 0x5a10)
	charge(9, 29+5*iterations, iterations)
	imageWord(p, 0xec7, f.bottom-1, p.Instructions+1, 0x5a12)
	imageWord(p, 0xec3, f.top+1, p.Instructions+2, 0x5a16)
	imageWord(p, 0x1171, f.remaining-1, p.Instructions+3, 0x5a1a)
	p.After.R[cpu.AX] = ax
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.SI] = 0x1252 + width
	p.After.R[cpu.DI] = bottomAt + width
	p.After.Flags = nativeCompareFlags(v.Flags(), f.remaining, 1) &^ cpu.CF
	p.After.IP = 0x5947
	p.LastIP = 0x5a20
	if f.remaining > 1 {
		charge(5, 27, 0)
	} else {
		charge(9, 65, 0)
		p.After.R[cpu.DI] = v.Read16(f.stack)
		p.After.R[cpu.SI] = v.Read16(f.stack + 2)
		p.After.Seg[cpu.DS] = v.Read16(f.stack + 4)
		p.After.R[cpu.BP] = v.Read16(f.stack + 6)
		p.After.R[cpu.SP] += 12
		p.After.IP = v.Read16(f.stack + 8)
		p.After.Seg[cpu.CS] = v.Read16(f.stack + 10)
		p.LastIP = 0x5a27
	}
	if p.Instructions > budget {
		return nil
	}
	return p
}
