import { useState } from "react";
import { createRoot } from "react-dom/client";
import { BottomNavigation } from "../src/atomic/organisms/BottomNavigation";
import { QuickAddSheet } from "../src/atomic/organisms/QuickAddSheet";
import { AiEntrySheet } from "../src/atomic/organisms/AiEntrySheet";
import { BaseButton } from "../src/atomic/atoms/BaseButton";
import { BaseFab } from "../src/atomic/atoms/BaseNavigation";
import { Text } from "../src/atomic/atoms/Text";
import type { EntrySession } from "../src/services/ai";
import "../src/styles.css";

const draft = {type:"expense",amount:42000,wallet_id:"w",category_id:null,occurred_at:"2026-09-21T05:00:00Z",note:"Lunch",included_in_reports:true};
const fixture = {
  calls: [] as {path:string;method:string;body:any}[], refreshes:0, aiOpens:0, manualOpens:0, plainClicks:0, failMessages:false, configured:true,
  createdSessions:0, failCreate:false, holdCreate:false, releaseCreate:null as null | (() => void),
  holdProposal:false, releaseProposal:null as null | (() => void),
  holdLatest:false, releaseLatest:null as null | (() => void),
  session: {id:"s",processing:false,messages:[{id:"m",role:"user",content:"Lunch and taxi",created_at:draft.occurred_at}],proposals:[
    {id:"p",session_id:"s",version:1,status:"pending",draft,questions:[]},
    {id:"missing",session_id:"s",version:1,status:"pending",draft:{...draft,amount:0,wallet_id:"",occurred_at:""},questions:["Chọn ví và ngày"]},
    {id:"transfer",session_id:"s",version:1,status:"pending",draft:{...draft,type:"transfer"},questions:[]},
  ]} as EntrySession,
};
(window as any).fixture = fixture;
window.fetch = async (url, options = {}) => {
  const path = String(url), method = options.method ?? "GET", body = options.body ? JSON.parse(String(options.body)) : null;
  fixture.calls.push({path,method,body});
  const response = (data: unknown) => new Response(JSON.stringify({data}),{headers:{"content-type":"application/json"}});
  if(path.endsWith("/capabilities")) return response({ai_configured:fixture.configured,ocr_configured:fixture.configured});
  if(path.endsWith("/latest")) {
    if(fixture.holdLatest) await new Promise<void>(resolve => { fixture.releaseLatest=resolve; });
    return response(fixture.session);
  }
  if(path.endsWith("/wallets")) return response([{id:"w",name:"Tiền mặt",type:"basic",currency:"VND",current_balance:100000,opening_balance:100000,is_in_total:true}]);
  if(path.endsWith("/categories")) return response([]);
  if(path.endsWith("/sessions") && method === "POST") {
    if(fixture.holdCreate) await new Promise<void>(resolve => { fixture.releaseCreate=resolve; });
    if(fixture.failCreate) return new Response(JSON.stringify({detail:"Could not create session"}),{status:503,headers:{"content-type":"application/problem+json"}});
    fixture.session={id:`new-${++fixture.createdSessions}`,processing:false,messages:[],proposals:[]};
    return response(fixture.session);
  }
  if(path.includes("/proposals/")) {
    if(fixture.holdProposal) await new Promise<void>(resolve => { fixture.releaseProposal=resolve; });
    const id=path.split("/proposals/")[1].split("/")[0];
    const p=fixture.session.proposals.find(item=>item.id===id)!;
    if (body.version !== p.version) return new Response(JSON.stringify({detail:"Stale proposal"}),{status:409,headers:{"content-type":"application/problem+json"}});
    if(method==="PATCH") {p.draft=body.draft;p.version++;}
    else if(path.endsWith("/approve")) {p.status="approved";p.transaction_id="tx";}
    else if(path.endsWith("/reject")) p.status="rejected";
    return response(p);
  }
  if(path.endsWith("/messages")) {
    if(fixture.failMessages) return new Response(JSON.stringify({detail:"Provider unavailable"}),{status:503,headers:{"content-type":"application/problem+json"}});
    fixture.session.messages.push({id:body.request_id,role:"user",content:body.text,created_at:draft.occurred_at});
    return response(fixture.session);
  }
  throw new Error(`Unexpected API call ${path}`);
};
function Fixture() {
  const [mode,setMode]=useState<"manual"|"ai"|null>(null);
  const [nav,setNav]=useState(true);
  const ai=()=>{fixture.aiOpens++;setMode("ai");};
  return <main className="mx-auto max-w-phone p-4">
    <Text>AI entry interaction fixture</Text>
    <BaseButton id="toggle-nav" onClick={()=>setNav(value=>!value)}>Toggle navigation</BaseButton>
    <BaseFab id="plain-fab" aria-label="Plain add" onClick={()=>fixture.plainClicks++}>+</BaseFab>
    {nav ? <BottomNavigation tab="overview" onTabChange={()=>{}} onAdd={()=>{fixture.manualOpens++;setMode("manual");}} onAiAdd={ai}/> : null}
    {mode==="manual" ? <QuickAddSheet onClose={()=>setMode(null)} onSaved={()=>fixture.refreshes++} onAiEntry={ai}/> : null}
    {mode==="ai" ? <AiEntrySheet onClose={()=>setMode(null)} onSaved={()=>fixture.refreshes++}/> : null}
  </main>;
}
createRoot(document.getElementById("root")!).render(<Fixture/>);
