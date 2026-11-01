package pod

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 局部分頁不替代正常路線。解開半形組合後比對全文，防止混合文字被截斷。
func TestJaKoPanels(t *testing.T) {
	root, out := os.Getenv("POD_JAKO_ROOT"), os.Getenv("POD_JAKO_OUT")
	if root == "" || out == "" {
		t.Skip("requires complete candidate data")
	}
	compact := func(s string) string { return strings.Join(strings.Fields(strings.ReplaceAll(s, `\n`, " ")), "") }
	flat := func(pages [][][]rune) string {
		var text strings.Builder
		for _, page := range pages {
			for _, line := range page {
				for _, r := range line {
					if isPair(r) {
						n := r - pairBase
						text.WriteRune(n >> 7)
						text.WriteRune(n & 127)
					} else {
						text.WriteRune(r)
					}
				}
			}
		}
		return compact(text.String())
	}
	proof := map[string]any{}
	for _, code := range []string{"ja", "ko"} {
		f, err := LoadFontScaled(filepath.Join(root, "font", "pod-"+code+".golemfnt"))
		if err != nil {
			t.Fatal(err)
		}
		if f.W != 16 || f.Scaled[3].Font == nil || f.Scaled[3].Font.W != 24 {
			t.Fatal(code, "missing native16/24")
		}
		p, err := LoadLangPack(filepath.Join(root, "text"), code, f)
		if err != nil {
			t.Fatal(err)
		}
		if p.ECL.Len() != 4634 || p.Journal == nil || len(p.Journal.Entries) != 99 || p.Guide == nil || len(p.Guide.Chapters) != 78 || len(p.Map.Guide.Titles) != 32 {
			t.Fatal(code, "incomplete pack")
		}
		for _, look := range p.UI.Exact {
			c, ok := look.(*Catalog)
			if !ok {
				continue
			}
			for key, entry := range c.byHash {
				n, _ := strconv.Atoi(strings.SplitN(key, ":", 2)[0])
				if len(PackHalfWidth([]rune(entry.Translation))) > n {
					t.Errorf("%s overflow %s %d>%d: %s", code, entry.Key, len(PackHalfWidth([]rune(entry.Translation))), n, entry.Translation)
				}
			}
		}
		jp, gp, maxJ, maxG := 0, 0, 0, 0
		for n, e := range p.Journal.Entries {
			pages := JournalPages(e)
			if len(pages) == 0 || flat(pages) != compact(e.Body) {
				t.Fatal(code, "journal body loss", n)
			}
			jp += len(pages)
			maxJ = max(maxJ, len(pages))
			p.Journal.OpenAuto(n, 0)
			if !strings.Contains(flat([][][]rune{{p.Journal.Lines()[0]}}), compact(e.Title)) {
				t.Fatal(code, "journal title loss", n, e.Title)
			}
			for page := range pages {
				p.Journal.page = page
				for _, line := range p.Journal.Lines() {
					if len(line) > journalCols {
						t.Fatal(code, "journal width", n, page, len(line))
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
				t.Fatal(code, "guide body loss", n+1)
			}
			gp += len(pages)
			maxG = max(maxG, len(pages))
			if !strings.Contains(flat([][][]rune{{p.Guide.Lines()[0]}}), compact(e.Title)) {
				t.Fatal(code, "guide title loss", n+1, e.Title)
			}
			for page := range pages {
				p.Guide.page = page
				for _, line := range p.Guide.Lines() {
					if len(line) > guideCols {
						t.Fatal(code, "guide width", n+1, page, len(line))
					}
				}
			}
		}
		samplePages := map[int]int{}
		missing := func(r rune) { t.Errorf("%s missing glyph U+%04X", code, r) }
		for _, scale := range []int{2, 3} {
			for _, n := range []int{1, 11, 57, 74} {
				p.Journal.OpenAuto(n, 0)
				samplePages[n] = len(p.Journal.pages)
				for _, page := range []int{0, len(p.Journal.pages) - 1} {
					p.Journal.page = page
					pixels := make([]byte, 320*200*scale*scale*4)
					for i := 3; i < len(pixels); i += 4 {
						pixels[i] = 255
					}
					p.Journal.Draw(pixels, scale, f, missing)
					dir := filepath.Join(out, "layout")
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(dir, fmt.Sprintf("%s-journal-%03d-p%d-%dx.png", code, n, page+1, scale)))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(file, &image.NRGBA{Pix: pixels, Stride: 320 * scale * 4, Rect: image.Rect(0, 0, 320*scale, 200*scale)})
					file.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
				p.Journal.CloseAuto()
			}
		}
		proof[code] = map[string]any{"journal_entries": 99, "journal_pages": jp, "journal_max_pages": maxJ, "guide_chapters": 78, "guide_pages": gp, "guide_max_pages": maxG, "sample_pages": samplePages}
	}
	b, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "panels.json"), append(b, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
}
