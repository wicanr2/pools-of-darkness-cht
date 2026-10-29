#!/usr/bin/env python3
"""驗證常駐字模Go單點與原生16／24像素英數呈現。"""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text())


def tree(path):
    return {str(p.relative_to(path)): sha(p) for p in sorted(path.rglob('*')) if p.is_file()}


def resolution(path):
    text = path.read_text()
    names = ('resident-whole-initial', 'resident-short-initial', 'resident-glyph-ready')
    anchors = {name: f'<a id="{name}"></a>' for name in names}
    assert all(text.count(anchor) == 1 for anchor in anchors.values())
    section = lambda name: text.split(anchors[name])[1].split('<a id=')[0]
    marker = 'RESIDENT-GLYPH-READY-20261029'
    key = 'DOS/GAME.EXE/35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69/runtime:1B2E:11A6'
    ready = section(names[-1])
    assert marker in ready and key in ready and '11EA' in ready and '11F1' in ready
    for name in names[:-1]:
        assert marker in section(name) and '(#resident-glyph-ready)' in section(name)
        assert '(#'+name+')' in ready
    return dict(key=key, correction=marker)


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--project', type=Path, required=True)
    ap.add_argument('--commit', required=True)
    ap.add_argument('--stage', choices=('trace', 'full'), default='full')
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    assert not a.out.exists(), '保留既有收據，請另定檔名'
    p = a.project
    module = p/'workplace/dosgolem'
    raw = p/'workplace/probe/draw-intercept-v1/native-resident-glyph-v1'
    baseline_path = p/'docs/re/019-native-surface-fill-spans-runtime.json'
    baseline = read(baseline_path)
    git = lambda *args: subprocess.check_output(['git', *args], cwd=module, text=True).strip()
    assert git('rev-parse', 'HEAD') == a.commit and not git('status', '--porcelain')
    base = baseline['program_commit']
    assert git('merge-base', base, a.commit) == base
    names = ['apps/pod/'+name for name in ('direct_compose.go', 'draw_blocks.go', 'live.go', 'glyph_text.go',
             'glyph_text_test.go', 'glyph_text_composition_test.go', 'glyph_text_duplicate_test.go',
             'resident_glyph_pixel_test.go', 'resident_glyph_guards_test.go', 'resident_glyph_edges_test.go')]
    names += ['apps/pod/drawblocks/effects.go', 'apps/pod/drawblocks/native_resident_glyph.go',
              'cmd/pod-play/main.go', 'cmd/pod-play/source.go']
    assert sorted(git('diff', '--name-only', base, a.commit).splitlines()) == sorted(names)
    for name in ('apps/pod/drawblocks/native_resident_glyph.go', 'apps/pod/glyph_text.go'):
        text = (module/name).read_text()
        assert all(token not in text for token in ('.Step(', 'b.run(', 'effect{', 'DecodeOpcode', '.Write8(', '.Write16('))
    for name, digest in baseline['source_sha256'].items():
        if name not in names:
            assert sha(module/name) == digest
    build = (raw/'build-v2.txt').read_text()
    assert 'go1.24.13' in build and 'vcs.modified=false' in build and 'vcs.revision='+a.commit in build
    original = p/'workplace/input/dos/pod/GAME.EXE'
    assert sha(original) == '35ad635e558852925b7028a921aacafc2df8e84be24a8a1136c6267471948d69'
    product_log = module/'workplace/native-resident-glyph-final-product-tests.log'
    text = product_log.read_text()
    for marker in (
        'resident glyph ordered original pixel comparisons=4096 phases=2',
        'resident glyph whole SaveState+PortLog cases=192 Go pixels=8192 full glyph plans=64 CPU control=72832 CPU resident pixels=0',
        'resident glyph positive phases=2 source/frame/cursor/model refusals=64 unchanged',
        'resident glyph corner/full-background/full-foreground SaveState+PortLog cases=48 CPU resident pixels=0',
        'native glyph font original-cell draws=20 two scales native-detail cases=17 pending masks and source invalidation passed',
        'native glyph source/icon/frame acceptance and refusal cases=9 guest unchanged',
        'native 16/24 font receipts, original toggle, translation priority and mode reset passed',
        '--- PASS: TestResidentPixelRanges', '--- PASS: TestGlyphTextDuplicateEntry',
        '--- PASS: TestWholeGlyphIndependentOriginal', '--- PASS: TestWholeGlyphBoundaryStates',
    ):
        assert marker in text, marker
    assert '\nPASS\n' in text and 'FAIL' not in text and 'SKIP' not in text
    regression = read(raw/'regression-inputs.json')
    assert git('rev-parse', 'HEAD^') == regression['commit']
    for name, digest in regression['drawblocks_source_sha256'].items():
        assert sha(module/name) == digest, name
    log = module/'workplace/native-resident-glyph-regression.log'
    assert sha(log) == regression['regression_log_sha256']
    assert 'github.com/wicanr2/dosgolem/apps/pod/drawblocks\t' in log.read_text() and 'FAIL' not in log.read_text()
    final_log = module/'workplace/native-resident-glyph-final-related-tests.log'
    final_text = final_log.read_text()
    assert 'FAIL' not in final_text
    for package in ('apps/pod', 'apps/pod/bgm', 'xlate', 'session', 'internal/machine', 'cmd/pod-play'):
        assert 'github.com/wicanr2/dosgolem/'+package+'\t' in final_text
    vet = module/'workplace/native-resident-glyph-final-vet.log'
    assert not vet.read_text()
    paths = [raw/'trace-native', p/'workplace/probe/draw-intercept-v1/native-surface-fill-spans-v1/trace-native',
             p/'workplace/probe/draw-intercept-v1/native-fade-v1/trace-original-entries']
    current, previous, oracle = [read(path/'calls.json') for path in paths]
    for field in ('Calls', 'Digest', 'VideoCount', 'AudioCount', 'FrameCount', 'DOSCount',
                  'VideoSHA256', 'AudioSHA256', 'FramesSHA256', 'DOSSHA256'):
        assert current[field] == previous[field] == oracle[field], field
    assert tree(paths[0]/'save') == tree(paths[1]/'save') == tree(paths[2]/'save')
    for field in ('Captures', 'DrawCaptures'):
        assert current[field] == previous[field]
        for row in current[field]:
            assert sha(paths[0]/row['Memory']) == sha(paths[1]/row['Memory']) == row['MemorySHA256']
    assert current['Commands']['Mismatches'] == 0
    assert len(oracle['Captures']) == 32
    for row in oracle['Captures']:
        data = (paths[2]/row['Memory']).read_bytes()
        assert hashlib.sha256(data[0x1c486:0x1c4e3]).hexdigest() == 'd1c978b914f28d1a978a2cae1b7a1c52f4b5d233176e8bd6c3029b05c9b4958b'
        word = lambda at: int.from_bytes(data[at:at+2], 'little')
        table = row['Seg'][3]*16+0x9d86
        font = word(table+2)*16+word(table)
        assert hashlib.sha256(data[font+256:font+1016]).hexdigest() == '72fb1adcd5698a5e9770b91a499c347c7c8b05c567b93a8cca1f7249c49b4a7e'
    counts = current['AlgorithmicPreparedByExit']
    observed = {key: value for key, value in counts.items() if key in ('1b2e:11ea', '1b2e:11f1')}
    assert set(observed) == {'1b2e:11ea', '1b2e:11f1'} and all(row[0] == row[1] > 0 for row in observed.values())
    for key in set(counts) | set(previous['AlgorithmicPreparedByExit']):
        if key not in observed:
            assert counts.get(key) == previous['AlgorithmicPreparedByExit'].get(key), key
    pixel_count = sum(row[0] for row in observed.values())
    assert current['AlgorithmicPlans']-previous['AlgorithmicPlans'] == pixel_count
    assert current['AlgorithmicInstructions']-previous['AlgorithmicInstructions'] == pixel_count
    assert current['Commands']['HostBytes']-previous['Commands']['HostBytes'] == pixel_count
    fixed = current['HostInstructions']+current['BlockInstructions']-current['AlgorithmicInstructions']
    assert fixed <= baseline['fixed_semantic_instructions']
    result = dict(scope='常駐8×8字模Go單點與來源認證ASCII原生16／24像素顯示；不外推其他繪圖及未驗路線',
        program_commit=a.commit, baseline_commit=base, parent_commit=regression['commit'],
        binary_sha256=sha(module/'workplace/bin/pod-play-nativeresidentglyph-v2'),
        original_state_and_ordered_events_equal=True, virtual_steps=current['Digest']['Steps'],
        guest_files=len(tree(paths[0]/'save')), guest_files_equal=True,
        direct_original_witnesses=len(current['Captures'])+len(current['DrawCaptures']), original_witnesses_equal=True,
        algorithmic_plans=current['AlgorithmicPlans'], algorithmic_instructions=current['AlgorithmicInstructions'],
        fixed_semantic_instructions=fixed, normal_resident_pixel_exits=observed, new_normal_go_pixel_bytes=pixel_count,
        controlled_original=dict(ordered_pixels=4096, pixel_phases=2, whole_calls=192, go_pixels=8192,
            full_glyph_plans=64, cpu_control=72832, cpu_resident_pixels=0, corner_calls=48, refusals=64),
        high_resolution_controlled=dict(native_sizes=[16,24], real_source_draws=20, native_detail_cases=17,
            source_icon_frame_cases=9, pending_mask=True, source_invalidation=True, original_toggle=True,
            translation_priority=True, duplicate_entry_idempotent=True),
        source_sha256={name:sha(module/name) for name in names},
        resolution_backlinks=resolution(p/'docs/spec/025-direct-render-commands.md'),
        spec_sha256=sha(p/'docs/spec/025-direct-render-commands.md'),
        regression=dict(drawblocks_verified_at=regression['commit'], drawblocks_sources_unchanged=True,
            inputs_sha256=sha(raw/'regression-inputs.json'), full_log_sha256=sha(log),
            final_related_log_sha256=sha(final_log), final_vet_sha256=sha(vet), product_log_sha256=sha(product_log)),
        build_info_sha256=sha(raw/'build-v2.txt'), trace_binary_sha256=sha(module/'workplace/bin/draw-call-probe-nativeresidentglyph-v2'),
        prior_full_receipt_sha256=sha(baseline_path), source_reports={str(path.relative_to(p)):sha(path/'calls.json') for path in paths},
        prior_full_route_note='前批四路demo、39對介面、六對真存讀檔與demo中斷另存原收據，本批正常路線另列')
    if a.stage == 'full':
        ui_path = raw/'ui-check.json'
        ui = read(ui_path)
        assert ui['tool_sha256'] == sha(p/'tools/native_resident_glyph_ui_check.py')
        assert ui['direct_ui_tool_sha256'] == sha(p/'tools/direct-ui-receipt.sh')
        assert ui['current_binary_sha256'] == result['binary_sha256']
        assert ui['image_metadata_negative_controls'] == 7 and ui['native_font_sizes'] == [16, 24]
        before = dict(hd=p/'workplace/probe/g9-hd/nativesurfacefillspans-play-v1/hd',
                      font=p/'workplace/probe/g9-hd/nativeresidentglyph-fontbaseline-v1/wall-direct')
        folders = dict(hd=p/'workplace/probe/g9-hd/nativeresidentglyph-play-v2/hd',
                       font=p/'workplace/probe/g9-hd/nativeresidentglyph-font-v2/wall-direct')
        for row in ui['rows']:
            assert row['unexpected_flash_count'] == 0 and row['original_state_equal'] and row['guest_files_equal']
            for item in row['images']:
                path = Path(item['image'])
                assert item['outside_cell_changes'] == 0 and sha(path) == item['image_sha256']
                assert sha(before[row['route']]/path.relative_to(folders[row['route']])) == item['previous_sha256']
                source = path.parent/'source' if path.name == 'frame.png' else path.parent.parent/'sources'/path.stem
                assert sha(source.with_suffix('.idx')) == item['source_sha256']
                assert sha(source.with_suffix('.glyphs.json')) == item['glyph_report_sha256']
        result['normal_frontend'] = ui
        result['ui_receipt_sha256'] = sha(ui_path)
        preview_path = raw/'preview/preview.json'
        preview = read(preview_path)
        assert preview['scale'] == 3 and preview['method'] == 'native same-state screenshot crop; no resampling'
        for name, digest in preview['files'].items():
            assert sha(raw/'preview'/name) == sha(p/'docs/screenshots'/name) == digest
        result['readme_preview'] = dict(**preview, manifest_sha256=sha(preview_path))
        if (raw/'activation.json').exists():
            activation = read(raw/'activation.json')
            assert sha(module/'workplace/bin/pod-play') == activation['candidate_sha256'] == result['binary_sha256']
            assert sha(module/'workplace/bin'/activation['previous_binary']) == activation['previous_sha256'] == baseline['binary_sha256']
            result['activation'] = activation
    result['tool_sha256'] = sha(Path(__file__))
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps({key:result[key] for key in ('algorithmic_plans','fixed_semantic_instructions','new_normal_go_pixel_bytes','normal_resident_pixel_exits')}))


if __name__ == '__main__':
    main()
