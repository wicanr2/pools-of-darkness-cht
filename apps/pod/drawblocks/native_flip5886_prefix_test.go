// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"testing"
)

func TestFlip5886Prefixes(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 3, 8, 9, 17, 160, 300} {
		for _, height := range []uint16{1, 2, 3, 7, 8} {
			for _, h := range []bool{false, true} {
				for _, ver := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						v := flip5886Seed(t, width, height, h, ver)
						v.dos = dos
						p := planFlip5886Prefix(v)
						if p == nil {
							t.Fatal("flip prefix refused")
						}
						t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
						if t.Failed() {
							t.Fatalf("flip prefix differs w%d h%d horizontal%v vertical%v dos%v", width, height, h, ver, dos)
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("flip5886 prefix comparisons", cases)
}
