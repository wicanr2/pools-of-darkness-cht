// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"testing"
)

func TestScale5615Prefixes(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{2, 1}, {3, 2}, {9, 7}, {17, 9}, {40, 20}} {
		for _, factor := range []uint16{1, 2, 3, 5, 8} {
			for _, center := range [][2]int16{{25, 15}, {20, 10}, {29, 19}, {-5, -3}, {70, 80}} {
				for _, dos := range []bool{false, true} {
					v := scale5615Seed(t, dims[0], dims[1], factor, center[0], center[1])
					v.dos = dos
					p := planScale5615Prefix(v)
					if p == nil {
						t.Fatal("scale prefix refused")
					}
					t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
					if t.Failed() {
						t.Fatalf("scale prefix differs dims%v factor%d center%v dos%v", dims, factor, center, dos)
					}
					cases++
				}
			}
		}
	}
	t.Log("scale5615 prefix comparisons", cases)
}
