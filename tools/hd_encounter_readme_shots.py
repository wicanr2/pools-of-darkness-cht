#!/usr/bin/env python3
"""將已通過正常遭遇圖收據的原始像素裁切保存README，不縮放或重畫。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil

import numpy as np
from PIL import Image


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('root','run-root','report'):
        ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args();root=a.root.resolve();run=a.run_root.resolve()
    assert root in run.parents
    report=json.loads(a.report.read_text())
    assert report['status']=='PASS' and report['negative_controls_all_pass'] and report['flash_count']==0
    name=next(r['capture'] for r in report['normal_shots'] if r['full_source'] and r['distance']==0)
    assert Path(name).name==name
    row=next(r for r in report['normal_shots'] if r['capture']==name)
    sources=[run/side/'window-shots'/(name+'.png') for side in ('old','hd')]
    assert [sha(p) for p in sources]==[row['before_png_sha256'],row['after_png_sha256']]
    idx=run/'hd/sources'/(name+'.idx');pal=idx.with_suffix('.pal')
    assert sha(idx)==row['indexed_sha256'] and sha(pal)==row['palette_sha256']
    ref=np.frombuffer(idx.read_bytes(),np.uint8).reshape(200,320)
    rgb=np.frombuffer(pal.read_bytes(),np.uint8).reshape(256,3)
    pixels=rgb[ref[24:112,24:112]].repeat(3,0).repeat(3,1)
    rect=(312,72,576,336)
    old=np.array(Image.open(sources[0]).convert('RGBA').crop(rect))
    assert np.array_equal(old[:,:,:3],pixels) and np.all(old[:,:,3]==255), '原版裁切含其他覆繪，不能標原版'
    target=root/'docs/screenshots';assert target.is_dir() and target.stat().st_uid==1000
    outputs=[]
    for source,name_out,role in zip(sources,('encounter-041-original-3x.png','encounter-041-hd-3x.png'),('original','HD')):
        path=target/name_out
        with Image.open(source) as image:
            assert image.size==(1440,600)
            image.convert('RGBA').crop(rect).save(path)
        outputs.append(dict(path=str(path.relative_to(root)),role=role,source=str(source.relative_to(root)),source_sha256=sha(source),
                            crop=list(rect),size=[264,264],sha256=sha(path)))
    window=target/'encounter-041-demo-window.png';shutil.copyfile(sources[1],window)
    outputs.append(dict(path=str(window.relative_to(root)),role='HD full window',source=str(sources[1].relative_to(root)),
                        source_sha256=sha(sources[1]),size=[1440,600],sha256=sha(window)))
    proof=dict(scope='正常冷啟動DEMO近景；原始像素裁切與完整視窗，無縮放、重畫或來源注入',
               capture=name,source_frame=int(name.rsplit('f',1)[1])-1,source_frame_offset=-1,
               key='sprit1-041@0',indexed_sha256=sha(idx),palette_sha256=sha(pal),
               play_receipt_sha256=sha(a.report),images=outputs,
               rights='private original-reference screenshots; no public distribution rights implied')
    (target/'encounter-041-runtime.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(proof,ensure_ascii=False,indent=2))


if __name__=='__main__':main()
