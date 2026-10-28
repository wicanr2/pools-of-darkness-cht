// Spec025/#49: original glyph output and complete-machine comparisons.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

func glyph5aa4Seed(t *testing.T, width, height uint16, mapped bool) *view {
	v := copySurfaceSeed(t, 30, 20, false, false)
	v.s.IP = 0x5aa4
	v.s.Seg[cpu.SS] = 0x8000
	v.s.R[cpu.SP] = 0xe000
	stack := uint32(v.SS())*16 + uint32(v.SP())
	putRowWord(v, stack+4, 65)
	putRowWord(v, stack, 0x1234)
	putRowWord(v, stack+2, 0x4567)
	putRowWord(v, 0x48a10+0x110b, 0x100)
	putRowWord(v, 0x48a10+0x110d, 0x6000)
	for i := 0; i < 26; i++ {
		v.mem[0x60100+i] = 0
	}
	for off, value := range map[uint32]uint16{0: 0, 2: 22, 4: 12, 6: 0x100, 8: 0x7000} {
		putRowWord(v, 0x60100+off, value)
	}
	font := uint32(0x70100)
	putRowWord(v, font, 256)
	putRowWord(v, font+2, height)
	putRowWord(v, font+6, 0)
	if mapped {
		putRowWord(v, font+6, 1)
	}
	for i := uint32(0); i < 256; i++ {
		v.mem[font+8+i] = byte(i ^ 0x37)
	}
	v.mem[font+8] = 255
	v.mem[font+8+3] = 255
	putRowWord(v, font+0x108+2*65, 0x400)
	putRowWord(v, font+0x400, width)
	for i := uint32(0); i < uint32(width)*uint32(height); i++ {
		v.mem[font+0x402+i] = byte(i % 5)
	}
	if v.Read16(stack) != 0x1234 || v.Read16(stack+2) != 0x4567 || v.Read16(stack+4) != 65 {
		t.Fatal("glyph fixture overwrote call frame")
	}
	return v
}

func TestNativeGlyph5AA4Pixels(t *testing.T) {
	cases := 0
	for _, width := range []uint16{1, 2, 3, 7, 9} {
		for _, height := range []uint16{1, 2, 5} {
			for _, mapped := range []bool{false, true} {
				v := glyph5aa4Seed(t, width, height, mapped)
				m := original(v)
				type pixel struct {
					at    uint32
					value byte
				}
				got := []pixel{}
				want := []pixel{}
				m.WatchWrite(0xa0000, 0xafa00, func(_ *machine.Machine, at uint32, _, value byte) { got = append(got, pixel{at, value}) })
				for y := uint16(0); y < height; y++ {
					for x := uint16(0); x < width; x++ {
						value := v.Read8(0x70502 + uint32(y)*uint32(width) + uint32(x))
						if mapped {
							value = v.Read8(0x70108 + uint32(value))
							if value == 255 {
								continue
							}
						}
						want = append(want, pixel{0xa0000 + uint32(12+y)*320 + uint32(22+x), value})
					}
				}
				returned := false
				for m.Steps < 10000 {
					if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
						returned = true
						break
					}
					if e := m.Step(); e != nil {
						t.Fatal(e)
					}
				}
				if !returned {
					t.Fatalf("glyph no return %dx%d mapped%v IP%x", width, height, mapped, m.CPU.IP)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("glyph pixels %dx%d mapped%v got%v want%v", width, height, mapped, got, want)
				}
				if m.Mem[0x60102] != byte(22+width) {
					t.Fatal("context advance differs")
				}
				cases++
			}
		}
	}
	t.Logf("variable glyph raw/mapped ordered pixel comparisons=%d", cases)
}
