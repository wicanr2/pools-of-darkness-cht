// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestFlip5886Selector(t *testing.T) {
	cases, algorithmic, fixed, metadata := 0, 0, 0, 0
	for _, dims := range [][2]uint16{{1, 1}, {2, 3}, {9, 7}, {17, 8}, {40, 3}} {
		for _, h := range []bool{false, true} {
			for _, ver := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 7, 27, 128} {
						v := flip5886Seed(t, dims[0], dims[1], h, ver)
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
									if flip5886FixedRaster(p.Writes, p.Transfers) {
										t.Fatal("fixed flip raster escaped")
									}
								}
							} else {
								if w.CS() == 0x48a1 && (w.IP() == 0x59c8 || w.IP() == 0x59cf || w.IP() == 0x59d3 || w.IP() == 0x59f1 || w.IP() == 0x59f8 || w.IP() == 0x59fc || w.IP() == 0x5a0c || w.IP() == 0x5a10) {
									t.Fatalf("CPU flip raster dims%v horizontal%v vertical%v dos%v budget%d IP%x", dims, h, ver, dos, budget, w.IP())
								}
								if err := b.Step(); err != nil {
									t.Fatal(err)
								}
								metadata++
							}
						}
						if h || ver {
							if used == 0 {
								t.Fatal("no Go flip plan")
							}
						}
						var sa, sb bytes.Buffer
						if err := a.SaveState(&sa); err != nil {
							t.Fatal(err)
						}
						if err := b.SaveState(&sb); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("flip selector SaveState differs dims%v horizontal%v vertical%v dos%v budget%d", dims, h, ver, dos, budget)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("flip5886 selector comparisons", cases, "algorithmic plans", algorithmic, "fixed metadata plans", fixed, "CPU metadata steps", metadata)
}
