#!/usr/bin/env python3
"""檢查英文面板正式資料、原校讀稿、數值及欄位；不改遊戲資料或補猜攻略。"""
import argparse
import collections
import csv
import hashlib
import json
from pathlib import Path
import re

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):
    with p.open(newline='') as f:
        reader=csv.DictReader(f,delimiter='\t',quoting=csv.QUOTE_NONE);rows=list(reader)
    if any(None in x or any(v is None for v in x.values()) for x in rows):raise ValueError('TSV欄數不符：'+str(p))
    return rows
def normalize(value):return value.replace('“','"').replace('”','"').replace('—','--')
def numbers(value):return collections.Counter(re.findall(r'[+-]?\d+',value))

def check(root):
    text=root/'text';source=read(root/'reference/journal-en.tsv');journal=read(text/'journal.en.tsv');guide=read(text/'walkthrough.en.tsv');base=read(text/'walkthrough.zh-TW.tsv')
    assert len(source)==len(journal)==99 and {int(x['entry']) for x in journal}==set(range(1,101))-{13}
    assert len(guide)==len(base)==78 and [int(x['chapter']) for x in guide]==list(range(1,79))
    originals={x['entry']:x for x in source}
    for row in journal:
        old=originals[row['entry']];title=normalize(old['title']);body=normalize(old['body'])
        if row['title']!=title:
            assert len(title)>38 and len(row['title'])<=38;body=title+'\\n'+body
        assert row['body']==body and row['source']=="SSI Adventurer's Journal p."+old['page']
        assert len(row['title'])<=38
    for old,new in zip(base,guide):
        assert old['chapter']==new['chapter'] and len(new['title'])<=33
        # 中文同時列出中英文寶物名時，其加值只計一次。
        dedup=re.sub(r'[（(][^()（）]*[A-Za-z][^()（）]*[）)]','',old['body'])
        body=new['body'];parts=body.split('\\n',1)
        if len(parts)==2 and re.match(r'^(Map|Chapter)\b',parts[0]):body=parts[1]
        assert numbers(dedup)==numbers(body),(old['chapter'],'數值或路線標號改變')
        source_numbers=re.findall(r'\d+',old['source']);assert source_numbers==re.findall(r'\d+',new['source'])
    geo=read(text/'map-geo.tsv');maps=read(text/'map-titles.en.tsv');points=read(text/'map-guide.en.tsv');old_points=read(text/'map-guide.zh-TW.tsv')
    assert len(maps)==len(geo)==32 and [x['map'] for x in maps]==[x['map'] for x in geo]
    for old,new in zip(geo,maps):assert bool(old['title'])==bool(new['title']) and old['title'].count('|')==new['title'].count('|')
    assert len(points)==len(old_points)==7
    for old,new in zip(old_points,points):assert all(old[k]==new[k] for k in ('map','n','x','y','source'))
    for rows,fields in ((journal,('title','body','source')),(guide,('title','body','source')),(maps,('title',)),(points,('label','summary'))):
        for row in rows:
            for field in fields:assert all(32<=ord(c)<127 for c in row[field]),(field,row)
    names=['journal.en.tsv','walkthrough.en.tsv','map-titles.en.tsv','map-guide.en.tsv']
    return dict(journals=99,guide_chapters=78,maps=32,unknown_maps=sum(not x['title'] for x in maps),guide_points=7,
        normalization='ASCII double quotes and double hyphens; long full titles preserved at start of body',
        source_journal_sha256=sha(root/'reference/journal-en.tsv'),
        original_source_sha256={n:sha(text/n) for n in ('journal.zh-TW.tsv','walkthrough.zh-TW.tsv','map-geo.tsv','map-guide.zh-TW.tsv')},
        english={n:sha(text/n) for n in names},numeric_and_location_checks='PASS',display_charset='printable ASCII; existing Courier16/24 unchanged')

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--receipt',type=Path);a=ap.parse_args();proof=check(a.root)
    if a.receipt:a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(proof,ensure_ascii=False,indent=2))

if __name__=='__main__':main()
