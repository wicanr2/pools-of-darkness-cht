package session

import (
	"errors"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type blockObserver struct {
	recordingObserver
	mode     string
	prepared int
	held     StepView
}

func (o *blockObserver) HostBlockRanges() []machine.HostCodeRange {
	base := uint32(machine.PSPSeg)*16 + 0x100
	return []machine.HostCodeRange{{Lo: base, Hi: base + 2}}
}
func (o *blockObserver) PrepareHostBlock(v StepView, budget uint64) (*machine.HostWorkPlan, error) {
	o.prepared++
	o.held = v
	copyBuf := make([]byte, 2)
	if !v.CopyRAM(uint32(v.CS())*16+uint32(v.IP()), copyBuf) {
		return nil, errors.New("plain RAM copy rejected")
	}
	copyBuf[0] ^= 255
	if v.Read8(uint32(v.CS())*16+uint32(v.IP())) == copyBuf[0] {
		return nil, errors.New("plain RAM alias escaped")
	}
	if o.mode == "panic" {
		panic("block")
	}
	if o.mode == "error" {
		return nil, errors.New("block")
	}
	if budget < 2 || v.IP() != 0x100 {
		return nil, nil
	}
	p := &machine.HostWorkPlan{After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: 0x102, Flags: v.Flags()}, Instructions: 2, Cycles: 4, LastCS: v.CS(), LastIP: 0x101}
	if o.mode == "invalid" {
		p.Writes = []machine.HostWrite{{Address: 0x449, Step: 1}}
	}
	return p, nil
}
func TestHostBlockOwnerLifecycle(t *testing.T) {
	if _, e := New(Config{NativeHostBlocks: true, Observer: &recordingObserver{}}); e == nil {
		t.Fatal("missing provider accepted")
	}
	code := []byte{0x90, 0x90, 0xeb, 0xfe}
	for _, mode := range []string{"good", "invalid", "panic", "error"} {
		obs := &blockObserver{mode: mode}
		o := startSyntheticOwner(t, code)
		o.observer = obs
		o.nativeHostBlocks = true
		bare := startSyntheticOwner(t, code)
		tick, err := runTurn(t, o, 11)
		if mode == "panic" || mode == "error" {
			if err == nil || tick.Reason != StopReasonObserverFault || o.viewOpen || tick.MachineStepsAfter != 0 {
				t.Fatal("fault escaped", mode, tick, err)
			}
		} else {
			other, e := runTurn(t, bare, 11)
			if err != nil || e != nil || tick.Steps != other.Steps || machineDigest(o) != machineDigest(bare) {
				t.Fatal("block state differs", mode, tick, err)
			}
			if mode == "good" && (tick.HostBlocks != 1 || tick.BlockInstructions != 2 || tick.HostCalls != 0) {
				t.Fatal("block receipt", tick, obs.prepared)
			}
			if mode == "invalid" && tick.HostBlocks != 0 {
				t.Fatal("invalid accepted")
			}
		}
		n := obs.prepared
		o.machine.CPU.IP = 0x100
		o.machine.CPU.Halted = false
		_ = o.machine.Step()
		if n != obs.prepared {
			t.Fatal("provider leaked")
		}
		o.Close()
		bare.Close()
	}
}
