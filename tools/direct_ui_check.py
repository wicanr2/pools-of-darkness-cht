#!/usr/bin/env python3
"""核對同一路線直接／舊合成、既有原版終點及逐張視窗截圖。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def sha(b):
    return hashlib.sha256(b).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes())
            for p in sorted(root.rglob('*')) if p.is_file()}


def endpoint(folder):
    log = (folder / 'run.log').read_text()
    a = re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)', log)
    b = re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)', log)
    assert a and b, folder
    assert '字型缺字' not in log and '截圖失敗' not in log
    return a.groups() + b.groups()


def audit(folder, route):
    log = (folder/'run.log').read_text()
    count = int(re.search(r'pod-play: 閃現 (\d+) 次', log)[1])
    if count == 0:
        return dict(flash_count=0, intentional_mode_change_cells=0, unexpected_flash_count=0)
    # The unchanged audit compares across deliberate F2 epochs. At 1220 this
    # exact script changes an original-only 2x title back to the HD 3x title.
    # Require every event, not only the total, to belong to that input edge.
    assert route == 'title', (folder, count)
    events = re.findall(r'pod-play: 閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄([^\n]+)', log)
    assert events and all((shown, changed) == ('1219', '1220') for shown, changed, _, _ in events)
    cells = sum(len(cols.split()) for _, _, _, cols in events)
    assert len(events) == 25 and cells == count == 998
    return dict(flash_count=count, intentional_mode_change_cells=cells, unexpected_flash_count=0)


def main():
    p = argparse.ArgumentParser()
    p.add_argument('--root', type=Path, required=True)
    p.add_argument('--project', type=Path, required=True)
    p.add_argument('--out', type=Path, required=True)
    p.add_argument('--route', action='append', choices=('wall', 'movie', 'title'))
    args = p.parse_args()
    rows = []
    for route, frames, baseline in (
        ('wall', 4800, 'workplace/probe/g9-hd/wall-play-v3/hd'),
        ('movie', 5100, 'workplace/probe/g9-hd/intro-play-v1/subtitles'),
        ('title', 1250, 'workplace/probe/g9-hd/title-play-v3/switch'),
    ):
        if args.route and route not in args.route:
            continue
        a, b = [args.root / f'{route}-{mode}' for mode in ('direct', 'legacy')]
        state = endpoint(a)
        assert state == endpoint(b) == endpoint(args.project / baseline), route
        assert tuple(map(int, state[:2])) == (frames, frames * 96250)
        records = [json.loads((folder/'inputs.json').read_text()) for folder in (a, b)]
        assert records[0] == records[1]
        # Compare original DOS files independently of presentation preferences.
        frontend = {'pod-map-visited.txt', 'pod-journal-seen.txt', 'pod-walkthrough-pos.txt',
                    'pod-lang.txt', 'pod-speed.txt', 'pod-barfont.txt'}
        guests = [{k: v for k, v in tree(folder/'save').items() if k not in frontend}
                  for folder in (a, b, args.project/baseline)]
        assert guests[0] == guests[1] == guests[2], route
        for name in ('source.idx', 'source.pal', 'original.rgba', 'frame.png'):
            assert (a/name).read_bytes() == (b/name).read_bytes(), (route, name)
        for name in ('shots', 'window-shots', 'sources'):
            assert tree(a/name) and tree(a/name) == tree(b/name), (route, name)
        measured = audit(a, route)
        assert measured == audit(b, route)
        draw = json.loads((a/'draw-report.json').read_text())
        old_draw = json.loads((b/'draw-report.json').read_text())
        assert draw['Frames'] > 0 and draw['Mismatches'] == draw['LegacyRGBFrames'] == 0
        assert draw['NativeBytes'] > 0 and draw['SelectedRGBPixels'] > 0 and old_draw['Commands'] == 0
        rows.append(dict(route=route, frames=frames, steps=int(state[1]), memory_sha256=state[2],
                         indexed_sha256=state[3], cpu_sha256=state[4], palette_sha256=state[5],
                         guest_files=len(guests[0]), original_state_equal=True, guest_state_equal=True,
                         screenshot_pairs=len(tree(a/'shots')), all_screenshots_equal=True, **measured,
                         commands=draw, inputs_sha256=sha((a/'inputs.json').read_bytes()),
                         frame_sha256=sha((a/'frame.png').read_bytes()),
                         direct_draw_report_sha256=sha((a/'draw-report.json').read_bytes())))
    result = dict(scope='既有正常讀檔、片頭與標題路線；命令／舊合成逐張同圖，原版終點另比既有收據；不外推未走路線',
                  rows=rows, tool_sha256=sha(Path(__file__).read_bytes()))
    args.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps({r['route']: dict(screenshot_pairs=r['screenshot_pairs'],
                                     commands=r['commands']['Commands'], flash_count=r['flash_count']) for r in rows}))


if __name__ == '__main__':
    main()
