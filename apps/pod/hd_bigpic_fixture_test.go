package pod

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func bigpicFixture(t *testing.T) (string, string) {
	t.Helper()
	theme, inv := os.Getenv("DOSGOLEM_POD_BIGPIC_THEME"), os.Getenv("DOSGOLEM_POD_BIGPIC_INVENTORY")
	if theme == "" || inv == "" {
		t.Skip("未提供BIGPIC主題與真實來源")
	}
	return theme, inv
}
func bigpicPixelsEqual(t *testing.T, got, want []byte) {
	t.Helper()
	for row := 0; row < 360; row++ {
		off := ((24+row)*960 + 24) * 4
		if !bytes.Equal(got[off:off+912*4], want[row*912*4:(row+1)*912*4]) {
			t.Fatalf("BIGPIC第%d列與已驗PNG不符", row)
		}
	}
}

func TestBIGPICAllSourcesAndPaletteGuards(t *testing.T) {
	theme, inv := bigpicFixture(t)
	items, e := LoadHDTheme(theme)
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, it := range items {
		if _, ok := bigpicSources[it.Key]; !ok {
			continue
		}
		count++
		idx := make([]byte, 64000)
		for y := 0; y < 120; y++ {
			copy(idx[(8+y)*320+8:], it.Reference[y*304:(y+1)*304])
		}
		p, e := os.ReadFile(filepath.Join(inv, it.Key+".pal"))
		if e != nil || len(p) != 768 {
			t.Fatal(e)
		}
		var pal [256][3]uint8
		for c := range pal {
			copy(pal[c][:], p[c*3:c*3+3])
		}
		original := append([]byte(nil), idx...)
		live := fixtureLive(t, theme)
		live.Frame(idx, pal)
		bigpicPixelsEqual(t, live.Compose(3, true, nil), it.Pix)
		want := vgaBaseRGBA(idx, pal, 3)
		for row := 0; row < 360; row++ {
			off := ((24+row)*960 + 24) * 4
			copy(want[off:off+912*4], it.Pix[row*912*4:(row+1)*912*4])
		}
		if !bytes.Equal(live.Compose(3, true, nil), want) {
			t.Fatal("BIGPIC改動矩形外像素")
		}
		if !bytes.Equal(idx, original) {
			t.Fatal("來源色號被修改")
		}
		if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(idx, pal, 2)) || !bytes.Equal(live.Compose(3, false, nil), vgaBaseRGBA(idx, pal, 3)) {
			t.Fatal("2倍或F12仍有BIGPIC")
		}
		// 真實使用色的任何改動都必須撤除；未使用色改動保持來源匹配。
		changed := pal
		changed[it.Reference[0]][0] ^= 1
		live.Frame(idx, changed)
		if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, changed, 3)) {
			t.Fatalf("%s改用色後仍留HD", it.Key)
		}
		var used [256]bool
		for _, c := range it.Reference {
			used[c] = true
		}
		changed = pal
		unused := -1
		for c, yes := range used {
			if !yes {
				unused = c
				break
			}
		}
		if unused < 0 {
			t.Fatal("預期有未使用色")
		}
		changed[unused][0] ^= 1
		live.Frame(idx, changed)
		bigpicPixelsEqual(t, live.Compose(3, true, nil), it.Pix)
		// 已啟用後，改寫的8×8格回退，其餘完整來源仍可覆繪。
		cell := append([]byte(nil), idx...)
		for y := 40; y < 48; y++ {
			for x := 40; x < 48; x++ {
				cell[y*320+x] ^= 255
			}
		}
		live.Frame(cell, pal)
		got, base := live.Compose(3, true, nil), vgaBaseRGBA(cell, pal, 3)
		for y := 120; y < 144; y++ {
			off := (y*960 + 120) * 4
			if !bytes.Equal(got[off:off+24*4], base[off:off+24*4]) {
				t.Fatalf("%s改寫小格未撤除", it.Key)
			}
		}
		clear(idx)
		live.Frame(idx, pal)
		if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)) {
			t.Fatal("清除後留HD")
		}
		live.ModeChange()
		live.Frame(idx, pal)
		if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)) {
			t.Fatal("模式重設留HD")
		}
	}
	if count != 7 {
		t.Fatalf("BIGPIC數量%d", count)
	}
}

func TestBIGPICNormal244AndRejectedMetadata(t *testing.T) {
	theme, _ := bigpicFixture(t)
	capture := os.Getenv("DOSGOLEM_POD_BIGPIC_CAPTURE")
	if capture == "" {
		t.Skip("未提供正常244來源")
	}
	idx, pal := readLocalizedCapture(t, capture, "f003635-bigpic1-244")
	live := fixtureLive(t, theme)
	live.Frame(idx, pal)
	items, e := LoadHDTheme(theme)
	if e != nil {
		t.Fatal(e)
	}
	var it HDItem
	for _, row := range items {
		if row.Key == "bigpic1-244" {
			it = row
		}
	}
	bigpicPixelsEqual(t, live.Compose(3, true, nil), it.Pix)
	legacy := NewHDLive(nil, nil, nil)
	if e = legacy.AddHD(theme); e != nil {
		t.Fatal(e)
	}
	legacy.Frame(idx, pal)
	bigpicPixelsEqual(t, legacy.Compose(3, true, nil), it.Pix)
	if _, _, e = loadBigpicHD([]HDItem{it}); e != nil {
		t.Fatal(e)
	}
	wrong := it
	wrong.X = 0
	if _, _, e = loadBigpicHD([]HDItem{wrong}); e == nil {
		t.Fatal("錯位置未拒收")
	}
	wrong = it
	wrong.Key = "bigpic1-999"
	if _, _, e = loadBigpicHD([]HDItem{wrong}); e == nil {
		t.Fatal("未知key未拒收")
	}
	wrong = it
	wrong.Reference = append([]byte(nil), it.Reference...)
	wrong.Reference[0] ^= 1
	if _, _, e = loadBigpicHD([]HDItem{wrong}); e == nil {
		t.Fatal("錯來源未拒收")
	}
	if _, _, e = loadBigpicHD([]HDItem{it, it}); e == nil {
		t.Fatal("重複來源未拒收")
	}
}
