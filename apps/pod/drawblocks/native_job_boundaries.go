package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

// A typed rendering job owns its data calculation. These checkpoints project
// only the ABI state and ordered notifications at scheduler boundaries.
type drawingJobCheckpoint struct {
	before, after machine.HostRegisters
	native, dos   uint64
	writes        []machine.HostWrite
}

func drawingJobCode(v CPUView, ip uint16, fp smallCode) bool {
	var raw [260]byte
	if fp.size < 1 || fp.size > len(raw) {
		return false
	}
	copyView(v, 0x48a10+uint32(ip), raw[:fp.size])
	return sha256.Sum256(raw[:fp.size]) == fp.sum
}

type drawingJobProjection struct {
	r     machine.HostRegisters
	steps []drawingJobCheckpoint
}

func (j *drawingJobProjection) add(next uint16, native, dos uint64, writes []machine.HostWrite, change func()) {
	before := j.r
	change()
	j.r.IP = next
	j.steps = append(j.steps, drawingJobCheckpoint{before, j.r, native, dos, writes})
}
func (j *drawingJobProjection) plan(v CPUView, budget uint64) *machine.HostWorkPlan {
	current := boundedMemoryRegisters(v)
	first := -1
	for i, c := range j.steps {
		if c.before == current {
			first = i
			break
		}
	}
	if first < 0 || budget == 0 {
		return nil
	}
	n := len(j.steps) - first
	if uint64(n) > budget {
		n = int(budget)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: j.steps[first+n-1].after, Instructions: uint64(n), LastCS: 0x48a1, LastIP: j.steps[first+n-1].before.IP}
	for i := 0; i < n; i++ {
		c := j.steps[first+i]
		if v.DOSBoxCost() {
			p.Cycles += c.dos
		} else {
			p.Cycles += c.native
		}
		for _, w := range c.writes {
			w.Step = uint64(i + 1)
			w.CS = 0x48a1
			w.IP = c.before.IP
			p.Writes = append(p.Writes, w)
		}
	}
	return p
}
func drawingJobWord(at uint32, value uint16) []machine.HostWrite {
	return []machine.HostWrite{{Address: at, Value: byte(value)}, {Address: at + 1, Value: byte(value >> 8)}}
}
func drawingJobAdmit(v CPUView, budget uint64) bool {
	if budget == 0 || v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 || v.Flags()&cpu.DF != 0 {
		return false
	}
	m, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return false
	}
	model, known := m.CPUModel()
	return known && model == cpu.Model80386
}
func drawingByteResultFlags(before uint16, value byte) uint16 {
	f := before &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if value == 0 {
		f |= cpu.ZF
	}
	if value&128 != 0 {
		f |= cpu.SF
	}
	if bits.OnesCount8(value)%2 == 0 {
		f |= cpu.PF
	}
	return f
}
func drawingByteAddFlags(before uint16, a, b byte) uint16 {
	n := uint16(a) + uint16(b)
	r := byte(n)
	f := drawingByteResultFlags(before, r)
	if n > 255 {
		f |= cpu.CF
	}
	if (a^b^r)&16 != 0 {
		f |= cpu.AF
	}
	if (^(a^b)&(a^r))&128 != 0 {
		f |= cpu.OF
	}
	return f
}
func drawingByteSubtractFlags(before uint16, a, b byte) uint16 {
	r := a - b
	f := drawingByteResultFlags(before, r)
	if a < b {
		f |= cpu.CF
	}
	if (a^b^r)&16 != 0 {
		f |= cpu.AF
	}
	if ((a^b)&(a^r))&128 != 0 {
		f |= cpu.OF
	}
	return f
}
func drawingShiftRightFlags(before, value uint16) uint16 {
	f := nativeCompareFlags(before, value>>1, 0)
	if value&1 != 0 {
		f |= cpu.CF
	}
	if value&0x8000 != 0 {
		f |= cpu.OF
	}
	return f
}
func drawingJobBoundaryIP(ip uint16) bool {
	switch ip {
	case 0x689e, 0x68a5, 0x68a7, 0x68a8, 0x68ac, 0x68af, 0x68b2, 0x68b4, 0x68b9, 0x68be, 0x68c0, 0x68c2, 0x68c5, 0x68c6, 0x68cb, 0x68cd, 0x68d0, 0x68d5, 0x68d7, 0x68da, 0x68dd, 0x68df, 0x68e2, 0x68e4, 0x68e6, 0x68eb, 0x68ed, 0x68f2,
		0x4b6c, 0x4b70, 0x4b74, 0x4b78, 0x4b7a, 0x4b7e, 0x4b82, 0x4b86, 0x4b8a, 0x4b8b,
		0x4b9a, 0x4b9d, 0x4b9f, 0x4ba1, 0x4ba2, 0x4ba5, 0x4ba6, 0x4ba9, 0x4baa, 0x4bad, 0x4bb1, 0x4bb5, 0x4bb9, 0x4bbd, 0x4bbe, 0x4bc1, 0x4bc4,
		0x4268, 0x426c, 0x4271, 0x4273, 0x4278, 0x427c, 0x427f, 0x4281, 0x4283, 0x4285, 0x4287, 0x4289, 0x428b, 0x428d, 0x428f, 0x4291, 0x4295, 0x4299, 0x429a, 0x429d, 0x429f:
		return true
	}
	return false
}
func planDrawingJobBoundary(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !drawingJobBoundaryIP(v.IP()) || !drawingJobAdmit(v, budget) {
		return nil
	}
	switch {
	case v.IP() >= 0x689e && v.IP() <= 0x68f2:
		return planImageHeaderBoundary(v, budget)
	case v.IP() >= 0x4b6c && v.IP() <= 0x4b8b:
		return planFadeBlendBoundary(v, budget)
	case v.IP() >= 0x4b9a && v.IP() <= 0x4bc4:
		return planFadeColorBoundary(v, budget)
	case v.IP() >= 0x4268 && v.IP() <= 0x429f:
		return planFillRowBoundary(v, budget)
	}
	return nil
}
