package drawblocks

import "testing"

func TestPrimitiveOriginalClipQuirk(t *testing.T) {
	cases := 0
	for _, dos := range []bool{false, true} {
		for _, budget := range []uint64{1, 7, 1024} {
			v := axisSeed(t, [4]uint16{0xfff6, 0, 20, 90}, 320)
			v.dos = dos
			for i := uint16(0); i < 2; i++ {
				seg, flags := uint16(2), uint16(0xc0)
				right, bottom := uint16(10), uint16(100)
				if i == 1 {
					seg, flags = 0xa000, 0x80
					right, bottom = 319, 199
				}
				for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: 0, 0x6bd: 0, 0x8bd: right, 0xabd: bottom, 0xcbd: flags} {
					putRowWord(v, 0x48a10+uint32(off+i*2), value)
				}
			}
			oracle := callMachine(v)
			pixels, maxX := 0, uint32(0)
			oracle.WatchWrites(0xa0000, 0xaffff, func(a uint32, _, _ uint8) {
				pixels++
				x := (a - 0xa0000) % 320
				if x > maxX {
					maxX = x
				}
			})
			runCall(t, oracle, v, false)
			oracle.WatchWrites(0, 0, nil)
			if pixels != 62 || maxX != 20 {
				t.Fatalf("original clipping witness changed pixels%d maxX%d", pixels, maxX)
			}
			primitiveWhole(t, v, budget)
			cases++
		}
	}
	t.Logf("primitive original clipping-count quirk SaveState+PortLog cases=%d original pixels=62 maxX=20 selected right=10 CPU/fixed primitive pixels=0", cases)
}
