package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNarrowScaleWhole(t *testing.T) {
	cases, g, f, c := 0, 0, 0, 0
	for _, height := range []uint16{1, 3, 9} {
		for _, factor := range []uint16{1, 2, 3, 8} {
			for _, center := range [][2]int16{{25, 15}, {-5, -3}, {70, 80}} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 7, 64, 10000} {
						v := scale5615Seed(t, 1, height, factor, center[0], center[1])
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
									if scale5615FixedRaster(p.Writes, p.Transfers) {
										t.Fatal("fixed narrow raster")
									}
								}
								if e := b.CommitHostBlock(p); e != nil {
									t.Fatal(e)
								}
							} else {
								switch w.IP() {
								case 0x5813, 0x5879, 0x587d, 0x5856, 0x585a:
									if w.CS() == 0x48a1 {
										t.Fatalf("CPU narrow raster h%d factor%d center%v dos%v budget%d IP%x", height, factor, center, dos, budget, w.IP())
									}
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
		}
	}
	t.Logf("width1 scale whole SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed row preparation/pixels=0", cases, g, f, c)
}
