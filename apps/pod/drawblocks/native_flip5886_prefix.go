// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func planFlip5886Prefix(v CPUView) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.IP() != 0x5886 || v.Read8(0x449) != 0x13 {
		return nil
	}
	fp := flip5886Code
	var code [418]byte
	copyView(v, 0x48a10+0x5886, code[:])
	if sha256.Sum256(code[:]) != fp {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 10)
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

	n, cost := uint64(50+4*links), uint64(189)
	if links > 0 {
		cost += 21*links - 10
	}
	horizontal, vertical := v.Read16(stack+6) != 0, v.Read16(stack+8) != 0
	phase := func(steps, cycles uint64) { n += steps; cost += cycles }
	phase(5, 46)
	p.After.R[cpu.BX] = slot * 2
	cursor := s
	for cursor.flags&64 != 0 {
		p.After.R[cpu.BX] = cursor.seg
		cursor = descriptor(cursor.seg)
	}
	p.After.Seg[cpu.ES] = 0xa000
	p.After.Seg[cpu.DS] = v.CS()
	p.After.R[cpu.SP] -= 8
	p.After.R[cpu.BP] = v.SP() - 2
	if !horizontal && !vertical {
		cost -= 5
		phase(6, 47)
		p.After.R[cpu.AX] = 0
		p.After.Seg[cpu.DS] = v.DS()
		p.After.R[cpu.SP] += 12
		p.After.R[cpu.BP] = v.BP()
		p.After.IP = v.Read16(stack)
		p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0) &^ cpu.DF
		p.LastIP = 0x5a27
	} else {
		rows := s.bottom - s.top + 1
		if !vertical {
			phase(4, 17)
		} else {
			rows /= 2
			if rows > 0 {
				phase(6, 21)
			} else {
				rows = 1
				phase(7, 18)
			}
		}
		imageWord(p, 0x1171, rows, n, 0x5944)
		phase(14, 52)
		if vertical {
			phase(15, 58)
		} else {
			phase(2, 13)
		}
		row := s.top
		if vertical {
			row = s.bottom
		}
		p.After.R[cpu.AX] = 320
		p.After.R[cpu.BX] = row*320 + s.left
		p.After.R[cpu.DX] = row * 64
		p.After.R[cpu.SI] = s.top*320 + s.left
		p.After.R[cpu.DI] = row*320 + s.left
		p.After.IP = 0x59a9
		if vertical {
			p.After.Flags = copyAddFlags(v.Flags(), s.left+row*256, row*64) &^ cpu.DF
			p.LastIP = 0x59a7
		} else {
			p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0) &^ cpu.DF
			p.LastIP = 0x597a
		}
	}
	p.Instructions, p.Cycles = n, cost
	if v.DOSBoxCost() {
		p.Cycles = n
	}
	return p
}
