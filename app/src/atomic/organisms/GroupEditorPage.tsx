import { Trash2 } from "lucide-react";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { IconButton } from "../atoms/IconButton";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { CategoryEditForm } from "../molecules/CategoryEditForm";
import { PageBackHeader } from "../molecules/PageBackHeader";
import { createCategory, deleteCategory, fetchCategories, updateCategory, updateCategoryWallets, type Category, type CategoryInput } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";

const EDITOR_TEXT = {
  createTitle: "Nhóm mới",
  editTitle: "Sửa nhóm",
  back: "Quay lại",
  loading: "Đang tải nhóm...",
  loadError: "Không tải được dữ liệu nhóm.",
  saveError: "Không thể lưu thay đổi. Vui lòng thử lại.",
  deleteError: "Không thể xóa nhóm. Vui lòng thử lại.",
  deleteConfirm: "Xóa nhóm này? Nhóm con sẽ trở thành nhóm cấp gốc.",
  deleteLabel: "Xóa nhóm",
} as const;

export function GroupEditorPage() {
  const { pathname } = useLocation(); const navigate = useNavigate();
  const id = pathname.match(/groups\/([^/]+)\/edit/)?.[1];
  const [items, setItems] = useState<Category[]>([]); const [wallets, setWallets] = useState<Wallet[]>([]); const [saving, setSaving] = useState(false); const [loading, setLoading] = useState(true); const [error, setError] = useState("");
  useEffect(() => { void Promise.all([fetchCategories(), fetchWallets()]).then(([categories, ownerWallets]) => { setItems(categories); setWallets(ownerWallets); }).catch(() => setError(EDITOR_TEXT.loadError)).finally(() => setLoading(false)); }, []);
  const category = id ? items.find((item) => item.id === id) ?? null : null;
  const save = async (input: CategoryInput) => { setSaving(true); setError(""); try { if (category?.is_system) { await updateCategoryWallets(category.id, input.wallet_ids); } else if (category) { await updateCategory(category.id, input); } else { await createCategory(input); } await navigate({ to: "/account/groups" }); } catch { setError(EDITOR_TEXT.saveError); } finally { setSaving(false); } };
  const remove = async () => { if (!category || !window.confirm(EDITOR_TEXT.deleteConfirm)) return; setSaving(true); setError(""); try { await deleteCategory(category.id); await navigate({ to: "/account/groups" }); } catch { setError(EDITOR_TEXT.deleteError); } finally { setSaving(false); } };
  const trailing = category && !category.is_system ? <IconButton label={EDITOR_TEXT.deleteLabel} variant="bare" disabled={saving} onClick={() => void remove()}><Trash2 size={19} className="text-rose-600" /></IconButton> : undefined;
  return <section className="space-y-5 px-1"><PageBackHeader title={category ? EDITOR_TEXT.editTitle : EDITOR_TEXT.createTitle} backTo="/account/groups" backLabel={EDITOR_TEXT.back} trailing={trailing} />{loading ? <SurfaceCard padding="md" className="text-sm text-slate-500">{EDITOR_TEXT.loading}</SurfaceCard> : null}{error ? <SurfaceCard padding="md" className="text-sm text-rose-600" role="alert">{error}</SurfaceCard> : null}{!loading && (!id || category) ? <CategoryEditForm category={category} categories={items} wallets={wallets} saving={saving} readOnly={Boolean(category?.is_system)} onSave={save} onCancel={() => void navigate({ to: "/account/groups" })} /> : null}</section>;
}
