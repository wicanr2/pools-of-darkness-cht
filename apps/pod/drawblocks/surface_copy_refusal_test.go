package drawblocks

import (
	"bytes"
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

func TestNativeCopyPrefixRefusals(t *testing.T) {
	seed := copySurfaceSeed(t, 8, 5, true, true)
	var code [593]byte
	copyView(seed, 0x48a10+0x3f5e, code[:])
	if sha256.Sum256(code[:]) != surfaceCopyCode {
		t.Fatal("source fingerprint changed")
	}
	cases := 0
	for _, kind := range []string{"mode", "CS", "IP", "code", "SP", "stack-video", "stack-driver", "source-slot", "dest-slot", "source-cycle", "dest-cycle", "source-root", "dest-root", "source-bounds", "dest-bounds", "root-width"} {
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
			v.mem[0x48a10+0x3f5e] ^= 1
		case "SP":
			v.s.R[cpu.SP] = 4
		case "stack-video":
			v.s.Seg[cpu.SS] = 0xa000
		case "stack-driver":
			v.s.Seg[cpu.SS] = 0x48a1
			v.s.R[cpu.SP] = 0x7000
		case "source-slot":
			putRowWord(v, stack+4, 256)
		case "dest-slot":
			putRowWord(v, stack+6, 256)
		case "source-cycle":
			putRowWord(v, 0x48a10+0xbd, 0)
		case "dest-cycle":
			putRowWord(v, 0x48a10+0xbf, 2)
		case "source-root":
			putRowWord(v, 0x48a10+0xc1, 0x9000)
		case "dest-root":
			putRowWord(v, 0x48a10+0xbf, 6)
			putRowWord(v, 0x48a10+0xc3, 0x9000)
		case "source-bounds":
			putRowWord(v, 0x48a10+0x4bd, 65535)
		case "dest-bounds":
			putRowWord(v, 0x48a10+0x8bf, 320)
		case "root-width":
			putRowWord(v, 0x48a10+0x8c1, 300)
		}
		before := append([]byte(nil), v.mem...)
		regs := v.s
		if planSurfaceCopyPrefix(v) != nil {
			t.Fatal("invalid copy prefix accepted", kind)
		}
		if !bytes.Equal(before, v.mem) || regs != v.s {
			t.Fatal("prefix refusal mutated")
		}
		cases++
	}
	t.Log("copy prefix refusals", cases)
}
