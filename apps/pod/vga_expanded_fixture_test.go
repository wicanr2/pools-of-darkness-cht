package pod

import (
	"bytes"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 本批尚無正常遇怪位置；來源插入只是呈現診斷，不是正常玩家路徑收據。
func TestVGAExpandedSourceTemplates(t *testing.T) {
	_, indexed, pal, baseDir := vgaFixture(t)
	dir := os.Getenv("DOSGOLEM_POD_HD_VGA_EXPANDED_DIR")
	if dir == "" {
		t.Skip("未提供b08候選主題")
	}
	before, after := fixtureLive(t, baseDir), fixtureLive(t, dir)
	before.Frame(indexed, pal)
	after.Frame(indexed, pal)
	if len(after.actors.matches) != 8 || len(after.combat.matches) != 5 {
		t.Fatal("新候選改變既有來源比對")
	}
	if !bytes.Equal(before.Compose(3, true, nil), after.Compose(3, true, nil)) {
		t.Fatal("新候選或開場圖層改變既有戰場合成")
	}
	var manifest struct {
		Items []struct {
			Key, PNG, Indexed string
			Size              [2]int
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "actors.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	var baseline struct{ Items []struct{ Key string } }
	b, err = os.ReadFile(filepath.Join(baseDir, "actors.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &baseline); err != nil {
		t.Fatal(err)
	}
	oldKeys := map[string]bool{}
	for _, item := range baseline.Items {
		oldKeys[item.Key] = true
	}
	if len(manifest.Items) <= len(oldKeys) {
		t.Fatal("沒有追加怪物候選")
	}
	for _, item := range manifest.Items {
		if oldKeys[item.Key] {
			continue
		}
		ref, err := os.ReadFile(filepath.Join(dir, item.Indexed))
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(filepath.Join(dir, item.PNG))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		w, h := item.Size[0], item.Size[1]
		pixels := image.NewNRGBA(image.Rect(0, 0, w*3, h*3))
		draw.Draw(pixels, pixels.Bounds(), im, im.Bounds().Min, draw.Src)
		for _, flip := range []bool{false, true} {
			scene := append([]byte(nil), indexed...)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					sx := x
					if flip {
						sx = w - 1 - x
					}
					c := ref[y*w+sx]
					if c != 8 {
						scene[(80+y)*320+8+x] = c
					}
				}
			}
			after.Frame(scene, pal)
			for _, mode := range []struct {
				scale   int
				enabled bool
			}{{2, true}, {3, false}} {
				raw := after.Compose(mode.scale, mode.enabled, nil)
				for y := 0; y < h*mode.scale; y++ {
					for x := 0; x < w*mode.scale; x++ {
						c := pal[scene[(80+y/mode.scale)*320+8+x/mode.scale]]
						dst := (((80*mode.scale + y) * 320 * mode.scale) + 8*mode.scale + x) * 4
						if !bytes.Equal(raw[dst:dst+4], []byte{c[0], c[1], c[2], 255}) {
							t.Fatalf("%s鏡像%v倍率%d覆繪%v沒有回原版", item.Key, flip, mode.scale, mode.enabled)
						}
					}
				}
			}
			found := false
			for _, match := range after.actors.matches {
				if !match.conflict && match.x == 8 && match.y == 80 && after.actors.patterns[match.pattern].key == item.Key {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s鏡像%v無完整前景比對", item.Key, flip)
			}
			out := after.Compose(3, true, nil)
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					sx := x
					if flip {
						sx = w - 1 - x
					}
					if ref[y*w+sx] == 8 {
						continue
					}
					for dy := 0; dy < 3; dy++ {
						for dx := 0; dx < 3; dx++ {
							px := x*3 + dx
							if flip {
								px = w*3 - 1 - px
							}
							src := ((y*3+dy)*w*3 + px) * 4
							dst := (((80+y)*3+dy)*960 + (8+x)*3 + dx) * 4
							if !bytes.Equal(out[dst:dst+4], pixels.Pix[src:src+4]) {
								t.Fatalf("%s鏡像%v成品前景不同", item.Key, flip)
							}
						}
					}
				}
			}
			// 同一實際素材改色或少一點時，不能仍以完整候選命中。
			badPal := pal
			badPal[144][0] ^= 1
			after.Frame(scene, badPal)
			if len(after.actors.matches) != 0 {
				t.Fatalf("%s鏡像%v色盤失配仍命中", item.Key, flip)
			}
			broken := append([]byte(nil), scene...)
			for y := 0; y < h; y++ {
				changed := false
				for x := 0; x < w; x++ {
					sx := x
					if flip {
						sx = w - 1 - x
					}
					if ref[y*w+sx] != 8 {
						broken[(80+y)*320+8+x] = 8
						changed = true
						break
					}
				}
				if changed {
					break
				}
			}
			after.Frame(broken, pal)
			for _, match := range after.actors.matches {
				if match.x == 8 && match.y == 80 && after.actors.patterns[match.pattern].key == item.Key {
					t.Fatalf("%s鏡像%v缺一前景點仍完整命中", item.Key, flip)
				}
			}
			outside := make([]byte, 64000)
			for y := 0; y < h; y++ {
				copy(outside[(80+y)*320+8:(80+y)*320+8+w], scene[(80+y)*320+8:(80+y)*320+8+w])
			}
			after.Frame(outside, pal)
			if len(after.actors.matches) != 0 {
				t.Fatalf("%s鏡像%v非戰場仍命中", item.Key, flip)
			}
		}
	}
}
