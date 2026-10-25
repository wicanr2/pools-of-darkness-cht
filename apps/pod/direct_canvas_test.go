package pod

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/xlate"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectCanvasCommandsAndFrameBoundary(t *testing.T) {
	d := NewDirectCanvas()
	c := machine.VideoCommand{Kind: machine.VideoCopy, Offset: 10, Count: 2, Width: 2, Backward: true, Native: true, Data: []byte{1, 2, 3, 4}}
	if e := d.Apply(c); e != nil {
		t.Fatal(e)
	}
	c.Data[0] = 99
	want := make([]byte, 64000)
	copy(want[8:12], []byte{3, 4, 1, 2})
	if e := d.Frame(want); e != nil {
		t.Fatal(e)
	}
	if e := d.Apply(machine.VideoCommand{Kind: machine.VideoFill, Offset: 8, Count: 2, Width: 2, Value: 0x1234, Native: true}); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(d.committed, want) {
		t.Fatal("post-retrace command changed presented frame")
	}
	before := append([]byte(nil), d.pending...)
	for _, bad := range []machine.VideoCommand{{Kind: machine.VideoCopy, Offset: 63999, Count: 2, Width: 1, Data: []byte{1, 2}}, {Kind: machine.VideoCopy, Count: 2, Width: 1, Data: []byte{1}}, {Kind: machine.VideoFill, Offset: 0, Count: 2, Width: 2, Backward: true}, {Kind: 99, Count: 1, Width: 1}} {
		if d.Apply(bad) == nil || !bytes.Equal(before, d.pending) {
			t.Fatal("malformed command mutated canvas")
		}
	}
	if d.Frame(want) == nil || d.Stats().Mismatches != 1 {
		t.Fatal("missing negative frame check")
	}
	d.Reset()
	if e := d.Frame(want); e != nil {
		t.Fatal(e)
	}
	if d.Stats().BootstrapFrames != 1 {
		t.Fatal("snapshot bootstrap missing")
	}
}

// Genuine indexed frames are a diagnostic oracle for RGB; normal input paths
// are verified separately by the frontend receipts.
func TestDirectRGBActualFixtures(t *testing.T) {
	root := os.Getenv("DOSGOLEM_POD_DIRECT_ROOT")
	if root == "" {
		t.Skip("未提供真實來源")
	}
	theme := filepath.Join(root, "workplace/hd/theme")
	pairs := []string{"workplace/probe/g9-hd/b25-play/hd/source", "workplace/probe/g9-hd/b25-play/hd/sources/pod-f003650", "workplace/probe/g9-hd/wall-play-v3/hd/source", "workplace/probe/g9-hd/intro-play-v1/subtitles/final"}
	for _, name := range []string{"title-001", "title-002", "title-003"} {
		pairs = append(pairs, "workplace/hd/inventory/"+name)
	}
	for _, name := range pairs {
		t.Run(name, func(t *testing.T) {
			idx, e := os.ReadFile(filepath.Join(root, name+".idx"))
			if e != nil {
				t.Fatal(e)
			}
			p, e := os.ReadFile(filepath.Join(root, name+".pal"))
			if e != nil || len(idx) != 64000 || len(p) != 768 {
				t.Fatal("invalid actual fixture", e)
			}
			var pal [256][3]uint8
			for c := range pal {
				copy(pal[c][:], p[c*3:c*3+3])
			}
			a, b := fixtureLive(t, theme), fixtureLive(t, theme)
			b.EnableDirect()
			for frame := 0; frame < 5; frame++ {
				cur := append([]byte(nil), idx...)
				if frame > 1 {
					for y := 40; y < 48; y++ {
						for x := 40; x < 48; x++ {
							cur[y*320+x] ^= byte(frame)
						}
					}
				}
				a.Frame(cur, pal)
				b.VideoCommand(machine.VideoCommand{Kind: machine.VideoCopy, Count: 64000, Width: 1, Data: cur, Native: true})
				b.Frame(cur, pal)
				for _, s := range []int{3, 2, 3} {
					for _, overlay := range []bool{true, false} {
						want, got := a.Compose(s, overlay, nil), b.Compose(s, overlay, nil)
						if !bytes.Equal(want, got) {
							for n := range want {
								if got[n] != want[n] {
									t.Fatalf("frame%d scale%d overlay%v byte%d got%d want%d", frame, s, overlay, n, got[n], want[n])
								}
							}
						}
					}
				}
			}
			if b.Direct.Stats().LegacyRGBFrames != 0 {
				t.Fatal("official opaque theme fell back")
			}
		})
	}
}

func TestBattleHoldDeferredHistory(t *testing.T) {
	var a, b battleHold
	var eligible [441]bool
	eligible[0] = true
	idx := make([]byte, 64000)
	var pal [256][3]uint8
	pal[1] = [3]byte{3, 7, 11}
	for frame := 0; frame < 6; frame++ {
		a.commit()
		b.commit()
		idx[8*320+8] = byte(frame % 2)
		source := xlate.NewIndexedSource(idx, pal, 320, 200, 3)
		base := make([]byte, 960*600*4)
		source.FillOriginal(base)
		old := append([]byte(nil), base...)
		sparse := make([]byte, len(base))
		if frame == 0 {
			for y := 24; y < 48; y++ {
				for x := 24; x < 48; x++ {
					o := (y*960 + x) * 4
					copy(old[o:o+4], []byte{21, 22, 23, 255})
					copy(sparse[o:o+4], old[o:o+4])
				}
			}
		}
		a.apply(idx, base, old, 1, eligible)
		b.applyDeferred(idx, sparse, 1, eligible, source)
		source.FillOriginal(sparse)
		if !bytes.Equal(old, sparse) {
			t.Fatal("battle history mismatch", frame)
		}
	}
}
