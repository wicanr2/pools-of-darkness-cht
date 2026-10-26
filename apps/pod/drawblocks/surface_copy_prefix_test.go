package drawblocks

import (
	"crypto/sha256"
	"testing"
)

func TestNativeCopyPrefixes(t *testing.T) {
	seed := copySurfaceSeed(t, 1, 1, false, false)
	var code [593]byte
	copyView(seed, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases := 0
	for _, width := range []uint16{1, 2, 3, 7, 8, 9, 20} {
		for _, height := range []uint16{1, 2, 5} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						v := copySurfaceSeed(t, width, height, bx, by)
						v.dos = dos
						p := planSurfaceCopyPrefix(v)
						if p == nil {
							t.Fatal("copy prefix refused")
						}
						t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
						if t.Failed() {
							t.Fatal("copy prefix mismatch", width, height, bx, by, dos, p.Instructions, p.Cycles)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("copy prefix independent comparisons", cases)
}
