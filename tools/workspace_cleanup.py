#!/usr/bin/env python3
"""Docker內盤點／清理可重建工作檔，保護原始輸入、存檔及唯一研究證據。"""
import argparse,collections,hashlib,json,os,subprocess,tarfile
from pathlib import Path

KEEP_SUFFIX={'.go','.py','.sh','.c','.h','.s','.asm','.md','.tsv','.csv','.json','.jsonl','.log','.txt','.diff','.i64','.idb','.state','.mem','.bin','.pty','.sav','.thg','.efx','.pc','.cfg'}
PACKAGE_ROOTS=['package-promo-v1','release-v1','input-fix-v1/package','package-hd-v1/candidate-v1','package-candidate/v.0.1.0-20261101']
def sha(p):
    h=hashlib.sha256()
    with p.open('rb') as f:
        for b in iter(lambda:f.read(4*1024**2),b''):h.update(b)
    return h.hexdigest()
def write(p,d):p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
def disk(p):return int(subprocess.check_output(['du','-sx','-B1',str(p)],text=True).split()[0])
def size(p):
    s=p.stat();return dict(size=s.st_size,mtime_ns=s.st_mtime_ns,inode=s.st_ino,blocks=s.st_blocks*512)
def files(p):
    if p.is_symlink():return
    if p.is_file() and not p.is_symlink():yield p;return
    for base,dirs,names in os.walk(p,followlinks=False):
        dirs[:]=[d for d in dirs if d!='.git' and not(Path(base)/d).is_symlink()]
        for n in names:
            q=Path(base)/n
            if not q.is_symlink() and q.is_file():yield q

def plan(r,out):
    w=r/'workplace';assert out.parent==w and not out.exists();out.mkdir()
    tracked=subprocess.check_output(['git','-C',str(r),'ls-files','-z']).decode().split('\0')
    png=collections.defaultdict(list)
    for n in tracked:
        p=r/n
        if n and p.is_file() and p.suffix.lower() in ('.png','.otf','.hex','.dict'):
            png[p.stat().st_size].append(p)
    dup_cache={}
    protected={};targets={};gitroots=[]
    for base,dirs,names in os.walk(w,followlinks=False):
        p=Path(base)
        if p==out:dirs[:]=[];continue
        if '.git' in dirs or '.git' in names:
            gitroots.append(p)
            if p!=w/'dosgolem':dirs[:]=[];continue
            dirs[:]=[d for d in dirs if d!='.git']
        if p.name.casefold()=='save':
            for q in files(p):protected[str(q.relative_to(r))]=sha(q)
        for n in names:
            q=p/n
            if q.is_file() and not q.is_symlink() and (q.suffix.lower() in ('.pty','.sav','.thg','.efx','.pc') or n.lower().startswith(('pod-lang.','pod-map-','pod-journal-','pod-walkthrough-','pod-speed.','pod-barfont.'))):
                protected[str(q.relative_to(r))]=sha(q)
    for report in ['037-promo-runtime.json','039-v1-promo-runtime.json']:
        m=json.loads((r/'docs/re'/report).read_text())
        for name,info in m['storyboard']['sources'].items():
            p=r/name;assert p.is_file() and sha(p)==info['sha256'];protected[name]=info['sha256']
    protected_roots=[w/'input',w/'font',w/'fonts-src',w/'dosgolem',w/'hd/theme',out,*gitroots]
    def add(p,reason,override=False,duplicate=False):
        if not p.exists():return
        for q in files(p):
            assert q.resolve()==q and w in q.parents,q
            rel=str(q.relative_to(r));s=q.stat()
            if rel in protected or any(q==x or x in q.parents for x in protected_roots):continue
            if any(part.casefold()=='save' for part in q.relative_to(w).parts):continue
            if not override and q.suffix.lower() in KEEP_SUFFIX:continue
            origin=None
            if duplicate:
                choices=png.get(s.st_size,[])
                if not choices:continue
                h=sha(q)
                for x in choices:
                    if x not in dup_cache:dup_cache[x]=sha(x)
                    if dup_cache[x]==h:origin=str(x.relative_to(r));break
                if not origin:continue
            assert (s.st_uid,s.st_gid)==(os.getuid(),os.getgid()),q
            targets[rel]=dict(path=rel,reason=reason,**size(q),**({'same_git_file':origin} if origin else {}))
    for root in PACKAGE_ROOTS:
        p=w/root
        for n in ['stage','stage-first-layout','packages','main','engine','assets','bin']:
            add(p/n,'rebuildable package staging/source/assets')
        for n in ['main-source.tar','engine-source.tar']:add(p/n,'Git archive copy',override=True)
        for smoke in [p/'smoke',p/'smoke-baseline-old-now']:
            if smoke.exists():
                for base,dirs,names in os.walk(smoke,followlinks=False):
                    for n in list(dirs):
                        if n in ('data','home','config','wine','unpack','empty','hd-cache'):
                            add(Path(base)/n,'disposable extracted/install test environment');dirs.remove(n)
        for n in ['assets','segments','design-v1']:add(p/'promo'/n,'rebuildable video intermediates')
    for base,dirs,names in os.walk(w,followlinks=False):
        p=Path(base)
        if p==out or p==w/'input':dirs[:]=[];continue
        if '.git' in dirs and p!=w/'dosgolem':dirs[:]=[];continue
        for n in list(dirs):
            q=p/n
            if n in ('gocache','go-cache','__pycache__'):
                add(q,'compiler/bytecode cache',override=True);dirs.remove(n)
            elif n in ('clean-main','candidate','candidate-preview','candidate-v2','verified-assets','formal-before','rebuilt-theme','user-data','frontend') or (n.startswith('theme-before') and q!=w/'hd/theme'):
                add(q,'superseded rebuild/asset/test snapshot');dirs.remove(n)
            elif n=='theme' and q!=w/'hd/theme' and (q/'theme.json').is_file():
                add(q,'rebuildable duplicate theme');dirs.remove(n)
    # This model was rejected for production; final JA/KO tables live in Git.
    add(w/'multilingual-jako/models','unused downloadable translation models',override=True)
    add(w/'hd/out','byte-identical copy of tracked art/font',duplicate=True)
    add(w/'probe/g9-hd','byte-identical copy of tracked art/font',duplicate=True)
    for q in (w/'hd').iterdir():
        if q.is_dir() and q.name.startswith('theme-') and q.name.endswith(('-stage','-previous','-baseline')):
            add(q,'superseded HD theme backup')
    for name in ['rebuild-clean-v1/repo','hd-import-core-clean-v1/main','multilingual-zhcn/main','multilingual-zhcn/final-main']:
        add(w/name,'rebuildable Git source snapshot')
    # Some test runners put an entire DOS working copy in a directory named
    # "save". Remove only byte-identical, non-save original program/assets.
    original_game={q.name:q for q in (w/'input/dos/pod').iterdir() if q.is_file() and q.suffix.lower() in ('.dax','.bma','.acf','.drv','.com','.ovr','.exe','.sms','.bat')}
    original_hash={name:sha(q) for name,q in original_game.items()}
    for q in files(w):
        if any(q==x or x in q.parents for x in protected_roots):continue
        rel=str(q.relative_to(r));s=q.stat()
        if rel in targets:continue
        if q.name in original_game and s.st_size==original_game[q.name].stat().st_size and sha(q)==original_hash[q.name]:
            assert q.resolve()==q and (s.st_uid,s.st_gid)==(os.getuid(),os.getgid()),q
            targets[rel]=dict(path=rel,reason='byte-identical disposable DOS working copy',same_original=str(original_game[q.name].relative_to(r)),sha256=original_hash[q.name],**size(q))
            protected.pop(rel,None)
        elif q.suffix in ('','.test') and rel not in protected:
            with q.open('rb') as f:magic=f.read(4)
            if magic==b'\x7fELF':
                assert q.resolve()==q and (s.st_uid,s.st_gid)==(os.getuid(),os.getgid()),q
                targets[rel]=dict(path=rel,reason='superseded compiled test executable',**size(q))
        elif q.name in ('clean-main.tar','engine-clean.tar'):
            add(q,'superseded Git archive copy',override=True)
    # Keep the current formal executable and the module cache required offline.
    b=w/'dosgolem/workplace/bin'
    for q in b.iterdir():
        if q.name!='pod-play' and q.is_file() and not q.is_symlink():
            assert q.stat().st_uid==os.getuid();targets[str(q.relative_to(r))]=dict(path=str(q.relative_to(r)),reason='superseded compiled probe',**size(q))
    delivery={str(p.relative_to(r)):sha(p) for p in (r/'dist-all').rglob('*') if p.is_file() and p.name=='SHA256SUMS.json'}
    original={str(q.relative_to(r)):sha(q) for q in files(w/'input')}
    record=dict(date='2026-11-02',issue=81,before_bytes=disk(w),protected_files=protected,originals=original,delivery_manifests=delivery,
                formal_binary_sha256=sha(b/'pod-play'),targets=sorted(targets.values(),key=lambda x:x['path']))
    write(out/'plan.json',record)
    print('PLAN',len(targets),'files;',round(sum(x['blocks'] for x in targets.values())/1024**3,2),'GiB; protected saves/media',len(protected))

def apply(r,out):
    w=r/'workplace';p=json.loads((out/'plan.json').read_text());assert not(out/'result.json').exists()
    for row in p['targets']:
        q=r/row['path'];assert w in q.parents and q.resolve()==q and not q.is_symlink();assert size(q)=={k:row[k] for k in ('size','mtime_ns','inode','blocks')},q
        if 'same_original' in row:assert sha(q)==row['sha256']==sha(r/row['same_original'])
    for row in p['targets']:(r/row['path']).unlink()
    for name,h in {**p['protected_files'],**p['originals'],**p['delivery_manifests']}.items():assert sha(r/name)==h,name
    assert sha(w/'dosgolem/workplace/bin/pod-play')==p['formal_binary_sha256']
    result=dict(status='PASS',before_bytes=p['before_bytes'],after_bytes=disk(w),removed_files=len(p['targets']),protected_files=len(p['protected_files']))
    result['freed_bytes']=result['before_bytes']-result['after_bytes'];write(out/'result.json',result);print(json.dumps(result))

def archive_mem(r,out):
    base=r/'workplace/probe/draw-intercept-v1';paths=sorted(base.rglob('*.mem'));dest=out/'draw-memory.tar.gz';assert not dest.exists()
    manifest=[]
    with tarfile.open(dest,'w:gz',compresslevel=1) as t:
        for p in paths:
            assert p.is_file() and not p.is_symlink();manifest.append(dict(path=str(p.relative_to(r)),sha256=sha(p),size=p.stat().st_size));t.add(p,arcname=str(p.relative_to(r)),recursive=False)
    with tarfile.open(dest,'r:gz') as t:
        for row,member in zip(manifest,t.getmembers(),strict=True):
            assert member.name==row['path'] and hashlib.sha256(t.extractfile(member).read()).hexdigest()==row['sha256']
    record=dict(archive=dest.name,archive_sha256=sha(dest),files=manifest,original_bytes=sum(x['size'] for x in manifest),compressed_bytes=dest.stat().st_size,originals_removed=False)
    write(out/'memory-archive.json',record)
    for row in manifest:
        p=r/row['path'];assert sha(p)==row['sha256'];p.unlink()
    record['originals_removed']=True;write(out/'memory-archive.json',record)
    print('ARCHIVE',len(manifest),'memory dumps;',round(sum(x['size'] for x in manifest)/1024**3,2),'->',round(dest.stat().st_size/1024**3,2),'GiB')

def restore_mem(r,out):
    record=json.loads((out/'memory-archive.json').read_text());dest=out/record['archive']
    assert sha(dest)==record['archive_sha256']
    with tarfile.open(dest,'r:gz') as t:
        for row,member in zip(record['files'],t.getmembers(),strict=True):
            p=r/row['path'];base=r/'workplace/probe/draw-intercept-v1'
            assert member.name==row['path'] and member.isfile() and p.suffix=='.mem' and base in p.parents and p.resolve()==p
            data=t.extractfile(member).read();assert hashlib.sha256(data).hexdigest()==row['sha256']
            if p.exists():assert sha(p)==row['sha256'];continue
            assert p.parent.is_dir();tmp=p.with_suffix('.mem.restore')
            with tmp.open('xb') as f:f.write(data)
            tmp.replace(p)
    print('RESTORED',len(record['files']),'memory dumps')

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('mode',choices=['plan','apply','archive-mem','restore-mem']);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);a=ap.parse_args()
    assert Path('/.dockerenv').exists() and a.root.is_absolute() and a.out.parent==a.root/'workplace'
    {'plan':plan,'apply':apply,'archive-mem':archive_mem,'restore-mem':restore_mem}[a.mode](a.root,a.out)
if __name__=='__main__':main()
