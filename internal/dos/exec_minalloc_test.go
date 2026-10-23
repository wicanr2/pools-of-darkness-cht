package dos

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// 這一份釘 `docs/spec/430`：MZ 的 e_minalloc 是程式自己的未初始化資料與堆疊。
// 反面不會報錯：EXEC 把子程式疊在父程式的堆疊上，子程式拿到的命令列是自己的程式碼。

// minallocParentMZ 造一支父程式：映像 2 段（hlt），e_minalloc＝0x40，
// 堆疊段緊接映像（SS＝2），SP＝0x200，整個堆疊都在 e_minalloc 那一段裡。
func minallocParentMZ() []byte {
	const hdrPar = 2
	body := make([]byte, 32)
	body[0] = 0xF4 // hlt
	out := make([]byte, hdrPar*16+len(body))
	h := out[:32]
	h[0], h[1] = 'M', 'Z'
	total := len(out)
	binary.LittleEndian.PutUint16(h[2:], uint16(total%512))
	binary.LittleEndian.PutUint16(h[4:], uint16((total+511)/512))
	binary.LittleEndian.PutUint16(h[8:], hdrPar)
	binary.LittleEndian.PutUint16(h[10:], 0x40) // e_minalloc
	binary.LittleEndian.PutUint16(h[12:], 0x40) // e_maxalloc
	binary.LittleEndian.PutUint16(h[14:], 2)    // SS（相對載入段）
	binary.LittleEndian.PutUint16(h[16:], 0x200)
	copy(out[hdrPar*16:], body)
	return out
}

// TestLoadEXEReservesMinAlloc：FreeSeg 跳過 e_minalloc。
func TestLoadEXEReservesMinAlloc(t *testing.T) {
	m := machine.New()
	if err := m.LoadEXE(minallocParentMZ()); err != nil {
		t.Fatal(err)
	}
	if want := uint16(machine.LoadSeg + 2 + 0x40 + 1); m.FreeSeg != want {
		t.Errorf("FreeSeg ＝ %04X，預期 %04X（載入段＋映像 2 段＋e_minalloc 40h＋1）", m.FreeSeg, want)
	}
	if n, err := machine.MZLoadParags(minallocParentMZ()); err != nil || n != 2+0x40 {
		t.Errorf("MZLoadParags ＝ %d, %v，預期 %d", n, err, 2+0x40)
	}
	if n, err := machine.MZImageParags(minallocParentMZ()); err != nil || n != 2 {
		t.Errorf("MZImageParags ＝ %d, %v，預期 2（只算映像）", n, err)
	}
}

// TestExecDoesNotLandOnParentStack：父程式把 EXEC 參數區與命令列放在自己的堆疊上
// （Turbo Pascal 的 Exec 就是這樣，《Pools of Darkness》`CONTROL.EXE`），
// 子程式要落在 e_minalloc 之後，拿到的命令列與父程式給的相同。
func TestExecDoesNotLandOnParentStack(t *testing.T) {
	m := machine.New()
	if err := m.LoadEXE(minallocParentMZ()); err != nil {
		t.Fatal(err)
	}
	d := New(m, t.TempDir())
	d.Install()
	if err := os.WriteFile(filepath.Join(d.Root, "CHILD.EXE"), buildChildMZ(), 0o644); err != nil {
		t.Fatal(err)
	}

	ss := m.CPU.Seg[cpu.SS] // 載入段＋2
	m.WriteBytes(cpu.Addr(ss, 0x0C0), append([]byte("child.exe"), 0))
	pb := cpu.Addr(ss, 0x100)
	m.Write16(pb, 0)        // env ＝ 繼承
	m.Write16(pb+2, 0x0120) // 尾巴 → SS:0120
	m.Write16(pb+4, ss)
	m.WriteBytes(cpu.Addr(ss, 0x120), []byte{3, '9', ' ', '1', 0x0D})
	m.Write16(pb+6, 0x0140) // FCB1 → SS:0140
	m.Write16(pb+8, ss)
	m.Write16(pb+10, 0x0150) // FCB2 → SS:0150
	m.Write16(pb+12, ss)

	c := m.CPU
	c.Seg[cpu.DS], c.R[cpu.DX] = ss, 0x0C0
	c.Seg[cpu.ES], c.R[cpu.BX] = ss, 0x100
	call(m, d, 0x21, 0x4B00)
	if c.Flags&cpu.CF != 0 {
		t.Fatalf("EXEC 失敗，AX=%04X Missing=%v", c.R[cpu.AX], d.Missing)
	}

	parentEnd := uint16(machine.LoadSeg + 2 + 0x40)
	if d.curPSP < parentEnd {
		t.Errorf("子程式 PSP ＝ %04X，落在父程式 e_minalloc 裡（父程式到 %04X）", d.curPSP, parentEnd)
	}
	tail := uint32(d.curPSP)*16 + 0x80
	got := []byte{m.Read8(tail), m.Read8(tail + 1), m.Read8(tail + 2), m.Read8(tail + 3)}
	if want := []byte{3, '9', ' ', '1'}; string(got) != string(want) {
		t.Errorf("子程式命令列 ＝ % X，預期 % X（被子程式自己的映像蓋掉）", got, want)
	}
}
