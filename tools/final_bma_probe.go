// 研究035：原版FINAL播放器直接入口；不宣稱正常通關，僅讀取BMA來源。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/apps/pod/drawblocks"
	"github.com/wicanr2/dosgolem/bootroot"
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/session"
	"os"
	"path/filepath"
)

type imageCall struct {
	Path                     string
	Update                   int
	Steps                    uint64
	CS, IP, SS, SP           uint16
	Head                     uint32
	Size                     uint32
	Frame, Slot, X, Y, Flags uint16
	SHA                      string
}
type event struct {
	ImageCall int
	Update    int
	Retrace   int
	Steps     uint64
	SHA       string
	File      string
}
type observer struct {
	fast             bool
	hostPlans        int
	hostEntryVisits  int
	hostInstructions uint64
	captured         bool
	lastCapture      uint64
	out              string
	update, retrace  int
	steps            uint64
	indexed          []byte
	palette          [256][3]uint8
	images           []imageCall
	events           []event
	last             string
	seen             map[string]string
}

func (o *observer) capture(v session.StepView, path string) {
	o.steps = v.Steps()
	if v.IP() == 0x6733 {
		if o.captured && o.lastCapture == v.Steps() {
			return
		}
		o.captured = true
		o.lastCapture = v.Steps()
		stack := uint32(v.SS())*16 + uint32(v.SP())
		arg := func(off uint32) uint16 { return v.Read16(stack + off) }
		head := uint32(arg(14))*16 + uint32(arg(12))
		size := uint32(v.Read16(head)) | uint32(v.Read16(head+2))<<16
		if size >= 10 && size <= 65535 && head >= 0x500 && head+size <= 0xa0000 {
			code := make([]byte, 255)
			if !v.CopyRAM(uint32(v.CS())*16+0x6733, code) {
				panic("code read")
			}
			if fmt.Sprintf("%x", sha256.Sum256(code)) != "d9ebc163f106be514229472090f2497771df8e36ab5aeca61681f758d32baa54" {
				return
			}
			data := make([]byte, size)
			if !v.CopyRAM(head, data) {
				panic("source read")
			}
			if len(o.images) >= 4096 {
				panic("image call cap")
			}
			o.images = append(o.images, imageCall{path, o.update, v.Steps(), v.CS(), v.IP(), v.SS(), v.SP(), head, size, arg(16), arg(4), arg(6), arg(8), arg(10), fmt.Sprintf("%x", sha256.Sum256(data))})
		}
	}
}
func (o *observer) BeforeStep(v session.StepView) error { o.capture(v, "BeforeStep"); return nil }
func (o *observer) HostBlockRanges() []machine.HostCodeRange {
	if o.fast {
		return pod.DrawBlockRanges()
	}
	return nil
}
func (o *observer) PrepareHostBlock(v session.StepView, budget uint64) (*machine.HostWorkPlan, error) {
	if !o.fast {
		return nil, nil
	}
	o.capture(v, "PrepareHostBlock")
	if v.IP() == 0x6733 {
		o.hostEntryVisits++
	}
	p := drawblocks.Plan(v, budget)
	if p != nil {
		o.hostPlans++
		o.hostInstructions += p.Instructions
	}
	return p, nil
}
func (o *observer) VideoWrite(machine.VideoWrite) {}
func (o *observer) Frame(idx []byte, pal [256][3]uint8) {
	o.retrace++
	o.indexed = idx
	o.palette = pal
}
func (o *observer) presented() {
	if len(o.indexed) != 64000 {
		return
	}
	// 只核對上方168列的完整索引與使用RGB，排除下方字幕和未使用色盤槽。
	// 全黑色盤也保留索引轉換，不將不可見圖樣誤當已清除。
	identity := append([]byte{}, o.indexed[:168*320]...)
	used := [256]bool{}
	for _, v := range o.indexed[:168*320] {
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
		if len(o.seen) >= 6000 {
			panic("intro-art source cap")
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
	o.events = append(o.events, event{len(o.images), o.update, o.retrace, o.steps, h, name})
	fmt.Printf("update=%d step=%d source=%s\n", o.update, o.steps, name)
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	fast := flag.Bool("go-draw", false, "exercise current Go drawing blocks")
	orig := flag.String("orig", "", "read-only original")
	seed := flag.String("seed", "", "G2 seed")
	out := flag.String("out", "", "empty receipt")
	frames := flag.Int("frames", 5100, "host updates")
	flag.Parse()
	if *out == "" || *orig == "" || *seed == "" || *frames < 1 || *frames > 6000 {
		panic("need original, seed, new output and bounded frame count")
	}
	must(os.Mkdir(*out, 0700))
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
	o := &observer{fast: *fast, out: *out, seen: map[string]string{}}
	owner, e := session.New(session.Config{InitialScale: s, InitialLayout: layout, AdLib: true, ClockPercent: 50, Observer: o, NativeHostBlocks: *fast})
	must(e)
	defer owner.Close()
	_, e = owner.BootOriginal(session.BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: root.SaveRoot, Format: session.BootFormatCOM, Queue: []session.BootQueued{{Name: "VACE.EXE", Args: "4 FINAL 0 1"}}, RootTSRShrink: true, ProgramPath: `C:\POD\SND.COM`})
	must(e)
	budget := session.InstructionBudget(machine.ScaleSteps(machine.DefaultVGAFrameEvery, 50) * 70 / 60)
	var committed, committedInstructions uint64
	for i := 1; i <= *frames; i++ {
		o.update = i
		v, e := owner.View()
		must(e)
		_, e = owner.Deliver(session.CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration})
		must(e)
		tick, e := owner.Advance(budget)
		must(e)
		committed += tick.HostBlocks
		committedInstructions += tick.BlockInstructions
		o.presented()
		if tick.RawError != nil {
			panic(tick.RawError)
		}
		if owner.Status().Phase != session.PhaseRunning {
			break
		}
		if i%500 == 0 {
			fmt.Printf("progress %d\n", i)
		}
	}
	d, e := owner.Digest()
	must(e)
	b, e := json.MarshalIndent(struct {
		GoDraw                                bool
		HostPlans, HostEntryVisits            int
		HostInstructions                      uint64
		CommittedPlans, CommittedInstructions uint64
		Images                                []imageCall
		Frames                                int
		Budget                                session.InstructionBudget
		Events                                []event
		Digest                                session.StateDigest
	}{*fast, o.hostPlans, o.hostEntryVisits, o.hostInstructions, committed, committedInstructions, o.images, o.update, budget, o.events, d}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "timing.json"), append(b, '\n'), 0600))
}
