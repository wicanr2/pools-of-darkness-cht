package drawblocks

// Spec025/#49: decode image runs directly with Go COPY/FILL spans.
import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var rleCode = smallFingerprint(40, "1926d07bb1c1e91594b6a388138e57b204d8efb327addc7cebf32d0dd7ec5818")
var flippedRLECode = smallFingerprint(42, "674e9d2236f6de2d9da5e9a1ff2e2bd26fa8833d9dc5e6f5502ee3e5ba3cf676")

// planRLERuns decodes the asset's length/repeat packets. The loop traverses
// compressed image runs, with no x86 opcode or instruction-address dispatch.
func planRLERuns(v CPUView, budget uint64) *machine.HostWorkPlan {
	return planRLERunsProgram(v, budget, v.IP() == 0x6b33)
}
func planRLERunsProgram(v CPUView, budget uint64, flipped bool) *machine.HostWorkPlan {
	entry, header, fp := uint16(0x68f4), uint16(0x689e), rleCode
	if flipped {
		entry, header, fp = 0x6b33, 0x6abe, flippedRLECode
	}
	if v.CS() != 0x48a1 || v.IP() != entry || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.CX()&0xff00 != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [42]byte
	copyView(v, 0x48a10+uint32(entry), code[:fp.size])
	if sha256.Sum256(code[:fp.size]) != fp.sum {
		return nil
	}
	si, end, di := uint32(v.SI()), uint32(v.BX()), uint32(v.DI())
	srcBase := uint32(v.DS()) * 16
	if si >= end || srcBase+end > 0xa0000 {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: entry, Flags: v.Flags()}, LastCS: 0x48a1}
	for si < end {
		control := v.Read8(srcBase + si)
		count := uint32(control>>1) + 1
		repeat := control&1 != 0
		used := count + 1
		if repeat {
			used = 2
		}
		if si+used > end || di+count > 64000 {
			break
		}
		last := si+used == end
		steps, cycles, wordStep := uint64(11), uint64(34), uint64(7)
		wordIP, byteIP, exitIP := uint16(0x68ff), uint16(0x6903), uint16(0x6907)
		loads := uint64(1)
		if repeat {
			steps = 13
			cycles = 48
			wordStep = 9
			wordIP = 0x6910
			byteIP = 0x6914
			exitIP = 0x6918
			loads = 2
		}
		if last {
			steps++
			cycles += 2
			if repeat {
				exitIP = 0x691a
			} else {
				exitIP = 0x6909
			}
		}
		if flipped {
			if repeat {
				wordIP, byteIP = 0x6b50, 0x6b54
				exitIP = 0x6b58
				if last {
					exitIP = 0x6b5a
				}
			} else {
				wordIP, byteIP = 0x6b3e, 0x6b42
				exitIP = 0x6b46
				if last {
					exitIP = 0x6b48
				}
			}
		}
		iterations := uint64(count/2 + count%2)
		cycles += 5 * iterations
		if v.DOSBoxCost() {
			cycles = steps + loads + iterations
		}
		if steps > budget-p.Instructions {
			break
		}
		n := p.Instructions
		value := uint16(0)
		var data []byte
		if repeat {
			value = uint16(v.Read8(srcBase+si+1)) * 257
			p.After.R[cpu.AX] = value
		} else {
			data = make([]byte, int(count))
			copyView(v, srcBase+si+1, data)
			p.After.R[cpu.AX] = p.After.R[cpu.AX]&0xff00 | uint16(control>>1)
		}
		if count/2 > 0 {
			s := machine.HostTransfer{Address: 0xa0000 + di, Value: value, Count: uint16(count / 2), Width: 2, Step: n + wordStep, CS: 0x48a1, IP: wordIP}
			if data != nil {
				s.Data = append([]byte(nil), data[:count/2*2]...)
			}
			p.Transfers = append(p.Transfers, s)
		}
		if count%2 != 0 {
			s := machine.HostTransfer{Address: 0xa0000 + di + count/2*2, Value: value, Count: 1, Width: 1, Step: n + wordStep + 2, CS: 0x48a1, IP: byteIP}
			if data != nil {
				s.Data = []byte{data[len(data)-1]}
			}
			p.Transfers = append(p.Transfers, s)
		}
		si += used
		di += count
		p.Instructions += steps
		p.Cycles += cycles
		p.LastIP = exitIP
		p.After.R[cpu.SI] = uint16(si)
		p.After.R[cpu.DI] = uint16(di)
		p.After.R[cpu.CX] = 0
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(si), uint16(end))
		if last {
			p.After.IP = header
			break
		}
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}
