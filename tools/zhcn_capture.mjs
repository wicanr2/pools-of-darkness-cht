// 固定Chromium CDP；只重排現有版面的簡中文字。/project唯讀，/out可寫。
import {spawn} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless','--no-sandbox','--disable-dev-shm-usage','--allow-file-access-from-files','--remote-debugging-port=9222','--user-data-dir=/tmp/pod-zhcn-cdp','about:blank'],{stdio:'ignore'});
const exited=new Promise(resolve=>browser.once('exit',resolve));
const sleep=ms=>new Promise(resolve=>setTimeout(resolve,ms));let ws;
try {
 let target;for(let i=0;i<100;i++){try{target=await(await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'})).json();break;}catch{await sleep(50);}}
 if(!target)throw Error('Chromium未就緒');ws=new WebSocket(target.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject;});
 let id=0;const pending=new Map();ws.onmessage=e=>{const m=JSON.parse(e.data);if(!m.id)return;const p=pending.get(m.id);if(!p)return;pending.delete(m.id);clearTimeout(p.timer);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result);};
 const send=(method,params={})=>new Promise((resolve,reject)=>{const n=++id;const timer=setTimeout(()=>{pending.delete(n);reject(Error('CDP逾時 '+method));},15000);pending.set(n,{resolve,reject,timer});ws.send(JSON.stringify({id:n,method,params}));});
 const evaluate=async expression=>(await send('Runtime.evaluate',{expression,returnByValue:true})).result.value;
 await send('Page.enable');let metrics=[];
 for(const scene of ['title-001','title-003','poster']) {
  const poster=scene==='poster';await send('Emulation.setDeviceMetricsOverride',{width:poster?1440:960,height:poster?684:600,deviceScaleFactor:1,mobile:false});
  await send('Page.navigate',{url:'file:///out/'+(poster?'poster':'title')+'.html?layout='+scene});
  let ready=false;for(let i=0;i<200;i++){if(await evaluate('document.readyState==="complete" && [...document.images].every(i=>i.complete && i.naturalWidth>0) && document.fonts.status==="loaded" && [...document.fonts].some(f=>f.family==="Rounded" && f.status==="loaded")')){ready=true;break;}await sleep(50);}
  if(!ready)throw Error('圖像或原生字型未載入');await send('Page.bringToFront');
  const boxes=await evaluate(poster?'[...document.querySelectorAll(".copy > div")].map(e=>{let b=e.getBoundingClientRect();return {x:b.x,y:b.y,width:b.width,height:b.height,scrollWidth:e.scrollWidth}})':'titleMetrics()');
  if(boxes.some(b=>b.x<0||b.y<0||b.x+b.width>(poster?1440:960)||b.y+b.height>(poster?656:600)||b.scrollWidth>b.width))throw Error('文字超出版面');metrics.push({scene,boxes});
  const screenshot=async(name,clip)=>{const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,...(clip?{clip}:{})});await writeFile('/out/'+name,Buffer.from(shot.data,'base64'));};
  if(poster){for(const [side,x]of [['left',0],['right',1200]])for(const scale of [2,3])await screenshot(`poster-full-${side}-${scale}x.png`,{x,y:56,width:240,height:600,scale:scale/3});}
  else await screenshot(scene+'.png');
 }
 await writeFile('/out/layout.json',JSON.stringify(metrics,null,2)+'\n');
} finally {if(ws)ws.close();browser.kill('SIGTERM');await Promise.race([exited,sleep(3000)]);if(browser.exitCode===null){browser.kill('SIGKILL');await exited;}}
