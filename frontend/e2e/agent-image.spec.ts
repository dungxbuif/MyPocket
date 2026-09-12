import { expect,test } from "@playwright/test";
test.use({serviceWorkers:"block"});
const receipt={name:"agent-receipt.png",mimeType:"image/png",buffer:Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aX1kAAAAASUVORK5CYII=","base64")};

test("owned receipt upload is attached to agent request",async({page})=>{
  let receiptID="";
  await page.route("**/api/v1/files/presign",route=>route.fulfill({status:200,contentType:"application/json",body:JSON.stringify({file:{id:"receipt-e2e",object_key:"users/test/receipts/e2e.png",content_type:"image/png",size_bytes:receipt.buffer.length,original_filename:receipt.name,created_at:new Date().toISOString()},upload_url:"http://127.0.0.1:18173/test-receipt-upload"})}));
  await page.route("http://127.0.0.1:18173/test-receipt-upload",route=>route.fulfill({status:200,body:""}));
  await page.route("**/api/v1/agent/intakes",async route=>{const body=route.request().postDataJSON();receiptID=body.receipt_id;return route.fulfill({status:202,contentType:"application/json",body:JSON.stringify({session:{id:"image-session",kind:"intake",title:"Nhập giao dịch",status:"active",created_at:new Date().toISOString(),updated_at:new Date().toISOString()},message:{id:"image-message",session_id:"image-session",run_id:"image-run",role:"user",text:"Đọc hóa đơn",created_at:new Date().toISOString()},run:{id:"image-run",session_id:"image-session",kind:"intake",status:"completed",request_text:"Đọc hóa đơn",response_text:"Tổng 120.000đ",attempts:1,draft_ids:[],tool_runs:[{id:"tool",status:"completed",attempts:1}]}})});});
  await page.goto("/");await page.getByRole("button",{name:"Đăng nhập bằng Google"}).click();await page.getByRole("button",{name:"Để sau",exact:true}).click();await page.getByRole("button",{name:"Trợ lý",exact:true}).click();await page.getByRole("group",{name:"Loại yêu cầu"}).getByRole("button",{name:"Thêm giao dịch"}).click();
  await page.getByLabel("Yêu cầu cho trợ lý").fill("Đọc hóa đơn");await page.getByLabel("Đính kèm ảnh hóa đơn").setInputFiles(receipt);await expect(page.getByText(receipt.name)).toBeVisible();await page.getByRole("button",{name:/Gửi yêu cầu/}).click();
  await expect(page.getByText("Tổng 120.000đ")).toBeVisible();expect(receiptID).not.toBe("");
});
