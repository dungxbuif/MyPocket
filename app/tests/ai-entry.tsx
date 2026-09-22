import { useState } from "react";
import { createRoot } from "react-dom/client";
import { BottomNavigation } from "../src/atomic/organisms/BottomNavigation";
import { QuickAddSheet } from "../src/atomic/organisms/QuickAddSheet";
import { AiEntrySheet } from "../src/atomic/organisms/AiEntrySheet";
import { BudgetsPanel } from "../src/atomic/organisms/BudgetsPanel";
import { TransactionsPanel } from "../src/atomic/organisms/TransactionsPanel";
import { BaseButton } from "../src/atomic/atoms/BaseButton";
import { BaseFab } from "../src/atomic/atoms/BaseNavigation";
import { Text } from "../src/atomic/atoms/Text";
import type { EntryProcess } from "../src/services/ai";
import "../src/styles.css";

const draft = {type:"expense",amount:42000,wallet_id:"w",category_id:null,occurred_at:"2026-09-21T05:00:00Z",note:"Lunch",included_in_reports:true};
const categoryFixture = [
  {id:"food",parent_id:null,kind:"expense",name:"Ăn uống",system_key:"expense_food",icon_key:"expense_food",is_system:true,wallet_ids:[]},
  {id:"coffee",parent_id:"food",kind:"expense",name:"Cà phê",system_key:null,icon_key:"expense_coffee",is_system:false,wallet_ids:[]},
  {id:"salary",parent_id:null,kind:"income",name:"Lương",system_key:"income_salary",icon_key:"income_salary",is_system:true,wallet_ids:[]},
];
const fixture = {
  calls: [] as {path:string;method:string;body:any}[], refreshes:0, aiOpens:0, manualOpens:0, plainClicks:0, failProcess:false, configured:true,
  nextProposals:[
    {id:"p",process_id:"",version:1,status:"pending",draft,questions:[]},
    {id:"missing",process_id:"",version:1,status:"pending",draft:{...draft,amount:0,wallet_id:"",occurred_at:""},questions:["Chọn ví và ngày"]},
    {id:"transfer",process_id:"",version:1,status:"pending",draft:{...draft,type:"transfer"},questions:[]},
  ] as any[],
  holdProposal:false, releaseProposal:null as null | (() => void),
  process: {id:"",processing:false,proposals:[]} as EntryProcess,
};
(window as any).fixture = fixture;
window.fetch = async (url, options = {}) => {
  const path = String(url), method = options.method ?? "GET", body = options.body instanceof FormData ? Object.fromEntries([...options.body.entries()].filter(([key])=>key!=="files")) : options.body ? JSON.parse(String(options.body)) : null;
  fixture.calls.push({path,method,body});
  const response = (data: unknown) => new Response(JSON.stringify({data}),{headers:{"content-type":"application/json"}});
  if(path.endsWith("/capabilities")) return response({ai_configured:fixture.configured,ocr_configured:fixture.configured,files_configured:fixture.configured});
  if(path.includes("/requests/")) return response(fixture.process);
  if(path.endsWith("/wallets")) return response([{id:"w",name:"Tiền mặt",type:"basic",currency:"VND",current_balance:100000,opening_balance:100000,is_in_total:true}]);
  if(path.endsWith("/categories")) return response(categoryFixture);
  if(path.endsWith("/transactions")) return method === "POST" ? response({ id:"manual-tx", ...body }) : response([
    { id:"tx1", type:"expense", amount:35000, wallet_id:"w", note:"Lunch", occurred_at:"2026-09-20T05:00:00Z", included_in_reports:true },
    { id:"tx2", type:"income", amount:100000, wallet_id:"w", note:"Refund", occurred_at:"2026-09-21T05:00:00Z", included_in_reports:true },
  ]);
  if(path.endsWith("/budgets")) return method === "POST" ? response({ id:"budget1", ...body }) : response({ items:[], spent:0, limit_amount:0 });
  if(path.endsWith("/process") && method === "POST") {
    if(fixture.failProcess) return new Response(JSON.stringify({detail:"Provider unavailable"}),{status:503,headers:{"content-type":"application/problem+json"}});
    fixture.process={id:body.request_id,processing:false,proposals:fixture.nextProposals.map(item=>({...item,process_id:body.request_id}))};
    return response(fixture.process);
  }
  if(path.includes("/proposals/")) {
    if(fixture.holdProposal) await new Promise<void>(resolve => { fixture.releaseProposal=resolve; });
    const id=path.split("/proposals/")[1].split("/")[0];
    const p=fixture.process.proposals.find(item=>item.id===id)!;
    if (body.version !== p.version) return new Response(JSON.stringify({detail:"Stale proposal"}),{status:409,headers:{"content-type":"application/problem+json"}});
    if(method==="PATCH") {p.draft=body.draft;p.version++;}
    else if(path.endsWith("/approve")) {p.status="approved";p.transaction_id="tx";}
    else if(path.endsWith("/reject")) p.status="rejected";
    return response(p);
  }
  throw new Error(`Unexpected API call ${path}`);
};
function Fixture() {
  const [mode,setMode]=useState<"manual"|"ai"|"budget"|"transactions"|null>(null);
  const [nav,setNav]=useState(true);
  const ai=()=>{fixture.aiOpens++;setMode("ai");};
  return <main className="mx-auto max-w-phone p-4">
    <Text>AI entry interaction fixture</Text>
    <BaseButton onClick={()=>setMode("budget")}>Budget fixture</BaseButton>
    <BaseButton onClick={()=>setMode("transactions")}>Transactions fixture</BaseButton>
    {mode === "budget" ? <BudgetsPanel masked={false} onChanged={()=>fixture.refreshes++} /> : null}
    {mode === "transactions" ? <TransactionsPanel onChanged={()=>fixture.refreshes++} /> : null}
    <BaseButton id="toggle-nav" onClick={()=>setNav(value=>!value)}>Toggle navigation</BaseButton>
    <BaseFab id="plain-fab" aria-label="Plain add" onClick={()=>fixture.plainClicks++}>+</BaseFab>
    {nav ? <BottomNavigation tab="overview" onTabChange={()=>{}} onAdd={()=>{fixture.manualOpens++;setMode("manual");}} onAiAdd={ai}/> : null}
    {mode==="manual" ? <QuickAddSheet onClose={()=>setMode(null)} onSaved={()=>fixture.refreshes++} onAiEntry={ai}/> : null}
    {mode==="ai" ? <AiEntrySheet onClose={()=>setMode(null)} onSaved={()=>fixture.refreshes++}/> : null}
  </main>;
}
createRoot(document.getElementById("root")!).render(<Fixture/>);
