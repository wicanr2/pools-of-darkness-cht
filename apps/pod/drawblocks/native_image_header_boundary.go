package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// Parse the source header and visible interval once. Register checkpoints are
// derived from that geometry, including excluded and clipped spans.
func planImageHeaderBoundary(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !imageAdmit(v, 0x689e, imageSpanCode) {
		return nil
	}
	marker := v.Read16(0x48a10 + 0x1250)
	initial := v.IP() <= 0x68a5
	j := drawingJobProjection{r: boundedMemoryRegisters(v)}
	none := func() {}
	if initial {
		j.r.IP = 0x689e
		j.add(0x68a5, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, marker, 0x8000) })
		next, cost := uint16(0x68a7), uint64(2)
		if marker >= 0x8000 {
			next, cost = 0x6850, 7
		}
		j.add(next, cost, 1, nil, none)
		if marker >= 0x8000 {
			return j.plan(v, budget)
		}
	}
	consumed := uint32(0)
	if v.IP() >= 0x68e2 {
		consumed = 4
	} else if v.IP() >= 0x68a8 {
		consumed = 2
	}
	si := uint32(v.SI())
	if si < consumed {
		return nil
	}
	si -= consumed
	base := uint32(v.DS()) * 16
	arg := uint32(v.SS())*16 + uint32(v.BP()) + 8
	if v.BP() > 65525 || arg < 0x500 || arg+2 > 0xa0000 || si+4 > 65536 || base+si < 0x500 || base+si+4 > 0xa0000 {
		return nil
	}
	raw := v.Read16(base + si)
	width, compressed := uint32(v.Read8(base+si+2)), uint32(v.Read8(base+si+3))
	x := int32(raw&0x7fff) + int32(int16(v.Read16(arg)))
	left, right := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5))
	row := int32(v.Read16(0x48a10 + 0x120c))
	if x < -255 || x > 574 || width == 0 || compressed == 0 || left > right || right >= 320 || si+4+compressed > 65535 || base+si+4+compressed > 0xa0000 || base+si < 0x48a10+0x1252 && base+si+4+compressed > 0x48a10+0x1250 || arg < 0x48a10+0x1252 && arg+2 > 0x48a10+0x1250 {
		return nil
	}
	end := x + int32(width) - 1
	visibleLeft, visibleRight := x, end
	if visibleLeft < left {
		visibleLeft = left
	}
	if visibleRight > right {
		visibleRight = right
	}
	if visibleLeft <= visibleRight {
		at, count := row+visibleLeft-left, visibleRight-visibleLeft+1
		if at < 0 || at+count > 64000 || at%320+count > 320 {
			return nil
		}
	}
	if v.IP() >= 0x68ac && marker != raw {
		return nil
	}
	if !initial {
		j.r.IP = 0x68a7
		j.r.R[cpu.SI] = uint16(si)
	}
	ux, dx := uint16(x), uint16(x+int32(width))
	offset := uint16(row + x - left)
	j.add(0x68a8, 7, 2, nil, func() { j.r.R[cpu.AX] = raw; j.r.R[cpu.SI] = uint16(si + 2) })
	j.add(0x68ac, 3, 1, drawingJobWord(0x48a10+0x1250, raw), none)
	j.add(0x68af, 2, 1, nil, func() { j.r.R[cpu.AX] = raw & 0x7fff; j.r.Flags = nativeCompareFlags(j.r.Flags, raw&0x7fff, 0) })
	j.add(0x68b2, 6, 1, nil, func() { j.r.Flags = imageRowAddFlags(j.r.Flags, raw&0x7fff, v.Read16(arg)); j.r.R[cpu.AX] = ux })
	j.add(0x68b4, 2, 1, nil, func() { j.r.R[cpu.DI] = ux })
	j.add(0x68b9, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, ux, uint16(left)); j.r.R[cpu.DI] = uint16(x - left) })
	j.add(0x68be, 7, 1, nil, func() { j.r.Flags = imageRowAddFlags(j.r.Flags, uint16(x-left), uint16(row)); j.r.R[cpu.DI] = offset })
	j.add(0x68c0, 2, 1, nil, func() { j.r.R[cpu.DX] = ux })
	j.add(0x68c2, 6, 1, nil, func() {
		j.r.Flags = drawingByteAddFlags(j.r.Flags, byte(ux), byte(width))
		j.r.R[cpu.DX] = ux&0xff00 | uint16(byte(dx))
	})
	carry := byte(0)
	if uint16(byte(ux))+uint16(width) > 255 {
		carry = 1
	}
	j.add(0x68c5, 2, 1, nil, func() { j.r.Flags = drawingByteAddFlags(j.r.Flags, byte(ux>>8), carry); j.r.R[cpu.DX] = dx })
	j.add(0x68c6, 2, 1, nil, func() {
		j.r.Flags = nativeCompareFlags(j.r.Flags, dx, 1)&^uint16(cpu.CF) | j.r.Flags&cpu.CF
		j.r.R[cpu.DX] = uint16(end)
	})
	j.add(0x68cb, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, ux, uint16(right)) })
	if x > right {
		j.add(0x68cd, 2, 1, nil, none)
		j.add(0x6840, 7, 1, nil, none)
		return j.plan(v, budget)
	}
	j.add(0x68d0, 7, 1, nil, none)
	j.add(0x68d5, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, uint16(end), uint16(left)) })
	if end < left {
		j.add(0x68d7, 2, 1, nil, none)
		j.add(0x6848, 7, 1, nil, none)
		return j.plan(v, budget)
	}
	j.add(0x68da, 7, 1, nil, none)
	j.add(0x68dd, 6, 1, nil, func() { j.r.R[cpu.BX] = j.r.R[cpu.BX]&0xff00 | uint16(compressed) })
	j.add(0x68df, 2, 1, nil, func() { j.r.R[cpu.BX] = uint16(compressed); j.r.Flags = nativeCompareFlags(j.r.Flags, 0, 0) })
	j.add(0x68e2, 2, 1, nil, func() { j.r.R[cpu.SI] = uint16(si + 4); j.r.Flags = imageRowAddFlags(j.r.Flags, uint16(si+2), 2) })
	j.add(0x68e4, 2, 1, nil, func() {
		j.r.R[cpu.BX] = uint16(si + 4 + compressed)
		j.r.Flags = imageRowAddFlags(j.r.Flags, uint16(compressed), uint16(si+4))
	})
	j.add(0x68e6, 2, 1, nil, func() { j.r.R[cpu.CX] &= 255; j.r.Flags = nativeCompareFlags(j.r.Flags, 0, 0) })
	j.add(0x68eb, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, ux, uint16(left)) })
	if x < left {
		j.add(0x691c, 7, 1, nil, none)
		return j.plan(v, budget)
	}
	j.add(0x68ed, 2, 1, nil, none)
	j.add(0x68f2, 7, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, uint16(end), uint16(right)) })
	next, cost := uint16(0x68f4), uint64(2)
	if end > right {
		next, cost = 0x691c, 7
	}
	j.add(next, cost, 1, nil, none)
	return j.plan(v, budget)
}
