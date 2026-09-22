import { createServer } from 'vite';
import { spawn } from 'node:child_process';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { strict as assert } from 'node:assert';

const server = await createServer({server:{host:'127.0.0.1',port:0}});
const profile = await mkdtemp(path.join(tmpdir(),'mypocket-ai-browser-'));
let chrome, socket;
try {
  await server.listen();
  const port = server.httpServer.address().port;
  chrome = spawn(process.env.CHROME_BIN ?? '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',[
    '--headless=new','--no-first-run','--no-default-browser-check','--disable-background-timer-throttling','--disable-renderer-backgrounding','--remote-debugging-port=0',`--user-data-dir=${profile}`,'about:blank',
  ],{stdio:['ignore','ignore','pipe']});
  const endpoint = await new Promise((resolve,reject)=>{
    let output=''; const timeout=setTimeout(()=>reject(new Error('Chrome startup timeout')),15000);
    chrome.once('error',reject);
    chrome.stderr.on('data',chunk=>{output+=chunk;const match=output.match(/DevTools listening on (ws:\/\/[^\s]+)/);if(match){clearTimeout(timeout);resolve(match[1]);}});
  });
  socket = new WebSocket(endpoint);
  await new Promise(resolve=>socket.addEventListener('open',resolve,{once:true}));
  let nextID=0; const pending=new Map();
  socket.addEventListener('message',event=>{const msg=JSON.parse(event.data); if(pending.has(msg.id)){const {resolve,reject}=pending.get(msg.id);pending.delete(msg.id);if(msg.error)reject(new Error(msg.error.message));else resolve(msg.result);}});
  const send=(method,params={},sessionId)=>new Promise((resolve,reject)=>{const id=++nextID;pending.set(id,{resolve,reject});socket.send(JSON.stringify({id,method,params,...(sessionId?{sessionId}:{})}));});
  const {targetId}=await send('Target.createTarget',{url:'about:blank'});
  const {sessionId}=await send('Target.attachToTarget',{targetId,flatten:true});
  const call=(method,params={})=>send(method,params,sessionId);
  const evaluate=async expression=>{const result=await call('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(result.exceptionDetails)throw new Error(JSON.stringify(result.exceptionDetails));return result.result.value;};
  const wait=ms=>new Promise(resolve=>setTimeout(resolve,ms));
  const until=async expression=>{for(let i=0;i<100;i++){if(await evaluate(expression))return;await wait(30);}throw new Error(`Timeout: ${expression}\n${await evaluate('JSON.stringify({text:document.body.innerText,fixture:window.fixture})')}`);};
  const clickText=async text=>{await evaluate(`Array.from(document.querySelectorAll('button')).find(b=>b.textContent===${JSON.stringify(text)})?.click()`);await wait(30);};
  const center=()=>evaluate(`(()=>{const r=document.querySelector('[aria-label="Thêm giao dịch"]').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()`);
  const down=async()=>{const p=await center();await call('Input.dispatchMouseEvent',{type:'mouseMoved',...p});await call('Input.dispatchMouseEvent',{type:'mousePressed',...p,button:'left',clickCount:1});return p;};
  const up=p=>call('Input.dispatchMouseEvent',{type:'mouseReleased',...p,button:'left',clickCount:1});
  await call('Emulation.setDeviceMetricsOverride',{width:390,height:844,deviceScaleFactor:1,mobile:false});
  await call('Page.navigate',{url:`http://127.0.0.1:${port}/tests/ai-entry.html`});
  await call('Page.bringToFront');
  await until(`!!document.querySelector('[aria-label="Thêm giao dịch"]')`);
  await evaluate(`document.querySelector('#plain-fab').focus();document.querySelector('#toggle-nav').focus()`);
  const plain=await evaluate(`(()=>{const r=document.querySelector('#plain-fab').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()`);
  await call('Input.dispatchMouseEvent',{type:'mousePressed',...plain,button:'left',clickCount:1});
  await call('Input.dispatchMouseEvent',{type:'mouseReleased',...plain,button:'left',clickCount:1});
  assert.equal(await evaluate('fixture.plainClicks'),1,'FAB without hold must preserve ordinary clicks after blur');
  let p=await down(); await wait(80); await up(p);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Thêm giao dịch"]')`);
  assert.deepEqual(await evaluate('[fixture.manualOpens,fixture.aiOpens]'),[1,0]);
  await clickText('Hủy');
  p=await down(); await wait(550); await up(p);

  await until(`!!document.querySelector('textarea[aria-label="Mô tả giao dịch"]')`);
  await evaluate(`(()=>{const input=document.querySelector('textarea[aria-label="Mô tả giao dịch"]');Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value').set.call(input,'Lunch and taxi');input.dispatchEvent(new Event('input',{bubbles:true}));})()`);
  await clickText('Gửi');
  await until(`document.querySelectorAll('[aria-label="Giao dịch đề xuất"]').length===3`);
  assert.ok(await evaluate(`fixture.calls.some(c=>c.path.endsWith('/process') && c.method==='POST')`),'AI entry must submit through the one-shot endpoint');
  assert.equal(await evaluate(`fixture.calls.some(c=>c.path.includes('/sessions') || c.path.endsWith('/messages'))`),false,'AI entry must not call session or conversation endpoints');
  assert.deepEqual(await evaluate('[fixture.manualOpens,fixture.aiOpens]'),[1,1]);
  assert.equal(await evaluate(`document.querySelectorAll('textarea[aria-label="Mô tả giao dịch"]').length`),1,'AI input should be a multiline assistant composer');
  assert.equal(await evaluate(`!!document.querySelector('[aria-label="Đính kèm ảnh hoặc PDF"]')`),true,'composer should expose a clear attachment action');
  assert.equal(await evaluate(`!!document.querySelector('[aria-label="Kết quả AI"]')`),true,'recognized transactions should be grouped in an assistant result card');
  assert.equal(await evaluate('!!document.querySelector(\'[role="log"]\')'),false);
  assert.equal(await evaluate('document.body.textContent.includes("Cuộc trò chuyện mới")'),false);
  if(process.env.AI_ENTRY_SCREENSHOT){const {data}=await call('Page.captureScreenshot',{format:'png'});await writeFile(process.env.AI_ENTRY_SCREENSHOT,Buffer.from(data,'base64'));}
  const rows = `Array.from(document.querySelectorAll('[aria-label="Giao dịch đề xuất"]'))`;
  const input = async (selector,value) => {
    await evaluate(`(()=>{const input=${selector};Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set.call(input,${JSON.stringify(value)});input.dispatchEvent(new Event('input',{bubbles:true}));})()`);
    await wait(40);
  };
  await input(`${rows}[0].querySelector('[aria-label="Số tiền"]')`,'50000');
  await clickText('Lưu giao dịch');
  await until('fixture.refreshes===1');
  assert.equal(await evaluate(`fixture.calls.find(c=>c.method==='PATCH').body.draft.amount`),50000);
  assert.equal(await evaluate(`fixture.calls.find(c=>c.path.endsWith('/approve')).body.version`),2);
  assert.equal(await evaluate(`${rows}.length`),2);
  await clickText('Xóa item');
  await until(`${rows}.length===1`);
  assert.equal(await evaluate('fixture.refreshes'),1,'remove candidate must not refresh ledger');
  assert.equal(await evaluate(`${rows}[0].querySelector('select').disabled`),true,'transfer remains blocked');
  await clickText('Xóa item');
  await until(`${rows}.length===0`);
  await clickText('Đóng');

  // Two edited valid cards: save-all persists both, exactly once per proposal.
  await evaluate(`fixture.nextProposals=[...Array(2)].map((_,i)=>({id:'all-'+i,process_id:'',version:1,status:'pending',draft:{type:'expense',amount:12000,wallet_id:'w',category_id:null,occurred_at:'2026-09-21T05:00:00Z',note:'Batch',included_in_reports:true},questions:[]}))`);
  p=await down();await wait(550);await up(p);
  await until(`!!document.querySelector('textarea[aria-label="Mô tả giao dịch"]')`);
  await evaluate(`(()=>{const input=document.querySelector('textarea[aria-label="Mô tả giao dịch"]');Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value').set.call(input,'Two expenses');input.dispatchEvent(new Event('input',{bubbles:true}));})()`);
  await clickText('Gửi');
  await until(`${rows}.length===2`);
  await input(`${rows}[1].querySelector('[aria-label="Số tiền"]')`,'25000');
  await clickText('Lưu tất cả');
  await until('fixture.refreshes===3');
  assert.deepEqual(await evaluate('fixture.process.proposals.map(p=>[p.status,p.draft.amount])'),[['approved',12000],['approved',25000]]);
  await clickText('Đóng');

  // Manual form calendar cancel/selection keeps time, nested Escape preserves sheet.
  p=await down();await wait(50);await up(p);
  await until(`document.querySelector('[role="dialog"][aria-label="Thêm giao dịch"]') && !document.body.textContent.includes('Đang tải dữ liệu')`);
  await input(`document.querySelector('[aria-label="Số tiền"]')`,'35000');
  await evaluate(`document.querySelector('[aria-label="Nhóm"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Chọn nhóm"]')`);
  assert.equal(await evaluate(`document.querySelectorAll('[role="dialog"][aria-label="Chọn nhóm"] [aria-pressed="false"]').length`),2,'transaction picker uses selectable category-tree rows');
  await evaluate(`Array.from(document.querySelectorAll('[role="dialog"][aria-label="Chọn nhóm"] button')).find(b=>b.textContent.includes('Cà phê')).click()`);
  await until(`!document.querySelector('[role="dialog"][aria-label="Chọn nhóm"]')`);
  assert.ok(await evaluate(`document.querySelector('[aria-label="Nhóm"]').textContent.includes('Cà phê')`),'selected group returns to transaction form');
  await clickText('Thêm chi tiết');
  const beforeTime=await evaluate(`document.querySelector('[aria-label="Giờ giao dịch"]').value`);
  await evaluate(`document.querySelector('[aria-label="Ngày giao dịch"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  const beforeDay=await evaluate(`document.querySelector('[aria-selected="true"][data-day]').dataset.day`);
  await evaluate(`document.querySelector('[aria-label="Tháng trước"]').click()`);
  await call('Input.dispatchKeyEvent',{type:'keyDown',key:'Escape',code:'Escape',windowsVirtualKeyCode:27});
  await until(`!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  assert.ok(await evaluate(`!!document.querySelector('[aria-label="Thêm giao dịch"]')`),'Escape closes only calendar');
  await evaluate(`document.querySelector('[aria-label="Ngày giao dịch"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  assert.equal(await evaluate(`document.querySelector('[aria-selected="true"][data-day]').dataset.day`),beforeDay,'cancel does not change date');
  await evaluate(`document.querySelector('[aria-label="Tháng trước"]').click()`);
  const selectedDay=await evaluate(`(()=>{const button=document.querySelectorAll('[data-day]')[15];const day=button.dataset.day;button.click();return day;})()`);
  await until(`!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  assert.equal(await evaluate(`document.querySelector('[aria-label="Giờ giao dịch"]').value`),beforeTime);
  if(process.env.FORM_SCREENSHOT_DIR){const {data}=await call('Page.captureScreenshot',{format:'png'});await writeFile(path.join(process.env.FORM_SCREENSHOT_DIR,'transaction.png'),Buffer.from(data,'base64'));}
  await evaluate(`document.querySelector('[aria-label="Ngày giao dịch"]').click()`);
  if(process.env.FORM_SCREENSHOT_DIR){const {data}=await call('Page.captureScreenshot',{format:'png'});await writeFile(path.join(process.env.FORM_SCREENSHOT_DIR,'calendar.png'),Buffer.from(data,'base64'));}
  await call('Input.dispatchKeyEvent',{type:'keyDown',key:'Escape',code:'Escape',windowsVirtualKeyCode:27});
  await clickText('Lưu');
  await until(`fixture.calls.some(c=>c.path.endsWith('/transactions') && c.method==='POST')`);
  const tx=await evaluate(`fixture.calls.find(c=>c.path.endsWith('/transactions') && c.method==='POST').body`);
  assert.equal(tx.amount,35000);
  assert.equal(tx.wallet_id,'w');
  assert.equal(tx.category_id,'coffee');
  assert.equal(await evaluate(`(()=>{const d=new Date(${JSON.stringify(tx.occurred_at)});return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0');})()`),selectedDay);
  assert.equal(await evaluate(`!!document.querySelector('[role="dialog"][aria-label="Thêm giao dịch"]')`),false);

  await clickText('Budget fixture');
  await until(`document.body.textContent.includes('Chưa có ngân sách')`);
  await clickText('Thêm');
  await input(`document.querySelector('[aria-label="Tên ngân sách"]')`,'Ăn uống');
  await input(`document.querySelector('[aria-label="Hạn mức"]')`,'2000000');
  await evaluate(`document.querySelector('[aria-label="Nhóm chi"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Nhóm chi"]')`);
  assert.equal(await evaluate(`document.querySelectorAll('[role="dialog"][aria-label="Nhóm chi"] [aria-pressed="false"]').length`),2,'budget picker uses the same selectable category tree');
  await evaluate(`Array.from(document.querySelectorAll('[role="dialog"][aria-label="Nhóm chi"] button')).find(b=>b.textContent.includes('Ăn uống')).click()`);
  await until(`!document.querySelector('[role="dialog"][aria-label="Nhóm chi"]')`);
  await evaluate(`document.querySelector('[aria-label="Khoảng thời gian"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Khoảng thời gian"]')`);
  await evaluate(`document.querySelector('[aria-label="Từ ngày"]').click()`);
  await until(`!!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  await call('Input.dispatchKeyEvent',{type:'keyDown',key:'Escape',code:'Escape',windowsVirtualKeyCode:27});
  await until(`!document.querySelector('[role="dialog"][aria-label="Chọn ngày"]')`);
  assert.equal(await evaluate(`document.querySelectorAll('[role="dialog"]').length`),2);
  await clickText('Xong');
  if(process.env.FORM_SCREENSHOT_DIR){const {data}=await call('Page.captureScreenshot',{format:'png'});await writeFile(path.join(process.env.FORM_SCREENSHOT_DIR,'budget.png'),Buffer.from(data,'base64'));}
  await clickText('Lưu');
  await until(`fixture.calls.some(c=>c.path.endsWith('/budgets') && c.method==='POST')`);
  assert.equal(await evaluate(`fixture.calls.find(c=>c.path.endsWith('/budgets') && c.method==='POST').body.limit_amount`),2000000);

  await clickText('Transactions fixture');
  await until(`document.body.textContent.includes('Lunch') && document.body.textContent.includes('Refund')`);
  assert.ok((await evaluate('document.body.innerText')).indexOf('Refund') < (await evaluate('document.body.innerText')).indexOf('Lunch'),'days sorted newest first');
  if(process.env.FORM_SCREENSHOT_DIR){const {data}=await call('Page.captureScreenshot',{format:'png'});await writeFile(path.join(process.env.FORM_SCREENSHOT_DIR,'transactions.png'),Buffer.from(data,'base64'));}
  await evaluate(`document.querySelector('[aria-label="Tùy chọn giao dịch"]').click()`);
  await clickText('Xem theo nhóm');
  assert.ok(await evaluate(`document.body.textContent.includes('1 giao dịch')`));

  console.log('Browser passed: shared form/card editing, save-all, reject/transfer guard, date cancel/select/time preservation, nested Escape, real-shaped transaction/budget payloads, grouped ledger.');

} finally {
  socket?.close();
  if(chrome && chrome.exitCode===null){chrome.kill();await new Promise(resolve=>chrome.once('exit',resolve));}
  await server.close();
  await rm(profile,{recursive:true,force:true});
}
