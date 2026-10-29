import { expect, test, type Page } from '@playwright/test';

const api = 'http://127.0.0.1:18173/api/v1';
async function login(page:Page){const owner=`lifecycle-${crypto.randomUUID()}`;await page.goto(`${api}/auth/google/callback?subject=${owner}&email=${owner}@example.com&email_verified=true&name=Lifecycle`);await page.getByRole('button',{name:'Tài khoản',exact:true}).click()}

test('mounted account UI creates export and review-first import jobs without mutating balances',async({page})=>{
  await login(page);
  const before=await (await page.request.get(`${api}/wallets`)).json();
  await page.getByRole('button',{name:'Tạo bản xuất CSV'}).click();
  await expect(page.getByRole('status')).toContainText('export: queued');
  const csv='occurred_at,type,amount_vnd,source_wallet_id,destination_wallet_id,category_id,note,excluded_from_reports\n';
  await page.getByLabel('Chọn CSV để nhập').setInputFiles({name:'import.csv',mimeType:'text/csv',buffer:Buffer.from(csv)});
  await expect(page.getByRole('status')).toContainText('import: queued');
  const after=await (await page.request.get(`${api}/wallets`)).json();
  expect(after.wallets).toEqual(before.wallets);
});
