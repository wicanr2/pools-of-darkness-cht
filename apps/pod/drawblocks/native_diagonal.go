// Interior diagonal rasterization uses Go integer error accumulators.
// Original positions retain only ABI, scratch and model-time notices.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planSteepLine(v CPUView, budget uint64) *machine.HostWorkPlan {
	plans := buildSteepLine(v)
	if plans == nil || plans.full.Instructions > budget {
		return nil
	}
	return plans.full
}
func buildSteepLine(v CPUView) *linePlans {
	var prefix *machine.HostWorkPlan
	if v.CS() != 0x48a1 || v.IP() != 0x2cfc || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := axisLineCode
	var code [1338]byte
	copyView(v, 0x48a10+0x2cfc, code[:])
	if sha256.Sum256(code[:]) != fp.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 16)
	if !ok {
		return nil
	}
	slot := v.Read16(stack + 4)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	s := descriptor(slot * 2)
	root := s
	links := uint64(0)
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return nil
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	p := imageBasePlan(v)
	entry := v.IP()
	delta := entry - 0x2b10
	for _, w := range []struct {
		off   uint32
		value uint16
		step  uint64
		ip    uint16
	}{{2, v.BP(), 1, 0x2b10}, {4, v.DS(), 3, 0x2b13}, {6, v.SI(), 4, 0x2b14}, {8, v.DI(), 5, 0x2b15}, {10, v.CS(), 6, 0x2b16}} {
		p.Writes = append(p.Writes, machine.HostWrite{Address: stack - w.off, Value: byte(w.value), Step: w.step, CS: 0x48a1, IP: w.ip + delta}, machine.HostWrite{Address: stack - w.off + 1, Value: byte(w.value >> 8), Step: w.step, CS: 0x48a1, IP: w.ip + delta})
	}
	// Common typed surface fields; source positions only retain ordered notices.
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xebd, s.seg, 11, 0x2b21}, {0xebf, s.extra, 13, 0x2b28}, {0xec1, s.left, 15, 0x2b2f}, {0xec3, s.top, 17, 0x2b36}, {0xec5, s.right, 19, 0x2b3d}, {0xecb, s.right - s.left + 1, 22, 0x2b45}, {0xec7, s.bottom, 24, 0x2b4c}, {0xecd, s.bottom - s.top + 1, 27, 0x2b54}, {0xec9, s.flags, 29, 0x2b5b}} {
		imageWord(p, w.off, w.value, w.step, w.ip+delta)
	}
	// The read variant shares the remaining shape at a different code offset.
	tailDelta := delta
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xedb, root.flags, 32, 0x2b70}, {0xecf, root.seg, 34, 0x2b77}, {0xed1, root.extra, 36, 0x2b7e}, {0xed5, root.top, 38, 0x2b85}, {0xed9, root.bottom, 40, 0x2b8c}, {0xedf, root.bottom - root.top + 1, 43, 0x2b94}, {0xed3, root.left, 45, 0x2b9b}, {0xed7, root.right, 47, 0x2ba2}, {0xedd, root.right - root.left + 1, 50, 0x2baa}} {
		imageWord(p, w.off, w.value, w.step+4*links, w.ip+tailDelta)
	}
	x0, y0, x1, y1, color := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12), v.Read16(stack+14)
	if x0 == x1 || y0 == y1 {
		return nil
	}
	ax, bx, cy, dy := x0, x1, y0, y1
	abs := func(x int32) int32 {
		if x < 0 {
			return -x
		}
		return x
	}
	if abs(int32(int16(x0))-int32(int16(x1))) > 32767 || abs(int32(int16(y0))-int32(int16(y1))) > 32767 {
		return nil
	}
	if int16(cy) >= int16(dy) {
		ax, bx, cy, dy = bx, ax, dy, cy
	}
	fromX, toX := ax, bx
	if x0 < s.left || x0 > s.right || x1 < s.left || x1 > s.right || y0 < s.top || y0 > s.bottom || y1 < s.top || y1 > s.bottom || abs(int32(int16(x0))-int32(int16(x1))) >= int32(dy-cy) {
		return nil
	}
	n, cost, iterations := uint64(50+4*links), uint64(189), uint64(0)
	if links > 0 {
		cost += 21*links - 10
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	finish := func() *linePlans {
		phase(6, 42)
		p.After.R[cpu.SP] += 4
		p.After.IP = v.Read16(stack)
		p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		p.Instructions, p.Cycles, p.LastIP = n, cost, 0x3235
		if v.DOSBoxCost() {
			p.Cycles = n + iterations
		}
		return &linePlans{full: p, prefix: prefix}
	}
	phase(6, 51)
	phase(2, 4)
	if int16(y0) < int16(y1) {
		cost += 5
	} else {
		phase(2, 4)
	}
	phase(1, 2)
	emit(0x112b, fromX, 0x2db2)
	phase(1, 6)
	emit(0x112d, toX, 0x2db5)
	phase(2, 13)
	phase(1, 6)
	emit(0x112f, cy, 0x2dc2)
	phase(2, 13)
	phase(1, 6)
	emit(0x1131, dy, 0x2dcf)
	phase(2, 4)
	if int16(fromX) < int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(4, 26)
	phase(2, 12)
	emit(0x1129, color, 0x2ded)
	difference := fromX - toX
	magnitude := uint16(abs(int32(int16(fromX)) - int32(int16(toX))))
	phase(3, 18)
	emit(0x1149, difference, 0x2df9)
	phase(1, 2)
	if int16(fromX) >= int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(1, 6)
	emit(0x114d, magnitude, 0x2e01)
	phase(3, 18)
	emit(0x1151, dy-cy, 0x2e0d)
	phase(2, 13)
	p.After.Seg[cpu.ES] = 0xa000
	// Both endpoints are inside; clipping is a separate unadmitted path.
	phase(6, 30)
	phase(2, 4)
	emit(0x1139, magnitude, 0x2f3d)
	phase(4, 16)
	emit(0x1137, (dy-cy)/2, 0x2f48)
	phase(4, 25)
	phase(2, 8)
	if int16(fromX) < int16(toX) {
		cost += 5
	}
	phase(5, 23)
	phase(9, 42)
	emit(0x1133, magnitude, 0x3053)
	phase(10, 50)
	if root.right-root.left+1 == 320 {
		phase(2, 0)
		cost -= 10
	}
	phase(6, 33)
	phase(2, 8)
	decrease := int16(fromX) > int16(toX)
	if decrease {
		cost += 5
	}
	prefix = imageBasePlan(v)
	prefix.Writes = append(prefix.Writes, p.Writes...)
	prefix.Instructions, prefix.Cycles = n, cost
	if v.DOSBoxCost() {
		prefix.Cycles = n
	}
	prefix.After.R[cpu.AX] = color
	prefix.After.R[cpu.BX] = (dy - cy) / 2
	prefix.After.R[cpu.CX] = dy - cy
	prefix.After.R[cpu.DX] = dy - cy
	prefix.After.R[cpu.SI] = magnitude
	prefix.After.R[cpu.BP] = dy - s.bottom
	prefix.After.R[cpu.DI] = uint16(uint32(cy-root.top)*uint32(root.right-root.left+1) + uint32(fromX-root.left))
	prefix.After.R[cpu.SP] -= 8
	prefix.After.Seg[cpu.DS] = v.CS()
	prefix.After.Seg[cpu.ES] = 0xa000
	prefix.After.Flags = nativeCompareFlags(v.Flags(), difference, 0)
	prefix.After.IP = 0x309c
	if decrease {
		prefix.After.IP = 0x30b9
	}
	prefix.LastIP = 0x309a
	major, minor := dy-cy, magnitude
	err := major / 2
	remaining := minor
	carry := false
	for step := uint16(0); step <= major; step++ {
		moved := (uint32(major/2) + uint32(step)*uint32(minor)) / uint32(major)
		x := int32(fromX) + int32(moved)
		pixelIP, declineIP := uint16(0x309c), uint16(0x30b0)
		if decrease {
			x = int32(fromX) - int32(moved)
			pixelIP, declineIP = 0x30b9, 0x30cd
		}
		at := uint32(cy+step-root.top)*uint32(root.right-root.left+1) + uint32(x-int32(root.left))
		phase(1, 7)
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0xa0000 + at, Value: byte(color), Step: n, CS: 0x48a1, IP: pixelIP})
		phase(2, 4)
		if step == major {
			cost += 5
			break
		}
		phase(4, 13)
		err += minor
		carry = err < major
		if carry {
			cost += 5
		} else {
			err -= major
			remaining--
			phase(4, 17)
			imageWord(p, 0x1133, remaining, n-1, declineIP)
		}
	}
	p.After.R[cpu.AX] = color
	p.After.R[cpu.BX] = err
	p.After.R[cpu.CX] = 0xffff
	p.After.R[cpu.DX] = major
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 1)
	if !carry {
		p.After.Flags &^= cpu.CF
	}
	phase(1, 7)
	return finish()
}

func planShallowLine(v CPUView, budget uint64) *machine.HostWorkPlan {
	plans := buildShallowLine(v)
	if plans == nil || plans.full.Instructions > budget {
		return nil
	}
	return plans.full
}
func buildShallowLine(v CPUView) *linePlans {
	var prefix *machine.HostWorkPlan
	if v.CS() != 0x48a1 || v.IP() != 0x2cfc || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := axisLineCode
	var code [1338]byte
	copyView(v, 0x48a10+0x2cfc, code[:])
	if sha256.Sum256(code[:]) != fp.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 16)
	if !ok {
		return nil
	}
	slot := v.Read16(stack + 4)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	s := descriptor(slot * 2)
	root := s
	links := uint64(0)
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return nil
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	p := imageBasePlan(v)
	entry := v.IP()
	delta := entry - 0x2b10
	for _, w := range []struct {
		off   uint32
		value uint16
		step  uint64
		ip    uint16
	}{{2, v.BP(), 1, 0x2b10}, {4, v.DS(), 3, 0x2b13}, {6, v.SI(), 4, 0x2b14}, {8, v.DI(), 5, 0x2b15}, {10, v.CS(), 6, 0x2b16}} {
		p.Writes = append(p.Writes, machine.HostWrite{Address: stack - w.off, Value: byte(w.value), Step: w.step, CS: 0x48a1, IP: w.ip + delta}, machine.HostWrite{Address: stack - w.off + 1, Value: byte(w.value >> 8), Step: w.step, CS: 0x48a1, IP: w.ip + delta})
	}
	// Common typed surface fields; source positions only retain ordered notices.
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xebd, s.seg, 11, 0x2b21}, {0xebf, s.extra, 13, 0x2b28}, {0xec1, s.left, 15, 0x2b2f}, {0xec3, s.top, 17, 0x2b36}, {0xec5, s.right, 19, 0x2b3d}, {0xecb, s.right - s.left + 1, 22, 0x2b45}, {0xec7, s.bottom, 24, 0x2b4c}, {0xecd, s.bottom - s.top + 1, 27, 0x2b54}, {0xec9, s.flags, 29, 0x2b5b}} {
		imageWord(p, w.off, w.value, w.step, w.ip+delta)
	}
	// The read variant shares the remaining shape at a different code offset.
	tailDelta := delta
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xedb, root.flags, 32, 0x2b70}, {0xecf, root.seg, 34, 0x2b77}, {0xed1, root.extra, 36, 0x2b7e}, {0xed5, root.top, 38, 0x2b85}, {0xed9, root.bottom, 40, 0x2b8c}, {0xedf, root.bottom - root.top + 1, 43, 0x2b94}, {0xed3, root.left, 45, 0x2b9b}, {0xed7, root.right, 47, 0x2ba2}, {0xedd, root.right - root.left + 1, 50, 0x2baa}} {
		imageWord(p, w.off, w.value, w.step+4*links, w.ip+tailDelta)
	}
	x0, y0, x1, y1, color := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12), v.Read16(stack+14)
	if x0 == x1 || y0 == y1 {
		return nil
	}
	ax, bx, cy, dy := x0, x1, y0, y1
	abs := func(x int32) int32 {
		if x < 0 {
			return -x
		}
		return x
	}
	if abs(int32(int16(x0))-int32(int16(x1))) > 32767 || abs(int32(int16(y0))-int32(int16(y1))) > 32767 {
		return nil
	}
	if int16(cy) >= int16(dy) {
		ax, bx, cy, dy = bx, ax, dy, cy
	}
	fromX, toX := ax, bx
	if x0 < s.left || x0 > s.right || x1 < s.left || x1 > s.right || y0 < s.top || y0 > s.bottom || y1 < s.top || y1 > s.bottom || abs(int32(int16(x0))-int32(int16(x1))) < int32(dy-cy) {
		return nil
	}
	n, cost, iterations := uint64(50+4*links), uint64(189), uint64(0)
	if links > 0 {
		cost += 21*links - 10
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	finish := func() *linePlans {
		phase(6, 42)
		p.After.R[cpu.SP] += 4
		p.After.IP = v.Read16(stack)
		p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		p.Instructions, p.Cycles, p.LastIP = n, cost, 0x3235
		if v.DOSBoxCost() {
			p.Cycles = n + iterations
		}
		return &linePlans{full: p, prefix: prefix}
	}
	phase(6, 51)
	phase(2, 4)
	if int16(y0) < int16(y1) {
		cost += 5
	} else {
		phase(2, 4)
	}
	phase(1, 2)
	emit(0x112b, fromX, 0x2db2)
	phase(1, 6)
	emit(0x112d, toX, 0x2db5)
	phase(2, 13)
	phase(1, 6)
	emit(0x112f, cy, 0x2dc2)
	phase(2, 13)
	phase(1, 6)
	emit(0x1131, dy, 0x2dcf)
	phase(2, 4)
	if int16(fromX) < int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(4, 26)
	phase(2, 12)
	emit(0x1129, color, 0x2ded)
	difference := fromX - toX
	magnitude := uint16(abs(int32(int16(fromX)) - int32(int16(toX))))
	phase(3, 18)
	emit(0x1149, difference, 0x2df9)
	phase(1, 2)
	if int16(fromX) >= int16(toX) {
		cost += 5
	} else {
		phase(1, 2)
	}
	phase(1, 6)
	emit(0x114d, magnitude, 0x2e01)
	phase(3, 18)
	emit(0x1151, dy-cy, 0x2e0d)
	phase(2, 13)
	p.After.Seg[cpu.ES] = 0xa000
	phase(7, 32)
	phase(4, 16)
	emit(0x1137, magnitude/2, 0x30de)
	phase(4, 20)
	decrease := int16(fromX) > int16(toX)
	if !decrease {
		cost += 5
		phase(3, 11)
	} else {
		phase(4, 17)
	}
	phase(2, 13)
	phase(6, 38)
	phase(10, 50)
	if root.right-root.left+1 == 320 {
		phase(2, 0)
		cost -= 10
	}
	phase(1, 6)
	emit(0xec7, cy, 0x31eb)
	phase(6, 28)
	if !decrease {
		cost += 5
	} else {
		phase(1, 2)
	}
	prefix = imageBasePlan(v)
	prefix.Writes = append(prefix.Writes, p.Writes...)
	prefix.Instructions, prefix.Cycles = n, cost
	if v.DOSBoxCost() {
		prefix.Cycles = n
	}
	prefix.After.R[cpu.AX] = color
	prefix.After.R[cpu.BX] = magnitude / 2
	prefix.After.R[cpu.CX] = magnitude
	prefix.After.R[cpu.DX] = magnitude
	prefix.After.R[cpu.SI] = s.bottom - cy
	prefix.After.R[cpu.BP] = dy - cy
	prefix.After.R[cpu.DI] = uint16(uint32(cy-root.top)*uint32(root.right-root.left+1) + uint32(fromX-root.left))
	prefix.After.R[cpu.SP] -= 8
	prefix.After.Seg[cpu.DS] = v.CS()
	prefix.After.Seg[cpu.ES] = 0xa000
	prefix.After.Flags = nativeCompareFlags(v.Flags(), difference, 0)
	prefix.After.IP = 0x321c
	prefix.LastIP = 0x3203
	if decrease {
		prefix.After.IP = 0x3206
		prefix.After.Flags |= cpu.DF
		prefix.LastIP = 0x3205
	}
	major, minor := magnitude, dy-cy
	err := major / 2
	carry := false
	for step := uint16(0); step <= major; step++ {
		moved := (uint32(major/2) + uint32(step)*uint32(minor)) / uint32(major)
		x := int32(fromX) + int32(step)
		pixelIP := uint16(0x321c)
		if decrease {
			x = int32(fromX) - int32(step)
			pixelIP = 0x3206
		}
		at := uint32(cy+uint16(moved)-root.top)*uint32(root.right-root.left+1) + uint32(x-int32(root.left))
		phase(1, 7)
		iterations++
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at, Value: color & 255, Count: 1, Width: 1, Step: n, CS: 0x48a1, IP: pixelIP, Backward: decrease})
		phase(2, 4)
		if step == major {
			cost += 5
			break
		}
		phase(3, 6)
		err += minor
		carry = err < major
		if carry {
			cost += 5
		} else {
			err -= major
			phase(4, 18)
		}
	}
	p.After.R[cpu.AX] = color
	p.After.R[cpu.BX] = err
	p.After.R[cpu.CX] = 0xffff
	p.After.R[cpu.DX] = major
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 1)
	if !carry {
		p.After.Flags &^= cpu.CF
	}
	return finish()
}

type linePlans struct{ full, prefix *machine.HostWorkPlan }

func chooseLinePlan(plans *linePlans, budget uint64) *machine.HostWorkPlan {
	if plans == nil {
		return nil
	}
	if plans.full.Instructions <= budget {
		return plans.full
	}
	if plans.prefix != nil && plans.prefix.Instructions <= budget {
		return plans.prefix
	}
	return nil
}
