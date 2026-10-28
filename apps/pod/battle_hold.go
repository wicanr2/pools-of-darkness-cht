package pod

import "github.com/wicanr2/dosgolem/xlate"

// 主專案規格010／研究017：戰場像素可能先於色盤，怪物也可能跨畫格畫完。
// 規格010四格修正：明確按D的6619–6622局部繪製需四格，6623完整。
// 未知來源不等待，不更動原版執行；第五格仍強制回退。
const battleHoldFrames = 4

type battleHold struct {
	prev, current                 []byte
	prevIndex, currentIndex       []byte
	prevPalette, currentPalette   uint64
	prevHeld, currentHeld         [21 * 21]uint8
	pending                       bool
	prevOriginal, currentOriginal *xlate.IndexedSource
}

func (h *battleHold) commit() {
	if h == nil || !h.pending {
		return
	}
	h.prev, h.current = h.current, h.prev[:0]
	h.prevIndex, h.currentIndex = h.currentIndex, h.prevIndex[:0]
	h.prevPalette, h.prevHeld = h.currentPalette, h.currentHeld
	h.prevOriginal = h.currentOriginal
	h.pending = false
}

func (h *battleHold) apply(indexed, base, out []byte, palette uint64, eligible [21 * 21]bool) {
	h.applySource(indexed, base, out, palette, eligible, nil)
}
func (h *battleHold) applyDeferred(indexed, out []byte, palette uint64, eligible [21 * 21]bool, source *xlate.IndexedSource) {
	h.applySource(indexed, nil, out, palette, eligible, source)
}
func (h *battleHold) applySource(indexed, base, out []byte, palette uint64, eligible [21 * 21]bool, source *xlate.IndexedSource) {
	if h == nil || len(indexed) != 64000 || len(out) != 960*600*4 || source == nil && len(base) != len(out) {
		return
	}
	clear(h.currentHeld[:])
	if len(h.prev) == len(out) && len(h.prevIndex) == len(indexed) && palette == h.prevPalette {
		for cy := 0; cy < 21; cy++ {
			for cx := 0; cx < 21; cx++ {
				n := cy*21 + cx
				if !eligible[n] || h.prevHeld[n] >= battleHoldFrames {
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
				if source != nil {
					original = source.CellIsOriginal(out, x/8, y/8)
				}
				for row := 0; row < 24 && original && source == nil; row++ {
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
				if h.prevOriginal != nil {
					h.prevOriginal.CopyCellTo(out, h.prev, x/8, y/8)
				}
				for row := 0; row < 24 && h.prevOriginal == nil; row++ {
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
	h.currentOriginal = source
}
