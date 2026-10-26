package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var mirrorPairCode = smallFingerprint(50, "71ffb9fb40b99437973306c9ef582b52f24b0014fb32e1feed59703a7ce12575")

func mirrorPairBase(ip uint16) uint16 {
	for _, base := range [...]uint16{0x6c9e, 0x6d42, 0x6df3} {
		if ip >= base && ip < base+48 && (ip-base)%12 == 0 {
			return base
		}
	}
	return 0
}

// planMirrorPairs reverses image pixel pairs, preserving the original ordered
// word writes. Entry stripe only determines compatibility timing metadata.
func planMirrorPairs(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 || v.CX() == 0 || v.CX() > 64 {
		return nil
	}
	base := mirrorPairBase(v.IP())
	if base == 0 {
		return nil
	}
	var code [50]byte
	copyView(v, 0x48a10+uint32(base), code[:])
	if sha256.Sum256(code[:]) != mirrorPairCode.sum {
		return nil
	}
	src := uint32(v.DS()) * 16
	si, di, pairs := uint32(v.SI()), uint32(v.DI()), uint32(v.CX())
	if si+pairs*2 > 65536 || src+si+pairs*2 > 0xa0000 || di+2 > 64000 || pairs*2 > di%320+2 || di%320 == 319 {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}, LastCS: 0x48a1}
	start := uint32((v.IP() - base) / 12)
	for x := uint32(0); x < pairs; x++ {
		stripe := (start + x) % 4
		last := x+1 == pairs
		steps, cycles := uint64(6), uint64(22)
		if last {
			cycles += 5
		} else if stripe == 3 {
			steps++
			cycles += 7
		}
		if steps > budget-p.Instructions {
			break
		}
		at := base + uint16(stripe)*12
		a, b := v.Read8(src+si+x*2), v.Read8(src+si+x*2+1)
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + di - x*2, Data: []byte{b, a}, Count: 1, Width: 2, Step: p.Instructions + 3, CS: 0x48a1, IP: at + 3})
		p.After.R[cpu.AX] = uint16(a)<<8 | uint16(b)
		p.After.R[cpu.SI] += 2
		p.After.R[cpu.DI] -= 2
		p.After.R[cpu.CX]--
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(pairs-x), 1) &^ cpu.CF
		if di-x*2 < 2 {
			p.After.Flags |= cpu.CF
		}
		p.LastIP = at + 10
		p.After.IP = at + 12
		if last {
			p.After.IP = base + 50
		} else if stripe == 3 {
			p.LastIP = base + 48
			p.After.IP = base
		}
		p.Instructions += steps
		if v.DOSBoxCost() {
			cycles = steps + 1
		}
		p.Cycles += cycles
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}
