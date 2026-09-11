import { expect,test } from "@playwright/test";
test.use({serviceWorkers:"block"});

test("agent text flow remains review-first",async({page})=>{
  let polls=0;
  await page.route("**/api/v1/agent/**",async route=>{
    const request=route.request();
    if(request.method()==="POST")return route.fulfill({status:202,contentType:"application/json",body:JSON.stringify({run:{id:"run-e2e",kind:"transaction_draft",status:"queued",request_text:"Lunch 120k",attempts:0,draft_ids:[]}})});
    polls++;return route.fulfill({status:200,contentType:"application/json",body:JSON.stringify({run:{id:"run-e2e",kind:"transaction_draft",status:"completed",request_text:"Lunch 120k",attempts:1,draft_ids:["draft-e2e"]}})});
  });
  await page.goto("/");await page.getByRole("button",{name:"Đăng nhập bằng Google"}).click();await page.getByRole("button",{name:"Để sau",exact:true}).click();await page.getByRole("button",{name:"Trợ lý",exact:true}).click();
  await page.getByLabel("Yêu cầu cho trợ lý").fill("Lunch 120k");await page.getByRole("button",{name:/Gửi yêu cầu/}).click();
  await expect(page.getByText("Đang chờ")).toBeVisible();await expect(page.getByRole("button",{name:"Mở bản nháp"})).toBeVisible({timeout:5000});expect(polls).toBeGreaterThan(0);
});
