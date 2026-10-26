package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var waitCode = smallFingerprint(24, "e10eb7c44d7cf6de7a37a2b893c7d9fa2e1e0dcf3b3d2a53917e2b19341929dc")

func waitTestFlags(flags uint16, value byte) uint16 {
	flags &^= uint16(cpu.CF | cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
	if value&8 == 0 {
		flags |= cpu.PF | cpu.ZF
	}
	return flags
}

// Two data states: leave the current retrace, then wait for its next edge.
// The short prefixes below only preserve event-boundary ABI continuations.
func planWait(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 || budget == 0 {
		return nil
	}
	phase := 0
	post, branch, finish := false, false, false
	switch v.IP() {
	case 0x2a90, 0x2a9d:
		phase = 1
	case 0x2a9e:
		phase = 1
		post = true
	case 0x2aa0:
		phase = 1
		branch = true
	case 0x2aa2:
		phase = 2
	case 0x2aa3:
		phase = 2
		post = true
	case 0x2aa5:
		phase = 2
		branch = true
	case 0x2aa7:
		finish = true
	default:
		return nil
	}
	var code [24]byte
	copyView(v, 0x48a10+0x2a90, code[:])
	if sha256.Sum256(code[:]) != waitCode.sum {
		return nil
	}
	stack := uint32(v.SS())*16 + uint32(v.SP())
	if v.SP() > 65532 || stack < 0x500 || stack+4 > 0xa0000 {
		return nil
	}
	p := imageBasePlan(v)
	charge := func(n, c uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n
		} else {
			p.Cycles += c
		}
	}
	result := func() *machine.HostWorkPlan {
		if p.Instructions == 0 {
			return nil
		}
		return p
	}
	if v.IP() == 0x2a90 {
		if budget < 4 || v.Read16(0x463) != 0x3d4 {
			return nil
		}
		p.After.R[cpu.AX] = 0x40
		p.After.R[cpu.DX] = 0x3da
		p.After.Seg[cpu.ES] = 0x40
		p.After.Flags = v.Flags()&^uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF) | cpu.SF
		charge(4, 31)
		p.After.IP, p.LastIP = 0x2a9d, 0x2a9a
	} else if !finish && v.DX() != 0x3da {
		return nil
	}
	transition := func(taken bool) {
		if phase == 1 {
			p.LastIP = 0x2aa0
			if taken {
				p.After.IP = 0x2a9d
			} else {
				phase = 2
				p.After.IP = 0x2aa2
			}
		} else {
			p.LastIP = 0x2aa5
			if taken {
				p.After.IP = 0x2aa2
			} else {
				finish = true
				p.After.IP = 0x2aa7
			}
		}
	}
	if branch {
		taken := v.Flags()&cpu.ZF == 0
		if phase == 2 {
			taken = !taken
		}
		cost := uint64(2)
		if taken {
			cost = 7
		}
		charge(1, cost)
		transition(taken)
	}
	if post {
		value := byte(v.AX())
		p.After.Flags = waitTestFlags(v.Flags(), value)
		if budget-p.Instructions == 1 {
			charge(1, 2)
			p.LastIP = v.IP()
			p.After.IP = v.IP() + 2
			return result()
		}
		taken := value&8 != 0
		if phase == 2 {
			taken = !taken
		}
		cost := uint64(4)
		if taken {
			cost = 9
		}
		charge(2, cost)
		transition(taken)
	}
	var snapshot machine.VGAStatusSnapshot
	known := false
	if reader, ok := v.(interface {
		VGAStatusSnapshot() (machine.VGAStatusSnapshot, bool)
	}); ok {
		snapshot, known = reader.VGAStatusSnapshot()
		known = known && snapshot.Steps == v.Steps()
	}
	for p.Instructions < budget {
		if finish {
			p.After.R[cpu.SP] += 4
			p.After.IP = v.Read16(stack)
			p.After.Seg[cpu.CS] = v.Read16(stack + 2)
			p.LastIP = 0x2aa7
			charge(1, 24)
			return result()
		}
		if !known {
			return result()
		}
		value, ok := snapshot.At(p.Instructions+1, uint64(len(p.VideoStatusReads)+1))
		if !ok {
			return result()
		}
		readIP := uint16(0x2a9d)
		if phase == 2 {
			readIP = 0x2aa2
		}
		p.After.R[cpu.AX] = p.After.R[cpu.AX]&0xff00 | uint16(value)
		moment, cost := p.Instructions+1, p.Cycles+16
		if v.DOSBoxCost() {
			cost = moment
		}
		p.VideoStatusReads = append(p.VideoStatusReads, machine.HostVideoStatusRead{Value: value, Step: moment, Cycles: cost})
		room := budget - p.Instructions
		if room < 3 {
			charge(1, 16)
			p.After.IP, p.LastIP = readIP+1, readIP
			if room == 2 {
				p.After.Flags = waitTestFlags(p.After.Flags, value)
				charge(1, 2)
				p.LastIP = readIP + 1
				p.After.IP = readIP + 3
			}
			return result()
		}
		p.After.Flags = waitTestFlags(p.After.Flags, value)
		taken := value&8 != 0
		if phase == 2 {
			taken = !taken
		}
		cycleCost := uint64(20)
		if taken {
			cycleCost = 25
		}
		charge(3, cycleCost)
		transition(taken)
	}
	return result()
}
