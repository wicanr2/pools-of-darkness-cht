#!/usr/bin/env python3
"""將逐件數值與目視雜湊通過的 VGA 戰鬥圖匯出到既有 actors／combat 清單。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('actor-req','actor-out','terrain-req','terrain-out','theme'):
        ap.add_argument('--'+name,type=Path,required=True)
    ap.add_argument('--merge-existing',action='store_true',help='保留已驗收圖層，拒絕跨批key、來源與鏡像歧義')
    ap.add_argument('--actor-keys',nargs='+',help='明示只接入本請求的指定怪物；未列出者不覆寫既有主題')
    a=ap.parse_args();files=[];outputs={};proof=[]
    for kind,req,out,target,sourcefile in (
        ('combat-actors',a.actor_req,a.actor_out,'actors.json','CPIC1.DAX'),
        ('combat-terrain',a.terrain_req,a.terrain_out,'combat.json','DUNGCOM.DAX')):
        request=json.loads((req/'manifest.json').read_text())
        validation=json.loads((out/'validation.json').read_text())
        visual=json.loads((out/'visual-review.json').read_text())
        if not validation['all_pass'] or validation['scope']!='all' or validation['request_manifest_sha256']!=sha(req/'manifest.json'):
            raise ValueError('驗收未過、非全批或請求已變更')
        bykey={r['key']:r for r in validation['items']};items=[];seen=set()
        selected=request['items']
        if kind=='combat-actors' and a.actor_keys:
            keys=set(a.actor_keys)
            if len(keys)!=len(a.actor_keys) or not keys<={it['key']for it in selected}:
                raise ValueError('指定怪物key重複或不在請求內')
            selected=[it for it in selected if it['key']in keys]
        for it in selected:
            key=it['key'];png=out/(key+'.png');idx=req/it['indexed'];pal=req/it['palette']
            if not re.fullmatch(r'[a-z0-9@-]+',key)or it['source_file']!=sourcefile:raise ValueError('來源種類或key不符本批')
            if sha(png)!=bykey[key]['png_sha256']or sha(png)!=visual.get(key)or sha(idx)!=it['indexed_sha256']or sha(pal)!=it['palette_sha256']:
                raise ValueError('缺目前版本目視或來源／成品雜湊不符')
            w,h=it['source_size'];ref=idx.read_bytes()
            if len(ref)!=w*h or len(pal.read_bytes())!=768:raise ValueError('來源長度不合')
            if kind=='combat-actors':
                if [w,h]not in ([24,24],[24,48],[48,24],[48,48],[72,48]):raise ValueError('怪物尺寸不合')
                variants=[ref,b''.join(ref[y*w:(y+1)*w][::-1]for y in range(h))]
                signatures=[(w,h,hashlib.sha256(v).hexdigest())for v in variants]
            else:
                if [w,h]!=[24,24]:raise ValueError('地形尺寸不合')
                signatures=[(w,h,sha(idx),hashlib.sha256(pal.read_bytes()[32*3:64*3]).hexdigest())]
            if len(set(signatures))!=len(signatures)or any(s in seen for s in signatures):raise ValueError('來源或鏡像重複，須先核對別名')
            seen.update(signatures)
            items.append(dict(key=key,size=[w,h],png=key+'.png',png_sha256=sha(png),indexed=key+'.idx',indexed_sha256=sha(idx)))
            files.extend(((png,key+'.png'),(idx,key+'.idx')))
        manifest=dict(version=1,scale=3,kind=kind,grid=[8,8,7,7,24],items=items)
        if kind=='combat-actors':manifest['background_index']=8
        outputs[target]=manifest
        proof.append(dict(kind=kind,request_sha256=sha(req/'manifest.json'),validation_sha256=sha(out/'validation.json'),visual_sha256=sha(out/'visual-review.json'),keys=[r['key']for r in items]))
    pending={name:source for source,name in files}
    if a.merge_existing:
        for target,manifest in outputs.items():
            path=a.theme/target
            if not path.exists():continue
            existing=json.loads(path.read_text())
            if any(existing.get(k)!=v for k,v in manifest.items()if k!='items'):raise ValueError('既有圖層契約不同')
            combined={item['key']:item for item in existing['items']}
            if len(combined)!=len(existing['items']):raise ValueError('既有key重複')
            for item in manifest['items']:
                if item['key']in combined and combined[item['key']]!=item:raise ValueError('同key已有不同內容，不能覆寫')
                combined[item['key']]=item
            merged=dict(existing);merged['items']=list(combined.values());outputs[target]=merged
        seen_by_kind={}
        for target,manifest in outputs.items():
            seen=seen_by_kind.setdefault(target,set())
            for item in manifest['items']:
                refpath=pending.get(item['indexed'],a.theme/item['indexed'])
                pngpath=pending.get(item['png'],a.theme/item['png'])
                if sha(refpath)!=item['indexed_sha256']or sha(pngpath)!=item['png_sha256']:raise ValueError('合併來源或成品雜湊失配')
                ref=refpath.read_bytes();w,h=item['size']
                if len(ref)!=w*h:raise ValueError('合併來源尺寸不合')
                variants=[ref,b''.join(ref[y*w:(y+1)*w][::-1]for y in range(h))]if target=='actors.json'else[ref]
                # 研究017證實各地形區塊的色盤雜湊互異；低色號沿既有EGA契約。
                bank=item['key'].split('@')[0]if target=='combat.json'and any(c>15 for c in ref)else''
                for value in variants:
                    signature=(w,h,hashlib.sha256(value).hexdigest(),bank)
                    if signature in seen:raise ValueError('跨批來源或鏡像重複，須先核對別名')
                    seen.add(signature)
        previous=a.theme/'vga-layers-provenance.json'
        old=json.loads(previous.read_text())['items']if previous.exists()else[]
        merged={(row['kind'],row['request_sha256']):row for row in old}
        merged.update({(row['kind'],row['request_sha256']):row for row in proof})
        proof=list(merged.values())
    a.theme.mkdir(parents=True,exist_ok=True)
    # 先驗完全部輸入，再寫到獨立 staging 主題；正式啟用前仍須正常前端及失配驗證。
    for source,name in files:shutil.copyfile(source,a.theme/name)
    for name,manifest in outputs.items():
        temporary=a.theme/(name+'.tmp');temporary.write_text(json.dumps(manifest,indent=2)+'\n');temporary.replace(a.theme/name)
    (a.theme/'vga-layers-provenance.json').write_text(json.dumps(dict(scope='numeric and asset visual accepted; runtime gate separate',items=proof),indent=2)+'\n')
    print('既有格式戰鬥主題已建置：'+', '.join(f'{k} {len(v["items"])}項'for k,v in outputs.items()))


if __name__=='__main__':main()
