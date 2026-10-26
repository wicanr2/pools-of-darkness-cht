// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestScale5615Selector(t *testing.T) {
	cases, algorithmic, fixed, metadata := 0, 0, 0, 0
	for _, dims := range [][2]uint16{{2, 1}, {3, 2}, {9, 7}, {17, 9}, {40, 20}} {
		for _, factor := range []uint16{1, 2, 3, 5, 8} {
			for _, center := range [][2]int16{{25, 15}, {20, 10}, {29, 19}, {-5, -3}, {70, 80}} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 7, 27, 128} {
						v := scale5615Seed(t, dims[0], dims[1], factor, center[0], center[1])
						v.dos = dos
						a, b := callMachine(v), callMachine(v)
						runCall(t, a, v, false)
						used := 0
						for b.Steps < 1000000 {
							if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
								break
							}
							w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
							if p := Plan(w, budget); p != nil {
								if err := b.CommitHostBlock(p); err != nil {
									t.Fatal(err)
								}
								if p.Algorithmic {
									used++
									algorithmic++
								} else {
									fixed++
									if scale5615FixedRaster(p.Writes, p.Transfers) {
										t.Fatal("fixed scale raster escaped")
									}
								}
							} else {
								if w.CS() == 0x48a1 && (w.IP() == 0x5813 || w.IP() == 0x5879 || w.IP() == 0x587d || w.IP() == 0x5856 || w.IP() == 0x585a) {
									t.Fatalf("CPU scale raster dims%v factor%d center%v dos%v budget%d IP%x", dims, factor, center, dos, budget, w.IP())
								}
								if err := b.Step(); err != nil {
									t.Fatal(err)
								}
								metadata++
							}
						}
						if used == 0 {
							t.Fatal("no Go scaling plan")
						}
						var sa, sb bytes.Buffer
						if err := a.SaveState(&sa); err != nil {
							t.Fatal(err)
						}
						if err := b.SaveState(&sb); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("scale selector SaveState differs dims%v factor%d center%v dos%v budget%d", dims, factor, center, dos, budget)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("scale5615 short selector comparisons", cases, "algorithmic plans", algorithmic, "fixed metadata plans", fixed, "CPU metadata steps", metadata)
}
