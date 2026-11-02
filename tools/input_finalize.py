#!/usr/bin/env python3
"""彙整本輪正常玩家路徑、實際滑鼠與原始位址證據。"""
import argparse,hashlib,json,re,subprocess
from pathlib import Path
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def state(p):
    text=p.read_text();a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',text)
    b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',text);assert a and b;return list(a.groups())+list(b.groups())
def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--work',type=Path,required=True);a=ap.parse_args()
    assert Path('/.dockerenv').exists();r,w=a.root,a.work
    runs={n:read(w/n/'pass.json') for n in ['auto-fixed','manual-fixed','retry-fixed','mouse-2','mouse-3']}
    assert all(x['status']=='PASS' for x in runs.values())
    auto,manual=state(w/'auto-fixed/play.log'),state(w/'manual-fixed/play.log');assert auto==manual
    assert runs['auto-fixed']['save']==runs['manual-fixed']['save']
    assert runs['auto-fixed']['fills']==[['3505','YOU']] and runs['manual-fixed']['fills']==[]
    assert runs['retry-fixed']['fills']==[['4200','SURVIVE']]
    old=read(w/'old-packages.json')
    for row in old:assert sha(r/row['path'])==row['sha256'] and (r/row['path']).stat().st_size==row['size']
    assert len(old)==24
    inputdb=r/'workplace/probe/draw-intercept-v1/ida-game'
    inputmeta=read(inputdb/'game-1b2e.bin.json');mousemeta=read(inputdb/'game-1d1f.bin.json')
    src=dict(grade='confirmed in sampled original executable paths',ida_version='9.4',
        input=dict(binary=inputmeta['input'],sha256=inputmeta['sha256'],runtime_cs=inputmeta['runtime_cs'],ida_base=0,
                   rows=[x for x in read(w/'ida-input.json')['rows'] if 0x711<=x['offset']<=0x789]),
        mouse=dict(binary=mousemeta['input'],sha256=mousemeta['sha256'],runtime_cs=mousemeta['runtime_cs'],ida_base=0,
                   runtime_memory_file='workplace/input-fix-v1/menu.mem',runtime_memory_sha256=sha(w/'menu.mem'),
                   address_space=read(w/'ida-mouse.json')['address_space'],
                   rows=[x for x in read(w/'ida-mouse.json')['rows'] if 0x138<=x['ida_ea']<=0x15c or 0x10072<=x['ida_ea']<=0x100c6]),
        quiz=dict(source_memory='workplace/probe/g7-quiz/mem.bin',sha256=sha(r/'workplace/probe/g7-quiz/mem.bin'),
                  address_space='runtime overlay offset; historical overlay CS=5509; relocated resident from bound path A',
                  call_offset='02B9',call_bytes='9a11072e1b',return_offset='02BE',answer='DS:7DEE',input='SS:BP-12Bh',committed_input='DS:7EEE'))
    (r/'docs/re/038-input-source.json').write_text(json.dumps(src,ensure_ascii=False,indent=2)+'\n')
    engine=subprocess.check_output(['git','-C',str(r/'workplace/dosgolem'),'rev-parse','HEAD'],text=True).strip()
    tests=(w/'tests-commit.log').read_text();assert tests.count('ok  ')==4 and 'FAIL' not in tests
    result=dict(date='2026-11-02',issue=79,status='PASS input sampled scope',engine_commit=engine,
        candidate_binary_sha256=sha(w/'pod-play-commit'),normal_paths=runs,
        auto_manual_same_state=auto,auto_manual_saves_equal=True,auto_waits_for_enter=True,
        F12_autofill=True,five_locale_quiz_display=True,right_button_original_semantics='Esc; left Enter',
        cursor='native OS cursor above canvas; no software cursor restoration',
        tests=dict(log='workplace/input-fix-v1/tests-commit.log',sha256=sha(w/'tests-commit.log'),packages=['host','session','apps/pod Quiz|Input','cmd/pod-play']),
        source_proof_sha256=sha(r/'docs/re/038-input-source.json'),old_packages=old,
        limits=['Native Windows/macOS input not tested','Focus cleanup covered by host/session tests; GUI captures cover clicks and outside drag','No full campaign; new music remains separate'],
        tools={n:sha(r/'tools'/n) for n in ['input_receipt.py','input_mouse_gui.py','input_finalize.py']})
    (r/'docs/re/038-input-runtime.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS input, auto/manual state and saves equal, 24 old packages unchanged')
if __name__=='__main__':main()
