package drawblocks

func validEllipseFrame(v CPUView) bool {
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 {
		return false
	}
	stack, ok := paletteSourceStack(v, 2, 24)
	if !ok || v.BP() != v.SP()+6 {
		return false
	}
	slot := v.Read16(stack + 12)
	if slot > 255 {
		return false
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	s := descriptor(slot * 2)
	root := s
	links := 0
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return false
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left != 0 || root.top != 0 || root.right != 319 || root.bottom != 199 || s.left > s.right || s.top > s.bottom || s.right > 319 || s.bottom > 199 {
		return false
	}
	for index, value := range []uint16{s.seg, s.extra, s.left, s.top, s.right, s.bottom, s.flags, s.right - s.left + 1, s.bottom - s.top + 1, root.seg, root.extra, root.left, root.top, root.right, root.bottom, root.flags, 320, 200} {
		if word(0xebd+uint16(index)*2) != value {
			return false
		}
	}
	cx, cy, rx, ry, color := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18), v.Read16(stack+20), v.Read16(stack+22)
	return word(0x112d) == cx && word(0x1131) == cy && word(0x1129) == uint16(byte(color))*257 && rx > 0 && rx <= 160 && ry > 0 && ry <= 100 && int16(cx) >= -160 && int16(cx) <= 479 && int16(cy) >= -100 && int16(cy) <= 299
}
