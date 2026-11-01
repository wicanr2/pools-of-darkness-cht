// Docker內固定Chromium CDP，原版面與原生字型，透明文字另與B底圖合成。
import {spawn} from 'node:child_process';
import {writeFile} from 'node:fs/promises';
const browser=spawn('/ms-playwright/chromium-1194/chrome-linux/chrome',['--headless','--no-sandbox','--disable-dev-shm-usage','--allow-file-access-from-files','--remote-debugging-port=9222','--user-data-dir=/tmp/pod-jako-cdp','about:blank'],{stdio:'ignore'});
const exited=new Promise(resolve=>browser.once('exit',resolve));
const sleep=ms=>new Promise(resolve=>setTimeout(resolve,ms));let ws;
try {
 let target;
 for(let i=0;i<100;i++){try{target=await(await fetch('http://127.0.0.1:9222/json/new?about:blank',{method:'PUT'})).json();break;}catch{await sleep(50);}}
 if(!target)throw Error('Chromium未就緒');
 ws=new WebSocket(target.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{ws.onopen=resolve;ws.onerror=reject;});
 let id=0;const pending=new Map();
 ws.onmessage=e=>{const m=JSON.parse(e.data);if(!m.id)return;const p=pending.get(m.id);if(!p)return;pending.delete(m.id);clearTimeout(p.timer);m.error?p.reject(Error(JSON.stringify(m.error))):p.resolve(m.result);};
 const send=(method,params={})=>new Promise((resolve,reject)=>{const n=++id;const timer=setTimeout(()=>{pending.delete(n);reject(Error('CDP逾時 '+method));},15000);pending.set(n,{resolve,reject,timer});ws.send(JSON.stringify({id:n,method,params}));});
 const evaluate=async expression=>(await send('Runtime.evaluate',{expression,returnByValue:true})).result.value;
 await send('Page.enable');const version=await send('Browser.getVersion');
 for(const lang of ['ja','ko']) {
  let metrics=[];
  for(const scene of ['title-001','title-003','title-002','poster']) {
   const poster=scene==='poster',cover=scene==='title-002';
   await send('Emulation.setDeviceMetricsOverride',{width:poster?1440:960,height:poster?684:600,deviceScaleFactor:1,mobile:false});
   await send('Emulation.setDefaultBackgroundColorOverride',{color:{r:0,g:0,b:0,a:poster?0:1}});
   await send('Page.navigate',{url:`file:///out/${lang}/${poster?'poster':cover?'cover':'title'}.html?layout=${scene}`});
   let ready=false;
   for(let i=0;i<200;i++){
    if(await evaluate('document.readyState==="complete" && [...document.images].every(i=>i.complete && i.naturalWidth>0) && document.fonts.status==="loaded" && [...document.fonts].some(f=>f.family==="Rounded" && f.status==="loaded")')){ready=true;break;}
    await sleep(50);
   }
   if(!ready)throw Error('圖像或原生字型未載入');
   if(poster)await evaluate(`(()=>{const t=document.getElementById('title');let size=parseFloat(getComputedStyle(t).fontSize);while(t.scrollWidth>t.clientWidth&&size>24)t.style.fontSize=(--size)+'px';return size})()`);
   await send('Page.bringToFront');
   let boxes;
   if(poster)boxes=await evaluate('[...document.querySelectorAll(".copy > div")].map(e=>{let b=e.getBoundingClientRect();return {x:b.x,y:b.y,width:b.width,height:b.height,scrollWidth:e.scrollWidth,font:parseFloat(getComputedStyle(e).fontSize)}})');
   else if(cover)boxes=await evaluate('[...document.querySelectorAll(".line")].map(e=>{let r=document.createRange();r.selectNodeContents(e);let b=r.getBoundingClientRect();return {x:b.x,y:b.y,width:b.width,height:b.height,scrollWidth:b.width}})');
   else boxes=await evaluate('titleMetrics()');
   if(boxes.some(b=>b.x<0||b.y<0||b.x+b.width>(poster?1440:960)||b.y+b.height>(poster?656:600)||b.scrollWidth>b.width+1))throw Error('文字超出版面 '+lang+' '+scene+' '+JSON.stringify(boxes));
   metrics.push({scene,boxes});
   const screenshot=async(name,clip)=>{const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:true,...(clip?{clip}:{})});await writeFile(`/out/${lang}/`+name,Buffer.from(shot.data,'base64'));};
   if(poster){for(const [side,x]of [['left',0],['right',1200]])for(const scale of [2,3])await screenshot(`poster-text-${side}-${scale}x.png`,{x,y:56,width:240,height:600,scale:scale/3});}
   else await screenshot(scene+'.png');
  }
  await writeFile(`/out/${lang}/layout.json`,JSON.stringify({version,metrics},null,2)+'\n');
 }
} finally {
 if(ws)ws.close();browser.kill('SIGTERM');await Promise.race([exited,sleep(3000)]);
 if(browser.exitCode===null){browser.kill('SIGKILL');await exited;}
}
