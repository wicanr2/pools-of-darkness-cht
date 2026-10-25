package cpu

// HostCallBus is optional and only runs after a completed 16-bit CALL.
// The bus owner must preserve the callee effects and all event boundaries.
type HostCallBus interface{ AfterGuestCall() }

func (c *CPU) afterGuestCall() {
	if c.HostCalls && !c.operand32 && !c.addr32 && c.fault == nil {
		if b, ok := c.Bus.(HostCallBus); ok {
			b.AfterGuestCall()
		}
	}
}

// ApplyHostRegisters records an equivalent completed 16-bit callee. The
// original instruction identity is for semantic notifications, not execution.
func (c *CPU) ApplyHostRegisters(r [8]uint16, seg [4]uint16, ip, flags, lastCS, lastIP uint16) {
	c.R, c.Seg, c.IP, c.Flags = r, seg, ip, flags
	c.opCS, c.opIP = lastCS, lastIP
	c.segOverride = noSegOverride
	c.repPrefix, c.lock, c.operand32, c.addr32, c.ext386 = 0, false, false, false, false
}
