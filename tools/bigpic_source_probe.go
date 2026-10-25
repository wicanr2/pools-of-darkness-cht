// Disposable source-only normal-DEMO observer. Never linked into pod-play.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/bootroot"
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/session"
)

type source struct {
	Key              string `json:"key"`
	File             string `json:"file"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	IndexedSHA       string `json:"indexed_sha256"`
	PaletteSHA       string `json:"palette_sha256"`
	ref, pal, anchor []byte
	ax, ay           int
}
type hit struct {
	Update       int
	Steps        uint64
	Key          string
	Rect         [4]int
	UsedRGBEqual bool
	Capture      string
}
type observer struct {
	out             string
	update, retrace int
	steps           uint64
	indexed         []byte
	palette         [256][3]uint8
	sources         []source
	seen            map[string]bool
	hits            []hit
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func hash(b []byte) string                              { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func read(p string) []byte                              { b, e := os.ReadFile(p); must(e); return b }
func (o *observer) BeforeStep(v session.StepView) error { o.steps = v.Steps(); return nil }
func (o *observer) VideoWrite(machine.VideoWrite)       {}
func (o *observer) Frame(idx []byte, pal [256][3]uint8) {
	o.retrace++
	o.indexed = idx
	o.palette = pal
}
func (o *observer) presented() {
	if len(o.indexed) != 64000 {
		return
	}
	for _, s := range o.sources {
		for from := 0; from+len(s.anchor) <= len(o.indexed); {
			at := bytes.Index(o.indexed[from:], s.anchor)
			if at < 0 {
				break
			}
			at += from
			from = at + 1
			x, y := at%320-s.ax, at/320-s.ay
			if x < 0 || y < 0 || x+s.Width > 320 || y+s.Height > 200 {
				continue
			}
			equal := true
			for row := 0; row < s.Height; row++ {
				if !bytes.Equal(o.indexed[(y+row)*320+x:(y+row)*320+x+s.Width], s.ref[row*s.Width:(row+1)*s.Width]) {
					equal = false
					break
				}
			}
			if !equal {
				continue
			}
			rgb := true
			for _, c := range s.ref {
				if !bytes.Equal(o.palette[c][:], s.pal[int(c)*3:int(c)*3+3]) {
					rgb = false
					break
				}
			}
			identity := fmt.Sprintf("%s:%d:%d:%t", s.Key, x, y, rgb)
			if o.seen[identity] {
				continue
			}
			o.seen[identity] = true
			name := fmt.Sprintf("f%06d-%s", o.update, s.Key)
			must(os.WriteFile(filepath.Join(o.out, name+".idx"), o.indexed, 0600))
			flat := make([]byte, 0, 768)
			for _, c := range o.palette {
				flat = append(flat, c[:]...)
			}
			must(os.WriteFile(filepath.Join(o.out, name+".pal"), flat, 0600))
			h := hit{o.update, o.steps, s.Key, [4]int{x, y, s.Width, s.Height}, rgb, name}
			o.hits = append(o.hits, h)
			fmt.Printf("BIGPIC %+v\n", h)
		}
	}
}
func main() {
	orig := flag.String("orig", "", "read-only original")
	seed := flag.String("seed", "", "G2 seed")
	out := flag.String("out", "", "empty output")
	inventory := flag.String("inventory", "", "read-only decoded sources")
	frames := flag.Int("frames", 26000, "bounded normal DEMO updates")
	flag.Parse()
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("output already exists")
	}
	must(os.Mkdir(*out, 0700))
	var catalog struct {
		Items []source `json:"items"`
	}
	must(json.Unmarshal(read(filepath.Join(*inventory, "inventory.json")), &catalog))
	o := &observer{out: *out, seen: map[string]bool{}, hits: []hit{}}
	for _, s := range catalog.Items {
		if s.File != "BIGPIC1.DAX" {
			continue
		}
		s.ref = read(filepath.Join(*inventory, s.Key+".idx"))
		s.pal = read(filepath.Join(*inventory, s.Key+".pal"))
		if hash(s.ref) != s.IndexedSHA || hash(s.pal) != s.PaletteSHA || len(s.ref) != s.Width*s.Height || len(s.pal) != 768 {
			panic("source changed")
		}
		best := -1
		for y := 0; y < s.Height; y++ {
			for x := 0; x+32 <= s.Width; x += 8 {
				var used [256]bool
				score := 0
				for _, v := range s.ref[y*s.Width+x : y*s.Width+x+32] {
					if !used[v] {
						score++
						used[v] = true
					}
				}
				if score > best {
					best = score
					s.ax = x
					s.ay = y
				}
			}
		}
		s.anchor = s.ref[s.ay*s.Width+s.ax : s.ay*s.Width+s.ax+32]
		o.sources = append(o.sources, s)
	}
	if len(o.sources) != 7 {
		panic("need seven BIGPIC sources")
	}
	required := []bootroot.RequiredFile{}
	for n, h := range map[string]string{"ADLIB.COM": "7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286", "CONTROL.EXE": "60e53bcf4266fb4e68e4fc22990159d1c1af978e9d10068708cfd11d49935366"} {
		b, e := hex.DecodeString(h)
		must(e)
		var a [32]byte
		copy(a[:], b)
		required = append(required, bootroot.RequiredFile{Name: n, SHA256: a})
	}
	root, e := bootroot.Prepare(bootroot.BootRootInput{OriginalRoot: *orig, SaveRoot: filepath.Join(*out, "save"), Required: required})
	must(e)
	for _, n := range []string{"pool4.cfg", "snd.com"} {
		must(os.WriteFile(filepath.Join(root.SaveRoot, n), read(filepath.Join(*seed, n)), 0600))
	}
	exe := read(filepath.Join(*orig, "ADLIB.COM"))
	scale := host.OutputScale2
	layout := host.MouseLayout{Epoch: 1, Scale: scale, ChromeHeight: 18 * int(scale), Canvas: host.Canvas{Width: 320, Height: 200}, FrameWidth: 320 * int(scale), FrameHeight: 218 * int(scale)}
	owner, e := session.New(session.Config{InitialScale: scale, InitialLayout: layout, AdLib: true, ClockPercent: 50, Observer: o})
	must(e)
	defer owner.Close()
	_, e = owner.BootOriginal(session.BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: root.SaveRoot, Format: session.BootFormatCOM, Queue: []session.BootQueued{{Name: "CONTROL.EXE"}}, RootTSRShrink: true, ProgramPath: `C:\POD\SND.COM`})
	must(e)
	budget := session.InstructionBudget(machine.ScaleSteps(machine.DefaultVGAFrameEvery, 50) * 70 / 60)
	completed := 0
	for i := 1; i <= *frames; i++ {
		o.update = i
		v, e := owner.View()
		must(e)
		input := session.CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration}
		if i == 900 || i == 1800 {
			input.BIOSKeys = []dos.Key{{Scan: 0x39, ASCII: ' '}}
		}
		_, e = owner.Deliver(input)
		must(e)
		tick, e := owner.Advance(budget)
		must(e)
		o.presented()
		completed = i
		if tick.RawError != nil {
			panic(tick.RawError)
		}
		if tick.Phase == session.PhaseStopped {
			break
		}
		if i%1000 == 0 {
			fmt.Printf("progress %d\n", i)
		}
	}
	digest, e := owner.Digest()
	must(e)
	report := struct {
		Scope  string
		Frames int
		Budget session.InstructionBudget
		Hits   []hit
		Digest session.StateDigest
	}{"source-only normal cold DEMO; BIOS spaces900/1800; no guest state injection", completed, budget, o.hits, digest}
	b, e := json.MarshalIndent(report, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "report.json"), append(b, '\n'), 0600))
}
