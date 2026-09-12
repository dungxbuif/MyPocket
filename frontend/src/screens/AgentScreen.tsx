import * as React from "react";
import { Bot, ImagePlus, Send, X } from "lucide-react";
import { effectiveAgentStatus, loadAgentRun, loadIntakeSession, submitAgentMessage, type AgentMessageView, type AgentRunView, type AgentSessionView } from "../app/agent";
import { Button } from "../components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { FilePickerInput } from "../components/inputs/FilePickerInput";
import { OperationError, operationFailure, type OperationFailure } from "../components/feedback/OperationError";
import { AgentMessage } from "../components/agent/AgentMessage";
import { AgentRunStatus } from "../components/agent/AgentRunStatus";
import { TextAreaControl } from "../app/components";

export function AgentScreen({ online, onDraftReady }: { online: boolean; onDraftReady: () => void }) {
  const [kind,setKind]=React.useState<AgentRunView["kind"]>("intake");
  const [message,setMessage]=React.useState("");
  const [image,setImage]=React.useState<File|null>(null);
  const [run,setRun]=React.useState<AgentRunView|null>(null);
  const [session,setSession]=React.useState<AgentSessionView|null>(null);
  const [messages,setMessages]=React.useState<AgentMessageView[]>([]);
  const [busy,setBusy]=React.useState(false);
  const [failure,setFailure]=React.useState<OperationFailure|null>(null);

  React.useEffect(()=>{
    if(!run || ["completed","failed"].includes(run.status)) return;
    let cancelled=false;
    const timer=window.setInterval(()=>{void loadAgentRun(run.id).then(async next=>{if(cancelled)return;setRun(next);if(next.status==="completed"&&session?.kind==="intake"){const history=await loadIntakeSession(session.id);if(!cancelled){setMessages(history.messages);onDraftReady();}}}).catch(error=>{if(!cancelled)setFailure(operationFailure(error,"Chưa cập nhật được tiến độ. Hãy thử lại."));});},1500);
    return()=>{cancelled=true;window.clearInterval(timer)};
  },[run?.id,run?.status,session?.id,session?.kind,onDraftReady]);

  function chooseImage(file:File){
    if(!file.type.startsWith("image/")||file.size>15*1024*1024){setFailure({message:"Chỉ nhận ảnh tối đa 15 MiB."});return}
    setFailure(null);setImage(file);
  }
  async function submit(){
    if(!online||busy||message.trim()==="")return;setBusy(true);setFailure(null);
    try{const next=await submitAgentMessage({kind,message:message.trim(),image:image??undefined});setRun(next.run);setSession(next.session);setMessages((current)=>[...current,next.message]);setMessage("");setImage(null)}catch(error){setFailure(operationFailure(error,"Chưa gửi được yêu cầu. Nội dung vẫn được giữ để thử lại."))}finally{setBusy(false)}
  }
  const status=run?effectiveAgentStatus(run):null;
  const hasAssistantMessage=messages.some((item)=>item.role==="assistant");
  const hasDraftAction=messages.some((item)=>item.action?.type==="drafts_created");
  return <section className="content-stack agent-screen" aria-label="Trợ lý tài chính">
    <div className="sub-header"><div><small>Review-first</small><h1><Bot aria-hidden="true"/> Trợ lý</h1></div></div>
    <Card>
      <CardHeader><CardTitle>Bạn muốn làm gì?</CardTitle><CardDescription>Phân tích dữ liệu hoặc tạo bản nháp. Agent không tự ghi sổ.</CardDescription></CardHeader>
      <div className="segmented agent-kind" role="group" aria-label="Loại yêu cầu">
        <Button variant={kind==="intake"?"default":"outline"} size="sm" onClick={()=>setKind("intake")}>Thêm giao dịch</Button>
        <Button variant={kind==="advisor"?"default":"outline"} size="sm" onClick={()=>setKind("advisor")}>Hỏi tài chính</Button>
      </div>
      <label className="agent-composer"><span>Yêu cầu</span><TextAreaControl aria-label="Yêu cầu cho trợ lý" value={message} onChange={event=>setMessage(event.target.value)} maxLength={8000} placeholder="Ví dụ: Tạo khoản chi 120.000đ cho bữa trưa…" /></label>
      <div className="agent-image-row">
        <FilePickerInput aria-label="Đính kèm ảnh hóa đơn" accept="image/*" capture="environment" disabled={!online||busy} onFileSelected={chooseImage}/>
        <ImagePlus aria-hidden="true"/>
        {image?<span>{image.name}<Button variant="icon" size="icon" aria-label="Bỏ ảnh" onClick={()=>setImage(null)}><X/></Button></span>:<small>Ảnh hóa đơn tùy chọn</small>}
      </div>
      {!online?<p className="offline-warning">Kết nối mạng để dùng agent và OCR.</p>:null}
      <OperationError failure={failure} onRetry={()=>void submit()} retryLabel="Gửi lại" busy={busy}/>
      <Button className="w-full" disabled={!online||message.trim()===""} loading={busy} onClick={()=>void submit()}><Send aria-hidden="true"/> Gửi yêu cầu</Button>
    </Card>
    {run?<div className="agent-thread">
      {messages.length>0?messages.map((item)=><React.Fragment key={item.id}><AgentMessage role={item.role==="assistant"?"agent":"user"} text={item.text}/>{item.action?.type==="drafts_created"?<Card><p>Đã tạo bản nháp để bạn kiểm tra, sửa, xác nhận hoặc từ chối.</p><Button variant="outline" onClick={onDraftReady}>Mở bản nháp</Button></Card>:null}</React.Fragment>):<AgentMessage role="user" text={run.request_text}/>}
      {status?<AgentRunStatus status={status}/>:null}
      {run.response_text&&!hasAssistantMessage?<AgentMessage role="agent" text={run.response_text}/>:null}
      {run.draft_ids.length>0&&!hasDraftAction?<Card><p>Đã tạo bản nháp để bạn kiểm tra, sửa, xác nhận hoặc từ chối.</p><Button variant="outline" onClick={onDraftReady}>Mở bản nháp</Button></Card>:null}
      {run.error_code?<p className="offline-warning" role="alert">Yêu cầu không hoàn tất. Mã: {run.error_code}</p>:null}
    </div>:null}
  </section>;
}
