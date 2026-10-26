package machine

// VGAStatusSnapshot is an owned prediction view of the existing status
// models, not an I/O read. Private parameters cannot expose mutable devices.
type VGAStatusSnapshot struct {
	Steps                 uint64
	readTicks, frameEvery uint64
	phase, valid          bool
}

func (m *Machine) VGAStatusSnapshot() (VGAStatusSnapshot, bool) {
	if m.Peek8(0x449) != 0x13 || m.planarOn || m.Unchained() || m.CRTCTiming && !m.Mode13Phase {
		return VGAStatusSnapshot{}, false
	}
	return VGAStatusSnapshot{Steps: m.Steps, readTicks: m.portTicks, frameEvery: m.VGAFrameEvery, phase: m.Mode13Phase, valid: true}, true
}

// At predicts readNumber (one based) at the relative instruction moment.
// It deliberately delegates phase calculation to the existing model.
func (s VGAStatusSnapshot) At(relativeStep, readNumber uint64) (byte, bool) {
	if !s.valid || relativeStep == 0 || readNumber == 0 || ^uint64(0)-s.Steps < relativeStep || ^uint64(0)-s.readTicks < readNumber {
		return 0, false
	}
	if s.phase {
		return mode13InputStatus(s.Steps+relativeStep, s.frameEvery), true
	}
	if ((s.readTicks+readNumber)>>4)&1 != 0 {
		return 9, true
	}
	return 0, true
}
