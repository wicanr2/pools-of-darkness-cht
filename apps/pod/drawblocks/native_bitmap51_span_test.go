// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestBitmap51SpansAdvance(t *testing.T) {
	spans, advances := 0, 0
	for _, width := range []uint16{1, 2, 7, 8, 9, 16, 17, 32, 160, 300} {
		for _, rows := range []uint16{1, 3} {
			for _, planar := range []bool{false, true} {
				for _, compressed := range []bool{false, true} {
					v := bitmap51RowSeed(t, width, rows, planar, compressed)
					m := original(v)
					for m.Steps < 100000 {
						if m.Steps > 0 && (m.CPU.Seg[cpu.CS] != 0x48a1 || m.CPU.IP == 0x5313) {
							break
						}
						ip := m.CPU.IP
						if ip == 0x53d8 || ip == 0x53f6 || ip == 0x53fa || ip == 0x53fc {
							for _, dos := range []bool{false, true} {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
								p := planBitmap51Span(w, 1)
								if ip == 0x53fc {
									p = planBitmap51Advance(w, 10)
									advances++
								} else {
									spans++
								}
								if p == nil {
									t.Fatalf("row handoff refused width%d rows%d planar%v compressed%v IP%x", width, rows, planar, compressed, ip)
								}
								compare(t, w, p)
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
	t.Log("bitmap51 single-step span comparisons", spans, "row advances", advances)
}
