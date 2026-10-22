package main

// 切換列字級（主專案規格 020）：小、中、大三段，只影響切換列與它的下拉選單，不影響遊戲內文字。
// 格子仍是 8 邏輯像素一格，字模置中畫在格內：3 倍時小＝16、中＝22、大＝24 像素；2 倍時小＝12、中與大＝16。
// 12、22 像素字模由 tools/font-dev.sh 以源泉圓體直接渲染（<字型>-bar12.golemfnt、-bar22.golemfnt），
// 缺檔時以最近鄰從相鄰字模縮放。

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wicanr2/dosgolem/apps/sosb"
	"github.com/wicanr2/dosgolem/xlate"
)

type barSize int

const (
	barSmall barSize = iota
	barMedium
	barLarge
)

// barSizeKeys 是下拉選單各字級的 host-text 鍵，順序同 barSize。
var barSizeKeys = []string{"barfont.small", "barfont.medium", "barfont.large"}

// barGlyphPx 回 scale 倍、字級 s 的字模邊長（實體像素）。
func barGlyphPx(scale int, s barSize) int {
	if scale >= 3 {
		return [...]int{16, 22, 24}[s]
	}
	return [...]int{12, 16, 16}[s]
}

func readBarSize(path string) barSize {
	b, err := os.ReadFile(path)
	if err != nil {
		return barMedium
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || n < int(barSmall) || n > int(barLarge) {
		return barMedium
	}
	return barSize(n)
}

// barFonts 是切換列各字模邊長的字型。
type barFonts map[int]*xlate.Font

// loadBarFonts 以 -font 的字型（16）、其 3 倍字模（24）與同目錄的 -bar12、-bar22 字模組成；缺的邊長以最近鄰補。
func loadBarFonts(main *xlate.Font, path string) (barFonts, error) {
	out := barFonts{16: main}
	if s, ok := main.Scaled[3]; ok && s.Font != nil {
		out[24] = s.Font
	}
	base := strings.TrimSuffix(path, ".golemfnt")
	for _, px := range []int{12, 22} {
		p := fmt.Sprintf("%s-bar%d.golemfnt", base, px)
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			continue
		}
		f, err := xlate.LoadFont(p)
		if err != nil {
			return nil, err
		}
		if f.W != px || f.H != px {
			return nil, fmt.Errorf("%s：要是 %d×%d", p, px, px)
		}
		f.Name = filepath.Base(p)
		sosb.AddHalfWidthPairs(f)
		out[px] = f
	}
	if out[24] == nil {
		out[24] = xlate.DeriveNearest(main, 24, 24)
	}
	if out[22] == nil {
		out[22] = xlate.DeriveNearest(out[24], 22, 22)
	}
	if out[12] == nil {
		out[12] = xlate.DeriveNearest(main, 12, 12)
	}
	return out, nil
}

// barFontMenu 是字級下拉選單，對齊切換列的字級項目。
func barFontMenu(text map[string]string, current barSize, bar []toolbarItem) []toolbarItem {
	col := dropMenuCol(bar, actBarFont)
	var out []toolbarItem
	for i, k := range barSizeKeys {
		state := itemEnabled
		if barSize(i) == current {
			state = itemActive
		}
		label := sosb.PackHalfWidthAll([]rune(text[k]))
		out = append(out, toolbarItem{key: k, label: label, act: actBarFont, state: state, col: col, cols: len(label)})
	}
	return out
}

// barGlyphs 回目前倍率與字級要用的字模；字型不是 -font（其他語言）時沿用該字型的倍率規則。
func (g *game) barGlyphs() (*xlate.Font, int) {
	f := g.barFont()
	px := barGlyphPx(g.scale, g.barSize)
	if f == g.mainFont && g.barFonts[px] != nil {
		return g.barFonts[px], px
	}
	v, k, _, _ := f.ForScale(g.scale)
	return v, v.W * k
}

func (g *game) setBarSize(s barSize) {
	g.barSize = s
	if g.barSizePath != "" {
		if err := writePref(g.barSizePath, int(s)); err != nil {
			fmt.Fprintln(os.Stderr, "sosb-play: 字級偏好寫入失敗：", err)
		}
	}
}

func (g *game) clickBarFont(x, y int) {
	items := barFontMenu(g.barText, g.barSize, g.toolbarItems())
	if i := dropMenuHit(items, x, y); i >= 0 {
		g.setBarSize(barSize(i))
	}
	g.bar.sizes = false
}
