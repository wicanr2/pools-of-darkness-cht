#!/usr/bin/env python3
"""在Docker內建立日韓翻譯草稿，只寫workplace，不接入正式語系。"""
import argparse
import csv
import hashlib
import json
from pathlib import Path
import re


def rows(path):
    return list(csv.DictReader(path.open(),delimiter='\t',quoting=csv.QUOTE_NONE))


def digest(s):return hashlib.sha256(s.encode('ascii')).hexdigest()
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for name in ('root','prior','common','out'):ap.add_argument('--'+name,type=Path,required=True)
    a=ap.parse_args();r=a.root;t=r/'text';a.out.mkdir(parents=True,exist_ok=True)
    originals={stem:{x['key']:x['original'] for x in rows(r/'workplace/text-src'/file)}
               for stem,file in [('ecl-text','ecl-source.tsv'),('ui-text','ui-source.tsv')]}
    events={stem:{x['event_key']:x for x in rows(t/(stem+'-events.tsv'))}
            for stem in ('ecl-text','ui-text','name')}
    name={};realm_specific=set()
    for row in rows(t/'glossary.tsv'):
        if row['category'] not in ('monster','person','place','item','spell'):continue
        for s in (row['en'].strip(),row['en'].strip().upper()):
            name['name.'+digest(s)[:12]]=s
            if row['category'] in ('person','place'):realm_specific.add(digest(s))
    originals['name']=name
    # 母串的每個詞以完整原文SHA回查，不依位址或中文字猜英文。
    words={digest(s):s for x in originals['ui-text'].values() for s in re.split('[ @]',x) if s}
    all_plans={};counts={}
    for lang in ('ja','ko'):
        prior={};templates={};masters={}
        for base in (a.common,a.prior):
            for stem in ('ecl-text','ui-text','name'):
                ev={x['event_key']:x['original_sha256'] for x in rows(base/'text'/(stem+'-events.tsv'))}
                prior.update({(stem,ev[x['key']]):x['translation'] for x in rows(base/'text'/(stem+'.'+lang+'.tsv'))
                              if x['key'] in ev and not (base==a.common and stem=='name' and ev[x['key']] in realm_specific)})
            templates.update({x['template']:x['translation'] for x in rows(base/'text'/('composed-templates.'+lang+'.tsv'))})
            masters.update({x['original_sha256']:x['translation'] for x in rows(base/'text'/('menu-masters.'+lang+'.tsv'))})
        plan=[];reuse=0
        def add(stem,base,fields,english,reused=None):
            nonlocal reuse
            header=list(base[0]);items=[]
            for index,row in enumerate(base):
                changes={}
                for column in fields:
                    source=english(index,row,column)
                    fixed=reused(index,row,column) if reused else None
                    if fixed is not None:reuse+=1
                    changes[column]=dict(english=source,reused=fixed)
                items.append(dict(row=row,changes=changes))
            plan.append(dict(stem=stem,header=header,rows=items))
        for stem in ('ecl-text','ui-text','name'):
            data=rows(t/(stem+'.zh-TW.tsv'))
            for x in data:
                s=originals[stem][x['key']];e=events[stem][x['key']]
                assert digest(s)==e['original_sha256'] and len(s.encode('ascii'))==int(e['original_length'])
            add(stem,data,['translation'],lambda i,x,c:originals[stem][x['key']],
                lambda i,x,c:prior.get((stem,events[stem][x['key']]['original_sha256'])))
        add('composed-templates',rows(t/'composed-templates.tsv'),['translation'],lambda i,x,c:x['template'],lambda i,x,c:templates.get(x['template']))
        add('menu-masters',rows(t/'menu-masters.tsv'),['translation'],lambda i,x,c:words[x['original_sha256']],lambda i,x,c:masters.get(x['original_sha256']))
        for stem in ('host-text','title','cover','poster','journal','walkthrough','map-titles','map-guide'):
            data=rows(t/(stem+'.en.tsv'))
            fields=['title','body'] if stem in ('journal','walkthrough') else ['title'] if stem=='map-titles' else ['label','summary'] if stem=='map-guide' else ['translation']
            add(stem,data,fields,lambda i,x,c:x[c])
        all_plans[lang]=plan;counts[lang]=dict(tables=len(plan),rows=sum(len(x['rows']) for x in plan),reused_fields=reuse)
    proof=dict(issue=76,scope='draft only; English sources local',counts=counts,
               source={str(p.relative_to(r)):sha(p) for p in sorted(t.glob('*.tsv')) if '.ja.' not in p.name and '.ko.' not in p.name},
               prior={str(base):{stem:sha(base/'text'/(stem+'.ja.tsv')) for stem in ('ecl-text','ui-text','name')} for base in (a.common,a.prior)})
    (a.out/'plan.json').write_text(json.dumps(all_plans,ensure_ascii=False,indent=2)+'\n')
    (a.out/'prepare.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(counts,indent=2))


if __name__=='__main__':main()
