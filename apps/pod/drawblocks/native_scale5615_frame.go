// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
)

type scale5615Frame struct {
	s, d                                         rectangleSurface
	stack                                        uint32
	width, height, factor, color, samples, destY uint16
	x, y                                         int32
}

func readScale5615Frame(v CPUView) (f scale5615Frame, ok bool) {
	if v.CS() != 0x48a1 || v.DS() != v.CS() || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return f, false
	}
	var code [625]byte
	copyView(v, 0x48a10+0x5615, code[:])
	if sha256.Sum256(code[:]) != smallFingerprint(625, "6d9aac63d38f8b0e5ddd9bdbdfcfdda9144318a34398436b6b468c52f0271106").sum {
		return f, false
	}
	stack, good := paletteSourceStack(v, 2, 24)
	if !good || v.BP() != v.SP()+6 {
		return f, false
	}
	f.stack = stack
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	desc := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	var surfaces [2]rectangleSurface
	var roots [2]rectangleSurface
	for i := range surfaces {
		slot := v.Read16(stack + 12 + uint32(i)*2)
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
		surfaces[i], roots[i] = s, root
	}
	f.s, f.d = surfaces[0], surfaces[1]
	f.width, f.height = f.d.right-f.d.left+1, f.d.bottom-f.d.top+1
	f.factor = v.Read16(stack + 22)
	f.color = v.Read16(stack + 20)
	cx, cy := int32(int16(v.Read16(stack+16))), int32(int16(v.Read16(stack+18)))
	if f.width == 0 || f.factor == 0 || f.factor > 8 || cx < -160 || cx > 479 || cy < -100 || cy > 299 {
		return f, false
	}
	f.samples = (f.width - 1 + f.factor - 1) / f.factor
	f.x = cx - int32((f.width-1)/2/f.factor)
	startY := cy - int32((f.height-1)/2/f.factor)
	f.y = int32(int16(word(0x1171)))
	group := f.y - startY
	if group < 0 || group > int32((f.height-1)/f.factor) {
		return f, false
	}
	f.destY = f.d.top + uint16(group)*f.factor
	if v.DX() != f.destY || word(0x1212) != f.samples || word(0x1141) != uint16(f.x-int32(f.s.left)) || word(0x113f) != uint16(startY-int32(f.s.top)) {
		return f, false
	}
	for k, s := range surfaces {
		root := roots[k]
		off := uint16(0xebd + k*36)
		for i, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
			if i == 7 {
				if k == 0 {
					value = f.samples * f.factor
				} else {
					value = f.width - 1
				}
			}
			if word(off+uint16(i)*2) != value {
				return f, false
			}
		}
	}
	return f, true
}
