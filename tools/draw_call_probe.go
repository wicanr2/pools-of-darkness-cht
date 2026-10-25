// Research-only normal-DEMO whole glyph call capture. Build in dosgolem/workplace.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/bootroot"
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/dos"
	"github.com/wicanr2/dosgolem/internal/machine"
	"github.com/wicanr2/dosgolem/session"
	"hash"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"
)

type call struct {
	R                                [8]uint16
	Seg                              [4]uint16
	IP, Flags                        uint16
	Step                             uint64
	Memory, MemorySHA256, CodeSHA256 string
}
type observer struct {
	out                                          string
	calls                                        uint64
	seen                                         map[byte]bool
	captures                                     []call
	canvas                                       *pod.DirectCanvas
	video, audio, frames, dosStream              hash.Hash
	videoCount, audioCount, frameCount, dosCount uint64
	dosPending                                   bool
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func (o *observer) BeforeStep(v session.StepView) error {
	state := func(label string) {
		fmt.Fprintf(o.dosStream, "%s %d %04x:%04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x %04x\n", label, v.Steps(), v.CS(), v.IP(), v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI(), v.ES(), v.CS(), v.SS(), v.DS(), v.Flags())
	}
	if o.dosPending {
		state("return")
		o.dosPending = false
	}
	pc := (uint32(v.CS()) << 4) + uint32(v.IP())
	if v.Read8(pc) == 0xcd && v.Read8((uint32(v.CS())<<4)+uint32(uint16(v.IP()+1))) == 0x21 {
		state("request")
		var buf [256]byte
		for i := range buf {
			buf[i] = v.Read8((uint32(v.DS()) << 4) + uint32(uint16(v.DX()+uint16(i))))
		}
		fmt.Fprintf(o.dosStream, "ds_dx_256 %x\n", sha256.Sum256(buf[:]))
		o.dosCount++
		o.dosPending = true
	}
	if v.CS() != 0x1b2e || v.IP() != 0x11a6 {
		return nil
	}
	o.calls++
	char := v.Read8((uint32(v.SS()) << 4) + uint32(uint16(v.SP()+4)))
	if len(o.captures) >= 32 || o.seen[char] {
		return nil
	}
	o.seen[char] = true
	mem := make([]byte, 1<<20)
	for i := range mem {
		mem[i] = v.Read8(uint32(i))
	}
	name := fmt.Sprintf("glyph-%03d.mem", char)
	must(os.WriteFile(filepath.Join(o.out, name), mem, 0600))
	o.captures = append(o.captures, call{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags(), Step: v.Steps(), Memory: name, MemorySHA256: fmt.Sprintf("%x", sha256.Sum256(mem)), CodeSHA256: fmt.Sprintf("%x", sha256.Sum256(mem[0x1b2e0+0x11a6:0x1b2e0+0x1203]))})
	return nil
}
func (o *observer) VideoWrite(w machine.VideoWrite) {
	var b [18]byte
	binary.LittleEndian.PutUint64(b[:], w.Step)
	binary.LittleEndian.PutUint16(b[8:], w.CS)
	binary.LittleEndian.PutUint16(b[10:], w.IP)
	binary.LittleEndian.PutUint32(b[12:], w.Offset)
	b[16] = w.Value
	b[17] = w.WriteMode
	o.video.Write(b[:])
	o.videoCount++
}
func (o *observer) Frame(idx []byte, pal [256][3]uint8) {
	must(o.canvas.Frame(idx))
	o.frames.Write(idx)
	for _, c := range pal {
		o.frames.Write(c[:])
	}
	o.frameCount++
}
func (o *observer) VideoModeChange(machine.ModeChange)  { o.canvas.Reset() }
func (o *observer) VideoCommand(c machine.VideoCommand) { must(o.canvas.Apply(c)) }
func (o *observer) HostCallTargets() []uint32           { return pod.GlyphHostCallTargets() }
func (o *observer) PrepareHostCall(v session.StepView, budget uint64) (*machine.HostWorkPlan, error) {
	return pod.PlanGlyphCall(v, budget), nil
}
func (o *observer) OPLWrite(w machine.OPLWrite) { fmt.Fprintf(o.audio, "O%+v\n", w); o.audioCount++ }
func (o *observer) SpeakerSample(w machine.SpeakerSample) {
	fmt.Fprintf(o.audio, "S%+v\n", w)
	o.audioCount++
}
func (o *observer) PITChannel2(w machine.PIT2Change) {
	fmt.Fprintf(o.audio, "P%+v\n", w)
	o.audioCount++
}
func main() {
	root := flag.String("root", "", "project")
	out := flag.String("out", "", "new output")
	whole := flag.Bool("whole-call", false, "enable whole glyph calls")
	flag.Parse()
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("output exists")
	}
	must(os.Mkdir(*out, 0700))
	orig := filepath.Join(*root, "workplace/input/dos/pod")
	seed := filepath.Join(*root, "workplace/probe/g2-receipt/scratch")
	required := []bootroot.RequiredFile{}
	for name, hash := range map[string]string{"CONTROL.EXE": "60e53bcf4266fb4e68e4fc22990159d1c1af978e9d10068708cfd11d49935366", "ADLIB.COM": "7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286"} {
		b, e := hex.DecodeString(hash)
		must(e)
		var want [32]byte
		copy(want[:], b)
		required = append(required, bootroot.RequiredFile{Name: name, SHA256: want})
	}
	br, e := bootroot.Prepare(bootroot.BootRootInput{OriginalRoot: orig, SaveRoot: filepath.Join(*out, "save"), Required: required})
	must(e)
	for _, n := range []string{"pool4.cfg", "snd.com"} {
		must(os.WriteFile(filepath.Join(br.SaveRoot, n), read(filepath.Join(seed, n)), 0600))
	}
	stamp := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	must(filepath.WalkDir(br.SaveRoot, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			return os.Chtimes(p, stamp, stamp)
		}
		return nil
	}))
	scale := host.OutputScale2
	layout := host.MouseLayout{Epoch: 1, Scale: scale, ChromeHeight: 36, Canvas: host.Canvas{Width: 320, Height: 200}, FrameWidth: 640, FrameHeight: 436}
	o := &observer{out: *out, seen: map[byte]bool{}, canvas: pod.NewDirectCanvas(), video: sha256.New(), audio: sha256.New(), frames: sha256.New(), dosStream: sha256.New()}
	owner, e := session.New(session.Config{InitialScale: scale, InitialLayout: layout, AdLib: true, ClockPercent: 50, Observer: o, NativeVideoCommands: true, NativeHostCalls: *whole})
	must(e)
	defer owner.Close()
	exe := read(filepath.Join(orig, "ADLIB.COM"))
	_, e = owner.BootOriginal(session.BootInput{EXE: exe, ExpectedEXESHA256: sha256.Sum256(exe), SaveRoot: br.SaveRoot, Format: session.BootFormatCOM, Queue: []session.BootQueued{{Name: "CONTROL.EXE"}}, RootTSRShrink: true, ProgramPath: `C:\POD\SND.COM`})
	must(e)
	profile, e := os.Create(filepath.Join(*out, "cpu.pprof"))
	must(e)
	must(pprof.StartCPUProfile(profile))
	start := time.Now()
	hostCalls, hostInstructions := uint64(0), uint64(0)
	for i := 1; i <= 7400; i++ {
		v, e := owner.View()
		must(e)
		input := session.CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration}
		if i == 900 || i == 1800 {
			input.BIOSKeys = []dos.Key{{Scan: 0x39, ASCII: ' '}}
		}
		_, e = owner.Deliver(input)
		must(e)
		tick, e := owner.Advance(96250)
		must(e)
		must(tick.RawError)
		hostCalls += tick.HostCalls
		hostInstructions += tick.HostInstructions
		if i%2000 == 0 {
			fmt.Printf("frame%d calls%d captures%d\n", i, o.calls, len(o.captures))
		}
	}
	pprof.StopCPUProfile()
	must(profile.Close())
	digest, e := owner.Digest()
	must(e)
	b, e := json.MarshalIndent(struct {
		Calls                                             uint64
		Captures                                          []call
		Digest                                            session.StateDigest
		Seconds                                           float64
		WholeCall                                         bool
		HostCalls, HostInstructions                       uint64
		VideoCount, AudioCount, FrameCount, DOSCount      uint64
		VideoSHA256, AudioSHA256, FramesSHA256, DOSSHA256 string
		Commands                                          pod.DirectStats
	}{o.calls, o.captures, digest, time.Since(start).Seconds(), *whole, hostCalls, hostInstructions, o.videoCount, o.audioCount, o.frameCount, o.dosCount, fmt.Sprintf("%x", o.video.Sum(nil)), fmt.Sprintf("%x", o.audio.Sum(nil)), fmt.Sprintf("%x", o.frames.Sum(nil)), fmt.Sprintf("%x", o.dosStream.Sum(nil)), o.canvas.Stats()}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(*out, "calls.json"), append(b, '\n'), 0600))
}
