#!/usr/bin/env python3
"""核對實際推廣片時長、音訊、預期靜態段、來源與文字安全區。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def image(p):return np.array(Image.open(p).convert('RGBA'))


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','work','receipt'):ap.add_argument('--'+key,type=Path,required=True)
    a=ap.parse_args();r,w=a.root,a.work;p=w/'promo';m=read(p/'storyboard.json');probe=read(p/'ffprobe.json');video=p/'pod-v.0.1.2-20261102-promo.mp4'
    assert m['duration']==63 and float(probe['format']['duration'])==63 and video.stat().st_size>1_000_000
    v,audio=probe['streams'];assert v['codec_type']=='video' and v['codec_name']=='h264' and (v['width'],v['height'])==(1920,1080)
    assert v['r_frame_rate']=='25/1' and v['nb_frames']=='1575' and float(v['duration'])==63
    assert audio['codec_type']=='audio' and audio['codec_name']=='aac' and audio['sample_rate']=='48000' and audio['channels']==2 and float(audio['duration'])==63
    log=(p/'audio-check.log').read_text();assert 'silence_start:' not in log
    mean=float(re.search(r'mean_volume: ([-0-9.]+) dB',log)[1]);peak=float(re.search(r'max_volume: ([-0-9.]+) dB',log)[1])
    assert -35<mean<-10 and -15<peak<-.1
    log=(p/'image-check.log').read_text();assert 'black_start:' not in log
    starts=list(map(float,re.findall(r'freeze_start: ([0-9.]+)',log)));ends=list(map(float,re.findall(r'freeze_end: ([0-9.]+)',log)))
    allowed=[(0,6),(6,14),(14,21),(21,25),(25,29),(44,51),(57,63)]
    assert len(starts)==len(ends)==len(allowed)==7
    for start,end,(lo,hi) in zip(starts,ends,allowed):assert lo<=start<end<=hi
    assert not any(start<37 and end>29 for start,end in zip(starts,ends)),'control clip frozen'
    for src,info in m['sources'].items():assert sha(r/src)==info['sha256'],src
    for b in m['text_safe_bounds']:
        x,y,width,height=b['rect'];sx,sy,sw,sh=b['safe'];assert sx<=x and sy<=y and x+width<=sx+sw and y+height<=sy+sh
    capture=w/'capture/intro';assert sha(capture/'audio.s16le')==m['capture']['audio_sha256'] and sha(capture/'opl.jsonl')==m['capture']['opl_sha256']
    metrics=read(capture/'audio-metrics.json');assert metrics['seconds']==85 and metrics['peak']<32767 and metrics['rms']>.01 and metrics['opl_events']==6660
    # F12原版畫面逐像素吻合實際來源；二倍只比較未翻譯的上方美術區。
    checked=[]
    for n,scale,full in [(622,3,True),(650,3,True),(1302,2,False),(1350,2,False)]:
        prefix=capture/f'sources/pod-f{n:06d}'
        idx=np.frombuffer(prefix.with_suffix('.idx').read_bytes(),np.uint8).reshape(200,320)
        pal=np.frombuffer(prefix.with_suffix('.pal').read_bytes(),np.uint8).reshape(256,3)
        raw=pal[idx].repeat(scale,0).repeat(scale,1);im=image(capture/f'window/pod-f{n:06d}.png')
        center=im[:,80*scale:400*scale,:3]
        assert np.array_equal(center if full else center[:168*scale],raw if full else raw[:168*scale])
        checked.append(dict(frame=n,scale=scale,F12=full,source_indexed_sha256=sha(prefix.with_suffix('.idx')),palette_sha256=sha(prefix.with_suffix('.pal'))))
    assert len(list((p/'review').glob('f*.png')))==9
    result=dict(date='2026-11-02',issue=78,status='PASS private promo technical and visual scope',version=m['version'],
        video=dict(path=video.name,sha256=sha(video),size=video.stat().st_size,duration=63,width=1920,height=1080,fps=25,frames=1575),
        audio=dict(codec='aac',sample_rate=48000,channels=2,mean_dB=mean,max_dB=peak,unexpected_silence=False,source=metrics,human_listening=False),
        image_checks=dict(unexpected_black=0,unexpected_freeze=0,black_min_seconds=.2,black_pixel_threshold=.02,black_picture_ratio=.98,
            freeze_min_seconds=2,freeze_noise_dB=-60,expected_stills=[dict(start=s,end=e) for s,e in zip(starts,ends)]),
        controls=checked,all_text_contained=True,text_rectangles=len(m['text_safe_bounds']),storyboard=m,
        ffprobe=probe,files={str(q.relative_to(p)):sha(q) for q in [p/'contact.png',p/'ffprobe.json',p/'audio-check.log',p/'image-check.log',*sorted((p/'review').glob('f*.png'))]},
        visual_review=dict(reviewer='Codex',contact_sha256=sha(p/'contact.png'),scope='nine encoded samples, title/captions, four layouts and full closing card'),
        rights=m['rights'],limits=m['limits']+['Audio technical checks do not constitute human listening approval'])
    a.receipt.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS 63s1080p, full closing, audio, controls, nine visual samples; black0 unexpected freeze0')


if __name__=='__main__':main()
