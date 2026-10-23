package xlate

// Hold 保留呈現（spec 202 §2.8，DRAFT）：原版正在畫、還沒被替換的格子，這一幀先顯示上一幀的呈現。
// 用途是 HD 換圖：整塊原版吻合才換圖，原版分幾幀畫完一張圖時，中間的半張圖不能原樣呈現。
//
// 同一幀可以 Apply 多次（畫面、量尺、截圖各合成一次），結果相同；Commit 在下一幀開始時把這一幀最後一次
// Apply 的結果定為「上一幀」。
type Hold struct {
	w, h, max int
	// 上一幀已確定的呈現、各格指紋與連續保留次數
	prev []uint8
	sums []uint64
	held []int
	// 這一幀最後一次 Apply 的結果，Commit 時才生效
	cur     []uint8
	curSums []uint64
	curHeld []int
	pending bool
}

// NewHold 建立保留層；w、h 是原版畫面大小（8 的倍數），max 是一格連續保留的上限（幀）。
func NewHold(w, h, max int) *Hold {
	n := (w / flashCellPx) * (h / flashCellPx)
	return &Hold{w: w, h: h, max: max, sums: make([]uint64, n), held: make([]int, n),
		curSums: make([]uint64, n), curHeld: make([]int, n)}
}

// Apply 就地改寫 out：indexed 是原版色號，base 是原版放大 scale 倍的 RGBA，out 是同一幀替換後的 RGBA。
// region 以格座標與格內色號指紋（cellContent）判斷這格是否可能保留，例如「內容等於某張 HD 圖的對應格」。
// 一格開始保留後，只要仍未被替換、仍在 region 內，就持續保留到 max 幀，不要求每幀都有變化
// （原版畫到一半停住時，半張圖也不能露出）。
//
// region 回 (hd, true) 且 hd 不是 nil 時，這格直接畫 hd（放大後這格的 RGBA，逐列連續），不計保留次數也沒有上限：
// 呼叫端已確定這格就是某張 HD 圖的這一格。hd 是 nil 時照上述規則保留上一幀。
func (h *Hold) Apply(indexed, base, out []uint8, scale int, region func(cx, cy int, sum uint64) ([]uint8, bool)) {
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
			h.curSums[c], h.curHeld[c] = sum, 0
			changed := sum != h.sums[c]
			if !ink || !cellSame(base, out, h.w, cx, cy, scale) {
				continue
			}
			hd, ok := region(cx, cy, sum)
			if ok && len(hd) == n*flashCellPx*scale {
				for y, k := cy*flashCellPx*scale, 0; y < (cy+1)*flashCellPx*scale; y, k = y+1, k+n {
					o := y*stride + cx*n
					copy(out[o:o+n], hd[k:k+n])
				}
				continue
			}
			if havePrev && ok && (changed || h.held[c] > 0) && h.held[c] < h.max {
				for y := cy * flashCellPx * scale; y < (cy+1)*flashCellPx*scale; y++ {
					o := y*stride + cx*n
					copy(out[o:o+n], h.prev[o:o+n])
				}
				h.curHeld[c] = h.held[c] + 1
			}
		}
	}
	h.cur = append(h.cur[:0], out...)
	h.pending = true
}

// Commit 在新的一幀開始時呼叫：上一幀最後一次 Apply 的結果成為比較基準。沒有 Apply 過就不變。
func (h *Hold) Commit() {
	if !h.pending {
		return
	}
	h.prev = append(h.prev[:0], h.cur...)
	copy(h.sums, h.curSums)
	copy(h.held, h.curHeld)
	h.pending = false
}

// CellContent 回第 (cx, cy) 個 8×8 格有沒有墨跡與色號指紋，與 Hold 傳給 region 的 sum 相同算法；w 是畫面寬。
func CellContent(indexed []uint8, w, cx, cy int) (ink bool, sum uint64) { return cellContent(indexed, w, cx, cy) }
