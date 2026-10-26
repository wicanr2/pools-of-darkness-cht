// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestBitmap51PlanarPrefixes(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 7, 8, 9, 16, 17, 32, 160, 300} {
		for _, rows := range []uint16{1, 3} {
			for _, compressed := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					v := bitmap51RowSeed(t, width, rows, true, compressed)
					v.dos = dos
					p := planBitmap51PlanarPrefix(v, 1000)
					if p == nil {
						t.Fatal("planar prefix refused")
					}
					compare(t, v, p)
					cases++
				}
			}
		}
	}
	t.Log("bitmap51 planar prefix comparisons", cases)
}
func TestBitmap51PlanarGroups(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 7, 8, 9, 16, 17, 32, 160, 300} {
		for _, compressed := range []bool{false, true} {
			v := bitmap51RowSeed(t, width, 1, true, compressed)
			m := original(v)
			lastDone := uint16(65535)
			for m.Steps < 100000 {
				if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
					break
				}
				if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x53c5 {
					done := width - uint16(m.Mem[0x48a10+0x11fd]) - uint16(m.Mem[0x48a10+0x11fe])<<8
					if done == 0 || done == width/2 || done == width-1 || done/8 != lastDone/8 {
						for _, dos := range []bool{false, true} {
							w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
							p := planBitmap51PlanarGroup(w, 10000)
							if p == nil {
								t.Fatal("planar group refused", width, compressed, done, dos)
							}
							t.Run("group", func(t *testing.T) { compare(t, w, p) })
							if t.Failed() {
								t.Fatal("planar group differs", width, compressed, done, dos, p.Instructions, p.Cycles)
							}
							cases++
						}
						lastDone = done
					}
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	t.Log("bitmap51 planar group comparisons", cases)
}
