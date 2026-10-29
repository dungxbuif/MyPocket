import { expect,test } from "@playwright/test";
test.use({serviceWorkers:"block"});

test("agent text flow remains review-first",async({page})=>{
  let polls=0;
  await page.route("**/api/v1/agent/**",async route=>{
    const request=route.request();
    const url=request.url();
    if(request.method()==="POST")return route.fulfill({status:202,contentType:"application/json",body:JSON.stringify({status:"ok",session:{id:"session-e2e",kind:"intake",title:"Nhập giao dịch",status:"active",created_at:"2026-09-12T00:00:00Z",updated_at:"2026-09-12T00:00:00Z"},message:{id:"message-user",session_id:"session-e2e",run_id:"run-e2e",role:"user",text:"Lunch 120k",created_at:"2026-09-12T00:00:00Z"},run:{id:"run-e2e",session_id:"session-e2e",kind:"intake",status:"queued",request_text:"Lunch 120k",attempts:0,draft_ids:[]}})});
    if(url.includes("/agent/intakes/session-e2e"))return route.fulfill({status:200,contentType:"application/json",body:JSON.stringify({status:"ok",session:{id:"session-e2e",kind:"intake",title:"Nhập giao dịch",status:"active",created_at:"2026-09-12T00:00:00Z",updated_at:"2026-09-12T00:00:01Z"},messages:[{id:"message-user",session_id:"session-e2e",run_id:"run-e2e",role:"user",text:"Lunch 120k",created_at:"2026-09-12T00:00:00Z"},{id:"message-assistant",session_id:"session-e2e",run_id:"run-e2e",role:"assistant",text:"Đã tạo nháp",action:{type:"drafts_created",draft_ids:["draft-e2e"]},created_at:"2026-09-12T00:00:01Z"}]})});
    polls++;return route.fulfill({status:200,contentType:"application/json",body:JSON.stringify({status:"ok",run:{id:"run-e2e",session_id:"session-e2e",kind:"intake",status:"completed",request_text:"Lunch 120k",attempts:1,draft_ids:["draft-e2e"]}})});
  });
  await page.goto("/");await page.getByRole("button",{name:"Đăng nhập bằng Google"}).click();await page.getByRole("button",{name:"Để sau",exact:true}).click();await page.getByRole("button",{name:"Trợ lý",exact:true}).click();
  await page.getByLabel("Yêu cầu cho trợ lý").fill("Lunch 120k");await page.getByRole("button",{name:/Gửi yêu cầu/}).click();
  await expect(page.getByText("Đang chờ")).toBeVisible();await expect(page.getByRole("button",{name:"Mở bản nháp"})).toBeVisible({timeout:5000});expect(polls).toBeGreaterThan(0);
});
