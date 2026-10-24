package main

import (
	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/host"
	"testing"
)

func TestPosterWindowInputAndLanguage(t *testing.T) {
	g := &game{posters: &pod.PosterSet{}, chinese: true, langs: []*pod.LangPack{{Code: "zh-TW"}, {Code: "en"}, {Code: "ko"}}}
	for _, scale := range []int{2, 3} {
		g.scale = scale
		w, h := g.frameSize()
		if w != 480*scale || h != 200*scale {
			t.Fatalf("視窗 %d×%d", w, h)
		}
		for _, v := range []struct{ x, y, wantX, wantY int }{{80*scale - 1, 0, -1, 0}, {80 * scale, 0, 0, 0}, {400*scale - 1, 200*scale - 1, 319, 199}, {400 * scale, 0, 320, 0}, {80 * scale, -1, 0, -1}} {
			x, y := g.canvasPoint(v.x, v.y)
			if x != v.wantX || y != v.wantY {
				t.Fatalf("%dx (%d,%d) → (%d,%d)，預期 (%d,%d)", scale, v.x, v.y, x, y, v.wantX, v.wantY)
			}
			outside := x < 0 || x >= 320 || y < 0 || y >= 200
			if outside && mouseEvent(host.MouseEventDown, x, y).Target != host.MouseTargetOutside {
				t.Fatal("海報點擊送入遊戲畫布")
			}
			if x < 0 || x >= 320 {
				g.bar = toolbar{}
				g.bar.track(x, y)
				if g.bar.shown || g.bar.takesPointer(x, y) {
					t.Fatal("海報觸發中央切換列")
				}
			}
		}
	}
	for _, v := range []struct {
		chinese bool
		lang    int
		want    string
	}{{true, 0, "zh-TW"}, {true, 1, "en"}, {true, 2, "ko"}, {false, 0, "en"}, {true, 0, "zh-TW"}} {
		g.chinese, g.lang = v.chinese, v.lang
		if got := g.posterLocale(); got != v.want {
			t.Fatalf("海報語系 %s，預期 %s", got, v.want)
		}
	}
	g.posters = nil
	g.scale = 2
	if w, h := g.frameSize(); w != 640 || h != 400 {
		t.Fatal("無海報時尺寸改變")
	}
	if x, y := g.canvasPoint(0, 0); x != 0 || y != 0 {
		t.Fatal("無海報時座標改變")
	}
}
