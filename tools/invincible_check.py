#!/usr/bin/env python3
"""核對F1切片的正常路線、受控原版呼叫與明列閃現限制。"""
import argparse,hashlib,json,re,sys
from pathlib import Path
import numpy as np
from PIL import Image

sys.path.insert(0,str(Path(__file__).resolve().parent))
from zhcn_check import state
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def tree(p):return {str(x.relative_to(p)):sha(x)for x in sorted(p.rglob('*'))if x.is_file()}
def pix(p):return np.array(Image.open(p).convert('RGBA'))
def flash(log):
 return dict(raw=int(re.search(r'閃現 (\d+) 次',log)[1]),events=re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log))
def guest(p):
 front={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
 return {n:h for n,h in tree(p).items()if n not in front}
def main():
 ap=argparse.ArgumentParser(description=__doc__)
 for n in ('root','out','receipt'):ap.add_argument('--'+n,type=Path,required=True)
 a=ap.parse_args();r=a.root;o=a.out;parent=read(r/'docs/re/028-english-panels-runtime.json');old=read(r/'docs/re/027-zhcn-runtime.json')
 assert sha(r/'workplace/dosgolem/workplace/bin/pod-play-before-invincible')==parent['binary_sha256']
 assert sha(r/'workplace/dosgolem/workplace/bin/pod-play')==sha(o/'pod-play')
 logs={n:(o/f'gui-{n}/run.log').read_text()for n in ('off','city-on','demo','audit','baseline')}
 states={n:state(s)for n,s in logs.items()}
 assert states['off']==states['city-on']==parent['baseline']['state']
 assert guest(o/'gui-off/game')==guest(o/'gui-city-on/game')==old['runtime']['zh-TW']['guest']
 assert states['demo']==states['audit']
 assert guest(o/'gui-demo/game')==guest(o/'gui-audit/game')==guest(o/'gui-baseline/game')
 counters={}
 for name in ('city-on','demo','audit'):
  log=logs[name];m=re.search(r'無敵 開啟true、還原 (\d+) 次、移除有害效果 (\d+) 筆、拒絕 (\d+) 次',log);assert m
  counters[name]=[int(x)for x in m.groups()];assert counters[name][2]==0
 assert counters['city-on']==[0,0,0]and counters['demo']==counters['audit']and counters['demo'][0]>0
 assert all('中止 1 次、中止後返回 1 次' in logs[n]for n in ('demo','audit','baseline'))
 for name,log in logs.items():
  assert '字型缺字' not in log and '截圖失敗' not in log
  if name!='baseline':
   draw=read(o/f'gui-{name}/draw-report.json');assert draw['NativeBytes']==draw['HostBytes'] and draw['FallbackBytes']==draw['Mismatches']==draw['LegacyRGBFrames']==0
 for n in ('off','city-on'):
  assert flash(logs[n])['raw']==121
  for p in (o/f'gui-{n}/shots').glob('*.png'):assert np.array_equal(pix(p),pix(r/'workplace/english-panels/gui/shots'/p.name))
  for p in (o/f'gui-{n}/window').glob('*.png'):
   diff=np.any(pix(p)!=pix(r/'workplace/english-panels/gui/window'/p.name),axis=2)
   if p.name=='pod-f004770.png':diff[8:26,697:731]=False
   assert not np.any(diff),(n,p.name)
 # 較長戰鬥路線的量尺事件如實保存，是否共同發生由新舊程式對照判定。
 flashes={n:flash(s)for n,s in logs.items()}
 assert flashes['demo']==flashes['audit']
 assert flashes['audit']==flashes['baseline']and flashes['audit']['raw']==19
 cells=[]
 for first,last,row,cols in flashes['audit']['events']:
  first,last,row=int(first),int(last),int(row)
  aa=[];bb=[]
  for frame in (first,last):
   stem=f'pod-f{frame:06d}'
   ai=np.frombuffer((o/'gui-audit/sources'/(stem+'.idx')).read_bytes(),np.uint8).reshape(200,320)
   bi=np.frombuffer((o/'gui-baseline/sources'/(stem+'.idx')).read_bytes(),np.uint8).reshape(200,320)
   assert sha(o/'gui-audit/sources'/(stem+'.pal'))==sha(o/'gui-baseline/sources'/(stem+'.pal'))
   ar=pix(o/'gui-audit/shots'/(stem+'.png'));br=pix(o/'gui-baseline/shots'/(stem+'.png'))
   aa.append(ai)
   for col in map(int,cols.split()):
    assert np.array_equal(ai[row*8:row*8+8,col*8:col*8+8],bi[row*8:row*8+8,col*8:col*8+8])
    assert np.array_equal(ar[row*24:row*24+24,col*24:col*24+24],br[row*24:row*24+24,col*24:col*24+24])
  for col in map(int,cols.split()):
   before=aa[0][row*8:row*8+8,col*8:col*8+8];after=aa[1][row*8:row*8+8,col*8:col*8+8]
   kind='existing HD occlusion question; unresolved'
   if first==8191:
    assert np.unique(before).tolist()==[32]and np.any(before!=after)
    kind='camera change from flat floor index32 to new wall'
   cells.append(dict(first=first,last=last,row=row,col=col,classification=kind,source_pixels_changed=int(np.count_nonzero(before!=after)),same_indices_palette_and_presented_pixels_as_old_binary=True))
 assert len(cells)==19 and sum(x['first']==8191 for x in cells)==13
 oracle=read(o/'oracle.json');assert len(oracle)==7
 assert all(x['HP']==162 and x['Health']==0 for x in oracle if x['On']and x['Target']=='player')
 assert oracle[-1]['Target']=='enemy'and oracle[-1]['HP']==0 and oracle[-1]['Health']==6
 for name in ('unit.log','oracle.log','frontend.log'):assert '\nok ' in '\n'+(o/name).read_text()
 assert not(o/'vet.log').read_text().strip()
 assert sha(o/'pod-play')==sha(o/'clean-pod-play')
 lock=read(r/'review/hd/runtime/manifest.json');theme=tree(r/'workplace/hd/theme')
 assert theme=={x['path']:x['sha256']for x in lock['theme']}
 assert len(theme)==1536
 assert hashlib.sha256(json.dumps(theme,sort_keys=True).encode()).hexdigest()==old['assets']['theme_tree_sha256']
 for n,h in old['assets']['fonts'].items():assert sha(r/'workplace/font'/n)==h
 proof=dict(date='2026-11-01',issue=72,status='PASS within F1 sample scope; long-combat flash audit retained separately',engine=lock['engine']['commit'],binary_sha256=sha(o/'pod-play'),engine_tree=(o/'engine-tree.txt').read_text().strip(),
  base_engine='2ec2dfa4616d47795f27c1f6f9eee92999fb4ac9',source_proof=dict(path='docs/re/029-invincible-source.json',sha256=sha(r/'docs/re/029-invincible-source.json')),
  normal=dict(states=states,counters=counters,guest_files={n:len(guest(o/f'gui-{n}/game'))for n in logs},guest_equal_to028_when_off_or_city_on=True,DEMO_guest_equal=True,
   scripts={n:sha(o/f'gui-{n}/script.txt')for n in logs},screenshots={n:{x.name:sha(x)for x in sorted((o/f'gui-{n}/window').glob('*.png'))}for n in logs},flashes=flashes,
   long_combat_cells=cells,camera_change_false_positives=13,existing_unresolved_HD_occlusion_cells=6,F1_flash_regression_in_these_cells=False),
  controlled_original_damage=dict(kind='explicit composition: original resident code from town snapshot plus battle records; not normal gameplay',results=oracle),
  assets=dict(theme_files=1536,fonts=10,theme_and_fonts_unchanged=True),clean_git_build_matches=True,
  logs={n:sha(o/n)for n in ('unit.log','oracle.log','frontend.log','vet.log')},
  engine_files={n:sha(r/'workplace/dosgolem'/n)for n in ('apps/pod/cheat.go','apps/pod/cheat_test.go','cmd/pod-play/main.go')},
  tools={p.name:sha(p)for p in sorted((r/'tools').glob('*invincible*'))},
  limits=['sampled routes only; harmful effect types outside verified five retained','effect removal uses existing unlink policy without reclaiming original Pascal nodes','long-combat raw flash events retained; no blanket zero-flash claim','no change to old nine packages; remaining original project scope pending'])
 a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
 print(json.dumps(dict(status=proof['status'],counters=counters,flash={n:d['raw']for n,d in flashes.items()},same_long_combat_events=flashes['audit']==flashes['baseline']),indent=2))
if __name__=='__main__':main()
