// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestFlip5886Whole(t *testing.T) {
	cases, plans := 0, 0
	for _, width := range []uint16{1, 2, 3, 8, 9, 17, 160, 300} {
		for _, height := range []uint16{1, 2, 3, 7, 8} {
			for _, h := range []bool{false, true} {
				for _, ver := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						v := flip5886Seed(t, width, height, h, ver)
						v.dos = dos
						a, b := callMachine(v), callMachine(v)
						runCall(t, a, v, false)
						for b.Steps < 1000000 {
							if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
								break
							}
							w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
							p := flip5886(w, 10000)
							if p == nil {
								t.Fatalf("flip whole CPU w%d h%d horizontal%v vertical%v dos%v IP%x", width, height, h, ver, dos, w.IP())
							}
							if err := b.CommitHostBlock(p); err != nil {
								t.Fatal(err)
							}
							plans++
						}
						var sa, sb bytes.Buffer
						if err := a.SaveState(&sa); err != nil {
							t.Fatal(err)
						}
						if err := b.SaveState(&sb); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("flip SaveState differs w%d h%d horizontal%v vertical%v dos%v", width, height, h, ver, dos)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("flip5886 whole Go machine comparisons", cases, "algorithmic plans", plans)
}
