import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import { AgentScreen } from "./AgentScreen";

beforeEach(()=>{vi.restoreAllMocks()});
test("submits text and renders provider output as text",async()=>{
  const fetchMock=vi.spyOn(globalThis,"fetch").mockResolvedValueOnce(new Response(JSON.stringify({run:{id:"r1",kind:"analysis",status:"completed",request_text:"Phân tích",response_text:"<b>An toàn</b>",attempts:1,draft_ids:[]}}),{status:202,headers:{"Content-Type":"application/json"}}));
  render(<AgentScreen online onDraftReady={()=>{}}/>);fireEvent.click(screen.getByRole("button",{name:"Phân tích"}));fireEvent.change(screen.getByLabelText("Yêu cầu cho trợ lý"),{target:{value:"Phân tích"}});fireEvent.click(screen.getByRole("button",{name:/Gửi yêu cầu/}));
  expect(await screen.findByText("<b>An toàn</b>")).toBeInTheDocument();expect(document.querySelector("b")?.textContent).not.toBe("An toàn");expect(fetchMock).toHaveBeenCalled();
});
test("rejects oversized image before upload",()=>{render(<AgentScreen online onDraftReady={()=>{}}/>);const file=new File([new Uint8Array(15*1024*1024+1)],"huge.jpg",{type:"image/jpeg"});fireEvent.change(screen.getByLabelText("Đính kèm ảnh hóa đơn"),{target:{files:[file]}});expect(screen.getByRole("alert")).toHaveTextContent("15 MiB")});
test("offline mode preserves composer and blocks submit",()=>{render(<AgentScreen online={false} onDraftReady={()=>{}}/>);fireEvent.change(screen.getByLabelText("Yêu cầu cho trợ lý"),{target:{value:"Giữ lại"}});expect(screen.getByRole("button",{name:/Gửi yêu cầu/})).toBeDisabled();expect(screen.getByDisplayValue("Giữ lại")).toBeInTheDocument()});
