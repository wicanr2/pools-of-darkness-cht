#!/usr/bin/env python3
"""在隔離前端副本追蹤像素來源，正式程式與planner保持原樣。"""
import argparse
import json
import os
from pathlib import Path
import shutil

ap = argparse.ArgumentParser()
ap.add_argument('--module', type=Path, required=True)
ap.add_argument('--replace', action='store_true')
a = ap.parse_args()
module = a.module.resolve()
src = module / 'cmd/pod-play'
out = module / 'workplace/draw-closure-audit-cmd'
assert src.is_dir()
if out.exists():
    assert a.replace and out.is_dir() and not out.is_symlink() and out.stat().st_uid == os.getuid()
    shutil.rmtree(out)
out.mkdir()
for f in src.glob('*.go'):
    shutil.copy2(f, out / f.name)
p = out / 'main.go'
text = p.read_text()
changes = {
    'func (o observer) BeforeStep(v session.StepView) error      { return o.live.BeforeStep(v) }':
    'func (o observer) BeforeStep(v session.StepView) error { drawClosureVisit(v); return o.live.BeforeStep(v) }',
    'return o.live.PrepareHostBlock(v, budget)':
    'p,e:=o.live.PrepareHostBlock(v,budget); drawClosurePrepared(v,p);return p,e',
    'return o.live.PrepareHostCall(v, budget)':
    'p,e:=o.live.PrepareHostCall(v,budget); drawClosurePrepared(v,p);return p,e',
    'func (o observer) VideoCommand(c machine.VideoCommand) { o.live.VideoCommand(c) }':
    'func (o observer) VideoCommand(c machine.VideoCommand) { drawClosureCommand(c);o.live.VideoCommand(c) }',
    'func (g *game) finish() error {':
    'func (g *game) finish() error {\n\tdrawClosureFinish(g.shotFrame,g.drawReport)',
}
for before, after in changes.items():
    assert text.count(before) == 1, before
    text = text.replace(before, after, 1)
p.write_text(text)
table = json.loads((module / 'apps/pod/drawblocks/blocks.json').read_text())
roots = '\n'.join(f'0x{cs:04x}{lo:04x}: {hi-lo},' for cs, lo, hi, _ in table['functions'])
(out / 'draw_closure.go').write_text('''package main
import("crypto/sha256";"encoding/json";"fmt";"os";"path/filepath";
"github.com/wicanr2/dosgolem/internal/machine";"github.com/wicanr2/dosgolem/session")
var closureRoots=map[uint32]int{
''' + roots + r'''
}
var closure struct{
 LastStart,LastEnd uint64
 LastClass string
 Bytes,Commands map[string]uint64
 Writers map[string]map[string]uint64
 Prepared map[string][2]uint64
 Visits map[string]uint64
 CodeSHA map[string]string
 LinkErrors uint64
}
func init(){
 closure.Bytes=map[string]uint64{};closure.Commands=map[string]uint64{}
 closure.Writers=map[string]map[string]uint64{};closure.Prepared=map[string][2]uint64{}
 closure.Visits=map[string]uint64{};closure.CodeSHA=map[string]string{}
}
func drawClosureVisit(v session.StepView){
 if v.CS()!=0x48a1&&v.CS()!=0x1ed4&&v.CS()!=0x1d1f{return}
 key:=uint32(v.CS())<<16|uint32(v.IP());size,ok:=closureRoots[key];if !ok{return}
 name:=fmt.Sprintf("%04x:%04x",v.CS(),v.IP());closure.Visits[name]++
 if _,ok:=closure.CodeSHA[name];!ok{
  raw:=make([]byte,size);for i:=range raw{raw[i]=v.Read8(uint32(v.CS())*16+uint32(v.IP())+uint32(i))}
  closure.CodeSHA[name]=fmt.Sprintf("%x",sha256.Sum256(raw))
 }
}
func drawClosurePrepared(v session.StepView,p *machine.HostWorkPlan){
 if p==nil{return}
 kind:="fixed";if p.Algorithmic{kind="typed_go"}
 closure.LastClass=kind;closure.LastStart=v.Steps();closure.LastEnd=v.Steps()+p.Instructions
 name:=fmt.Sprintf("%s/%04x:%04x",kind,v.CS(),v.IP());row:=closure.Prepared[name];row[0]++;row[1]+=p.Instructions;closure.Prepared[name]=row
 drawClosureVisit(v)
}
func drawClosureCommand(c machine.VideoCommand){
 kind:="cpu_pixel";if c.Native{kind="x86_native_string"}
 if c.Host{
  kind=closure.LastClass
  if c.Step<=closure.LastStart||c.Step>closure.LastEnd||kind==""{closure.LinkErrors++;kind="unlinked_host"}
 }
 n:=uint64(c.Width)*uint64(c.Count);closure.Bytes[kind]+=n;closure.Commands[kind]++
 key:=fmt.Sprintf("%04x:%04x",c.CS,c.IP);if closure.Writers[kind]==nil{closure.Writers[kind]=map[string]uint64{}};closure.Writers[kind][key]+=n
}
func drawClosureFinish(framePath,reportPath string){
 if framePath==""&&reportPath==""{return} // 一格安裝檢查沒有要求輸出收據。
 path:=reportPath+".coverage.json";if framePath!=""{path=filepath.Join(filepath.Dir(framePath),"coverage.json")}
 b,e:=json.MarshalIndent(closure,"","  ");if e!=nil{panic(e)}
 if e=os.WriteFile(path,append(b,10),0600);e!=nil{panic(e)}
}
''')
print(out)
