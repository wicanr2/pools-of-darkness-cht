package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func runSurfaceSpanWhole(t *testing.T, v *view, budget uint64) (int, int, int) {
	t.Helper()
	a, b := callMachine(v), callMachine(v)
	runCall(t, a, v, false)
	goPlans, fixed, control := 0, 0, 0
	for b.Steps < 1000000 {
		if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
			break
		}
		w := paletteMachineView{b}
		p := Plan(w, budget)
		if p != nil {
			if p.Algorithmic {
				goPlans++
			} else {
				fixed++
				if surfaceCopyFixedRaster(p.Writes, p.Transfers) {
					t.Fatal("fixed surface pixels")
				}
			}
			if err := b.CommitHostBlock(p); err != nil {
				t.Fatal(err)
			}
		} else {
			if w.CS() == 0x48a1 && (w.IP() == 0x4169 || w.IP() == 0x4177) {
				t.Fatalf("CPU surface pixel IP%x CX%d", w.IP(), w.CX())
			}
			if err := b.Step(); err != nil {
				t.Fatal(err)
			}
			control++
		}
	}
	paletteMachinesEqual(t, a, b)
	return goPlans, fixed, control
}

func TestSurfaceSpanWhole(t *testing.T) {
	cases, goPlans, fixed, control := 0, 0, 0, 0
	for _, width := range []uint16{1, 2, 3, 8, 19} {
		for _, height := range []uint16{1, 3} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						for _, budget := range []uint64{1, 2, 7, 17, 1024} {
							v := surfaceSpanSeed(t, width, height, bx, by, dos)
							g, f, c := runSurfaceSpanWhole(t, v, budget)
							goPlans += g
							fixed += f
							control += c
							cases++
						}
					}
				}
			}
		}
	}
	t.Logf("surface span whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed surface pixels=0", cases, goPlans, fixed, control)
}

func TestSurfaceSpanEdges(t *testing.T) {
	cases := 0
	for _, kind := range []string{"full-canvas", "eight-parents", "zero-origin-backward"} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 1024} {
				v := surfaceSpanSeed(t, 3, 3, false, false, dos)
				switch kind {
				case "full-canvas":
					for slot := uint16(0); slot < 2; slot++ {
						for off, value := range map[uint16]uint16{0x4bd: 0, 0x6bd: 0, 0x8bd: 319, 0xabd: 199} {
							putRowWord(v, 0x48a10+uint32(off+slot*2), value)
						}
					}
				case "eight-parents":
					for slot := uint16(2); slot <= 9; slot++ {
						seg, flags := uint16(0xa000), uint16(0x80)
						if slot < 9 {
							seg, flags = (slot+1)*2, 0xc0
						}
						for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: 0, 0x6bd: 0, 0x8bd: 319, 0xabd: 199, 0xcbd: flags} {
							putRowWord(v, 0x48a10+uint32(off+slot*2), value)
						}
					}
				case "zero-origin-backward":
					for slot := uint16(0); slot < 2; slot++ {
						x := slot * 2
						for off, value := range map[uint16]uint16{0x4bd: x, 0x6bd: 0, 0x8bd: x, 0xabd: 1} {
							putRowWord(v, 0x48a10+uint32(off+slot*2), value)
						}
					}
				}
				runSurfaceSpanWhole(t, v, budget)
				cases++
			}
		}
	}
	t.Logf("surface span full canvas/eight parents/zero-origin backward SaveState cases=%d CPU/fixed surface pixels=0", cases)
}
