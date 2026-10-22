package xlate

import "testing"

// ghostSetup 回 16×8 畫面：第 0 格有墨跡（錨定格）、第 1 格純色；一筆兩格的疊字已定色顯示。
func ghostSetup(t *testing.T) (l *Layer, s *Stamp, idx, rgb []uint8) {
	t.Helper()
	idx, rgb = make([]uint8, 16*8), make([]uint8, 3*16*8)
	idx[2*16+3] = 7
	l = &Layer{W: 16, H: 8}
	s = &Stamp{Key: "k", X: 0, Y: 0, Cells: 2, CellW: 8, CellH: 8, State: Pending}
	l.Stamps = []*Stamp{s}
	l.Frame(idx, rgb)
	if s.State != Shown {
		t.Fatal("定色後應該顯示")
	}
	s.BG = [3]uint8{200, 0, 0} // 用醒目的背景色判斷有沒有畫
	return l, s, idx, rgb
}

// drawn 回第 1 格（純色、非錨定格）左上像素是不是疊字背景色。
func drawn(l *Layer) bool {
	dst := make([]uint8, 16*8*4)
	l.Draw(dst, 1, nil)
	return dst[8*4] == 200
}

func TestGhostDrawsRemovedStampUntilNextFrame(t *testing.T) {
	l, _, idx, rgb := ghostSetup(t)
	l.Stamps = nil // 呼叫端移除（呼叫外的顯存寫入等）
	if !drawn(l) {
		t.Fatal("上一幀仍有效的疊字被移除後，同一幀要照畫")
	}
	l.Frame(idx, rgb) // 原版內容不變：轉殘影且有效
	if !drawn(l) {
		t.Fatal("原版墨跡與定色時相同，殘影要畫")
	}
}

func TestGhostFollowsOriginalContent(t *testing.T) {
	l, _, idx, rgb := ghostSetup(t)
	l.Stamps = nil
	other := append([]uint8(nil), idx...)
	other[2*16+3] = 9 // 錨定格內容不同
	other[2*16+11] = 9
	l.Frame(other, rgb)
	if drawn(l) {
		t.Fatal("錨定格不同不能畫")
	}
	l.Frame(idx, rgb) // 同一段原文回到原位（原版還原顯存）
	if !drawn(l) {
		t.Fatal("同樣墨跡回來要畫")
	}
	blank := make([]uint8, 16*8) // 整塊清成純色：非錨定格仍相同，不足以畫
	l.Frame(blank, rgb)
	if drawn(l) {
		t.Fatal("只有非錨定格相同不能畫")
	}
}

func TestGhostForgetAndAge(t *testing.T) {
	l, _, idx, rgb := ghostSetup(t)
	l.Stamps = nil
	l.Frame(idx, rgb)
	l.ForgetGhosts()
	if drawn(l) {
		t.Fatal("ForgetGhosts 之後不能畫")
	}

	l, _, idx, rgb = ghostSetup(t)
	l.Stamps = nil
	for i := 0; i <= GhostFrames; i++ {
		l.Frame(idx, rgb)
	}
	if drawn(l) {
		t.Fatalf("超過 %d 次 Frame 不能畫", GhostFrames)
	}
}

func TestGhostSkipsStampAddedBack(t *testing.T) {
	l, s, idx, rgb := ghostSetup(t)
	l.Stamps = nil
	l.Frame(idx, rgb)
	l.Stamps = []*Stamp{s}
	if got := l.ghostsToDraw(); len(got) != 0 {
		t.Fatalf("加回的疊字不該再當殘影畫：%d 筆", len(got))
	}
}

// 呼叫端自行管理（Frozen）的疊字沒有 Layer 算的指紋，殘影以最後一幀的畫面為準。
func TestGhostCoversFrozenStamp(t *testing.T) {
	idx, rgb := make([]uint8, 16*8), make([]uint8, 3*16*8)
	idx[2*16+3] = 7
	s := &Stamp{Key: "b", X: 0, Y: 0, Cells: 2, CellW: 8, CellH: 8, State: Shown, BG: [3]uint8{200, 0, 0}}
	l := &Layer{W: 16, H: 8, Stamps: []*Stamp{s}, Frozen: func(*Stamp) bool { return true }}
	l.Frame(idx, rgb)
	l.Stamps = nil
	blank := make([]uint8, 16*8)
	l.Frame(blank, rgb) // 原版清掉這段
	if drawn(l) {
		t.Fatal("原文不在畫面上不能畫")
	}
	l.Frame(idx, rgb) // 原版還原同一段
	if !drawn(l) {
		t.Fatal("同一段原文回來要畫")
	}
}
