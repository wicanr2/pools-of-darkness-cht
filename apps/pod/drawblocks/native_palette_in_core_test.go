package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func fadeInSeed(t *testing.T, entry, count uint16, dos bool) *view {
	return &view{mem: input(t), dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, count, 23, 0xa500, 0xe000, 0xd000, 0, 0}, Seg: [4]uint16{0x48a1, 0x48a1, 0x8000, 0x48a1}, IP: entry, Flags: 0xfad7}}
}
func TestFadeInDifference(t *testing.T) {
	cases, plans := 0, 0
	for _, count := range []uint16{1, 3, 17, 768} {
		for _, maximum := range []byte{0, 31, 63} {
			for _, dos := range []bool{false, true} {
				v := fadeInSeed(t, 0x4892, count, dos)
				v.s.R[3] |= uint16(maximum)
				for i := uint32(0); i < uint32(count); i++ {
					v.mem[0x48a10+0x1b72+i] = byte((i*13 + 3) % 64)
					v.mem[0x48a10+0x1872+i] = byte((i*7 + 61) % 64)
				}
				for _, budget := range []uint64{10, 11, 13, 32, 4096, 12000} {
					p := planFadeDifferenceLayout(v, budget, fadeInRGBLayout)
					if p != nil {
						compare(t, v, p)
						plans++
					}
				}
				cases++
			}
		}
	}
	if plans == 0 {
		t.Fatal("no difference compared")
	}
	t.Logf("fade-in RGB difference cases=%d original plans=%d", cases, plans)
}
func TestFadeInBlend(t *testing.T) {
	cases, plans := 0, 0
	for _, count := range []uint16{1, 3, 17, 768} {
		for _, maximum := range []byte{0, 1, 31, 63} {
			for _, dos := range []bool{false, true} {
				v := fadeInSeed(t, 0x48e5, count, dos)
				v.mem[0x48a10+0x1173] = maximum
				for i := uint32(0); i < uint32(count); i++ {
					v.mem[0x48a10+0x1e72+i] = maximum / 2
					v.mem[0x48a10+0x2172+i] = byte((i*13 + 7) % 64)
					direction, color := byte(1), byte(10)
					if i%2 != 0 {
						direction, color = 255, 40
					}
					v.mem[0x48a10+0x2472+i] = direction
					v.mem[0x48a10+0x1b72+i] = color
				}
				for _, budget := range []uint64{7, 10, 14, 32, 4096, 12000} {
					p := planFadeBlendLayout(v, budget, fadeInRGBLayout)
					if p != nil {
						compare(t, v, p)
						plans++
					}
				}
				cases++
			}
		}
	}
	if plans == 0 {
		t.Fatal("no blend compared")
	}
	t.Logf("fade-in RGB blend cases=%d original plans=%d", cases, plans)
}
