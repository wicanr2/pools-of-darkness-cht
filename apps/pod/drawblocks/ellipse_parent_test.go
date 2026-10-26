package drawblocks

import (
	"bytes"
	"testing"
)

func TestNativeEllipseParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, dos := range []bool{false, true} {
			v := ellipseSeed(t, 50, 40, 40, 20, [4]uint16{25, 15, 70, 150})
			v.dos = dos
			for i := uint16(1); i <= links; i++ {
				seg, flags := uint16(0xa000), uint16(0x80)
				if i < links {
					seg, flags = (i+1)*2, 0xc0
				}
				for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, 319}, {0xabd, 199}, {0xcbd, flags}} {
					putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
				}
			}
			p := planEllipsePrefix(v)
			if p == nil {
				t.Fatal("parent ellipse prefix refused")
			}
			compare(t, v, p)
			cases++
			a, b := callMachine(v), callMachine(v)
			runCall(t, a, v, false)
			runCall(t, b, v, true)
			var sa, sb bytes.Buffer
			if err := a.SaveState(&sa); err != nil {
				t.Fatal(err)
			}
			if err := b.SaveState(&sb); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
				t.Fatal("parent ellipse full machine differs", links, dos)
			}
			cases++
		}
	}
	t.Log("ellipse parent prefix and machine comparisons", cases)
}
