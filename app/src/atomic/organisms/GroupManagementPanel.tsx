import { useEffect, useMemo, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { BaseButton } from "../atoms/BaseButton";
import { Heading } from "../atoms/Heading";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { CategoryEditForm } from "../molecules/CategoryEditForm";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { BaseCategoryTree, type BaseCategoryTreeItem } from "../molecules/BaseCategoryTree";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { createCategory, deleteCategory, fetchCategories, updateCategory, type Category, type CategoryInput } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";

const GROUP_TEXT = {
  title: "Quản lý nhóm",
  add: "Nhóm mới",
  loading: "Đang tải nhóm...",
  empty: "Chưa có nhóm nào.",
  retry: "Thử lại",
  loadError: "Không tải được danh sách nhóm.",
  saveError: "Không thể lưu nhóm. Dữ liệu trong form được giữ nguyên.",
  deleteError: "Không thể xóa nhóm.",
  deleteConfirm: "Xóa nhóm cá nhân này? Hành động này không thể hoàn tác.",
  expense: "Khoản chi",
  income: "Khoản thu",
  debt: "Vay/Nợ",
  back: "Quay lại Tài khoản",
  backLabel: "Quay lại",
  walletAll: "Áp dụng tất cả ví",
  walletCount: (count: number) => `Hoạt động trong ${count} ví`,
  createTitle: "Nhóm mới",
  editTitle: "Sửa nhóm",
  closeEditor: "Đóng form nhóm",
} as const;

const KIND_OPTIONS = [
  { value: "expense", label: GROUP_TEXT.expense },
  { value: "income", label: GROUP_TEXT.income },
  { value: "debt", label: GROUP_TEXT.debt },
] as const;
type CategoryKind = typeof KIND_OPTIONS[number]["value"];
type ScreenState = "loading" | "ready" | "error";

export function GroupManagementPanel() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [kind, setKind] = useState<CategoryKind>("expense");
  const [state, setState] = useState<ScreenState>("loading");
  const [editing, setEditing] = useState<Category | null | undefined>(undefined);
  const [saving, setSaving] = useState(false);
  const [actionError, setActionError] = useState("");
  const load = async () => { setState("loading"); try { const [items, ownerWallets] = await Promise.all([fetchCategories(), fetchWallets()]); setCategories(items); setWallets(ownerWallets); setState("ready"); } catch { setState("error"); } };
  useEffect(() => { void load(); }, []);
  const roots = useMemo(() => categories.filter((item) => !item.parent_id && item.kind === kind).sort((left, right) => Number(categories.some((item) => item.parent_id === right.id)) - Number(categories.some((item) => item.parent_id === left.id))), [categories, kind]);
  const save = async (input: CategoryInput) => { setSaving(true); setActionError(""); try { if (editing) { await updateCategory(editing.id, input); } else { await createCategory(input); } await load(); setEditing(undefined); } catch { setActionError(GROUP_TEXT.saveError); } finally { setSaving(false); } };
  const remove = async (item: Category) => { if (!window.confirm(GROUP_TEXT.deleteConfirm)) return; setSaving(true); setActionError(""); try { await deleteCategory(item.id); await load(); setEditing(undefined); } catch { setActionError(GROUP_TEXT.deleteError); } finally { setSaving(false); } };
  return <section className="space-y-3">
    <header className="grid grid-cols-[1fr_auto_1fr] items-center px-1 pb-1"><Link to="/account" aria-label={GROUP_TEXT.back}><BaseButton variant="outline" size="sm" className="gap-1 px-3 text-slate-700"><ArrowLeft size={14} />{GROUP_TEXT.backLabel}</BaseButton></Link><Heading as="h1" size="section" className="text-center">{GROUP_TEXT.title}</Heading><span /></header>
    <SegmentedControl value={kind} options={KIND_OPTIONS} onChange={setKind} />
    <BaseButton variant="outline" className="w-full gap-2" onClick={() => { setActionError(""); setEditing(null); }}><span className="grid h-5 w-5 place-items-center rounded-full bg-emerald-500 text-white">+</span>{GROUP_TEXT.add}</BaseButton>
    {actionError ? <SurfaceCard padding="md" className="text-sm text-rose-600" role="alert">{actionError}</SurfaceCard> : null}
    {state === "loading" ? <SurfaceCard padding="md" className="text-sm text-slate-500">{GROUP_TEXT.loading}</SurfaceCard> : null}
    {state === "error" ? <SurfaceCard padding="md" className="space-y-3 text-sm text-rose-600"><p>{GROUP_TEXT.loadError}</p><BaseButton variant="secondary" size="sm" onClick={() => void load()}>{GROUP_TEXT.retry}</BaseButton></SurfaceCard> : null}
    {state === "ready" && roots.map((root) => <BaseCategoryTree key={root.id} root={treeItem(root)} children={categories.filter((item) => item.parent_id === root.id).map(treeItem)} onSelect={(id) => setEditing(categories.find((item) => item.id === id))} />)}
    {state === "ready" && roots.length === 0 ? <SurfaceCard padding="md" className="text-sm text-slate-500">{GROUP_TEXT.empty}</SurfaceCard> : null}
    {editing !== undefined ? <BaseBottomSheet title={editing ? GROUP_TEXT.editTitle : GROUP_TEXT.createTitle} closeLabel={GROUP_TEXT.closeEditor} onClose={() => { if (!saving) { setEditing(undefined); setActionError(""); } }}><CategoryEditForm category={editing} categories={categories} wallets={wallets} saving={saving} onSave={save} onDelete={editing ? () => remove(editing) : undefined} onCancel={() => { setEditing(undefined); setActionError(""); }} /></BaseBottomSheet> : null}
  </section>;
}

function treeItem(item: Category): BaseCategoryTreeItem {
  const presentation = categoryPresentationFor(item.icon_key || item.system_key);
  return { id: item.id, name: item.name, subtitle: item.wallet_ids.length ? GROUP_TEXT.walletCount(item.wallet_ids.length) : GROUP_TEXT.walletAll, isSystem: item.is_system, isEditable: Boolean(item.owner_id), ...presentation };
}
