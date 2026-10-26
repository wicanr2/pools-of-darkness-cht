// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

type scalePixel struct {
	at    uint32
	value byte
}

func scale5615Seed(t *testing.T, width, height, factor uint16, cx, cy int16) *view {
	v := copySurfaceSeed(t, width, height, false, false)
	v.s.IP = 0x5615
	stack := uint32(v.SS())*16 + uint32(v.SP())
	for _, w := range [][2]uint16{{4, 0}, {6, 1}, {8, uint16(cx)}, {10, uint16(cy)}, {12, 0x7b}, {14, factor}} {
		putRowWord(v, stack+uint32(w[0]), w[1])
	}
	// Source 20,10..29,19. Destination 18,8 with a separately sized frame.
	for _, w := range [][2]uint16{{0x4bd, 20}, {0x6bd, 10}, {0x8bd, 29}, {0xabd, 19}} {
		putRowWord(v, 0x48a10+uint32(w[0]), w[1])
	}
	v.s.Flags |= cpu.DF
	return v
}
func scale5615Pixels(v *view, width, height, factor uint16, cx, cy int16) []scalePixel {
	canvas := append([]byte(nil), v.mem...)
	out := []scalePixel{}
	x0 := int32(cx) - int32((width-1)/2/factor)
	y0 := int32(cy) - int32((height-1)/2/factor)
	for row := uint16(0); row < height; {
		data := make([]byte, width)
		sy := y0 + int32(row/factor)
		for col := range data {
			sx := x0 + int32(uint16(col)/factor)
			value := byte(0x7b)
			if sx >= 20 && sx <= 29 && sy >= 10 && sy <= 19 {
				value = canvas[0xa0000+uint32(sy*320+sx)]
			}
			data[col] = value
		}
		for rep := uint16(0); rep < factor && row < height; rep++ {
			for col, value := range data {
				at := uint32(0xa0000) + uint32((8+row)*320+18) + uint32(col)
				out = append(out, scalePixel{at, value})
				canvas[at] = value
			}
			row++
		}
	}
	return out
}
func TestScale5615Pixels(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{2, 1}, {3, 2}, {9, 7}, {17, 9}, {40, 20}} {
		for _, factor := range []uint16{1, 2, 3, 5, 8} {
			for _, center := range [][2]int16{{25, 15}, {20, 10}, {29, 19}, {-5, -3}, {70, 80}} {
				v := scale5615Seed(t, dims[0], dims[1], factor, center[0], center[1])
				m := original(v)
				got := []scalePixel{}
				m.WatchWrite(0xa0000, 0xb0000, func(m *machine.Machine, at uint32, old, value byte) { got = append(got, scalePixel{at, value}) })
				for m.Steps < 1000000 {
					if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
						break
					}
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
				want := scale5615Pixels(v, dims[0], dims[1], factor, center[0], center[1])
				if !reflect.DeepEqual(got, want) {
					index := 0
					for index < len(got) && index < len(want) && got[index] == want[index] {
						index++
					}
					t.Fatalf("scale pixel difference dims%v factor%d center%v got%d want%d first%d", dims, factor, center, len(got), len(want), index)
				}
				cases++
			}
		}
	}
	t.Log("scale5615 ordered pixel comparisons", cases)
}
