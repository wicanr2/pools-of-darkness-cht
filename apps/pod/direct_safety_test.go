package pod

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectTransparentThemeFallback(t *testing.T) {
	dir := t.TempDir()
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for n := 3; n < len(img.Pix); n += 4 {
		img.Pix[n] = 255
	}
	save := func() {
		var b bytes.Buffer
		if e := png.Encode(&b, img); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(filepath.Join(dir, "one.png"), b.Bytes(), 0600); e != nil {
			t.Fatal(e)
		}
	}
	save()
	if e := os.WriteFile(filepath.Join(dir, "theme.json"), []byte(`{"items":[{"png":"one.png"}]}`), 0600); e != nil {
		t.Fatal(e)
	}
	if !deferredThemeSafe(dir) {
		t.Fatal("opaque rejected")
	}
	img.Pix[3] = 127
	save()
	if deferredThemeSafe(dir) {
		t.Fatal("partial alpha accepted")
	}
	img.Pix[3] = 0
	save()
	if deferredThemeSafe(dir) {
		t.Fatal("transparent source accepted")
	}
	l := NewLive(nil, nil, nil)
	l.EnableDirect()
	l.deferredUnsafe = true
	l.Frame(make([]byte, 64000), [256][3]uint8{})
	if !bytes.Equal(l.Compose(3, true, nil), l.composeLegacy(3, true, nil)) || l.Direct.Stats().LegacyRGBFrames != 1 {
		t.Fatal("legacy fallback changed output")
	}
}
