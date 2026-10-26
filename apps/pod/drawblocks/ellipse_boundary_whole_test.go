package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeEllipseBoundaryWhole(t *testing.T) {
	cases, plans := 0, 0
	for _, shape := range [][4]int32{{160, 100, 160, 100}, {160, 100, 1, 1}, {160, 100, 2, 1}, {479, 299, 160, 100}, {-160, -100, 160, 100}, {50, 40, 160, 1}, {50, 40, 1, 100}, {50, 40, 101, 73}} {
		for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{256} {
					v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
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
							if !p.Algorithmic {
								t.Fatal("full ellipse not algorithmic")
							}
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
							used++
						} else {
							t.Fatalf("ellipse stepped original code shape%v box%v dos%v budget%d IP%x", shape, box, dos, budget, w.IP())
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
						t.Fatal("ellipse full machine differs", shape, box, dos, budget)
					}
					cases++
					plans += used
				}
			}
		}
	}
	t.Log("ellipse boundary whole comparisons", cases, "algorithmic plans", plans)
}
