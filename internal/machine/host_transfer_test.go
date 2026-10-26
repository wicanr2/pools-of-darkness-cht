package machine

import (
	"bytes"
	"testing"
)

func transferFixture() (*Machine, *HostWorkPlan) {
	m := New()
	m.SetVideoMode(0x13)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	p := &HostWorkPlan{After: HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, Flags: m.CPU.Flags}, Instructions: 3, Cycles: 6, Transfers: []HostTransfer{{Address: vgaLo + 64, Count: 3, Width: 2, Backward: true, Data: []byte{1, 2, 3, 4, 5, 6}, Step: 2, CS: 0x1234, IP: 8}}}
	return m, p
}
func TestHostTransfersGuardsAndOwnedSource(t *testing.T) {
	for _, change := range []func(*Machine, *HostWorkPlan){func(m *Machine, p *HostWorkPlan) { p.Transfers[0].Data = []byte{1} }, func(m *Machine, p *HostWorkPlan) { p.Transfers[0].Address = 0x501; p.Transfers[0].Count = 8 }, func(m *Machine, p *HostWorkPlan) { p.Transfers[0].Step = 4 }, func(m *Machine, p *HostWorkPlan) { p.Transfers[0].Width = 4 }, func(m *Machine, p *HostWorkPlan) { p.Transfers[0].Count = 0 }, func(m *Machine, p *HostWorkPlan) { p.Writes = []HostWrite{{Address: 0x800, Step: 2}} }, func(m *Machine, p *HostWorkPlan) {
		m.WatchWrite(vgaLo, vgaLo+100, func(*Machine, uint32, byte, byte) {})
	}} {
		m, p := transferFixture()
		before := append([]byte(nil), m.Mem...)
		change(m, p)
		if err := m.CommitHostBlock(p); err == nil || m.Steps != 0 || !bytes.Equal(before, m.Mem) {
			t.Fatal("bad transfer changed guest")
		}
	}
	m, p := transferFixture()
	m.SetNativeVideo(true)
	var values []byte
	m.ObserveVideoWrites(func(w VideoWrite) {
		values = append(values, w.Value)
		p.Transfers[0].Data[1] = 99
		p.Transfers[0].Address = 0x500
	})
	var command VideoCommand
	m.ObserveVideoCommands(func(c VideoCommand) { command = c })
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(values, []byte{1, 2, 3, 4, 5, 6}) || !bytes.Equal(command.Data, values) || command.Count != 3 || command.Width != 2 || !command.Backward || !command.Host {
		t.Fatal("proposal source leaked into commit")
	}
	if m.Mem[vgaLo+64] != 1 || m.Mem[vgaLo+65] != 2 || m.Mem[vgaLo+60] != 5 {
		t.Fatal("word order")
	}
}
