package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteAdvanceCode = smallFingerprint(82, "9363734089cb9e9c5dc14386327d0a47e6a34f223e6a5cf6f53784710d006049")
var paletteFrameCode = smallFingerprint(23, "8e520c0068ea63ec8f32c89b0e63881b1949480d3e1483988d54dacd48e97bea")

func planPaletteFrame(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x4b55 || v.DS() != 0x48a1 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [23]byte
	copyView(v, 0x48a10+0x4b55, code[:])
	if sha256.Sum256(code[:]) != paletteFrameCode.sum {
		return nil
	}
	remaining := v.Read8(0x48a10 + 0x1174)
	if remaining > 63 {
		return nil
	}
	p := imageBasePlan(v)
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1174, Value: remaining - 1, Step: 1, CS: 0x48a1, IP: 0x4b55}}
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	if remaining != 0 {
		count := v.Read16(0x48a10 + 0x116f)
		if count < 1 || count > 256 {
			return nil
		}
		p.After.R[cpu.CX] = count * 3
		p.After.R[cpu.SI] = 0
		p.After.IP = 0x4b6c
		return imageTimingReturn(v, p, budget, 7, 30, 0x4b6a)
	}
	for i, b := range []byte{0x5f, 0x5e, 0x1f, 0x5d, 0xcb} {
		if v.Read8(0x48a10+0x4bc6+uint32(i)) != b {
			return nil
		}
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if v.SP() > 65524 || stack < 0x500 || stack+12 > 0xa0000 || stack < 0x48a10+0x7580 && stack+12 > 0x48a10 {
		return nil
	}
	p.After.R[cpu.DI] = v.Read16(stack)
	p.After.R[cpu.SI] = v.Read16(stack + 2)
	p.After.Seg[cpu.DS] = v.Read16(stack + 4)
	p.After.R[cpu.BP] = v.Read16(stack + 6)
	p.After.R[cpu.SP] += 12
	p.After.IP = v.Read16(stack + 8)
	p.After.Seg[cpu.CS] = v.Read16(stack + 10)
	return imageTimingReturn(v, p, budget, 8, 59, 0x4bca)
}

func planPaletteAdvance(v CPUView, budget uint64) *machine.HostWorkPlan {
	advance := v.IP() == 0x4a72
	if v.CS() != 0x48a1 || (!advance && v.IP() != 0x4a31) || v.ES() != 0x48a1 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [82]byte
	copyView(v, 0x48a10+0x4a31, code[:])
	if sha256.Sum256(code[:]) != paletteAdvanceCode.sum {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if v.SP() < 2 || v.SP() > 65530 || stack < 0x502 || stack+4 > 0xa0000 || stack-2 < 0x48a10+0x7580 && stack+4 > 0x48a10 {
		return nil
	}
	y := v.BP()
	bottom := word(0xec7)
	if y > bottom || bottom >= 200 {
		return nil
	}
	p := imageBasePlan(v)
	if advance {
		if v.Read16(stack) != 0x48a1 || v.DS() != word(0xecf) {
			return nil
		}
		y++
		p.After.R[cpu.BP] = y
		if y > bottom {
			p.After.R[cpu.BP] = v.Read16(stack + 2)
			p.After.R[cpu.BX] = 255
			p.After.R[cpu.DI] = 0x1572
			p.After.R[cpu.SP] += 4
			p.After.Seg[cpu.DS] = 0x48a1
			p.After.Seg[cpu.ES] = 0x48a1
			p.After.IP = 0x4a83
			p.After.Flags = nativeCompareFlags(v.Flags(), y, bottom)
			p.Writes = []machine.HostWrite{{Address: stack + 2, Value: 0xa1, Step: 8, CS: 0x48a1, IP: 0x4a81}, {Address: stack + 3, Value: 0x48, Step: 8, CS: 0x48a1, IP: 0x4a81}}
			return imageTimingReturn(v, p, budget, 9, 30, 0x4a82)
		}
	} else if v.DS() != 0x48a1 {
		return nil
	}
	stride, width := uint32(word(0xedd)), uint32(word(0x1135))
	x, dy := int32(word(0xec1))-int32(word(0xed3)), int32(y)-int32(word(0xed5))
	if word(0xecf) != 0xa000 || stride == 0 || stride > 320 || width == 0 || x < 0 || uint32(x)+width > stride || dy < 0 || dy >= 200 {
		return nil
	}
	row := uint32(x) + uint32(dy)*stride
	if row+width > 64000 {
		return nil
	}
	n, cost, saveStep := uint64(18), uint64(111), uint64(15)
	p.After.R[cpu.AX] = uint16(uint32(dy) * stride)
	p.After.R[cpu.DX] = 0
	p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x), uint16(uint32(dy)*stride))
	if stride == 320 {
		n, cost, saveStep = 20, 101, 17
		p.After.R[cpu.AX] = 320
		p.After.R[cpu.DX] = uint16(dy * 64)
		p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x+dy*256), uint16(dy*64))
	}
	at := stack
	if !advance {
		n -= 4
		cost -= 19
		saveStep -= 4
		at -= 2
		p.After.R[cpu.SP] -= 2
	}
	p.Writes = []machine.HostWrite{{Address: at, Value: 0xa1, Step: saveStep, CS: 0x48a1, IP: 0x4a5a}, {Address: at + 1, Value: 0x48, Step: saveStep, CS: 0x48a1, IP: 0x4a5a}}
	p.After.Seg[cpu.DS] = 0xa000
	p.After.R[cpu.DI] = uint16(row)
	p.After.R[cpu.SI] = uint16(row)
	p.After.R[cpu.CX] = uint16(width)
	p.After.IP = 0x4a67
	return imageTimingReturn(v, p, budget, n, cost, 0x4a62)
}
