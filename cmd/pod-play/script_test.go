package main

import "testing"

// TestParseScriptHostActs 釘住腳本的面板動作名稱，以及 F 鍵名照舊送進 DOS（既有收據不換畫面）。
func TestParseScriptHostActs(t *testing.T) {
	got, err := parseScript("10:guide,11:pgdn,12:next,13:journal,14:map,15:F6,16:shot,17:scale,18:F2,19:mapall,20:V")
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]hostAction{10: actGuide, 11: actPageDown, 12: actNextEntry, 13: actJournal, 14: actMap, 16: actShot, 17: actScale, 19: actReveal}
	for f, act := range want {
		if len(got[f]) != 1 || got[f][0].act != act {
			t.Fatalf("畫格 %d 得到 %+v，要動作 %d", f, got[f], act)
		}
	}
	if a := got[15]; len(a) != 1 || a[0].act != 0 || a[0].key == nil {
		t.Fatalf("F6 應送 DOS 鍵，得到 %+v", a)
	}
	if a := got[18]; len(a) != 1 || a[0].act != 0 || a[0].key == nil {
		t.Fatalf("F2 應保留 DOS 鍵，前端倍率動作用 scale，得到 %+v", a)
	}
	if a := got[20]; len(a) != 1 || a[0].act != 0 || a[0].key == nil {
		t.Fatalf("V 應保留 DOS 鍵，全圖動作用 mapall：%+v", a)
	}
}
