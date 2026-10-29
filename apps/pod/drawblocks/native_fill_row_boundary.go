package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFillRowBoundary(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.DS() != 0x48a1 || v.ES() != 0xa000 {
		return nil
	}
	var code [260]byte
	copyView(v, 0x48a10+0x41af, code[:])
	if sha256.Sum256(code[:]) != rectangleCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 16)
	if !ok || v.BP() != v.SP()+6 {
		return nil
	}
	s, root, ok := primitiveSurface(v, v.Read16(stack+12))
	if !ok || v.BX() < s.top || v.BX() > s.bottom {
		return nil
	}
	color := v.Read16(stack + 14)
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	if word(0x1129) != color || word(0x113b) != s.left || word(0x113d) != s.right {
		return nil
	}
	x, y, stride := s.left-root.left, v.BX()-root.top, root.right-root.left+1
	width := s.right - s.left + 1
	at := uint32(y)*uint32(stride) + uint32(x)
	if at+uint32(width) > 64000 {
		return nil
	}
	j := drawingJobProjection{r: boundedMemoryRegisters(v)}
	none := func() {}
	j.r.IP = 0x4268
	j.add(0x426c, 6, 1, nil, func() { j.r.R[cpu.DI] = s.left })
	j.add(0x4271, 7, 1, nil, func() { j.r.R[cpu.DI] = x; j.r.Flags = nativeCompareFlags(j.r.Flags, s.left, root.left) })
	j.add(0x4273, 2, 1, nil, func() { j.r.R[cpu.DX] = v.BX() })
	j.add(0x4278, 7, 1, nil, func() { j.r.R[cpu.DX] = y; j.r.Flags = nativeCompareFlags(j.r.Flags, v.BX(), root.top) })
	j.add(0x427c, 3, 1, nil, func() { j.r.R[cpu.AX] = stride })
	j.add(0x427f, 2, 1, nil, func() { j.r.Flags = nativeCompareFlags(j.r.Flags, stride, 320) })
	if stride == 320 {
		j.add(0x4287, 7, 1, nil, none)
		j.add(0x4289, 2, 1, nil, func() { j.r.R[cpu.DX] = y << 8 })
		j.add(0x428b, 2, 1, nil, func() { j.r.R[cpu.DI] = x + y*256; j.r.Flags = imageRowAddFlags(j.r.Flags, x, y*256) })
		j.add(0x428d, 2, 1, nil, func() { j.r.R[cpu.DX] = y << 7; j.r.Flags = drawingShiftRightFlags(j.r.Flags, y<<8) })
		j.add(0x428f, 2, 1, nil, func() { j.r.R[cpu.DX] = y << 6; j.r.Flags = drawingShiftRightFlags(j.r.Flags, y<<7) })
		j.add(0x4291, 2, 1, nil, func() { j.r.R[cpu.DI] = uint16(at); j.r.Flags = imageRowAddFlags(j.r.Flags, x+y*256, y*64) })
	} else {
		j.add(0x4281, 2, 1, nil, none)
		j.add(0x4283, 16, 1, nil, func() {
			j.r.R[cpu.AX] = y * stride
			j.r.R[cpu.DX] = 0
			j.r.Flags = j.r.Flags&^uint16(cpu.CF|cpu.PF|cpu.ZF|cpu.SF|cpu.OF) | cpu.PF | cpu.ZF
		})
		j.add(0x4285, 2, 1, nil, func() { j.r.R[cpu.DI] = uint16(at); j.r.Flags = imageRowAddFlags(j.r.Flags, x, y*stride) })
		j.add(0x4291, 7, 1, nil, none)
	}
	j.add(0x4295, 6, 1, nil, func() { j.r.R[cpu.CX] = s.right })
	j.add(0x4299, 6, 1, nil, func() { j.r.R[cpu.CX] = width - 1; j.r.Flags = nativeCompareFlags(j.r.Flags, s.right, s.left) })
	j.add(0x429a, 2, 1, nil, func() {
		j.r.R[cpu.CX] = width
		j.r.Flags = imageRowAddFlags(j.r.Flags, width-1, 1)&^uint16(cpu.CF) | j.r.Flags&cpu.CF
	})
	j.add(0x429d, 2, 1, nil, func() { j.r.R[cpu.AX] = color })
	j.add(0x429f, 2, 1, nil, func() { j.r.R[cpu.AX] = uint16(byte(color)) * 257 })
	j.add(0x42a1, 2, 1, nil, func() { j.r.R[cpu.CX] = width / 2; j.r.Flags = drawingShiftRightFlags(j.r.Flags, width) })
	return j.plan(v, budget)
}
