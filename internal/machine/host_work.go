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
	Transfers            []HostTransfer
}

// HostTransfer is an owned, nonwrapping string-write effect. Data==nil means
// fill; otherwise Data contains bytes in the original iteration/write order.
type HostTransfer struct {
	Address      uint32
	Value, Count uint16
	Width        uint8
	Backward     bool
	Data         []byte
	Step         uint64
	CS, IP       uint16
}

// HostCodeRange is a half-open physical code interval. Providers still check
// the exact segmented entry and code bytes before proposing any effect.
type HostCodeRange struct {
	Lo, Hi uint32
	CS     uint16
}

func (m *Machine) ObserveHostBlocks(ranges []HostCodeRange, fn func() error) {
	m.onHostBlock = fn
	m.hostBlockRanges = append([]HostCodeRange(nil), ranges...)
	if fn == nil {
		m.hostBlockRanges = nil
	}
}
func (m *Machine) hostBlockMatch() bool {
	at := uint32(m.CPU.Seg[1])*16 + uint32(m.CPU.IP)
	for _, r := range m.hostBlockRanges {
		if (r.CS == 0 || r.CS == m.CPU.Seg[1]) && r.Lo <= at && at < r.Hi {
			return true
		}
	}
	return false
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
	if limit == 0 || m.hostPredicate || m.CycleClock || m.CPU.Halted || m.CPU.Flags&0x100 != 0 || m.irq0Pending || len(m.cbQueue) > 0 || len(m.keyQueue) > 0 || len(m.timedKeys) > 0 || len(m.wordWatches) > 0 || len(m.breaks) > 0 || m.Coverage != nil || m.TraceSegs || m.WatchDSOn || m.onRead != nil || m.accelCredit > 0 {
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
	last = 0
	wi := 0
	for _, s := range p.Transfers {
		if s.Step == 0 || s.Step <= last || s.Step > p.Instructions || s.Count == 0 || (s.Width != 1 && s.Width != 2) {
			return fmt.Errorf("machine: invalid host transfer")
		}
		last = s.Step
		lo, hi := int64(s.Address), int64(s.Address)+int64(s.Width)*int64(s.Count)
		if s.Backward {
			lo = int64(s.Address) - int64(s.Count-1)*int64(s.Width)
			hi = int64(s.Address) + int64(s.Width)
		}
		if lo < 0x500 || hi > vgaLo+64000 || lo < vgaLo && hi > vgaLo || s.Data != nil && len(s.Data) != int(s.Count)*int(s.Width) {
			return fmt.Errorf("machine: unsupported host transfer surface/source")
		}
		if hi > vgaLo && (m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13) {
			return fmt.Errorf("machine: unsupported host transfer mode")
		}
		if m.onWrite != nil && int64(m.watchLo) < hi && int64(m.watchHi) >= lo {
			return fmt.Errorf("machine: host transfer intersects observer")
		}
		for _, w := range m.writeWatches {
			if int64(w.lo) < hi && int64(w.hi) >= lo {
				return fmt.Errorf("machine: host transfer intersects guard")
			}
		}
		for wi < len(p.Writes) && p.Writes[wi].Step < s.Step {
			wi++
		}
		if wi < len(p.Writes) && p.Writes[wi].Step == s.Step {
			return fmt.Errorf("machine: ambiguous host write order")
		}
	}
	return nil
}
func (m *Machine) CommitHostWork(p *HostWorkPlan) error {
	return m.commitHostWork(p, false)
}
func (m *Machine) CommitHostBlock(p *HostWorkPlan) error {
	return m.commitHostWork(p, true)
}
func (m *Machine) commitHostWork(p *HostWorkPlan, block bool) error {
	if p == nil {
		return fmt.Errorf("machine: missing host plan")
	}
	owned := *p
	owned.Writes = append([]HostWrite(nil), p.Writes...)
	owned.Transfers = append([]HostTransfer(nil), p.Transfers...)
	for i := range owned.Transfers {
		if owned.Transfers[i].Data != nil {
			owned.Transfers[i].Data = append([]byte{}, owned.Transfers[i].Data...)
		}
	}
	p = &owned
	if err := m.ValidateHostWork(p); err != nil {
		return err
	}
	defer func() { m.hostEffect = nil; m.inVideoCommand = false }()
	for i, ti := 0, 0; i < len(p.Writes) || ti < len(p.Transfers); {
		if ti < len(p.Transfers) && (i == len(p.Writes) || p.Transfers[ti].Step < p.Writes[i].Step) {
			m.commitHostTransfer(p.Transfers[ti])
			ti++
			continue
		}
		n, c := m.hostVideoGroup(p.Writes[i:])
		if n > 0 {
			m.inVideoCommand = true
			for j := i; j < i+n; j++ {
				m.hostEffect = &p.Writes[j]
				m.Write8(p.Writes[j].Address, p.Writes[j].Value)
			}
			m.onVideoCommand(c)
			m.inVideoCommand = false
			i += n
		} else {
			m.hostEffect = &p.Writes[i]
			m.Write8(p.Writes[i].Address, p.Writes[i].Value)
			i++
		}
	}
	m.hostEffect = nil
	m.Steps += p.Instructions
	m.CPU.Cycles += p.Cycles
	m.CPU.ApplyHostRegisters(p.After.R, p.After.Seg, p.After.IP, p.After.Flags, p.LastCS, p.LastIP)
	m.insnCS, m.insnIP = p.LastCS, p.LastIP
	if block {
		m.HostBlockCommits++
		m.HostBlockInstructions += p.Instructions
	} else {
		m.HostCallCommits++
		m.HostCallInstructions += p.Instructions
	}
	return nil
}

func (m *Machine) commitHostTransfer(s HostTransfer) {
	context := HostWrite{Step: s.Step, CS: s.CS, IP: s.IP}
	m.hostEffect = &context
	video := s.Address >= vgaLo && m.nativeVideo && m.onVideoCommand != nil
	if video {
		m.inVideoCommand = true
	}
	for i := 0; i < int(s.Count); i++ {
		off := int(s.Address) + i*int(s.Width)
		if s.Backward {
			off = int(s.Address) - i*int(s.Width)
		}
		for b := 0; b < int(s.Width); b++ {
			value := byte(s.Value >> uint(8*b))
			if s.Data != nil {
				value = s.Data[i*int(s.Width)+b]
			}
			m.Write8(uint32(off+b), value)
		}
	}
	if video {
		kind := VideoFill
		if s.Data != nil {
			kind = VideoCopy
		}
		m.onVideoCommand(VideoCommand{Kind: kind, Step: m.Steps + s.Step, CS: s.CS, IP: s.IP, Offset: s.Address - vgaLo, Count: s.Count, Value: s.Value, Width: s.Width, Backward: s.Backward, Native: true, Host: true, Data: s.Data})
		m.inVideoCommand = false
	}
	m.hostEffect = nil
}
func (m *Machine) videoWriteContext() (step uint64, cs, ip uint16) {
	if w := m.hostEffect; w != nil {
		return m.Steps + w.Step, w.CS, w.IP
	}
	cs, ip = m.CPU.OpAddr()
	return m.Steps, cs, ip
}
