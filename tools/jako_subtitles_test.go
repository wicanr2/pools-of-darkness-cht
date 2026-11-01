package pod

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// 以Go overlay使用既有正常片頭擷取。這是字幕元件抽驗，不是新跑完整片頭。
// POD_JAKO_ROOT指向專案，POD_JAKO_OUT指向含candidate/text、candidate/font的工作區。
func TestJakoSubtitles(t *testing.T) {
	root, out := os.Getenv("POD_JAKO_ROOT"), os.Getenv("POD_JAKO_OUT")
	if root == "" || out == "" {
		t.Skip("requires existing normally captured intro sources and ja/ko candidates")
	}
	sha := func(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	// 解開半形組合字後比較全部非空白字元，換行不能吃掉正文。
	compact := func(s string) string {
		s = strings.ReplaceAll(s, `\n`, "")
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, s)
	}
	unpack := func(lines [][]rune) string {
		var b strings.Builder
		for _, line := range lines {
			for _, r := range line {
				if isPair(r) {
					v := r - pairBase
					b.WriteRune((v >> 7) & 0x7f)
					b.WriteRune(v & 0x7f)
				} else {
					b.WriteRune(r)
				}
			}
		}
		return compact(b.String())
	}
	paths, err := filepath.Glob(filepath.Join(root, "workplace/probe/g9-hd/b44-source-v2/source-*.idx"))
	if err != nil {
		t.Fatal(err)
	}
	more, err := filepath.Glob(filepath.Join(root, "workplace/probe/g9-hd/intro-play-v1/subtitles/sources/*.idx"))
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, more...)
	if len(paths) == 0 {
		t.Fatal("missing existing normally captured intro indices")
	}
	samples := []map[string]any{}
	inputs := map[string]string{}
	proof := map[string]any{
		"scope":      "component rendering of eight existing normally captured subtitle cues; no new full intro route",
		"go_version": runtime.Version(), "languages": []string{"ja", "ko"}, "scales": []int{2, 3},
		"subtitle_start_row": 168, "subtitle_columns": 38, "subtitle_max_lines": 4,
	}
	defer func() {
		proof["passed"] = !t.Failed()
		proof["samples"], proof["inputs_sha256"] = samples, inputs
		data, err := json.MarshalIndent(proof, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		if err = os.MkdirAll(filepath.Join(out, "test"), 0755); err == nil {
			err = os.WriteFile(filepath.Join(out, "test/subtitles.json"), append(data, '\n'), 0644)
		}
		if err != nil {
			t.Error(err)
		}
	}()
	for _, lang := range []string{"ja", "ko"} {
		fontPath := filepath.Join(out, "candidate/font/pod-"+lang+".golemfnt")
		font3Path := filepath.Join(out, "candidate/font/pod-"+lang+"-3x.golemfnt")
		f, err := LoadFontScaled(fontPath)
		if err != nil {
			t.Fatal(err)
		}
		if f.W != 16 || f.H != 16 || f.Scaled[3].Font == nil || f.Scaled[3].Font.W != 24 || f.Scaled[3].Font.H != 24 {
			t.Fatal("requires native16/24 fonts", lang)
		}
		inputs["font/"+filepath.Base(fontPath)] = sha(read(fontPath))
		inputs["font/"+filepath.Base(font3Path)] = sha(read(font3Path))
		textDir := filepath.Join(out, "candidate/text")
		inputs["text/host-text."+lang+".tsv"] = sha(read(filepath.Join(textDir, "host-text."+lang+".tsv")))
		p, err := LoadLangPack(textDir, lang, f)
		if err != nil {
			t.Fatal(err)
		}
		l := NewLive(f, nil, nil)
		l.SetLang(p)
		seen := map[string]bool{}
		for _, path := range paths {
			idx := read(path)
			if len(idx) != 64000 {
				t.Fatal("invalid source index length", path, len(idx))
			}
			flat := read(strings.TrimSuffix(path, ".idx") + ".pal")
			if len(flat) != 768 {
				t.Fatal("invalid source palette length", path, len(flat))
			}
			var pal [256][3]uint8
			for i := range pal {
				copy(pal[i][:], flat[3*i:3*i+3])
			}
			cue, known := introSources[sha(idx[168*320:])]
			if !known || seen[cue.key] || pal[7][0] == 0 || pal[7][0] > 170 || pal[7][0] != pal[7][1] || pal[7][0] != pal[7][2] || pal[cue.background] != [3]uint8{} {
				continue
			}
			seen[cue.key] = true
			body := p.Host[cue.key]
			lines := wrapCols(body, 38)
			widths, packedCells := []int{}, 0
			fits := len(lines) > 0 && len(lines) <= 4
			for _, line := range lines {
				widths = append(widths, len(line))
				packedCells += len(line)
				fits = fits && len(line) <= 38
			}
			complete := compact(body) != "" && unpack(lines) == compact(body)
			if !fits || !complete {
				t.Errorf("%s %s: raw=%d runes packed=%d cells lines=%v complete=%v fits=%v; no translation changed", lang, cue.key, utf8.RuneCountInString(body), packedCells, widths, complete, fits)
			}
			l.Frame(idx, pal)
			rel, err := filepath.Rel(root, path)
			if err != nil {
				t.Fatal(err)
			}
			for _, scale := range []int{2, 3} {
				missing := 0
				base := l.Compose(scale, false, nil)
				got := l.Compose(scale, true, func(r rune) {
					missing++
					t.Errorf("%s %s %dx missing U+%04X", lang, cue.key, scale, r)
				})
				animationSame := bytes.Equal(base[:168*320*scale*scale*4], got[:168*320*scale*scale*4])
				overlaid := !bytes.Equal(base, got)
				if !animationSame || !overlaid {
					t.Errorf("%s %s %dx subtitle present=%v animation unchanged=%v", lang, cue.key, scale, overlaid, animationSame)
				}
				delete(p.Host, cue.key)
				textFallback := bytes.Equal(base, l.Compose(scale, true, nil))
				p.Host[cue.key] = body
				if !textFallback {
					t.Error("missing translation did not preserve complete source", lang, cue.key, scale)
				}
				font, _, _, _ := f.ForScale(scale)
				var removed rune
				for _, line := range lines {
					for _, r := range line {
						if r != ' ' {
							removed = r
							break
						}
					}
					if removed != 0 {
						break
					}
				}
				glyph, exists := font.Glyphs[removed]
				if !exists {
					t.Fatal("negative control glyph was already missing", lang, cue.key)
				}
				delete(font.Glyphs, removed)
				glyphFallback := bytes.Equal(base, l.Compose(scale, true, nil))
				font.Glyphs[removed] = glyph
				if !glyphFallback {
					t.Error("missing actual glyph did not preserve complete source", lang, cue.key, scale)
				}
				sample := map[string]any{
					"language": lang, "cue": cue.key, "scale": scale, "source": filepath.ToSlash(rel),
					"source_indexed_sha256": sha(idx), "source_palette_sha256": sha(flat), "text_sha256": sha([]byte(body)),
					"raw_runes": utf8.RuneCountInString(body), "packed_cells": packedCells, "line_cells": widths,
					"all_text_preserved": complete, "fits_complete_cue": fits, "animation_168_rows_unchanged": animationSame,
					"subtitle_present": overlaid, "missing_glyphs": missing, "missing_text_fallback": textFallback,
					"missing_glyph_fallback": glyphFallback, "rgba_sha256": sha(got),
				}
				if scale == 3 {
					name := "subtitle-samples/" + lang + "." + strings.ReplaceAll(cue.key, ".", "") + ".png"
					if err := os.MkdirAll(filepath.Dir(filepath.Join(out, name)), 0755); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(out, name))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(file, &image.NRGBA{Pix: got, Stride: 960 * 4, Rect: image.Rect(0, 0, 960, 600)})
					closeErr := file.Close()
					if err != nil || closeErr != nil {
						t.Fatal("saving subtitle sample", err, closeErr)
					}
					sample["file"], sample["png_sha256"] = name, sha(read(filepath.Join(out, name)))
				}
				samples = append(samples, sample)
			}
		}
		if len(seen) != 8 {
			t.Errorf("%s only %d of8 existing normally captured cues", lang, len(seen))
		}
	}
	if len(samples) != 32 {
		t.Errorf("only %d of32 language/scale/cue checks", len(samples))
	}
	t.Log("ja/ko eight existing subtitle sources at2/3x; complete wrapping, animation unchanged, no missing glyphs and missing-text/glyph fallbacks; component scope only")
}
