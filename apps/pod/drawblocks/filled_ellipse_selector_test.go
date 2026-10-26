package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeFilledEllipseSelector(t *testing.T) {
	cases, plans, metadataSteps, fixedPlans := 0, 0, 0, 0
	for _, shape := range [][4]int32{{160, 100, 1, 2}, {160, 100, 2, 5}, {160, 100, 10, 20}, {50, 40, 40, 20}, {0, 0, 20, 10}, {319, 199, 20, 10}} {
		for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 7, 27, 128} {
					v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
					v.dos = dos
					v.s.IP = 0x39b2
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
							} else {
								fixedPlans++
								if filledEllipseFixedTransfers(p.Transfers) {
									t.Fatal("fixed filled ellipse span escaped")
								}
							}
						} else {
							if _, _, _, pixel := filledEllipseSpanStage(w.IP()); pixel && w.CS() == 0x48a1 {
								t.Fatal("ellipse CPU pixel", shape, box, dos, budget, w.IP())
							}
							if err := b.Step(); err != nil {
								t.Fatal(err)
							}
							metadataSteps++
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
	t.Log("filled ellipse selector comparisons", cases, "algorithmic plans", plans, "CPU metadata steps", metadataSteps, "fixed metadata plans", fixedPlans)
}
