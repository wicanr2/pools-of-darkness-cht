package pod

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/xlate"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestHDCompositeRealSources(t *testing.T) {
	root := os.Getenv("POD_OCC_ROOT")
	out := os.Getenv("POD_OCC_OUT")
	if root == "" || out == "" {
		t.Skip("需要本機合法素材及正常擷取")
	}
	theme := filepath.Join(root, "workplace/hd/theme")
	capture := filepath.Join(root, "workplace/probe/invincible-v1/gui-audit/sources")
	load := func() *HDLayers {
		h, _, e := LoadHDLayers(theme)
		if e != nil {
			t.Fatal(e)
		}
		return h
	}
	read := func(frame int) ([]byte, [256][3]byte) {
		t.Helper()
		prefix := filepath.Join(capture, fmt.Sprintf("pod-f%06d", frame))
		b, e := os.ReadFile(prefix + ".idx")
		if e != nil {
			t.Fatal(e)
		}
		p, e := os.ReadFile(prefix + ".pal")
		if e != nil {
			t.Fatal(e)
		}
		if len(b) != 64000 || len(p) != 768 {
			t.Fatal("畫格長度")
		}
		var pal [256][3]byte
		for i := range pal {
			copy(pal[i][:], p[i*3:i*3+3])
		}
		return b, pal
	}
	render := func(h *HDLayers, idx []byte, pal [256][3]byte, scale int) []byte {
		p := make([]byte, 320*200*scale*scale*4)
		h.Draw(p, idx, scale)
		if xlate.NewIndexedSource(idx, pal, 320, 200, scale).FillOriginal(p) < 0 {
			t.Fatal("alpha不完整")
		}
		return p
	}
	source := func(idx []byte, pal [256][3]byte, scale int) []byte {
		p := make([]byte, 320*200*scale*scale*4)
		xlate.NewIndexedSource(idx, pal, 320, 200, scale).FillOriginal(p)
		return p
	}
	cellSame := func(a, b []byte, row, col int) bool {
		for y := row * 24; y < row*24+24; y++ {
			n := (y*960 + col*24) * 4
			if !bytes.Equal(a[n:n+96], b[n:n+96]) {
				return false
			}
		}
		return true
	}
	valid := func(ms []compositeSprite) int {
		n := 0
		for _, m := range ms {
			if !m.conflict {
				n++
			}
		}
		return n
	}
	hasTerrain := func(h *HDLayers, x, y int) bool {
		for _, m := range h.composite.terrainNow {
			if m.x == x && m.y == y {
				return true
			}
		}
		return false
	}
	proof := []map[string]any{}
	for _, frame := range []int{8192, 8193, 8196, 8899} {
		idx, pal := read(frame)
		before := sha256.Sum256(idx)
		h := load()
		h.Frame(idx, pal)
		if valid(h.composite.effectsNow) != 1 {
			t.Fatalf("%d效果候選不唯一：%v", frame, h.composite.effectsNow)
		}
		if frame == 8899 {
			if valid(h.composite.actorsNow) != 1 {
				t.Fatal("未辨識特效下方CPIC102")
			}
		} else if !hasTerrain(h, 32, 32) {
			t.Fatalf("%d未接住已知效果下方地形", frame)
		}
		got, base := render(h, idx, pal, 3), source(idx, pal, 3)
		var cells [][2]int
		switch frame {
		case 8192:
			for x := 1; x <= 6; x++ {
				cells = append(cells, [2]int{1, x})
			}
			for x := 2; x <= 6; x++ {
				cells = append(cells, [2]int{2, x})
			}
			cells = append(cells, [2]int{3, 5}, [2]int{3, 6})
		case 8196:
			cells = [][2]int{{4, 4}, {4, 5}, {4, 6}, {5, 5}, {5, 6}}
		case 8899:
			cells = [][2]int{{10, 15}}
		}
		for _, c := range cells {
			if cellSame(got, base, c[0], c[1]) {
				t.Fatalf("%d格%v仍是原版", frame, c)
			}
		}
		if !bytes.Equal(render(h, idx, pal, 2), source(idx, pal, 2)) {
			t.Fatal("二倍誤畫HD")
		}
		// 同一來源的直接與既有RGB合成必須一致，重複Compose不能前進動畫或保留狀態。
		live := NewLive(&xlate.Font{W: 16, H: 16, Glyphs: map[rune][]byte{}}, nil, nil)
		live.HDLayers = h
		live.EnableDirect()
		copy(live.Direct.pending, idx)
		live.Frame(idx, pal)
		direct := live.Compose(3, true, nil)
		legacy := live.composeLegacy(3, true, nil)
		if !bytes.Equal(direct, legacy) || !bytes.Equal(direct, live.Compose(3, true, nil)) || !bytes.Equal(live.Compose(3, false, nil), base) {
			t.Fatal("直接／舊合成或F12／重複呼叫失配")
		}
		if before != sha256.Sum256(idx) {
			t.Fatal("修改原版來源")
		}
		path := filepath.Join(out, fmt.Sprintf("component-%06d.png", frame))
		f, e := os.Create(path)
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, &image.NRGBA{Pix: got, Stride: 960 * 4, Rect: image.Rect(0, 0, 960, 600)})
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		composite := h.composite
		h.composite = nil
		oldPixels := render(h, idx, pal, 3)
		h.composite = composite
		f, e = os.Create(filepath.Join(out, fmt.Sprintf("component-old-%06d.png", frame)))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, &image.NRGBA{Pix: oldPixels, Stride: 960 * 4, Rect: image.Rect(0, 0, 960, 600)})
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		proof = append(proof, map[string]any{"capture_label": frame, "advanced_frame": frame - 1, "source_sha256": fmt.Sprintf("%x", before), "effect_matches": valid(h.composite.effectsNow), "occluded_actors": valid(h.composite.actorsNow), "terrain_matches": len(h.composite.terrainNow), "verified_cells": cells})
	}
	// 以首個真實畫格做來源、色盤、清除、歧義及呼叫間改動的拒絕樣本。
	idx, pal := read(8192)
	h := load()
	h.Frame(idx, pal)
	m := h.composite.effectsNow[0]
	p := h.actors.patterns[m.pattern]
	checkEmpty := func(h *HDLayers) {
		t.Helper()
		if len(h.composite.source) != 0 {
			t.Fatal("錯誤來源仍接受合成")
		}
	}
	bad := append([]byte(nil), idx...)
	for _, pt := range p.ink {
		if pt.color >= 64 {
			bad[(m.y+pt.y)*320+m.x+pt.x] = 127
			break
		}
	}
	h.Frame(bad, pal)
	if valid(h.composite.effectsNow) != 0 {
		t.Fatal("未知特效點仍被接受")
	}
	h = load()
	bad = append([]byte(nil), idx...)
	bad[8*320+32] = 127
	h.Frame(bad, pal)
	if hasTerrain(h, 32, 8) {
		t.Fatal("未知地形點仍被遮罩忽略")
	}
	for _, color := range []int{0, 32} {
		h = load()
		changed := pal
		changed[color][0] ^= 1
		h.Frame(idx, changed)
		checkEmpty(h)
	}
	h = load()
	h.Frame(idx, pal)
	h.Frame(bytes.Repeat([]byte{32}, 64000), pal)
	checkEmpty(h)
	h = load()
	h.Frame(idx, pal)
	h.Frame(nil, pal)
	checkEmpty(h)
	// 完整效果前景歧義須全部拒絕。
	h = load()
	h.actors.patterns = append(h.actors.patterns, h.actors.patterns[m.pattern])
	h.composite = newCombatComposite(h.combat, h.actors)
	h.Frame(idx, pal)
	if valid(h.composite.effectsNow) != 0 {
		t.Fatal("特效歧義未拒絕")
	}
	h = load()
	h.Frame(idx, pal)
	changed := append([]byte(nil), idx...)
	changed[0] ^= 1
	old := h.composite
	h.composite = nil
	base := render(h, changed, pal, 3)
	h.composite = old
	if !bytes.Equal(base, render(h, changed, pal, 3)) {
		t.Fatal("來源改動後仍沿用合成遮罩")
	}
	// 8為不透明背景的原完整矩形路徑也保留原樣，不能把黑底當地形。
	h = load()
	opaque := append([]byte(nil), idx...)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			if p.full[y*24+x] == 8 {
				opaque[(m.y+y)*320+m.x+x] = 8
			}
		}
	}
	h.Frame(opaque, pal)
	got := render(h, opaque, pal, 3)
	original := source(opaque, pal, 3)
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			if p.full[y*24+x] != 8 {
				continue
			}
			for dy := 0; dy < 3; dy++ {
				n := (((m.y+y)*3+dy)*960 + (m.x+x)*3) * 4
				if !bytes.Equal(got[n:n+12], original[n:n+12]) {
					t.Fatal("不透明背景8遭覆畫")
				}
			}
		}
	}
	// CPIC色盤失配只能停用該層，已證實效果仍可獨立呈現。
	ci, cp := read(8899)
	h = load()
	wrong := cp
	wrong[144][0] ^= 1
	h.Frame(ci, wrong)
	if valid(h.composite.actorsNow) != 0 || valid(h.composite.effectsNow) != 1 {
		t.Fatal("CPIC／效果色盤未隔離")
	}
	h = load()
	wrong = cp
	wrong[64][0] ^= 1
	h.Frame(ci, wrong)
	if valid(h.composite.effectsNow) != 0 || len(h.composite.terrainNow) == 0 {
		t.Fatal("效果色盤失配不應停用已確認CPIC下的地形")
	}
	// 被特效藏住的原始像素不同、可見部分相同，也不能猜選其中一張HD。
	h = load()
	h.Frame(ci, cp)
	under := h.composite.actorsNow[0]
	duplicate := h.actors.patterns[under.pattern]
	duplicate.ink = append([]actorInk(nil), duplicate.ink...)
	for n, pt := range duplicate.ink {
		if h.composite.effectMask[(under.y+pt.y)*320+under.x+pt.x] {
			duplicate.ink[n].color ^= 1
			break
		}
	}
	h.actors.patterns = append(h.actors.patterns, duplicate)
	h.composite = newCombatComposite(h.combat, h.actors)
	h.classified = false
	h.Frame(ci, cp)
	if valid(h.composite.actorsNow) != 0 {
		t.Fatal("被遮擋CPIC歧義未拒絕")
	}
	h = load()
	h.Frame(idx, pal)
	var bg compositeTerrain
	for _, candidate := range h.composite.terrainNow {
		if candidate.x == 8 && candidate.y == 8 {
			bg = candidate
		}
	}
	copyTile := h.combat.tiles[bg.tile]
	copyTile.reference = append([]byte(nil), copyTile.reference...)
	for n := range copyTile.reference {
		if h.composite.cover[(bg.y+n/24)*320+bg.x+n%24] {
			copyTile.reference[n] ^= 1
			break
		}
	}
	n := len(h.combat.tiles)
	h.combat.tiles = append(h.combat.tiles, copyTile)
	h.combat.vgaLookup[vgaTerrainSignature{sha256.Sum256(copyTile.reference), paletteBankHash(pal, 32, 64)}] = n
	h.composite = newCombatComposite(h.combat, h.actors)
	h.classified = false
	h.Frame(idx, pal)
	if hasTerrain(h, 8, 8) {
		t.Fatal("地形合成來源歧義未拒絕")
	}
	data, _ := json.MarshalIndent(map[string]any{"samples": proof, "negative_classes": 11, "source_unchanged": true, "direct_legacy_F12_2x_repeat": true}, "", "  ")
	if e := os.WriteFile(filepath.Join(out, "component.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
}

// 完整CPIC下的地形不能等不相關特效出現，否則原本不相干的位置也會閃現。
func TestHDCompositeForegroundIndependent(t *testing.T) {
	root, out := os.Getenv("POD_OCC_ROOT"), os.Getenv("POD_OCC_OUT")
	if root == "" || out == "" {
		t.Skip("需要正常DEMO來源")
	}
	count := 0
	for _, event := range []struct {
		frame, row int
		cols       []int
	}{
		{6592, 16, []int{20, 21}}, {7955, 10, []int{8, 9}},
		{8210, 7, []int{6, 7, 8}}, {8750, 8, []int{9, 10}},
	} {
		// 這裡只檢查當幀的HD或暫留資格。8210擷取已在分段繪製途中，
		// 三張截圖沒有更早的暫留歷史，不能以全新Live宣稱正常零閃現。
		// 四格上限由TestBattleHold覆蓋，連續歷史另由完整10500格GUI驗證。
		for _, label := range []int{event.frame + 1, event.frame + 2} {
			prefix := filepath.Join(out, "gui-diagnose/sources", fmt.Sprintf("pod-f%06d", label))
			idx, e := os.ReadFile(prefix + ".idx")
			if e != nil {
				t.Fatal(e)
			}
			raw, e := os.ReadFile(prefix + ".pal")
			if e != nil {
				t.Fatal(e)
			}
			var pal [256][3]byte
			for i := range pal {
				copy(pal[i][:], raw[i*3:i*3+3])
			}
			h, _, e := LoadHDLayers(filepath.Join(root, "workplace/hd/theme"))
			if e != nil {
				t.Fatal(e)
			}
			h.Frame(idx, pal)
			got, base := make([]byte, 960*600*4), make([]byte, 960*600*4)
			h.Draw(got, idx, 3)
			source := xlate.NewIndexedSource(idx, pal, 320, 200, 3)
			source.FillOriginal(got)
			source.FillOriginal(base)
			for _, col := range event.cols {
				same := true
				for y := event.row * 24; y < (event.row+1)*24; y++ {
					n := (y*960 + col*24) * 4
					same = same && bytes.Equal(got[n:n+96], base[n:n+96])
				}
				pending := h.pendingBattleCells()
				cell := (event.row-1)*21 + col - 1
				if same && !pending[cell] {
					t.Fatalf("擷取%d列%d欄%d仍原樣呈現", label, event.row, col)
				}
			}
			count++
		}
	}
	t.Logf("四處前後%d個來源獨立合成通過", count)
}

func TestHDCompositeSmallPrefixGuards(t *testing.T) {
	root, out := os.Getenv("POD_OCC_ROOT"), os.Getenv("POD_OCC_OUT")
	if root == "" || out == "" {
		t.Skip("需要正常DEMO來源")
	}
	prefix := filepath.Join(out, "gui-diagnose/sources/pod-f008751")
	idx, e := os.ReadFile(prefix + ".idx")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(prefix + ".pal")
	if e != nil {
		t.Fatal(e)
	}
	var pal [256][3]byte
	for i := range pal {
		copy(pal[i][:], raw[i*3:i*3+3])
	}
	load := func() *HDLayers {
		h, _, e := LoadHDLayers(filepath.Join(root, "workplace/hd/theme"))
		if e != nil {
			t.Fatal(e)
		}
		return h
	}
	cell := (8-1)*21 + 9 - 1
	h := load()
	h.Frame(idx, pal)
	if !h.composite.deferCells[cell] {
		t.Fatal("完整小型特效前綴未認出")
	}
	bad := append([]byte(nil), idx...)
	bad[64*320+64] = 127
	h.Frame(bad, pal)
	if h.composite.deferCells[cell] {
		t.Fatal("未知背景點未拒絕")
	}
	for _, color := range []int{0, 32, 64} {
		h = load()
		wrong := pal
		wrong[color][0] ^= 1
		h.Frame(idx, wrong)
		if h.composite.deferCells[cell] {
			t.Fatal("前綴色盤失配未拒絕")
		}
	}
	h = load()
	h.Frame(idx, pal)
	for _, candidate := range h.composite.prefixes {
		if candidate.rows == 8 && candidate.source[0] == 32 {
			p := h.actors.patterns[candidate.pattern]
			if p.key == "comspr-005" {
				h.actors.patterns = append(h.actors.patterns, p)
				break
			}
		}
	}
	h.composite = newCombatComposite(h.combat, h.actors)
	h.classified = false
	h.Frame(idx, pal)
	if h.composite.deferCells[cell] {
		t.Fatal("前綴歧義未拒絕")
	}
	h = load()
	h.Frame(idx, pal)
	h.Frame(bytes.Repeat([]byte{32}, 64000), pal)
	if h.composite.deferCells[cell] {
		t.Fatal("清除後前綴殘留")
	}
}

func TestHDCompositeHoldRemainsFourFrames(t *testing.T) {
	var hold battleHold
	var eligible [441]bool
	eligible[0] = true
	idx := make([]byte, 64000)
	var pal [256][3]byte
	pal[64] = [3]byte{4, 8, 12}
	base := make([]byte, 960*600*4)
	xlate.NewIndexedSource(idx, pal, 320, 200, 3).FillOriginal(base)
	previous := append([]byte(nil), base...)
	for y := 24; y < 48; y++ {
		for x := 24; x < 48; x++ {
			at := (y*960 + x) * 4
			copy(previous[at:at+4], []byte{31, 32, 33, 255})
		}
	}
	hold.apply(idx, base, previous, 1, eligible)
	idx[8*320+8] = 64
	xlate.NewIndexedSource(idx, pal, 320, 200, 3).FillOriginal(base)
	for frame := 1; frame <= 5; frame++ {
		hold.commit()
		got := append([]byte(nil), base...)
		hold.apply(idx, base, got, 1, eligible)
		at := (24*960 + 24) * 4
		if frame <= 4 && got[at] != 31 {
			t.Fatalf("第%d格未暫留", frame)
		}
		if frame == 5 && !bytes.Equal(got, base) {
			t.Fatal("第五格未強制回退")
		}
	}
}
