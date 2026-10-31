package pod

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func encounterFixture(t *testing.T) (string, string) {
	t.Helper()
	dir, req := os.Getenv("DOSGOLEM_POD_ENCOUNTER_THEME"), os.Getenv("DOSGOLEM_POD_ENCOUNTER_REQUEST")
	if dir == "" || req == "" {
		t.Skip("未提供b41實際SPRIT來源與成品")
	}
	return dir, req
}

func encounterFixtureScene(ref []byte, background byte) []byte {
	idx := bytes.Repeat([]byte{background}, 64000)
	for n, v := range ref {
		if v != 0 {
			idx[(24+n/88)*320+24+n%88] = v
		}
	}
	return idx
}

func encounterFixtureExpected(idx []byte, pal [256][3]uint8, ref, pixels []byte) []byte {
	out := vgaBaseRGBA(idx, pal, 3)
	for n, v := range ref {
		if v == 0 {
			continue
		}
		for row := 0; row < 3; row++ {
			src := ((n/88*3+row)*264 + n%88*3) * 4
			dst := (((24+n/88)*3+row)*960 + (24+n%88)*3) * 4
			copy(out[dst:dst+12], pixels[src:src+12])
		}
	}
	return out
}

func TestVGAEncounterThreeDistancesSourceAndComposition(t *testing.T) {
	dir, req := encounterFixture(t)
	h, err := loadEncounterHD(dir)
	if err != nil || h == nil || len(h.items) != 3 {
		t.Fatal("三種距離載入", err)
	}
	for _, item := range h.items {
		t.Run(item.key, func(t *testing.T) {
			ref, err := os.ReadFile(filepath.Join(dir, item.key+".idx"))
			if err != nil {
				t.Fatal(err)
			}
			pal := backgroundPalette(t, req, item.key)
			live := fixtureLive(t, dir)
			if live.deferredUnsafe {
				t.Fatal("遭遇圖錯誤停用直接合成")
			}
			for _, bg := range []byte{32, 79} {
				idx := encounterFixtureScene(ref, bg)
				original := append([]byte(nil), idx...)
				live.Frame(idx, pal)
				if len(live.encounters.matches) != 1 || live.encounters.items[live.encounters.matches[0]].key != item.key {
					t.Fatal("完整來源未唯一辨識")
				}
				want := encounterFixtureExpected(idx, pal, ref, item.pixels)
				if !bytes.Equal(live.Compose(3, true, nil), want) {
					t.Fatal("完整PNG前景、零區或框外像素不同")
				}
				live.Direct = &DirectCanvas{}
				if !bytes.Equal(live.Compose(3, true, nil), want) {
					t.Fatal("直接合成不同於獨立預期")
				}
				if live.Direct.stats.LegacyRGBFrames != 0 || live.Direct.stats.SelectedRGBPixels == 0 {
					t.Fatal("直接合成未先選HD")
				}
				live.Direct = nil
				if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(idx, pal, 2)) || !bytes.Equal(live.Compose(3, false, nil), vgaBaseRGBA(idx, pal, 3)) {
					t.Fatal("2倍或F12有HD")
				}
				if !bytes.Equal(idx, original) {
					t.Fatal("呈現改動原版來源")
				}
				changed := pal
				changed[item.used[0]][0] ^= 1
				live.Frame(idx, changed)
				if len(live.encounters.matches) != 0 || !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, changed, 3)) {
					t.Fatal("使用RGB失配仍留HD")
				}
				changed = pal
				changed[144][0] ^= 1
				live.Frame(idx, changed)
				if !bytes.Equal(live.Compose(3, true, nil), encounterFixtureExpected(idx, changed, ref, item.pixels)) {
					t.Fatal("未使用RGB停用遭遇圖")
				}
				// 在舊背景中的透明區改一點，不得影響前景辨識或覆蓋它。
				idx[24*320+24] = 47
				live.Frame(idx, pal)
				if !bytes.Equal(live.Compose(3, true, nil), encounterFixtureExpected(idx, pal, ref, item.pixels)) {
					t.Fatal("來源0區改動被覆寫")
				}
				pt := item.ink[len(item.ink)/2]
				idx[pt.offset] ^= 255
				live.Frame(idx, pal)
				if len(live.encounters.matches) != 0 {
					t.Fatal("一點前景失配未撤除候選")
				}
				got, base := live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)
				cx, cy := (24+pt.x)/8*8, (24+pt.y)/8*8
				for row := cy * 3; row < (cy+8)*3; row++ {
					off := (row*960 + cx*3) * 4
					if !bytes.Equal(got[off:off+96], base[off:off+96]) {
						t.Fatal("改動格仍有HD")
					}
				}
				cold := fixtureLive(t, dir)
				cold.Frame(idx, pal)
				if !bytes.Equal(cold.Compose(3, true, nil), base) {
					t.Fatal("冷失配仍繪HD")
				}
				// 搬到未證實位置時整個候選撤除，不推測座標。
				moved := bytes.Repeat([]byte{bg}, 64000)
				for n, v := range ref {
					if v != 0 {
						moved[(32+n/88)*320+32+n%88] = v
					}
				}
				cold.Frame(moved, pal)
				if len(cold.encounters.matches) != 0 {
					t.Fatal("未證實位置仍辨識")
				}
				clear(idx)
				live.Frame(idx, pal)
				if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, pal, 3)) {
					t.Fatal("清除後殘留")
				}
			}
			// 重疊來源只供診斷，全部回退。
			conflict := &encounterHD{items: []encounterItem{item, item}}
			idx := encounterFixtureScene(ref, 32)
			conflict.frame(idx, pal)
			if len(conflict.matches) != 0 {
				t.Fatal("重疊未全部回退")
			}
		})
	}
}

func TestVGAEncounterLoaderRejectsSubstitution(t *testing.T) {
	dir, _ := encounterFixture(t)
	raw, err := os.ReadFile(filepath.Join(dir, "encounters.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	for _, v := range manifest["items"].([]any) {
		it := v.(map[string]any)
		for _, field := range []string{"png", "indexed"} {
			name := it[field].(string)
			data, e := os.ReadFile(filepath.Join(dir, name))
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(tmp, name), data, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, kind := range []string{"unknown", "source", "size", "position", "duplicate"} {
		t.Run(kind, func(t *testing.T) {
			var bad map[string]any
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			items := bad["items"].([]any)
			it := items[0].(map[string]any)
			switch kind {
			case "unknown":
				it["key"] = "sprit1-999@0"
			case "source":
				other := items[1].(map[string]any)
				it["indexed"], it["indexed_sha256"] = other["indexed"], other["indexed_sha256"]
			case "size":
				it["size"] = []int{88, 44}
			case "position":
				it["position"] = []int{16, 24}
			case "duplicate":
				bad["items"] = append(items, it)
			}
			encoded, e := json.Marshal(bad)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(tmp, "encounters.json"), encoded, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e = loadEncounterHD(tmp); e == nil {
				t.Fatal("失配遭遇圖未拒收")
			}
		})
	}
}
