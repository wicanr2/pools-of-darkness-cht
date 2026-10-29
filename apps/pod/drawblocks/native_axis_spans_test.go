package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestAxisSpanBoundaries(t *testing.T) {
	cases, zero := 0, 0
	seen := map[uint16]bool{}
	for _, stride := range []uint16{160, 320} {
		for _, xy := range [][4]uint16{{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 159, 0}, {159, 10, 0, 10}, {0xffff, 20, 320, 20}, {17, 0, 17, 0}, {17, 0, 17, 1}, {17, 0, 17, 6}, {17, 0, 17, 7}, {17, 0, 17, 8}, {17, 199, 17, 0}, {17, 0xffff, 17, 200}} {
			for _, dos := range []bool{false, true} {
				for _, flags := range []uint16{2, 0xed7} {
					v := axisSeed(t, xy, stride)
					v.dos = dos
					v.s.Flags = flags
					m := callMachine(v)
					for m.Steps < 100000 {
						if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
							break
						}
						if m.CPU.Seg[cpu.CS] == 0x48a1 && axisSpanPixelIP(m.CPU.IP) {
							w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							p := planAxisSpan(w, 1)
							if p == nil {
								t.Fatalf("axis span refused xy%v stride%d dos%v IP%x regs%+v", xy, stride, dos, w.IP(), w.s)
							}
							compare(t, w, p)
							cases++
							seen[w.IP()] = true
							if w.CX() == 0 {
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
	if len(seen) != 10 {
		t.Fatalf("missing axis pixel position %v", seen)
	}
	t.Logf("axis span independently reached original comparisons=%d zero-count=%d pixel phases=%d", cases, zero, len(seen))
}

func axisWhole(t *testing.T, v *view, budget uint64) (int, int, int) {
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
				if axisSpanFixedPixels(p.Writes, p.Transfers) {
					t.Fatal("fixed axis pixel")
				}
			}
			if err := b.CommitHostBlock(p); err != nil {
				t.Fatal(err)
			}
		} else {
			if w.CS() == 0x48a1 && axisSpanPixelIP(w.IP()) {
				t.Fatalf("CPU axis pixel IP%x budget%d regs%+v", w.IP(), budget, boundedMemoryRegisters(w))
			}
			if err := b.Step(); err != nil {
				t.Fatal(err)
			}
			c++
		}
	}
	paletteMachinesEqual(t, a, b)
	return g, f, c
}

func TestAxisSpanWhole(t *testing.T) {
	cases, g, f, c := 0, 0, 0, 0
	for _, xy := range [][4]uint16{{0, 0, 0, 0}, {0, 0, 1, 0}, {0, 0, 319, 0}, {319, 10, 0, 10}, {0xffff, 20, 320, 20}, {17, 0, 17, 0}, {17, 0, 17, 1}, {17, 0, 17, 6}, {17, 0, 17, 7}, {17, 0, 17, 8}, {17, 199, 17, 0}, {17, 0xffff, 17, 200}} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 2, 7, 1024} {
				v := axisSeed(t, xy, 320)
				v.dos = dos
				a, b, d := axisWhole(t, v, budget)
				g += a
				f += b
				c += d
				cases++
			}
		}
	}
	t.Logf("axis span whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed axis pixels=0", cases, g, f, c)
}

func TestAxisSpanZeroEllipse(t *testing.T) {
	cases, g, f, c := 0, 0, 0, 0
	for _, radius := range [][2]int32{{0, 0}, {0, 2}, {2, 0}, {0, 100}, {160, 0}} {
		for _, filled := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 7, 1024} {
					v := ellipseSeed(t, 160, 100, radius[0], radius[1], [4]uint16{0, 0, 319, 199})
					v.dos = dos
					if filled {
						v.s.IP = 0x39b2
					}
					a, b, d := axisWhole(t, v, budget)
					g += a
					f += b
					c += d
					cases++
				}
			}
		}
	}
	t.Logf("zero-radius outline/filled whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed axis pixels=0", cases, g, f, c)
}
