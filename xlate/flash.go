package xlate

// FlashAudit 量測「原版內容先原樣呈現、之後才被疊字或 HD 圖換掉」的閃現（spec 202 §2.6）。
//
// 只比對呈現結果與原版畫面，不看疊字層的內部狀態：以 8×8 原版像素為一格，某格有墨跡且原樣呈現，
// 窗內同樣的原版內容又被改掉，就是玩家先看到原版、再看到替換結果。
type FlashAudit struct {
	w, h, window int
	cells        []flashCell
	events       []FlashEvent
}

type flashCell struct {
	shown bool
	frame int
	sum   uint64
}

// FlashEvent 是一次閃現：Shown 畫格原樣呈現，Frame 畫格被改；CX、CY 是格座標（原版像素 ÷8）。
type FlashEvent struct {
	Frame, Shown int
	CX, CY       int
}

const flashCellPx = 8

// NewFlashAudit 建立量尺；w、h 是原版畫面大小（8 的倍數），window 是原樣呈現到被改之間最多幾個畫格。
func NewFlashAudit(w, h, window int) *FlashAudit {
	return &FlashAudit{w: w, h: h, window: window, cells: make([]flashCell, (w/flashCellPx)*(h/flashCellPx))}
}

// Observe 比對一個畫格。indexed 是原版色號；base 是原版放大 scale 倍的 RGBA，out 是同一畫格疊字、HD 合成後的 RGBA。
func (a *FlashAudit) Observe(frame int, indexed, base, out []uint8, scale int) {
	if scale <= 0 || len(indexed) < a.w*a.h || len(base) < a.w*a.h*scale*scale*4 || len(out) != len(base) {
		return
	}
	cols := a.w / flashCellPx
	for cy := 0; cy < a.h/flashCellPx; cy++ {
		for cx := 0; cx < cols; cx++ {
			c := &a.cells[cy*cols+cx]
			ink, sum := cellContent(indexed, a.w, cx, cy)
			same := cellSame(base, out, a.w, cx, cy, scale)
			switch {
			case same && ink:
				c.shown, c.frame, c.sum = true, frame, sum
			case !same:
				if c.shown && c.sum == sum && frame > c.frame && frame-c.frame <= a.window {
					a.events = append(a.events, FlashEvent{Frame: frame, Shown: c.frame, CX: cx, CY: cy})
				}
				c.shown = false
			}
		}
	}
}

// cellContent 回第 (cx, cy) 格有沒有墨跡（兩種以上色號）與色號的 FNV-1a 64 指紋；w 是畫面寬。
func cellContent(indexed []uint8, w, cx, cy int) (bool, uint64) {
	sum := uint64(14695981039346656037)
	first, ink := indexed[cy*flashCellPx*w+cx*flashCellPx], false
	for y := 0; y < flashCellPx; y++ {
		row := indexed[(cy*flashCellPx+y)*w+cx*flashCellPx:]
		for x := 0; x < flashCellPx; x++ {
			v := row[x]
			ink = ink || v != first
			sum = (sum ^ uint64(v)) * 1099511628211
		}
	}
	return ink, sum
}

// cellSame 回 out 與 base 在第 (cx, cy) 格的放大像素是否全部相同；w 是原版畫面寬。
func cellSame(base, out []uint8, w, cx, cy, scale int) bool {
	stride := w * scale * 4
	n := flashCellPx * scale * 4
	for y := cy * flashCellPx * scale; y < (cy+1)*flashCellPx*scale; y++ {
		o := y*stride + cx*n
		b, q := base[o:o+n], out[o:o+n]
		for i := range b {
			if b[i] != q[i] {
				return false
			}
		}
	}
	return true
}

// Events 回目前記到的閃現。
func (a *FlashAudit) Events() []FlashEvent { return a.events }

// Count 回閃現次數。
func (a *FlashAudit) Count() int { return len(a.events) }
