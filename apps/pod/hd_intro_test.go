package pod

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIntroRealSources(t *testing.T) {
	dir, capture, proof := os.Getenv("POD_INTRO_THEME"), os.Getenv("POD_INTRO_CAPTURE"), os.Getenv("POD_INTRO_PROOF")
	if dir == "" || capture == "" || proof == "" {
		t.Skip("需要自備原版及研究031正常來源")
	}
	h, err := loadIntroHD(dir)
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		Events []struct {
			Source       string `json:"source"`
			Kind         string `json:"kind"`
			Block, Frame int
			ImageCall    int `json:"image_call"`
		}
	}
	b, err := os.ReadFile(proof)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, e := range p.Events {
		if e.ImageCall == 0 || e.Block == 0 || e.Kind == "outside-BMA" {
			continue
		}
		f := h.byCall[[2]int{e.Block, e.Frame}]
		if f == nil {
			t.Fatal(e)
		}
		idx, err := os.ReadFile(filepath.Join(capture, e.Source+".idx"))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(capture, e.Source+".pal"))
		if err != nil || len(raw) != 768 {
			t.Fatal("palette", err)
		}
		var pal [256][3]uint8
		for n := range pal {
			copy(pal[n][:], raw[n*3:n*3+3])
		}
		saved := append([]byte(nil), idx...)
		h.active = f
		h.generation++
		h.frame(idx, pal)
		if len(h.rgba) == 0 {
			t.Fatalf("來源未覆蓋：%s %s b%d f%d", e.Source, e.Kind, e.Block, e.Frame)
		}
		if !bytes.Equal(saved, idx) {
			t.Fatal("改寫來源")
		}
		counts[e.Kind]++
		out := bytes.Repeat([]byte{31}, 960*600*4)
		h.draw(out, 2)
		if !bytes.Equal(out, bytes.Repeat([]byte{31}, len(out))) {
			t.Fatal("二倍出圖")
		}
		h.draw(out, 3)
		if !bytes.Equal(out[960*504*4:], bytes.Repeat([]byte{31}, 960*96*4)) {
			t.Fatal("覆蓋字幕")
		}
		if e.Kind == "complete" {
			for _, palette := range h.meta.Blocks[e.Block].Palettes {
				flat := make([]byte, 0, 768)
				for _, c := range pal {
					flat = append(flat, c[:]...)
				}
				if palette.SHA == fmt.Sprintf("%x", sha256.Sum256(flat)) && palette.Light == 65535 && !bytes.Equal(h.rgba, f.pixels) {
					t.Fatal("完整HD像素不同", e.Source)
				}
			}
		}
		broken := append([]byte(nil), idx...)
		broken[12345] ^= 127
		h.frame(broken, pal)
		if len(h.rgba) != 0 {
			t.Fatal("未知像素出圖", e.Source)
		}
		pal[255][0] ^= 1
		h.frame(idx, pal)
		if len(h.rgba) != 0 {
			t.Fatal("未知色盤出圖", e.Source)
		}
	}
	if counts["complete"] < 100 || counts["draw-prefix"] < 20 || counts["clear-prefix"] < 1 {
		t.Fatal("抽樣不齊", counts)
	}
	t.Log(counts)
	h.reset()
	if h.active != nil || h.rgba != nil {
		t.Fatal("重設殘留")
	}
}

func TestIntroPrefixRejectsOutOfOrder(t *testing.T) {
	base := make([]byte, introPixels)
	target := append([]byte(nil), base...)
	target[100] = 1
	target[900] = 2
	target[1200] = 3
	ops := []uint16{100, 900, 1200}
	idx := append([]byte(nil), base...)
	idx[900] = 2
	if _, ok := introPrefix(idx, base, target, ops); ok {
		t.Fatal("接受亂序寫入")
	}
	idx[100] = 1
	if cut, ok := introPrefix(idx, base, target, ops); !ok || cut != 2 {
		t.Fatal("拒絕真前綴")
	}
	idx[400] = 1
	if _, ok := introPrefix(idx, base, target, ops); ok {
		t.Fatal("接受範圍外改寫")
	}
}

func TestIntroManifestGuards(t *testing.T) {
	dir := os.Getenv("POD_INTRO_THEME")
	if dir == "" {
		t.Skip("需要本機BMA重建主題")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "intro.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*introMeta)
	}{
		{"version", func(m *introMeta) { m.Version = 2 }},
		{"count", func(m *introMeta) { m.Frames = m.Frames[:132] }},
		{"duplicate", func(m *introMeta) { m.Frames[1].Block = 0; m.Frames[1].Frame = 0 }},
		{"traversal", func(m *introMeta) { m.Frames[1].PNG.Path = "../escape.png" }},
		{"index-hash", func(m *introMeta) { m.Frames[0].Index.SHA = string(bytes.Repeat([]byte{'0'}, 64)) }},
		{"palette", func(m *introMeta) { m.Blocks[1].Palettes[0].Light = 65536 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m introMeta
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.change(&m)
			tmp := t.TempDir()
			if err := os.Symlink(filepath.Join(dir, "intro"), filepath.Join(tmp, "intro")); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.Marshal(m)
			if err := os.WriteFile(filepath.Join(tmp, "intro.json"), encoded, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadIntroHD(tmp); err == nil {
				t.Fatal("未拒絕失配清單")
			}
		})
	}
}

// 診斷觀測器：模擬已證實的來源呼叫參數；正常入口由獨立GUI收據覆蓋。
type introTestCPU struct {
	mem            []byte
	step           uint64
	cs, ip, ss, sp uint16
}

func (v *introTestCPU) Read8(a uint32) uint8   { return v.mem[a] }
func (v *introTestCPU) Read16(a uint32) uint16 { return binary.LittleEndian.Uint16(v.mem[a:]) }
func (v *introTestCPU) Steps() uint64          { return v.step }
func (v *introTestCPU) CS() uint16             { return v.cs }
func (v *introTestCPU) IP() uint16             { return v.ip }
func (v *introTestCPU) SS() uint16             { return v.ss }
func (v *introTestCPU) SP() uint16             { return v.sp }
func (v *introTestCPU) DS() uint16             { return 0 }

func TestIntroSourceCallGuards(t *testing.T) {
	dir, orig := os.Getenv("POD_INTRO_THEME"), os.Getenv("POD_INTRO_ORIGINAL")
	if dir == "" || orig == "" {
		t.Skip("需要自備原版")
	}
	h, err := loadIntroHD(dir)
	if err != nil {
		t.Fatal(err)
	}
	bma, err := os.ReadFile(filepath.Join(orig, "INTRO.BMA"))
	if err != nil {
		t.Fatal(err)
	}
	drv, err := os.ReadFile(filepath.Join(orig, "VGA.DRV"))
	if err != nil {
		t.Fatal(err)
	}
	v := &introTestCPU{mem: make([]byte, 1<<20), step: 1, cs: 0x48a1, ip: 0x6733, ss: 0x100, sp: 0x100}
	copy(v.mem[0x4f143:], drv[0x6733:0x6733+255])
	off := int(binary.LittleEndian.Uint32(bma))
	size := int(binary.LittleEndian.Uint32(bma[off:]))
	copy(v.mem[0x20000:], bma[off:off+size])
	put := func(off uint32, x uint16) { binary.LittleEndian.PutUint16(v.mem[0x1100+off:], x) }
	put(4, 2)
	put(14, 0x2000)
	h.observe(v)
	if h.active == nil || h.active.block != 1 || h.active.frame != 0 {
		t.Fatal("來源未綁定")
	}
	h.observe(v)
	if h.stats.Calls != 1 {
		t.Fatal("相同步未去重")
	}
	v.step++
	put(6, 1)
	h.observe(v)
	if h.active != nil {
		t.Fatal("未知座標綁定")
	}
	v.step++
	put(6, 0)
	v.mem[0x20010] ^= 1
	h.observe(v)
	if h.active != nil {
		t.Fatal("來源失配綁定")
	}
	v.step++
	v.mem[0x20010] ^= 1
	v.mem[0x4f143] ^= 1
	h.observe(v)
	if h.active != nil {
		t.Fatal("程式碼失配綁定")
	}
}
