package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func primitiveSeed(t *testing.T, kind string, xy [4]uint16, stride uint16) *view {
	switch kind {
	case "point":
		return pointSeed(t, false, xy[0], xy[1], stride)
	case "region":
		return regionSeed(t, xy, stride, false)
	}
	return axisSeed(t, xy, stride)
}

func TestPrimitiveBoundaries(t *testing.T) {
	counts := map[string]int{}
	zero := 0
	seen := map[uint16]bool{}
	shapes := map[string][][4]uint16{
		"point":  {{0, 0, 0, 0}, {17, 99, 0, 0}, {159, 199, 0, 0}, {319, 199, 0, 0}, {0xffff, 0, 0, 0}},
		"region": {{0, 0, 0, 0}, {0, 0, 1, 1}, {17, 19, 23, 23}, {0xffff, 0xffff, 320, 3}, {319, 199, 319, 199}},
		"line":   {{17, 19, 35, 25}, {35, 19, 17, 25}, {17, 19, 23, 39}, {23, 19, 17, 39}, {0, 0, 159, 100}, {159, 0, 0, 100}, {0xffff, 0xffff, 40, 20}, {70, 0, 20, 80}},
	}
	for kind, inputs := range shapes {
		for _, xy := range inputs {
			for _, stride := range []uint16{160, 320} {
				for _, dos := range []bool{false, true} {
					for _, flags := range []uint16{2, 0xed7} {
						v := primitiveSeed(t, kind, xy, stride)
						v.dos = dos
						v.s.Flags = flags
						m := callMachine(v)
						for m.Steps < 100000 {
							if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
								break
							}
							if m.CPU.Seg[cpu.CS] == 0x48a1 && primitivePixelIP(m.CPU.IP) {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
								p := planPrimitivePixel(w, 1)
								if p == nil {
									t.Fatalf("primitive refused %s xy%v stride%d IP%x regs%+v", kind, xy, stride, w.IP(), w.s)
								}
								compare(t, w, p)
								counts[kind]++
								seen[w.IP()] = true
								if w.CX() == 0 && (w.IP() == 0x345c || w.IP() == 0x3460) {
									zero++
								}
							}
							if e := m.Step(); e != nil {
								t.Fatal(e)
							}
						}
					}
				}
			}
		}
	}
	if len(seen) != 7 {
		t.Fatal("missing primitive phases", seen)
	}
	t.Logf("primitive ordered original point=%d region=%d diagonal=%d zero-count=%d pixel phases=%d", counts["point"], counts["region"], counts["line"], zero, len(seen))
}

func primitiveWhole(t *testing.T, v *view, budget uint64) (int, int, int) {
	t.Helper()
	a, b := callMachine(v), callMachine(v)
	runCall(t, a, v, false)
	g, f, c := 0, 0, 0
	for b.Steps < 1000000 {
		if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
			break
		}
		w := paletteMachineView{b}
		p := Plan(w, budget)
		if p != nil {
			if p.Algorithmic {
				g++
			} else {
				f++
				if primitiveFixedPixels(p.Writes, p.Transfers) {
					t.Fatal("fixed primitive pixels")
				}
			}
			if e := b.CommitHostBlock(p); e != nil {
				t.Fatal(e)
			}
		} else {
			if w.CS() == 0x48a1 && primitivePixelIP(w.IP()) {
				t.Fatalf("CPU primitive pixel IP%x budget%d regs%+v", w.IP(), budget, boundedMemoryRegisters(w))
			}
			if e := b.Step(); e != nil {
				t.Fatal(e)
			}
			c++
		}
	}
	paletteMachinesEqual(t, a, b)
	return g, f, c
}

func TestPrimitiveWhole(t *testing.T) {
	cases, g, f, c := 0, 0, 0, 0
	shapes := map[string][][4]uint16{
		"point":  {{0, 0, 0, 0}, {17, 99, 0, 0}, {319, 199, 0, 0}, {0xffff, 0, 0, 0}},
		"region": {{0, 0, 0, 0}, {0, 0, 1, 1}, {17, 19, 23, 23}, {0xffff, 0xffff, 320, 3}},
		"line":   {{17, 19, 35, 25}, {35, 19, 17, 25}, {17, 19, 23, 39}, {23, 19, 17, 39}, {0xffff, 0xffff, 40, 20}, {70, 0, 20, 80}},
	}
	for kind, inputs := range shapes {
		for _, xy := range inputs {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 2, 7, 1024} {
					v := primitiveSeed(t, kind, xy, 320)
					v.dos = dos
					a, b, d := primitiveWhole(t, v, budget)
					g += a
					f += b
					c += d
					cases++
				}
			}
		}
	}
	t.Logf("primitive whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed primitive pixels=0", cases, g, f, c)
}
