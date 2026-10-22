package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSpeedControlChooseAndLabel(t *testing.T) {
	s := speedControl{base: 50, set: 1, pct: 50, adaptive: true}
	if p := s.choose(8); p != 50 {
		t.Fatalf("未登記 TP Delay 時應維持 1x，得 %d", p)
	}
	s.ready = true
	if p := s.choose(8); p != 50 || s.target() != 400 {
		t.Fatalf("互動模式從 1x 起調：%d、目標 %d", p, s.target())
	}
	if s.label() != "8x" || s.actual() != "1x" {
		t.Fatalf("label %q actual %q", s.label(), s.actual())
	}
	s.pct = 115
	if s.actual() != "2.3x" {
		t.Fatalf("actual %q", s.actual())
	}
	s.pct = 50
	fixed := speedControl{base: 50, set: 1, pct: 50, ready: true}
	if p := fixed.choose(32); p != 1600 || fixed.label() != "32x" || fixed.actual() != "32x" {
		t.Fatalf("自動模式固定為設定值：%d %q", p, fixed.label())
	}
	if p := fixed.observe(time.Second, true); p != 1600 {
		t.Fatal("自動模式不調節")
	}
}

func TestSpeedControlAdapts(t *testing.T) {
	s := speedControl{base: 50, set: 1, pct: 50, adaptive: true, ready: true}
	s.choose(8)
	for i := 0; i < 200; i++ {
		s.observe(5*time.Millisecond, false)
	}
	if s.pct != 400 {
		t.Fatalf("主機有餘裕時應升到設定值，得 %d", s.pct)
	}
	for i := 0; i < 200; i++ {
		s.observe(40*time.Millisecond, false)
	}
	if s.pct != 50 {
		t.Fatalf("主機跟不上時應降回 1x，得 %d", s.pct)
	}
	s.choose(1)
	if p := s.observe(time.Second, true); p != 50 {
		t.Fatal("1x 不調節")
	}
}

func TestSpeedPreference(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sosb-speed.txt")
	if readSpeed(path) != 1 {
		t.Fatal("沒有偏好檔應為 1x")
	}
	if err := writePref(path, 16); err != nil || readSpeed(path) != 16 {
		t.Fatalf("寫入後讀回 %d（%v）", readSpeed(path), err)
	}
	if err := os.WriteFile(path, []byte("3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if readSpeed(path) != 1 {
		t.Fatal("無效倍率應為 1x")
	}
}

func TestSpeedMenu(t *testing.T) {
	bar := toolbarItems(testBarText, toolbarState{scale: 2, chinese: true, speed: "8x"})
	menu := speedMenu(8, "2.3x", bar)
	if len(menu) != len(speedChoices)+1 || menu[0].state != itemDisabled || menu[4].state != itemActive {
		t.Fatalf("選單 %d 項、實際列 %v、8x 狀態 %v", len(menu), menu[0].state, menu[4].state)
	}
	x := (menu[0].col) * 8
	if i := dropMenuHit(menu, x, barH*5+1); i != 4 || speedChoices[i-1] != 8 {
		t.Fatalf("第 5 列點擊得 %d", i)
	}
	if i := dropMenuHit(menu, x, barH*8+1); i != -1 {
		t.Fatal("選單外點擊應為 -1")
	}
}

func TestScriptSpeed(t *testing.T) {
	s, err := parseScript("10:speed=8")
	if err != nil || s[10][0].speed != 8 {
		t.Fatalf("%v %v", s, err)
	}
	if _, err := parseScript("10:speed=3"); err == nil {
		t.Fatal("無效倍率應回錯誤")
	}
}

// TestBarFontSizes 釘住字級對應的字模邊長（主專案規格 020，使用者 2026-10-22 定案）。
func TestBarFontSizes(t *testing.T) {
	want := map[int][3]int{2: {12, 16, 16}, 3: {16, 22, 24}}
	for scale, px := range want {
		for s := barSmall; s <= barLarge; s++ {
			if got := barGlyphPx(scale, s); got != px[s] {
				t.Fatalf("%d× 字級 %d：%d 像素，要 %d", scale, s, got, px[s])
			}
		}
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "sosb-barfont.txt")
	if readBarSize(path) != barMedium {
		t.Fatal("沒有偏好檔時應為中")
	}
	if err := writePref(path, int(barSmall)); err != nil || readBarSize(path) != barSmall {
		t.Fatal("偏好檔讀寫不一致")
	}
	os.WriteFile(path, []byte("7\n"), 0o644)
	if readBarSize(path) != barMedium {
		t.Fatal("無效偏好應回中")
	}
}

func TestBarFontMenuAndScript(t *testing.T) {
	bar := toolbarItems(testBarText, toolbarState{scale: 2, chinese: true, speed: "1x", size: "中"})
	menu := barFontMenu(testBarText, barLarge, bar)
	if len(menu) != 3 || menu[2].state != itemActive || string(menu[0].label) != "小" {
		t.Fatalf("字級選單 %+v", menu)
	}
	if i := dropMenuHit(menu, (menu[0].col-1)*8+2, barH*2+3); i != 1 {
		t.Fatalf("點第二列得 %d，要 1（中）", i)
	}
	got, err := parseScript("5:barfont=1,6:barfonts")
	if err != nil || got[5][0].size != 1 || got[6][0].act != actBarFont {
		t.Fatalf("腳本 %+v %v", got, err)
	}
	if _, err := parseScript("5:barfont=4"); err == nil {
		t.Fatal("字級 4 應失敗")
	}
}

// TestDropMenusInsideScreen：速度、字級選單在切換列最右邊時往左移，不超出 320 邏輯像素。
func TestDropMenusInsideScreen(t *testing.T) {
	bar := toolbarItems(testBarText, toolbarState{scale: 3, hasHD: true, chinese: true, speed: "32x", size: "中"})
	for _, menu := range [][]toolbarItem{speedMenu(32, "17.9x", bar), barFontMenu(testBarText, barMedium, bar)} {
		left := (menu[0].col - 1) * 8
		if left < 0 || left+dropMenuCols*8 > 320 {
			t.Fatalf("選單 x=%d..%d 超出畫面", left, left+dropMenuCols*8)
		}
	}
}
