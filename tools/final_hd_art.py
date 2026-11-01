#!/usr/bin/env python3
"""以手繪主畫面及原版FINAL相位產生HD素材；材質動作為美術近似。"""
import argparse
import hashlib
import json
from pathlib import Path
import sys

import cv2
import numpy as np
from PIL import Image, ImageDraw

sys.path.insert(0, str(Path(__file__).resolve().parent))
from hd_metrics import read_png, structural_score, hue_score


def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    for key in ('requests', 'plates', 'out'): ap.add_argument('--'+key, type=Path, required=True)
    a = ap.parse_args(); cv2.setNumThreads(1)
    if a.out.exists(): raise ValueError('輸出須為新目錄')
    request = json.loads((a.requests/'manifest.json').read_text())
    rows = request['items']; mapping = {(x['block'], x['frame']): x['key'] for x in request['frames']}
    src = {}; rgb = {}
    for row in rows:
        key = row['key']; raw = a.requests/row['indexed']; p = a.requests/row['palette']
        assert sha(raw) == row['source_indexed_sha256'] and sha(p) == row['palette_sha256']
        src[key] = np.frombuffer(raw.read_bytes(), np.uint8).reshape(168, 320)
        rgb[key] = np.frombuffer(p.read_bytes(), np.uint8).reshape(256, 3)[src[key]]
    def key(b, f): return mapping[(b, f)]
    anchors = {'dragon': key(0, 0), 'pool-empty': key(1, 11), 'lake-empty': key(2, 0), 'city': key(3, 2),
               'wizard': key(4, 0), 'wizard-wave': key(4, 13), 'portal-closed': key(5, 1)}
    plates = {n: np.array(Image.open(a.plates/(n+'.png')).convert('RGB')) for n in anchors}
    assert all(p.shape == (504, 960, 3) for p in plates.values())
    # 同一原版石框，固定使用第一張手繪框，防止不同生成稿在不動邊界跳動。
    for name, plate in plates.items():
        if name != 'dragon':
            fixed = plates['dragon'].copy(); fixed[27*3:156*3, 60*3:260*3] = plate[27*3:156*3, 60*3:260*3]; plates[name] = fixed
    yy, xx = np.mgrid[:504, :960].astype(np.float32)
    flow = cv2.DISOpticalFlow_create(cv2.DISOPTICAL_FLOW_PRESET_MEDIUM); flow.setFinestScale(0)
    def warp(target, anchor, painting, box):
        if np.array_equal(src[target], src[anchor]): return painting.copy()
        field = flow.calc(cv2.cvtColor(rgb[target], cv2.COLOR_RGB2GRAY), cv2.cvtColor(rgb[anchor], cv2.COLOR_RGB2GRAY), None)
        field = cv2.resize(field, (960, 504), interpolation=cv2.INTER_LINEAR)*3
        warped = cv2.remap(painting, xx+field[:, :, 0], yy+field[:, :, 1], cv2.INTER_LINEAR, borderMode=cv2.BORDER_REPLICATE)
        result = painting.copy(); x, y, x2, y2 = [n*3 for n in box]; result[y:y2, x:x2] = warped[y:y2, x:x2]
        return result
    def bounded(mask, box):
        result = np.zeros((168, 320), np.float32); x, y, x2, y2 = box; result[y:y2, x:x2] = mask[y:y2, x:x2]; return result
    def hdmask(mask):
        mask = cv2.morphologyEx(mask.astype(np.float32), cv2.MORPH_CLOSE, np.ones((3, 3), np.uint8))
        mask = cv2.dilate(mask, np.ones((3, 3), np.uint8))
        return cv2.GaussianBlur(cv2.resize(mask, (960, 504)), (5, 5), 0)[:, :, None]
    def over(bg, fg, mask): return np.clip(bg*(1-mask)+fg*mask, 0, 255).astype(np.uint8)
    results = []; a.out.mkdir(parents=True)
    for row in rows:
        k = row['key']; b, f = row['block'], row['frame']
        if b == 0:
            if f == 0: image = plates['dragon'].copy()
            else:
                body = bounded(src[k] != src[key(1, 11)], (60, 29, 259, 116))
                image = over(plates['pool-empty'], warp(k, key(0, 0), plates['dragon'], (60, 29, 259, 116)), hdmask(body))
            glow = bounded((src[k] != src[key(0, 0)]) & (rgb[k][:, :, 2] > 180) & (rgb[k][:, :, 1] > 150) & (rgb[k][:, :, 0] > 110), (61, 28, 259, 116))
            light = np.zeros_like(image); light[:] = [218, 247, 255]; image = over(image, light, hdmask(glow))
        elif b == 1:
            image = plates['pool-empty'].copy()
            glow = bounded(src[k] != src[key(1, 11)], (61, 28, 259, 116))
            light = np.zeros_like(image); light[:] = [218, 247, 255]; image = over(image, light, hdmask(glow))
        elif b in (2, 3):
            if k == key(2, 0): image = plates['lake-empty'].copy()
            elif k == key(3, 2): image = plates['city'].copy()
            else:
                end = rgb[key(3, 2)].astype(float); target = rgb[k].astype(float)
                change = src[key(2, 0)] != src[key(3, 2)]
                weight = np.clip(np.sum(target*np.array([.299, .587, .114]), axis=2)/np.maximum(1, np.sum(end*np.array([.299, .587, .114]), axis=2)), 0, 1)
                weight[src[k] == src[key(2, 0)]] = 0
                # 城市先以稀疏線條顯露；不得用前景的擴張／閉合把稀疏線條填成整座城。
                alpha = cv2.resize(bounded(weight*change, (60, 53, 260, 156)), (960, 504), interpolation=cv2.INTER_LINEAR)
                alpha = cv2.GaussianBlur(alpha, (3, 3), 0)[:, :, None]
                image = over(plates['lake-empty'], plates['city'], alpha)
        elif b in (4, 5):
            if k == key(5, 1): image = plates['portal-closed'].copy()
            else:
                name = 'wizard-wave' if b == 4 and 12 <= f <= 28 else 'wizard'
                image = warp(k, anchors[name], plates[name], (96, 34, 241, 153))
                if b == 5:
                    # 此圖是原版關門前的較暗人物相位，不以時間猜亮度。
                    rect = (slice(34, 153), slice(96, 241))
                    ratio = min(1., float(rgb[k][rect].sum())/max(1., float(rgb[anchors[name]][rect].sum())))
                    image[34*3:153*3, 96*3:241*3] = (image[34*3:153*3, 96*3:241*3]*ratio).astype(np.uint8)
        else: raise ValueError((b, f))
        path = a.out/(k+'.png'); Image.fromarray(image).convert('RGBA').save(path)
        ct, check = read_png(path); score = structural_score(rgb[k], check); hue = hue_score(rgb[k], check)[0]
        results.append(dict(key=k, block=b, frame=f, png=path.name, sha256=sha(path),
            source_indexed_sha256=row['source_indexed_sha256'], source_palette_sha256=row['palette_sha256'],
            structure=score, hue=hue, passed=bool(ct == 6 and score >= .4 and hue < 30),
            method='generated painting and source-guided texture motion/reveal; artistic approximation',
            rights='Codex generated HD, private project patch scope'))
    contact = Image.new('RGB', (5*320, ((len(results)+4)//5)*188)); draw = ImageDraw.Draw(contact)
    for n, row in enumerate(results):
        x=n%5*320; y=n//5*188; contact.paste(Image.open(a.out/row['png']).convert('RGB').resize((320,168), Image.Resampling.LANCZOS), (x,y)); draw.text((x+2,y+169), row['key'], fill='white')
    contact.save(a.out/'contact.png')
    proof = dict(issue=77, style='B', all_pass=len(results) == 43 and all(x['passed'] for x in results),
        thresholds=dict(structure_min=.4, hue_median_max=30), items=results,
        inputs={str(p.name): sha(p) for p in sorted(a.plates.glob('*.png')) if not p.name.startswith('caption-')},
        tools=dict(opencv=cv2.__version__, numpy=np.__version__))
    (a.out/'manifest.json').write_text(json.dumps(proof, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(dict(count=len(results), all_pass=proof['all_pass'], failures=[x for x in results if not x['passed']]), ensure_ascii=False))
    if not proof['all_pass']: raise SystemExit(1)


if __name__ == '__main__': main()
