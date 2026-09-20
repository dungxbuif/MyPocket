import { useEffect, useState } from "react";
import { Tags } from "lucide-react";
import { Text } from "../atoms/Text";
import { BudgetGauge } from "../atoms/Progress";
import { MetricBox } from "../atoms/MetricBox";
import { SectionTitle } from "../atoms/SectionTitle";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { StatusMessage } from "../atoms/StatusMessage";
import { BudgetProgressItem } from "../molecules/BudgetProgressItem";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { fetchBudgets, saveBudget, deleteBudget, type Budget, type BudgetSummary } from "../../services/budgets";
import { fetchWallets, type Wallet } from "../../services/wallets";
import { fetchCategories, type Category } from "../../services/categories";
import { formatVND, ratioPercent } from "../utils/format";
import { budgetDateInput, budgetDateLabel as dateLabel } from "../../services/budgetDates";

type Draft={name:string;amount:string;wallet:string;category:string;start:string;end:string};
function initialDraft(budget?:Budget):Draft {
  const now=new Date();
  return {name:budget?.name??"",amount:budget?String(budget.limit_amount):"",wallet:budget?.wallet_id??"",category:budget?.category_id??"",start:dateLabel(budget?new Date(budget.start_at):new Date(now.getFullYear(),now.getMonth(),1)),end:dateLabel(budget?new Date(Date.parse(budget.end_at)-1):new Date(now.getFullYear(),now.getMonth()+1,0))};
}

export function BudgetsPanel({ masked, refreshKey=0, onChanged }: { masked:boolean; refreshKey?:number; onChanged:()=>void }) {
  const [data,setData]=useState<BudgetSummary|null>(null);
  const [wallets,setWallets]=useState<Wallet[]>([]);
  const [categories,setCategories]=useState<Category[]>([]);
  const [loading,setLoading]=useState(true);
  const [error,setError]=useState("");
  const [reload,setReload]=useState(0);
  const [editor,setEditor]=useState<Budget|"new"|null>(null);
  const [draft,setDraft]=useState<Draft>(()=>initialDraft());
  const [saving,setSaving]=useState(false);
  const [formError,setFormError]=useState("");
  useEffect(()=>{
    let cancelled=false;
    setLoading(true);setError("");
    Promise.all([fetchBudgets(),fetchWallets(),fetchCategories()]).then(([budgets,nextWallets,nextCategories])=>{
      if(cancelled)return;
      setData(budgets);setWallets(nextWallets.filter(wallet=>wallet.type!=="credit"));setCategories(nextCategories.filter(category=>category.kind==="expense"));
    }).catch(()=>{if(!cancelled)setError("Không thể tải ngân sách. Vui lòng thử lại.");}).finally(()=>{if(!cancelled)setLoading(false);});
    return ()=>{cancelled=true;};
  },[refreshKey,reload]);
  const open=(budget?:Budget)=>{setFormError("");setDraft(initialDraft(budget));setEditor(budget??"new");};
  const saved=()=>{setEditor(null);setReload(value=>value+1);onChanged();};
  const submit=async()=>{
    if(saving || (editor!=="new" && editor?.ended))return;
    const amount=Number(draft.amount);
    const start=new Date(`${draft.start}T00:00:00`), end=new Date(`${draft.end}T00:00:00`);
    end.setDate(end.getDate()+1);
    if(!draft.name.trim() || !Number.isSafeInteger(amount) || amount<=0 || !Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime()) || end<=start) {setFormError("Nhập tên, hạn mức nguyên dương và khoảng ngày hợp lệ.");return;}
    setSaving(true);setFormError("");
    try {await saveBudget({name:draft.name.trim(),limit_amount:amount,wallet_id:draft.wallet||null,category_id:draft.category||null,...budgetDateInput(draft.start,draft.end,editor&&editor!=="new"?editor:undefined)},editor&&editor!=="new"?editor.id:undefined);saved();}
    catch(error){setFormError(error instanceof Error?error.message:"Không thể lưu ngân sách.");}
    finally {setSaving(false);}
  };
  const remove=async()=>{
    if(saving || !editor || editor==="new" || !window.confirm("Xóa ngân sách này? Giao dịch và số dư ví không thay đổi."))return;
    setSaving(true);setFormError("");
    try {await deleteBudget(editor.id);saved();}catch(error){setFormError(error instanceof Error?error.message:"Không thể xóa ngân sách.");}finally {setSaving(false);}
  };
  const locked=saving || (editor!==null && editor!=="new" && editor.ended);
  const money=(value:number)=>masked?"••••••":formatVND(value);
  return <>
    <SectionTitle title="Ngân sách" action="Thêm" onAction={()=>open()} />
    {loading?<StatusMessage>Đang tải ngân sách...</StatusMessage>:error?<StatusMessage tone="danger">{error}<BaseButton variant="ghost" onClick={()=>setReload(value=>value+1)}>Thử lại</BaseButton></StatusMessage>:data?<>
      {data.items.length===0?<StatusMessage variant="plain">Chưa có ngân sách.</StatusMessage>:<>
        <SurfaceCard padding="lg"><Text weight="semibold">Các ngân sách đang chạy</Text><BudgetGauge value={ratioPercent(data.spent,data.limit_amount)} label="Đã dùng" /><div className="grid grid-cols-3 gap-2"><MetricBox label="Ngân sách" value={money(data.limit_amount)} /><MetricBox label="Đã chi" value={money(data.spent)} danger /><MetricBox label="Còn lại" value={money(data.limit_amount-data.spent)} /></div></SurfaceCard>
        {data.items.map(budget=>{
          const category=categories.find(item=>item.id===budget.category_id);
          const icon=category?categoryPresentationFor(category.system_key??category.icon_key).icon:Tags;
          return <div key={budget.id} className="space-y-2"><BudgetProgressItem masked={masked} budget={{name:budget.name,spent:budget.spent,limit:budget.limit_amount,icon}} /><Text size="xs" tone="secondary">{dateLabel(new Date(budget.start_at))} — {dateLabel(new Date(Date.parse(budget.end_at)-1))} · {budget.ended?"Đã kết thúc":Date.parse(budget.start_at)>Date.now()?"Sắp bắt đầu":`Còn ${budget.days_remaining} ngày`}</Text>{!budget.ended?<Text size="sm">Gợi ý chi/ngày: {money(Math.max(0,Math.floor((budget.limit_amount-budget.spent)/Math.max(1,budget.days_remaining))))}</Text>:null}{budget.spent>=budget.limit_amount*0.75?<Text tone={budget.spent>budget.limit_amount?"danger":"secondary"}>{budget.spent>budget.limit_amount?"Đã vượt ngân sách":"Đã dùng ít nhất 75% ngân sách"}</Text>:null}<BaseButton variant="chip" onClick={()=>open(budget)}>{budget.ended?"Xem / Xóa":"Sửa / Xóa"}</BaseButton></div>;
        })}
      </>}
    </>:null}
    {editor?<BaseBottomSheet title={editor==="new"?"Thêm ngân sách":"Ngân sách"} onClose={()=>{if(!saving)setEditor(null);}} closeLabel="Hủy" presentation="form" closingDisabled={saving}><form className="space-y-3" onSubmit={event=>{event.preventDefault();void submit();}}>
      {formError?<StatusMessage tone="danger">{formError}</StatusMessage>:null}
      {editor!=="new" && editor.ended?<StatusMessage>Ngân sách đã kết thúc, chỉ có thể xem hoặc xóa.</StatusMessage>:null}
      <FormField label="Tên ngân sách"><BaseTextInput required disabled={locked} maxLength={200} value={draft.name} onChange={event=>setDraft({...draft,name:event.target.value})} /></FormField>
      <FormField label="Hạn mức (VND)"><BaseTextInput required disabled={locked} inputMode="numeric" value={draft.amount} onChange={event=>setDraft({...draft,amount:event.target.value.replace(/\D/g,"")})} /></FormField>
      <FormField label="Ví"><BaseSelect disabled={locked} value={draft.wallet} onChange={event=>setDraft({...draft,wallet:event.target.value})}><option value="">Tất cả ví</option>{wallets.map(wallet=><option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></FormField>
      <FormField label="Nhóm chi"><BaseSelect disabled={locked} value={draft.category} onChange={event=>setDraft({...draft,category:event.target.value})}><option value="">Tất cả nhóm chi</option>{categories.map(category=><option key={category.id} value={category.id}>{category.name}</option>)}</BaseSelect></FormField>
      <FormField label="Từ ngày"><BaseTextInput required type="date" disabled={locked} value={draft.start} onChange={event=>setDraft({...draft,start:event.target.value})} /></FormField>
      <FormField label="Đến hết ngày"><BaseTextInput required type="date" disabled={locked} min={draft.start} value={draft.end} onChange={event=>setDraft({...draft,end:event.target.value})} /></FormField>
      <Text size="xs" tone="secondary">Khoảng ngày tùy chỉnh, không tự lặp. Xóa ví hoặc nhóm được chọn cũng xóa cấu hình ngân sách này.</Text>
      <div className="flex gap-2"><BaseButton type="submit" loading={saving} disabled={locked}>Lưu ngân sách</BaseButton>{editor!=="new"?<BaseButton variant="danger" disabled={saving} onClick={()=>void remove()}>Xóa ngân sách</BaseButton>:null}</div>
    </form></BaseBottomSheet>:null}
  </>;
}
