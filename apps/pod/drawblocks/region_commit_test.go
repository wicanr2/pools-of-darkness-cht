package drawblocks

import (
	"bytes"
	"testing"
)

func TestNativeRegionCommit(t *testing.T) {
	cases := 0
	for _, xy := range [][4]uint16{{0, 0, 0, 0}, {17, 19, 20, 22}, {0xffff, 0xffff, 320, 200}, {0, 200, 1, 201}} {
		for _, sub := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				v := regionSeed(t, xy, 320, sub)
				v.dos = dos
				p := Plan(v, 10000)
				if p == nil || !p.Algorithmic {
					t.Fatal("region did not select Go spans")
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
				var originalState, nativeState bytes.Buffer
				if err := a.SaveState(&originalState); err != nil {
					t.Fatal(err)
				}
				if err := b.SaveState(&nativeState); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(originalState.Bytes(), nativeState.Bytes()) {
					t.Fatal("region complete state differs", xy, sub, dos)
				}
				c := callMachine(v)
				c.RunLimit = p.Instructions - 1
				var before, after bytes.Buffer
				if err := c.SaveState(&before); err != nil {
					t.Fatal(err)
				}
				if c.CommitHostBlock(p) == nil {
					t.Fatal("region crossed run boundary")
				}
				if err := c.SaveState(&after); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(before.Bytes(), after.Bytes()) {
					t.Fatal("event refusal mutated region state")
				}
				cases++
			}
		}
	}
	t.Log("region committed machine original comparisons", cases)
}
