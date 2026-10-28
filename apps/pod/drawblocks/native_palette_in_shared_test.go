package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestFadeInSharedJobs(t *testing.T) {
	scan, list, target, errors := 0, 0, 0, 0
	for _, dos := range []bool{false, true} {
		for _, n := range []uint16{1, 2, 17, 320} {
			v := fadeInSeed(t, 0x47ba, n, dos)
			v.s.Seg[cpu.DS] = 0xa000
			v.s.R[cpu.SI] = 3200
			v.s.R[cpu.BX] = 0
			for i := uint32(0); i < uint32(n); i++ {
				v.mem[0xa0000+3200+i] = byte(i * 7)
			}
			for _, budget := range []uint64{4, 8, 37, 4096} {
				p := planPaletteScan(v, budget)
				if p == nil {
					t.Fatal("scan refused")
				}
				compare(t, v, p)
				scan++
			}
		}
		for _, n := range []uint16{0, 1, 17, 255} {
			v := fadeInSeed(t, 0x47d6, 1, dos)
			v.s.R[cpu.BX] = n
			v.s.R[cpu.DI] = 0x1572
			for i := uint32(0); i < 256; i++ {
				v.mem[0x48a10+0x1772+i] = byte(i % 3)
			}
			for _, budget := range []uint64{7, 17, 4096} {
				p := planPaletteList(v, budget)
				if p == nil {
					t.Fatal("list refused")
				}
				compare(t, v, p)
				list++
			}
		}
		for _, n := range []uint16{1, 2, 17, 256} {
			v := fadeInSeed(t, 0x4867, n, dos)
			v.s.R[cpu.DI] = 0x1b72
			v.mem[0x48a10+0x1179] = 3
			v.mem[0x48a10+0x117b] = 31
			v.mem[0x48a10+0x117d] = 63
			for _, budget := range []uint64{7, 17, 4096} {
				p := planPaletteTarget(v, budget)
				if p == nil {
					t.Fatal("target refused")
				}
				compare(t, v, p)
				target++
			}
		}
		for _, n := range []uint16{1, 2, 17, 256} {
			for _, maximum := range []uint16{0, 1, 31, 63} {
				v := fadeInSeed(t, 0x48b3, 1, dos)
				v.s.R[cpu.BX] = maximum
				putRowWord(v, 0x48a10+0x116f, n)
				p := planPaletteError(v, 9)
				if p == nil {
					t.Fatal("errors refused")
				}
				compare(t, v, p)
				errors++
			}
		}
	}
	t.Logf("fade-in shared scan=%d list=%d target=%d error=%d original plans", scan, list, target, errors)
}
