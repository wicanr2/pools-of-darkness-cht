// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"math/bits"
)

type dissolve4337Frame struct {
	s, d                                    rectangleSurface
	stack                                   uint32
	width, height, remaining, period, hbits uint16
	state, poly, mask                       uint32
}

var dissolve4337Code = smallFingerprint(681, "6888a66a54dd072b40f5c95f990d81d53f75257852bce465c95dff15e1168796").sum

func readDissolve4337Frame(v CPUView) (f dissolve4337Frame, ok bool) {
	if v.CS() != 0x48a1 || v.DS() != v.CS() || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
		return f, false
	}
	var code [681]byte
	copyView(v, 0x48a10+0x4337, code[:])
	if sha256.Sum256(code[:]) != dissolve4337Code {
		return f, false
	}
	stack, good := paletteSourceStack(v, 4, 16)
	if !good || v.BP() != v.SP()+6 {
		return f, false
	}
	f.stack = stack
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	desc := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	for k := 0; k < 2; k++ {
		slot := v.Read16(stack + 12 + uint32(k)*2)
		if slot > 255 {
			return f, false
		}
		s := desc(slot * 2)
		root := s
		links := 0
		for root.flags&64 != 0 {
			if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
				return f, false
			}
			root = desc(root.seg)
			links++
		}
		if root.seg != 0xa000 || root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 || s.left > s.right || s.right > 319 || s.top > s.bottom || s.bottom > 199 {
			return f, false
		}
		for i, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
			if word(uint16(0xebd+k*36)+uint16(i)*2) != value {
				return f, false
			}
		}
		if k == 0 {
			f.s = s
		} else {
			f.d = s
		}
	}
	f.width, f.height = f.s.right-f.s.left+1, f.s.bottom-f.s.top+1
	if dw := f.d.right - f.d.left + 1; dw < f.width {
		f.width = dw
	}
	if dh := f.d.bottom - f.d.top + 1; dh < f.height {
		f.height = dh
	}
	area := f.width * f.height
	f.remaining = word(0x1171)
	if f.remaining == 0 || f.remaining > area || word(0x1133) != f.height || word(0x1135) != f.width {
		return f, false
	}
	f.period = area / 140
	if word(0x1175) != f.period {
		return f, false
	}
	f.hbits = uint16(bits.Len16(f.height))
	degree := f.hbits + uint16(bits.Len16(f.width))
	f.mask = (1 << f.hbits) - 1
	if word(0x1139) != f.hbits || uint32(word(0x1165)) != f.mask {
		return f, false
	}
	for n, off := range []uint32{0x42b3, 0x42d4, 0x42f5, 0x4316} {
		f.poly |= uint32(v.Read8(0x48a10+off+uint32(degree))) << uint(8*n)
	}
	if word(0x116b) != uint16(f.poly) || word(0x116d) != uint16(f.poly>>16) {
		return f, false
	}
	f.state = uint32(word(0x1167)) | uint32(word(0x1169))<<16
	if f.state == 0 || f.state >= uint32(1)<<degree {
		return f, false
	}
	accepted := area - f.remaining
	if v.IP() == 0x45c4 {
		if accepted == 0 {
			return f, false
		}
		accepted--
	}
	counter := uint16(0) - accepted
	if f.period > 0 {
		counter = f.period - accepted%f.period
	}
	if v.IP() == 0x45ce {
		counter = 0
	}
	if word(0x1177) != counter {
		return f, false
	}
	return f, true
}

// This routine has no CLD; arithmetic preserves its incoming DF.
func dissolve4337CompareFlags(before, a, b uint16) uint16 {
	return nativeCompareFlags(before, a, b) | before&cpu.DF
}
