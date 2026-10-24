// Disposable source-only timing probe; not linked into pod-play.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/bootroot"
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/session"
	"os"
	"path/filepath"
)

type event struct {
	Update  int
	Retrace int
	Steps   uint64
	SHA     string
	File    string
}
type observer struct {
	out             string
	update, retrace int
	steps           uint64
	indexed         []byte
	palette         [256][3]uint8
	events          []event
	last            string
	seen            map[string]string
}

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
	// RGB fingerprint avoids irrelevant palette slots outside the subtitle band.
	rgb := make([]byte, 0, 320*32*3)
	for _, v := range o.indexed[168*320:] {
		rgb = append(rgb, o.palette[v][:]...)
	}
	// Keep invisible indexed transitions too: an all-black DAC can hide ink.
	identity := append([]byte{}, o.indexed[168*320:]...)
	used := [256]bool{}
	for _, v := range o.indexed[168*320:] {
		used[v] = true
	}
	for v, yes := range used {
		if yes {
			identity = append(identity, byte(v))
			identity = append(identity, o.palette[v][:]...)
		}
	}
	sum := sha256.Sum256(identity)
	h := hex.EncodeToString(sum[:])
	if h == o.last {
		return
	}
	o.last = h
	name, ok := o.seen[h]
	if !ok {
		if len(o.seen) >= 1000 {
			panic("subtitle source cap")
		}
		name = fmt.Sprintf("source-%03d", len(o.seen))
		o.seen[h] = name
		must(os.WriteFile(filepath.Join(o.out, name+".idx"), o.indexed, 0600))
		flat := make([]byte, 0, 768)
		for _, c := range o.palette {
			flat = append(flat, c[:]...)
		}
		must(os.WriteFile(filepath.Join(o.out, name+".pal"), flat, 0600))
	}
	o.events = append(o.events, event{o.update, o.retrace, o.steps, h, name})
	fmt.Printf("update=%d step=%d source=%s\n", o.update, o.steps, name)
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	orig := flag.String("orig", "", "read-only original")
	seed := flag.String("seed", "", "G2 seed")
	out := flag.String("out", "", "empty receipt")
	frames := flag.Int("frames", 11000, "host updates")
	flag.Parse()
	must(os.MkdirAll(*out, 0700))
	required := []bootroot.RequiredFile{}
	for name, h := range map[string]string{"ADLIB.COM": "7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286", "CONTROL.EXE": "60e53bcf4266fb4e68e4fc22990159d1c1af978e9d10068708cfd11d49935366"} {
		b, e := hex.DecodeString(h)
		must(e)
		var a [32]byte
		copy(a[:], b)
		required = append(required, bootroot.RequiredFile{Name: name, SHA256: a})
	}
	root, e := bootroot.Prepare(bootroot.BootRootInput{OriginalRoot: *orig, SaveRoot: filepath.Join(*out, "save"), Required: required})
	must(e)
	for _, n := range []string{"pool4.cfg", "snd.com"} {
		b, e := os.ReadFile(filepath.Join(*seed, n))
		must(e)
		must(os.WriteFile(filepath.Join(root.SaveRoot, n), b, 0600))
	}
	exe, e := os.ReadFile(filepath.Join(*orig, "ADLIB.COM"))
	must(e)
	s := host.OutputScale2
	layout := host.MouseLayout{Epoch: 1, Scale: s, ChromeHeight: 18 * int(s), Canvas: host.Canvas{Width: 320, Height: 200}, FrameWidth: 320 * int(s), FrameHeight: 218 * int(s)}
	o := &observer{out: *out, seen: map[string]string{}}
	owner, e := session.New(session.Config{InitialScale: s, InitialLayout: layout, AdLib: true, ClockPercent: 50, Observer: o})
	must(e)
	defer owner.Close()
	_, e = owner.BootOriginal(session.BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: root.SaveRoot, Format: session.BootFormatCOM, Queue: []session.BootQueued{{Name: "CONTROL.EXE"}}, RootTSRShrink: true, ProgramPath: `C:\POD\SND.COM`})
	must(e)
	budget := session.InstructionBudget(machine.ScaleSteps(machine.DefaultVGAFrameEvery, 50) * 70 / 60)
	for i := 1; i <= *frames; i++ {
		o.update = i
		v, e := owner.View()
		must(e)
		_, e = owner.Deliver(session.CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration})
		must(e)
		tick, e := owner.Advance(budget)
		must(e)
		o.presented()
		if tick.RawError != nil {
			panic(tick.RawError)
		}
		if i%500 == 0 {
			fmt.Printf("progress %d\n", i)
		}
	}
	d, e := owner.Digest()
	must(e)
	b, e := json.MarshalIndent(struct {
		Frames int
		Budget session.InstructionBudget
		Events []event
		Digest session.StateDigest
	}{*frames, budget, o.events, d}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "timing.json"), append(b, '\n'), 0600))
}
