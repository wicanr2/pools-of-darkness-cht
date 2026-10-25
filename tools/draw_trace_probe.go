// Research-only normal-DEMO video writer inventory. Built inside dosgolem/workplace.
package main

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "flag"
 "fmt"
 "os"
 "path/filepath"
 "sort"
 "time"

 "github.com/wicanr2/dosgolem/bootroot"
 "github.com/wicanr2/dosgolem/host"
 "github.com/wicanr2/dosgolem/internal/dos"
 "github.com/wicanr2/dosgolem/internal/machine"
 "github.com/wicanr2/dosgolem/session"
)

type writer struct {
 CS,IP uint16
 FirstStep,LastStep,Count uint64
 FirstUpdate int
 Lo,Hi uint32
 Code,Stack string
 Registers []uint16
 Memory string
}
type observer struct {
 out string
 update int
 writers map[uint32]*writer
 pending []*writer
 memories int
 indexed []byte
 palette [256][3]uint8
}
func must(e error){if e!=nil{panic(e)}}
func read(p string)[]byte{b,e:=os.ReadFile(p);must(e);return b}
func (o *observer) BeforeStep(v session.StepView)error{
 if len(o.pending)==0{return nil}
 for _,w:=range o.pending{
  base:=uint32(w.CS)<<4;code:=make([]byte,64);for i:=range code{code[i]=v.Read8(base+uint32(w.IP)+uint32(i))};w.Code=hex.EncodeToString(code)
  stack:=make([]byte,128);for i:=range stack{stack[i]=v.Read8((uint32(v.SS())<<4)+uint32(v.SP())+uint32(i))};w.Stack=hex.EncodeToString(stack)
  w.Registers=[]uint16{v.AX(),v.CX(),v.DX(),v.BX(),v.SP(),v.BP(),v.SI(),v.DI(),v.ES(),v.CS(),v.SS(),v.DS(),v.IP(),v.Flags()}
  if o.memories<16{
   memory:=make([]byte,1<<20);for i:=range memory{memory[i]=v.Read8(uint32(i))}
   w.Memory=fmt.Sprintf("writer-%04x-%04x.mem",w.CS,w.IP);must(os.WriteFile(filepath.Join(o.out,w.Memory),memory,0600));o.memories++
  }
 }
 o.pending=nil;return nil
}
func(o *observer)VideoWrite(w machine.VideoWrite){
 key:=uint32(w.CS)<<16|uint32(w.IP);s:=o.writers[key]
 if s==nil{if len(o.writers)>=256{panic("writer cap")};s=&writer{CS:w.CS,IP:w.IP,FirstStep:w.Step,FirstUpdate:o.update,Lo:w.Offset,Hi:w.Offset};o.writers[key]=s;o.pending=append(o.pending,s)}
 s.Count++;s.LastStep=w.Step;if w.Offset<s.Lo{s.Lo=w.Offset};if w.Offset>s.Hi{s.Hi=w.Offset}
}
func(o *observer)Frame(idx []byte,pal [256][3]uint8){o.indexed=idx;o.palette=pal}
func main(){
 orig:=flag.String("orig","","read-only original");seed:=flag.String("seed","","read-only G2 configuration");out:=flag.String("out","","new output");frames:=flag.Int("frames",7400,"normal DEMO frames");flag.Parse()
 if _,e:=os.Stat(*out);!os.IsNotExist(e){panic("output already exists")};must(os.Mkdir(*out,0700))
 required:=[]bootroot.RequiredFile{}
 for n,h:=range map[string]string{"ADLIB.COM":"7c39949b5c740d92449110e896f08f2d11b16b8602d7d3d6d59a9f250d574286","CONTROL.EXE":"60e53bcf4266fb4e68e4fc22990159d1c1af978e9d10068708cfd11d49935366"}{b,e:=hex.DecodeString(h);must(e);var a [32]byte;copy(a[:],b);required=append(required,bootroot.RequiredFile{Name:n,SHA256:a})}
 root,e:=bootroot.Prepare(bootroot.BootRootInput{OriginalRoot:*orig,SaveRoot:filepath.Join(*out,"save"),Required:required});must(e)
 for _,n:=range []string{"pool4.cfg","snd.com"}{must(os.WriteFile(filepath.Join(root.SaveRoot,n),read(filepath.Join(*seed,n)),0600))}
 exe:=read(filepath.Join(*orig,"ADLIB.COM"));scale:=host.OutputScale2
 layout:=host.MouseLayout{Epoch:1,Scale:scale,ChromeHeight:18*int(scale),Canvas:host.Canvas{Width:320,Height:200},FrameWidth:320*int(scale),FrameHeight:218*int(scale)}
 o:=&observer{out:*out,writers:map[uint32]*writer{}}
 owner,e:=session.New(session.Config{InitialScale:scale,InitialLayout:layout,AdLib:true,ClockPercent:50,Observer:o});must(e);defer owner.Close()
 _,e=owner.BootOriginal(session.BootInput{EXE:exe,ExpectedEXESHA256:sha256.Sum256(exe),SaveRoot:root.SaveRoot,Format:session.BootFormatCOM,Queue:[]session.BootQueued{{Name:"CONTROL.EXE"}},RootTSRShrink:true,ProgramPath:`C:\POD\SND.COM`});must(e)
 budget:=session.InstructionBudget(machine.ScaleSteps(machine.DefaultVGAFrameEvery,50)*70/60);start:=time.Now()
 for i:=1;i<=*frames;i++{
  o.update=i;v,e:=owner.View();must(e);input:=session.CapturedUpdate{StartedPanel:v.Panel,Layout:v.Layout,SourceGeneration:v.SourceGeneration}
  if i==900||i==1800{input.BIOSKeys=[]dos.Key{{Scan:0x39,ASCII:' '}}};_,e=owner.Deliver(input);must(e);tick,e:=owner.Advance(budget);must(e);if tick.RawError!=nil{panic(tick.RawError)}
  if i%1000==0{fmt.Printf("progress %d writers%d\n",i,len(o.writers))}
 }
 digest,e:=owner.Digest();must(e);rows:=[]*writer{};for _,w:=range o.writers{rows=append(rows,w)};sort.Slice(rows,func(i,j int)bool{return rows[i].Count>rows[j].Count})
 report:=struct{Scope string;Frames int;Budget session.InstructionBudget;Seconds float64;Writers []*writer;Digest session.StateDigest}{"read-only normal DEMO; BIOS spaces900/1800; register/stack snapshots immediately after first observed store; no replacement",*frames,budget,time.Since(start).Seconds(),rows,digest}
 b,e:=json.MarshalIndent(report,"","  ");must(e);must(os.WriteFile(filepath.Join(*out,"report.json"),append(b,'\n'),0600))
 must(os.WriteFile(filepath.Join(*out,"final.idx"),o.indexed,0600));flat:=make([]byte,0,768);for _,c:=range o.palette{flat=append(flat,c[:]...)};must(os.WriteFile(filepath.Join(*out,"final.pal"),flat,0600))
}
