package xlate

import "testing"

// holdScene 模擬 HD 換圖：16×8 畫面兩格是一張圖，兩格都畫好才整塊換成 HD（紅色）。
func holdScene(s *flashScreen, step int) {
	if step >= 1 {
		s.ink(0, 7) // 第一幀只畫好左半
	}
	if step >= 2 {
		s.ink(1, 7)
	}
	if step >= 2 {
		s.render(0, 1)
	} else {
		s.render()
	}
}

func runHoldScene(t *testing.T, hold *Hold, region func(cx, cy int) bool) (*FlashAudit, [][]uint8) {
	t.Helper()
	a, s := NewFlashAudit(16, 8, 8), newFlashScreen()
	var outs [][]uint8
	for f, step := range []int{0, 1, 2, 2} {
		holdScene(s, step)
		if hold != nil {
			hold.Apply(s.idx, s.base, s.out, 2, region)
		}
		a.Observe(f+1, s.idx, s.base, s.out, 2)
		outs = append(outs, append([]uint8(nil), s.out...))
	}
	return a, outs
}

func TestHoldHidesHalfDrawnPicture(t *testing.T) {
	all := func(int, int) bool { return true }
	if a, _ := runHoldScene(t, nil, all); a.Count() == 0 {
		t.Fatal("對照組：沒有保留時應量到閃現")
	}
	a, outs := runHoldScene(t, NewHold(16, 8, 2), all)
	if a.Count() != 0 {
		t.Fatalf("保留後不應有閃現：%+v", a.Events())
	}
	if outs[1][3*2*4] != 0 { // 第二幀左半格的墨跡位置要是上一幀的空白，不是原版色號 7
		t.Fatal("半張圖不能原樣呈現")
	}
	if outs[2][0] != 0xF0 {
		t.Fatal("畫完那幀要是 HD")
	}
}

func TestHoldRegionAndLimit(t *testing.T) {
	none := func(int, int) bool { return false }
	if a, _ := runHoldScene(t, NewHold(16, 8, 2), none); a.Count() == 0 {
		t.Fatal("region 外不保留，應量到閃現")
	}
	// 連續變化：每幀換一個色號，保留上限 2 幀之後放行
	h, s := NewHold(16, 8, 2), newFlashScreen()
	shown := 0
	for f := 0; f < 6; f++ {
		s.ink(0, uint8(10+f))
		s.render()
		h.Apply(s.idx, s.base, s.out, 2, func(int, int) bool { return true })
		if s.out[3*2*4] == uint8(10+f) {
			shown++
		}
	}
	if shown < 2 {
		t.Fatalf("連續變化的格子要定期放行，6 幀只顯示 %d 次", shown)
	}
}
