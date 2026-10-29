// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planScale5615Prefix(v CPUView) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x5615 || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := smallFingerprint(625, "6d9aac63d38f8b0e5ddd9bdbdfcfdda9144318a34398436b6b468c52f0271106").sum
	var code [625]byte
	copyView(v, 0x48a10+0x5615, code[:])
	if sha256.Sum256(code[:]) != fp {
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
		imageWord(p, w.off, w.value, oldN+w.step, w.ip+0x16b7)
	}
	n += 43 + 4*destLinks
	cost += 163
	if destLinks > 0 {
		cost += 21*destLinks - 10
	}

	factor := v.Read16(stack + 14)
	dw, dh := d.right-d.left+1, d.bottom-d.top+1
	cx, cy := v.Read16(stack+8), v.Read16(stack+10)
	if factor == 0 || factor > 8 || dw == 0 || int16(cx) < -160 || int16(cx) > 479 || int16(cy) < -100 || int16(cy) > 299 {
		return nil
	}
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	emit := func(off, value, ip uint16) { imageWord(p, off, value, n, ip) }
	// Typed scaling grid: centered source sample, integer repeats, scratch span.
	span := dw - 1
	samples := (span + factor - 1) / factor
	phase(2, 8)
	emit(0xeef, span, 0x5748)
	// Existing CPU contract: each unsigned DIV adds 25, MUL adds 14.
	phase(6, 16+25)
	emit(0x1212, samples, 0x575a)
	phase(2, 8+14)
	emit(0xecb, samples*factor, 0x5760)
	xhalf := span / 2 / factor
	xoff := cx - s.left - xhalf
	phase(8, 36+25)
	emit(0x1141, xoff, 0x5777)
	yhalf := (dh - 1) / 2 / factor
	yoff := cy - s.top - yhalf
	phase(8, 28+25)
	emit(0x113f, yoff, 0x578c)
	phase(2, 12)
	emit(0x1171, yoff+s.top, 0x5794)
	phase(1, 6)
	p.Instructions, p.Cycles, p.LastIP = n, cost, 0x5798
	if v.DOSBoxCost() {
		p.Cycles = n
	}
	p.After.R[cpu.AX] = yhalf
	p.After.R[cpu.BX] = xoff
	p.After.R[cpu.CX] = yoff + s.top
	p.After.R[cpu.DX] = d.top
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.BP] = v.SP() - 2
	p.After.Seg[cpu.DS] = v.CS()
	p.After.Flags = copyAddFlags(v.Flags(), yoff, s.top) &^ cpu.DF
	p.After.IP = 0x579c
	return p
}
