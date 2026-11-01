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

func TestFinalRealSources(t *testing.T) {
	dir, capture, proof := os.Getenv("POD_FINAL_THEME"), os.Getenv("POD_FINAL_CAPTURE"), os.Getenv("POD_FINAL_PROOF")
	if dir == "" || capture == "" || proof == "" {
		t.Skip("requires original FINAL player evidence")
	}
	h, err := loadFinalHD(dir)
	if err != nil || h == nil {
		t.Fatal(err)
	}
	var p struct {
		Events []struct {
			Source, Kind string
			Block, Frame int
			ImageCall    int `json:"image_call"`
		}
	}
	raw, err := os.ReadFile(proof)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, e := range p.Events {
		if e.Kind == "outside-BMA" {
			continue
		}
		idx, err := os.ReadFile(filepath.Join(capture, e.Source+".idx"))
		if err != nil {
			t.Fatal(err)
		}
		flat, err := os.ReadFile(filepath.Join(capture, e.Source+".pal"))
		if err != nil || len(flat) != 768 {
			t.Fatal(err)
		}
		var pal [256][3]uint8
		for n := range pal {
			copy(pal[n][:], flat[n*3:n*3+3])
		}
		before := append([]byte(nil), idx...)
		h.active = h.byCall[[2]int{e.Block, e.Frame}]
		h.generation++
		h.frame(idx, pal)
		if len(h.rgba) == 0 {
			t.Fatalf("missing %s b%d f%d %s", e.Source, e.Block, e.Frame, e.Kind)
		}
		if !bytes.Equal(before, idx) {
			t.Fatal("source changed")
		}
		counts[e.Kind]++
		if e.Kind == "full" {
			for _, p := range h.meta.Blocks[e.Block].Palettes {
				if p.SHA == fmt.Sprintf("%x", sha256.Sum256(flat)) && p.Light == 65535 && !bytes.Equal(h.rgba, h.active.pixels) {
					t.Fatal("full PNG differs", e.Source)
				}
			}
		}
		out := bytes.Repeat([]byte{31}, 960*600*4)
		h.draw(out, 2)
		if !bytes.Equal(out, bytes.Repeat([]byte{31}, len(out))) {
			t.Fatal("2x changed")
		}
		h.draw(out, 3)
		if !bytes.Equal(out[960*504*4:], bytes.Repeat([]byte{31}, 960*96*4)) {
			t.Fatal("bottom32 changed")
		}
		bad := append([]byte(nil), idx...)
		bad[10] = 255
		h.frame(bad, pal)
		if len(h.rgba) != 0 {
			t.Fatal("unknown pixel accepted")
		}
		pal[255][0] ^= 0x55
		h.frame(idx, pal)
		if len(h.rgba) != 0 {
			t.Fatal("unknown palette accepted")
		}
	}
	if counts["full"] != 53 || counts["write-prefix"] != 9 || counts["zero-write-prefix"] != 3 {
		t.Fatal(counts)
	}
	h.reset()
	if h.active != nil || h.rgba != nil {
		t.Fatal("reset stale")
	}
	t.Log(counts)
}

func TestFinalSourceCallGuards(t *testing.T) {
	dir, orig := os.Getenv("POD_FINAL_THEME"), os.Getenv("POD_FINAL_ORIGINAL")
	if dir == "" || orig == "" {
		t.Skip("requires original files")
	}
	h, err := loadFinalHD(dir)
	if err != nil {
		t.Fatal(err)
	}
	bma, err := os.ReadFile(filepath.Join(orig, "FINAL.BMA"))
	if err != nil {
		t.Fatal(err)
	}
	drv, err := os.ReadFile(filepath.Join(orig, "VGA.DRV"))
	if err != nil {
		t.Fatal(err)
	}
	size := binary.LittleEndian.Uint32(bma)
	for _, cs := range []uint16{0x46eb, 0x48a1} {
		v := &introTestCPU{mem: make([]byte, 1<<20), step: 1, cs: cs, ip: 0x6733, ss: 0x100, sp: 0x100}
		code := linear(cs, 0x6733)
		copy(v.mem[code:], drv[0x6733:0x6733+255])
		copy(v.mem[0x20000:], bma[:size])
		put := func(off uint32, x uint16) { binary.LittleEndian.PutUint16(v.mem[0x1100+off:], x) }
		put(4, 2)
		put(14, 0x2000)
		h.reset()
		saved := sha256.Sum256(v.mem)
		calls := h.stats.Calls
		h.observe(v)
		if h.active == nil || h.active.block != 0 || h.active.frame != 0 {
			t.Fatal("source not bound", cs)
		}
		if sha256.Sum256(v.mem) != saved {
			t.Fatal("guest memory changed")
		}
		h.observe(v)
		if h.stats.Calls != calls+1 {
			t.Fatal("step not deduplicated")
		}
		for _, off := range []uint32{4, 6, 8, 10, 16} {
			v.step++
			old := v.Read16(0x1100 + off)
			put(off, 99)
			h.observe(v)
			if h.active != nil {
				t.Fatal("bad arg accepted", off)
			}
			put(off, old)
		}
		v.step++
		v.mem[0x20010] ^= 1
		h.observe(v)
		if h.active != nil {
			t.Fatal("bad block accepted")
		}
		v.mem[0x20010] ^= 1
		v.step++
		v.mem[code] ^= 1
		h.observe(v)
		if h.active != nil {
			t.Fatal("bad code accepted")
		}
		v.step++
		v.cs = 0xf000
		h.observe(v)
		if h.active != nil {
			t.Fatal("out-of-RAM code accepted")
		}
	}
}

func TestFinalManifestGuards(t *testing.T) {
	dir := os.Getenv("POD_FINAL_THEME")
	if dir == "" {
		t.Skip("requires theme")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "final.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*introMeta){
		func(m *introMeta) { m.Version = 2 }, func(m *introMeta) { m.Frames = m.Frames[:57] },
		func(m *introMeta) { m.Blocks = m.Blocks[:5] }, func(m *introMeta) { m.Frames[1].Frame = 99 },
		func(m *introMeta) { m.Frames[0].PNG.Path = "../escape.png" },
		func(m *introMeta) { m.Frames[0].Index.SHA = string(bytes.Repeat([]byte{'0'}, 64)) },
		func(m *introMeta) { m.Blocks[0].Palettes[0].Light = 65536 },
	} {
		var m introMeta
		json.Unmarshal(raw, &m)
		change(&m)
		tmp := t.TempDir()
		if err = os.Symlink(filepath.Join(dir, "final"), filepath.Join(tmp, "final")); err != nil {
			t.Fatal(err)
		}
		b, _ := json.Marshal(m)
		os.WriteFile(filepath.Join(tmp, "final.json"), b, 0600)
		if _, err = loadFinalHD(tmp); err == nil {
			t.Fatal("bad manifest accepted")
		}
	}
	tmp := t.TempDir()
	if h, err := loadFinalHD(tmp); err != nil || h != nil {
		t.Fatal("missing manifest not optional")
	}
}
