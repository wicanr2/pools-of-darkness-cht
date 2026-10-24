package pod

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func localizedFixture(t *testing.T) (string, string) {
	t.Helper()
	dir, capture := os.Getenv("DOSGOLEM_POD_HD_TITLE_DIR"), os.Getenv("DOSGOLEM_POD_HD_TITLE_CAPTURE")
	if dir == "" || capture == "" {
		t.Skip("未提供開場主題及正常冷啟動來源")
	}
	return dir, capture
}

func readLocalizedPNG(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewNRGBA(image.Rect(0, 0, 960, 600))
	draw.Draw(out, out.Bounds(), im, im.Bounds().Min, draw.Src)
	return out.Pix
}

func readLocalizedCapture(t *testing.T, capture, name string) ([]byte, [256][3]uint8) {
	t.Helper()
	idx, err := os.ReadFile(filepath.Join(capture, name+".idx"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(capture, name+".pal"))
	if err != nil || len(data) != 768 {
		t.Fatal("來源色盤不合", err)
	}
	var pal [256][3]uint8
	for i := range pal {
		copy(pal[i][:], data[i*3:i*3+3])
	}
	return idx, pal
}

func TestLocalizedTitleNormalSourceAndLanguage(t *testing.T) {
	dir, capture := localizedFixture(t)
	for key, name := range map[string]string{"title-001": "pod-f001000", "title-002": "pod-f001100", "title-003": "pod-f001200"} {
		live := fixtureLive(t, dir)
		idx, pal := readLocalizedCapture(t, capture, name)
		original := append([]byte(nil), idx...)
		live.Frame(idx, pal)
		for _, code := range []string{DefaultLang, EnglishLang, "ja", DefaultLang} {
			live.SetLang(&LangPack{Code: code, UI: &Translator{}})
			folder := code
			if code == "ja" {
				folder = "base"
			}
			want := readLocalizedPNG(t, filepath.Join(dir, "localized", folder, key+".png"))
			if got := live.Compose(3, true, nil); !bytes.Equal(got, want) {
				t.Fatalf("%s/%s開場成品或語系殘留不合", key, code)
			}
			if !bytes.Equal(live.Compose(3, false, nil), vgaBaseRGBA(idx, pal, 3)) {
				t.Fatal("F12未回原版")
			}
			if !bytes.Equal(live.Compose(2, true, nil), vgaBaseRGBA(idx, pal, 2)) {
				t.Fatal("2倍仍有開場HD")
			}
		}
		if !bytes.Equal(idx, original) {
			t.Fatal("開場改了原始色號")
		}
		changed := pal
		changed[200][0] ^= 1
		live.Frame(idx, changed)
		if !bytes.Equal(live.Compose(3, true, nil), vgaBaseRGBA(idx, changed, 3)) {
			t.Fatal("未知色盤仍留HD或舊語系")
		}
	}
}

func TestLocalizedTitleCreditsBottomAndSourceMismatch(t *testing.T) {
	dir, capture := localizedFixture(t)
	live := fixtureLive(t, dir)
	idx, pal := readLocalizedCapture(t, capture, "pod-f001200")
	live.Frame(idx, pal)
	want := live.Compose(3, true, nil)
	changed := append([]byte(nil), idx...)
	clear(changed[192*320:])
	live.Frame(changed, pal)
	got := live.Compose(3, true, nil)
	base := vgaBaseRGBA(changed, pal, 3)
	if !bytes.Equal(got[:576*960*4], want[:576*960*4]) || !bytes.Equal(got[576*960*4:], base[576*960*4:]) {
		t.Fatal("credits底列改寫未逐格撤除")
	}
	changed = append([]byte(nil), idx...)
	changed[100*320+100] ^= 255
	live.Frame(changed, pal)
	got = live.Compose(3, true, nil)
	base = vgaBaseRGBA(changed, pal, 3)
	for y := 288; y < 312; y++ {
		o := (y*960 + 288) * 4
		if !bytes.Equal(got[o:o+24*4], base[o:o+24*4]) {
			t.Fatal("來源失配小格沒有回退")
		}
	}
}
