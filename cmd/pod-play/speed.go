package main

// CPU 模擬速度（主專案規格 019，由幽靈騎士主專案規格 019 移植）：倍率是虛擬 CPU 每秒遊戲時間的指令數，
// 遊戲時鐘不變。1x 是 -clock 的時脈，行為與加入本功能前相同；大於 1x 時依主機負擔自動調節實際倍率。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wicanr2/dosgolem/apps/pod"
)

var speedChoices = []int{1, 2, 4, 8, 16, 32}

const (
	speedTarget   = 18 * time.Millisecond // 每格模擬耗時目標（30fps 的 33 ms 扣掉合成與上傳）
	speedHigh     = 22 * time.Millisecond // 超過就降
	speedLow      = 14 * time.Millisecond // 低於且未被截斷就升
	speedFrameCap = 25 * time.Millisecond // 單格模擬上限：超過就不再跑下一段
	speedSettle   = 10                    // 兩次調整至少隔幾格
	maxClockPct   = 1600
	renderSlow    = time.Second / 30 // 畫面低於30fps時停止加速
	renderFast    = time.Second / 35 // 恢復至35fps後才重新升速
)

func validSpeed(m int) bool {
	for _, c := range speedChoices {
		if c == m {
			return true
		}
	}
	return false
}

func readSpeed(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 1
	}
	m, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || !validSpeed(m) {
		return 1
	}
	return m
}

// writePref 原子寫入一個整數偏好（速度倍率、切換列字級）。
func writePref(path string, m int) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pod-pref-*")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if _, err = fmt.Fprintf(file, "%d\n", m); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// speedControl 決定每格的時脈比例。base 為 1x 的時脈（-clock），set 為設定倍率，pct 為目前實際時脈。
type speedControl struct {
	base, set, pct int
	adaptive       bool          // 互動模式才調節；自動模式固定為設定值
	ema            time.Duration // 每格模擬耗時的指數平均
	settle         int
	ready          bool // 已登記 TP Delay：沒有它就不能加速（Delay 會跟著縮短）
	drawAt         time.Time
	drawInterval   time.Duration
}

// rendered 觀察完整Draw回呼的開始間距，包含兩次回呼之間的模擬與呈現負擔。
func (s *speedControl) rendered(now time.Time) {
	if !s.adaptive {
		return
	}
	if !s.drawAt.IsZero() {
		if interval := now.Sub(s.drawAt); interval > 0 {
			if s.drawInterval == 0 {
				s.drawInterval = interval
			} else {
				s.drawInterval += (interval - s.drawInterval) / 8
			}
		}
	}
	s.drawAt = now
}

func (s *speedControl) target() int {
	p := s.base * s.set
	if p > maxClockPct {
		p = maxClockPct
	}
	return p
}

// choose 設定倍率；回應採用的時脈比例。未 ready 時維持 1x。
func (s *speedControl) choose(m int) int {
	s.set = m
	switch {
	case !s.ready || m == 1:
		s.pct = s.base
	case s.adaptive:
		if s.pct > s.target() || s.pct < s.base {
			s.pct = s.base
		}
	default:
		s.pct = s.target()
	}
	s.settle = speedSettle
	return s.pct
}

// observe 在每格模擬後呼叫；capped 表示本格因上限而未跑完。回應新的時脈比例（不變則同原值）。
func (s *speedControl) observe(spent time.Duration, capped bool) int {
	if !s.adaptive || !s.ready || s.set == 1 {
		return s.pct
	}
	if s.ema == 0 {
		s.ema = spent
	} else {
		s.ema += (spent - s.ema) / 8
	}
	if s.settle > 0 {
		s.settle--
		return s.pct
	}
	p := s.pct
	switch {
	case s.ema > speedHigh || capped || s.drawInterval > renderSlow:
		p = p * 4 / 5
	case s.ema < speedLow && (s.drawInterval == 0 || s.drawInterval <= renderFast):
		p = p * 5 / 4
	}
	if p < s.base {
		p = s.base
	}
	if t := s.target(); p > t {
		p = t
	}
	if p != s.pct {
		s.pct, s.settle = p, speedSettle
	}
	return s.pct
}

// label 是切換列速度項目的倍率文字（設定值），例如 8x。
func (s *speedControl) label() string { return fmt.Sprintf("%dx", s.set) }

// actual 是實際倍率文字，例如 2.3x；整數倍時不帶小數。
func (s *speedControl) actual() string {
	if s.pct%s.base == 0 {
		return fmt.Sprintf("%dx", s.pct/s.base)
	}
	return fmt.Sprintf("%.1fx", float64(s.pct)/float64(s.base))
}

// speedMenu 是速度下拉選單，對齊切換列的速度項目：第一列是實際倍率（灰色、不能點），其下為各倍率。
func speedMenu(current int, actual string, bar []toolbarItem) []toolbarItem {
	col := dropMenuCol(bar, actSpeed)
	info := pod.PackHalfWidthAll([]rune(actual))
	out := []toolbarItem{{key: "actual", label: info, state: itemDisabled, col: col, cols: len(info)}}
	for _, m := range speedChoices {
		state := itemEnabled
		if m == current {
			state = itemActive
		}
		label := pod.PackHalfWidthAll([]rune(fmt.Sprintf("%dx", m)))
		out = append(out, toolbarItem{key: strconv.Itoa(m), label: label, act: actSpeed, state: state, col: col, cols: len(label)})
	}
	return out
}

// registerDelay 在覆繪綁定找到 CRT.Delay 後登記一次，之後才套用設定倍率（Delay 已在開機時以 1x 校準）。
func (g *game) registerDelay() {
	if g.tpRegistered || !g.delayAccel || g.live.TPDelayLoop == 0 {
		return
	}
	g.tpRegistered = true
	if err := g.owner.SetTPDelay(g.live.TPDelayLoop, g.live.TPDelayCount); err != nil {
		fmt.Fprintln(os.Stderr, "pod-play: TP Delay 登記失敗：", err)
		return
	}
	g.speed.ready = true
	if err := g.applyClock(g.speed.choose(g.speed.set)); err != nil {
		fmt.Fprintln(os.Stderr, "pod-play: 速度設定失敗：", err)
	}
}

// applyClock 在回合之間改時脈比例，每格預算與混音器速率跟著改。
func (g *game) applyClock(p int) error {
	if p == g.owner.ClockPercent() {
		return nil
	}
	if err := g.owner.SetClockPercent(p); err != nil {
		return err
	}
	g.budget = stepsPerHostFrame(p)
	g.mix.SetStepsPerSecond(g.budget*60, g.lastSteps)
	return nil
}

func (g *game) setSpeed(m int) {
	if err := g.applyClock(g.speed.choose(m)); err != nil {
		fmt.Fprintln(os.Stderr, "pod-play: 速度設定失敗：", err)
	}
}

func (g *game) clickSpeed(x, y int) {
	items := speedMenu(g.speed.set, g.speed.actual(), g.toolbarItems())
	if i := dropMenuHit(items, x, y); i >= 1 {
		m := speedChoices[i-1]
		g.setSpeed(m)
		if g.speedPath != "" {
			if err := writePref(g.speedPath, m); err != nil {
				fmt.Fprintln(os.Stderr, "pod-play: 速度偏好寫入失敗：", err)
			}
		}
	}
	g.bar.speeds = false
}

// perfStats 是 -perf 的量測（主專案規格 019 §4）：每格模擬耗時、每次繪製耗時、實際時脈的平均。
type perfStats struct {
	on          bool
	start       time.Time
	adv, draw   []time.Duration
	pctSum, pcN int
}

func (p *perfStats) report(base int) {
	if !p.on {
		return
	}
	stat := func(name string, d []time.Duration) {
		if len(d) == 0 {
			return
		}
		s := append([]time.Duration(nil), d...)
		sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
		var sum time.Duration
		for _, v := range s {
			sum += v
		}
		fmt.Fprintf(os.Stderr, "pod-play: perf %s n=%d mean=%v p50=%v p90=%v\n", name, len(s), sum/time.Duration(len(s)), s[len(s)/2], s[len(s)*9/10])
	}
	stat("advance", p.adv)
	stat("draw", p.draw)
	w := time.Since(p.start).Seconds()
	avg := 0.0
	if p.pcN > 0 {
		avg = float64(p.pctSum) / float64(p.pcN) / float64(base)
	}
	fmt.Fprintf(os.Stderr, "pod-play: perf wall=%.1fs drawfps=%.1f updatefps=%.1f 平均實際倍率=%.2fx\n",
		w, float64(len(p.draw))/w, float64(len(p.adv))/w, avg)
}
