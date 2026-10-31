package pod

// 規格010 b39：僅透過go test -overlay加入，不進正式程式。
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestVGAEffectBatchSourceFixture(t *testing.T) {
	dir, proofPath := os.Getenv("DOSGOLEM_POD_EFFECT_DIR"), os.Getenv("DOSGOLEM_POD_EFFECT_BATCH_PROOF")
	if dir == "" || proofPath == "" {
		t.Skip("未提供完整COMSPR候選與固定來源預期")
	}
	data, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		Sampled []string `json:"sampled_keys"`
		Cases   []struct {
			Key     string
			Flip    bool
			X, Y    int
			Matches []struct {
				Key      string
				SHA      string `json:"indexed_sha256"`
				X, Y     int
				Conflict bool
			}
		} `json:"diagnostic_cases"`
	}
	if err := json.Unmarshal(data, &proof); err != nil {
		t.Fatal(err)
	}
	if len(proof.Sampled) != 6 || len(proof.Cases) != 24 {
		t.Fatal("必須使用事先固定的六類24種方向／位置")
	}
	a, err := loadEffectHD(dir)
	if err != nil {
		t.Fatal(err)
	}
	if a == nil || a.sources != 36 {
		t.Fatal("完整COMSPR應為36個去重來源")
	}
	for _, key := range proof.Sampled {
		t.Run(key, func(t *testing.T) {
			ref, err := os.ReadFile(filepath.Join(dir, key+".idx"))
			if err != nil || len(ref) != 576 {
				t.Fatal("來源長度", err)
			}
			f, err := os.Open(filepath.Join(dir, key+".png"))
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := png.Decode(f)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
			pixels := image.NewNRGBA(image.Rect(0, 0, 72, 72))
			draw.Draw(pixels, pixels.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
			count := 0
			for _, c := range proof.Cases {
				if c.Key != key {
					continue
				}
				count++
				if len(c.Matches) != 1 || c.Matches[0].Conflict {
					t.Fatal("固定獨立來源預期不同，須回到證據審查")
				}
				source := make([]byte, 576)
				for y := 0; y < 24; y++ {
					for x := 0; x < 24; x++ {
						sx := x
						if c.Flip {
							sx = 23 - x
						}
						source[y*24+x] = ref[y*24+sx]
					}
				}
				scene := bytes.Repeat([]byte{8}, 64000)
				for row := 0; row < 24; row++ {
					copy(scene[(c.Y+row)*320+c.X:], source[row*24:(row+1)*24])
				}
				original := append([]byte(nil), scene...)
				a.effectPaletteOK = true
				a.frame(scene, true)
				if len(a.matches) != 1 {
					t.Fatalf("完整來源候選%d，固定預期1", len(a.matches))
				}
				m, want := a.matches[0], c.Matches[0]
				p := a.patterns[m.pattern]
				sum := sha256.Sum256(p.full)
				if p.key != want.Key || hex.EncodeToString(sum[:]) != want.SHA || m.x != want.X || m.y != want.Y || m.conflict != want.Conflict {
					t.Fatal("Go辨識不同於完整來源獨立預期")
				}
				out := bytes.Repeat([]byte{17}, 960*600*4)
				baseline := append([]byte(nil), out...)
				a.draw(out, scene, 3)
				for y := 0; y < 600; y++ {
					for x := 0; x < 960; x++ {
						o := (y*960 + x) * 4
						expected := baseline[o : o+4]
						if x >= c.X*3 && x < (c.X+24)*3 && y >= c.Y*3 && y < (c.Y+24)*3 {
							px, py := x-c.X*3, y-c.Y*3
							if source[(py/3)*24+px/3] != 8 {
								sx := px
								if c.Flip {
									sx = 71 - px
								}
								s := (py*72 + sx) * 4
								expected = pixels.Pix[s : s+4]
							}
						}
						if !bytes.Equal(out[o:o+4], expected) {
							t.Fatal("完整成品、背景8或框外像素不符")
						}
					}
				}
				if !bytes.Equal(scene, original) {
					t.Fatal("呈現改動原版索引")
				}
				for _, scale := range []int{2, 3} {
					if scale == 3 {
						a.effectPaletteOK = false
					}
					got := append([]byte(nil), baseline...)
					a.draw(got, scene, scale)
					if !bytes.Equal(got, baseline) {
						t.Fatal("2倍或失配色盤仍繪HD")
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
				for n, v := range source {
					if v != 8 {
						scene[(c.Y+n/24)*320+c.X+n%24] ^= 255
						break
					}
				}
				a.frame(scene, true)
				for _, m := range a.matches {
					if m.x == c.X && m.y == c.Y && a.patterns[m.pattern].key == key {
						t.Fatal("缺前景點仍完整辨識")
					}
				}
				a.frame(bytes.Repeat([]byte{8}, 64000), true)
				if len(a.matches) != 0 {
					t.Fatal("清除後殘留")
				}
			}
			if count != 4 {
				t.Fatal("抽樣未涵蓋原向／鏡像及兩種位置")
			}
		})
	}
}
