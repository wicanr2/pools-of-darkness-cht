package pod

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func backgroundFixture(t *testing.T) (string, string) {
	t.Helper()
	dir, req := os.Getenv("DOSGOLEM_POD_BACKGROUND_THEME"), os.Getenv("DOSGOLEM_POD_BACKGROUND_REQUEST")
	if dir == "" || req == "" {
		t.Skip("未提供b40真實背景來源與候選主題")
	}
	return dir, req
}

func backgroundPalette(t *testing.T, req, key string) (pal [256][3]uint8) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(req, key+".pal"))
	if err != nil || len(raw) != 768 {
		t.Fatal("背景色盤", err)
	}
	for n := range pal {
		copy(pal[n][:], raw[n*3:n*3+3])
	}
	return
}

func backgroundScene(ref []byte, x, y int) []byte {
	idx := make([]byte, 64000)
	for row := 0; row < 88; row++ {
		copy(idx[(y+row)*320+x:], ref[row*88:(row+1)*88])
	}
	return idx
}

func backgroundExpected(idx []byte, pal [256][3]uint8, pixels []byte, x, y int) []byte {
	out := vgaBaseRGBA(idx, pal, 3)
	for row := 0; row < 264; row++ {
		copy(out[((y*3+row)*960+x*3)*4:], pixels[row*264*4:(row+1)*264*4])
	}
	return out
}

func TestVGABackgroundSixSourceClassesAndPaletteIsolation(t *testing.T) {
	dir, req := backgroundFixture(t)
	h, err := loadExactHD(dir)
	if err != nil || h == nil || len(h.items) != 9 {
		t.Fatal("九個完整背景載入", err)
	}
	for _, key := range []string{"back1-001", "back1-003", "back1-005", "back1-006", "skygrnd-002", "skygrnd-004"} {
		t.Run(key, func(t *testing.T) {
			var p *exactHDItem
			for n := range h.items {
				if h.items[n].key == key {
					p = &h.items[n]
				}
			}
			if p == nil {
				t.Fatal("缺來源")
			}
			pal := backgroundPalette(t, req, key)
			live := fixtureLive(t, dir)
			for _, pos := range [][2]int{{24, 24}, {25, 25}} {
				x, y := pos[0], pos[1]
				idx := backgroundScene(p.ref, x, y)
				original := append([]byte(nil), idx...)
				live.Frame(idx, pal)
				if !bytes.Equal(live.Compose(3, true, nil), backgroundExpected(idx, pal, p.pixels, x, y)) {
					t.Fatal("完整成品、位置、移動清除或框外像素不同")
				}
				if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(idx, pal, 2)) || !bytes.Equal(live.Compose(3, false, nil), vgaBaseRGBA(idx, pal, 3)) {
					t.Fatal("2倍或F12仍有HD")
				}
				changed := pal
				changed[p.ref[0]][0] ^= 1
				live.Frame(idx, changed)
				if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, changed, 3)) {
					t.Fatal("使用RGB改變後未撤除HD")
				}
				changed = pal
				changed[0][0] ^= 1 // 所有背景只使用144–175。
				live.Frame(idx, changed)
				if !bytes.Equal(live.Compose(3, true, nil), backgroundExpected(idx, changed, p.pixels, x, y)) {
					t.Fatal("未使用RGB錯誤停用背景")
				}
				if !bytes.Equal(idx, original) {
					t.Fatal("呈現改動來源色號")
				}
				idx[(y+10)*320+x+10] ^= 255
				live.Frame(idx, pal)
				if len(live.HDLayers.exact.matches) != 0 {
					t.Fatal("一點失配後仍完整辨識")
				}
				got, base := live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)
				cx, cy := (x+10)/8*8, (y+10)/8*8
				for row := cy * 3; row < (cy+8)*3; row++ {
					off := (row*960 + cx*3) * 4
					if !bytes.Equal(got[off:off+24*4], base[off:off+24*4]) {
						t.Fatal("改動的8×8格仍殘留HD")
					}
				}
				cold := fixtureLive(t, dir)
				cold.Frame(idx, pal)
				if !bytes.Equal(cold.Compose(3, true, nil), base) {
					t.Fatal("未曾完整辨識的背景仍繪HD")
				}
				clear(idx)
				live.Frame(idx, pal)
				if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)) {
					t.Fatal("清除後殘留HD")
				}
			}
		})
	}
	// 同索引的兩種天空不能互相衝突或誤選；實際RGB先決定候選。
	for _, key := range []string{"skygrnd-002", "skygrnd-003"} {
		ref, err := os.ReadFile(filepath.Join(dir, key+".idx"))
		if err != nil {
			t.Fatal(err)
		}
		h.frame(backgroundScene(ref, 24, 24), backgroundPalette(t, req, key))
		if len(h.matches) != 1 || h.matches[0].conflict || h.items[h.matches[0].item].key != key {
			t.Fatal("相同索引的天空未依RGB唯一辨識")
		}
	}
}

func TestVGABackgroundLoaderRejectsSubstitution(t *testing.T) {
	dir, _ := backgroundFixture(t)
	raw, err := os.ReadFile(filepath.Join(dir, "exact.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Version int              `json:"version"`
		Scale   int              `json:"scale"`
		Kind    string           `json:"kind"`
		Items   []map[string]any `json:"items"`
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	for _, it := range m.Items {
		for _, field := range []string{"png", "indexed"} {
			name := it[field].(string)
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(tmp, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, kind := range []string{"unknown", "substitute", "size", "algorithm", "duplicate"} {
		t.Run(kind, func(t *testing.T) {
			var bad map[string]any
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			items := bad["items"].([]any)
			it := items[0].(map[string]any)
			switch kind {
			case "unknown":
				it["key"] = "background-unknown"
			case "substitute":
				// 保持正確檔案SHA，換成另一個真實來源，仍須拒收。
				other := items[3].(map[string]any)
				it["indexed"], it["indexed_sha256"] = other["indexed"], other["indexed_sha256"]
			case "size":
				it["size"] = []int{44, 176}
			case "algorithm":
				it["algorithm"] = "scale3x"
			case "duplicate":
				bad["items"] = append(items, it)
			}
			encoded, err := json.Marshal(bad)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(tmp, "exact.json"), encoded, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = loadExactHD(tmp); err == nil {
				t.Fatal("不合法背景未拒收")
			}
		})
	}
}

func TestExactLegacyEGAAndOverlapFallback(t *testing.T) {
	ref := make([]byte, 64)
	for n := range ref {
		ref[n] = byte(n%7 + 1)
	}
	p := exactHDItem{key: "legacy", w: 8, h: 8, ref: ref, pixels: bytes.Repeat([]byte{9}, 24*24*4)}
	p.makeAnchor()
	h := &exactHD{items: []exactHDItem{p}}
	idx := make([]byte, 64000)
	for y := 0; y < 8; y++ {
		copy(idx[(24+y)*320+24:], ref[y*8:(y+1)*8])
	}
	var pal [256][3]uint8
	copy(pal[:16], combatEGA[:])
	h.frame(idx, pal)
	if len(h.matches) != 1 || h.matches[0].conflict {
		t.Fatal("EGA來源行為不同")
	}
	pal[0][0] ^= 1
	h.frame(idx, pal)
	if len(h.matches) != 0 {
		t.Fatal("EGA色盤守衛失效")
	}
	pal[0][0] ^= 1
	// 注入重疊候選只供診斷；全數回退，不挑任一個圖。
	h.items = append(h.items, p)
	h.frame(idx, pal)
	if len(h.matches) != 2 {
		t.Fatal("缺重疊診斷候選")
	}
	out := bytes.Repeat([]byte{17}, 960*600*4)
	before := append([]byte(nil), out...)
	h.draw(out, idx, 3)
	if !h.matches[0].conflict || !h.matches[1].conflict || !bytes.Equal(out, before) {
		t.Fatal("重疊未全部回退")
	}
	// 低色號也不能借用已保留的VGA背景key。
	if _, err := exactBackgroundPalette("back1-001", ref, [2]int{8, 8}, ""); err == nil {
		t.Fatal("VGA key接受EGA替代來源")
	}
	if palette, err := exactBackgroundPalette("legacy", ref, [2]int{8, 8}, ""); err != nil || palette != "" {
		t.Fatal("既有EGA來源被拒收")
	}
}
