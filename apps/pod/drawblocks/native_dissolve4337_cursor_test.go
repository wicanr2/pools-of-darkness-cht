// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337Cursors(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}} {
		v := dissolve4337Seed(t, dims[0], dims[1])
		m := original(v)
		sample := 0
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.IP == 0x456c {
				sample++
				if sample < 20 || sample%17 == 0 {
					for _, dos := range []bool{false, true} {
						w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
						p := planDissolve4337Cursor(w, 1000)
						if p == nil {
							t.Fatal("dissolve cursor refused", dims, sample, dos)
						}
						t.Run("cursor", func(t *testing.T) { compare(t, w, p) })
						if t.Failed() {
							t.Fatal("dissolve cursor differs", dims, sample, dos)
						}
						cases++
					}
				}
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("dissolve4337 cursor comparisons", cases)
}
