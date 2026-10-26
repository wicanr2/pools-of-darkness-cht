package drawblocks

import (
	"math/rand"
	"reflect"
	"testing"
)

func ellipseSeed(t *testing.T, cx, cy, rx, ry int32, box [4]uint16) *view {
	v := axisSeed(t, [4]uint16{uint16(cx), uint16(cy), uint16(rx), uint16(ry)}, 320)
	v.s.IP = 0x346a
	for _, w := range [][2]uint16{{0xbd, 2}, {0x4bd, box[0]}, {0x6bd, box[1]}, {0x8bd, box[2]}, {0xabd, box[3]}, {0xcbd, 0xc0}, {0xbf, 0xa000}, {0x2bf, 0}, {0x4bf, 0}, {0x6bf, 0}, {0x8bf, 319}, {0xabf, 199}, {0xcbf, 0x80}} {
		putRowWord(v, 0x48a10+uint32(w[0]), w[1])
	}
	return v
}
func midpointEllipsePixels(cx, cy, rx, ry int32, box [4]uint16) []uint32 {
	x, y := int32(0), ry
	a, b := rx*rx, ry*ry
	gx, gy := int32(0), 2*a*ry
	err := a/4 + b - a*ry
	out := []uint32{}
	plot := func() {
		for _, xy := range [][2]int32{{cx + x, cy + y}, {cx + x, cy - y}, {cx - x, cy + y}, {cx - x, cy - y}} {
			if xy[0] >= int32(box[0]) && xy[0] <= int32(box[2]) && xy[1] >= int32(box[1]) && xy[1] <= int32(box[3]) {
				out = append(out, 0xa0000+uint32(xy[1]*320+xy[0]))
			}
		}
	}
	for gx < gy {
		plot()
		if err >= 0 {
			y--
			gy -= 2 * a
			err -= gy
		}
		x++
		gx += 2 * b
		err += gx + b
	}
	diff := a - b
	half := func(x int32) int32 {
		if x < 0 {
			return (x - 1) / 2
		}
		return x / 2
	}
	err += half(diff + half(diff) - gx - gy)
	for y >= 0 {
		plot()
		if err < 0 {
			x++
			gx += 2 * b
			err += gx
		}
		y--
		gy -= 2 * a
		err -= gy - a
	}
	return out
}
func TestNativeEllipsePixels(t *testing.T) {
	rng := rand.New(rand.NewSource(0x346a))
	inputs := [][4]int32{{160, 100, 1, 1}, {160, 100, 2, 1}, {160, 100, 1, 2}, {160, 100, 100, 80}, {0, 0, 20, 10}, {319, 199, 20, 10}, {50, 40, 40, 20}}
	for i := 0; i < 100; i++ {
		inputs = append(inputs, [4]int32{rng.Int31n(361) - 20, rng.Int31n(241) - 20, rng.Int31n(160) + 1, rng.Int31n(100) + 1})
	}
	cases := 0
	for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}, {0, 0, 10, 100}} {
		for _, shape := range inputs {
			v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
			old := originalLinePixels(t, v)
			goPixels := midpointEllipsePixels(shape[0], shape[1], shape[2], shape[3], box)
			if !reflect.DeepEqual(old, goPixels) {
				index := 0
				for index < len(old) && index < len(goPixels) && old[index] == goPixels[index] {
					index++
				}
				t.Fatalf("ellipse pixels differ box%v shape%v old%d Go%d first%d", box, shape, len(old), len(goPixels), index)
			}
			cases++
		}
	}
	t.Log("ellipse ordered pixel comparisons", cases, "seed", 0x346a)
}
