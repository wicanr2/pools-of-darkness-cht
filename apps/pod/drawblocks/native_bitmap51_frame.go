// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
)

var bitmap51Code = smallFingerprint(584, "dd209e837a213ee07c0b9602563167e0eed60702d8a96f4c20d98f27ebb157a2").sum

type bitmap51Frame struct {
	s, root                                  rectangleSurface
	stack, source                            uint32
	offset, width, stride, rowBytes, y, rows uint16
	planar                                   bool
}

func readBitmap51Frame(v CPUView) (f bitmap51Frame, ok bool) {
	if v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 || v.Flags()&cpu.DF != 0 {
		return f, false
	}
	var code [584]byte
	copyView(v, 0x48a10+0x51cc, code[:])
	if sha256.Sum256(code[:]) != bitmap51Code {
		return f, false
	}
	stack, good := paletteSourceStack(v, 2, 18)
	if !good || v.BP() != v.SP()+6 {
		return f, false
	}
	f.stack = stack
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	slot := v.Read16(stack + 12)
	if slot > 255 {
		return f, false
	}
	f.s = descriptor(slot * 2)
	f.root = f.s
	links := 0
	for f.root.flags&64 != 0 {
		if links >= 8 || f.root.seg > 510 || f.root.seg&1 != 0 {
			return f, false
		}
		f.root = descriptor(f.root.seg)
		links++
	}
	s, root := f.s, f.root
	if root.seg != 0xa000 || root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 || s.left > s.right || s.right > 319 || s.top > s.bottom || s.bottom > 199 {
		return f, false
	}
	f.y = word(0xec3)
	if f.y < s.top || f.y > s.bottom {
		return f, false
	}
	for index, value := range []uint16{s.seg, s.extra, s.left, f.y, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
		if word(0xebd+uint16(index)*2) != value {
			return f, false
		}
	}
	f.width, f.stride, f.rowBytes, f.rows = word(0x11fb), word(0x11f7), word(0x1135), word(0x11ff)
	if f.stride < 2 || f.stride > 40 || f.stride&1 != 0 || f.rowBytes != f.stride*8 || f.width == 0 || f.width > s.right-s.left+1 || f.width > f.rowBytes || f.rows == 0 || f.rows > s.bottom-f.y+1 || word(0x11f3) > 255 {
		return f, false
	}
	f.planar = word(0x11f5) == 0
	if word(0x11f3) == 1 {
		if v.DS() != v.CS() {
			return f, false
		}
		f.offset = 0x1252
	} else {
		if v.DS() != v.Read16(stack+16) || word(0x1201) < f.rowBytes {
			return f, false
		}
		f.offset = word(0x1201) - f.rowBytes
	}
	if uint32(f.offset)+uint32(f.rowBytes) > 65536 {
		return f, false
	}
	f.source = uint32(v.DS())*16 + uint32(f.offset)
	end := f.source + uint32(f.rowBytes)
	if f.source < 0x500 || end > 0xa0000 || (f.source < stack+18 && stack-2 < end) || (f.source < 0x48a10+0x1252 && 0x48a10+0xebd < end) {
		return f, false
	}
	return f, true
}
