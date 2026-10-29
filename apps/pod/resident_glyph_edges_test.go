package pod

import (
	"encoding/binary"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestResidentPixelEdges(t *testing.T) {
	base := glyphViews(t)[0]
	cases := 0
	for _, pos := range [][2]uint16{{0, 0}, {39, 0}, {0, 24}, {39, 24}} {
		for _, fill := range []byte{0, 255} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 7, 1024} {
					v := *base
					v.memory = append([]byte(nil), base.memory...)
					v.dos = dos
					stack := uint32(v.SS())*16 + uint32(v.SP())
					table := uint32(v.DS())*16 + 0x9d86
					v.memory[stack+4] = 0
					v.memory[stack+6] = 255
					v.memory[stack+8] = 128
					binary.LittleEndian.PutUint16(v.memory[stack+12:], pos[0])
					binary.LittleEndian.PutUint16(v.memory[stack+10:], pos[1])
					src := uint32(v.Read16(table+2))*16 + uint32(v.Read16(table))
					for i := 0; i < 8; i++ {
						v.memory[src+uint32(i)] = fill
					}
					a, b := glyphMachine(&v, cpu.Model80386), glyphMachine(&v, cpu.Model80386)
					ip, cs := v.Read16(stack), v.Read16(stack+2)
					for n := 0; n < 1000; n++ {
						if n > 0 && a.CPU.IP == ip && a.CPU.Seg[cpu.CS] == cs {
							break
						}
						if e := a.Step(); e != nil {
							t.Fatal(e)
						}
					}
					returned := false
					for n := 0; n < 1000; n++ {
						if n > 0 && b.CPU.IP == ip && b.CPU.Seg[cpu.CS] == cs {
							returned = true
							break
						}
						w := residentViewAt(b, dos)
						p := residentGlyphPlan(w, budget)
						if p == nil {
							p = PlanGlyphCall(w, budget)
						}
						if p != nil {
							if e := b.CommitHostWork(p); e != nil {
								t.Fatal(e)
							}
						} else {
							if w.CS() == 0x1b2e && (w.IP() == 0x11ea || w.IP() == 0x11f1) {
								t.Fatal("CPU edge glyph")
							}
							if e := b.Step(); e != nil {
								t.Fatal(e)
							}
						}
					}
					if !returned {
						t.Fatal("edge glyph did not return")
					}
					residentEqual(t, a, b)
					cases++
				}
			}
		}
	}
	t.Logf("resident glyph corner/full-background/full-foreground SaveState+PortLog cases=%d CPU resident pixels=0", cases)
}
