// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type dissolve4337WaitView struct {
	*view
	snapshot machine.VGAStatusSnapshot
	known    bool
}

func (v dissolve4337WaitView) VGAStatusSnapshot() (machine.VGAStatusSnapshot, bool) {
	return v.snapshot, v.known
}
func (v dissolve4337WaitView) Steps() uint64 { return v.snapshot.Steps }
func TestNativeDissolve4337Whole(t *testing.T) {
	cases, plans, waits, metadata := 0, 0, 0, 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}, {17, 9}, {20, 10}} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{512, 4096} {
				v := dissolve4337Seed(t, dims[0], dims[1])
				v.dos = dos
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
					p := dissolve4337(w, budget)
					if p != nil {
						plans++
					} else {
						snapshot, known := b.VGAStatusSnapshot()
						p = planWait(dissolve4337WaitView{w, snapshot, known}, budget)
						if p != nil {
							waits++
						}
					}
					if p != nil {
						if err := b.CommitHostBlock(p); err != nil {
							t.Fatal(err)
						}
					} else {
						if w.CS() == 0x48a1 && w.IP() != 0x45ca && w.IP() != 0x45cb && !(w.IP() >= 0x2a90 && w.IP() <= 0x2aa7) {
							t.Fatalf("dissolve whole CPU dims%v dos%v budget%d IP%x", dims, dos, budget, w.IP())
						}
						if err := b.Step(); err != nil {
							t.Fatal(err)
						}
						metadata++
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
					t.Fatalf("dissolve whole SaveState differs dims%v dos%v budget%d steps%d/%d", dims, dos, budget, a.Steps, b.Steps)
				}
				cases++
			}
		}
	}
	t.Log("dissolve4337 whole machine comparisons", cases, "algorithmic plans", plans, "shared VGA wait plans", waits, "CPU control steps", metadata)
}
