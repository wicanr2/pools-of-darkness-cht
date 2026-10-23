package xlate

import "testing"

// latchStep：用 flashScreen（16×8、兩格、2 倍）模擬一幀；cover 的格被圖層換成紅色。
func latchStep(l *Latch, s *flashScreen, pal uint64, cover ...int) {
	s.render(cover...)
	l.Tick()
	l.Apply(s.idx, pal, s.base, s.out, 2)
}

func red(s *flashScreen, cx int) bool { return s.out[(cx*16)*4] == 0xF0 }

func TestLatchKeepsHDWhileOriginalUnchanged(t *testing.T) {
	l, s := NewLatch(16, 8, 4, 600), newFlashScreen()
	s.ink(0, 7)
	s.ink(1, 7)
	latchStep(l, s, 1, 0, 1) // 兩格都被換成 HD
	latchStep(l, s, 1, 1)    // 圖層這幀只認出第 1 格（第 0 格那件被蓋到別處）
	if !red(s, 0) {
		t.Fatal("原版沒變的格要沿用 HD")
	}
	s.ink(0, 9) // 原版這格真的變了
	latchStep(l, s, 1)
	if red(s, 0) {
		t.Fatal("原版變了不能沿用")
	}
	if !red(s, 1) {
		t.Fatal("第 1 格原版沒變，應沿用")
	}
	latchStep(l, s, 2) // 色盤變了
	if red(s, 1) {
		t.Fatal("色盤變了不能沿用")
	}
}

func TestLatchAlternatingContentAndExpiry(t *testing.T) {
	l, s := NewLatch(16, 8, 4, 3), newFlashScreen()
	s.ink(0, 7)
	latchStep(l, s, 1, 0)
	s.ink(0, 9)
	latchStep(l, s, 1, 0) // 第二種內容也被換過
	s.ink(0, 7)
	latchStep(l, s, 1)
	if !red(s, 0) {
		t.Fatal("回到第一種內容要取回它的 HD")
	}
	for i := 0; i < 4; i++ { // 內容不同、不用第一筆，超過 3 幀移除
		s.ink(0, 11)
		latchStep(l, s, 1)
	}
	s.ink(0, 7)
	latchStep(l, s, 1)
	if red(s, 0) {
		t.Fatal("超過保留幀數不能沿用")
	}
	latchStep(l, s, 1, 0)
	l.Forget()
	latchStep(l, s, 1)
	if red(s, 0) {
		t.Fatal("Forget 後不能沿用")
	}
}
