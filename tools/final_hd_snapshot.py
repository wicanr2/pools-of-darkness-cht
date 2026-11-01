#!/usr/bin/env python3
"""保存已目視的片尾素材及原有版本1重建配方，不帶入原版索引。"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil


def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def write(p, obj): p.write_text(json.dumps(obj, ensure_ascii=False, indent=2)+'\n')


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--engine',required=True)
    a=ap.parse_args();r=a.root;w=r/'workplace/probe/final-hd';art=w/'artwork-v3'
    mpath=r/'review/hd/runtime/manifest.json';m=json.loads(mpath.read_text())
    assert len(m['theme'])==1888 and len(m['fonts'])==18
    before={row['path']:row['sha256'] for row in m['theme']}
    metrics=json.loads((art/'manifest.json').read_text());assert metrics['all_pass'] and len(metrics['items'])==43
    out=r/'review/hd/b48-final';assert not out.exists();out.mkdir()
    for row in metrics['items']:
        src=art/row['png'];assert sha(src)==row['sha256']
        shutil.copy2(src,out/row['png'])
    shutil.copy2(w/'candidate/theme/final.json',out/'final.json')
    shutil.copy2(art/'manifest.json',out/'metrics.json')
    shutil.copy2(art/'contact.png',out/'contact.png')
    plates=out/'plates';plates.mkdir();generated=r/'workplace/hd/out/b48-final'
    names=('dragon','wizard','city','pool-empty','lake-empty','portal-closed','wizard-wave')
    for name in names:
        for suffix in ('.png','-raw.png'):
            shutil.copy2(generated/(name+suffix),plates/(name+suffix))
    shutil.copy2(generated/'generations.json',plates/'generations.json')
    refs=out/'original';refs.mkdir()
    shutil.copy2(generated/'caption-inventory.png',refs/'caption-inventory.png')
    for p in sorted((r/'workplace/hd/requests/b48-final').glob('*-1x.png')):shutil.copy2(p,refs/p.name)
    # 綁定實際目視的43張成品；原尺寸另抽查龍、城市顯露、揮手及關門。
    write(out/'visual-review.json',dict(date='2026-11-02',reviewer='Codex',issue=77,
        contact_sha256=sha(out/'contact.png'),status='accepted 43-phase contact and native-size samples',
        items=[dict(key=x['key'],sha256=x['sha256']) for x in metrics['items']],
        limits=['Source-guided movement, light and texture are artistic approximations',
                'Private original comparisons do not belong in distributable patches']))
    recipes=json.loads((w/'candidate/rebuild-recipes.json').read_text());assert len(recipes)==146
    assert not before.keys() & {x['path'] for x in recipes}
    for row in m['theme']:
        assert sha(w/'candidate/theme'/row['path'])==row['sha256']
    for row in recipes: assert sha(w/'candidate/theme'/row['path'])==row['sha256']
    m['theme']+=recipes;m['theme'].sort(key=lambda x:x['path'])
    proof=json.loads((w/'source-proof.json').read_text())
    for name in ('FINAL.BMA','FINAL.ACF'):m['original_archives'][name]=proof['input_sha256'][name]
    m['engine']['commit']=a.engine
    m['scope']='片尾FINAL增補；2034檔主題、18份字型；私人素材與本機重建，歷史DEMO預期值不變'
    write(mpath,m)
    (r/'tools/rebuild-engine.commit').write_text(a.engine+'\n')
    write(r/'docs/re/035-final-source.json',proof)
    print('saved',len(m['theme']),'theme files;',len(recipes),'new;',len(list(plates.glob('*.png'))),'plates')


if __name__=='__main__':main()
