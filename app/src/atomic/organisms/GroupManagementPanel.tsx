import { Text } from "../atoms/Text";
import { useEffect, useMemo, useState } from "react";
import { PlusCircle } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { PageBackHeader } from "../molecules/PageBackHeader";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { BaseButton } from "../atoms/BaseButton";
import { Heading } from "../atoms/Heading";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { BaseSwitch } from "../atoms/BaseSwitch";
import { BaseCategoryTree, type BaseCategoryTreeItem } from "../molecules/BaseCategoryTree";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { fetchCategories, type Category } from "../../services/categories";

const GROUP_TEXT = {
  title: "Quản lý nhóm",
  add: "Nhóm mới",
  loading: "Đang tải nhóm...",
  empty: "Chưa có nhóm nào.",
  retry: "Thử lại",
  loadError: "Không tải được danh sách nhóm.",
  expense: "Khoản chi",
  income: "Khoản thu",
  debt: "Vay/Nợ",
  back: "Quay lại Tài khoản",
  backLabel: "Quay lại",
  showInactive: "Hiển thị nhóm không hoạt động",
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
  const [kind, setKind] = useState<CategoryKind>("expense");
  const [showInactive, setShowInactive] = useState(false);
  const [state, setState] = useState<ScreenState>("loading");
  const navigate = useNavigate();
  const load = async () => { setState("loading"); try { setCategories(await fetchCategories()); setState("ready"); } catch { setState("error"); } };
  useEffect(() => { void load(); }, []);
  const visibleCategories = useMemo(() => {
    if (showInactive) return categories;
    const visibleIDs = new Set(categories.filter(item => item.is_active !== false).map(item => item.id));
    for (const item of categories) {
      if (item.is_active !== false) {
        let parentID = item.parent_id;
        while (parentID) {
          visibleIDs.add(parentID);
          parentID = categories.find(candidate => candidate.id === parentID)?.parent_id ?? null;
        }
      }
    }
    return categories.filter(item => visibleIDs.has(item.id));
  }, [categories, showInactive]);
  const roots = useMemo(() => visibleCategories.filter((item) => !item.parent_id && item.kind === kind).sort((left, right) => Number(visibleCategories.some((item) => item.parent_id === right.id)) - Number(visibleCategories.some((item) => item.parent_id === left.id))), [visibleCategories, kind]);
  return <section className="space-y-3">
    <PageBackHeader title={GROUP_TEXT.title} backTo="/account" backLabel={GROUP_TEXT.back} />
    <SegmentedControl value={kind} options={KIND_OPTIONS} onChange={setKind} />
    <SurfaceCard padding="md"><BaseSwitch label={GROUP_TEXT.showInactive} checked={showInactive} onChange={event => setShowInactive(event.target.checked)} /></SurfaceCard>
    <BaseButton variant="outline" className="w-full" onClick={() => void navigate({ to: "/account/groups/new", search: { kind } })}><PlusCircle size={20} />{GROUP_TEXT.add}</BaseButton>
    {state === "loading" ? <SurfaceCard padding="md"><Text tone="secondary">{GROUP_TEXT.loading}</Text></SurfaceCard> : null}
    {state === "error" ? <SurfaceCard padding="md" className="space-y-3"><Text tone="danger">{GROUP_TEXT.loadError}</Text><BaseButton variant="secondary" size="sm" onClick={() => void load()}>{GROUP_TEXT.retry}</BaseButton></SurfaceCard> : null}
    {state === "ready" && roots.map((root) => <BaseCategoryTree key={root.id} root={treeItem(root, !showInactive && root.is_active === false)} children={visibleCategories.filter((item) => item.parent_id === root.id).map(item => treeItem(item, !showInactive && item.is_active === false))} onSelect={(id) => void navigate({ to: "/account/groups/$categoryId/edit", params: { categoryId: id } })} />)}
    {state === "ready" && roots.length === 0 ? <SurfaceCard padding="md"><Text tone="secondary">{GROUP_TEXT.empty}</Text></SurfaceCard> : null}
  </section>;
}

function treeItem(item: Category, contextOnly = false): BaseCategoryTreeItem {
  const presentation = categoryPresentationFor(item.icon_key || item.system_key);
  const state = item.is_active === false ? " · Đã lưu trữ" : "";
  return { id: item.id, name: item.name, subtitle: `${item.wallet_ids.length ? `Hoạt động trong ${item.wallet_ids.length} ví` : "Áp dụng tất cả ví"}${state}`, isSystem: item.is_system, isEditable: !contextOnly, ...presentation };
}
