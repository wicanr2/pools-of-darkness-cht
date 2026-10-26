// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestScale5615Outputs(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{2, 1}, {3, 2}, {9, 7}, {17, 9}, {40, 20}} {
		for _, factor := range []uint16{1, 2, 3, 5, 8} {
			for _, center := range [][2]int16{{25, 15}, {20, 10}, {29, 19}, {-5, -3}, {70, 80}} {
				v := scale5615Seed(t, dims[0], dims[1], factor, center[0], center[1])
				m := original(v)
				for m.Steps < 1000000 {
					if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
						break
					}
					if m.CPU.IP == 0x581a {
						for _, dos := range []bool{false, true} {
							w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							p := planScale5615Output(w, 10000)
							if p == nil {
								t.Fatal("scale output refused", dims, factor, center)
							}
							t.Run("output", func(t *testing.T) { compare(t, w, p) })
							if t.Failed() {
								t.Fatal("scale output differs", dims, factor, center, dos)
							}
							cases++
						}
					}
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
	t.Log("scale5615 output group comparisons", cases)
}
