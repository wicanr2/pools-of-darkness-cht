// Spec025/#49: source-driven Go variable-width glyph pixels.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

func glyph5aa4Compare8Flags(before uint16, a, b byte) uint16 {
	r := a - b
	f := before &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if a < b {
		f |= cpu.CF
	}
	if (a^b^r)&16 != 0 {
		f |= cpu.AF
	}
	if r == 0 {
		f |= cpu.ZF
	}
	if r&128 != 0 {
		f |= cpu.SF
	}
	if ((a^b)&(a^r))&128 != 0 {
		f |= cpu.OF
	}
	if bits.OnesCount8(r)%2 == 0 {
		f |= cpu.PF
	}
	return f
}
func glyph5aa4DecrementFlags(before, value uint16) uint16 {
	return ellipseSubFlags(before, value, 1, 0)&^uint16(cpu.CF) | before&cpu.CF
}

// Row traversal is driven by glyph dimensions and the font's color table.
// Source IPs below label ABI notices only, not an opcode interpreter.
func planGlyph5aa4Rows(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || (v.IP() != 0x5c00 && v.IP() != 0x5c26 && v.IP() != 0x5c2f) {
		return nil
	}
	f, ok := readGlyph5aa4Frame(v)
	if !ok {
		return nil
	}
	row := uint16(uint32(f.y+f.row)*320 + uint32(f.x))
	p := &machine.HostWorkPlan{Algorithmic: true, After: boundedMemoryRegisters(v), LastCS: 0x48a1}
	remaining := f.height - f.row
	add := func(n, cycles, iterations uint64, last, next uint16) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + iterations
		} else {
			p.Cycles += cycles
		}
		p.LastIP = last
		p.After.IP = next
	}
	if !f.mapped {
		if v.IP() != 0x5c00 || v.BX() != f.fontOff || v.CX() != f.width || v.DI() != row || uint32(v.SI())+f.fontBase != f.data {
			return nil
		}
		for remaining > 0 {
			n := uint64(7)
			cycles := uint64(30) + 5*uint64(f.width)
			last, next := uint16(0x5c12), uint16(0x5c00)
			if remaining == 1 {
				n++
				cycles += 2
				last, next = 0x5c14, 0x5c3f
			}
			if n > budget-p.Instructions {
				break
			}
			data := make([]byte, int(f.width))
			copyView(v, f.data, data)
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(row), Data: data, Width: 1, Count: f.width, Step: p.Instructions + 2, CS: 0x48a1, IP: 0x5c02})
			imageWord(p, 0x11ff, remaining-1, p.Instructions+6, 0x5c0d)
			flags := copyAddFlags(p.After.Flags, row, 320)
			p.After.Flags = glyph5aa4DecrementFlags(flags, remaining)
			p.After.R[cpu.AX] = row
			p.After.R[cpu.CX] = f.width
			p.After.R[cpu.SI] += f.width
			p.After.R[cpu.DI] = row + 320
			add(n, cycles, uint64(f.width), last, next)
			remaining--
			row += 320
			f.data += uint32(f.width)
		}
	} else {
		if v.BX() != f.fontOff+8 || v.BP() != row {
			return nil
		}
		col := uint16(0)
		if v.IP() == 0x5c26 {
			if v.CX() == 0 || v.CX() > f.width {
				return nil
			}
			col = f.width - v.CX()
			if v.SI() != uint16(f.data-f.fontBase+uint32(col)) || v.DI() != row+col {
				return nil
			}
		} else {
			col = f.width
			if v.CX() != 0 || v.SI() != uint16(f.data-f.fontBase+uint32(f.width)) {
				return nil
			}
			raw := v.Read8(f.data + uint32(f.width) - 1)
			lastColor := v.Read8(f.fontBase + uint32(f.fontOff) + 8 + uint32(raw))
			di := row + f.width
			if lastColor == 255 {
				di--
			}
			if v.DI() != di || byte(v.AX()) != lastColor {
				return nil
			}
		}
		for remaining > 0 {
			for col < f.width {
				value := v.Read8(f.data + uint32(col))
				color := v.Read8(f.fontBase + uint32(f.fontOff) + 8 + uint32(value))
				n, cycles, last, next := uint64(6), uint64(22), uint16(0x5c2d), uint16(0x5c26)
				transparent := color == 255
				if transparent {
					cycles = 24
					n = 7
					last = 0x5c25
					if col+1 == f.width {
						cycles = 27
						n = 6
						last, next = 0x5c23, 0x5c2f
					}
				} else if col+1 == f.width {
					next = 0x5c2f
				}
				if n > budget-p.Instructions {
					break
				}
				p.After.R[cpu.AX] = p.After.R[cpu.AX]&0xff00 | uint16(color)
				p.After.R[cpu.SI]++
				p.After.Flags = glyph5aa4Compare8Flags(p.After.Flags, color, 255)
				if !transparent {
					p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(row+col), Value: uint16(color), Width: 1, Count: 1, Step: p.Instructions + 5, CS: 0x48a1, IP: 0x5c2c})
					p.After.R[cpu.DI]++
					p.After.R[cpu.CX]--
				} else {
					p.After.Flags = glyph5aa4DecrementFlags(p.After.Flags, p.After.R[cpu.CX])
					p.After.R[cpu.CX]--
					if col+1 < f.width {
						before := p.After.Flags
						p.After.Flags = copyAddFlags(before, p.After.R[cpu.DI], 1)&^uint16(cpu.CF) | before&cpu.CF
						p.After.R[cpu.DI]++
					}
				}
				iterations := uint64(2)
				if transparent {
					iterations = 1
				}
				add(n, cycles, iterations, last, next)
				col++
			}
			if col < f.width || 5 > budget-p.Instructions {
				break
			}
			flags := copyAddFlags(p.After.Flags, row, 320)
			p.After.Flags = glyph5aa4DecrementFlags(flags, remaining)
			imageWord(p, 0x11ff, remaining-1, p.Instructions+4, 0x5c38)
			p.After.R[cpu.BP] = row + 320
			p.After.R[cpu.DI] = row + 320
			p.After.R[cpu.CX] = f.width
			cycles, last, next := uint64(25), uint16(0x5c3d), uint16(0x5c26)
			if remaining == 1 {
				cycles = 20
				next = 0x5c3f
			}
			add(5, cycles, 0, last, next)
			remaining--
			row += 320
			f.data += uint32(f.width)
			col = 0
		}
	}
	if p.Instructions == 0 {
		return nil
	}
	return p
}
