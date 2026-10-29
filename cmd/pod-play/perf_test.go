package main

import (
	"testing"
	"time"

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

func TestSpeedControlUsesRenderingFeedback(t *testing.T) {
	s := speedControl{base: 50, set: 8, pct: 200, adaptive: true, ready: true}
	now := time.Unix(100, 0)
	s.rendered(now)
	for i := 0; i < 160; i++ {
		now = now.Add(50 * time.Millisecond)
		s.rendered(now)
		if got := s.observe(5*time.Millisecond, false); got > 200 {
			t.Fatal("畫面20fps時仍升速")
		}
	}
	if s.pct != 50 {
		t.Fatalf("慢繪製應退回1x，得%d", s.pct)
	}
	for i := 0; i < 240; i++ {
		now = now.Add(10 * time.Millisecond)
		s.rendered(now)
		s.observe(5*time.Millisecond, false)
	}
	if s.pct != 400 {
		t.Fatalf("畫面恢復後應重新升速，得%d", s.pct)
	}
	// 32fps位於遲滯區，不因模擬很快反覆提高倍率。
	s = speedControl{base: 50, set: 8, pct: 100, adaptive: true, ready: true}
	s.rendered(now)
	for i := 0; i < 80; i++ {
		now = now.Add(31 * time.Millisecond)
		s.rendered(now)
		if s.observe(5*time.Millisecond, false) != 100 {
			t.Fatal("遲滯區應保持時脈")
		}
	}
	s.choose(1)
	if s.observe(time.Second, true) != 50 {
		t.Fatal("繪製回饋不能改變1x基準")
	}
}
