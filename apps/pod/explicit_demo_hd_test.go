package pod

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/xlate"
)

func TestExplicitDemoHDTransition(t *testing.T) {
	theme := os.Getenv("DOSGOLEM_POD_HD_DIR")
	capture := os.Getenv("DOSGOLEM_POD_HD_EXPLICIT_TRANSITION")
	if theme == "" || capture == "" {
		t.Skip("未提供明確按D的正常逐格來源")
	}
	read := func(path string) []byte {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	readPNG := func(path string) []byte {
		img, err := png.Decode(bytes.NewReader(read(path)))
		if err != nil || img.Bounds().Dx() != 960 || img.Bounds().Dy() != 600 {
			t.Fatal("正常截圖格式不合", err)
		}
		out := image.NewNRGBA(image.Rect(0, 0, 960, 600))
		draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
		return out.Pix
	}
	for _, direct := range []bool{false, true} {
		live := fixtureLive(t, theme)
		if direct {
			live.EnableDirect()
		}
		live.Flash = xlate.NewFlashAudit(320, 200, 8)
		baseline := xlate.NewFlashAudit(320, 200, 8)
		changedFrames, changedCells := map[int]bool{}, map[[2]int]bool{}
		for f := 6614; f <= 6624; f++ {
			stem := fmt.Sprintf("pod-f%06d", f)
			idx := read(filepath.Join(capture, "sources", stem+".idx"))
			raw := read(filepath.Join(capture, "sources", stem+".pal"))
			if len(idx) != 64000 || len(raw) != 768 {
				t.Fatal("來源長度不合")
			}
			var pal [256][3]uint8
			for i := range pal {
				copy(pal[i][:], raw[i*3:i*3+3])
			}
			original := append([]byte(nil), idx...)
			if direct {
				// Rehydrate captured indices; this tests composition, not a normal CPU path.
				copy(live.Direct.pending, idx)
			}
			live.Frame(idx, pal)
			live.AuditFlash(f-1, 3)
			out := live.Compose(3, true, nil)
			if !bytes.Equal(out, live.Compose(3, true, nil)) || !bytes.Equal(idx, original) {
				t.Fatal("重複合成不一致或修改原始來源")
			}
			captured := readPNG(filepath.Join(capture, "shots", stem+".png"))
			baseline.Observe(f-1, idx, vgaBaseRGBA(idx, pal, 3), captured, 3)
			// A full GUI screenshot has earlier hold history. Use the pinned old
			// renderer with the same fresh history for the exact area comparison.
			previous := readPNG(filepath.Join(capture, "controlled-original-replay", stem+".png"))
			for y := 8 * 3; y < 176*3; y++ {
				for x := 8 * 3; x < 176*3; x++ {
					o := (y*960 + x) * 4
					if bytes.Equal(out[o:o+4], previous[o:o+4]) {
						continue
					}
					changedFrames[f-1] = true
					changedCells[[2]int{x / 24, y / 24}] = true
					if f-1 != 6622 || x/24 != 7 || y/24 != 4 {
						t.Fatalf("修正影響已證實格之外 direct=%v frame=%d cell=%d,%d", direct, f-1, x/24, y/24)
					}
				}
			}
		}
		events := baseline.Events()
		if len(events) != 1 || events[0] != (xlate.FlashEvent{Frame: 6623, Shown: 6622, CX: 7, CY: 4}) {
			t.Fatalf("舊版負對照未重現：%+v", events)
		}
		if live.Flash.Count() != 0 || len(changedFrames) != 1 || len(changedCells) != 1 {
			t.Fatalf("四格修正未閉合：events=%+v frames=%v cells=%v", live.Flash.Events(), changedFrames, changedCells)
		}
		t.Logf("explicit demo transition direct=%v original flashes=1 fixed flashes=0 frames=11 changed frames=1 changed cells=1", direct)
	}
}
