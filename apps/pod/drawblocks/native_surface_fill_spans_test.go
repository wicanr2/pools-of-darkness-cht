package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func surfaceFillSpanSeed(t *testing.T, width, height uint16) *view {
	v := pointSeed(t, false, 0, 0, 320)
	v.s.IP = 0x41af
	stack := uint32(v.SS())*16 + uint32(v.SP())
	putRowWord(v, stack+6, 0xab2b)
	for i := uint16(0); i < 2; i++ {
		seg, flags := uint16(2), uint16(0xc0)
		left, top, right, bottom := uint16(0), uint16(0), width-1, height-1
		if i == 1 {
			seg, flags = 0xa000, 0x80
			right, bottom = 319, 199
		}
		for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: left, 0x6bd: top, 0x8bd: right, 0xabd: bottom, 0xcbd: flags} {
			putRowWord(v, 0x48a10+uint32(off+i*2), value)
		}
	}
	return v
}
func TestSurfaceFillSpanBoundaries(t *testing.T) {
	cases, zero := 0, 0
	for _, dims := range [][2]uint16{{1, 1}, {2, 1}, {3, 2}, {9, 9}, {17, 3}, {64, 48}, {320, 200}} {
		for _, dos := range []bool{false, true} {
			v := surfaceFillSpanSeed(t, dims[0], dims[1])
			v.dos = dos
			m := callMachine(v)
			for m.Steps < 100000 {
				if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
					break
				}
				if m.CPU.Seg[cpu.CS] == 0x48a1 && surfaceFillPixelIP(m.CPU.IP) {
					w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
					p := planSurfaceFillSpan(w, 1)
					if p == nil {
						t.Fatalf("surface span rejected dims%v IP%x", dims, w.IP())
					}
					compare(t, w, p)
					cases++
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
	t.Logf("surface-fill span ordered original comparisons=%d zero-count=%d", cases, zero)
}
func TestSurfaceFillSpanWhole(t *testing.T) {
	cases, g, f, c := 0, 0, 0, 0
	for _, dims := range [][2]uint16{{1, 1}, {2, 1}, {3, 2}, {9, 9}, {17, 3}, {64, 48}, {320, 200}} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 2, 7, 27, 1024} {
				v := surfaceFillSpanSeed(t, dims[0], dims[1])
				v.dos = dos
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
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
							if surfaceFillFixedPixels(p.Writes, p.Transfers) {
								t.Fatal("fixed surface fill")
							}
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
					} else {
						if w.CS() == 0x48a1 && surfaceFillPixelIP(w.IP()) {
							t.Fatalf("CPU surface fill IP%x", w.IP())
						}
						if e := b.Step(); e != nil {
							t.Fatal(e)
						}
						c++
					}
				}
				paletteMachinesEqual(t, a, b)
				cases++
			}
		}
	}
	t.Logf("surface-fill span whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed surface-fill pixels=0", cases, g, f, c)
}
