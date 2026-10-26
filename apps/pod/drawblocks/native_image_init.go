package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var imageInitCode = smallFingerprint(255, "d9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54")

// Resolve the destination surface and an indexed asset frame using Go data
// structures. The location table below describes ABI writes, not raster work.
func planImageInitialize(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x6733 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [255]byte
	copyView(v, 0x48a10+0x6733, code[:])
	if sha256.Sum256(code[:]) != imageInitCode.sum {
		return nil
	}
	sp := uint32(v.SP())
	stack := uint32(v.SS())*16 + sp
	if sp < 10 || sp+18 > 65536 || stack < 0x50a || stack+18 > 0xa0000 || stack-10 < 0x48a10+0x7580 && stack+18 > 0x48a10 {
		return nil
	}
	arg := func(off uint32) uint16 { return v.Read16(stack + off) }
	slot := arg(4)
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
	base, off := uint32(arg(14))*16, uint32(arg(12))
	head := base + off
	if off+6 > 65536 || head < 0x500 || head+6 > 0xa0000 || head < 0x48a10+0x1252 && head+6 > 0x48a10+0xebd || head < stack+18 && head+6 > stack-10 {
		return nil
	}
	frames, index := v.Read16(head+4), arg(16)
	valid := index < frames
	physical := uint32(0)
	if valid {
		at := off + 6 + uint32(index)*4
		if index > 16381 || at+4 > 65536 || base+at+4 > 0xa0000 || base+at < 0x48a10+0x1252 && base+at+4 > 0x48a10+0xebd || base+at < stack+18 && base+at+4 > stack-10 {
			return nil
		}
		relative := uint32(v.Read16(base+at)) | uint32(v.Read16(base+at+2))<<16
		if relative > 0xfffff || head+relative < head {
			return nil
		}
		physical = head + relative
		if physical < 0x500 || physical+4 > 0xa0000 || physical < 0x48a10+0x1252 && physical+4 > 0x48a10+0xebd {
			return nil
		}
	}
	p := imageBasePlan(v)
	emit := func(a uint32, value uint16, step uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: a, Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: a + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
	}
	for _, w := range []struct {
		off   uint32
		value uint16
		step  uint64
		ip    uint16
	}{{2, v.BP(), 1, 0x6733}, {4, v.DS(), 3, 0x6736}, {6, v.SI(), 4, 0x6737}, {8, v.DI(), 5, 0x6738}, {10, v.CS(), 6, 0x6739}} {
		emit(stack-w.off, w.value, w.step, w.ip)
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{
		{0xebd, s.seg, 11, 0x6744}, {0xebf, s.extra, 13, 0x674b}, {0xec1, s.left, 15, 0x6752}, {0xec3, s.top, 17, 0x6759}, {0xec5, s.right, 19, 0x6760}, {0xecb, s.right - s.left + 1, 22, 0x6768}, {0xec7, s.bottom, 24, 0x676f}, {0xecd, s.bottom - s.top + 1, 27, 0x6777}, {0xec9, s.flags, 29, 0x677e},
	} {
		imageWord(p, w.off, w.value, w.step, w.ip)
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{
		{0xedb, root.flags, 32, 0x6793}, {0xecf, root.seg, 34, 0x679a}, {0xed1, root.extra, 36, 0x67a1}, {0xed5, root.top, 38, 0x67a8}, {0xed9, root.bottom, 40, 0x67af}, {0xedf, root.bottom - root.top + 1, 43, 0x67b7}, {0xed3, root.left, 45, 0x67be}, {0xed7, root.right, 47, 0x67c5}, {0xedd, root.right - root.left + 1, 50, 0x67cd},
	} {
		imageWord(p, w.off, w.value, w.step+links*4, w.ip)
	}
	p.After.Seg[cpu.ES] = root.seg
	n, cost := uint64(61)+4*links, uint64(304)
	if links > 0 {
		cost += 21*links - 10
	}
	if !valid {
		p.After.R[cpu.AX] = root.right - root.left + 1
		p.After.R[cpu.BX] = index
		p.After.R[cpu.SP] += 4
		p.After.IP = arg(0)
		p.After.Seg[cpu.CS] = arg(2)
		p.After.Flags = nativeCompareFlags(v.Flags(), index, frames) | v.Flags()&cpu.DF
		return imageTimingReturn(v, p, budget, n, cost, 0x6831)
	}
	p.After.R[cpu.BP] = v.SP() - 2
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.AX] = uint16(physical >> 4)
	p.After.R[cpu.BX] = index * 4
	p.After.R[cpu.DX] = 0
	p.After.Seg[cpu.DS] = uint16(physical >> 4)
	p.After.R[cpu.SI] = uint16(physical & 15)
	n, cost = 88+4*links, 353
	if links > 0 {
		cost += 21*links - 10
	}
	p.After.Flags = nativeCompareFlags(v.Flags(), arg(10)&3, 0)
	p.After.IP = 0x6a0c
	last := uint16(0x6825)
	if arg(10)&3 == 0 {
		n++
		cost += 7
		p.After.Flags = imageRowAddFlags(v.Flags()&^cpu.DF, p.After.R[cpu.SI], 4)
		p.After.R[cpu.SI] += 4
		p.After.IP = 0x6850
		last = 0x682b
	}
	return imageTimingReturn(v, p, budget, n, cost, last)
}

func imageTimingReturn(v CPUView, p *machine.HostWorkPlan, budget, n, cost uint64, last uint16) *machine.HostWorkPlan {
	if n > budget {
		return nil
	}
	if v.DOSBoxCost() {
		cost = n
	}
	p.Instructions, p.Cycles, p.LastIP = n, cost, last
	return p
}

func planImageReturn(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || (v.IP() != 0x682d && v.IP() != 0x6e5e) || v.Read8(0x449) != 0x13 || budget < 5 {
		return nil
	}
	for i, b := range []byte{0x5f, 0x5e, 0x1f, 0x5d, 0xcb} {
		if v.Read8(0x48a10+uint32(v.IP())+uint32(i)) != b {
			return nil
		}
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if v.SP() > 65524 || stack < 0x500 || stack+12 > 0xa0000 {
		return nil
	}
	p := imageBasePlan(v)
	p.After.R[cpu.DI] = v.Read16(stack)
	p.After.R[cpu.SI] = v.Read16(stack + 2)
	p.After.Seg[cpu.DS] = v.Read16(stack + 4)
	p.After.R[cpu.BP] = v.Read16(stack + 6)
	p.After.R[cpu.SP] += 12
	p.After.IP = v.Read16(stack + 8)
	p.After.Seg[cpu.CS] = v.Read16(stack + 10)
	return imageTimingReturn(v, p, budget, 5, 40, v.IP()+4)
}
