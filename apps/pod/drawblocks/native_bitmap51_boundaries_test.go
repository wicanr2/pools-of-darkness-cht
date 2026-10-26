// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestBitmap51Boundaries(t *testing.T) {
	cases := 0
	for _, box := range [][4]uint16{{0, 0, 319, 1}, {319, 199, 319, 199}, {0, 199, 8, 199}, {303, 0, 319, 2}} {
		for _, planar := range []bool{false, true} {
			for _, compressed := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					for _, df := range []bool{false, true} {
						v := bitmap51CallSeed(t, 320, 3, planar, compressed)
						v.dos = dos
						if df {
							v.s.Flags |= cpu.DF
						}
						for i, off := range []uint16{0x4bd, 0x6bd, 0x8bd, 0xabd} {
							putRowWord(v, 0x48a10+uint32(off), box[i])
						}
						a, b := callMachine(v), callMachine(v)
						runCall(t, a, v, false)
						for b.Steps < 1000000 {
							if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
								break
							}
							w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
							if p := Plan(w, 1024); p != nil {
								if !p.Algorithmic && bitmap51FixedRaster(p.Writes, p.Transfers) {
									t.Fatal("boundary fixed raster")
								}
								if err := b.CommitHostBlock(p); err != nil {
									t.Fatal(err)
								}
							} else {
								if w.CS() == 0x48a1 && (w.IP() == 0x53d8 || w.IP() == 0x53f6 || w.IP() == 0x53fa) {
									t.Fatalf("boundary CPU raster box%v planar%v compressed%v dos%v df%v IP%x", box, planar, compressed, dos, df, w.IP())
								}
								if err := b.Step(); err != nil {
									t.Fatal(err)
								}
							}
						}
						var sa, sb bytes.Buffer
						if err := a.SaveState(&sa); err != nil {
							t.Fatal(err)
						}
						if err := b.SaveState(&sb); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
							t.Fatalf("bitmap boundary differs box%v planar%v compressed%v dos%v df%v", box, planar, compressed, dos, df)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("bitmap51 boundary full machine comparisons", cases)
}
