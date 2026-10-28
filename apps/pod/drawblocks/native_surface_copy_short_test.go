package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestSurfaceShortAudit(t *testing.T) {
	cases, goPixels, fixedPixels, cpuWrites := 0, 0, 0, 0
	fixedLocations := map[uint16]int{}
	for _, bx := range []bool{false, true} {
		for _, by := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 7, 17, 24, 1024} {
					v := copySurfaceSeed(t, 3, 2, bx, by)
					v.dos = dos
					v.s.Seg[cpu.SS] = 0x8000
					v.s.R[cpu.SP] = 0xe000
					for off, value := range map[uint32]uint16{0: 0x1234, 2: 0x4567, 4: 0, 6: 1} {
						putRowWord(v, 0x8e000+off, value)
					}
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					for b.Steps < 100000 {
						if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
							break
						}
						w := paletteMachineView{b}
						p := Plan(w, budget)
						if p != nil {
							pixels := 0
							for _, x := range p.Writes {
								if x.Address >= 0xa0000 && x.Address < 0xafa00 {
									pixels++
									if !p.Algorithmic {
										fixedLocations[x.IP]++
									}
								}
							}
							for _, x := range p.Transfers {
								if x.Address >= 0xa0000 && x.Address < 0xafa00 {
									pixels += int(x.Count) * int(x.Width)
									if !p.Algorithmic {
										fixedLocations[x.IP]++
									}
								}
							}
							if p.Algorithmic {
								goPixels += pixels
							} else {
								fixedPixels += pixels
							}
							if e := b.CommitHostBlock(p); e != nil {
								t.Fatal(e)
							}
						} else {
							if w.CS() == 0x48a1 && (w.IP() == 0x4169 || w.IP() == 0x4177) {
								cpuWrites++
								if cpuWrites <= 12 {
									t.Logf("row-valid=%v flag-want=%x width=%d", planSurfaceCopyRows(surfaceRowProjection{w}, 24) != nil, copyAddFlags(w.Flags(), w.SI()-1, 1), w.Read16(0x48a10+0x1135))
									t.Logf("CPU gap bx%v by%v dos%v budget%d IP%x AX%x BX%x CX%x SI%x DI%x flags%x", bx, by, dos, budget, w.IP(), w.AX(), w.BX(), w.CX(), w.SI(), w.DI(), w.Flags())
								}
							}
							if e := b.Step(); e != nil {
								t.Fatal(e)
							}
						}
					}
					paletteMachinesEqual(t, a, b)
					cases++
				}
			}
		}
	}
	if fixedPixels != 0 || cpuWrites != 0 {
		t.Fatalf("surface span fallback fixed%d CPU%d", fixedPixels, cpuWrites)
	}
	t.Logf("surface-copy short SaveState cases=%d Go pixel bytes=%d fixed pixel bytes=%d CPU write instructions=%d fixed source IPs=%v", cases, goPixels, fixedPixels, cpuWrites, fixedLocations)
}
