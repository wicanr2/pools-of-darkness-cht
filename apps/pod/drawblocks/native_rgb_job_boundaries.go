package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFadeBlendBoundary(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.DS() != 0x48a1 || !drawingJobCode(v, 0x4b6c, fadeBlendCode) {
		return nil
	}
	ip := v.IP()
	index := v.SI()
	if ip == 0x4b8b {
		if index == 0 {
			return nil
		}
		index--
	}
	used := v.Read16(0x48a10 + 0x116f)
	if used < 1 || used > 256 || v.CX() == 0 || uint32(index)+uint32(v.CX()) != uint32(used)*3 {
		return nil
	}
	threshold := v.Read8(0x48a10 + 0x1173)
	if threshold > 63 {
		return nil
	}
	j := drawingJobProjection{r: boundedMemoryRegisters(v)}
	none := func() {}
	if ip >= 0x4b8a {
		if v.Read8(0x48a10+0x1e72+uint32(index)) != byte(v.AX()) || byte(v.AX()) > 126 {
			return nil
		}
		j.r.IP = 0x4b8a
		j.r.R[cpu.SI] = index
		j.add(0x4b8b, 2, 1, nil, func() { j.r.Flags = fadeAdvanceFlags(j.r.Flags, index, j.r.Flags&cpu.CF != 0); j.r.R[cpu.SI]++ })
		next := uint16(0x4b6c)
		if v.CX() == 1 {
			next = 0x4b8d
		}
		j.add(next, 2, 1, nil, func() { j.r.R[cpu.CX]-- })
		return j.plan(v, budget)
	}
	err := v.Read8(0x48a10 + 0x1e72 + uint32(index))
	distance := v.Read8(0x48a10 + 0x2172 + uint32(index))
	if err > 63 || distance > 63 {
		return nil
	}
	sum := err + distance
	advance := sum > threshold
	value := sum
	color, direction := byte(0), byte(0)
	if advance {
		value -= threshold
		color = v.Read8(0x48a10 + 0x1872 + uint32(index))
		direction = v.Read8(0x48a10 + 0x2472 + uint32(index))
		if ip >= 0x4b86 {
			if direction == 1 && color > 0 {
				color--
			} else if direction == 255 && color < 63 {
				color++
			} else {
				return nil
			}
		}
		if color > 63 || (direction != 1 && direction != 255) || (direction == 1 && color == 63) || (direction == 255 && color == 0) {
			return nil
		}
	}
	j.r.IP = 0x4b6c
	j.add(0x4b70, 6, 1, nil, func() { j.r.R[cpu.AX] = j.r.R[cpu.AX]&0xff00 | uint16(err) })
	j.add(0x4b74, 6, 1, nil, func() {
		j.r.R[cpu.AX] = j.r.R[cpu.AX]&0xff00 | uint16(sum)
		j.r.Flags = drawingByteAddFlags(j.r.Flags, err, distance)
	})
	j.add(0x4b78, 6, 1, nil, func() { j.r.Flags = drawingByteSubtractFlags(j.r.Flags, sum, threshold) })
	if advance {
		j.add(0x4b7a, 2, 1, nil, none)
		j.add(0x4b7e, 6, 1, nil, func() {
			j.r.R[cpu.AX] = j.r.R[cpu.AX]&0xff00 | uint16(value)
			j.r.Flags = drawingByteSubtractFlags(j.r.Flags, sum, threshold)
		})
		j.add(0x4b82, 6, 1, nil, func() { j.r.R[cpu.BX] = j.r.R[cpu.BX]&0xff00 | uint16(direction) })
		j.add(0x4b86, 6, 1, []machine.HostWrite{{Address: 0x48a10 + 0x1872 + uint32(index), Value: color + direction}}, func() { j.r.Flags = drawingByteAddFlags(j.r.Flags, color, direction) })
	} else {
		j.add(0x4b86, 7, 1, nil, none)
	}
	j.add(0x4b8a, 6, 1, []machine.HostWrite{{Address: 0x48a10 + 0x1e72 + uint32(index), Value: value}}, none)
	j.add(0x4b8b, 2, 1, nil, func() { j.r.Flags = fadeAdvanceFlags(j.r.Flags, index, j.r.Flags&cpu.CF != 0); j.r.R[cpu.SI]++ })
	next := uint16(0x4b6c)
	if v.CX() == 1 {
		next = 0x4b8d
	}
	j.add(next, 2, 1, nil, func() { j.r.R[cpu.CX]-- })
	return j.plan(v, budget)
}

// Build the next palette-index/RGB call from its source array, with an
// independently validated caller frame at every interrupted handoff.
func planFadeColorBoundary(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.DS() != 0x48a1 || !drawingJobCode(v, 0x4b9a, fadeColorCode) {
		return nil
	}
	used := v.Read16(0x48a10 + 0x116f)
	if used < 1 || used > 256 || v.BP() < 18 {
		return nil
	}
	sp := uint32(v.BP()) - 6
	stack := uint32(v.SS())*16 + sp
	if sp+12 > 65536 || stack < 0x50c || stack+12 > 0xa0000 || stack-12 < 0x50000 && stack+12 > 0x48a10 {
		return nil
	}
	ip := v.IP()
	j := drawingJobProjection{r: boundedMemoryRegisters(v)}
	none := func() {}
	if ip == 0x4bc1 || ip == 0x4bc4 {
		if v.DI()&1 != 0 || v.DI() >= used*2 || v.SI() != 0x1872+(used-v.DI()/2)*3 {
			return nil
		}
		j.r.IP = 0x4bc1
		j.r.R[cpu.SP] = uint16(sp) - 8
		j.add(0x4bc4, 2, 1, nil, func() { j.r.R[cpu.SP] = uint16(sp); j.r.Flags = imageRowAddFlags(j.r.Flags, uint16(sp)-8, 8) })
		j.add(0x4b9a, 7, 1, nil, none)
		return j.plan(v, budget)
	}
	initialDI := v.DI()
	if ip != 0x4b9a {
		initialDI += 2
	}
	if initialDI&1 != 0 || initialDI > used*2 {
		return nil
	}
	base := uint16(0x1872 + (used-initialDI/2)*3)
	j.r.IP = 0x4b9a
	j.r.R[cpu.SP] = uint16(sp)
	j.r.R[cpu.DI] = initialDI
	j.r.R[cpu.SI] = base
	j.add(0x4b9d, 2, 1, nil, func() { j.r.R[cpu.DI] = initialDI - 2; j.r.Flags = nativeCompareFlags(j.r.Flags, initialDI, 2) })
	if initialDI == 0 {
		j.add(0x4b55, 7, 1, nil, none)
		return j.plan(v, budget)
	}
	index := v.Read16(0x48a10 + 0x1572 + uint32(initialDI-2))
	rgb := [3]byte{v.Read8(0x48a10 + uint32(base)), v.Read8(0x48a10 + uint32(base) + 1), v.Read8(0x48a10 + uint32(base) + 2)}
	if index > 255 || rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return nil
	}
	for i, after := range []uint16{0x4ba5, 0x4ba9, 0x4bad} {
		if ip >= after && v.Read16(0x48a10+0x117f+uint32(i)*2) != uint16(rgb[i]) {
			return nil
		}
	}
	j.add(0x4b9f, 2, 1, nil, none)
	j.add(0x4ba1, 2, 1, nil, func() { j.r.R[cpu.AX] &= 255; j.r.Flags = nativeCompareFlags(j.r.Flags, 0, 0) })
	nexts := [][2]uint16{{0x4ba2, 0x4ba5}, {0x4ba6, 0x4ba9}, {0x4baa, 0x4bad}}
	for i, channel := range rgb {
		i, channel := i, channel
		j.add(nexts[i][0], 7, 2, nil, func() { j.r.R[cpu.AX] = uint16(channel); j.r.R[cpu.SI] = base + uint16(i) + 1 })
		j.add(nexts[i][1], 2, 1, drawingJobWord(0x48a10+0x117f+uint32(i)*2, uint16(channel)), none)
	}
	values := [6]uint16{uint16(rgb[2]), uint16(rgb[1]), uint16(rgb[0]), index, 0x48a1, 0x4bc1}
	next := [6]uint16{0x4bb1, 0x4bb5, 0x4bb9, 0x4bbd, 0x4bbe, 0x2aa8}
	for i, value := range values {
		i, value := i, value
		cost := uint64(8)
		if i == 4 {
			cost = 4
		}
		if i == 5 {
			cost = 11
		}
		j.add(next[i], cost, 1, drawingJobWord(stack-uint32(i+1)*2, value), func() { j.r.R[cpu.SP] = uint16(sp) - uint16(i+1)*2 })
	}
	return j.plan(v, budget)
}
