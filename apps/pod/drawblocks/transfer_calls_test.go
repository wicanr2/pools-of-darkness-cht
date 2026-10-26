package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestSmallCallsIndependentBoundaries(t *testing.T) {
	mem := input(t)
	cases := 0
	for _, entry := range []struct{ cs, ip uint16 }{{0x1ed4, 0x1a5c}, {0x1ed4, 0x1a7f}, {0x1d1f, 0x46e}} {
		for _, length := range []uint16{0, 1, 2, 3, 31, 64, 255, 256} {
			for _, source := range []uint16{8, 128} {
				for _, dest := range []uint16{32, 512} {
					for _, dos := range []bool{false, true} {
						v := &view{mem: mem, dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}, Seg: [4]uint16{0x6000, entry.cs, 0x8000, 0x6000}, IP: entry.ip, Flags: 0xfed7}}
						word := func(off, val uint16) {
							at := 0x80000 + int(v.SP()+off)
							mem[at] = byte(val)
							mem[at+1] = byte(val >> 8)
						}
						word(0, 0x1234)
						word(2, 0x1111)
						switch entry.ip {
						case 0x1a5c:
							word(4, length)
							word(6, dest)
							word(8, 0x6000)
							word(10, source)
							word(12, 0x6000)
						case 0x1a7f:
							word(4, 0xcd55)
							word(6, length)
							word(8, dest)
							word(10, 0x6000)
						case 0x46e:
							word(4, length)
							word(6, dest)
							word(8, 0x6000)
						}
						p := planSmallCall(v, 4096)
						if p == nil {
							t.Fatal("valid full call refused", entry, length, source, dest)
						}
						compare(t, v, p)
						cases++
					}
				}
			}
		}
	}
	t.Log("full ABI, zero/odd/even lengths, overlap and both directions/costs", cases)
}
