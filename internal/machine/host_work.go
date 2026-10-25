package machine

import "fmt"

type HostRegisters struct {
	R         [8]uint16
	Seg       [4]uint16
	IP, Flags uint16
}
type HostWrite struct {
	Address uint32
	Value   byte
	Step    uint64
	CS, IP  uint16
}
type HostWorkPlan struct {
	After                HostRegisters
	Instructions, Cycles uint64
	LastCS, LastIP       uint16
	Writes               []HostWrite
}

func (m *Machine) ObserveHostCalls(fn func()) {
	m.onHostCall = fn
	m.CPU.HostCalls = fn != nil
	m.hostCallTargets = nil
}
func (m *Machine) SetHostCallTargets(targets []uint32) {
	m.hostCallTargets = append([]uint32(nil), targets...)
}
func (m *Machine) AfterGuestCall() {
	if m.onHostCall == nil {
		return
	}
	if len(m.hostCallTargets) > 0 {
		at := (uint32(m.CPU.Seg[1]) << 4) + uint32(m.CPU.IP)
		found := false
		for _, target := range m.hostCallTargets {
			if at == target {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}
	before := m.Steps
	m.onHostCall()
	if m.Steps > before {
		m.accelCredit = m.Steps - before + 1
	}
}

// HostWorkBudget follows the pinned Step event model, not hardware cycles.
func (m *Machine) HostWorkBudget(limit uint64) uint64 {
	if limit == 0 || m.CycleClock || m.CPU.Halted || m.CPU.Flags&0x100 != 0 || m.irq0Pending || len(m.cbQueue) > 0 || len(m.keyQueue) > 0 || len(m.timedKeys) > 0 || len(m.wordWatches) > 0 || len(m.breaks) > 0 || m.Coverage != nil || m.TraceSegs || m.WatchDSOn || m.onRead != nil || m.accelCredit > 0 {
		return 0
	}
	before := func(on bool, next uint64) {
		if !on {
			return
		}
		if next <= m.Steps {
			limit = 0
		} else if next-m.Steps < limit {
			limit = next - m.Steps
		}
	}
	before(m.IRQ0Every > 0, m.nextIRQ0)
	before(m.VGAFrameEvery > 0, m.nextFrame)
	before(m.periodic.on, m.periodic.next)
	before(m.RunLimit > 0, m.RunLimit)
	return limit
}
func (m *Machine) ValidateHostWork(p *HostWorkPlan) error {
	if p == nil || p.Instructions == 0 || p.Instructions > m.HostWorkBudget(p.Instructions) || ^uint64(0)-m.Steps < p.Instructions || ^uint64(0)-m.CPU.Cycles < p.Cycles {
		return fmt.Errorf("machine: host work exceeds event/counter budget")
	}
	last := uint64(0)
	for _, w := range p.Writes {
		if w.Step == 0 || w.Step < last || w.Step > p.Instructions || w.Address < 0x500 || w.Address >= vgaLo+64000 {
			return fmt.Errorf("machine: unsupported host write")
		}
		last = w.Step
		if w.Address >= vgaLo && (m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13) {
			return fmt.Errorf("machine: unsupported host video surface")
		}
		if m.onWrite != nil && m.watchLo <= w.Address && w.Address <= m.watchHi {
			return fmt.Errorf("machine: host write intersects observer")
		}
		for _, watch := range m.writeWatches {
			if watch.lo <= w.Address && w.Address <= watch.hi {
				return fmt.Errorf("machine: host write intersects guard")
			}
		}
	}
	return nil
}
func (m *Machine) CommitHostWork(p *HostWorkPlan) error {
	if err := m.ValidateHostWork(p); err != nil {
		return err
	}
	defer func() { m.hostEffect = nil }()
	for i := range p.Writes {
		m.hostEffect = &p.Writes[i]
		m.Write8(p.Writes[i].Address, p.Writes[i].Value)
	}
	m.hostEffect = nil
	m.Steps += p.Instructions
	m.CPU.Cycles += p.Cycles
	m.CPU.ApplyHostRegisters(p.After.R, p.After.Seg, p.After.IP, p.After.Flags, p.LastCS, p.LastIP)
	m.insnCS, m.insnIP = p.LastCS, p.LastIP
	m.HostCallCommits++
	m.HostCallInstructions += p.Instructions
	return nil
}
func (m *Machine) videoWriteContext() (step uint64, cs, ip uint16) {
	if w := m.hostEffect; w != nil {
		return m.Steps + w.Step, w.CS, w.IP
	}
	cs, ip = m.CPU.OpAddr()
	return m.Steps, cs, ip
}
