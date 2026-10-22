package main

// 視窗頂端切換列（death-knights-of-krynn-cht 規格 011）：平常隱藏，游標移到頂端時疊在畫面上方出現。
// 每個項目送出的前端動作與對應的功能鍵相同（F2、F12、F5…），由 game.apply 處理。這一層不認識 ebiten。

import (
	"fmt"
	"strings"

	"github.com/wicanr2/dosgolem/apps/sosb"
	"github.com/wicanr2/dosgolem/xlate"
)

const (
	barH        = 12 // 切換列高度（邏輯像素）
	barReveal   = 3  // 游標 y 小於此值時出現
	barTextY    = 2  // 文字列起點
	barFirstCol = 1
	barCols     = 40 // 320 邏輯像素 ÷ 8
)

type itemState int

const (
	itemEnabled  itemState = iota
	itemActive             // 目前狀態（例如目前倍率），已在這個狀態時按下不動作
	itemDisabled           // 灰色，不能按
)

// toolbarItem 是切換列的一段文字；col、cols 以 8 邏輯像素為一格。
type toolbarItem struct {
	key   string
	label []rune
	act   hostAction
	state itemState
	col   int
	cols  int
}

// toolbarState 是決定項目文字與狀態的前端狀態。
type toolbarState struct {
	scale                             int
	hasHD, chinese, muted, fullscreen bool
	hasMusic, music                   bool // 有配樂素材、配樂開啟中（規格 014）
	answers                           bool // 查驗題答案顯示中
	cheat                             bool // 我方無敵開啟中（主專案規格 017）
	journal, guide, mapOpen           bool
	langs                             int  // 已載入的語言數（規格 015）；兩種以上才顯示語言項
	noJournal, noGuide                bool // 目前語言沒有手札、攻略（規格 015），項目停用
	speed                             string // 速度項目的倍率文字（主專案規格 019）；空字串表示不提供
	size                              string // 字級項目的目前字級文字（主專案規格 020）；空字串表示不提供
}

// toolbarItems 依規格 011 §2.2 排出項目，項目間空一格。text 為 host-text 的 toolbar.* 鍵。
func toolbarItems(text map[string]string, st toolbarState) []toolbarItem {
	on := func(b bool) itemState {
		if b {
			return itemActive
		}
		return itemEnabled
	}
	panel := func(open bool) itemState {
		if !st.chinese {
			return itemDisabled // 原版顯示時 F5／F6／F7 不動作
		}
		return on(open)
	}
	scale3 := "toolbar.scale3"
	if st.hasHD {
		scale3 = "toolbar.scale3hd"
	}
	lang, langAct := "toolbar.original", actToggle
	if st.chinese {
		lang = "toolbar.chinese"
	}
	if st.langs > 1 { // 規格 015：語言項顯示目前語言名稱，按下依序換下一種語言，最後一種之後回原版
		langAct = actLangCycle
		if st.chinese {
			lang = "lang.name"
		}
	}
	// 倍率合成一項（使用者 2026-10-22 定案，主專案規格 019 §3.3）：顯示目前倍率，點下等同 F2 切換；不另列「無HD」
	scale := "toolbar.scale2"
	if st.scale == 3 {
		scale = scale3
	}
	items := []toolbarItem{{key: scale, act: actScale}}
	items = append(items,
		toolbarItem{key: lang, act: langAct, state: on(st.chinese)},
		toolbarItem{key: "toolbar.journal", act: actJournal, state: absent(st.noJournal, panel(st.journal))},
		toolbarItem{key: "toolbar.guide", act: actGuide, state: absent(st.noGuide, panel(st.guide))},
		toolbarItem{key: "toolbar.map", act: actMap, state: panel(st.mapOpen)},
		toolbarItem{key: "toolbar.answers", act: actAnswers, state: on(st.answers)},
		toolbarItem{key: "toolbar.cheat", act: actCheat, state: absent(!sosb.CheatVerified, on(st.cheat))},
		toolbarItem{key: "toolbar.music", act: actMusic, state: musicState(st)},
		toolbarItem{key: "toolbar.mute", act: actMute, state: on(st.muted)},
		toolbarItem{key: "toolbar.shot", act: actShot},
		toolbarItem{key: "toolbar.fullscreen", act: actFullscreen, state: on(st.fullscreen)},
	)
	if st.speed != "" {
		items = append(items, toolbarItem{key: "toolbar.speed", act: actSpeed})
	}
	if st.size != "" {
		items = append(items, toolbarItem{key: "toolbar.barfont", act: actBarFont})
	}
	width := 0
	for i := range items {
		// 半形兩字併一格（規格 005 §2；#18），純英文標籤也併：日文「BGM」原本三格，使切換列超過 320 像素
		label := text[items[i].key]
		switch items[i].act {
		case actSpeed:
			label = strings.ReplaceAll(label, "{speed}", st.speed)
		case actBarFont:
			label = strings.ReplaceAll(label, "{size}", st.size)
		}
		items[i].label = sosb.PackHalfWidthAll([]rune(label))
		width += len(items[i].label) + 1
	}
	col := barFirstCol
	if col+width-1 > barCols { // 放不下時從第 0 格開始（3倍HD 加 32x 時剛好 40 格）
		col = 0
	}
	for i := range items {
		items[i].col, items[i].cols = col, len(items[i].label)
		col += items[i].cols + 1
	}
	return items
}

// toolbarHit 回邏輯像素 (x, y) 所在的項目索引，不在任何項目上回 -1。
func toolbarHit(items []toolbarItem, x, y int) int {
	if y < 0 || y >= barH {
		return -1
	}
	for i, it := range items {
		if x >= it.col*8 && x < (it.col+it.cols)*8 {
			return i
		}
	}
	return -1
}

// toolbarClick 回點擊 (x, y) 要送出的動作；停用、已是目前倍率或不在項目上回 0。
func toolbarClick(items []toolbarItem, x, y int) hostAction {
	i := toolbarHit(items, x, y)
	if i < 0 || items[i].state == itemDisabled {
		return 0
	}
	return items[i].act
}

// toolbar 是切換列的顯示狀態。forced 由自動模式腳本的 bar 動作設定。
type toolbar struct {
	shown, forced bool
	speeds        bool // 速度選單展開中（主專案規格 019）
	sizes         bool // 字級選單展開中（主專案規格 020）
}

func (t *toolbar) menuOpen() bool { return t.speeds || t.sizes }

// track 依游標邏輯位置更新顯示：頂端 barReveal 列內出現，離開切換列範圍收起。
func (t *toolbar) track(x, y int) {
	inside := x >= 0 && x < 320 && y >= 0
	switch {
	case t.forced, t.menuOpen():
		t.shown = true // 選單展開時由點擊或 Esc 收起，游標可移到列下方
	case !t.shown:
		t.shown = inside && y < barReveal
	default:
		t.shown = inside && y < barH
	}
}

// takesPointer 回切換列是否吃掉 (x, y) 的滑鼠按下（顯示中且在切換列範圍內）。
func (t *toolbar) takesPointer(x, y int) bool {
	return t.menuOpen() || t.shown && x >= 0 && x < 320 && y >= 0 && y < barH
}

var (
	barBG       = [3]uint8{0, 0, 0}
	barLine     = [3]uint8{170, 170, 170}
	barText     = [3]uint8{255, 255, 255}
	barHover    = [3]uint8{255, 255, 85}
	barActive   = [3]uint8{85, 255, 255}
	barDisabled = [3]uint8{110, 110, 110}
)

// drawToolbar 把切換列畫進 scale 倍的 RGBA（320×200 畫面）；hover 為游標所在項目（-1 為無）。
// glyph 為字級對應的字模（主專案規格 020），1:1 置中畫在 8 邏輯像素的格內。
func drawToolbar(dst []uint8, scale int, glyph *xlate.Font, items []toolbarItem, hover int, missing func(rune)) {
	if glyph == nil || scale <= 0 || len(dst) != 4*320*scale*200*scale {
		return
	}
	w := 320 * scale
	for y := 0; y < barH*scale; y++ {
		c := barBG
		if y >= (barH-1)*scale {
			c = barLine
		}
		for x := 0; x < w; x++ {
			setRGB(dst, w, x, y, c)
		}
	}
	for i, it := range items {
		drawToolbarLabel(dst, scale, glyph, it, i == hover && toolbarClick(items, it.col*8, barTextY) != 0, barTextY, missing)
	}
}

// drawToolbarLabel 畫一個項目的文字；rowY 是文字列起點（邏輯像素）。
func drawToolbarLabel(dst []uint8, scale int, f *xlate.Font, it toolbarItem, hover bool, rowY int, missing func(rune)) {
	w, cell := 320*scale, 8*scale
	c := barText
	switch {
	case it.state == itemDisabled:
		c = barDisabled
	case hover:
		c = barHover
	case it.state == itemActive:
		c = barActive
	}
	rb := (f.W + 7) / 8
	ox, oy := (cell-f.W)/2, (cell-f.H)/2
	for n, r := range it.label {
		if r == ' ' {
			continue
		}
		g, ok := f.Glyphs[r]
		if !ok {
			if missing != nil {
				missing(r)
			}
			continue
		}
		cx, cy := (it.col+n)*cell, rowY*scale
		for gy := 0; gy < f.H; gy++ {
			for gx := 0; gx < f.W; gx++ {
				if g[gy*rb+gx/8]&(0x80>>uint(gx%8)) == 0 {
					continue
				}
				if lx, ly := ox+gx, oy+gy; lx >= 0 && lx < cell && ly >= 0 && ly < cell {
					setRGB(dst, w, cx+lx, cy+ly, c)
				}
			}
		}
	}
}

// dropMenuCol 回下拉選單文字起始格：對齊切換列上 act 項目的下一格，選單（寬 6 格）超出右緣時往左移。
func dropMenuCol(bar []toolbarItem, act hostAction) int {
	col := barFirstCol
	for _, it := range bar {
		if it.act == act {
			col = it.col
			break
		}
	}
	if col+dropMenuCols > barCols {
		col = barCols - dropMenuCols
	}
	return col + 1
}

const dropMenuCols = 6

// dropMenuHit 回下拉選單（切換列下方每列一項、寬 6 格）中 (x, y) 所在的項目，不在選單上回 -1。
func dropMenuHit(items []toolbarItem, x, y int) int {
	n := len(items)
	if n == 0 || y < barH || y >= barH*(n+1) {
		return -1
	}
	left := (items[0].col - 1) * 8
	if x < left || x >= left+dropMenuCols*8 {
		return -1
	}
	return y/barH - 1
}

// drawDropMenu 畫下拉選單；hover 為游標所在項目（-1 為無）。停用列（例如速度選單的實際倍率）不反白。
func drawDropMenu(dst []uint8, scale int, glyph *xlate.Font, items []toolbarItem, hover int, missing func(rune)) {
	n := len(items)
	if n == 0 || glyph == nil {
		return
	}
	left := (items[0].col - 1) * 8
	for y := barH * scale; y < (n+1)*barH*scale; y++ {
		for x := left * scale; x < (left+dropMenuCols*8)*scale; x++ {
			setRGB(dst, 320*scale, x, y, barBG)
		}
	}
	for i, it := range items {
		drawToolbarLabel(dst, scale, glyph, it, i == hover && it.state != itemDisabled, (i+1)*barH+barTextY, missing)
	}
}

func setRGB(dst []uint8, w, x, y int, c [3]uint8) {
	i := 4 * (y*w + x)
	if x < 0 || x >= w || i < 0 || i+3 >= len(dst) {
		return
	}
	dst[i], dst[i+1], dst[i+2], dst[i+3] = c[0], c[1], c[2], 0xFF
}

// toolbarKeys 是切換列用到的全部 host-text 鍵。
var toolbarKeys = []string{"toolbar.scale2", "toolbar.scale3", "toolbar.scale3hd", "toolbar.chinese",
	"toolbar.original", "toolbar.journal", "toolbar.guide", "toolbar.map", "toolbar.answers", "toolbar.cheat", "toolbar.mute", "toolbar.music",
	"toolbar.shot", "toolbar.fullscreen", "toolbar.speed", "toolbar.barfont", "barfont.small", "barfont.medium", "barfont.large"}

// checkToolbarText 確認 host-text 有全部切換列文字。
func checkToolbarText(text map[string]string) error {
	for _, k := range toolbarKeys {
		if text[k] == "" {
			return fmt.Errorf("host-text.zh-TW.tsv 缺 %s", k)
		}
	}
	return nil
}

// musicState：沒有配樂素材時停用；開啟中為目前狀態，但與倍率不同，按下仍會切換（toolbarClick 只擋倍率）。
func musicState(st toolbarState) itemState {
	switch {
	case !st.hasMusic:
		return itemDisabled
	case st.music:
		return itemActive
	}
	return itemEnabled
}

// absent：目前語言沒有這個面板時停用。
func absent(none bool, st itemState) itemState {
	if none {
		return itemDisabled
	}
	return st
}
