import { strict as assert } from 'node:assert';
import { createServer } from 'vite';
import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';

const server = await createServer({ server: { middlewareMode: true }, appType: 'custom' });
const originalFetch = globalThis.fetch;
const originalStorage = globalThis.localStorage;
globalThis.localStorage = { getItem: () => null };
try {
  const { createLongPress } = await server.ssrLoadModule('/src/atomic/atoms/longPress.ts');
  let now = 0, timer, manual = 0, ai = 0;
  const press = createLongPress(() => manual++, () => ai++, {
    schedule(fn, delay) { timer = { fn, at: now + delay }; return 1; },
    clear() { timer = undefined; },
  });
  const tick = ms => { now += ms; if (timer && now >= timer.at) { const fn = timer.fn; timer = undefined; fn(); } };
  press.start(); tick(499); press.release(); press.click();
  assert.deepEqual([manual, ai], [1, 0], 'short press opens only manual entry');
  press.start(); tick(500); tick(500); press.release(); press.click();
  assert.deepEqual([manual, ai], [1, 1], 'hold fires once and consumes release click');
  for (const reason of ['move', 'pointercancel', 'blur', 'unmount']) {
    press.start(); tick(200); press.cancel(); tick(500); press.click();
    assert.deepEqual([manual, ai], [1, 1], `${reason} cancels both activation paths`);
  }
  press.click(true);
  assert.equal(manual, 2, 'keyboard activation remains available after cancellation');

  const api = await server.ssrLoadModule('/src/services/ai.ts');
  const { proposalIssues, sameDraft, validateEntryFiles } = await server.ssrLoadModule('/src/services/aiEntryLogic.ts');
  const draft = { type:'expense', amount:42000, wallet_id:'w', category_id:null, occurred_at:'2026-09-21T05:00:00Z', note:'Lunch', included_in_reports:true };
  const proposal = { id:'p', process_id:'request-1', version:3, status:'pending', draft, questions:[] };
  const wallets = [{id:'w',name:'Cash',type:'basic'}], categories = [];
  assert.deepEqual(proposalIssues(draft, wallets, categories), []);
  for (const patch of [{amount:0},{wallet_id:''},{occurred_at:''},{type:'transfer'},{type:'unknown'},{wallet_id:'credit'},{category_id:'missing'}]) {
    assert.ok(proposalIssues({...draft,...patch}, wallets, categories).length, JSON.stringify(patch));
  }
  assert.ok(proposalIssues({...draft,category_id:'c'}, wallets, [{id:'c',kind:'income',wallet_ids:[]}]).length);
  assert.equal(sameDraft(draft, {...draft,note:'Edited'}), false);
  assert.equal(sameDraft(draft, {...draft}), true);
  assert.equal(validateEntryFiles([{type:'image/png',size:5*1024*1024}]), '');
  for (const files of [[{type:'image/gif',size:1}],[{type:'image/jpeg',size:5*1024*1024+1}],Array(21).fill({type:'image/png',size:1})]) assert.ok(validateEntryFiles(files));

  const calls = [];
  let current = {...proposal};
  globalThis.fetch = async (url, options) => {
    const body = options.body instanceof FormData ? options.body : options.body ? JSON.parse(options.body) : null;
    calls.push({url:String(url),method:options.method ?? 'GET',body,headers:new Headers(options.headers)});
    if (String(url).endsWith('/capabilities')) return response({ai_configured:true,ocr_configured:true,files_configured:false});
    if (String(url).endsWith('/requests/request-1')) return response({id:'request-1',proposals:[current],processing:false});
    if (options.method === 'PATCH') { assert.equal(body.version,3); current = {...current, version:4, draft:body.draft}; return response(current); }
    if (String(url).endsWith('/approve')) { assert.deepEqual(body,{version:4}); current = {...current,status:'approved',transaction_id:'tx'}; return response(current); }
    if (String(url).endsWith('/reject')) { assert.deepEqual(body,{version:3}); return response({...proposal,status:'rejected'}); }
    if (String(url).endsWith('/process')) return response({id:'request-1',proposals:[current],processing:false});
    throw new Error(`Unexpected request ${url}`);
  };
  const capabilities = await api.fetchEntryCapabilities();
  assert.deepEqual(capabilities,{ai_configured:true,ocr_configured:true,files_configured:false},'file upload availability includes private storage, not OCR credentials alone');
  const latest = await api.fetchEntryRequest('request-1');
  assert.equal(latest.id, 'request-1');
  const saved = await api.saveEntryProposal(proposal, {...draft,note:'Edited'});
  assert.equal(saved.draft.note,'Edited'); assert.equal(saved.version,4);
  const approved = await api.decideEntryProposal(saved, 'approve');
  assert.equal(approved.transaction_id,'tx'); assert.equal(approved.status,'approved');
  const rejected = await api.decideEntryProposal(proposal, 'reject'); assert.equal(rejected.status,'rejected');
  const file = new File([new Uint8Array([37,80,68,70])], 'receipt.pdf', {type:'application/pdf'});
  await api.processEntry({request_id:'request-1',text:'Receipt',timezone:'Asia/Ho_Chi_Minh',files:[file]});
  const submitted = calls.at(-1);
  assert.ok(submitted.url.endsWith('/process'));
  assert.equal(submitted.body.get('request_id'),'request-1');
  assert.equal(submitted.body.get('files').name,'receipt.pdf');
  assert.equal(submitted.body.get('files').type,'application/pdf');
  assert.deepEqual([...submitted.body.keys()],['request_id','text','timezone','files'],'browser sends raw multipart file bytes, not a base64 file field');
  assert.equal(submitted.headers.get('Content-Type'),null,'browser must set the multipart boundary itself');
  assert.ok(calls.every(call=>call.url.startsWith('/api/v1/ai/entry/')), 'draft APIs never call ledger endpoint');
  globalThis.fetch = async () => new Response(JSON.stringify({detail:'Provider unavailable'}),{status:503,headers:{'content-type':'application/problem+json'}});
  await assert.rejects(()=>api.processEntry({request_id:'failed',text:'hello',timezone:'UTC',files:[]}),/Provider unavailable/);

  const { EntryProposalRow } = await server.ssrLoadModule('/src/atomic/molecules/EntryProposalRow.tsx');
  const render = p => renderToStaticMarkup(React.createElement(EntryProposalRow,{proposal:p,wallets,categories,onUpdated(){},onSaved(){}}));
  assert.match(render(proposal),/42\.000/);
  for (const status of ['approved','rejected']) assert.doesNotMatch(render({...proposal,status}),/<input|<select|<button/,'terminal rows are immutable');
  const { AiEntrySheet } = await server.ssrLoadModule('/src/atomic/organisms/AiEntrySheet.tsx');
  const sheet = renderToStaticMarkup(React.createElement(AiEntrySheet,{onClose(){},onSaved(){}}));
  assert.doesNotMatch(sheet,/Cuộc trò chuyện mới|Hội thoại|session|messages/,'AI entry is one-shot and exposes no conversation state');
  assert.match(sheet,/<textarea[^>]+aria-label="Mô tả giao dịch"/,'AI entry uses the shared multiline composer');
  assert.match(sheet,/Đính kèm ảnh hoặc PDF/,'AI composer exposes attachment action');
  const { BaseFileUpload } = await server.ssrLoadModule('/src/atomic/atoms/BaseFileUpload.tsx');
  const upload = renderToStaticMarkup(React.createElement(BaseFileUpload,{label:'Images',onFiles(){}}));
  assert.match(upload,/type="file"/); assert.match(upload,/accept="image\/jpeg,image\/png,application\/pdf"/); assert.match(upload,/multiple=""/);
  console.log('AI entry contracts passed: hold/cancel, draft validation, edit/version/approve/reject, real service payloads/errors, terminal rows and upload.');
} finally { globalThis.fetch = originalFetch; globalThis.localStorage = originalStorage; await server.close(); }

function response(data) { return new Response(JSON.stringify({data}),{headers:{'content-type':'application/json'}}); }
