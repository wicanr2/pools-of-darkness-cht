package drawblocks

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func TestNativeRegionSelector(t *testing.T) {
	cases, rows := 0, 0
	for _, dos := range []bool{false, true} {
		for _, budget := range []uint64{27, 71, 1024} {
			for _, sub := range []bool{false, true} {
				v := regionSeed(t, [4]uint16{0, 0, 319, 199}, 320, sub)
				v.dos = dos
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				used := 0
				native := 0
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.IP == 0x1234 && b.CPU.Seg[cpu.CS] == 0x4567 {
						break
					}
					w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
					if p := Plan(w, budget); p != nil {
						if p.Algorithmic {
							native++
						}
						if w.IP() == 0x3418 && p.Algorithmic {
							used++
						}
						if err := b.CommitHostBlock(p); err != nil {
							t.Fatal(err)
						}
					} else if err := b.Step(); err != nil {
						t.Fatal(err)
					}
				}
				if native == 0 || used == 0 && !sub {
					t.Fatal("production selector missing Go region path", dos, budget, sub, used, native)
				}
				var sa, sb bytes.Buffer
				if err := a.SaveState(&sa); err != nil {
					t.Fatal(err)
				}
				if err := b.SaveState(&sb); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
					t.Fatal("production bounded region differs", dos, budget, sub)
				}
				rows += used
				cases++
			}
		}
	}
	t.Log("production bounded region whole comparisons", cases, "row plans", rows)
}
