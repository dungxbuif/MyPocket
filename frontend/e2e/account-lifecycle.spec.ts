import { expect, test, type Page } from '@playwright/test';

const api='http://127.0.0.1:18173/api/v1';
async function login(page:Page){const owner=`reset-${crypto.randomUUID()}`;await page.goto(`${api}/auth/google/callback?subject=${owner}&email=${owner}@example.com&email_verified=true&name=Reset`);return (await page.context().cookies()).find(cookie=>cookie.name==='mypocket_csrf')!.value}

test('destructive preview and cancel are mounted and non-mutating',async({page})=>{
  const csrf=await login(page);
  const created=await page.request.post(`${api}/wallets`,{headers:{'X-CSRF-Token':csrf,'Idempotency-Key':crypto.randomUUID()},data:{name:'Keep me',type:'cash'}});expect(created.status()).toBe(201);
  await page.goto('/');await page.getByRole('button',{name:'Tài khoản',exact:true}).click();
  await page.getByRole('button',{name:'Xem trước đặt lại dữ liệu'}).click();
  await expect(page.getByRole('dialog',{name:'Xác nhận đặt lại'})).toBeVisible();
  await page.getByRole('button',{name:'Hủy'}).click();
  await expect(page.getByRole('dialog',{name:'Xác nhận đặt lại'})).toHaveCount(0);
  const wallets=await (await page.request.get(`${api}/wallets`)).json();expect(wallets.wallets.some((wallet:{name:string})=>wallet.name==='Keep me')).toBe(true);
});
