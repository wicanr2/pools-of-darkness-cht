package drawblocks

import (
	"bytes"
	"testing"
)

func TestNativePointCommit(t *testing.T) {
	cases := 0
	for _, read := range []bool{false, true} {
		for _, xy := range [][2]uint16{{0, 0}, {17, 99}, {319, 199}, {0xffff, 0}} {
			for _, dos := range []bool{false, true} {
				v := pointSeed(t, read, xy[0], xy[1], 320)
				v.dos = dos
				p := Plan(v, 200)
				if p == nil || !p.Algorithmic {
					t.Fatal("point did not select native geometry")
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
					t.Fatal("point complete machine state differs", read, xy, dos)
				}
				cases++
			}
		}
	}
	t.Log("point committed machine original comparisons", cases)
}
