// Spec025/#49: source-driven Go variable-width glyph pixels.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeGlyph5AA4Spans(t *testing.T) {
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
					if m.CPU.Seg[cpu.CS] == 0x48a1 && (m.CPU.IP == 0x5c02 || m.CPU.IP == 0x5c2c) {
						seen++
						for _, dos := range []bool{false, true} {
							v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							for _, budget := range []uint64{1, 17} {
								p := planGlyph5aa4Spans(v, budget)
								if p == nil {
									t.Fatalf("glyph span refused width%d height%d mapped%v IP%x regs%+v", width, height, mapped, v.IP(), v.s)
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
				if seen == 0 && !(mapped && width*height == 1) {
					t.Fatal("no pixel comparisons")
				}
				cases++
			}
		}
	}
	t.Logf("glyph source-driven single-step spans cases=%d original comparisons=%d", cases, plans)
}
