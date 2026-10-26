// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestScale5615Boundaries(t *testing.T) {
	cases, plans := 0, 0
	for _, box := range [][4]uint16{{0, 0, 319, 1}, {318, 198, 319, 199}, {0, 199, 8, 199}, {303, 0, 319, 3}} {
		for _, factor := range []uint16{1, 3, 8} {
			for _, center := range [][2]int16{{-160, -100}, {25, 15}, {479, 299}} {
				for _, dos := range []bool{false, true} {
					for _, parents := range []bool{false, true} {
						v := scale5615Seed(t, box[2]-box[0]+1, box[3]-box[1]+1, factor, center[0], center[1])
						v.dos = dos
						for i, off := range []uint16{0x4bd, 0x6bd, 0x8bd, 0xabd} {
							putRowWord(v, 0x48a10+uint32(off+2), box[i])
						}
						if parents {
							for slot := uint16(2); slot <= 9; slot++ {
								seg, flags := uint16(0xa000), uint16(0x80)
								if slot < 9 {
									seg, flags = (slot+1)*2, 0xc0
								}
								for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, 319}, {0xabd, 199}, {0xcbd, flags}} {
									putRowWord(v, 0x48a10+uint32(w[0]+slot*2), w[1])
								}
							}
						}
						a, b := callMachine(v), callMachine(v)
						runCall(t, a, v, false)
						for b.Steps < 1000000 {
							if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
								break
							}
							w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
							p := scale5615(w, 10000)
							if p == nil {
								t.Fatalf("boundary scale CPU box%v factor%d center%v dos%v parents%v IP%x", box, factor, center, dos, parents, w.IP())
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
							t.Fatalf("boundary scale SaveState differs box%v factor%d center%v dos%v parents%v", box, factor, center, dos, parents)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("scale5615 boundary parent whole comparisons", cases, "algorithmic plans", plans)
}
