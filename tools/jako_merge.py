#!/usr/bin/env python3
"""合併經人工語境校讀的日韓批次；完整驗收前只輸出候選text。"""
import argparse
from collections import Counter
import csv
import hashlib
import json
from pathlib import Path
import re
import shutil

PH=re.compile(r'\{[^{}]+\}|%[ds]')


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def rows(p):return list(csv.DictReader(p.open(),delimiter='\t',quoting=csv.QUOTE_NONE))


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('root','work','out'):ap.add_argument('--'+key,type=Path,required=True)
    ap.add_argument('--preview',action='store_true',help='不讀未完劇情批次，僅供隔離版面預覽')
    a=ap.parse_args();r=a.root;w=a.work
    assert a.out.resolve()!= (r/'text').resolve()
    plans=json.loads((w/'plan.json').read_text());small=json.loads((w/'small-translations.json').read_text())
    dictionaries={}
    for stem,file in (('ui-text','ui-translations.tsv'),('name','name-translations.tsv')):
        source=json.loads((w/'batches'/(stem+'.json')).read_text());done=rows(w/file)
        assert {int(x['id']) for x in done}=={x['id'] for x in source}
        assert len(done)==len(source)
        byid={x['id']:x for x in source}
        dictionaries[stem]={byid[int(x['id'])]['english'].casefold():x for x in done}
    agents={};agent_files={}
    if not a.preview:
        for part in range(1,4):
            expected={(x['stem'],x['index']) for x in json.loads((w/'parallel'/f'part-{part}.json').read_text())}
            actual={}
            for p in sorted((w/'parallel'/f'part-{part}').glob('b*.json')):
                batch=json.loads(p.read_text());assert isinstance(batch,list)
                agent_files[str(p.relative_to(w))]=sha(p)
                for x in batch:
                    key=x['stem'],x['index'];assert key in expected and key not in actual
                    assert set(x)=={'stem','index','ja','ko'},(p,key,set(x))
                    actual[key]=x
            assert set(actual)==expected,('incomplete part',part,len(actual),len(expected))
            for key,x in actual.items():assert key not in agents;agents[key]=x
    a.out.mkdir(parents=True,exist_ok=True)
    # 候選完整保留既有三語、查驗答案與音譯資料。
    for p in (r/'text').iterdir():
        if p.is_dir():shutil.copytree(p,a.out/p.name,dirs_exist_ok=True)
        elif p.suffix=='.tsv' or p.name.startswith('LICENSE'):shutil.copyfile(p,a.out/p.name)
    warnings=[];proof={};canonical={x['english'].casefold():x for x in rows(w/'glossary.tsv')}
    term_path=r/'text/jako-term-fixes.tsv'
    term_fixes={(x['lang'],x['stem'],int(x['index']),x['field']):x for x in rows(term_path)}
    term_applied=[]
    for lang in ('ja','ko'):
        host_path=r/'text'/('host-text.'+lang+'.draft.tsv')
        if not host_path.exists():host_path=r/'text'/('host-text.'+lang+'.tsv')
        host={x['key']:x['translation'] for x in rows(host_path)}
        results=[]
        for table in plans[lang]:
            stem=table['stem'];out=[]
            for index,item in enumerate(table['rows']):
                row=dict(item['row']);missing=False
                for field,cell in item['changes'].items():
                    source=cell['english'];value=cell['reused']
                    if stem in ('ecl-text','journal','walkthrough') and value is None:
                        if a.preview:missing=True;break
                        value=agents[(stem,index)][lang][field]
                    elif stem in dictionaries and value is None:value=dictionaries[stem][source.casefold()][lang]
                    elif stem=='host-text':value=host[row['key']]
                    elif stem=='title':value=small['title'].get(row['key'],{}).get(lang,source)
                    elif stem in ('cover','poster'):value=small[stem][row['key']][lang]
                    elif stem=='map-titles':value=small[stem][row['map']][lang]
                    elif stem=='map-guide':value=small[stem][source][lang]
                    elif stem=='composed-templates' and value is None:value=small[stem][source][lang]
                    elif stem=='menu-masters' and value is None:value=small['menu-words'][source][lang]
                    assert isinstance(value,str),(lang,stem,index,field)
                    if stem=='ui-text' and source.casefold() in canonical:
                        value=canonical[source.casefold()][lang]
                    fix=term_fixes.get((lang,stem,index,field))
                    if fix is not None:
                        assert value==fix['before'] and hashlib.sha256(value.encode()).hexdigest()==fix['before_sha256'],('stale term fix',lang,stem,index)
                        value=fix['after'];term_applied.append((lang,stem,index,field))
                    assert not any(x in value for x in ('\t','\r','\n')),('raw TSV control',lang,stem,index)
                    # 組句的原文{n}是捕獲語法；譯文用{1}等位置參數，沿既有繁中格式。
                    contract=item['row'][field] if stem=='composed-templates' else source
                    assert Counter(PH.findall(contract))==Counter(PH.findall(value)),('placeholder',lang,stem,index,contract,value)
                    assert not re.search('[\uac00-\ud7a3]',value) if lang=='ja' else not re.search('[\u3040-\u30ff\u3400-\u9fff]',value),(lang,stem,index,value)
                    if source.strip() and not value.strip() and not (stem=='title' and row['key']=='title.article'):raise ValueError(('empty',lang,stem,index))
                    if stem in ('ecl-text','journal','walkthrough'):
                        old=Counter(re.findall(r'\d+',source));new=Counter(re.findall(r'\d+',value))
                        if old!=new:warnings.append(dict(lang=lang,stem=stem,index=index,field=field,kind='numbers',expected=dict(old),actual=dict(new)))
                        if source.count('\\n')!=value.count('\\n'):warnings.append(dict(lang=lang,stem=stem,index=index,field=field,kind='paragraphs',expected=source.count('\\n'),actual=value.count('\\n')))
                    row[field]=value
                if missing:continue
                if stem in ('ui-text','ecl-text','name'):row['source']='machine-translation-76'
                out.append(row)
            header=table['header'];dest=a.out/(stem+'.'+lang+'.tsv')
            if not out and stem in ('journal','walkthrough'):
                if dest.exists():dest.unlink()
                continue
            assert all(set(x)==set(header) for x in out)
            dest.write_text('\n'.join(['\t'.join(header)]+['\t'.join(x[k] for k in header) for x in out])+'\n')
            results.append(dict(path=dest.name,rows=len(out),sha256=sha(dest)))
        proof[lang]=results
    if not a.preview:assert set(term_applied)==set(term_fixes), 'unapplied term fix'
    receipt=dict(status='PREVIEW only' if a.preview else 'MERGED; validation pending',issue=76,counts=proof,
                 term_fixes_sha256=sha(term_path),term_fixes_applied=len(term_applied),
                 agent_files=agent_files,warnings=warnings,plan_sha256=sha(w/'plan.json'),glossary_sha256=sha(w/'glossary.tsv'),
                 root_sources={p:sha(w/p) for p in ('ui-translations.tsv','name-translations.tsv','small-translations.json')})
    (a.out.parent/'merge.json').write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(status=receipt['status'],warnings=len(warnings),rows={lang:sum(x['rows'] for x in proof[lang]) for lang in proof}),ensure_ascii=False))


if __name__=='__main__':main()
