package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/rand"
	"reflect"
	"testing"
)

func clipLineSeed(t *testing.T, xy [4]int32, child bool) *view {
	v := axisSeed(t, [4]uint16{uint16(xy[0]), uint16(xy[1]), uint16(xy[2]), uint16(xy[3])}, 320)
	if child {
		for _, w := range [][2]uint16{{0xbd, 2}, {0x4bd, 25}, {0x6bd, 15}, {0x8bd, 70}, {0xabd, 40}, {0xcbd, 0xc0}, {0xbf, 0xa000}, {0x2bf, 0}, {0x4bf, 0}, {0x6bf, 0}, {0x8bf, 319}, {0xabf, 199}, {0xcbf, 0x80}} {
			putRowWord(v, 0x48a10+uint32(w[0]), w[1])
		}
	}
	return v
}

func originalLinePixels(t *testing.T, v *view) []uint32 {
	m := original(v)
	out := []uint32{}
	m.WatchWrite(0xa0000, 0xb0000, func(m *machine.Machine, at uint32, old, value byte) { out = append(out, at) })
	for m.Steps < 100000 {
		if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
			return out
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("line did not return")
	return nil
}

func counterLinePixels(xy [4]int32, box [4]int32) []uint32 {
	x0, y0, x1, y1 := xy[0], xy[1], xy[2], xy[3]
	if y0 > y1 {
		x0, x1, y0, y1 = x1, x0, y1, y0
	}
	dx, direction := x1-x0, int32(1)
	if dx < 0 {
		dx = -dx
		direction = -1
	}
	dy := y1 - y0
	major, minor := dx, dy
	steep := dy > dx
	if steep {
		major, minor = dy, dx
	}
	left, top, right, bottom := box[0], box[1], box[2], box[3]
	atStep := func(k int32) (int32, int32) {
		moved := (major/2 + k*minor) / major
		if steep {
			return x0 + direction*moved, y0 + k
		}
		return x0 + direction*k, y0 + moved
	}
	first := int32(-1)
	for k := int32(0); k <= major; k++ {
		x, y := atStep(k)
		if x >= left && x <= right && y >= top && y <= bottom {
			first = k
			break
		}
	}
	out := []uint32{}
	if first < 0 {
		return out
	}
	_, firstY := atStep(first)
	remaining := major - first
	err := (major/2 + first*minor) % major
	available := bottom - firstY
	if !steep {
		if x1 < left {
			remaining -= left - x1
		}
		if x1 > right {
			remaining -= x1 - right
		}
	}
	horizontalRemaining := dx
	if steep {
		if y1 > bottom {
			remaining -= y1 - bottom
		}
		skipTop := int32(0)
		if y0 < top {
			skipTop = top - y0
		}
		horizontalRemaining -= (major/2 + skipTop*minor) / major
		margin := int32(0)
		if x1 < left {
			margin = x1 - left
		}
		if x1 > right {
			margin = right - x1
		}
		horizontalRemaining += margin
	}
	for k := first; k <= major; k++ {
		x, y := atStep(k)
		out = append(out, 0xa0000+uint32(uint16(y*320+x)))
		remaining--
		if remaining < 0 {
			break
		}
		err += minor
		if err >= major {
			err -= major
			if steep {
				horizontalRemaining--
				if horizontalRemaining < 0 {
					break
				}
			} else {
				available--
				if available < 0 {
					break
				}
			}
		}
	}
	return out
}

func TestNativeClipCounterPixels(t *testing.T) {
	rng := rand.New(rand.NewSource(0x2cfc))
	inputs := [][4]int32{{-10, 0, 20, 90}, {20, 0, -10, 90}, {-10, -10, 100, 100}, {100, -10, -10, 100}, {-10, 20, 100, 30}, {100, 20, -10, 30}, {30, -10, 50, 100}, {50, -10, 30, 100}, {-100, -50, 420, 250}, {420, -50, -100, 250}}
	for i := 0; i < 1000; i++ {
		xy := [4]int32{rng.Int31n(521) - 100, rng.Int31n(301) - 50, rng.Int31n(521) - 100, rng.Int31n(301) - 50}
		if xy[0] != xy[2] && xy[1] != xy[3] {
			inputs = append(inputs, xy)
		}
	}
	cases := 0
	for _, box := range [][4]int32{{0, 0, 319, 199}, {25, 15, 70, 40}, {0, 0, 10, 100}, {25, 15, 70, 150}, {25, 15, 200, 40}} {
		for _, xy := range inputs {
			v := clipLineSeed(t, xy, true)
			for _, w := range [][2]uint16{{0x4bd, uint16(box[0])}, {0x6bd, uint16(box[1])}, {0x8bd, uint16(box[2])}, {0xabd, uint16(box[3])}} {
				putRowWord(v, 0x48a10+uint32(w[0]), w[1])
			}
			original := originalLinePixels(t, v)
			candidate := counterLinePixels(xy, box)
			if !reflect.DeepEqual(original, candidate) {
				index := 0
				for index < len(original) && index < len(candidate) && original[index] == candidate[index] {
					index++
				}
				t.Fatalf("counter pixels differ seed2cfc box%v xy%v original%d Go%d first%d original%v Go%v", box, xy, len(original), len(candidate), index, original, candidate)
			}
			cases++
		}
	}
	t.Log("counter clipping pixel comparisons", cases, "seed", 0x2cfc)
}
