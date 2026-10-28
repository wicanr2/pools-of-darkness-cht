// Spec025/#49: source-driven Go variable-width glyph pixels.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeGlyph5AA4Rows(t *testing.T) {
	cases, plans := 0, 0
	for _, width := range []uint16{1, 2, 3, 7, 9} {
		for _, height := range []uint16{1, 2, 5} {
			for _, mapped := range []bool{false, true} {
				seed := glyph5aa4Seed(t, width, height, mapped)
				m := original(seed)
				seen := 0
				for m.Steps < 10000 {
					if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
						break
					}
					if m.CPU.Seg[cpu.CS] == 0x48a1 && (m.CPU.IP == 0x5c00 || m.CPU.IP == 0x5c26 || m.CPU.IP == 0x5c2f) {
						seen++
						for _, dos := range []bool{false, true} {
							v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							for _, budget := range []uint64{5, 6, 7, 8, 17, 1024} {
								p := planGlyph5aa4Rows(v, budget)
								if p == nil {
									if budget < 8 {
										continue
									}
									t.Fatalf("glyph row refused width%d height%d mapped%v IP%x regs%+v", width, height, mapped, v.IP(), v.s)
								}
								compare(t, v, p)
								plans++
							}
						}
					}
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
				if seen == 0 {
					t.Fatal("no row comparisons")
				}
				cases++
			}
		}
	}
	t.Logf("glyph source-driven row cursor cases=%d original comparisons=%d", cases, plans)
}
