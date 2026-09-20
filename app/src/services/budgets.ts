import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type BudgetInput = {name:string; limit_amount:number; wallet_id?:string|null; category_id?:string|null; start_at:string; end_at:string};
export type Budget = BudgetInput & {id:string; spent:number; days_remaining:number; ended:boolean};
export type BudgetSummary = {items:Budget[]; limit_amount:number; spent:number};
const PATH="/api/v1/budgets";
export const fetchBudgets=()=>apiRequest<BudgetSummary>(PATH,{},getStoredToken());
export const saveBudget=(input:BudgetInput,id?:string)=>apiRequest<Budget>(id?`${PATH}/${id}`:PATH,{method:id?"PATCH":"POST",body:JSON.stringify(input)},getStoredToken());
export const deleteBudget=(id:string)=>apiRequest<void>(`${PATH}/${id}`,{method:"DELETE"},getStoredToken());
