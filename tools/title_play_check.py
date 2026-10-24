"""核對三張開場正常冷啟動、語系切換與原版狀態。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def rgba(p):return np.array(Image.open(p).convert('RGBA'))


def raw(folder,frame,scale=3):
    p=folder/'sources'/f'pod-f{frame:06d}'
    idx=np.frombuffer(p.with_suffix('.idx').read_bytes(),np.uint8).reshape(200,320)
    pal=np.frombuffer(p.with_suffix('.pal').read_bytes(),np.uint8).reshape(256,3)
    rgb=pal[idx].repeat(scale,0).repeat(scale,1)
    return np.concatenate((rgb,np.full((*rgb.shape[:2],1),255,np.uint8)),axis=2)


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--out',type=Path,required=True);ap.add_argument('--theme',type=Path,required=True)
    a=ap.parse_args();states={};inputs={};guests={}
    for case in ('hd','base','switch'):
        folder=a.out/case;log=(folder/'run.log').read_text()
        end=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
        digest=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
        assert end and digest and int(end[1])==1250
        states[case]=end.groups()+digest.groups()
        inputs[case]=json.loads((folder/'inputs.json').read_text())
        frontend={'pod-map-visited.txt','pod-journal-seen.txt','pod-walkthrough-pos.txt','pod-lang.txt','pod-speed.txt','pod-barfont.txt'}
        guests[case]={str(p.relative_to(folder/'save')):sha(p)for p in (folder/'save').rglob('*')if p.is_file()and str(p.relative_to(folder/'save'))not in frontend}
        if case=='hd':assert 'pod-play: 閃現 0 次' in log,'首次開場閃現未歸零'
    assert states['hd']==states['base']==states['switch'],'原版CPU／記憶體／色號／色盤不同'
    assert inputs['hd']==inputs['base']==inputs['switch'],'輸入不同'
    assert guests['hd']==guests['base']==guests['switch'],'DOS檔案不同'
    screenshots=[]
    for key,frame in [('title-001',1000),('title-002',1100),('title-003',1200)]:
        for case in ('hd','switch'):
            p=a.out/case/'shots'/f'pod-f{frame:06d}.png'
            assert np.array_equal(rgba(p),rgba(a.theme/'localized/zh-TW'/f'{key}.png')),(case,key,'繁中完整成品不符')
            screenshots.append(dict(case=case,key=key,frame=frame,sha256=sha(p),locale='zh-TW'))
        p=a.out/'base/shots'/f'pod-f{frame:06d}.png'
        assert np.array_equal(rgba(p),raw(a.out/'base',frame)),'無HD不是配對原版'
        p=a.out/'switch/shots'/f'pod-f{frame+6:06d}.png'
        assert np.array_equal(rgba(p),rgba(a.theme/'localized/en'/f'{key}.png')),(key,'F8英文成品不符')
        screenshots.append(dict(case='switch',key=key,frame=frame+6,sha256=sha(p),locale='en'))
    folder=a.out/'switch'
    assert np.array_equal(rgba(folder/'shots/pod-f001211.png'),rgba(a.theme/'localized/zh-TW/title-003.png'))
    assert np.array_equal(rgba(folder/'shots/pod-f001216.png'),raw(folder,1216,2)),'2倍仍有開場HD'
    assert np.array_equal(rgba(folder/'shots/pod-f001221.png'),rgba(a.theme/'localized/zh-TW/title-003.png'))
    assert np.array_equal(rgba(folder/'shots/pod-f001226.png'),raw(folder,1226)),'F12不是原版'
    assert np.array_equal(rgba(folder/'shots/pod-f001231.png'),rgba(a.theme/'localized/zh-TW/title-003.png'))
    poster_checks=[]
    for case in ('hd','base','switch'):
        for p in sorted((a.out/case/'window-shots').glob('*.png')):
            frame=int(p.stem.split('f')[-1]);window=rgba(p)
            scale=window.shape[0]//200
            locale='en'if case=='switch'and frame in (1006,1106,1206,1226)else'zh-TW'
            assert np.array_equal(window[:,80*scale:400*scale],rgba(a.out/case/'shots'/p.name)), '完整視窗中央不同'
            for side,offset in [('left',0),('right',400*scale)]:
                base=Image.open(a.theme/'posters'/f'poster-{side}-{scale}x.png').convert('RGBA')
                text=Image.open(a.theme/'posters'/f'poster-text-{side}.{locale}-{scale}x.png').convert('RGBA')
                expected=np.array(Image.alpha_composite(base,text))
                assert np.array_equal(window[:,offset:offset+80*scale],expected),(case,frame,side,'海報語系不符')
            poster_checks.append(dict(case=case,frame=frame,locale=locale,scale=scale,sha256=sha(p)))
    report=dict(all_pass=True,scope='正常冷啟動三張TITLE、F8繁中英文、F2、F12；未知語系／改色／底列另見真實畫格測試',
        state=states['hd'],guest_files=len(guests['hd']),guest_state_equal=True,
        binary_sha256=inputs['hd']['binary_sha256'],inputs_sha256=hashlib.sha256(json.dumps(inputs['hd'],sort_keys=True).encode()).hexdigest(),
        flashes=0,screenshots=screenshots,poster_checks=poster_checks)
    (a.out/'summary.json').write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(report,ensure_ascii=False,indent=2))


if __name__=='__main__':main()
