package drawblocks

import (
	"bytes"
	"testing"
)

func TestNativeAxisCommit(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{0, 0, 0, 0}, {0xffff, 19, 320, 19}, {17, 0, 17, 8}, {17, 199, 17, 0}} {
		for _, stride := range []uint16{160, 320} {
			for _, dos := range []bool{false, true} {
				v := axisSeed(t, xy, stride)
				v.dos = dos
				p := Plan(v, 10000)
				if p == nil || !p.Algorithmic {
					t.Fatal("axis did not select Go geometry")
				}
				a, b := callMachine(v), callMachine(v)
				for i := uint64(0); i < p.Instructions; i++ {
					if err := a.Step(); err != nil {
						t.Fatal(err)
					}
				}
				if err := b.CommitHostBlock(p); err != nil {
					t.Fatal(err)
				}
				var sa, sb bytes.Buffer
				if err := a.SaveState(&sa); err != nil {
					t.Fatal(err)
				}
				if err := b.SaveState(&sb); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
					t.Fatal("axis complete machine differs", xy, stride, dos)
				}
				cases++
			}
		}
	}
	t.Log("axis committed machine original comparisons", cases)
}
