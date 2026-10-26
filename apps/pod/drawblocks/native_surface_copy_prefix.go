// Typed source/destination surface initialization, spec025/#49.
// Original positions retain only ABI, scratch and model-time notices.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planSurfaceCopyPrefix(v CPUView) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x3f5e || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := surfaceCopyCode
	var code [593]byte
	copyView(v, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != fp {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 8)
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

	if root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 {
		return nil
	}
	destSlot := v.Read16(stack + 6)
	if destSlot > 255 {
		return nil
	}
	d := descriptor(destSlot * 2)
	dr := d
	destLinks := uint64(0)
	for dr.flags&64 != 0 {
		if destLinks >= 8 || dr.seg > 510 || dr.seg&1 != 0 {
			return nil
		}
		dr = descriptor(dr.seg)
		destLinks++
	}
	if dr.seg != 0xa000 || dr.left != 0 || dr.top != 0 || dr.right != 319 || dr.bottom != 199 || d.left > d.right || d.right > 319 || d.top > d.bottom || d.bottom > 199 {
		return nil
	}
	n, cost := uint64(50+4*links), uint64(189)
	if links > 0 {
		cost += 21*links - 10
	}
	oldN := n
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{
		{0xee1, d.seg, 4, 0x4004}, {0xee3, d.extra, 6, 0x400b}, {0xee5, d.left, 8, 0x4012}, {0xee7, d.top, 10, 0x4019}, {0xee9, d.right, 12, 0x4020}, {0xeef, d.right - d.left + 1, 15, 0x4028}, {0xeeb, d.bottom, 17, 0x402f}, {0xef1, d.bottom - d.top + 1, 20, 0x4037}, {0xeed, d.flags, 22, 0x403e},
		{0xeff, dr.flags, 25 + 4*destLinks, 0x4053}, {0xef3, dr.seg, 27 + 4*destLinks, 0x405a}, {0xef5, dr.extra, 29 + 4*destLinks, 0x4061}, {0xef9, dr.top, 31 + 4*destLinks, 0x4068}, {0xefd, dr.bottom, 33 + 4*destLinks, 0x406f}, {0xf03, dr.bottom - dr.top + 1, 36 + 4*destLinks, 0x4077}, {0xef7, dr.left, 38 + 4*destLinks, 0x407e}, {0xefb, dr.right, 40 + 4*destLinks, 0x4085}, {0xf01, dr.right - dr.left + 1, 43 + 4*destLinks, 0x408d}} {
		imageWord(p, w.off, w.value, oldN+w.step, w.ip)
	}
	n += 43 + 4*destLinks
	cost += 163
	if destLinks > 0 {
		cost += 21*destLinks - 10
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	height, width := s.bottom-s.top+1, s.right-s.left+1
	dh, dw := d.bottom-d.top+1, d.right-d.left+1
	if height < dh {
		phase(3, 15)
	} else {
		height = dh
		phase(4, 12)
	}
	phase(1, 2)
	emit(0x1133, height, 0x409c)
	if width < dw {
		phase(3, 15)
	} else {
		width = dw
		phase(4, 12)
	}
	phase(1, 2)
	emit(0x1135, width, 0x40ab)
	sx, sy, dx, dy := s.left, s.top, d.left, d.top
	backX, backY := sx < dx, sy < dy
	direction := uint16(1)
	phase(5, 14)
	if !backY {
		cost += 5
	} else {
		phase(5, 18)
		sy += height - 1
		dy += height - 1
		direction = 0xffff
	}
	phase(1, 2)
	emit(0x113f, sy, 0x40c8)
	phase(1, 6)
	emit(0x1143, dy, 0x40cb)
	phase(1, 6)
	emit(0x1147, direction, 0x40cf)
	phase(6, 16)
	if !backX {
		cost += 5
	} else {
		phase(6, 20)
		sx += width - 2
		dx += width - 2
	}
	reverse := uint16(0)
	if backX {
		reverse = 1
	}
	phase(1, 2)
	emit(0x1141, sx, 0x40f1)
	phase(1, 6)
	emit(0x1145, dx, 0x40f4)
	phase(1, 6)
	emit(0x1139, reverse, 0x40f8)
	// Two segment loads include the existing model's 18-cost far reload each.
	phase(27, 148)
	p.Instructions, p.Cycles = n, cost
	if v.DOSBoxCost() {
		p.Cycles = n
	}
	p.LastIP = 0x4159
	p.After.R[cpu.AX] = 320
	p.After.R[cpu.BX] = dx
	p.After.R[cpu.CX] = reverse
	p.After.R[cpu.DX] = dy * 64
	p.After.R[cpu.SI] = sy*320 + sx
	p.After.R[cpu.DI] = dy*320 + dx
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.BP] = v.SP() - 2
	p.After.Seg[cpu.DS] = 0xa000
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = 0x415e
	p.After.Flags = copyAddFlags(v.Flags(), dx+dy*256, dy*64)
	if backX {
		p.After.Flags |= cpu.DF
	}
	return p
}
func copyAddFlags(before, a, b uint16) uint16 {
	result := a + b
	f := nativeCompareFlags(before, result, 0) &^ uint16(cpu.AF|cpu.OF|cpu.CF)
	if uint32(a)+uint32(b) > 65535 {
		f |= cpu.CF
	}
	if (a^b^result)&16 != 0 {
		f |= cpu.AF
	}
	if (^(a^b)&(a^result))&0x8000 != 0 {
		f |= cpu.OF
	}
	return f
}
