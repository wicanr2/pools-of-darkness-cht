#!/usr/bin/env python3
"""將候選主題對應到Git中的已存PNG，更新原樣設定與重建清單。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil


def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()


def inside(root,name):
    p=root/name
    if not name or Path(name).is_absolute() or root.resolve() not in p.resolve().parents:
        raise ValueError('檔案路徑超出根目錄')
    return p


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--theme',type=Path,required=True)
    ap.add_argument('--tracked-list',type=Path,required=True,help='git ls-files -z的檔案清單，包含已暫存的新素材')
    ap.add_argument('--batch',required=True)
    a=ap.parse_args();r=a.root
    if not re.fullmatch(r'b\d+',a.batch):raise ValueError('批次名稱不合')
    target=r/'review/hd/runtime/manifest.json';lock=json.loads(target.read_text())
    for name,want in lock['source_files'].items():
        if sha(inside(r,name))!=want:raise ValueError('字型來源不同，須獨立重驗')
    pngs={}
    for entry in a.tracked_list.read_bytes().split(b'\0'):
        if not entry:continue
        name=entry.decode('utf-8');p=inside(r,name)
        if p.suffix=='.png':pngs[sha(p)]=name
    rows=[];copies=[]
    old={row['path']:row for row in lock['theme']}
    for p in sorted(a.theme.rglob('*')):
        if not p.is_file():continue
        name=str(p.relative_to(a.theme));digest=sha(p);row=dict(path=name,sha256=digest)
        if p.suffix=='.png':
            source=pngs.get(digest)
            if source is None:raise ValueError(f'正式PNG尚未保存Git：{name}')
            row.update(kind='repo-file',source=source,rights='private original-reference review' if name.startswith('posters/') else 'accepted HD redraw')
        elif p.suffix=='.json':
            source='review/hd/runtime/theme/'+name;destination=inside(r,source)
            row.update(kind='repo-file',source=source,rights='project metadata; private review')
            copies.append((p,destination))
        elif p.suffix=='.idx':
            if name in old:
                if old[name]['kind']!='original-index' or old[name]['sha256']!=digest:
                    raise ValueError('既有來源索引改變，須獨立重驗')
                row=old[name]
            else:
                m=re.fullmatch(r'([a-z0-9]+)-(\d+)(?:@(\d+))?\.idx',name)
                if not m:raise ValueError('未知來源索引檔名')
                archive=m[1].upper()+'.DAX'
                if archive not in lock['original_archives']:raise ValueError('新原版圖族須獨立固定來源SHA')
                row.update(kind='original-index',archive=archive,block=int(m[2]),frame=int(m[3] or 0),
                           decoder='wall-8x8' if archive=='8X8D1.DAX' else 'sprite-vga' if archive=='PIC1.DAX' else 'block-vga')
        else:raise ValueError('未知主題檔案種類')
        rows.append(row)
    new={row['path']:row for row in rows}
    if not set(old)<=set(new):raise ValueError('候選少了舊素材')
    if any(new[k]['sha256']!=old[k]['sha256'] for k in old if not k.endswith('.json')):
        raise ValueError('舊素材改變，須獨立驗收')
    # 全部輸入確認後才更新可由Git回復的設定檔，不寫正式主題。
    for source,destination in copies:
        destination.parent.mkdir(parents=True,exist_ok=True)
        if not destination.exists() or sha(destination)!=sha(source):shutil.copyfile(source,destination)
    lock['theme']=rows
    lock['scope']=a.batch+'主題；私人素材保存與本機重建，不擴張遊戲驗收或公開權利'
    target.write_text(json.dumps(lock,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(files=len(rows),png=sum(p['path'].endswith('.png') for p in rows),
                         original_indexes=sum(p['kind']=='original-index' for p in rows),manifest_sha256=sha(target))))


if __name__=='__main__':main()
