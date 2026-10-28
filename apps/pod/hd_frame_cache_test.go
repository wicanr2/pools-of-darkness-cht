package pod

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func TestHDFrameCacheOriginalClassification(t *testing.T) {
	theme, capture := os.Getenv("DOSGOLEM_POD_HD_DIR"), os.Getenv("DOSGOLEM_POD_HD_FRAME_CAPTURE")
	if theme == "" || capture == "" {
		t.Skip("未提供本機HD主題與正常畫格")
	}
	load := func() *HDLayers {
		h, _, err := LoadHDLayers(theme)
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(capture, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	indexed, raw := read("source.idx"), read("source.pal")
	if len(indexed) != 64000 || len(raw) != 768 {
		t.Fatal("畫格長度不合")
	}
	var pal [256][3]uint8
	for i := range pal {
		copy(pal[i][:], raw[i*3:i*3+3])
	}
	original, cached := load(), load()
	cases := 0
	compare := func(idx []byte, palette [256][3]uint8) {
		original.classify(idx, palette)
		cached.Frame(idx, palette)
		a, b := make([]byte, 960*600*4), make([]byte, 960*600*4)
		original.Draw(a, idx, 3)
		cached.Draw(b, idx, 3)
		if !bytes.Equal(a, b) || original.pendingBattleCells() != cached.pendingBattleCells() {
			t.Fatalf("快取改變HD或待呈現格，第%d組", cases)
		}
		cases++
	}
	before := sha256.Sum256(indexed)
	for i := 0; i < 12; i++ {
		compare(indexed, pal)
	}
	if len(original.actors.matches) == 0 {
		t.Fatal("正常夾具沒有已辨識人物")
	}
	m := original.actors.matches[0]
	pixel := original.actors.patterns[m.pattern].ink[0]
	at := (m.y+pixel.y)*320 + m.x + pixel.x
	changed := append([]byte(nil), indexed...)
	changed[at] ^= 1
	for i := 0; i < 3; i++ {
		compare(changed, pal)
	}
	changed[at] ^= 1
	compare(changed, pal)
	alternate := pal
	alternate[144][0] ^= 1
	compare(changed, alternate)
	compare(changed, alternate)
	compare(changed, pal)
	compare(nil, pal)
	compare(nil, pal)
	compare(indexed, pal)
	if sha256.Sum256(indexed) != before {
		t.Fatal("修改原始畫格")
	}
	t.Log("HD cached/uncached original output comparisons", cases)
}
