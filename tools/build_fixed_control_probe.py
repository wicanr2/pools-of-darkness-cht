#!/usr/bin/env python3
"""在Docker產生正常路線固定控制觀測器；產品planner保持原樣。"""
import argparse
from pathlib import Path


def render(project):
    source = (project/'tools/draw_call_probe.go').read_text()
    source = source.replace('type observer struct {', 'type fixedRecord struct {\n Before call\n Budget uint64\n Plan *machine.HostWorkPlan\n}\ntype observer struct {\n fixedRecords []fixedRecord', 1)
    source = source.replace('p := drawblocks.Plan(v, budget)\n\to.recordAlgorithmic(p)', '''p := drawblocks.Plan(v, budget)
 if p != nil && !p.Algorithmic {
  mem:=make([]byte,1<<20); for i:=range mem { mem[i]=v.Read8(uint32(i)) }
  name:=fmt.Sprintf("fixed-%d.mem",v.Steps());must(os.WriteFile(filepath.Join(o.out,name),mem,0600))
  pc:=uint32(v.CS())*16+uint32(v.IP())
  before:=call{R:[8]uint16{v.AX(),v.CX(),v.DX(),v.BX(),v.SP(),v.BP(),v.SI(),v.DI()},Seg:[4]uint16{v.ES(),v.CS(),v.SS(),v.DS()},IP:v.IP(),Flags:v.Flags(),Step:v.Steps(),Memory:name,MemorySHA256:fmt.Sprintf("%x",sha256.Sum256(mem)),CodeSHA256:fmt.Sprintf("%x",sha256.Sum256(mem[pc:pc+64]))}
  o.fixedRecords=append(o.fixedRecords,fixedRecord{before,budget,p})
 }
 o.recordAlgorithmic(p)''', 1)
    source = source.replace('must(os.WriteFile(filepath.Join(*out, "calls.json"), append(b, \'\\n\'), 0600))', '''must(os.WriteFile(filepath.Join(*out, "calls.json"), append(b, '\\n'), 0600))
 b,e=json.MarshalIndent(o.fixedRecords,"","  ");must(e);must(os.WriteFile(filepath.Join(*out,"fixed-records.json"),append(b,'\\n'),0600))''', 1)
    assert 'fixed-records.json' in source and 'o.fixedRecords=append' in source
    return source


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--package-dir', default='workplace/dosgolem/workplace/native-fixed-control-audit')
    ap.add_argument('--receipt-root', default='workplace/probe/draw-intercept-v1/native-fixed-control-v1')
    a = ap.parse_args()
    project = a.project.resolve()
    out, raw = [(project/name).resolve() for name in (a.package_dir, a.receipt_root)]
    for path in (out, raw):
        assert path.is_relative_to(project/'workplace') and not path.exists(), path
    out.mkdir()
    raw.mkdir()
    (out/'main.go').write_text(render(project))
    print(out)


if __name__ == '__main__':
    main()
