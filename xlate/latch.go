package xlate

// Latch 是 HD 鎖存（spec 202 §2.9）：某格原版內容與色盤都和它上次被 HD 替換時相同，就沿用那次的 HD 畫面。
// 圖層逐幀整件比對，一件被蓋到一小塊就整件退回原版；鎖存讓沒被改到的格維持 HD。
type Latch struct {
	w, h, keep, frames int
	cells              [][]latchEntry
}

type latchEntry struct {
	sum, pal uint64
	scale    int
	pix      []uint8 // 這格放大後的 RGBA，逐列連續
	idle     int     // 連續沒用到的幀數
}

// NewLatch 建立鎖存；w、h 是原版畫面大小（8 的倍數），每格最多記 keep 筆，超過 frames 幀沒用到就移除。
func NewLatch(w, h, keep, frames int) *Latch {
	return &Latch{w: w, h: h, keep: keep, frames: frames, cells: make([][]latchEntry, (w/flashCellPx)*(h/flashCellPx))}
}

// Apply 就地改寫 out：indexed 是原版色號，pal 是色盤指紋，base 是原版放大 scale 倍的 RGBA，out 是圖層合成後的 RGBA。
func (l *Latch) Apply(indexed []uint8, pal uint64, base, out []uint8, scale int) {
	if scale <= 0 || len(indexed) < l.w*l.h || len(base) != l.w*l.h*scale*scale*4 || len(out) != len(base) {
		return
	}
	cols := l.w / flashCellPx
	n := flashCellPx * scale * 4
	stride := l.w * scale * 4
	for cy := 0; cy < l.h/flashCellPx; cy++ {
		for cx := 0; cx < cols; cx++ {
			ink, sum := cellContent(indexed, l.w, cx, cy)
			if !ink {
				continue
			}
			c := cy*cols + cx
			found := -1
			for i, e := range l.cells[c] {
				if e.sum == sum && e.pal == pal && e.scale == scale {
					found = i
					break
				}
			}
			if !cellSame(base, out, l.w, cx, cy, scale) {
				if found >= 0 && l.same(found, c, out, cx, cy, scale) {
					l.cells[c][found].idle = 0 // 同樣的 HD 畫面：不重新配置
					continue
				}
				pix := make([]uint8, 0, n*flashCellPx*scale)
				for y := cy * flashCellPx * scale; y < (cy+1)*flashCellPx*scale; y++ {
					o := y*stride + cx*n
					pix = append(pix, out[o:o+n]...)
				}
				e := latchEntry{sum: sum, pal: pal, scale: scale, pix: pix}
				list := l.cells[c]
				if found >= 0 {
					list = append(list[:found], list[found+1:]...)
				}
				list = append([]latchEntry{e}, list...)
				if len(list) > l.keep {
					list = list[:l.keep]
				}
				l.cells[c] = list
				continue
			}
			if found < 0 {
				continue
			}
			e := &l.cells[c][found]
			e.idle = 0
			for y, k := cy*flashCellPx*scale, 0; y < (cy+1)*flashCellPx*scale; y, k = y+1, k+n {
				o := y*stride + cx*n
				copy(out[o:o+n], e.pix[k:k+n])
			}
		}
	}
}

// same 回第 c 格第 i 筆記錄的畫面是否與 out 這格相同。
func (l *Latch) same(i, c int, out []uint8, cx, cy, scale int) bool {
	n := flashCellPx * scale * 4
	stride := l.w * scale * 4
	pix := l.cells[c][i].pix
	for y, k := cy*flashCellPx*scale, 0; y < (cy+1)*flashCellPx*scale; y, k = y+1, k+n {
		o := y*stride + cx*n
		for j := 0; j < n; j++ {
			if out[o+j] != pix[k+j] {
				return false
			}
		}
	}
	return true
}

// Tick 每幀呼叫一次：記錄的閒置幀數加一，超過 frames 的移除。
func (l *Latch) Tick() {
	for c, list := range l.cells {
		keep := list[:0]
		for _, e := range list {
			if e.idle++; e.idle <= l.frames {
				keep = append(keep, e)
			}
		}
		l.cells[c] = keep
	}
}

// Forget 清空全部記錄。
func (l *Latch) Forget() {
	for c := range l.cells {
		l.cells[c] = nil
	}
}
