// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
	"reflect"
	"testing"
)

type dissolve4337Pixel struct {
	at    uint32
	value byte
}

func dissolve4337Seed(t *testing.T, width, height uint16) *view {
	v := copySurfaceSeed(t, width, height, false, false)
	v.s.IP = 0x4337
	return v
}
func dissolve4337Pixels(v *view, width, height uint16) []dissolve4337Pixel {
	canvas := append([]byte(nil), v.mem...)
	out := []dissolve4337Pixel{}
	emit := func(x, y uint16) {
		value := canvas[0xa0000+uint32((10+y)*320+20+x)]
		at := uint32(0xa0000) + uint32((8+y)*320+18+x)
		out = append(out, dissolve4337Pixel{at, value})
		canvas[at] = value
	}
	emit(0, 0)
	heightBits := bits.Len16(height)
	degree := heightBits + bits.Len16(width)
	poly := uint32(0)
	for n, off := range []uint32{0x42b3, 0x42d4, 0x42f5, 0x4316} {
		poly |= uint32(v.Read8(0x48a10+off+uint32(degree))) << uint(8*n)
	}
	mask := uint32((1 << heightBits) - 1)
	state := uint32(1)
	remaining := int(width) * int(height)
	for turns := 0; turns < 1000000 && remaining > 0; turns++ {
		low := state & 1
		state >>= 1
		if low != 0 {
			state ^= poly
		}
		x, y := state>>uint(heightBits), state&mask
		if x >= uint32(width) || y >= uint32(height) {
			continue
		}
		emit(uint16(x), uint16(y))
		remaining--
	}
	return out
}
func TestNativeDissolve4337Pixels(t *testing.T) {
	cases := 0
	for _, dims := range [][2]uint16{{1, 2}, {2, 1}, {2, 2}, {3, 3}, {7, 5}, {8, 8}, {9, 9}, {16, 7}, {17, 7}} {
		v := dissolve4337Seed(t, dims[0], dims[1])
		m := original(v)
		got := []dissolve4337Pixel{}
		m.WatchWrite(0xa0000, 0xb0000, func(m *machine.Machine, at uint32, old, value byte) { got = append(got, dissolve4337Pixel{at, value}) })
		returned := false
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				returned = true
				break
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
		if !returned {
			t.Fatal("dissolve did not return", dims, m.CPU.IP)
		}
		want := dissolve4337Pixels(v, dims[0], dims[1])
		if !reflect.DeepEqual(got, want) {
			index := 0
			for index < len(got) && index < len(want) && got[index] == want[index] {
				index++
			}
			t.Fatalf("dissolve pixel mismatch dims%v got%d want%d first%d", dims, len(got), len(want), index)
		}
		cases++
	}
	t.Log("dissolve4337 ordered pixel comparisons", cases, "seed", 1)
}
