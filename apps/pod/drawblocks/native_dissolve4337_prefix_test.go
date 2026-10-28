// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"testing"
)

func TestNativeDissolve4337Prefixes(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{1, 1}, {1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}, {160, 100}, {300, 180}} {
		for _, size := range []uint16{0, 1, 2} {
			for _, dos := range []bool{false, true} {
				v := dissolve4337Seed(t, dims[0], dims[1])
				v.dos = dos
				if size == 1 {
					putRowWord(v, 0x48a10+0x8bd+2, 18+dims[0])
					putRowWord(v, 0x48a10+0xabd+2, 8+dims[1])
				}
				if size == 2 && dims[0] > 1 && dims[1] > 1 {
					putRowWord(v, 0x48a10+0x8bd+2, 18+dims[0]-2)
					putRowWord(v, 0x48a10+0xabd+2, 8+dims[1]-2)
				}
				p := planDissolve4337Prefix(v)
				if p == nil {
					t.Fatal("dissolve prefix refused", dims, size, dos)
				}
				t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
				if t.Failed() {
					t.Fatal("dissolve prefix differs", dims, size, dos)
				}
				cases++
			}
		}
	}
	t.Log("dissolve4337 prefix comparisons", cases)
}
