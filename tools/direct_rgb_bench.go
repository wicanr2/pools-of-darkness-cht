// Renderer-only paired measurement on existing genuine indexed frames.
// Build in dosgolem/workplace; this is not a normal player-path receipt.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/apps/pod"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func main() {
	root := flag.String("root", "", "project")
	out := flag.String("out", "", "new report")
	flag.Parse()
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("report exists")
	}
	type row struct {
		Name, SourceSHA256, PaletteSHA256, RGBSHA256 string
		LegacyNS, DirectNS                           []int64
		LegacyMedianNS, DirectMedianNS               int64
		OriginalPixels, SelectedPixels               uint64
	}
	rows := []row{}
	for _, name := range []string{"workplace/hd/inventory/title-001", "workplace/probe/g9-hd/b25-play/hd/sources/pod-f003650", "workplace/probe/g9-hd/b25-play/hd/source", "workplace/probe/g9-hd/wall-play-v3/hd/source"} {
		idx, p := read(filepath.Join(*root, name+".idx")), read(filepath.Join(*root, name+".pal"))
		if len(idx) != 64000 || len(p) != 768 {
			panic("source dimensions")
		}
		var pal [256][3]uint8
		for i := range pal {
			copy(pal[i][:], p[i*3:i*3+3])
		}
		old, direct := pod.NewLive(nil, nil, nil), pod.NewLive(nil, nil, nil)
		must(old.AddHD(filepath.Join(*root, "workplace/hd/theme")))
		must(direct.AddHD(filepath.Join(*root, "workplace/hd/theme")))
		direct.EnableDirect()
		old.Frame(idx, pal)
		direct.Frame(idx, pal)
		want, got := old.Compose(3, true, nil), direct.Compose(3, true, nil)
		if !bytes.Equal(want, got) {
			panic("independent legacy RGB mismatch")
		}
		r := row{Name: name, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(idx)), PaletteSHA256: fmt.Sprintf("%x", sha256.Sum256(p)), RGBSHA256: fmt.Sprintf("%x", sha256.Sum256(got))}
		before := direct.Direct.Stats()
		direct.Compose(3, true, nil)
		after := direct.Direct.Stats()
		r.OriginalPixels = after.OriginalRGBPixels - before.OriginalRGBPixels
		r.SelectedPixels = after.SelectedRGBPixels - before.SelectedRGBPixels
		if r.OriginalPixels+r.SelectedPixels != 576000 || after.LegacyRGBFrames != 0 {
			panic("invalid deferred selection counts")
		}
		if name == "workplace/hd/inventory/title-001" && r.OriginalPixels != 0 {
			panic("full HD title painted original RGB")
		}
		measure := func(l *pod.Live) int64 {
			start := time.Now()
			for i := 0; i < 12; i++ {
				if len(l.Compose(3, true, nil)) != len(want) {
					panic("render dimensions")
				}
			}
			return time.Since(start).Nanoseconds() / 12
		}
		for i := 0; i < 7; i++ {
			var a, b int64
			if i%2 == 0 {
				a = measure(old)
				b = measure(direct)
			} else {
				b = measure(direct)
				a = measure(old)
			}
			r.LegacyNS = append(r.LegacyNS, a)
			r.DirectNS = append(r.DirectNS, b)
		}
		median := func(v []int64) int64 {
			c := append([]int64(nil), v...)
			sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
			return c[len(c)/2]
		}
		r.LegacyMedianNS = median(r.LegacyNS)
		r.DirectMedianNS = median(r.DirectNS)
		rows = append(rows, r)
	}
	result := struct {
		Scope string
		Rows  []row
	}{"實際索引來源的CPU RGB合成局部量測；七組交錯、每組12次；不含執行遊戲、觀測器、flashaudit、面板、GPU或視窗；不外推FPS", rows}
	b, e := json.MarshalIndent(result, "", "  ")
	must(e)
	must(os.WriteFile(*out, append(b, '\n'), 0600))
}
