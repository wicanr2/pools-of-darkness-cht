package pod

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 輸入為正常讀檔路線的最後原版畫格；這裡只驗呈現，不把直入當正常試玩。
func TestVGAWallSourceFixture(t *testing.T) {
	dir, capture := os.Getenv("DOSGOLEM_POD_HD_VGA_WALL_DIR"), os.Getenv("DOSGOLEM_POD_HD_VGA_WALL_CAPTURE")
	if dir == "" || capture == "" {
		t.Skip("未提供本機VGA街景主題與正常來源")
	}
	read := func(path string) []byte {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	idx, rawpal := read(filepath.Join(capture, "source.idx")), read(filepath.Join(capture, "source.pal"))
	if len(idx) != 64000 || len(rawpal) != 768 {
		t.Fatal("原版來源長度")
	}
	initial := append([]byte(nil), idx...)
	var pal [256][3]uint8
	for i := range pal {
		copy(pal[i][:], rawpal[i*3:i*3+3])
	}
	l := fixtureLive(t, dir)
	l.Frame(idx, pal)
	if l.walls == nil || !l.walls.active || l.walls.sources != 18 {
		t.Fatal("未載入18個VGA街景來源")
	}
	for pos, c := range l.walls.matches {
		if c == nil {
			t.Fatalf("完整來源缺格%d", pos)
		}
	}
	var m wallManifest
	if err := json.Unmarshal(read(filepath.Join(dir, "walls.json")), &m); err != nil {
		t.Fatal(err)
	}
	assets := map[string]image.Image{}
	refs := map[string][]byte{}
	for _, a := range m.Assets {
		im, err := png.Decode(bytes.NewReader(read(filepath.Join(dir, a.PNG))))
		if err != nil {
			t.Fatal(err)
		}
		assets[a.Key] = im
		refs[a.Key] = read(filepath.Join(dir, a.Indexed))
	}
	out := l.Compose(3, true, nil)
	for _, c := range m.Cells {
		if len(c.Positions) != 1 {
			t.Fatal("診斷只驗明列位置")
		}
		pos := c.Positions[0]
		cx, cy := pos%11, pos/11
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				want := color.NRGBAModel.Convert(assets["skygrnd-001"].At(cx*24+x, cy*24+y)).(color.NRGBA)
				if len(c.Layers) == 2 {
					key := c.Layers[1].Key
					if refs[key][y/3*8+x/3] != 13 {
						want = color.NRGBAModel.Convert(assets[key].At(x, y)).(color.NRGBA)
					}
				}
				dst := (((24+cy*8)*3+y)*960 + (24+cx*8)*3 + x) * 4
				if !bytes.Equal(out[dst:dst+4], []byte{want.R, want.G, want.B, want.A}) {
					t.Fatalf("組合格%d %d,%d不同", pos, x, y)
				}
			}
		}
	}
	for _, mode := range []struct {
		s       int
		enabled bool
	}{{2, true}, {3, false}} {
		raw := l.Compose(mode.s, mode.enabled, nil)
		for y := 24 * mode.s; y < 112*mode.s; y++ {
			for x := 24 * mode.s; x < 112*mode.s; x++ {
				c := pal[idx[y/mode.s*320+x/mode.s]]
				dst := (y*320*mode.s + x) * 4
				if !bytes.Equal(raw[dst:dst+4], []byte{c[0], c[1], c[2], 255}) {
					t.Fatal("2倍或F12沒有回到原版")
				}
			}
		}
	}
	for _, index := range []int{6, 32, 144} {
		changed := pal
		changed[index][0] ^= 1
		l.Frame(idx, changed)
		if l.walls.active {
			t.Fatalf("色盤%d改變仍畫街景", index)
		}
	}
	l.Frame(idx, pal)
	broken := append([]byte(nil), idx...)
	broken[24*320+24] ^= 1
	l.Frame(broken, pal)
	if l.walls.matches[0] != nil {
		t.Fatal("被改寫格沒有撤除")
	}
	bad := *l.walls
	bad.vgaAssets = append([]wallAsset(nil), bad.vgaAssets...)
	for n, a := range bad.vgaAssets {
		if a.wall {
			a.pix = append([]byte(nil), a.pix...)
			a.pix[0] ^= 1
			bad.vgaAssets[n] = a
			break
		}
	}
	bad.frame(idx, pal)
	if bad.active {
		t.Fatal("PNG接縫失配仍畫街景")
	}
	// 單色格沒有足夠資訊，不可單獨啟用。
	var constant [64]byte
	for n := range constant {
		constant[n] = 144
	}
	h := &wallHD{vga: true}
	h.cells[0] = map[[64]byte]*wallCell{constant: {ref: constant}}
	blank := make([]byte, 64000)
	for y := 24; y < 32; y++ {
		for x := 24; x < 32; x++ {
			blank[y*320+x] = 144
		}
	}
	h.frame(blank, pal)
	if h.active || h.matches[0] != nil {
		t.Fatal("單色非街景誤命中")
	}
	if !bytes.Equal(initial, idx) {
		t.Fatal("來源色號被改寫")
	}
	if path := os.Getenv("DOSGOLEM_POD_HD_VGA_WALL_SHOT"); path != "" {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err = png.Encode(f, &image.NRGBA{Pix: out, Stride: 960 * 4, Rect: image.Rect(0, 0, 960, 600)}); err != nil {
			t.Fatal(err)
		}
	}
}
