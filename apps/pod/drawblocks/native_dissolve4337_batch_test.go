// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337Batches(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}, {17, 9}, {20, 10}} {
		v := dissolve4337Seed(t, dims[0], dims[1])
		m := original(v)
		sample := 0
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.IP == 0x456c {
				sample++
				if sample < 5 || sample%47 == 0 {
					for _, dos := range []bool{false, true} {
						for _, budget := range []uint64{64, 128, 512, 4096} {
							w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							p := planDissolve4337Batch(w, budget)
							if p == nil {
								t.Fatal("dissolve batch refused", dims, sample, dos, budget)
							}
							if p.Instructions > budget {
								t.Fatal("dissolve batch exceeded budget")
							}
							t.Run("batch", func(t *testing.T) { compare(t, w, p) })
							if t.Failed() {
								t.Fatal("dissolve batch differs", dims, sample, dos, budget)
							}
							cases++
						}
					}
				}
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("dissolve4337 bounded batch comparisons", cases)
}
