// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
)

type flip5886Frame struct {
	s                             rectangleSurface
	stack                         uint32
	width, remaining, top, bottom uint16
	horizontal, vertical          bool
}

var flip5886Code = smallFingerprint(418, "2ff20b76959b7f3972f76530c6142c8337d0f935893752f9a23d417dde2eac36").sum

func readFlip5886Frame(v CPUView) (f flip5886Frame, ok bool) {
	if v.CS() != 0x48a1 || v.DS() != v.CS() || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return f, false
	}
	var code [418]byte
	copyView(v, 0x48a10+0x5886, code[:])
	if sha256.Sum256(code[:]) != flip5886Code {
		return f, false
	}
	stack, good := paletteSourceStack(v, 4, 18)
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
	f.horizontal, f.vertical = v.Read16(stack+14) != 0, v.Read16(stack+16) != 0
	if !f.horizontal && !f.vertical {
		return f, false
	}
	f.s = s
	f.width = s.right - s.left + 1
	rows := s.bottom - s.top + 1
	if f.vertical {
		rows /= 2
		if rows == 0 {
			rows = 1
		}
	}
	f.remaining = word(0x1171)
	if f.remaining == 0 || f.remaining > rows {
		return f, false
	}
	done := rows - f.remaining
	f.top, f.bottom = s.top+done, s.bottom-done
	for i, value := range []uint16{s.seg, s.extra, s.left, f.top, s.right, f.bottom, s.flags, f.width, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
		if word(0xebd+uint16(i)*2) != value {
			return f, false
		}
	}
	return f, true
}
