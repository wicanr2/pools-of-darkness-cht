package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestTransferOverlayReadbackAndOverlap(t *testing.T) {
	for _, back := range []bool{false, true} {
		for _, word := range []bool{false, true} {
			mem := make([]byte, 1<<20)
			for i := 0; i < 512; i++ {
				mem[0x60000+i] = byte(i)
			}
			flags := uint16(0xf002)
			if back {
				flags |= 0x400
			}
			v := &view{mem: mem, s: machine.HostRegisters{R: [8]uint16{0x6655, 12, 0, 0, 0, 0, 64, 128}, Seg: [4]uint16{0x6000, 0x1000, 0x8000, 0x6000}, Flags: flags}}
			e := &effect{v: v, s: v.s, shadow: map[uint32]byte{}}
			op := 0xaa
			if word {
				op = 0xab
			}
			e.mark(0, 2)
			e.str(op, 0xf3, 3)
			for i := 0; i < 12; i++ {
				off := 128 + i
				if back {
					off = 128 - i
				}
				if got := e.read(0x6000, uint16(off), 8); word && got != 0x55 && got != 0x66 || !word && got != 0x55 {
					t.Fatal("fill readback", back, word, i, got)
				}
			}
			e.write(0x6000, 128, 9, 8)
			if e.read(0x6000, 128, 8) != 9 {
				t.Fatal("scalar after transfer")
			}
			e.s.R[0] = 0x8877
			e.s.R[1] = 12
			e.s.R[7] = 128
			e.mark(1, 2)
			e.str(op, 0xf3, 3)
			if e.read(0x6000, 128, 8) != 0x77 {
				t.Fatal("transfer after scalar")
			}
			// Forward overlapping copy repeats the newly written byte, as REP MOVSB.
			e.s.Flags &^= 0x400
			e.s.R[6] = 0
			e.s.R[7] = 1
			e.s.R[1] = 16
			e.mark(2, 2)
			e.str(0xa4, 0xf3, 3)
			for i := 1; i <= 16; i++ {
				if e.read(0x6000, uint16(i), 8) != uint16(mem[0x60000]) {
					t.Fatal("overlap readback")
				}
			}
		}
	}
}
func TestVerifiedPrefixStopsBeforeBadBlock(t *testing.T) {
	mem := input(t)
	v := &view{mem: mem, s: machine.HostRegisters{R: [8]uint16{0, 0, 0, 0, 0xe000}, Seg: [4]uint16{0, 0x1ed4, 0x8000, 0x6000}, IP: 0x1a5c, Flags: 0xf002}}
	word := func(off, val uint16) { mem[0x80000+int(off)] = byte(val); mem[0x80000+int(off)+1] = byte(val >> 8) }
	word(0xe000, 0x1234)
	word(0xe002, 0x1111)
	word(0xe004, 16)
	word(0xe006, 0)
	word(0xe008, 0x6000)
	word(0xe00a, 32)
	word(0xe00c, 0x6000)
	at := uint32(0x1ed40 + 0x1a78)
	mem[at] ^= 1
	p := Plan(v, 4096)
	if p == nil || p.After.IP != 0x1a78 || p.Instructions != 8 {
		t.Fatal("lost verified prefix", p)
	}
	compare(t, v, p)
	// An alias segmented entry must not consume the same physical code.
	alias := *v
	alias.s.Seg[1]--
	alias.s.IP += 16
	if Plan(&alias, 4096) != nil {
		t.Fatal("segmented alias accepted")
	}
}
