package drawblocks

import "testing"

func TestNativeEllipsePrefixes(t *testing.T) {
	cases := 0
	for _, shape := range [][4]int32{{160, 100, 1, 2}, {160, 100, 2, 5}, {160, 100, 10, 20}, {50, 40, 40, 20}, {0, 0, 20, 10}, {319, 199, 20, 10}} {
		for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}} {
			for _, dos := range []bool{false, true} {
				v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
				v.dos = dos
				p := planEllipsePrefix(v)
				if p == nil {
					t.Fatal("ellipse prefix refused")
				}
				t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
				if t.Failed() {
					t.Fatal("ellipse prefix differs", shape, box, dos, p.Instructions, p.Cycles)
				}
				cases++
			}
		}
	}
	t.Log("ellipse prefix original comparisons", cases)
}
