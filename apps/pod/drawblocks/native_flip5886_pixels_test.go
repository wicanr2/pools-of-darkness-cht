// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

type flip5886Pixel struct {
	at    uint32
	value byte
}

func flip5886Seed(t *testing.T, width, height uint16, horizontal, vertical bool) *view {
	v := copySurfaceSeed(t, width, height, false, false)
	v.s.IP = 0x5886
	stack := uint32(v.SS())*16 + uint32(v.SP())
	h, ver := uint16(0), uint16(0)
	if horizontal {
		h = 1
	}
	if vertical {
		ver = 1
	}
	putRowWord(v, stack+4, 0)
	putRowWord(v, stack+6, h)
	putRowWord(v, stack+8, ver)
	v.s.Flags |= cpu.DF
	return v
}
func flip5886Pixels(v *view, width, height uint16, horizontal, vertical bool) []flip5886Pixel {
	canvas := append([]byte(nil), v.mem...)
	out := []flip5886Pixel{}
	if !horizontal && !vertical {
		return out
	}
	rows := height
	if vertical {
		rows = height / 2
		if rows == 0 {
			rows = 1
		}
	}
	for row := uint16(0); row < rows; row++ {
		top, bottom := uint16(10)+row, uint16(10)+row
		if vertical {
			bottom = 10 + height - 1 - row
		}
		saved := make([]byte, width)
		for x := uint16(0); x < width; x++ {
			sx := x
			if horizontal {
				sx = width - 1 - x
			}
			saved[x] = canvas[0xa0000+uint32(top)*320+20+uint32(sx)]
		}
		// In the single-row case source and destination alias during the second
		// stage. Read each source byte after the preceding ordered write.
		for x := uint16(0); x < width; x++ {
			sx := x
			if horizontal {
				sx = width - 1 - x
			}
			value := canvas[0xa0000+uint32(bottom)*320+20+uint32(sx)]
			at := uint32(0xa0000) + uint32(top)*320 + 20 + uint32(x)
			out = append(out, flip5886Pixel{at, value})
			canvas[at] = value
		}
		for x, value := range saved {
			at := uint32(0xa0000) + uint32(bottom)*320 + 20 + uint32(x)
			out = append(out, flip5886Pixel{at, value})
			canvas[at] = value
		}
	}
	return out
}
func TestFlip5886Pixels(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 3, 8, 9, 17, 160, 300} {
		for _, height := range []uint16{1, 2, 3, 7, 8} {
			for _, h := range []bool{false, true} {
				for _, ver := range []bool{false, true} {
					v := flip5886Seed(t, width, height, h, ver)
					m := original(v)
					got := []flip5886Pixel{}
					m.WatchWrite(0xa0000, 0xb0000, func(m *machine.Machine, at uint32, old, value byte) { got = append(got, flip5886Pixel{at, value}) })
					for m.Steps < 1000000 {
						if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
							break
						}
						if err := m.Step(); err != nil {
							t.Fatal(err)
						}
					}
					want := flip5886Pixels(v, width, height, h, ver)
					if !reflect.DeepEqual(got, want) {
						index := 0
						for index < len(got) && index < len(want) && got[index] == want[index] {
							index++
						}
						t.Fatalf("flip pixels differ w%d h%d horizontal%v vertical%v got%d want%d first%d", width, height, h, ver, len(got), len(want), index)
					}
					cases++
				}
			}
		}
	}
	t.Log("flip5886 ordered pixel comparisons", cases)
}
