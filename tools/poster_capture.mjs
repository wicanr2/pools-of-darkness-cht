// 容器內以 Chromium CDP 明示 viewport，避免 CLI window-size 混入視窗外框高度。
import {spawn} from 'node:child_process';
import {writeFile, mkdir, readFile} from 'node:fs/promises';
const out=process.argv[2]||'/out';
const assets=process.argv[3]==='assets';
const cover=process.argv[3]==='cover';
const titles=process.argv[3]==='titles';
await mkdir(out,{recursive:true});
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',[
  '--headless','--no-sandbox','--disable-dev-shm-usage','--allow-file-access-from-files',
  '--remote-debugging-port=9222','--user-data-dir=/tmp/pod-poster-cdp','about:blank'],{stdio:['ignore','ignore','pipe']});
let browserLog='';browser.stderr.on('data',b=>{browserLog=(browserLog+b.toString()).slice(-16000)});
const browserExit=new Promise(r=>browser.once('exit',r));
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
  const send=(method,params={})=>new Promise((resolve,reject)=>{const id=++sequence;const timer=setTimeout(()=>{waiting.delete(id);reject(Error('CDP 逾時：'+method+' '+(params.expression||'').slice(0,100)))},15000);waiting.set(id,{resolve,reject,timer});ws.send(JSON.stringify({id,method,params}))});
  await send('Page.enable');
  const shots=[];
  const scenes=titles?['title-001','title-003'].flatMap(key=>['zh-TW','en'].map(lang=>[`${key}.${lang}`,lang,key,960])):cover?[['cover-zh','zh-TW','wide',960],['cover-en','en','wide',960]]:assets?[['asset-zh','zh-TW','wide',1440],['asset-en','en','wide',1440]]:[['wide-zh','zh-TW','wide',1440],['wide-en','en','wide',1440],['narrow-zh','zh-TW','narrow',1066]];
  for(const [name,lang,layout,width]of scenes){
    await send('Emulation.setDefaultBackgroundColorOverride',{});
    const height=cover||titles?600:684;
    await send('Emulation.setDeviceMetricsOverride',{width,height,deviceScaleFactor:1,mobile:false});
    const page=titles?'b07-title':cover?'b06-cover':'poster';
    const url=`file:///project/review/hd/${page}/prototype.html?lang=${lang}&layout=${layout}`;
    await send('Page.navigate',{url});
    let ready=false;
    for(let i=0;i<200;i++){
      const r=await send('Runtime.evaluate',{expression:`location.href===${JSON.stringify(url)} && document.readyState==="complete" && [...document.images].every(i=>i.complete)`,returnByValue:true});
      if(r.result.value){ready=true;break}await sleep(50);
    }
    if(!ready)throw Error('圖片未載入');
    await send('Page.bringToFront');
    let fontReady=false;
    for(let i=0;i<200;i++){
      const r=await send('Runtime.evaluate',{expression:'document.fonts.status==="loaded" && [...document.fonts].some(f=>f.family==="Rounded" && f.status==="loaded")',returnByValue:true});
      if(r.result.value){fontReady=true;break}await sleep(50);
    }
    if(!fontReady)throw Error('指定源泉圓體未載入');
    const metrics=await send('Runtime.evaluate',{expression:'JSON.stringify({viewport:[innerWidth,innerHeight],game:{x:game.getBoundingClientRect().x,y:game.getBoundingClientRect().y,w:game.width,h:game.height}})',returnByValue:true});
    if(titles){
      const info=await send('Runtime.evaluate',{expression:'JSON.stringify(titleMetrics())',returnByValue:true});
      const boxes=JSON.parse(info.result.value);
      if(boxes.some(b=>b.x<0||b.y<0||b.x+b.width>960||b.y+b.height>600||b.scrollWidth>b.width))throw Error('開場文字溢位：'+JSON.stringify(boxes));
      await writeFile(`${out}/${name}.layout.json`,JSON.stringify(boxes,null,2)+'\n');
    }
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
      const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,clip:{x:0,y:0,width,height,scale:1}});
      await writeFile(`${out}/${name}.png`,Buffer.from(shot.data,'base64'));
    }
    shots.push({file:name+'.png',lang,layout,...JSON.parse(metrics.result.value)});
  }
  await writeFile(out+(assets?'/assets-capture.json':'/capture.json'),JSON.stringify(shots,null,2)+'\n');
  console.log(titles?'四張開場標誌／名單與文字邊界已匯出':cover?'兩語系封面預覽已匯出':assets?'海報底圖與各語系透明文字層已匯出':'三張完整 viewport 截圖已產生');
}catch(error){
  console.error(browserLog);
  for(const name of ['pids.current','pids.events','memory.events']){
    try{console.error(name+': '+await readFile('/sys/fs/cgroup/'+name,'utf8'))}catch{}
  }
  throw error;
}finally{ws?.close();browser.kill('SIGTERM');await browserExit;}
