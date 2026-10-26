// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestFlip5886Rows(t *testing.T) {
	rows, addresses := 0, 0
	for _, width := range []uint16{1, 2, 3, 8, 9, 17, 160, 300} {
		for _, height := range []uint16{1, 2, 3, 7, 8} {
			for _, h := range []bool{false, true} {
				for _, ver := range []bool{false, true} {
					if !h && !ver {
						continue
					}
					v := flip5886Seed(t, width, height, h, ver)
					m := original(v)
					for m.Steps < 1000000 {
						if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
							break
						}
						ip := m.CPU.IP
						if ip == 0x59a9 || ip == 0x5947 {
							for _, dos := range []bool{false, true} {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
								var p *machine.HostWorkPlan
								if ip == 0x59a9 {
									p = planFlip5886Rows(w, 10000)
									rows++
								} else {
									p = planFlip5886Address(w, 10000)
									addresses++
								}
								if p == nil {
									t.Fatalf("flip row refused w%d h%d horizontal%v vertical%v dos%v IP%x", width, height, h, ver, dos, ip)
								}
								t.Run("row", func(t *testing.T) { compare(t, w, p) })
								if t.Failed() {
									t.Fatal("flip row differs", width, height, h, ver, dos, ip)
								}
							}
						}
						if err := m.Step(); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		}
	}
	t.Log("flip5886 paired row comparisons", rows, "address plans", addresses)
}
