package session

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// 這一份釘 `docs/spec/432-session-boot-com-queue`：.COM 疊底、監督佇列、疊底常駐縮回。
// 零值欄位等於 235 的行為由 boot_original_test.go 釘住，這裡不重複。

// residentCOM：MOV DX,8000h；INT 27h（保留 8000h bytes ＝ 800h 段後常駐）。
var residentCOM = []byte{0xBA, 0x00, 0x80, 0xCD, 0x27}

// exitCOM：INT 20h。
var exitCOM = []byte{0xCD, 0x20}

// bootChain 以 residentCOM 疊底、佇列 NEXT.COM 開機，跑到程式結束，回佇列程式的 PSP。
func bootChain(t *testing.T, shrink bool) uint16 {
	t.Helper()
	save := t.TempDir()
	if err := os.WriteFile(filepath.Join(save, "NEXT.COM"), exitCOM, 0o600); err != nil {
		t.Fatal(err)
	}
	o, err := New(Config{InitialScale: host.OutputScale2, InitialLayout: sessionLayout(1, host.OutputScale2, false)})
	if err != nil {
		t.Fatal(err)
	}
	in := BootInput{EXE: residentCOM, ExpectedEXESHA256: sha256.Sum256(residentCOM), SaveRoot: save,
		Format: BootFormatCOM, Queue: []BootQueued{{Name: "NEXT.COM"}}, RootTSRShrink: shrink}
	if _, err := o.BootOriginal(in); err != nil {
		t.Fatal(err)
	}
	view, err := o.View()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := o.Deliver(CapturedUpdate{StartedPanel: view.Panel, Layout: view.Layout, SourceGeneration: view.SourceGeneration}); err != nil {
		t.Fatal(err)
	}
	r, err := o.Advance(10000)
	if err != nil || r.Reason != StopReasonProgramStopped {
		t.Fatalf("%+v %v", r, err)
	}
	for _, e := range o.dos.ExecLog {
		if e.Base == "NEXT.COM" || strings.EqualFold(e.Name, "NEXT.COM") {
			return e.PSP
		}
	}
	t.Fatalf("佇列程式沒有載入：ExecLog=%+v Missing=%v", o.dos.ExecLog, o.dos.Missing)
	return 0
}

// TestBootCOMQueueShrinkReclaimsTheGap：開縮回時佇列程式緊接在常駐區之後（PSP＋800h＋1）。
func TestBootCOMQueueShrinkReclaimsTheGap(t *testing.T) {
	want := uint16(machine.PSPSeg) + 0x800 + 1
	if got := bootChain(t, true); got != want {
		t.Fatalf("佇列程式 PSP=%04X，預期 %04X", got, want)
	}
}

// TestBootCOMQueueNoShrinkKeepsTheSegment：不開縮回時照 197 §5.1，佇列程式落在頂層 .COM 名義上擁有的整段之後。
func TestBootCOMQueueNoShrinkKeepsTheSegment(t *testing.T) {
	if got, floor := bootChain(t, false), uint16(machine.PSPSeg)+0x1000; got <= floor {
		t.Fatalf("佇列程式 PSP=%04X，預期在 %04X 之後", got, floor)
	}
}

// TestBootCOMQueueRejectsWithoutFailing：432 §2 第 1 條的每一種拒絕都不動 phase、不計步，之後仍可正常開機。
func TestBootCOMQueueRejectsWithoutFailing(t *testing.T) {
	save := t.TempDir()
	sum := sha256.Sum256(residentCOM)
	base := BootInput{EXE: residentCOM, ExpectedEXESHA256: sum, SaveRoot: save, Format: BootFormatCOM}
	big := make([]byte, maxBootCOMSize+1)
	for _, tc := range []struct {
		name string
		mod  func(*BootInput)
	}{
		{"未知格式", func(b *BootInput) { b.Format = 7 }},
		{"COM 太大", func(b *BootInput) { b.EXE, b.ExpectedEXESHA256 = big, sha256.Sum256(big) }},
		{"佇列太長", func(b *BootInput) { b.Queue = make([]BootQueued, maxBootQueue+1) }},
		{"空名稱", func(b *BootInput) { b.Queue = []BootQueued{{Name: ""}} }},
		{"名稱太長", func(b *BootInput) { b.Queue = []BootQueued{{Name: "ABCDEFGHI.COM"}} }},
		{"名稱含路徑", func(b *BootInput) { b.Queue = []BootQueued{{Name: `A\B.EXE`}} }},
		{"名稱含磁碟代號", func(b *BootInput) { b.Queue = []BootQueued{{Name: "C:B.EXE"}} }},
		{"名稱含控制字元", func(b *BootInput) { b.Queue = []BootQueued{{Name: "A\x01.EXE"}} }},
		{"參數太長", func(b *BootInput) { b.Queue = []BootQueued{{Name: "A.EXE", Args: strings.Repeat("x", 127)}} }},
		{"參數含 CR", func(b *BootInput) { b.Queue = []BootQueued{{Name: "A.EXE", Args: "a\rb"}} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := bootTestOwner(t)
			before := o.machine.Steps
			in := base
			tc.mod(&in)
			if _, err := o.BootOriginal(in); err == nil {
				t.Fatal("want rejection")
			}
			if o.Status().Phase != PhaseBooting || o.machine.Steps != before {
				t.Fatalf("拒絕後 phase=%d steps %d→%d", o.Status().Phase, before, o.machine.Steps)
			}
			if _, err := o.BootOriginal(base); err != nil {
				t.Fatalf("拒絕後應可正常開機：%v", err)
			}
		})
	}
}
