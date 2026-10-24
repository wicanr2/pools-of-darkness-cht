// Sealed owner private boot: LoadEXE, then Install, then Running.
//
// The original tree (read-only input) stays with the composition root; the
// owner only verifies the executable bytes and the save root handed to it.
// Observer installation, multi-layer composition and save/load are out of
// scope (Buck #16/#18 DRAFT).

package session

import (
	"github.com/wicanr2/dosgolem/internal/fsaccess"

	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
)

// maxBootEXESize is a corrupt-input guard, not a version assertion.
// START.EXE measures 67,619 bytes; loadability is decided by LoadEXE.
const maxBootEXESize = 4 << 20

// maxBootCOMSize 與 machine.loadCOMAt 的上限相同（映像從 0100h 起、堆疊從 FFFEh 往下）；
// 驗證先擋，載入器就不會在開機步驟裡失敗（`docs/spec/432-session-boot-com-queue` §2）。
const maxBootCOMSize = 0xFEFE

// maxBootQueue 是開機佇列的程式數上限（432 §2）。
const maxBootQueue = 8

// BootFormat 是開機映像的格式（432）。零值是 MZ，即 235 的原行為。
type BootFormat uint8

const (
	BootFormatEXE BootFormat = iota // machine.LoadEXE（預設）
	BootFormatCOM                   // machine.LoadCOM
)

// BootQueued 是開機後排入監督佇列的一支程式（`docs/spec/009-exec` §4）。
// Name 照 basename 在 save root 下解析，Args 寫進它的命令列尾。
type BootQueued struct{ Name, Args string }

// BootInput carries the executable bytes and the writable save root for one
// private boot.  Paths to the read-only original tree never enter the owner.
// Format、Queue、RootTSRShrink 是 432 新增，零值等於 235 的現行行為。
type BootInput struct {
	EXE               []byte
	ExpectedEXESHA256 [32]byte
	SaveRoot          string
	Format            BootFormat
	Queue             []BootQueued
	RootTSRShrink     bool   // `docs/spec/431-root-tsr-shrink-option` §2
	ProgramPath       string // `docs/spec/433-session-boot-program-path`；空值保留原行為
}

func validBootProgramPath(path string) bool {
	if path == "" {
		return true
	}
	if len(path) < 4 || len(path) > 127 || path[1:3] != `:\` ||
		!((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) {
		return false
	}
	for _, c := range []byte(path) {
		if c < 0x20 || c >= 0x7f {
			return false
		}
	}
	for _, part := range strings.Split(path[3:], `\`) {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, `/:*?"<>|`) {
			return false
		}
	}
	return true
}

// validBootQueueName 照 432 §2：非空、至多 12 bytes、不含路徑分隔、磁碟代號與控制字元。
func validBootQueueName(name string) bool {
	if name == "" || len(name) > 12 {
		return false
	}
	for i := 0; i < len(name); i++ {
		if c := name[i]; c < 0x20 || c == 0x7F || c == '/' || c == '\\' || c == ':' {
			return false
		}
	}
	return true
}

// validBootQueueArgs 照 432 §2：至多 126 bytes（命令列尾上限），不含 CR、NUL。
func validBootQueueArgs(args string) bool {
	if len(args) > 126 {
		return false
	}
	for i := 0; i < len(args); i++ {
		if args[i] == '\r' || args[i] == 0 {
			return false
		}
	}
	return true
}

// BootReceipt proves one completed boot without exposing machine state.
type BootReceipt struct {
	Phase     Phase
	EXESHA256 [32]byte
}

// BootOriginal loads the executable, installs DOS and enters Running.
// Steps 1-3 below reject without touching phase; only step 4 may fail the
// owner.  The caller's EXE slice is never retained.
func (o *Owner) BootOriginal(input BootInput) (BootReceipt, error) {
	if o == nil {
		return BootReceipt{}, errors.New("session: Owner 不得為 nil")
	}
	if o.phase != PhaseBooting {
		return BootReceipt{}, fmt.Errorf("session: phase %d 不可開機", o.phase)
	}
	if len(input.EXE) == 0 || len(input.EXE) > maxBootEXESize {
		return BootReceipt{}, fmt.Errorf("session: EXE 長度 %d 不合法", len(input.EXE))
	}
	switch input.Format {
	case BootFormatEXE:
	case BootFormatCOM:
		if len(input.EXE) > maxBootCOMSize {
			return BootReceipt{}, fmt.Errorf("session: COM 長度 %d 超過 %#x", len(input.EXE), maxBootCOMSize)
		}
	default:
		return BootReceipt{}, fmt.Errorf("session: 開機格式 %d 不合法", input.Format)
	}
	if len(input.Queue) > maxBootQueue {
		return BootReceipt{}, fmt.Errorf("session: 開機佇列 %d 支超過 %d", len(input.Queue), maxBootQueue)
	}
	if !validBootProgramPath(input.ProgramPath) {
		return BootReceipt{}, errors.New("session: DOS 程式路徑不合法")
	}
	for i, q := range input.Queue {
		if !validBootQueueName(q.Name) {
			return BootReceipt{}, fmt.Errorf("session: 開機佇列第 %d 支名稱 %q 不合法", i, q.Name)
		}
		if !validBootQueueArgs(q.Args) {
			return BootReceipt{}, fmt.Errorf("session: 開機佇列第 %d 支參數不合法", i)
		}
	}
	var zero [32]byte
	if input.ExpectedEXESHA256 == zero {
		return BootReceipt{}, errors.New("session: 缺少 EXE SHA-256 期望值")
	}
	if input.SaveRoot == "" {
		return BootReceipt{}, errors.New("session: 缺少 save root")
	}
	info, err := os.Lstat(input.SaveRoot)
	if err != nil {
		return BootReceipt{}, fmt.Errorf("session: save root 不可用: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return BootReceipt{}, errors.New("session: save root 不可為 symlink")
	}
	if !info.IsDir() {
		return BootReceipt{}, errors.New("session: save root 不是目錄")
	}
	if err := fsaccess.Writable(input.SaveRoot); err != nil {
		return BootReceipt{}, fmt.Errorf("session: save root 不可寫: %v", err)
	}
	exe := append([]byte(nil), input.EXE...)
	sum := sha256.Sum256(exe)
	if sum != input.ExpectedEXESHA256 {
		return BootReceipt{}, errors.New("session: EXE SHA-256 不符")
	}
	o.dos.Root = input.SaveRoot
	// 存檔根本身就是已驗證可寫的複製樹；設成 Scratch，遊戲寫檔才會真的落地
	// （dosgolem 規格 009：Scratch 為空時寫入只記帳）。見規格 235 補充。
	o.dos.Scratch = input.SaveRoot
	if input.ProgramPath != "" {
		o.dos.SetProgramPath(input.ProgramPath)
	}
	load := o.machine.LoadEXE
	if input.Format == BootFormatCOM {
		load = o.machine.LoadCOM
	}
	if err := load(exe); err != nil {
		return BootReceipt{}, o.fail(fmt.Errorf("session: 載入 EXE: %w", err))
	}
	// 432 §2：縮回選項與佇列只在載入成功後才設；dos.Install 不動佇列。
	o.dos.RootTSRShrink = input.RootTSRShrink
	for _, q := range input.Queue {
		o.dos.Enqueue(q.Name, q.Args)
	}
	// dos.Install 無回錯值：簽名上沒有錯分支可標；它只在 LoadEXE 成功後跑。
	if err := o.startLoadedMachine(); err != nil {
		return BootReceipt{}, err
	}
	return BootReceipt{Phase: PhaseRunning, EXESHA256: sum}, nil
}
