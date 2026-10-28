// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337Selector(t *testing.T) {
	cases, algorithmic, fixed, metadata := 0, 0, 0, 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {3, 3}, {7, 5}, {17, 9}} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 7, 27, 128} {
				v := dissolve4337Seed(t, dims[0], dims[1])
				v.dos = dos
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				used := 0
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
					snapshot, known := b.VGAStatusSnapshot()
					wrapped := dissolve4337WaitView{w, snapshot, known}
					if p := Plan(wrapped, budget); p != nil {
						if err := b.CommitHostBlock(p); err != nil {
							t.Fatal(err)
						}
						if p.Algorithmic {
							used++
							algorithmic++
						} else {
							fixed++
							if dissolve4337FixedPixels(p.Writes, p.Transfers) {
								t.Fatal("fixed dissolve pixel escaped")
							}
						}
					} else {
						if w.CS() == 0x48a1 && (w.IP() == 0x4569 || w.IP() == 0x45bb) {
							t.Fatalf("CPU dissolve pixel dims%v dos%v budget%d IP%x", dims, dos, budget, w.IP())
						}
						if err := b.Step(); err != nil {
							t.Fatal(err)
						}
						metadata++
					}
				}
				if used == 0 {
					t.Fatal("no native dissolve plan")
				}
				var sa, sb bytes.Buffer
				if err := a.SaveState(&sa); err != nil {
					t.Fatal(err)
				}
				if err := b.SaveState(&sb); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
					t.Fatalf("dissolve selector SaveState differs dims%v dos%v budget%d", dims, dos, budget)
				}
				cases++
			}
		}
	}
	t.Log("dissolve4337 selector comparisons", cases, "algorithmic plans including shared wait", algorithmic, "fixed metadata plans", fixed, "CPU control steps", metadata)
}
