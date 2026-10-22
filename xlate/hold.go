package xlate

// Hold 保留呈現（spec 202 §2.8，DRAFT）：原版正在畫、還沒被替換的格子，這一幀先顯示上一次的呈現。
// 用途是 HD 換圖：整塊原版吻合才換圖，原版分幾幀畫完一張圖時，中間的半張圖不能原樣呈現。
type Hold struct {
	w, h, max int
	prev      []uint8 // 上一次 Apply 結束時的 out
	sums      []uint64
	held      []int
}

// NewHold 建立保留層；w、h 是原版畫面大小（8 的倍數），max 是一格連續保留的上限（幀）。
func NewHold(w, h, max int) *Hold {
	n := (w / flashCellPx) * (h / flashCellPx)
	return &Hold{w: w, h: h, max: max, sums: make([]uint64, n), held: make([]int, n)}
}

// Apply 就地改寫 out：indexed 是原版色號，base 是原版放大 scale 倍的 RGBA，out 是同一幀替換後的 RGBA。
// region 回 true 的格才可能保留。
func (h *Hold) Apply(indexed, base, out []uint8, scale int, region func(cx, cy int) bool) {
	if scale <= 0 || len(indexed) < h.w*h.h || len(base) != h.w*h.h*scale*scale*4 || len(out) != len(base) {
		return
	}
	havePrev := len(h.prev) == len(out)
	cols := h.w / flashCellPx
	n := flashCellPx * scale * 4
	stride := h.w * scale * 4
	for cy := 0; cy < h.h/flashCellPx; cy++ {
		for cx := 0; cx < cols; cx++ {
			c := cy*cols + cx
			ink, sum := cellContent(indexed, h.w, cx, cy)
			changed := sum != h.sums[c]
			h.sums[c] = sum
			if havePrev && changed && ink && h.held[c] < h.max && region(cx, cy) && cellSame(base, out, h.w, cx, cy, scale) {
				for y := cy * flashCellPx * scale; y < (cy+1)*flashCellPx*scale; y++ {
					o := y*stride + cx*n
					copy(out[o:o+n], h.prev[o:o+n])
				}
				h.held[c]++
				continue
			}
			h.held[c] = 0
		}
	}
	h.prev = append(h.prev[:0], out...)
}
