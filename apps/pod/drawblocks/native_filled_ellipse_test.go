package drawblocks

import (
	"math/rand"
	"reflect"
	"testing"
)

func filledEllipsePixels(cx, cy, rx, ry int32, box [4]uint16) []uint32 {
	x, y := int32(0), ry
	a, b := rx*rx, ry*ry
	gx, gy := int32(0), 2*a*ry
	err := a/4 + b - a*ry
	out := []uint32{}
	plot := func() {
		left, right := cx-x, cx+x
		if left < int32(box[0]) {
			left = int32(box[0])
		}
		if right > int32(box[2]) {
			right = int32(box[2])
		}
		if left > right {
			return
		}
		for _, row := range []int32{cy + y, cy - y} {
			if row < int32(box[1]) || row > int32(box[3]) {
				continue
			}
			for col := left; col <= right; col++ {
				out = append(out, 0xa0000+uint32(row*320+col))
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
	err += (diff + (diff >> 1) - gx - gy) >> 1
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
func TestNativeFilledEllipsePixels(t *testing.T) {
	rng := rand.New(rand.NewSource(0x39b2))
	inputs := [][4]int32{{160, 100, 1, 1}, {160, 100, 2, 1}, {160, 100, 1, 2}, {160, 100, 100, 80}, {0, 0, 20, 10}, {319, 199, 20, 10}, {50, 40, 40, 20}}
	for i := 0; i < 100; i++ {
		inputs = append(inputs, [4]int32{rng.Int31n(361) - 20, rng.Int31n(241) - 20, rng.Int31n(160) + 1, rng.Int31n(100) + 1})
	}
	cases := 0
	for _, box := range [][4]uint16{{0, 0, 319, 199}, {25, 15, 70, 150}, {0, 0, 10, 100}} {
		for _, shape := range inputs {
			v := ellipseSeed(t, shape[0], shape[1], shape[2], shape[3], box)
			v.s.IP = 0x39b2
			old := originalLinePixels(t, v)
			pixels := filledEllipsePixels(shape[0], shape[1], shape[2], shape[3], box)
			if !reflect.DeepEqual(old, pixels) {
				index := 0
				for index < len(old) && index < len(pixels) && old[index] == pixels[index] {
					index++
				}
				t.Fatal("filled ellipse pixels differ", shape, box, len(old), len(pixels), index)
			}
			cases++
		}
	}
	t.Log("filled ellipse ordered pixel comparisons", cases, "seed", 0x39b2)
}
