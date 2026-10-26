"""在Docker內建置入口觀測探針；只分割研究用固定區塊，不改產品來源。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--module', type=Path, required=True)
    ap.add_argument('--name', required=True)
    ap.add_argument('--program-commit', required=True)
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert re.fullmatch(r'[a-z0-9-]+', a.name)
    assert re.fullmatch(r'[0-9a-f]{40}', a.program_commit)
    actual_commit = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=a.module, text=True).strip()
    assert actual_commit == a.program_commit, '來源提交與module HEAD不符，拒絕建立觀測產物'
    assert not a.out.exists()
    folder = a.module/'workplace'/f'{a.name}-blocks'
    probe = a.module/'workplace'/f'{a.name}-probe'
    folder.mkdir()
    probe.mkdir()
    records = {}
    for source in sorted((a.module/'apps/pod/drawblocks').glob('*.go')):
        if source.name.endswith('_test.go'):
            continue
        target = folder/source.name
        target.write_bytes(source.read_bytes())
        records[source.name] = dict(original_sha256=sha(source), observed_sha256=sha(target))
    effects = folder/'effects.go'
    s = effects.read_text()
    signature = 'func Plan(v CPUView, budget uint64) *machine.HostWorkPlan {'
    boundary = 'if e.p.Instructions > 0 && algorithmBoundary(e.s.Seg[1], e.s.IP) {'
    assert s.count(signature) == s.count(boundary) == 1
    s = s.replace(signature, 'func Plan(v CPUView, budget uint64, entries map[uint32]uint16) *machine.HostWorkPlan {')
    s = s.replace(boundary, 'if e.p.Instructions > 0 && (algorithmBoundary(e.s.Seg[1], e.s.IP) || observedEntry(entries, e.s.Seg[1], e.s.IP)) {')
    s += '\nfunc observedEntry(entries map[uint32]uint16, cs, ip uint16) bool { _, ok := entries[uint32(cs)<<16|uint32(ip)]; return ok }\n'
    effects.write_text(s)
    original_probe = a.project/'tools/draw_call_probe.go'
    s = original_probe.read_text()
    source_import = '"github.com/wicanr2/dosgolem/apps/pod/drawblocks"'
    source_call = 'drawblocks.Plan(v, budget)'
    assert s.count(source_import) == s.count(source_call) == 1
    s = s.replace(source_import, f'"github.com/wicanr2/dosgolem/workplace/{folder.name}"')
    s = s.replace(source_call, 'drawblocks.Plan(v, budget, o.drawEntries)')
    main_go = probe/'main.go'
    main_go.write_text(s)
    subprocess.run(['gofmt', '-w', str(effects), str(main_go)], check=True)
    records['effects.go']['observed_sha256'] = sha(effects)
    binary = a.module/'workplace/bin'/f'draw-call-probe-{a.name}'
    assert not binary.exists()
    subprocess.run(['go', 'build', '-o', str(binary), f'./workplace/{probe.name}'], cwd=a.module, check=True)
    result = dict(program_commit=a.program_commit,
                  scope='Research-only fixed-plan boundaries before registered entries; no asset, input or guest-state injection; product planner unchanged',
                  binary=binary.name, binary_sha256=sha(binary),
                  original_probe_sha256=sha(original_probe), observed_probe_sha256=sha(main_go),
                  builder_sha256=sha(Path(__file__)), source_files=records)
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(dict(binary=binary.name, binary_sha256=sha(binary))))


if __name__ == '__main__':
    main()
