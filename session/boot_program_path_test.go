package session

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/host"
)

// 規格 433：真實 DOS 開檔服務的正負對照，避免只比對同名欄位。
func TestBootProgramPathOpenAbsoluteSave(t *testing.T) {
	for _, path := range []string{"", `C:\POD\SND.COM`} {
		t.Run(path, func(t *testing.T) {
			save := t.TempDir()
			if err := os.Mkdir(filepath.Join(save, "SAVE"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(save, "SAVE", "PROOF.DAT"), []byte("proof"), 0o600); err != nil {
				t.Fatal(err)
			}
			// mov ax,3d00h; mov dx,0110h; int 21h; int 20h; padding; ASCIIZ filename.
			exe := append([]byte{0xB8, 0x00, 0x3D, 0xBA, 0x10, 0x01, 0xCD, 0x21, 0xCD, 0x20}, make([]byte, 6)...)
			exe = append(exe, []byte("C:\\POD\\SAVE\\PROOF.DAT\x00")...)
			o, err := New(Config{InitialScale: host.OutputScale2, InitialLayout: sessionLayout(1, host.OutputScale2, false)})
			if err != nil {
				t.Fatal(err)
			}
			in := BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: save, Format: BootFormatCOM, ProgramPath: path}
			if _, err := o.BootOriginal(in); err != nil {
				t.Fatal(err)
			}
			if path != "" && (o.dos.Dir != "POD" || o.dos.RootDir != "POD" || !bytes.Contains(o.machine.Mem, []byte(path+"\x00"))) {
				t.Fatal("目錄與環境路徑未同步")
			}
			v, err := o.View()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := o.Deliver(CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration}); err != nil {
				t.Fatal(err)
			}
			if _, err := o.Advance(1000); err != nil {
				t.Fatal(err)
			}
			opened := false
			for _, name := range o.dos.Opened {
				if strings.EqualFold(name, "PROOF.DAT") { // DOS.Opened 的既有契約只記 basename
					opened = true
				}
			}
			if opened != (path != "") {
				t.Fatalf("path=%q opened=%v missing=%v", path, opened, o.dos.Missing)
			}
		})
	}
}

func TestBootProgramPathValidation(t *testing.T) {
	for _, path := range []string{`POD\SND.COM`, `/tmp/x`, `C:REL.COM`, `C:\..\X.COM`, `C:\.\X.COM`, `C:\\X.COM`, `C:\`, `C:\X?.COM`, "C:\\X\x00.COM", "C:\\遊戲.COM", `C:\` + strings.Repeat("x", 125)} {
		t.Run(path, func(t *testing.T) {
			o := bootTestOwner(t)
			exe := []byte{0xCD, 0x20}
			in := BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: t.TempDir(), Format: BootFormatCOM, ProgramPath: path}
			if _, err := o.BootOriginal(in); err == nil {
				t.Fatal("應拒絕非法程式路徑")
			}
			if o.Status().Phase != PhaseBooting || o.machine.Steps != 0 || o.machine.ProgramPath != "" {
				t.Fatal("拒絕時改了狀態")
			}
			in.ProgramPath = `C:\SND.COM`
			if _, err := o.BootOriginal(in); err != nil {
				t.Fatal(err)
			}
			if o.dos.Dir != "" || o.dos.RootDir != "" {
				t.Fatal("根目錄程式仍沿用舊目錄")
			}
		})
	}
}
