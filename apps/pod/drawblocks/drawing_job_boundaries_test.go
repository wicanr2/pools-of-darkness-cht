package drawblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"testing"
)

func drawingBoundarySamples(t *testing.T) []*view {
	t.Helper()
	root := os.Getenv("POD_ROOT")
	if root == "" {
		t.Skip("original drawing boundary fixtures absent")
	}
	dir := filepath.Join(root, "workplace/probe/draw-intercept-v1/native-fixed-control-v1/trace-native")
	raw, e := os.ReadFile(filepath.Join(dir, "fixed-records.json"))
	if e != nil {
		t.Fatal(e)
	}
	var records []struct {
		Before struct {
			R                    [8]uint16
			Seg                  [4]uint16
			IP, Flags            uint16
			Memory, MemorySHA256 string
		}
	}
	if e = json.Unmarshal(raw, &records); e != nil {
		t.Fatal(e)
	}
	var out []*view
	for _, r := range records {
		b := r.Before
		if b.IP == 0x2a66 || b.IP == 0x2a77 {
			continue
		}
		mem, e := os.ReadFile(filepath.Join(dir, b.Memory))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(mem)) != b.MemorySHA256 {
			t.Fatal("fixture hash")
		}
		out = append(out, &view{mem: mem, s: machine.HostRegisters{R: b.R, Seg: b.Seg, IP: b.IP, Flags: b.Flags}})
	}
	if len(out) != 16 {
		t.Fatal("boundary sample count", len(out))
	}
	return out
}
func drawingView(m *machine.Machine, dos bool) *view {
	return &view{mem: append([]byte(nil), m.Mem...), dos: dos, s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
}
func drawingBoundaryCompare(t *testing.T, v *view, p *machine.HostWorkPlan) {
	t.Helper()
	if p == nil || !p.Algorithmic {
		t.Fatalf("typed boundary refused %04x", v.IP())
	}
	compare(t, v, p)
	a, b := callMachine(v), callMachine(v)
	for n := uint64(0); n < p.Instructions; n++ {
		if e := a.Step(); e != nil {
			t.Fatal(e)
		}
	}
	if e := b.CommitHostBlock(p); e != nil {
		t.Fatal(e)
	}
	paletteMachinesEqual(t, a, b)
}
func TestDrawingBoundaryNormalSamples(t *testing.T) {
	n := 0
	for _, seed := range drawingBoundarySamples(t) {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 2, 6, 13, 128} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
				p := planDrawingJobBoundary(v, budget)
				drawingBoundaryCompare(t, v, p)
				if p.Instructions > budget {
					t.Fatal("budget crossed")
				}
				n++
			}
		}
	}
	t.Logf("drawing boundary normal original samples comparisons=%d", n)
}

func walkDrawingBoundaries(t *testing.T, v *view, stop func(uint16) bool) (int, int) {
	t.Helper()
	m := callMachine(v)
	n, steps := 0, 0
	for m.Steps < 100 {
		if stop(m.CPU.IP) {
			return n, steps
		}
		w := drawingView(m, v.dos)
		if !drawingJobBoundaryIP(w.IP()) {
			t.Fatalf("missing handoff IP%x", w.IP())
		}
		for _, budget := range []uint64{1, 3, 128} {
			p := planDrawingJobBoundary(w, budget)
			drawingBoundaryCompare(t, w, p)
			n++
			steps += int(p.Instructions)
		}
		if e := m.Step(); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("boundary did not hand off")
	return 0, 0
}
func TestDrawingHeaderBoundaryPaths(t *testing.T) {
	samples := drawingBoundarySamples(t)
	var seed *view
	for _, v := range samples {
		if v.IP() == 0x689e {
			seed = v
			break
		}
	}
	cases, n := 0, 0
	for _, data := range []struct {
		raw, offset       uint16
		width, compressed byte
		marker            uint16
	}{{0, 0, 1, 2, 0}, {319, 0, 1, 2, 0}, {300, 0, 30, 2, 0}, {0, 65531, 10, 2, 0}, {320, 0, 3, 2, 0}, {0, 65526, 3, 2, 0}, {0x80f1, 0, 79, 69, 0}, {0, 0, 1, 2, 0x8000}} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			v.s.Seg[cpu.DS] = 0x6000
			v.s.R[cpu.SI] = 0x100
			v.s.R[cpu.CX] = 0xabcd
			v.s.R[cpu.BX] = 0xcdef
			v.s.R[cpu.AX] = 0xabcd
			v.s.R[cpu.DI] = 0x5555
			putRowWord(v, 0x60100, data.raw)
			v.mem[0x60102] = data.width
			v.mem[0x60103] = data.compressed
			putRowWord(v, uint32(v.SS())*16+uint32(v.BP())+8, data.offset)
			putRowWord(v, 0x48a10+0x120c, 3200)
			putRowWord(v, 0x48a10+0xec1, 0)
			putRowWord(v, 0x48a10+0xec5, 319)
			putRowWord(v, 0x48a10+0x1250, data.marker)
			x, _ := walkDrawingBoundaries(t, v, func(ip uint16) bool {
				return ip == 0x68f4 || ip == 0x691c || ip == 0x6840 || ip == 0x6848 || ip == 0x6850
			})
			n += x
			cases++
		}
	}
	t.Logf("drawing header paths=%d ordered and SaveState checkpoints=%d", cases, n)
}
func TestDrawingBlendBoundaryPaths(t *testing.T) {
	var seed *view
	for _, v := range drawingBoundarySamples(t) {
		if v.IP() == 0x4b6c {
			seed = v
			break
		}
	}
	cases, n := 0, 0
	for _, data := range [][5]byte{{0, 0, 0, 1, 1}, {9, 26, 63, 10, 255}, {43, 2, 63, 10, 1}, {30, 33, 63, 10, 1}, {49, 60, 63, 20, 255}, {60, 60, 63, 20, 1}, {63, 63, 0, 1, 1}} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			v.s.R[cpu.SI] = 2
			v.s.R[cpu.CX] = 1
			v.s.R[cpu.AX] = 0xabcd
			v.s.R[cpu.BX] = 0xcdef
			putRowWord(v, 0x48a10+0x116f, 1)
			v.mem[0x48a10+0x1173] = data[2]
			for i, off := range []uint32{0x1e72, 0x2172, 0x1872, 0x2472} {
				value := []byte{data[0], data[1], data[3], data[4]}[i]
				v.mem[0x48a10+off+2] = value
			}
			x, _ := walkDrawingBoundaries(t, v, func(ip uint16) bool { return ip == 0x4b8d })
			n += x
			cases++
		}
	}
	t.Logf("drawing RGB blend paths=%d ordered and SaveState checkpoints=%d", cases, n)
}
func TestDrawingColorBoundaryPaths(t *testing.T) {
	var seed *view
	for _, v := range drawingBoundarySamples(t) {
		if v.IP() == 0x4b9a {
			seed = v
			break
		}
	}
	cases, n := 0, 0
	for _, used := range []uint16{1, 3, 109, 256} {
		for _, done := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
				putRowWord(v, 0x48a10+0x116f, used)
				v.s.R[cpu.DI] = used * 2
				v.s.R[cpu.SI] = 0x1872
				if done {
					v.s.R[cpu.DI] = 0
					v.s.R[cpu.SI] = 0x1872 + used*3
				} else {
					putRowWord(v, 0x48a10+0x1572+uint32(used*2-2), 255)
					v.mem[0x48a10+0x1872] = 0
					v.mem[0x48a10+0x1873] = 31
					v.mem[0x48a10+0x1874] = 63
				}
				x, _ := walkDrawingBoundaries(t, v, func(ip uint16) bool { return ip == 0x2aa8 || ip == 0x4b55 })
				n += x
				cases++
			}
		}
	}
	t.Logf("drawing RGB call paths=%d ordered and SaveState checkpoints=%d", cases, n)
}
func TestDrawingFillBoundaryPaths(t *testing.T) {
	cases, n := 0, 0
	for _, stride := range []uint16{319, 320} {
		for _, width := range []uint16{1, 2, 9} {
			for _, dos := range []bool{false, true} {
				v := surfaceFillSpanSeed(t, width, 3)
				v.dos = dos
				putRowWord(v, 0x48a10+0x8bf, stride-1)
				m := callMachine(v)
				for m.CPU.IP != 0x4268 && m.Steps < 1000 {
					if e := m.Step(); e != nil {
						t.Fatal(e)
					}
				}
				if m.CPU.IP != 0x4268 {
					t.Fatal("no row entry")
				}
				w := drawingView(m, dos)
				x, _ := walkDrawingBoundaries(t, w, func(ip uint16) bool { return ip == 0x42a1 })
				n += x
				cases++
			}
		}
	}
	t.Logf("drawing fill row paths=%d ordered and SaveState checkpoints=%d", cases, n)
}
func TestDrawingBoundaryRefusals(t *testing.T) {
	cases := 0
	for _, seed := range drawingBoundarySamples(t) {
		for _, kind := range []string{"budget", "mode", "CS", "IP", "DF", "model", "8086", "code"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(128)
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS]++
			case "IP":
				v.s.IP = 0x7500
			case "DF":
				v.s.Flags |= cpu.DF
			case "code":
				start := uint32(0x689e)
				if v.IP() < 0x4300 {
					start = 0x41af
				} else if v.IP() >= 0x4b9a && v.IP() <= 0x4bc4 {
					start = 0x4b9a
				} else if v.IP() >= 0x4b6c && v.IP() <= 0x4b8b {
					start = 0x4b6c
				}
				v.mem[0x48a10+start] ^= 1
			}
			var w CPUView = v
			if kind == "model" {
				w = primitiveUnknownModel{v}
			}
			if kind == "8086" {
				w = primitive8086Model{v}
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planDrawingJobBoundary(w, budget) != nil {
				t.Fatal("bad boundary admitted", seed.IP(), kind)
			}
			if !bytes.Equal(before, v.mem) || regs != v.s {
				t.Fatal("refusal mutation")
			}
			cases++
		}
	}
	t.Logf("drawing boundary unchanged refusals=%d", cases)
}
