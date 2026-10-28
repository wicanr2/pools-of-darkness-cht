// Spec025/#49: original glyph output and complete-machine comparisons.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeGlyph5AA4Whole(t *testing.T) {
	cases, goPlans, fixed, cpuSteps := 0, 0, 0, 0
	for _, width := range []uint16{1, 2, 3, 7, 9} {
		for _, height := range []uint16{1, 2, 5} {
			for _, mapped := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 2, 7, 27, 1024} {
						v := glyph5aa4Seed(t, width, height, mapped)
						v.dos = dos
						a, b := callMachine(v), callMachine(v)
						runCall(t, a, v, false)
						for b.Steps < 10000 {
							if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
								break
							}
							w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
							p := Plan(w, budget)
							if p != nil {
								if p.Algorithmic {
									goPlans++
								} else {
									if glyph5aa4FixedPixels(p.Writes, p.Transfers) {
										t.Fatal("fixed glyph pixel")
									}
									fixed++
								}
								if e := b.CommitHostBlock(p); e != nil {
									t.Fatal(e)
								}
							} else {
								if w.CS() == 0x48a1 && (w.IP() == 0x5c02 || w.IP() == 0x5c2c) {
									t.Fatalf("CPU glyph pixel width%d height%d mapped%v budget%d IP%x", width, height, mapped, budget, w.IP())
								}
								if e := b.Step(); e != nil {
									t.Fatal(e)
								}
								cpuSteps++
							}
						}
						var sa, sb bytes.Buffer
						if e := a.SaveState(&sa); e != nil {
							t.Fatal(e)
						}
						if e := b.SaveState(&sb); e != nil {
							t.Fatal(e)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("whole glyph differs %dx%d mapped%v dos%v budget%d steps%d/%d IP%x", width, height, mapped, dos, budget, a.Steps, b.Steps, b.CPU.IP)
						}
						cases++
					}
				}
			}
		}
	}
	t.Logf("glyph whole SaveState comparisons=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed glyph pixels=0", cases, goPlans, fixed, cpuSteps)
}
