// Local dev persistence proof. Only deletes IDs created by this run.
import assert from 'node:assert/strict';
const origin=process.env.TEST_API_ORIGIN??'http://localhost:4173';
let token='';
async function request(path,method='GET',body,expected=200) {
 const response=await fetch(`${origin}/api/v1${path}`,{method,headers:{'Content-Type':'application/json',...(token?{Authorization:`Bearer ${token}`}:{})},...(body?{body:JSON.stringify(body)}:{})});
 const payload=response.status===204?null:await response.json();
 assert.equal(response.status,expected,`${method} ${path}: ${JSON.stringify(payload)}`);
 return payload?.data;
}
const login=await request('/login','POST',{email:process.env.TEST_EMAIL??'admin@mypocket.local',password:process.env.TEST_PASSWORD??'12345678'});
token=login.token;
const wallets=[],budgets=[];
try {
 const categories=await request('/categories');
 const id=key=>{const category=categories.find(item=>item.system_key===key || item.icon_key===key);assert.ok(category,`missing ${key}`);return category.id;};
 const goal=await request('/wallets','POST',{name:'API proof goal',type:'goal',currency:'VND',opening_balance:100000,is_in_total:false,target_amount:1000000,target_date:'2026-12-31'},201);wallets.push(goal.id);
 const basic=await request('/wallets','POST',{name:'API proof basic',type:'basic',currency:'VND',opening_balance:0,is_in_total:false},201);wallets.push(basic.id);
 const balance=async(expected)=>{const rows=await request('/wallets');const row=rows.find(item=>item.id===goal.id);assert.equal(row.current_balance,expected);assert.equal(row.target_date.slice(0,10),'2026-12-31');};
 const input={wallet_id:goal.id,category_id:id('income_transfer_in'),type:'income',amount:900000,occurred_at:new Date().toISOString(),included_in_reports:true};
 const deposit=await request('/transactions','POST',input,201);await balance(1000000);
 const withdrawal=await request('/transactions','POST',{...input,type:'expense',category_id:id('expense_transfer_out'),amount:200000},201);await balance(800000);
 const interest=await request('/transactions','POST',{...input,category_id:id('income_interest'),amount:10000},201);await balance(810000);
 await request(`/transactions/${deposit.id}`,'PATCH',{...input,amount:500000});await balance(410000);
 await request(`/transactions/${withdrawal.id}`,'DELETE',undefined,204);await balance(610000);
 await request(`/transactions/${interest.id}`,'DELETE',undefined,204);await balance(600000);
 await request('/transactions','POST',{...input,category_id:id('income_salary')},400);
 const now=new Date(),start=new Date(now.getFullYear(),now.getMonth(),1),end=new Date(now.getFullYear(),now.getMonth()+1,1);
 const config={name:'API proof budget',limit_amount:1000000,wallet_id:basic.id,category_id:id('expense_food'),start_at:start.toISOString(),end_at:end.toISOString()};
 const budget=await request('/budgets','POST',config,201);budgets.push(budget.id);
 await request('/budgets','POST',config,409);
 const txInput={...input,wallet_id:basic.id,type:'expense',category_id:id('expense_food_coffee'),amount:50000};
 const expense=await request('/transactions','POST',txInput,201);
 const spent=async(expected)=>{const result=await request('/budgets');assert.equal(result.items.find(item=>item.id===budget.id).spent,expected);};
 await spent(50000);
 await request(`/transactions/${expense.id}`,'PATCH',{...txInput,amount:75000});await spent(75000);
 await request(`/budgets/${budget.id}`,'PATCH',{...config,limit_amount:2000000});await spent(75000);
 await request(`/transactions/${expense.id}`,'PATCH',{...txInput,included_in_reports:false});await spent(0);
 await request(`/transactions/${expense.id}`,'DELETE',undefined,204);await spent(0);
 await request(`/budgets/${budget.id}`,'DELETE',undefined,204);budgets.pop();
 assert.ok((await request('/transactions')).some(row=>row.id===deposit.id),'budget deletion changed unrelated ledger');
 const unauthorized=await fetch(`${origin}/api/v1/budgets`);assert.equal(unauthorized.status,401);
 console.log('PASS: FE proxy → API → PostgreSQL. Goal date/deposit/withdrawal/interest/edit/delete/balance; real catalog rejection; budget CRUD/overlap/child category/report flag/recalculation; unauthorized read.');
} finally {
 for(const id of budgets)await request(`/budgets/${id}`,'DELETE',undefined,204);
 for(const id of wallets)await request(`/wallets/${id}`,'DELETE',undefined,204);
 console.log('Cleanup: only test-created wallets, their transactions and budgets removed.');
}
