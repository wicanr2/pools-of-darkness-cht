package pod

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/xlate"
)

func vgaFixture(t *testing.T) (*HDLayers, []byte, [256][3]uint8, string) {
	t.Helper()
	dir, capture := os.Getenv("DOSGOLEM_POD_HD_VGA_DIR"), os.Getenv("DOSGOLEM_POD_HD_VGA_CAPTURE")
	if dir == "" || capture == "" {
		t.Skip("未提供本機VGA主題與正常DEMO畫格")
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(capture, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	idx, rawpal := read("battle.idx"), read("battle.pal")
	if len(idx) != 64000 || len(rawpal) != 768 {
		t.Fatal("來源畫格長度不合")
	}
	var pal [256][3]uint8
	for i := range pal {
		copy(pal[i][:], rawpal[i*3:i*3+3])
	}
	h, count, err := LoadHDLayers(dir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 8 {
		t.Fatalf("首批圖層數%d，預期8", count)
	}
	return h, idx, pal, dir
}

func fixtureLive(t *testing.T, dir string) *Live {
	t.Helper()
	l := NewLive(nil, nil, nil)
	if err := l.AddHD(dir); err != nil {
		t.Fatal(err)
	}
	return l
}

func TestVGABattleTransitionFixture(t *testing.T) {
	_, _, _, dir := vgaFixture(t)
	capture := os.Getenv("DOSGOLEM_POD_HD_VGA_TRANSITION")
	if capture == "" {
		t.Skip("未提供正常DEMO逐格來源")
	}
	for _, keep := range []bool{false, true} {
		flashes := 0
		for _, span := range [][2]int{{6466, 6475}, {6688, 6695}} {
			live := fixtureLive(t, dir)
			if !keep {
				live.battleHold = nil
			}
			live.Flash = xlate.NewFlashAudit(320, 200, 8)
			for frame := span[0]; frame <= span[1]; frame++ {
				prefix := filepath.Join(capture, fmt.Sprintf("pod-f%06d", frame))
				idx, err := os.ReadFile(prefix + ".idx")
				if err != nil {
					t.Fatal(err)
				}
				rawpal, err := os.ReadFile(prefix + ".pal")
				if err != nil || len(rawpal) != 768 {
					t.Fatal("逐格色盤不合", err)
				}
				var pal [256][3]uint8
				for i := range pal {
					copy(pal[i][:], rawpal[i*3:i*3+3])
				}
				original := append([]byte(nil), idx...)
				live.Frame(idx, pal)
				live.AuditFlash(frame-1, 3)
				a := live.Compose(3, true, nil)
				b := live.Compose(3, true, nil)
				if !bytes.Equal(a, b) {
					t.Fatal("同格多次合成不一致")
				}
				if !bytes.Equal(original, idx) {
					t.Fatal("保留修改原始色號")
				}
			}
			flashes += live.Flash.Count()
			if keep && live.Flash.Count() > 0 {
				t.Logf("未消除事件：%+v", live.Flash.Events())
			}
		}
		t.Logf("保留%v：閃現%d", keep, flashes)
		if keep && flashes != 0 || !keep && flashes == 0 {
			t.Fatalf("保留%v：閃現%d", keep, flashes)
		}
	}
}

func TestVGABattleHoldUnknownBoundAndPalette(t *testing.T) {
	_, idx, pal, dir := vgaFixture(t)
	live := fixtureLive(t, dir)
	live.Frame(idx, pal)
	live.Compose(3, true, nil)
	offset := 100*320 + 100 // 沒有來源的場中央，不能等候猜測中的HD。
	pixel := (300*960 + 300) * 4
	for frame := 0; frame < 4; frame++ {
		changed := append([]byte(nil), idx...)
		changed[offset] = byte(9 + frame)
		live.Frame(changed, pal)
		got := live.Compose(3, true, nil)
		base := vgaBaseRGBA(changed, pal, 3)
		if !bytes.Equal(got[pixel:pixel+4], base[pixel:pixel+4]) {
			t.Fatal("未知圖形未立即回退")
		}
		if !bytes.Equal(live.Compose(3, false, nil), base) {
			t.Fatal("F12仍保留")
		}
		if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(changed, pal, 2)) {
			t.Fatal("2倍仍保留")
		}
	}
	changedPalette := pal
	changedPalette[200][0] ^= 1
	live.Frame(idx, changedPalette)
	want := vgaBaseRGBA(idx, changedPalette, 3)
	live.HDLayers.Draw(want, idx, 3)
	if !bytes.Equal(live.Compose(3, true, nil), want) {
		t.Fatal("改色後保留舊HD")
	}
	blank := bytes.Repeat([]byte{8}, 64000)
	live.Frame(blank, changedPalette)
	if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(blank, changedPalette, 3)) {
		t.Fatal("非戰場仍保留")
	}
}

func TestVGABattleHoldBound(t *testing.T) {
	_, idx, pal, _ := vgaFixture(t)
	h := &battleHold{}
	var eligible [21 * 21]bool
	for n := range eligible {
		eligible[n] = true
	}
	previous := vgaBaseRGBA(idx, pal, 3)
	h.apply(idx, previous, previous, 1, eligible)
	offset, pixel := 100*320+100, (300*960+300)*4
	for frame := 0; frame < 8; frame++ {
		h.commit()
		changed := append([]byte(nil), idx...)
		changed[offset] = byte(9 + frame)
		base := vgaBaseRGBA(changed, pal, 3)
		got := append([]byte(nil), base...)
		h.apply(changed, base, got, 1, eligible)
		if frame%4 < 3 {
			if !bytes.Equal(got[pixel:pixel+4], previous[pixel:pixel+4]) {
				t.Fatal("未按已證實的三格間隔保留")
			}
		} else if !bytes.Equal(got[pixel:pixel+4], base[pixel:pixel+4]) {
			t.Fatal("連續變化超過三格未回退")
		}
		previous = got
	}
}

func vgaBaseRGBA(idx []byte, pal [256][3]uint8, scale int) []byte {
	out := make([]byte, 320*200*scale*scale*4)
	for y := 0; y < 200*scale; y++ {
		for x := 0; x < 320*scale; x++ {
			c := pal[idx[y/scale*320+x/scale]]
			o := (y*320*scale + x) * 4
			copy(out[o:o+3], c[:])
			out[o+3] = 255
		}
	}
	return out
}

func TestVGABattleFixtureForegroundAndBackground(t *testing.T) {
	h, idx, pal, dir := vgaFixture(t)
	original := append([]byte(nil), idx...)
	originalPal := pal
	h.Frame(idx, pal)
	if len(h.combat.matches) != 5 || len(h.actors.matches) != 8 {
		t.Fatalf("地形%d、怪物%d，預期5／8", len(h.combat.matches), len(h.actors.matches))
	}
	base := vgaBaseRGBA(idx, pal, 3)
	terrain := append([]byte(nil), base...)
	h.combat.draw(terrain, idx, 3)
	out := append([]byte(nil), base...)
	h.Draw(out, idx, 3)
	if bytes.Equal(out, base) {
		t.Fatal("HD沒有實際繪製")
	}
	allowed := make([]bool, 64000)
	for _, m := range h.combat.matches {
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				allowed[(m.y+y)*320+m.x+x] = true
			}
		}
	}
	for _, m := range h.actors.matches {
		p := h.actors.patterns[m.pattern]
		foreground := make([]bool, p.w*p.h)
		for _, point := range p.ink {
			foreground[point.y*p.w+point.x] = true
			allowed[(m.y+point.y)*320+m.x+point.x] = true
		}
		for y := 0; y < p.h; y++ {
			for x := 0; x < p.w; x++ {
				if !foreground[y*p.w+x] {
					for sy := 0; sy < 3; sy++ {
						o := (((m.y+y)*3+sy)*960 + (m.x+x)*3) * 4
						if !bytes.Equal(out[o:o+12], terrain[o:o+12]) {
							t.Fatalf("%s背景8未保留底圖", p.key)
						}
					}
				}
			}
		}
	}
	for y := 0; y < 600; y++ {
		for x := 0; x < 960; x++ {
			o := (y*960 + x) * 4
			if !allowed[y/3*320+x/3] && !bytes.Equal(base[o:o+4], out[o:o+4]) {
				t.Fatal("HD超出原始前景與完整地形")
			}
		}
	}
	// 已觀測DRIDER的整個前景必須等於成品的水平鏡像，連同每個3×3格內的細節一起翻轉。
	f, err := os.Open(filepath.Join(dir, "cpic1-041.png"))
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	ref, err := os.ReadFile(filepath.Join(dir, "cpic1-041.idx"))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < 24; y++ {
		for x := 0; x < 48; x++ {
			if ref[y*48+x] == 8 {
				continue
			}
			for sy := 0; sy < 3; sy++ {
				for sx := 0; sx < 3; sx++ {
					r, g, b, a := im.At(x*3+2-sx, y*3+sy).RGBA()
					want := []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)}
					o := (((56+y)*3+sy)*960 + (8+47-x)*3 + sx) * 4
					if !bytes.Equal(out[o:o+4], want) {
						t.Fatal("鏡像HD細節不符原版朝向")
					}
				}
			}
		}
	}
	if !bytes.Equal(idx, original) || pal != originalPal {
		t.Fatal("呈現修改來源色號或色盤")
	}
	for _, scale := range []int{2} {
		a := vgaBaseRGBA(idx, pal, scale)
		b := append([]byte(nil), a...)
		h.Draw(b, idx, scale)
		if !bytes.Equal(a, b) {
			t.Fatal("2倍畫了HD")
		}
	}
}

func TestVGABattleFixtureFallbacks(t *testing.T) {
	h, idx, pal, _ := vgaFixture(t)
	bad := append([]byte(nil), idx...)
	// 保留怪物前景但移除所有具資訊地形，不能把相似圖形當作戰鬥場景。
	for y := 8; y < 176; y++ {
		for x := 8; x < 176; x++ {
			o := y*320 + x
			if bad[o] >= 32 && bad[o] < 64 {
				bad[o] = 32
			}
		}
	}
	h.Frame(bad, pal)
	if len(h.combat.matches) != 0 || len(h.actors.matches) != 0 {
		t.Fatal("缺少地形守衛仍繪怪物")
	}
	for _, color := range []int{0, 200} {
		changed := pal
		changed[color][0] ^= 1
		h.Frame(idx, changed)
		if len(h.actors.matches) != 0 {
			t.Fatalf("色盤%d失配仍繪怪物", color)
		}
		if color == 0 && len(h.combat.matches) != 0 {
			t.Fatal("保留色失配仍繪地形")
		}
	}
	changedTerrain := pal
	changedTerrain[32][0] ^= 1
	h.Frame(idx, changedTerrain)
	if len(h.combat.matches) != 0 || len(h.actors.matches) != 0 {
		t.Fatal("地形色盤失配仍通過戰鬥守衛")
	}
	h.Frame(idx, pal)
	m := h.actors.matches[0]
	p := h.actors.patterns[m.pattern]
	point := p.ink[0]
	bad = append([]byte(nil), idx...)
	bad[(m.y+point.y)*320+m.x+point.x] ^= 255
	h.Frame(bad, pal)
	for _, match := range h.actors.matches {
		if match.x == m.x && match.y == m.y {
			t.Fatal("來源前景失配仍匹配")
		}
	}
	// 高色號改色必須同時讓正式Live的鎖存失效，不能只讓當幀比對失效。
	live := NewLive(nil, nil, nil)
	live.HDLayers = h
	live.hdLatch = xlate.NewLatch(320, 200, hdLatchKeep, hdLatchFrames)
	live.Frame(idx, pal)
	before := live.Compose(3, true, nil)
	if bytes.Equal(before, vgaBaseRGBA(idx, pal, 3)) {
		t.Fatal("正對照沒有HD")
	}
	changed := pal
	changed[200][0] ^= 1
	live.Frame(idx, changed)
	want := vgaBaseRGBA(idx, changed, 3)
	h.Draw(want, idx, 3)
	if got := live.Compose(3, true, nil); !bytes.Equal(got, want) {
		t.Fatal("高色號變更後鎖存仍沿用舊HD")
	}
}
