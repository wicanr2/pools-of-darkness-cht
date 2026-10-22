package xlate

import (
	"bytes"
	"testing"
)

// stickyWatcher：16×8 兩格的圖面，兩格都有墨跡（色號 2 的一點）；畫成 col 色。
func stickyWatcher(key string, want []byte, col [4]uint8, sticky bool) *Watcher {
	return &Watcher{Key: key, Art: true, ArtKeys: []string{key + ":0"}, W: 16, H: 8, Want: want, Sticky: sticky,
		Make: func() []*Stamp {
			pix := make([]uint8, 16*8*9*4)
			for i := 0; i < len(pix); i += 4 {
				copy(pix[i:i+4], col[:])
			}
			return []*Stamp{{Key: key + ":0", Cells: 2, CellW: 8, CellH: 8, Art: true, PixScale: 3, Pix: pix,
				Reference: append([]byte(nil), want...)}}
		}}
}

func stickyWant() []byte {
	w := bytes.Repeat([]byte{1}, 128)
	w[3*16+2], w[3*16+10] = 2, 2
	return w
}

func paint(idx, want []byte) { copy(idx[:128], want) }

func TestStickyArtKeepsUnchangedCells(t *testing.T) {
	for _, sticky := range []bool{true, false} {
		l := &Layer{W: 16, H: 16}
		l.Watch(stickyWatcher("a", stickyWant(), artRed, sticky))
		idx, rgb := artBoard()
		paint(idx, stickyWant())
		l.Frame(idx, rgb)
		dst, _ := drawSmall(l)
		regionIs(t, dst, 0, 0, 16, 8, artRed)
		idx[3*16+10] = 5 // 第 1 格被原版改掉（例如另一張圖開始畫）
		l.Frame(idx, rgb)
		dst, drew := drawSmall(l)
		if sticky {
			regionIs(t, dst, 0, 0, 8, 8, artRed) // 沒被改的格照畫
			regionIs(t, dst, 8, 0, 16, 8, [4]byte{})
		} else if drew {
			t.Fatal("對照：非 Sticky 不整張吻合就整組停用（spec 204 §2.4）")
		}
		idx[3*16+2] = 5 // 墨跡格全都不吻合：停用
		l.Frame(idx, rgb)
		if _, drew := drawSmall(l); drew {
			t.Fatalf("sticky=%v：沒有墨跡格吻合仍在畫", sticky)
		}
		idx[3*16+2] = 2 // 只恢復一格：已停用，要重新整張吻合才啟用
		l.Frame(idx, rgb)
		if _, drew := drawSmall(l); drew {
			t.Fatalf("sticky=%v：停用後只靠部分吻合就重新啟用", sticky)
		}
	}
}

func TestStickyArtYieldsToFullMatch(t *testing.T) {
	blue := [4]uint8{20, 40, 220, 255}
	l := &Layer{W: 16, H: 16}
	l.Watch(stickyWatcher("bg", stickyWant(), artRed, true))
	logo := stickyWant()
	logo[3*16+10] = 5
	l.Watch(stickyWatcher("logo", logo, blue, true))
	idx, rgb := artBoard()
	paint(idx, stickyWant())
	l.Frame(idx, rgb)
	paint(idx, logo) // 新圖整張吻合
	l.Frame(idx, rgb)
	dst, _ := drawSmall(l)
	regionIs(t, dst, 0, 0, 16, 8, blue) // 舊圖讓位，不在底下殘留
}
