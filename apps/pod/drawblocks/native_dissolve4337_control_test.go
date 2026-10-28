// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeDissolve4337PixelsControls(t *testing.T) {
	pixels, controls := 0, 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}, {17, 9}, {20, 10}} {
		v := dissolve4337Seed(t, dims[0], dims[1])
		m := original(v)
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			ip := m.CPU.IP
			if ip == 0x4569 || ip == 0x45bb || ip == 0x45c4 || ip == 0x45ce {
				for _, dos := range []bool{false, true} {
					w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
					p := planDissolve4337Pixel(w, 1)
					if ip == 0x45c4 || ip == 0x45ce {
						p = planDissolve4337Control(w, 4)
						controls++
					} else {
						pixels++
					}
					if p == nil {
						t.Fatalf("dissolve handoff refused dims%v dos%v IP%x ES%x DS%x", dims, dos, ip, w.ES(), w.DS())
					}
					t.Run("handoff", func(t *testing.T) { compare(t, w, p) })
					if t.Failed() {
						t.Fatal("dissolve handoff differs", dims, dos, ip)
					}
				}
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Log("dissolve4337 single-step pixel comparisons", pixels, "control plans", controls)
}
