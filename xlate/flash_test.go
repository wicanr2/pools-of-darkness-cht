package xlate

import "testing"

// flashScreen 是 16×8 原版畫面（兩格）與放大 2 倍的 RGBA；色號 v 的 RGB 為 (v, v, v)。
type flashScreen struct{ idx, base, out []uint8 }

func newFlashScreen() *flashScreen {
	return &flashScreen{idx: make([]uint8, 16*8), base: make([]uint8, 32*16*4), out: make([]uint8, 32*16*4)}
}

// ink 在第 cx 格畫一個色號 v 的點（有墨跡）。
func (s *flashScreen) ink(cx int, v uint8) { s.idx[cx*8+3] = v }

// render 依 idx 重畫 base，out 先等於 base；cover 列出的格在 out 改成紅色（被疊字或 HD 換掉）。
func (s *flashScreen) render(cover ...int) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			v := s.idx[(y/2)*16+x/2]
			o := (y*32 + x) * 4
			s.base[o], s.base[o+1], s.base[o+2], s.base[o+3] = v, v, v, 0xFF
		}
	}
	copy(s.out, s.base)
	for _, cx := range cover {
		for y := 0; y < 16; y++ {
			for x := cx * 16; x < cx*16+16; x++ {
				s.out[(y*32+x)*4] = 0xF0
			}
		}
	}
}

func TestFlashAuditCountsShownThenReplaced(t *testing.T) {
	a, s := NewFlashAudit(16, 8, 8), newFlashScreen()
	s.ink(0, 7)
	s.render()
	a.Observe(1, s.idx, s.base, s.out, 2) // 原版字原樣呈現
	s.render(0)
	a.Observe(2, s.idx, s.base, s.out, 2) // 同樣的原版字被譯文蓋住
	a.Observe(3, s.idx, s.base, s.out, 2) // 持續蓋住不再記
	if a.Count() != 1 || a.Events()[0] != (FlashEvent{Frame: 2, Shown: 1, CX: 0, CY: 0}) {
		t.Fatalf("要記一次第 0 格，得到 %+v", a.Events())
	}
}

func TestFlashAuditIgnoresOtherContentWindowAndBlank(t *testing.T) {
	a, s := NewFlashAudit(16, 8, 8), newFlashScreen()
	s.ink(0, 7)
	s.render()
	a.Observe(1, s.idx, s.base, s.out, 2)
	s.ink(0, 9) // 原版內容換了（動畫、換畫面）才被蓋
	s.render(0)
	a.Observe(2, s.idx, s.base, s.out, 2)
	if a.Count() != 0 {
		t.Fatalf("內容不同不應記：%+v", a.Events())
	}

	s.render()
	a.Observe(3, s.idx, s.base, s.out, 2)
	s.render(0)
	a.Observe(12, s.idx, s.base, s.out, 2) // 超過 8 格的窗
	if a.Count() != 0 {
		t.Fatalf("超過窗不應記：%+v", a.Events())
	}

	s.render() // 第 1 格沒有墨跡（整格同色）
	a.Observe(13, s.idx, s.base, s.out, 2)
	s.render(1)
	a.Observe(14, s.idx, s.base, s.out, 2)
	if a.Count() != 0 {
		t.Fatalf("沒有墨跡不應記：%+v", a.Events())
	}

	s.render() // 反向對照：第 0 格同內容窗內被改要記
	a.Observe(15, s.idx, s.base, s.out, 2)
	s.render(0)
	a.Observe(16, s.idx, s.base, s.out, 2)
	if a.Count() != 1 {
		t.Fatalf("反向對照要記一次：%+v", a.Events())
	}
}
