package drawblocks

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

func fadeColorSeed(t *testing.T, ip, used, remaining uint16) *view {
	v := fadeSeed(t, ip, 1)
	v.s.Seg[cpu.SS] = 0x7000
	v.s.R[cpu.SP], v.s.R[cpu.BP] = 0x8000, 0x8100
	if ip == 0x4bc1 {
		v.s.R[cpu.SP] -= 8
	}
	v.s.R[cpu.DI] = remaining * 2
	v.s.R[cpu.SI] = 0x1872 + (used-remaining)*3
	v.mem[0x48a10+0x116f], v.mem[0x48a10+0x1170] = byte(used), byte(used>>8)
	for i := uint16(0); i < used; i++ {
		color := byte(i * 37)
		v.mem[0x48a10+0x1572+int(i)*2], v.mem[0x48a10+0x1573+int(i)*2] = color, 0
		copy(v.mem[0x48a10+0x1872+int(i)*3:], []byte{12, 34, 56})
	}
	return v
}

func TestNativeFadeColorOriginal(t *testing.T) {
	plans := 0
	for _, ip := range []uint16{0x4b9a, 0x4bc1} {
		for _, used := range []uint16{1, 109, 256} {
			positions := []uint16{0, 1}
			if used > 1 {
				positions = append(positions, used)
			}
			for _, remaining := range positions {
				for _, dos := range []bool{false, true} {
					v := fadeColorSeed(t, ip, used, remaining)
					v.dos = dos
					p := planFadeColor(v, 1024)
					if p == nil {
						t.Fatal("valid palette color job rejected")
					}
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	t.Log("fade color original job comparisons", plans)
}

func TestNativeFadeColorRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "direction", "segment", "count", "index", "cursor", "stack", "alias", "rgb", "palette-index", "cleanup-wrap"} {
		t.Run(kind, func(t *testing.T) {
			v := fadeColorSeed(t, 0x4bc1, 109, 109)
			budget := uint64(1024)
			switch kind {
			case "budget":
				budget = 16
			case "code":
				v.mem[0x48a10+0x4b9a] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "segment":
				v.s.Seg[cpu.DS] = 0x5000
			case "count":
				v.mem[0x48a10+0x116f], v.mem[0x48a10+0x1170] = 0, 0
			case "index":
				v.s.R[cpu.DI] = 3
			case "cursor":
				v.s.R[cpu.SI]++
			case "stack":
				v.s.Seg[cpu.SS] = 0xa000
			case "alias":
				v.s.Seg[cpu.SS], v.s.R[cpu.SP] = 0x48a1, 0x1800
			case "rgb":
				v.mem[0x48a10+0x1872] = 64
			case "palette-index":
				v.mem[0x48a10+0x1572+108*2+1] = 1
			case "cleanup-wrap":
				v.s.R[cpu.SP] = 65530
			}
			if planFadeColor(v, budget) != nil {
				t.Fatal("unproved palette job accepted")
			}
		})
	}
}
