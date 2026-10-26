// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestFlip5886Spans(t *testing.T) {
	cases := 0
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
						target := ip == 0x59c8 || ip == 0x59cf || ip == 0x59d3 || ip == 0x59f1 || ip == 0x59f8 || ip == 0x59fc || ip == 0x5a0c || ip == 0x5a10
						if target {
							for _, dos := range []bool{false, true} {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
								p := planFlip5886Span(w, 1)
								if p == nil {
									t.Fatalf("flip span refused w%d h%d horizontal%v vertical%v dos%v IP%x CX%d", width, height, h, ver, dos, ip, w.CX())
								}
								t.Run("span", func(t *testing.T) { compare(t, w, p) })
								if t.Failed() {
									t.Fatal("flip span differs", width, height, h, ver, dos, ip)
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
	}
	t.Log("flip5886 one-step span comparisons", cases)
}
