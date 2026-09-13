import { useEffect, useMemo, useState } from "react";
import { ArrowLeft, ChevronRight, FolderTree } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { fetchCategories, type Category } from "../../services/categories";

const GROUP_KIND_LABELS: Record<string, string> = { expense: "Khoản chi", income: "Khoản thu", debt: "Vay/Nợ" };

export function GroupManagementPanel() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [state, setState] = useState<"loading" | "ready" | "error">("loading");
  useEffect(() => { fetchCategories().then((items) => { setCategories(items); setState("ready"); }).catch(() => setState("error")); }, []);
  const roots = useMemo(() => categories.filter((item) => !item.parent_id), [categories]);
  return <section className="space-y-3">
    <div className="flex items-center gap-3"><Link to="/account" className="grid h-10 w-10 place-items-center rounded-full border border-[#e3e2e2] bg-white"><ArrowLeft size={18} /></Link><div><h1 className="text-xl font-bold">Quản lý nhóm</h1><p className="text-xs text-[#6f7a6b]">Nhóm hệ thống và nhóm của bạn</p></div></div>
    {state === "loading" ? <SurfaceCard className="p-5 text-sm text-[#6f7a6b]">Đang tải nhóm...</SurfaceCard> : null}
    {state === "error" ? <SurfaceCard className="p-5 text-sm text-[#ba1a1a]">Không tải được danh sách nhóm.</SurfaceCard> : null}
    {state === "ready" && roots.map((root) => <SurfaceCard key={root.id} className="p-3" radius="lg"><div className="flex items-center gap-3"><IconBadge icon={FolderTree} /><div className="min-w-0 flex-1"><p className="font-bold">{root.name}</p><p className="text-xs text-[#6f7a6b]">{GROUP_KIND_LABELS[root.kind] ?? root.kind}</p></div><ChevronRight size={17} className="text-[#6f7a6b]" /></div><div className="mt-2 space-y-1 border-l-2 border-[#e3e2e2] pl-5">{categories.filter((item) => item.parent_id === root.id).map((child) => <div key={child.id} className="flex items-center gap-2 rounded-2xl px-2 py-2"><span className="h-2 w-2 rounded-full bg-[#006e1c]" /><span className="text-sm">{child.name}</span></div>)}</div></SurfaceCard>)}
    {state === "ready" && roots.length === 0 ? <SurfaceCard className="p-5 text-sm text-[#6f7a6b]">Chưa có nhóm nào.</SurfaceCard> : null}
  </section>;
}
