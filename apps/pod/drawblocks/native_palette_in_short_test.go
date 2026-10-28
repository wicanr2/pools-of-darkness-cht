package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

func TestPaletteWholeShort(t *testing.T) {
	cases, goPlans, fixed, cpuSteps := 0, 0, 0, 0
	phases := map[uint16]int{}
	for _, entry := range []uint16{0x45e0, 0x4944} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 2, 7, 27, 1024} {
				v := fadeInWholeSeed(t, 3, 2)
				v.s.IP = entry
				v.dos = dos
				a, b := fadeInMachine(v), fadeInMachine(v)
				runCall(t, a, v, false)
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
					p := Plan(paletteMachineView{b}, budget)
					if p != nil {
						if p.Algorithmic {
							goPlans++
							if paletteSingleWriteIP(w.IP()) {
								phases[w.IP()]++
							}
						} else {
							if paletteFixedSingleWrites(p.Writes, p.Transfers) {
								t.Fatalf("fixed palette write IP%x", w.IP())
							}
							fixed++
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
					} else {
						if w.CS() == 0x48a1 && (paletteSingleWriteIP(w.IP()) || w.IP() == 0x2ab2 || w.IP() == 0x2ab7 || w.IP() == 0x2abb || w.IP() == 0x2abf) {
							t.Fatalf("CPU palette write %x dos%v budget%d regs%+v", w.IP(), dos, budget, w.s)
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
				if !bytes.Equal(sa.Bytes(), sb.Bytes()) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
					t.Fatalf("palette short differs entry%x dos%v budget%d steps%d/%d IP%x", entry, dos, budget, a.Steps, b.Steps, b.CPU.IP)
				}
				cases++
			}
		}
	}
	if len(phases) != 16 {
		t.Fatalf("not all single phases used %v", phases)
	}
	t.Logf("palette in/out short SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d CPU/fixed selected palette writes=0; CPU DAC OUT=0", cases, goPlans, fixed, cpuSteps)
}
