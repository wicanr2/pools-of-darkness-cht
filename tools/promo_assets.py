#!/usr/bin/env python3
"""用實際遊戲畫面製作字卡與有界操作片段，不改畫面內容或冒充遊戲UI。"""
import argparse
import hashlib
import json
import re
from pathlib import Path
from PIL import Image,ImageDraw,ImageFont
import numpy as np


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--work',type=Path,required=True)
    ap.add_argument('--version',default='v.0.1.2-20261102');ap.add_argument('--player',type=Path);a=ap.parse_args()
    assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}',a.version)
    r,w=a.root,a.work;o=w/'promo/assets';assert not o.exists();o.mkdir()
    W,H=1920,1080;fg='#e9e4d8';dim='#a7c7c6';accent='#55c9c7'
    fontpath=r/'font/source/rounded/GenSenRounded2TW-B.otf'
    font=lambda n:ImageFont.truetype(str(fontpath),n)
    sources={};bounds=[]
    def relative(path):
        if path.is_relative_to(r):return str(path.relative_to(r))
        return 'workplace/package-promo-v1/'+str(path.relative_to(w))
    def text(im,words,y,size=46,color=fg):
        d=ImageDraw.Draw(im);f=font(size);box=d.textbbox((0,0),words,font=f);width=box[2]-box[0];height=box[3]-box[1]
        x=(W-width)//2;assert 90<=x and x+width<=1830 and 34<=y and y+height<1040,(words,box)
        d.text((x-box[0],y-box[1]),words,font=f,fill=color)
        bounds.append(dict(text=words,rect=[x,y,width,height],safe=[80,30,1760,1020]))
    def bg():
        y=np.linspace(0,1,H)[:,None,None];top=np.array([9,25,30]);bottom=np.array([3,7,12]);rgb=np.broadcast_to((top*(1-y)+bottom*y).astype(np.uint8),(H,W,3)).copy()
        im=Image.fromarray(rgb);d=ImageDraw.Draw(im);d.line((90,95,1830,95),fill=accent,width=3);d.line((90,1006,1830,1006),fill='#23454a',width=2);return im
    def src(path,crop=False):
        assert path.is_file(),path
        key=relative(path);sources[key]=dict(sha256=sha(path),size=list(Image.open(path).size),kind='actual game renderer output')
        im=Image.open(path).convert('RGB')
        if crop:
            assert im.width*5==im.height*12
            box=(im.width//6,0,im.width*5//6,im.height)
            sources[key]['crop']=list(box);im=im.crop(box)
        return im
    def frame(path,title,caption,note='實際遊戲畫面',crop=False):
        im=bg();text(im,title,42,54,accent);s=src(path,crop);scale=min(1720/s.width,718/s.height)
        size=(round(s.width*scale),round(s.height*scale));s=s.resize(size,Image.Resampling.NEAREST)
        x=(W-size[0])//2;y=150+(718-size[1])//2;im.paste(s,(x,y));d=ImageDraw.Draw(im);d.rectangle((x-2,y-2,x+size[0]+1,y+size[1]+1),outline='#437c84',width=2)
        text(im,caption,905,40);text(im,note,974,26,dim);return im
    def save(im,name):im.save(o/name,compress_level=2)
    cap=w/'capture/intro/window';final=r/'workplace/probe/final-hd/runtime-v2/hd/window'
    im=bg();cover=src(r/'docs/screenshots/game-cover-hd.png');cover.thumbnail((1100,690),Image.Resampling.NEAREST);im.paste(cover,((W-cover.width)//2,130))
    text(im,'1991 年的經典，重新呈現。',42,58);text(im,'Pools of Darkness',840,52,accent);text(im,'繁體中文化與 HD',923,44)
    if not a.version.startswith('v.0.'):text(im,a.version+' 正式版',980,28,dim)
    save(im,'s00.png')
    save(frame(cap/'pod-f001362.png','重返菲蘭','你們曾經拯救的家園，再度被黑暗籠罩。','背景整理自本作手札導言與原版開場'),'s01.png')
    player=a.player or r/'workplace/multilingual-jako/gui-zh-TW/window'
    player_note='正式版程式的正常讀檔路線畫面' if a.player else '已驗正常玩家讀檔路線畫面，來源收據034'
    save(frame(player/'pod-f004200.png','再次踏上冒險','組成隊伍，探索城鎮，在回合制戰鬥中迎戰。',player_note),'s02.png')
    map_source=player/'pod-f004320.png' if a.player else r/'docs/screenshots/map-east-phlan.png'
    save(frame(map_source,'線索留在遊戲裡','修正手札、攻略摘要與地圖，減少來回翻書。','地圖介面實際畫面，僅裁出中央遊戲區域',crop=True),'tools-map.png')
    save(frame(player/'pod-f004346.png','線索留在遊戲裡','在遊戲內查閱攻略，保留已知地點與出處。',player_note),'tools-guide.png')
    # 兩段都是正式前端的實際F12／F2事件；抽樣時間依60次更新／秒換算。
    switch=o/'switch';switch.mkdir()
    first=sorted(n for n in range(550,751,4));second=sorted(n for n in range(1250,1421,4))
    for i in range(200):
        t=i/25
        if t<.8:n=first[0];label='F12 切回英文原版'
        elif t<4.2:n=first[min(len(first)-1,int((t-.8)*15))];label='F12 切回英文原版'
        elif t<7.08:n=second[min(len(second)-1,int((t-4.2)*15))];label='F2 切換二倍原版與三倍 HD'
        else:n=second[-1];label='F2 切換二倍原版與三倍 HD'
        save(frame(cap/f'pod-f{n:06d}.png','隨時切換呈現',label,'實際前端操作擷取，未重啟遊戲'),f'switch/{i:04d}.png')
    langs=[('zh-TW',1992,'繁體中文'),('en',1952,'英文'),('ja',1962,'日文'),('ko',1972,'韓文'),('zh-CN',1982,'簡體中文')]
    for i,(code,n,name) in enumerate(langs):save(frame(cap/f'pod-f{n:06d}.png','F8 切換五種語言',name,'日文、韓文為機器翻譯並經語境校讀'),f'lang{i}.png')
    save(frame(r/'workplace/probe/final-hd/runtime-v2/demo/frame.png','經典戰鬥，重新呈現','保留原版規則，重繪怪物、場景與已辨識特效。'),'s06.png')
    for i,n in enumerate((10,1000,1532)):save(frame(final/f'pod-f{n:06d}.png','從片頭到片尾','沿原版畫格與構圖，呈現 HD 動畫。','片尾播放器直接入口畫面展示'),f'ending{i}.png')
    im=bg();text(im,'黑暗之池',220,112);text(im,'繁體中文化與 HD',390,64,accent);text(im,'Linux   Windows   macOS',535,50)
    text(im,'自備原版，重返月之海。',670,54);text(im,a.version+('  私人試玩候選' if a.version.startswith('v.0.') else '  正式版'),790,35,dim)
    text(im,'Windows 為 Wine 抽驗；macOS 真機尚未驗證。',906,28,dim);save(im,'s08.png')
    scenes=[dict(id=i,duration=d,kind=k) for i,d,k in [(0,6,'title'),(1,8,'story'),(2,7,'game'),(3,8,'panels'),(4,8,'actual controls'),(5,7,'languages'),(6,7,'battle'),(7,6,'final direct player'),(8,6,'closing')]]
    manifest=dict(version=a.version,width=W,height=H,fps=25,duration=sum(s['duration'] for s in scenes),scenes=scenes,
        theme=dict(deep='#03070c',pool='#55c9c7',text=fg,motif='original stone frame retained, pool cyan cards'),sources=sources,
        text_safe_bounds=bounds,all_text_contained=True,font=dict(path=str(fontpath.relative_to(r)),sha256=sha(fontpath)),
        capture=dict(metadata=relative(w/'capture/capture-build.json'),script_sha256=sha(w/'capture/intro/script.txt'),
            audio_sha256=sha(w/'capture/intro/audio.s16le'),opl_sha256=sha(w/'capture/intro/opl.jsonl'),audio_metrics=json.loads((w/'capture/intro/audio-metrics.json').read_text())),
        rights='Local private promo. Original game images/music not cleared for public distribution. No upload or Git inclusion.',
        limits=['Not a completed campaign; final footage uses direct original VACE entry','No newly arranged music or native macOS validation','Cards and panel samples are intentional still sequences'])
    (w/'promo/storyboard.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
    contact=Image.new('RGB',(1440,810));names=['s00.png','s01.png','s02.png','tools-map.png','switch/0030.png','lang2.png','s06.png','ending2.png','s08.png']
    for i,name in enumerate(names):im=Image.open(o/name);im.thumbnail((480,270));contact.paste(im,((i%3)*480,(i//3)*270))
    contact.save(w/'promo/design-contact.png');print('PASS cards and source manifest',len(sources),'sources,',len(bounds),'text rectangles,',manifest['duration'],'seconds')


if __name__=='__main__':main()
