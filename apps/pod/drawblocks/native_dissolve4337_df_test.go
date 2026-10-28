// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337Direction(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {7, 5}, {17, 9}} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{512, 4096} {
				v := dissolve4337Seed(t, dims[0], dims[1])
				v.dos = dos
				v.s.Flags |= cpu.DF
				p := planDissolve4337Prefix(v)
				if p == nil {
					t.Fatal("DF prefix refused")
				}
				compare(t, v, p)
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
					p := dissolve4337(w, budget)
					if p == nil {
						snapshot, known := b.VGAStatusSnapshot()
						p = planWait(dissolve4337WaitView{w, snapshot, known}, budget)
					}
					if p != nil {
						if err := b.CommitHostBlock(p); err != nil {
							t.Fatal(err)
						}
					} else {
						if w.CS() == 0x48a1 && w.IP() != 0x45ca && w.IP() != 0x45cb {
							t.Fatalf("DF dissolve CPU dims%v IP%x", dims, w.IP())
						}
						if err := b.Step(); err != nil {
							t.Fatal(err)
						}
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
					t.Fatalf("DF dissolve SaveState differs dims%v dos%v budget%d", dims, dos, budget)
				}
				cases++
			}
		}
	}
	t.Log("dissolve4337 DF prefix and machine comparisons", cases)
}
