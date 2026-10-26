package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var imageSpanCode = smallFingerprint(86, "56ecc68cd838cf71e5de2da105cf370ccd3c2f7656159bdb2de14800a9c888a1")

// planImageSpan parses one unclipped image span's geometry and payload bounds.
// This is an asset-header operation, with separate ABI and timing effects.
func planImageSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x689e || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 || budget < 26 || v.BP() > 65525 {
		return nil
	}
	var code [86]byte
	copyView(v, 0x48a10+0x689e, code[:])
	if sha256.Sum256(code[:]) != imageSpanCode.sum || v.Read16(0x48a10+0x1250) >= 0x8000 {
		return nil
	}
	src, si := uint32(v.DS())*16, uint32(v.SI())
	arg := uint32(v.SS())*16 + uint32(v.BP()) + 8
	if si+4 > 65536 || src+si+4 > 0xa0000 || arg+2 > 0xa0000 {
		return nil
	}
	raw := v.Read16(src + si)
	x := (raw & 0x7fff) + v.Read16(arg)
	width, compressed := uint32(v.Read8(src+si+2)), uint32(v.Read8(src+si+3))
	left, right, row := v.Read16(0x48a10+0xec1), v.Read16(0x48a10+0xec5), v.Read16(0x48a10+0x120c)
	if width == 0 || compressed == 0 || left > right || right >= 320 || x < left || uint32(x)+width-1 > uint32(right) || si+4+compressed > 65536 || src+si+4+compressed > 0xa0000 {
		return nil
	}
	di := uint32(row) + uint32(x-left)
	if di+width > 64000 || di%320+width > 320 || (src+si < 0x48a10+0x1252 && src+si+4+compressed > 0x48a10+0x1250) || (arg < 0x48a10+0x1252 && arg+2 > 0x48a10+0x1250) {
		return nil
	}
	endX := x + uint16(width) - 1
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 26, Cycles: 115, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: 0x68f4, Flags: nativeCompareFlags(v.Flags(), endX, right)}, LastCS: 0x48a1, LastIP: 0x68f2}
	p.Writes = []machine.HostWrite{{Address: 0x48a10 + 0x1250, Value: byte(raw), Step: 4, CS: 0x48a1, IP: 0x68a8}, {Address: 0x48a10 + 0x1251, Value: byte(raw >> 8), Step: 4, CS: 0x48a1, IP: 0x68a8}}
	p.After.R[cpu.AX] = x
	p.After.R[cpu.CX] &= 0xff
	p.After.R[cpu.DX] = endX
	p.After.R[cpu.BX] = uint16(si + 4 + compressed)
	p.After.R[cpu.SI] += 4
	p.After.R[cpu.DI] = uint16(di)
	if v.DOSBoxCost() {
		p.Cycles = 27
	}
	return p
}
