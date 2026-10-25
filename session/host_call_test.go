package session

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type hostObserver struct {
	recordingObserver
	failPrepare, failEntry, bad bool
	prepared                    int
}

func (o *hostObserver) PrepareHostCall(v StepView, budget uint64) (*machine.HostWorkPlan, error) {
	if v.IP() != 0x106 {
		return nil, nil
	}
	o.prepared++
	if o.failPrepare {
		panic("prepare failure")
	}
	after := machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP() + 2, v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.Read16(cpu.Linear(v.SS(), v.SP())), Flags: v.Flags()}
	p := &machine.HostWorkPlan{After: after, Instructions: 2, Cycles: 14, LastCS: v.CS(), LastIP: 0x107}
	if o.bad {
		p.Writes = []machine.HostWrite{{Address: 0x449, Step: 1}}
	}
	return p, nil
}
func (o *hostObserver) BeforeStep(v StepView) error {
	if o.failEntry && v.IP() == 0x106 {
		panic("entry failure")
	}
	return o.recordingObserver.BeforeStep(v)
}
func TestHostCallOwnerLifecycleAndFaults(t *testing.T) {
	if _, e := New(Config{NativeHostCalls: true, Observer: &recordingObserver{}}); e == nil {
		t.Fatal("missing provider accepted")
	}
	code := []byte{0xe8, 3, 0, 0xeb, 0xfe, 0x90, 0x90, 0xc3}
	for _, mode := range []string{"good", "prepare-fault", "entry-fault", "bad-plan"} {
		obs := &hostObserver{failPrepare: mode == "prepare-fault", failEntry: mode == "entry-fault", bad: mode == "bad-plan"}
		o := startSyntheticOwner(t, code)
		o.observer = obs
		o.nativeHostCalls = true
		bare := startSyntheticOwner(t, code)
		tick, e := runTurn(t, o, 21)
		if mode == "prepare-fault" || mode == "entry-fault" {
			if e == nil || tick.Reason != StopReasonObserverFault || o.viewOpen {
				t.Fatal("host callback failure escaped", mode, tick, e)
			}
		} else {
			other, oe := runTurn(t, bare, 21)
			if e != nil || oe != nil || tick.Steps != other.Steps || machineDigest(o) != machineDigest(bare) {
				t.Fatal("owner equivalence", mode, tick, e)
			}
			if mode == "good" && (tick.HostCalls != 1 || tick.HostInstructions != 2) {
				t.Fatal("missing host receipt")
			}
			if mode == "bad-plan" && tick.HostCalls != 0 {
				t.Fatal("invalid proposal accepted")
			}
		}
		if o.machine.CPU.HostCalls {
			t.Fatal("host callback retained after Advance")
		}
		n := obs.prepared
		o.machine.CPU.IP = 0x100
		o.machine.CPU.Halted = false
		_ = o.machine.Step()
		if obs.prepared != n {
			t.Fatal("host callback leaked")
		}
		o.Close()
		bare.Close()
	}
}
