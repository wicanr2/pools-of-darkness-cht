// Typed Go ellipse geometry and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"sort"
)

func planEllipsePrefix(v CPUView) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x346a || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := ellipseOutlineCode
	var code [1352]byte
	copyView(v, 0x48a10+0x346a, code[:])
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

	if root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 {
		return nil
	}
	cx, cy, rx, ry, color := v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12), v.Read16(stack+14)
	if rx == 0 || rx > 160 || ry == 0 || ry > 100 || int16(cx) < -160 || int16(cx) > 479 || int16(cy) < -100 || int16(cy) > 299 {
		return nil
	}
	a, b := uint32(rx)*uint32(rx), uint32(ry)*uint32(ry)
	twiceA, twiceB := 2*a, 2*b
	gy := 2 * a * uint32(ry)
	partial := a/4 + b
	product := a * uint32(ry)
	err := partial - product
	n := uint64(55 + 4*links)
	cost := uint64(235)
	if links > 0 {
		cost += 21*links - 10
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{
		{0x1129, uint16(byte(color)) * 257, 3, 0x3549}, {0x112d, cx, 5, 0x354f}, {0x1131, cy, 7, 0x3555}, {0x112b, 0, 8, 0x3558}, {0x112f, ry, 10, 0x3561},
		{0x1155, uint16(a), 13, 0x3569}, {0x1157, uint16(a >> 16), 14, 0x356c}, {0x115d, uint16(twiceA), 17, 0x3574}, {0x115f, uint16(twiceA >> 16), 18, 0x3577},
		{0x1159, uint16(b), 21, 0x3580}, {0x115b, uint16(b >> 16), 22, 0x3583}, {0x1161, uint16(twiceB), 25, 0x358b}, {0x1163, uint16(twiceB >> 16), 26, 0x358e},
		{0x114d, 0, 28, 0x3594}, {0x114f, 0, 29, 0x3597}, {0x1151, uint16(gy), 40, 0x35af}, {0x1153, uint16(gy >> 16), 41, 0x35b2},
		{0x1149, uint16(partial), 50, 0x35cd}, {0x114b, uint16(partial >> 16), 51, 0x35d0}, {0x1149, uint16(err), 62, 0x35e9}, {0x114b, uint16(err >> 16), 63, 0x35ed}} {
		imageWord(p, w.off, w.value, n+w.step, w.ip)
	}
	for _, w := range []struct {
		step      uint64
		ip, value uint16
	}{{33, 0x35a4, uint16(twiceA)}, {55, 0x35de, uint16(a)}} {
		p.Writes = append(p.Writes, machine.HostWrite{Address: stack - 10, Value: byte(w.value), Step: n + w.step, CS: 0x48a1, IP: w.ip}, machine.HostWrite{Address: stack - 9, Value: byte(w.value >> 8), Step: n + w.step, CS: 0x48a1, IP: w.ip})
	}
	p.Instructions, p.Cycles, p.LastIP = n+64, cost+312, 0x35f1
	if v.DOSBoxCost() {
		p.Cycles = p.Instructions
	}
	p.After.R[cpu.AX], p.After.R[cpu.DX], p.After.R[cpu.BX], p.After.R[cpu.CX] = uint16(product), uint16(product>>16), ry, 0
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.BP] = v.SP() - 2
	p.After.Seg[cpu.DS] = v.CS()
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = 0x35f4
	borrow := uint16(0)
	if uint16(partial) < uint16(product) {
		borrow = 1
	}
	p.After.Flags = ellipseSubFlags(v.Flags(), uint16(partial>>16), uint16(product>>16), borrow)
	sort.SliceStable(p.Writes, func(i, j int) bool { return p.Writes[i].Step < p.Writes[j].Step })
	return p
}
