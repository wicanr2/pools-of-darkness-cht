package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNarrowScaleEdges(t *testing.T) {
	cases := 0
	for _, kind := range []string{"full-height", "eight-parents", "overlap", "stale-buffer"} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 1024} {
				v := scale5615Seed(t, 1, 3, 2, 25, 15)
				v.dos = dos
				switch kind {
				case "full-height":
					putRowWord(v, 0x48a10+0x6bd+2, 0)
					putRowWord(v, 0x48a10+0xabd+2, 199)
				case "eight-parents":
					for i := uint16(2); i <= 9; i++ {
						seg, flags := uint16(0xa000), uint16(0x80)
						if i < 9 {
							seg, flags = (i+1)*2, 0xc0
						}
						for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: 0, 0x6bd: 0, 0x8bd: 319, 0xabd: 199, 0xcbd: flags} {
							putRowWord(v, 0x48a10+uint32(off+i*2), value)
						}
					}
				case "stale-buffer":
					stack := uint32(v.SS())*16 + uint32(v.SP())
					putRowWord(v, stack+10, 80)
					v.mem[0x48a10+0x1252] = 0xa5
				case "overlap":
					for off, value := range map[uint16]uint16{0x4bd: 25, 0x6bd: 15, 0x8bd: 25, 0xabd: 17} {
						putRowWord(v, 0x48a10+uint32(off+2), value)
					}
				}
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := paletteMachineView{b}
					p := Plan(w, budget)
					if p != nil {
						if !p.Algorithmic && scale5615FixedRaster(p.Writes, p.Transfers) {
							t.Fatal("fixed narrow edge")
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
					} else {
						switch w.IP() {
						case 0x5813, 0x5879, 0x587d, 0x5856, 0x585a:
							if w.CS() == 0x48a1 {
								t.Fatalf("CPU narrow edge %s IP%x", kind, w.IP())
							}
						}
						if e := b.Step(); e != nil {
							t.Fatal(e)
						}
					}
				}
				paletteMachinesEqual(t, a, b)
				if kind == "stale-buffer" {
					for row := uint32(8); row <= 10; row++ {
						if a.Mem[0xa0000+row*320+18] != 0xa5 {
							t.Fatal("oracle did not preserve stale byte")
						}
					}
				}
				cases++
			}
		}
	}
	t.Logf("width1 full height/eight-parent/overlap/stale-buffer SaveState+PortLog cases=%d CPU/fixed row preparation/pixels=0", cases)
}
