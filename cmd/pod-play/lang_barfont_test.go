package main

import (
	"testing"

	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/xlate"
)

func TestLanguageBarFontSizes(t *testing.T) {
	main := &xlate.Font{W: 16, H: 16}
	cn := &xlate.Font{W: 16, H: 16}
	fonts := barFonts{}
	for _, px := range []int{12, 16, 22, 24} {
		fonts[px] = &xlate.Font{W: px, H: px}
	}
	live := &pod.Live{Font: cn}
	g := &game{live: live, mainFont: main, langs: []*pod.LangPack{{Code: "zh-CN", Font: cn}}, langBarFonts: map[*xlate.Font]barFonts{cn: fonts}}
	for _, scale := range []int{2, 3} {
		for _, size := range []barSize{barSmall, barMedium, barLarge} {
			g.scale, g.barSize = scale, size
			font, px := g.barGlyphs()
			want := barGlyphPx(scale, size)
			if font != fonts[want] || px != want {
				t.Fatalf("scale=%d size=%d got=%d want=%d", scale, size, px, want)
			}
		}
	}
	// 英文工具列仍用主字型，不能因加入簡中快取而混入簡中字模。
	g.langs[0].Code = pod.EnglishLang
	g.barFonts = fonts
	if got, _ := g.barGlyphs(); got != fonts[24] {
		t.Fatal("英文工具列沒有沿用主字型")
	}
}
