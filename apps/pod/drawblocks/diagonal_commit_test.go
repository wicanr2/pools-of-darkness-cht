package drawblocks

import (
	"bytes"
	"testing"
)

func TestNativeDiagonalCommit(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{0, 0, 1, 2}, {1, 2, 0, 0}, {20, 10, 30, 30}, {30, 10, 20, 30}, {0, 0, 1, 1}, {1, 1, 0, 0}, {0, 0, 159, 10}, {159, 20, 0, 30}} {
		for _, stride := range []uint16{160, 320} {
			for _, dos := range []bool{false, true} {
				v := axisSeed(t, xy, stride)
				v.dos = dos
				p := Plan(v, 10000)
				if p == nil || !p.Algorithmic {
					t.Fatal("diagonal did not select Go geometry")
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
					t.Fatal("diagonal complete machine differs", xy, stride, dos)
				}
				c := callMachine(v)
				c.RunLimit = p.Instructions - 1
				var before, after bytes.Buffer
				if err := c.SaveState(&before); err != nil {
					t.Fatal(err)
				}
				if c.CommitHostBlock(p) == nil {
					t.Fatal("diagonal crossed event limit")
				}
				if err := c.SaveState(&after); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(before.Bytes(), after.Bytes()) {
					t.Fatal("diagonal refusal mutated state")
				}
				cases++
			}
		}
	}
	t.Log("diagonal committed machine original comparisons", cases)
}
