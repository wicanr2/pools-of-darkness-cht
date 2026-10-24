package pod

// 主專案規格010／研究017：戰場像素可能先於色盤，怪物也可能跨畫格畫完。
// 已證實的6690–6692局部繪製最多保留三格；未知來源不等待，不更動原版執行。
type battleHold struct {
	prev, current               []byte
	prevIndex, currentIndex     []byte
	prevPalette, currentPalette uint64
	prevHeld, currentHeld       [21 * 21]uint8
	pending                     bool
}

func (h *battleHold) commit() {
	if h == nil || !h.pending {
		return
	}
	h.prev, h.current = h.current, h.prev[:0]
	h.prevIndex, h.currentIndex = h.currentIndex, h.prevIndex[:0]
	h.prevPalette, h.prevHeld = h.currentPalette, h.currentHeld
	h.pending = false
}

func (h *battleHold) apply(indexed, base, out []byte, palette uint64, eligible [21 * 21]bool) {
	if h == nil || len(indexed) != 64000 || len(base) != 960*600*4 || len(out) != len(base) {
		return
	}
	clear(h.currentHeld[:])
	if len(h.prev) == len(out) && len(h.prevIndex) == len(indexed) && palette == h.prevPalette {
		for cy := 0; cy < 21; cy++ {
			for cx := 0; cx < 21; cx++ {
				n := cy*21 + cx
				if !eligible[n] || h.prevHeld[n] >= 3 {
					continue
				}
				x, y := 8+cx*8, 8+cy*8
				changed, original := false, true
				for row := 0; row < 8; row++ {
					for col := 0; col < 8; col++ {
						o := (y+row)*320 + x + col
						changed = changed || indexed[o] != h.prevIndex[o]
					}
				}
				if !changed && h.prevHeld[n] == 0 {
					continue
				}
				for row := 0; row < 24 && original; row++ {
					o := ((y*3+row)*960 + x*3) * 4
					for k := 0; k < 24*4; k++ {
						if base[o+k] != out[o+k] {
							original = false
							break
						}
					}
				}
				if !original {
					continue
				}
				for row := 0; row < 24; row++ {
					o := ((y*3+row)*960 + x*3) * 4
					copy(out[o:o+24*4], h.prev[o:o+24*4])
				}
				h.currentHeld[n] = h.prevHeld[n] + 1
			}
		}
	}
	h.current = append(h.current[:0], out...)
	h.currentIndex = append(h.currentIndex[:0], indexed...)
	h.currentPalette, h.pending = palette, true
}
