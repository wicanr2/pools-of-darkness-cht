package pod

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/wicanr2/dosgolem/xlate"
	"os"
	"path/filepath"
	"testing"
)

// Uses real normally reached source/fade frames; no synthetic original art.
func TestIntroSourceFrames(t *testing.T) {
	root := os.Getenv("POD_INTRO_CAPTURE")
	text := os.Getenv("POD_TEXT_DIR")
	fonts := os.Getenv("POD_FONT_DIR")
	if root == "" || text == "" || fonts == "" {
		t.Skip("requires local original intro receipt and fonts")
	}
	files, err := filepath.Glob(filepath.Join(root, "source-*.idx"))
	if err != nil || len(files) == 0 {
		t.Fatal("no source frames", err)
	}
	for _, code := range []string{DefaultLang, EnglishLang} {
		file := "pod-rounded.golemfnt"
		if code == EnglishLang {
			file = "pod-en.golemfnt"
		}
		f, err := LoadFontScaled(filepath.Join(fonts, file))
		if err != nil {
			t.Fatal(err)
		}
		p, err := LoadLangPack(text, code, f)
		if err != nil {
			t.Fatal(err)
		}
		l := NewLive(f, nil, nil)
		l.SetLang(p)
		seen := map[string]bool{}
		for _, path := range files {
			idx, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			flat, err := os.ReadFile(path[:len(path)-4] + ".pal")
			if err != nil {
				t.Fatal(err)
			}
			var pal [256][3]uint8
			for i := range pal {
				copy(pal[i][:], flat[i*3:i*3+3])
			}
			cue, known := introSources[fmt.Sprintf("%x", sha256.Sum256(idx[168*320:]))]
			l.Frame(idx, pal)
			for _, scale := range []int{2, 3} {
				base := l.Compose(scale, false, nil)
				out := l.Compose(scale, true, func(r rune) { t.Errorf("%s missing U+%04X", code, r) })
				if !bytes.Equal(base[:168*320*scale*scale*4], out[:168*320*scale*scale*4]) {
					t.Fatal("changed animation")
				}
				if !known {
					if !bytes.Equal(base, out) {
						t.Fatal("unknown source changed")
					}
					continue
				}
				seen[cue.key] = true
				if pal[7][0] > 0 && bytes.Equal(base, out) {
					t.Fatalf("%s %s scale%d not translated", code, cue.key, scale)
				}
				// Missing locale and wrong palette/source must return the complete original.
				delete(p.Host, cue.key)
				if !bytes.Equal(base, l.Compose(scale, true, nil)) {
					t.Fatal("missing locale fallback")
				}
				p.Host[cue.key] = mustHost(t, text, code)[cue.key]
				old := l.pal
				l.pal[7] = [3]uint8{170, 0, 0}
				if !bytes.Equal(l.Compose(scale, false, nil), l.Compose(scale, true, nil)) {
					t.Fatal("palette fallback")
				}
				l.pal = old
				l.indexed[168*320] ^= 1
				if !bytes.Equal(l.Compose(scale, false, nil), l.Compose(scale, true, nil)) {
					t.Fatal("source fallback")
				}
				l.indexed[168*320] ^= 1
			}
		}
		if len(seen) != 8 {
			t.Fatalf("%s only %d cues", code, len(seen))
		}
		// Positive control: delay the first visible cue one presented frame.
		idx, err := os.ReadFile(filepath.Join(root, "source-007.idx"))
		if err != nil {
			t.Fatal(err)
		}
		flat, err := os.ReadFile(filepath.Join(root, "source-007.pal"))
		if err != nil {
			t.Fatal(err)
		}
		var pal [256][3]uint8
		for i := range pal {
			copy(pal[i][:], flat[i*3:i*3+3])
		}
		l.Frame(idx, pal)
		l.Flash = xlate.NewFlashAudit(320, 200, 3)
		body := p.Host["intro.001"]
		delete(p.Host, "intro.001")
		l.AuditFlash(1, 3)
		p.Host["intro.001"] = body
		l.AuditFlash(2, 3)
		if len(l.Flash.Events()) == 0 {
			t.Fatal("delayed overlay positive control was not detected")
		}
		l.Flash = xlate.NewFlashAudit(320, 200, 3)
		l.AuditFlash(1, 3)
		l.AuditFlash(2, 3)
		if len(l.Flash.Events()) != 0 {
			t.Fatal("immediate overlay flashed")
		}
		// A missing actual glyph must preserve the complete source cue.
		glyph := f.Scaled[3].Font
		var r rune
		for _, v := range []rune(body) {
			if v != ' ' && v != '\\' {
				r = v
				break
			}
		}
		g := glyph.Glyphs[r]
		delete(glyph.Glyphs, r)
		if !bytes.Equal(l.Compose(3, false, nil), l.Compose(3, true, nil)) {
			t.Fatal("missing glyph fallback")
		}
		glyph.Glyphs[r] = g
	}
}

func mustHost(t *testing.T, dir, code string) map[string]string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(dir, "host-text."+code+".tsv"))
	if e != nil {
		t.Fatal(e)
	}
	h, e := LoadHostText(b)
	if e != nil {
		t.Fatal(e)
	}
	return h
}
