// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

type bitmap51FrameView struct {
	CPUView
	sp, bp uint16
}

func (v bitmap51FrameView) SP() uint16 { return v.sp }
func (v bitmap51FrameView) BP() uint16 { return v.bp }
func planBitmap51PlanarPrefix(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x5373 || budget < 22 {
		return nil
	}
	f, ok := readBitmap51Frame(v)
	if !ok || !f.planar || v.SI() != f.offset {
		return nil
	}
	p := imageBasePlan(v)
	p.Instructions, p.Cycles, p.LastIP = 22, 109, 0x53c3
	if v.DOSBoxCost() {
		p.Cycles = 22
	}
	imageWord(p, 0x11f9, f.stride, 17, 0x53b1)
	imageWord(p, 0x11fd, f.width, 19, 0x53b9)
	p.Writes = append(p.Writes, machine.HostWrite{Address: f.stack - 2, Value: byte(v.BP()), Step: 20, CS: 0x48a1, IP: 0x53bd}, machine.HostWrite{Address: f.stack - 1, Value: byte(v.BP() >> 8), Step: 20, CS: 0x48a1, IP: 0x53bd})
	p.After.R[cpu.AX] = f.width
	p.After.R[cpu.DX] = f.y*64&0xff00 | 0x80
	p.After.R[cpu.DI] = f.y*320 + f.s.left
	p.After.R[cpu.BP] = f.stride
	p.After.R[cpu.SP] -= 2
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = 0x53c5
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	return p
}
func bitmap51PlanarState(v CPUView) (f bitmap51Frame, stack uint32, remaining, done uint16, ok bool) {
	stack, ok = paletteSourceStack(v, 0, 20)
	if !ok || v.SP() > 65533 || v.ES() != 0xa000 {
		return f, 0, 0, 0, false
	}
	savedBP := v.Read16(stack)
	if savedBP != v.SP()+8 {
		return f, 0, 0, 0, false
	}
	f, ok = readBitmap51Frame(bitmap51FrameView{v, v.SP() + 2, savedBP})
	if !ok || !f.planar || v.BP() != f.stride {
		return f, 0, 0, 0, false
	}
	remaining = v.Read16(0x48a10 + 0x11fd)
	if remaining == 0 || remaining > f.width {
		return f, 0, 0, 0, false
	}
	done = f.width - remaining
	if v.Read16(0x48a10+0x11f9) != f.stride-done/8 || v.SI() != f.offset+done/8 || v.DI() != f.y*320+f.s.left+done || v.DX() != f.y*64&0xff00|uint16(0x80>>(done&7)) {
		return f, 0, 0, 0, false
	}
	return f, stack, remaining, done, true
}
func bitmap51Color(v CPUView, f bitmap51Frame, done uint16) (color, last, mask byte) {
	mask = 0x80 >> (done & 7)
	for plane := uint16(0); plane < 8; plane++ {
		value := v.Read8(f.source + uint32(plane*f.stride+done/8))
		if value&mask != 0 {
			color |= 1 << plane
		}
		if plane == 7 {
			last = value & mask
		}
	}
	return color, last, mask
}
func planBitmap51PlanarGroup(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x53c5 {
		return nil
	}
	f, stack, remaining, done, ok := bitmap51PlanarState(v)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	charge := func(n, c uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + 1
		} else {
			p.Cycles += c
		}
	}
	for {
		color, last, mask := bitmap51Color(v, f, done)
		ones := uint64(bits.OnesCount8(color))
		oldN := p.Instructions
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(f.y*320+f.s.left+done), Value: uint16(color), Count: 1, Width: 1, Step: oldN + 51 + ones, CS: 0x48a1, IP: 0x53d8})
		imageWord(p, 0x11fd, remaining-1, oldN+52+ones, 0x53d9)
		p.After.R[cpu.AX], p.After.R[cpu.BX], p.After.R[cpu.CX] = uint16(color), 8*f.stride, v.CX()&0xff00|uint16(last)
		p.After.R[cpu.DI] = f.y*320 + f.s.left + done + 1
		p.After.R[cpu.SI] = f.offset + done/8
		if remaining == 1 {
			charge(55+ones, 239-3*ones)
			p.After.R[cpu.DX] = f.y*64&0xff00 | uint16(mask)
			p.After.R[cpu.BP] = v.Read16(stack)
			p.After.R[cpu.SP] += 2
			p.After.IP = 0x53fc
			p.LastIP = 0x53ed
			p.After.Flags = nativeCompareFlags(v.Flags(), 1, 1)&^cpu.CF | cpu.CF
			if p.Instructions > budget {
				return nil
			}
			return p
		}
		if mask == 1 {
			oldBytes := f.stride - done/8
			imageWord(p, 0x11f9, oldBytes-1, oldN+57+ones, 0x53e5)
			charge(59+ones, 245-3*ones)
			p.After.R[cpu.DX] = f.y*64&0xff00 | 0x80
			p.After.R[cpu.SI]++
			p.After.IP = 0x53c5
			p.LastIP = 0x53c3
			p.After.Flags = nativeCompareFlags(v.Flags(), oldBytes, 1)&^cpu.CF | cpu.CF
			if p.Instructions > budget {
				return nil
			}
			return p
		}
		charge(55+ones, 232-3*ones)
		remaining--
		done++
	}
}
