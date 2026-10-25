package xlate

import (
	"bytes"
	"testing"
)

func deferredBase(idx []byte, pal [256][3]uint8, w, h, scale int) []byte {
	o := make([]byte, w*h*scale*scale*4)
	for y := 0; y < h*scale; y++ {
		for x := 0; x < w*scale; x++ {
			c := pal[idx[(y/scale)*w+x/scale]]
			a := (y*w*scale + x) * 4
			o[a], o[a+1], o[a+2], o[a+3] = c[0], c[1], c[2], 255
		}
	}
	return o
}
func paintDeferred(p []byte, w, scale, x, y int) {
	for yy := y; yy < y+2*scale; yy++ {
		for xx := x; xx < x+3*scale; xx++ {
			at := (yy*w*scale + xx) * 4
			p[at], p[at+1], p[at+2], p[at+3] = 251, 17, 91, 255
		}
	}
}
func TestDeferredHoldEquivalentHistoryAndMixing(t *testing.T) {
	for _, scale := range []int{2, 3} {
		for _, mixed := range []bool{false, true} {
			old, newh := NewHold(16, 8, 3), NewHold(16, 8, 3)
			for frame := 0; frame < 24; frame++ {
				idx := make([]byte, 128)
				var pal [256][3]uint8
				pal[1] = [3]uint8{byte(31 + frame/4), 63, 127}
				pal[2] = pal[1]
				pal[3] = [3]uint8{19, 127, 65}
				for n := range idx {
					idx[n] = byte(1 + (n+frame)%3)
				}
				base := deferredBase(idx, pal, 16, 8, scale)
				source := NewIndexedSource(idx, pal, 16, 8, scale)
				region := func(cx, cy int, _ uint64) ([]byte, bool) {
					if frame%9 == 8 && cx == 1 {
						p := make([]byte, 8*scale*8*scale*4)
						for i := range p {
							p[i] = 239
						}
						for i := 3; i < len(p); i += 4 {
							p[i] = 255
						}
						return p, true
					}
					return nil, cx == 0
				}
				for repeat := 0; repeat < 2; repeat++ {
					a, b := append([]byte(nil), base...), make([]byte, len(base))
					if mixed && frame%2 == 0 {
						copy(b, base)
					}
					if frame%5 == 1 {
						paintDeferred(a, 16, scale, 0, 0)
						paintDeferred(b, 16, scale, 0, 0)
					}
					old.Apply(idx, base, a, scale, region)
					if mixed && frame%2 == 0 {
						newh.Apply(idx, base, b, scale, region)
					} else {
						newh.ApplyDeferred(source, b, region)
						if source.FillOriginal(b) < 0 {
							t.Fatal("deferred output rejected")
						}
					}
					if !bytes.Equal(a, b) {
						t.Fatalf("hold history/mixing differs frame%d scale%d mixed%t", frame, scale, mixed)
					}
				}
				old.Commit()
				newh.Commit()
				old.Commit()
				newh.Commit()
			}
		}
	}
}
func TestDeferredLatchEquivalentPartialAndPalette(t *testing.T) {
	for _, scale := range []int{2, 3} {
		old, newl := NewLatch(16, 8, 4, 5), NewLatch(16, 8, 4, 5)
		for frame := 0; frame < 32; frame++ {
			idx := make([]byte, 128)
			for n := range idx {
				idx[n] = byte(1 + n%2)
			}
			var pal [256][3]uint8
			pal[1] = [3]uint8{byte(43 + frame/8), 71, 98}
			pal[2] = pal[1]
			base := deferredBase(idx, pal, 16, 8, scale)
			source := NewIndexedSource(idx, pal, 16, 8, scale)
			a, b := append([]byte(nil), base...), make([]byte, len(base))
			if frame%4 == 0 {
				paintDeferred(a, 16, scale, scale, scale)
				paintDeferred(b, 16, scale, scale, scale)
			}
			old.Apply(idx, uint64(frame/8), base, a, scale)
			newl.ApplyDeferred(source, uint64(frame/8), b)
			source.FillOriginal(b)
			if !bytes.Equal(a, b) {
				t.Fatalf("latch differs frame%d scale%d", frame, scale)
			}
			old.Tick()
			newl.Tick()
		}
	}
}
func TestIndexedSourceImmutableAndSelectedPixels(t *testing.T) {
	idx := []byte{1, 2}
	var pal [256][3]uint8
	pal[1] = [3]uint8{31, 47, 63}
	pal[2] = [3]uint8{71, 83, 97}
	source := NewIndexedSource(idx, pal, 2, 1, 2)
	want := deferredBase(idx, pal, 2, 1, 2)
	idx[0] = 9
	pal[1] = [3]uint8{0, 0, 0}
	out := make([]byte, len(want))
	copy(out[0:4], []byte{255, 3, 7, 255})
	copy(want[:4], out[:4])
	if count := source.FillOriginal(out); count != 7 || !bytes.Equal(out, want) {
		t.Fatal("source mutated or selected pixel repainted", count)
	}
	bad := make([]byte, len(out))
	bad[3] = 127
	before := append([]byte(nil), bad...)
	if source.FillOriginal(bad) != -1 || !bytes.Equal(bad, before) {
		t.Fatal("partial alpha must reject atomically")
	}
}
