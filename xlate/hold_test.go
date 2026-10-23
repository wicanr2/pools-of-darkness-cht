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

func runHoldScene(t *testing.T, hold *Hold, region func(cx, cy int, sum uint64) ([]uint8, bool)) (*FlashAudit, [][]uint8) {
	t.Helper()
	a, s := NewFlashAudit(16, 8, 8), newFlashScreen()
	var outs [][]uint8
	for f, step := range []int{0, 1, 2, 2} {
		holdScene(s, step)
		if hold != nil {
			hold.Commit()
			hold.Apply(s.idx, s.base, s.out, 2, region)
		}
		a.Observe(f+1, s.idx, s.base, s.out, 2)
		outs = append(outs, append([]uint8(nil), s.out...))
	}
	return a, outs
}

func TestHoldHidesHalfDrawnPicture(t *testing.T) {
	all := func(int, int, uint64) ([]uint8, bool) { return nil, true }
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
	none := func(int, int, uint64) ([]uint8, bool) { return nil, false }
	if a, _ := runHoldScene(t, NewHold(16, 8, 2), none); a.Count() == 0 {
		t.Fatal("region 外不保留，應量到閃現")
	}
	// 連續變化：每幀換一個色號，保留上限 2 幀之後放行
	h, s := NewHold(16, 8, 2), newFlashScreen()
	shown := 0
	for f := 0; f < 6; f++ {
		s.ink(0, uint8(10+f))
		s.render()
		h.Commit()
		h.Apply(s.idx, s.base, s.out, 2, func(int, int, uint64) ([]uint8, bool) { return nil, true })
		if s.out[3*2*4] == uint8(10+f) {
			shown++
		}
	}
	if shown < 2 {
		t.Fatalf("連續變化的格子要定期放行，6 幀只顯示 %d 次", shown)
	}
}

// 原版畫到一半停住幾幀：已開始保留的格持續保留，不因「這幀沒變」就露出半張圖。
func TestHoldSurvivesStall(t *testing.T) {
	a, s, h := NewFlashAudit(16, 8, 8), newFlashScreen(), NewHold(16, 8, 8)
	all := func(int, int, uint64) ([]uint8, bool) { return nil, true }
	for f, step := range []int{0, 1, 1, 1, 2} {
		holdScene(s, step)
		h.Commit()
		h.Apply(s.idx, s.base, s.out, 2, all)
		holdScene(s, step) // 同一幀再合成一次（新的 out），結果與狀態不變
		h.Apply(s.idx, s.base, s.out, 2, all)
		a.Observe(f+1, s.idx, s.base, s.out, 2)
	}
	if a.Count() != 0 {
		t.Fatalf("停住期間露出半張圖：%+v", a.Events())
	}
}

// 呼叫端給出確定的 HD 格：直接畫，不受保留上限限制。
func TestHoldPaintsKnownCellWithoutLimit(t *testing.T) {
	h, s := NewHold(16, 8, 1), newFlashScreen()
	hd := make([]uint8, 16*16*4)
	for i := range hd {
		hd[i] = 0xAB
	}
	known := func(cx, cy int, sum uint64) ([]uint8, bool) { return hd, cx == 0 }
	s.ink(0, 7)
	for f := 0; f < 5; f++ { // 遠超過上限 1 幀
		s.render()
		h.Commit()
		h.Apply(s.idx, s.base, s.out, 2, known)
		if s.out[0] != 0xAB {
			t.Fatalf("第 %d 幀：確定的 HD 格沒畫", f)
		}
	}
}
