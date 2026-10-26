// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func bitmap51CallSeed(t *testing.T, width, rows uint16, planar, compressed bool) *view {
	v := bitmap51RowSeed(t, width, rows, planar, compressed)
	frame := uint32(v.SS())*16 + uint32(v.SP())
	v.s.R[cpu.BP] = v.Read16(frame + 6)
	v.s.R[cpu.DI] = v.Read16(frame)
	v.s.R[cpu.SI] = v.Read16(frame + 2)
	v.s.Seg[cpu.DS] = v.Read16(frame + 4)
	v.s.R[cpu.SP] += 8
	v.s.IP = 0x51cc
	file := uint32(0x80100)
	copy(v.mem[file:file+12], []byte("FORM\x00\x00\x00\x00PBM "))
	if planar {
		copy(v.mem[file+8:file+12], []byte("ILBM"))
	}
	copy(v.mem[file+12:file+20], []byte("BMHD\x00\x00\x00\x14"))
	putBE := func(a uint32, value uint16) { v.mem[a] = byte(value >> 8); v.mem[a+1] = byte(value) }
	for i := uint32(20); i < 40; i++ {
		v.mem[file+i] = 0
	}
	putBE(file+20, width)
	putBE(file+22, rows)
	v.mem[file+28] = 8
	v.mem[file+33] = 0x7b
	if compressed {
		v.mem[file+30] = 1
	}
	copy(v.mem[file+40:file+48], []byte("BODY\x00\x00\x00\x00"))
	at := file + 48
	stride := ((width + 15) / 16) * 2
	rowBytes := int(stride * 8)
	for row := uint16(0); row < rows; row++ {
		data := make([]byte, rowBytes)
		for i := range data {
			data[i] = byte(i*37 + int(width) + int(row)*19)
		}
		if compressed {
			for len(data) > 0 {
				n := len(data)
				if n > 128 {
					n = 128
				}
				v.mem[at] = byte(n - 1)
				at++
				copy(v.mem[at:at+uint32(n)], data[:n])
				at += uint32(n)
				data = data[n:]
			}
		} else {
			copy(v.mem[at:at+uint32(rowBytes)], data)
			at += uint32(rowBytes)
		}
	}
	putBE(file+46, uint16(at-file-48))
	return v
}
func TestBitmap51Whole(t *testing.T) {
	cases, algorithmic, fixed, metadata := 0, 0, 0, 0
	for _, width := range []uint16{1, 9, 17, 160, 300} {
		for _, rows := range []uint16{1, 3} {
			for _, planar := range []bool{false, true} {
				for _, compressed := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						for _, budget := range []uint64{1, 7, 27, 128, 1024} {
							v := bitmap51CallSeed(t, width, rows, planar, compressed)
							v.dos = dos
							a, b := callMachine(v), callMachine(v)
							runCall(t, a, v, false)
							used := 0
							for b.Steps < 1000000 {
								if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
									break
								}
								w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
								if p := Plan(w, budget); p != nil {
									if err := b.CommitHostBlock(p); err != nil {
										t.Fatal(err)
									}
									if p.Algorithmic {
										algorithmic++
										used++
									} else {
										fixed++
										if bitmap51FixedRaster(p.Writes, p.Transfers) {
											t.Fatal("fixed bitmap raster escaped")
										}
									}
								} else {
									if w.CS() == 0x48a1 && (w.IP() == 0x53d8 || w.IP() == 0x53f6 || w.IP() == 0x53fa) {
										t.Fatalf("CPU bitmap raster width%d rows%d planar%v compressed%v dos%v budget%d IP%x", width, rows, planar, compressed, dos, budget, w.IP())
									}
									if err := b.Step(); err != nil {
										t.Fatal(err)
									}
									metadata++
								}
							}
							if used == 0 {
								t.Fatal("no typed bitmap plan")
							}
							var sa, sb bytes.Buffer
							if err := a.SaveState(&sa); err != nil {
								t.Fatal(err)
							}
							if err := b.SaveState(&sb); err != nil {
								t.Fatal(err)
							}
							if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
								t.Fatalf("bitmap full machine differs width%d rows%d planar%v compressed%v dos%v budget%d steps%d/%d", width, rows, planar, compressed, dos, budget, a.Steps, b.Steps)
							}
							cases++
						}
					}
				}
			}
		}
	}
	t.Log("bitmap51 whole original comparisons", cases, "algorithmic plans", algorithmic, "fixed metadata plans", fixed, "CPU metadata steps", metadata)
}
