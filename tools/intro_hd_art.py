#!/usr/bin/env python3
"""片頭HD相位合成。來源定位見研究031；材質變形只屬美術近似。"""
import argparse
import hashlib
import json
from pathlib import Path

import cv2
import numpy as np
from PIL import Image, ImageDraw


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root', type=Path, required=True)
    ap.add_argument('--out', type=Path, required=True)
    a = ap.parse_args()
    root, out = a.root, a.out
    if out.exists():
        raise ValueError('輸出必須是新目錄')
    cv2.setNumThreads(1)
    req = root/'workplace/probe/intro-hd-v1/requests-v1'
    raw = root/'workplace/hd/out/intro74'
    old = root/'review/hd/b44-intro'
    rows = json.loads((root/'docs/re/031-intro-bma-requests.json').read_text())['items']
    src, rgb = {}, {}
    for r in rows:
        key = (r['block'], r['frame'])
        src[key] = np.frombuffer((req/r['indexed']).read_bytes(), np.uint8).reshape(168, 320)
        pal = np.frombuffer((req/r['palette']).read_bytes(), np.uint8).reshape(256, 3)
        rgb[key] = pal[src[key]]
    # 重複的完整色號圖面共用同一美術，不以影格編號臆造動作。
    import sys
    sys.path.insert(0, str(root/'tools'))
    from intro_bma import decode, states
    for f, before, after in states(decode((root/'workplace/input/dos/pod/INTRO.BMA').read_bytes())):
        key = (f.block, f.index)
        if key not in src:
            arr = np.frombuffer(after[:53760], np.uint8).reshape(168,320)
            equal = next(k for k,v in src.items() if np.array_equal(v, arr))
            src[key], rgb[key] = src[equal], rgb[equal]
    art = {name: np.array(Image.open(old/f'intro-{name}.png').convert('RGB'))
           for name in ('council','phlan','horned','heroes','followers','dreadlord')}
    plates = {name: np.array(Image.open(raw/f'{name}-raw.png').convert('RGB').resize((960,504),Image.Resampling.LANCZOS))
              for name in ('council-empty','horned-empty','horned-cage','followers-burnt','dreadlord-empty','dreadlord-fireball')}
    # 新背景只替換原版動作區域；核准石框與其他不動區域逐位元組保留。
    regions = {'council-empty':(58,46,252,106),'horned-empty':(91,25,240,134),'horned-cage':(91,25,240,134),
               'followers-burnt':(63,34,247,133),'dreadlord-empty':(59,29,258,156),
               'dreadlord-fireball':(59,29,258,156)}
    for name,box in regions.items():
        base = 'council' if name.startswith('council') else 'horned' if name.startswith('horned') else 'followers' if name.startswith('followers') else 'dreadlord'
        keep = art[base].copy(); x,y,x2,y2 = [v*3 for v in box]
        keep[y:y2,x:x2] = plates[name][y:y2,x:x2]
        plates[name] = keep
    yy,xx = np.mgrid[:504,:960].astype(np.float32)
    flow = cv2.DISOpticalFlow_create(cv2.DISOPTICAL_FLOW_PRESET_MEDIUM)
    flow.setFinestScale(0)

    def warp(target, anchor, painting):
        if np.array_equal(src[target],src[anchor]):
            return painting.copy()
        t = cv2.cvtColor(rgb[target],cv2.COLOR_RGB2GRAY)
        s = cv2.cvtColor(rgb[anchor],cv2.COLOR_RGB2GRAY)
        field = flow.calc(t,s,None)
        field = cv2.resize(field,(960,504),interpolation=cv2.INTER_LINEAR)*3
        return cv2.remap(painting,xx+field[:,:,0],yy+field[:,:,1],cv2.INTER_LINEAR,borderMode=cv2.BORDER_REPLICATE)

    def maskHD(mask):
        mask = cv2.morphologyEx(mask.astype(np.uint8),cv2.MORPH_CLOSE,np.ones((3,3),np.uint8))
        mask = cv2.dilate(mask,np.ones((3,3),np.uint8))
        return cv2.GaussianBlur(cv2.resize(mask.astype(np.float32),(960,504)),(5,5),0)[:,:,None]

    def over(bg, fg, mask):
        return np.clip(bg*(1-mask)+fg*mask,0,255).astype(np.uint8)

    def bounded(mask, box):
        r = np.zeros((168,320),bool); x,y,x2,y2=box; r[y:y2,x:x2]=mask[y:y2,x:x2]; return r

    skull = Image.open(raw/'phlan-skull-raw.png').convert('RGBA')
    alpha=np.array(skull)[:,:,3]; ys,xs=np.where(alpha>128)
    skull=skull.crop((int(xs.min()),int(ys.min()),int(xs.max()+1),int(ys.max()+1)))
    outputs, reports = {}, []
    out.mkdir(parents=True)
    for row in rows:
        b,f=row['block'],row['frame']; k=(b,f); key=row['key']
        if b == 0:
            continue  # SSI標誌保持原版，本批不以近似商標補洞。
        if b in (1,12,13):
            base=(1,0); anchor=(1,13); bg=plates['council-empty']; original=src[k]
            if b==1 and f==13:
                image=art['council'].copy()
            else:
                m=bounded(original != src[base],regions['council-empty'])
                active=np.unique(original[m]); solid=len(active)<=3 and m.sum()>0
                if solid:
                    color=np.median(rgb[k][m],axis=0)
                    texture=np.zeros_like(bg);texture[:]=color
                    image=over(bg,texture,maskHD(m))
                else:
                    image=over(bg,warp(k,anchor,art['council']),maskHD(m))
        elif b in (2,3):
            moving=bounded(src[k]!=src[(2,0)],(63,28,257,155))
            image=over(art['phlan'],warp(k,(2,0),art['phlan']),maskHD(moving))
            if k==(2,0):image=art['phlan'].copy()
            # 骷髏的獨立色段160–175在本段不含石框／城市，位置與大小取自當幀。
            m=bounded(((src[k]>=160)&(src[k]<=175))|(src[k]==17),(100,27,245,91))
            if m.sum()>6:
                ys,xs=np.where(m); box=(int(xs.min())*3,int(ys.min())*3,int(xs.max()+1)*3,int(ys.max()+1)*3)
                layer=Image.new('RGBA',(960,504)); layer.alpha_composite(skull.resize((box[2]-box[0],box[3]-box[1]),Image.Resampling.LANCZOS),(box[0],box[1]))
                image=np.array(Image.alpha_composite(Image.fromarray(image).convert('RGBA'),layer).convert('RGB'))
        elif b in (4,5):
            if k==(4,0):image=art['horned'].copy()
            else:
                bg=plates['horned-empty']; m=bounded(src[k]!=src[(5,7)],regions['horned-empty'])
                colors=rgb[k]; white=m&(colors.min(axis=2)>190)
                if m.sum()>0 and white.sum()>m.sum()*.7:
                    light=np.zeros_like(bg);light[:]=[220,238,255]; image=over(bg,light,maskHD(m))
                else:
                    body=warp(k,(4,0),art['horned'])
                    changed=bounded(src[k]!=src[(4,0)],regions['horned-empty'])
                    cage=warp(k,(4,13),plates['horned-cage'])
                    energy=over(body,cage,maskHD(changed))
                    image=over(bg,energy,maskHD(m))
        elif b==6:image=art['heroes'].copy()
        elif b in (7,8):
            anchor=(7,0) if b==7 or f<4 else (8,11)
            paint=art['followers'] if anchor==(7,0) else plates['followers-burnt']
            image=warp(k,anchor,paint)
            # 固定背景與石框以核准圖保護。
            mask=bounded(np.ones((168,320),bool),regions['followers-burnt'])
            image=over(art['followers'],image,cv2.resize(mask.astype(np.float32),(960,504),interpolation=cv2.INTER_NEAREST)[:,:,None])
        elif b in (9,10,11):
            if k==(9,0):image=art['dreadlord'].copy()
            else:
                bg=plates['dreadlord-empty']; m=bounded(src[k]!=src[(11,3)],regions['dreadlord-empty'])
                anchor=(10,6) if b==10 and f>=3 else (9,0)
                paint=plates['dreadlord-fireball'] if anchor==(10,6) else art['dreadlord']
                if b==11:
                    # 原版是布幕從兩側遮住固定人物，不對人物作材質拉伸。
                    # 本段32–63是黑底、紫布幕與布幕金邊；人物用64以上色段。
                    visible=bounded(src[k]>=64,(63,29,257,156)).astype(np.uint8)
                    contours,_=cv2.findContours(visible,cv2.RETR_EXTERNAL,cv2.CHAIN_APPROX_SIMPLE)
                    cv2.drawContours(visible,contours,-1,1,cv2.FILLED)
                    image=over(bg,paint,maskHD(visible))
                    path=out/(key+'.png');Image.fromarray(image).convert('RGBA').save(path)
                    outputs[k]=image
                    reports.append(dict(key=key,block=b,frame=f,png=path.name,sha256=sha(path),source_indexed_sha256=row['art_indexed_sha256'],method='approved painting revealed through original curtain phase',scope='original BMA curtain mask; artwork approximation',rights='Codex generated HD, private project patch scope'))
                    continue
                # 手繪手指與衣料在原輪廓的細微差異，以同一動作區的凸包包住，避免切斷手指。
                points=cv2.findNonZero(m.astype(np.uint8))
                silhouette=np.zeros((168,320),np.uint8)
                if points is not None:cv2.fillConvexPoly(silhouette,cv2.convexHull(points),1)
                image=over(bg,warp(k,anchor,paint),maskHD(silhouette))
                if b==10 and f<3:
                    colors=rgb[k]
                    sparks=bounded((colors[:,:,2]>180)&(colors[:,:,1]>160)&(src[k]!=src[(9,0)]),(62,96,144,130))
                    glow=np.zeros_like(image);glow[:]=[210,246,255]
                    image=over(image,glow,maskHD(sparks))
        else:raise ValueError(k)
        path=out/(key+'.png');Image.fromarray(image).convert('RGBA').save(path)
        outputs[k]=image
        reports.append(dict(key=key,block=b,frame=f,png=path.name,sha256=sha(path),
                            source_indexed_sha256=row['art_indexed_sha256'],
                            method='approved generated painting/layer; source-guided artistic texture deformation',
                            scope='original BMA phase selects image; artwork motion is an approximation',
                            rights='Codex generated HD, private project patch scope'))
    contact=Image.new('RGB',(5*320,((len(outputs)+4)//5)*188))
    d=ImageDraw.Draw(contact)
    for n,(k,im) in enumerate(outputs.items()):
        x=n%5*320;y=n//5*188;contact.paste(Image.fromarray(im).resize((320,168),Image.Resampling.LANCZOS),(x,y));d.text((x+2,y+169),str(k),fill='white')
    contact.save(out/'contact.jpg')
    for n,plate in plates.items():Image.fromarray(plate).convert('RGBA').save(out/(n+'.png'))
    (out/'manifest.json').write_text(json.dumps(dict(issue=74,style='B',items=reports,tools=dict(opencv=cv2.__version__,numpy=np.__version__)),ensure_ascii=False,indent=2)+'\n')
    print(len(reports),'HD相位')


if __name__=='__main__':main()
