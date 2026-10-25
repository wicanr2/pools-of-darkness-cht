package cpu

// StringCommand is an optional atomic MOVS/STOS replacement. A refusal must
// leave the bus untouched; CPU state and instruction accounting remain here.
type StringCommand struct {
	SourceSeg, SourceOff, DestSeg, DestOff uint16
	Count, Value                           uint16
	Width                                  uint8
	Copy, Backward                         bool
}

type StringCommandBus interface{ WriteStringCommand(StringCommand) bool }

func (c *CPU) nativeString(op uint8, delta uint16) bool {
	if !c.NativeStrings || (op != 0xA4 && op != 0xA5 && op != 0xAA && op != 0xAB) {
		return false
	}
	bus, ok := c.Bus.(StringCommandBus)
	if !ok {
		return false
	}
	n := uint16(1)
	if c.repPrefix != 0 {
		n = c.R[CX]
	}
	width := uint8(1)
	if op&1 != 0 {
		width = 2
	}
	copying := op == 0xA4 || op == 0xA5
	if !bus.WriteStringCommand(StringCommand{SourceSeg: c.dataSeg(DS), SourceOff: c.R[SI], DestSeg: c.Seg[ES], DestOff: c.R[DI], Count: n, Value: c.R[AX], Width: width, Copy: copying, Backward: c.Flag(DF)}) {
		return false
	}
	c.charge(cycString * int(n))
	if c.DOSBoxCost {
		c.Cycles += uint64(n)
	}
	c.R[DI] += delta * n
	if copying {
		c.R[SI] += delta * n
	}
	if c.repPrefix != 0 {
		c.R[CX] = 0
	}
	return true
}
