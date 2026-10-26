package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNativeEllipsePrefixRefusals(t *testing.T) {
	seed := ellipseSeed(t, 160, 100, 10, 20, [4]uint16{0, 0, 319, 199})
	cases := 0
	for _, kind := range []string{"mode", "CS", "IP", "code", "SP", "stack-video", "stack-driver", "slot", "cycle", "root", "surface-bounds", "root-width", "zero-x", "zero-y", "large-x", "large-y", "center-x", "center-y"} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		stack := uint32(v.SS())*16 + uint32(v.SP())
		switch kind {
		case "mode":
			v.mem[0x449] = 0xd
		case "CS":
			v.s.Seg[cpu.CS]++
		case "IP":
			v.s.IP++
		case "code":
			v.mem[0x48a10+0x346a] ^= 1
		case "SP":
			v.s.R[cpu.SP] = 4
		case "stack-video":
			v.s.Seg[cpu.SS] = 0xa000
		case "stack-driver":
			v.s.Seg[cpu.SS] = 0x48a1
			v.s.R[cpu.SP] = 0x7000
		case "slot":
			putRowWord(v, stack+4, 256)
		case "cycle":
			putRowWord(v, 0x48a10+0xbd, 0)
		case "root":
			putRowWord(v, 0x48a10+0xbf, 0x9000)
		case "surface-bounds":
			putRowWord(v, 0x48a10+0x4bd, 65535)
		case "root-width":
			putRowWord(v, 0x48a10+0x8bf, 300)
		case "zero-x":
			putRowWord(v, stack+10, 0)
		case "zero-y":
			putRowWord(v, stack+12, 0)
		case "large-x":
			putRowWord(v, stack+10, 161)
		case "large-y":
			putRowWord(v, stack+12, 101)
		case "center-x":
			putRowWord(v, stack+6, 32767)
		case "center-y":
			putRowWord(v, stack+8, 32767)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planEllipsePrefix(v) != nil {
			t.Fatal("invalid ellipse prefix accepted", kind)
		}
		if !bytes.Equal(before, v.mem) || v.s != regs {
			t.Fatal("ellipse prefix refusal changed state")
		}
		cases++
	}
	t.Log("ellipse prefix refusals", cases)
}
