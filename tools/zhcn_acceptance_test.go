package pod

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 透過Go overlay加入固定乾淨引擎，僅檢查既有正常片頭來源，不改產品程式。
func TestZhCNAcceptance(t *testing.T) {
	root, out := os.Getenv("POD_ZHCN_ROOT"), os.Getenv("POD_ZHCN_OUT")
	if root == "" || out == "" {
		t.Skip("requires local normally captured original frames")
	}
	f, err := LoadFontScaled(filepath.Join(out, "font/pod-zh-CN.golemfnt"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := LoadLangPack(filepath.Join(root, "text"), "zh-CN", f)
	if err != nil {
		t.Fatal(err)
	}
	if p.Journal == nil || len(p.Journal.Entries) != 99 || p.Guide == nil || len(p.Guide.Chapters) != 78 || p.Map == nil || p.UI.Players == nil {
		t.Fatal("incomplete zh-CN pack")
	}
	l := NewLive(f, nil, nil)
	l.SetLang(p)
	seen := map[string]bool{}
	paths, _ := filepath.Glob(filepath.Join(root, "workplace/probe/g9-hd/b44-source-v2/source-*.idx"))
	more, _ := filepath.Glob(filepath.Join(root, "workplace/probe/g9-hd/intro-play-v1/subtitles/sources/*.idx"))
	paths = append(paths, more...)
	for _, path := range paths {
		idx, err := os.ReadFile(path)
		if err != nil || len(idx) != 64000 {
			t.Fatal("source indices", err)
		}
		flat, err := os.ReadFile(path[:len(path)-4] + ".pal")
		if err != nil || len(flat) != 768 {
			t.Fatal("source palette", err)
		}
		var pal [256][3]uint8
		for i := range pal {
			copy(pal[i][:], flat[3*i:3*i+3])
		}
		cue, known := introSources[fmt.Sprintf("%x", sha256.Sum256(idx[168*320:]))]
		if !known || seen[cue.key] || pal[7][0] == 0 || pal[7][0] > 170 || pal[7][0] != pal[7][1] || pal[7][0] != pal[7][2] || pal[cue.background] != [3]uint8{} {
			continue
		}
		l.Frame(idx, pal)
		for _, scale := range []int{2, 3} {
			base := l.Compose(scale, false, nil)
			got := l.Compose(scale, true, func(r rune) { t.Errorf("missing U+%04X", r) })
			if bytes.Equal(base, got) || !bytes.Equal(base[:168*320*scale*scale*4], got[:168*320*scale*scale*4]) {
				t.Fatal("subtitle absent or animation changed", cue.key, scale)
			}
			body := p.Host[cue.key]
			delete(p.Host, cue.key)
			if !bytes.Equal(base, l.Compose(scale, true, nil)) {
				t.Fatal("missing translation did not fall back")
			}
			p.Host[cue.key] = body
			if scale == 3 {
				dir := filepath.Join(out, "subtitle-samples")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				file, err := os.Create(filepath.Join(dir, cue.key+".png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(file, &image.NRGBA{Pix: got, Stride: 960 * 4, Rect: image.Rect(0, 0, 960, 600)})
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
		seen[cue.key] = true
	}
	if len(seen) != 8 {
		t.Fatalf("only %d of8 normally captured cues", len(seen))
	}
	t.Log("99 journals,78 guide chapters,map and transliteration loaded;8 subtitle cues at2/3x;missing-text negative control")
}
