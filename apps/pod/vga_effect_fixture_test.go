package pod

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// 原版檔缺少時明確skip；合成來源只證明呈現，不聲稱正常施法事件。
func TestVGAEffectSourceFixture(t *testing.T) {
	dir := os.Getenv("DOSGOLEM_POD_EFFECT_DIR")
	if dir == "" {
		t.Skip("未提供首批COMSPR候選")
	}
	a, err := loadEffectHD(dir)
	if err != nil {
		t.Fatal(err)
	}
	if a == nil || a.sources != 6 {
		t.Fatal("首批特效應為六個來源")
	}
	a.effectPaletteOK = true
	for _, key := range []string{"comspr-000", "comspr-005", "comspr-006", "comspr-009", "comspr-010", "comspr-013"} {
		t.Run(key, func(t *testing.T) {
			for _, p := range a.patterns {
				if p.key != key {
					continue
				}
				for _, pos := range [][2]int{{8, 80}, {9, 81}} {
					x, y := pos[0], pos[1]
					scene := bytes.Repeat([]byte{8}, 64000)
					for row := 0; row < 24; row++ {
						copy(scene[(y+row)*320+x:], p.full[row*24:(row+1)*24])
					}
					original := append([]byte(nil), scene...)
					a.frame(scene, true)
					ambiguous := len(a.matches) > 1
					if len(a.matches) == 0 {
						t.Fatal("完整來源未辨識")
					}
					if ambiguous {
						if key != "comspr-000" {
							t.Fatal("非預期來源歧義")
						}
						for _, m := range a.matches {
							if !m.conflict {
								t.Fatal("歧義候選未全部回退")
							}
						}
						t.Log("箭矢鏡像與相鄰位置歧義，安全保留原版")
					} else if a.matches[0].conflict {
						t.Fatal("唯一來源被標記歧義")
					}
					out := bytes.Repeat([]byte{17}, 960*600*4)
					baseline := append([]byte(nil), out...)
					a.draw(out, scene, 3)
					for yy := 0; yy < 600; yy++ {
						for xx := 0; xx < 960; xx++ {
							o := (yy*960 + xx) * 4
							want := baseline[o : o+4]
							if xx >= x*3 && xx < (x+24)*3 && yy >= y*3 && yy < (y+24)*3 {
								px, py := xx-x*3, yy-y*3
								if !ambiguous && p.full[(py/3)*24+px/3] != 8 {
									s := (py*72 + px) * 4
									want = p.pixels[s : s+4]
								}
							}
							if !bytes.Equal(out[o:o+4], want) {
								t.Fatal("HD像素或背景8／矩形外改變")
							}
						}
					}
					for _, scale := range []int{2, 3} {
						if scale == 3 {
							a.effectPaletteOK = false
						}
						got := append([]byte(nil), baseline...)
						a.draw(got, scene, scale)
						if !bytes.Equal(got, baseline) {
							t.Fatal("2倍或失配色盤仍畫HD")
						}
					}
					a.frame(scene, true)
					if len(a.matches) != 0 {
						t.Fatal("失配色盤仍辨識")
					}
					a.effectPaletteOK = true
					a.frame(scene, false)
					if len(a.matches) != 0 {
						t.Fatal("非戰場仍辨識")
					}
					point := p.ink[0]
					scene[(y+point.y)*320+x+point.x] ^= 255
					a.frame(scene, true)
					if len(a.matches) != 0 {
						t.Fatal("少一點仍完整辨識")
					}
					a.frame(bytes.Repeat([]byte{8}, 64000), true)
					if len(a.matches) != 0 {
						t.Fatal("清除後殘留")
					}
					scene[(y+point.y)*320+x+point.x] ^= 255
					if !bytes.Equal(original, scene) {
						t.Fatal("輸入被修改")
					}
				}
			}
		})
	}
}

func TestVGAEffectNormalFixture(t *testing.T) {
	dir, capture := os.Getenv("DOSGOLEM_POD_EFFECT_DIR"), os.Getenv("DOSGOLEM_POD_EFFECT_CAPTURE")
	if dir == "" || capture == "" {
		t.Skip("未提供正常DEMO來源")
	}
	read := func(name string) []byte {
		b, e := os.ReadFile(filepath.Join(capture, name))
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	idx, raw := read("source.idx"), read("source.pal")
	if len(idx) != 64000 || len(raw) != 768 {
		t.Fatal("正常來源長度不符")
	}
	var pal [256][3]uint8
	for n := range pal {
		copy(pal[n][:], raw[n*3:n*3+3])
	}
	live := fixtureLive(t, dir)
	check := func(p [256][3]uint8, want bool) {
		live.Frame(idx, p)
		found := false
		for _, m := range live.actors.matches {
			if live.actors.patterns[m.pattern].key == "comspr-009" && m.x == 56 && m.y == 32 && !m.conflict {
				found = true
			}
		}
		if found != want {
			t.Fatalf("009正常來源辨識%v，預期%v", found, want)
		}
	}
	check(pal, true)
	bad := pal
	bad[64][0] ^= 1
	check(bad, false)
	if !live.actors.vgaPaletteOK {
		t.Fatal("特效色盤失配波及CPIC色盤")
	}
	bad = pal
	bad[144][0] ^= 1
	check(bad, true)
	if live.actors.vgaPaletteOK {
		t.Fatal("CPIC負對照未失配")
	}
	check(pal, true)
	if !bytes.Equal(live.Compose(3, false, nil), vgaBaseRGBA(idx, pal, 3)) {
		t.Fatal("F12未回原版")
	}
	if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(idx, pal, 2)) {
		t.Fatal("2倍未回原版")
	}
}

func TestVGAEffectKnownIDsAndColors(t *testing.T) {
	for _, key := range []string{"comspr-000", "comspr-021", "comspr-128", "comspr-149"} {
		if ok, err := vgaEffectReference(key, []byte{8, 64, 95}); !ok || err != nil {
			t.Fatal(key, err)
		}
	}
	for _, key := range []string{"comspr-022", "comspr-127", "comspr-150", "comspr-1", "cpic1-009"} {
		if _, err := vgaEffectReference(key, []byte{64}); err == nil {
			t.Fatal("未知ID通過", key)
		}
	}
	for _, c := range []byte{16, 63, 96, 144, 255} {
		if _, err := vgaEffectReference("comspr-009", []byte{64, c}); err == nil {
			t.Fatal("越界色號通過", c)
		}
	}
	if vga, err := vgaEffectReference("legacy", []byte{0, 8, 15}); vga || err != nil {
		t.Fatal("EGA相容性改變")
	}
}
