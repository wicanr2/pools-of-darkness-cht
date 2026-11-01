#!/usr/bin/env python3
"""建立可試玩候選的AppDir／Windows／macOS payload；原版只進本機完整版。"""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import stat
import struct
import zipfile


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def version_ok(value):
    if not re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}',value):raise ValueError('版號格式不合')
    return value


def files(root):return {str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*')) if p.is_file()}


def zip_tree(root,out):
    with zipfile.ZipFile(out,'w',zipfile.ZIP_DEFLATED,compresslevel=9) as z:
        for p in sorted(root.rglob('*')):
            if not p.is_file():continue
            name=root.name+'/'+str(p.relative_to(root));entry=zipfile.ZipInfo(name,(2026,11,1,0,0,0))
            entry.create_system=3;entry.external_attr=(stat.S_IFREG|stat.S_IMODE(p.stat().st_mode))<<16
            entry.compress_type=zipfile.ZIP_DEFLATED
            z.writestr(entry,p.read_bytes())
    with zipfile.ZipFile(out) as z:
        assert z.testzip() is None
        assert all(i.flag_bits&0x800 for i in z.infolist() if not i.filename.isascii())
        for i in z.infolist():
            if i.filename.endswith('/README.txt'):assert z.read(i).startswith(b'\xef\xbb\xbf')


def readme(version,platform,edition,full):
    intro='本機自用完整版，內含合法原版資料，請勿上傳或散布。' if full else ('HD中文化patch，不附原版，請自備合法DOS英文版1.10。' if edition=='hd' else '基本中文化patch，不附原版，請自備合法DOS英文版1.10。')
    return f'''黑暗之池 繁中候選 {version}
{intro}
平台：{platform}；呈現：{edition}。

這是可試玩候選，尚未宣稱原案全部完成。
繁中、已啟用HD、高解析英數、手札、攻略／地圖、工具列與demo中止沿目前驗收範圍。
繁中／英文／簡中、F1無敵及97種片頭敘事HD已接入。配樂試聽、日韓、片尾HD及推廣片仍待完成。

啟動：Windows執行PODCHT.exe；macOS開啟PODCHT.app；Linux執行AppImage。
patch請把原版47個遊戲檔放在程式旁original目錄，也可用-original指定。
HD patch首次啟動會從自備原版重建辨識資料，存於使用者資料目錄的hd-cache；之後核對並重用。
兩側採B版HD生圖海報，無須另備盒繪。2倍保留原版美術，3倍顯示已辨識的HD；SSI標誌維持原版。
Linux缺FUSE時可用APPIMAGE_EXTRACT_AND_RUN=1啟動；需要OpenGL、X11與ALSA，以及glibc2.36或相容環境。
首次啟動會核對並匯入原版，之後不改自備原版或發行目錄。
存檔位置：Windows %APPDATA%\\PODCHT\\game；macOS ~/Library/Application Support/PODCHT/game；Linux $XDG_DATA_HOME/pod-cht/game或~/.local/share/pod-cht/game。
不附1995年玩家SAVE資料，新角色與進度由你們建立。

F1無敵（預設關）、F2倍率、F8繁中／英文／簡中、F12原版、F5手札、F6攻略、F7地圖、F9查驗題答案、F3靜音。
demo中按一般按鍵可中止並回選單；功能鍵優先由前端處理。
macOS為未簽章候選，僅驗通用二進位結構，尚未在真機啟動驗收。
授權與第三方說明見LICENSE、THIRD-PARTY.md及licenses目錄。
來源與已知限制：本專案README、CONTEXT及PACKAGE-MANIFEST.json。
'''


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','engine','original','assets','build','stage','modcache'):
        ap.add_argument('--'+n,type=Path,required=True)
    ap.add_argument('--version',type=version_ok,required=True);ap.add_argument('--commit',required=True)
    ap.add_argument('--hd-patch-only',action='store_true')
    a=ap.parse_args();r=a.root
    assert re.fullmatch('[0-9a-f]{40}',a.commit)
    assert not a.stage.exists();a.stage.mkdir()
    lock=json.loads((r/'review/hd/runtime/manifest.json').read_text())
    assert files(a.assets/'theme')=={row['path']:row['sha256'] for row in lock['theme']}
    assert 1<=len(lock['theme'])<=4096
    assert all(sha(a.assets/'font'/row['path'])==row['sha256'] for row in lock['fonts'])
    binaries={'linux':a.build/'linux/pod-play','windows':a.build/'windows/PODCHT.exe','macos':a.build/'macos/pod-play'}
    assert all(p.is_file() for p in binaries.values())
    elf=binaries['linux'].read_bytes();assert elf[:4]==b'\x7fELF' and struct.unpack_from('<H',elf,18)[0]==62
    pe=binaries['windows'].read_bytes();off=struct.unpack_from('<I',pe,0x3c)[0];assert pe[:2]==b'MZ' and pe[off:off+4]==b'PE\0\0' and struct.unpack_from('<H',pe,off+4)[0]==0x8664
    mach=binaries['macos'].read_bytes();magic,n=struct.unpack_from('>II',mach);assert magic in (0xcafebabe,0xcafebabf) and n==2
    stride=20 if magic==0xcafebabe else 32;assert {struct.unpack_from('>I',mach,8+i*stride)[0] for i in range(2)}=={0x1000007,0x100000c}
    common=a.stage/'common';common.mkdir();shutil.copytree(a.assets/'font',common/'font')
    text=common/'text';text.mkdir()
    for p in (r/'text').iterdir():
        if p.is_file() and (p.suffix=='.tsv' or p.name.startswith('LICENSE')):shutil.copyfile(p,text/p.name)
        if p.is_dir() and (p.name=='cmudict' or p.name.startswith('translit-')):shutil.copytree(p,text/p.name)
    shutil.copyfile(r/'LICENSE',common/'LICENSE');lic=common/'licenses';lic.mkdir()
    for src,name in [(r/'font/source/rounded/SIL_Open_Font_License_1.1.txt','OFL-GenSenRounded.txt'),(r/'font/source/courier/OFL.txt','OFL-CourierPrime.txt'),(r/'font/source/unifont/OFL-1.1.txt','OFL-Unifont.txt'),(a.engine/'LICENSE','LICENSE-dosgolem'),(a.engine/'audio/nukedopl/COPYING.LGPL','COPYING.LGPL'),(a.build/'linux/Go-LICENSE','Go-LICENSE'),(r/'tools/vendor/appimage/appimage-LICENSE','AppImage-LICENSE')]:shutil.copyfile(src,lic/name)
    shutil.copyfile(r/'font/source/unifont/COPYING',lic/'COPYING-Unifont')
    shutil.copytree(a.engine/'audio/nukedopl',common/'source/nukedopl')
    source_tar=a.engine.parent/'engine-source.tar'
    with (common/'source/dosgolem.tar.gz').open('wb') as f:
        with gzip.GzipFile(fileobj=f,mode='wb',mtime=0) as z:z.write(source_tar.read_bytes())
    third=[]
    for line in (a.build/'linux/build-info.txt').read_text().splitlines():
        parts=line.strip().split('\t')
        if not parts or parts[0]!='dep':continue
        module,version=parts[1:3];escaped=lambda s:''.join('!'+c.lower() if c.isupper() else c for c in s)
        directory=a.modcache/(escaped(module)+'@'+escaped(version));assert directory.is_dir(),module
        notices=[p for p in directory.iterdir() if p.is_file() and p.name.lower().startswith(('license','copying'))];assert notices,module
        dest=lic/(module.replace('/','_')+'@'+version);dest.mkdir()
        for p in notices:shutil.copyfile(p,dest/p.name)
        third.append(dict(module=module,version=version,notices=[str((dest/p.name).relative_to(common)) for p in notices]))
    (common/'THIRD-PARTY.md').write_text('# 第三方來源與授權\n\nGo模組授權保存在licenses，版本由受測二進位建置資訊取得。Nuked OPL的Go來源、授權及完整固定引擎來源隨包提供於source。字型來源與授權保存於本專案font/source。\n\n'+''.join('- '+x['module']+' '+x['version']+'\n' for x in third))
    runtime=r/'tools/vendor/appimage/runtime-x86_64';assert sha(runtime)=='2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d'
    original={p.name:sha(p) for p in a.original.iterdir() if p.is_file()};assert len(original)==47
    recipes=[]
    for platform in ('linux','windows','macos'):
        for full,edition in (((False,'hd'),) if a.hd_patch_only else ((False,'base'),(True,'base'),(True,'hd'))):
            kind='full-local' if full else 'patch';name=f'pod-{a.version}-{platform}-{edition}-{kind}'
            top=a.stage/name;top.mkdir()
            if platform=='macos':
                app=top/'PODCHT.app/Contents';(app/'MacOS').mkdir(parents=True);data=app/'Resources'
                shutil.copytree(common,data);shutil.copyfile(binaries[platform],app/'MacOS/pod-play');(app/'MacOS/pod-play').chmod(0o755)
                plist=dict(CFBundleDisplayName='黑暗之池',CFBundleExecutable='pod-play',CFBundleIdentifier='io.github.wicanr2.podcht',CFBundleInfoDictionaryVersion='6.0',CFBundleName='PODCHT',CFBundlePackageType='APPL',CFBundleShortVersionString=a.version.split('-')[0][2:],CFBundleVersion=a.version,LSMinimumSystemVersion='11.0',NSHighResolutionCapable=True)
                (app/'Info.plist').write_bytes(plistlib.dumps(plist))
            elif platform=='linux':
                data=top/'usr/bin';shutil.copytree(common,data);shutil.copyfile(binaries[platform],data/'pod-play');(data/'pod-play').chmod(0o755)
                (top/'AppRun').write_text('#!/bin/sh\nset -eu\nexec "$APPDIR/usr/bin/pod-play" "$@"\n');(top/'AppRun').chmod(0o755)
                (top/'pod-cht.desktop').write_text('[Desktop Entry]\nType=Application\nName=黑暗之池\nExec=pod-play\nIcon=pod-cht\nCategories=Game;RolePlaying;\n')
                (top/'pod-cht.svg').write_text('<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256"><rect width="256" height="256" rx="24" fill="#18122b"/><ellipse cx="128" cy="145" rx="93" ry="55" fill="#57258a" stroke="#d0b46d" stroke-width="8"/><ellipse cx="128" cy="145" rx="67" ry="33" fill="#05010c"/><circle cx="128" cy="62" r="14" fill="#d0b46d"/></svg>')
                (top/'.DirIcon').symlink_to('pod-cht.svg')
            else:
                data=top;shutil.copytree(common,data,dirs_exist_ok=True);shutil.copyfile(binaries[platform],top/'PODCHT.exe')
            if full:
                (data/'original').mkdir()
                for n in original:shutil.copyfile(a.original/n,data/'original'/n)
                assert not(data/'original/SAVE').exists()
            source_count=0
            if edition=='hd' and not full:
                bundle=data/'hd-source';bundle.mkdir()
                (bundle/'manifest.json').write_text(json.dumps(lock,ensure_ascii=False,indent=2)+'\n')
                for row in lock['theme']:
                    if row['kind']!='repo-file':continue
                    assert 'original-reference' not in row.get('rights',''),row['path']
                    src=(r/row['source']).resolve()
                    assert r.resolve() in src.parents and sha(src)==row['sha256']
                    dst=bundle/row['source'];dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(src,dst)
                source_count=len(files(bundle))
            elif edition=='hd':shutil.copytree(a.assets/'theme',data/'hd')
            body=readme(a.version,platform,edition,full).replace('\n','\r\n');(data/'README.txt').write_bytes(b'\xef\xbb\xbf'+body.encode())
            if data!=top:shutil.copyfile(data/'README.txt',top/'README.txt')
            manifest=dict(version=a.version,engine_commit=a.commit,candidate=True,platform=platform,arch='amd64+arm64' if platform=='macos' else 'amd64',edition=edition,kind=kind,original_files=47 if full else 0,theme_files=len(lock['theme']) if edition=='hd' else 0,hd_source_files=source_count,fonts=len(lock['fonts']),files=files(top),missing_features=['Japanese/Korean','new music approval','FINAL movie HD','promo video','native macOS verification'],rights='local-only original data; never publish' if full else 'private patch; original game required')
            (top/'PACKAGE-MANIFEST.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2)+'\n')
            if not full:
                protected=set(original.values())
                assert not(set(files(top).values())&protected)
                assert not any(p.suffix.lower() in ('.dax','.bma','.acf','.idx','.ops') for p in top.rglob('*') if p.is_file())
            recipe=dict(name=name,platform=platform,edition=edition,kind=kind,tree=str(top),manifest_sha256=sha(top/'PACKAGE-MANIFEST.json'))
            if platform!='linux':
                archive=a.stage/(name+'.zip');zip_tree(top,archive);recipe.update(archive=str(archive),sha256=sha(archive))
            recipes.append(recipe)
    (a.stage/'recipes.json').write_text(json.dumps(dict(version=a.version,engine_commit=a.commit,items=recipes,modules=third),ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(dict(version=a.version,payloads=len(recipes),zips=sum(r['platform']!='linux' for r in recipes),appdirs=sum(r['platform']=='linux' for r in recipes)),indent=2))


if __name__=='__main__':main()
