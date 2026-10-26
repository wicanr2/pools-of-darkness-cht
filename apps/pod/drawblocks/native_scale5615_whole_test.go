// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestScale5615Whole(t *testing.T) {
	cases, plans := 0, 0
	for _, dims := range [][2]uint16{{2, 1}, {3, 2}, {9, 7}, {17, 9}, {40, 20}} {
		for _, factor := range []uint16{1, 2, 3, 5, 8} {
			for _, center := range [][2]int16{{25, 15}, {20, 10}, {29, 19}, {-5, -3}, {70, 80}} {
				for _, dos := range []bool{false, true} {
					v := scale5615Seed(t, dims[0], dims[1], factor, center[0], center[1])
					v.dos = dos
					a, b := callMachine(v), callMachine(v)
					runCall(t, a, v, false)
					for b.Steps < 1000000 {
						if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
							break
						}
						w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
						p := scale5615(w, 10000)
						if p == nil {
							t.Fatalf("scale whole fell to CPU dims%v factor%d center%v dos%v IP%x", dims, factor, center, dos, w.IP())
						}
						if err := b.CommitHostBlock(p); err != nil {
							t.Fatal(err)
						}
						plans++
					}
					var sa, sb bytes.Buffer
					if err := a.SaveState(&sa); err != nil {
						t.Fatal(err)
					}
					if err := b.SaveState(&sb); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
						t.Fatalf("scale whole SaveState differs dims%v factor%d center%v dos%v", dims, factor, center, dos)
					}
					cases++
				}
			}
		}
	}
	t.Log("scale5615 full Go machine comparisons", cases, "algorithmic plans", plans)
}
