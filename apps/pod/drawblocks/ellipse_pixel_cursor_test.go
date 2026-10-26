package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeEllipsePixelCursors(t *testing.T) {
	cases := 0
	for _, shape := range [][4]int32{{160, 100, 1, 2}, {160, 100, 2, 5}, {160, 100, 10, 20}, {50, 40, 40, 20}, {0, 0, 20, 10}, {319, 199, 20, 10}} {
		for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}} {
			v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
			m := original(v)
			for m.Steps < 100000 {
				if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
					break
				}
				if _, pixel := ellipsePixelStage(m.CPU.IP); pixel && m.CPU.Seg[cpu.CS] == 0x48a1 {
					for _, dos := range []bool{false, true} {
						w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
						p := planEllipsePixel(w, 1)
						if p == nil {
							t.Fatalf("ellipse pixel refused shape%v box%v IP%x", shape, box, w.IP())
						}
						compare(t, w, p)
						cases++
					}
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	t.Log("ellipse one-step Go pixel comparisons", cases)
}
