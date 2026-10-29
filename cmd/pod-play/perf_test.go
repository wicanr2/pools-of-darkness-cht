package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dosgolem/apps/pod"
)

// 工具列隱藏時仍須計完整Draw；單獨疊工具列的截圖路徑不能灌入畫面次數。
func TestPerfCountsDrawWhenToolbarHidden(t *testing.T) {
	g := &game{live: &pod.Live{}, scale: 2, perf: perfStats{on: true}}
	g.Draw(ebiten.NewImage(640, 400))
	if len(g.perf.draw) != 1 {
		t.Fatalf("隱藏工具列時Draw次數=%d", len(g.perf.draw))
	}
	g.drawBar(nil)
	if len(g.perf.draw) != 1 {
		t.Fatal("截圖疊工具列不應增加Draw次數")
	}
	g.perf.on = false
	g.Draw(ebiten.NewImage(640, 400))
	if len(g.perf.draw) != 1 {
		t.Fatal("關閉量測仍增加Draw次數")
	}
}
