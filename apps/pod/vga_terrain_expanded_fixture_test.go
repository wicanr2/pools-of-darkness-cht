package pod

import (
	"bytes"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// 主專案規格010 b26。插入原始來源只驗呈現守衛，不代表正常遭遇。
func TestVGATerrainExpandedSourceTemplates(t *testing.T) {
	baseDir, dir, capture := os.Getenv("POD_TERRAIN_BASE"), os.Getenv("POD_TERRAIN_THEME"), os.Getenv("POD_TERRAIN_CAPTURE")
	if baseDir == "" || dir == "" || capture == "" {
		t.Skip("未提供本機地形主題與正常DEMO來源")
	}
	read := func(root, name string) []byte {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	indexed, rawpal := read(capture, "source.idx"), read(capture, "source.pal")
	if len(indexed) != 64000 || len(rawpal) != 768 {
		t.Fatal("正常來源長度不合")
	}
	var pal [256][3]uint8
	for i := range pal {
		copy(pal[i][:], rawpal[i*3:i*3+3])
	}
	before, after := fixtureLive(t, baseDir), fixtureLive(t, dir)
	before.Frame(indexed, pal)
	after.Frame(indexed, pal)
	if !bytes.Equal(before.Compose(3, true, nil), after.Compose(3, true, nil)) {
		t.Fatal("新增地形改變既有正常終點合成")
	}
	// b26 defaults remain unchanged. Other banks provide their original high
	// palette and expected counts explicitly; this is source insertion only.
	expected := func(name string, fallback int) int {
		if value := os.Getenv(name); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				t.Fatal("invalid fixture count", name)
			}
			return n
		}
		return fallback
	}
	wantAdded := expected("POD_TERRAIN_EXPECTED_ADD", 23)
	wantTotal := expected("POD_TERRAIN_EXPECTED_TOTAL", 25)
	if source := os.Getenv("POD_TERRAIN_SOURCE_PALETTE"); source != "" {
		b, err := os.ReadFile(source)
		if err != nil || len(b) != 768 {
			t.Fatal("invalid source palette", source, err)
		}
		for i := 32; i < 64; i++ {
			copy(pal[i][:], b[i*3:i*3+3])
		}
	}
	// 同一已載入且驗證的素材可重用；每個反例仍建立全新的呈現狀態。
	fresh := func() *Live {
		l := NewLive(nil, nil, nil)
		l.HDLayers = after.HDLayers
		return l
	}
	type row struct {
		Key, PNG, Indexed string
		Size              [2]int
	}
	var old, current struct{ Items []row }
	if err := json.Unmarshal(read(baseDir, "combat.json"), &old); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(read(dir, "combat.json"), &current); err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, item := range old.Items {
		known[item.Key] = true
	}
	var contextRefs [][]byte
	if os.Getenv("POD_TERRAIN_SOURCE_PALETTE") != "" {
		for _, item := range current.Items {
			if !known[item.Key] && len(contextRefs) < 2 {
				ref := read(dir, item.Indexed)
				if len(ref) != 576 {
					t.Fatal("invalid context terrain")
				}
				contextRefs = append(contextRefs, ref)
			}
		}
		if len(contextRefs) != 2 {
			t.Fatal("new palette needs two supporting source tiles")
		}
	}
	added := 0
	for _, item := range current.Items {
		if known[item.Key] {
			continue
		}
		added++
		t.Run(item.Key, func(t *testing.T) {
			if item.Size != [2]int{24, 24} {
				t.Fatal("來源尺寸不合")
			}
			ref := read(dir, item.Indexed)
			if len(ref) != 576 {
				t.Fatal("來源長度不合")
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
			if im.Bounds().Dx() != 72 || im.Bounds().Dy() != 72 {
				t.Fatal("成品尺寸不合")
			}
			scene := append([]byte(nil), indexed...)
			// The unchanged production guard requires three matching tiles.
			// A new palette makes the original bank's tiles ineligible, so the
			// controlled scene supplies two authentic same-bank neighbors.
			for i, support := range contextRefs {
				for y := 0; y < 24; y++ {
					x := 32 + i*24
					copy(scene[(80+y)*320+x:(80+y)*320+x+24], support[y*24:y*24+24])
				}
			}
			for y := 0; y < 24; y++ {
				copy(scene[(80+y)*320+8:(80+y)*320+32], ref[y*24:y*24+24])
			}
			original := append([]byte(nil), scene...)
			live := fresh()
			live.Frame(scene, pal)
			has := func(l *Live) bool {
				for _, m := range l.HDLayers.combat.matches {
					if m.x == 8 && m.y == 80 && l.HDLayers.combat.tiles[m.tile].key == item.Key {
						return true
					}
				}
				return false
			}
			if !has(live) {
				t.Fatal("真實地形來源未命中")
			}
			out := live.Compose(3, true, nil)
			for y := 0; y < 72; y++ {
				for x := 0; x < 72; x++ {
					r, g, b, a := im.At(x, y).RGBA()
					want := []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)}
					o := ((240+y)*960 + 24 + x) * 4
					if !bytes.Equal(out[o:o+4], want) {
						t.Fatalf("成品像素不符 %d,%d", x, y)
					}
				}
			}
			for _, mode := range []struct {
				scale   int
				enabled bool
			}{{2, true}, {3, false}} {
				s := mode.scale
				got := live.Compose(s, mode.enabled, nil)
				for y := 0; y < 24*s; y++ {
					for x := 0; x < 24*s; x++ {
						c := pal[scene[(80+y/s)*320+8+x/s]]
						o := ((80*s+y)*320*s + 8*s + x) * 4
						if !bytes.Equal(got[o:o+4], []byte{c[0], c[1], c[2], 255}) {
							t.Fatal("2倍／F12未回原版")
						}
					}
				}
			}
			broken := append([]byte(nil), scene...)
			broken[80*320+8] ^= 255
			bad := fresh()
			bad.Frame(broken, pal)
			if has(bad) {
				t.Fatal("來源失配仍命中")
			}
			changed := pal
			changed[32][0] ^= 1
			bad = fresh()
			bad.Frame(scene, changed)
			if len(bad.HDLayers.combat.matches) != 0 {
				t.Fatal("地形色盤失配仍命中")
			}
			outside := make([]byte, 64000)
			for y := 0; y < 24; y++ {
				copy(outside[(80+y)*320+8:(80+y)*320+32], ref[y*24:y*24+24])
			}
			bad = fresh()
			bad.Frame(outside, pal)
			if len(bad.HDLayers.combat.matches) != 0 {
				t.Fatal("孤立圖案仍當作戰場")
			}
			if !bytes.Equal(scene, original) {
				t.Fatal("呈現改變來源色號")
			}
		})
	}
	if added != wantAdded || len(current.Items) != wantTotal {
		t.Fatalf("新增%d／共%d，預期%d／%d", added, len(current.Items), wantAdded, wantTotal)
	}
}
