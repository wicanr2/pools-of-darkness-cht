package main

import (
	"testing"

	"github.com/wicanr2/dosgolem/apps/pod"
)

// 從實際鍵盤對映走到面板動作，確認 F7、V、Esc 不洩漏 BIOS 鍵，並驗 F5/F6 互斥。
func TestMapPanelKeyRouting(t *testing.T) {
	m := &pod.AutoMap{Key: "test", Guide: &pod.MapGuide{Points: map[string][]pod.MapPoint{
		"test": {{N: 1, X: 3, Y: 4, Label: "未走過的點"}},
	}}}
	g := &game{chinese: true, live: &pod.Live{Map: m,
		Guide:   &pod.GuidePanel{Chapters: []pod.GuideChapter{{Title: "攻略", Body: "測試"}}},
		Journal: &pod.JournalPanel{Entries: map[int]pod.JournalEntry{57: {Title: "手札", Body: "測試"}}, Seen: map[int]bool{57: true}},
	}}
	press := func(key string, chars ...rune) {
		t.Helper()
		keys, acts := mapKeys(frameInput{Held: map[string]int{key: 1}, Chars: chars}, g.panelMode())
		if len(keys) != 0 || len(acts) != 1 {
			t.Fatalf("%s：BIOS=%v actions=%v", key, keys, acts)
		}
		g.apply(acts)
	}
	press("F7")
	if !m.Open() || len(m.VisiblePoints()) != 0 {
		t.Fatal("F7 未開地圖或預設洩漏未走過的點")
	}
	press("V", 'v')
	if len(m.VisiblePoints()) != 1 {
		t.Fatal("V 未顯示全圖攻略點")
	}
	press("V", 'v')
	if len(m.VisiblePoints()) != 0 {
		t.Fatal("V 未切回走過模式")
	}
	press("F6")
	if m.Open() || !g.live.Guide.Open() {
		t.Fatal("F6 未關地圖並開攻略")
	}
	press("F7")
	if !m.Open() || g.live.Guide.Open() {
		t.Fatal("F7 未關攻略並開地圖")
	}
	press("F5")
	if m.Open() || !g.live.Journal.Browsing() {
		t.Fatal("F5 未關地圖並開手札")
	}
	press("F7")
	if !m.Open() || g.live.Journal.Browsing() {
		t.Fatal("F7 未關手札並開地圖")
	}
	press("Escape")
	if m.Open() {
		t.Fatal("Esc 未關地圖")
	}
	press("F7")
	press("F7")
	if m.Open() {
		t.Fatal("F7 再按未關地圖")
	}
}
