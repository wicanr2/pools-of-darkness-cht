package pod

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPostersOfficialAssets(t *testing.T) {
	dir := os.Getenv("DOSGOLEM_POD_POSTERS_DIR")
	if dir == "" {
		t.Skip("未提供本機海報素材")
	}
	p, err := LoadPosters(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.text) != 2 {
		t.Fatalf("語系層數 = %d，要繁中與英文", len(p.text))
	}
	for _, scale := range []int{2, 3} {
		w, h := 320*scale, 200*scale
		game := make([]byte, w*h*4)
		for i := 0; i < len(game); i += 4 {
			game[i] = byte(i / 4)
			game[i+1] = byte(i / w)
			game[i+2] = 37
			game[i+3] = 255
		}
		before := append([]byte(nil), game...)
		var dst []byte
		var chinese []byte
		for _, locale := range []string{"zh-TW", "en", "ko", "zh-TW"} {
			dst = p.Compose(dst, game, scale, locale)
			if len(dst) != 480*scale*h*4 {
				t.Fatal("完整視窗尺寸錯誤")
			}
			for y := 0; y < h; y++ {
				if !bytes.Equal(dst[(y*480*scale+80*scale)*4:(y*480*scale+400*scale)*4], game[y*w*4:(y+1)*w*4]) {
					t.Fatalf("%dx %s 中央畫面第 %d 列改變", scale, locale, y)
				}
			}
			if locale == "zh-TW" {
				if chinese == nil {
					chinese = append([]byte(nil), dst...)
				} else if !bytes.Equal(chinese, dst) {
					t.Fatal("切回繁中未還原")
				}
			} else if bytes.Equal(chinese, dst) {
				t.Fatalf("切換 %s 後仍沿用繁中", locale)
			}
			if locale == "ko" {
				for y := 0; y < h; y++ {
					if !bytes.Equal(dst[y*480*scale*4:(y*480*scale+80*scale)*4], p.base[scale][0].Pix[y*80*scale*4:(y+1)*80*scale*4]) {
						t.Fatal("未知語系未清除舊文字")
					}
				}
			}
			if reference := os.Getenv("DOSGOLEM_POD_POSTERS_REFERENCE_DIR"); reference != "" && locale != "ko" {
				for side, x := range []int{0, 400 * scale} {
					name := []string{"left", "right"}[side]
					f, err := os.Open(filepath.Join(reference, fmt.Sprintf("poster-full-%s.%s-%dx.png", name, locale, scale)))
					if err != nil {
						t.Fatal(err)
					}
					im, err := png.Decode(f)
					f.Close()
					if err != nil {
						t.Fatal(err)
					}
					want := image.NewNRGBA(im.Bounds())
					draw.Draw(want, want.Bounds(), im, im.Bounds().Min, draw.Src)
					for y := 0; y < h; y++ {
						if !bytes.Equal(dst[(y*480*scale+x)*4:(y*480*scale+x+80*scale)*4], want.Pix[y*80*scale*4:(y+1)*80*scale*4]) {
							t.Fatalf("%s %s %dx 未保持定版字緣", locale, name, scale)
						}
					}
				}
			}
		}
		if !bytes.Equal(before, game) {
			t.Fatal("輸入畫面被修改")
		}
	}
	if p.Compose(nil, []byte{1}, 3, "zh-TW") != nil {
		t.Fatal("接受不完整畫格")
	}
	// 明確指定的素材有缺檔或錯尺寸時，不能悄悄接上部分海報。
	bad := t.TempDir()
	for _, name := range []string{"poster-left-2x.png", "poster-right-2x.png", "poster-left-3x.png", "poster-right-3x.png"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(bad, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.Remove(filepath.Join(bad, "poster-right-3x.png")); err != nil {
		t.Fatal(err)
	}
	if _, err = LoadPosters(bad); err == nil {
		t.Fatal("缺底圖仍成功")
	}
	f, err := os.Create(filepath.Join(bad, "poster-right-3x.png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, image.NewNRGBA(image.Rect(0, 0, 1, 1)))
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = LoadPosters(bad); err == nil {
		t.Fatal("錯尺寸仍成功")
	}
	data, err := os.ReadFile(filepath.Join(dir, "poster-right-3x.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(bad, "poster-right-3x.png"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(bad, "poster-text-right.ko-3x.png"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = LoadPosters(bad); err == nil {
		t.Fatal("語系層只剩單側仍成功")
	}
}
