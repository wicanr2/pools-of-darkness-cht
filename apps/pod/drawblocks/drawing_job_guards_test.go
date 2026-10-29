package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"reflect"
	"testing"
)

func TestDrawingBoundarySemanticGuards(t *testing.T) {
	seeds := map[string]*view{}
	for _, v := range drawingBoundarySamples(t) {
		switch v.IP() {
		case 0x689e:
			seeds["header"] = v
		case 0x4b6c:
			seeds["blend"] = v
		case 0x4b9a:
			seeds["color"] = v
		case 0x4268:
			seeds["fill"] = v
		}
	}
	cases := 0
	for kind, seed := range seeds {
		changes := []func(*view){func(v *view) { v.s.Seg[cpu.DS] = 0xffff }}
		switch kind {
		case "header":
			changes = append(changes, func(v *view) { v.s.Seg[cpu.ES]-- }, func(v *view) { v.s.R[cpu.SI] = 65535 }, func(v *view) { v.s.R[cpu.BP] = 65535 }, func(v *view) { v.mem[uint32(v.DS())*16+uint32(v.SI())+2] = 0 }, func(v *view) { v.mem[uint32(v.DS())*16+uint32(v.SI())+3] = 0 }, func(v *view) { putRowWord(v, 0x48a10+0xec5, 320) }, func(v *view) { putRowWord(v, 0x48a10+0xec1, 320) })
		case "blend":
			changes = append(changes, func(v *view) { putRowWord(v, 0x48a10+0x116f, 0) }, func(v *view) { v.s.R[cpu.CX]++ }, func(v *view) { v.mem[0x48a10+0x1173] = 64 }, func(v *view) { v.mem[0x48a10+0x1e72+uint32(v.SI())] = 64 }, func(v *view) { v.mem[0x48a10+0x2172+uint32(v.SI())] = 64 }, func(v *view) {
				for off, value := range map[uint32]byte{0x1e72: 63, 0x2172: 63, 0x1872: 63, 0x2472: 1} {
					v.mem[0x48a10+off+uint32(v.SI())] = value
				}
			})
		case "color":
			changes = append(changes, func(v *view) { v.s.R[cpu.SP]++ }, func(v *view) { v.s.R[cpu.BP] = 2 }, func(v *view) { v.s.R[cpu.DI]++ }, func(v *view) { v.s.R[cpu.SI]++ }, func(v *view) { putRowWord(v, 0x48a10+0x116f, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x116f, 257) }, func(v *view) { putRowWord(v, 0x48a10+0x1572+uint32(v.DI()-2), 256) }, func(v *view) { v.mem[0x48a10+uint32(v.SI())] = 64 })
		case "fill":
			changes = append(changes, func(v *view) { v.s.Seg[cpu.ES]-- }, func(v *view) { v.s.R[cpu.BP]++ }, func(v *view) { putRowWord(v, uint32(v.SS())*16+uint32(v.SP())+12, 256) }, func(v *view) { putRowWord(v, 0x48a10+0xedd, 319) }, func(v *view) { putRowWord(v, 0x48a10+0x1129, v.Read16(0x48a10+0x1129)+1) }, func(v *view) { v.s.R[cpu.BX] = 200 }, func(v *view) { putRowWord(v, 0x48a10+0x113b, v.Read16(0x48a10+0x113b)+1) })
		}
		for _, change := range changes {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			change(v)
			before := append([]byte(nil), v.mem...)
			state := v.s
			if planDrawingJobBoundary(v, 128) != nil {
				t.Fatal("bad semantic input admitted", kind)
			}
			if !bytes.Equal(before, v.mem) || state != v.s {
				t.Fatal("semantic refusal changed input")
			}
			cases++
		}
		if planDrawingJobBoundary(struct{ CPUView }{seed}, 128) != nil {
			t.Fatal("missing CPU model")
		}
		cases++
	}
	t.Logf("drawing boundary semantic/frame/source refusals=%d unchanged", cases)
}

func TestDrawingBoundaryIgnoredData(t *testing.T) {
	samples := drawingBoundarySamples(t)
	cases := 0
	for _, kind := range []string{"end-row", "no-color-advance", "no-next-color"} {
		var seed *view
		for _, v := range samples {
			if kind == "end-row" && v.IP() == 0x689e || kind == "no-color-advance" && v.IP() == 0x4b6c || kind == "no-next-color" && v.IP() == 0x4b9a {
				seed = v
				break
			}
		}
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		switch kind {
		case "end-row":
			putRowWord(v, 0x48a10+0x1250, 0x8000)
			v.s.Seg[cpu.DS] = 0xffff
			v.s.R[cpu.SI] = 65535
		case "no-color-advance":
			i := uint32(v.SI())
			v.mem[0x48a10+0x1e72+i] = 0
			v.mem[0x48a10+0x2172+i] = 0
			v.mem[0x48a10+0x1872+i] = 255
			v.mem[0x48a10+0x2472+i] = 7
		case "no-next-color":
			v.s.R[cpu.DI] = 0
			v.s.R[cpu.SI] = 0x1872 + v.Read16(0x48a10+0x116f)*3
			v.mem[0x48a10+uint32(v.SI())] = 255
		}
		drawingBoundaryCompare(t, v, planDrawingJobBoundary(v, 128))
		cases++
	}
	t.Logf("drawing boundary unused payload positive controls=%d", cases)
}

func TestDrawingBoundaryResumeGuards(t *testing.T) {
	seeds := map[uint16]*view{}
	for _, v := range drawingBoundarySamples(t) {
		seeds[v.IP()] = v
	}
	cases := 0
	for _, c := range []struct {
		entry, target uint16
		change        func(*view)
	}{
		{0x689e, 0x68b4, func(v *view) { v.s.R[cpu.DI]++ }},
		{0x689e, 0x68c6, func(v *view) { v.s.R[cpu.DX]++ }},
		{0x689e, 0x68ac, func(v *view) { putRowWord(v, 0x48a10+0x1250, v.Read16(0x48a10+0x1250)+1) }},
		{0x689e, 0x68e4, func(v *view) { v.s.R[cpu.BX]++ }},
		{0x4b6c, 0x4b82, func(v *view) { v.s.R[cpu.BX] ^= 1 }},
		{0x4b6c, 0x4b86, func(v *view) { v.s.R[cpu.AX]++ }},
		{0x4b9a, 0x4ba5, func(v *view) { putRowWord(v, 0x48a10+0x117f, v.Read16(0x48a10+0x117f)+1) }},
		{0x4268, 0x429f, func(v *view) { v.s.R[cpu.AX]++ }},
	} {
		seed := seeds[c.entry]
		m := callMachine(seed)
		for m.CPU.IP != c.target && m.Steps < 50 {
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
		if m.CPU.IP != c.target {
			t.Fatal("resume target absent")
		}
		v := drawingView(m, false)
		drawingBoundaryCompare(t, v, planDrawingJobBoundary(v, 1))
		c.change(v)
		before := append([]byte(nil), v.mem...)
		state := v.s
		if planDrawingJobBoundary(v, 128) != nil {
			t.Fatal("corrupt resumed state admitted", c.target)
		}
		if !bytes.Equal(before, v.mem) || state != v.s {
			t.Fatal("resume refusal mutation")
		}
		cases++
	}
	t.Logf("drawing boundary real resumed positive and rejection pairs=%d", cases)
}

func TestDrawingBoundarySelector(t *testing.T) {
	cases := 0
	for _, v := range drawingBoundarySamples(t) {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			want := planDrawingJobBoundary(v, 1)
			got := Plan(v, 1)
			if want == nil || !reflect.DeepEqual(want, got) {
				t.Fatalf("selector did not choose typed boundary %x", v.IP())
			}
			cases++
		}
	}
	t.Logf("drawing boundary production selector cases=%d", cases)
}
