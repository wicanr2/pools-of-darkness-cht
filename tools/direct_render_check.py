#!/usr/bin/env python3
"""直接繪圖命令的正常 DEMO、事件與相容畫布驗證。原始資料限本機。"""
import argparse
import hashlib
import json
from pathlib import Path


def sha(b):
    return hashlib.sha256(b).hexdigest()


def tree(root):
    return {str(p.relative_to(root)): sha(p.read_bytes())
            for p in sorted(root.rglob('*')) if p.is_file()}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--root', type=Path, required=True)
    parser.add_argument('--writers', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True)
    a = parser.parse_args()
    folders = [a.root / name for name in ('native-audit-demo', 'legacy-audit-demo', 'normal-demo')]
    rows = [json.loads((p / 'report.json').read_text()) for p in folders]
    native, legacy, previous = rows
    assert native['Native'] and not legacy['Native']
    assert native['FixedFileTime'] and legacy['FixedFileTime']
    assert all(r['Frames'] == 7400 and r['Budget'] == 96250 for r in rows)
    assert native['Digest'] == legacy['Digest']
    # Changing all input file dates changes the guest directory buffers. The
    # unfrozen pair independently checks the exact historical endpoint.
    unfrozen = [json.loads((a.root/name/'report.json').read_text())
                for name in ('native-demo', 'legacy-demo')]
    assert unfrozen[0]['Digest'] == unfrozen[1]['Digest'] == previous['Digest']
    assert (native['AudioSHA256'], native['AudioEvents']) == (legacy['AudioSHA256'], legacy['AudioEvents'])
    assert native['AudioEvents'] > 0
    assert native['DOSAudit'] and legacy['DOSAudit']
    assert native['DOSEvents'] == legacy['DOSEvents'] > 0
    assert native['DOSSHA256'] == legacy['DOSSHA256']
    assert tree(folders[0] / 'save') == tree(folders[1] / 'save') == tree(folders[2] / 'save')
    for name in ('final.idx', 'final.pal'):
        assert len({sha((p / name).read_bytes()) for p in folders}) == 1
    c = native['Commands']
    assert c['NativeBytes'] == sum(w['Count'] for w in native['Writers']) == 3343240
    assert c['FallbackBytes'] == c['Mismatches'] == 0 and c['Frames'] > 7400
    assert len(c['Writers']) == len(native['Writers']) == 10
    writers = []
    for w in native['Writers']:
        key = str(w['CS'] << 16 | w['IP'])
        cmd = c['Writers'][key]
        assert cmd['NativeBytes'] == w['Count'] and cmd['FallbackBytes'] == 0
        original = next(v for v in previous['Writers'] if (v['CS'], v['IP']) == (w['CS'], w['IP']))
        for k in ('Count', 'FirstStep', 'LastStep', 'Lo', 'Hi', 'Code'):
            assert w[k] == original[k]
        paired = next(v for v in legacy['Writers'] if (v['CS'], v['IP']) == (w['CS'], w['IP']))
        for k in ('Count', 'FirstStep', 'LastStep', 'Lo', 'Hi', 'Code', 'Registers', 'Stack'):
            assert w[k] == paired[k]
        witness = sha((folders[0]/w['Memory']).read_bytes())
        assert witness == sha((folders[1]/paired['Memory']).read_bytes())
        writers.append(dict(cs_ip=f"{w['CS']:04X}:{w['IP']:04X}", witness_memory_sha256=witness, **cmd))
    # Input provenance, without private memory or stack dumps.
    result = dict(scope='正常冷啟動SND→CONTROL→VACE→GAME；7400更新；空白鍵900／1800；新舊成對固定暫存檔時間；另比歷史終點；不代表未觀測路線',
                  fixed_scratch_file_time='2000-01-01T00:00:00Z',
                  witness_comparison='十份完整記憶體及暫存器／堆疊，固定檔案時間的原指令與命令攔截逐位元組相同；不遮蔽位址',
                  frames=7400, steps=native['Digest']['Steps'], indexed_retraces=c['Frames'],
                  digest={k: bytes(v).hex() if isinstance(v, list) else v for k, v in native['Digest'].items()},
                  historical_input_digest={k: bytes(v).hex() if isinstance(v, list) else v for k, v in previous['Digest'].items()},
                  historical_input_endpoint_equal=True,
                  input_metadata_boundary='固定檔案日期的兩側比較完整RAM；歷史日期輸入另由未固定時間的兩側比較完整終點。不同日期的RAM不宣稱相同。',
                  command_count=c['Commands'], native_bytes=c['NativeBytes'], fallback_bytes=0,
                  canvas_mismatches=0, writers=writers, audio_sha256=native['AudioSHA256'],
                  audio_events=native['AudioEvents'], guest_files=len(tree(folders[0] / 'save')),
                  dos_events=native['DOSEvents'], dos_sha256=native['DOSSHA256'], dos_events_equal=True,
                  dos_scope='直接INT21請求的步數、定位、完整16位暫存器／旗標、DS:DX起256 bytes雜湊及立即返回狀態；不寫入客體、不對外保存參數原文',
                  original_state_equal=True, guest_state_equal=True, audio_events_equal=True,
                  source_report_sha256={p.name: sha((p/'report.json').read_bytes()) for p in folders},
                  historical_pair_report_sha256={name:sha((a.root/name/'report.json').read_bytes()) for name in ('native-demo','legacy-demo')},
                  writer_proof_sha256=sha(a.writers.read_bytes()),
                  tool_sha256=sha(Path(__file__).read_bytes()),
                  trace_source_sha256=sha((a.root.parents[2]/'tools/draw_trace_probe.go').read_bytes()),
                  timing={'native_seconds': native['Seconds'], 'legacy_seconds': legacy['Seconds'],
                          'scope': '同時執行的完整探針成本，含追蹤與獨立畫布核對；不作效能結論'})
    a.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
