// 容器內以 Chromium CDP 明示 viewport，避免 CLI window-size 混入視窗外框高度。
import {spawn} from 'node:child_process';
import {writeFile, mkdir} from 'node:fs/promises';
const out=process.argv[2]||'/out';
const assets=process.argv[3]==='assets';
await mkdir(out,{recursive:true});
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',[
  '--headless','--no-sandbox','--disable-dev-shm-usage','--allow-file-access-from-files',
  '--remote-debugging-port=9222','--user-data-dir=/tmp/pod-poster-cdp','about:blank'],{stdio:'ignore'});
const sleep=ms=>new Promise(r=>setTimeout(r,ms));
let ws;
try{
  let target;
  for(let i=0;i<100;i++){
    try{target=await(await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'})).json();break}catch{}
    await sleep(50);
  }
  if(!target)throw Error('Chromium 未就緒');
  ws=new WebSocket(target.webSocketDebuggerUrl);
  await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject});
  let sequence=0;const waiting=new Map();
  ws.onmessage=e=>{const m=JSON.parse(e.data);if(!m.id)return;const p=waiting.get(m.id);if(!p)return;waiting.delete(m.id);clearTimeout(p.timer);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result)};
  const send=(method,params={})=>new Promise((resolve,reject)=>{const id=++sequence;const timer=setTimeout(()=>{waiting.delete(id);reject(Error('CDP 逾時：'+method))},15000);waiting.set(id,{resolve,reject,timer});ws.send(JSON.stringify({id,method,params}))});
  await send('Page.enable');
  const shots=[];
  const scenes=assets?[['asset-zh','zh-TW','wide',1440],['asset-en','en','wide',1440]]:[['wide-zh','zh-TW','wide',1440],['wide-en','en','wide',1440],['narrow-zh','zh-TW','narrow',1066]];
  for(const [name,lang,layout,width]of scenes){
    await send('Emulation.setDefaultBackgroundColorOverride',{});
    await send('Emulation.setDeviceMetricsOverride',{width,height:684,deviceScaleFactor:1,mobile:false});
    await send('Page.navigate',{url:`file:///project/review/hd/poster/prototype.html?lang=${lang}&layout=${layout}`});
    let ready=false;
    for(let i=0;i<200;i++){
      const r=await send('Runtime.evaluate',{expression:'document.readyState==="complete" && [...document.images].every(i=>i.complete)',returnByValue:true});
      if(r.result.value){ready=true;break}await sleep(50);
    }
    if(!ready)throw Error('圖片未載入');
    await send('Page.bringToFront');
    await send('Runtime.evaluate',{expression:'document.fonts.ready',awaitPromise:true});
    const metrics=await send('Runtime.evaluate',{expression:'JSON.stringify({viewport:[innerWidth,innerHeight],game:{x:game.getBoundingClientRect().x,y:game.getBoundingClientRect().y,w:game.width,h:game.height}})',returnByValue:true});
    if(assets){
      // 保留不透明背景下的原定版字緣。透明背景會停用 Chromium 的 LCD 字緣。
      for(const [side,x]of[['left',0],['right',1200]])for(const scale of [3,2]){
        const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,clip:{x,y:56,width:240,height:600,scale:scale/3}});
        await writeFile(`${out}/poster-full-${side}.${lang}-${scale}x.png`,Buffer.from(shot.data,'base64'));
      }
      await send('Emulation.setDefaultBackgroundColorOverride',{color:{r:0,g:0,b:0,a:0}});
      for(const kind of ['base','text']){
        const css=kind==='base'?'.copy{visibility:hidden!important}':
          'body,.stage,.side{background:transparent!important}.side{border-color:transparent!important}.side svg,.side:after{visibility:hidden!important}.copy{visibility:visible!important}';
        await send('Runtime.evaluate',{expression:`(()=>{let e=document.getElementById('export-style');if(!e){e=document.createElement('style');e.id='export-style';document.head.append(e)}e.textContent=${JSON.stringify(css)}})()`});
        for(const [side,x]of[['left',0],['right',1200]])for(const scale of [2,3]){
          if(kind==='base'&&lang!=='zh-TW')continue;
          const file=kind==='base'?`poster-${side}-${scale}x.png`:`poster-text-${side}.${lang}-${scale}x.png`;
          const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,clip:{x,y:56,width:240,height:600,scale:scale/3}});
          await writeFile(`${out}/${file}`,Buffer.from(shot.data,'base64'));
        }
      }
    }else{
      const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,clip:{x:0,y:0,width,height:684,scale:1}});
      await writeFile(`${out}/${name}.png`,Buffer.from(shot.data,'base64'));
    }
    shots.push({file:name+'.png',lang,layout,...JSON.parse(metrics.result.value)});
  }
  await writeFile(out+(assets?'/assets-capture.json':'/capture.json'),JSON.stringify(shots,null,2)+'\n');
  console.log(assets?'海報底圖與各語系透明文字層已匯出':'三張完整 viewport 截圖已產生');
}finally{ws?.close();browser.kill('SIGTERM');await new Promise(r=>browser.once('exit',r));}
