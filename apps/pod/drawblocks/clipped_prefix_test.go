package drawblocks

import (
	"math/rand"
	"testing"
)

func TestNativeClippedPrefixes(t *testing.T) {
	rng := rand.New(rand.NewSource(0x2cfc))
	inputs := [][4]int32{{-10, 20, 100, 30}, {100, 20, -10, 30}, {-10, 0, 20, 90}, {20, 0, -10, 90}, {30, -10, 50, 100}, {50, -10, 30, 100}, {30, 100, 50, -10}}
	for i := 0; i < 100; i++ {
		inputs = append(inputs, [4]int32{rng.Int31n(121) - 30, rng.Int31n(251) - 30, rng.Int31n(121) - 30, rng.Int31n(251) - 30})
	}
	cases := 0
	for _, box := range [][4]int32{{0, 0, 319, 199}, {25, 15, 70, 40}, {0, 0, 10, 100}, {25, 15, 70, 150}, {25, 15, 200, 40}} {
		for _, xy := range inputs {
			if xy[0] == xy[2] || xy[1] == xy[3] {
				continue
			}
			v := clipLineSeed(t, xy, true)
			for _, w := range [][2]uint16{{0x4bd, uint16(box[0])}, {0x6bd, uint16(box[1])}, {0x8bd, uint16(box[2])}, {0xabd, uint16(box[3])}} {
				putRowWord(v, 0x48a10+uint32(w[0]), w[1])
			}
			for _, dos := range []bool{false, true} {
				v.dos = dos
				p := planClippedLinePrefix(v)
				if p == nil {
					continue
				}
				t.Run("prefix", func(t *testing.T) { compare(t, v, p) })
				if t.Failed() {
					t.Fatalf("shallow clipped prefix mismatch box%v xy%v dos%v steps%d cycles%d", box, xy, dos, p.Instructions, p.Cycles)
				}
				cases++
			}
		}
	}
	t.Log("shallow clipped prefix comparisons", cases, "seed", 0x2cfc)
}
