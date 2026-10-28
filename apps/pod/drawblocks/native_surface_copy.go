package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type surfaceCopyRowCursor struct {
	stack                          uint32
	width, remaining, source, dest uint16
	backX, backY                   bool
}

func surfaceCopyRowGeometry(v CPUView) (surfaceCopyRowCursor, bool) {
	if v.CS() != 0x48a1 || v.IP() != 0x415e || v.Read8(0x449) != 0x13 || v.DS() != 0xa000 || v.ES() != 0xa000 {
		return surfaceCopyRowCursor{}, false
	}
	var code [593]byte
	copyView(v, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		return surfaceCopyRowCursor{}, false
	}
	stack, ok := paletteSourceStack(v, 0, 16)
	if !ok || v.BP() != v.SP()+6 {
		return surfaceCopyRowCursor{}, false
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	for _, pair := range [][2]uint16{{v.Read16(stack + 12), 0xebd}, {v.Read16(stack + 14), 0xee1}} {
		slot, off := pair[0], pair[1]
		if slot > 255 {
			return surfaceCopyRowCursor{}, false
		}
		s := descriptor(slot * 2)
		r := s
		links := 0
		for r.flags&64 != 0 {
			if links >= 8 || r.seg > 510 || r.seg&1 != 0 {
				return surfaceCopyRowCursor{}, false
			}
			r = descriptor(r.seg)
			links++
		}
		if s.left > s.right || s.top > s.bottom || s.right > 319 || s.bottom > 199 {
			return surfaceCopyRowCursor{}, false
		}
		for i, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, r.seg, r.extra, r.left, r.top, r.right, r.bottom, r.flags, r.right - r.left + 1, r.bottom - r.top + 1} {
			if word(off+uint16(i)*2) != value {
				return surfaceCopyRowCursor{}, false
			}
		}
	}
	width, height, remaining := word(0x1135), uint16(0), word(0x1133)
	sh, dh := word(0xecd), word(0xef1)
	height = sh
	if dh < height {
		height = dh
	}
	sw, dw := word(0xecb), word(0xeef)
	if dw < sw {
		sw = dw
	}
	if width != sw || width == 0 || width > 320 || height == 0 || height > 200 || remaining == 0 || remaining > height {
		return surfaceCopyRowCursor{}, false
	}
	// This prototype admits full VGA roots and in-bounds selected rectangles only.
	for _, o := range []uint16{0xecf, 0xef3} {
		if word(o) != 0xa000 {
			return surfaceCopyRowCursor{}, false
		}
	}
	for _, o := range []uint16{0xed3, 0xef7} {
		if word(o) != 0 || word(o+2) != 0 || word(o+4) != 319 || word(o+6) != 199 {
			return surfaceCopyRowCursor{}, false
		}
	}
	if word(0xedd) != 320 || word(0xf01) != 320 {
		return surfaceCopyRowCursor{}, false
	}
	sx, sy, dx, dy := word(0xec1), word(0xec3), word(0xee5), word(0xee7)
	if sx+width > 320 || dx+width > 320 || sy+height > 200 || dy+height > 200 {
		return surfaceCopyRowCursor{}, false
	}
	backX, backY := sx < dx, sy < dy
	direction := uint16(1)
	if backY {
		direction = 0xffff
	}
	if word(0x1147) != direction || word(0x1139) != map[bool]uint16{false: 0, true: 1}[backX] {
		return surfaceCopyRowCursor{}, false
	}
	df := uint16(0)
	if backX {
		df = cpu.DF
	}
	if v.Flags()&cpu.DF != df {
		return surfaceCopyRowCursor{}, false
	}
	completed := height - remaining
	firstSX, firstDX, firstSY, firstDY := sx, dx, sy, dy
	if backX {
		firstSX += width - 2
		firstDX += width - 2
	}
	if backY {
		firstSY += height - 1
		firstDY += height - 1
	}
	source, dest := firstSY*320+firstSX, firstDY*320+firstDX
	if backY {
		source -= completed * 320
		dest -= completed * 320
	} else {
		source += completed * 320
		dest += completed * 320
	}
	if v.SI() != source || v.DI() != dest || word(0x113f) != firstSY || word(0x1141) != firstSX || word(0x1143) != firstDY || word(0x1145) != firstDX {
		return surfaceCopyRowCursor{}, false
	}
	return surfaceCopyRowCursor{stack, width, remaining, source, dest, backX, backY}, true
}

func planSurfaceCopyRows(v CPUView, budget uint64) *machine.HostWorkPlan {
	row, ok := surfaceCopyRowGeometry(v)
	if !ok {
		return nil
	}
	stack, width, remaining, source, dest, backX, backY := row.stack, row.width, row.remaining, row.source, row.dest, row.backX, row.backY
	df := uint16(0)
	if backX {
		df = cpu.DF
	}
	p := imageBasePlan(v)
	shadow := map[uint32]byte{}
	iterations := uint64(0)
	read := func(at uint32) byte {
		if b, ok := shadow[at]; ok {
			return b
		}
		return v.Read8(at)
	}
	for remaining > 0 {
		terminal := remaining == 1
		n, c := uint64(17), uint64(81)
		if backY {
			c -= 5
		}
		if terminal {
			n += 6
			c += 37
			if backY {
				n++
				c += 7
			}
		}
		if n > budget-p.Instructions {
			break
		}
		oldN := p.Instructions
		words, tail := width/2, width%2
		data := make([]byte, int(words)*2)
		si, di := source, dest
		for i := uint16(0); i < words; i++ {
			lo, hi := read(0xa0000+uint32(si)), read(0xa0000+uint32(si)+1)
			data[i*2], data[i*2+1] = lo, hi
			shadow[0xa0000+uint32(di)], shadow[0xa0000+uint32(di)+1] = lo, hi
			if backX {
				si -= 2
				di -= 2
			} else {
				si += 2
				di += 2
			}
		}
		if words > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(dest), Count: words, Width: 2, Backward: backX, Data: data, Step: oldN + 5, CS: 0x48a1, IP: 0x4169})
		}
		if backX {
			si++
			di++
		}
		if tail > 0 {
			value := read(0xa0000 + uint32(si))
			shadow[0xa0000+uint32(di)] = value
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(di), Count: 1, Width: 1, Backward: backX, Data: []byte{value}, Step: oldN + 9, CS: 0x48a1, IP: 0x4177})
		}
		p.After.R[cpu.AX], p.After.R[cpu.BX], p.After.R[cpu.CX] = source, dest, 0
		priorDest := dest
		carry := false
		if backY {
			source -= 320
			dest -= 320
			carry = priorDest < 320
		} else {
			source += 320
			dest += 320
			carry = dest < priorDest
		}
		oldRemaining := remaining
		remaining--
		ip := uint16(0x41a2)
		p.LastIP = 0x41a7
		if backY {
			ip, p.LastIP = 0x418f, 0x4194
		}
		imageWord(p, 0x1133, remaining, oldN+16, ip)
		p.Instructions += n
		p.Cycles += c + uint64(words+tail)*5
		iterations += uint64(words + tail)
		if v.DOSBoxCost() {
			p.Cycles = p.Instructions + iterations
		}
		p.After.Flags = nativeCompareFlags(p.After.Flags, oldRemaining, 1)&^cpu.CF | df
		if carry {
			p.After.Flags |= cpu.CF
		}
		p.After.R[cpu.SI], p.After.R[cpu.DI] = source, dest
		if terminal {
			p.After.Flags &^= cpu.DF
			p.After.R[cpu.DI] = v.Read16(stack)
			p.After.R[cpu.SI] = v.Read16(stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(stack + 4)
			p.After.R[cpu.BP] = v.Read16(stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(stack + 10)
			p.LastIP = 0x41ae
			return p
		}
		p.After.IP = 0x415e
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}

var surfaceCopyCode = smallFingerprint(593, "559f234ab6f1b97b5c6970fc84c40982e5525fdcac06436fd355bf5ec8c8d4fa").sum

func planSurfaceCopy(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 {
		return nil
	}
	if v.IP() == 0x3f5e {
		p := planSurfaceCopyPrefix(v)
		if p != nil && p.Instructions <= budget {
			return p
		}
		return nil
	}
	return planSurfaceCopyRows(v, budget)
}
