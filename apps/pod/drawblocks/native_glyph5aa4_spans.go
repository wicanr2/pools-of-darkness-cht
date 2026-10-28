// Spec025/#49: source-driven Go variable-width glyph pixels.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type glyph5aa4Frame struct {
	fontBase, data                    uint32
	fontOff, width, height, row, x, y uint16
	mapped                            bool
}

func readGlyph5aa4Frame(v CPUView) (f glyph5aa4Frame, ok bool) {
	if v.CS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return f, false
	}
	var code [467]byte
	copyView(v, 0x48a10+0x5aa4, code[:])
	if sha256.Sum256(code[:]) != smallFingerprint(467, "0ac343cdf83afd9cff35de6fa8dc5d42035e3be7f885c6da18e090dd2c2906aa").sum {
		return f, false
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if uint32(v.SP())+14 > 65536 || stack < 0x500 || stack+14 > 0xa0000 || stack < 0x48a10+0x757f && 0x48a10 < stack+14 {
		return f, false
	}
	ch := v.Read16(stack + 12)
	if ch > 255 || ch == 10 || ch == 13 {
		return f, false
	}
	context := uint32(word(0x110d))*16 + uint32(word(0x110b))
	if uint32(word(0x110b))+26 > 65536 || context < 0x500 || context+26 > 0xa0000 || context < 0x48a10+0x757f && 0x48a10 < context+26 || context < stack+14 && stack < context+26 {
		return f, false
	}
	for off := uint16(0); off < 10; off += 2 {
		if v.Read16(context+uint32(off)) != word(0x110f+off) {
			return f, false
		}
	}
	f.fontOff = word(0x1115)
	f.fontBase = uint32(word(0x1117)) * 16
	if v.DS() != word(0x1117) || uint32(f.fontOff)+0x108+uint32(ch)*2+2 > 65536 {
		return f, false
	}
	font := f.fontBase + uint32(f.fontOff)
	directory := font + 0x108 + uint32(ch)*2
	if font < 0x500 || directory+2 > 0xa0000 || ch >= v.Read16(font) {
		return f, false
	}
	f.height = v.Read16(font + 2)
	f.mapped = v.Read16(font+6) != 0
	record := uint32(f.fontOff) + uint32(v.Read16(directory))
	if record+2 > 65536 || f.fontBase+record+2 > 0xa0000 {
		return f, false
	}
	f.width = v.Read16(f.fontBase + record)
	size := uint32(f.width) * uint32(f.height)
	if f.width == 0 || f.width > 320 || f.height == 0 || f.height > 200 || record+2+size > 65536 || f.fontBase+record+2+size > 0xa0000 {
		return f, false
	}
	lo, hi := font, f.fontBase+record+2+size
	if directory+2 > hi {
		hi = directory + 2
	}
	if hi < lo || lo < 0x48a10+0x757f && 0x48a10 < hi || lo < context+26 && context < hi || lo < stack+14 && stack < hi {
		return f, false
	}
	remaining := word(0x11ff)
	if remaining == 0 || remaining > f.height {
		return f, false
	}
	f.row = f.height - remaining
	f.data = f.fontBase + record + 2 + uint32(f.row)*uint32(f.width)
	descriptor := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	slot := word(0x110f)
	if slot > 255 {
		return f, false
	}
	s := descriptor(slot * 2)
	root := s
	links := 0
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return f, false
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 || s.left > s.right || s.right > 319 || s.top > s.bottom || s.bottom > 199 {
		return f, false
	}
	for i, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
		if word(0xebd+uint16(i)*2) != value {
			return f, false
		}
	}
	f.x, f.y = word(0x1111), word(0x1113)
	if f.x < s.left || uint32(f.x)+uint32(f.width) > uint32(s.right)+1 || f.y < s.top || uint32(f.y)+uint32(f.height) > uint32(s.bottom)+1 {
		return f, false
	}
	if v.DX() != f.width {
		return f, false
	}
	return f, true
}

// Parse the immutable font directory and glyph row, then produce a native COPY
// row or mapped pixel. No guest instruction is decoded or executed here.
func planGlyph5aa4Spans(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || (v.IP() != 0x5c02 && v.IP() != 0x5c2c) {
		return nil
	}
	f, ok := readGlyph5aa4Frame(v)
	if !ok {
		return nil
	}
	row := uint16(uint32(f.y+f.row)*320 + uint32(f.x))
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, After: boundedMemoryRegisters(v), LastCS: 0x48a1, LastIP: v.IP()}
	if v.IP() == 0x5c02 {
		if f.mapped || v.CX() != f.width || v.BX() != f.fontOff || uint32(v.SI())+f.fontBase != f.data || v.DI() != row || v.AX() != row {
			return nil
		}
		data := make([]byte, int(f.width))
		copyView(v, f.data, data)
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + uint32(row), Data: data, Count: f.width, Width: 1, Step: 1, CS: 0x48a1, IP: 0x5c02}}
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.SI] += f.width
		p.After.R[cpu.DI] += f.width
		p.After.IP = 0x5c04
		p.Cycles = 3 + 5*uint64(f.width)
		if v.DOSBoxCost() {
			p.Cycles = 1 + uint64(f.width)
		}
	} else {
		if !f.mapped || v.CX() == 0 || v.CX() > f.width || v.BX() != f.fontOff+8 || v.BP() != row {
			return nil
		}
		col := f.width - v.CX()
		if v.SI() != uint16(f.data-f.fontBase+uint32(col)+1) || v.DI() != row+col {
			return nil
		}
		color := v.Read8(f.fontBase + uint32(f.fontOff) + 8 + uint32(v.Read8(f.data+uint32(col))))
		if color == 255 || byte(v.AX()) != color {
			return nil
		}
		p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + uint32(row+col), Value: uint16(color), Count: 1, Width: 1, Step: 1, CS: 0x48a1, IP: 0x5c2c}}
		p.After.R[cpu.DI]++
		p.After.IP = 0x5c2d
		p.Cycles = 7
		if v.DOSBoxCost() {
			p.Cycles = 2
		}
	}
	return p
}

func glyph5aa4FixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && (w.IP == 0x5c02 || w.IP == 0x5c2c) {
			return true
		}
	}
	for _, s := range transfers {
		if s.CS == 0x48a1 && (s.IP == 0x5c02 || s.IP == 0x5c2c) {
			return true
		}
	}
	return false
}
