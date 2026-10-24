"""英文新字型的既有正常DEMO對話抽樣，不將其擴大為全英文面板完成。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def state(folder):
    log=(folder/'run.log').read_text()
    value=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    assert value and value[1]=='2210'and value[2]=='212712500'
    return value.groups(),log


def guest(folder):
    frontend={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
    return {str(p.relative_to(folder/'save')):sha(p)for p in (folder/'save').rglob('*')if p.is_file()and str(p.relative_to(folder/'save'))not in frontend}


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out',type=Path,required=True);ap.add_argument('--baseline',type=Path,required=True)
    a=ap.parse_args();current,log=state(a.out);before,_=state(a.baseline)
    assert current==before,'英文正常DEMO改變原版步數／記憶體／色號'
    assert guest(a.out)==guest(a.baseline),'DOS檔案改變'
    assert 'pod-play: 閃現 0 次'in log and '字型缺字'not in log
    assert re.search(r'pod-play: A map\[[^\n]*覆繪:',log)and re.search(r'pod-play: B map\[[^\n]*覆繪:',log)
    rows=[]
    for frame in (2050,2200):
        p=a.out/'shots'/f'pod-f{frame:06d}.png';image=np.array(Image.open(p).convert('RGB'))
        idx=np.frombuffer((a.out/'sources'/f'pod-f{frame:06d}.idx').read_bytes(),np.uint8).reshape(200,320)
        pal=np.frombuffer((a.out/'sources'/f'pod-f{frame:06d}.pal').read_bytes(),np.uint8).reshape(256,3)
        base=pal[idx].repeat(3,0).repeat(3,1);diff=np.any(image!=base,axis=2)
        diff[72:336,72:336]=False
        changed=int(np.count_nonzero(diff));assert changed>0,'英文對話沒有新字型覆繪'
        rows.append(dict(frame=frame,sha256=sha(p),changed_pixels_outside_portrait=changed))
    result=dict(all_pass=True,scope='英文前端及2210格正常DEMO對話；英文手札、攻略及其他語系未完成',
        original_state=current,original_memory_index_equal=True,guest_files=len(guest(a.out)),guest_files_equal=True,
        flashes=0,missing_glyphs=False,screenshots=rows,binary_sha256=json.loads((a.out/'inputs.json').read_text())['binary_sha256'],
        baseline_run_log_sha256=sha(a.baseline/'run.log'),run_log_sha256=sha(a.out/'run.log'))
    (a.out/'summary.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(result,ensure_ascii=False,indent=2))


if __name__=='__main__':main()
