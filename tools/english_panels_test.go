package pod

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 以overlay測正式資料及原面板；局部長文截圖不替代正常玩家路線。
func TestEnglishPanels(t *testing.T) {
	root, out := os.Getenv("POD_EN_ROOT"), os.Getenv("POD_EN_OUT")
	if root == "" || out == "" {
		t.Skip("requires project catalogs and fonts")
	}
	f, err := LoadFontScaled(filepath.Join(root, "workplace/font/pod-en.golemfnt"))
	if err != nil {
		t.Fatal(err)
	}
	if f.Scaled[3].Font == nil || f.Scaled[3].Font.W != 24 {
		t.Fatal("no native24 font")
	}
	p, err := LoadLangPack(filepath.Join(root, "text"), EnglishLang, f)
	if err != nil {
		t.Fatal(err)
	}
	if p.Journal == nil || len(p.Journal.Entries) != 99 || p.Guide == nil || len(p.Guide.Chapters) != 78 || p.Map == nil || len(p.Map.Guide.Titles) != 32 {
		t.Fatal("incomplete panels")
	}
	compact := func(s string) string { return strings.Join(strings.Fields(strings.ReplaceAll(s, `\n`, " ")), "") }
	flat := func(pages [][][]rune) string {
		var b strings.Builder
		for _, page := range pages {
			for _, line := range page {
				b.WriteString(string(line))
			}
		}
		return compact(b.String())
	}
	journalPages, guidePages, maxJournal, maxGuide := 0, 0, 0, 0
	for n, e := range p.Journal.Entries {
		pages := JournalPages(e)
		if len(pages) == 0 || flat(pages) != compact(e.Body) {
			t.Fatal("journal text lost", n)
		}
		journalPages += len(pages)
		maxJournal = max(maxJournal, len(pages))
		p.Journal.OpenAuto(n, 0)
		for page := range pages {
			p.Journal.page = page
			for _, line := range p.Journal.Lines() {
				if len(line) > journalCols {
					t.Fatal("journal overflow", n, page, len(line))
				}
			}
		}
		p.Journal.CloseAuto()
	}
	for n, e := range p.Guide.Chapters {
		p.Guide.SetPos(n+1, 0)
		if !p.Guide.Open() {
			p.Guide.Toggle()
		}
		pages := p.Guide.pages
		if len(pages) == 0 || flat(pages) != compact(e.Body) {
			t.Fatal("guide text lost", n+1)
		}
		guidePages += len(pages)
		maxGuide = max(maxGuide, len(pages))
		for page := range pages {
			p.Guide.page = page
			for _, line := range p.Guide.Lines() {
				if len(line) > guideCols {
					t.Fatal("guide overflow", n+1, page, len(line))
				}
			}
		}
	}
	samples := []map[string]any{}
	save := func(name string, pixels []byte, scale int) {
		dir := filepath.Join(out, "layout")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		file, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, &image.NRGBA{Pix: pixels, Stride: 320 * scale * 4, Rect: image.Rect(0, 0, 320*scale, 200*scale)})
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	missing := func(r rune) { t.Errorf("missing glyph U+%04X", r) }
	for _, scale := range []int{2, 3} {
		for _, n := range []int{1, 11, 57, 74} {
			p.Journal.OpenAuto(n, 0)
			for _, page := range []int{0, len(p.Journal.pages) - 1} {
				p.Journal.page = page
				pixels := make([]byte, 320*200*scale*scale*4)
				for i := 3; i < len(pixels); i += 4 {
					pixels[i] = 255
				}
				p.Journal.Draw(pixels, scale, f, missing)
				name := fmt.Sprintf("journal-%03d-p%d-%dx", n, page+1, scale)
				save(name, pixels, scale)
				samples = append(samples, map[string]any{"file": name + ".png", "kind": "journal", "entry": n, "page": page + 1, "pages": len(p.Journal.pages), "scale": scale})
			}
			p.Journal.CloseAuto()
		}
		for _, n := range []int{1, 16, 55} {
			p.Guide.SetPos(n, 0)
			if !p.Guide.Open() {
				p.Guide.Toggle()
			}
			for _, page := range []int{0, len(p.Guide.pages) - 1} {
				p.Guide.page = page
				pixels := make([]byte, 320*200*scale*scale*4)
				for i := 3; i < len(pixels); i += 4 {
					pixels[i] = 255
				}
				p.Guide.Draw(pixels, scale, f, missing)
				name := fmt.Sprintf("guide-%03d-p%d-%dx", n, page+1, scale)
				save(name, pixels, scale)
				samples = append(samples, map[string]any{"file": name + ".png", "kind": "guide", "chapter": n, "page": page + 1, "pages": len(p.Guide.pages), "scale": scale})
			}
		}
	}
	// 三語都已有內容時，已讀與目前攻略位置不可因切換遺失。
	twFont, err := LoadFontScaled(filepath.Join(root, "workplace/font/pod-rounded.golemfnt"))
	if err != nil {
		t.Fatal(err)
	}
	tw, err := LoadLangPack(filepath.Join(root, "text"), DefaultLang, twFont)
	if err != nil {
		t.Fatal(err)
	}
	live := NewLive(twFont, nil, nil)
	live.Journal, live.Guide, live.Map = tw.Journal, tw.Guide, tw.Map
	live.SetLang(tw)
	live.Journal.Seen = map[int]bool{57: true}
	live.Guide.SetPos(55, 1)
	live.Guide.Toggle()
	live.SetLang(p)
	if ch, page := live.Guide.Pos(); ch != 55 || page != 1 || !live.Guide.Open() || !live.Journal.Seen[57] {
		t.Fatal("locale lost progress")
	}
	// 錯誤的實際英文表頭必須由現有解析器拒絕。
	if _, err := LoadJournal([]byte("bad\ttitle\tbody\n1\tx\ty\n")); err == nil {
		t.Fatal("bad header accepted")
	}
	proof := map[string]any{"journals": 99, "guide_chapters": 78, "journal_pages": journalPages, "guide_pages": guidePages, "max_journal_pages": maxJournal, "max_guide_pages": maxGuide, "all_text_preserved": true, "all_lines_bounded": true, "locale_progress_preserved": true, "bad_header_rejected": true, "samples": samples, "scope": "local panel rendering; not fresh journal-trigger player proof"}
	b, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "test/layout.json"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("all177 records retain every nonspace character; %d journal pages,%d guide pages; 4/3 classes sampled", journalPages, guidePages)
}
